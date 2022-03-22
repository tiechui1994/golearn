# Mutex 实现原理

Mutex 可以处于2种操作模式: 正常和饥饿

在正常模式下, 等待着按 FIFO 顺序排队, 当唤醒一个等待队列中的 goroutine 时, 此 goroutine 并不会错直接获取到锁, 而是
会和新到达的 goroutine 竞争. 通常新到达的 goroutine 更容易获得锁(因为它已经在 CPU 上运行, 大概率可以直接执行到获取锁
的逻辑). 在这种情况下, 被唤醒的 goroutine 排在等待队列的前面. 

如果等待者超过1ms未能获得互斥锁, 它会将互斥锁切换到饥饿模式.

在饥饿模式下, 互斥锁的所有权直接从 goroutine 移交给队列前面的等待者. 而新到达的 goroutine 不会尝试获取互斥锁, 即使它
已解锁, 也不会尝试自旋. 相反, 它们将字节排在等待队列的尾部.

当获取锁的这个 goroutine "是队列中最后一个等待者", 或者"等待的时间不到1ms", 它会将互斥锁切换回正常操作模式.

正常模式的性能比较好, 因为即使有阻塞的等待者, goroutine 也可以连续多次获取互斥锁.

饥饿模式对于防止尾部 goroutine 延迟过长很重要.

![image](/images/develop_sync_mutex_state.png)

waiter_num: 记录当前等待抢占该锁 goroutine 数量.

starving: 当前锁**是否处于饥饿状态**

woken: 当前锁**是否有goroutine已经被唤醒**

locked: 当前锁**是否有goroutine持有**

sema 信号量作用:

当持有锁的 goroutine 释放锁后, 会释放 sema 信号量, 该信号量会唤醒之前抢占阻塞的 goroutine 来获取锁.

> locked 是版本1引进的.
> woken 是版本2引进的. 目的在于给新到达的 goroutine 机会. (被唤醒的goroutine与新到达的goroutine同等竞争)
> spin(自旋)是版本3引进的. 目的在于给新到达的 goroutine 更多机会. 在自旋过程中尝试去修改 woken 状态.
> starving 是版本4引进的. 目的在于解决饥饿问题. 如果被唤醒的 goroutine 在一定时间无法抢占到锁, 就直接在占有锁.

