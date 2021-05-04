# go 调度

G, M, P

- G: 代表一个 goroutine, 每个 goroutine 都有自己独立的栈存放当前运行内存及状态. 可以认为一个 G 就是一个任务.

- M: 代表内核线程(pthread), 它本身与一个 `内核线程` 进行绑定, goroutine 运行在M上.

- P: 代表一个资源, 可以认为一个 "有运行任务" 的P占了一个 `CPU线程` 的资源, 且只要处于调度的时候就有P.

> 注: 内核线程与CPU线程的区别, 在系统里可以有上万个内核线程, 但 CPU 线程并没有那么多, CPU线程也就是 top 命令里看到的
%Cpu0, %Cpu1, ...的数量. 


### 进程启动都做了什么?

```cgo
// runtime/go_tls.h

// 宏定义函数
#ifdef GOARCH_amd64
#define	get_tls(r)	MOVQ TLS, r
#define	g(r)	0(r)(TLS*1)
#endif
```

```
// runtime/asm_amd64.s

TEXT runtime·rt0_go(SB),NOSPLIT,$0
    ....
ok:
	// set the per-goroutine and per-mach "registers"
	get_tls(BX) // 宏定义函数, "go_tls.h" 文件中, 获取 TLS 变量, tls
	LEAQ	runtime·g0(SB), CX // CX = runtime.g0
	MOVQ	CX, g(BX)  // tls = g0
	LEAQ	runtime·m0(SB), AX // AX = runtime.m0
    
    // 全局 m0 与 g0 进行绑定
	// save m0->g0 = g0
	MOVQ	CX, m_g0(AX)
	// save g0->m = m0
	MOVQ	AX, g_m(CX)

	CLD				// convention is D is always left cleared
	CALL	runtime·check(SB)

	MOVL	16(SP), AX		// copy argc
	MOVL	AX, 0(SP)
	MOVQ	24(SP), AX		// copy argv
	MOVQ	AX, 8(SP)
	CALL	runtime·args(SB)   // 解析命令行
	CALL	runtime·osinit(SB) // 初始化CPU核数
	CALL	runtime·schedinit(SB) // 内存分配器, 栈, P, GC 回收等初始化

	// create a new goroutine to start program
	MOVQ	$runtime·mainPC(SB), AX
	PUSHQ	AX
	PUSHQ	$0			// arg size
	CALL	runtime·newproc(SB) // 创建一个新的 G 来启动 runtime.main
	POPQ	AX
	POPQ	AX

	// start this M
	CALL	runtime·mstart(SB) // 启动 m0, 开始等待空闲 G, 正式进入调度循环
	
	....
```

> M0 是什么? 程序会启动多个 M, 第一个启动的是 M0
>
> G0 是什么? G 分为三种, 第一种是用户任务的叫做 G. 第二种是执行 runtime 下调度工作的叫 G0, 每一个 M 都绑定一个 
> G0. 第三种是启动 runtime.main 用到的 G. 程序用到是基本上就是第一种.


### runtime.osinit(SB) 针对系统环境的初始化

### runtime.schedinit(SB) 调度相关的初始化

```cgo
func schedinit() {
	_g_ := getg()
    
    // 设置 M 最大数量
	sched.maxmcount = 10000

	tracebackinit()
	moduledataverify()
	stackinit()
	mallocinit()
	mcommoninit(_g_.m) // 初始化当前 M, 全局的 M0
	cpuinit()       // must run before alginit
	alginit()       // 初始化哈希算法, 在此之前不能使用 map
	modulesinit()   // provides activeModules
	typelinksinit() // uses maps, activeModules
	itabsinit()     // uses activeModules

	msigsave(_g_.m)
	initSigmask = _g_.m.sigmask

	goargs()
	goenvs()
	parsedebugvars()
	gcinit()

	sched.lastpoll = uint64(nanotime())
	procs := ncpu
	if n, ok := atoi32(gogetenv("GOMAXPROCS")); ok && n > 0 {
		procs = n
	}
	if procresize(procs) != nil {
		throw("unknown runnable goroutine during bootstrap")
	}
    
    ....	
}
```

sched.maxmcount 设置了 M 最大数量, 而 M 代表系统内核线程, 因此一个进程最大只能启动10000个系统内核线程. 

