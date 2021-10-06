## TLS (Thread Local Storage)

线程本地存储又叫线程局部存储, 简称 TLS, 其实质是线程私有的全局变量.

多线程编程当中, 普通的全局变量在多线程中是共享的, 一个线程队其进行了修改, 所有线程都可以看到这个修改, 但是线程私有的全局
变量与普通变量不同, 线程私有全局变量是线程私有的数据, 每个线程都有自己的一份副本, 某个线程对其所做的修改只会修改它自己的
副本, 并不会修改到其他线程的副本. 如果把线程看做一个类, 那么线程私有全局变量就是类的私有变量, 每个类的实例都有这个私有变
量, 每次实例修改只能修改自己的, 但不影响其他的实例.

在 C 当中, 线程私有的全局变量使用 `__thread` 关键字进行修饰.

```cgo
#include <stdio.h>
#include <unistd.h>
#include <pthread.h>

__thread int g = 0;

void* start(void* arg) {
    printf("start, g[%p]: %d\n", &g, g);

    g++;

    return NULL;
}

int main(int argc, char* argv[]) {
    pthread_t tid;
    g = 100;
    pthread_create(&tid, NULL, start, NULL);
    pthread_join(tid, NULL);

    printf("main, g[%p]: %d\n", &g, g);

    return 0;
}
```

gcc 编译: `gcc -o tls -lpthread main.c`

使用 `gdb ./main` 打开 gdb 调试. 使用 `disass main` 查看 main 函数的汇编代码, 如下:

```
(gdb) disass main
Dump of assembler code for function main:
   0x0000000000401194 <+0>:     push   %rbp
   0x0000000000401195 <+1>:     mov    %rsp,%rbp
   0x0000000000401198 <+4>:     sub    $0x20,%rsp
   0x000000000040119c <+8>:     mov    %edi,-0x14(%rbp)
   0x000000000040119f <+11>:    mov    %rsi,-0x20(%rbp)
   0x00000000004011a3 <+15>:    movl   $0x64,%fs:0xfffffffffffffffc
   0x00000000004011af <+27>:    lea    -0x8(%rbp),%rax
   0x00000000004011b3 <+31>:    mov    $0x0,%ecx
   0x00000000004011b8 <+36>:    mov    $0x401142,%edx
   0x00000000004011bd <+41>:    mov    $0x0,%esi
   0x00000000004011c2 <+46>:    mov    %rax,%rdi
   0x00000000004011c5 <+49>:    callq  0x401030 <pthread_create@plt>
   0x00000000004011ca <+54>:    mov    -0x8(%rbp),%rax
   0x00000000004011ce <+58>:    mov    $0x0,%esi
   0x00000000004011d3 <+63>:    mov    %rax,%rdi
   0x00000000004011d6 <+66>:    callq  0x401050 <pthread_join@plt>
   0x00000000004011db <+71>:    mov    %fs:0xfffffffffffffffc,%eax
   0x00000000004011e3 <+79>:    mov    %fs:0x0,%rdx
   0x00000000004011ec <+88>:    lea    -0x4(%rdx),%rcx
   0x00000000004011f3 <+95>:    mov    %eax,%edx
   0x00000000004011f5 <+97>:    mov    %rcx,%rsi
   0x00000000004011f8 <+100>:   mov    $0x402016,%edi
   0x00000000004011fd <+105>:   mov    $0x0,%eax
   0x0000000000401202 <+110>:   callq  0x401040 <printf@plt>
   0x0000000000401207 <+115>:   mov    $0x0,%eax
   0x000000000040120c <+120>:   leaveq 
   0x000000000040120d <+121>:   retq 
```

汇编指令:

```
0x00000000004011a3 <+15>: movl $0x64,%fs:0xfffffffffffffffc
```

