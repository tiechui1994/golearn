## Go 内存管理

在 Go 1.10 之前的版本, 堆区的内存空间是连续的; 但是在 1.11版本, 使用稀疏的堆内存空间替代了连续的内存, 解决了连续内存
带来的限制以及在特殊场景下肯能出现的问题.

### 线性内存

1.10 版本在启动时会初始化整片虚拟内存区域, 如下所示的三个区域 `spans`, `bitmap` 和 `arena` 分别预留了 512MB, 16GB,
512GB的内存空间, 这些内存并不是真正存在的物理内存, 而是虚拟内存:

![image](/images/develop_memory_linemem.jpeg)

- `arena` 区域是真正的堆区, 运行时将**8KB**看做一页, 这些内存页中存储了所有堆上初始化的对象. 一些页组合起来称为 `mspan`.

- `bitmap` 用于标识 `arena` 区域中哪些地址保存了对象, 并且用 4bit 标志位表示对象是否包含指针, GC标记信息. bitmap 
中的一个 byte(8bit) 大小的内存对应 `area` 区域中 4 个指针大小(指针大小为 8byte)的内存. 1byte -> 32byte, 所以 
`bitmap` 区域的大小是 `512GB/(4*8)=16GB`

![image](/images/develop_memory_bitmap.jpeg)

> 从上图可以看到, bitmap的高地址指向 arena 区域的低地址部分. 即 bitmap 的地址是由高地址向低地址增长的.


- `spans` 区域存储了指向内存管理单元 `runtime.mspan`(`arena`分割的页组合起来的内存管理基本单元) 的指针, 
每个指针对应一页, 所以 `spans` 区域的大小是 `512GB/8KB*8B=512MB`. 除以 `8KB` 是计算 `arena` 区域的页数, 乘以 
8B是计算 `spans` 区域所有指针的大小. 每个内存管理单元会管理几页(页大小为 8KB)的内存空间. 创建 `mspan` 时, 按页填充
对应的的 `spans` 区域, 在回收 object 的时候, 根据地址很容易获得它所属的 `mspan`.



对于任意一个地址, 都可以根据 `arena` 的基地址计算该地址所在的页数, 并通过 `spans` 数组获得管理该片内存的管理单元
`runtime.mspan`, `spans` 数组中多个连续的位置可能对应同一个 `runtime.mspan`. 

Go 在垃圾回收时会根据指针的地址判断对象是否在堆中, 并通过上面介绍的过程找到管理该对象的 `runtime.mspan`. 这些都是建立
在**堆区的内存是连续的**这一假设上. 这种设计虽然简单方便, 但是 C 和 GO 混合使用时会导致程序崩溃:

1.分配的内存地址发生冲突, 导致堆的初始化和扩容失败

2.没有被预留的大块内存会被分配给 C 语言的二进制, 导致扩容后的堆不连续;

### 稀疏内存

稀疏内存是 GO 在 1.11 中提出的方案, 使用稀疏的内存布局不仅能够移除堆大小的上限,  还能解决 C 和 Go 混合使用的地址空间
冲突问题. 不过因为稀疏内存的内存管理失去了内存的连续性这一个假设, 内存管理变得更加复杂.

![image](/images/develop_memory_heap.png)

运行时使用二维的 `runtime.heapArena` 数组管理所有的内存, 每个单元会管理 64MB 的内存空间:

