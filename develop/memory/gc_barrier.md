## Go 垃圾回收 - 屏障技术

Go 垃圾扫描时, 为了更快, 需要业务与扫描并发进行, 这时候就会可能出现下面的情况:

初始状态:

![image](/images/develop_gc_why_barrier1.png)

业务与扫描并发执行:

![image](/images/develop_gc_why_barrier2.png)

结果:

![image](/images/develop_gc_why_barrier3.png)

这样就会导致程序出 bug 了.

如何解决上述问题, 就需要使用到了 "内存屏障" 技术.

**内存屏障本质是什么?**

**1) 内存屏障只是对应一段特殊的代码, 该代码是咋编译期间生成**

**2) 内存屏障本质上是在运行期间拦截内存写操作, 相当于一个 hook 调用**

内存屏障的作用是通过 hook 内存的写操作时机, 做一些标记工作, 从而保证垃圾回收的正确性.

在 Go 当中涉及了三个写屏障, 分别是插入写屏障, 删除写屏障, 混合写屏障.

### 内存逃逸与写屏障

先看一段代码 gc.go:
```
package main

type user struct {
    name string
    age  *int
}

func fun1(u *user) {
    u.age = new(int)
}

func fun2(u *user) {
    var b = new(user)
    b.name = "google"
}

func main() {
    u1 := new(user)
    u2 := new(user)
	
    go fun1(u1)
    go fun2(u2)
}
```

内存逃逸的原则:
1. 在保证程序的正确性的前提下, 尽可能把对象分配到栈上, 这样性能最好; 栈上的对象生成周期跟随 goroutine, 协程结束了,
它就没了.

2. 明确一定要分配到堆上的对象, 或者不确定是否要分配在堆上的对象, 那么就全部分配到堆上; 这种对象的生命周期始于业务程序
的创建, 终于垃圾回收器的回收.

进行内存逃逸分析的结果如下: go1.17
```
# command-line-arguments
./gc.go:8:11: u does not escape
./gc.go:9:13: new(int) escapes to heap
./gc.go:12:11: u does not escape
./gc.go:13:13: new(user) does not escape
./gc.go:18:11: new(user) escapes to heap
./gc.go:19:11: new(user) escapes to heap
```

代码当中一共 4 次使用了 new, 但只有 3 次逃逸到了堆.

```
- 'u.age = new(int)' 逃逸到堆上, 因为其生命周期需要超越该 goroutine

- 'var b = new(user)' 分配到栈上, b 对象就只是在栈上使用

- 'u1 := new(user)', 逃逸到堆上, 因为其他的 goroutine 要使用该参数

- 'u2 := new(user)', 逃逸到堆上, 这种属于无法很确定的类型
```

func1 的代码汇编 'go tool compile -S ./gc.go'

```
"".fun1 STEXT size=98 args=0x8 locals=0x18 funcid=0x0
    0x0000 00000 (gc.go:8)  TEXT    "".fun1(SB), ABIInternal, $24-8
    0x0000 00000 (gc.go:8)  CMPQ    SP, 16(R14)
    0x0004 00004 (gc.go:8)  JLS     79
    0x0006 00006 (gc.go:8)  SUBQ    $24, SP
    0x000a 00010 (gc.go:8)  MOVQ    BP, 16(SP)
    0x000f 00015 (gc.go:8)  LEAQ    16(SP), BP
    0x0014 00020 (gc.go:8)  MOVQ    AX, "".u+32(SP)
    0x0019 00025 (gc.go:9)  LEAQ    type.int(SB), AX
    0x0020 00032 (gc.go:9)  CALL    runtime.newobject(SB) # new(int) 对应调用的函数
    0x0025 00037 (gc.go:9)  MOVQ    "".u+32(SP), CX       # CX = &u
    0x002a 00042 (gc.go:9)  TESTB   AL, (CX)
    0x002c 00044 (gc.go:9)  CMPL    runtime.writeBarrier(SB), $0 # 屏障判断, 全局变量 runtime.writeBarrier
    0x0033 00051 (gc.go:9)  JNE     59                           # 不相等, 开启内存屏障, 跳转到 59 行
    0x0035 00053 (gc.go:9)  MOVQ    AX, 16(CX)                   # 相等, 未开启内存屏障, 进行赋值
    0x0039 00057 (gc.go:9)  JMP     69
    0x003b 00059 (gc.go:9)  LEAQ    16(CX), DI                   # 获取 u.age 的指针地址, DI
    0x003f 00063 (gc.go:9)  NOP
    0x0040 00064 (gc.go:9)  CALL    runtime.gcWriteBarrier(SB)   # 内存屏障 hook, 汇编代码实现. 使用寄存器传递参数
    0x0045 00069 (gc.go:10) MOVQ    16(SP), BP
    0x004a 00074 (gc.go:10) ADDQ    $24, SP
    0x004e 00078 (gc.go:10) RET
    0x004f 00079 (gc.go:10) NOP
    0x004f 00079 (gc.go:8)  MOVQ    AX, 8(SP)
    0x0054 00084 (gc.go:8)  CALL    runtime.morestack_noctxt(SB)
    0x0059 00089 (gc.go:8)  MOVQ    8(SP), AX
    0x005e 00094 (gc.go:8)  NOP
    0x0060 00096 (gc.go:8)  JMP     0
```

