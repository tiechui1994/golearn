# go 调度

内核堆系统线程的调度可以归纳为:**在执行操作系统代码时, 内核调度器按照一定的算法挑选处一个线程并把该线程保存在内存之中的寄
存器值放入CPU对于的寄存器从而恢复该线程的运行.**

万变不离其宗, 系统线程对 goroutine 的调度与内核对系统线程的调度原理是一样的, `实质都是通过保存和修改CPU寄存器的值来达
到切换线程/gorotine的目的.`

因此, 为了实现对 gorotine 的调度, 需要引入一个数据结构来保存CPU寄存器的值以及goroutine的其他一些状态信息, 这个数据结
构就是 g, 它保存了 goroutine 的所有信息, 该结构体的每一个实例对象都代表了一个 goroutine, 调度器代码可以通过 g 对象来
对 goroutine 进行调度, 当 goroutine 被调离 CPU 时, 调度器代码负责把 CPU 寄存器的值保存在 g 对象的成员变量当中, 当
goroutine 被调度起来运行时, 调度器代码负责把 g 对象的成员变量所保存的寄存器的值恢复到 CPU 寄存器.

任何一个由编译型语言所编写的程序在被操作系统加载起来运行时都会顺序经过如下几个阶段:

1. 从磁盘上把可执行程序读入内存;

2. 创建进程和主线程; 

3. 为主线程分配栈空间;

4. 把由用户在命令行输入的参数拷贝到主线程的栈;

5. 把主线程放入操作系统的运行队列等待被调度执行起来运行.


G, M, P

- G: 代表一个 goroutine, 每个 goroutine 都有自己独立的栈存放当前运行内存及状态. 可以认为一个 G 就是一个任务.

- M: 代表内核线程(pthread), 它本身与一个 `内核线程` 进行绑定, goroutine 运行在M上.

- P: 代表一个资源, 可以认为一个 "有运行任务" 的P占了一个 `CPU线程` 的资源, 且只要处于调度的时候就有P.

> 注: 内核线程与CPU线程的区别, 在系统里可以有上万个内核线程, 但 CPU 线程并没有那么多, CPU线程也就是 top 命令里看到的
%Cpu0, %Cpu1, ...的数量. 


### 进程启动都做了什么?

通过 gdb 的 `info files` 可以查找到go编译文件的函数入口地址, 通过单步调试 `si`, 最终到 `runtime.rt0_go` 这个汇编
函数当中开始进行调度器的初始化.

```cgo
// runtime/go_tls.h

// 宏定义函数
#ifdef GOARCH_amd64
#define	get_tls(r)	MOVQ TLS, r    // 获取 TLS 的位置
#define	g(r)	0(r)(TLS*1)        // 获取 TLS 当中存储 g 的位置 
#endif
```

// runtime/asm_amd64.s

```cgo
TEXT runtime·rt0_go(SB),NOSPLIT,$0
    ....
    
    # 给 g0 分配栈空间, 大约是64k
    MOVQ	$runtime·g0(SB), DI  # DI=g0 
    LEAQ	(-64*1024+104)(SP), BX # 预分配大约64k的栈
    MOVQ	BX, g_stackguard0(DI) # g0.stackguard0=BX
    MOVQ	BX, g_stackguard1(DI) # g0.stackguard1=BX
    MOVQ	BX, (g_stack+stack_lo)(DI) # g0.stack.lo=BX
    MOVQ	SP, (g_stack+stack_hi)(DI) # g0.stack.hi=SP
    
   ....
   
    # 开始初始化tls, settls本质上是通过系统调用 arch_prctl 实现的.
    LEAQ	runtime·m0+m_tls(SB), DI # DI=&m0.tls
   	CALL	runtime·settls(SB) # 调用 settls 设置本地存储, settls 的参数在 DI 当中
   	
   	# 测试settls 是否可以正常工作
    get_tls(BX) # 获取fs段寄存器地址并放入 BX 寄存器, 其实就是 m0.tls[1] 的地址
    MOVQ	$0x123, g(BX) # 将 0x123 拷贝到 fs 段寄存器的地址偏移-8的内存位置, 也就是 m0.tls[0]=0x123 
    MOVQ	runtime·m0+m_tls(SB), AX # AX=m0.tls[0]
    CMPQ	AX, $0x123 # 检查 m0.tls[0] 是否是通过本地存储存入的 0x123 来验证功能是否正常
    JEQ 2(PC) # 相等
    CALL	runtime·abort(SB) # 线程本地存储功能不正常, 退出程序
ok:
	// set the per-goroutine and per-mach "registers"
	get_tls(BX) // 获取fs段基地址BX寄存器
	LEAQ	runtime·g0(SB), CX // CX=&g0
	MOVQ	CX, g(BX)  // 把g0的地址保存到线程本地存储,即 m0.tls[0]=&g0
	LEAQ	runtime·m0(SB), AX // AX=&m0
    
    // 全局 m0 与 g0 进行关联
	MOVQ	CX, m_g0(AX) # m0.g0=&g0 
	MOVQ	AX, g_m(CX) # g0.m=&m0
    
    # 到此位置, m0与g0绑定在一起, 之后主线程可以通过 get_tls 获取到 g0, 通过 g0 又获取到 m0 
    # 这样就实现了 m0, g0 与主线程直接的关联. 
        
	CLD				// convention is D is always left cleared
	CALL	runtime·check(SB)
    
    # 命令行参数拷贝
	MOVL	16(SP), AX		// copy argc
	MOVL	AX, 0(SP)
	MOVQ	24(SP), AX		// copy argv
	MOVQ	AX, 8(SP) 
	CALL	runtime·args(SB)   // 解析命令行
	
	// 对于 linux, osinit 唯一功能就是获取CPU核数并放到变量 ncpu 中
	CALL	runtime·osinit(SB) 
	CALL	runtime·schedinit(SB) // 调度系统初始化
```

