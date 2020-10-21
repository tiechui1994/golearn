
```cgo
// 地址偏移(内存必须连续)
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

// 2^b, b的有效范围[0-63]
func bucketShift(b uint8) uintptr {
	return uintptr(1) << (b & (sys.PtrSize*8 - 1))
}

// 2^b-1
func bucketMask(b uint8) uintptr {
	return bucketShift(b) - 1
}

// hash值的高8位
func tophash(hash uintptr) uint8 {
	top := uint8(hash >> (sys.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}
	return top
}

// 获取 bucket 的状态是否在 evacuated (迁移状态)
func evacuated(b *bmap) bool {
	h := b.tophash[0]
	return h > emptyOne && h < minTopHash
}
```

// 结构体
```cgo
// A header for a Go map.
type hmap struct {
	count     int    // 代表哈希表中的元素个数, 调用len(map)时, 返回的就是该字段值.
	flags     uint8  // 状态标志, 下文常量中会解释四种状态位含义.
	B         uint8  // buckets（桶）的对数log_2（哈希表元素数量最大可达到装载因子*2^B）
	noverflow uint16 // 溢出桶的大概数量.
	hash0     uint32 // 哈希种子.

	buckets    unsafe.Pointer // 指向buckets数组的指针, 数组大小为2^B, 如果元素个数为0, 它为nil.
	oldbuckets unsafe.Pointer // 如果发生扩容, oldbuckets是指向老的buckets数组的指针, 老的buckets数组大小是新
	                          // 的buckets的1/2.非扩容状态下, 它为nil.
	                          
	nevacuate  uintptr        // 表示扩容进度, 小于此地址的buckets代表已搬迁完成.

	extra *mapextra // 这个字段是为了优化GC扫描而设计的.当key和value均不包含指针, 并且都可以inline时使用.
	                // extra是指向mapextra类型的指针.
}

// mapextra holds fields that are not present on all maps.
type mapextra struct {
	// 如果 key 和 value 都不包含指针, 并且可以被 inline(<=128 字节), 就使用 hmap 的extra字段来存储 overflow 
	// buckets, 这样可以避免 GC 扫描整个 map, 然而 bmap.overflow 也是个指针. 这时候我们只能把这些 overflow 的
	// 指针都放在 hmap.extra.overflow 和 hmap.extra.oldoverflow 中了. overflow 包含的是 hmap.buckets 的 
	// overflow 的 buckets, oldoverflow 包含扩容时的 hmap.oldbuckets 的 overflow 的 bucket.
	overflow    *[]*bmap
	oldoverflow *[]*bmap

	// 指向空闲的 overflow bucket 的指针
	nextOverflow *bmap
}

// A bucket for a Go map.
type bmap struct {
	// tophash包含此桶中每个键的哈希值最高字节（高8位）信息（也就是前面所述的high-order bits）.
	// 如果tophash[0] < minTopHash, tophash[0]则代表桶的搬迁（evacuation）状态.
	tophash [bucketCnt]uint8
}
```

