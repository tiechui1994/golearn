### main goroutine 

下来, 看看 runtime.main 函数:

```cgo
func main() {
    g := getg() // 当前是 g, 即 m.curg
    
    // Racectx of m0->g0 is used only as the parent of the main goroutine.
    // It must not be used for anything else.
    g.m.g0.racectx = 0
    
    // 最大栈空间
    if sys.PtrSize == 8 {
        maxstacksize = 1000000000
    } else {
        maxstacksize = 250000000
    }
    
    // Allow newproc to start new Ms.
    mainStarted = true
    
    if GOARCH != "wasm" { 
        // 现在执行的是 main goroutine, 因此使用的是 main goroutine 的栈, 需要切换
        // 到 g0 栈去执行 newm() 函数.
        systemstack(func() {
            // 创建监控线程, 该线程独立于调度器, 不需要关联 p 即可运行(只在 g0 上运行)
            newm(sysmon, nil, -1)
        })
    }
    
    // Lock the main goroutine onto this, the main OS thread,
    // during initialization. Most programs won't care, but a few
    // do require certain calls to be made by the main thread.
    // Those can arrange for main.main to run in the main thread
    // by calling runtime.LockOSThread during initialization
    // to preserve the lock.
    lockOSThread()
    
    if g.m != &m0 {
        throw("runtime.main not on m0")
    }
    
    // 执行 runtime 包的 init 函数
    doInit(&runtime_inittask) // must be before defer
    if nanotime() == 0 {
        throw("nanotime returning zero")
    }
    
    // Defer unlock so that runtime.Goexit during init does the unlock too.
    needUnlock := true
    defer func() {
        if needUnlock {
            unlockOSThread()
        }
    }()
    
    // Record when the world started.
    runtimeInitTime = nanotime()
    
    // 开启垃圾回收器
    gcenable()
    
    main_init_done = make(chan bool)
    
    ...
    
    // 执行 main 包的 init 函数
    doInit(&main_inittask)
    
    close(main_init_done)
    
    needUnlock = false
    unlockOSThread()
    
    // 静态库或动态库
    if isarchive || islibrary {
        // A program compiled with -buildmode=c-archive or c-shared
        // has a main, but it is not executed.
        return
    }
    
    // 开始执行 main 函数
    fn := main_main 
    fn()
    if raceenabled {
        racefini()
    }
    
    // 可恢复的 panic
    // 使正常的客户端程序正常工作: 如果在 main 返回的时同时另一个goroutine 产生 panic, 
    // 则让另一个 goroutine 完成对 panic 的打印. 一旦完成, 它将退出.
    if atomic.Load(&runningPanicDefers) != 0 {
        // 最多执行 1000 次调度
        for c := 0; c < 1000; c++ {
            if atomic.Load(&runningPanicDefers) == 0 {
                break
            }
            
            // 切换到 g0, 主动让出 CPU 执行(当前的 gp 的 m 与 p 解绑, 将 gp 放入全局队列, 开启新一轮调度)
            Gosched() 
        }
    }
    
    // 不可恢复的 panic, 直接休眠
    if atomic.Load(&panicking) != 0 {
        gopark(nil, nil, waitReasonPanicWait, traceEvGoStop, 1)
    }
    
    // 系统调用, 退出进程.
    exit(0)
    
    // 保护性代码
    for {
        var x *int32
        *x = 0
    }
}
```

runtime.main 函数工作:

1. 启用 sysmon 系统监控线程, 该线程负责整个程序的 pc, 抢占调度, netpoll 等功能监控

2. 执行 runtime 包的 init 初始化

3. 执行 main 包的 init 初始化

4. 执行 main.main 函数

5. 调用 exit 系统调用退出进程.

`runtime.main()` 在程序初始化(汇编函数 `rt0_go()`) 快结尾时, 通过调用 `runtime.newproc()` 添加到 p (在 `schedinit()` 当中的 `procresize()`
m 与 p 进行了绑定) 当中. 它使用了一些特殊手段, 让 `runtime.main()` 执行完的返回值指向了 `goexit()`

从上述流程来看, runtime.main 在执行完 main.main 函数之后就直接调用 exit 结束进程了, 它并没有返回到调用它函数. 需要注意的是, 
runtime.main 是 main goroutine 的入口函数, 并不是直接被调用的, 而是在 `schedule() -> execute() -> gogo()` 这个调用链
的 gogo 函数中使用汇编代码跳过来的, 从这个角度, goroutine 没有地方可以返回. 

