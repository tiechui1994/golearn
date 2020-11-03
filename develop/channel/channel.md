# golang channel 源码解析


## 数据结构

```cgo
type hchan struct {
	qcount   uint           // chan 里元素的数量
	dataqsiz uint           // chan 底层循环数组的长度
	buf      unsafe.Pointer // 指向底层循环数组的指针, 只针对缓存型 chan. 内存肯定连续
	elemsize uint16 // chan 中元素大小
	closed   uint32 // chan 是否被关闭的标志
	elemtype *_type // chan 在元素类型 (element type)
	sendx    uint   // 已发送元素在循环数组中的索引 (send index)
	recvx    uint   // 已接收元素在循环数组中的索引 (receive index)
	
	// 注意: recvq, sendq 都表示被阻塞的 goroutine, 这些 goroutine 由于尝试读取 channel 或者 向
	// channel 发送数据而被阻塞 
	recvq    waitq  // 等待接收的 goroutine 队列 (list of recv waiters)
	sendq    waitq  // 等待发送的 goroutine 队列 (list of send waiters)

	// lock保护了hchan中的所有字段, 以及被此 channel blocked 的 sudog 中的几个字段.
    //
    // 不要在持有此锁的同时更改另一个G的状态(特别是, 处于ready状态的G), 因为这可能会因堆栈收缩而导致死锁.
	lock mutex
}
```

// waitq 是 `sudog` 的一个双向链表, 而 `sudog` 实际上是对 goroutine 的一个封装.

```cgo
type waitq struct {
    first *sudog
    last  *sudog
}
```

// 常量

```cgo
const (
    maxAlign  = 8
    // 在 64 位系统上 hchan 大小为 96, hchanSize=96
    // 在 32 位系统上 hchan 大小为 48, hchanSize=48
    hchanSize = unsafe.Sizeof(hchan{}) + uintptr(-int(unsafe.Sizeof(hchan{}))&(maxAlign-1))
)
```

### 相关函数

```cgo
// 获取buffer上第i个slot的位置. 建立在内存连续的基础上
func chanbuf(c *hchan, i uint) unsafe.Pointer {
	return add(c.buf, uintptr(i)*uintptr(c.elemsize))
}

func (c *hchan) raceaddr() unsafe.Pointer {
	// 将 cahn 上的读取和写入操作视为在此地址发生.
	// 避免使用qcount或dataqsiz的地址. 
	// 因为内置的len()和cap()会读取这些地址, 并且我们不希望它们与close()之类的操作竞争.
	return unsafe.Pointer(&c.buf)
}

```

### 创建 chan

```cgo
func makechan(t *chantype, size int) *hchan {
	elem := t.elem

	// 编译器的检查, 元素的大小, 以及基本的校验
	if elem.size >= 1<<16 {
		throw("makechan: invalid channel element type")
	}
	if hchanSize%maxAlign != 0 || elem.align > maxAlign {
		throw("makechan: bad alignment")
	}
    
    // 计算 elem.size 与 size 的乘积是否溢出
	mem, overflow := math.MulUintptr(elem.size, uintptr(size))
	if overflow || mem > maxAlloc-hchanSize || size < 0 {
		panic(plainError("makechan: size out of range"))
	}

	// 当存储在buf中的元素不包含指针时, hchan将不包含GC感兴趣的指针. buf指向相同的分配, elemtype是可持久化的.
    // sudog 从它们自己的线程中引用, 因此无法将其收集.
	var c *hchan
	switch {
	case mem == 0:
		// elem.size*size 结果为0, 要么是 elem.size 为0, 要么是创建的size为0
		// 此时只分配 hchan 大小的内存.
		c = (*hchan)(mallocgc(hchanSize, nil, true))
		// Race detector uses this location for synchronization.
		c.buf = c.raceaddr()
	case elem.ptrdata == 0:
		// elem 当中不包含指针. hchan和buf的地址是连续的. 此时分配 hchanSize + mem
		c = (*hchan)(mallocgc(hchanSize+mem, nil, true))
		c.buf = add(unsafe.Pointer(c), hchanSize) // buf 的地址的在 hchan 之后
	default:
		// elem 包含指针, 两次内存操作. hchan 和 buf的地址是不连续的
		c = new(hchan)
		c.buf = mallocgc(mem, elem, true)
	}

	c.elemsize = uint16(elem.size)
	c.elemtype = elem
	c.dataqsiz = uint(size)

	if debugChan {
		print("makechan: chan=", c, "; elemsize=", elem.size, "; elemalg=", elem.alg, "; dataqsiz=", size, "\n")
	}
	
	return c
}
```

