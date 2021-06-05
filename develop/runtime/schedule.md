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
			// - sched.nmspinning 或 sched.npidle 不为0,
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

#### 系统调用让出 CPU

Go 中没有直接对系统内核函数调用, 而是封装了 syscall.Syscall 方法.

```cgo
// syscall/syscall_unix.go

func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
```

```
// func Syscall(trap int64, a1, a2, a3 uintptr) (r1, r2, err uintptr);
// trap in AX, args in DI SI DX R10 R8 R9, return in AX DX
// 注意, 这与 "标准" ABI 约定不同, 后者将在 CX 中传递第4个arg, 而不是R10.

TEXT ·Syscall(SB),NOSPLIT,$0-56
	CALL	runtime·entersyscall(SB)
	MOVQ	a1+8(FP), DI
	MOVQ	a2+16(FP), SI
	MOVQ	a3+24(FP), DX
	MOVQ	trap+0(FP), AX	// 系统调用号
	SYSCALL                 // 系统调用
	CMPQ	AX, $0xfffffffffffff001  // 0xfffffffffffff001 是 linux MAX_ERRNO 取反转无符号
	JLS	ok
	MOVQ	$-1, r1+32(FP)
	MOVQ	$0, r2+40(FP)
	NEGQ	AX
	MOVQ	AX, err+48(FP)
	CALL	runtime·exitsyscall(SB)
	RET
ok:
	MOVQ	AX, r1+32(FP)
	MOVQ	DX, r2+40(FP)
	MOVQ	$0, err+48(FP)
	CALL	runtime·exitsyscall(SB)
	RET
```

汇编代码先是执行了 runtime.entersyscall 方法, 然后进行系统调用, 最后执行 runtime.exitsyscall 方法. 字面上是进入
系统调用先执行一些逻辑, 退出系统调用之后执行了一些逻辑.

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
func reentersyscall(pc, sp uintptr) {
	_g_ := getg()

	// Disable preemption because during this function g is in Gsyscall status,
	// but can have inconsistent g->sched, do not let GC observe it.
	_g_.m.locks++

	// Entersyscall must not call any function that might split/grow the stack.
	// (See details in comment above.)
	// Catch calls that might, by replacing the stack guard with something that
	// will trip any stack check and leaving a flag to tell newstack to die.
	_g_.stackguard0 = stackPreempt
	_g_.throwsplit = true

	// 保存执行现场
	save(pc, sp)
	_g_.syscallsp = sp
	_g_.syscallpc = pc
	casgstatus(_g_, _Grunning, _Gsyscall) // 状态切换, _Gsyscall
	// sp 合法性检查
	if _g_.syscallsp < _g_.stack.lo || _g_.stack.hi < _g_.syscallsp {
		systemstack(func() {
			print("entersyscall inconsistent ", hex(_g_.syscallsp), " [", hex(_g_.stack.lo), ",", hex(_g_.stack.hi), "]\n")
			throw("entersyscall")
		})
	}

	if trace.enabled {
		systemstack(traceGoSysCall)
		// systemstack itself clobbers g.sched.{pc,sp} and we might
		// need them later when the G is genuinely blocked in a
		// syscall
		save(pc, sp)
	}

	if atomic.Load(&sched.sysmonwait) != 0 {
		systemstack(entersyscall_sysmon)
		save(pc, sp)
	}

	if _g_.m.p.ptr().runSafePointFn != 0 {
		// runSafePointFn may stack split if run on this stack
		systemstack(runSafePointFn)
		save(pc, sp)
	}

    // 保存 syscalltick, 同时将 g.oldp 设置为当前的 p, g.p 置空.
    // p 和 m 解绑
	_g_.m.syscalltick = _g_.m.p.ptr().syscalltick
	_g_.sysblocktraced = true
	pp := _g_.m.p.ptr()
	pp.m = 0
	_g_.m.oldp.set(pp)
	_g_.m.p = 0
	atomic.Store(&pp.status, _Psyscall) // 切换 P 的状态
	if sched.gcwaiting != 0 {
		systemstack(entersyscall_gcwait)
		save(pc, sp)
	}

	_g_.m.locks--
}

```


#### 常用的函数

```cgo
wakep() // 在 sched.npidle >0 && sched.nmspinning == 0 的状况下启动一个自旋的 M, 也就是唤醒一个 P

dropg() // 解除 curg 和 m 的绑定关系

acquirep(p) // 将 p 与 m 进行绑定, 同时设置 p 的状态为 _Prunning
releasep()  // 将 p 与 m 解绑, 同时设置 p 的状态为 _Pidle


acquirem()  // 禁止抢占调度(m.locks++)
releasem(m) // 启用抢占调度(m.locks--), 如果当前处于抢占当中(m.locks==0&&_g_.preempt ), 
            // 修改 _g_.stackguard0 的值
```