```cgo
type heapArena struct {
    // 存储 arena 上的 pointer/scalar bitmap 
	bitmap [heapArenaBitmapBytes]byte
	
	// spans 映射 "arena的虚拟地址页面ID" 到 *mspan
    // 对于 allocted spans, pages 映射到 span 本身,
    // 对于 free spans, 只有 lowest 和 highest pages 会映射到 span 本身.
    // 对于 internal pages 映射到任意 span.
    // 对于 never allocated pages, span 条目为nil.
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

Go 语言的内存分配器包含内存管理单元, 线程缓存, 中心缓存和页堆几个重要组件. 分别对应数据结构 `runtime.mspan`, 
`runtime.mcache`, `runtime.mcentral` 和 `runtime.mheap`. 

![image](/images/develop_memory_layout.jpeg)

所有的 Go 语言程序都会在启动时初始化如上图的内存布局, 每一个处理器都会被分配一个线程缓存 `runtime.mcache` 用于处理微
对象和小对象的分配, 它们会持有内存管理单元 `runtime.mspan`

每个类型的内存管理单元都会管理特定大小的对象, 当内存管理单元中不存在空闲对象时, 它们会从 `runtime.mheap` 持有的 134
个中心缓存 `runtime.mcentral` 中获取新的内存单元, 中心缓存属于全局的堆结构体 `runtime.mheap`, 它会从操作系统中申
请内存.

在 amd64 的 Linux 系统上, `runtime.mheap` 会持有 4194304 个 `runtime.heapArena`, 每一个 `runtime.heapArena`
都会管理 64 MB 的内存, 单个 Go 语言程序的内存上限就是 256TB.


### 内存管理单元

`runtime.mspan` 是 Go内存管理的基本单元, 是一片连续的 `8KB` 的页组成的大块内存. 注意: 这里的页和操作系统本身的页并
不是一回事, 它一般是操作系统页大小的几倍. `runtime.msapn` 是一个包含起始地址, `mspan`规格, 页的数量等内容的双端链表.


每个 `runtime.mspan` 按照它自身的属性 `Size Class` 的大小分割成若干个 `object`, 每个 `object` 可存储一个对象.
并且会使用一个位图标记其尚未使用的 `object`. 属性 `Size Class` 决定 `object` 大小, 而 `runtime.msapn` 只会分配
给和 `object` 尺寸大小接近的对象, 当然了, 对象的大小要小于 `object` 大小. 还有一个概念: `Span Class`. 它和 `Size
Class` 的含义差不多.

```
Size_Class = Span_Class/2
```

因为其实每个 `Size Class` 有两个 `runtime.msapn`, 也就是两个 `Span Class`. 其中一个分配给含有指针的对象, 另一个
分配给不含指针的对象. 这会给垃圾回收机制带来利好.

下图展示了 `runtime.mspan` 由一组连续的页组成, 按照一定大小划分成 `object`:

![image](/images/develop_memory_mspan_pic.jpeg)


在 Go1.9.2里 `runtime.mspan` 的 `Size Class` 共有 67 种. 每种 `mspan` 分割的 `object` 大小是 8*2n 的倍数.
这个是写死在代码里面的:

```
// path: src/runtime/sizeclasses.go
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
```

根据 `mspan` 的 `Size Class` 可以得到它划分的 `object` 大小. 比如 `Size Class` 等于3, `object` 大小就是 32B,
`object` 大小就是 32B. 32B大小的object可以存储对象的范围在 17B~32B 的对象. 而对于微小对象(小于16B), 分配器会将其
进行合并, 将几个对象分配到同一个 `object` 中.

数组当中最大的数是 32768, 也就是32KB, 超过此大小就是大对象了, 它会被特别对待. `Size Class` 为 0 表示大对象, 它直接
由堆内存分配, 而小对象都要通过 `mspan` 来分配.

对于 `mspan`, 它的 `Size Class` 会决定它所能分到的页数, 这也是写死的:

```
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

比如, 当申请一个 `object` 大小为 32B 的 `mspan` 时, 在 `class_to_size` 里对应的索引是 3, 而索引3在 `class_to_allocnpages`
数组里对应的页数是1.


