# gdb 调试

> GDB, GNU 项目调试器, 允许您查看另一个程序在执行时 "内部" 发生了什么 --- 或者另一个程序在崩溃时正在做什么.
>
> GDB 可以做四种主要的事情(加上其他支持这些的事情)来帮助你在行动中捕捉错误:
>
> 1.启动你的程序, 指定任何可能影响它的行为的东西.
>
> 2.让你的程序在指定的条件下停止.
>
> 3.当你的程序停止时，检查发生了什么.
>
> 4.改变你的程序，这样你就可以尝试纠正一个错误的影响, 然后继续学习另一个错误.
>
> 这些程序可能与 GDB 在同一台机器上(本地), 另一台机器(远程)或模拟器上执行.
>
> GDB 可以在大多数流行的 UNIX 和 Microsoft Windows 变体以及 Mac OS X 上运行.

## gdb 调试命令

执行命令:

- run, r: 运行程序, 当遇到断点后, 程序会在断点处停止运行, 等待用户输入下一步命令
- continue, c: 继续执行, 到下一个断点处(或运行结束)
- next, n: 单步调试, 如果有函数调用, 不进入此函数体.
- step, s: 单步调试, 如果有函数调用, 则进入函数. 
- si: 执行单条指令.
- until <line>, u <line>: 运行到函数某一行.
- finish: 运行程序, 知道当前函数完成后返回, 并打印函数返回时堆栈地址和返回值以及参数值等信息.
- call func(args), 调用程序中可见的函数, 并传递"参数", 如: call gdb_test(55)
- return <value>, 改变程序执行流程, 直接结束当前函数, 并将指定值返回.
- quit, q: 退出 gdb


断点设置:

- break <n>, b <n>: 在第n行处设置断点
- b <func>: 在函数 func() 的入口设置断点, 如: b main
- delete <断点号n>: 删除第n个断点
- disable/enable <断点号n>: 暂停/启用第n个断点
- info b: 显示断点设置状况


打印命令:

- info functions `[regex]`: 查询函数

- info frame: 当前调用栈的状况
```
(gdb) info frame
Stack level 0, frame at 0xc000066f80:
 rip = 0x47f780 in main.main (/home/user/workspace/videos/golang/main.go:18); saved rip = 0x432067
 source language unknown.
 Arglist at 0xc000066f70, args: 
 Locals at 0xc000066f70, Previous frame's sp is 0xc000066f80
 Saved registers:
  rip at 0xc000066f78
```

- info registers `[rsp, rip, rax, ...]`: 当前`所有`寄存器的状况
- info stack: stack backtrace
- info args: 当前栈上的所有参数
- info locals: 显示当前堆栈页的所有变量.
- info threads: 显示当前所有的线程信息.

- info files: **正在调试程序 file 名称,  Entry point, 各个 section 地址分布**
```
(gdb) info files 
Symbols from "/home/user/workspace/videos/golang/video".
Local exec file:
    `/home/user/workspace/videos/golang/video', file type elf64-x86-64.
    Entry point: 0x45c220
    0x0000000000401000 - 0x000000000047f8b1 is .text
    0x0000000000480000 - 0x00000000004b52e5 is .rodata
    0x00000000004b5480 - 0x00000000004b595c is .typelink
    0x00000000004b5960 - 0x00000000004b59b8 is .itablink
    0x00000000004b59b8 - 0x00000000004b59b8 is .gosymtab
    0x00000000004b59c0 - 0x000000000050ebc8 is .gopclntab
    0x000000000050f000 - 0x000000000050f020 is .go.buildinfo
    0x000000000050f020 - 0x000000000051f600 is .noptrdata
    0x000000000051f600 - 0x0000000000526e10 is .data
    0x0000000000526e20 - 0x0000000000555d28 is .bss
    ...
```

- info symbol ADDR: 打印存储在地址 ADDR 的符号名称. 如果没有符号恰好存储在 ADDR 中, GDB打印最近的符号及其偏移量.
```
(gdb) info symbol 0x47f8a0
main.main.func1 in section .text of /home/user/workspace/videos/golang/video
```

- info address SYMBOL: 符号 SYMBOL 存储位置. 对于 register 变量, 显示存储在哪个寄存器. 对于 non-register 本
地变量, 会输出 stack 偏移量.
```
(gdb) info address main.main
Symbol "main.main" is a function at address 0x47f780.
```

- info display: 打印程序暂停时display的表达式
- where: 显示所有帧栈的backtrace

- display FMT

- x/FMT ADDRESS, 检查内存. ADDRESS是要检查的内存地址的表达式. FMT是显示的格式. 格式字母包括:

```
o(octal)

x(hex)

d(decimal)

u(unsigned decimal)

t(binary)

f(float)

c(char)

s(string)

z(hex, 在左侧填充0)

a(address) // 地址

i(instruction) // 指令
```

```
b 表示单字节
h 表示双字节
w 表示四字节
g 表示八字节
```

- print <表达式>, p <表达式>: 其中的"表达式"是当前测试程序的有效表达式, 例如: p var(打印变量var的值), p fun(22) 
调用函数fun()

- display <表达式>: 在每次单步进行指令后, 紧接着输出被设置的表达式及值. 如 dispaly a


堆栈相关:

- info registers, i r `[rsp,rip,rbp]`: 查看寄存器使用情况
- info stack, 查看栈使用状况
- up/down, 调到上一层/下一层函数


获取 fs, gs 寄存器的基地址:

```
(gdb) call (int)arch_prctl(0x1003, $rsp-0x8)    
$2 = 0

