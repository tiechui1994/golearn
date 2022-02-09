## go 调度

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

宏定义: `runtime/go_tls.h`

```cgo
#ifdef GOARCH_amd64
#define	get_tls(r)	MOVQ TLS, r    // 获取 TLS 的位置
#define	g(r)	0(r)(TLS*1)        // 获取 TLS 当中存储 g 的位置 
#endif
```

使用上述的两个代码可以获取当前线程当中存储的 g 对象, 从而获取 m, p 对象.


## Go程序是从哪里启动的?

通过 gdb 的 `info files` 可以查找到 Go 编译文件的函数入口地址(`Entry point:0x45bc80`), 对该地址进行打断点, 执行,
到达 `_rt0_amd64_linux` 函数, 该函数就是 Go 程序的入口地址.

// runtime/rt0_linux_amd64.s

```asm
TEXT _rt0_amd64_linux(SB),NOSPLIT,$-8
	JMP	_rt0_amd64(SB)
```

// runtime/asm_amd64.s

```asm
TEXT _rt0_amd64(SB),NOSPLIT,$-8
	MOVQ	0(SP), DI	// argc
	LEAQ	8(SP), SI	// argv
	JMP	runtime·rt0_go(SB)

TEXT runtime·rt0_go(SB),NOSPLIT,$0
    MOVQ	DI, AX		 // argc
    MOVQ	SI, BX		 // argv
    
    // 调整SP
    SUBQ	$(4*8+7), SP // 预留39字节空间, 2args, 2auto
    ANDQ	$~15, SP     // 调整栈顶寄存器按16字节对齐
    MOVQ	AX, 16(SP)   // 保存 argc
    MOVQ	BX, 24(SP)   // 保存 argv
    
    // 给 g0 预留栈空间, 大约是64k
    MOVQ	$runtime·g0(SB), DI  // 将 g0 的地址存入 DI 
    LEAQ	(-64*1024+104)(SP), BX // 预分配大约64k的栈
    MOVQ	BX, g_stackguard0(DI) // g0.stackguard0=SP-64*1024+104
    MOVQ	BX, g_stackguard1(DI) // g0.stackguard1=SP-64*1024+104
    MOVQ	BX, (g_stack+stack_lo)(DI) // g0.stack.lo=SP-64*1024+104
    MOVQ	SP, (g_stack+stack_hi)(DI) // g0.stack.hi=SP
    
    // 省略了 CPU 检查代码
    ....
    
    // 初始化 m0 的 tls
    LEAQ	runtime·m0+m_tls(SB), DI // DI=&m0.tls
    CALL	runtime·settls(SB) // 调用 settls 设置线程的TLS, settls 的参数在DI当中.
                               // 之后, 可以通过 fs 段寄存器获取 m.tls
    
    // 测试 tls
    // 获取 fs 段基址并放入到 BX, 其实就是 m0.tls[1] 的地址(原因后面会讲到).
    // get_tls 是代码由编译器生成
    get_tls(BX) 
    MOVQ	$0x123, g(BX) // 将 0x123 拷贝到 fs基址-8 的位置, 即: m0.tls[0]=0x123 
    MOVQ	runtime·m0+m_tls(SB), AX // AX=m0.tls[0]
    CMPQ	AX, $0x123
    JEQ 2(PC) // 相等,则跳过2条指令(包含本身这条)
    CALL	runtime·abort(SB) // 线程本地存储功能不正常, 退出程序
ok:
    // set the per-goroutine and per-mach "registers"
    // 获取 fs 段基址到 BX
    get_tls(BX) 
    LEAQ	runtime·g0(SB), CX // CX=&g0
    MOVQ	CX, g(BX)  // 把g0的地址保存到线程本地存储, 即 m0.tls[0]=&g0
    LEAQ	runtime·m0(SB), AX // AX=&m0
    
    // m0 与 g0 进行关联
    MOVQ	CX, m_g0(AX) // m0.g0=&g0 
    MOVQ	AX, g_m(CX) // g0.m=&m0
    
    // 到此位置, m0与g0绑定在一起, 之后通过 getg 可获取到 g0, 通过 g0 又获取到 m0.
    // 这样就实现了 m0, g0 与主线程直接的关联.
    CLD		// convention is D is always left cleared
    CALL	runtime·check(SB)
    
    MOVL	16(SP), AX		// copy argc
    MOVL	AX, 0(SP)
    MOVQ	24(SP), AX		// copy argv
    MOVQ	AX, 8(SP) 
    CALL	runtime·args(SB)   // 解析命令行
    
    // 初始化系统核心数, osinit 唯一功能就是获取CPU核数并放到变量 ncpu 中
    CALL	runtime·osinit(SB) 
    // 调度器初始化
    CALL	runtime·schedinit(SB) 
    
    // 创建一个 main goroutine 来启动程序
    MOVQ	$runtime·mainPC(SB), AX	// goroutine 函数入口. AX=&funcval{runtime.main}
    PUSHQ	AX // newproc 第二个参数, 新的 goroutine 需要执行的函数. 
    PUSHQ	$0 // newproc 第一个参数, runtime.main 函数需要的参数大小. 这里是0
    CALL	runtime·newproc(SB)
    POPQ	AX
    POPQ	AX
    
    // 主线程启动, 进入调度循环, 运行刚刚创建的 main goroutine
    CALL	runtime·mstart(SB)
    
    CALL	runtime·abort(SB)	// mstart启动之后永远不会返回, 万一返回了, 需要 crash
    RET
```

