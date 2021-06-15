# go 调度

内核堆系统线程的调度可以归纳为:**在执行操作系统代码时, 内核调度器按照一定的算法挑选处一个线程并把该线程保存在内存之中的寄
存器值放入CPU对于的寄存器从而恢复该线程的运行.**

万变不离其宗, 系统线程对 goroutine 的调度与内核对系统线程的调度原理是一样的, `实质都是通过保存和修改CPU寄存器的值来达
到切换线程/goroutine的目的.`

因此, 为了实现对 goroutine 的调度, 需要引入一个数据结构来保存CPU寄存器的值以及goroutine的其他一些状态信息, 这个数据结
构就是g, 它保存了goroutine的所有信息, 该结构体的每一个实例对象都代表了一个goroutine, 调度器代码可以通过 g 对象来对
goroutine进行调度, 当goroutine被调离 CPU 时, 调度器代码负责把 CPU 寄存器的值保存在g对象的成员变量当中, 当goroutine
被调度起来运行时, 调度器代码负责把g对象的成员变量所保存的寄存器的值恢复到 CPU 寄存器.


任何一个由编译型语言所编写的程序在被操作系统加载起来运行时都会顺序经过如下几个阶段:

1. 从磁盘上把可执行程序读入内存;

2. 创建进程和主线程; 

3. 为主线程分配栈空间;

4. 把由用户在命令行输入的参数拷贝到主线程的栈;

5. 把主线程放入操作系统的运行队列等待被调度执行起来运行.

宏定义函数: `runtime/go_tls.h`

```cgo
#ifdef GOARCH_amd64
#define	get_tls(r)	MOVQ TLS, r    // 获取 TLS 的位置
#define	g(r)	0(r)(TLS*1)        // 获取 TLS 当中存储 g 的位置 
#endif
```


### Go程序是从哪里启动的?

通过 gdb 的 `info files` 可以查找到go编译文件的函数入口地址, 通过单步调试 `si`, 最终到 `runtime.rt0_go` 这个汇编
函数当中开始进行调度器的初始化.

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
    
    ...
    
    msigsave(_g_.m) // 初始化 m0.gsignal
    initSigmask = _g_.m.sigmask
    
    ... 
    
    sched.lastpoll = uint64(nanotime())
    procs := ncpu // 系统有多少个核, 就创建多少个 p 对象
    if n, ok := atoi32(gogetenv("GOMAXPROCS")); ok && n > 0 {
        procs = n // 通过修改环境变量 GOMAXPROCS, 指定创建p的数量 
    }
    
    // 创建和初始化全局变量 allp
    if procresize(procs) != nil {
        throw("unknown runnable goroutine during bootstrap")
    }
    
    ...	
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
    
    ...
    
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
    ...
    
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
共占用 24 个字节, 所以传递给 newproc 的第一个参数是 24, 表示 sum 函数的大小.

为什么需要传递 fn 函数的参数大小给 newproc 函数? 原因在于 newproc 函数创建一个新的 goroutine 来执行 fn 函数, 而这
个新创建的 goroutine 与当前 goroutine 使用不同的栈, 因此需要在创建 goroutine 的时候把 fn 需要用到的参数从当前栈上
拷贝到新 goroutine 的栈上之后才能开始执行, 而 newproc 函数本身并不知道需要拷贝多少数据到新创建的 goroutine 的栈上去,
所以需要用参数的方式指定拷贝数据的大小.


newproc 函数是对 newproc1 的一个包装, 最重要的工作:

- 获取 fn 函数的第一个参数的地址(argp)

- 使用 systemstack 函数切换到 g0 栈(对于初始化场景来说现在本身就在 g0 栈, 不需要切换, 但是对于用户创建的 goroutine
则需要进行栈切换), 创建一个 newg, 同时进行栈参数拷贝, 同时设置 newg.sched (调整sp, pc指向 goexit, g), 以及 newg
其它相关变量, 完成后 newg 的状态是 _Grunnable