// bmap的构建
```cgo
// src/cmd/compile/internal/gc/reflect.go:bmap
// bucket 结构 
func bmap(t *types.Type) *types.Type {
	if t.MapType().Bucket != nil {
		return t.MapType().Bucket
	}

	bucket := types.New(TSTRUCT)
	keytype := t.Key()
	elemtype := t.Elem()
	dowidth(keytype)
	dowidth(elemtype)
	if keytype.Width > MAXKEYSIZE {
		keytype = types.NewPtr(keytype)
	}
	if elemtype.Width > MAXELEMSIZE {
		elemtype = types.NewPtr(elemtype)
	}

	field := make([]*types.Field, 0, 5)

	// The first field is: uint8 topbits[BUCKETSIZE].
	arr := types.NewArray(types.Types[TUINT8], BUCKETSIZE)
	field = append(field, makefield("topbits", arr))

	arr = types.NewArray(keytype, BUCKETSIZE)
	arr.SetNoalg(true)
	keys := makefield("keys", arr)
	field = append(field, keys)

	arr = types.NewArray(elemtype, BUCKETSIZE)
	arr.SetNoalg(true)
	elems := makefield("elems", arr)
	field = append(field, elems)
	
	// 确保 overflow 指针是结构中的最后一个内存, 因为运行时假定它可以使用size-ptrSize作为 overflow 指针的偏移量. 
	// 一旦计算了偏移量和大小, 我们就要仔细检查下面的属性(在已经忽略检查代码).
    //
    // BUCKETSIZE为8, 因此该结构在此处已对齐为64位.
    // 在32位系统上, 最大对齐方式为32位, 并且溢出指针将添加另一个32位字段, 并且该结构将以无填充结尾.
    // 在64位系统上, 最大对齐方式为64位, 并且溢出指针将添加另一个64位字段, 并且该结构将以无填充结尾.
    // 但是, 在nacl/amd64p32上, 最大对齐方式是64位, 但是溢出指针只会添加一个32位字段, 因此, 如果该结构需要64位填充
    // (由于key或elem的原因), 则它将最后带有一个额外的32位填充字段.
    // 通过在此处发出填充.
	if int(elemtype.Align) > Widthptr || int(keytype.Align) > Widthptr {
		field = append(field, makefield("pad", types.Types[TUINTPTR]))
	}
	
	// 如果keys和elems没有指针, 则map实现可以在侧面保留一个 overflow 指针列表, 以便可以将 buckets 标记为没有指针.
    // 在这种情况下, 通过将 overflow 字段的类型更改为 uintptr, 使存储桶不包含任何指针.
	otyp := types.NewPtr(bucket)
	if !types.Haspointers(elemtype) && !types.Haspointers(keytype) {
		otyp = types.Types[TUINTPTR]
	}
	overflow := makefield("overflow", otyp)
	field = append(field, overflow)

	// link up fields
	bucket.SetNoalg(true)
	bucket.SetFields(field[:])
	dowidth(bucket)

	t.MapType().Bucket = bucket

	bucket.StructType().Map = t
	return bucket
}
```

// hmap的构建
```cgo
// src/cmd/compile/internal/gc/reflect.go:hmap
func hmap(t *types.Type) *types.Type {
	if t.MapType().Hmap != nil {
		return t.MapType().Hmap
	}

	bmap := bmap(t)
	
	// type hmap struct {
	//    count      int
	//    flags      uint8
	//    B          uint8
	//    noverflow  uint16
	//    hash0      uint32
	//    buckets    *bmap
	//    oldbuckets *bmap
	//    nevacuate  uintptr
	//    extra      unsafe.Pointer // *mapextra
	// }
	// must match runtime/map.go:hmap.
	fields := []*types.Field{
		makefield("count", types.Types[TINT]),
		makefield("flags", types.Types[TUINT8]),
		makefield("B", types.Types[TUINT8]),
		makefield("noverflow", types.Types[TUINT16]),
		makefield("hash0", types.Types[TUINT32]), // Used in walk.go for OMAKEMAP.
		makefield("buckets", types.NewPtr(bmap)), // Used in walk.go for OMAKEMAP.
		makefield("oldbuckets", types.NewPtr(bmap)),
		makefield("nevacuate", types.Types[TUINTPTR]),
		makefield("extra", types.Types[TUNSAFEPTR]),
	}

	hmap := types.New(TSTRUCT)
	hmap.SetNoalg(true)
	hmap.SetFields(fields)
	dowidth(hmap)

	// The size of hmap should be 48 bytes on 64 bit and 28 bytes on 32 bit platforms.
	// 5("count", "buckets", "oldbuckets", "nevacuate", "extra")
	if size := int64(8 + 5*Widthptr); hmap.Width != size {
		Fatalf("hmap size not correct: got %d, want %d", hmap.Width, size)
	}

	t.MapType().Hmap = hmap
	hmap.StructType().Map = t
	return hmap
}
```

如图:

![image](/images/develop_map_bmap.png)