procresize 初始化P的数量, procs 参数为初始化的数量, 而在初始化之前先做数量的判断, 默认是ncpu(CPU核数). 也可以通过
GOMAXPROCS来控制P的数量. _MaxGomaxprocs 控制了最大数量只能是1024.


### runtime.mainPC(SB) 启动监控任务

```cgo
func main() {
	g := getg()

	// m0->g0的 racectx 仅用作main goroutine的parent, 不得将其用于其他任何用途.
	g.m.g0.racectx = 0

	// 设置 maxstacksize, 64-bit是 1GB, 32-bit 是 250M. 之所以使用十进制, 主要是方便查看信息.
	if sys.PtrSize == 8 {
		maxstacksize = 1000000000
	} else {
		maxstacksize = 250000000
	}

	// Allow newproc to start new Ms.
	mainStarted = true
    
    // 启动 sysmon 系统线程, 后台监控
	if GOARCH != "wasm" { // no threads on wasm yet, so no sysmon
		systemstack(func() {
			newm(sysmon, nil, -1)
		})
	}
    
    ....
}
```

在 runtime 下启动一个全程运行的监控任务, 该任务用于标记抢占时间过长的 G, 以及检测 epoll 里面有没有可执行的G. 

### runtime.mstart(SB) 启动调度循环

调度循环:

![image](/images/develop_runtime_schedule.png)

- 图1代表 M 启动的过程, 把 M 跟一个 P 绑定在一起. 在程序初始化的过程中进程启动的最后一步启动第一个M(即M0), 这个M从全
局的空闲的P列表里拿到一个P, 然后与其绑定. 而 P 里面有2个管理G的链表(runq存储待运行的G列表, gfree存储空闲的G列表), M
启动后等待可执行的G

- 图2代表创建 G 的过程. 创建完一个 G 先扔到当前 P 的 runq 带运行队列里. 在图 3 的执行过程, M 从绑定的P的 runq 列表
里获取一个 G 来执行. 当中执行完成之后, 图4的流程里把 G 扔到 gfree 列表里. 注意: 此时的 G 并没有销毁(只重置了G的栈以
及状态), 当再次创建G的时候优先从 gfree 当中获取, 这样达到复用 G 的目的.

- 图3代表执行一个 G 的过程.

- 图4代表释放一个 G 的过程.

> M 启动后处于一个自循环状态, 执行完一个 G 之后继续执行下一个 G, 反复上面的 `图2 ~ 图4` 的过程. 当第一个 M 正在繁忙而
又有新的 G 需要执行时, 会再开启一个 M 来执行.


### 调度器如何开启调度循环

mstart() 启动 M, 调用的链: `mstart()` -> `mstart1()` -> `schedule()`

```cgo
func mstart() {
	_g_ := getg()

	...
	
	mstart1()

	...
	
	mexit(osStack)
}
```

```cgo
func mstart1() {
	_g_ := getg()

	if _g_ != _g_.m.g0 {
		throw("bad runtime·mstart")
	}

	// Record the caller for use as the top of stack in mcall and
	// for terminating the thread.
	// We're never coming back to mstart1 after we call schedule,
	// so other calls can reuse the current frame.
	save(getcallerpc(), getcallersp())
	asminit()
	minit()
    
    // 只在汇编当中调用才执行此函数。
	// 初始化信号处理程序; minit之后, 以便 minit 可以准备线程以处理信号.
	if _g_.m == &m0 {
		mstartm0()
	}

    // 初始化函数
	if fn := _g_.m.mstartfn; fn != nil {
		fn()
	}

	if _g_.m != &m0 {
		acquirep(_g_.m.nextp.ptr())
		_g_.m.nextp = 0
	}
	// 执行调度函数
	schedule()
}
```


// 创建一个 M, 并绑定到一个系统内核线程, 系统内核线程的启动函数是 mstart (在 g0 上执行), 初始化函数是 fn(在 mstart 
当中被调用执行, 也是运行在 g0 上的.)