### 发送

```cgo
// 如果 block 不为 true, 则该 protocol 不会休眠
// 当关闭与休眠有关的chan时, 可以使用 g.param == nil 唤醒休眠. 
func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
    // channnl 为nil时, 当进行非阻塞状况下(即 block 为 false), 立即返回, 否则休眠当前的 goroutine 到永远.
	if c == nil {
		if !block {
			return false
		}
		gopark(nil, nil, waitReasonChanSendNilChan, traceEvGoStop, 2)
		throw("unreachable")
	}
    
    // 开启了 -race 调试 
	if raceenabled {
		racereadpc(c.raceaddr(), callerpc, funcPC(chansend))
	}

	// 快速判断: 非阻塞模式下, 快速检测到失败, 不用获取锁, 快速返回(false)
    //
    // 考量:
    // 在观察到通道未关闭之后, 我们观察到该通道尚未准备好发送. 这些观察中的每一个都是 world-size 大小的读取(取决于通
    // 道的类型, 第一个 c.closed 和第二个 c.recvq.first 或 c.qcount).
    // 
    // 由于关闭的通道无法从"准备发送"转换为"未准备发送", 因此即使在两个观测值之间通道已关闭, 它们也隐含着两者之间的一个
    // 时刻, 即通道既未关闭又未关闭. 我们的行为就好像我们当时在观察该通道, 并报告发送无法继续进行.
    //
    // 如果在此处对读取进行了重新排序, 也可以: 如果我们观察到该通道尚未准备好发送, 然后观察到它没有关闭, 则表明该通道在
    // 第一次观察期间没有关闭.
    
    // c.closed == 0, channel 未关闭
    // c.dataqsiz == 0 && c.recvq.first == nil, 非缓冲型 chan, 且 "接收队列(recvq)" 当中没有元素
    // c.dataqsiz > 0 && c.qcount == c.dataqsiz, 缓存型 chan, 且 "缓冲队列(buf)" 已满
	if !block && c.closed == 0 && ((c.dataqsiz == 0 && c.recvq.first == nil) ||
		(c.dataqsiz > 0 && c.qcount == c.dataqsiz)) {
		return false
	}

	var t0 int64
	if blockprofilerate > 0 {
		t0 = cputicks() // 记录当前的 cpu 的 ticket
	}

    // 加锁状况下操作.
	lock(&c.lock)
    
    // 当前的 chan 已经关闭, 那当然就无法发送了!!
	if c.closed != 0 {
		unlock(&c.lock)
		panic(plainError("send on closed channel"))
	}
    
    // 从接收队列当中寻找一个接收者, 直接将当前的内容发送给这个接收者, 绕过缓存区
	if sg := c.recvq.dequeue(); sg != nil {
	    // 找到了等待的接收者. 可以绕过通道缓冲区(如果有)将要发送的值直接发送到接收者.
		send(c, sg, ep, func() { unlock(&c.lock) }, 3)
		return true
	}
    
    // 当前的接收队列为空. 这个时候需要借助缓冲区来存储发送的元素了.
	if c.qcount < c.dataqsiz {
		// 存在可用的缓冲区, 将发送元素入队列
		qp := chanbuf(c, c.sendx)
		
		// 开启 -race
		if raceenabled {
			raceacquire(qp)
			racerelease(qp)
		}
		
		// 数据移动
		typedmemmove(c.elemtype, qp, ep)
		
		// 修正 sendx 和 qcount 变量
		c.sendx++
		if c.sendx == c.dataqsiz {
			c.sendx = 0
		}
		c.qcount++
		unlock(&c.lock)
		return true
	}
    
    // 当前没有合适的缓冲区来存储发送的元素. 这个时候需要根据 "是否是阻塞发送" 来做出抉择
    // 如果是 "阻塞发送", 那么就需要休眠当前的 goroutine
    // 如果是 "非阻塞发送", 那么就快速失败返回(false) 
	if !block {
		unlock(&c.lock)
		return false
	}

	// 阻塞当前的 goroutine 
	gp := getg()
	mysg := acquireSudog()
	mysg.releasetime = 0
	if t0 != 0 {
		mysg.releasetime = -1
	}
	
	// 在 "分配elem" 和 "将mysg入队到gp.waitcopy可以找到它的地方" 之间没有堆栈拆分.
	mysg.elem = ep
	mysg.waitlink = nil
	mysg.g = gp
	mysg.isSelect = false
	mysg.c = c
	gp.waiting = mysg
	gp.param = nil
	c.sendq.enqueue(mysg) // 将当前的 sudog 入队列 sendq
	
	// 带 lock 休眠的休眠操作, 意味着在 goparkunlock 需要解锁 lock
	goparkunlock(&c.lock, waitReasonChanSend, traceEvGoBlockSend, 3)
	
	// 确保 "发送的值" 保持活跃状态, 直到接收者将其复制出来. 也就是延长发送的值的生存周期, 防止提前被 GC 回收
	// sudog具有指向堆栈对象的指针, 但是sudog不被视为stack跟踪器的root.
	KeepAlive(ep)
    
    // 唤醒后的一些检查和清理操作
	if mysg != gp.waiting {
		throw("G waiting list is corrupted")
	}
	gp.waiting = nil
	
	// 注意: 这里的 gp.param 的值是 mysg
	if gp.param == nil {
	    // 唤醒的以后, 发现 chan 已经关闭了, 那么发送肯定会存在问题
		if c.closed == 0 {
			throw("chansend: spurious wakeup")
		}
		panic(plainError("send on closed channel"))
	}
	gp.param = nil
	if mysg.releasetime > 0 {
		blockevent(mysg.releasetime-t0, 2)
	}
	mysg.c = nil
	releaseSudog(mysg)
	return true
}
```


