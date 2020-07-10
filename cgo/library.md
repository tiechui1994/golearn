## 静态库和动态库

[cgo使用注意](https://www.bandari.net/blog/24)

CGO 在使用 C/C++ 资源的时候一般有三种形式: 直接使用源码; 链接静态库; 链接动态库.

直接使用源码就是在 `import "C"` 之前的注释部分包含C代码, 或者在当前包中包含 C/C++ 源文件.

链接静态库和链接动态库方式比较类似, 都是通过在 `LDFLAGS` 选项指定要链接的库方式链接.


### 使用 C 静态库

如果 CGO 中引入的 C/C++ 资源有代码而且代码规模比较小, 直接使用源码是最理想的方式, 但是很多时候并没有源代码, 或者从 
C/C++ 源代码开始构建的过程异常复杂, 这种时候使用 C 静态库是一个不错的选择.

静态库因为是静态链接, 最终的目标程序并不会产生额外的运行时依赖, 也不会出现动态库特有的跨运行时资源管理的错误. 不过静态
库在链接阶段会有一定要求: 静态库一般包含了全部的代码, 里面会有大量的符号, 如果不同静态库之间出现了符号冲突则会导致链接
的失败.

```cgo
// math.h
#ifdef  _ADD_H
#define _ADD_H

int add(int,int);

#endif

// math.c
#include "add.h"
int add(int a, int b) {
    return a+b;
}
```

因为 CGO 使用的是 GCC 命令来编译和链接 C 和 Go 桥接的代码. 因此静态库也必须是GCC兼容的格式.

生成 libmath.a 静态库:

```
gcc -c -o math.o math.c
ar rcs libmath.a math.o
```

由于有 math 库的全部代码, 所以可以使用 `go generate` 工具来生成静态库, 或者是通过 `Makefile` 来构建静态库. 因此
发布 CGO 源码包, 并不需要提前构建 C 静态库.


为了支持 `go get` 命令直接下载并安装, C 语言的 `#include` 语法可以将 `math` 库的源文件链接到当前的包.

创建 `z_link_math.c` 文件如下:

```
#include "./math.c"
```

然后在执行 `go get` 或 `go build` 之类的命令时, CGO 就是自动构建 `math` 库对应的代码. 这种技术是在不改变静态库源
代码组织结构的前提下, 将静态库转化为源代码方式引用.  这种 CGO 包是最完美的.

---

如果使用的是第三方的静态库, 需要先下载安装静态库到合适的位置. 然后在 `#cgo` 指令中通过 `CFLAGS` 和 `LDFLAGS` 来指
定头文件和库的位置. 


### 使用 C 动态库

动态库出现的初衷是对于相同的库, 多个进程可以共享同一个, 以节省内存和磁盘资源. 

从库开发角度来说, 动态库可以隔离不同的动态库之间的关系, 减少链接时出现符号冲突的风险. 而且对于 windows 等平台,动态库
是跨越 VC 和 GCC 不同编译器平台的唯一的可行方式.

对于 CGO 来说, 使用动态库和静态库是一样的, 因为动态库也必须要有一个小的静态导出库(Linux下可以直接链接so文件, 但是在
windows 必须为 dll 创建一个 `.a` 文件用于链接). 

生成 libmath.so 动态库:

```
gcc -shared -fPIC -o libmath.so math.c
```


### 导出 C 静态库 和 动态库

```
import "C"

func main() {}

//export add
func add(a, b C.int) C.int {
	return a + b
}
```

> **注意, `export` 和 `//` 之间没有空格, 否则无法导出.** 

根据 CGO 文档的要求, 需要在 main 包中导出 C 函数. 对于 C 静态库构建方式来说, 会忽略 main 包中的 main 函数,只是简
单导出 C 函数. 采用以下命令构建:

```
go build -buildmode=c-archive -o math.a
```

导出动态库:

```
go build -buildmode=c-shared -o math.so
```


### 导出非 main 包的函数

C 静态库和 C 动态库的构建说明:

```
-buildmode=c-archive
    Build the listed main package, plus all packages it imports,
    into a C archive file. The only callable symbols will be those
    functions exported using a cgo //export comment. Requires
    exactly one main package to be listed.

-buildmode=c-shared
    Build the listed main package, plus all packages it imports,
    into a C shared library. The only callable symbols will
    be those functions exported using a cgo //export comment.
    Requires exactly one main package to be listed.
``` 

文档说明导出的 C 函数必须是在 main 包导出, 然后才能在生成的头文件包含声明的语句. 但是很多时候可能希望不同类型的导出函
数组织到不同的 Go 包中, 然后统一导出为一个静态库或动态库.

要实现从非 main 包导出 C 函数, 或者是多个包导出 C 函数, 需要自己提供导出 C 函数对应的头文件(因为CGO无法为非 main 包
的导出函数生成头文件)

