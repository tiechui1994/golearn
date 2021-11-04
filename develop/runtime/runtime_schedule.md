### 调度循环

任何 goroutine 被调度运行起来都是通过 schedule() -> execute() -> gogo() 调用链, 而且这个调用链中的函数一直没有
返回. 一个 goroutine 从调度到退出的路径:

```cgo
schedule()->execute()->gogo()->xxx()->goexit()->goexit1()->mcall()->goexit0()->schedule()
```

其中 `schedule()->execute()->gogo()` 是在 g0 上执行的. `xxx()->goexit()->goexit1()->mcall()` 是在 curg 
上执行的. `goexit0()->schedule()` 又是在 g0 上执行的.

一轮调度是从调用 schedule 开始, 然后经过一系列代码的执行到最后再次通过调用 schedule 函数进行新一轮的调度. 

从一轮调度到新一轮调度这一个过程称为一个调度循环. 这里的调度循环是指某一个工作线程的调度循环, 而同一个 go 程序中可能存在
多个工作线程, 每个工作线程都有自己的调度循环.

每次执行mcall切换到g0栈时都是切换到 g0.sched.sp 所指的固定位置, 这样 g0 栈不会增加的.

### goroutine 调度策略

所谓的 goroutine 调度, 是指程序代码按照一定算法在适当的时候选出合适的 goroutine 并放到 CPU 上去运行的过程. 调度系统
需要解决的三大问题:

1. 调度时机: 什么时候发生调度?

2. 调度策略: 使用什么策略挑选下一个进入运行的 goroutine?

3. 切换时机: 如何把挑选出来的 goroutine 放到 CPU 上运行?

schedule 函数分三步分别查找各种运行队列中寻找可运行的 goroutine:

- 从全局队列中寻找 goroutine. 为了保证调度公平性, 每个工作线程每经过61次调度就优先从全局队列中找到一个 goroutine 运
行.

- 从工作线程本地运行队列中查找 goroutine

- 从其他工作线程的运行队列中偷取 goroutine. 如果上一步也没有找到需要运行的 goroutine, 则调用 findrunnable 从其他
工作线程的运行队列中偷取 goroutine, findrunnable 函数在偷取之前会再次尝试从全局运行队列和当前线程本地运行队列中查找需
要运行的 goroutine.


全局运行队列中获取 goroutine:

globrunqget() 函数的 `_p_` 是当前工作线程绑定的 p, 第二个参数 max 表示最多可以从全局队列拿多少个 g 到当前工作线程
的本地运行队列.

```cgo
func globrunqget(_p_ *p, max int32) *g {
    // 全局运行队列为空
	if sched.runqsize == 0 {
		return nil
	}
    
    // 根据 p 的数量平分全局运行队列中的 goroutines
	n := sched.runqsize/gomaxprocs + 1
	if n > sched.runqsize {
		n = sched.runqsize // 最多获取全局队列中 goroutine 总量
	}
	if max > 0 && n > max {
		n = max // 最多获取 max 个 goroutine
	}
	if n > int32(len(_p_.runq))/2 {
		n = int32(len(_p_.runq)) / 2 // 最多只能获取本地队列容量的一半
	}

	sched.runqsize -= n 
    
    // 从全局队列 pop 出一个
	gp := sched.runq.pop()
	n--
	
	// 剩余的直接存储到 _p_ 的本地队列当中
	for ; n > 0; n-- {
		gp1 := sched.runq.pop()
		runqput(_p_, gp1, false)
	}
	return gp
}

```

从工作线程本地运行队列当中获取:

runqget() 的参数是本地运行队列 `_p_`. 工作线程的本地运行队列分为两部分, 一部分是 p 的 runq, runhead, runtail
三个成员组成的无锁循环队列, 该队列最多可存储 256 个 goroutine; 另一部分是 p 的 runnext 成员, 它是指向一个 g 结构体
对象的指针, 最多包含 1 个 goroutine.

本地运行队列优先从 runnext 当中获取 goroutine, 然后从循环队列当中获取 goroutine.

```cgo
func runqget(_p_ *p) (gp *g, inheritTime bool) {
    // 从 runnext 当中获取
    for {
        next := _p_.runnext
        if next == 0 {
            break
        }
        if _p_.runnext.cas(next, 0) {
            return next.ptr(), true
        }
    }
    
    // 从队列当中获取
    for {
        h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with other consumers
        t := _p_.runqtail
        if t == h {
            return nil, false // 当前队列不存在可用的 goroutine
        }
        gp := _p_.runq[h%uint32(len(_p_.runq))].ptr()
        if atomic.CasRel(&_p_.runqhead, h, h+1) { // cas-release, commits consume
            return gp, false // 注意: 当从无锁队列中获取 inheritTime 是 false, 该值会导致 p 的 schedtick 增加
        }
    }
}
```

atomic.LoadAcq 和 atomic.CasRel 分别提供了 load-acquire 和 cas-release 语义.

atomic.LoadAcq:

- 原子读取, 保证读取过程中不会有其他线程对该变量进行写入

- 位于 atomic.LoadAcq 之后的代码, 对内存的读取和写入必须在 atomic.LoadAcq 读取完成后才能执行, 编译器和CPU都不能
打乱.

- 当前线程执行 atomic.LoadAcq 时可以读取到其他线程最近一次通过 atomic.CasRel 对同一个变量值写入的值.


atomic.CasRel:

- 原子的执行比较并交换的操作

- 位于 atomic.CasRel 之前的代码, 对内存的读取和写入必须在 atomic.CasRel 对内存的写入之前完成, 编译器和CPU不能打乱
这个顺序.

- 线程执行atomic.CasRel 完成后其他线程通过 atomic.LoadAcq 读取同一变量可以读到最新的值.


当全局队列和本地队列当中找不到要执行的 goroutine 时, 这时就要从其他工作线程的本地运行队列当中盗取 goroutine.

findrunnable() 函数负责处理与盗取相关的逻辑.

