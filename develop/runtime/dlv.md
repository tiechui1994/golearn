## dlv 调试

### 安装

1. dlv 版本下载

```
# amd64
curl -o dlv -L https://hub.fastgit.org/tiechui1994/jobs/releases/download/dlv_v1.7.3/dlv-1.16-amd64

# arm64
curl -o dlv -L https://hub.fastgit.org/tiechui1994/jobs/releases/download/dlv_v1.7.3/dlv-1.16-arm64
```

2. 相关内核变量设置

```
sudo bash -c "echo 0 > /proc/sys/kernel/yama/ptrace_scope"
```

### 调试

- 开启 dlv 调试

使用命令 'dlv attach PID' 开启对运行时进程的调试

```
user@master:~$ ./dlv attach 17938
Type 'help' for list of commands.
(dlv) goroutines
 Goroutine 1 - User: ./main.go:25 main.main (0x49bbb7) [chan receive]
 Goroutine 2 - User: /opt/local/local/go/src/runtime/proc.go:307 runtime.gopark (0x434e85) [force gc (idle)]
 Goroutine 3 - User: /opt/local/local/go/src/runtime/proc.go:307 runtime.gopark (0x434e85) [GC sweep wait]
 Goroutine 4 - User: /opt/local/local/go/src/runtime/proc.go:307 runtime.gopark (0x434e85) [GC scavenge wait]
 Goroutine 5 - User: /opt/local/local/go/src/runtime/proc.go:307 runtime.gopark (0x434e85) [finalizer wait]
 Goroutine 6 - User: ./main.go:16 main.main.func1 (0x49bc94) (thread 17942)
 Goroutine 7 - User: /opt/local/local/go/src/runtime/proc.go:307 runtime.gopark (0x434e85) [select]
 Goroutine 8 - User: /opt/local/local/go/src/runtime/sigqueue.go:147 os/signal.signal_recv (0x46223d) (thread 17941)
[8 goroutines]
```

> 这里看到线程 17942 绑定的 goroutine 是 6. 接下来就是切换到 'Goroutine 6' 上, 并查看栈信息. 从而定位问题.

```
(dlv) goroutine 6
Switched from 0 to 6 (thread 17942)
(dlv) stack
0  0x000000000040602f in runtime.selectnbrecv
   at /opt/local/local/go/src/runtime/chan.go:687
1  0x000000000049bc94 in main.main.func1
   at ./main.go:16
2  0x00000000004651c1 in runtime.goexit
   at /opt/local/local/go/src/runtime/asm_amd64.s:1374

```


### dlv 常用的指令

#### 断点类

- `break, b <locspec>` 添加断点, 其中 <locspec> 的格式可以: 

```
*<address>  指定内存地址的位置. 内存地址一般来源于汇编代码当中 

<filename>:<line>  指定文件名中的行line.

<line>  当前文件的行号

+<offset> 当前位置正向偏移
-<offset> 当前位置反向偏移


<function>:[<line>] '<function>'的格式是 '<package>.(*<reciver type>).<function name>'
```

- `breakpoints, bp` 打印所有的断点.


#### 调试执行

- `step, s`              单步调试, 函数文件, 汇编代码单步跳跃
- `step-instruction, si` 单步调试 CPU 指令(使用 disass 汇编的代码)
- `stepout, so`          跳出当前函数

- `next, n`     跳跃到下一个断点
- `continue, c` 运作直到遇到下一个断点或程序终止

- `restart, r`, 重启进程


#### 测试

- `regs`, 打印寄存器的值. 常常关注的寄存器有: **rip, rsp(当前函数栈顶), rbp(当前函数栈基)**, **rdi, rsi, rdx, rcx, r8, r9(C语言函数调用传参首选的寄存器, 如果不够, 则将剩余的参数入栈)**, 
**rax(C语言函数返回值使用的寄存器)**. 

- `examinemem`, `x`, 测试内存地址, 如果知道地址, 可以用于测试内存存储结构.

用法:

```
x [-fmt <formt>] [-count <count>] [-size <size>] <address>
x [-fmt <formt>] [-count <count>] [-size <size>] -x <expression>
```

例如:

```
x -count 20 -size 8 0xc00008af38
x -count 20 -size 8 -x 0xc00008af38 - 16
x -count 20 -size 8 -x &myVar
x -count 20 -size 8 -x myPtrVar
```

- `args`, 打印函数参数
- `locals [<regex>]`, 打印局部变量
- `display <expression>` 打印表达式的值(每执行一次). 有两个选项: '-a' 选项将一个表达式添加到每次调试执行完成时打印的表达式列表中;  '-d' 选项从列表中删除指定的表达式.
- `print, p [%format] <expression>`, 解析表达式.

关于 `<expression>`, 格式:

```
1. 变量



2. 接口

<interface name>(<concret type>) <value>
```

- `set <variable> = <value>`, 设置变量的值


#### 协程, 线程, 堆栈

- `goroutine, gr <id>`, 切换 goroutine
- `goroutines, grs`
- `thread, tr`
- `threads`, 所有的线程
- `frame <m>`, 栈帧 <m> 信息
- `stack, bt`, 查看当前的栈信息(函数调用链)

#### 代码

- `disass`, 汇编代码

用法:

```
[grountine <n>] [frame <m>] disass [-a <start> <end>] [-l <locspec>] '-a' 用于指定内存范围. '-l' 用于指定函数位置(更常用).
```


- `list, ls [<locspec>]`, 查看源代码. `<locspec>` 参考 'break' 当中的含义.