rt0_go 函数的大体工作:

1. 调整SP, 然后给 g0 分配栈空间.

2. 主线程与m0绑定: 先调用 settls 设置 fs 段基址, fs段当中写入数据,  最后比较地址当中值和写入数据是否一致. 当测试通过之后,
就将 g0 地址写入到线程本地存储当中. 

3. m0 和 g0 进行绑定.  

4. 调用 runtime.args() 解析 args, 调用 runtime.osinit() 初始化系统核心数.调用 runtime.schedinit() 调度器初始
化.

5. 创建 main goroutine(运行的函数是 runtime.main)

6. 调用 runtime.mstart() 函数启动主线程, 进入调度循环.


M0 是什么? 程序会启动多个 M, 第一个启动的是 M0, 并且 M0 是操作系统启动进程的时候创建的. 其他的 M 都是 runtime 通过
系统调用 clone 进行创建的.

G0 是什么? 在 Go 当中 G 分为三种, 第一种是用户创建任务的叫做 G. 第二种是执行 runtime 下调度工作的叫 G0, 每一个 M 都
绑定一个 G0. 第三种是启动 runtime.main 用到的 G. 程序用到是基本上就是第一种.


### 调整 SP

```
SUBQ	$(4*8+7), SP // 预留39字节空间, 2args, 2auto
ANDQ	$~15, SP     // 调整栈顶寄存器按16字节对齐
```

先是将 SP 减掉 39, 即: 向下移动39 Byte. 然后进行与运算.

`~15` 表示对15进行取反操作. 15 的二进制是 `1111`, 其他位都是0; 取反后, 变成 `0000`, 高位全是1. 这样在与SP进行与运
算后, 低四位变成了0, 高位不变. 这样就达到了SP地址16字节对齐.

为什么要进行16字节对齐? 因为CPU有一组SSE指令, 这些指令中出现的内存地址必须是16的倍数.

### 主线程绑定 m0

```
    // 初始化 m0 的 tls
    LEAQ	runtime·m0+m_tls(SB), DI // 获取 m0.tls 的地址, 存放到 DI 当中
    CALL	runtime·settls(SB) // 调用 settls 设置线程的TLS, settls 的参数在DI当中.
                               // 之后, 可以通过 fs 段寄存器获取 m.tls
    
    // 测试 tls
    // 获取 fs 段基址并放入到 BX, 其实就是 m0.tls[1] 的地址(原因后面会讲到).
    // get_tls 是代码由编译器生成
    get_tls(BX) 
    MOVQ	$0x123, g(BX) // 将 0x123 拷贝到 fs基址-8 的位置, 即: m0.tls[0]=0x123 
    MOVQ	runtime·m0+m_tls(SB), AX // AX=m0.tls[0]
    CMPQ	AX, $0x123
    JEQ 2(PC) // 相等,则跳过2条指令(包含本身这条)
    CALL	runtime·abort(SB) // 线程本地存储功能不正常, 退出程序
ok:
    // set the per-goroutine and per-mach "registers"
    // 获取 fs 段基址到 BX
    get_tls(BX) 
    LEAQ	runtime·g0(SB), CX // CX=&g0
    MOVQ	CX, g(BX)  // 把g0的地址保存到线程本地存储, 即 m0.tls[0]=&g0
    LEAQ	runtime·m0(SB), AX // AX=&m0
    
    // m0 与 g0 进行关联
    MOVQ	CX, m_g0(AX) // m0.g0=&g0 
    MOVQ	AX, g_m(CX) // g0.m=&m0
```