```cgo
// Tries to steal from other P's, get g from global queue, poll network.
func findrunnable() (gp *g, inheritTime bool) {
    _g_ := getg() // g0
    
    // 这里的条件和 handoffp 中的条件必须一致: 如果 findrunnable 将返回G以运行,
    // 则 handoffp 必须启动一个 M.
top:
    _p_ := _g_.m.p.ptr()
    
    ... 
    
    now, pollUntil, _ := checkTimers(_p_, 0) // 运行计时器
    
    ... 
    
    // local runq
    // 再次查看一下本地运行队列是否有需要运行的 goroutine
    if gp, inheritTime := runqget(_p_); gp != nil {
        return gp, inheritTime
    }
    
    // global runq
    // 再次查看一下全局运行队列是否有需要运行的 goroutine
    if sched.runqsize != 0 {
        lock(&sched.lock)
        gp := globrunqget(_p_, 0)
        unlock(&sched.lock)
        if gp != nil {
            return gp, false
        }
    }
    
    // 检查 netpoll 当中是否存在就绪的 g
    // netpollinited() 返回当前 netpoll 是否就绪
    if netpollinited() && atomic.Load(&netpollWaiters) > 0 && atomic.Load64(&sched.lastpoll) != 0 {
        if list := netpoll(0); !list.empty() { // non-blocking
            gp := list.pop()
            // injectglist 将 list 中的每个可运行 G 添加到某个运行队列中, 并清除 glist.
            // 如果当前 M 没有绑定 _p_, 向全局队列批量添加 len(list) 个 G, 同时启动 min(len(list), sched.npidle)
            // 个 M;
            // 如果当前 M 绑定了 _p_, 则分 min(len(list), sched.npidle) 次向全局队列添加 G, 同时启动 
            // min(len(list), sched.npidle) 个 M, 对于剩余的 G 则添加到当前 _p_ 的本地运行队列中.
            injectglist(&list) 
            casgstatus(gp, _Gwaiting, _Grunnable)
            if trace.enabled {
                traceGoUnpark(gp, 0)
            }
            return gp, false
        }
    }
    
    // Steal work from other P's.
    procs := uint32(gomaxprocs)
    ranTimer := false
    // 当 "处于自旋状态的M的数量" >= "处于运行当中的P的数量" 时, 阻塞运行
    // 当 GOMAXPROCS>>1 但是程序并行度很低, 这对于防止过度的 CPU 消耗是必要的
    // 这个判断主要是为了防止因为寻找可运行的goroutine而消耗太多的CPU.
    // 因为已经有足够多的工作线程正在寻找可运行的goroutine, 让他们去找就好了, 自己偷个懒去睡觉.
    if !_g_.m.spinning && 2*atomic.Load(&sched.nmspinning) >= procs-atomic.Load(&sched.npidle) {
        goto stop
    }
    
    // 让当前的线程进入自旋状态 
    if !_g_.m.spinning {
        _g_.m.spinning = true
        atomic.Xadd(&sched.nmspinning, 1)
    }
    
    for i := 0; i < 4; i++ {
        // stealOrder.start() 开启一次随机偷取的枚举
        // enum.next() 是计算下一个位置
        // enum.done() 是否结束
        // enum.position() 获取当前的位置
        for enum := stealOrder.start(fastrand()); !enum.done(); enum.next() {
            if sched.gcwaiting != 0 {
                goto top
            }
            stealRunNextG := i > 2 // first look for ready queues with more than 1 g
            p2 := allp[enum.position()] // allp 当中的某个位置
            if _p_ == p2 { // p2 刚好是当前的线程绑定的 p, 则不用查找, 本身就没有
                continue
            }
            
            // 从 p2 当中偷取(如果偷取到, 至少应该是一个, 剩下的会保存在 _p_ 当中的)
            // stealRunNextG 表示是否偷取 p2.runnext 
            if gp := runqsteal(_p_, p2, stealRunNextG); gp != nil {
                return gp, false
            }
            
            // 从 p2 当中没有偷取到. 
            // i = 2 时, shouldStealTimers() 决定是否能从 p2 当中偷取 g
            // 当 p2.status != _Prunning || (p2.m.curg.status == _Grunning && preempt)
            // i = 3 时, 直接从 p2 当中偷取
            if i > 2 || (i > 1 && shouldStealTimers(p2)) {
                // 为 p2 运行计时器, tnow是当前时间, w 是下一个运行计时器时间, ran是否运行了计时器
                tnow, w, ran := checkTimers(p2, now)
                now = tnow
                if w != 0 && (pollUntil == 0 || w < pollUntil) {
                    pollUntil = w
                }
                if ran {
                    // 运行 timer 可能导致任意数量处于ready当中的 G 被添加到这个 P 的本地运行队列中. 
                    // 这会导致 runqsteal 总是有空间添加盗取 G 的这一假设无效. 所以现在检查是否有本地 G 运行。
                    if gp, inheritTime := runqget(_p_); gp != nil {
                        return gp, inheritTime
                    }
                    ranTimer = true
                }
            }
        }
    }
    if ranTimer {
        // Running a timer may have made some goroutine ready.
        goto top
    }
    
stop:
     
    // GC 检查
    // We have nothing to do. If we're in the GC mark phase, can
    // safely scan and blacken objects, and have work to do, run
    // idle-time marking rather than give up the P.
    if gcBlackenEnabled != 0 && _p_.gcBgMarkWorker != 0 && gcMarkWorkAvailable(_p_) {
        _p_.gcMarkWorkerMode = gcMarkWorkerIdleMode
        gp := _p_.gcBgMarkWorker.ptr()
        casgstatus(gp, _Gwaiting, _Grunnable)
        if trace.enabled {
            traceGoUnpark(gp, 0)
        }
        return gp, false
    }
    
    delta := int64(-1)
    if pollUntil != 0 {
        // checkTimers ensures that polluntil > now.
        delta = pollUntil - now
    }
    
    // wasm only:
    // If a callback returned and no other goroutine is awake,
    // then wake event handler goroutine which pauses execution
    // until a callback was triggered.
    gp, otherReady := beforeIdle(delta)
    if gp != nil {
        casgstatus(gp, _Gwaiting, _Grunnable)
        if trace.enabled {
            traceGoUnpark(gp, 0)
        }
        return gp, false
    }
    if otherReady {
        goto top
    }
    
    // Before we drop our P, make a snapshot of the allp slice,
    // which can change underfoot once we no longer block
    // safe-points. We don't need to snapshot the contents because
    // everything up to cap(allp) is immutable.
    allpSnapshot := allp
    
    // return P and block
    lock(&sched.lock)
    if sched.gcwaiting != 0 || _p_.runSafePointFn != 0 {
        unlock(&sched.lock)
        goto top
    }
    if sched.runqsize != 0 {
        gp := globrunqget(_p_, 0)
        unlock(&sched.lock)
        return gp, false
    }
    if releasep() != _p_ {
        throw("findrunnable: wrong p")
    }
    pidleput(_p_)
    unlock(&sched.lock)
    
    // Delicate dance: 线程从 "自旋状态" 转换为 "非自旋状态", 可能与提交新的goroutine并发. 
    // 我们必须先丢弃 nmspinning, 然后再次检查所有 P 本地队列 (在两者之间使用 #StoreLoad 内存屏障). 
    // 如果上述两件事执行顺序相反, 在检查完所有运行队列之后但在丢弃 nmspinning 之前, 另一个线程可以提交 
    // goroutine. 结果, 没有人会释放线程来运行 goroutine.
    //
    // 如果后面查到了新 goroutine, 则需要恢复 m.spinning 作为自旋的信号, 以释放新的工作线程 (因为可能有
    // 多个饥饿的goroutine). 但是, 如果在发现新goroutine之后也没有查找到空闲的P, 则可以仅休眠当前线程: 
    // 系统已满载, 因此不需要旋转线程.
    // 另请参见文件顶部的 "Worker thread parking/unparking" 注释.
    
    wasSpinning := _g_.m.spinning
    if _g_.m.spinning {
        // 先将 m 更改为 "非自旋状态"
        _g_.m.spinning = false
        if int32(atomic.Xadd(&sched.nmspinning, -1)) < 0 {
            throw("findrunnable: negative nmspinning")
        }
    }
    
    // 接着检查 allp 当中是否有先新的 goroutine 加入
    for _, _p_ := range allpSnapshot {
        // _p_ 的本地队列有 g 加入
        if !runqempty(_p_) {
            lock(&sched.lock)
            _p_ = pidleget() // 从全局空闲 sched.pidle 当中获取一个 _p_ 
            unlock(&sched.lock)
            if _p_ != nil {
                acquirep(_p_) // 将 _p_ 当前运行的 m 绑定. acquirep() -> wirep()
                if wasSpinning {
                    _g_.m.spinning = true // 当前的 m 再次进入自旋当中
                    atomic.Xadd(&sched.nmspinning, 1)
                }
                goto top
            }
            
            break
        }
    }
    
    ... // 再次进行 GC 和 netpoll 检查 
    
    
    // 休眠
    stopm()
    goto top
}
```


偷取是由 runqsteal() 函数完成, 从 p2 当中偷取 G 放入 `_p_` 当中. 批量偷取的细节函数由 runqgrab() 完成. 偷取完成
之后, 对 `_p_` 的runqtail进行修正.

```cgo
// 从 _p_ 当中偷取 g 放入到 p2
func runqsteal(_p_, p2 *p, stealRunNextG bool) *g {
    t := _p_.runqtail
    // 从 p2 当中偷取 g 存放到_p_ 当中, t 是 _p_ 当中存储的开始位置
    // 返回偷取的数量
    n := runqgrab(p2, &_p_.runq, t, stealRunNextG)
    if n == 0 {
        return nil
    }
    
    // 至少偷取了一个, 将偷取到的最后一个位置的 gp 返回
    n--
    gp := _p_.runq[(t+n)%uint32(len(_p_.runq))].ptr()
    if n == 0 {
        return gp
    }
    // 调整 _p_ 的 runqtail 的值.
    h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with consumers
    if t-h+n >= uint32(len(_p_.runq)) {
        throw("runqsteal: runq overflow")
    }
    atomic.StoreRel(&_p_.runqtail, t+n) // store-release, makes the item available for consumption
    return gp
}
```

runqgrab() 完成偷取工作:

- 计算需要批量偷取的 G 的数量

- 根据计算的数量并且结合 stealRunNextG 进行偷取操作(进行 G 拷贝), 最后修正被偷取的 p 的 runqhead 的值

