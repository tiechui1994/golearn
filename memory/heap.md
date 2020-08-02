## 虚拟内存布局

在 Go 1.10 之前的版本, 堆区的内存空间是连续的; 但是在 1.11版本, 使用稀疏的堆内存空间替代了连续的内存, 解决了连续内存
带来的限制以及在特殊场景下肯能出现的问题.

### 线性内存

1.10 版本在启动时会初始化整片虚拟内存区域, 如下所示的三个区域 `spans`, `bitmap` 和 `arena` 分别预留了 512MB, 16GB,
512GB的内存空间, 这些内存并不是真正存在的物理内存, 而是虚拟内存:

- `spans` 区域存储了指向内存管理单元 `runtime.mspan` 的指针, 每个内存管理单元会管理几页的内存空间, 每页大小为 8KB

- `bitmap` 用于标识 `arena` 区域中那些地址保存了对象, 位图中的每个字节都会标示堆区域中的32字节是否空闲. 1b -> 32b

- `arena` 区域是真正的堆区, 运行时将8KB看做一页, 这些内存页中存储了所有堆上初始化的对象.


对于任意一个地址, 都可以根据 `arena` 的基地址计算该地址所在的页数, 并通过 `spans` 数组获得管理该片内存的管理单元
`runtime.mspan`, `spans` 数组中多个连续的位置可能对应同一个 `runtime.mspan`. 

Go 在垃圾回收时会根据指针的地址判断对象是否在堆中, 并通过上面介绍的过程找到管理该对象的 `runtime.mspan`. 这些都是建立
在**堆区的内存是连续的**这一假设上. 这种设计虽然简单方便, 但是 C 和 GO 混合使用时会导致程序崩溃:

1.分配的内存地址发生冲突, 导致堆的初始化和扩容失败

2.没有被预留的大块内存会被分配给 C 语言的二进制, 导致扩容后的堆不连续;

### 稀疏内存

稀疏内存是 GO 在 1.11 中提出的方案, 使用稀疏的内存布局不仅能够移除堆大小的上限,  还能解决 C 和 Go 混合使用的地址空间
冲突问题. 不过因为稀疏内存的内存管理失去了内存的连续性这一个假设, 内存管理变得更加复杂.

![image](/images/mem_spe_heap.png)

运行时使用二维的 `runtime.heapArena` 数组管理所有的内存, 每个单元会管理 64MB 的内存空间:

```cgo
type heapArena struct {
    // 存储 arena 上的 pointer/scalar bitmap 
	bitmap [heapArenaBitmapBytes]byte
	
	// spans 映射 "arena的虚拟地址页面ID" 到 *mspan
    // 对于allocted spans, pages 映射到 span 本身,
    // 对于free spans, 只有 lowest 和 highest pages 会映射到 span 本身.
    // 对于 internal pages 映射到任意 span.
    // 对于never allocated pages, span 条目为nil.
    // 
    // 修改受 mheap.lock 保护. 可以在不锁定的情况下执行读取, 但是 ONLY 从已知包含 in-use 或 stack spans 的索引
    // 中进行. 这意味着在确定地址是否存在以及在spans数组中查找地址之间一定没有安全点.
	spans [pagesPerArena]*mspan
	
	// pageInUse是一个bitmap, 指示哪些 spans 处于 mSpanInUse 状态. 该bitmap由页码(page number)索引, 但是仅
	// 使用每个 span 中第一页相对应的位.
    //
    // 写入受 mheap_.lock 保护.
	pageInUse [pagesPerArena / 8]uint8
	
	// pageMarks是一个bitmap, 它指示哪些 spans 上有标记的对象. 与 pageInUse 一样, 仅使用每个 span 中与第一页相
	// 对应的位.
    //
    // 在标记期间自动完成写入. 读取是非原子且无锁的, 因为它们仅在扫描期间发生 (因此从不与写入竞争).
    //
    // 用于快速查找可以被释放的 spans.
    //
    // TODO: 如果这是 uint64 可以进行更快的扫描, 那很好, 但是我们没有64位原子位操作.
	pageMarks [pagesPerArena / 8]uint8
	zeroedBase uintptr
}
```