> M0 是什么? 程序会启动多个 M, 第一个启动的是 M0
>
> G0 是什么? G 分为三种, 第一种是用户任务的叫做 G. 第二种是执行 runtime 下调度工作的叫 G0, 每一个 M 都绑定一个 
> G0. 第三种是启动 runtime.main 用到的 G. 程序用到是基本上就是第一种.


### runtime.osinit(SB) 针对系统环境的初始化

唯一的作用的就是初始化全局变量 ncpu

### runtime.schedinit(SB) 调度初始化

schedinit 做的重要事情:

- 初始化 m0, mcommoninit函数

- 创建和初始化 allp, procresize 函数, 并将 m0 绑定到 `allp[0]` 即 `m0.p = allp[0], allp[0].m = m0` 

```cgo
func schedinit() {
    // getg 最终是插入代码, 格式如下:
    // gettls(CX)
    // MOVQ g(CX), BX; BX 当中就是当前 g 的结构体对象的地址
	_g_ := getg() // _g_ = &g0
    
    // 设置最多启动 10000 个操作系统线程, 即最多 10000 个M
	sched.maxmcount = 10000

	tracebackinit()
	moduledataverify()
	stackinit()
	mallocinit()
	mcommoninit(_g_.m) // 初始化 m0, 因为 g0.m = &m0
	
	......  

	msigsave(_g_.m) // 初始化 m0.gsignal
	initSigmask = _g_.m.sigmask

    ...... 
    
	sched.lastpoll = uint64(nanotime())
	procs := ncpu // 系统有多少个核, 就创建多少个 p 对象
	if n, ok := atoi32(gogetenv("GOMAXPROCS")); ok && n > 0 {
		procs = n // 通过修改环境变量 GOMAXPROCS, 指定创建p的数量 
	}
	
	// 创建和初始化全局变量 allp
	if procresize(procs) != nil {
		throw("unknown runnable goroutine during bootstrap")
	}
    
    ....	
}
```

前面汇编当中, g0的地址已经被设置到了本地存储之中, schedinit通过 getg 函数(getg函数是编译器实现的, 源码当中找不到其定
义的)从本地存储中获取当前正在运行的 g, 这里获取的是 g0, 然后调用 mcommoninit() 函数对 m0 进行必要的初始化, 对 m0 
初始化完成之后, 调用 procresize() 初始化系统需要用到的 p 结构体对象. 它的数量决定了最多同时有多少个 goroutine 同时
并行运行. 

sched.maxmcount 设置了 M 最大数量, 而 M 代表系统内核线程, 因此一个进程最大只能启动10000个系统内核线程. 

procresize 初始化P的数量, procs 参数为初始化的数量, 而在初始化之前先做数量的判断, 默认是ncpu(CPU核数). 也可以通过
GOMAXPROCS来控制P的数量. _MaxGomaxprocs 控制了最大数量只能是1024.

重点关注一下 mcommoninit() 函数如何初始化 m0 以及 procresize() 函数如何创建和初始化 p 结构体对象.

````cgo
func mcommoninit(mp *m, id int64) {
	_g_ := getg() // 初始化过程中 _g_ = &g0

	// g0 stack won't make sense for user (and is not necessary unwindable).
	// 函数调用栈 traceback
	if _g_ != _g_.m.g0 {
		callers(1, mp.createstack[:])
	}

	lock(&sched.lock)
    
    // 初始化过程中, 这里 id 是 -1, 生成 m 的 id
	if id >= 0 {
		mp.id = id
	} else {
		mp.id = mReserveID()
	}

	mp.fastrand[0] = uint32(int64Hash(uint64(mp.id), fastrandseed))
	mp.fastrand[1] = uint32(int64Hash(uint64(cputicks()), ^fastrandseed))
	if mp.fastrand[0]|mp.fastrand[1] == 0 {
		mp.fastrand[1] = 1
	}
    
    // 创建用于信号处理的 gsignal, 只是简单从堆上分配一个 g 对象, 绑定到 mp.
	mpreinit(mp)
	if mp.gsignal != nil {
		mp.gsignal.stackguard1 = mp.gsignal.stack.lo + _StackGuard
	}

    // 将 m 挂入全局链表 allm 之中, 在这里 allm[0]=&m0
	// Add to allm so garbage collector doesn't free g->m
	// when it is just in a register or thread-local storage.
	mp.alllink = allm

	// NumCgoCall() iterates over allm w/o schedlock,
	// so we need to publish it safely.
	atomicstorep(unsafe.Pointer(&allm), unsafe.Pointer(mp))
	unlock(&sched.lock)

	// Allocate memory to hold a cgo traceback if the cgo call crashes.
	if iscgo || GOOS == "solaris" || GOOS == "illumos" || GOOS == "windows" {
		mp.cgoCallers = new(cgoCallers)
	}
}
````

mcommoninit 函数重点就是初始化了 m0 的 `id, fastrand, gsignal ...` 等变量,  然后 m0 放入到全局链表 allm 之中, 
然后就返回了.