// 常量值
```cgo
const (
	// 一个桶中最多能装载的键值对(key-value)的个数为8
	bucketCntBits = 3
	bucketCnt     = 1 << bucketCntBits // 8

	// 触发扩容的装载因子为13/2=6.5
	loadFactorNum = 13
	loadFactorDen = 2

	// 键和值超过128个字节, 就会被转换为指针
	maxKeySize  = 128
	maxElemSize = 128

	// 数据偏移量应该是bmap结构体的大小,它需要正确地对齐. 
	// 对于amd64p32而言, 这意味着: 即使指针是32位的, 也是64位对齐. 
	dataOffset = unsafe.Offsetof(struct {
		b bmap
		v int64
	}{}.v)

	// 每个桶(如果有溢出, 则包含它的overflow的链桶) 在搬迁完成状态(evacuated states)下, 要么会包含它所有的键值对,
	// 要么一个都不包含(但不包括调用evacuate()方法阶段,该方法调用只会在对map发起write时发生,在该阶段其他goroutine
	// 是无法查看该map的). 简单的说,桶里的数据要么一起搬走,要么一个都还未搬.
	//
	// tophash除了放置正常的高8位hash值, 还会存储一些特殊状态值(标志该cell的搬迁状态). 正常的tophash值, 
	// 最小应该是5,以下列出的就是一些特殊状态值. 
	emptyRest      = 0 // 表示cell为空,并且比它高索引位的cell或者overflows中的cell都是空的. (初始化bucket时,就是该状态)
	emptyOne       = 1 // 空的cell, cell已经被搬迁到新的bucket
	evacuatedX     = 2 // 键值对已经搬迁完毕,key在新buckets数组的前半部分
	evacuatedY     = 3 // 键值对已经搬迁完毕,key在新buckets数组的后半部分
	evacuatedEmpty = 4 // cell为空,整个bucket已经搬迁完毕
	minTopHash     = 5 // tophash的最小正常值

	// flags
	iterator     = 1 // 可能有迭代器在使用buckets
	oldIterator  = 2 // 可能有迭代器在使用oldbuckets
	hashWriting  = 4 // 有协程正在向map写人key
	sameSizeGrow = 8 // 等量扩容

	// 用于迭代器检查的bucket ID
	noCheck = 1<<(8*sys.PtrSize) - 1 // 系统的最大值
)
```

// map 创建 
```cgo
// 如果编译器认为map和第一个bucket可以直接创建在栈上, h和bucket可能都是非空
// 如果h != nil, 那么map可以直接在h中创建.
// 如果h.buckets != nil, 那么h指向的bucket可以作为map的第一个bucket使用.
func makemap(t *maptype, hint int, h *hmap) *hmap {
	// math.MulUintptr返回hint与t.bucket.size的乘积, 并判断该乘积是否溢出.
	mem, overflow := math.MulUintptr(uintptr(hint), t.bucket.size)
	// maxAlloc的值, 根据平台系统的差异而不同，具体计算方式参照src/runtime/malloc.go
	if overflow || mem > maxAlloc {
		hint = 0
	}

	// initialize Hmap
	if h == nil {
		h = new(hmap)
	}
	// 通过fastrand得到哈希种子
	h.hash0 = fastrand()

	// 根据输入的元素个数hint, 找到能装下这些元素的B值
	B := uint8(0)
	//  hint > 8 && uintptr(hint) > bucketShift(B)*6.5
	for overLoadFactor(hint, B) {
		B++
	}
	h.B = B

	// 分配初始哈希表
	// 如果B为0，那么buckets字段后续会在mapassign方法中lazily分配
	if h.B != 0 {
		var nextOverflow *bmap
		// makeBucketArray创建一个map的底层保存buckets的数组，它最少会分配h.B^2的大小。
		h.buckets, nextOverflow = makeBucketArray(t, h.B, nil)
		if nextOverflow != nil {
			h.extra = new(mapextra)
			h.extra.nextOverflow = nextOverflow
		}
	}
	return h
}
```

