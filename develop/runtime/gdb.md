## gdb 调试常用命令

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

- info functions: 查询函数
- info frame: 当前调用栈的状况
- info registers `[rsp, rip, rax, ...]`: 当前`所有`寄存器的状况
- info stack: stack backtrace
- info args: 当前栈上的所有参数
- info locals: 显示当前堆栈页的所有变量.
- info threads: 显示当前所有的线程信息.
- info files: 正在调试的 targets 和 file 的名称
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