```cgo
// send 是将发送方发送的ep值复制到接收方sg中(跳过缓冲区). 然后将接收方sg唤醒, 继续执行后续的操作.
//
// chan c 必须为空且已锁定. 使用unlockf()在发送后解锁c.
// sg 必须已经从 recvq 中出队.
// ep 必须为非nil, 并且指向堆或调用者的堆栈.
func send(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int) {
    // 开启了 -race 追踪 
	if raceenabled {
		if c.dataqsiz == 0 {
			racesync(c, sg)
		} else {
			qp := chanbuf(c, c.recvx)
			raceacquire(qp)
			racerelease(qp)
			raceacquireg(sg.g, qp)
			racereleaseg(sg.g, qp)
			c.recvx++
			if c.recvx == c.dataqsiz {
				c.recvx = 0
			}
			c.sendx = c.recvx // c.sendx = (c.sendx+1) % c.dataqsiz
		}
	}
	
	if sg.elem != nil {
		sendDirect(c.elemtype, sg, ep) // 直接发送
		sg.elem = nil
	}
	
	gp := sg.g
	unlockf()
	
	// 唤醒正在休眠的接收者
	gp.param = unsafe.Pointer(sg)
	if sg.releasetime != 0 {
		sg.releasetime = cputicks()
	}
	goready(gp, skip+1)
}
```

// 直接发送

```cgo
// 仅在 "一个正在运行的goroutine写入到另一个正在运行的goroutine的stack" 的操作是在 "无缓冲" 或 "空缓冲" chan 上
// 发送和接收.
// 
// GC假定 stack 写入仅在goroutine运行时发生, 并且仅由该goroutine完成. 
// 使用写屏障足以弥补违反该假设的缺点, 但是写屏障必须起作用.
// typedmemmove 将调用 bulkBarrierPreWrite, 但是目标字节不在 heap 中, 因此这无济于事. 
// 因此这里安排调用 memmove 和 typeBitsBulkBarrier.
func sendDirect(t *_type, sg *sudog, src unsafe.Pointer) {
	// src在我们的stack上, dst是另一个stack上的插槽.
    
    // 一旦我们从sg中读出了sg.elem, 如果目标 stack 被复制(缩小), 它将不再被更新.
    // 因此, 请确保在 read 和 use 之间没有任何抢占点.
	dst := sg.elem
	typeBitsBulkBarrier(t, uintptr(dst), uintptr(src), t.size)
	
	// 不需要cgo写屏障检查, 因为dst始终是Go内存.
	memmove(dst, src, t.size)
}
```

