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

```plan9_x86
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
> G0 是什么? G 分为三种, 第一种是用户任务的叫做 G. 第二种是执行 runtime 下调度工作的叫 G0, 每一个 M 都绑定一个 G0.
> 第三种是启动 runtime.main 用到的 G. 程序用到是基本上就是第一种.
