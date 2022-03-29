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
	// 在 [tail, head) 区间的 slot 归消费者所有.
	// 消费者继续拥有这个范围之外的 slot, 直到它把 slot 设置为 nil, 此时所有权传递给消费者.
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