> typedmemmove() VS memmove()
>
> typedmemmove() 是将类型 _type 的对象从 src 拷贝到 dst, src和dst都是unsafe.Pointer, 这当中涉及到了内存操作.
由于GC的存在, 在拷贝前, 如果 _type 包含指针, 需要开启写屏障 bulkBarrierPreWrite. 然后调用 memmove() 进行内存
拷贝的操作. 
> 
> memmove() 直接是进行内存移动, 使用的是汇编实现的, 代码位置在 `src/runtime/memmove_amd64.s`. 当然了, 这个操作
本身就是单纯的内存操作, 不涉及其他任何内容.


---

> bulkBarrierPreWrite(dst, src, size uintptr)
>
> bulkBarrierPreWrite 使用[dst, dst + size) 的 pointer/scalar 信息对内存范围[src, src + size) 中的每个指针
slot 执行写屏障. 这会在执行 memmove 之前执行必要的写屏障.
> 
> src, dst和size必须是 pointer-aligned.
> 范围 [dst，dst + size) 必须位于单个对象内.
> 它不执行实际的写操作.
>
> 作为一种特殊情况, src == 0 表示正在将其用于memclr. bulkBarrierPreWrite将为每个写障碍的src传递0.
>
> 调用方应在调用memmove(dst,src,size) 之前立即调用bulkBarrierPreWrite. 此函数标记为 nosplit, 以避免被抢占; GC
不得停止memmove和执行屏障之间的goroutine. 调用者还负责cgo指针检查, 这可能是将Go指针写入非Go内存中.
>
> 对于分配的内存中完全不包含指针的情况(即src和dst不包含指针), 不维护 pointer bitmap; 通常, 通过检查 typ.ptrdata/ 
bulkBarrierPreWrite 的任何调用者都必须首先确保基础分配包含指针.

---

> typeBitsBulkBarrier(typ *_type, dst, src, size uintptr)
>
> typeBitsBulkBarrier 对将由 memmove 使用 bitmap 类型定位这些指针slot的每个指针从[src, src+size) 复制到
[dst, dst+size) 的每个指针执行写屏障.
>
> 类型typ必须与[src, src+size) 和 [dst, dst + size) 完全对应.
> dst, src和size必须是 pointer-aligned.
> 类型typ必须具有plain bitmap, 而不是GC程序.
> 此功能的唯一用途是在 chan 发送中, 并且64kB chan元素大小限制为我们解决了这一问题.
>
> 不能被抢占, 因为它通常在 memmove 之前运行, 并且GC必须将其视为原子动作.

### 接收

