## go 调度(三) - M 是如何创建的? g0 是什么时候创建的

启动一个新的系统线程 M, 调用函数 startm, 该函数的所做的事情不是很复杂:

首先, 获取一个空闲的 p (当参数的 p 为nil时, 则从 sched.pidle 当中获取一个), 如果获取到 p, 停止启动新的系统线程 M.

其次, 从 sched.midle 当中获取一个空闲的 m, 如果不存在, 则调用 newm() 创建一个, 设置自旋(可选), 完成启动新的系统线程M.

最后, 当 sched.midle 当中获取到了休眠的 m, 将 `_p_` 绑定到 m.nextp, 然后调用 notewakeup() 函数唤醒休眠的 m

> 调用 startm 的函数. handoffp(), sysmon(), schedEnableUser()

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

> notewakeup, 在 `runtime_schedule.md` 当中已经有详细讲解.

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
    
    // 对于已经创建的 mp 对象, 会被添加对象, 则被添加到 newmHandoff.new 队列当中. 
    // gp.m.incgo, 由 C 启动的线程.
    // gp.m.lockedExt, 当前线程是否锁定.
    if gp := getg(); gp != nil && gp.m != nil && (gp.m.lockedExt != 0 || gp.m.incgo) && GOOS != "plan9" {
        // 当前处于锁定的 m 或 可能由C启动的线程上. 此线程的内核状态可能很奇怪(用户可能已为此目的将其锁定).
        // 我们不想将其克隆到另一个线程中. 而是要求一个已知良好的线程为我们创建线程.    
        lock(&newmHandoff.lock)
        if newmHandoff.haveTemplateThread == 0 {
            throw("on a locked thread with no template thread")
        }
        // 将 mp 添加到 newmHandoff.newm 链表当中
        mp.schedlink = newmHandoff.newm
        newmHandoff.newm.set(mp)
        // newmHandoff 处于等待当中, 则唤醒之
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
        asmcgocall(_cgo_thread_start, unsafe.Pointer(&ts)) // 执行 _cgo_thread_start 函数, 参数是 ts
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
    
    // 在进行系统线程 clone 期间中禁用信号, 这样新线程可以从禁用的信号开始, 之后它将调用 minit 
    var oset sigset
    sigprocmask(_SIG_SETMASK, &sigset_all, &oset)
    // 系统调用 clone 函数, 克隆一个新的系统线程
    // 第一个参数 cloneFlags 克隆的 flags, 是否共享内存等
    // 第二个参数 stk, 设置新线程的栈顶
    // 第三个参数是 mp
    // 第四个参数是 g0
    // 第五个参数是新线程启动之后开始执行的函数地址. 这里设置的是 mstart, 在程序启动时, 最后一步调用的也是 mstart 函数.
    // 注: mp, g0 复制到新线程当中, 并赋值到 tls 当中.
    ret := clone(cloneFlags, stk, unsafe.Pointer(mp), unsafe.Pointer(mp.g0), unsafe.Pointer(funcPC(mstart)))
    sigprocmask(_SIG_SETMASK, &oset, nil) // 启用信号
}
```

> clone() 是汇编实现(系统调用 clone). 
> sigprocmask() 更改当前阻塞信号的列表

分配 m 对象: (同时还创建了 g0, g0 栈)

```cgo
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


> mstart(), M启动. 调用的地方: clone 之后, asmcgocall 之后. runtime·rt0_go 该函数在 `runtime_bootstrap.md` 
当中有详解.

systemstack() 函数: 切换到 g0 栈上, 执行函数 fn

```cgo
// func systemstack(fn func())
TEXT runtime·systemstack(SB), NOSPLIT, $0-8
    MOVQ    fn+0(FP), DI    // DI = fn
    get_tls(CX)
    MOVQ    g(CX), AX    // AX = g
    MOVQ    g_m(AX), BX    // BX = g.m, 当前工作线程 m
    
    CMPQ    AX, m_gsignal(BX) // g == m.gsignal
    JEQ    noswitch // 相等跳转到 noswitch
    
    MOVQ    m_g0(BX), DX // DX = m.g0
    CMPQ    AX, DX // g == m.g0
    JEQ    noswitch // 相等则跳转 noswitch, 当前在 g0 栈上
    
    CMPQ    AX, m_curg(BX) // g == m.curg
    JNE    bad // 不相等, 程序异常
    
    // stack 切换, 从 curg 切换到 g0
    // 将 curg 上下文保存到 sched 当中
    MOVQ    $runtime·systemstack_switch(SB), SI // SI=runtime·systemstack_switch, 空函数地址
    MOVQ    SI, (g_sched+gobuf_pc)(AX) // g.sched.pc=SI
    MOVQ    SP, (g_sched+gobuf_sp)(AX) // g.sched.sp=SP
    MOVQ    AX, (g_sched+gobuf_g)(AX)  // g.sched.g=g
    MOVQ    BP, (g_sched+gobuf_bp)(AX) // g.sched.bp=BP
    
    MOVQ    DX, g(CX) // 切换 tls 到 g0
    MOVQ    (g_sched+gobuf_sp)(DX), BX // BX=g0.sched.sp
    
    // 栈调整, 伪装成 mstart() 调用函数 systemstack(), 目的是停止追踪
    SUBQ    $8, BX
    MOVQ    $runtime·mstart(SB), DX // DX=runtime·mstart
    MOVQ    DX, 0(BX) // 将 runtime·mstart 函数地址入栈
    MOVQ    BX, SP // 调整当前的 SP 
    
    // 调用 target 函数
    MOVQ    DI, DX   // DX=fn 
    MOVQ    0(DI), DI // 判断 fn 非空
    CALL    DI // 函数调用, 没有参数和返回值
    
    // 函数调用完成, 切换到 curg 栈上
    get_tls(CX)
    MOVQ    g(CX), AX   
    MOVQ    g_m(AX), BX // BX=m
    MOVQ    m_curg(BX), AX // AX=m.curg
    MOVQ    AX, g(CX) // 设置本地保存 m.curg 
    MOVQ    (g_sched+gobuf_sp)(AX), SP // SP = m.curg.sched.sp
    MOVQ    $0, (g_sched+gobuf_sp)(AX) // m.curg.sched.sp = 0
    RET

noswitch:
    // 当前已经在 g0 栈上了, 直接跳跃
    MOVQ    DI, DX
    MOVQ    0(DI), DI
    JMP    DI

bad:
    // Bad: g is not gsignal, not g0, not curg. What is it?
    MOVQ    $runtime·badsystemstack(SB), AX
    CALL    AX
    INT    $3
```
