# Go 汇编基础

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
个局部变量的结束位置(为啥? 看寄存器内存布局图). 使用形如 `symbol+offset(SP)` 的形式, 引用函数的局部变量. 加入局部变
量都是8个字节, 那么第一个局部变量就可以使用 localvar0-8(SP) 来表示. 与 `硬件SP寄存器` 是两个不同的东西, 在栈帧为 0
的情况下, `伪SP寄存器` 和 `硬件SP寄存器` 指向同一位置. 手写汇编时, 如果是 `symbol+offset(SP)` 则表示 `伪寄存器SP`,
如果是 `offset(SP)` 则表示 `硬件SP寄存器`.

> 注意: 对于编译输出 (`go tool compile -S / go tool objdump`) 的代码来讲, 所有的 SP 都是 `硬件SP寄存器`. 无论
是否带 symbol

- PC: 实质上就是体系结构当中的 pc 寄存器. 在 x86 平台对应 ip 寄存器, amd64 上则是 rip. 大部分都是用在跳转.

> 小结: 分析汇编, 重点是 SP, SB 寄存器 (FP寄存器在这里看不到的), 手写汇编, 重点看 SP, FP 寄存器.


伪寄存器的内存模型:

> 注意: 要站在 callee 的角度来看.

![image](/images/develop_assembly_regmem.png)
 
 
汇编指令:

常用的汇编指令(指令后缀带 `Q` 说明是 64 位上的汇编指令)

| 指令 | 指令种类 | 用途 | 案例 |
| ---- | ----- | ---- | ---- |
| MOVQ | 传送 | 数据传送 |  |
| LEAQ | 传送 | 地址传送 | LEAQ AX, BX // 把AX的有效地址传送到BX |
| PUSHQ | 传送 | 栈压入 |  |
| POPQ | 传送 | 栈弹出 | |
| ADDQ | 运算 | 相加并赋值 | ADDQ BX, AX // 等价 AX+=BX |
| SUBQ | 运算 | 相减并赋值 | SUBQ BX, AX // 等价 AX-=BX |
| CMPQ | 运算 | 比较大小 | CMPQ SI CX // 比较 SI 和 CX 大小 |
| CALL | 转移 | 调用函数 | CALL ·Sum(SB) |
| JMP | 转移 | 无条件转移指令 | JMP LOOP // 无条件转至LOOP标签处, goto语句 |
| JE | 转移 | 等于则跳转 | JE LOOP // 等于则跳转, 跳转到 LOOP 地址处 |
| JZ | 转移 | 为0则跳转 | JZ LOOP // 为0则跳转, 则跳转到 LOOP 地址处, 一般前面都有一个 CMPQ 指令 |
| JL | 转移 | 有符号小于则跳转 | JL LOOP // 有符号小于则跳转, 跳转到 LOOP 地址处 |
| JB | 转移 | 无符号小于则跳转 | JB LOOP // 无符号小于则跳转, 跳转到 LOOP 地址处 |
| JG | 转移 | 有符号大于则跳转 | JG LOOP // 有符号大于则跳转, 跳转到 LOOP 地址处 |
| JA | 转移 | 无符号大于则跳转 | JA LOOP // 无符号大于则跳转, 跳转到 LOOP 地址处 |


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

   MOVQ AX, 0(SP) // AX 入函数调用栈
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
TEXT ·Add(SB), $8
    MOVQ a+0(FP), AX  // a
    MOVQ b+8(FP), BX  // b

    MOVQ AX, 0(SP)  // a 入函数参数栈
    MOVQ BX, 8(SP)  // b 入函数参数栈
    CALL ·Print2(SB) // 调用 Print2(a,b) 
    MOVQ 16(SP), CX // 获取函数返回值. 

    MOVQ CX, 0(SP)  // c 入函数栈
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