(gdb) x/gx $rsp-0x8
0x7fffffffe6e8: 0x00007ffff7fe0700 

(gdb) call (int)arch_prctl(0x1004, $rsp-0x8)
$3 = 0 
(gdb) x/gx $rsp-0x8
0x7fffffffe6e8: 0x0000000000000000
```

注: `arch_prctl` 是系统调用函数, 用于获取具体进程或线程状态. `0x1003` 表示 ARCH_GET_FS, `0x1004` 表示 ARCH_GET_GS

## objdump

objdump 显示有关一个或多个可执行文件的信息.

```
objdump <option(s)> <file(s)>
```

option 选项:

- `-f` 或 `--file-headers`: 显示 file 的 header 信息(start address, arch, file format)

- `-h`, `--section-headers` 或 `--headers`: 显示 file 的 section header 摘要信息. 常用的 section 有 text,
rodata, data, bss 等

- `-p` 或 `--private-headers`: 显示 file 特定的 header 信息(依赖的动态库版本, dynamic section)

- `-r` 或 `--reloc`: 重定位

- `-t` 或 `--syms`: file 符号表. 类似 `nm`. 

基于 ELF 的输出格式:

```
SYMBOL TABLE:
0000000000000000 l    df *ABS*	0000000000000000 go.go
0000000000401000 l     F .text	0000000000000000 runtime.text
0000000000401d40 l     F .text	000000000000022d cmpbody
```

第一个字段是符号地址, 第二个字段是标志位, 第三个字段是 section(或 `*ABS*`,`*UND*`), 第四个字段, 对于普通符号是对
齐字节, 对于其他符号是当前符号大小; 最后一个字段是符号名称.

关于标志位:

```
1, g, u, !: 分别表示 symbol 是一个 local(l), global(g), unique global(u), "'not global and not local' 或
'golbal and local'" (!). unique global symbol 是标准 ELF symbol 集合的 GNU 扩展, 对于这样的 symbol, 动态
链接器需要确保整个过程中只有一个 symbol 使用此名称

C, 表示 symbol 是一个 constructor(C)

I, i: 分别表示 symbol 是一个 indirect reference[间接引用] to another symbol(I), function to be evaluated during reloc processing (i)

D, d: 分别表示 symbol 是一个 debuging symbol(d), dynamic symbol(D)

F, f, O: 分别表示 symbol 是一个 function(F), file(f), Object(O)
```

- `-x` 或 `--all-headers`: 显示 file 所有的 header 信息. 等价于参数 `-a -f -h -p -r -t`


- `-T` 或 `--dynamic-syms`: 显示 file 动态符号表. 这个只对动态库 file 有效. 类似 `nm --dynamic`. 输出格式与
`--syms` 类似, 除了在 symbol 的名称之前, 给出与之关联的版本信息 symbol.

- `-R` 或 `--dynamic-reloc`: 动态重定位. 这个只对动态库 file 有效. 

- `-S` 或 `--source`:  显示源代码的反汇编.

- `-d` 或 `--disassemble`: 源代码反汇编

- `-j name` 或 `--section name`: 显示具体 section 的内容, 常与 `-d`, `-r`, `-S` 一起使用

- `-l` 或 `--line-numbers`: 将源代码的行号与目标代码或重定位对应起来. 常与 `-d`, `-r` 一起使用. 

- `--start-address=address`, `--stop-address=address`, 从指定指定地址处显示数据. 常与 `-d`, `-S`, `-r` 一
起使用. 


### 案例

- 查看指定函数的汇编代码

1) 查询函数 `fmt.Println` 地址: 
```
> objdump video -t | grep fmt.Println -C 1
0000000000479020 g     F .text	00000000000000e5 fmt.Fprintln
0000000000479120 g     F .text	0000000000000067 fmt.Println
00000000004791a0 g     F .text	0000000000000087 fmt.getField
```

2) 查看汇编代码(指定开始地址和结束地址): 
```
> objdump video -d -l --start-address=0x479120 --stop-address=0x4791a0
video:     file format elf64-x86-64

Disassembly of section .text:

0000000000479120 <fmt.Println>:
fmt.Println():
/opt/local/go/src/fmt/print.go:273
  479120:	49 3b 66 10          	cmp    0x10(%r14),%rsp
  479124:	76 3c                	jbe    479162 <fmt.Println+0x42>
  479126:	48 83 ec 30          	sub    $0x30,%rsp
  47912a:	48 89 6c 24 28       	mov    %rbp,0x28(%rsp)
  47912f:	48 8d 6c 24 28       	lea    0x28(%rsp),%rbp
  479134:	48 89 44 24 38       	mov    %rax,0x38(%rsp)
/opt/local/go/src/fmt/print.go:274
  479139:	48 8b 15 30 dd 0a 00 	mov    0xadd30(%rip),%rdx        # 526e70 <os.Stdout>
  479140:	48 89 df             	mov    %rbx,%rdi
  479143:	48 89 ce             	mov    %rcx,%rsi
  479146:	48 89 d3             	mov    %rdx,%rbx
  479149:	48 89 c1             	mov    %rax,%rcx

...
```

## nm

## strip