```cgo
// makeBucket为map创建用于保存buckets的数组. 
func makeBucketArray(t *maptype, b uint8, dirtyalloc unsafe.Pointer) (buckets unsafe.Pointer, nextOverflow *bmap) {
	base := bucketShift(b)
	nbuckets := base
	
	// 对于小的b值(小于4),即桶的数量小于16时,使用溢出桶的可能性很小. 对于此情况, 就避免计算开销. 
	if b >= 4 {
		// 当桶的数量大于等于16个时, 正常情况下就会额外创建2^(b-4)个溢出桶
		nbuckets += bucketShift(b - 4)
		sz := t.bucket.size * nbuckets // 计算内存大小
		up := roundupsize(sz)          // 计算mallocgc将分配的内存块的大小(需要以此为准)
		if up != sz {
			nbuckets = up / t.bucket.size
		}
	}

	// 这里, dirtyalloc 分两种情况. 如果它为nil, 则会分配一个新的底层数组. 
	// 如果它不为nil,则它指向的是曾经分配过的底层数组, 该底层数组是由之前同样的t和b参数通过makeBucketArray分配的,
	// 如果数组不为空,需要把该数组之前的数据清空并复用. 
	if dirtyalloc == nil {
		buckets = newarray(t.bucket, int(nbuckets))
	} else {
		buckets = dirtyalloc
		size := t.bucket.size * nbuckets
		if t.bucket.ptrdata != 0 {
			memclrHasPointers(buckets, size) // 开启了写屏障
		} else {
			memclrNoHeapPointers(buckets, size) // 最终都会调用此方法
		}
	}

	// 即b大于等于4的情况下, 会预分配一些溢出桶. 
	// 为了把跟踪这些溢出桶的开销降至最低, 使用了以下约定:
	// 如果预分配的溢出桶的overflow指针为nil, 那么可以通过指针碰撞(bumping the pointer)获得更多可用桶. 
	// (关于指针碰撞: 假设内存是绝对规整的,所有用过的内存都放在一边,空闲的内存放在另一边,中间放着一个指针作为分界点的
	// 指示器, 那所分配内存就仅仅是把那个指针向空闲空间那边挪动一段与对象大小相等的距离, 这种分配方式称为"指针碰撞")
	// 对于最后一个溢出桶, 需要一个安全的非nil指针指向它. 
	if base != nbuckets {
	    // buckets(基地址) + base(2^B)*bucketsize, 即获得第一个 overflow
		nextOverflow = (*bmap)(add(buckets, base*uintptr(t.bucketsize)))
		// 最后一个 overflow
		last := (*bmap)(add(buckets, (nbuckets-1)*uintptr(t.bucketsize)))
		last.setoverflow(t, (*bmap)(buckets)) // 最后一个 overflow 指针指向 buckets(基地址, 也是安全的指针)
	}
	return buckets, nextOverflow
}
```