但是, 前面的分析当中得知, 在创建 goroutine 时在其栈上已经放好了一个返回地址, 伪造成 goexit 函数调用了 goroutine 的入口函数, 在这里并没有
使用到这个返回地址, 其实这个地址是为非 main goroutine 准备的, 让其在执行完成之后返回到 goexit 继续执行.


### 非 main goroutine 退出流程

非 main goroutine 的退出流程:

首先来看下 goexit 汇编函数

```cgo
// 在 goroutine 上运行的最顶层函数.
// returns to goexit+PCQuantum.
TEXT runtime·goexit(SB),NOSPLIT,$0-0
    BYTE	$0x90	// NOP
    CALL	runtime·goexit1(SB)	// does not return
    // traceback from goexit1 must hit code range of goexit
    BYTE	$0x90	// NOP
```

非 main goroutine 返回时直接返回到 goexit 的第二条指令: `CALL runtime·goexit1(SB)`, 该指令继续调用 goexit1 函数. 

```cgo
func goexit1() {
	if raceenabled {
		racegoend()
	}
	if trace.enabled {
		traceGoEnd()
	}
	mcall(goexit0)
}
```

goexit1 函数通过调用 mcall 从当前运行的用户 g 切换到 g0, 然后在 g0 栈上调用和执行 goexit0 函数.

```cgo
// func mcall(fn func(*g)), 切换到 m->g0 栈上, 然后调用 fn(g) 函数, fn 函数必须不能返回(因为返回了就会导致程序panic).
// g 对象, 是 m->curg
// mcall 的参数是一个指向 funcval 对象的指针.
TEXT runtime·mcall(SB), NOSPLIT, $0-8
    // 获取参数的值放入 DI 寄存器, 它是 funcval 对象的指针. 当前场景是 goexit0 的地址
    MOVQ	fn+0(FP), DI 
    
    get_tls(CX)
    MOVQ	g(CX), AX	// AX=g, g 是用户 goroutine
    MOVQ	0(SP), BX	// 将 mcall 的返回地址 (rip寄存器的值, 调用 mcall 函数的下一条指令) 放入 BX
    
    // 保存 g 的调度信息.
    MOVQ	BX, (g_sched+gobuf_pc)(AX) // g.sched.pc = BX 
    LEAQ	fn+0(FP), BX               // fn 是调用方的栈顶元素, 其地址就是调用方的栈顶
    MOVQ	BX, (g_sched+gobuf_sp)(AX) // g.sched.sp = BX, 用户 goroutine 的 rsp 
    MOVQ	AX, (g_sched+gobuf_g)(AX)  // g.sched.g = AX
    MOVQ	BP, (g_sched+gobuf_bp)(AX) // g.sched.bp = BP, 用户 goroutine 的 rbp 
    
    // 切换到 g0 栈, 然后调用 fn 
    MOVQ	g(CX), BX    // BX = g 
    MOVQ	g_m(BX), BX  // BX = g.m 
    MOVQ	m_g0(BX), SI // SI = g0 
    
    // 此时, SI=g0, AX=g, 这里需要判断 g 是否是 g0 
    CMPQ	SI, AX	// if g == m->g0 call badmcall
    JNE	3(PC) // 不相等
    MOVQ	$runtime·badmcall(SB), AX
    JMP	AX
    MOVQ	SI, g(CX) // 将本地存储设置为 g0
    MOVQ	(g_sched+gobuf_sp)(SI), SP	// 从 g0.sched.sp 当中恢复 SP, 即 rsp 寄存器, 此时栈已经发生变更
    PUSHQ	AX        // fn 参数 g 入栈
    MOVQ	DI, DX    // DX=fn 
    MOVQ	0(DI), DI // 判断fn不为nil
    CALL	DI        // 调用 fn 函数(已经准备好了栈参数, 因此这里是 CALL), 该函数不会返回, 这里调用的函数是 goexit0 
    POPQ	AX // 正常状况下, 这里及其之后的指令不会执行的
    MOVQ	$runtime·badmcall2(SB), AX
    JMP	AX
    RET
```

mcall 的参数是一个函数, 在 Go 当中, 函数变量并不是一个直接指向函数代码的指针, 而是一个指向 funcval 结构体对象的指针, funcval 
结构体对象的第一个成员 fn 才是真正指向函数代码的指针.

mcall 函数的功能:

1. 首先从当前运行的 g 切换到 g0, 这一步包括保存当前 g 的调度信息(pc,sp,bp,g), 把 g0 设置到 TLS 当中, 修改 rsp 寄存器
使其指向 g0 的栈.