```cgo
type mspan struct {
    // 链表后向地址, 用于将 span 链接起来
	next *mspan  
	// 链表前向地址, 用于将 span 链接起来
	prev *mspan    
	
	list *mSpanList // For debugging. TODO: Remove.

	startAddr uintptr // 起始地址, 即所管理页的地址
	npages    uintptr // 管理的页数, 每个页大小为 8KB

	nelems uintptr // 块个数, 表示有多少个块可供分配
    
    allocCount  uint16 // 已经分配的个数
	allocCache uint64 // allocBits的补码, 用于快速查找未被使用的内存

	allocBits  *gcBits // 标记内存的占用情况
	gcmarkBits *gcBits // 标记内存的回收情况
	
	sweepgen    uint32
	divMul      uint16     // for divide by elemsize - divMagic.mul
	baseMask    uint16     // if non-0, elemsize is a power of 2, & this will get object allocation base
	spanclass   spanClass  // 跨度类, 与 Size Class 相关. (sizeclass<<1) | bool2int(noscan)
	state       mSpanState // 状态
	needzero    uint8      // needs to be zeroed before allocation
	divShift    uint8      // for divide by elemsize - divMagic.shift
	divShift2   uint8      // for divide by elemsize - divMagic.shift2
	scavenged   bool       // whether this span has had its pages released to the OS
	elemsize    uintptr    // class表中的对象大小, 即块大小
	limit       uintptr    // end of data in span
	speciallock mutex      // guards specials list
	specials    *special   // linked list of special records sorted by offset.
}
```

> Size Class 实质就是 class_to_size 的索引. 通过它可以直接查找到elemsize和npages. 从而可以计算出 nelems. 

串联后的 `mspan` 是双向链表, 运行时会使用 `runtime.mSpanList` 存储双向链表的头结点和尾节点, 并在线程缓存以及中心缓
存中使用.

> 页和内存

每个 `runtime.mspan` 都管理 `npages` 个大小为 8KB 的页, 这里的页不是操作系统的内存页, 它是操作系统内存页的整倍数.

- `freeindex`, 扫描页中空闲对象的初始索引;

- `allocBits` 和 `gcmarkBits` 分别用于标记内存的占用和回收情况;

- `allocCache`, 是 `allocBits` 的补码, 可以用于快速查找内存中未被使用的内存;

`runtime.mspan` 会以三种不同是视角看待管理的内存. 当结构体管理的内存不足时, 运行时会以页为单位向堆申请内存:

![image](/images/develop_memory_mspan_inheap.png)


当用户程序或者线程向 `runtime.mspan` 申请内存时, 该结构体会使用 `allocCache` 字段**以对象为单位**在管理的内存中快
速查找待分配的空间:

![image](/images/develop_memory_mspan_inuser.png)


运行时的 `runtime.mspan` 结构如下:

![image](/images/develop_memory_mspan_inglobal.jpeg)


> 状态

运行时会使用 `runtime.mSpanStateBox` 结构体存储内存管理单元的状态 `runtime.mSpanState`.

该状态可能处于 `mSpanDead`, `mSpanInUse`, `mSpanManual` 和 `mSpanFree` 四种情况. 当 `runtime.mspan` 在空闲
堆中, 它会处于 `mSpanFree` 状态; 当 `runtime.mspan` 已经被分配时, 它会处于 `mSpanInUse`, `mSpanManul` 状态,
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
    
    // 微对象分配器
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

> 微分配器

线程缓存中还包含几个用于分配微对象的字段. `tiny`, `tinyoffset`, `local_tinyallocs` 三个字段组成了微对象分配器,
专门为 16 字节以下的对象申请和管理内存.

微分配器只会用于分配非指针类型的内存, 上述三个字段中 `tiny` 会指向堆中的一片内存, `tinyOffset` 是下一个空闲内存所在的
偏移量, 最后的 `local_tinyallocs` 会记录内存分配器中分配的对象个数.


### 中心缓存

`runtime.mcentral` 是内存分配器的中心缓存, 与线程缓存不同, 访问中心缓存中的内存管理单元需要使用互斥锁:

````cgo
type mcentral struct {
	lock      mutex
	
	// 跨度类 
	spanclass spanClass
	nonempty  mSpanList // list of spans with a free object, ie a nonempty free list
	empty     mSpanList // list of spans with no free objects (or cached in an mcache)

	// nmalloc是从此 mcentral 分配的对象的累积计数, 假定 mcache 中的所有 spans 都已完全分配. 
	// 写入是的, 读取是在STW下进行的
	nmalloc uint64
}
````

每一个中心缓存都会管理某个跨度类的内存管理单元, 它会同时持有两个 `runtime.mSpanList`, 分别存储包含空闲对象的列表和不
包含空闲对象的链表.

该结构体在初始化时, 两个链表都不包含任何内存, 程序运行时会扩容结构体持有的两个链表.

> 内存管理单元

线程缓存会通过中心缓存的 `runtime.mcentral.cacheSpan()` 方法获取新的内存管理单元. 包括以下几个部分:

1.从有空闲对象 `runtime.mspan` 列表中查找可以使用的内存管理单元;

2.从没有空闲对象的 `runtime.mspan` 链表中查找可以使用的内存管理单元;

3.调用 `runtime.mcentral.grow()` 从堆中申请新的内存管理单元;

4.更新内存管理单元的 `allocCache` 等字段帮助快速分配内存;

首先, 会在中心缓存的非空链表中查找可用的 `runtime.mspan`, 根据 `sweepgen` 字段分别进行不同的处理:

1.当内存单元**等待回收时**, 将其插入 `empty` 链表, 调用 `runtime.mspan.sweep` 清理该单元并返回;

2.当内存单元**正在被后台回收时**, 跳过该内存单元;

3.当内存单元**已经被回收时**, 将内存单元插入 `empty` 链表并返回;

```cgo
func (c *mcentral) cacheSpan() *mspan {
	// Deduct credit for this span allocation and sweep if necessary.
	spanBytes := uintptr(class_to_allocnpages[c.spanclass.sizeclass()]) * _PageSize
	deductSweepCredit(spanBytes, 0)

	lock(&c.lock)
	traceDone := false
	if trace.enabled {
		traceGCSweepStart()
	}
	sg := mheap_.sweepgen
retry:
	var s *mspan
	for s = c.nonempty.first; s != nil; s = s.next {
		if s.sweepgen == sg-2 && atomic.Cas(&s.sweepgen, sg-2, sg-1) {
		    // 等待回收
			c.nonempty.remove(s)
			c.empty.insertBack(s)
			unlock(&c.lock)
			s.sweep(true) // 回收
			goto havespan
		}
		if s.sweepgen == sg-1 {
		    // 正在回收(后台)
			continue
		}
		// 已经回收
		c.nonempty.remove(s)
		c.empty.insertBack(s)
		unlock(&c.lock)
		goto havespan
	}
	...
}
```

如果中心缓存没有在 `nonempty` 找到可用的内存管理单元, 就会继续遍历其持有的 `empty` 列表, 在这里的处理与包含空闲对象的
链表几乎完全相同. 当找到需要回收的内存单元时, 也会触发 `runtime.mspan.sweep` 进行清理, 如果清理后的内存单元仍然不包含
空闲对象, 就会重新执行相应的代码:
 
```cgo
func (c *mcentral) cacheSpan() *mspan {
    ...
    for s = c.empty.first; s != nil; s = s.next {
    		if s.sweepgen == sg-2 && atomic.Cas(&s.sweepgen, sg-2, sg-1) {
    		    // 等待回收, 有一个empty span, 需要sweeping, sweep之后看是否可以释放其中的一些空间.
    			c.empty.remove(s)
    			c.empty.insertBack(s)
    			unlock(&c.lock)
    			s.sweep(true) // sweep
    			freeIndex := s.nextFreeIndex()
    			if freeIndex != s.nelems {
    				s.freeindex = freeIndex
    				goto havespan
    			}
    			lock(&c.lock)
    			// 扫描后 span 仍然为空, 因此它已经在空列表中, 因此只需重试.
    			goto retry
    		}
    		if s.sweepgen == sg-1 {
    		    // 正在回收
    			continue
    		}
    		// 已经清理过了 empty span, 所有随后的扫描也都是已经扫描或者扫描过程中(GC)
    		break
    	}
    	
        ...
}
```