// 插入操作
```cgo
func mapassign(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
	// 如果h是空指针,赋值会引起panic
	// 例如以下语句
	// var m map[string]int
	// m["k"] = 1
	if h == nil {
		panic(plainError("assignment to entry in nil map"))
	}
	
	// 如果开启了竞态检测 -race
	if raceenabled {
		callerpc := getcallerpc()
		pc := funcPC(mapassign)
		racewritepc(unsafe.Pointer(h), callerpc, pc)
		raceReadObjectPC(t.key, key, callerpc, pc)
	}
	// 如果开启了memory sanitizer -msan
	if msanenabled {
		msanread(key, t.key.size)
	}
	// 有其他goroutine正在往map中写key, 会抛出以下错误
	if h.flags&hashWriting != 0 {
		throw("concurrent map writes")
	}
	// 通过key和哈希种子, 算出对应哈希值
	hash := t.hasher(key, uintptr(h.hash0))

	// 将flags的值与hashWriting做按位 "异或" 运算
	// 因为在当前goroutine可能还未完成key的写入, 再次调用t.hasher会发生panic.
	h.flags ^= hashWriting

	if h.buckets == nil {
		h.buckets = newobject(t.bucket) // newarray(t.bucket, 1)
	}

again:
    // bucketMask返回值是2的B次方减1
    // 因此,通过hash值与bucketMask返回值做按位与操作,返回的在buckets数组中的第几号桶
	bucket := hash & bucketMask(h.B) // 获取bucket的位置
	// 如果map正在搬迁(即h.oldbuckets != nil)中, 则先进行搬迁工作. 
	if h.growing() {
		growWork(t, h, bucket)
	}
	
	
	// 计算出上面求出的第几号bucket的内存位置
	// post = start + bucketNumber * bucketsize
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.bucketsize)))
	top := tophash(hash)

	var inserti *uint8
	var insertk unsafe.Pointer
	var elem unsafe.Pointer
	
bucketloop:
	for {
		// 遍历桶中的8个cell
		for i := uintptr(0); i < bucketCnt; i++ {
			// 这里分两种情况:
			// 第一种情况是cell位的tophash值和当前tophash值不相等
			// 在 b.tophash[i] != top 的情况下
			// 理论上有可能会是一个空槽位
			// 一般情况下 map 的槽位分布是这样的, e 表示 empty:
			// [h0][h1][h2][h3][h4][e][e][e]
			// 但在执行过 delete 操作时,可能会变成这样:
			// [h0][h1][e][e][h5][e][e][e]
			// 所以如果再插入的话,会尽量往前面的位置插
			// [h0][h1][e][e][h5][e][e][e]
			//          ^
			//          ^
			//       这个位置
			// 所以在循环的时候还要顺便把前面的空位置先记下来
			// 因为有可能在后面会找到相等的key,也可能找不到相等的key
			if b.tophash[i] != top {
				// 如果cell位为空,那么就可以在对应位置进行插入
				if isEmpty(b.tophash[i]) && inserti == nil {
					inserti = &b.tophash[i]
					insertk = add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
					elem = add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
				}
				if b.tophash[i] == emptyRest {
					break bucketloop
				}
				continue
			}
			
			// 第二种情况是cell位的tophash值和当前的tophash值相等
			k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			if t.indirectkey() {
				k = *((*unsafe.Pointer)(k))
			}
			// 注意,即使当前cell位的tophash值相等,不一定它对应的key也是相等的,所以还要做一个key值判断
			if !t.key.equal(key, k) {
				continue
			}
			// 如果已经有该key了,就更新它
			if t.needkeyupdate() {
				typedmemmove(t.key, k, key)
			}
			// 这里获取到了要插入key对应的value的内存地址
			// pos = start + dataOffset + 8*keysize + i*elemsize
			elem = add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
			// 如果顺利到这,就直接跳到done的结束逻辑中去
			goto done
		}
		// 如果桶中的8个cell遍历完,还未找到对应的空cell或覆盖cell,那么就进入它的溢出桶中去遍历
		ovf := b.overflow(t)
		// 如果连溢出桶中都没有找到合适的cell,跳出循环. 
		if ovf == nil {
			break
		}
		b = ovf
	}

	// 在已有的桶和溢出桶中都未找到合适的cell供key写入,那么有可能会触发以下两种情况
	// 情况一:
	// 判断当前map的装载因子是否达到设定的6.5阈值,或者当前map的溢出桶数量是否过多. 如果存在这两种情况之一,则进行扩容操作. 
	// hashGrow()实际并未完成扩容,对哈希表数据的搬迁(复制)操作是通过growWork()来完成的. 
	// 重新跳入again逻辑,在进行完growWork()操作后,再次遍历新的桶. 
	if !h.growing() && (overLoadFactor(h.count+1, h.B) || tooManyOverflowBuckets(h.noverflow, h.B)) {
		hashGrow(t, h)
		goto again // Growing the table invalidates everything, so try again
	}

	// 情况二:
	// 在不满足情况一的条件下,会为当前桶再新建溢出桶,并将tophash,key插入到新建溢出桶的对应内存的0号位置
	if inserti == nil {
		// all current buckets are full, allocate a new one.
		newb := h.newoverflow(t, b)
		inserti = &newb.tophash[0]
		insertk = add(unsafe.Pointer(newb), dataOffset)
		elem = add(insertk, bucketCnt*uintptr(t.keysize))
	}

	// 在插入位置存入新的key和value
	if t.indirectkey() {
		kmem := newobject(t.key)
		*(*unsafe.Pointer)(insertk) = kmem
		insertk = kmem
	}
	if t.indirectelem() {
		vmem := newobject(t.elem)
		*(*unsafe.Pointer)(elem) = vmem
	}
	typedmemmove(t.key, insertk, key)
	*inserti = top
	// map中的key数量+1
	h.count++

done:
	if h.flags&hashWriting == 0 {
		throw("concurrent map writes")
	}
	h.flags &^= hashWriting
	if t.indirectelem() {
		elem = *((*unsafe.Pointer)(elem))
	}
	return elem
}
```