```cgo
func newproc(siz int32, fn *funcval) {
    // 注: siz 当中包含了保存 rip 使用的栈空间(8)
    // 函数调用参数入栈是从右往左, 而且栈是从高地址向低地址增长的
    // 注: argp 指向 fn 函数的第一个参数
    // 参数 fn 在栈上的地址 + 8 = fn 函数的第一个参数. (参考上面的汇编代码)
    argp := add(unsafe.Pointer(&fn), sys.PtrSize)
    gp := getg() // 当前场景是 m0.g0
    
    // getcallerpc() 返回一个地址, 也就是调用 newproc 时 call 指令压栈的函数返回地址,
    // 对于当前场景来说, pc 就是 'CALL runtime·newproc(SB)' 后面的 'POPQ AX' 这条指令地址
    pc := getcallerpc() 
    
    // 切换到 g0 执行作为参数的函数
    systemstack(func() {
        newg := newproc1(fn, argp, siz, gp, pc)
        
        // 获取当前的 g0 绑定的 _p_, 然后将新创建的 newg 放入到 _p_ 本地队列当中.
        // 注: newg 当前还没有和任何 m 进行关联, 只有被调度运行的时才和 m 进行关联
        _p_ := getg().m.p.ptr() 
        runqput(_p_, newg, true)
        
        // 当 runtime.main 执行后, 值为 true
        if mainStarted {
            wakep() // 唤醒一个 p, sched.npidle > 0 && sched.nmspinning == 0, startm(), 开启一个线程
        }
    })
}
```

newproc1 函数参数:

- 第一个参数 fn 是新创建的 goroutine 需要执行的函数, 注意 fn 的类型是 funcval; 
- 第二个参数 argp 是 fn 函数的第一个参数的地址; 
- 第三个参数是 fn 函数以字节为单位的大小;
- 第四个参数是当前运行的 g; 
- 第五个参数调用 newproc 的函数的返回地址.