```cgo
func procresize(nprocs int32) *p {
	old := gomaxprocs // 系统初始化时, gomaxprocs = 0
	if old < 0 || nprocs <= 0 {
		throw("procresize: invalid arg")
	}
	
	...... 

    // 初始化时 len(allp) == 0
	// Grow allp if necessary.
	if nprocs > int32(len(allp)) {
		// Synchronize with retake, which could be running
		// concurrently since it doesn't run on a P.
		lock(&allpLock)
		if nprocs <= int32(cap(allp)) {
			allp = allp[:nprocs]
		} else {
		    // 初始化, 进入次分支, 创建 allp 切片
			nallp := make([]*p, nprocs)
			// Copy everything up to allp's cap so we
			// never lose old allocated Ps.
			copy(nallp, allp[:cap(allp)])
			allp = nallp
		}
		unlock(&allpLock)
	}

	// 初始化 nprocs 个 p
	for i := old; i < nprocs; i++ {
		pp := allp[i]
		if pp == nil {
			pp = new(p) // 直接从堆上分配一个 p
		}
		pp.init(i) // 初始化(status,deferpool,mcache等)
		atomicstorep(unsafe.Pointer(&allp[i]), unsafe.Pointer(pp)) // 存储到allp对应的位置
	}

	_g_ := getg() // 当前处于初始化状况下 _g_ = g0, 并且此时 m0.p 还未初始化, 因此在这里会初始化 m0.p
	if _g_.m.p != 0 && _g_.m.p.ptr().id < nprocs {
		_g_.m.p.ptr().status = _Prunning
		_g_.m.p.ptr().mcache.prepareForSweep()
	} else {
		// release the current P and acquire allp[0].
		//
		// We must do this before destroying our current P
		// because p.destroy itself has write barriers, so we
		// need to do that from a valid P.
		if _g_.m.p != 0 {
			if trace.enabled {
				// Pretend that we were descheduled
				// and then scheduled again to keep
				// the trace sane.
				traceGoSched()
				traceProcStop(_g_.m.p.ptr())
			}
			_g_.m.p.ptr().m = 0
		}
		_g_.m.p = 0 // 初始化时的 m0.p 
		p := allp[0] // 第一个 p 
		p.m = 0
		p.status = _Pidle
		acquirep(p) // 初始化, 将 allp[0] 和 m0 关联起来, 并且 allp[0] 的状态为  _Prunning
		if trace.enabled {
			traceGoStart()
		}
	}

	// g.m.p is now set, so we no longer need mcache0 for bootstrapping.
	mcache0 = nil

	// 释放未使用的 p, 此时还不能释放 p(因为处于系统调用中的 m 可能引用 p, 即 m.p = p)
	// 初始化时 old 是 0
	for i := nprocs; i < old; i++ {
		p := allp[i]
		p.destroy()
	}

	// 再次确定切片大小是 nprocs
	if int32(len(allp)) != nprocs {
		lock(&allpLock)
		allp = allp[:nprocs]
		unlock(&allpLock)
	}
    
    // 把所有空闲的p放入空闲链表
	var runnablePs *p
	for i := nprocs - 1; i >= 0; i-- {
		p := allp[i]
		if _g_.m.p.ptr() == p { // 当前 m 关联的 p 
			continue
		}
		// 状态修改
		p.status = _Pidle
		
		// 判断当 p 的本地队列是否为空 即 runqhead == runtail && runnext=0
		if runqempty(p) { 
			pidleput(p)
		} else {
		    // 给 p 绑定一个 m, 并且把这些非空闲的 p 组成一个单向链表
		    // 最终链表的头是 runnablePs
			p.m.set(mget())
			p.link.set(runnablePs)
			runnablePs = p
		}
	}
	stealOrder.reset(uint32(nprocs))
	var int32p *int32 = &gomaxprocs // make compiler check that gomaxprocs is an int32
	atomic.Store((*uint32)(unsafe.Pointer(int32p)), uint32(nprocs))
	return runnablePs
}
```

procresize 函数主要干的事情:

- 使用 `make([]*p nprocs)` 初始化全局变量 allp. 

- 使用循环初始化 nprocs 个 p 结构体对象并依次保存在 allp 切片之中.

- 把 m0 和 `allp[0]` 绑定在一起, 即 `m0.p = allp[0], allp[0].m = m0` (比较关键)

- 把除了 `allp[0]` 之外的所有 p 放入全局变量 sched 的 pidle 空闲队列之中. 对于非空闲的 p 组装成一个链表, 并返回.

### runtime.mainPC(SB) main goroutine

在进行 schedinit 完成调度系统初始化后, `m0 与 g0`, `m0 与 allp[0]` 已经相互关联了. 接下来的操作就行需要创建 main
goroutine 用于执行 runtime.main 函数. 

调用 `newproc()` 创建一个新的 goroutine 用于执行 mainPC 所对应的 runtime.main 函数. 

// runtime/asm_amd64.s

```cgo
TEXT runtime·rt0_go(SB),NOSPLIT,$0
    ..... 
    
    # 创建 main goroutine, 也是系统的第一个 goroutine 
    // create a new goroutine to start program
    MOVQ	$runtime·mainPC(SB), AX # mainPC 是 runtime.main 
    PUSHQ	AX # AX=&funcval{runtime.main}, newproc 的第二个参数(新的goroutine需要执行的函数)
    PUSHQ	$0 # newproc 的第一个参数, 该参数表示 runtime.main 函数需要的参数大小, 因为没有参数, 所以这里是0
    CALL	runtime·newproc(SB) // 创建 main goutine
    POPQ	AX
    POPQ	AX
    
    # 主线程进入调度循环，运行刚刚创建的goroutine
    CALL  runtime·mstart(SB)  #
    
    # 上面的mstart永远不应该返回的, 如果返回了, 一定是代码逻辑有问题, 直接abort
    CALL  runtime·abort(SB)
    RET


# runtime·mainPC 的定义.
DATA  runtime·mainPC+0(SB)/8,$runtime·main(SB)
GLOBL runtime·mainPC(SB),RODATA,$8
```

先分析下 newproc 函数:

newproc 用于创建新的 goroutine, 有两个参数, 第二个参数 fn, 新创建出来的 goroutine 将从 fn 这个函数开始执行, 而这
个 fn 函数可能会有参数, newproc 的第一个参数是 fn 函数的参数以字节为单位的大小. 比如如下 go 片段代码:

```cgo
func sum(a, b, c int64) {
}

func main() {
    go sum(1,2,3)
}
```