上述进行内存屏障的伪代码如下:

```
if runtime.writeBarrier.enabled != false {
    runtime.gcWriteBarrier(ptr, val)
} else {
    *ptr = val
}
```

接下来可以研究下 `runtime.gcWriteBarrier` 代码, 其使用的是汇编代码实现 `asm_amd64.s`

```
// gcWriteBarrier 没有遵循 Go ABI(使用栈传递参数), 使用了2个寄存器传递参数:
// - DI is the destination of the write
// - AX is the value being written at DI
TEXT runtime·gcWriteBarrier<ABIInternal>(SB),NOSPLIT,$112
    // Save the registers clobbered by the fast path. This is slightly
    // faster than having the caller spill these.
    MOVQ    R12, 96(SP)
    MOVQ    R13, 104(SP)
    // TODO: 考虑将 g.m.p 作为参数传递, 以便它们可以在一系列写入屏障之间共享.
    get_tls(R13)
    MOVQ    g(R13), R13
    MOVQ    g_m(R13), R13
	
    MOVQ    m_p(R13), R13
    MOVQ    (p_wbBuf+wbBuf_next)(R13), R12  # R12 = p.wbBuf.next
    // 增加 wbBuf.next 的位置 
    LEAQ    16(R12), R12                    # R12 = p.wbBuf.next 偏移 16 bit
    MOVQ    R12, (p_wbBuf+wbBuf_next)(R13)  # p.wbBuf.next = R12, 分配了 2 Byte
    CMPQ    R12, (p_wbBuf+wbBuf_end)(R13)   # p.wbBuf.next == p.wbBuf.end 
    // 记录写屏障操作, 这里是使用的是当前的混合写屏障技术.
    MOVQ    AX, -16(R12)    // Record value, 插入写屏障, 第2个字节
    MOVQ    (DI), R13       // 传入的第一个参数 DI, 指针
    MOVQ    R13, -8(R12)    // Record *slot, 删除写屏障, 第1个字节
    // (flags set in CMPQ above)
    JEQ    flush                              # wbBuf 已经满了, 需要进行刷新 wbBuf 了
ret:
    MOVQ    96(SP), R12
    MOVQ    104(SP), R13
    // 执行写操作
    MOVQ    AX, (DI)                       # 赋值
    RET

flush:
    ...
    // This takes arguments DI and AX
    CALL    runtime·wbBufFlush(SB)        # 批量刷新操作, 最终调用 wbBufFlush1, 将 mspan.gcmarkBits 进行标记 
    ...
    JMP    ret
```

触发了写屏障之后, 核心目的是为了能够把赋值的前后两个值记录下来, 以便 GC 垃圾回收器能得到通知, 从而避免错误的回收.
记录下来是最本质的, 但是并不是要立马处理, 所以这里做的优化就是, 攒满一个 buffer, 然后批量处理, 这样效率会非常高的.

总结下, 写屏障到底做了什么?

- hook 写操作
- hook 写操作后, 把赋值语句的前后的值都记录下来, 存储在 p.wbBuf 队列
- 当 p.wbBuf 满了之后, 批量刷写到扫描队列(置灰), 也就是将 mspan.gcmarkBits 相关位置进行标记

### 插入写屏障

**对象丢失的必要条件**

之前有提到三色标记法, 如果要想出现对象丢失(错误的回收) 那么必须是同时满足两个条件:

条件1: 赋值器把白色对象的引用写入给黑色对象了(即, 黑色对象指向白色对象了)

条件2: 从灰色对象出发, 最终到达该白色对象的所有路径都被赋值器破坏(换句话说, 这个已经被黑色指向的白色对象, 没有在灰色
对象的保护下)

举个例子:

![image](/images/develop_gc_lostobject.png)

