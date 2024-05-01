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


### 删除写屏障