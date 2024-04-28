# sync.WaitGroup

WaitGroup 的设计初衷是多个任务并发执行, 一起结束.

> 注: WaitGroup 对象不能被 copy 传递, 也就是说其非指针对象不能作为参数跨函数传递. 

## 数据结构

WaitGroup 是基于信号量(PV操作)实现的. 在 state1 当中存储了 sema 的值.

```
type WaitGroup struct {
    noCopy noCopy // 如果被 copy 复制, 该数据会发生变更.

    // 64-bit value: 高32位是计数, 低32位是waiter数量
    // 最后4字节存储信号量 sema
    state1 [3]uint32
}
```

sema 对应于 golang 中 runtime 内部的信号量的实现. WaitGroup 中会用到 sema 的两个相关函数, runtime_Semacquire 
和 runtime_Semrelease. runtime_Semacquire 表示增加一个信号量, 并挂起当前 goroutine. runtime_Semrelease 表
示减少一个信号量, 并唤醒 sema 上其中一个正在等待的 goroutine.


// state 计算, 获取 state(counter, waiter) 和 sema 
```cgo
// 这里主要使用到来 OS 的内存地址字节对齐.
func (wg *WaitGroup) state() (statep *uint64, semap *uint32) {
    if uintptr(unsafe.Pointer(&wg.state1))%8 == 0 {
        // 注: 因为这里的 state1 是数组, 其数组第一个元素的地址与数组的地址是一致的.
        // 如果 state1 是切片, 切片的第一个元素的地址与切片的地址是不一致的.
        return (*uint64)(unsafe.Pointer(&wg.state1)), &wg.state1[2]
    } else {
        return (*uint64)(unsafe.Pointer(&wg.state1[1])), &wg.state1[0]
    }
}
```


// Add, Done 操作.
```cgo
// Add 方法将 delta 值加上计数器, delta 可以为负数. 
// 如果计数器变为 0,则所有在 Wait 上阻塞的 Goroutine 都会被释放.
// 如果计数器变为负数，则 Add 方法会 panic.
//
// 注意: 当计数器为 0 时调用 delta 值为正数的 Add 方法必须在 Wait 方法之前执行.
// 而 delta 值为负数或者 delta 值为正数但计数器大于 0 时, 则可以在任何时间点执行
func (wg *WaitGroup) Add(delta int) {
    statep, semap := wg.state()
    state := atomic.AddUint64(statep, uint64(delta)<<32)
    v := int32(state >> 32)
    w := uint32(state)
    if v < 0 {
        panic("sync: negative WaitGroup counter")
    }
    // Wait, Add 并发调用
    if w != 0 && delta > 0 && v == int32(delta) {
        panic("sync: WaitGroup misuse: Add called concurrently with Wait")
    }
    if v > 0 || w == 0 {
        return
    }
    
    // 当 waiters > 0 时, 这个 Goroutine 将 counter 设置为 0.
    // 现在不可能有对状态的并发修改:
    // - Add 方法不能与 Wait 方法同时执行.
    // - Wait 不会在看到 counter 为 0 时增加等待者.
    // state 一致性检测
    if *statep != state {
        panic("sync: WaitGroup misuse: Add called concurrently with Wait")
    }
    // Reset waiters count to 0.
    *statep = 0
    for ; w != 0; w-- {
        runtime_Semrelease(semap, false, 0)
    }
}

// Done decrements the WaitGroup counter by one.
func (wg *WaitGroup) Done() {
    wg.Add(-1)
}
```

```cgo
// Wait blocks until the WaitGroup counter is zero.
func (wg *WaitGroup) Wait() {
    statep, semap := wg.state()
    for {
        state := atomic.LoadUint64(statep)
        v := int32(state >> 32)
        w := uint32(state)
        if v == 0 {
            // Counter is 0, no need to wait.
            return
        }
        // Increment waiters count.
        if atomic.CompareAndSwapUint64(statep, state, state+1) {
            runtime_Semacquire(semap)
            if *statep != 0 {
                panic("sync: WaitGroup is reused before previous Wait has returned")
            }
            return
        }
    }
}
```