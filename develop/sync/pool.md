# pool 原理

### 一种巧妙的双端队列

针对一个 uint64 的数值, 组装和拆除为两个 uint32 的值. 原理很简单: 高32位, 低32位.

```cgo
const dequeueBits = 32

func (d *poolDequeue) unpack(ptrs uint64) (head, tail uint32) {
	const mask = 1<<dequeueBits - 1
	head = uint32((ptrs >> dequeueBits) & mask)
	tail = uint32(ptrs & mask)
	return
}

func (d *poolDequeue) pack(head, tail uint32) uint64 {
	const mask = 1<<dequeueBits - 1
	return (uint64(head) << dequeueBits) |
		uint64(tail&mask)
}
```


双端队列的设计:

```cgo
type poolDequeue struct {
	// 保存 head, tail 的索引位置
	// tail = oldest data 位置
	// head = 下一个要填充的 slot 位置
	// 在 [tail, head) 区间的 slot 归 consumer 所有.
	// consumer 继续拥有这个范围之外的 slot, 直到它把 slot 设置为 nil, 此时所有权传递给producer.
	headTail uint64

    // vals 是存储在队列中的 interface{} 值的环形缓冲区.  
    // 它的大小必须是 2 的幂.
    //
    // 如果 slot 为空, 则 vals[i].typ 为 nil, 否则为非 nil.
    // 一个 slot 在使用中, 直到 tail index 超出 head index, 并且 typ 已设置为 nil.
    // 由 consumer 原子地设置为 nil, 并由 producer 原子地读取.
	vals []eface
}

// 就是 interface{} 的底层结构.
type eface struct {
	typ, val unsafe.Pointer
}
```


```cgo
// 添加元素. 添加到 head, 只能被单个 producer 调用
func (d *poolDequeue) pushHead(val interface{}) bool {
	ptrs := atomic.LoadUint64(&d.headTail)
	head, tail := d.unpack(ptrs)
	if (tail+uint32(len(d.vals)))&(1<<dequeueBits-1) == head {
		// Queue is full.
		return false
	}
	slot := &d.vals[head&uint32(len(d.vals)-1)]

	// Check head slot has been released by popTail.
	typ := atomic.LoadPointer(&slot.typ)
	if typ != nil {
		// Another goroutine is still cleaning up the tail, so
		// the queue is actually still full.
		return false
	}

	// The head slot is free, so we own it.
	if val == nil {
		val = dequeueNil(nil)
	}
	*(*interface{})(unsafe.Pointer(slot)) = val

	// Increment head. This passes ownership of slot to popTail
	// and acts as a store barrier for writing the slot.
	atomic.AddUint64(&d.headTail, 1<<dequeueBits)
	return true
}

// 从队列尾删除元素. 被 consumers 调用
func (d *poolDequeue) popTail() (interface{}, bool) {
	var slot *eface
	
	// 这里使用 for + cas 是保证并发的关键.
	for {
		ptrs := atomic.LoadUint64(&d.headTail)
		head, tail := d.unpack(ptrs)
		if tail == head {
			// Queue is empty.
			return nil, false
		}

		// 确定 head 和 tail, 并增加 tail. 如果执行成功, 则调用者将获得 tail slot
		ptrs2 := d.pack(head, tail+1)
		if atomic.CompareAndSwapUint64(&d.headTail, ptrs, ptrs2) {
			slot = &d.vals[tail&uint32(len(d.vals)-1)]
			break
		}
	}

	// 从 slot 当中获取值
	val := *(*interface{})(unsafe.Pointer(slot))
	if val == dequeueNil(nil) {
		val = nil
	}

	// 告知 pushHead 当前的 slot 已经完成了消费. 需要将 slot 归零很重要, 因此不会留下可能使该对象
	// 存活时间超过必要的引用.
	slot.val = nil
	atomic.StorePointer(&slot.typ, nil)
	// At this point pushHead owns the slot.

	return val, true
}

// 从队列尾删除元素. 只能被单个 producer 调用
func (d *poolDequeue) popHead() (interface{}, bool) {
	var slot *eface
	for {
		ptrs := atomic.LoadUint64(&d.headTail)
		head, tail := d.unpack(ptrs)
		if tail == head {
			// Queue is empty.
			return nil, false
		}

		// 确定 tail 和 减小 head. 如果执行成功, 则调用者将获得 head slot
		head--
		ptrs2 := d.pack(head, tail)
		if atomic.CompareAndSwapUint64(&d.headTail, ptrs, ptrs2) {
			// We successfully took back slot.
			slot = &d.vals[head&uint32(len(d.vals)-1)]
			break
		}
	}

	val := *(*interface{})(unsafe.Pointer(slot))
	if val == dequeueNil(nil) {
		val = nil
	}
	
	*slot = eface{} // slot 归零. 因为是非并发调用, 因此不会出现数据竞争.
	return val, true
}
```


在 poolDequeue 的基础上的 `双链表队列`(poolChain):

