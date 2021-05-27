### 调度循环中如何让出 CPU

#### 正常完成让出 CPU

正常执行过程:

```
schedule()->execute()->gogo()->xxx()->goexit()->goexit1()->mcall()->goexit0()->schedule()
```

其中 `schedule()->execute()->gogo()` 是在 g0 上执行的. `xxx()->goexit()->goexit1()->mcall()` 是在 curg 
上执行的. `goexit0()->schedule()` 又是在 g0 上执行的.


#### 主动让出 CPU

在实际场景中还有一些没有执行完成的 G, 而又需要临时停止执行. 比如, time.Sleep, IO阻塞等, 就要挂起该 G, 把CPU让出来
给其他 G 使用. 在 runtime 的 gopark 方法:

```cgo
// runtime/proc.go 

func gopark(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, 
    traceEv byte, traceskip int) {
	if reason != waitReasonSleep {
		checkTimeouts() // timeouts may expire while two goroutines keep the scheduler busy
	}
	mp := acquirem()
	gp := mp.curg
	status := readgstatus(gp)
	if status != _Grunning && status != _Gscanrunning {
		throw("gopark: bad g status")
	}
	mp.waitlock = lock
	mp.waitunlockf = unlockf
	gp.waitreason = reason
	mp.waittraceev = traceEv
	mp.waittraceskip = traceskip
	releasem(mp)
	
	// mcall 在M里从当前正在运行的G切换到g0
    // park_m 在切换到的g0下先把传过来的 G 切换为_Gwaiting 状态挂起该G
    // 调用回调函数waitunlockf()由外层决定是否等待解锁, 返回true则等待解锁不在执行G, 返回false则不等待解锁继续执行.
	mcall(park_m)
}

// 在 g0 上执行
func park_m(gp *g) {
	_g_ := getg()

	if trace.enabled {
		traceGoPark(_g_.m.waittraceev, _g_.m.waittraceskip)
	}
    
    // 切换状态, _Grunning -> _Gwaiting
	casgstatus(gp, _Grunning, _Gwaiting)
	dropg() // 把g0从M的"当前运行"里剥离出来.
    
    // 如果不需要等待解锁, 则切换到_Grunnable状态并直接执行G
	if fn := _g_.m.waitunlockf; fn != nil {
		ok := fn(gp, _g_.m.waitlock)
		_g_.m.waitunlockf = nil
		_g_.m.waitlock = nil
		if !ok {
			if trace.enabled {
				traceGoUnpark(gp, 2)
			}
			casgstatus(gp, _Gwaiting, _Grunnable)
			execute(gp, true) // Schedule it back, never returns.
		}
	}
	
	// 需要等待解锁, 不再执行当前的 G
	schedule()
}
``` 

```cgo
// func mcall(fn func(*g)), 切换到 m->g0 的栈(SP=g0.sched.sp), 调用 fn(g), fn必须永远不返回. 
// 在 g 当中保存了运行状态, bp, g, bp
TEXT runtime·mcall(SB), NOSPLIT, $0-8
	MOVQ	fn+0(FP), DI  // fn 函数

	get_tls(CX)  // CX=tls
	MOVQ	g(CX), AX	// AX = g
	MOVQ	0(SP), BX	// caller's PC
	MOVQ	BX, (g_sched+gobuf_pc)(AX) // 设置 g.sched.pc 的值为当前的 SP
	LEAQ	fn+0(FP), BX	// caller's SP, 函数地址传送
	
	// 保存 sp, g, bp 到 g->sched 当中
	MOVQ	BX, (g_sched+gobuf_sp)(AX) // sp 保存了 mcall 调用的函数地址 
	MOVQ	AX, (g_sched+gobuf_g)(AX)  // g 保存了当前线程的 g
	MOVQ	BP, (g_sched+gobuf_bp)(AX) // bp 保存了BP寄存器(SP伪寄存器, 一个标记, 包含在调用栈当中)

	// 切换到 m->g0 & its stack, 执行 fn
	MOVQ	g(CX), BX  // g
	MOVQ	g_m(BX), BX // g.m
	MOVQ	m_g0(BX), SI // g.m.g0
	CMPQ	SI, AX	// g == m->g0
	JNE	3(PC)  // 不等于则跳转 pc+3 的位置, 即 "MOVQ	SI, g(CX)" 位置
	MOVQ	$runtime·badmcall(SB), AX // 获取 badmcall 函数地址
	JMP	AX                            // 跳转到对应函数地址, 执行函数 
	MOVQ	SI, g(CX)	// 切换, 设置 g = m->g0
	MOVQ	(g_sched+gobuf_sp)(SI), SP	// sp = m->g0->sched.sp
	PUSHQ	AX // 调用方的 g 
	MOVQ	DI, DX // fn 函数
	MOVQ	0(DI), DI // 获取 DI 偏移量为0的数据, 即判断 fn ！= nil
	CALL	DI // 调用 fn 函数, 参数是调用方的 g
	POPQ	AX // 正常状况下, 执行不到这里, 因为 fn 函数永不返回.
	MOVQ	$runtime·badmcall2(SB), AX
	JMP	AX
	RET
```


