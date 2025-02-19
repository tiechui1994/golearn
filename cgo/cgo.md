# CGO 原理

[相关文章](https://colobu.com/2018/06/13/cgo-articles/)

[CGO编程](https://chai2010.cn/advanced-go-programming-book/ch2-cgo/ch2-02-basic.html)

## cgo语句

### cgo 编译参数与链接参数 

在 `import "C"`语句前的注释可以通过 `#cgo` 语句设置 `编译阶段` 和 `链接阶段` 的相关参数.

**编译阶段**的参数主要用于 `定义相关的宏` 和 `指定头文件检索路径`.

**链接阶段**的参数主要是 `指定库文件检索路径` 和 `要链接的库文件`.

案例:

```cgo
// #cgo CFLAGS: -D PNG_DEBUG=1 -I ./include
// #cgo LDFLAGS: -L /usr/local/lib -l png
// #include <png.h>
import "C"
```

上面的代码中:
 
`CFLAGS` 部分, `-D` 部分定义了 `PNG_DEBUG`, 值为 `1`; `-I` 定义了头文件包含的检索目录. 

`LDFLAGS` 部分, `-L` 指定了链接时文件检索目录, `-l` 指定了链接时需要链接 `png` 库.  

因为 C/C++ 遗留的问题, C 头文件检索目录可以是相对目录, 但是 `库文件检索目录` 则需要是绝对路径.

在库文件的检索目录中可以通过 `${SRCDIR}` 变量表示当前包含目录的绝对路径:

```cgo
// #cgo LDFLAGS: -L ${SRCDIR}/libs -l foo
```

上面的代码在链接时将被展开为:

```cgo
// #cgo LDFLAGS: -L /go/src/foo/libs -l foo
```

`#cgo` 语句主要影响 **`CFLAGS`, `CPPFLAGS`, `CXXFLAGS`, `FFLAGS`, `LDFLAGS`** 这几个编译器环境变量. 

① `LDFLAGS` 用于设置链接阶段的参数.

② `CFLAGS`, `CXXFLAGS`,`CPPFLAGS`, `FFLAGS` 这几个变量用于改变编译阶段的构建参数.


对于在cgo环境混合使用C和C++的用户来说, 可能有三种不同的编译选项: 其中 `CFLAGS` 对应 C 语言特有的编译选项, 
`CXXFLAGS` 对应是 C++ 特有的编译选项, `CPPFLAGS` 则对应 C 和 C++ 共有的编译选项.


但是在链接阶段, C 和 C++的链接选项是通用的, 因此这个时候已经不再有C和C++语言的区别, 它们的目标文件的类型是相同的.

### cgo 条件选择

`#cgo` 指令还支持条件选择, 当满足某个操作系统或某个CPU架构类型时类型时后面的编译或链接选项生效. 

条件选择:

```cgo
// #cgo windows CFLAGS: -D X86=1
// #cgo !windows LDFLAGS: -l math
```

宏定义案例:

```cgo
package main

/*
#cgo windows CFLAGS: -D CGO_OS_WINDOWS=1
#cgo darwin  CFLAGS: -D CGO_OS_DRWIN=1
#cgo linux   CFLAGS: -D CGO_OS_LINUX=1

#if defined(CGO_OS_WINDOWS)
   const char* os = "windows";
#elif defined(CGO_OS_DARWIN)
   const char* os = "darwin";
#elif defined(CGO_OS_LINUX)
   const char* os = "linux";
#else
   const char* os = "unknown";
#endif
*/
import "C"

func main() {
    print(C.GoString(C.os))
}
```

> 注意: 
>
> 1.在链接C库的使用, 不支持条件选择. 并且CGO参数有严格的格式 `#cgo CFLAGS:...` 或者 `#cgo LDFLAGS: ... `, 
> 即 `#cgo` 和 参数(`CFLAGS`, `LDFLAGS`) 
> 
> 2.对于C语言库(`.h` 文件定义内容 和 `.c` 文件实现 `.h` 的定义), 在CGO当中引用 `.h` 文件, 必须采用 `动态库/静态库` 
> 链接的方式, 否则可能无法编译通过.  
>
> 3.上述的 `CFLAGS`, `LDFLAGS` 会设置到编译环境变量 `CGO_LDFLAGS`, `CGO_CFLAGS` 当中, 作为 gcc 编译和链接的参数.

### 使用 PKG-CONFIG

使用 `#cgo CFLAGS`,`#cgo CXXFLAGS`, `#cgo LDFLAGS` 的方式是一种 hard code 的方式, 一旦第三方库发生变更, 其代码
也需要跟着变动. 可以使用 `pkg-config` 来避免这种情况.

一个 /lib/xxx/pkgconfig/libxxx.pc 文件
```
prefix=/lib/xxx
exec_prefix=${prefix}
liddir=${exec_prefix}/lib
includedir=${exec_prefix}/include

Name: xxx
Description: The xxx libary
Version: 0.1
Libs: -lxxx -L${liddir}
Cflags: -I${includedir}
```

编译的时候, 需要将上述的 pkgconfig 目录添加到 PKG_CONFIG_PATH 环境变量当中. `export PKG_CONFIG_PATH=/lib/xxx/pkgconfig`,
同时在动态库添加到 LD_LIBRARY_PATH 环境变量当中.  `export LD_LIBRARY_PATH=/lib/xxx/lib`. 最后, 就是在 CGO 当
中使用了.

```cgo
package main

/*
#cgo pkg-config: libxxx
#include <stdlib.h>
#include <xxx.h>
*/
import "C"

func main() {
    // xxxx
}
```

使用 `#cgo pkg-config` 去编译和链接. 


## 常用的cgo类型

### 数值类型

| C | CGO | Go |
| -- | -- | -- |
| char | C.char | byte | 
| singed char | C.schar | int8 |
| unsigned char | C.uchar | uint8 |
| short | C.short | int16 |
| int | C.int | int32 |
| long | C.long | int32 |
| long long int | C.longlong | int64 |
| float | C.float | float32 |
| double | C.double | float64 |
| size_t | C.size_t | uint | 

数值类型使用案例:

```cgo
/*
#include <stdio.h>

int call(int arg1, int* arg2, const char* arg3) {
    printf("arg1: %d\n", arg1);
    printf("arg2: %d\n", *arg2);
    printf("arg3: %s\n", arg3);
    
    *arg2 = 13579;
    return 0;
}
*/
import "C"
import (
    "fmt"
    "unsafe"
)

func main() {
    p1, p2, p3 := 1, 10, "Hello World"
    
    arg1 := C.int(p1)
    arg2 := (*C.int)(unsafe.Pointer(&p2))
    arg3 := C.CString(p3)
    
    var res C.int
    res = C.call(arg1, arg2, arg3)
    
    fmt.Println(*arg2, *(*int)(unsafe.Pointer(arg2)), res)
}
```

`call()` 函数有三个参数, 类型分别是 `int`, `int*`, `const char*`. 都是 C 语言类型. 

在 Go 当中要调用 `call()` 函数, 那么参数必须也是 C 类型的. `int` 类型使用 `C.int(p1)`, `int*` 类型使用 unsafe
包直接转换成 `*C.int`, `const char*` 使用 `C.CString(p3)` 直接将 string 转换成 `char*`.

在 Go 获取 `call()` 函数返回值(一般状况下, C语言返回值都是数值类型的), 那么返回值也必须是 C 类型的. 这里的返回值类型
是 `int`, 先声明一个变量为 `C.int`, 然后接收. 对于整数类型, 使用 `int(res)` 进行转换成 Go 类型, 但是对于 `指针`类
型, 必须要使用 unsafe 包转换成 Go 类型, 之后才能使用转换后的值, 否则会出现类型不匹配的问题. 

> 注: 在 C 语言当中 const char* ptr(指向字符常量的指针), char const* ptr(指向字符常量的指针), char* const ptr
> (指向字符的指针常数, 即const指针), 在 cgo 调用的时候, 全部都转换成 `*C.char`.

### 结构体, 联合, 枚举类型

#### 结构体

在 Go 当中, 可以通过 **C.struct_xxx** 来访问 C 语言中定义的 `struct xxx` 结构体类型.

结构体的内存按照C语言的通用对齐规则, 在 32 位 Go 语言环境 C 语言结构也按照 32 位对齐规则, 在 64 位Go语言环境按照 64 
位对齐规则. 对于指定了特殊对齐规则的结构体, 无法在 CGO 中访问.

> 注: 结构体当中出现了 Go 语言的关键字, 通过下划线的方式进行访问.

```cgo
/*
struct A {
    int   i;
    float f;
    int   type;
    char  chan;
};
*/
import "C"

func main() {
    var a C.struct_A
    
    fmt.Println(a.i)
    fmt.Println(a.f)
    fmt.Println(a._type)
    fmt.Println(a._chan)
}
```

> 如果有两个成员, 一个是以 Go 语言关键字命名, 另外一个刚好是以下划线和Go语言关键字命名, 那么以 Go 语言关键字命名的成员
将无法访问(被屏蔽)


C 语言结构体中 `位字段` 对应的成员无法在 Go 语言当中访问. 如果需要操作 `位字段` 成员, 需要通过在 C 语言当中定义辅助函
数来完成. 对应 `零长数组` 的成员, 无法在 Go 语言中直接访问数组的元素, 但其中 `零长数组` 的成员所在的位置偏移量依然可以
通过 `unsafe.Offset(a.arr)` 来访问.

```cgo
/*
struct A {
    int size:10; // 位字段无法访问
    float arr[]; // 零长的数组也无法访问
}
*/
import "C"

func main() {
    var a C.struct_A
    fmt.Println(a.size) 
    fmt.Println(a.arr)
}
```

> 注: 在 C 语言当中, 无法直接访问 Go 语言定义的结构体类型. 如果要想访问, 则必须通过函数的方式进行参数传递已达到间接访问
的目的.


结构体复杂的案例: (展示在 C 函数当中传递 `C结构体` 和 `C结构体指针的方式`)

```cgo
/*
#include <stdio.h>

typedef struct option {
    int iarg;
    float farg;
    const char* carg;
    int* iptr;
} option;

void call(option arg1, option* arg2) {
    printf("arg1 iarg: %d, farg: %0.2f, carg: %s, iptr: %d\n", arg1.iarg, arg1.farg, arg1.carg,
        *arg1.iptr);

    printf("\n==========================\n\n");

    printf("arg2 iarg: %d, farg: %0.2f, carg: %s, iptr: %d\n", arg2->iarg, arg2->farg, arg2->carg,
        *(arg1.iptr));
}
*/
import "C"
import (
    "unsafe"
)

func main() {
    val := 100
    opt := C.struct_option{
        iarg: C.int(10),
        farg: C.float(100.00),
        carg: C.CString("Hello World"),
        iptr: (*C.int)(unsafe.Pointer(&val)),
    }
    arg1 := *(*C.struct_option)(unsafe.Pointer(&opt))

    // 只能走 C malloc 路线
    // 确定申请内存大小, 并进行内存申请
    size := 2 * int(unsafe.Sizeof(C.struct_option{}))
    arg2 := (*C.struct_option)(C.malloc(C.size_t(size)))
    // unsafe 转换成数组, 对数组的元素进行赋值. 注意: p 的长度是 2, 内存占用是 48
    p := (*[2]C.struct_option)(unsafe.Pointer(arg2))[:]
    p[0] = *(*C.struct_option)(unsafe.Pointer(&opt))
    p[1] = *(*C.struct_option)(unsafe.Pointer(&opt))
	
    C.call(arg1, arg2)
}
```

在 C 中, 定义了一个 `option` 类型, call() 函数当中需要的参数类型分别是 `option` 和 `*option`,  call() 函数代码
就是打印参数里的值.

在 Go 中, 调用 C 中的 call() 函数. 创建一个 `struct_option` 类型的变量, 使用 `unsafe` 包直接将 `option` 变量直
接转换成 `C.struct_option` (call函数参数 arg1). 对于 call 函数参数 arg2, 转换相对复杂了一些, 详细步骤参考代码当中
的的注释.

Go 调用 C 的指针传递原则: **Go传递给C的 Go Pointer 所指向的 Go Memory 中不能包含任何指向Go Memory的Pointer**. 接
下来分情况讨论:

- 传递一个指向 struct 的指针.

```cgo
/*
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

typedef struct foo {
    int  a;
    int* p;
} foo;

void plus(foo* f) {
    (f->a)++;
    *(f->p)++;
}
*/
import "C"
import (
    "fmt"
    "unsafe"
)

func main() {
    f := C.struct_foo{
        a:5,
        p: (*C.int)(unsafe.Pointer(new(int))), 
    }
    // f 是指向 Go Memory(Go分配的), 指针 f.p 也是指向 Go Memory (new(int), 违反了Go 调用 C 的指针传递原则.
    // 将产生错误: cgo argument has Go pointer to Go pointer
    args := (*C.struct_foo)(unsafe.Pointer(&f))
    C.plus(args)
    fmt.Println(f.a)
}
```

- 传递一个指向 struct field 的指针.

```cgo
/*
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

typedef struct foo {
    int  a;
    int* p;
} foo;

void plus(int* f) {
    (*f)++;
}
*/
import "C"
import (
    "fmt"
    "unsafe"
)

func main() {
    f := C.struct_foo{
        a: 5,
        p: (*C.int)(unsafe.Pointer(new(int))),
    }
    args := (*C.int)(unsafe.Pointer(&f.a))
    
    // f.a, f.p 是指向 Go Memory 的, 因此可以成功执行
    C.plus(args)
    fmt.Println(f.a, *f.p)
}
```

- 传递一个指向 slice 或 array 中的 element 的指针.

```cgo
/*
void plus(int** f) {
    (**f)++;
}
*/
import "C"
import (
    "fmt"
    "unsafe"
)

func main() {
    val := make([]*int, 5)
    val[0] = new(int)
    args := (**C.int)(unsafe.Pointer(&val[0]))
    
    // val[0] 的地址指向的是 Go Memory, 违反了Go 调用 C 的指针传递原则.
    // 将产生错误: cgo argument has Go pointer to Go pointer
    C.plus(args)
    fmt.Println(*val[0])
}
```

针对上述的 1 和 3 的破局: 主要的问题出现在传递给 C 的 Go Pointer 所指向的 Go Memory 中包含任何指向 Go Memory 的Pointer, 
要想解决这个问题. 方式一, 在传递给 C 的 Go Pointer 所指向的 Go Memory 中, 将任何指向 Go Memory 的 Pointer 修改
为指向 C 的 Pointer. 方式二, 传递给 C 全部使用 C 内存空间(推荐).
 
方式一: 使用 C.malloc 分配子节点指针, 然后强制转换成相关类型
```
// 思路一, 消除Go子节点指针
args := make([]C.struct_foo, 1)
args[0] = C.struct_foo{
    a: 5,
    p: (*C.int)(unsafe.Pointer(C.malloc(C.size_t(4)))),
}
argsp := (*C.struct_foo)(unsafe.Pointer(&args[0]))
C.plus(argsp)
```

方式二: 先使用 C.malloc 开辟出 C 内存空间, 然后使用将开辟的 C 地址转换为 Go 的 slice, 最后往 slice 当中填充数据.
```
// 思路二, 开辟 C 内存空间
size := int(unsafe.Sizeof(C.struct_foo{}))
args := (*C.struct_foo)(unsafe.Pointer(C.malloc(C.size_t(size)))) 

// 使用数组的方式转换成 slice.
pa := (*[10]C.struct_foo)(unsafe.Pointer(args))[:] // [10]C.struct_foo 与 args 本质上是统一的
pa[0] = C.struct_foo{
    a: 5,
    p: (*C.int)(unsafe.Pointer(new(int))),
}

// 使用 SliceHeadrer 方式转换. 
sh := reflect.SliceHeader{
    Data: uintptr(unsafe.Pointer(args)),
    Len:1,
    Cap:1,
}
ps := *(*[]C.struct_foo)(unsafe.Pointer(&sh)) // []C.struct_foo 与 reflect.SliceHeader 本质上是统一的
pa[0] = C.struct_foo{
    a: 5,
    p: (*C.int)(unsafe.Pointer(new(int))),
}
```

**CGO函数调用的细节点:**

1. 在 C 语言当中, 数组和指针是等价的.

2. 在 C 语言当中 `const xxx* ptr`(指向xxx常量的指针), `xxx const* ptr`(指向xxx常量的指针), `xxx* const ptr`
(指向xxx的指针常数, 即const指针), 在 cgo 调用的时候, 全部都转换成 `*C.struct_xxx`.

3. 在 CGO 当中, 需要在 Go 当中定义和 C 中 "一致" 的结构体, 方便 `unsafe` 转换. 所谓一致, 就是 Go 结构体当中成员的
类型都是 `C.xxx` 或 `*C.xxx` 类型.

4. 在调用 C 语言函数时, 若参数类型需要转换成 `C.struct_xxx`, 则可以直接使用 `unsafe` 包进行直接转换.

5. 在调用 C 语言函数时, 若参数类型需要转换成 `*C.struct_xxx`, 必须按照 `C数组` 的方式进行转换. 即 `确定内存长度`,
`malloc分配内存`, `unsafe包转换成数组`, `对数组的元素进行赋值`. 这一点非常的重要.


> 函数指针 vs 回调函数:
> 如果在程序中定义了一个函数, 那么在编译时系统就会为这个函数代码分配一段存储空间, 这段存储空间的首地址称为这个函数的
地址. 而且函数名表示的就是这个地址. 既然是地址我们就可以定义一个指针变量来存放, 这个指针变量叫做函数指针变量, 简称函
数指针.
> 
> 回调函数, 就是函数指针作为某个函数的参数.


> 关于 C 当中类型的定义:

```
1. 自定义数据类型, 方便移植

typedef struct XXX {
    
} XXX;

或

typedef struct XXX {
} 

// 使用的类型: "Xxx x", "Xxx* x"


2. 别名

struct XXX {
    
} XXX;

// 使用的类型: "struct Xxx x", "struct Xxx* x" 
```

由于 C 和 Go 函数调用的方式不同, 这就导致了 C 和 Go 的函数指针无法进行直接调用, 这就需要各自在自己的空间内去调用.

Go 当中调用 C 的回调函数, 需要在 C 当中提供执行函数 invoke 函数, 然后将 C 的回调函数交给 invoke 去执行.

C 当中调用 Go 的回调函数, 只需要将 Go 的函数导出即可.

CGO 函数指针桥梁: Go 指针(`*[0]byte`) 或 C 函数指针(需要使用 typedef 定义) 传递给函数. 具体案例参考 [callback.go](./callback.go) 代码.

#### 联合类型

对于联合类型,可以通过 `C.union_xxx` 来访问 C 语言中定义的 `union xxx` 类型. **但是 Go 语言中并不支持C语言联合类型,
它们会被转换为对应大小的字节数组.**

```cgo
/*
#include <stdint.h>

union B {
    int   i;
    float f;
};

union C {
    int8_t  i8;
    int64_t i64;
};
*/
import "C"

func main() {
    var b C.union_B
    fmt.Printf("%T \n", b) // [4]uint8 
    
    var c C.union_C
    fmt.Printf("%T \n", c) // [8]uint8
}
```

如果需要操作C语言的联合类型变量, 一般有三种办法:

第一种是在C语言当中定义辅助函数;

第二种是通过 Go 语言的 `encoding/binary` 手工解码成员(需要注意大小端的问题)

第三种是使用 `unsafe` 包强制转换为对应的类型(性能最好的方式)

```
var ub C.union_B
fmt.Println("b.i:", binary.LittleEndian.Uint32(ub[:]))
fmt.Println("b.f:", binary.LittleEndian.Uint32(ub[:]))

fmt.Println("b.i:", *(*C.int)(unsafe.Pointer(&ub)))
fmt.Println("b.f:", *(*C.float)(unsafe.Pointer(&ub)))
```

> 虽然unsafe包访问最简单, 性能也最好, 但是对于有**嵌套联合类型**的情况处理会导致问题复杂化. 对于复杂的联合类型, 推荐通
过在C语言中定义辅助函数的方式处理.

#### 枚举类型

对于枚举类型, 通过 `C.enum_xxx` 访问 C 语言当中定义的 `enum xxx` 结构体类型

```cgo
/*
enum C {
    ONE,
    TWO,
};
*/
import "C"

main(){
    var c C.enum_C = C.TWO
    fmt.Println(c)
    fmt.Println(C.ONE)
}
```

> 在 C 语言当中, 枚举类型底层对应 int 类型, 支持负数类型的值. 可以通过 C.ONE, C.TWO 等直接访问定义的枚举值.

### 数组, 字符串和切片

- C 语言当中的`数组,指针和字符串`

在 C 语言当中, 数组名其实对应着一个指针, 指向特定类型特定长度的一段内存, 但是这个指针不能被修改.

当把数组名传递给一个函数时, 实际上传递的是数组第一个元素的地址.

C 语言的字符串是一个 char 类型的数组, 字符串的长度需要根据表示结尾的NULL字符的位置确定.


- Go 语言当中的 `数组,字符串和切片`

Go 当中, 数组是一种值类型, 而且数组的长度是数组类型的一部分.

Go 当中, 字符串对应一个长度确定的只读 byte 类型的内存.

Go 当中, 切片是一个简化版的动态数组.


- CGO 当中`数组,字符串和切片`转换的本质

Go 语言 和 C 语言的数组, 字符串和切片之间的相互转换可以简化为 `Go 语言的切片和 C 语言中指向一定长度内存的指针` 之间的
转换.


- 系统内存的转换函数(主要针对的是字符串和字节数组)

```
// Go String to C String 
func C.CString(string) *C.char

// Go []byte to C Array
func C.CBytes([]byte) unsafe.Pointer


// C String to Go String
func C.GoString(*C.char) string

// C data with explicit length to Go String
func C.GoStringN(*C.char, C.int) string

// C data with explicit length to Go []byte
func C.GoBytes(unsafe.Pointer, C.int) []byte
```

> C.CString 针对输入 Go 字符串, 克隆一个 C 语言格式的字符串; 返回的字符串由 C 语言的 malloc 函数分配, 不使用时需要
> 通过 C 语言的 free 函数释放.
>
> C.Cbytes 函数和 C.CString 类似, 针对输入的 Go 语言切片克一个 C 语言版本的字符数组.
>
> C.GoString 用于将从 NULL 结尾的 C 语言字符串克隆一个 Go语言字符串.
>
> C.GoStringN 是另一个字符数组克隆函数.
> 
> C.GoBytes 用于从 C 语言数组, 克隆一个 Go 语言字节切片.

当 C 语言字符串或数组向 Go 语言转换时, 克隆的内存由 Go 语言分配管理. 通过该组转换函数, 转换前和转换后的内存依然在各自的
语言环境中, 它们并没有跨域 Go 语言和 C 语言.

克隆方式实现转换的优点是接口和内存管理简单. 缺点是克隆需要分配新的内存和复制操作都会导致额外的开销.

reflect 中字符串和切片的定义:

```
type StringHeader struct{
    Data uintptr
    Len int
}

type SliceHeader struct {
    Data uintptr
    Len int
    Cap int
}
```

`C 语言字符串, 数组转换成 Go 语言的字符串, 数组` 案例:

```cgo
/*
#include <string.h>

char arr[10];
char *s = "Hello";
**/
import "C"

func main() {
    // 通过 reflect.SliceHeader 转换
       var arr []byte
       array := (*reflect.SliceHeader)(unsafe.Pointer(&arr))
       array.Data = uintptr(unsafe.Pointer(&C.arr[0]))
       array.Len = 10
       array.Cap = 10
   
       // 切片
       arr1 := (*[31]byte)(unsafe.Pointer(&C.arr[0]))[:10:10]
   
       // 通过 reflect.StringHeader 转换
       var s string
       str := (*reflect.StringHeader)(unsafe.Pointer(&s))
       str.Data = uintptr(unsafe.Pointer(C.s))
       str.Len = int(C.strlen(C.s))
   
       // 切片
       length := int(C.strlen(C.s))
       s1 := string((*[31]byte)(unsafe.Pointer(C.s))[:length:length])
   
       fmt.Println("arr:", string(arr), "arr1:", string(arr1), "s:", s, "s1:", s1)
}
```

需要注意上述代码的使用场景, 是在 Go 当中访问 C 当中的变量. 反之, 在调用 C 函数, 将 Go 变量转换成 C 变量的时候, 上述代
码是非法无效的 (`runtime error: cgo argument has Go pointer to Go pointer`), 必须要通过前面介绍的方法进行转换.

> 注: 上述代码中, 是在Go当中直接访问C的内存空间. 即 C 和 Go 共享变量.

> Go 字符串是只读的, 用户需要自己保证 Go 字符串在使用期间, 底层对应的 C 字符串内容不会发生变化, 内存不会被提前释放掉.


### 类型转换

#### 指针间的转换

- C 和 Go 关于指针的区别

在 C 语言中, 不同类型的指针是可以显式或隐式转换的, 如果是隐式只是会在编译时给出一些警告信息.

Go 语言对于不同类型的转换非常严格, 任何 C 语言中可能出现的警告信息在 Go 语言中都可能是错误!

指针是 C 语言的灵魂, 指针间的自由转换也是 cgo 代码中经常要解决的第一个问题.


- Go 当中两个指针之间的转换

在 Go 语言中两个指针的类型完全一致则不需要转换可以直接使用. 如果一个指针类型是用 type 命令在另一个指针类型基础上构建的, 
换言之 `两个指针是 "底层结构完全相同" 的指针`, 那么可以通过直接强制转换语法进行指针间的转换.  但是 cgo 经常要面对是2个
完全不同类型的指针间的转换, 原则上这种操作在纯Go 语言代码是严格禁止的.

```
var p *X
var q *Y

q = (*X)(unsafe.Pointer(p)) // *X => *Y
p = (*Y)(unsafe.Pointer(q)) // *Y => *X
```

为了实现 X 类型和 Y 类型的指针的转换, 需要借助 `unsafe.Pointer` 作为中间桥接类型实现不同类型指针之间的转换. 
`unsafe.Pointer` 指针类型类似 C 语言中的 `void*` 类型的指针.

指针简单转换流程图:

![image](/images/cgo_xtoy.png)


- C 和 Go 关于空指针的传递

在 C 语言当中, 指针类型 `char*`, `void*`, `int*` 等等, 其空值为 `NULL`. 

> 特别注意: NULL 是一个宏变量, 定义在 `<stdlib.h>` 当中, 在 Go 中可以通过 `C.NULL` 获取其值(0), 不能将 `C.NULL` 传递给 C

在 Go 语言当中, 指针类型的空值为 `nil`. 

在 Go 调用 C 的时候, 如果函数的参数需要传值为 `NULL` 或某个参数内的成员变量为指针类型, 且此时需要传值为 `NULL`,
这个时候该怎么解决呢? 对于后一种情况, 比较简单 `成员变量不要赋值即可`. 但是前一种状况, 是必须要有值传递的, 此时传递的值应
该是 `(*C.char)(unsafe.Pointer(nil))` (这里是 `char*` 类型), `unsafe.Pointer(nil)` (这里是 `void*` 类型) . 

> 注意: 在第一种状况下, 传递 C.NULL, 程序是会报错的. 因为 C.NULL 在 Go 当中就是一个整数类型, 并不是指针类型.

#### 数值和指针的转换

为了严格控制指针的使用, Go 语言禁止将数值类型直接转为指针类型! 不过, Go 语言针对 `unsafe.Pointer` 指针类型特别定义了
一个 `unitptr` 类型. 可以以 `unitptr` 为中介, 实现数值类型到 `unsafe.Pointer` 指针类型的转换. 再结合前面提到的方
法, 就可以实现数值类型和指针的转换了.

int32 类型到 C 语言的 `char*` 字符串指针类型的相互转换:

![image](/images/cgo_numtoptr.png)

#### 切片间的转换

在 C 语言当中数组也是一种指针, 因此两个不同类型数组之间的转换和指针类型间转换基本类似.

在 Go 语言当中, 数组或数组对应的切片不再是指针类型, 因此无法直接实现不同类型的切片之间的转换.

在 Go 的 `reflect`包提供了切片类型的底层结构, 再结合前面不同类型直接的指针转换, 可以实现 []X 到 []Y 类型的切片转换.

```
var p []X
var q []Y

src := (*reflect.SliceHeader)(unsafe.Pointer(&p))
dst := (*reflect.SliceHeader)(unsafe.Pointer(&q))

dst.Data = src.Data
dst.Len = src.Len * int(unsafe.Sizeof(p[0])) / int(unsafe.Sizeof(q[0]))
dst.Cap = src.Len * int(unsafe.Sizeof(p[0])) / int(unsafe.Sizeof(q[0]))
```


## 内部机制

### CGO 生成的中间文件

在构建一个 cgo 包时增加一个 `-work` 输出中间生成所在目录并且在构建完成时保留中间文件.

对于比较简单的 cgo 代码可以直接手工调用 `go tool cgo` 命令来查看生成的中间文件.

在一个 Go 源文件当中, 如果出现 `import "C"` 指令则表示将调用 cgo 命令生成对应的中间文件. 下面是生成的中间文件的简单
示意图:

![image](/images/cgo_middle_process.png)


包含有 4 个 Go 文件, 其中 nocgo 开头的文件中没有 `import "C"` 指令, 其他的 2 个文件则包含了 cgo 代码. cgo 命令
会为每个包含 cgo 代码的 Go 文件创建 2 个中间文件, 比如 main.go 会分别创建 **main.cgo1.go** 和 **main.cgo2.c** 
两个中间文件, cgo当中是 Go 代码部分和 C 代码部分.

然后会为整个包创建一个 **_cgo_gotypes.go** Go 文件, 其中包含 Go 语言部分辅助代码. 此外还创建一个 **_cgo_export.h**
和 **_cgo_export.c** 文件, 对应 Go 语言导出到 C 语言的类型和函数.

```cgo
package main

/*
int sum(int a, int b) { 
  return a+b; 
}
*/
import "C"

func main() {
    println(C.sum(1, 1))
}
```

使用 `go tool cgo main.go` 生成的中间文件如下:

```
_cgo_export.c
_cgo_export.h
_cgo_flags
_cgo_gotypes.go
_cgo_main.c
_cgo_.o
main.cgo1.go
main.cgo2.c
```

#### main.cgo1.go

```
//main.cgo1.go 

package main

/*
int sum(int a, int b) {
  return a+b; 
}
*/
import _ "unsafe"

func main() {
    println((_Cfunc_sum)(1, 1))
}
```

其中 `C.sum(1,1)` 函数调用替换成了 `(_Cfunc_sum)(1,1)`. 每一个 `C.xxx` 形式的函数都会被替换成 `_Cfunc_xxx` 格
式的纯 Go 函数, 其中前缀 `_Cfunc_` 表示这是一个C函数, 对应一个私有的 Go 桥接函数.

#### _cgo_gotypes.go

`_Cfunc_sum` 函数在 cgo 生成的 _cgo_gotypes.go 文件当中定义:

```
// _cgo_gotypes.go

//go:linkname _cgo_runtime_cgocall runtime.cgocall
func _cgo_runtime_cgocall(unsafe.Pointer, uintptr) int32

//go:linkname _cgo_runtime_cgocallback runtime.cgocallback
func _cgo_runtime_cgocallback(unsafe.Pointer, unsafe.Pointer, uintptr, uintptr)

//go:cgo_import_static _cgo_7b5139e7c7da_Cfunc_sum
//go:linkname __cgofn__cgo_7b5139e7c7da_Cfunc_sum _cgo_7b5139e7c7da_Cfunc_sum
var __cgofn__cgo_7b5139e7c7da_Cfunc_sum byte
var _cgo_7b5139e7c7da_Cfunc_sum = unsafe.Pointer(&__cgofn__cgo_7b5139e7c7da_Cfunc_sum)

//go:cgo_unsafe_args
func _Cfunc_sum(p0 _Ctype_int, p1 _Ctype_int) (r1 _Ctype_int) {
    _cgo_runtime_cgocall(_cgo_7b5139e7c7da_Cfunc_sum, uintptr(unsafe.Pointer(&p0)))
    if _Cgo_always_false {
        _Cgo_use(p0)
        _Cgo_use(p1)
    }
    return
}
```

`_Cfunc_sum` 函数的参数和返回值 `_Ctype_int` 类型对应 `C.int` 类型, 命名的规则和 `_C_func_xxx` 类似, 不同的前
缀用于区分函数和类型. 

`_cgo_7b5139e7c7da_Cfunc_sum` 指针来自 `__cgofn__cgo_7b5139e7c7da_Cfunc_sum`, 而该值是通过 `go:linkname`
连接到 `_cgo_7b5139e7c7da_Cfunc_sum`, 该值是又是通过 `go:cgo_import_static` 从 C 的静态库当中导入的符号. 该
符号定义在了 `main.cgo2.c` 文件当中.

`_cgo_runtime_cgocall` 是对 `runtime.cgocall` 函数的连接, 声明如下:

```cgo
func runtime.cgocall(fn, arg unsafe.Pointer) int32
```

> 第一个参数是 C 语言函数的地址
> 第二个参数是存储 C 语言函数对应的参数结构体(参数和返回值)的地址

在此例当中, 被传入C语言函数 `_cgo_7b5139e7c7da_Cfunc_sum` 也是 cgo 生成的中间函数. 函数定义在 main.cgo2.c 当中.

#### main.cgo2.c

```
// main.cgo2.c

extern char* _cgo_topofstack(void);

#define CGO_NO_SANITIZE_THREAD
#define _cgo_tsan_acquire()
#define _cgo_tsan_release()
#define _cgo_msan_write(addr, sz)

void _cgo_7b5139e7c7da_Cfunc_sum(void *v)
{
    struct {
        int p0;
        int p1;
        int r;
        char __pad12[4];
    } __attribute__((__packed__, __gcc_struct__)) *_cgo_a = v;
    char *_cgo_stktop = _cgo_topofstack();
    __typeof__(_cgo_a->r) _cgo_r;
    _cgo_tsan_acquire();
    _cgo_r = sum(_cgo_a->p0, _cgo_a->p1);
    _cgo_tsan_release();
    _cgo_a = (void*)((char*)_cgo_a + (_cgo_topofstack() - _cgo_stktop));
    _cgo_a->r = _cgo_r;
    _cgo_msan_write(&_cgo_a->r, sizeof(_cgo_a->r));
}
```

此函数参数只有一个 `void*` 的指针, 函数没有返回值. 真实的 sum 函数的函数参数和返回值通过唯一的参数指针类实现.

其中 `_cgo_tsan_acquire` 和 `_cgo_tsan_acquire`, 函数是定义在 `libcgo.h` 当中. 函数 `_cgo_topofstack`
是汇编实现(获取栈顶位置, g->m->curg.stack.hi)

```
struct {
        int p0;
        int p1;
        int r;
        char __pad12[4];
} __attribute__((__packed__, __gcc_struct__)) *_cgo_a = v;
```

其中, p0, p1分别对应 sum 的第一个和第二个参数, r 对应 sum 的返回值. `_pad12` 用于填充结构体保证对齐CPU机器字的整数
倍.

> 然后从参数执行的结构体获取参数后开始调用真实的C语言版sum函数, 并且将返回值保存到结构体的返回值对应的成员.


因为 Go 语言和 C 语言有着不同的内存模型和函数调用规范, 其中 `_cgo_topofstack` 函数相关的代码用于 C 函数调用后恢复调
用栈. `_cgo_tsan_acquire` 和 `_cgo_tsan_release` 则是用于扫描 CGO 相关函数的指针总相关检查.


```cgo
// Call from Go to C.
//
// 这里使用 nosplit, 是因为它用于某些平台上的系统调用. 系统调用可能在栈上有 untyped 参数, 因此 grow 或 scan
// 只能都是不安全的. 
//
//go:nosplit
func cgocall(fn, arg unsafe.Pointer) int32 {
    if !iscgo && GOOS != "solaris" && GOOS != "illumos" && GOOS != "windows" {
        throw("cgocall unavailable")
    }
    
    // 函数检查
    if fn == nil {
        throw("cgocall nil")
    }
    
    if raceenabled {
        racereleasemerge(unsafe.Pointer(&racecgosync))
    }
    
    mp := getg().m // 获取当前的 m
    mp.ncgocall++  // 统计 m 调用 CGO 的次数
    mp.ncgo++      // 周期内调用的次数
    
    // Reset traceback.
    mp.cgoCallers[0] = 0 // 如果在 cgo 中 crash. 记录 CGO 的 traceback
    
    // 进入系统调用, 以便调度程序创建新的 M 来执行 G
    // 对于 asmcgocall 的调用保证不会增加 stack 并且不会分配内存, 因此在超出 $GOMAXPROCS
    // 之外的 "系统调用中" 中调用是安全的.
    // fn 可能会会掉到 Go 代码中, 这种情况下, 将退出 "system call", 运行 Go 代码(可能会增加堆栈),
    // 然后重新进入 "system call". PC 和 SP在这里被保存.
    entersyscall() // 进入系统调用前期准备工作. M, P 分离, 防止系统调用阻塞 P 的调度, 保存上下文.
    
    // 告诉异步抢占我们正在进入外部代码. 在 entersyscall 之后这样做, 因为这可能会阻塞并导致异步抢占失败,
    // 但此时同步抢占会成功(尽管这不是正确性的问题)
    osPreemptExtEnter(mp) // 在 linux 当中是空函数
    
    mp.incgo = true
    errno := asmcgocall(fn, arg) // 切换到 g0, 调用 C 函数 fn, 汇编实现
    
    // Update accounting before exitsyscall because exitsyscall may
    // reschedule us on to a different M.
    mp.incgo = false
    mp.ncgo--
    
    osPreemptExtExit(mp) // 在 linux 当中是空函数
    
    exitsyscall() // 退出系统调用, 寻找 P 来绑定 M
    
    // Note that raceacquire must be called only after exitsyscall has
    // wired this M to a P.
    if raceenabled {
        raceacquire(unsafe.Pointer(&racecgosync))
    }
    
    // 防止 Go 的 gc, 在 C 函数执行期间回收相关参数, 用法与前述_Cgo_use类似
    KeepAlive(fn) 
    KeepAlive(arg)
    KeepAlive(mp)
    
    return errno
}
```

Go 调入 C 之后, 程序的运行将不受 Go 的 runtime 的管控. 一个正常的 Go 函数是需要 runtime 的管控的, 即函数的运行时间
过长导致 goroutine 的抢占, 以及 GC 的执行会导致所有的 goroutine 被挂起.

C 程序的执行, 限制了 Go 的 runtime 的调度行为. 为此, Go 的 runtime 会在进入 C 程序之前, 标记这个运行 C 的线程 M,
将其排除在调度之外.

由于正常的 Go 程序运行在一个 2k 的栈上, 而 C 程序需要一个无穷大的栈, 因此在进入 C 函数之前需要把当前线程的栈从 2k 切换
到线程本身的系统栈上, 即切换到 g0.

asmcgocall 采用汇编实现:

// runtime/asm_amd64.s

```cgo
// func asmcgocall(fn, arg unsafe.Pointer) int32
// fn 是函数地址, arg 是第一个参数地址
// 在 g0 上调用 fn(arg) 函数.
TEXT ·asmcgocall(SB),NOSPLIT,$0-20
    MOVQ    fn+0(FP), AX
    MOVQ    arg+8(FP), BX
    
    MOVQ    SP, DX // 保存当前的 SP 到 DX
    
    // Figure out if we need to switch to m->g0 stack.
    // We get called to create new OS threads too, and those
    // come in on the m->g0 stack already.
    // 切换 g 之前的检查
    get_tls(CX)
    MOVQ    g(CX), R8 // R8 = g
    CMPQ    R8, $0    // g == 0
    JEQ    nosave // 相等跳转, 则说明当前 g 为空
    MOVQ    g_m(R8), R8 // 当前 m
    MOVQ    m_g0(R8), SI // SI = m.g0
    MOVQ    g(CX), DI    // DI = g  
    CMPQ    SI, DI // m.g0 == g
    JEQ    nosave // 相等跳转, 当前在 g0 上
    MOVQ    m_gsignal(R8), SI // SI = m.gsignal
    CMPQ    SI, DI // m.gsignal == g
    JEQ    nosave // 相等跳转, 当前 m.gsignal 上
    
    // 切换到 g0 上
    MOVQ    m_g0(R8), SI // SI=m.g0
    CALL    gosave<>(SB) // 调用 gosave, 参数是 gobuf
    MOVQ    SI, g(CX) // 切换到 g0
    MOVQ    (g_sched+gobuf_sp)(SI), SP // 恢复 g0 的 SP 
    
    // Now on a scheduling stack (a pthread-created stack).
    // Make sure we have enough room for 4 stack-backed fast-call
    // registers as per windows amd64 calling convention.
    SUBQ    $64, SP     // SP=SP-64
    ANDQ    $~15, SP    // SP=SP+16, 偏移 gcc ABI
    MOVQ    DI, 48(SP)    // 保存 g 
    MOVQ    (g_stack+stack_hi)(DI), DI // DI=g.stack.hi
    SUBQ    DX, DI       // 计算 g 栈大小, 保存到 DI 当中
    MOVQ    DI, 40(SP)    // 保存 g 栈大小(这里不能保存 SP, 因为在回调时栈可能被拷贝)
    MOVQ    BX, DI        // DI = first argument in AMD64 ABI
    MOVQ    BX, CX        // CX = first argument in Win64
    CALL    AX          // 调用函数, 参数 DI, SI, CX, DX, R8
    
    // 函数调用完成, 恢复到 g, stack
    get_tls(CX)
    MOVQ    48(SP), DI // DI=g
    MOVQ    (g_stack+stack_hi)(DI), SI // SI=g.stack.hi
    SUBQ    40(SP), SI // SI=SI-size
    MOVQ    DI, g(CX)  // tls 保存, 恢复到 g 
    MOVQ    SI, SP     // 恢复 SP
    
    MOVL    AX, ret+16(FP) // 函数返回错误码
    RET

nosave:
    // 在系统栈上运行, 甚至可能没有 g.
    // 在线程创建或线程拆除期间可能没有 g 发生(例如, 参见 Solaris 上的 needm/dropm).
    // 这段代码和上面的代码作用是一样的, 但没有saving/restoring g, 并且不用担心 stack 移动(因为我们在系统栈上,
    // 而不是在 goroutine 堆栈上).
    // 如果上面的代码已经在系统栈上, 则可以直接使用, 但是通过此代码的唯一路径在 Solaris 上很少见.
    SUBQ    $64, SP
    ANDQ    $~15, SP
    MOVQ    $0, 48(SP)    // where above code stores g, in case someone looks during debugging
    MOVQ    DX, 40(SP)    // save original stack pointer
    MOVQ    BX, DI        // DI = first argument in AMD64 ABI
    MOVQ    BX, CX        // CX = first argument in Win64
    CALL    AX
    MOVQ    40(SP), SI    // restore original stack pointer
    MOVQ    SI, SP
    MOVL    AX, ret+16(FP)
    RET
```

**当Go调用C函数时, 会单独占用一个系统线程. 因此如果在 Go协程中并发调用C函数, 而C函数中又存在阻塞操作,就很可能会造成Go
程序不停的创建新的系统线程,而Go并不会回收系统线程,过多的线程会拖垮整个系统**


调用链(从调用者角度来看): 
```cgo
C.sum (调用方) =>
    _Cfunc_sum(main.cgo1.go, 调用_cgo_gotypes.go当中的代码) =>
        _cgo_runtime_cgocall(_cgo_gotypes.go当中导入C符号表, 调用汇编实现的runtime.cgocall函数, asm_amd64.s) =>
            _cgo_7b5139e7c7da_Cfunc_sum(main.cgo2.c, 调用 C 库的 sum 函数) =>
                sum(C当中的sum函数)
```

文件链: 
```
main.go -> main.cgo1.go -> _cgo_gotypes.go -> main.cgo2.c
```


理解 CGO 调用的黑盒子是 `go:cgo_import_static` 将 C 函数加载到 Go 空间中, `go:linkname` Go 当中的链接. 接下来就
是 `runtime.cgocall`. 关于该函数, 在 export.go 当中有详细介绍.


## CGO 内存模型

CGO 是架接 Go 语言和C语言的桥梁, 它使二者在二进制接口层面实现了互通, 但是要注意两种语言的内存模型的差异而可能引起的问
题.

如果在 CGO 处理的跨语言函数调用时涉及到了指针的传递, 则可能会出现 Go 语言和 C 语言共享某一段内存的场景. 

> C 与 Go 之间的内存差别差别:

C 语言的内存在分配之后就是稳定的, 但是 Go 语言因为函数栈的动态伸缩可能导致栈中内存地址的移动(这是Go和C内存模型的最大差
异.) 如果 C 语言持有的是移动之前的 Go 指针, 那么以旧指针访问 Go 对象就会导致程序崩溃.


### Go 访问 C 内存

C 语言空间的内存是稳定的, 只要不是被人为提前释放, 那么在 Go 语言空间可以放心大胆地使用. 在 Go 语言访问 C 语言内存是最
简单的情形.

案例: 因为 Go 语言限制, 无法在 Go 语言当中的创建大于 2GB 内存的切片. 借助 cgo 技术, 可以实现在 C 语言环境创建大于 
2GB 的内存, 然后转为 Go 语言的切片使用.

C 内存管理函数:

```
头文件: <stdlib.h>
声明: void* malloc(int n)
含义: 在堆上, 分配n个字节, 并返回 void 指针类型.
返回值: 分配成功, 返回分配在堆上存储空间的首地址; 否则返回 NULL


头文件: <stdlib.h>
声明: void* calloc(int n, int size)
含义: 在堆上, 分配 n*size 个字节, 并初始化为0, 并返回 void 指针类型.
返回值: 同 malloc


头文件: <stdlib.h>
声明: void* realloc(void* p, int n)
含义: 重新分配堆上的void指针p所指的空间为n个字节, 同时会赋值原有内容到新分配的堆上存储空间. 注意, 若原来的void指针p在
堆上的空间不大于n个字节, 则保持不变.
返回值: 同 malloc


头文件: <stdlib.h>
声明: void free(void* p)
含义: 释放 void 指针 p 所指的堆上的空间.


头文件: <string.h>
声明: void* memset(void* p, int c, int n)
含义: 对于void指针p为首地址的n个字节, 将其中的每个字节设置为c.
返回值: 返回指向存储区域 p 的void类型指针.
```

C 内存管理案例(`malloc`, `realloc`, `memset`, `free`):

```cgo
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define SIZE 5

void main() {
    long int* arr, d;
    int i;
    int length = SIZE;

    arr = (long int*)malloc(length*sizeof(long int));
    memset(arr, 0, length*sizeof(long int));

    printf("input numbers util you input zero: \n");

    for(;;) {
        printf("> ");
        scanf("%ld", &d);
        if (d==0) {
            arr[i++]=0;
            break;
        } else {
            arr[i++]=d;
        }

        if (i >= length) {
            // 重新分配 2*(length*size) 大小的内存, 并把后半部分 length*size 大小的内存设置为 0
            arr = (long int*)realloc(arr, 2*length*sizeof(long int));
            memset(arr+length, 0, length*sizeof(long int));
            length *= 2;
        }
    }

    printf("\n");
    printf("output all numbers: \n");
    for (int idx=0; idx<i; idx++) {
        if (idx && idx%5==0) {
            printf("\n");
        }

        printf("%ld\t", *(arr+idx));
    }

    printf("\n");

    free(arr);
}
```

内存分配案例:

```cgo
/*
#include <stdlib.h>

void* makeslice(size_t memsize) {
    return malloc(memsize);
}
*/
import "C"

func makeByteSlice(n int) []byte {
    p := C.makeslice(C.size_t(n))
    return ((*[1<<31]byte)(p))[0:n:n]
}

func freeByteSlice(p []byte) {
    C.free(unsafe.Pointer(&p[0]))
}
```

> 注: `free()` 是 `<stdlib.h>` 当中的函数.


### C 临时访问传入的 Go 内存

一个极端场景: 将一块位于某 goroutine 的栈上的 Go 语言内存传入了 C 语言函数后, 在此 C 语言函数执行期间, 此 goroutine 
的栈因为空间不足的原因发生了扩展, 也就是导致了原来的 Go 语言内存被移动到了新的位置. 但是此时此刻 C 语言函数并不知道该 
Go 语言内存已经移动了位置, 仍然使用之前的地址来操作该内存 -- 这将导致内存越界.

以上是一个推论(真实情况有些差异), 也就是说 **C 访问传入的 Go 内存可能是不安全的!**

当然有 RPC 远程过程调用经验的用户可能会考虑通过 `完全传值` 的方式处理: 借助 C 语言内存稳定的特性, 在 C 语言空间先开辟
同样大小的内存, 然后将 Go 的内存填充到 C 的内存空间; 返回的内存也是如此处理.


```cgo
/*
#include <stdio.h>
#include <stdlib.h>

void printString(const char* s) {
    printf("%s \n", s);
}
*/
import "C"

func printString(s string) {
    cs := C.CString(s) // 构建 C 语言字符串
    defer C.free(unsafe.Pointer(cs)) // 释放 C 语言字符串
    
    C.printString(cs) // 函数调用, 使用的是 C 自己创建的字符串
}

func main() {
    printString("Hello")
}
```

上面的处理思路是安全的, 但是效率极其低下(因为要多次分配内存并逐个复制元素), 同时也及其繁琐.

为了简化并高效处理此种 C 语言传入 Go 语言内存的问题, cgo 针对该场景定义了专门的规则: **在CGO调用的C语言函数返回前, 
cgo保证传入的 Go 语言内存在此期间不会发生移动, C 语言函数可以大胆地使用 Go 语言的内存!**

```cgo
/*
#include <stdio.h>

void printString(const char* s, int n) {
    printf("len is:%d, data: %s \n", n, s);
}
*/
import "C"

func printString(s string) {
   C.printString(C.CString(s), C.int(len(s)))
}
```

任何完美的技术都有滥用的时候, CGO的这种看似完美的规则也是存在隐患的. 假设调用的 C 语言函数需要长时间运行, 那么将会导致
它引用的 Go 语言内存在 C 语言返回前不能被移动, 从而可能间接地导致这个 Go 内存栈对应的 goroutine 不能动态伸缩栈内存, 
也就是可能导致整个 goroutine 被阻塞. 因此, 在需要长时间运行的 C 语言函数 (特别是在纯CPU运算之外, 还可能因为需要等待
其他的资源而需要不确定时间才能完成的函数), 需要谨慎处理传入的 Go 语言内存.

> 需要小心的是在取得 Go 内存后需要马上传入 C 语言函数, 不能保存到临时变量后再间接传入 C 语言函数. 因为CGO 只能保证在
> C 函数调用之后被传入的 Go 语言内存不会发生移动, 它并不能保证在传入 C 函数之前内存不发生变化.


错误的案例:

```
tmp := uintptr(unsafe.Pointer(&x)) // 临时获取 x 的指针
pb := (*int16)(unsafe.Pointer(tmp)) // 转换为其他类型指针
*pb = 42
``` 

由于 tmp 并不是指针类型, 在它获取到 Go 对象地址之后, `x 对象可能会被移动` (垃圾回收), 但是因为不是指针类型, 所有不会
被 Go 语言运行时更新成新内存的地址. 在非指针类型的 tmp 保持 Go 对象的地址, 和在 C 语言环境保持 Go 对象的地址的效果是
一样的: 如果原始的 Go 对象内存发生了移动, Go 语言运行时并不会同步更新它们. 


### C长时间持有 Go 指针对象

当 C 语言函数调用 Go 语言函数的时候, C 语言函数就成了函数的调用方, Go 语言函数返回的 Go 对象的生命周期超出了 Go 语言
运行时的管理, 简言之, 不能在 C 语言函数中直接使用 Go 语言对象的内存.


### 导出 C 函数不能返回 Go 内存


## C++ 类包装

CGO 是 C 语言和 Go语言之间的桥梁, 原则上无法支持 C++ 的类. CGO 不支持 C++ 语法的根本原因是 C++ 至今为止还没有一个
二进制接口规范(ABI).  一个 C++ 类的构造函数在编译为目标文件时如何生成链接符号,方法在不同的平台甚至是C++的不同版本之间
都是不一样的.

但是 C++ 是兼容 C 语言, 所以可以通过增加一组 C 语言函数接口作为 C++ 类和 CGO 之间的桥梁, 这样可以间接地实现 C++ 和 
Go 之间的互联. 因为 CGO 只支持 C 语言中值类型的数据类型, 所以无法直接使用 C++ 的引用参数等特性的.

### C++ 类到 GO 语言对象

实现 C++ 类到 Go 语对象的包装需要以下几个步骤: 首先是用纯 C 函数接口包装该 C++ 类; 其次是通过 CGO 将纯 C 函数接口映
射到 Go 函数; 最后是做一个 Go 包装对象, 将 C++ 类到方法用 Go 对象的方法实现.

> 需要注意的点:
>
> 1.必须采用了静态库/动态库链接的方式去编译链接.(使用代码的方式编译不了)
>
> 2.在使用 gcc 编译包装了 C++ 库的时候, 一定要链接 `stdc++` 库 (即一定要有 `LDFLAGS` 参数, 且必须包含 
> `-l stdc++` 选项).

下面是一个详细的案例:

C++ 代码:

> buffer.h, buffer.cpp 是使用 c++ 实现的一个简单缓存类

// buffer.h
```cgo
#include <string>

#ifdef __cplusplus

class Buffer {
    private:
        std::string* s_;

    public:
        Buffer(int size);
        ~Buffer(){}

        int Size();
        char* Data();
};

#endif
```

// buffer.cpp
```cgo
#include "buffer.h"

Buffer::Buffer(int size) {
    this->s_ = new std::string(size, char('\0'));
}

int Buffer::Size() {
    return this->s_->size();
}

char* Buffer::Data() {
    return (char*)this->s_->data();
}
```


> **bridge.h 和 bridge.cpp 是使用 C 代码包装 C++ 的类 Buffer, 这个也称为桥接**
>
> **1. 需要深刻理解 `extern "C"` 的含义, 这是包装成 C++ 成 C 的关键环节.**
>
> **2. 在 bridge.h `typedef struct Buffer_T Buffer_T` 只是预定义结构体(有点类似 `void*` 的感觉), 真正的 
> `Buffer_T` 是在 bridge.cpp 实现的时候具体化, 而且这里采用了继承, 也可以使用组合的方式. bridge.cpp 当中可以使用
> C++ 的方式用.**

// bridge.h
```cgo
#ifdef __cplusplus
extern "C" {
#endif

typedef struct Buffer_T Buffer_T;

extern Buffer_T* NewBuffer(int size);
extern void DeleteBuffer(Buffer_T* p);

extern char* Buffer_Data(Buffer_T* p);
extern int Buffer_Size(Buffer_T* p);

#ifdef __cplusplus
}
#endif
```

// bridge.cpp
```cgo
#include "buffer.h"
#include "bridge.h"

#ifdef __cplusplus
extern "C" {
#endif

// 注意这里的包装继承机制
struct Buffer_T: Buffer {
    Buffer_T(int size): Buffer(size) {}
    ~Buffer_T() {}
};

Buffer_T* NewBuffer(int size) {
    Buffer_T* p = new Buffer_T(size);
    return p;
}

void DeleteBuffer(Buffer_T* p) {
    delete p;
}

char* Buffer_Data(Buffer_T* p) {
    return p->Data();
}

int Buffer_Size(Buffer_T* p) {
    return p->Size();
}

#ifdef __cplusplus
extern "C" {
#endif
```

> **注意: 在链接的时候一定要链接库stdc++(因为使用了C调用C++), buffer(Buffer库)**

// main.go, 最终的 go 测试函数
```cgo
package main

/*
#cgo CXXFLAGS: -std=c++11 -I .
#cgo LDFLAGS: -L . -lbuffer -lstdc++

#include <stdio.h>
#include "bridge.h"
*/
import "C"
import "unsafe"

/**
这里采用了静态库链接的方式, 链接 libbuffer.a 文件

在使用 gcc 编译包装了 C++ 库的时候, 一定要链接 "stdc++" 库, 这个非常重要.
**/

type Buffer struct {
    ptr *C.Buffer_T
}

func NewBuffer(size int) *Buffer {
    p := C.NewBuffer(C.int(size))

    return &Buffer{
        ptr: (*C.Buffer_T)(p),
    }
}

func (p *Buffer) Delete() {
    C.DeleteBuffer(p.ptr)
}

func (p *Buffer) Data() []byte {
    data := C.Buffer_Data(p.ptr)
    size := C.Buffer_Size(p.ptr)
    return ((*[1 << 31]byte)(unsafe.Pointer(data)))[0:int(size):int(size)]
}

func main() {
    buf := NewBuffer(1024)
    defer buf.Delete()

    copy(buf.Data(), []byte("Hello World. \x00"))
    C.puts((*C.char)(unsafe.Pointer(&(buf.Data()[0]))))
}
```

// test.c, C 测试文件
```cgo
#include "bridge.h"
#include <stdio.h>

int main() {
    void* buf = NewBuffer(10);

    char* data = Buffer_Data(buf);
    int size = Buffer_Size(buf);
    printf("data: %s, size: %d \n", data, size);
    DeleteBuffer(buf);
}
```

// makefile
```
GO = $(wildcard *.go)
CPLUS = $(wildcard *.cpp)
C = $(wildcard *.c)

gotest: static
    go build -o gotest -ldflags "-w" -x $(GO)

ctest: static
    $(CC) -o ctest $(C) -static -L. -lbuffer -lstdc++

static: clean
    $(CXX) -c -std=c++11 $(CPLUS)
    $(AR) -r libbuffer.a *.o
    @rm -rf *.o

.PHONY : clean
clean:
    @rm -rf ctest gotest *.o *.a *.gch
```