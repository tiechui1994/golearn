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

- `top` 查看进程具体线程信息

使用 'top -H -p PID' 查询进程的具体线程信息.

```
user@master:~$ top -H -p 17938

top - 14:59:48 up 14 days, 23:52,  1 user,  load average: 1.02, 0.94, 0.91
Threads:   8 total,   1 running,   7 sleeping,   0 stopped,   0 zombie
%Cpu(s): 13.1 us,  0.2 sy,  0.0 ni, 86.4 id,  0.0 wa,  0.0 hi,  0.2 si,  0.0 st
KiB Mem : 16285756 total,  3835032 free,  4565472 used,  7885252 buff/cache
KiB Swap:  8000508 total,  8000252 free,      256 used. 10544932 avail Mem 

  PID USER      PR  NI    VIRT    RES    SHR S %CPU %MEM     TIME+ COMMAND                                                                                             
17942 user      20   0  703248   1712   1196 R 99.9  0.0   0:20.56 main                                                                                                
17938 user      20   0  703248   1712   1196 S  0.0  0.0   0:00.00 main                                                                                                
17939 user      20   0  703248   1712   1196 S  0.0  0.0   0:00.01 main                                                                                                
17940 user      20   0  703248   1712   1196 S  0.0  0.0   0:00.00 main                                                                                                
17941 user      20   0  703248   1712   1196 S  0.0  0.0   0:00.00 main                                                                                                
17943 user      20   0  703248   1712   1196 S  0.0  0.0   0:00.00 main                                                                                                
17944 user      20   0  703248   1712   1196 S  0.0  0.0   0:00.00 main                                                                                                
17945 user      20   0  703248   1712   1196 S  0.0  0.0   0:00.00 main
```

> 这里可以看到 PID 为 17942 的线程, 其CPU相对比较高.

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

- `break, b <linespec>` 添加断点, 其中 <linespec> 的格式可以是 `*<address>`, `<filename>:<line>`, 
`<function>:[<line>]`
- `breakpoints, bp` 打印所有的断点.


- `step, s` 单步调试
- `step-instruction, si` 单步调试CPU指令
- `stepout, so` 跳出当前函数
- `next, n` 跳跃到下一个断点
- `continue, c` 运作直到遇到下一个断点或程序终止

- `restart, r`, 重启进程


- `display -a [%format] <expression>`, '-a' 选项将一个表达式添加到每次调试执行完成时打印的表达式列表中.
- `display -d <number>`, '-d' 选项从列表中删除指定的表达式.
- `args`, 打印函数参数
- `locals [<regex>]`, 打印局部变量
- `print, p [%format] <expression>`, 解析表达式
- `regs`, 打印寄存器的值
- `set <variable> = <value>`, 设置变量的值


- `goroutine, gr <id>`, 切换 goroutine
- `goroutines, grs`
- `thread, tr`
- `threads`, 所有的线程


- `disass`, 汇编当前的代码.
- `list, ls [<linespec>]`, 查看源代码. `<linespec>` 可以是 `<filename>:<line>`, `<function>:[<line>]`,
`<line>`


- `stack, bt`, 查看当前的栈信息