m0 是全局变量, m0 需要绑定到工作线程, 才能进行调度执行.

这里需要说明一下 `runtime.settls()` 函数, 它在 `runtime/sys_linux_amd64.s` 当中. 内容如下:

```
TEXT runtime·settls(SB),NOSPLIT,$32
	ADDQ	$8, DI	// DI=DI+8
	MOVQ	DI, SI  // 系统调用第二个参数,
	MOVQ	$0x1002, DI	// 系统调用第一个参数, ARCH_SET_FS, 表示设置FS的基址
	MOVQ	$SYS_arch_prctl, AX // 系统调用号
	SYSCALL
	CMPQ	AX, $0xfffffffffffff001 // AX 与 -1 进行比较
	JLS	2(PC)
	MOVL	$0xf1, 0xf1  // crash
	RET
```

前面说道了 DI 里面存放的是 m0.tls 的地址, 那么 `ADDQ	$8, DI` 表示对 DI 地址偏移8字节, 也就在指向了 `m0.tls[1]`
的位置处. 

接下来就是准备 `arch_prctl` 系统调用的参数. Linux 系统调用是使用特定的寄存器传递参数的. 其中 DI, SI, DX, R10, R8,
R9 用于传递系统调用参数, AX 用于传递系统调用号. 系统调用返回后, AX 用于传递系统调用失败错误码(0表示成功). 

arch_prctl 在操作码是 $0x1002 (ARCH_SET_FS) 时, 表示设置 fs 的基址. 这里也就是 `m0.tls[1]` 的地址. 当设置好 FS
基址之后, 每次可以通过 `fs基址 + 偏移量` 来获取工作线程的线程本地存储的值了(线程全局私有变量).

> `arch_prctl` 系统调用详情, 参考: https://man7.org/linux/man-pages/man2/arch_prctl.2.html

`arch_prctl` 系统调用完成之后, 比较返回的错误码(AX)与-1的关系. 当错误码小于-1时, 系统调用失败, 函数就会 crash 掉,
否则, 系统调用成功, settls 返回.


在 `settls` 之后, 使用 `get_tls(BX)` 获取tls, 该代码由编译器生成. 可以理解为将 `m.tls` 的地址存储到 BX. 然后将
0x123 存放到 `m.tls[0]` 处. 这两行代码在实际汇编时只会生成一行代码:

```
movq  $0x123, %fs:0xfffffffffffffff8
```

接下来是比较 `m.tls[0]` 处的值和 0x123 是否一致. 如果一致, 则说明 tls 可以工作. 接下来就是将 g0 的地址存放到 tls 
当中了. 原因在于通过 g 可以获取 m, 然后通过 m 可以获取到 p. 

之后就是 m0 与 g0 进行绑定了. 需要注意的是, 这里的 g0 是工作在系统栈上, 只能进行调度, 不能用于执行任务.


### runtime.osinit() 针对系统环境的初始化

osinit() 唯一的作用的就是初始化全局变量 ncpu

### runtime.schedinit() 调度初始化

schedinit 做的重要事情:

- 初始化 m0, mcommoninit() 函数, 通用 m 初始化.

- 调用 msigsave() 初始化 m0.gsignal

- 调用 procresize() 初始化 allp, 同时将 m0 绑定到 `allp[0]` 即 `m0.p = allp[0], allp[0].m = m0` 