```cgo
// batchHead 是开始的位置, stealRunNextG 是否尝试偷取 runnext 
func runqgrab(_p_ *p, batch *[256]guintptr, batchHead uint32, stealRunNextG bool) uint32 {
    for {
        h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with other consumers
        t := atomic.LoadAcq(&_p_.runqtail) // load-acquire, synchronize with the producer
        
        n := t - h // 计算队列中有多少个 goroutine
        n = n - n/2 // 取队列中 goroutine 个数的一半
        if n == 0 {
            if stealRunNextG {
                // 尝试从 _p_.runnext 当中 steal 
                if next := _p_.runnext; next != 0 {
                    // 当前的 _p_ 处于运行中的
                    if _p_.status == _Prunning {
                        // 休眠以确保 _p_ 不会 run 我们将要窃取的 g.
                        // 这里的重要用例是当 _p_ 中的一个 gp 正在运行在 ready() 中(g0栈, 将 gp 变为 _Grunnable), 
                        // 其他一个 g 几乎同时被阻塞. 不要在此期间中偷取 gp 当中 runnext, 而是退后给 _p_ 一个调
                        // 度 runnext 的机会. 这将避免在不同 Ps 之间传递 gs.
                        // 同步 chan send/recv 在写入时需要约 50ns, 因此 3us 会产生约 50 次写入.
                        if GOOS != "windows" {
                            usleep(3)
                        } else {
                            // 在 windows 系统定时器粒度是 1-15ms
                            osyield() // windows 系统
                        }
                    }
                    
                    if !_p_.runnext.cas(next, 0) {
                        continue
                    }
                    
                    // 成功偷取 runnext
                    batch[batchHead%uint32(len(batch))] = next
                    return 1
                }
            }
            return 0
        }
        
        // 细节: 按理说队列中的goroutine个数最多就是 len(_p_.runq)
        // 所以n的最大值也就是len(_p_.runq)/2, 这里的判断是为啥?
        // 读取 runqhead 和 runqtail 是两个操作而非一个原子操作, 当读取 runhead 之后未读取 runqtail
        // 之前, 如果其他线程快速的增加这两个值(其他偷取者偷取g会增加 runqhead, 队列所有者添加 g 会增加 runqtail), 
        // 则导致读取出来的 runqtail 已经远远大于之前读取到的放在局部变量的 h 里面的 runqhead 了, 也就是说 h 和 t 
        // 已经不一致了.
        if n > uint32(len(_p_.runq)/2) { // read inconsistent h and t
            continue
        }
        
        // 获取 n 个 goroutine, 将其存放在 batch 当中
        for i := uint32(0); i < n; i++ {
            g := _p_.runq[(h+i)%uint32(len(_p_.runq))]
            batch[(batchHead+i)%uint32(len(batch))] = g
        }
        // 修改 _p_ 本地队列的 runqhead
        if atomic.CasRel(&_p_.runqhead, h, h+n) { // cas-release, commits consume
            return n
        }
    }
}
```

工作线程在放弃可运行的 goroutine 而进入睡眠之前, 会反复尝试从各个运行队列寻找需要运行的 goroutine, 可谓是尽心尽力.
需要注意:

1. 工作线程M的自旋状态(spinning). 工作线程在从其他工作线程的本地运行队列中盗取goroutine时的状态称为自旋状态. 当前 M
在去其他 P 的运行队列盗取 goroutine 之前先把 spinning 标志设为 true, 同时增加处于自旋状态 M 的数量, 而盗取结束之后
则把 spinning 标志还原为 false, 同时减少处于自旋状态的 M 的数量. 当有空闲 P 又有 goroutine 需要运行时, 这个处于自
旋状态 M 的数量决定了是否需要唤醒或者创建新的工作线程.

2.盗取算法. 盗取过程用了两个嵌套 for 循环. 内层循环实现盗取逻辑, 从代码可以看到盗取的实质就是遍历 allp 中所有的 p, 查
看其运行队列是否有goroutine, 如果有, 则取其一半到当前工作线程的运行队列, 然后从findrunnable 返回. 如果没有, 则继续
遍历下一个p. **但是为了保证公平性, 遍历 allp 时并不是从固定的 `allp[0]` 开始, 而是从随机位置上的 p 开始, 而且遍历的
顺序也是随机化了, 并不是现在访问了第i个p下一次就访问第i+1个p, 而是使用了一种伪随机方式遍历allp中的每个p, 防止每次遍历时
使用同样的顺序访问allp中的元素**

3.当真的没有可以偷取的 G 时候, 则休眠当前的 M, 并在 M 唤醒之后再次重新去偷取 G, 直到偷取成功之后返回.


当前工作线程休眠:

```cgo
func stopm() {
    _g_ := getg() // 当前是 g0
    
    // 当前运行的 M 的状态判断(未锁定, 与p绑定, 非自旋)
    if _g_.m.locks != 0 {
        throw("stopm holding locks")
    }
    if _g_.m.p != 0 {
        throw("stopm holding p")
    }
    if _g_.m.spinning {
        throw("stopm spinning")
    }
    
    lock(&sched.lock)
    mput(_g_.m) // 将 M 放入全局 sched.midle 当中
    unlock(&sched.lock)
    notesleep(&_g_.m.park)  // 当前工作线程休眠在 park 成员上, 直到唤醒才返回.
    noteclear(&_g_.m.park)  // 唤醒之后的清理工作
    acquirep(_g_.m.nextp.ptr()) // 唤醒之后, 将 m 与 m.nextp 绑定, 并开始运行
    _g_.m.nextp = 0
}
```

stopm的核心是调用mput把m结构体对象放入sched的midle空闲队列, 然后通过notesleep(&m.park)函数让自己进入睡眠状态.

**note是go runtime实现的一次性睡眠和唤醒机制, 一个线程可以通过调用 `notesleep(*note)` 进入睡眠状态, 而另外一个线
程则可以通过 `notewakeup(*note)` 把其唤醒**.

note 的底层实现机制与操作系统相关, 不同操作系统不同机制. Linux下使用 futext 系统调用, Mac 下则使用的是 pthread_cond_t
条件变量. note 对这些底层机制做了抽象个封装.

```cgo
func notesleep(n *note) {
    gp := getg() // 当前的 g0 
    if gp != gp.m.g0 {
        throw("notesleep not on g0")
    }
    ns := int64(-1) // 休眠的时间, ns
    if *cgo_yield != nil {
        // Sleep for an arbitrary-but-moderate interval to poll libc interceptors.
        ns = 10e6 // cgo 当中的单位是 ms
    }
    
    // 为了防止意外唤醒, 使用了 for 循环. 也就是说, 当处于休眠状态 n.key 的值一直是 0
    for atomic.Load(key32(&n.key)) == 0 {
        gp.m.blocked = true
        futexsleep(key32(&n.key), 0, ns) // 系统调用
        if *cgo_yield != nil {
            asmcgocall(*cgo_yield, nil)
        }
        gp.m.blocked = false
    }
}
```

Linux下线程休眠系统调用: 当 addr 的值为 val 时, 线程休眠

```cgo
func futexsleep(addr *uint32, val uint32, ns int64) {
    // ns < 0, 一直进行休眠. 
    // op 的值为 _FUTEX_WAIT_PRIVATE
    if ns < 0 {
        futex(unsafe.Pointer(addr), _FUTEX_WAIT_PRIVATE, val, nil, nil, 0)
        return
    }
    
    // ns >= 0, 休眠时间为 ns
    var ts timespec
    ts.setNsec(ns)
    futex(unsafe.Pointer(addr), _FUTEX_WAIT_PRIVATE, val, unsafe.Pointer(&ts), nil, 0)
}
```

futex 是使用汇编实现的, 这里不再展开了.

说了线程休眠, 也顺带说下一线程唤醒, 线程唤醒原理很简单, 就是将 note.key 的值设为 1, 并调用 futexwakeup 函数.

```cgo
func notewakeup(n *note) {
    old := atomic.Xchg(key32(&n.key), 1)
    if old != 0 {
        print("notewakeup - double wakeup (", old, ")\n")
        throw("notewakeup - double wakeup")
    }
    futexwakeup(key32(&n.key), 1)
}


func futexwakeup(addr *uint32, cnt uint32) {
    // 系统调用, op 的值为 _FUTEX_WAKE_PRIVATE
    ret := futex(unsafe.Pointer(addr), _FUTEX_WAKE_PRIVATE, cnt, nil, nil, 0)
    if ret >= 0 {
        return
    }
    
    // 下面是系统调用异常, 则程序退出
    // I don't know that futex wakeup can return
    // EAGAIN or EINTR, but if it does, it would be
    // safe to loop and call futex again.
    systemstack(func() {
        print("futexwakeup addr=", addr, " returned ", ret, "\n")
    })
    
    *(*int32)(unsafe.Pointer(uintptr(0x1006))) = 0x1006 // 访问非法地址, 让程序退出
}
```

### 调度循环中如何让出 CPU

#### 正常完成

正常执行过程:

```
schedule()->execute()->gogo()->xxx()->goexit()->goexit1()->mcall()->goexit0()->schedule()
```

其中 `schedule()->execute()->gogo()` 是在 g0 上执行的. `xxx()->goexit()->goexit1()->mcall()` 是在 curg 
上执行的. `goexit0()->schedule()` 又是在 g0 上执行的.

#### 被动调度

在实际场景中还有一些没有执行完成的 G, 而又需要临时停止执行. 比如, time.Sleep, IO阻塞等, 就要挂起该 G, 把CPU让出来
给其他 G 使用. 在 runtime 的 gopark 方法:

```cgo
// runtime/proc.go 

func gopark(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, traceEv byte, traceskip int) {
    if reason != waitReasonSleep {
        checkTimeouts() // 空函数
    }
    mp := acquirem() // 获取当前的 m, 并加锁.
    gp := mp.curg // 当前执行的 gp
    status := readgstatus(gp) // 获取 gp 的状态
    if status != _Grunning && status != _Gscanrunning {
        throw("gopark: bad g status")
    }
    
    // 设置 mp 和 gp 在挂起之前的一些参数
    mp.waitlock = lock
    mp.waitunlockf = unlockf
    gp.waitreason = reason
    mp.waittraceev = traceEv
    mp.waittraceskip = traceskip
    releasem(mp) // 释放锁
    
    // mcall, 从 curg 切换到 g0 上执行 park_m 函数
    // park_m 在切换到的 g0 下先将 gp 切换为_Gwaiting 状态挂起该G
    // 调用回调函数waitunlockf()由外层决定是否等待解锁, true表示等待解锁不在执行G, false则不等待解锁可以继续执行.
    mcall(park_m)
}

// 运行在 g0 上
func park_m(gp *g) {
    _g_ := getg() // 当前是 g0
    
    if trace.enabled {
        traceGoPark(_g_.m.waittraceev, _g_.m.waittraceskip)
    }
    
    // gp 状态修改: _Grunning -> _Gwaiting
    casgstatus(gp, _Grunning, _Gwaiting)
    dropg() // 将 curg 与 m 进行解绑
    
    // 执行 waitunlockf 函数, 根据返回结果做进一步处理
    if fn := _g_.m.waitunlockf; fn != nil {
        ok := fn(gp, _g_.m.waitlock)
        _g_.m.waitunlockf = nil
        _g_.m.waitlock = nil
        
        // 返回结果是 false, 将 gp 状态重新切换为 _Grunnable, 并重新执行 gp 
        if !ok {
            if trace.enabled {
                traceGoUnpark(gp, 2)
            }
            casgstatus(gp, _Gwaiting, _Grunnable)
            execute(gp, true) // Schedule it back, never returns.
        }
    }
    
    // 返回结果是 true, 需要等待解锁, 进入下一次调度
    schedule()
}
``` 

park_m 函数主要做的事情:

- 将 curg 状态设置为 _Gwaiting (等待当中), 调用 dropg 解除 curg 与 m 的关联关系

- 调用 waitunlockf 函数, 根据函数返回结果决定是再次执行(false), 还是调度执行其他的 g(true)

关于 mcall 函数, 在 `runtime.md` 当中有详细的讲解.


g 有挂起(_Gwaiting), 肯定就存在唤醒 (_Grunnable), 函数是 runtime.goready()

```cgo
// 唤醒 gp
func goready(gp *g, traceskip int) {
    systemstack(func() {
        ready(gp, traceskip, true)
    })
}

// 这里的 ready 是运行在 g0 栈上的
func ready(gp *g, traceskip int, next bool) {
    if trace.enabled {
        traceGoUnpark(gp, traceskip)
    }
    
    status := readgstatus(gp) // 读取 gp 的状态
    
    // Mark runnable.
    _g_ := getg() // 当前的 g0
    mp := acquirem() // 当中的工作线程 M, 增加 locks, 阻止对 _g_ 的抢占调度
    if status&^_Gscan != _Gwaiting { // 
        dumpgstatus(gp)
        throw("bad g->status in ready")
    }
    
    // gp 的状态切换回 _Grunnable
    casgstatus(gp, _Gwaiting, _Grunnable)
    runqput(_g_.m.p.ptr(), gp, next) // 将 gp 放入到当前 _p_ 的本地运行队列, 优先调度
    wakep() // 可能启动新的 M
    releasem(mp) // 减少locks, 并进行可能的 _g_ 抢占调度标记
}
```

ready() 函数的工作是将 gp 的状态恢复为 _Grunnable, 然后将 gp 放入到本地运行队列当中等待调度. 同时可能会启动新的 M
来调度执行 G

关于 M 的的启动和休眠在 runtime.md 当中有详细的讲解.

#### 主动调度

goroutine 的主动调度是指当前正在运行的 goroutine 通过调用 runtime.Gosched() 函数暂时放弃运行而发生的调度.

主动调度完全是用户代码自己控制的, 根据代码可以预计什么地方一定会发生调度. 

```cgo
func Gosched() {
    checkTimeouts() // amd64 linux 下是空函数
    
    // 切换到 g0 上, 执行 gosched_m 函数
    // 在 mcall 函数当中会保存好当前 gp 的 sp, pc, g, bp 信息, 以便后续的恢复执行.
    mcall(gosched_m)
}
```

```cgo
func gosched_m(gp *g) {
    // 追踪信息
    if trace.enabled {
        traceGoSched()
    }
    
    goschedImpl(gp) // 这里的 gp 是切换到 g0 之前的运行的 goroutine
}

func goschedImpl(gp *g) {
    // 读取 gp 的状态
    status := readgstatus(gp) 
    // gp 当前的状态必须是 _Grunning, 因为正在运行主动调度的嘛
    if status&^_Gscan != _Grunning {
        dumpgstatus(gp)
        throw("bad g status")
    }
    
    // 将 gp 的状态切换为 _Grunnable
    casgstatus(gp, _Grunning, _Grunnable)
    dropg() // 解除 gp 与 m 的绑定关系
    lock(&sched.lock)
    globrunqput(gp) // 将 gp 放入全局队列
    unlock(&sched.lock)
    
    schedule() // 执行新一轮调度
}
```

#### 抢占调度

需要关注的内容:

- 什么情况下会发生抢占调度

- 因运行时间过长而发生的抢占调度有什么特点

sysmon 系统监控线程会定期(10ms) 通过 retake 函数对 goroutine 发起抢占. 先从 sysmon 函数讲起.

sysmon 的启动是在 runtime.main 函数当中启动的. 启动代码片段如下:

```cgo
systemstack(func() {
    newm(sysmon, nil, -1)
})
```

newm() 的第一个参数 fn 是创建 m 的 mstartfn 的值, 该函数是在 m 启动后(调用 mstart() 函数), 未调度之前需要执行的函
数. sysmon() 函数本身是一个死循环, 因此创建的 m 不会和 p 进行绑定.

sysmon() 函数就做一件事, 没个 10 ms 发起抢占调度.

```cgo
// 始终在不带 p 的情况下运行, 因此不存在写屏障.
func sysmon() {
    lock(&sched.lock)
    sched.nmsys++
    checkdead()
    unlock(&sched.lock)
    
    lasttrace := int64(0)
    idle := 0 // 循环次数
    delay := uint32(0)
    for {
        // delay参数用于控制for循环的间隔, 不至于无限死循环.
        // 控制逻辑是前50次每次sleep 20us, 超过50次则每次翻2倍, 直到最大10ms.
        if idle == 0 { 
            delay = 20
        } else if idle > 50 { 
            delay *= 2
        }
        if delay > 10*1000 { 
            delay = 10 * 1000
        }
        usleep(delay) // 休眠 delay 时间
        now := nanotime() // 当前时间
        next, _ := timeSleepUntil() // 返回下一次需要休眠的时间
        // double check
        if debug.schedtrace <= 0 && (sched.gcwaiting != 0 || atomic.Load(&sched.npidle) == uint32(gomaxprocs)) {
            lock(&sched.lock)
            // sched.gcwaiting gc 等待的时间
            // sched.npidle 当前空闲的 p
            if atomic.Load(&sched.gcwaiting) != 0 || atomic.Load(&sched.npidle) == uint32(gomaxprocs) {
                // 下一次休眠时间未到, 需要进行休眠
                if next > now { 
                    atomic.Store(&sched.sysmonwait, 1)
                    unlock(&sched.lock)
                    // 使唤醒周期足够小以使采样正确
                    // forcegcperiod 是 120 s
                    sleep := forcegcperiod / 2 
                    if next-now < sleep {
                        sleep = next - now
                    }
                    
                    // osRelaxMinNS=0
                    shouldRelax := sleep >= osRelaxMinNS
                    if shouldRelax {
                        osRelax(true) // amd64 linux 是空函數
                    }
                    
                    // 休眠 sysmon 线程
                    notetsleep(&sched.sysmonnote, sleep)
                    if shouldRelax {
                        osRelax(false)
                    }
                    
                    // 唤醒 sysmon 后需要重新计算 now 和 next
                    now = nanotime()
                    next, _ = timeSleepUntil()
                    lock(&sched.lock)
                    atomic.Store(&sched.sysmonwait, 0)
                    noteclear(&sched.sysmonnote) // 唤醒后清理工作
                }
                idle = 0
                delay = 20
            }
            unlock(&sched.lock)
        }
        
        // 重新校对 now 和 next
        lock(&sched.sysmonlock)
        {
            // If we spent a long time blocked on sysmonlock
            // then we want to update now and next since it's
            // likely stale.
            now1 := nanotime()
            if now1-now > 50*1000 /* 50µs */ {
                next, _ = timeSleepUntil()
            }
            now = now1
        }
    
        // trigger libc interceptors if needed
        if *cgo_yield != nil {
            asmcgocall(*cgo_yield, nil)
        }
        // 如果超过 10 毫秒没有查找 epoll
        lastpoll := int64(atomic.Load64(&sched.lastpoll))
        if netpollinited() && lastpoll != 0 && lastpoll+10*1000*1000 < now {
            atomic.Cas64(&sched.lastpoll, uint64(lastpoll), uint64(now))
            list := netpoll(0) // non-blocking - returns list of goroutines
            if !list.empty() {
                // 需要在 injectglist 之前减少 "空闲锁定的m" 的数量(假设还有一个正在运行). 
                // 否则, 可能导致以下情况: injectglist 唤醒了所有的 p. 但在启动 m 运行 p 之前, 另一个 m 从
                // syscall返回, 完成它运行g, 观察到 "没有工作要执行" 并且 "也没有正在运行的m", 就报告了死锁.
                incidlelocked(-1)
                
                // 把 epoll ready 的 g 放入到全局队列. 这里不会放入到本地队列的, 因为 m 没有绑定 p
                injectglist(&list)
                incidlelocked(1)
            }
        }
        
        // next < now, 说明定时器已经执行了
        if next < now {
            // 可能是因为有一个不可抢占的 P.
            // 尝试启动一个 M 来运行它们.
            startm(nil, false)
        }
        if atomic.Load(&scavenge.sysmonWake) != 0 {
            // Kick the scavenger awake if someone requested it.
            wakeScavenger()
        }
        // 重新获取在 syscall 中阻塞的 P 并抢占长时间运行的 G
        if retake(now) != 0 {
            idle = 0
        } else {
            idle++
        }
        
        // GC 相关内容
        // check if we need to force a GC
        if t := (gcTrigger{kind: gcTriggerTime, now: now}); t.test() && atomic.Load(&forcegc.idle) != 0 {
            lock(&forcegc.lock)
            forcegc.idle = 0
            var list gList
            list.push(forcegc.g)
            injectglist(&list)
            unlock(&forcegc.lock)
        }
        if debug.schedtrace > 0 && lasttrace+int64(debug.schedtrace)*1000000 <= now {
            lasttrace = now
            schedtrace(debug.scheddetail > 0)
        }
        unlock(&sched.sysmonlock)
    }
}
```

