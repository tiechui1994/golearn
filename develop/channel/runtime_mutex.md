# runtime 线程锁 - runtime.mutex

Go 的 runtime 中封装了 mutex, 这个 mutex 在 channel, netpoll, 检测活跃的定时器等使用.

sync.Mutex vs runtime.mutex

sync.Mutex 是用户层的锁, Lock 失败的 goroutine 被阻塞(调用 gopark).

runtime.mutex 是 runtime 使用的锁, Lock 抢锁失败, 会造成 m 阻塞(线程阻塞, 底层调用 futex)

### 基础内容

futex(Fast Userspace Mutexes), Linux 下实现锁定和构建高级抽象锁如信号量和POSIX互斥的基本工具.

futex 由一块能够被多个进程共享的内存空间(一个对齐后的整型变量)组成; 这个整型变量的值能够通过汇编语言调用 CPU 提供的
原子操作指令来增加或减少, 并且一个进程可以等待直到那个值变成正数. futex 的操作几乎是在用户空间完成, 只有当操作结果不
一致需要仲裁时, 才需要进入操作系统内核空间执行. 

futex 的基本思想是竞争态总是很少发生的, 只有在竞争态才需要进入内核, 否则在用户态即可完成.

```cgo
/*
 * val: 0, unlock
 * val: 1, lock, no waiters
 * val: 2, one or more waiters
 */

int val = 0;

void lock() {
    int c;
    if ((c = cmpxchg(val, 0, 1)) != 0) {
        if ( c != 2 ) {
            c = xchg(val, 2);
        }
        
        while ( c != 0 ) {
            fmutex_wait(&val, 2);
            c = xchg(val, 2);
        }
    }
}

void unlick() {
    if ( atomic_dec(val) != 1 ) {
        val = 0;
        fmutex_wake(&val, 1);
    }
}

# val 代表期待的值, 当 *addr == val, 才会进行 wait (线程阻塞)
fmutex_wait(int* addr, int val);

# 唤醒 n 个阻塞在 addr 指向的锁变量上挂起等待的进程
fmutex_wake(int* addr, int n);
```

### runtime 实现的阻塞等待 procyield, osyield

// procyield, 执行 n(传入的参数) 次 `PAUSE` 指令 (x86指令集, 主要用于提高自旋的性能). 
```
TEXT runtime.procyiled(SB), NOSPLIT,$0-0
    MOVL cycles+0(FP), AX
again:
    PAUSE
    SUBL $1, AX
    JNZ again
    RET
```

// osyield, 主要是系统调用 (AX=24, sched_yield), 让当前的线程放弃 CPU 执行权限, 把线程移到队列尾部, 优先执行其
他的线程. 与 runtime.Gosched 类似.
```
TEXT runtime.osyield(SB), NOSPLIT,$0
    MOVL $SYS_sched_yield, AX
    SYSCALL
    RET
```

Go 的 `futexsleep` 和 `fmutexwake` 是对 `futex` 的包装, 线程级别的锁, 实现如下:

```
// 如果 *addr == val { 当前线程进入sleep状态 }; 不会阻塞超过ns, ns<0表示永远休眠
futexsleep(addr *uint32, val uint32, ns int64)

// 如果任何线程阻塞在addr上, 则唤醒至少cnt个阻塞的任务
futexwakeup(addr *uint32, cnt uint32) 
```

// runtime.futex 系统调用实现
```
// int64 futex(int32 *uaddr, int32 op, int32 val,
//    struct timespec *timeout, int32 *uaddr2, int32 val2);
// op: 代表操作, wait: 128, wake: 129
TEXT runtime·futex(SB),NOSPLIT,$0
    MOVQ  addr+0(FP), DI
    MOVL  op+8(FP), SI
    MOVL  val+12(FP), DX
    MOVQ  ts+16(FP), R10
    MOVQ  addr2+24(FP), R8
    MOVL  val3+32(FP), R9
    MOVL  $SYS_futex, AX
    SYSCALL
    MOVL  AX, ret+40(FP)
    RET
```

