# Go 函数编译

Go 当中, 有四类函数:

- top-level func

- method with value receiver

- method with pointer receiver

- func literal

`top-level func` 就是平常的普通函数:
```
func Println() {}
```

`method with value receiver` 和 `method with pointer receiver`指的是结构体方法的 '值接收者方法' 和 '指针接收
者方法'.


`func literal` 的定义如下:

> A function literal represents an anonymous function.

也就是说包含匿函数和闭包.

`top-level func` 类型的函数调用比较常见, 也比较简单, 这里不展开分析了.

## 值接收者与指针接收者方法

```cgo
package main

type Value struct {
    X uint64
    Y uint64
}

func (v Value) Vinc(inc uint64) {
    v.X += inc
    v.Y += inc
}

func (v *Value) Pinc(inc uint64) {
    v.X += inc
    v.Y += inc
}

func main() {
    val := Value{X: 2, Y: 5}
    val.Vinc(10)
    val.Pinc(20)
}
```

使用 `go tool compile -N -l -S main.go` 得到汇编代码.

### 调用值接收者(value receiver)方法

- 在汇编层面, 结构体是一段连续内存. 因此 `val := Value{X: 2, Y: 5}` 初始化代码如下:

```
00029 (main.go:19)   XORPS   X0, X0
00032 (main.go:19)   MOVUPS  X0, "".val+24(SP) // 初始化16字节内存
00037 (main.go:19)   MOVQ    $2, "".val+24(SP) // 初始化val.X
00046 (main.go:19)   MOVQ    $5, "".val+32(SP) // 初始化val.Y
```

`XORPS SRC, DST`, 源操作数 SRC 与 目标操作数 DST 进行异或, 结果保存到 DST 当中.

`MOVUPS` 和 `MOVQ` 两者含义相同, 但是 `MOVUPS` 操作数是16字节对齐, `MOVQ` 是8字节对齐.
  

- 接下来是准备调用 `val.Vinc(10)` 的代码:

```
00055 (main.go:20)   MOVQ    $2, (SP)  // 函数调用第一个参数, Value值对象
00063 (main.go:20)   MOVQ    $5, 8(SP)
00072 (main.go:20)   MOVQ    $10, 16(SP) // 函数调用第二个参数, 10
00081 (main.go:20)   CALL    "".Value.Vinc(SB)
```

- 接下来是 `Vinc` 方法的汇编实现:

```
"".Value.Vinc STEXT nosplit size=31 args=0x18 locals=0x0
    00000 (main.go:8)    TEXT    "".Value.Vinc(SB), NOSPLIT|ABIInternal, $0-24
    00000 (main.go:9)    MOVQ    "".v+8(SP), AX    // AX=v.X
    00005 (main.go:9)    ADDQ    "".inc+24(SP), AX // AX+=inc
    00010 (main.go:9)    MOVQ    AX, "".v+8(SP)    // v.X=AX
    00015 (main.go:10)   MOVQ    "".v+16(SP), AX   // AX=v.Y
    00020 (main.go:10)   ADDQ    "".inc+24(SP), AX // AX+=inc
    00025 (main.go:10)   MOVQ    AX, "".v+16(SP)   // v.Y=AX
    00030 (main.go:11)   RET
```

注: 在这里 `0(SP)` 保存的是函数的返回地址. 从 `8(SP) - 32(SP)` 保存的调用的参数(包括返回值, 但是这里方法没有返回值) 

### 调用值接收者(pointer receiver)方法

- 调用 `val.Pinc(20)` 的代码:

```
00086 (main.go:21)   LEAQ    "".val+24(SP), AX // 将 val 指针放入 AX
00091 (main.go:21)   MOVQ    AX, (SP)          // 函数调用第一个参数, Value指针
00095 (main.go:21)   MOVQ    $20, 8(SP)        // 函数调用第二个参数, 10
00104 (main.go:21)   CALL    "".(*Value).Pinc(SB)
```

需要注意的是, `"".val+24(SP)` 当中保存的是已经初始化的 val, 这个在前面的汇编当中已经看到了.

- 接下来是 `Pinc` 方法的汇编代码:

```
"".(*Value).Pinc STEXT nosplit size=53 args=0x10 locals=0x0
    00000 (main.go:13)   TEXT    "".(*Value).Pinc(SB), NOSPLIT|ABIInternal, $0-16
    00000 (main.go:14)   MOVQ    "".v+8(SP), AX
    00005 (main.go:14)   TESTB   AL, (AX)
    00007 (main.go:14)   MOVQ    "".v+8(SP), CX
    00012 (main.go:14)   TESTB   AL, (CX)
    00014 (main.go:14)   MOVQ    (AX), AX
    00017 (main.go:14)   ADDQ    "".inc+16(SP), AX
    00022 (main.go:14)   MOVQ    AX, (CX)
    00025 (main.go:15)   MOVQ    "".v+8(SP), AX
    00030 (main.go:15)   TESTB   AL, (AX)
    00032 (main.go:15)   MOVQ    "".v+8(SP), CX
    00037 (main.go:15)   TESTB   AL, (CX)
    00039 (main.go:15)   MOVQ    8(AX), AX
    00043 (main.go:15)   ADDQ    "".inc+16(SP), AX
    00048 (main.go:15)   MOVQ    AX, 8(CX)
    00052 (main.go:16)   RET
```

整块代码可以分为两部分:

```
00000 (main.go:14)   MOVQ    "".v+8(SP), AX // 将Value指针拷贝给AX
00005 (main.go:14)   TESTB   AL, (AX)
00007 (main.go:14)   MOVQ    "".v+8(SP), CX // 将Value指针拷贝给CX
00012 (main.go:14)   TESTB   AL, (CX)
00014 (main.go:14)   MOVQ    (AX), AX       // 将AX指针指向值的前8字节拷贝给AX
00017 (main.go:14)   ADDQ    "".inc+16(SP), AX // AX+=inc
00022 (main.go:14)   MOVQ    AX, (CX)       // 将AX的值赋值给CX指针指向的对象.
```

`MOVQ (AX), AX`, AX 当中保存的是 Value 指针, 这里就是将 Value 指针指向的值的前8字节拷贝给AX. 类似于 AX=*AX.

`MOVQ AX, (CX)`, CX 当中保存的是 Value 指针, AX 里面是计算的结果. 类似于 *CX=AX.

关于寄存器引用(不包括 SP, BP, SB, PC 伪寄存器).

```
// 加括号代表是指针的引用
MOVQ (AX), BX   // BX = *AX, 将 AX 指向的内存区域 8B 赋值给 BX
MOVQ 16(AX), BX // BX = *(AX+16)

// 不加括号是值的引用
MOVQ AX, BX // BX = AX, 将 AX 中存储的内存拷贝给 BX
```

第二部分: 与第一部分基本上是类似的

```
00025 (main.go:15)   MOVQ    "".v+8(SP), AX // 将Value指针拷贝给AX
00030 (main.go:15)   TESTB   AL, (AX)
00032 (main.go:15)   MOVQ    "".v+8(SP), CX // 将Value指针拷贝给CX
00037 (main.go:15)   TESTB   AL, (CX)
00039 (main.go:15)   MOVQ    8(AX), AX      // AX=*(AX+8)
00043 (main.go:15)   ADDQ    "".inc+16(SP), AX // AX+=inc
00048 (main.go:15)   MOVQ    AX, 8(CX)      // *(CX+8)=AX
```

调用值接收者方法的时候, 调用者caller将参数值写入到栈上, 调用函数callee操作的是调用者caller栈上的参数值.

调用指针接收者方法的时候, 与值接收者方法的区别在于调用者caller写入栈的参数的地址值, 所以调用完成之后可以直接体现在指针的
结构体中.

## 匿名和闭包函数

匿名函数和闭包函数看起来很像, 但是底层的实现却不一样.

### 匿名函数

```cgo
package main

func main() {
    f := func(x uint64) uint64 {
        x += x
        return x
    }
	
    f(200)
}
```

函数调用汇编:

```
00029 (main.go:23)   LEAQ    "".main.func1·f(SB), DX // 拷贝匿名函数地址
00036 (main.go:23)   MOVQ    DX, "".f+16(SP)         // f=DX
00041 (main.go:28)   MOVQ    $200, (SP)              // 调用第一个参数入栈
00049 (main.go:28)   MOVQ    "".main.func1·f(SB), AX // 拷贝匿名函数
00056 (main.go:28)   CALL    AX                      // 函数调用
```

这里需要注意的是 `CALL AX`, AX 的值通过 MOVQ 进行拷贝的.
".main.func1.1

匿名函数汇编:

```
"".main.func1 STEXT nosplit size=30 args=0x10 locals=0x0
    00000 (main.go:23)   TEXT    "".main.func1(SB), NOSPLIT|ABIInternal, $0-16
    00000 (main.go:23)   MOVQ    $0, "".~r1+16(SP) // 初始化函数返回值
    00009 (main.go:24)   MOVQ    "".x+8(SP), AX    // AX=x
    00014 (main.go:24)   ADDQ    "".x+8(SP), AX    // AX+=x
    00019 (main.go:24)   MOVQ    AX, "".x+8(SP)    // x=AX
    00024 (main.go:25)   MOVQ    AX, "".~r1+16(SP) // 设置函数返回值
    00029 (main.go:25)   RET
```

### 闭包函数

闭包函数, 就相对比较复杂一些了.

```cgo
package main

func main() {
    f := func() func() uint64 {
        x := uint64(100)
        return func() uint64 {
            x += 100
            return x
        }
    }()

    f()
    f()
    f()
}
```

- 函数调用代码:

```
00032 (main.go:29)   CALL    "".main.func1(SB)
00037 (main.go:29)   MOVQ    (SP), DX       // 将返回的地址放入DX当中
00041 (main.go:23)   MOVQ    DX, "".f+8(SP) // 返回地址拷贝到栈上, 以便后面调用.
00046 (main.go:31)   MOVQ    (DX), AX       // AX=*DX, 其实质就是 main.func1.1 的函数地址.
00049 (main.go:31)   CALL    AX             // 1
00051 (main.go:32)   MOVQ    "".f+8(SP), DX
00056 (main.go:32)   MOVQ    (DX), AX
00059 (main.go:32)   CALL    AX             // 2
00061 (main.go:33)   MOVQ    "".f+8(SP), DX
00066 (main.go:33)   MOVQ    (DX), AX
00069 (main.go:33)   CALL    AX             // 3
```


- 匿名函数 `main.func1`, 该含义用于返回一个匿名函数:

```
"".main.func1 STEXT size=181 args=0x8 locals=0x28
    00000 (main.go:23)   TEXT    "".main.func1(SB), ABIInternal, $40-8
    00000 (main.go:23)   MOVQ    (TLS), CX
    00009 (main.go:23)   CMPQ    SP, 16(CX)
    00013 (main.go:23)   JLS     171
    00019 (main.go:23)   SUBQ    $40, SP
    00023 (main.go:23)   MOVQ    BP, 32(SP)
    00028 (main.go:23)   LEAQ    32(SP), BP
    00033 (main.go:23)   MOVQ    $0, "".~r0+48(SP)   // 初始化返回值(指针)
    00042 (main.go:24)   LEAQ    type.uint64(SB), AX // 获取 type.uint64 地址
    00049 (main.go:24)   MOVQ    AX, (SP)            // 函数(runtime.newobjec)参数. 
    00053 (main.go:24)   CALL    runtime.newobject(SB) // 创建一个 uint64 指针, 逃逸到堆上.
    00058 (main.go:24)   MOVQ    8(SP), AX             // 获取函数返回值, 一个 *uint64 指针
    00063 (main.go:24)   MOVQ    AX, "".&x+24(SP)      // &x=AX, 将返回的指针保存在栈上
    00068 (main.go:24)   MOVQ    $100, (AX)            // *AX=100, 设置指针指向的值
    
    00075 (main.go:25)   LEAQ    type.noalg.struct { F uintptr; "".x *uint64 }(SB), AX // 匿名结构体地址
    00082 (main.go:25)   MOVQ    AX, (SP)                 // 函数(runtime.newobjec)参数
    00086 (main.go:25)   CALL    runtime.newobject(SB)
    00091 (main.go:25)   MOVQ    8(SP), AX                // 获取函数返回值, 一个 type.noalg.struct 指针
    00096 (main.go:25)   MOVQ    AX, ""..autotmp_4+16(SP) // 将返回的地址保存在栈上
    00101 (main.go:25)   LEAQ    "".main.func1.1(SB), CX  // 获取 main.func1.1 函数地址
    00108 (main.go:25)   MOVQ    CX, (AX)                 // *AX=CX, 设置 type.noalg.struct 指针指向指向值的 F 字段
    00111 (main.go:25)   MOVQ    ""..autotmp_4+16(SP), AX 
    00116 (main.go:25)   TESTB   AL, (AX)
    00118 (main.go:25)   MOVQ    "".&x+24(SP), CX         // CX=&x, *uint64 指针
    00123 (main.go:25)   LEAQ    8(AX), DI                // DX是 type.noalg.struct 指针指向的 x 字段
    00127 (main.go:25)   CMPL    runtime.writeBarrier(SB), $0 // 写屏障
    00134 (main.go:25)   JEQ     138
    00136 (main.go:25)   JMP     164
    00138 (main.go:25)   MOVQ    CX, 8(AX)             // *(AX+8)=CX, 设置 type.noalg.struct 指针指向指向值的 x 字段
    00142 (main.go:25)   JMP     144
    00144 (main.go:25)   MOVQ    ""..autotmp_4+16(SP), AX 
    00149 (main.go:25)   MOVQ    AX, "".~r0+48(SP)     // 设置返回值
    
    00154 (main.go:25)   MOVQ    32(SP), BP
    00159 (main.go:25)   ADDQ    $40, SP
    00163 (main.go:25)   RET
    00164 (main.go:25)   CALL    runtime.gcWriteBarrierCX(SB)
    00169 (main.go:25)   JMP     144
    00171 (main.go:25)   NOP
    00171 (main.go:23)   CALL    runtime.morestack_noctxt(SB)
    00176 (main.go:23)   JMP     0
```

代码分为两大部分, 第一部分是在堆上创建 uint64 指针.

```
00033 (main.go:23)   MOVQ    $0, "".~r0+48(SP)   // 初始化返回值(指针)
00042 (main.go:24)   LEAQ    type.uint64(SB), AX // 获取 type.uint64 地址
00049 (main.go:24)   MOVQ    AX, (SP)            // 函数(runtime.newobjec)参数
00053 (main.go:24)   CALL    runtime.newobject(SB)
00058 (main.go:24)   MOVQ    8(SP), AX           // 获取函数返回值, 一个 *uint64 指针
00063 (main.go:24)   MOVQ    AX, "".&x+24(SP)    // &x=AX, 将返回的指针保存在栈上
00068 (main.go:24)   MOVQ    $100, (AX)          // *AX=100, 设置指针指向的值
```

> 注: 在闭包当中, 有关闭包的所有的参数都会使用 `newobject` 去分配成一个指针, 从而逃逸到堆上. 这样做的目的除了, 将闭包
参数的生命周期延长外, 还有就是很容易管理. 

当此部分代码完成之后, 栈上的数据如下:

![image](/images/develop_runtime_closure_1.png)

> 注: 栈的地址是由高地址向低地址增长, SP指向栈顶(低地址处)


第二部分是在堆上创建 `type.noalg.struct {F uintptr; x *uint64 }` 指针, 并将指针作为返回值返回.