1) 插入写操作: X -> Z
2) 删除写操作: Y -> null
3) 回收操作: scan Y
4) 回收操作: 回收 Z (这就是问题了)

在这个两个条件同时出现的时候, 才会出现对象被错误的回收. 如果破坏了第一个条件, 那么就会消除垃圾错误的回收了, 于是就有了
插入写屏障.

**插入写屏障**

```
writePointer(slot, ptr):
    shade(ptr)
    *slot = ptr
```

> `shade(ptr)` 会将尚未变成灰色或黑色的指针 ptr 标记为灰色. 通过保守的假设 *slot 可能会变为黑色, 并确保 ptr 不会
> 在将赋值为 *slot 前变为白色, 进而确保了强三色不变性.
>
> 不足:
> 1) 由于栈上的对象无写屏障(不hook), 那么导致黑色的栈可能指向白色的堆对象, 所以必须假设赋值器是灰色赋值器, 扫描结束
> 后, 必须 STW 重新扫描栈才能确保不丢对象;
> 2) STW 重新扫描栈, 若 goroutine 量大且活跃的场景, 延迟不可控. 经验值平均 10-100ms

使用了插入写屏障后:

![image](/images/develop_gc_insertwb.png)

### 删除写屏障

三色不变式

强三色: 不允许黑色对象指向白色对象

弱三色: 允许黑色对象指向白色对象, 但必须保证一个前提, 这个白色对象必须处于灰色对象的保护下.

强三色不变式要求黑色赋值器的根只能引用灰色或者黑色对象, 不能引用白色对象(因为黑色赋值器不再被扫描, 引用白色).
弱三色不变式允许黑色赋值器的根引用白色对象, 但前提是白色对象必须处于灰色保护下.

获取赋值器的快照, 意味着回收器需要扫描其根并将其着为黑色. 必须在回收起始阶段完成赋值器快照的获取, 并保证其不持有任何
白色对象. 否则一旦赋值器持有某白色对象的唯一引用并将其写入黑色对象, 然后再删除该指针, 则会违背弱三色不变式的要求. 为
黑色对象增加写屏障可以捕捉这一内存写操作, 但如此依赖, 该方案将退化到强三色不变式的框架下. 因此, 基于其快照的解决方案
将只允许黑色赋值器的存在.

回到之前的那个问题, 如果破坏了第二个条件, 那么就会消除垃圾错误的回收了, 于是就有了删除写屏障.

```
writePointer(slot, ptr)
    shade(*slot)
    *slot = ptr
```

在对象的引用(slot)被删除时, 将白色的对象的引用涂为黑色, 这样删除写屏障就可以保证弱三色不变性, 对象引用的下游对象一定
可以被灰色对象保护.

> 1) 删除写屏障也叫基于快照的写屏障, 必须在开始时, STW 扫描整个栈(注意, 是所有的goroutine栈), **保证所有堆上的在用的
> 对象处于灰色保护下, 保证弱三色不变性**
> 2) 由于起始快照的原因, 起始是执行STW, 删除写屏障不适用于栈特别大的场景, 栈越大, STW 扫描时间越长.
> 3) 删除写屏障会导致扫描进度(波面)后退, 所以扫描精度不如插入写屏障.


使用了删除写屏障后:

![image](/images/develop_gc_insertwb.png)


思考: 如果不整机暂停 STW 栈, 而是一个栈一个栈的快照, 这样也没有 STW 了, 是否可以满足要求? (这个就是当前 golang 混合
写屏障的时候做的, 虽然没有 STW 了, 但是扫描到某一个具体的栈的时候, 还是要暂停这一个 goroutine 的)

不行, 纯粹的删除写屏障, 起始必须整个栈打快照, 要把所有的堆对象都处于灰色保护中才行. 举例如下?

初始状态:

1. A 是 g1 栈的一个对象, g1 已结扫描完了, 并且 C 也是扫黑了的对象.
2. B 是 g2 栈的对象, 指向了 C 和 D, g2 完全还没扫描, B 是一个灰色对象, D 的白色对象.

![image](/images/develop_gc_why_deletewb1.png)

操作一: g2 进行赋值变更, 把 C 指向 D 对象, 这个时候黑色的 C 就指向了白色的 D (由于是删除写屏障, 这里不会触发 hook)

操作二: 把 B 指向 C 的引用删除, 由于是栈对象操作, 不会触发删除写屏障

![image](/images/develop_gc_why_deletewb2.png)

操作三: 清理, 因为 C 已经是黑色对象了, 所以不会再扫描, 因此 D 就会被错误的清理掉.

