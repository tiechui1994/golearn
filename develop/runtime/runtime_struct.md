### GMP数据结构

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