编译器在编译上面的代码时, 会把其替换为对 newproc 函数的调用, 编译后的代码逻辑上等同于下面的汇编代码:

```cgo
0x001d 00029 (sum.go:7) MOVL    $24, (SP) # newproc第一个参数
0x0024 00036 (sum.go:7) LEAQ    "".sum·f(SB), AX 
0x002b 00043 (sum.go:7) MOVQ    AX, 8(SP)  # newproc第二个参数
0x0030 00048 (sum.go:7) MOVQ    $1, 16(SP) # 函数参数1
0x0039 00057 (sum.go:7) MOVQ    $2, 24(SP) # 函数参数2 
0x0042 00066 (sum.go:7) MOVQ    $3, 32(SP) # 函数参数3
0x004b 00075 (sum.go:7) CALL    runtime.newproc(SB)
```

编译器编译时首先把 sum 函数需要用到的 3 个参数压入栈, 然后调用 newproc 函数. 因为 sum 函数的 3 个 int64 类型的参数
公占用 24 个字节, 所以传递给 newproc 的第一个参数是 24, 表示 sum 函数的大小.

为什么需要传递 fn 函数的参数大小给 newproc 函数? 原因在于 newproc 函数创建一个新的 goroutine 来执行 fn 函数, 而这
个新创建的 goroutine 与当前 goroutine 使用不同的栈, 因此需要在创建 goroutine 的时候把 fn 需要用到的参数从当前栈上
拷贝到新 goroutine 的栈上之后才能开始执行, 而 newproc 函数本身并不知道需要拷贝多少数据到新创建的 goroutine 的栈上去,
所以需要用参数的方式指定拷贝数据的大小.


newproc 函数是对 newproc1 的一个包装, 最重要的工作:

- 获取 fn 函数的第一个参数的地址(argp)

- 使用 systemstack 函数切换到 g0 栈(对于初始化场景来说现在本身就在 g0 栈, 不需要切换, 但是对于用户创建的 goroutine
则需要进行栈切换)

```cgo
func newproc(siz int32, fn *funcval) {
	// 函数调用参数入栈是从右往左, 而且栈是从高地址向低地址增长的
	// 注: argp 指向 fn 函数的第一个参数
	// 参数 fn 在栈上的地址 + 8 = fn 函数的第一个参数. (参考上面的汇编代码)
	argp := add(unsafe.Pointer(&fn), sys.PtrSize)
	gp := getg() // 获取当前运行的 g, 初始化是 m0.g0
	
	// getcallerpc() 返回一个地址, 也就是调用 newproc 时 call 指令压栈的函数返回地址,
	// 对于当前场景来说, pc 就是 'CALL	runtime·newproc(SB)' 后面的 'POPQ AX' 这条指令地址
	pc := getcallerpc() 
	
	// 切换到 g0 执行作为参数的函数
	systemstack(func() {
		newg := newproc1(fn, argp, siz, gp, pc)
        
        // 获取当前的 g0 绑定的 _p_, 然后将新创建的 newg 放入到 _p_ 本地队列当中.
        // 注: newg 当前还没有和任何 m 进行关联, 只有被调度运行的时才和 m 进行关联
		_p_ := getg().m.p.ptr() 
		runqput(_p_, newg, true)

		if mainStarted {
			wakep()
		}
	})
}
```

newproc1 函数的第一个参数 fn 是新创建的 goroutine 需要执行的函数, 注意 fn 的类型是 funcval; 第二个参数 argp 是 
fn 函数的第一个参数的地址; 第三个参数是 fn 函数以字节为单位的大小. 第四个参数是当前运行的 g; 第五个参数调用 newproc
的函数的返回地址.