```cgo
// chanrecv在chan c上接收并将接收到的数据写入 ep.
// ep可能为nil, 在这种情况下, 接收到的数据将被忽略.
// 
// 如果 block == false 并且 没有可用元素, 则返回(false,false).
// 否则, 如果 c 关闭, 则*ep为零并返回(true, false).
// 否则, 用一个元素填充 *ep并返回(true, true)
// ep 必须是非nil且指向heap或调用者的stack.
func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
    // raceenabled: 不需要检查ep, 因为它始终在stack中, 或者是反射所分配的新内存.
    
    // 当前的 chan 为nil, 在非阻塞的状况下, 立即返回 (false,false), 否则阻塞当前的 goroutine 到永远
	if c == nil {
		if !block {
			return
		}
		gopark(nil, nil, waitReasonChanReceiveNilChan, traceEvGoStop, 2)
		throw("unreachable")
	}

	
	// 快速判断: 非阻塞模式下, 快速检测到失败, 不用获取锁, 快速返回(false, false)
    //
    // 在观察到 chan 尚未准备好接收之后, 我们观察到通道未关闭. 这些观察中的每一个都是单个 word-sized 的读取(第一个
    // c.sendq.first或c.qcount, 第二个c.closed)
    //
    // 由于无法重新打开通道, 因此对通道未关闭的后续观察意味着它在第一次观察时也未关闭. 我们的行为就好像我们当时在观察该
    // 通道, 并报告接收无法继续进行.
    //
    // 这里操作顺序在这里很重要: 在进行近距离追踪时, 反转操作可能导致错误的行为.
    //
    // c.dataqsiz == 0 && c.sendq.first == nil, 非缓冲型chan, 当前的 "发送队列(sendq)" 为空
    // c.dataqsiz > 0 && atomic.Loaduint(&c.qcount) == 0 缓冲型chan, 当前的 "缓冲队列(buf)" 为空
    // atomic.Load(&c.closed) == 0, chan 已关闭
	if !block && (c.dataqsiz == 0 && c.sendq.first == nil ||
		c.dataqsiz > 0 && atomic.Loaduint(&c.qcount) == 0) &&
		atomic.Load(&c.closed) == 0 {
		return
	}

	var t0 int64
	if blockprofilerate > 0 {
		t0 = cputicks()
	}
    
    // 锁定处理
	lock(&c.lock)
    
    // 当前 chan 未关闭的状况下, "缓存队列" 的元素数量为0, 说明目前没有接收到元素.
	if c.closed != 0 && c.qcount == 0 {
		if raceenabled {
			raceacquire(c.raceaddr())
		}
		unlock(&c.lock)
		if ep != nil {
			typedmemclr(c.elemtype, ep) // 清空 ep 内存
		}
		return true, false
	}
    
    // 阻塞的 "发送队列(sendq)" 当中存在 sender. 则说明当前的缓冲队列已满.
	if sg := c.sendq.dequeue(); sg != nil {
	    // 详细参考 recv 当中的解释
		recv(c, sg, ep, func() { unlock(&c.lock) }, 3)
		return true, true
	}
    
    // 当前的缓存队列存在元素
	if c.qcount > 0 {
		qp := chanbuf(c, c.recvx)
		if raceenabled {
			raceacquire(qp)
			racerelease(qp)
		}
		
		//  将 qp 的数据拷贝到 ep
		if ep != nil {
			typedmemmove(c.elemtype, ep, qp)
		}
		typedmemclr(c.elemtype, qp) // 清除 qp 内容
		
		// 修正 recvx 和 qcount 的值
		c.recvx++
		if c.recvx == c.dataqsiz {
			c.recvx = 0
		}
		c.qcount--
		unlock(&c.lock)
		return true, true
	}
    
    // 当前缓队列没有任何元素, 当然了, "发送队列(sendq)肯定为空".
    // 非阻塞状况下(block为false), 快速返回 (false, false)
    // 阻塞状况下(block为true), 休眠当前的 goroutine
	if !block {
		unlock(&c.lock)
		return false, false
	}

	gp := getg()
	mysg := acquireSudog()
	mysg.releasetime = 0
	if t0 != 0 {
		mysg.releasetime = -1
	}
	
	mysg.elem = ep
	mysg.waitlink = nil
	gp.waiting = mysg
	mysg.g = gp
	mysg.isSelect = false
	mysg.c = c
	gp.param = nil
	c.recvq.enqueue(mysg) // 将当前的 sudog 入队列recvq
	
	// 持有锁的状况下休眠
	goparkunlock(&c.lock, waitReasonChanReceive, traceEvGoBlockRecv, 3)
    
    // 休眠被唤醒的后需要的验证和清理工作
	if mysg != gp.waiting {
		throw("G waiting list is corrupted")
	}
	gp.waiting = nil
	if mysg.releasetime > 0 {
		blockevent(mysg.releasetime-t0, 2)
	}
	closed := gp.param == nil // gp.param 的值应该是 mysg. 只有在关闭的时候这个值才被设为 nil
	gp.param = nil
	mysg.c = nil
	releaseSudog(mysg)
	return true, !closed
}
```