该结构体的 `bitmap` 和 `spans` 与线性内存中的 `bitmap` 和 `spans` 区域一一对应, `zeroBase` 字段指向了该结构体
管理的内存的基地址. 这种设计将原有的连续大内存切分成稀疏的小内存, 而用于管理这些内存的元信息也被切分成了小块.

不同平台和架构的二维数组大小可能完全不同, 如果 Go 语言服务在 **Linux x86-64 架构上运行, 二维数组的一维大小会是1, 二
维大小是 4194304 (4M)**, 因为每一个指针占用8字节的内存空间, 所以元信息的总大小为 32 MB. 由于每个 `runtime.heapArea`
都会管理 64MB 的内存, 整个堆区最多可以管理 256TB (4M*64MB).


## 内存管理单元

Go 语言的内存分配器包含内存管理单元, 线程缓存, 中心缓存和页堆几个重要组件.  分别对应数据结构 `runtime.mspan`, 
`runtime.mcache`, `runtime.mcentral` 和 `runtime.mheap`. 

![image](/images/mem_layout.jpeg)

所有的 Go 语言程序都会在启动时初始化如上图的内存布局, 每一个处理器都会被分配一个线程缓存 `runtime.mcache` 用于处理微
对象和小对象的分配, 它们会持有内存管理单元 `runtime.mspan`

每个类型的内存管理单元都会管理特定大小的对象, 当内存管理单元中不存在空闲对象时, 它们会从 `runtime.mheap` 持有的 134
个中心缓存 `runtime.mcentral` 中获取新的内存单元, 中心缓存属于全局的堆结构体 `runtime.mheap`, 它会从操作系统中申
请内存.

在 amd64 的 Linux 系统上, `runtime.mheap` 会持有 4194304 个 `runtime.heapArena`, 每一个 `runtime.heapArena`
都会管理 64 MB 的内存, 单个 Go 语言程序的内存上限就是 256TB.


### 内存管理单元

`runtime.mspan` 是 Go内存管理的基本单元, 该结构体包含了 `next` 和 `prev` 两个字段, 它们分别指向了前一个和后一个
`runtime.mspan`.

```cgo
type mspan struct {
	next *mspan     // next span in list, or nil if none
	prev *mspan     // previous span in list, or nil if none
	list *mSpanList // For debugging. TODO: Remove.

	startAddr uintptr // 起始地址
	npages    uintptr // 页数, 每个页大小为 8KB

	manualFreeList gclinkptr // list of free objects in mSpanManual spans

	freeindex uintptr // 扫描页中空闲对象的初始索引
	nelems uintptr // 当前 mspan 当中存储的对象的个数

	allocCache uint64 // allocBits的补码, 用于快速查找未被使用的内存

	allocBits  *gcBits // 标记内存的占用情况
	gcmarkBits *gcBits // 标记内存的回收情况
	
	sweepgen    uint32
	divMul      uint16     // for divide by elemsize - divMagic.mul
	baseMask    uint16     // if non-0, elemsize is a power of 2, & this will get object allocation base
	allocCount  uint16     // number of allocated objects
	spanclass   spanClass  // 跨度类, size class and noscan (uint8)
	state       mSpanState // 状态, mspaninuse etc
	needzero    uint8      // needs to be zeroed before allocation
	divShift    uint8      // for divide by elemsize - divMagic.shift
	divShift2   uint8      // for divide by elemsize - divMagic.shift2
	scavenged   bool       // whether this span has had its pages released to the OS
	elemsize    uintptr    // computed from sizeclass or from npages
	limit       uintptr    // end of data in span
	speciallock mutex      // guards specials list
	specials    *special   // linked list of special records sorted by offset.
}
```