```cgo
// 调度 M 的执行 P 当中的 G, 即 M 的启动函数
// 如果p == nil, 则尝试获取一个空闲P, 如果没有空闲P则不执行任何操作.
// 可以与 m.p == nil 一起运行, 因此不允许写入障碍.
// 如果设置了spin, 则调用者已增加nmspinning, 而startm将减少nmspinning或在新启动的M中设置m.spinning.
func startm(_p_ *p, spinning bool) {
	lock(&sched.lock)
	if _p_ == nil {
	    // 从空闲的 p 当中获取一个 p
		_p_ = pidleget()
		if _p_ == nil {
			unlock(&sched.lock)
			if spinning {
				if int32(atomic.Xadd(&sched.nmspinning, -1)) < 0 {
					throw("startm: negative nmspinning")
				}
			}
			return
		}
	}
	
    // 获取一个空闲的 M
	mp := mget()
	if mp == nil {
		// 没有可用的M, 此时调用newm. 但是, 我们已经拥有一个P来分配给M.
        //
        // 释放sched.lock后, 另一个G(例如,在系统调用中)可能找不到空闲的P, 而checkdead发现可运行的G但没有运行的M, 
        // 因为此新M尚未启动, 因此引发了明显的死锁.
        //
        // 通过为新M预先分配ID来避免这种情况, 从而在释放 sched.lock 之前将其标记为 "正在运行". 这个新的M最终将运行
        // 调度程序以执行所有排队的G.
		id := mReserveID()
		unlock(&sched.lock)

		var fn func()
		if spinning {
			// The caller incremented nmspinning, so set m.spinning in the new M.
			fn = mspinning
		}
		newm(fn, _p_, id)
		return
	}
	unlock(&sched.lock)
	if mp.spinning {
		throw("startm: m is spinning")
	}
	if mp.nextp != 0 {
		throw("startm: m has p")
	}
	if spinning && !runqempty(_p_) {
		throw("startm: p has runnable gs")
	}
	// The caller incremented nmspinning, so set m.spinning in the new M.
	mp.spinning = spinning
	mp.nextp.set(_p_)
	notewakeup(&mp.park) // 唤醒 M
}

// newm -> newm1 -> newosproc(传入 mstart函数指针, 作为启动函数, 运行在 g0 上)
// 创建一个 M, 设置 M 的初始化函数是 fn (这个函数是在 g 上执行）
func newm(fn func(), _p_ *p, id int64) {
    // 创建一个M对象, 且与P关联
	mp := allocm(_p_, fn, id)
	// 暂存 P
	mp.nextp.set(_p_)
	mp.sigmask = initSigmask
	
	// 当前的 G 绑定了一个锁定的 M
	if gp := getg(); gp != nil && gp.m != nil && (gp.m.lockedExt != 0 || gp.m.incgo) && GOOS != "plan9" {
		// 我们处于锁定的M或可能由C启动的线程上. 此线程的内核状态可能很奇怪(用户可能已为此目的将其锁定).
		// 我们不想将其克隆到另一个线程中. 而是要求一个已知良好的线程为我们创建线程.	
		lock(&newmHandoff.lock)
		if newmHandoff.haveTemplateThread == 0 {
			throw("on a locked thread with no template thread")
		}
		mp.schedlink = newmHandoff.newm
		newmHandoff.newm.set(mp)
		if newmHandoff.waiting {
			newmHandoff.waiting = false
			notewakeup(&newmHandoff.wake)
		}
		unlock(&newmHandoff.lock)
		return
	}
	newm1(mp)  // 创建系统内核线程
}

func newm1(mp *m) {
    // cgo 相关的创建系统内核线程
	if iscgo {
		...
		return
	}
	
	// 创建系统内核线程
	execLock.rlock() 
	newosproc(mp) // 创建系统内核线程
	execLock.runlock()
}

// 创建系統内核线程, 初始化运行的函数是 mstart. 
func newosproc(mp *m) {
    // 栈顶位置
	stk := unsafe.Pointer(mp.g0.stack.hi)
	
	// 在 clone 过程中禁用信号, 以便新线程以禁用的信号开始.
	// 之后它将调用 minit 
	var oset sigset
	sigprocmask(_SIG_SETMASK, &sigset_all, &oset)
	// clone, funcPC函数用户获取函数的位置指针
	ret := clone(cloneFlags, stk, unsafe.Pointer(mp), unsafe.Pointer(mp.g0), unsafe.Pointer(funcPC(mstart)))
	sigprocmask(_SIG_SETMASK, &oset, nil)
}


// 分配与任何线程无关的新 m. 
// 如果需要, 可以将 p 用于分配上下文.
// fn 被记录为新 m 的 m.mstartfn. 
// id 是可选的预分配的 m ID. -1 会被忽略
//
// 此函数即使没有调用者也被允许有写障碍, 因为它借用了_p_.
func allocm(_p_ *p, fn func(), id int64) *m {
	_g_ := getg()
	acquirem() // disable GC because it can be called from sysmon
	if _g_.m.p == 0 {
		acquirep(_p_) // 在此函数中临时为mallocs借用p
	}

	// 释放 freem 链表, 一旦 M 已经被释放, 则会从 freem 当中删除
	if sched.freem != nil {
		lock(&sched.lock)
		var newList *m
		for freem := sched.freem; freem != nil; {
			if freem.freeWait != 0 {
				next := freem.freelink
				freem.freelink = newList
				newList = freem
				freem = next
				continue
			}
			stackfree(freem.g0.stack) // 释放 stack
			freem = freem.freelink
		}
		sched.freem = newList
		unlock(&sched.lock)
	}

	mp := new(m)
	mp.mstartfn = fn   // 设置初始化函数
	mcommoninit(mp, id) // 初始化m
    
    // 创建 g0, 并将 g0 与 m 进行绑定.
	// 如果是 cgo 或 Solaris 或 illumos 或 Darwin, pthread_create将创建堆栈.
    // Windows 和 Plan9 将调度的堆栈安排在OS堆栈上.
	if iscgo || GOOS == "solaris" || GOOS == "illumos" || GOOS == "windows" || GOOS == "plan9" || GOOS == "darwin" {
		mp.g0 = malg(-1)
	} else {
		mp.g0 = malg(8192 * sys.StackGuardMultiplier) // 8K
	}
	mp.g0.m = mp

	if _p_ == _g_.m.p.ptr() {
		releasep()
	}
	releasem(_g_.m)

	return mp
}
```