```cgo
// recv 在 full chan(缓冲通道已满或者非缓冲通道) c上处理接收操作.
// 包含2个部分:
// 1) 由发送方 sg 发送的值被放入通道, 并且唤醒发送方以继续执行其后续的逻辑.
// 2) 接收方接收到的值(当前G)被写入 ep.
// 对于同步通道, 两个值相同.
// 对于异步通道, 接收方从通道缓冲区获取其数据, 而发送方的数据放入通道缓冲区.
//
// chan c 必须已满并且已锁定. recv用 unlockf 函数解锁c.
// sg 必须已经从 sendq 中出队列.
// ep 必须为非nil, 并且指向 heap 或调用者的stack.
func recv(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int) {
	if c.dataqsiz == 0 {
	    // 非缓冲型通道, 或者同步通道.
	    // 设置 -race
		if raceenabled {
			racesync(c, sg)
		}
		if ep != nil {
			recvDirect(c.elemtype, sg, ep) // 直接从 sender 复制数据.
		}
	} else {
	    // 缓冲型通道, 或者异步通道
	    // 
        // 先获取 recvx 对应的 slot, 
        // 先 slot 当中读取数据到 ep 当中, 然后将 sender 发送的数据再插入到 slot 当中.
        // 由于缓冲队列是满的, 所以经过上述的操作之后, 缓冲队列依旧是满的, 但是 recvx 需要向前移动一位(保证chan的元素
        // 都是先进先出的), 而且 sendx 也是需要向前移动一位. 但是 sendx 和 recvx 的值相等.
		
		// 获取 slot 的位置
		qp := chanbuf(c, c.recvx)
		
		// copy data from "buf" to "receiver"
		if ep != nil {
			typedmemmove(c.elemtype, ep, qp)
		}
		// copy data from "sender" to "buf"
		typedmemmove(c.elemtype, qp, sg.elem) 
		
		// 更新 recvx 和 sendx
		c.recvx++
		if c.recvx == c.dataqsiz {
			c.recvx = 0
		}
		c.sendx = c.recvx // c.sendx = (c.sendx+1) % c.dataqsiz
	}
	
	// sg 清理工作
	sg.elem = nil
	gp := sg.g
	unlockf()
	
	// 唤醒对应的发送者
	gp.param = unsafe.Pointer(sg)
	if sg.releasetime != 0 {
		sg.releasetime = cputicks()
	}
	goready(gp, skip+1)
}
```

```cgo
// 原理和 sendDirect 类似
func recvDirect(t *_type, sg *sudog, dst unsafe.Pointer) {
	src := sg.elem
	typeBitsBulkBarrier(t, uintptr(dst), uintptr(src), t.size)
	memmove(dst, src, t.size)
}
```

### 关闭

```cgo
func closechan(c *hchan) {
    // 当前的 chan 为 nil. 关闭自然会报错
	if c == nil {
		panic(plainError("close of nil channel"))
	}

	lock(&c.lock)
	
	// 重复关闭一个 chan
	if c.closed != 0 {
		unlock(&c.lock)
		panic(plainError("close of closed channel"))
	}
    
    // 设置了 -race 追踪
	if raceenabled {
		callerpc := getcallerpc()
		racewritepc(c.raceaddr(), callerpc, funcPC(closechan))
		racerelease(c.raceaddr())
	}
    
    // 关闭
	c.closed = 1

	var glist gList // 将所有处于阻塞状态的 g 串在一起. 链接点是 g.schedlink

	// 释放所有的 recvq
	for {
		sg := c.recvq.dequeue()
		if sg == nil {
			break
		}
		if sg.elem != nil {
			typedmemclr(c.elemtype, sg.elem) // 清除 sg 当中的元素值
			sg.elem = nil
		}
		if sg.releasetime != 0 {
			sg.releasetime = cputicks()
		}
		gp := sg.g
		gp.param = nil // 关闭的标志之一
		if raceenabled {
			raceacquireg(gp, c.raceaddr())
		}
		glist.push(gp) // 将 goroutine 存入 glist
	}

	// 释放所有的 sendq (they will panic)
	for {
		sg := c.sendq.dequeue()
		if sg == nil {
			break
		}
		sg.elem = nil
		if sg.releasetime != 0 {
			sg.releasetime = cputicks()
		}
		
		gp := sg.g
		gp.param = nil
		if raceenabled {
			raceacquireg(gp, c.raceaddr())
		}
		glist.push(gp) // 将 goroutine 存入 glist
	}
	unlock(&c.lock)

	// 唤醒 glist 当中所有的 gp. 这样做的好处是减少锁定的时间
	for !glist.empty() {
		gp := glist.pop()
		gp.schedlink = 0
		goready(gp, 3)
	}
}
```