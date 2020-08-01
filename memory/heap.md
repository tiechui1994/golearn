### 内存分配

堆上的所有的对象都会通过调用 `runtime.newobject` 函数分配, 该函数会调用 `runtime.mallocgc` 分配指定大小的内存空
间, 这也是用户程序向堆上申请空间的必经函数.

```cgo
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	mp := acquirem()
	mp.mallocing = 1

	shouldhelpgc := false
    dataSize := size
    c := gomcache()
	var x unsafe.Pointer
	noscan := typ == nil || typ.ptrdata == 0
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

将小于16字节的对象划分为微对象, 它会使用线程缓存上的微分配器提高对象分配的性能, 我们主要使用它来分配较小的字符串以及逃逸
的临时变量. 微分配器可以将多个较小的内存分配请求合入同一个内存块中, 只有当内存块中的所有对象都需要被回收时, 整片内存才可
能被回收.

**微分配器管理的对象不可以是指针类型**, 管理多个对象的内存块大小 `maxTinySize` 是可以调整的, 在默认情况下, 内存块的
大小为16字节. `maxTinySize` 的值越大, 结合多个对象的可能性就越高, 内存浪费也就越严重; `maxTinySize` 越小, 内存浪
费就会越小, 不过无论如何调整, 8的倍数是一个很好的选择.

![image](/images/mem_tiny_alloc.png)

微分配器已经在16字节的内存块中分配了12字节的对象, 如果下一个待分配的对象小于4字节, 它就会使用上述的内存块的剩余部分, 减
少内存碎片, 不过该内存块只有3个对象都被标记为垃圾时才会被回收.

线程缓存 `runtime.mcache` 中的 `tiny` 字段指向了 `maxTinySize` 大小的块, 如果当前块中还包含大小合适的空闲内存, 
运行时会通过基地址和偏移量获取并返回这块内存:

```cgo
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	...
	if size <= maxSmallSize {
		if noscan && size < maxTinySize {
			off := c.tinyoffset
			if off+size <= maxTinySize && c.tiny != 0 {
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
			// [1]
			span := c.alloc[tinySpanClass]
			v := nextFreeFast(span)
			if v == 0 {
			    // [2]
				v, _, _ = c.nextFree(tinySpanClass)
			}
			x = unsafe.Pointer(v)
			(*[2]uint64)(x)[0] = 0
			(*[2]uint64)(x)[1] = 0
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
			if size <= smallSizeMax-8 {
				sizeclass = size_to_class8[(size+smallSizeDiv-1)/smallSizeDiv]
			} else {
				sizeclass = size_to_class128[(size-smallSizeMax+largeSizeDiv-1)/largeSizeDiv]
			}
			size = uintptr(class_to_size[sizeclass])
			spc := makeSpanClass(sizeclass, noscan)
			span := c.alloc[spc]
			v := nextFreeFast(span)
			if v == 0 {
				v, span, _ = c.nextFree(spc)
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
	theBit := sys.Ctz64(s.allocCache)
	if theBit < 64 {
		result := s.freeindex + uintptr(theBit)
		if result < s.nelems {
			freeidx := result + 1
			if freeidx%64 == 0 && freeidx != s.nelems {
				return0
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
存, 运行时会通过 `runtime.mcache,nextFree`找到新的内存管理单元:

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

