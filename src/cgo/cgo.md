# golang cgo

[相关文章](https://colobu.com/2018/06/13/cgo-articles/)

[文档](https://chai2010.cn/advanced-go-programming-book/ch2-cgo/ch2-02-basic.html)

## cgo语句

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
| short | C.short | int16 |
| int | C.int | int32 |
| long | C.long | int32 |
| long long int | C.longlong | int64 |
| float | C.float | float32 |
| double | C.double | float64 |
| size_t | C.size_t | uint | 


### 结构体, 联合, 枚举类型

#### 结构体

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

#### 联合类型

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

在 C 语言当中, 数组名其实对应着一个指针, 指向特定类型特定长度的一段内存, 但是这个指针不能被修改.

当把数组名传递给一个函数时, 实际上传递的是数组第一个元素的地址.

C 语言的字符串是一个 char 类型的数组, 字符串的长度需要根据表示结尾的NULL字符的位置确定.

---

Go 当中, 数组是一种值类型, 而且数组的长度是数组类型的一部分.

Go 当中, 字符串对应一个长度确定的只读 byte 类型的内存.

Go 当中, 切片是一个简化版的动态数组.

---

Go 语言 和 C 语言的数组, 字符串和切片之间的相互转换可以简化为 `Go 语言的切片和 C 语言中指向一定长度内存的指针` 
之间的转换.


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

> C.CString 针对输入 Go 字符串, 克隆一个 C 语言格式的字符串; 返回的字符串由 C 语言的 malloc 函数分配,
> 不使用时需要通过 C 语言的 free 函数释放.
>
> C.Cbytes 函数和 C.CString 类似, 针对输入的 Go 语言切片克一个 C 语言版本的字符数组.
>
> C.GoString 用于将从 NULL 结尾的 C 语言字符串克隆一个 Go语言字符串.
>
> C.GoStringN 是另一个字符数组克隆函数.
> 
> C.GoBytes 用于从 C 语言数组, 克隆一个 Go 语言字节切片.

当 C 语言字符串或数组向 Go 语言转换时, 克隆的内存由 Go 语言分配管理. 通过该组转换函数, 转换前和转换后的内存
依然在各自的语言环境中, 它们并没有跨域 Go 语言和 C 语言.

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

如果不希望单独分配内存, 可以在Go当中直接访问C的内存空间:

```cgo
/**
#include <string.h>
char arr[10];
cahr *s = "Hello";
**/
import "C"

func main() {
    // 通过 reflect.SliceHeader 转换
    var arr0 []byte
    var arr0Hdr = (*reflect.SliceHeader)(unsafe.Pointer(&arr0))
    arr0Hdr.Data = uintptr(unsafe.Pointer(&C.arr[0])
    arr0Hdr.Len = 10
    arr0Hdr.Cap = 10
    
    // 通过切片语法转换
    arr1 := (*[31]byte)(unsafe.Pointer(&C.arr[0]))[:10:10]
    
    var s0 string
    var s0Hdr = (*reflect.SliceHeader)(unsafe.Pointer(&s0))
    s0Hdr.Data = uintptr(unsafe.Pointer(C.s))
    s0Hdr.Len = int(C.strlen(C.s))
    
    sLen := int(C.strlen(C.s))
    s1 := string(*[31]byte)(unsafe.Pointer(C.s))[:SLen:SLen]
}
```

> Go 字符串是只读的, 用户需要自己保证 Go 字符串在使用期间, 底层对应的 C 字符串内容不会发生变化, 内存不会被
> 提前释放掉.


### 指针间的转换

在 C 语言中, 不同类型的指针是可以显式或隐式转换的, 如果是隐式只是会在编译时给出一些警告信息.

Go 语言对于不同类型的转换非常严格, 任何 C 语言中可能出现的警告信息在 Go 语言中都可能是错误!

指针是 C 语言的灵魂, 指针间的自由转换也是 cgo 代码中经常要解决的第一个问题.

---

在 Go 语言中两个指针的类型完全一致则不需要转换可以直接使用. 如果一个指针类型是用 type 命令在另
一个指针类型基础上构建的, 换言之 `两个指针是 "底层结构完全相同" 的指针`, 那么可以通过直接强制转
换语法进行指针间的转换.  但是 cgo 经常要面对是2个完全不同类型的指针间的转换, 原则上这种操作在纯
Go 语言代码是严格禁止的.

```
var p *X
var q *Y

q = (*X)(unsafe.Pointer(p)) // *X => *Y
p = (*Y)(unsafe.Pointer(q)) // *Y => *X
```

为了实现 X 类型和 Y 类型的指针的转换, 需要借助 `unsafe.Pointer` 作为中间桥接类型实现不同类型
指针之间的转换. `unsafe.Pointer` 指针类型类似 C 语言中的 `void*` 类型的指针.

指针简单转换流程图:

[!image](../../images/xtoy.png)


### 数值和指针的转换

为了严格控制指针的使用, Go 语言禁止将数值类型直接转为指针类型! 不过, Go 语言针对 `unsafe.Pointer`
指针类型特别定义了一个 `unitptr` 类型. 可以以 `unitptr` 为中介, 实现数值类型到 `unsafe.Pointer`
指针类型的转换. 再结合前面提到的方法, 就可以实现数值类型和指针的转换了.

int32 类型到 C 语言的 `char*` 字符串指针类型的相互转换:

[!image](../../images/numtoptr.png)

### 切片间的转换

在 C 语言当中数组也是一种指针, 因此两个不同类型数组之间的转换和指针类型间转换基本类似.

在 Go 语言当中, 数组或数组对应的切片不再是指针类型, 因此无法直接实现不同类型的切片之间的转换.

在 Go 的 `reflect`包提供了切片类型的底层结构, 再结合前面不同类型直接的指针转换, 可以实现 []X
到 []Y 类型的切片转换.

```
var p []X
var q []Y

pHdr := (*reflect.SliceHeader)(unsafe.Pointer(&p))
qHdr := (*reflect.SliceHeader)(unsafe.Pointer(&q))

qHdr.Data = pHdr.Data
qHdr.Len = pHdr.Len * int(unsafe.Sizeof(p[0])) / int(unsafe.Sizeof(q[0]))
qHdr.Cap = pHdr.Len * int(unsafe.Sizeof(p[0])) / int(unsafe.Sizeof(q[0]))
```


## 内部机制

### CGO 生成的中间文件

在构建一个 cgo 包时增加一个 `-work` 输出中间生成所在目录并且在构建完成时保留中间文件.

对于比较简单的 cgo 代码可以直接手工调用 `go tool cgo` 命令来查看生成的中间文件.

在一个 Go 源文件当中, 如果出现 `import "C"` 指令则表示将调用 cgo 命令生成对应的中间文件. 下面是生成的中间
文件的简单示意图:

[!image](../../images/cgo_mid.png)


包含有 4 个 Go 文件, 其中 nocgo 开头的文件中没有 `import "C"` 指令, 其他的 2 个文件则包含了 cgo 代码. 
cgo 命令会为每个包含 cgo 代码的 Go 文件创建 2 个中间文件, 比如 main.go 会分别创建 `main.cgo1.go` 和 
`main.cgo2.c` 两个中间文件, cgo当中是 Go 代码部分和 C 代码部分.

然后会为整个包创建一个 `_cgo_gotypes.go` Go 文件, 其中包含 Go 语言部分辅助代码. 此外还创建一个 
`_cgo_export.h` 和 `_cgo_export.c` 文件, 对应 Go 语言导出到 C 语言的类型和函数.

```cgo
/*
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

其中 `main.cgo1.go [1]` 的代码如下:

```
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

其中 C.sum(1,1) 函数调用替换成了 (_Cfunc_sum)(1,1). 每一个 `C.xxx` 形式的函数都会被替换成 `_Cfunc_xxx`
格式的纯 Go 函数, 其中前缀 `_Cfunc_` 表示这是一个C函数, 对应一个私有的 Go 桥接函数.

`_Cfunc_sum` 函数在 cgo 生成的 `_cgo_gotypes.go [2]` 文件当中定义:

```
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

`_Cfunc_sum` 函数的参数和返回值 `_Ctype_int` 类型对应 `C.int` 类型, 命名的规则和 `_C_func_xxx`
类似, 不同的前缀用于区分函数和类型.

`_cgo_runtime_cgocall` 对应 `runtime.cgocall` 函数, 声明如下:

```
func runtime.cgocall(fn, arg unsafe.Pointer) int32
```

> 第一个参数是 C 语言函数的地址
> 第二个参数是存储 C 语言函数对应的参数结构体(参数和返回值)的地址

在此例当中, 被传入C语言函数 `_cgo_7b5139e7c7da_Cfunc_sum` 也是 cgo 生成的中间函数. 函数定义在
`main.cgo2.c [3]` 当中.

```
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

```
struct {
		int p0;
		int p1;
		int r;
		char __pad12[4];
	} __attribute__((__packed__, __gcc_struct__)) *_cgo_a = v;
```

其中, p0, p1分别对应 sum 的第一个和第二个参数, r 对应 sum 的返回值. `_pad12` 用于填充结构体保证对齐CPU机器字
的整数倍.

> 然后从参数执行的结构体获取参数后开始调用真实的C语言版sum函数, 并且将返回值保存到结构体的返回值对应的成员.


因为 Go 语言和 C 语言有着不同的内存模型和函数调用规范, 其中 `_cgo_topofstack` 函数相关的代码用于 C 函数调用后
恢复调用栈. `_cgo_tsan_acquire` 和 `_cgo_tsan_release` 则是用于扫描 CGO 相关函数的指针总相关检查.

调用链:

```
main.go     main.cgo1.go    _cgo_types.go       runtime.cgocall     main.cgo2.c
C.sum(1,1)
            _Cfunc_sum(1,1)
                            _cgo_runtime_cgocall(...)
                                                _cgo_runtime_cgocall(...)   
                                                                    _cgo_xxx_Cfunc_sum(void*)
                                                _cgo_runtime_cgocall(...)
2
``` 

文件链:

```
main.go -> main.cgo1.go -> _cgo_gotypes.go -> main.cgo2.c
```


## CGO 内存模型

CGO 是架接 Go 语言和C语言的桥梁, 它使二者在二进制借口层面实现了互通, 但是要注意两种语言的内存模型
的差异而可能引起的问题.

如果在 CGO 