// 线程 sleep, wakeup 实现原理 
```cgo
// Atomically,
//    if(*addr == val) sleep
// 可能被误唤醒; 这是允许的.
// 睡眠时间不要超过 ns; ns < 0 表示永远.
func futexsleep(addr *uint32, val uint32, ns int64) {
    // 某些 Linux 内核存在bug, 即 FUTEX_WAIT 的 futex 返回内部错误代码作为 errno.
    // Libpthread 忽略这里的返回值, 我们也可以: 正如它所说的几行, 允许虚假唤醒.
    if ns < 0 {
        futex(unsafe.Pointer(addr), _FUTEX_WAIT_PRIVATE, val, nil, nil, 0)
        return
    }

    var ts timespec
    ts.setNsec(ns)
    futex(unsafe.Pointer(addr), _FUTEX_WAIT_PRIVATE, val, unsafe.Pointer(&ts), nil, 0)
}

// If any procs are sleeping on addr, wake up at most cnt.
func futexwakeup(addr *uint32, cnt uint32) {
    ret := futex(unsafe.Pointer(addr), _FUTEX_WAKE_PRIVATE, cnt, nil, nil, 0)
    if ret >= 0 {
        return
    }

    systemstack(func() {
        print("futexwakeup addr=", addr, " returned ", ret, "\n")
    })

    *(*int32)(unsafe.Pointer(uintptr(0x1006))) = 0x1006
}
```

### runtime.mutex 实现

数据结构, 定义在 runtime/runtime2.go

```
type mutex struct {
    // Empty struct if lock ranking is disabled, otherwise includes the lock rank
    lockRankStruct
    // Futex-based impl treats it as uint32 key,
    // while sema-based impl as M* waitm.
    // Used to be a union, but unions break precise GC.
    key uintptr
}
```

lockRankStruct 是 runtime 死锁检测用的, 只有开启特定设置才会有具体实现, 否则是一个空struct(空struct只要不是最后
一个字段是不占用任何空间的). 不用关注.

lock 实现: CPU 自旋 + CPU 调度 + futexsleep 

```cgo
func lock(l *mutex) {
    lockWithRank(l, getLockRank(l))
}

func lockWithRank(l *mutex, rank lockRank) {
    lock2(l)
}

func lock2(l *mutex) {
    // 获取当前的 g, 将 m.locks 增加(只要 m.locks !=0, 就不能抢占)
    gp := getg()
    if gp.m.locks < 0 {
        throw("runtime·lock: lock count")
    }
    gp.m.locks++
    
    // l.key 只有三种状态 mutex_unlocked, mutex_locked, mutex_sleeping
    // 分别代表无锁状态, 正常加锁状态, 有线程调用 futexsleep 阻塞状态.
    // 设置状态为 mutex_unlocked, 这里是 Xchg, 是交换(设置), 返回的是旧的值
    v := atomic.Xchg(key32(&l.key), mutex_locked)
    if v == mutex_unlocked { // l.key 之前的状态是 mutex_unlocked, 加锁成功
        return
    }

    // 这里的 v 不是 mutex_unlocked, 因此只能是 mutex_locked 或 mutex_sleeping
    // 因此 wait 可能是 mutex_locked 或 mutex_sleeping
    // 注意: 如果将 l.key 从 mutex_sleeping 更改为其他值, 必须小心在返回之前将其改回 mutex_sleeping
    wait := v

    // 多核情况下尝试自旋4次, 单个就不用自旋了
    spin := 0
    if ncpu > 1 {
        spin = active_spin
    }
    for {
        // 自旋, 尝试获取锁
        for i := 0; i < spin; i++ {
            // 注意, 前面已经将 l.key 设置为 mutex_locked, 这里的 l.key == mutex_unlocked, 
            // 说明其他持有锁的线程进行了释放. 
            for l.key == mutex_unlocked {
                // CAS 抢锁成功直接返回
                if atomic.Cas(key32(&l.key), mutex_unlocked, wait) {
                    return
                }
            }
            // CPU 自旋, 执行 active_spin_cnt=30 次 PAUSE 指令
            procyield(active_spin_cnt)
        }
        
        // passive_spin=1, 再尝试抢一次锁
        for i := 0; i < passive_spin; i++ {
            for l.key == mutex_unlocked {
                if atomic.Cas(key32(&l.key), mutex_unlocked, wait) {
                    return
                }
            }
            osyield() // 系统调用 sched_yield, 让出 cpu
        }
        
        // 设置 l.key 为 mutex_sleeping
        v = atomic.Xchg(key32(&l.key), mutex_sleeping)
        if v == mutex_unlocked {
            // 注意, 这里是从 mutex_unlocked => mutex_sleeping 也认为加锁成功, 直接返回.
            // 这里没有走 futexsleep 阻塞当中的线程, 造成的影响是解锁时执行 fmutexwake 没有唤醒的线程.
            return
        }
        
        // 设置 wait 的状态为 mutex_sleeping
        wait = mutex_sleeping
        // 如果 l.key == mutext_sleeping 就进行休眠, 直到被唤醒或l.key的值发生变更.
        futexsleep(key32(&l.key), mutex_sleeping, -1)
    }
}
```