gopark 是进行调度让出 CPU 资源的方法, 里面调用了 mcall(), 注释是这样描述的:

> 从当前运行的 g 切换到 g0 的运行栈, 然后调用 fn(g), 这里被调用的 g 是调用 mcall 方法时的 G. mcall 方法保存当前
运行 G 的 PC/SP 到 g->sched 里, 因此该 G 可以在以后被重新恢复执行.


M 创建时绑定一个 g0, 调度工作运行在 g0 栈上的. mcall 方法通过 g0 先把当前调用的 G 的执行栈暂存到 g->sched 变量里,
然后切换到 g0 的执行栈上执行 park_m. park_m 方法的状态把 gp(当前调用的G) 的状态从 _Grunning 切换到 _Gwaiting 表
明进入到了等待唤醒状态, 此时休眠 G 的操作就完成了. 接下来既然 G 休眠了, CPU 线程总不能闲下来, 在 park_m 当中又出现了
schedule 方法, 开启新的一轮调度循环了.

> 进入调度循环之前还有个对 waitunlockf 方法的判断, 该方法是如果不需要等待解锁则调用调用 execute 方法继续执行之前的 G,
而该方法永远不会 return, 也不会再次进入下一次调度. 这是一个外部一个控制是否要进行下一轮调度的选择.

#### 抢占让出 CPU

```cgo

// 始终在不带P的情况下运行, 因此不存在写屏障.
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
		usleep(delay)
		now := nanotime()
		next, _ := timeSleepUntil()
		// 双检查
		if debug.schedtrace <= 0 && (sched.gcwaiting != 0 || atomic.Load(&sched.npidle) == uint32(gomaxprocs)) {
			lock(&sched.lock)
			if atomic.Load(&sched.gcwaiting) != 0 || atomic.Load(&sched.npidle) == uint32(gomaxprocs) {
				if next > now {
					atomic.Store(&sched.sysmonwait, 1)
					unlock(&sched.lock)
					// Make wake-up period small enough
					// for the sampling to be correct.
					sleep := forcegcperiod / 2
					if next-now < sleep {
						sleep = next - now
					}
					shouldRelax := sleep >= osRelaxMinNS
					if shouldRelax {
						osRelax(true)
					}
					notetsleep(&sched.sysmonnote, sleep)
					if shouldRelax {
						osRelax(false)
					}
					now = nanotime()
					next, _ = timeSleepUntil()
					lock(&sched.lock)
					atomic.Store(&sched.sysmonwait, 0)
					noteclear(&sched.sysmonnote)
				}
				idle = 0
				delay = 20
			}
			unlock(&sched.lock)
		}
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
		// poll network if not polled for more than 10ms
		lastpoll := int64(atomic.Load64(&sched.lastpoll))
		if netpollinited() && lastpoll != 0 && lastpoll+10*1000*1000 < now {
			atomic.Cas64(&sched.lastpoll, uint64(lastpoll), uint64(now))
			list := netpoll(0) // non-blocking - returns list of goroutines
			if !list.empty() {
				// Need to decrement number of idle locked M's
				// (pretending that one more is running) before injectglist.
				// Otherwise it can lead to the following situation:
				// injectglist grabs all P's but before it starts M's to run the P's,
				// another M returns from syscall, finishes running its G,
				// observes that there is no work to do and no other running M's
				// and reports deadlock.
				// 需要在 injectglist 之前减少 "空闲锁定的M" 的数量(假设还有一个正在运行). 
				// 否则, 可能导致以下情况: injectglist 捕获所有P, 但在启动M运行P之前, 另一个M从syscall返回, 完
				// 成运行G, 观察到没有工作要做没有其他正在运行的M, 并报告了死锁.
				incidlelocked(-1)
				
				// 把 epoll ready 的G列表注入到全局runq里
				injectglist(&list)
				incidlelocked(1)
			}
		}
		if next < now {
			// There are timers that should have already run,
			// perhaps because there is an unpreemptible P.
			// Try to start an M to run them.
			startm(nil, false)
		}
		if atomic.Load(&scavenge.sysmonWake) != 0 {
			// Kick the scavenger awake if someone requested it.
			wakeScavenger()
		}
		// retake P's blocked in syscalls
		// and preempt long running G's
		if retake(now) != 0 {
			idle = 0
		} else {
			idle++
		}
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