retake() 函数是 sysmon 的一个核心函数, 目的是获取在 syscall 中阻塞的 P 并抢占长时间运行的 G

```cgo
func retake(now int64) uint32 {
    n := 0
    // Prevent allp slice changes. This lock will be completely
    // uncontended unless we're already stopping the world.
    lock(&allpLock)
    
    // 我们不能在 allp 上使用 range 循环, 因为我们可能会暂时删除 allpLock. 
    // 因此, 我们需要在循环中每次都重新获取.
    for i := 0; i < len(allp); i++ {
        _p_ := allp[i]
        if _p_ == nil {
            // This can happen if procresize has grown
            // allp but not yet created new Ps.
            continue
        }
        
        // _p_.sysmontick用于 sysmon 线程记录被监控 p 的系统调用的次数和调度的次数
        pd := &_p_.sysmontick
        s := _p_.status
        sysretake := false
        // 当前 p 处于系统调用或正在运行当中
        if s == _Prunning || s == _Psyscall {
            // 如果运行时间过长, 则进行抢占
            // _p_.schedtick: 每发生一次调度, 调度器++该值
            t := int64(_p_.schedtick)
            if int64(pd.schedtick) != t { // 监控到一次新的调度, 重置 schedtick 和 schedwhen
                pd.schedtick = uint32(t) 
                pd.schedwhen = now
            } else if pd.schedwhen+forcePreemptNS <= now { // 运行时间超过 10 ms 运行时间
                // 抢占标记, 设置运行的 gp.preempt=true, gp.stackguard0=stackPreempt
                preemptone(_p_) 
                // 对于系统调用, preemptone() 不会工作, 因为此时的 m 和 p 已经解绑.
                sysretake = true
            }
        }
        
        // 当前 p 处于系统调用当中
        if s == _Psyscall {
            // 在运行时间未超过 10 ms状况下, 监控到一次新的调度(至少是20us), 修正相关的值
            t := int64(_p_.syscalltick)
            if !sysretake && int64(pd.syscalltick) != t {
                // 监控线程监控到一次新的调度, 需要重置跟 sysmon 相关的 schedtick 和 schedwhen 变量
                pd.syscalltick = uint32(t)
                pd.syscallwhen = now
                continue
            }
            
            // 在满足下面所有条件下, 不会发生抢占调度:
            // - 当前的 _p_ 本地队列为空, 说明没有可以执行的 g;
            // - sched.nmspinning 或 sched.npidle 不为0, 说明没有空闲的 p
            // - 运行时间未超过 10ms;
            if runqempty(_p_) && atomic.Load(&sched.nmspinning)+atomic.Load(&sched.npidle) > 0 && 
                pd.syscallwhen+10*1000*1000 > now {
                continue
            }
            // Drop allpLock so we can take sched.lock.
            unlock(&allpLock)
            
            // 减少 sched.nmidlelocked 的数量
            incidlelocked(-1)
            if atomic.Cas(&_p_.status, s, _Pidle) { // 将当前的 _p_ 状态切换成 _Pidle
                if trace.enabled {
                    traceGoSysBlock(_p_)
                    traceProcStop(_p_)
                }
                n++
                _p_.syscalltick++ // 系统调度次数增加
                handoffp(_p_) // 启动新的 m 
            }
            incidlelocked(1)
            lock(&allpLock)
        }
    }
    unlock(&allpLock)
    return uint32(n)
}
```

sysmon 监控线程如果监控到某个 goroutine 连续运行超过 10ms, 则会调用 preemptone() 函数向该 goroutine 发起抢占请
求.

```cgo
func preemptone(_p_ *p) bool {
    mp := _p_.m.ptr()
    if mp == nil || mp == getg().m {
        return false
    }
    gp := mp.curg
    if gp == nil || gp == mp.g0 {
        return false
    }
    
    // gp 是被抢占的 goroutine, 因此是 curg
    // 设置抢占标志
    gp.preempt = true 
    
    
    // go 协程中的每个调用都通过将当前栈指针与 gp->stackguard0 进行比较来检查栈溢出.
    // 将 gp->stackguard0 设置为 StackPreempt 会将抢占合并到正常的栈溢出检查中.
    // stackPreempt 是一个常亮 0xfffffffffffffade, 非常大的一个数
    gp.stackguard0 = stackPreempt
    
    // Request an async preemption of this P.
    if preemptMSupported && debug.asyncpreemptoff == 0 {
        _p_.preempt = true
        preemptM(mp)
    }
    
    return true
}
```

preemptone 函数只是简单的设置了抢占 goroutine 对于 g 结构体中的 preempt 成员为 true 和 stackguard0 成员的值为
stackPreempt就返回了, 并未真正强制被抢占的 goroutine 暂停下来.

既然设置了抢占标志, 那么就一定需要对这些标志进行处理.

处理抢占请求的函数是 newstack(), 该函数如果发现自己被抢占, 则会暂停当前 goroutine 的执行. 而调用 newstack函数的调用
链是: `morestack_noctxt()->morestack()->newstack()`, 而 `runtime.morestack_noctxt()` 函数会被插入到每个 go
函数的结尾处.

```cgo
0x0000 00000 (call.go:17)       MOVQ    (TLS), CX # CX=g 
0x0009 00009 (call.go:17)       CMPQ    SP, 16(CX) # g.stackguard0 与当前 SP 进行比较
0x000d 00013 (call.go:17)       JLS     163 # SP < g.stackguard0, 则跳转

... 

0x00a3 00163 (call.go:17)       CALL    runtime.morestack_noctxt(SB) # stack 扩容
0x00a8 00168 (call.go:17)       JMP     0 # 跳转到开始
```

上面代码的核心就是当 sp < stackguard0, 则说明当前 g 的栈快用完了. 有溢出风险, 需要栈扩容. 

如果当前的 g 设置了抢占标记, 那么 sp 的值就会远远小于 stackguard0, 这样就会执行 runtime.morestack_noctxt() 函数.

```cgo
TEXT runtime·morestack_noctxt(SB),NOSPLIT,$0
    MOVL	$0, DX
    JMP	runtime·morestack(SB)
```

morestack_noctxt 函数使用 JMP 指令直接跳转到 morestack 继续执行. 注意这里没有使用 CALL 指令调用 morestack, 因此
rsp 栈顶寄存器没有发生变化(栈顶值为调用 morestack_noctxt 函数的下一条指令的地址)