```
00075 (main.go:25)   LEAQ    type.noalg.struct { F uintptr; "".x *uint64 }(SB), AX // 匿名结构体地址
00082 (main.go:25)   MOVQ    AX, (SP)             // 函数(runtime.newobjec)参数
00086 (main.go:25)   CALL    runtime.newobject(SB)
00091 (main.go:25)   MOVQ    8(SP), AX            // 获取函数返回值, 一个 type.noalg.struct 指针
00096 (main.go:25)   MOVQ    AX, ""..autotmp_4+16(SP) // 将返回的地址保存在栈上
00101 (main.go:25)   LEAQ    "".main.func1.1(SB), CX  // 获取 main.func1.1 函数地址
00108 (main.go:25)   MOVQ    CX, (AX)                 // *AX=CX, 设置 type.noalg.struct 指针指向指向值的 F 字段
00111 (main.go:25)   MOVQ    ""..autotmp_4+16(SP), AX // *uint64
00116 (main.go:25)   TESTB   AL, (AX)
00118 (main.go:25)   MOVQ    "".&x+24(SP), CX        // CX=&x
00123 (main.go:25)   LEAQ    8(AX), DI               // DX是 type.noalg.struct 指针指向的 x 字段
00127 (main.go:25)   CMPL    runtime.writeBarrier(SB), $0 // 对 F 字段进行写屏障保护
00134 (main.go:25)   JEQ     138
00136 (main.go:25)   JMP     164
00138 (main.go:25)   MOVQ    CX, 8(AX) // *(AX+8)=CX, 设置 type.noalg.struct 指针指向指向值的 x 字段
00142 (main.go:25)   JMP     144
00144 (main.go:25)   MOVQ    ""..autotmp_4+16(SP), AX 
00149 (main.go:25)   MOVQ    AX, "".~r0+48(SP)  // 设置返回值
```

当此部分代码完成之后, 栈上的数据如下:

![image](/images/develop_runtime_closure_2.png)

`runtime.writeBarrier` 是进行数据进行写屏障保护.


- 匿名函数 `main.func1.1`:

```
"".main.func1.1 STEXT nosplit size=71 args=0x8 locals=0x10
    00000 (main.go:25)   TEXT    "".main.func1.1(SB), NOSPLIT|NEEDCTXT|ABIInternal, $16-8
    00000 (main.go:25)   SUBQ    $16, SP
    00004 (main.go:25)   MOVQ    BP, 8(SP)
    00009 (main.go:25)   LEAQ    8(SP), BP
    
    00014 (main.go:25)   MOVQ    8(DX), AX     // 获取 AX=*(DX+8), 也就是 *uint64 指针
    00018 (main.go:25)   MOVQ    AX, "".&x(SP) // 将 AX 当中记录 *uint64 地址写入栈中
    00022 (main.go:25)   MOVQ    $0, "".~r0+24(SP) // 初始化函数返回值.
    00031 (main.go:26)   MOVQ    "".&x(SP), AX     // AX=&x
    00035 (main.go:26)   MOVQ    (AX), AX          // 获取指针当中的值
    00038 (main.go:26)   MOVQ    "".&x(SP), CX     // AX=&x
    00042 (main.go:26)   ADDQ    $100, AX          // 进行增加操作, AX+=100
    00046 (main.go:26)   MOVQ    AX, (CX)          // *CX=AX, 回写
    00049 (main.go:27)   MOVQ    "".&x(SP), AX     // AX=&x 
    00053 (main.go:27)   MOVQ    (AX), AX          // 获取指针当中的值
    00056 (main.go:27)   MOVQ    AX, "".~r0+24(SP) // 函数返回值.
    
    00061 (main.go:27)   MOVQ    8(SP), BP
    00066 (main.go:27)   ADDQ    $16, SP
    00070 (main.go:27)   RET
```

注: 在调用 main.func1.1 函数的时候, DX 被设置为指向 `type.noalg.struct` 的指针, 这个是操作的前提条件.

匿名函数是闭包的一种, 只是没有传递变量而已. 在闭包的调用当中, 会将上下文信息逃逸到堆上, 避免栈帧调用结束而被回收.

闭包的实现所做的事情(依赖某个寄存器):

1) 创建上下文信息内存块(逃逸到堆上)

2) 将上下文信息地址返回, 并保存到某个寄存器上.

3) 在调用返回的函数的时候, 从寄存器上获取函数地址, 然后进行 `CALL`, 在函数当中, 从寄存器上获取上下文信息, 然后进行相关
操作.


指针操作的操作: `读取指针 -> 读取指针的值 -> 值操作`, `读取指针 -> 指针赋值`

```
MOVQ 8(SP), AX
MOVQ 8(SP), CX

MOVQ (AX), AX
SUBQ $100, AX

MOVQ AX, (CX)
```