串联后的 `mspan` 是双向链表, 运行时会使用 `runtime.mSpanList` 存储双向链表的头结点和尾节点, 并在线程缓存以及中心缓
存中使用.

> 页和内存

每个 `runtime.mspan` 都管理 `npages` 个大小为 8KB 的页, 这里的页不是操作系统的内存页, 它是操作系统内存页的整倍数.

- `freeindex`, 扫描页中空闲对象的初始索引;

- `allocBits` 和 `gcmarkBits` 分别用于标记内存的占用和回收情况;

- `allocCache`, 是 `allocBits` 的补码, 可以用于快速查找内存中未被使用的内存;

`runtime.msapn` 会以两种不同是视角看待管理的内存. 当结构体管理的内存不足时, 运行时会以页为单位向堆申请内存:

![image](/images/mem_mspan_heap.png)


当用户程序或者线程向 `runtime.mspan` 申请内存时, 该结构体会使用 `allocCache` 字段**以对象为单位**在管理的内存中快
速查找待分配的空间:

![image](/images/mem_mspan_user.png)


> 状态

运行时会使用 `runtime.mSpanStateBox` 结构体存储内存管理单元的状态 `runtime.mSpanState`.

该状态可能处于 `mSpanDead`, `mSpanInUse`, `mSpanManual` 和 `mSpanFree` 四种情况. 当 `runtime.msapn` 在空闲
堆中, 它会处于 `mSpanFree` 状态; 当 `runtime.msapn` 已经被分配时, 它会处于 `mSpanInUse`, `mSpanManul` 状态,
这些状态会在遵循以下规则发生转换:

- 在垃圾回收的**任意阶段**, 可能从 `mSpanFree` 转换到 `mSpanInUse` 和 `mSpanManual`;

- 在垃圾回收的**清除阶段**, 可能从 `mSpanInUse` 和 `mSpanManual` 转换到 `mSpanFree`;

- 在垃圾回收的**标记阶段**, 不能从 `mSpanInUse` 和 `mSpanManual` 转换到 `mSpanFree`;

设置 `runtime.mspan` 结构体状态的读写操作必须是原子性的, 避免垃圾回收造成的线程竞争问题.

> 跨度类

`runtime.spanClass` 是 `runtime.mspan` 结构体的跨度类, 它决定了内存管理单元中存储的对象大小和个数.

Go 语言的内存管理模块一共包含 67 种跨度类, 每一个跨度类都会存储特定大小的对象并且包含特定数量的页数以及对象, 所有的数据
都会预先计算好斌存储在 `runtime.class_to_size` 和 `runtime.class_to_allocnpages` 等变量中:

```cgo
var class_to_size = [67]uint16{
0, 8, 16, 32, 48, 64, 80, 96, 112, 128, 
144, 160, 176, 192, 208, 224, 240, 256, 
288, 320, 352, 384, 416, 448, 480, 512, 
576, 640, 704, 768, 896, 1024, 
1152, 1280, 1408, 1536, 1792, 2048, 
2304, 2688, 3072, 3200, 3456, 4096, 
4864, 5376, 6144, 6528, 6784, 6912, 8192, 
9472, 9728, 10240, 10880, 12288, 13568, 14336, 16384, 
18432, 19072, 20480, 21760, 24576, 27264, 28672, 32768}

var class_to_allocnpages = [67]uint8{
0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 
1, 1, 1, 1, 1, 1, 1, 1, 
1, 1, 1, 1, 1, 1, 1, 1, 
1, 1, 1, 1, 1, 1, 
1, 1, 2, 1, 2, 1, 
2, 1, 3, 2, 3, 1, 
3, 2, 3, 4, 5, 6, 1, 
7, 6, 5, 4, 3, 5, 7, 2, 
9, 7, 5, 8, 3, 10, 7, 4}
```