将 0x64(100) 存储到 `%fs:0xfffffffffffffffc` 这个地址上. 可以看到全局变量 g 的地址为 `%fs:0xfffffffffffffffc`,
fs 是段寄存器, 0xfffffffffffffffc 是有符号 -4, 因此全局变量 g 的地址为: `fs段基址 - 4`.

通过系统调用 `arch_prctl` 获取 fs 段基地址. `0x1003` 表示 ARCH_GET_FS, `$rsp-8` 表示获取的数据存储的地址. `x`
命令用于显示内存地址.

> arch_prctl 设置/获取特定架构的线程状态.
> 参考: https://man7.org/linux/man-pages/man2/arch_prctl.2.html

```
(gdb) call (int)arch_prctl(0x1003, $rsp-8)
$6 = 0

(gdb) x/a $rsp-8
0x7fffffffcb98: 0x7ffff7dce740

(gdb) x/ha &g
0x7ffff7dce73c: 0x64
```

`0x7ffff7dce740` 为 fs 段基地址. `0x7ffff7dce73c` 为 &g 的地址. 可以看出两者相差值是 `4`.


接着, 跳转到 start 函数. 反汇编 `start` 函数.

```
(gdb) disass start
Dump of assembler code for function start:
   0x0000000000401142 <+0>:     push   %rbp
   0x0000000000401143 <+1>:     mov    %rsp,%rbp
   0x0000000000401146 <+4>:     sub    $0x10,%rsp
   0x000000000040114a <+8>:     mov    %rdi,-0x8(%rbp)
   0x000000000040114e <+12>:    mov    %fs:0xfffffffffffffffc,%eax
=> 0x0000000000401156 <+20>:    mov    %fs:0x0,%rdx
   0x000000000040115f <+29>:    lea    -0x4(%rdx),%rcx
   0x0000000000401166 <+36>:    mov    %eax,%edx
   0x0000000000401168 <+38>:    mov    %rcx,%rsi
   0x000000000040116b <+41>:    mov    $0x402004,%edi
   0x0000000000401170 <+46>:    mov    $0x0,%eax
   0x0000000000401175 <+51>:    callq  0x401040 <printf@plt>
   0x000000000040117a <+56>:    mov    %fs:0xfffffffffffffffc,%eax
   0x0000000000401182 <+64>:    add    $0x1,%eax
   0x0000000000401185 <+67>:    mov    %eax,%fs:0xfffffffffffffffc
   0x000000000040118d <+75>:    mov    $0x0,%eax
   0x0000000000401192 <+80>:    leaveq 
   0x0000000000401193 <+81>:    retq
```


```
(gdb) call (int)arch_prctl(0x1003, $rsp-8)
$8 = 0

(gdb) x/a $rsp-8
0x7ffff7dcced8: 0x7ffff7dcd700

(gdb) x/ha &g
0x7ffff7dcd6fc: 0x0
```

`0x7ffff7dcd700` 和 `0x7ffff7dcd6fc` 也是相差 4. 

注: `main` 线程和 `start` 线程当中的 fs 基地址是完全不一样的, 这也就导致了同样的全局变量 g, 在两个线程当中的地址是完
全不一样的. 那么修改任何一个线程当中的 g 的值, 对另外一个线程都是没有影响的.

上述得出结论: **gcc编译器(线程库以及内核的支持)使用了CPU的fs段寄存器来实现线程本地存储,* 不同的线程中 fs 段基地址是不
一样的, 这样看似一个全局变量但在不同线程中却拥有不同的内存地址, 从而实现了线程私有的全局变量.

> 在 gdb 当中, 命令 `info threads` 可以查看当前所有的线程信息. 在其中的 `Target Id` 当中就包含有有线程的 fs 基地址.
fs+offset, offset为正, 表示系统定义的线程私有变量. offset为负数, 表示用户自定义的线程私有变量.

```
0x04  线程堆栈顶部
0x08  线程堆栈底部

0x20  进程PID
0x24  线程ID
```