```cgo
//go:systemstack
func newproc1(fn *funcval, argp unsafe.Pointer, narg int32, callergp *g, callerpc uintptr) *g {
    // 当前已经切换到 g0 栈, 因此无论什么状况下, _g_ = g0 (工作线程的 g0)
    _g_ := getg()  // 对于当前的场景, 这里的 g0 = m0.g0
    
    if fn == nil {
        _g_.m.throwing = -1 // do not dump full stacks
        throw("go of nil func value")
    }
    
    // 禁止抢占
    acquirem() // 增加当前 m 的 locks 值
    siz := narg
    siz = (siz + 7) &^ 7 // size 进行 8 字节对齐
    
    // 对于参数大小的限制, 不能超过 2008 (2048-4*8-8), 如果是 int64 类型, 最多只能是 249 个参数
    // 4*sizeof(uintreg): extra space added below
    // sizeof(uintreg): caller's LR (arm) or return address (x86, in gostartcall).
    if siz >= _StackMin-4*sys.RegSize-sys.RegSize {
        throw("newproc: function arguments too large for new goroutine")
    }
    
    // 当前与 m 绑定的 p, 初始化时, 这里的 _p_ 其实就是 allp[0]
    _p_ := _g_.m.p.ptr() 
    
    // 从本地队列中获取一个 g. 在初始化时没有, 返回值是 nil 
    newg := gfget(_p_)
    if newg == nil {
        // new一个g, 然后从堆上为其分配栈, 并设置 g 的 stack 成员和两个 stackguard 成员
        newg = malg(_StackMin) // 2k栈大小
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
    
    // 调整 newg 的栈顶指针.
    // sys.RegSize=8, sys.MinFrameSize=0, sys.SpAlign=1
    // totalSize 最终大小是 siz+32
    totalSize := 4*sys.RegSize + uintptr(siz) + sys.MinFrameSize // extra space in case of reads slightly beyond frame
    totalSize += -totalSize & (sys.SpAlign - 1)                  // align to spAlign
    sp := newg.stack.hi - totalSize
    spArg := sp
    if usesLR { // 值是 false, 不会执行到此的
        // caller's LR
        *(*uintptr)(unsafe.Pointer(sp)) = 0
        prepGoExitFrame(sp)
        spArg += sys.MinFrameSize
    }
    
    if narg > 0 {
        // 把参数从 newproc 函数的栈(初始化是g0栈)拷贝到新的 newg 的栈
        // 注: 这里是从 sp 的位置开始拷贝的.
        memmove(unsafe.Pointer(spArg), argp, uintptr(narg))
        
        // 这是一个 stack-to-stack 的复制. 
        // 如果启用了写屏障并且 source stack为 grey (目标始终为黑色), 则执行屏障复制.
        // 我们 "after" memmove 之后这样做, 因为目标 stack 上可能有垃圾.
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
    
    // 重新调整 sched 成员和 newg 的栈(参考下面的分析)
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
    
    // 减少当前 m 的 locks 的值, 同时当 m.locks=0 && _g_.preempt 时, 进行抢占标记
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
完成 goroutine 的调度. 每次调度 goroutine 都是从 schedule 函数开始的.


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
        // 保证调度的公平性, 每进行 61 次调度需要优先从全局运行队列中获取 goroutine
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


### main gorutine 

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

从上述流程来看, runtime.main 在执行完 main.main 函数之后就直接调用 exit 结束进程了, 它并没有返回到调用它函数. 这里
需要注意, runtime.main 是 main goroutine 的入口函数, 并不是直接被调用的, 而是在 `schedule() -> execute() ->
gogo()` 这个调用链的 gogo 函数中使用汇编代码跳过来的, 从这个角度, goroutine 没有地方可以返回. 但是, 前面的分析当中
得知, 在创建 goroutine 时在其栈上已经放好了一个返回地址, 伪造成 goexit 函数调用了 goroutine 的入口函数, 在这里并没有
使用到这个返回地址, 其实这个地址是为非 main goroutine 准备的, 让其在执行完成之后返回到 goexit 继续执行.


### 非 main gorutine 退出流程

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

非 main goroutine 返回时直接返回到 goexit 的第二条指令: `CALL	runtime·goexit1(SB)`, 该指令继续调用 goexit1 函
数. 

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

goexit1 函数通过调用 mcall 从当前运行的用户 goroutine 切换到 g0, 然后在 g0 栈上调用和执行 goexit0 函数.

```cgo
// func mcall(fn func(*g))
// 切换到 m->g0 栈上, 然后调用 fn(g) 函数
// fn 函数必须不能返回.
// It should gogo(&g->sched) to keep running g.
// mcall 的参数是一个指向 funcval 对象的指针.
TEXT runtime·mcall(SB), NOSPLIT, $0-8
    # 获取参数的值放入 DI 寄存器, 它是 funcval 对象的指针. 当前场景是 goexit0 的地址
    MOVQ	fn+0(FP), DI 
    
    get_tls(CX)
    MOVQ	g(CX), AX	# AX=g, 这里的 g 是用户 goroutine
    MOVQ	0(SP), BX	# 将 mcall 的返回地址(rip寄存器的值)放入 BX
    
    # 保存 g 的调度信息, 即将切换到 g0 栈
    MOVQ	BX, (g_sched+gobuf_pc)(AX) # g.sched.pc = AX 
    LEAQ	fn+0(FP), BX # fn 是调用方的栈顶元素, 其地址就是调用方的栈顶
    MOVQ	BX, (g_sched+gobuf_sp)(AX) # g.sched.sp = BX, 用户 goroutine 的 rsp 
    MOVQ	AX, (g_sched+gobuf_g)(AX) # g.sched.g = AX
    MOVQ	BP, (g_sched+gobuf_bp)(AX) # g.sched.bp = BP, 用户 goroutine 的 rbp 
    
    # 切换到 g0 栈, 然后调用 fn 
    MOVQ	g(CX), BX    # BX = g 
    MOVQ	g_m(BX), BX  # BX = g.m 
    MOVQ	m_g0(BX), SI # SI = g0 
    
    # 此时, SI=g0, AX=g, 这里需要判断 g 是否是 g0 
    CMPQ	SI, AX	// if g == m->g0 call badmcall
    JNE	3(PC) # 不相等
    MOVQ	$runtime·badmcall(SB), AX
    JMP	AX
    MOVQ	SI, g(CX) # 将本地存储设置为 g0
    MOVQ	(g_sched+gobuf_sp)(SI), SP	# 从 g0.sched.sp 当中恢复 SP, 即 rsp 寄存器  
    PUSHQ	AX  # fn 的参数 g 入栈
    MOVQ	DI, DX # DX=fn 
    MOVQ	0(DI), DI # 判断fn不为nil
    CALL	DI # 调用 fn 函数, 该函数不会返回, 这里调用的函数是 goexit0 
    POPQ	AX # 正常状况下, 这里及其之后的指令不会执行的
    MOVQ	$runtime·badmcall2(SB), AX
    JMP	AX
    RET