> 非 M0 的启动函数首先从 startm 方法开始启动, 要进行调度工作必须有调度处理器P, 因此先从空闲的 P 链表里获取一个 P, 在 
newm 方法创建一个 M 与 P 绑定.


> newm 中通过调用 newosproc 新创建一个内核线程, 并把内核线程与M以及 mstart 方法进行关联, 这样内核线程执行时就可以找
到 M 并且启动调度循环的方法. 最后 schedule 启动调度循环.


> allocm 方法中创建M的同时创建一个 G 与自己关联, 这个 G 就是 g0. 为何 M 要关联一个 g0 ? 因为 runtime 下执行一个 G
也需要用到栈空间来完成调度工作, 而拥有执行栈的地方只有 G, 因此需要为每个 M 配置一个 g0


### 调度器如何进入调度循环的

调用 schedule 进入调度循环后, 这个方法永远不会再返回.

```cgo
// runtime/proc.go

func schedule() {
	_g_ := getg()

	if _g_.m.locks != 0 {
		throw("schedule: holding locks")
	}

	if _g_.m.lockedg != 0 {
		stoplockedm()
		execute(_g_.m.lockedg.ptr(), false) // Never returns.
	}

	// We should not schedule away from a g that is executing a cgo call,
	// since the cgo call is using the m's g0 stack.
	if _g_.m.incgo {
		throw("schedule: in cgo")
	}

top:
    // 获取 P, 并将抢占变量设置为 false
	pp := _g_.m.p.ptr()
	pp.preempt = false

	if sched.gcwaiting != 0 {
		gcstopm()
		goto top
	}
	if pp.runSafePointFn != 0 {
		runSafePointFn()
	}

	
	// 进行完整性检查: 如果我们正在 spinning, 则 runq 应该为空.
    // 在调用checkTimers之前检查它, 因为这可能会调用 goready 将就绪的 goroutine 放在本地运行队列中.
	if _g_.m.spinning && (pp.runnext != 0 || pp.runqhead != pp.runqtail) {
		throw("schedule: spinning with local work")
	}

	checkTimers(pp, 0)

	var gp *g
	var inheritTime bool

	// 普通的goroutine会检查是否需要就绪, 但 GCworkers 和 tracereaders 不会这样做, 而必须在此处进行检查.
	tryWakeP := false
	if trace.enabled || trace.shutdown {
		gp = traceReader()
		if gp != nil {
			casgstatus(gp, _Gwaiting, _Grunnable)
			traceGoUnpark(gp, 0)
			tryWakeP = true
		}
	}
	
	// 进入gc MarkWorker 工作模式
	if gp == nil && gcBlackenEnabled != 0 {
		gp = gcController.findRunnableGCWorker(_g_.m.p.ptr())
		tryWakeP = tryWakeP || gp != nil
	}
	
	// 开始查找 gp (需要调度的任务)
	if gp == nil {
	    // 全局队列
		if _g_.m.p.ptr().schedtick%61 == 0 && sched.runqsize > 0 {
			lock(&sched.lock)
			gp = globrunqget(_g_.m.p.ptr(), 1)
			unlock(&sched.lock)
		}
	}
	if gp == nil {
	    // M 绑定的 P 的本地队列
		gp, inheritTime = runqget(_g_.m.p.ptr())
	}
	if gp == nil {
	    // 从其他的 M 当中获取, 这个方法是阻塞执行的
		gp, inheritTime = findrunnable() // blocks until work is available
	}

	// 此线程将运行goroutine并且不再自旋, 因此, 如果将其标记为正在旋转, 则需要立即将其重置, 
	// 并可能启动新的旋转M.
	if _g_.m.spinning {
		resetspinning()
	}

	if sched.disable.user && !schedEnabled(gp) {
		// Scheduling of this goroutine is disabled. Put it on
		// the list of pending runnable goroutines for when we
		// re-enable user scheduling and look again.
		lock(&sched.lock)
		if schedEnabled(gp) {
			// Something re-enabled scheduling while we
			// were acquiring the lock.
			unlock(&sched.lock)
		} else {
			sched.disable.runnable.pushBack(gp)
			sched.disable.n++
			unlock(&sched.lock)
			goto top
		}
	}

	// If about to schedule a not-normal goroutine (a GCworker or tracereader),
	// wake a P if there is one.
	if tryWakeP {
		wakep()
	}
	if gp.lockedm != 0 {
		// Hands off own p to the locked m,
		// then blocks waiting for a new p.
		startlockedm(gp)
		goto top
	}
    
    // 执行查找到的 G
	execute(gp, inheritTime)
}
```

