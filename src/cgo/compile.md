# C/C++ 静态链接库(.a) 和动态链接库(.so)

## 目标文件

目标文件常常按照特定格式来组织, 在 Linux 下, 它是 ELF 格式 (Executable Linkable Format, 可执行可
链接格式).

通常目标文件有三种形式:

- 可执行目标文件. 即可直接运行的二进制文件.

- 可重定位目标文件. 它包含了二进制的代码和数据, 可以与其他可重定位目标文件合并, 并创建一个可执行目标文件.

- 共享目标文件. 它是一种在加载或者运行时进行链接的特殊可重定位目标文件.

```cgo
// main.c
#include <stdio.h>
#include <math.h>

int main() {
    printf("hello \n");
    int b = 2;
    double a = exp(b);
    printf("%lf\n", a);
    return 0;
}
```

可重定位目标文件 `main.o`

```
$ gcc -c main.c
$ file main.o 
main.o: ELF 64-bit LSB relocatable, x86-64, version 1 (SYSV), not stripped

$ readelf -h main.o
ELF 头：
  Magic：  7f 45 4c 46 02 01 01 00 00 00 00 00 00 00 00 00 
  类别:                              ELF64
  数据:                              2 补码，小端序 (little endian)
  版本:                              1 (current)
  OS/ABI:                            UNIX - System V
  ABI 版本:                          0
  类型:                              REL (可重定位文件)
  ...
```

共享目标文件libm.so

```
$ file /lib/x86_64-linux-gnu/libm-2.27.so 
/lib/x86_64-linux-gnu/libm-2.27.so: ELF 64-bit LSB shared object x86-64, version 1 
(GNU/Linux), dynamically linked ...


$ readelf -h /lib/x86_64-linux-gnu/libm-2.27.so 
ELF 头：
  Magic：  7f 45 4c 46 02 01 01 03 00 00 00 00 00 00 00 00 
  类别:                              ELF64
  数据:                              2 补码，小端序 (little endian)
  版本:                              1 (current)
  OS/ABI:                            UNIX - GNU
  ABI 版本:                          0
  类型:                              DYN (共享目标文件)
  ...
```

可执行目标文件 `main`

```
$ gcc -o main main.o -l m

$ file main
main: ELF 64-bit LSB pie executable x86-64, version 1 (SYSV), dynamically linked, 
interpreter /lib64/ld-linux-x86-64.so.2, ...

$ readelf -h main
ELF 头：
  Magic：  7f 45 4c 46 02 01 01 00 00 00 00 00 00 00 00 00 
  类别:                              ELF64
  数据:                              2 补码，小端序 (little endian)
  版本:                              1 (current)
  OS/ABI:                            UNIX - System V
  ABI 版本:                          0
  类型:                              DYN (共享目标文件)
  ...
```


C语言当中库有两种, 一种是 `静态链接库`, 一种是 `动态链接库`. 不管是哪一种库, 要使用它们, 都要在程序中包
含相应的 `include` 头文件.

程序编译的过程:

[!image](../../images/compile.png)


结合 gcc 指令看一下每个阶段生成的文件:

```
gcc -c main.c
```

生成一个 `main.o` 文件, 该文件是将源文件编译成汇编文件, 在链接之前, 该文件不是可执行文件.

```
gcc -o main main.c
```

生成一个 `main` 可执行文件, 格式是 `ELF`. 该文件是链接后的可执行文件.


## 静态链接库

静态链接, 即在链接阶段, 将源文件中用到的库函数 与 汇编生成的目标文件 `.o` 合并生成可执行文件. 该可执行
文件可能比较大. 

这种链接方式的好处是: 方便移植, 因为可执行程序与库函数再无关系. 缺点是: 文件太大.

静态编译:

```
gcc -static -o main main.c
```

> 在 linux 当中, 静态库为 lib*.a, 动态库为 lib*.so

```cgo
// add.h
#ifdef  _ADD_H
#define _ADD_H
int add(int,int);
#endif

// add.c
#include "add.h"
int add(int a, int b) {
    return a+b;
}
```

生成静态库:

```
// 生成 .o 文件
gcc -c add.c

// 生成 .a 文件
ar -crv libadd.a add.o
```

编译链接静态文件:

```
gcc -o main main.c -L ./add -l add
```

> -L 是指定加载库文件的路径
> -l 是指定加载的库文件



## 动态链接库

静态加载:

[!image](../../images/aload.png)


对于静态编译的程序1和程序2, 都使用了库 staticMath. 在内存中就有两份相同的 staticMath 目标文件, 这样
很浪费空间, 一旦程序数量很多就很可能导致内存不足.


动态加载:

[!image](../../images/soload.png)

在动态库当中, 两个程序使用相同的库, 这个目标文件在内存中只有一份, 供所有程序使用.


生成一个动态库:

```
gcc -fPIC -shared -o libmath.so math.c
```

动态链接:

```
gcc -o main main.c -L ./ -l math
```

动态链接和静态链接是一样的. 注意: -l 后面的是 lib 与 so 中间的库名称.


## gcc(g++) 编译选项

- `-shared`, 指定生成动态链接库

- `-static`, 指定生成静态链接库

- `-fPIC`, 表示编译为位置独立的代码, 用于编译共享库(动态库). 目标文件需要创建成位置无关码, 概念上就是
可执行程序装载它们的时候, 它们可以放置可执行程序的内存里的任何地方.

- `-L`, 表示要链接的库所在目录

- `-l`, 指定链接时需要的动态库.

- `-Wall`, 生成所有警告信息.

- `-ggdb`, 此选项尽可能的生成gdb的可以使用的调试信息.

- `-g`, 编译器在编译的时候产生调试信息

- `-c`, 只激活预处理, 编译和汇编, 也就是把程序生成目标文件(.o文件)

- `-Wl,options`, 把参数(options) 传递给链接库ld, 如果options中间有逗号, 就将options分成多个选项,
然后传递给链接程序.