```

mcall 的参数是一个函数, 在 Go 当中, 函数变量并不是一个直接指向函数代码的指针, 而是一个指向 funcval 结构体对象的指针,
funcval 结构体对象的第一个成员 fn 才是真正指向函数代码的指针.

mcall 函数的功能:

1. 首先从当前运行的 g 切换到 g0, 这一步包括保存当前 g 的调度信息(pc,sp,bp,g), 把 g0 设置到 tls 当中, 修改 CPU 的 
rsp 寄存器使其指向 g0 的栈.

2. 以当前运行的 g 为参数调用 fn 函数(此处是 goexit0). 

> 注: 在 g0 保存有 pc, 但是这里并不会从 pc 处开始调用, 而是直接 call 给得的 fn 函数.

从 mcall 的功能看, mcall 做的事情与 gogo 函数完全相反, gogo 实现了从 g0 切换到某个 goroutine 去运行, 而 mcall
实现了从某个 goroutine 切换到 g0 来运行. 因此, mcall 和 gogo 的代码很相似.

mcall 和 gogo 在做切换时有个重要的区别: gogo 函数在从 g0 切换到其他 goroutine 时, 首先切换了栈, 然后通过跳转指令
从 runtime 切换到了用户 goroutine 的代码. 而 mcall 函数在从其他 goroutine 切换回 g0 时只切换了栈, 并未使用跳转
指令跳转到 runtime 代码去执行. 为何有这种差别? 原因在于从 g0 切换到其他 goroutine 之前执行的是 runtime 的代码并且
使用的是 g0 栈, 因此切换时先切换栈然后从 runtime 代码跳转某个 goroutine 的代码去执行(切换栈和跳转指令不能颠覆), 然而
从某个 goroutine 切换回 g0 时, goroutine 使用的是 call 指令来调用 mcall 函数, **mcall 本身就是 runtime 的代码,
所以 call 指令其实已经完成从 goroutine 代码切换 runtime 代码的跳转, 因此 mcall 函数自身无需再跳转了, 只需要把栈切
回来即可.**


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


### M 是如何创建的? g0 是什么时候创建的

启动一个新的系统线程 M, 调用函数 startm, 该函数的所做的事情不是很复杂:

首先, 获取一个空闲的 p (当参数的 p 为nil时, 则从 sched.pidle 当中获取一个)

其次, 从 sched.midle 当中获取一个空闲的 m, 如果不存在, 则调用 newm 创建一个, 并自旋.

最后, 当 sched.midle 当中获取到了休眠的 m, 将 `_p_` 绑定到 m.nextp, 然后调用 notewakeup() 函数唤醒休眠的 m


```cgo
// 如果p == nil, 则尝试获取一个空闲p, 如果没有空闲 p, 则直接返回, 也就是说所有的 p 都绑定了 m, 无需再启动 m 了.
// 如果设置了spinning, 则调用者已增加nmspinning, 而startm将减少nmspinning或在新启动的M中设置m.spinning.
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
```

在没有可用的 m 时, 则需要创建一个新的系统线程 m, 来执行 g.

```cgo
// newm -> newm1 -> newosproc(传入 mstart函数指针, 作为启动函数, 运行在 g0 上)
// 创建一个 M, 设置 M 的初始化函数是 fn (这个函数是在 g 上执行）
func newm(fn func(), _p_ *p, id int64) {
    // 在堆上分配一个 m, 并且与 _p_ 进行绑定. 
    // 注: allocm() 在创建 m 的同时会创建 g0, 并且分配相关的栈内存
    mp := allocm(_p_, fn, id)
    mp.nextp.set(_p_) // 将 m.nextp 也设置为 _p_
    mp.sigmask = initSigmask
    
    // 当前处于锁定的 m, 则不能通过clone 当前的 m 来创建系统线程. 
    // 对于已经创建的 mp 对象, 会被添加对象, 则被添加到 newmHandoff.new 队列当中. 
    if gp := getg(); gp != nil && gp.m != nil && (gp.m.lockedExt != 0 || gp.m.incgo) && GOOS != "plan9" {
        // 当前处于锁定的m 或 可能由C启动的线程上. 此线程的内核状态可能很奇怪(用户可能已为此目的将其锁定).
        // 我们不想将其克隆到另一个线程中. 而是要求一个已知良好的线程为我们创建线程.	
        lock(&newmHandoff.lock)
        if newmHandoff.haveTemplateThread == 0 {
            throw("on a locked thread with no template thread")
        }
        // 将 mp 添加到 newmHandoff.newm 链表当中
        mp.schedlink = newmHandoff.newm
        newmHandoff.newm.set(mp)
        if newmHandoff.waiting {
            newmHandoff.waiting = false
            notewakeup(&newmHandoff.wake)
        }
        unlock(&newmHandoff.lock)
        return
    }
    
    // 创建系统线程, 并且与 mp 进行关联起来
    newm1(mp)  
}

