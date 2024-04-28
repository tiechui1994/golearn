## Go 底层函数

### typedmemmove() vs memmove()

```cgo
// src/runtime/mbarrier.go
func typedmemmove(typ *_type, dst, src unsafe.Pointer)
```

typedmemmove() 是将类型 _type 的对象从 src 拷贝到 dst, src和dst都是 unsafe.Pointer, 这当中涉及到了内存操作.
由于GC的存在, 在拷贝前, 如果 _type 包含指针, 需要开启写屏障 bulkBarrierPreWrite. 然后调用 memmove() 进行内存
拷贝的操作. 


```cgo
// src/runtime/stubs.go
func memmove(dst, src unsafe.Pointer, size uintptr)
```

memmove() 直接是进行内存移动, 使用的是汇编实现的, 代码位置在 `src/runtime/memmove_amd64.s`. 当然了, 这个操作本身
就是单纯的内存操作, 不涉及其他任何内容.

### bulkBarrierPreWrite() vs typeBitsBulkBarrier()

这里需要先了解一下 go 的内存布局情况. 参考 `虚拟内存布局 > 线性内存` 部分. 简单说, 在 go 当中, 内存分为 spans, bitmap,
area 三部分. 其中 spans 管理页(page, page是对arena的粗粒度划分)的,  bitmap是保存arena区域保存了哪些对象,并且对象
中哪些地址包含了指针. 

一个 bitmap 指针(8bit) 可以表示 4 个 arena(也就是堆区域)指针. 即 2bit bitmap代表1个 arena 指针. 

结论: (arena/页大小)*指针大小 = spans.  (arena/指针大小)/4 = bitmap 


```cgo
// src/runtime/mbitmap.go
func bulkBarrierPreWrite(dst, src, size uintptr)
```

bulkBarrierPreWrite 使用 `[dst, dst + size)` 的 pointer/scalar 信息对内存范围 `[src, src + size)` 中的每
个指针 slot 执行写屏障. 这会在执行 memmove 之前执行必要的写屏障.

1. src, dst 和 size 必须是 pointer-aligned(内存对齐). 

2. 范围 `[dst, dst+size)` 必须位于单个对象内. 

3. 它不执行实际的写操作.

作为一种特殊情况, src == 0 表示正在将其用于 memclr. bulkBarrierPreWrite 将为每个写障碍的src传递0.


调用方应在调用 memmove(dst,src,size) 之前立即调用 bulkBarrierPreWrite. 此函数标记为 nosplit, 以避免被抢占; GC
不得停止memmove和执行屏障之间的goroutine. 调用者还负责cgo指针检查, 这可能是将Go指针写入非Go内存中.

对于分配的内存中完全不包含指针的情况(即src和dst不包含指针), 不维护 pointer bitmap; 通常, 通过检查 typ.ptrdata, 
bulkBarrierPreWrite 的任何调用者都必须首先确保基础分配包含指针.


```cgo
// src/runtime/mbitmap.go
func typeBitsBulkBarrier(typ *_type, dst, src, size uintptr)
```

typeBitsBulkBarrier 对将由 memmove 使用 bitmap 类型定位这些指针slot的每个指针从 `[src, src+size)` 复制到
`[dst, dst+size)` 的每个指针执行写屏障.

1. 类型 typ 必须与 `[src, src+size)` 和 `[dst, dst + size)` 完全对应.

2. dst, src 和 size必须是 pointer-aligned(内存对齐).

3. 类型typ必须具有plain bitmap, 而不是GC程序.

此功能的唯一用途是在 chan 发送中, 并且64kB chan元素大小限制为我们解决了这一问题.

不能被抢占, 因为它通常在 memmove 之前运行, 并且GC必须将其视为原子动作.


```cgo
// src/runtime/mbitmap.go
func bulkBarrierPreWriteSrcOnly(dst, src, size uintptr)
```

bulkBarrierPreWriteSrcOnly 类似于 bulkBarrierPreWrite, 但它不对 `[dst, dst+size)` 范围执行写屏障.

除了 bulkBarrierPreWrite 的要求外, 调用者需要确保区间 `[dst, dst+size)` 是0.

这用于特殊情况, 例如 dst 刚刚创建并使用 malloc 归零.


```cgo
// src/runtime/mbitmap.go
func bulkBarrierBitmap(dst, src, size, maskOffset uintptr, bits *uint8)
```

bulkBarrierBitmap 执行 1-bit pointer bitmap 执行写屏障. 用于从 `[src, src+size)` 复制到 `[dst, dst+size)`.
假定 src 将 maskOffset 字节开始到 bitmap 覆盖的数据中, 已\以bit为单位(可能不是8的倍数).

> 被 bulkBarrierPreWrite 用于写入数据和BSS.



## runtime 常用函数

### acquirem(), releasem()

acquirem(), 获取当前的 m(禁止抢占)

releasem(), 释放 m, 如果发现存在抢占请求(`_g_.m.locks`为0的前提下), 则设置抢占标记.

acquirem() 和 releasem() 这一对函数, 可以看做锁的获取和释放. 锁定的内容是不能进行抢占请求.

```cgo
// src/runtime/runtime1.go

//go:nosplit
func acquirem() *m {
    _g_ := getg()
    _g_.m.locks++
    return _g_.m
}

//go:nosplit
func releasem(mp *m) {
    _g_ := getg()
    mp.locks--
    if mp.locks == 0 && _g_.preempt {
        // restore the preemption request in case we've cleared it in newstack
        // 设置抢占标记 stackguard0, 这样在进入下一次执行前需要检查 stackguard0 的值, 从而进行抢占.
        _g_.stackguard0 = stackPreempt
    }
}
```


### acquirep(), releasep()

acquirep(), 当前 m 与 p 进行绑定, 同时将 p 的状态设置为 _Pidle -> _Prunning.

releasep(), 当前 m 与 p 进行解绑, 同时将 p 的状态由 _Prunning -> _Pidle

```cgo
func acquirep(_p_ *p) {
    // Do the part that isn't allowed to have write barriers.
    wirep(_p_)

    // Have p; write barriers now allowed.

    // Perform deferred mcache flush before this P can allocate
    // from a potentially stale mcache.
    _p_.mcache.prepareForSweep()

    if trace.enabled {
        traceProcStart()
    }
}

func wirep(_p_ *p) {
    _g_ := getg()

    if _g_.m.p != 0 {
        throw("wirep: already in go")
    }
    if _p_.m != 0 || _p_.status != _Pidle {
        id := int64(0)
        if _p_.m != 0 {
            id = _p_.m.ptr().id
        }
        print("wirep: p->m=", _p_.m, "(", id, ") p->status=", _p_.status, "\n")
        throw("wirep: invalid p state")
    }
    _g_.m.p.set(_p_)
    _p_.m.set(_g_.m)
    _p_.status = _Prunning
}

func releasep() *p {
    _g_ := getg()

    if _g_.m.p == 0 {
        throw("releasep: invalid arg")
    }
    _p_ := _g_.m.p.ptr()
    if _p_.m.ptr() != _g_.m || _p_.status != _Prunning {
        print("releasep: m=", _g_.m, " m->p=", _g_.m.p.ptr(), " p->m=", hex(_p_.m), " p->status=", _p_.status, "\n")
        throw("releasep: invalid p state")
    }
    if trace.enabled {
        traceProcStop(_g_.m.p.ptr())
    }
    _g_.m.p = 0
    _p_.m = 0
    _p_.status = _Pidle
    return _p_
}
```