```cgo
TEXT runtime·morestack(SB),NOSPLIT,$0-0
    # 获取 m 
    get_tls(CX)
    MOVQ	g(CX), BX    # BX=g 
    MOVQ	g_m(BX), BX  # BX=g.m 
    
    # m->g0 与 g 比较
    MOVQ	m_g0(BX), SI # SI=m.g0 
    CMPQ	g(CX), SI    # m.g0 == g 
    JNE	3(PC) # 不相等
    CALL	runtime·badmorestackg0(SB)
    CALL	runtime·abort(SB)
    
    # m->gsignal 与 g 比较 
    MOVQ	m_gsignal(BX), SI # SI = m.signal 
    CMPQ	g(CX), SI # m.signal == g 
    JNE	3(PC) # 不相等
    CALL	runtime·badmorestackgsignal(SB)
    CALL	runtime·abort(SB)
    
    // f 是通过 call 调用 morestack_noctxt 的那个函数
    // f's caller 是通过 call 调用 f 的那个函数
    // 保存 f's caller 的信息到 m->morebuf
    NOP	SP	// tell vet SP changed - stop checking offsets
    // 这里的 SP 和 PC 的获取是 hard code 的.
    // 因为 morestack_noctxt 函数是汇编编译插入到代码当中, 并且是在函数调用的开始的位置
    // 而且插入的指令并没有占用栈空间. 那么 8(SP) 就是 f's caller 的 PC, 即调用 f 函数完成后的返回地址
    // 16(SP) 就是 f's caller 的 SP, 即调用 f 函数的栈顶指针.
    // 需要记录 f's caller 的 SP, 是因为当前 goroutine 要进行栈扩容, 那么就会发生栈内容的拷贝, 拷贝的结束位置就是
    // f's caller 的 SP. 
    // 需要记录 f's caller 的 PC, 是因为栈扩容完成之后, 需要重新将 PC 压人栈, 调用 f 函数.
    MOVQ	8(SP), AX	// f's caller's PC 
    MOVQ	AX, (m_morebuf+gobuf_pc)(BX)
    LEAQ	16(SP), AX	// f's caller's SP
    MOVQ	AX, (m_morebuf+gobuf_sp)(BX)
    get_tls(CX)
    MOVQ	g(CX), SI
    MOVQ	SI, (m_morebuf+gobuf_g)(BX)
    
    // 保存 f 的信息到 g->sched
    MOVQ	0(SP), AX // f's PC, 栈顶是函数的返回地址, 即当前函数返回后需要执行的指令
    MOVQ	AX, (g_sched+gobuf_pc)(SI)
    MOVQ	SI, (g_sched+gobuf_g)(SI)
    LEAQ	8(SP), AX // f's SP, f函数的栈顶指针
    MOVQ	AX, (g_sched+gobuf_sp)(SI)
    MOVQ	BP, (g_sched+gobuf_bp)(SI)
    MOVQ	DX, (g_sched+gobuf_ctxt)(SI) # 0 
    
    // 调用 newstack 函数
    MOVQ	m_g0(BX), BX // BX=m.g0
    MOVQ	BX, g(CX)    // g=BX, 切换到 g0 上
    MOVQ	(g_sched+gobuf_sp)(BX), SP // 恢复 g0.sched.sp 
    CALL	runtime·newstack(SB) // 函数不会返回
    CALL	runtime·abort(SB)	// crash if newstack returns
    RET
```

```cgo
//go:nowritebarrierrec
func newstack() {
    thisg := getg() // g0
    
    ... // 省略一些检查性的代码
    
    gp := thisg.m.curg
    
    ... // 省略一些检查性的代码
    
    morebuf := thisg.m.morebuf
    thisg.m.morebuf.pc = 0
    thisg.m.morebuf.lr = 0
    thisg.m.morebuf.sp = 0
    thisg.m.morebuf.g = 0
    
    // 判断当前是否处于抢占调度当中
    preempt := atomic.Loaduintptr(&gp.stackguard0) == stackPreempt
    
    // Be conservative about where we preempt.
    // We are interested in preempting user Go code, not runtime code.
    // If we're holding locks, mallocing, or preemption is disabled, don't
    // preempt.
    // This check is very early in newstack so that even the status change
    // from Grunning to Gwaiting and back doesn't happen in this case.
    // That status change by itself can be viewed as a small preemption,
    // because the GC might change Gwaiting to Gscanwaiting, and then
    // this goroutine has to wait for the GC to finish before continuing.
    // If the GC is in some way dependent on this goroutine (for example,
    // it needs a lock held by the goroutine), that small preemption turns
    // into a real deadlock.
    if preempt {
        // 检查被抢占 goroutine 的 M 的状态
        if !canPreemptM(thisg.m) {
            // Let the goroutine keep running for now.
            // gp->preempt is set, so it will be preempted next time.
            // 还原 stackguard0 为正常值, 表示已经处理过了抢占请求了
            gp.stackguard0 = gp.stack.lo + _StackGuard
            // 不抢占, gogo 继续执行 m.curg, 不用调用 schedule() 函数挑选 g 了
            gogo(&gp.sched) // never return
        }
    }
    
    ... // 省略一些检查性的代码 
    
    if preempt {
        if gp == thisg.m.g0 {
            throw("runtime: preempt g0")
        }
        if thisg.m.p == 0 && thisg.m.locks == 0 {
            throw("runtime: g is running but p is not")
        }
        
        // 收缩栈
        if gp.preemptShrink {
            gp.preemptShrink = false
            shrinkstack(gp) // 栈收缩
        }
        
        // 停止抢占, 但是也会进入下一轮调度
        if gp.preemptStop {
            preemptPark(gp) // never returns
        }
    
        // 其行为和 runtime.Gosched() 是一样的. gopreempt_m 和 gosched_m 函数的逻辑是一样的
        // 响应抢占请求, 最终调用到 goschedImpl() 函数, 
        // 设置 gp 的状态为 _Grunnable, 同时与 gp 与 m 解绑
        // 进入到下一轮调度. 
        gopreempt_m(gp) // never return
    }
    
    // 下面的内容是扩容栈
    // 栈扩容到原来的 2 倍
    oldsize := gp.stack.hi - gp.stack.lo
    newsize := oldsize * 2
    
    // 确保我们的增长至少与适应新的frame所需的一样多.
    // (这只是一个优化 - morestack 的调用者将在返回时重新检查边界)
    if f := findfunc(gp.sched.pc); f.valid() {
        max := uintptr(funcMaxSPDelta(f))
        for newsize-oldsize < max+_StackGuard {
            newsize *= 2
        }
    }
    
    // 栈溢出, 超过 1 个G
    if newsize > maxstacksize {
        print("runtime: goroutine stack exceeds ", maxstacksize, "-byte limit\n")
        print("runtime: sp=", hex(sp), " stack=[", hex(gp.stack.lo), ", ", hex(gp.stack.hi), "]\n")
        throw("stack overflow")
    }
    
    // 修改当前的 gp 的状态到 _Gcopystack
    casgstatus(gp, _Grunning, _Gcopystack)
    
    // 栈内容拷贝. 创建新的栈, 相关内容拷贝, 释放掉旧的栈 
    // 在由于 gp 的状态是 _Gcopystack, 因此 GC 不会扫描当前的栈
    copystack(gp, newsize)
    if stackDebug >= 1 {
        print("stack grow done\n")
    }
    
    // 切换回 gp 原来的状态, 并继续执行之前的代码
    casgstatus(gp, _Gcopystack, _Grunning)
    gogo(&gp.sched)
}
```

抢占调度的过程, 在 sysmon 当中监控发现某个 g 运行时间过长(超过10ms), 则对该 g 设置抢占标记. 该 g 在下一次函数调用的
时候, 在检查栈的时候会进行抢占调度. 

#### 系统调用

前面分析了当运行时间超过 10ms 之后会发生抢占. 接下来说下系统调用导致的抢占.

对于处于 _Psyscall 状态的 p, 只要满足下面三个条件之一, 就会发生抢占:

- p 的运行队列不为空. 这用来保证当前 p 的本地运行队列中的 goroutine 得到及时的调度, 因为该 p 对应的工作线程正处于系统
调用之中, 无法调度队列中的 goroutine, 所以需要寻找另外一个工作线程来接管这个 p 从而达到调度这些 goroutine 的目的.

- 没有空闲的 p(sched.nmspinning+sched.npidle==0). 表示其他所有的 p 都已经与工作线程绑定且正在执行 go 代码, 这说
明系统比较繁忙, 因此需要抢占当前正在处于系统调用之中而实际上系统调用并不需要的这个 p 并把它分配给其他的工作线程去调度其他
的 goroutine

- 观察到 p 对应的 m 处于系统调用之中到现在已经超过 10ms. 这表示只要系统调用超时, 就对其进行抢占.