主要步骤:
1) 调用 `atomic.Xchg` 设置 l.key 的值为 mutex_locked.

2) 根据 `atomic.Xchg` 返回值 `v` 进行判断,  v == mutex_unlocked, 原来的状态是 mutex_unlocked, 加锁成功, 直
接返回. 否则加锁失败, `v` 可能的 mutex_locked 或 mutex_sleeping

3) 在多核状况下, 进行4次自旋, 在自旋过程中发现 l.key == mutex_unlocked, 表示其他线程对象释放了锁, 使用 CAS 抢占
锁, 抢占成功, 直接返回. 失败的话, 调用 `procyied` 让出 CPU 执行权限(并不切换)

4) 再尝试一次 `CAS`, 失败的话调用 `osyield` 让出 CPU.

5) `osyield` 完成后, 继续执行, 这个时候调用 `atomic.Xchg` 设置 `l.key = mutex_sleeping` 表示当前准备调用
`futexsleep` 进行 `sleep`.

6) 使用系统调用 `futexsleep`, 如果 l.key == mutex_sleeping, 则当前线程进入休眠状态， 直到有其他地方调用 `futexwakeup` 
来唤醒. 如果这个时候 `l.key != mutex_sleeping`, 说明在步骤5设置完这短短时间内, 其他线程设置又重新设置了 l.key 
状态比如设置为了 mutex_locked 或者 mutex_unlocked. 这个时候不会进入sleep, 而是会去循环执行步骤3


unlock 实现: futexwakeup 

```cgo
func unlock(l *mutex) {
    unlockWithRank(l)
}

func unlockWithRank(l *mutex) {
    unlock2(l)
}

func unlock2(l *mutex) {
    // 设置 l.key = mutex_unlocked
    v := atomic.Xchg(key32(&l.key), mutex_unlocked)
    if v == mutex_unlocked { // 重复调用 unlock, 直接抛出异常.
        throw("unlock of unlocked lock")
    }
    
    // 之前的状态是 mutex_sleeping, 说明其他有线程在"sleep", 唤醒一个 sleep 的线程.
    // 注意: 这里的唤醒的队列可能的空的
    if v == mutex_sleeping { 
        futexwakeup(key32(&l.key), 1)
    }

    // m.locks 减少.
    gp := getg()
    gp.m.locks--
    if gp.m.locks < 0 {
        throw("runtime·unlock: lock count")
    }
    if gp.m.locks == 0 && gp.preempt { 
        gp.stackguard0 = stackPreempt
    }
}
```
 
        