```cgo
func schedinit() {
    // getg 最终是插入代码, 格式如下:
    // gettls(CX)
    // MOVQ g(CX), BX; BX 当中就是当前 g 的结构体对象的地址
    _g_ := getg() // _g_ = &g0
    if raceenabled {
        _g_.racectx, raceprocctx0 = raceinit()
    }
    
    // 设置最多启动 10000 个操作系统线程, 即 m 最多 10000 个
    sched.maxmcount = 10000
    
    tracebackinit()
    moduledataverify()
    stackinit()
    mallocinit()
    fastrandinit() // must run before mcommoninit
    mcommoninit(_g_.m, -1) // 初始化 m0, 因为 g0.m = &m0
    cpuinit()       // must run before alginit
    alginit()       // maps must not be used before this call
    modulesinit()   // provides activeModules
    typelinksinit() // uses maps, activeModules
    itabsinit()     // uses activeModules
    
    msigsave(_g_.m) // 初始化 m0.gsignal
    initSigmask = _g_.m.sigmask
    
    goargs()
    goenvs()
    parsedebugvars()
    gcinit()
    
    sched.lastpoll = uint64(nanotime())
    procs := ncpu // 系统核数量, 在 osinit() 当中调用 getproccount 获取.
    if n, ok := atoi32(gogetenv("GOMAXPROCS")); ok && n > 0 {
        procs = n // 环境变量 GOMAXPROCS, 修改  procs 数量
    }
    
    // 创建和初始化全局变量 allp, 这里会进一步限制 procs 的值.
    if procresize(procs) != nil {
        throw("unknown runnable goroutine during bootstrap")
    }
    
    ...	
}
```

前面汇编当中, g0的地址已经被设置到了线程 m0 的 TLS之中, schedinit 通过 getg 函数(getg函数是编译器实现的, 源码当中找
不到其定义的) 从 TLS 中获取当前正在运行的 g (这里是g0).

调用 mcommoninit() 函数对 m0 进行必要的初始化(这是一个通用的 m 初始化函数), m0 初始化完成之后, 调用 procresize() 
初始化系统需要用到的 p 结构体对象. 它的数量决定了最多同时有多少个 goroutine 同时并行运行. 

sched.maxmcount 设置了 M 最大数量, 而 M 代表系统内核线程, 因此一个进程最大只能启动 `10000` 个系统内核线程. 

procresize 初始化P的数量, procs 参数为初始化的数量, 而在初始化之前先做数量的判断, 默认是 ncpu(CPU核数). 也可以通过
GOMAXPROCS来控制 P 的数量. _MaxGomaxprocs 控制了最大数量只能是1024.

> 注: go 运行时当中中, m 最多是10000个, p 最多是 1024 个.


关注一下 mcommoninit() 函数如何初始化 m0 以及 procresize() 函数如何创建和初始化 p 结构体对象.

