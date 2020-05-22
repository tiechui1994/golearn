# C/C++ 静态链接库(.a) 和动态链接库(.so)

库有两种, 一种是 `静态链接库`, 一种是 `动态链接库`. 不管是哪一种库, 要使用它们, 都要在程序中包含相应的 `include` 头文件.

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

静态链接, 即在链接阶段, 将源文件中用到的库函数 与 汇编生成的目标文件 `.o` 合并生成可执行文件. 该可执行文件可能比较大. 

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


对于静态编译的程序1和程序2, 都使用了库 staticMath. 在内存中就有两份相同的 staticMath 目标文件, 这样很浪费空间, 一旦程序数量
很多就很可能导致内存不足.


动态加载:

[!image](../../images/soload.png)