![image](/images/develop_sync_pool_poolChain.png)

```cgo
// 双链表队列元素, 值是双端队列
type poolChainElt struct {
	poolDequeue
	
	// next 可以被 producer 原子写入. 被 consumer 原子读. 它只能由 nil 到 non-nil
	// prev 可以被 consumer 原子写入, 被 producer 原子读, 它只能由 non-nil 到 nil
	next, prev *poolChainElt
}

// 双链表的具体实现.
// consumers 从 c.tail 获取数据, producer 将数据插入 c.head.
type poolChain struct {
	// push, head 只能被 producer 访问, 并写入数据.
	head *poolChainElt
    
    // popTail, tail 可以被 consumers 访问, 读写都必须是原子性的.
	tail *poolChainElt
}
```

```cgo
const dequeueBits = 32
const dequeueLimit = (1 << dequeueBits) / 4

// 生产数据, 单一 producer 访问
func (c *poolChain) pushHead(val interface{}) {
	d := c.head
	if d == nil {
		// Initialize the chain.
		const initSize = 8 // Must be a power of 2
		d = new(poolChainElt)
		d.vals = make([]eface, initSize)
		c.head = d
		storePoolChainElt(&c.tail, d) // 初始化的时候, head 和 tail 都指向的一个位置.
	}

	if d.pushHead(val) {
		return
	}
    
    // 当前队列已满. 重新分配一个 dequeue, 之前的 2 倍.
	newSize := len(d.vals) * 2
	if newSize >= dequeueLimit {
		// Can't make it any bigger.
		newSize = dequeueLimit
	}
    
    // 新的 poolChainElt 在双向队列尾部插入. 
	d2 := &poolChainElt{prev: d}
	d2.vals = make([]eface, newSize)
	c.head = d2
	storePoolChainElt(&d.next, d2)
	d2.pushHead(val)
}

// 消费数据, 多个 consumers 访问
func (c *poolChain) popTail() (interface{}, bool) {
	d := loadPoolChainElt(&c.tail)
	if d == nil {
		return nil, false
	}

    // 注: d 在初始化的时值为 c.tail 
	for {
		// 加载 d.next 指针. 一般情况, d 可能暂时为空, 但如果在 popTail 之前 next 为非nil, 并且 popTail 失败,
		// 则 d 永久为空, 此时需要将 d 删除掉
		d2 := loadPoolChainElt(&d.next)

		if val, ok := d.popTail(); ok {
			return val, ok
		}

		if d2 == nil {
			// 此时 head 和 tail 指向的是同一个位置 d, 这种情况下, 需要等待 producer pushHead 数据. 
			return nil, false
		}

		// 此时, 要安全的移除掉 d. 即: 先修改 c.tail 的指向, 然后修改 d2.prev 置空. 
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&c.tail)), unsafe.Pointer(d), unsafe.Pointer(d2)) {
			// 清空 d2.prev 的指针, 之后 d 变成了没有任何引用的的对象, 被清理掉.
			storePoolChainElt(&d2.prev, nil)
		}
		d = d2 // 下一轮.
	}
}

// 从 head 当中移除数据, 单一 producer 访问
func (c *poolChain) popHead() (interface{}, bool) {
	d := c.head
	for d != nil {
		if val, ok := d.popHead(); ok {
			return val, ok
		}
		// 之前的队列中可能还有未消费的元素, 所以尝试回朔.
		// 这导致的问题: 回朔之后空出来的 slot 不会再被放入数据, 直到这个 poolChainElt 满足条件被清理.
		d = loadPoolChainElt(&d.prev)
	}
	return nil, false
}
```


### pool 原理

pool 主要分为两级缓存. 一级是 private, 优先获取. 二级是 shared (双向链表数组, 数组的每个双向链表代表一个P)

![image](/images/develop_sync_pool_pool.png)

```cgo
type poolLocalInternal struct {
	private interface{} // 只能被相应的 P 使用.
	shared  poolChain   // 本地 P 可以 pushHead/popHead; 其他 P 只能 popTail.
}

type poolLocal struct {
	poolLocalInternal

	// Prevents false sharing on widespread platforms with
	// 128 mod (cache line size) = 0 .
	pad [128 - unsafe.Sizeof(poolLocalInternal{})%128]byte
}

type Pool struct {
	noCopy noCopy

	local     unsafe.Pointer // local fixed-size per-P pool, actual type is [P]poolLocal
	localSize uintptr        // size of the local array

	victim     unsafe.Pointer // local from previous cycle
	victimSize uintptr        // size of victims array

	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() interface{}
}
```

在 Get 的时候, 根据 pid 获取对应的 poolLocal, 优先获取 private 数据, 如果不存在, 从 shared.popHead 获取, 如果
仍然不存在, 则尝试从每个的 poolLocal 的 shared.popTail 获取. (特殊情况, 在 STW 期间, 会从 victim 当中获取). 在
尝试了上述所有的方法的都无法获取的时候, 使用 New 去创建一个.