如果两个链表中都没有找到可用的内存单元, 它会调用 `runtime.mcentral.grow()` 触发扩容操作从堆中申请新的内存:

```cgo
func (c *mcentral) cacheSpan() *mspan {
    ...
	unlock(&c.lock)

	// Replenish central list if empty.
	s = c.grow()
	if s == nil {
		return nil
	}
	lock(&c.lock)
	c.empty.insertBack(s)
	unlock(&c.lock)

	// At this point s is a non-empty span, queued at the end of the empty list,
	// c is unlocked.
havespan:
	if trace.enabled && !traceDone {
		traceGCSweepDone()
	}
	n := int(s.nelems) - int(s.allocCount)
	if n == 0 || s.freeindex == s.nelems || uintptr(s.allocCount) == s.nelems {
		throw("span has no free objects")
	}
	// Assume all objects from this span will be allocated in the
	// mcache. If it gets uncached, we'll adjust this.
	atomic.Xadd64(&c.nmalloc, int64(n))
	usedBytes := uintptr(s.allocCount) * s.elemsize
	atomic.Xadd64(&memstats.heap_live, int64(spanBytes)-int64(usedBytes))
	if trace.enabled {
		// heap_live changed.
		traceHeapAlloc()
	}
	if gcBlackenEnabled != 0 {
		// heap_live changed.
		gcController.revise()
	}
	freeByteBase := s.freeindex &^ (64 - 1)
	whichByte := freeByteBase / 8
	// Init alloc bits cache.
	s.refillAllocCache(whichByte)

	// Adjust the allocCache so that s.freeindex corresponds to the low bit in
	// s.allocCache.
	s.allocCache >>= s.freeindex % 64

	return s
```

无论何种方式获取的内存单元, 该方法最后都会对内存单元的 `allocBits` 和 `allocCache` 等字段进行更新, 让运行时在分配内
存时能够快速找到空闲的对象.

> 扩容

中心缓存的扩容方法 `runtime.mcentral.grow()` 会根据预先计算的 `class_to_allocnpages` 和 `class_to_size` 获
取待分配的页数以及跨度类并调用 `runtime.mheap.alloc()` 获取新的 `runtime.mspan` 结构:

```cgo
func (c *mcentral) grow() *mspan {
    // 获取要分配的页数
	npages := uintptr(class_to_allocnpages[c.spanclass.sizeclass()])
	// 获取要分配的跨度类大小(对象的大小)
	size := uintptr(class_to_size[c.spanclass.sizeclass()])

	s := mheap_.alloc(npages, c.spanclass, false, true)
	if s == nil {
		return nil
	}

    // 计算分配页的对象个数 n 
	// n := (npages << _PageShift) / size
	// _PageShift 13
	n := (npages << _PageShift) >> s.divShift * uintptr(s.divMul) >> s.divShift2
	s.limit = s.base() + size*n // 计算最大的位置, 该位置之后的内存会浪费掉
	heapBitsForAddr(s.base()).initSpan(s) // 初始化
	return s
}
```

### 页堆

`runtime.mheap` 是内存分配的核心结构体, Go语言程序只会存在一个全局的结构, 而堆上初始化的所有对象都由该结构体统一管理,
该结构体中包含两组非常重要的字段, 其中一个是全局中心缓存列表 `central`, 另一个是管理堆区内存区域的 `arena` 以及相关
字段.

页堆中包含一个长度为 134 的 `runtime.mcentral` 的数组, 其中 67 个为跨度类需要 `scan` 的中心缓存, 另外 67 个是
`noscan` 的中心缓存.