```cgo
func newproc1(fn *funcval, argp unsafe.Pointer, narg int32, callergp *g, callerpc uintptr) *g {
	// 当前已经切换到 g0 栈, 因此无论什么状况下, _g_ = g0 (工作线程的 g0)
	// 对于当前的场景, 这里的 g0 = m0.g0
	_g_ := getg() 

	if fn == nil {
		_g_.m.throwing = -1 // do not dump full stacks
		throw("go of nil func value")
	}
	
	// 禁用抢占, 因为它可以将 p 保留在本地变量中
	acquirem() 
	siz := narg
	siz = (siz + 7) &^ 7 // size 进行 8 字节对齐

	// We could allocate a larger initial stack if necessary.
	// Not worth it: this is almost always an error.
	// 4*sizeof(uintreg): extra space added below
	// sizeof(uintreg): caller's LR (arm) or return address (x86, in gostartcall).
	if siz >= _StackMin-4*sys.RegSize-sys.RegSize {
		throw("newproc: function arguments too large for new goroutine")
	}
    
    // 初始化时, 这里的 _p_ 其实就是 allp[0]
	_p_ := _g_.m.p.ptr() 
	
	// 从 _p_ 本地缓存中获取一个 g, 初始化时没有, 返回 nil 
	newg := gfget(_p_)
	if newg == nil {
	    // new一个g, 然后从堆上为其分配栈, 并设置 g 的 stack 成员和两个 stackguard 成员
		newg = malg(_StackMin)
		casgstatus(newg, _Gidle, _Gdead)
		
		// 放入全局 allgs 切片当中
		allgadd(newg) // publishes with a g->status of Gdead so GC scanner doesn't look at uninitialized stack.
	}
	if newg.stack.hi == 0 {
		throw("newproc1: newg missing stack")
	}

	if readgstatus(newg) != _Gdead {
		throw("newproc1: new g is not Gdead")
	}
    
    // 调整 g 的栈顶指针.
    // sys.MinFrameSize 是 0
    // sys.SpAlign 是 1
    // totalSize 最终大小是 siz+32
	totalSize := 4*sys.RegSize + uintptr(siz) + sys.MinFrameSize // extra space in case of reads slightly beyond frame
	totalSize += -totalSize & (sys.SpAlign - 1)                  // align to spAlign
	sp := newg.stack.hi - totalSize
	spArg := sp
	if usesLR {
		// caller's LR
		*(*uintptr)(unsafe.Pointer(sp)) = 0
		prepGoExitFrame(sp)
		spArg += sys.MinFrameSize
	}
	if narg > 0 {
	    // 把参数从 newproc 函数的栈(初始化是g0栈)拷贝到新的 g 的栈
	    // 注: 这里是从 sp 的位置开始拷贝的.
		memmove(unsafe.Pointer(spArg), argp, uintptr(narg))
		// This is a stack-to-stack copy. If write barriers
		// are enabled and the source stack is grey (the
		// destination is always black), then perform a
		// barrier copy. We do this *after* the memmove
		// because the destination stack may have garbage on
		// it.
		if writeBarrier.needed && !_g_.m.curg.gcscandone {
			f := findfunc(fn.fn)
			stkmap := (*stackmap)(funcdata(f, _FUNCDATA_ArgsPointerMaps))
			if stkmap.nbit > 0 {
				// We're in the prologue, so it's always stack map index 0.
				bv := stackmapdata(stkmap, 0)
				bulkBarrierBitmap(spArg, spArg, uintptr(bv.n)*sys.PtrSize, 0, bv.bytedata)
			}
		}
	}
    
    // 清空 newg.sched 里面的字段, 然后重新设置 sched 对应的值
	memclrNoHeapPointers(unsafe.Pointer(&newg.sched), unsafe.Sizeof(newg.sched))
	newg.sched.sp = sp 
	newg.stktopsp = sp 
	
	// newg.sched.pc 表示当 newg 被调度起来运行时从这个地址开始执行指令.
	// 把 pc 设置成 goexit 函数偏移 1 (sys.PCQuantum是1) 的位置.
	// 为啥这样做, 暂时不清楚
	newg.sched.pc = funcPC(goexit) + sys.PCQuantum 
	newg.sched.g = guintptr(unsafe.Pointer(newg))
	
	// 调整 sched成员和 newg 的栈(参考下面的分析)
	gostartcallfn(&newg.sched, fn)
	
	// traceback
	newg.gopc = callerpc
	newg.ancestors = saveAncestors(callergp)
	
	// 设置 newg 的 startpc 为 fn.fn, 该成员主要用于函数调用栈的 traceback 和栈收缩
	// newg 真正从哪里执行不依赖次成员, 而是 sched.pc
	newg.startpc = fn.fn
	if _g_.m.curg != nil {
		newg.labels = _g_.m.curg.labels
	}
	if isSystemGoroutine(newg, false) {
		atomic.Xadd(&sched.ngsys, +1)
	}
	// newg 的状态为 _Grunnable, 表示 g 可以进行运行了
	// 注: 前面获取 newg 的时候, newg 添加到 allg 当中, 但是当前的 newg 并没有关联到某个 p
	casgstatus(newg, _Gdead, _Grunnable)
    
    // 下面主要进行 goid 设置和 goid 缓存(缓存在 _p_ 当中)
	if _p_.goidcache == _p_.goidcacheend {
		// Sched.goidgen is the last allocated id,
		// this batch must be [sched.goidgen+1, sched.goidgen+GoidCacheBatch].
		// At startup sched.goidgen=0, so main goroutine receives goid=1.
		_p_.goidcache = atomic.Xadd64(&sched.goidgen, _GoidCacheBatch)
		_p_.goidcache -= _GoidCacheBatch - 1
		_p_.goidcacheend = _p_.goidcache + _GoidCacheBatch
	}
	newg.goid = int64(_p_.goidcache)
	_p_.goidcache++
	if raceenabled {
		newg.racectx = racegostart(callerpc)
	}
	if trace.enabled {
		traceGoCreate(newg, newg.startpc)
	}
	releasem(_g_.m)

	return newg
}
```

gostartcallfn 函数, 在 newg.sched.pc 设置为 `funcPC(goexit) + 1` 之后进行调用, 在这个函数当中可以找到为何这样做
的原因.

```cgo
// fn 是 goroutine 的入口地址, 在初始化的时候对应是是 runtime.main
func gostartcallfn(gobuf *gobuf, fv *funcval) {
	var fn unsafe.Pointer
	if fv != nil {
		fn = unsafe.Pointer(fv.fn)
	} else {
		fn = unsafe.Pointer(funcPC(nilfunc))
	}
	gostartcall(gobuf, fn, unsafe.Pointer(fv))
}


func gostartcall(buf *gobuf, fn, ctxt unsafe.Pointer) {
	sp := buf.sp // newg 的栈顶, 目前 newg 栈上只有 fn 函数的参数, sp 指向的是 fn 的第一个参数
	if sys.RegSize > sys.PtrSize {
		sp -= sys.PtrSize
		*(*uintptr)(unsafe.Pointer(sp)) = 0
	}
	sp -= sys.PtrSize // 为返回地址预留空间, 然后将返回地址写入当前预留的位置
	*(*uintptr)(unsafe.Pointer(sp)) = buf.pc
	buf.sp = sp // 重新设置 newg 的栈顶寄存器
	
	// 这里才正在让 newg 的 ip 寄存器指向 fn 函数. 这里只是设置 newg 的信息, newg 还未执行,
	// 等到 newg 被调度起来之后, 调度器会把 buf.pc 放入到 CPU 的 ip 寄存器, 从而使得 cpu 真正执行起来.
	buf.pc = uintptr(fn)
	buf.ctxt = ctxt // fv 地址
}
```

gostartcallfn 函数的主要作用:

- 获取 fn 函数地址

