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

### runtime.mainPC(SB) 启动监控任务

### runtime.mstart(SB) 启动调动循环

mstart() 函数是启动 M, 调用的链: `mstart()` -> `mstart1()` -> `schedule()`

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

	// Install signal handlers; after minit so that minit can
	// prepare the thread to be able to handle the signals.
	if _g_.m == &m0 {
		mstartm0()
	}

	if fn := _g_.m.mstartfn; fn != nil {
		fn()
	}

	if _g_.m != &m0 {
		acquirep(_g_.m.nextp.ptr())
		_g_.m.nextp = 0
	}
	schedule()
}
```


### 相关数据结构

````cgo
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
````

