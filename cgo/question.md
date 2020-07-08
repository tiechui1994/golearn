# CGO 常见的问题

1. 明明已经指定动态库目录, 编译通过后, 运行错误: `找不到so目录`

```
error while loading shared libraries: libheader.so: cannot open shared object file: No such file 
or directory
```

原因分析:

> 问题就是出现在 `LDFLAGS: -L ./ -l xxx` 这条指令上. 这条指令的意思是指定动态库的目录. 
> 编译时指定 -L 目录, 只是在程序链接成可执行文件时使用, 但是运行程序时, 并不会去指定的目录寻找动
> 态库, 所以就找不到该动态库了. `自行百度链接器和动态链接器`.
>
> ps: 如果编译的是静态库, 就不会出现该问题了, 因为编译的时候已经加载到程序里了.


解决方案:

> 1.`//#cgo LDFLAGS: -L ./ -l xxx -Wl,-rpath -Wl,/path/to/lib` 编译时指定动态库路径
> 
> 2.设置系统动态库的路径. 在 `/etc/ld.so.conf` 目录下添加动态库的路径(运行的机器)
>
> 3.用 C 手动打开动态库. 将编译的动态库添加到系统动态库 `/usr/local/lib` 下.


2. C 代码抽出来后, 运行时提示找不到函数

```
/tmp/go-build649472735/b001/_x002.o: In function `_cgo_4167bb085d83_Cfunc_SayHello':
/tmp/go-build/cgo-gcc-prolog:49: undefined reference to `SayHello'
collect2: error: ld returned 1 exit status
```

原因分析:

> 在 main package 中, go run xxx 会出现这个问题.

解决方案:

> 进入到目标执行 go build xxx, 然后执行


3. 头文件声明变量, 提示重复定义错误

```
duplicate symbol _a in:
$WORK/b002/_x002.o
$WORK/b002/_x003.o
ld: 1 duplicate symbol for architecture x86_64
clang: error: linker command failed with exit code 1 (use -v to see invocation)
```

重现:

c代码:
```cgo
// header.h
#ifndef _BAR_H
#define _BAR_H
int a = 100;
void SayHello();
#endif


// header.c
#include <stdio.h>
#include "header.h"
void SayHello() {
    printf("hello world! \n");
}
```

go 代码:
```cgo
package bar

// #cgo CFLAGS: -I ./
// #cgo LDFLAGS: -L ./ -l header
// #include "header.h"
import "C"

func Hello() {
    C.SayHello()
}


package main

import "bar"

func main(){
    bar.Hello()
}
```

> 注意: 使用的是静态链接库

> 编译执行过程如下:

```
# 静态链接库
gcc -c -o header.o header.c 
ar rcs libheader.a header.o 

# 执行
go run main.go bar.go
```

原因分析:

> 重复引用头文件导致的. 这和 cgo 的编译相关. 运行 `go run -x main.go` 可以看到 (linux下的错误) 
> /tmp/go-build/bar/_obj/bar.o: (.data+0x0): multiple define of 'a' /tmp/go-build/bar/_obj/bar.cgo
> 2.o:(.data+0x0): first defined here. 每个文件是先单独编译成 .o 文件, 然后合并头的时候文件就被引用了两次, 所
> 以重复定义了.


解决方案:

> 去掉 header.c 的 `#include "header.h"` 就行了, cgo 会自动查找同名 c 文件. 如果要使用 header.h 里面定义的变
> 量 a, 使用 `extern int a;` 即可.
>
> 或者使用动态链接库, 编译的时候添加 `-fPIC` 参数.
 