| class | bytes/obj | bytes/span | objects | tail waste | max waste |
| --- | --- | --- | --- | --- | --- |
| 1 | 8 | 8192 | 1024 | 0 | 87.50% |
| 2 | 16 | 8192 | 512 | 0 | 43.75% |
| 3 | 32 | 8192 | 256 | 0 | 46.88% |
| 4 | 48 | 8192 | 170 | 32 | 31.52% |
| 5 | 64 | 8192 | 128 | 0 | 31.52% |
| 6 | 80 | 8192 | 102 | 32 | 19.07% |
| ... | ... | ... | ... | ... | ... |
| 65 | 28672 | 57344 | 2 | 0 | 4.91% |
| 66 | 32768 | 32768 | 1 | 0 | 12.50% |

对象大小从8B到32KB, 总共66个跨度类的大小, 存储的对象数以及浪费的内存空间. 以第四个跨度类为例, 跨度类为 4 的 `runtime.mspan`
中对象的大小上限为 48 字节, 管理一个 1 页面(8KB), 最多可以存储 170 个对象, 尾部有 32 字节的内存被填充. 当页中存储的
对象都是 33 字节时, 最多浪费 31.52% 的资源:

((48-33)*170+32)/8192= 0.3158

// 65
((28672-27265)*2)/57344=0.0491

// 66
(32768-28673)/32768=0.1250

除了上述66个跨度类之外,运行时还包括ID为0的特殊跨度类, 它能够管理大于 32KB 的特殊对象.

跨度类中除了存储类别的ID之外, 它还会存储一个 noscan 标记位, 该标记表示对象是否包含指针,垃圾回收会堆包含指针的 `runtime.mspan`
结构体进行扫描. 下面是ID和标记位底层存储方式:

```cgo
// 通过ID和标记位构建 spanClass, 前7位存储跨度类的ID, 最后一位存储标记位
func makeSpanClass(sizeclass uint8, noscan bool) spanClass {
	return spanClass(sizeclass<<1) | spanClass(bool2int(noscan))
}

func (sc spanClass) sizeclass() int8 {
	returnint8(sc >> 1)
}

func (sc spanClass) noscan() bool {
	return sc&1 != 0
}
```


### 线程缓存

`runtime.mcache` 是 Go 语言中的线程缓存, 他会与线程的处理器(m)一一绑定, 主要用来缓存用户程序申请的微对象和小对象.

每一个线程缓存都持有 67*2 个 `runtime.mspan`, 这些内存管理单元都存储在结构体 `alloc` 字段中:

```cgo
type mcache struct {
	next_sample uintptr // trigger heap sample after allocating this many bytes
	local_scan  uintptr // bytes of scannable heap allocated

	tiny             uintptr
	tinyoffset       uintptr
	local_tinyallocs uintptr // number of tiny allocs not counted in other stats
    
    // 存储内存管理单元
	alloc [numSpanClasses]*mspan // spans to allocate from, indexed by spanClass

	stackcache [_NumStackOrders]stackfreelist

	local_largefree  uintptr                  // bytes freed for large objects (>maxsmallsize)
	local_nlargefree uintptr                  // number of frees for large objects (>maxsmallsize)
	local_nsmallfree [_NumSizeClasses]uintptr // number of frees for small objects (<=maxsmallsize)

	flushGen uint32
}
```

线程缓存在刚刚被初始化时是不包含 `runtime.mspan` 的, 只有当用户程序申请内存时才会从上一级组件获取新的 `runtime.mspan`
满足内存分配的需求.

> 初始化

运行时在初始化处理器时会调用 `runtime.allocmcache` 初始化线程缓存, 该函数在**系统栈**中使用 `runtime.mheap` 中的
*线程缓存分配器*初始化新的 `runtime.mcache` 结构体:

```cgo
func allocmcache() *mcache {
	var c *mcache
	// 系统栈, 使用 mheap 的 cachealloc 分配
	systemstack(func() {
		lock(&mheap_.lock)
		c = (*mcache)(mheap_.cachealloc.alloc())
		c.flushGen = mheap_.sweepgen
		unlock(&mheap_.lock)
	})
	// 空的占位符 `emptymspan`
	for i := range c.alloc {
		c.alloc[i] = &emptymspan
	}
	c.next_sample = nextSample()
	return c
}
```