schedule 中首先尝试从本地 P 获取(runqget)一个可执行的 G, 如果没有则从其他地方获取(findrunnable), 最终通过 execute
方法执行 G.

runqget 先通过 runnext 拿到待运行 G, 没有的话, 再从 runq 里获取.

findrunnable 从全局队列, epoll, 其他的 P 里获取.

> 在调度开通做了个优化: 每处理一些任务之后, 就会优先从全局队列里获取任务, 以保障公平性, 防止由于每个 P 里的 G 过多, 而
全局队列里的任务得不到执行机会.


### 调度循环中如何让出 CPU

- 正常完成让出 CPU

执行过程:

schedule -> execute -> gogo -> 执行G

创建 G:
     
newproc1 -> pc指向 goexit( goexit -> goexit0 -> goexit1-> schedule)


绝大多数场景下程序都行执行完成一个 G, 再执行另一个 G. 

```cgo
// runtime/proc.go

func execute(gp *g, inheritTime bool) {
	_g_ := getg()

	// 在进入 _Grunning 之前绑定 gp.m, 以便运行的G具有M.
	_g_.m.curg = gp
	gp.m = _g_.m
	
	// 修改状态为 _Grunning
	casgstatus(gp, _Grunnable, _Grunning)
	gp.waitsince = 0
	gp.preempt = false
	gp.stackguard0 = gp.stack.lo + _StackGuard 
	if !inheritTime {
		_g_.m.p.ptr().schedtick++
	}

	// 检查是否需要打开或关闭 profiler, 这个是进行 pprof 开启状况下进行捕捉信息使用的.
	hz := sched.profilehz
	if _g_.m.profilehz != hz {
		setThreadCPUProfiler(hz)
	}

	if trace.enabled {
		// GoSysExit has to happen when we have a P, but before GoStart.
		// So we emit it here.
		if gp.syscallsp != 0 && gp.sysblocktraced {
			traceGoSysExit(gp.sysexitticks)
		}
		traceGoStart()
	}

    // 真正的执行G, 切换到该G的栈帧上执行(汇编实现)
	gogo(&gp.sched)
}
```