- 调用 gostartcall 函数, 在函数当中调整 newg 的 sp (将 goexit 函数的第二条指令地址入栈, 伪造成 goexit 函数调用了 fn,
从而使 fn 执行完成后 ret 指令返回到 goexit 继续执行完成清理工作) 和 pc (重新设置 pc 为需要执行的函数的地址, 即fn, 当前
场景为 runtime.main 函数地址) 的值.


在调用 newproc 函数之后, runtime.main 对应的 newg 已经创建完成, 并且加入到当前线程绑定的 p 当中, 等待被调度执行.

- newg 的 pc 指向的是 runtime.main 函数的第一条指令, sp 成员指向是的 newg 栈顶单元, 该单元保存了 runtime.main 函数
执行完成之后的返回地址( runtime.goexit 的第二条指令, 预期 runtime.main 执行完成之后就会执行 `CALL runtime.goexit1(SB)`
这条指令)

- newg 放入当前线程绑定的 p 对象的本地运行队列, 它是第一个真正意义上的用户 goroutine

- **newg 的 m 成员是 nil, 因为它还没有被调度起来运行, 没有任何 m 进行绑定.**

### runtime.mstart(SB) 调度循环开始

到目前为止, runtime.main goroutine已经创建, 并放到了 m0 线程绑定的 `allp[0]` 的本地队列当中, 接下来就需要启动调度循
环, 开始去查找 goroutine 并执行. 

汇编代码从 `newproc` 返回之后, 开始执行 mstart 函数.

mstart() 启动调度循环, 调用链: `mstart()` -> `mstart1()` -> `schedule()`

```cgo
func mstart() {
	_g_ := getg() // _g_ = g0
    
    // 对于启动过程, g0 的 stack 早完成初始化, osStack = false
	osStack := _g_.stack.lo == 0
	if osStack {
		// Initialize stack bounds from system stack.
		// Cgo may have left stack size in stack.hi.
		// minit may update the stack bounds.
		size := _g_.stack.hi
		if size == 0 {
			size = 8192 * sys.StackGuardMultiplier
		}
		_g_.stack.hi = uintptr(noescape(unsafe.Pointer(&size)))
		_g_.stack.lo = _g_.stack.hi - size + 1024
	}
	// Initialize stack guard so that we can start calling regular
	// Go code.
	_g_.stackguard0 = _g_.stack.lo + _StackGuard
	// This is the g0, so we can also call go:systemstack
	// functions, which check stackguard1.
	_g_.stackguard1 = _g_.stackguard0
	
	// 执行 mstart1
	mstart1()

	// Exit this thread.
	switch GOOS {
	case "windows", "solaris", "illumos", "plan9", "darwin", "aix":
		// Windows, Solaris, illumos, Darwin, AIX and Plan 9 always system-allocate
		// the stack, but put it in _g_.stack before mstart,
		// so the logic above hasn't set osStack yet.
		osStack = true
	}
	mexit(osStack)
}
``` 


```cgo
func mstart1() {
	_g_ := getg() // 启动过程为 g0, 当前是运行在 g0 栈上的

	if _g_ != _g_.m.g0 {
		throw("bad runtime·mstart")
	}

	// getcallerpc() 获取调用 mstart1 执行完的返回地址
	// getcallersp() 获取调用 mstart1 时的栈顶地址
	save(getcallerpc(), getcallersp())
	asminit() // AMD64 Linux 是空函数
	minit() // 信号相关初始化
    
    // 启动时, _g_.m 是 &m0, 因此会执行下面的 mstartm0 函数
	if _g_.m == &m0 {
		mstartm0() // 信号初始化
	}

    // 在这里 fn 为 nil 
	if fn := _g_.m.mstartfn; fn != nil {
		fn()
	}
    
    // 在这里, 将 nextp 与当前的 m 进行绑定
	if _g_.m != &m0 {
		acquirep(_g_.m.nextp.ptr())
		_g_.m.nextp = 0
	}
	
	// 执行调度函数
	schedule()
}
```

mstart1 首先调用 save() 函数来保存 g0的调用信息, **save这一行代码非常关键, 是理解调度循环的关键点之一**. 注意这里
getcallerpc() 返回的的 mstart 调用 mstart1 时被 call 指令压栈的返回地址, getcallersp() 返回的是调用 mstart1
函数之前 mstart 函数的栈顶地址.

```cgo
//go:nosplit
//go:nowritebarrierrec
func save(pc, sp uintptr) {
	_g_ := getg()

	_g_.sched.pc = pc // 再次运行时的指令地址
	_g_.sched.sp = sp // 再次运行时的栈顶
	_g_.sched.lr = 0
	_g_.sched.ret = 0
	_g_.sched.g = guintptr(unsafe.Pointer(_g_)) // 保存当前的 _g_
	
	// 需要确保ctxt为零, 但此处不能有写障碍. 但是, 它应该始终已经为零.
	// 断言.
	if _g_.sched.ctxt != nil {
		badctxt()
	}
}
```

save 函数保存了调度相关的所有信息, 包括最为重要的当前正在运行 g 的下一条指令地址和栈顶地址. 不管想 g0 还是其他的 g 来
说这些信息在调度过程中都是必不可少的.

到目前为止, 上述的 mstart() 是在 g0 上执行的, 所有的操作都是针对 g0 而言. 

为何 g0 已经执行到 mstart1 这个函数而且还会继续调用其他函数, 但 g0 的调度信息中的 pc 和 sp 却要设置在 mstart 函数
中? 难道下次切换到 g0 时需要从 `switch` 语句继续执行? 从 mstart 函数可以看到, switch 语句之后就要退出线程了!

save 函数执行之后, 返回到 mstart1 继续其他跟 m 相关的一些初始化, 完成这些初始化后则调用调度系统的核心函数 schedule()
完成 gorotine 的调度. 每次调度 goroutine 都是从 schedule 函数开始的.


