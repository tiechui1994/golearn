## golang cgo

### cgo语句

在 `import "C"`语句前的注释可以通过 `#cgo` 语句设置 `编译阶段` 和 `链接阶段` 的相关参数.

编译阶段的参数主要用于 `定义相关的宏` 和 `指定头文件检索路径`.

链接阶段的参数主要是 `指定库文件检索路径` 和 `要链接的库文件`.

```cgo
// #cgo CFLAGS: -D PNG_DEBUG=1 -I ./include
// #cgo LDFLAGS: -L /usr/local/lib -l png
// #include <png.h>
import "C"
```

上面的代码中:
 
`CFLAGS` 部分, `-D` 部分定义了 `PNG_DEBUG`, 值为 `1`; `-I` 定义了头文件包含的检索目录. 

`LDFLAGS` 部分, `-L` 指定了链接时文件检索目录, `-l` 指定了链接时需要链接 `png` 库.  

因为 C/C++ 遗留的问题, C头文件检索目录可以是相对目录, 但是 `库文件检索目录` 则需要是绝对路径.
在库文件的检索目录中可以通过 `${SRCDIR}` 变量表示当前包含目录的绝对路径:

```
// #cgo LDFLAGS: -L ${SRCDIR}/libs -l foo
```

上面的代码在链接时将被展开为:

```
// #cgo LDFLAGS: -L /go/src/foo/libs -l foo
```

`#cgo` 语句主要影响 `CFLAGS`, `CPPFLAGS`, `CXXFLAGS`, `FFLAGS`, `LDFLAGS` 几个编译器环境编辑. 

`LDFLAGS` 用于设置链接时的参数, 除此之外的几个变量用于改变编译阶段的构建参数(CFLAGS用于针对C语言代码设置编译参数).

对于在cgo环境混合使用C和C++的用户来说, 可能有三种不同的编译选项: 其中 CFLAGS 对应C语言特有的编译选项, 
CXXFLAGS 对应是 C++ 特有的编译选项, CPPFLAGS 则对应 C 和 C++ 共有的编译选项.

但是在链接阶段, C和C++的链接选项是通用的, 因此这个时候已经不再有C和C++语言的区别, 它们的目标文件的类型是相同的.

`#cgo` 指令还支持条件选择, 当满足某个操作系统或某个CPU架构类型时类型时后面的编译或链接选项生效. 

```cgo
// #cgo windows CFLAGS: -D X86=1
// #cgo !windows LDFLAGS: -l math
```


```cgo
package main

/*
#cgo windows CFLAGS: -D CGO_OS_WINDOWS=1
#cgo darwin  CFLAGS: -D CGO_OS_DRWIN=1
#cgo linux CFLAGS: -D CGO_OS_LINUX=1

#if defined(CGO_OS_WINDOWS)
#   const char* os = "windows";
#elif defined(CGO_OS_DARWIN)
#   const char* os = "darwin";
#elif defined(CGO_OS_LINUX)
#   const char* os = "linux";
#else
#   error(unknown os)
#endif
*/
import "C"

func main(){
    print(C.GoString(C.os))
}
```


## 常用的cgo类型

### 数值类型

| C | CGO | Go |
| -- | -- | -- |
| char | C.char | byte | 
| singed char | C.schar | int8 |
| unsigned char | C.uchar | uint8 |
|  short | C.short | int16 |
| int | C.int | int32 |
| long | C.long | int32 |
| long long int | C.longlong | int64 |
| float | C.float | float32 |
| double | C.double | float64 |
| size_t | C.size_t | uint | 

### 结构体, 联合, 枚举类型

- 结构体

在 Go 当中, 可以通过 `C.struct_xxx` 来访问C语言中定义的 `struct xxx` 结构体类型.

结构体的内存按照C语言的通用对齐规则, 在 32 位Go语言环境 C 语言结构也按照 32 位对齐规则,
在 64 位Go语言环境按照 64 位对齐规则. 对于指定了特殊对齐规则的结构体, 无法在 CGO 中访问.

```cgo
/*
struct A {
    int i;
    float f;
};
*/
import "C"

func main() {
    var a C.struct_A
    fmt.Println(a.i)
}
```

结构体当中出现了 Go 语言的关键字, 通过下划线的方式进行访问.

```cgo
/*
struct A {
    int type;
    char chan;
};
*/
import "C"

func main() {
    var a C.struct_A
    fmt.Println(a._type)
    fmt.Println(a._chan)
}
```

> 如果有两个成员, 一个是以 Go 语言关键字命名, 另外一个刚好是以下划线和Go语言关键字命名, 那么以 Go 语言关键字命名的
成员将无法访问(被屏蔽)


C 语言结构体中 `位字段` 对应的成员无法在 Go 语言当中访问. 如果需要操作 `位字段` 成员, 需要通过在 C 语言当中定义辅
助函数来完成.  对应 `零长数组` 的成员, 无法在 Go 语言中直接访问数组的元素, 但其中 `零长数组` 的成员所在的位置偏移
量依然可以通过 `unsafe.Offset(a.arr)` 来访问.

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

> 在 C 语言当中, 无法直接访问 Go 语言定义的结构体类型

- 联合类型

对于联合类型,可以通过 `C.union_xxx` 来访问 C 语言中定义的 `union xxx` 类型. **但是 Go 语言中并不支持C语言联合类型,
它们会被转换为对应大小的字节数组.**

```cgo
/*
#include <stdint.h>

union B {
    int i;
    float f;
};

union C {
    int8_t i8;
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

> 虽然unsafe包访问最简单, 性能也最好, 但是对于有**嵌套联合类型**的情况处理会导致问题复杂化. 对于复杂的联合类型,
推荐通过在C语言中定义辅助函数的方式处理.


- 枚举类型

对于枚举类型, 通过 `C.enum_xxx` 访问 C 语言当中定义的 `enum xxx` 结构体类型