// 汇编实现的 gogo 函数。

主要做两件事: 一是绑定 gobuf.g 到当前线程的 tls 当中, 二是跳转到 gobuf.pc 的位置, 继续往下执行.

```
// runtime/asm_amd64.s

// func gogo(buf *gobuf), 从 gobuf 当中恢复状态
// gogo函数的stack大小是16, arg大小是8
TEXT runtime·gogo(SB), NOSPLIT, $16-8
	MOVQ	buf+0(FP), BX		// gobuf
	MOVQ	gobuf_g(BX), DX     // gobuf.g
	MOVQ	0(DX), CX		    // make sure g != nil
	get_tls(CX)   // CX 当中存储 gobuf.g 的 tls 内容
	MOVQ	DX, g(CX) // 将 gobuf.g 存储到当前内核线程的 tls 当中
	MOVQ	gobuf_sp(BX), SP	// gobuf.sp
	MOVQ	gobuf_ret(BX), AX   // gobuf.ret
	MOVQ	gobuf_ctxt(BX), DX  // gobuf.ctxt
	MOVQ	gobuf_bp(BX), BP    // gobuf.bp
	MOVQ	$0, gobuf_sp(BX)	// 清空 gobuf 的 sp, ret, ctxt, bp
	MOVQ	$0, gobuf_ret(BX)
	MOVQ	$0, gobuf_ctxt(BX)
	MOVQ	$0, gobuf_bp(BX)
	MOVQ	gobuf_pc(BX), BX    // gobuf.pc
	JMP	BX // 跳跃到 gobuf.pc 指定的位置
```

// gobuf 数据结构
```cgo
type gobuf struct {
  // libmach已知(硬编码)sp, pc 和 g 的偏移量.
  //
  // ctxt 对于GC比较特殊: 它可能是在 heap 上分配的funcval, 因此GC需要对其进行跟踪, 但需要对其进行设置并将其从汇编中
  // 清除, 这在其中很难有写障碍. 但是, ctxt实际上是一个实时寄存器, 我们只在真实寄存器和gobuf之间交换它. 因此,
  // 我们在堆栈扫描期间将其视为根, 这意味着保存和恢复它的程序集不需要写障碍. 它仍然被用作指针, 以便Go进行的任何
  // 其他写入都会遇到写入障碍.
  sp   uintptr
  pc   uintptr
  g    guintptr
  ctxt unsafe.Pointer
  ret  sys.Uintreg
  lr   uintptr
  bp   uintptr // for GOEXPERIMENT=framepointer
}
```

gogo 方法传的参数是 gp.sched, 这个结构体里保存了函数栈寄存器 SP/PC/BP, G(需要执行的goroutine). gogo函数中实质就只
是做了函数栈指针的移动.

C 语言里栈帧创建的时候有个IP寄存器指向 "return address", 即主调函数的一条指令的地址, 被调函数退出的时候通过该指针回到
调用函数里. 在Go语言里有个 PC 寄存器指向退出函数. 那么该PC寄存器指向哪里呢?

```cgo
// 从fn开始创建状态为_Grunnable 的新g, 其中narg个字节的参数从argp开始.
// callerpc 是创建它的go语句的地址. 调用者负责将新的 g 添加到调度程序中.
func newproc1(fn *funcval, argp unsafe.Pointer, narg int32, callergp *g, callerpc uintptr) *g {
	_g_ := getg()

    ......

	_p_ := _g_.m.p.ptr()
	
	// 从当前P里面复用一个空闲G
	newg := gfget(_p_)
	// 如果没有空闲G则新建一个, 默认堆大小为_StackMin=2048 bytes
	if newg == nil {
		newg = malg(_StackMin)
		casgstatus(newg, _Gidle, _Gdead)
		// 把新创建的G添加到全局allg里
		allgadd(newg) // publishes with a g->status of Gdead so GC scanner doesn't look at uninitialized stack.
	}
	
	......
	
	newg.sched.sp = sp
	newg.stktopsp = sp
	// 记录当前任务的pc寄存器为goexit方法, 用于当执行G结束后找到退出方法, 从而再次进入调度循环.
	newg.sched.pc = funcPC(goexit) + sys.PCQuantum // +PCQuantum so that previous instruction is in same function
	newg.sched.g = guintptr(unsafe.Pointer(newg))
	gostartcallfn(&newg.sched, fn)
	newg.gopc = callerpc
	newg.ancestors = saveAncestors(callergp)
	newg.startpc = fn.fn
	
	...... 

	return newg
}
```