上述的条件是根据 sysmon() 函数当当中 p 为 `_Psyscall` 反推来的. 那么当满足上述的条件之后, 会调用函数 handoffp() 函
数去启动新的工作线程来接管 p 从而达到抢占的目的.

```cgo
func handoffp(_p_ *p) {
    // handoffp must start an M in any situation where
    // findrunnable would return a G to run on _p_.
    
    // 本地运行队列不为空, 需要启动 m 来接管
    if !runqempty(_p_) || sched.runqsize != 0 {
        startm(_p_, false)
        return
    }
    // 有垃圾回收工作要做, 也需要启动 m 来接管
    if gcBlackenEnabled != 0 && gcMarkWorkAvailable(_p_) {
        startm(_p_, false)
        return
    }
    
    // 所有其他的 p 都在运行 goroutine, 说明系统比较忙, 需要启动 m 
    if atomic.Load(&sched.nmspinning)+atomic.Load(&sched.npidle) == 0 && atomic.Cas(&sched.nmspinning, 0, 1) { // TODO: fast atomic
        startm(_p_, true)
        return
    }
    
    lock(&sched.lock)
    
    // 如果 gc 正在等待 Stop The World
    if sched.gcwaiting != 0 {
        _p_.status = _Pgcstop
        sched.stopwait--
        if sched.stopwait == 0 {
            notewakeup(&sched.stopnote)
        }
        unlock(&sched.lock)
        return
    }
    if _p_.runSafePointFn != 0 && atomic.Cas(&_p_.runSafePointFn, 1, 0) {
        sched.safePointFn(_p_)
        sched.safePointWait--
        if sched.safePointWait == 0 {
            notewakeup(&sched.safePointNote)
        }
    }
    
    // 全局运行队列不为空, 启动 m 来接管 p
    if sched.runqsize != 0 {
        unlock(&sched.lock)
        startm(_p_, false)
        return
    }
    
    // 所有的 p 都空闲下来了, 但是需要监控网络连接读写时间, 需要启动 m
    if sched.npidle == uint32(gomaxprocs-1) && atomic.Load64(&sched.lastpoll) != 0 {
        unlock(&sched.lock)
        startm(_p_, false)
        return
    }
    if when := nobarrierWakeTime(_p_); when != 0 {
        wakeNetPoller(when)
    }
    pidleput(_p_) // 无事可做, 将 p 放入到全局空闲队列
    unlock(&sched.lock)
}
```

handoffp 函数就是在适当的状况下启动 m 接收 p, 主要包含以下条件:

- "p 本地运行队列" 或 "全局队列" 存在待运行的 g

- 需要帮助 gc 完成标记工作

- 系统繁忙, 所有其他的 p 都在运行 g, 需要帮忙

- 所有的 p 都处于空闲状态, 但是需要监控网络连接读写事件, 则需要启动新的 m 来 poll 网络连接.

到此为止, sysmon 监控线程对处于系统调用当中的 p 抢占就已经完成.

从上面来看, **对于正在进行系统调用的goroutine的抢占实质是剥夺与其对应的工作线程所绑定的p**, 虽然说处于系统调用之中的工
作线程不需要 p, 但一旦从操作系统内核返回到用户空间之后就必须绑定一个 p 才能运行 go 代码. 那么, 工作线程从系统调用返回之
后如果发现进入系统调用之前的所使用的 p 被监控线程拿走了, 该咋办?


Go 中没有直接对系统内核函数调用, 而是封装了 syscall.Syscall 方法.

```cgo
// syscall/syscall_unix.go

func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
```


```cgo
// func Syscall(trap int64, a1, a2, a3 uintptr) (r1, r2, err uintptr);
// trap in AX, 
// args in DI SI DX R10 R8 R9, 
// return in AX DX
TEXT ·Syscall(SB),NOSPLIT,$0-56
    CALL	runtime·entersyscall(SB) # 系统调用的准备工作
    MOVQ	a1+8(FP), DI
    MOVQ	a2+16(FP), SI
    MOVQ	a3+24(FP), DX
    MOVQ	trap+0(FP), AX	# AX, 系统调用号
    SYSCALL               
     
    // 0xfffffffffffff001 是 -4095
    // 从内核返回, 判断返回值, linux 使用 -1 ~ -4095 作为错误码
    CMPQ	AX, $0xfffffffffffff001 
    JLS	ok  // AX < $0xfffffffffffff001
    
    // 系统调用异常返回
    MOVQ	$-1, r1+32(FP)
    MOVQ	$0, r2+40(FP)
    NEGQ	AX  // 错误码取反
    MOVQ	AX, err+48(FP)
    CALL	runtime·exitsyscall(SB)
    RET
    
    // 系统调用正常返回
ok:
    MOVQ	AX, r1+32(FP)
    MOVQ	DX, r2+40(FP)
    MOVQ	$0, err+48(FP)
    CALL	runtime·exitsyscall(SB)
    RET
```

Syscall 函数主要依次做了三件事:

- 调用 runtime.entersyscall 函数

- 使用 SYSCALL 指令进入系统调用

- 调用 runtime.exitsyscall 函数

根据前面的分析猜测, exitsyscall 函数将会处理当前工作线程进入系统调用之后所拥有的 p 被监控线程抢占剥夺的情况. 但是 
entersyscall 又会做啥呢?

```cgo
func entersyscall() {
	reentersyscall(getcallerpc(), getcallersp())
}

// goroutine g 即将进入系统调用.
// 记录它不再使用cpu.
// 仅从 go syscall 库和 cgocall 中调用此方法, 而不从运行时使用的 low-level 系统调用中调用此方法.
//
// entersyscall不能拆分堆栈: gosave 必须使 g->sched 引用调用者的堆栈段, 因为 entersyscall 将在之后立即返回.
//
// 在对syscall进行有效调用期间, 我们无法安全地移动堆栈, 因为我们不知道哪个 uintptr 参数是真正的指针(返回堆栈).
// 实际上, 我们需要以最短的路径去到达 entersyscall, 并执行该函数.
//
// reentersyscall 是 cgo 回调所使用的入口点, 在该处显式保存的SP和PC. 当将从调用堆栈中比父级更远的函数中调用 
// exitsyscall时, 这是必需的, 因为 g->syscallsp 必须始终指向有效的 stack frame.  entersyscall是syscall的常
// 规入口点, 它从调用方获取SP和PC.
//
// Syscall trace:
// 在系统调用开始时, 发出 traceGoSysCall 以捕获stack trace.
// 如果系统调用没有被阻塞, 我们不会发出任何其他事件.
// 如果系统调用被阻塞(即, P 被重新获取), 则 retaker 触发 traceGoSysBlock; 当syscall返回时, 触发 traceGoSysExit,
// 当 goroutine 开始运行时(可能是瞬间, 如果 exitsyscallfast 返回true), 触发 traceGoStart.
//
// 为了确保在 traceGoSysBlock 之后严格发出 traceGoSysExit, 需要存储 m 当中 syscalltick 的当前值( 
// _g_.m.syscalltick = _g_.m.p.ptr().syscalltick), 谁发出 traceGoSysBlock, 增加其对应的 p.syscalltick;
// 我们在触发 traceGoSysExit 之前等待该增量.
// 
// 注意: 即使未启用跟踪, 增量也会完成, 因为可以在syscall的中间启用跟踪. 我们不希望等待挂起.
//go:nosplit
func reentersyscall(pc, sp uintptr) {
    _g_ := getg() // 执行系统调用的 g 
    
    // 增加 m.locks 的值, 禁止抢占, 因此 g 处于 Gsyscall 状态, 但可以有不一致的 g.sched
    _g_.m.locks++
    
    // Entersyscall 不得调用任何可能 split/grow stack 的函数.
    // 通过将 stackguad0 设置为 stackPreempt 可以触发 stack 检查, 并设置 throwsplit 为 true 让
    // newstack 函数抛出异常, 从而捕获可能的调用.
    _g_.stackguard0 = stackPreempt
    _g_.throwsplit = true // 此标志会导致 newstack 抛出异常.
    
    // Leave SP around for GC and traceback.
    // 保存当前 g 的 pc 和 sp, 之前分析过此函数
    save(pc, sp) 
    _g_.syscallsp = sp
    _g_.syscallpc = pc
    
    // g 的状态切换到 _Gsyscall
    casgstatus(_g_, _Grunning, _Gsyscall) 
    if _g_.syscallsp < _g_.stack.lo || _g_.stack.hi < _g_.syscallsp {
        systemstack(func() {
            print("entersyscall inconsistent ", hex(_g_.syscallsp), " [", hex(_g_.stack.lo), ",", hex(_g_.stack.hi), "]\n")
            throw("entersyscall")
        })
    }
    
    // 追踪
    if trace.enabled {
        systemstack(traceGoSysCall)
        save(pc, sp)
    }
    
    // 需要等待监控
    if atomic.Load(&sched.sysmonwait) != 0 {
        systemstack(entersyscall_sysmon)
        save(pc, sp)
    }
    
    // p 设置了安全检查函数
    if _g_.m.p.ptr().runSafePointFn != 0 {
        // 如果在此堆栈上运行 runSafePointFn 可能将 stack 拆分
        systemstack(runSafePointFn)
        save(pc, sp)
    }
    
    _g_.m.syscalltick = _g_.m.p.ptr().syscalltick // 将 m 和 p 的 syscalltick 设置一致
    _g_.sysblocktraced = true
    
    // 解除 m, p 之间的绑定关系
    // 注意: 将 p 设置到了 m 的 oldp 上.
    pp := _g_.m.p.ptr()
    pp.m = 0
    _g_.m.oldp.set(pp) 
    _g_.m.p = 0
    
    // 设置 p 的状态
    atomic.Store(&pp.status, _Psyscall)
    
    // 需要等待 gc 
    if sched.gcwaiting != 0 {
        systemstack(entersyscall_gcwait)
        save(pc, sp)
    }
    
    // 减少 m.locks 的值, 解除禁止抢占
    _g_.m.locks--
}
```