[第二版 mutex 设计](https://zhuanlan.zhihu.com/p/341887600), 这个版本当中增加了 worken, 让刚到达的 goroutine
与被唤醒的goroutine 同等竞争(正常状况下, 新到达的goroutine是需要排队获取锁, 并且刚唤醒的goroutine需要在获取到P才能
执行, 如此一来, 就相当于给了新到来goroutine更多的机会.)

[第三版 mutex 设计](https://zhuanlan.zhihu.com/p/342706674), 这个版本当中增加了自旋, 子旋的过程中尝试修改唤醒标
志, 再一次提升新到达goroutine获得锁的机会.

第四版本 mutex 设计如下, 解决的是队列当中 goroutine 长时间无法获得锁的情况(第二, 三版本导致的).

### Lock 逻辑

```cgo
func (m *Mutex) Lock() {
    if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
        return
    }
    // 缓慢之路, 尝试自旋竞争或饥饿状态下饥饿goroutine竞争
    m.lockSlow()
}

func (m *Mutex) lockSlow() {
    var waitStartTime int64
    starving := false 
    awoke := false 
    iter := 0 // 自旋次数
    old := m.state // 当前的锁的状态
    for {
        // 锁是非饥饿状态(正常状态), 尝试自旋. 版本3引进的内容
        if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
            // 没有唤醒的状况下, 尝试设置唤醒(因为当前自旋的 goroutine 可以看作是被唤醒的 goroutine)
            if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
                atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
                awoke = true
            }
            runtime_doSpin()
            iter++
            old = m.state // 再次获取锁的状态
            continue
        }
        
        // 状态转移. 
        new := old
        if old&mutexStarving == 0 {
            new |= mutexLocked // 正常状态, 确保加锁
        }
        if old&(mutexLocked|mutexStarving) != 0 {
            new += 1 << mutexWaiterShift // waiter数量加1
        }
        if starving && old&mutexLocked != 0 {
            new |= mutexStarving // 复制饥饿状态到新的状态.(饥饿持续)
        }
        
        if awoke {
            if new&mutexWoken == 0 {
                throw("sync: inconsistent mutex state")
            }
            new &^= mutexWoken // 当前已经有唤醒的 goroutine, 则新状态将唤醒状态去掉.
        }
        
        // 尝试切换
        if atomic.CompareAndSwapInt32(&m.state, old, new) {
            // 正常模式下成功获取锁. 
            // old & mutexLocked == 0 是成功获取到锁的基础 
            if old&(mutexLocked|mutexStarving) == 0 {
                break // locked the mutex with CAS
            }
            
            // waitStartTime 为0, 说明当前的 goroutine 是新到达的
            // waitStartTime 不为0, 说明当前的 goroutine 是唤醒的
            queueLifo := waitStartTime != 0
            if waitStartTime == 0 {
                waitStartTime = runtime_nanotime()
            }
            // 阻塞当前的 goroutine 等待
            // 当 queueLifo 为 true, 需要将该 goroutine 排在队列头
            runtime_SemacquireMutex(&m.sema, queueLifo, 1)
            // 唤醒之后, 重新设置 starving 状态
            starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs
            old = m.state
            // 被唤醒的 goroutine, 又处于饥饿状态, 直接获取
            if old&mutexStarving != 0 {
                if old&(mutexLocked|mutexWoken) != 0 || old>>mutexWaiterShift == 0 {
                    throw("sync: inconsistent mutex state")
                }
                
                // delta 是修改 state 的一个偏移量. 加锁, 减少等待者 
                delta := int32(mutexLocked - 1<<mutexWaiterShift)
                
                // 解除饥饿状态: 正常状况(获取锁的时间小于1ms) 或 最后一个等待者
                if !starving || old>>mutexWaiterShift == 1 {
                    delta -= mutexStarving // 清除饥饿标记
                }
                atomic.AddInt32(&m.state, delta) 
                break
            }
            
            // 正常状况下的唤醒
            awoke = true
            iter = 0
        } else {
            old = m.state
        }
    }
}
```

### Unlock 逻辑

```cgo
func (m *Mutex) Unlock() {
	if race.Enabled {
		_ = m.state
		race.Release(unsafe.Pointer(m))
	}

	// Fast path: drop lock bit.
	new := atomic.AddInt32(&m.state, -mutexLocked)
	if new != 0 {
		m.unlockSlow(new)
	}
}

func (m *Mutex) unlockSlow(new int32) {
	if (new+mutexLocked)&mutexLocked == 0 {
		throw("sync: unlock of unlocked mutex")
	}
	
	if new&mutexStarving == 0 {
	    // 正常状态下
		old := new
		for {
			// 如果等待的人数为0, 或有一个 goroutine 已经被唤醒, 或有一个 goroutine 抢占到了锁.
			// 在上述任意一种情况下, 不需要唤醒任何人.
			// 在饥饿模式下, 锁的所有权直接从解锁的 goroutine 移交给下一个等待的 goroutine.
			// 由于当前是正常模式, 因此解锁之后, 并不能直接移交 goroutine
			if old>>mutexWaiterShift == 0 || old&(mutexLocked|mutexWoken|mutexStarving) != 0 {
				return
			}
			// 唤醒某个等待的 goroutine.
			new = (old - 1<<mutexWaiterShift) | mutexWoken // 因为是唤醒, 因此必须设置 mutexWoken
			if atomic.CompareAndSwapInt32(&m.state, old, new) {
				runtime_Semrelease(&m.sema, false, 1) // 第二参数表示能否直接移交锁的所有权.
				return
			}
			old = m.state
		}
	} else {
	    // 饥饿状态下, 直接将锁移交给下一个 waiter, 并让出自身的时间片, 以便下一个 waiter 
	    // 可以立即开始运行.
	    // 注: mutexLocked 并没有设置, waiter 被唤醒后自己设置. 但是如果设置了 mutexStarving,
	    // mutex 仍然被认为是锁定的, 因此新到达的 goroutine 不会获取它.
		runtime_Semrelease(&m.sema, true, 1)
	}
}
```