```cgo
type mheap struct {
	// 必须仅在 system stack 上获取锁, 否则,如果 g 的堆栈在持有锁的情况下增长, 则g可能会自锁.
	lock      mutex
	free      mTreap // free spans
	sweepgen  uint32 // sweep generation, see comment in mspan
	sweepdone uint32 // all spans are swept
	sweepers  uint32 // number of active sweepone calls

	// allspans 是曾经创建的所有 mspan 的一部分. 每个mspan恰好出现一次.
    // 
    // allspan 的内存是手动管理的, 可以随着堆的增长重新分配和移动.
    //
    // 通常, allspans 受 mheap_.lock 保护, 这可以防止并发访问以及释放 backing store.
    // STW 期间的访问可能不持有锁定, 但必须确保在访问 span 不会发生 allocation(因为这可能释放backing store).
	allspans []*mspan // all spans out there

	// scanSpans 包含两个 mspan 堆栈: 一个是 swept in-use spans, 另一个是 unswept spans.
	// 在每个GC周期中, 这两个角色都会交换. 由于扫掠根在每个周期上增加2, 这意味着扫掠跨度以
	// sweepSpans[sweepgen/2%2]为单位, 而未扫掠跨度以 sweepSpans[1-sweepgen/2%2]为单位.
	// 从未扫描的堆栈中扫出持久性有机污染物跨度, 并推送仍在扫描的堆栈中使用的跨度. 
	// 同样, 分配使用中的跨度会将其推入扫掠堆栈.
	sweepSpans [2]gcSweepBuf

	_ uint32 // align uint64 fields on 32-bit for atomics

	// Proportional sweep
	//
	// These parameters represent a linear function from heap_live
	// to page sweep count. The proportional sweep system works to
	// stay in the black by keeping the current page sweep count
	// above this line at the current heap_live.
	//
	// The line has slope sweepPagesPerByte and passes through a
	// basis point at (sweepHeapLiveBasis, pagesSweptBasis). At
	// any given time, the system is at (memstats.heap_live,
	// pagesSwept) in this space.
	//
	// It's important that the line pass through a point we
	// control rather than simply starting at a (0,0) origin
	// because that lets us adjust sweep pacing at any time while
	// accounting for current progress. If we could only adjust
	// the slope, it would create a discontinuity in debt if any
	// progress has already been made.
	pagesInUse         uint64  // pages of spans in stats mSpanInUse; R/W with mheap.lock
	pagesSwept         uint64  // pages swept this cycle; updated atomically
	pagesSweptBasis    uint64  // pagesSwept to use as the origin of the sweep ratio; updated atomically
	sweepHeapLiveBasis uint64  // value of heap_live to use as the origin of sweep ratio; written with lock, read without
	sweepPagesPerByte  float64 // proportional sweep ratio; written with lock, read without
	// TODO(austin): pagesInUse should be a uintptr, but the 386
	// compiler can't 8-byte align fields.

	// Scavenger pacing parameters
	//
	// The two basis parameters and the scavenge ratio parallel the proportional
	// sweeping implementation, the primary differences being that:
	//  * Scavenging concerns itself with RSS, estimated as heapRetained()
	//  * Rather than pacing the scavenger to the GC, it is paced to a
	//    time-based rate computed in gcPaceScavenger.
	//
	// scavengeRetainedGoal represents our goal RSS.
	//
	// All fields must be accessed with lock.
	//
	// TODO(mknyszek): Consider abstracting the basis fields and the scavenge ratio
	// into its own type so that this logic may be shared with proportional sweeping.
	scavengeTimeBasis     int64
	scavengeRetainedBasis uint64
	scavengeBytesPerNS    float64
	scavengeRetainedGoal  uint64
	scavengeGen           uint64 // incremented on each pacing update

	// Page reclaimer state

	// reclaimIndex is the page index in allArenas of next page to
	// reclaim. Specifically, it refers to page (i %
	// pagesPerArena) of arena allArenas[i / pagesPerArena].
	//
	// If this is >= 1<<63, the page reclaimer is done scanning
	// the page marks.
	//
	// This is accessed atomically.
	reclaimIndex uint64
	// reclaimCredit is spare credit for extra pages swept. Since
	// the page reclaimer works in large chunks, it may reclaim
	// more than requested. Any spare pages released go to this
	// credit pool.
	//
	// This is accessed atomically.
	reclaimCredit uintptr

	// Malloc stats.
	largealloc  uint64                  // bytes allocated for large objects
	nlargealloc uint64                  // number of large object allocations
	largefree   uint64                  // bytes freed for large objects (>maxsmallsize)
	nlargefree  uint64                  // number of frees for large objects (>maxsmallsize)
	nsmallfree  [_NumSizeClasses]uint64 // number of frees for small objects (<=maxsmallsize)

	// arenas is the heap arena map. It points to the metadata for
	// the heap for every arena frame of the entire usable virtual
	// address space.
	//
	// Use arenaIndex to compute indexes into this array.
	//
	// For regions of the address space that are not backed by the
	// Go heap, the arena map contains nil.
	//
	// Modifications are protected by mheap_.lock. Reads can be
	// performed without locking; however, a given entry can
	// transition from nil to non-nil at any time when the lock
	// isn't held. (Entries never transitions back to nil.)
	//
	// In general, this is a two-level mapping consisting of an L1
	// map and possibly many L2 maps. This saves space when there
	// are a huge number of arena frames. However, on many
	// platforms (even 64-bit), arenaL1Bits is 0, making this
	// effectively a single-level map. In this case, arenas[0]
	// will never be nil.
	arenas [1 << arenaL1Bits]*[1 << arenaL2Bits]*heapArena

	// heapArenaAlloc is pre-reserved space for allocating heapArena
	// objects. This is only used on 32-bit, where we pre-reserve
	// this space to avoid interleaving it with the heap itself.
	heapArenaAlloc linearAlloc

	// arenaHints is a list of addresses at which to attempt to
	// add more heap arenas. This is initially populated with a
	// set of general hint addresses, and grown with the bounds of
	// actual heap arena ranges.
	arenaHints *arenaHint

	// arena is a pre-reserved space for allocating heap arenas
	// (the actual arenas). This is only used on 32-bit.
	arena linearAlloc

	// allArenas is the arenaIndex of every mapped arena. This can
	// be used to iterate through the address space.
	//
	// Access is protected by mheap_.lock. However, since this is
	// append-only and old backing arrays are never freed, it is
	// safe to acquire mheap_.lock, copy the slice header, and
	// then release mheap_.lock.
	allArenas []arenaIdx

	// sweepArenas is a snapshot of allArenas taken at the
	// beginning of the sweep cycle. This can be read safely by
	// simply blocking GC (by disabling preemption).
	sweepArenas []arenaIdx

	// curArena is the arena that the heap is currently growing
	// into. This should always be physPageSize-aligned.
	curArena struct {
		base, end uintptr
	}

	_ uint32 // ensure 64-bit alignment of central

	// central free lists for small size classes.
	// the padding makes sure that the mcentrals are
	// spaced CacheLinePadSize bytes apart, so that each mcentral.lock
	// gets its own cache line.
	// central is indexed by spanClass.
	central [numSpanClasses]struct {
		mcentral mcentral
		pad      [cpu.CacheLinePadSize - unsafe.Sizeof(mcentral{})%cpu.CacheLinePadSize]byte
	}

	spanalloc             fixalloc // allocator for span*
	cachealloc            fixalloc // allocator for mcache*
	treapalloc            fixalloc // allocator for treapNodes*
	specialfinalizeralloc fixalloc // allocator for specialfinalizer*
	specialprofilealloc   fixalloc // allocator for specialprofile*
	speciallock           mutex    // lock for special record allocators.
	arenaHintAlloc        fixalloc // allocator for arenaHints

	unused *specialfinalizer // never set, just here to force the specialfinalizer type into DWARF
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

![image](/develop_memory_tiny_alloc.png)

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