func newm1(mp *m) {
    // 当前需要创建 cgo 使用的系统线程
    if iscgo {
        var ts cgothreadstart
        if _cgo_thread_start == nil {
            throw("_cgo_thread_start missing")
        }
        ts.g.set(mp.g0)
        ts.tls = (*uint64)(unsafe.Pointer(&mp.tls[0]))
        ts.fn = unsafe.Pointer(funcPC(mstart))
        if msanenabled {
            msanwrite(unsafe.Pointer(&ts), unsafe.Sizeof(ts))
        }
        execLock.rlock() // Prevent process clone.
        asmcgocall(_cgo_thread_start, unsafe.Pointer(&ts))
        execLock.runlock()
        return
    }
    
    // 创建正常的工作线程
    execLock.rlock() 
    newosproc(mp) // 真正的创建工作
    execLock.runlock()
}

// 创建系統内核线程, 初始化运行的函数是 mstart. 
func newosproc(mp *m) {
    // g0 的栈顶位置
    stk := unsafe.Pointer(mp.g0.stack.hi)
    
    // 在进行系统线程 clone 期间中禁用 signal, 这样新线程可以从禁用的信号开始,
    // 之后它将调用 minit 
    var oset sigset
    sigprocmask(_SIG_SETMASK, &sigset_all, &oset)
    // 系统调用 clone 函数, 克隆一个新的系统线程
    // 第一个参数 cloneFlags 克隆的 flags, 是否共享内存等
    // 第二个参数 stk, 设置新线程的栈顶
    // 第三个参数是 mp
    // 第四个参数是 g0
    // 第五个参数是新线程启动之后开始执行的函数地址. 这里设置的是 mstart, 在程序启动时, 最后一步调用的也是 mstart
    // 这个函数.
    ret := clone(cloneFlags, stk, unsafe.Pointer(mp), unsafe.Pointer(mp.g0), unsafe.Pointer(funcPC(mstart)))
    sigprocmask(_SIG_SETMASK, &oset, nil) // 启用信号
}


