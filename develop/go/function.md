## go 底层函数


### typedmemmove() VS memmove()

```
func typedmemmove(typ *_type, dst, src unsafe.Pointer)
```

typedmemmove() 是将类型 _type 的对象从 src 拷贝到 dst, src和dst都是 unsafe.Pointer, 这当中涉及到了内存操作.
由于GC的存在, 在拷贝前, 如果 _type 包含指针, 需要开启写屏障 bulkBarrierPreWrite. 然后调用 memmove() 进行内存
拷贝的操作. 


```
func memmove(dst, src unsafe.Pointer, size uintptr)
```

memmove() 直接是进行内存移动, 使用的是汇编实现的, 代码位置在 `src/runtime/memmove_amd64.s`. 当然了, 这个操作本
身就是单纯的内存操作, 不涉及其他任何内容.


### bulkBarrierPreWrite() VS typeBitsBulkBarrier()

```cgo
func bulkBarrierPreWriteSrcOnly(dst, src, size uintptr)
```

bulkBarrierPreWrite 使用 `[dst, dst + size)` 的 pointer/scalar 信息对内存范围 `[src, src + size)` 中的每
个指针 slot 执行写屏障. 这会在执行 memmove 之前执行必要的写屏障.

1. src, dst 和 size 必须是 pointer-aligned(内存对齐). 

2. 范围 `[dst, dst+size)` 必须位于单个对象内. 

3. 它不执行实际的写操作.

作为一种特殊情况, src == 0 表示正在将其用于memclr. bulkBarrierPreWrite将为每个写障碍的src传递0.


调用方应在调用memmove(dst,src,size) 之前立即调用bulkBarrierPreWrite. 此函数标记为 nosplit, 以避免被抢占; GC不
得停止memmove和执行屏障之间的goroutine. 调用者还负责cgo指针检查, 这可能是将Go指针写入非Go内存中.

对于分配的内存中完全不包含指针的情况(即src和dst不包含指针), 不维护 pointer bitmap; 通常, 通过检查 typ.ptrdata, 
bulkBarrierPreWrite 的任何调用者都必须首先确保基础分配包含指针.



```cgo
func typeBitsBulkBarrier(typ *_type, dst, src, size uintptr)
```

typeBitsBulkBarrier 对将由 memmove 使用 bitmap 类型定位这些指针slot的每个指针从 `[src, src+size)` 复制到
`[dst, dst+size)` 的每个指针执行写屏障.

1. 类型 typ 必须与 `[src, src+size)` 和 `[dst, dst + size)` 完全对应.

2. dst, src 和 size必须是 pointer-aligned(内存对齐).

3. 类型typ必须具有plain bitmap, 而不是GC程序.

此功能的唯一用途是在 chan 发送中, 并且64kB chan元素大小限制为我们解决了这一问题.

不能被抢占, 因为它通常在 memmove 之前运行, 并且GC必须将其视为原子动作.