> 替换

`mcache.refill` 方法会为线程缓存获取一个指定跨度类的内存管理单元, 被替换的单元不能包含空闲的内存空间, 而获取的单元至少
包含一个空闲对象用于分配内存:

```cgo
func (c *mcache) refill(spc spanClass) {
	// Return the current cached span to the central lists.
	s := c.alloc[spc]

	if s != &emptymspan {
		atomic.Store(&s.sweepgen, mheap_.sweepgen)
	}
    
    // 从中心缓存中申请新的 runtime.mspan 存储到线程缓存中, 这也是向线程缓存中插入内存管理单元的唯一方法.
	s = mheap_.central[spc].mcentral.cacheSpan()
	s.sweepgen = mheap_.sweepgen + 3

	c.alloc[spc] = s
}
```


## 内存分配

堆上的所有的对象都会通过调用 `runtime.newobject` 函数分配, 该函数会调用 `runtime.mallocgc` 分配指定大小的内存空
间, 这也是用户程序向堆上申请空间的必经函数.

```cgo
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	mp := acquirem()
	mp.mallocing = 1

	shouldhelpgc := false
    dataSize := size
    c := gomcache() // get mcache
	var x unsafe.Pointer
	noscan := typ == nil || typ.ptrdata == 0 // true 表示非指针, false 表示指针
	// maxSmallSize 32768, maxTinySize 16
	if size <= maxSmallSize {
		if noscan && size < maxTinySize {
			// 微对象分配
		} else {
			// 小对象分配
		}
	} else {
		// 大对象分配
	}

	publicationBarrier()
	mp.mallocing = 0
	releasem(mp)

	return x
}
```

上述代码使用 `runtime.gocache` 获取线程缓存并通过类型判断是否为指针类型. `runtime.mallocgc` 会根据对象的大小执行
不同的分配逻辑, 运行时根据对象大小将它们分为微对象, 小对象和大对象. 这里会根据大小选择不同的分配逻辑:

- 微对象 `(0, 16B)`, 先使用微型分配器, 再依次尝试线程缓存, 中心缓存 和 堆分配内存.

- 小对象 `[16B, 32KB]`, 依次尝试使用线程缓存, 中心缓存和堆分配内存;

- 大对象 `(32KB, +OO)`, 直接在堆上分配内存;

#### 微对象

**将小于16字节的对象划分为微对象**, 它会使用线程缓存上的微分配器提高对象分配的性能, 我们**主要使用它来分配较小的字符串以
及逃逸的临时变量**. 微分配器可以将多个较小的内存分配请求合入同一个内存块中, 只有当内存块中的所有对象都需要被回收时, 整片
内存才可能被回收.

**微分配器管理的对象不可以是指针类型**, 管理多个对象的内存块大小 `maxTinySize` 是可以调整的, 在默认情况下, 内存块的
大小为16字节. `maxTinySize` 的值越大, 结合多个对象的可能性就越高, 内存浪费也就越严重; `maxTinySize` 越小, 内存浪
费就会越小, 不过无论如何调整, 8 的倍数是一个很好的选择. 

> 8个字节完全不会浪费，但是合并的机会较少. 32字节提供了更多的合并机会，但可能导致最坏情况下浪费4倍.

不能显式释放从微分配器获得的对象. 因此, 当明确释放对象时, 我们确保其 size >= maxTinySize.

下面是一个例子:

![image](/images/mem_tiny_alloc.png)

微分配器已经在16字节的内存块中分配了12字节的对象, 如果下一个待分配的对象小于4字节, 它就会使用上述的内存块的剩余部分, 减
少内存碎片, 不过该内存块只有2个对象都被标记为垃圾时才会被回收.

---

