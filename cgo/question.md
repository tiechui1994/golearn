## CGO 常见的问题

### 明明已经指定动态库目录, 编译通过后, 运行错误: `找不到so目录`

```
error while loading shared libraries: libheader.so: cannot open shared object file: No such file 
or directory
```

原因分析:

> 问题就是出现在 `LDFLAGS: -L ./ -l xxx` 这条指令上. 这条指令的意思是指定动态库的目录. 编译时指定 -L 目录, 只是在
> 程序链接成可执行文件时使用, 但是运行程序时, 并不会去指定的目录寻找动态库, 所以就找不到该动态库了. `自行百度链接器和动
> 态链接器`.
>
> ps: 如果编译的是静态库, 就不会出现该问题了, 因为编译的时候已经加载到程序里了.


解决方案:

> 1.`//#cgo LDFLAGS: -L ./ -lxxx -Wl,-rpath -Wl,/path/to/lib` 编译时指定动态库路径.
> 
> 2.设置系统动态库的路径. 在 `/etc/ld.so.conf` 目录下添加动态库的路径(运行的机器)
>
> 3.用 C 手动打开动态库. 将编译的动态库添加到系统动态库 `/usr/local/lib` 下.


### C 代码抽出来后, 运行时提示找不到函数

```
/tmp/go-build649472735/b001/_x002.o: In function `_cgo_4167bb085d83_Cfunc_SayHello':
/tmp/go-build/cgo-gcc-prolog:49: undefined reference to `SayHello'
collect2: error: ld returned 1 exit status
```

原因分析:

> 在 cgo 的编译(依赖 C 源文件)过程中, 先会根据 `main.go` 生成一系列文件, 其中 `_cgo_export.c`, `main.cgo2.c`, 
`_cgo_main.c`, 这几个文件, 至于每个文件的内容前面讲过. 然后, gcc 分别编译每个文件和外部的 C 源文件, 生成 `*.o` 文件,
最后, 会将这些 `*.o` 文件合并到 `_cgo_.o` 文件当中. 由于没有编译外部的 C 源文件, 导致无法成功合并 `_cgo_.o` 文件.

解决方案:

> 1. 先将外部的 C 源文件编译成库, 然后使用库依赖的方式.
>
> 2. 在目标目录, 然后执行 `go build -o xxx` (xxx代表输出的文件名称, 注意这里没有 `xxx.go` 的参数), 这样 go 会自动
编译外部的 C 源文件(对于命令 `go run xx.go` 或`go build -o xx xx.go`, go 不会自动编译外部依赖的 C 源文件).


> 顺便说一下, 如果库依赖错误了, 比如要静态链接(包含了`#cgo LDFLAGS:-static`这样的选项), 但是却只提供了动态库(.so文件), 
也会产生上述的错误, 这种状况下要么修改链接方式, 要么增加静态库.


### 头文件声明变量, 提示重复定义错误

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

> 注意: **使用的是静态链接库**
>
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

> 1.去掉 header.c 的 `#include "header.h"` 就行了, cgo 会自动查找同名 c 文件. 如果要使用 header.h 里面定义的变
> 量 a, 使用 `extern int a;` 即可.
>
> 2.使用动态链接库, 编译的时候添加 `-fPIC` 参数, 这样会生产与位置无关的内容. (参考cgo/question案例)
>
> 3.增加static修饰符, 这样可以隐藏文件.
 

### 提示找不到头文件

```
fatal error: libxxx.h: No such file or directory
#include "libxxx.h"
          ^~~~~~~~~
compilation terminated.
```

原因分析: `libxxx.h` 没有添加到系统 include 下, 也没有设置编译查找路径.

解决方案:

> 增加编译选项 `#cgo CFLAGS: -I /path/to/header`


### 导出函数名称错误

```
export comment has wrong name "xxx", want "xxx"
```

在 Go 当中导出函数的限制:

- "Go函数名称" 和 "导出的C函数名称" 必须保持一致, 否则无导出.

-  导出的 Go 函数参数或返回值当中不能包含自定义的结构体(string, int, slice, array, map, chan, interface的别名
struct不算)