## RWMutex 实现

Lock/Unlock: 写操作时调用的方法. 如果锁被读协程或者写协程持有, Lock 会一直阻塞, 直到 "全部的读锁" 或 "写锁" 被释放为
止; Unlock 是对应的解锁操作.
   
RLock/RUnlock: 读操作调用的方法. 如果锁被写协程持有, RLock 会一直阻塞, 直到写协程释放锁; 如果被读进程持有, 会立即生
效; RUnlock 是对应的解锁操作.


```cgo
type RWMutex struct {
  w           Mutex   // 互斥锁解决多个writer的竞争
  writerSem   uint32  // writer信号量
  readerSem   uint32  // reader信号量
  readerCount int32   // reader的数量
  readerWait  int32   // writer等待完成的reader的数量
}
```

readerCount, 记录当前读 goroutine 数量, 以及是否有写 goroutine 在等待.

readerWait, 写 goroutine 请求写锁时, 需要等待完成读 goroutine 的数量.


需要将 RLock, RUnLock, Lock, UnLock 当做一个有机整体进行解读. 基本上就是四条:

- RLock 时, 如果有写锁, 则需要休眠. 

- Lock 时, 如果有读锁, 则需要休眠. 同时 readerWait = rederCount. 在这里将 readerCount 设置为负数.

- RUnLock 时, 如果 readerWait == 1, 则需要唤醒.

- UnLock 时, 如果 readerWait > 0, 则需要依次唤醒. 在这里将 readerCount 重置为正数.

```cgo
// 不能用于递归读锁定; 阻塞的 Lock 调用会阻止新的 goroutine 获取到读锁.
func (rw *RWMutex) RLock() {
    // readerCount 为负值时, 意味着有 goroutine 请求写锁, 因为写锁优先级更高, 
    // 需要将后续请求读锁的 goroutine 休眠
    if atomic.AddInt32(&rw.readerCount, 1) < 0 {
        // A writer is pending, wait for it.
        runtime_SemacquireMutex(&rw.readerSem, false, 0)
    }
}
```

```cgo
// 释放读锁. 它不会影响其他的读锁. 
// 在没有获取读锁的状况下释放读锁, 将产生一个运行时错误.
func (rw *RWMutex) RUnlock() {
    // rederCount 减一操作.
    // 如果 readerCount 为0, 意味着在没有获得读锁的状况下释放读锁, 将产生一个错误.
    // 如果 readerCount 为负数, 意味着有 goroutine 请求写锁. 不影响当前的操作
    if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {
        // Outlined slow-path to allow the fast-path to be inlined
        rw.rUnlockSlow(r)
    }
}

func (rw *RWMutex) rUnlockSlow(r int32) {
    if r+1 == 0 || r+1 == -rwmutexMaxReaders {
        race.Enable()
        throw("sync: RUnlock of unlocked RWMutex")
    }
    // readerWait 是当请求写锁时, 休眠的请求读锁的 goroutine 数量.
    // 如果 readerWait 为 1, 则需要唤醒休眠的请求读锁的 goroutine, 告知它们可以继续获取读锁了.
    if atomic.AddInt32(&rw.readerWait, -1) == 0 {
        // The last reader unblocks the writer.
        runtime_Semrelease(&rw.writerSem, false, 1)
    }
}
```


```cgo
func (rw *RWMutex) Lock() {
    // 解决写锁竞争的问题
    rw.w.Lock()
	
    // Announce to readers there is a pending writer.
    // 让 readerCount 变为负数, 标志者当前处于写锁状态.(存在 goroutine 在获取写锁)
    r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
    // 如果有读锁, 那么需要等待, 直到这些读锁全部释放
    // 此时需要 readerCount 的值赋值给 readerWait. 因为等待的释放 goroutine 的数量是 readerCount
    if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
        runtime_SemacquireMutex(&rw.writerSem, false, 0)
    }
}
```


```cgo
// 如果在没有获取到写锁的状况下释放写锁, 将产生一个运行时错误.
//
// 与 Mutex 一样, 锁定的 RWMutex 与特定的 goroutine 无关. 
// 一个 goroutine 可以 RLock (Lock) 一个 RWMutex, 然后安排另一个 goroutine 对它进行 RUnlock( UnLock ).
func (rw *RWMutex) Unlock() {
    // 恢复 readerCount 的数量. (因为在获取到写锁的时候, 该值被被设置为负数)
    // readerCount 为正数, 也表明当前已经没有了写锁.
    r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)
    if r >= rwmutexMaxReaders {
        race.Enable()
        throw("sync: Unlock of unlocked RWMutex")
    }
	
    // 唤醒所有的读锁的请求者. 因为在 readerCount 为负数的时候, 将所有的读锁请求者全部休眠
    for i := 0; i < int(r); i++ {
        runtime_Semrelease(&rw.readerSem, false, 0)
    }
    // 释放写锁
    rw.w.Unlock()
}
```