线程缓存 `runtime.mcache` 中的 `tiny` 字段指向了 `maxTinySize` 大小的块, 如果当前块中还包含大小合适的空闲内存, 
运行时会通过基地址和偏移量获取并返回这块内存:

```cgo
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	...
	if size <= maxSmallSize {
		if noscan && size < maxTinySize {
		    // 微对象, 在 mcache 当前块中找到合适大小的内存
			off := c.tinyoffset // mcache 的偏移量
			// align tiny pointer 以进行必需的对齐.
			if size&7 == 0 {
                off = round(off, 8)
            } else if size&3 == 0 {
                off = round(off, 4)
            } else if size&1 == 0 {
                off = round(off, 2)
            }
			if off+size <= maxTinySize && c.tiny != 0 {
			    //  The object fits into existing tiny block.
				x = unsafe.Pointer(c.tiny + off)
				c.tinyoffset = off + size
				c.local_tinyallocs++
				releasem(mp)
				return x
			}
			...
		}
		...
	}
	...
}
```

当内存块中不包含空闲的内存时, 会先从线程缓存找到跨度类对应的内存管理单元 `runtime.mspan`, 调用 `runtime.nextFreeFast`
获取空闲的内存 `[1]`; 当不存在空闲时, 会调用 `runtime.mcache.nextFree` 从中心缓存或者页堆中获取可分配的内存块 `[2]`

```cgo
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	...
	if size <= maxSmallSize {
		if noscan && size < maxTinySize {
			...
			// [1] 线程缓存中分配一个新的 maxTinySize 块
			span := c.alloc[tinySpanClass]
			v := nextFreeFast(span)
			if v == 0 {
			    // [2] 中心缓存或者堆中获取可分配的内存块
				v, _, _ = c.nextFree(tinySpanClass)
			}
			x = unsafe.Pointer(v)
			(*[2]uint64)(x)[0] = 0
			(*[2]uint64)(x)[1] = 0
			// 根据剩余的可用空间量, 看看是否需要用新的小块替换现有的小块.
			if size < c.tinyoffset || c.tiny == 0 {
				c.tiny = uintptr(x)
				c.tinyoffset = size
			}
			size = maxTinySize
		}
		...
	}
	...
	return x
}
```

获取新的空闲内存块之后, 上述代码会清空空闲内存中的数据, 更新构成微对象分配器的几个字段 `tiny`, `tinyoffset` 并返回新
的空闲内存.


#### 小对象

小对象是指大小为16字节到32768字节的对象以及所有小于16字节的指针类型的对象, 小对象的分配可以被分为三个步骤:

1. 确定分配对象的大小以及跨度类 `runtime.spanClass`

2. 从线程缓存, 中心缓存或者堆中获取内存管理单元并从内存管理单元找到空闲的内存空间

3. 调用 `runtime.memclrNoHeapPointers` 清空空闲内存中的所有数据.

确定待分配的大小以及跨度类需要使用预先计算好的 `size_to_class8`, `size_to_class128`, 以及 `class_to_size` 字
典, 这些字典能够帮助我们快速获取对应的值并构建 `runtime.spanClass`:

```cgo
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	...
	if size <= maxSmallSize {
		...
		} else {
			var sizeclass uint8
			// smallSizeMax: 1024, smallSizeDiv:8, largeSizeDiv:128
            if size <= smallSizeMax-8 {
                sizeclass = size_to_class8[(size+smallSizeDiv-1)/smallSizeDiv]
            } else {
                sizeclass = size_to_class128[(size-smallSizeMax+largeSizeDiv-1)/largeSizeDiv]
            }
            size = uintptr(class_to_size[sizeclass])
            // noscan 是否没有指针
            spc := makeSpanClass(sizeclass, noscan)
            // mcache 获取一个 spanClass 大小的 mspan 
            span := c.alloc[spc]
            v := nextFreeFast(span)
            if v == 0 {
                // mcenter 或 heap 分配
                v, span, shouldhelpgc = c.nextFree(spc)
            }
            x = unsafe.Pointer(v)
            if needzero && span.needzero != 0 {
                memclrNoHeapPointers(unsafe.Pointer(v), size)
            }
        }
	} else {
		...
	}
	...
	return x
}
```