```cgo
func schedule() {
	_g_ := getg() // _g_ 是工作线程 m 对于的 g0, 在初始化时是 m0.g0

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
	    // 保证调度的公平性, 每进行 61 次调度需要优先从全局运行队列中获取 gorotine
		if _g_.m.p.ptr().schedtick%61 == 0 && sched.runqsize > 0 {
			lock(&sched.lock)
			gp = globrunqget(_g_.m.p.ptr(), 1)
			unlock(&sched.lock)
		}
	}
	
	if gp == nil {
	    // 从与 m 关联的 p 的本地运行队列中获取 goroutine
		gp, inheritTime = runqget(_g_.m.p.ptr())
	}
	if gp == nil {
	    // 当本地队列和全局队列都没有找到要运行的 goroutine. 调用 findrunable 函数总其他工作线程
	    // 的运行队列中偷取, 如果偷取不到, 则当前工作线程进入休眠, 直到获取到需要运行的 goroutine
	    // 之后函数才返回.
		gp, inheritTime = findrunnable() // blocks until work is available
	}

	// 此时线程将要开始执行 gp, 因此对于自旋状况必须重置
	if _g_.m.spinning {
		resetspinning()
	}
    
	if sched.disable.user && !schedEnabled(gp) {
	    // 当前的 gp 被禁用调用时, 需要重新启用用户调度并再次查看, 
	    // 否则, 将其放在待处理的可运行goroutine列表中.
	    // 一般只有系统的 goroutine 可以被禁用调度
		lock(&sched.lock)
		if schedEnabled(gp) {
			unlock(&sched.lock)
		} else {
			sched.disable.runnable.pushBack(gp)
			sched.disable.n++
			unlock(&sched.lock)
			goto top
		}
	}

	// 当 goroutine 不是一般的 goroutine 时 (a GCworker or tracereader),
	// 唤醒一个 P(可能会开启一个线程, sched.npidle>0 且 sched.nmspinning=0 状况下)
	if tryWakeP {
		wakep()
	}
	
	// gp锁定在某个 m 上, 则需要重新进行查询 gp 
	if gp.lockedm != 0 {
		// Hands off own p to the locked m,
		// then blocks waiting for a new p.
		startlockedm(gp)
		goto top
	}
    
    // 当前运行的是 runtime 代码, 函数栈使用的是 g0 的栈空间
    // 调用 execute 切换到 gp 代码和栈空间去运行.
	execute(gp, inheritTime)
}
```

schedule 函数通过 globalrunqget() 和 runqget() 函数分别从全局队列和当前工作线程的本地运行队列获取要下一个要运行的
goroutine, 如果这两个队列都没有需要运行的 goroutine, 则通过 findrunnable() 函数从其他的 p 的运行队列当中偷取
goroutine, 一旦找到一个, 则调用 execute 函数从 g0 切换到该 goroutine 去运作. 


```cgo
// 这里的 g 是要即将运行的 goroutine, 一般是用户代码
func execute(gp *g, inheritTime bool) {
	_g_ := getg() // 当前是 g0

	// 在进入 _Grunning 之前, 将 m 和 g 进行关联
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

    // 完成 g0 到 gp 正在的切换.
	gogo(&gp.sched)
}
```

execute 的第一个参数 gp 是需要调度起来的 goroutine. 这里首先进行了 m 和 gp 的绑定, 接着就是 gp 状态修改为 _Grunning.
完成 gp 运行前的准备工作之后, gogo 函数完成 g0 到 gp 的切换: **CPU执行权的转让和栈的切换**

gogo 函数是使用汇编实现的, 之所以使用汇编, 是因为 goroutine 的调度涉及不同执行流之间的切换, 执行流的切换本质上就是
CPU寄存器以及函数栈帧的切换.

```cgo
// func gogo(buf *gobuf)
// 从 gobuf 当中恢复 state, longjmp
TEXT runtime·gogo(SB), NOSPLIT, $16-8
	MOVQ	buf+0(FP), BX   // buf=&gp.sched
	MOVQ	gobuf_g(BX), DX // DX=gp.sched.g
	
	# 检查 gp.sched.g 不为 nil 
	MOVQ	0(DX), CX	// make sure g != nil
	
	get_tls(CX)
	
	// 将运行的 g 的指针存放到本地存储, 这样后面可以直接通过线程本地存储获取到当前正在执行的
	// goroutine 的 g 对象, 从而找到与之关联的 m 和 p
	MOVQ	DX, g(CX) 
	
	# CPU 的 SP 寄存器设置为 sched.sp, 栈切换
	MOVQ	gobuf_sp(BX), SP // restore SP
	
	# CPU其他寄存器
	MOVQ	gobuf_ret(BX), AX
	MOVQ	gobuf_ctxt(BX), DX
	MOVQ	gobuf_bp(BX), BP
	
	# 清空 sched 的值
	MOVQ	$0, gobuf_sp(BX)	// clear to help garbage collector
	MOVQ	$0, gobuf_ret(BX)
	MOVQ	$0, gobuf_ctxt(BX)
	MOVQ	$0, gobuf_bp(BX)
	
	# sched.pc 的值( goroutine函数执行的入口地址 )放入 BX 寄存器,
	# JMP 把 BX 寄存器的值放入到 CPU 的 IP 寄存器, 于是, CPU 就跳转到该地址继续执行指令  
	MOVQ	gobuf_pc(BX), BX
	JMP	BX
```

gogo 主要做的事情:

1. 将 sched.g 存储到 tls 当中, 这样通过 tls 直接获取到 g, 然后获取到与之关联的 m 和 p 

2. sched 的成员恢复到 CPU 寄存器完成状态和栈的切换

3. 跳转到 gp.sched.pc 所指的指令地址(goroutine 函数入口)处执行.

到目前为止, runtime.main 函数就被调度执行起来了.

下来, 看看 main 函数:

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
   
    ......  
    
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
    
    // 系统调用, 退出进程, main goroutine 并没有返回, 而是直接进入系统调用退出进程
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