````cgo
func mcommoninit(mp *m, id int64) {
    _g_ := getg() // 在初始化时 _g_ = &g0
    
    // g0 stack won't make sense for user (and is not necessary unwindable).
    // 函数调用栈 traceback
    if _g_ != _g_.m.g0 {
        callers(1, mp.createstack[:])
    }
    
    lock(&sched.lock)
    
    // 初始化时, id 为 -1, 生成 m 的 id
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
    
    // 将 m 插入到全局链表 allm 之中.
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

当 `mp.alllink = allm` 执行后, m0 与 allm 的关系如下图①所示. 利用 alllink 字段, 将所有的 m 形成一个闭合的单向链
表(通过任意一个 m 可以遍历完所有的 m).

![image](/images/develop_runtime_allmlink.png)

mcommoninit() 函数重点就是初始化了 m0 的 `id, fastrand, gsignal ...` 等变量,  然后 m0 放入到全局链表 allm 之中, 
然后就返回了.


在 schedinit() 的最后, 是初始化 allp. 函数内容如下:

```cgo
func procresize(nprocs int32) *p {
    old := gomaxprocs // 系统初始化时, gomaxprocs = 0
    if old < 0 || nprocs <= 0 {
        throw("procresize: invalid arg")
    }
    
    // 更新统计时间. procresize() 函数在运行时的任意时刻被调用(不建议这样做).
    now := nanotime()
    if sched.procresizetime != 0 {
        sched.totaltime += int64(old) * (now - sched.procresizetime)
    }
    sched.procresizetime = now
    
    // 初始化时 len(allp) == 0. 扩展 allp.
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
    
    // 初始化 nprocs - gomaxprocs 个 p
    for i := old; i < nprocs; i++ {
        pp := allp[i]
        if pp == nil {
            pp = new(p) // 直接从堆上分配一个 p
        }
        pp.init(i) // 初始化(status,deferpool,mcache等)
        atomicstorep(unsafe.Pointer(&allp[i]), unsafe.Pointer(pp)) // 存储到allp对应的位置
    }
    
    // 当前处于初始化状况下 _g_ = g0, 并且此时 m0.p 还未初始化, 因此在这里会初始化 m0.p
    _g_ := getg() 
    if _g_.m.p != 0 && _g_.m.p.ptr().id < nprocs {
        _g_.m.p.ptr().status = _Prunning
        _g_.m.p.ptr().mcache.prepareForSweep()
    } else {
        // 条件: m 没有关联 p, 或者 p 需要被销毁.
        // 释放当前的 P, 同时获取 all[0]
        //
        // We must do this before destroying our current P
        // because p.destroy itself has write barriers, so we
        // need to do that from a valid P.
        // 原因: 因为 p.destroy 是带有写屏障的, 因此需要获取一个合法的 P
        if _g_.m.p != 0 {
            if trace.enabled {
                // Pretend that we were descheduled
                // and then scheduled again to keep
                // the trace sane.
                traceGoSched()
                traceProcStop(_g_.m.p.ptr())
            }
            _g_.m.p.ptr().m = 0 // p 解绑 m
        }
        _g_.m.p = 0 // m 解绑 p
        
        p := allp[0] // 获取 allp[0], 临时绑定到当前的 m 上
        p.m = 0
        p.status = _Pidle
        acquirep(p) // p, m 关联, 并且 p 的状态为  _Prunning
        if trace.enabled {
            traceGoStart()
        }
    }
    
    // g.m.p is now set, so we no longer need mcache0 for bootstrapping.
    mcache0 = nil
    
    // 销毁多余的 p, 此时还不能释放 p (因为处于系统调用中的 m 可能引用 p, 即 m.p = p)
    // 初始化时 old 是 0
    for i := nprocs; i < old; i++ {
        p := allp[i]
        p.destroy()
    }
    
    // resize allp大小
    if int32(len(allp)) != nprocs {
        lock(&allpLock)
        allp = allp[:nprocs]
        unlock(&allpLock)
    }
    
    // 将所有空闲的p放入空闲链表. 这是遍历 allp 数组.
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
            pidleput(p) // 放入 sched.pidle 列表当中
        } else {
            // 给 p 绑定一个 m(从 sched.midle 当中获取).
            // 通过 link 将非空的的 p 组成一个闭合的单向链表, 链表的头是 runnablePs(最终的返回值)
            // 结论: 当重新进行 resize P 时, P(除了当前m绑定的P)的状态都将被标记为  _Pidle, 
            // 同时, 重新进行去绑定 m 
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

- 使用初始化 nprocs - gomaxprocs (gomaxprocs记录上一次生成p的数量)  个 p 结构体对象并依次保存在 allp 切片之中.

- 给当前的 m 绑定一个 p. 如果当前 m 和 p 已经绑定, 并且 m.id 不在移除范围之内, 设置一下 m.p 的状态. 否则, 把 m 和 
`allp[0]` 临时绑定, 即 `m0.p = allp[0], allp[0].m = m0`. 为后续删除做准备.

- 删除多余的 p. (初始化的时候, 不会进入此操作的)

- 调整 allp, 遍历 allp, 对于除了当前 m.p 之外的 p, 如果空闲, 则放入全局变量 sched.pidle 空闲队列之中. 如果非空闲,
则重新绑定一个 m, 同时通过 link 组装成一个单向闭合链表(返回内容).

### 创建 main goroutine

在进行 schedinit() 完成调度系统初始化后, `m0 与 g0`, `m0 与 allp[0]` 已经相互关联了. 接下来的操作就行需要创建 main
goroutine 用于执行 runtime.main() 函数. 

调用 `newproc()` 创建一个新的 goroutine 用于执行 mainPC 所对应的 runtime.main 函数. 

// runtime/asm_amd64.s

```cgo
// 创建 main goroutine, 也是系统的第一个 goroutine 
// create a new goroutine to start program
MOVQ	$runtime·mainPC(SB), AX // mainPC 是 runtime.main 
PUSHQ	AX // AX=&funcval{runtime.main}, newproc 的第二个参数(新的goroutine需要执行的函数)
PUSHQ	$0 // newproc 的第一个参数, 该参数表示 runtime.main 函数需要的参数大小, 因为没有参数, 所以这里是0
CALL	runtime·newproc(SB) // 创建 main goutine
POPQ	AX
POPQ	AX

// 主线程进入调度循环，运行刚刚创建的goroutine
CALL  runtime·mstart(SB)  

// 上面的mstart永远不应该返回的, 如果返回了, 一定是代码逻辑有问题, 直接abort
CALL  runtime·abort(SB)
```

关于 mainPC 的定义:

```cgo
// runtime·mainPC 的定义.
DATA  runtime·mainPC+0(SB)/8,$runtime·main(SB)
GLOBL runtime·mainPC(SB),RODATA,$8
```


先分析下 newproc():

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
0x001d 00029 (sum.go:7) MOVL    $24, (SP) // newproc 第一个参数, 参数大小
0x0024 00036 (sum.go:7) LEAQ    "".sum·f(SB), AX 
0x002b 00043 (sum.go:7) MOVQ    AX, 8(SP)  // newproc 第二个参数, 函数指针
0x0030 00048 (sum.go:7) MOVQ    $1, 16(SP) // 函数参数1
0x0039 00057 (sum.go:7) MOVQ    $2, 24(SP) // 函数参数2 
0x0042 00066 (sum.go:7) MOVQ    $3, 32(SP) // 函数参数3
0x004b 00075 (sum.go:7) CALL    runtime.newproc(SB)
```

首先把 sum 函数需要用到的 3 个参数入栈, 然后是 newproxc() 的 2 个参数入栈 然后调用 newproc 函数. 因为 sum 函数的
3 个 int64 类型的参数共占用 24 个字节, 所以传递给 newproc 的第一个参数是 24, 表示 sum 函数的大小.

为什么需要传递 fn 函数的参数大小给 newproc 函数? 原因在于 newproc 函数创建一个新的 goroutine 来执行 fn 函数, 而这
个新创建的 goroutine 与当前 goroutine 使用不同的栈, 因此需要在创建 goroutine 的时候把 fn 需要用到的参数从当前栈上
拷贝到新 goroutine 的栈上之后才能开始执行, 而 newproc 函数本身并不知道需要拷贝多少数据到新创建的 goroutine 的栈上去,
所以需要用参数的方式指定拷贝数据的大小.


newproc() 是对 newproc1() 的一个包装, 重要的工作:

- 获取 fn 函数的第一个参数的地址(argp), fn 函数的返回地址(pc).

- 使用 systemstack 函数切换到 g0 栈(初始化时, 当前就在 g0 栈, 不需要切换, 但是对于用户创建的 goroutine 则需要进行
栈切换), 创建一个 newg, 同时进行栈参数拷贝, 同时设置 newg.sched (调整sp, pc指向 goexit, g), 以及 newg 其它相关变量, 
完成后 newg 的状态是 _Grunnable.

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
- 第三个参数 narg 是 fn 函数以字节为单位的大小;
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
    acquirem() // 增加当前 m.locks 值
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
    
    // 从 _p_.gFree 或 sched.gFree 当中获取一个 g. 如果获取到了 g, 则其已经分配好了 stack
    // 注: 如果 _p_.gFree 为空, 则先从 sched.gFree 当中获取一个 g 放入到 _p_.gFree, 然后从 _p_.gFree 当中获取.
    newg := gfget(_p_)
    if newg == nil {
        // new 一个g, 然后从堆上为其分配栈, 并设置 g 的 stack 成员和两个 stackguard 成员
        newg = malg(_StackMin) // 2k 栈大小
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
    
    // 调整 newg 的栈顶指针. 以下是在 amd64 架构下:
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
        // 把参数从 newproc 函数的栈(初始化是g0栈)拷贝到新的 newg 的栈. "汇编实现".
        // argp 是第一个参数位置. spArg 是预分配栈的低地址.
        // dst: [spArg, spArg+narg), src: [argp, argp+narg)
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
    
    // 清空 newg.sched 里面的字段, 然后重新设置 sched 对应的值. 
    memclrNoHeapPointers(unsafe.Pointer(&newg.sched), unsafe.Sizeof(newg.sched))
    newg.sched.sp = sp 
    newg.stktopsp = sp 
    
    // newg.sched.pc 表示当 newg 被调度起来运行时从这个地址开始执行指令.
    // 把 pc 设置成 goexit 函数偏移 1 (sys.PCQuantum是1) 的位置.
    // 要了解为什么这样做, 需要看后面的详细分析的. 
    newg.sched.pc = funcPC(goexit) + sys.PCQuantum // funcPC 用于获取函数地址
    newg.sched.g = guintptr(unsafe.Pointer(newg))
    
    // 重新调整 sched 成员和 newg 的栈(参考下面的分析)
    gostartcallfn(&newg.sched, fn)
    
    // traceback
    newg.gopc = callerpc
    newg.ancestors = saveAncestors(callergp)
    
    // 设置 newg 的 startpc 为 fn.fn, 该成员主要用于函数调用栈的 traceback 和栈收缩.
    // newg 真正从哪里执行不依赖此成员, 而是 sched.pc
    newg.startpc = fn.fn
    if _g_.m.curg != nil {
        newg.labels = _g_.m.curg.labels
    }
    if isSystemGoroutine(newg, false) {
        atomic.Xadd(&sched.ngsys, +1) // 系统 goroutine 统计
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

gostartcallfn(), 在 newg.sched.pc 设置为 `funcPC(goexit) + 1` 之后进行调用, 在这个函数当中可以找到为何这样做的原因.

这里需要注意, gobuf 的 sp 和 pc, g 成员变量已经设置. 且 `gobuf.pc=funcPC(goexit) + 1`, `sp` 指向的位置是要调用
函数的第一个参数位置处(参数已经拷贝到新的g的栈上).

```cgo
// fn 是 goroutine 的入口地址, 在初始化的时候对应是是 runtime.main, 如果是用户创建的函数, 则是函数地址.
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
    // sp 是 newg 的栈顶, 目前 newg 栈上拷贝了 fn 函数的参数, sp 指向的是 fn 的第一个参数.
    sp := buf.sp 
    if sys.RegSize > sys.PtrSize {
        sp -= sys.PtrSize
        *(*uintptr)(unsafe.Pointer(sp)) = 0
    }
    
    // 为返回地址预留空间, 然后将返回地址写入当前预留的位置. 注意: 这里是 pc 是 "funcPC(goexit)+1",
    // 也就说, 函数执行完成之后, 会跳到 goexit 函数当中.
    sp -= sys.PtrSize 
    *(*uintptr)(unsafe.Pointer(sp)) = buf.pc
    
    // 重新设置 newg 的栈顶寄存器.
    buf.sp = sp 
    
    // 这里才正在让 newg 的 ip 寄存器指向 fn 函数. 这里只是设置 newg 的信息, newg 还未执行,
    // 等到 newg 被调度起来之后, 调度器会把 buf.pc 放入到 CPU 的 ip 寄存器, 从而使得 cpu 真正执行起来.
    buf.pc = uintptr(fn)
    buf.ctxt = ctxt // fv 地址
}
```

gostartcallfn() 所做的事情:

- 获取 fn 函数地址

- 调用 gostartcall 函数, 调整 newg 的 sp (将 goexit() 的第二条指令地址入栈, 伪造成 goexit 函数调用了 fn, 从而使
fn 执行完成后 ret 指令返回到 goexit 继续执行完成清理工作) 和 pc (重新设置 pc 为需要执行的函数的地址, 即fn, 当前场景
为 runtime.main 函数地址) 的值.

- 重新设置 newg.sched.pc 指向的是 runtime.main() 的第一条指令, 重新设置 newg.sched.sp 指向是的 newg 栈顶单元, 
该单元保存了 runtime.main 函数执行完成之后的返回地址.


小结:

在调用 newproc 函数之后, runtime.main 对应的 newg 已经创建完成, 并且加入到当前线程绑定的 p 当中, 等待被调度执行.

newg 放入当前线程绑定的 p 对象的本地运行队列, 它是第一个真正意义上的用户 goroutine

**此时 newg 的 m 成员是 nil, 因为它还没有被调度起来运行, 因此没有与 m 绑定.**

> 注: 创建 newg 是在 m.g0 栈上, 并且最终被加入到 m.p 上. 此时的 newg 不绑定任何 m, 只有当 newg 被调度起来执行的时
候, 才会绑定一个新的 m (不一定是当前的m).

### runtime.mstart(SB) 调度循环开始

到目前为止, runtime.main goroutine已经创建, 并放到了 m0 线程绑定的 `allp[0]` 的本地队列当中, 接下来就需要启动调度
循环, 开始去查找 goroutine 并执行. 

汇编代码从 `newproc` 返回之后, 开始执行 mstart().

mstart() 是新线程的起点(通常是在 g0 栈上), 启动调度循环, 调用链: `mstart()` -> `mstart1()` -> `schedule()`

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
    // Initialize stack guard so that we can start calling regular Go code.
    // 在 linux amd64 下, _StackGuard=928
    _g_.stackguard0 = _g_.stack.lo + _StackGuard
    // This is the g0, so we can also call go:systemstack
    // functions, which check stackguard1.
    _g_.stackguard1 = _g_.stackguard0
    
    mstart1() // 真正启动 m 的函数
    
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


func mstart1() {
    // 在启动 m 的时, _g_ 必须是在 g0 上执行.
    _g_ := getg()
    if _g_ != _g_.m.g0 {
        throw("bad runtime·mstart")
    }
    
    // getcallerpc() 获取调用 mstart1 执行完的返回地址(即在 mstart() 函数当中的 "switch" 部分的地址).
    // getcallersp() 获取调用 mstart1 时的栈顶地址.
    // 由于即将调度执行用户的 g, 则需要将 g0 栈上的状态信息进行保存.
    save(getcallerpc(), getcallersp())
    asminit() // amd64 Linux 是空函数
    minit() // 信号相关初始化, 这个点与 start() 最后调用 sigprocmask() 是呼应的.
    
    // 如果当前运行在 m0 上, 需要执行额外的 mstartm0()
    if _g_.m == &m0 {
        mstartm0() // 信号初始化
    }
    
    // 在这里 fn 为 nil. mstartfn 意为 m 启动前的执行函数. 
    // sysmon 系统 goroutine 就是阻塞在这里执行的.
    if fn := _g_.m.mstartfn; fn != nil {
        fn()
    }
    
    // 如果当前没有运行在 m0 上, 则将 m.nextp 与当前 m 进行绑定.
    if _g_.m != &m0 {
        acquirep(_g_.m.nextp.ptr())
        _g_.m.nextp = 0
    }
    
    // 执行调度函数
    schedule()
}
```

mstart1 首先调用 save() 函数来保存 g0 的调用信息, **save这一行代码非常关键, 是理解调度循环的关键点之一**. 

getcallerpc() 返回的的 mstart 调用 mstart1 时被 call 指令压栈的返回地址. (注释当中已经说明)

getcallersp() 返回的是调用 mstart1 函数之前 mstart 函数的栈顶地址.

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

save 函数保存了调度相关的所有信息, 最为重要的当前正在运行 g 的下一条指令地址和栈顶地址. 不管想 g0 还是其他的 g 来说这
些信息在调度过程中都是必不可少的.

到目前为止, 上述的 mstart() 是在 g0 上执行的, 所有的操作都是针对 g0 而言. 


为何 g0 已经执行到 mstart1() 而且还会继续调用其他函数, 但 g0 的调度信息中的 pc 和 sp 却要设置在 mstart() 中? 难道
下次切换到 g0 时需要从 mstart 的 `switch` 语句继续执行? 从 mstart 函数可以看到, switch 语句之后就要退出线程了!

save() 执行之后, 返回到 mstart1() 继续进行一些 m 相关的一些初始化, 完成这些初始化后则调用调度系统的核心 schedule(),
在 schedule() 当中完成 goroutine 的切换(g0 -> g), 并且每次调度 goroutine 都是从 schedule() 开始的.

```cgo
func schedule() {
    // _g_ 是工作线程 m 对于的 g0, 在初始化时是 m0.g0
    _g_ := getg()
    
    // m 不能被锁定.
    if _g_.m.locks != 0 {
        throw("schedule: holding locks")
    }
    
    // 当 m 与某个 lockedg 锁定, 则 stop 掉 m(其实就是休眠), m 与 m.p 解绑, 执行调度 lockedg
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
    // 获取 p, 并将抢占变量设置为 false
    pp := _g_.m.p.ptr()
    pp.preempt = false
    
    // 当前处于 GC 的 stopTheWorld 阶段. 这个时候需要将当前的 m 休眠掉
    if sched.gcwaiting != 0 {
        gcstopm() // 会调用 stopm(), 休眠阻塞
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
    
    // 获取一个可以进行执行的 g
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
    
    // 当前进入gc MarkWorker 模式
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
