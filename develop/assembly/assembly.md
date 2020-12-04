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
| CALL | 转移 | 调用函数 | |
| JMP | 转移 | 无条件转移指令 | JMP 0x0185 // 无条件转至 0x0185 地址处 |
| JLS | 转移 | 条件转移指令 | JLS 0x185 // 左边小于右边, 则跳转到 0x0185 地址处, 一般前面都有一个 CMPQ 指令 |