在上述片段中, 我们重点分析两个函数和方法的实现原理. 它们分别是 `runtime.nextFreeFast` 和 `runtime.mcache.nextFree`
这两个函数会帮助我们获取空闲的内存空间. `runtime.nextFreeFast` 会利用内存管理单元的 `allocCache` 字段, 快速找到
该字段中位1的位数, 在上面介绍过 1 表示该位对应的内存空间是空闲的:

```cgo
func nextFreeFast(s *mspan) gclinkptr {
    // allocCache 中是否有空闲对象?
	theBit := sys.Ctz64(s.allocCache)
	if theBit < 64 {
		result := s.freeindex + uintptr(theBit)
		if result < s.nelems {
			freeidx := result + 1
			if freeidx%64 == 0 && freeidx != s.nelems {
				return 0
			}
			s.allocCache >>= uint(theBit + 1)
			s.freeindex = freeidx
			s.allocCount++
			return gclinkptr(result*s.elemsize + s.base())
		}
	}
	return 0
}
```

找到了空闲的对象后, 就可以更新内存管理单元的 `allocCache`, `freeindex` 等字段并返回该片内存了; 如果没有找到空闲的内
存, 运行时会通过 `runtime.mcache.nextFree` 找到新的内存管理单元:

```cgo
func (c *mcache) nextFree(spc spanClass) (v gclinkptr, s *mspan, shouldhelpgc bool) {
	s = c.alloc[spc]
	freeIndex := s.nextFreeIndex()
	if freeIndex == s.nelems {
		c.refill(spc)
		s = c.alloc[spc]
		freeIndex = s.nextFreeIndex()
	}

	v = gclinkptr(freeIndex*s.elemsize + s.base())
	s.allocCount++
	return
}
```

如果我们的线程缓存中没有找到可用的内存管理单元, 会通过 `runtime.mcache.refill` 使用中心缓存中的内存管理单元替换已经
不存在可用对象的结构体, 该方法会调用新结构体的 `runtime.mspan.nextFreeIndex` 获取空闲的内存并返回.

### 大对象

运行时对于大于 32KB 的大对象会单独处理, 不会从线程缓存或者中心缓存中获取内存管理单元, 而是直接在系统的栈中调用函数
`runtime.largeAlloc` 函数分配大片的内存:


```cgo
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	...
	if size <= maxSmallSize {
		...
	} else {
		var s *mspan
		systemstack(func() {
			s = largeAlloc(size, needzero, noscan)
		})
		s.freeindex = 1
		s.allocCount = 1
		x = unsafe.Pointer(s.base())
		size = s.elemsize
	}

	publicationBarrier()
	mp.mallocing = 0
	releasem(mp)

	return x
}
```

`runtime.largeAlloc` 函数会计算分配该对象所需要的页数, 它会按照 8KB 的倍数为对象在堆上申请内存:

```cgo
func largeAlloc(size uintptr, needzero bool, noscan bool) *mspan {
    // _PageSize 8192
	if size+_PageSize < size {
		throw("out of memory")
	}
	
	// _PageShift 13, 一个页的大小是 8KB = 2^13B, 这里右移是计算
	npages := size >> _PageShift
	// 是否是整倍数
	if size&_PageMask != 0 {
		npages++
	}

	deductSweepCredit(npages*_PageSize, npages)

	s := mheap_.alloc(npages, makeSpanClass(0, noscan), true, needzero)
	if s == nil {
		throw("out of memory")
	}
	s.limit = s.base() + size
	heapBitsForAddr(s.base()).initSpan(s)
	return s
}
```

申请内存时会创建一个跨度类为0的 `runtime.spanClass` 并调用 `runtime.mheap_.alloc` 分配一个管理对应内存的管理单元.