G的执行环境里的 pc 变量赋值了一个 goexit 的函数地址, 也就是说G正常执行完退出时执行的是 goexit 函数.


goexit 函数的汇编实现, 调用函数 goexit1 函数.

```
TEXT runtime·goexit(SB),NOSPLIT,$0-0
	BYTE	$0x90	// NOP
	CALL	runtime·goexit1(SB)	// does not return
	// traceback from goexit1 must hit code range of goexit
	BYTE	$0x90	// NOP
```


```cgo
// G执行结束后回到这里放到P的本地队列里
func goexit1() {
	if raceenabled {
		racegoend()
	}
	if trace.enabled {
		traceGoEnd()
	}
	
	// 切换到g0来释放G
	mcall(goexit0)
}

// g0下当G执行结束后回到这里放到P的本地队列里
func goexit0(gp *g) {
	_g_ := getg()

	......
	
	gfput(_g_.m.p.ptr(), gp)
	
	...... 
	
	schedule()
}
```

代码中切换到了g0下执行了 schedule 方法, 再次进度了下一轮调度循环. 

以上就是正常执行一个G并正常退出的实现.


- 主动让出 CPU

- 抢占让出 CPU

- 系统调用让出 CPU



### 相关数据结构

// g, Groutine
```cgo
type g struct{
    stack     stack   // g自身的栈 
    m         *m      // 隶属于哪个m 
    sched     gobuf   // 保存了g的现场, goroutine切换时通过它来恢复
    atomicstatus uint32 // G的状态
    schedlink guintptr // 下一个 G, 链表
    lockedm   muintptr  // 锁定的M, G中断恢复指定M执行
    gopc      uintptr   // 创建 goroutine的指令地址
    startpc   uintptr   // gorotine 函数指令地址
    ... 
}
```

// m, 系统线程
```cgo
type m struct{
    g0    *g        // g0， 每个M都有自己独有的 g0，负责调度
    curg  *g        // 当前正在运行的 g
    p     puintptr  // 绑定 P 执行代表(如果是nil, 则处于空闲状态)
    nextp puintptr  // 当 M 被唤醒时,首先拥有这个 P
    oldp  puintptr  // 之前执行 syscall 绑定的 P
    
    schedlink muintptr // 下一个m, 构成链表
    lockedg   guintptr // 和 G 的 lockedm 对应
    freelink  *m       // sched.freem  
    
    tls [6]uintptr 
    mstartfn func()    // m 启动执行的函数
    ... 
}
```

// p, Procs
```cgo
type p struct{
    status int32 // P 的状态
    link   puintptr // 下一个P, 构成链表
    m      muintptr // 拥有这个P的M
    
    // 一个比 runq 优先级更高的的 runable G
    runnext guintptr
    
    // P 本地 runable G 队列, 无锁访问
    runqhead uint32
    runqtail uint32
    runq     [256]guintptr 
    
    // 状态为 dead 的 G 链表, 获取 G 时会从这里获取.
    gFree struct{
        gList
        n int32
    }
}
```

// sched 全局调度
```cgo
// 全局的调度 sched
type schedt struct{
    midle muintptr // 空闲的 M 链表
    
    pidle puintptr // 空闲的 P 链表
    
    runq gQueue    // 全局 runnable 的 G 队列
    runqsize int32 // 全局 runnable 的 G 队列大小
    
    // 全局 dead G
    gFree struct{
        lock    mutex
        stack   gList // 带有 stack 的 G
        noStack gList // 不带 stack 的 G
        n int32
    } 
    
    // 全局等待释放的 M (m.exited已经被设置), 链接到 m.freelink
    freem *m 
}
```