// 分配与任何线程无关的新 m. 
// 如果需要, 可以将 p 用于分配上下文.
// fn 是为新 m 的 m.mstartfn. 
// id 是可选的预分配的 m.id. -1 会被忽略
//
// 此函数即使没有调用者也被允许有写障碍, 因为它借用了_p_.
func allocm(_p_ *p, fn func(), id int64) *m {
    _g_ := getg()
    acquirem() // 当前的 m 禁止抢占
    
    // 当的 m 没有绑定任何 p, 则需要绑定一个 p 
    if _g_.m.p == 0 {
        acquirep(_p_) 
    }
    
    // 开始是否 sched.freem 当中 m.g0 的栈空间
    if sched.freem != nil {
        lock(&sched.lock)
        var newList *m // 记录不能被释放 m.g0 栈空间的 m 列表
        for freem := sched.freem; freem != nil; {
            if freem.freeWait != 0 {
                next := freem.freelink
                freem.freelink = newList
                newList = freem
                freem = next
                continue
            }
            stackfree(freem.g0.stack) // 释放 g0 stack
            freem = freem.freelink
        }
        sched.freem = newList // 设置新的 freem 列表
        unlock(&sched.lock)
    }
    
    mp := new(m)
    mp.mstartfn = fn   // 在 M 启动之后, 调度之前执行的函数
    mcommoninit(mp, id) // 开始初始化m, 将 mp 添加到全局队列当中
    
    // 创建 g0, 并将 g0 与 m 进行绑定.
    // 如果是 cgo 或 Solaris 或 illumos 或 Darwin, pthread_create 将创建堆栈.
    // Windows 和 Plan9 将调度的堆栈安排在OS堆栈上.
    if iscgo || GOOS == "solaris" || GOOS == "illumos" || GOOS == "windows" || GOOS == "plan9" || GOOS == "darwin" {
        mp.g0 = malg(-1)
    } else {
        mp.g0 = malg(8192 * sys.StackGuardMultiplier) // 8K
    }
    mp.g0.m = mp
    
    if _p_ == _g_.m.p.ptr() {
        releasep() // 释放掉前面临时绑定的 p 
    }
    releasem(_g_.m) // 取消抢占
    
    return mp
}
```


systemstack() 函数: 切换到 g0 栈上, 执行函数 func

```cgo
// func systemstack(fn func())
TEXT runtime·systemstack(SB), NOSPLIT, $0-8
    MOVQ	fn+0(FP), DI	// DI = fn
    get_tls(CX)
    MOVQ	g(CX), AX	// AX = g
    MOVQ	g_m(AX), BX	// BX = g.m, 当前工作线程 m
    
    CMPQ	AX, m_gsignal(BX) // g == m.gsignal
    JEQ	noswitch // 相等跳转到 noswitch
    
    MOVQ	m_g0(BX), DX // DX = m.g0
    CMPQ	AX, DX // g == m.g0
    JEQ	noswitch // 相等则跳转 noswitch, 当前在 g0 栈上
    
    CMPQ	AX, m_curg(BX) // g == m.curg
    JNE	bad // 不相等, 程序异常
    
    // 切换 stack, 从 curg 切换到 g0
    // 将 curg 保存到 sched 当中
    MOVQ	$runtime·systemstack_switch(SB), SI // SI=runtime·systemstack_switch, 空函数地址
    MOVQ	SI, (g_sched+gobuf_pc)(AX) // g.sched.pc=SI
    MOVQ	SP, (g_sched+gobuf_sp)(AX) // g.sched.sp=SP
    MOVQ	AX, (g_sched+gobuf_g)(AX)  // g.sched.g=g
    MOVQ	BP, (g_sched+gobuf_bp)(AX) // g.sched.bp=BP
    
    MOVQ	DX, g(CX) // 切换 tls 到 g0
    MOVQ	(g_sched+gobuf_sp)(DX), BX // BX=g0.sched.sp
    
    // 栈调整, 伪装成 mstart() 调用函数 systemstack(), 目的是停止回溯
    SUBQ	$8, BX
    MOVQ	$runtime·mstart(SB), DX // DX=runtime·mstart
    MOVQ	DX, 0(BX) // 将 runtime·mstart 函数地址入栈
    MOVQ	BX, SP // 调整当前的 SP 
    
    // 调用 target 函数
    MOVQ	DI, DX   // DX=fn 
    MOVQ	0(DI), DI // 判断 fn 非空
    CALL	DI // 函数调用, 没有参数和返回值
    
    // 函数调用完成, 切换到 curg 栈上
    get_tls(CX)
    MOVQ	g(CX), AX   
    MOVQ	g_m(AX), BX // BX=m
    MOVQ	m_curg(BX), AX // AX=m.curg
    MOVQ	AX, g(CX) // 设置本地保存 m.curg 
    MOVQ	(g_sched+gobuf_sp)(AX), SP // SP = m.curg.sched.sp
    MOVQ	$0, (g_sched+gobuf_sp)(AX) // m.curg.sched.sp = 0
    RET

noswitch:
    // 当前已经在 g0 栈上了, 直接调用函数
    MOVQ	DI, DX
    MOVQ	0(DI), DI
    JMP	DI

bad:
    // Bad: g is not gsignal, not g0, not curg. What is it?
    MOVQ	$runtime·badsystemstack(SB), AX
    CALL	AX
    INT	$3
```