解决上述问题的方法:

方法1: 栈对象也 hook, 所有对象赋值(插入, 删除)都 hook. 这种方法可以解决问题, 但是对于栈, 寄存器的赋值 hook 是不现
实的.

方法2: 加入插入写屏障逻辑, C 指向 D 的时候, 将 D 置灰, 这样扫描没有问题. 也可以去掉起始 STW 扫描, 从而可以并发, 一
个一个栈扫描. **这种方式就是当前的混合写屏障 = 删除写屏障 + 插入写屏障**

```
writePointer(slot, ptr)
    shade(*slot) # 删除
    shade(ptr)   # 插入
    *slot = ptr
```

> 1) 混合写屏障继承了插入写屏障的优点, 起始无需 STW 快照, 直接并发扫描垃圾即可;
> 2) 混合写屏障继承了删除写屏障的优点, 赋值器是黑色赋值器, 扫描一遍就不需要再扫描了, 这样就消除了插入写屏障时期最后 STW
> 的重新扫描栈;
> 3) 混合写屏障扫描精度继承了删除写屏障, 比插入写屏障更低, 随着带来的是 GC 过程全程无 STW(注, 扫描某一个具体的栈的
> 时候, 是要停止 goroutine 赋值器工作的)

### GC 当中置灰

三色(白, 灰, 黑)三色是抽象出来的概念, 所谓的三色标记法, 在 Go 内部对象并没有保存颜色的属性, 三色对它们只是状态的描述,
是通过一个队列+掩码位图来实现的:

- 白色对象: 对象所在 mspan 的 gcmarkBits 中对应的 bit 为 0, 不在队列;
- 灰色对象: 对象所在 mspan 的 gcmarkBits 中对应的 bit 为 1, 且对象在扫描队列中;
- 黑色对象: 对象所在 mspan 的 gcmarkBits 中对应的 bit 为 1, 且对象已经从扫描队列中处理并被删除;


置灰的代码参考 mwbbuf.go 当中的 `wbBufFlush1` 函数.

通过 ptr 可以快速定位到 mspan(findObject), 从而快速查找到 gcmarkBits, 检查相关的 bit 位.

每个 p 持有一个 gcWork 对象, 它持有两个队列 wbuf1 和 wbuf2, 该队列当中保存了待扫描的灰色对象, 扫描线程会从队列当中
获取要扫描的对象(tryGet(), 一旦获取不到, 会从 work.full 当中获取); 一旦其中的某个队列满了, 它会将当前队列当中的对象 
push 到全局的 work.full 当中. 这套机制类似 goroutine 的调度. 

### GC 过程

实现过程在 mgc.go 当中

Go 语言的垃圾收集可以分为`清除终止`, `标记`, `标记终止`, 和 `清除` 四个不同阶段, 它们分别完成了不同的工作:

1. 清理终止阶段;

- 启动后台 MarkWorker 协程.

- **暂停程序**, 所有的处理器在这时会进入安全点(Safe point);

- 如果当前垃圾收集循环是强制触发的, 还需要处理还未被清理的内存管理单元; 清理 sync.Pool

2. 标记阶段 (_GCmark);

- 将状态切换至 `_GCmark`, 开启写屏障, 用户程序协助(Mutator Assiste)并将根对象入队, work 全局变量初始化, p 当中 tiny 标记为灰色;

- **恢复执行程序**, 标记进程和用户协助的用户程序会开始并发标记内存中的对象, **写屏障会将被覆盖的指针和新指针都标记成灰色, 而
所有新创建的对象都会直接标记成黑色**;

- 开始扫描根对象, 包括所有的 Goroutine 的栈, 全局对象以及不在堆中的运行时数据结构, **扫描 Goroutine 栈期间会暂停当
前处理器**;

- 依次处理灰色队列中的对象, 将对象标记成黑色并将它们指向的对象标记成灰色;

- 使用分布式的终止算法检查剩余的工作, 发现标记阶段完成后进入标记终止阶段; 

- **暂停程序**, 唤醒所有的休眠的辅助标记协程

3. 标记终止阶段(_GCmarktermination);

-  将状态切换至 `_GCmarktermination` 并关闭辅助标记的用户程序;

- 清理处理器上的线程缓存;

4. 清理阶段;

- 将状态切换至 `_GCoff` 开始清理阶段, 初始化清理状态并关闭写屏障;

- **恢复用户程序**, 所有新创建的对象会标记成白色;

- 后台并发清理所有的内存管理单元, 当 Goroutine 申请新的内存管理单元时就会触发清理;



