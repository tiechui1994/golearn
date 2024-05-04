# runtime 协程锁 - gopark 与 goready

goroutine 休眠, `gopark`, `goparkunlock`(对 gopark 的进一步封装)

goroutine 唤醒, `goready` 

### gopark

gopark, 休眠当前的 goroutine, 切换 g 的状态; g 与 m 分离; 开启新一轮调度;

```cgo
func gopark(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, traceEv byte, traceskip int) {
    if reason != waitReasonSleep {
        checkTimeouts() // timeouts may expire while two goroutines keep the scheduler busy
    }
    // m 禁止抢占
    mp := acquirem()
    gp := mp.curg // 用户 g 堆栈
	
    // 读取当前 g 的状态, 进行状态判断. 此时的状态一般是 _Grunning
    status := readgstatus(gp)
    if status != _Grunning && status != _Gscanrunning {
        throw("gopark: bad g status")
    }

    mp.waitlock = lock       // unlockf 参数, 绑定到 m
    mp.waitunlockf = unlockf // 休眠前 callback 函数, 绑定到 m
    gp.waitreason = reason
    mp.waittraceev = traceEv
    mp.waittraceskip = traceskip
    releasem(mp)
    
    // 切换到 g0 栈上, 执行 park_m 函数. 后续不再切回 g, 因此执行的函数永不返回.
    mcall(park_m)
}

// 在 g0 上执行(回调)
func park_m(gp *g) {
    _g_ := getg()
    
    // 修改 gp 的状态
    casgstatus(gp, _Grunning, _Gwaiting)
    dropg() // curg 与 m 解绑
    
    // m 上绑定的回调函数. 这个函数是在休眠前执行的, 函数执行成功才能休眠, 
    // 否则当前的 g 重新执行
    if fn := _g_.m.waitunlockf; fn != nil {
        ok := fn(gp, _g_.m.waitlock)
        _g_.m.waitunlockf = nil
        _g_.m.waitlock = nil
        if !ok {
            if trace.enabled {
                traceGoUnpark(gp, 2)
            }
            casgstatus(gp, _Gwaiting, _Grunnable)
            // 在这个函数当中, 会重新绑定 m 与 g, 函数永不返回.
            execute(gp, true) 
        }
    }
    
    // 新一轮调度, 永不返回
    schedule()
}
```

// goparkunlock 是对 gopark 的封装. 特殊在于 lock 对象是线程锁
```cgo
func goparkunlock(lock *mutex, reason waitReason, traceEv byte, traceskip int) {
    gopark(parkunlock_c, unsafe.Pointer(lock), reason, traceEv, traceskip)
}

func parkunlock_c(gp *g, lock unsafe.Pointer) bool {
    unlock((*mutex)(lock))
    return true
}
```

### goready

goready, 唤醒一个 goroutine; 切换到 g0 上, 将 gp 的状态调整为 _Grunnable, 重新放回队列

```cgo
func goready(gp *g, traceskip int) {
    // 切换的 g0 栈上, 执行 ready 函数. 然后切回到 g 上.
    systemstack(func() {
        ready(gp, traceskip, true)
    })
}

func ready(gp *g, traceskip int, next bool) {
    status := readgstatus(gp)

    // Mark runnable.
    _g_ := getg()
    mp := acquirem() // 禁止 mp 的抢占
    if status&^_Gscan != _Gwaiting {
        dumpgstatus(gp)
        throw("bad g->status in ready")
    }

    // 切换 gp 的状态 _Gwaiting => _Grunable
    casgstatus(gp, _Gwaiting, _Grunnable)
    // 将 gp 放到 p 的运行队列当中, 后续进行调度执行
    runqput(_g_.m.p.ptr(), gp, next)
    wakep()
    releasem(mp)
}
```