有几个问题:

- 有 sysmon 监控线程来抢占剥夺, 为什么这里还需要解除 m 和 p 之间的绑定关系? 原因在于这里主动解除 m 和 p 的绑定关系之
后, sysmon 线程就不需要加锁或cas操作来修改 m.p 成员从而解除 m 和 p 直接的关系.

- 为什么要记录工作线程进入系统调用前的所绑定的 p? 因为记录下来可以让工作线程从系统调用返回之后快速找到一个可能可用的 p,
而不需要加锁从 sched 的 pidle 全局列表中去寻找空闲的 p


exitsyscall 函数:

```cgo
//go:nosplit
//go:nowritebarrierrec
//go:linkname exitsyscall
func exitsyscall() {
    _g_ := getg()
    
    _g_.m.locks++ 
    if getcallersp() > _g_.syscallsp {
        throw("exitsyscall: syscall frame is no longer valid")
    }
    
    _g_.waitsince = 0
    oldp := _g_.m.oldp.ptr() // 进入系统调用前所绑定的 p
    _g_.m.oldp = 0
    
    // 当前的 m 可以成功绑定到 oldp 或 sched.pilde 当中的 p
    if exitsyscallfast(oldp) {
        if trace.enabled {
            if oldp != _g_.m.p.ptr() || _g_.m.syscalltick != _g_.m.p.ptr().syscalltick {
                systemstack(traceGoStart)
            }
        }
        // 增加 syscalltick 计数. sysmon 线程依据它判断是否是同一次系统调用
        _g_.m.p.ptr().syscalltick++
        casgstatus(_g_, _Gsyscall, _Grunning) // 切换 g 的状态
    
        _g_.syscallsp = 0
        _g_.m.locks-- // 解除禁止抢占
        
        // 如果当前 g 处于抢占当中, 设置 stackguard0 的值
        if _g_.preempt {
            // 恢复抢占请求, 以防我们在 newstack 中清除它
            _g_.stackguard0 = stackPreempt
        } else {
            // 恢复正常的 stackguard0, 因为在进入系统调用前, 破坏了该值
            _g_.stackguard0 = _g_.stack.lo + _StackGuard
        }
        _g_.throwsplit = false  // 解除不允许 stack 分裂
    
        if sched.disable.user && !schedEnabled(_g_) {
            // Scheduling of this goroutine is disabled.
            Gosched()
        }
    
        return
    }
    
    // 当前没有空闲的 p 来绑定当前的 m
    _g_.sysexitticks = 0
    if trace.enabled {
        // Wait till traceGoSysBlock event is emitted.
        // This ensures consistency of the trace (the goroutine is started after it is blocked).
        for oldp != nil && oldp.syscalltick == _g_.m.syscalltick {
            osyield()
        }
        // We can't trace syscall exit right now because we don't have a P.
        // Tracing code can invoke write barriers that cannot run without a P.
        // So instead we remember the syscall exit time and emit the event
        // in execute when we have a P.
        _g_.sysexitticks = cputicks()
    }
    
    _g_.m.locks-- // 解除抢占
    
    // 切换到 g0, 调用 exitsyscall0
    // 将 gp 状态设置为 _Grunnable
    // 解除 gp 与 m 的绑定关系
    // 将 gp 放入到全局队列当中, 休眠当前的 m, 在 m 唤醒之后开启新一轮的调度
    mcall(exitsyscall0)
    
    // 调度程序返回, 所以我们现在可以运行了.
    // 删除我们在系统调用期间留给垃圾收集器的 syscallsp 信息. 必须等到现在, 因为在 gosched 返回之前, 不能确定垃圾
    // 收集器没有运行.
    _g_.syscallsp = 0
    _g_.m.p.ptr().syscalltick++
    _g_.throwsplit = false
}
```


```cgo
//go:nosplit
func exitsyscallfast(oldp *p) bool {
    _g_ := getg()
    
    // Freezetheworld sets stopwait but does not retake P's.
    if sched.stopwait == freezeStopWait {
        return false
    }
    
    // 尝试重新绑定 oldp
    if oldp != nil && oldp.status == _Psyscall && atomic.Cas(&oldp.status, _Psyscall, _Pidle) {
        wirep(oldp) // 当前 g 与 oldp 绑定
        exitsyscallfast_reacquired() // 增加 p.syscalltick 计数
        return true
    }
    
    // 尝试获取其他的空闲的 p
    if sched.pidle != 0 {
        var ok bool
        systemstack(func() {
            // 从 sched.pidle 当中成功获取的 p, 并且绑定到当前的 m
            ok = exitsyscallfast_pidle()
            if ok && trace.enabled {
                if oldp != nil {
                    // Wait till traceGoSysBlock event is emitted.
                    // This ensures consistency of the trace (the goroutine is started after it is blocked).
                    for oldp.syscalltick == _g_.m.syscalltick {
                        osyield()
                    }
                }
                traceGoSysExit(0)
            }
        })
        if ok {
            return true
        }
    }
    
    return false
}
```


当没有空闲的 p 可供当前 m 进行绑定, 则需要切换到 g0 上, 执行 exitsyscall0 函数:

```cgo
//go:nowritebarrierrec
func exitsyscall0(gp *g) {
    _g_ := getg() // 当前就是 g0
    
    casgstatus(gp, _Gsyscall, _Grunnable) // 切换 gp 的状态
    dropg() // 将 gp 与 m 解绑
    
    lock(&sched.lock)
    var _p_ *p
    
    // 当前 g0 是否可以被调度, 返回值是 true
    if schedEnabled(_g_) {
        _p_ = pidleget() // 尝试从 sched.pidle 当中获取 p
    }
    
    if _p_ == nil {
        // 仍然没有获取到 p, 只能将 gp 放入全局队列
        globrunqput(gp)
    } else if atomic.Load(&sched.sysmonwait) != 0 {
        // 当前处于 sysmon 线程等待, 则需要唤醒 sysmon
        atomic.Store(&sched.sysmonwait, 0)
        notewakeup(&sched.sysmonnote)
    }
    unlock(&sched.lock)
    if _p_ != nil {
        // 获取到了 p, 则将 p 与 m 绑定, 并且去执行当前的 gp 
        acquirep(_p_)
        execute(gp, false) // Never returns.
    }
    
    if _g_.m.lockedg != 0 {
        // 虽然没有获取的 p, 但是当前 m 锁定在 g (g0) 上.
        // 休眠当前的 m, 等到唤醒时, 则与新的 p 进行了绑定, 则可以重新执行 gp
        // 疑问: 这里 gp 不会被其他 m 调度执行吗?
        stoplockedm()
        execute(gp, false) // Never returns.
    }
    
    // 比较悲催, 只能休眠当前的 m, 等到唤醒之后, 则与新的 p 绑定, 再次进入调度
    // 注: 在这里 gp 已经放入全局队列, 可能已经被执行了.
    stopm()
    schedule() // Never returns.
}
```

#### 常用的函数

```cgo
wakep() // 在 sched.npidle >0 && sched.nmspinning == 0 的状况下启动一个自旋的 M, 也就是唤醒一个 P

wirep(p) // 将当前的 m 与 p 绑定

dropg() // 解除 curg 和 m 的绑定关系

acquirep(p) // 将 p 与 m 进行绑定, 同时设置 p 的状态为 _Prunning
releasep()  // 将 p 与 m 解绑, 同时设置 p 的状态为 _Pidle


acquirem()  // 禁止抢占调度(m.locks++)
releasem(m) // 启用抢占调度(m.locks--), 如果当前处于抢占当中(m.locks==0&&_g_.preempt ), 
            // 修改 _g_.stackguard0 的值
```
