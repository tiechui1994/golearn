# Go 汇编指令基础

> 小知识点: 在 Linux 当中 `·` 怎么输入?
> 按下快捷键: Ctrl+Shift+U, 然后输入: 00B7 或者 B7, 回车之后 `·` 就显示在 Shell 终端了.
> [中点号相关的文档](https://zh.wikipedia.org/zh-cn/%E9%97%B4%E9%9A%94%E5%8F%B7)

几个概念

- 栈: 进程, 线程, goroutine 都有字节的调用栈, 先进先出(FIFO)
- 栈帧: 函数调用时, 在栈上为函数所分配的内存区域
- 调用者: caller, 比如: A函数调用了B函数, 那么 A 就是调用者
- 被调者: callee, 比如: A函数调用了B函数, 那么 B 就是被调者

### Go 汇编中的伪寄存器

![image](/images/develop_assembly_registers.png)

> 在 AMD64 环境, 伪PC寄存器其实是IP指令计数寄存器的别名. 伪FP寄存器对应的是函数的帧指针, 一般来访问函数的参数和返回值.
> 伪SP栈指针对应的是当前函数栈的底部(不包括参数和返回值部分), 一般用于定位局部变量. 伪SP是一个比较特殊的寄存器, 因为还存
> 在一个同名的SP真寄存器. 真SP寄存器对应的是栈的顶部, 一般用于定位调用其他函数的参数和返回值.


| 寄存器 | 说明 |
| ---- | ---- | 
| SB(Static base pointer) | **global symbols** |
| FP(Frame pointer) | **arguments and locals** |
| PC(Program counter | **jumps and branches**  |
| SP(Stack pointer) | **top of stack, function local var** |

- FP: 使用如 `symbol+offset(FP)` 的方式, **引用 callee 函数的入参参数**. 例如 `arg0+0(FP)`, `arg0+8(FP)`,
使用 FP 必须加 symbol, 否则无法通过编译(symbol主要为了提升代码可读性, 实际没啥用)

> 注意: 编写 go 汇编代码时, 要站在 callee 的角度看 (FP), 在 callee 来看, (FP) 指向的是 caller 调用 callee 时
传递的第一个参数的位置. 加入当前的 callee 函数是 add, 在 add 的代码中引用FP, 该FP指向的位置是在 caller 的 stack 
frame 上, 指向调用 add 函数时传递的第一个参数的位置. **经常在 callee 中使用 `symbol+offset(FP)` 来获取入参的参
数值**.

- SB: 全局静态基指针, 一般用在声明函数, 全局变量中.

- SP: 栈指针, 具有迷惑性, 因为会有 `伪SP寄存器` 和 `硬件SP寄存器` 之分. plan9 的这个 `伪SP寄存器` 指向当前栈帧第一
个局部变量的结束位置(为啥? 看寄存器内存布局图). 使用形如 `symbol+offset(SP)` 的形式, 表示函数的局部变量. 如果局部变
量是8字节, 那么第一个局部变量就可以使用 localvar0-8(SP) 来表示. 与 `硬件SP寄存器` 是两个不同的东西, 在栈帧为 0
的情况下, `伪SP寄存器` 和 `硬件SP寄存器` 指向同一位置. 手写汇编时, 如果是 `symbol+offset(SP)` 则表示 `伪寄存器SP`,
如果是 `offset(SP)` 则表示 `硬件SP寄存器`.

> 注意: 对于编译输出 (`go tool compile -S / go tool objdump`) 的代码来讲, 所有的 SP 都是 `硬件SP寄存器`. 无论
是否带 symbol

- PC: 实质上就是体系结构当中的 pc 寄存器. 在 x86 平台对应 ip 寄存器, amd64 上则是 rip. 大部分都是用在跳转.

> 小结: 分析汇编, 重点是 SP, SB 寄存器 (FP寄存器在这里看不到的), 手写汇编, 重点看 SP, FP 寄存器.


伪寄存器的内存模型:

> 注意: 要站在 callee 的角度来看.

```
高地址
 /\ 
/__\ 
 ||   +-----------------+
 ||   | other caller... |
 ||   +-----------------+=======
 ||   | caller parent BP|      /\
 ||   +-----------------+      || 
 ||   | local var1      |      ||
 ||   +-----------------+      ||
 ||   |     ...         |      ||
 ||   +-----------------+      ||
 ||   | local varN      |      ||
 ||   +-----------------+      ||
 ||   | callee retN     |      ||
 ||   +-----------------+
 ||   |     ...         | caller frame stack
 ||   +-----------------+
 ||   | callee ret1     |      ||
 ||   +-----------------+      ||
 ||   | callee argN     |      ||
 ||   +-----------------+      ||
 ||   |     ...         |      ||
 ||   +-----------------+      ||
 ||   | callee arg1     |      ||
 ||   +-----------------+------||--------------> FP伪寄存器
 ||   | return address  |      \/
 ||   +-----------------+=======
 ||   | caller BP       |      /\
 ||   +-----------------+------||--------------> SP伪寄存器 (BP寄存器)
 ||   | local var1      |      
 ||   +-----------------+ callee frame stack
 ||   |     ...         |     
 ||   +-----------------+      ||
 ||   | local varN      |      \/
 ||   +-----------------+=======---------------> SP硬件寄存器
 ||
低地址
```

汇编代码:

```
TEXT ·add(SB), NOSPLIT, $8-24
    MOVQ $0, r1+16(FP)

    MOVQ a+0(FP), AX
    ADDQ b+8(FP), AX
    MOVQ AX, r1+16(FP)
    RET
```

对应的伪汇编代码:

```
"".add STEXT nosplit size=48 args=0x18 locals=0x10
	0x0000 00000 (pkg/add_amd64.s:66)	TEXT	"".add(SB), NOSPLIT, $16-24
	0x0000 00000 (pkg/add_amd64.s:66)	SUBQ	$16, SP
	0x0004 00004 (pkg/add_amd64.s:66)	MOVQ	BP, 8(SP)
	0x0009 00009 (pkg/add_amd64.s:66)	LEAQ	8(SP), BP
	0x000e 00014 (pkg/add_amd64.s:66)	FUNCDATA	$0, "".add.args_stackmap(SB)
	0x000e 00014 (pkg/add_amd64.s:67)	MOVQ	$0, r1+40(FP)
	0x0017 00023 (pkg/add_amd64.s:69)	MOVQ	a+24(FP), AX
	0x001c 00028 (pkg/add_amd64.s:70)	ADDQ	b+32(FP), AX
	0x0021 00033 (pkg/add_amd64.s:71)	MOVQ	AX, r1+40(FP)
	0x0026 00038 (pkg/add_amd64.s:72)	MOVQ	8(SP), BP
	0x002b 00043 (pkg/add_amd64.s:72)	ADDQ	$16, SP
	0x002f 00047 (pkg/add_amd64.s:72)	RET
```

在这段代码当中可以得到的信息: (代码当中的 SP 都是硬件SP寄存器, 假设当前被调用的是 add 函数)

0-8 为本地变量( add 函数当中的声明的本地变量长度是 8)
8-16 为 caller 的 BP. (SP伪寄存器)
16-24 return address. 伪汇编代码当中没有任何说明.
24-40 为 callee args. 两个参数a, b
40-48 为 callee rets. 一个返回参数.

> 在汇编代码当中:
> SP寄存器和伪SP寄存器之间的长度 = "函数局部变量长度"
> 伪SP寄存器和FP寄存器之间的长度 = "Caller BP" + "Return Addr" = 16
>
> 基于以上的结论, 在编写Go汇编函数细节点(案例):
>
> 获取函数参数和返回值(FP寄存器): 
>   MOVQ arg1+0(FP),  AX
>   MOVQ arg2+8(FP),  BX
>   MOVQ ret1+16(FP), CX
>   MOVQ ret2+24(FP), SI
>
> 创建局部变量(伪SP寄存器):
>   MOVQ $0, var0-0(SP), AX
>   MOVQ $0, var1-8(SP), BX
>
>
> 函数调用的参数和返回值(SP寄存器):
>   MOVQ AX, 0(SP)
>   MOVQ BX, 8(SP)
> 
> 如果觉得不麻烦, 也是可以使用上述关系进行搞定参数.


![image](/images/develop_assembly_regmem.png)
 

### 汇编指令

常用的汇编指令(指令后缀带 `Q` 说明是 64 位上的汇编指令)

运算和传送指令:

| 指令 | 指令种类 | 用途 | 案例 |
| ---- | ----- | ---- | ---- |
| MOVQ | 传送 | 数据传送 |  |
| LEAQ | 传送 | 地址传送 | LEAQ AX, BX // 把AX的有效地址传送到BX |
| PUSHQ | 传送 | 栈压入 |  |
| POPQ | 传送 | 栈弹出 | |
| ADDQ | 运算 | 相加并赋值 | ADDQ BX, AX // 等价 AX+=BX |
| SUBQ | 运算 | 相减并赋值 | SUBQ BX, AX // 等价 AX-=BX |


函数调用指令:

| 指令 | 指令种类 | 用途 | 案例 |
| ---- | ----- | ---- | ---- |
| CALL | 转移 | 调用函数, 返回地址(call后面那条指令的地址)入栈, 跳转到调用函数开始处 | CALL ·Sum(SB) |
| LEAVE | 转移 | 函数返回, 为返回准备好栈, 为 ret 准备好栈, 将栈顶的返回地址传递为 PC 寄存器 | LEAVE |


转移指令：

| 指令 | 指令种类 | 用途 | 案例 |
| ---- | ----- | ---- | ---- |
| JMP | 转移 | 无条件转移指令 | JMP LOOP // 无条件转至LOOP标签处, goto语句 |
| JE | 转移 | 等于跳转 | JE LOOP // 等于则跳转, 跳转到 LOOP 地址处 |
| JZ | 转移 | 为0则跳转 | JZ LOOP // 为0则跳转, 跳转到 LOOP 地址处, 一般前面都有一个 CMPQ 指令 |
| JS | 转移 | 负数跳转 | JS LOOP // 负数则跳转, 跳转到 LOOP 地址处 |
| JL | 转移 | 有符号小于则跳转 | JL LOOP // 有符号小于则跳转, 跳转到 LOOP 地址处 |
| JG | 转移 | 有符号大于则跳转 | JG LOOP // 有符号大于则跳转, 跳转到 LOOP 地址处 |
| JB | 转移 | 无符号小于则跳转 | JB LOOP // 无符号小于则跳转, 跳转到 LOOP 地址处 |
| JA | 转移 | 无符号大于则跳转 | JA LOOP // 无符号大于则跳转, 跳转到 LOOP 地址处 |


测试指令:

| 指令 | 指令种类 | 用途 | 案例 |
| ---- | ----- | ---- | ---- |
| TESTB | 测试 | 测试字节, 与关系 | TESTB AX AX // AX & AX |
| COMPB | 测试 | 比较字节, 差关系 | COMPB AX BX // BX - AX | 


---

在 Go 当中, 需要注意的一些细节:

1. **当 xxx 是`函数`, `变量` 时, `MOVQ $xxx(SB) AX` 和 `LEVQ xxx(SB) AX`是等价的, 都是用于计算 xxx 的地址**
这里需要注意的时, 如果 xxx 是 go 代码当中的函数或变量, 一般格式是 `package·var`. 但是对于对于汇编当中定义的函数或变量
需要根据实际情况指定.

2. 寄存器引用(不包括 SP, BP, SB, PC 伪寄存器).

```cgo
// 加括号代表是指针的引用
MOVQ (AX), BX // BX = *AX, 将 AX 指向的内存区域 8B 赋值给 BX
MOVQ 16(AX), BX // BX = *(AX+16)

// 不加括号是值的引用
MOVQ AX, BX // BX = AX, 将 AX 中存储的内存拷贝给 BX
```

3. 地址运算

```cgo
LEAQ (AX)(AX*2), CX // CX = AX + AX*2 = AX * 3
```

上面的代码当中的 2 代表 scale, scale 只能是0, 2, 4, 8

### 汇编案例

> CMPQ, JMP, JL 使用案例

```armasm
CMPQ AX, BX // AX 和 BX 比较
JL LOOP_FALSE // AX < BX
JMP LOOP_TRUE // 无调整转移, AX >= BX

LOOP_TRUE:
    CALL xxx(SB)

LOOP_FALSE:
    CALL xxx(SB)
```

> CALL 使用案例 -- 求和公式

```
#include "textflag.h"

// func Sum(n int) int
TEXT ·Sum(SB), $8
    MOVQ n+0(FP), AX   // n
    MOVQ ret+8(FP), BX // ret

    CMPQ AX, $0
    JG STEP
    JMP RETURN

STEP:
   SUBQ $1, AX  // AX-=1

   MOVQ AX, 0(SP) // AX 入函数调用栈. 局部变量
   CALL ·Sum(SB) 
   MOVQ 8(SP), BX // BX=Sum(AX-1), 获取函数返回值.

   MOVQ n+0(FP), AX   // AX=n
   ADDQ AX, BX        // BX+=AX
   MOVQ BX, ret+8(FP) // return BX
   RET


RETURN:
    MOVQ $0, ret+8(FP) // return 0
    RET
```

> CALL 使用案例 -- 加法公式

```
#include "textflag.h"

// func Add(a, b int) int
TEXT ·Add(SB), $16
    MOVQ a+0(FP), AX  // a
    MOVQ b+8(FP), BX  // b

    MOVQ AX, 0(SP)  // a 入函数参数栈. 局部变量
    MOVQ BX, 8(SP)  // b 入函数参数栈. 局部变量
    CALL ·Print2(SB) // 调用 Print2(a,b) 
    MOVQ 16(SP), CX // 获取函数返回值. 

    MOVQ CX, 0(SP)  // c 入函数栈. 局部变量
    CALL ·Print1(SB) // 调用 Printl(a)

    MOVQ a+0(FP), AX // AX=a, 避免AX污染
    MOVQ b+8(FP), BX // BX=b, 避免BX污染

    ADDQ AX, BX    // BX+=AX
    MOVQ BX, ret+16(FP) // return BX

    RET
```

函数调用小结:

完整函数调用分为三个部分: 函数参数准备, 调用函数, 获取函数返回值.

- 函数参数准备. **需要将准备好的参数压入到SP寄存器当中, 除非调用函数不需要参数**. 例如: 

```
MOVQ $10, 0(SP)  // 将 10 压入函数调用参数栈
MOVQ $20, 8(SP)  // 将 20 压入函数调用参数栈
MOVQ $30, 16(SP) // 将 30 压入函数调用参数栈
```

> 上面的案例是将 10, 20, 30 这三个数压入到函数参数调用栈.

- 调用函数. 这部分最为简单, 直接就是一个 `CALL` 指令后面跟上需要调用的函数. 例如:

```
CALL ·Print3(SB) // 调用 Print3(a,b,c)
```

- 获取函数返回值. **函数的返回值和函数参数共用一个栈, 因此参数返回值的偏移量是在参数压入栈的基础上进行偏移的**. 例如:

```
MOVQ 24(SP), AX // Print3函数三个 int 参数, 那么返回值的偏移量是从 24 开始
```


变量声明小结:

- 整数变量

```
// var INT int
GLOBL ·INT(SB), $8
DATA ·INT+0(SB)/8, $0x10
```

- 数组变量

// 数组的方式
```
// var ARRAY [2]byte
GLOBL ·ARRAY(SB), $16
DATA ·ARRAY+0(SB)/1, $0x10
DATA ·ARRAY+1(SB)/1, $0x20
```


- 字符串

// 采用结构体的方式
```
// var STRING string
GLOBL ·STRING(SB), NOPTR, $16
DATA  ·STRING+0(SB)/8, $·private<>(SB)
DATA  ·STRING+8(SB)/8, $23

GLOBL ·private<>(SB), NOPTR, $16
DATA ·private<>+0(SB)/8, $"12345678"      // ...string data...
DATA ·private<>+8(SB)/8, $"12345678"      // ...string data...
DATA ·private<>+16(SB)/8,$"12345678"      // ...string data...
```

> 注意: 上述的 `private<>` 是一个私有变量. 当然也可以是一个全局变量(相对私有变量而言), 但是这种方式定义的全局变量是非
> 法的. 例如, 下面的例子就无法赋值成功:

```
// var STRING string
GLOBL ·STRING(SB), NOPTR, $16
DATA ·STRING+0(SB)/8, $"12345678"      // ...string data...
DATA ·STRING+8(SB)/8, $"12345678"      // ...string data...
```

> 上述定义会报 `unexpected fault address` 错误

- 切片

// 字节切片
```
GLOBL ·hello(SB), $24           // var hello []byte("hello world!")
DATA ·hello+0(SB)/8,$text<>(SB) // SliceHeader.Data
DATA ·hello+8(SB)/8,$12         // SliceHeader.Len
DATA ·hello+16(SB)/8,$16        // SliceHeader.Cap

GLOBL text<>(SB), NOPTR, $16
DATA text<>+0(SB)/8,$"hello wo"      // ...string data...
DATA text<>+8(SB)/8,$"rld!"          // ...string data...
```

// 整数切片
```
// var SLICE []int
GLOBL ·SLICE(SB), $24
DATA ·SLICE+0(SB)/8, $slice<>(SB)
DATA ·SLICE+8(SB)/8, $4
DATA ·SLICE+16(SB)/8, $6

GLOBL slice<>(SB), NOPTR, $16
DATA slice<>+0(SB)/8, $10
DATA slice<>+8(SB)/8, $20
DATA slice<>+16(SB)/8, $21
DATA slice<>+24(SB)/8, $21
```

> 需要注意的一个细节点, 无论是字节切片(本质是一个字符串)还是整数切片, 在定义私有数据的时候, 都使用了 NOPTR 标记.这一点
需要被铭记. 

- map, chan 等内置的数据结构是在 runtime 当中实现的, 无法使用汇编创建(原因是汇编当中无法调用 runtime 的函数)

- 结构体, 上述的 数组, 字符串, 切片都是使用结构体的方式进行定义的, 其他的结构体也是类似的. 这这里不在举例.

> 最后, 需要铭记一点, 在汇编当中初始化的时候只是分配了一段内存, 然后为这段内存设置相应的数据, 仅此而已. 在汇编当中是没有
数据类型概念的.