从上述流程来看, runtime.main 在执行玩 main.main 函数之后就直接调用 exit 结束进程了, 它并没有返回到调用它函数. 这里
需要注意, runtime.main 是 main goroutine 的入口函数, 并不是直接被调用的, 而是在 `schedule() -> execute() ->
gogo()` 这个调用链的 gogo 函数中使用汇编代码跳过来的, 从这个角度, goroutine 没有地方可以返回. 但是, 前面的分析当中
得知, 在创建 gorotine 时在其栈上已经放好了一个返回地址, 伪造成 goexit 函数调用了 goroutine 的入口函数, 在这里并没有
使用到这个返回地址, 其实这个地址是为非 main goroutine 准备的, 让其在执行完成之后返回到 goexit 继续执行.




### 调度器如何开启调度循环

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


### 调度循环中如何让出 CPU

#### 正常完成让出 CPU

执行过程:

schedule -> execute -> gogo -> 执行G

创建 G:
     
newproc1 -> goexit0(通过调整 ip 寄存器设置的) -> goexit1 -> schedule)


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
	MOVQ	gobuf_pc(BX), BX    // gobuf.pc, 初始化时指向 goexit
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

### 相关数据结构

// g, Groutine
```cgo
// CPU寄存器相关内容保存, 非常核心.
type gobuf struct {
    sp   uintptr  // 保存CPU的rsp寄存器的值
    pc   uintptr  // 保存CPU的rip寄存器的值
    g    guintptr // 记录当前这个gobuf对象属于哪个goroutine
    
    // ctxt 可能是一个在堆上分配的 funcval, GC本应要追踪它. 但是, 它只是在汇编代码当中进行设置和清楚,
    // 这种状况下, 很难出现写屏障.
    // 并且, ctxt 是实时保存在寄存器当中, 并且我们只在真实的寄存器当中交换它, 它也可以看做一个根节点, 这意味着ctxt
    // 的保存和恢复不存在写屏障.
    ctxt unsafe.Pointer 
 
    // 保存系统调用的返回值, 因为从系统调用返回之后, 如果 p 被其它工作线程抢占，
    // 则这个 goroutine 会被放入全局运行队列被其它工作线程调度, 其它线程需要知道系统调用的返回值.
    ret  sys.Uintreg  
    lr   uintptr
 
    // 保存CPU的rbp寄存器的值
    bp   uintptr // for GOEXPERIMENT=framepointer
}

type g struct{
    stack     stack   // 记录 gorutine 使用的栈
    
    // 栈溢出检查, 实现栈的自动伸缩, 抢占调度也会用到 stackguard0
    stackguard0 uintptr // Go的栈
    stackguard1 uintptr // C的栈
     
    m         *m      // 此goroutine正在被哪个工作线程执行
    sched     gobuf   // 保存调度信息, 主要就是几个CPU寄存器的值
    atomicstatus uint32 // G的状态
    
    // schedlink 指向全局运行队列中的下一个 g, 所有位于全局运行队列中的 g 形成一个链表.
    schedlink guintptr
    
    // 抢占标志, 如果需要抢占调度, 设置为 true
    preempt   bool
    
    lockedm   muintptr  // 锁定的M, G中断恢复指定M执行
    gopc      uintptr   // 创建 goroutine的指令地址
    startpc   uintptr   // gorotine 函数指令地址
    ... 
}
```

// m, 系统线程(工作线程), 保存了 m 自身使用的栈信息, 当前正在运行的 goroutine 以及与 m 绑定的 p 等信息.
```cgo
type m struct{
    // g0 主要用来记录工作线程使用的栈信息, 在执行调度代码时需要使用这个栈,
    // 执行用户 goroutine 代码时, 使用用户 goroutine 自己的栈, 调度时发生栈的切换.
    g0    *g  
    
    // 执行工作线程正在运行的 goroutine 的 g 的结构体对象.
    curg  *g        // 当前正在运行的 g
    p     puintptr  // 绑定 P 执行代表(如果是nil, 则处于空闲状态)
    nextp puintptr  // 当 M 被唤醒时,首先拥有这个 P
    oldp  puintptr  // 执行系统调用时的 p
    
    // 通过 TLS 实现 m 结构体与工作线程之间的绑定. 第1个元素保存的是当前运行的 g, 
    // 通过 g 可以找到 m, 然后通过 m 再找到 p
    tls [6]uintptr 
    mstartfn func() 
    
    // spinning 状态: 表示当前工作线程正在视图从其他工作线程的本地运行队列偷取 goroutine
    spinning bool
    blocked  bool // m is blocked on a note
    
    // 没有 goroutine 需要运行时, 工作线程休眠在这个 park 成员上, 其他线程通过这个 park 唤醒
    // 该工作线程
    park note 
    // 记录所有工作线程的一个链表
    alllink *m 
    schedlink muintptr // 下一个m, 构成链表
    
    // Linux 平台 thread 值就是操作系统线程 ID
    thread    uintptr
    freelink  *m       // on sched.freem  
    ... 
}
```

// p, Procs, 用于保存工作线程执行 go 代码时所必需的资源, 比如 goroutine 的运行队列, 内存分配使用到的缓存等
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


// 全局变量
allgs []*g // 保存所有的 g
allm  *m   // 保存的 m 构成一个链表, 包括 m0
allp  []*p // 保存所有的p, len(allp) == gomaxprocs

ncpu int32 // 系统中cpu核的数量,程序启动时由 runtime 代码初始化
gomaxprocs int32 // p 的最大值, 默认等于 ncpu, 但是可以通过 GOMAXPROCS 修改

sched schedt // 调度器对象, 记录调度器工作状态

m0 m // 代表进程的主线程
g0 g // m0 的 g0, 即 m0.g0 = &g0
```