// 搬移操作
```cgo
func growWork(t *maptype, h *hmap, bucket uintptr) {
	// 为了确认搬迁的 bucket 是我们正在使用的 bucket
	// 即如果当前key映射到老的bucket1, 那么就搬迁该bucket1.
	evacuate(t, h, bucket&h.oldbucketmask())

	// 如果还未完成扩容工作，则再搬迁一个bucket.
	if h.growing() {
		evacuate(t, h, h.nevacuate)
	}
}
```

// 查询操作
```cgo
func mapaccess1(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer {
	// 如果开启了竞态检测 -race
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		pc := funcPC(mapaccess1)
		racereadpc(unsafe.Pointer(h), callerpc, pc)
		raceReadObjectPC(t.key, key, callerpc, pc)
	}
	// 如果开启了memory sanitizer -msan
	if msanenabled && h != nil {
		msanread(key, t.key.size)
	}
	
	// 如果map为空或者元素个数为0, 返回零值
	if h == nil || h.count == 0 {
		if t.hashMightPanic() {
			t.hasher(key, 0) // see issue 23734
		}
		return unsafe.Pointer(&zeroVal[0])
	}
	// 注意, 这里是按位与操作
	// 当h.flags对应的值为hashWriting(代表有其他goroutine正在往map中写key)时, 那么位计算的结果不为0, 因此抛出以
	// 下错误. 这也表明, go的map是非并发安全的
	if h.flags&hashWriting != 0 {
		throw("concurrent map read and map write")
	}
	
	// 不同类型的key, 会使用不同的hash算法, 可详见src/runtime/alg.go中typehash函数中的逻辑.
	hash := t.hasher(key, uintptr(h.hash0))
	m := bucketMask(h.B)
	
	// 按位与操作, 找到对应的bucket
	b := (*bmap)(add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
	// 如果oldbuckets不为空, 那么证明map发生了扩容
	// 如果有扩容发生, 老的buckets中的数据可能还未搬迁至新的buckets里
	// 所以需要先在老的buckets中找
	if c := h.oldbuckets; c != nil {
		if !h.sameSizeGrow() {
			m >>= 1
		}
		oldb := (*bmap)(add(c, (hash&m)*uintptr(t.bucketsize)))
		// 如果在oldbuckets中tophash[0]的值, 为evacuatedX, evacuatedY, evacuatedEmpty 其中之一
		// 则evacuated()返回为true, 代表搬迁完成. 
		// 因此, 只有当搬迁未完成时, 才会从此oldbucket中遍历
		if !evacuated(oldb) {
			b = oldb
		}
	}
	// 取出当前key值的tophash值
	top := tophash(hash)
	// 以下是查找的核心逻辑
	// 双重循环遍历: 外层循环是从桶到溢出桶遍历; 内层是桶中的cell遍历
	// 跳出循环的条件有三种: 
	// 第一种是已经找到key值;
	// 第二种是当前桶再无溢出桶;
	// 第三种是当前桶中有cell位的tophash值是emptyRest, 这个值在前面解释过, 它代表此时的桶后面的cell还未利用, 
	// 所以无需再继续遍历. 
bucketloop:
	for ; b != nil; b = b.overflow(t) {
		for i := uintptr(0); i < bucketCnt; i++ {
			// 判断tophash值是否相等
			if b.tophash[i] != top {
				if b.tophash[i] == emptyRest {
					break bucketloop
				}
				continue
			}
			// 因为在bucket中key是用连续的存储空间存储的, 因此可以通过bucket地址+数据偏移量(bmap结构体的大小)+keysize的大小, 
			// 得到k的地址. 同理, value的地址也是相似的计算方法, 只是再要加上8个keysize的内存地址.
			k := add(unsafe.Pointer(b), dataOffset+i*uintptr(t.keysize))
			if t.indirectkey() {
				k = *((*unsafe.Pointer)(k))
			}
			// 判断key是否相等
			if t.key.equal(key, k) {
				e := add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.keysize)+i*uintptr(t.elemsize))
				if t.indirectelem() {
					e = *((*unsafe.Pointer)(e))
				}
				return e
			}
		}
	}
	
	// 所有的bucket都未找到, 则返回零值
	return unsafe.Pointer(&zeroVal[0])
}
```