2. 以当前运行的 g 为参数调用 fn 函数(此处是 goexit0). 

> 注: 虽然在 g0 保存有 pc, 但是这里并不会从 pc 处开始调用, 而是直接调用 fn 函数, mcall 的目的是在 g0 栈上执行 fn, 在 fn 当中去进行新一轮的调度.

从 mcall 的功能看, mcall 做的事情与 gogo 函数完全相反, gogo 实现了从 g0 切换到某个 goroutine 去运行, 而 mcall 实现了从某个 goroutine 切换
到 g0 来运行. 因此, mcall 和 gogo 的代码很相似.

mcall 和 gogo 在做切换时有个重要的区别: gogo 函数在从 g0 切换到其他 goroutine 时, 首先切换了栈, 然后通过跳转指令(JMP) 跳转到
用户 goroutine 的代码. mcall 函数在从其他 goroutine 切换回 g0 时只切换了栈, 使用 CALL 指令执行需要函数. 

为何有上述的差别? 原因在于从 g0 切换到其他 goroutine 之前执行的是 runtime 的代码并且使用的是 g0 栈, 因此切换时先切换栈然后
从 runtime 代码跳转某个 goroutine 的代码去执行(切换栈和跳转指令不能颠覆), 然而从某个 goroutine 切换回 g0 时, goroutine 
使用的是 call 指令来调用 mcall 函数, **mcall 本身就是 runtime 的代码, 所以 call 指令其实已经完成从 goroutine 代码切换 
runtime 代码的跳转, 因此 mcall 函数自身无需再跳转了, 只需要把栈切回来即可.**


从 goroutine 切换到 g0 之后, 在 g0 栈上执行 goexit0 函数, 完成最后的清理工作:

- 把 g 的状态设置为 _Gdead

- 把 g 的一些字段清空

- 调用 dropg 解除 g 和 m 的关系, 即 g.m=nil m.curg=nil

- 将 g 放入 p 的 freeg 队列缓存起来供下次创建 g 时快速获取而不是从内存分配.

- 调用 schedule 函数再次进行调度.

```cgo
func goexit0(gp *g) {
	_g_ := getg() // 当前是 g0

	casgstatus(gp, _Grunning, _Gdead) // 状态设置
	if isSystemGoroutine(gp, false) {
		atomic.Xadd(&sched.ngsys, -1)
	}
	
	// gp 的状态清空
	gp.m = nil
	locked := gp.lockedm != 0
	gp.lockedm = 0
	_g_.m.lockedg = 0
	gp.preemptStop = false
	gp.paniconfault = false
	gp._defer = nil // should be true already but just in case.
	gp._panic = nil // non-nil for Goexit during panic. points at stack-allocated data.
	gp.writebuf = nil
	gp.waitreason = 0
	gp.param = nil
	gp.labels = nil
	gp.timer = nil

	if gcBlackenEnabled != 0 && gp.gcAssistBytes > 0 {
		// Flush assist credit to the global pool. This gives
		// better information to pacing if the application is
		// rapidly creating an exiting goroutines.
		scanCredit := int64(gcController.assistWorkPerByte * float64(gp.gcAssistBytes))
		atomic.Xaddint64(&gcController.bgScanCredit, scanCredit)
		gp.gcAssistBytes = 0
	}
    
    // 将 m 和 g 的关系解除. g.m=nil m.curg=nil
	dropg()

	if GOARCH == "wasm" { // no threads yet on wasm
		gfput(_g_.m.p.ptr(), gp) // 将 gp 放入到 p 的 freeg 队列
		schedule() // 再次调度, 将不再返回
	}

	if _g_.m.lockedInt != 0 {
		print("invalid m->lockedInt = ", _g_.m.lockedInt, "\n")
		throw("internal lockOSThread error")
	}
	gfput(_g_.m.p.ptr(), gp) // 将 gp 放入到 p 的 freeg 队列
	if locked {
		// goroutine可能已锁定此线程, 因为它将其置于异常的内核状态. 杀死它, 而不是将其返回到线程池.
        // 返回 mstart, 它将释放P并退出线程.
		if GOOS != "plan9" {
			gogo(&_g_.m.g0.sched) // 这里将跳转到 g0.sched.pc 处执行. 该 pc 是 mstart 函数当中设置的.
		} else {
		    // plan9 系统上重用该线程
			_g_.m.lockedExt = 0
		}
	}
	schedule() // 再次调度, 不再返回.
}
```

到此为止, 一个 goroutine 就运行结束了, 工作线程再次调用 schedule 进入新一轮的调度循环.
