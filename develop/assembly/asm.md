# go 汇编

## 变量定义(热身)

### 定义整数变量

先看一个简单的例子:

```cgo
package pkg

var ID = 9527
```

> 代码只定义了一个 int 类型的包级别变量, 并进行初始化. 使用 `go tool compile -S pkg.go` 生成伪汇编代码:

```
go.cuinfo.packagename. SDWARFINFO dupok size=0
        0x0000 70 6b 67                                         pkg
"".ID SNOPTRDATA size=8
        0x0000 37 25 00 00 00 00 00 00                          7%......
```

> "".ID 对应 ID 变量符号, 变量的内存大小是 8 个字节. 变量初始化内容为`37 25 00 00 00 00 00 00`, 对应十六进制格式
为0x2537, 十进制为 9527. SNOPTRDATA 是相关的标记. 其中 `NOPTR` 表示数据中不包含指针数据.

Go 汇编语言提供了 DATA 命令用于初始化包变量, DATA 命令语法:

```
DATA symbol+offset(SB)/width, value
```

其中 symbol 为变量在汇编语言中对应的标识符, offset 是符号开始地址的偏移量, width 是初始化内存的宽度大小, value 是要
初始化的值. 其中当前包中 Go 语言定义的符号 symbol, 在汇编中对应 `·symbol`, 其中 `·` 中点符号为一个特殊的 unicode 
符号.

采用以下命令可以给 ID 变量初始化为十六进制的 0x2537 (常量需要以 `$` 开头表示):

```
DATA ·ID+0(SB)/1,$0x37
DATA ·ID+1(SB)/1,$0x25
```

变量定义好了之后需要导出已供其他代码引用. Go 汇编语言提供了 GLOBL 命令用于符号导出:

```
GLOBL symbol(SB), width
```

> 其中 symbol 对应汇编中符号的名字, width 为符号对应内存的大小.

```
GLOBL ·ID, $8
```

现在, 已经初步完成了用汇编定义整数变量的工作.


现在将上述的工作完整的展示一遍:

// pkg.go, 声明变量 ID
```cgo
package pkg

var ID int
``` 

// pkg_amd64.s, 使用汇编代码初始化变量 ID

```
GLOBL ·ID(SB), $8

DATA ·ID+0(SB)/1,$0x37
DATA ·ID+1(SB)/1,$0x25
DATA ·ID+2(SB)/1,$0x00
DATA ·ID+3(SB)/1,$0x00
DATA ·ID+4(SB)/1,$0x00
DATA ·ID+5(SB)/1,$0x00
DATA ·ID+6(SB)/1,$0x00
DATA ·ID+7(SB)/1,$0x00
```

文件名 `pkg_amd64.s` 的后缀名表示 AMD64 环境下的汇编代码文件.

// main.go, 测试代码
```cgo
package main

import "golearn/develop/assembly/pkg"

func main() {
	println(pkg.ID)
}
```

> 对于 Go 包的用户来说, 用 Go 汇编语言或 Go 语言实现并无任何区别.

### 定义字符串变量

虽然从 Go 语言角度来看, 定义字符串和整数变量的写法基本相同, 但是字符串底层却有着比单个整数更复杂的数据结构.

先看一个简单的例子:

```cgo
package pkg

var Name = "gopher"
```

生成的伪汇编代码如下:

```
go.cuinfo.packagename. SDWARFINFO dupok size=0
        0x0000 70 6b 67                                         pkg
go.string."gopher" SRODATA dupok size=6
        0x0000 67 6f 70 68 65 72                                gopher
"".Name SDATA size=16
        0x0000 00 00 00 00 00 00 00 00 06 00 00 00 00 00 00 00  ................
        rel 0+8 t=1 go.string."gopher"+0
```

> 符号 `go.string."gopher"` 它表示字符串 "gopher", 该符号有一个 SRODATA 标志表示这个数据在只读内存段, dupok 表
> 示出现多个相同标识符的数据时只保留一个就可以了.
>
> 真正的 Go 字符串变量 Name对于的大小却只有16个字节了. 其实 Name 变量并没有直接对于 "gopher" 字符串, 而是对于 16 字
> 节大小的 reflect.StringHeader 结构体.

```
type reflect.StringHeader struct {
	Data uintptr
	Len int
}
```

从汇编角度看, Name 变量对应的是 `reflect.StringHeader` 结构体类型. 前 8 字节对于底层真实字符串数据的指针, 也就是
符号 `go.string."gopher"` 对应的地址. 后 8 字节对应底层真实字符串数据的有效长度(这里是6).

与上面整数的案例一样, 通过汇编重新定义初始化 Name 字符串:

// pkg_amd64.s
```
GLOBL ·NameData(SB),$8
DATA  ·NameData(SB)/8,$"gopher"

GLOBL ·Name(SB),$16
DATA  ·Name+0(SB)/8, $·NameData(SB)
DATA  ·Name+8(SB)/8, $6
```

> 在 Go 汇编语言当中, go.string."gopher" 不是一个合法的符号, 因此我们无法手工创建. 因此我们新创建一个 `·NameData`
> 符号表示底层的字符串数据. 然后定义 `·Name` 符号, 内存大小为 16 字节, 其中钱8个字节使用 `·NameData` 符号对于的地址
> 初始化, 后 8 个字节为常量 6, 表示字符串长度.

测试代码:

```cgo
package main

import "golearn/develop/assembly/pkg"

func main() {
	println(pkg.Name)
}
```

> 运行出错: `pkg.NameData: missing Go type information for global symbol: size 8`.
>
> 错误提示汇编中定义的 NameData 符号没有类型信息. 其实 Go 汇编语言中定义的数据并没有所谓的类型, 每个符号只不过是对应一
> 块内存而已, 因此 NameData 符号也是没有类型的. 但是 Go 语言是带垃圾回收器的语言, 而 Go 汇编语言是工作在自动垃圾回收
> 体系框架内的. 当 Go 语言的垃圾回收器在扫描到 NameData 变量的时候, 无法知晓该变量内部是否包含指针, 因此就出现了这种
> 错误. **错误的根本原因并不是 NameData 没有类型, 而是 NameData 变量没有标注是否含义指针信息.**

修复方式:

方式一: 通过给 `NameData` 变量增加一个 NOPTR 标志, 表示其中不会包含指针数据可以修复该错误.

```
#include "textflag.h"

GLOBL ·NameData(SB), NOPTR, $8
DATA  ·NameData(SB)/8,$"gopher"

GLOBL ·Name(SB),$16
DATA  ·Name+0(SB)/8,$·NameData(SB)
DATA  ·Name+8(SB)/8,$6
```

方式二: 通过给 `NameData` 变量在 Go 语言中增加一个不包含指针并且大小为8个字节的类型来修复该错误.

```cgo
package pkg

var NameData [8]byte
var Name string
```

> 将 NameData 声明为长度为 8 的字节数组. 编译器可以通过类型分析处该变量不包含指针, 因此汇编代码中可以省略 NOPTR 标志.

### 定义 main 函数

前面的例子主要展示汇编定义整数和字符串变量. 现在尝试汇编实现一个函数.

// main.go

```cgo
package main

var hello = "hello world"

//go:noescape
func main()
``` 

```
TEXT ·main(SB), $16-0
    MOVQ ·hello+0(SB), AX
    MOVQ AX, 0(SP)
    MOVQ ·hello+8(SB), BX
    MOVQ BX, 8(SP)
    CALL runtime·printstring(SB)
    CALL runtime·printnl(SB)
    RET
```

> `TEXT ·main(SB), $16-0` 用于定义 main 函数, 其中 $16-0 表示 main 函数的帧大小是 16 个字节(对于string头部结
> 构体的大小, 用于给 runtime·printstring 函数传递参数), 0 表示 main 函数没有参数和返回值. main 函数内部通过调用
> 运行时内部 runtime·printstring(SB) 函数来打印字符串. 然后调用 runtime·printnl 打印换行符.

### 特殊字符串

Go 语言函数和方法符号在编译为目标文件后, 目标文件中的每个符号均包含对于包的绝对导入路径. 因此目标文件的符号可能非常复杂.
例如 `"/path/to/pkg.(*SomeType).SomeMethod` 或 `go.string."abc"` 等名字. 目标文件的符号名不仅仅包含普通的字符,
还可能包含点号, 星号, 小括号, 双引号等诸多特殊字符.

## 全局变量

在 Go 中, 变量根据作用域和生命周期有全局变量和局部变量之分. 全局变量是包一级的变量, 全局变量一般有着较为固定的内存地址, 
声明周期跨域整个程序运行时间. 局部变量一般是函数内定义的变量, 只有在函数被执行的时间才在栈上创建, 当函数调用完成后将回收(
暂时不考虑闭包对局部变量捕获的问题)

在 Go 汇编中, 内存是通过 SB 伪寄存器定位. SB (Static base pointer), 意为静态内存的开始地址. 可以将 SB 想象为一个
和内存容量相同大小的字节数组, 所有的静态全局符号通常可以通过SB加上一个偏移量定位, 而我们定义的符号其实就是相对于SB内存开
始地址偏移量. 对于SB伪寄存器, 全局变量和全局函数的符号并没有任何区别.

**要定义全局变量, 首先要声明一个变量对应的符号, 以及变量对应的内存大小.**

```
GLOBL symbol(SB), width
```

GLOBL 指令用于定义名为 symbol 的变量, 变量对应的内存宽度为 width (必须是2的指数倍), 内存宽度部分必须使用常量初始化.

> 注意: 在 Go 汇编中我们无法为变量指定具体的类型. 在汇编中定义全局变量时, 只关心变量的名字和内存大小, 变量最终的类型只
> 能在Go语言中声明.

**变量定义之后, 可以通过 `DATA` 汇编指令指定对应内存中的数据.**

```
DATA symbol+offset(SB)/width, value
```

含义是从 symbol+offset 偏移量开始, width宽度的内存, 用 value 变量对应的值初始化. DATA 初始化内存时, width 必须是
1, 2, 4, 8 几个宽度之一, 因为再大的内存无法一次性无法用一个 uint64 大小的值表示.

### 数组类型

汇编中数组是非常简单的类型.

先声明一个 `[2]int` 类型的数组变量 num:

```
var num [2]int
```

然后在汇编中定义一个对应16字节大小的变量, 并使用零值初始化:

```
GLOBL ·num(SB), $16
DATA  ·num+0/8(SB), $0
DATA  ·num+1/8(SB), $0
```

### string 类型变量

Go 汇编角度, 字符串只是一种结构体. string头的定义:

```
type StringHeader struct{
    Data uintptr
    Len  int
}
```

在 AMD64 环境中 StringHeader 有16个字节大小. 

先声明一个Go字符串变量.

```
var hello string
```

然后在汇编中定义一个对应16字节大小的变量:

```
GLOBL ·hello(SB), $16
```

同时可以为字符串准备真正的数据. 下面的汇编代码定义一个text当前文件内的私有变量(以 `<>` 为后缀名), 内容为 "hello world!"

```
#include "textflag.h"

GLOBL text<>(SB), NOPTR, $16
DATA text<>+0(SB)/8, $"hello wo"
DATA text<>+8(SB)/8, $"rld!"
```

虽然 `text<>` 私有变量表示的字符串只有12个字符长度, 这也是一个字符串, 长度就是16.

接下来然后使用 text 私有变量对应的内存地址对应的常量来初始化字符串头结构体中的Data部分.

```cgo
DATA ·hello+0(SB)/8, $text<>(SB)
DATA ·hello+8(SB)/8, $12
```

完整的内容是:

```
#include "textflag.h"

GLOBL text<>(SB), NOPTR, $16
DATA text<>+0(SB)/8, $"hello wo"
DATA text<>+8(SB)/8, $"rld!"

GLOBL ·hello(SB), $16
DATA ·hello+0(SB)/8, $text<>(SB)
DATA ·hello+8(SB)/8, $12
```

> 注意: 下面的赋值是非法的.

```
GLOBL ·hello(SB), $16
DATA ·hello+0(SB)/8, $"1234567"
DATA ·hello+8(SB)/8, $12
```

> 运行的时候, 会报 `unexpected fault address` 错误


### slice 类型变量

slice 变量和 string变量类似, 只不过对应的切片头结构体而已. 切片头的结构如下:

```
type SliceHeader struct {
    Data uintptr
    Len  int
    Cap  int
}
```

例子:

```cgo
GLOBL ·hello(SB), $24            // var hello []byte("Hello World!")
DATA ·hello+0(SB)/8,$text<>(SB) // SliceHeader.Data
DATA ·hello+8(SB)/8,$12         // SliceHeader.Len
DATA ·hello+16(SB)/8,$16        // SliceHeader.Cap

GLOBL text<>(SB),$16
DATA text<>+0(SB)/8,$"Hello Wo"      // ...string data...
DATA text<>+8(SB)/8,$"rld!"          // ...string data...
```

### map/channel类型变量

map/channel等类型并没有公开的内部结构, 它们只是一种未知类型的指针, 无法直接初始化. 在汇编中只能为类似变量定义并进行0值
初始化:

```cgo
var m map[string]int

var ch chan int
```

```cgo
GLBOL ·m(SB), $8 // var m map[string]int
DATA  ·m(SB)+0/8, $0

GLBOL ·ch(SB), $8 // var ch chan int
DATA  ·ch(SB)+0/8, $0
```

在 `runtime` 包当中为汇编提供了辅助函数. 比如可以通过 `runtime.makemap` 和 `runtime.makechan` 内部函数来创建 
map 和 chan 变量. 函数签名如下:

```
func makemap(mapType *byte, hint int, mapbuf *any) (hmap map[any]any)
func makechan(chanType *byte, size int) (hchan chan any)
```

## 函数

### 函数声明

函数标识符通过 TEXT 汇编指令定义, 表示该行开始的指令在TEXT内存段. TEXT 语句后的指令一般对应函数的实现, 但是对于TEXT
指令本身来说并不关系后面是否有指令. 因此 TEXT 和 LABEL 定义的符号是类型的, 区别只是 LABEL 是用于跳转标号, 但是本质
上它们都是通过标识符映射一个内存地址.

```
TEXT symbol(SB), [flags,] $framesize[-argsize]
```

函数定义由5个部分组成: TEXT指令, 函数名(symbol), 可选的flag标记, 函数帧大小和可选的函数参数大小.

其中, TEXT 用于定义函数符号, 函数名中当前包的路径可以忽略. 函数名字后面是 `(SB)`, 表示函数名符号相对于SB伪寄存器的偏
移量, 二者组合在一起最终是绝对地址. 作为全局的标识符的全局变量和全局函数的名字一般都是基于SB伪寄存器的相对地址.

标志部分, 用于指示函数的一些特殊行为, 标记在 `textflag.h` 文件中定义. 常见的 `NOSPLIT` 主要用于指示叶子函数不进行
栈分裂. `WRAPPER` 标志则表示这个是一个包装函数, 在panic或runtime.caller等某些处理函数帧的地方不会增加函数帧计数.
`NEEDCTXT` 表示需要一个上下文参数, 一般用于闭包函数.

framesize部分, 表示函数的局部变量需要多少栈空间, 其中包含调用其他函数时准备调用参数的隐式栈空间. 

最后是可以省略的参数大小, 之所以可以省略是因为编译器可以从Go语言函数声明中推导出函数参数的大小.

一个简单的函数 `Swap`.

```go
package main

//go:nosplit
func Swap(a, b int) (int, int)
```

下面是汇编定义 `Swap` (两种定义)

```
TEXT ·Swap(SB), NOSPLIT, $0-32

TEXT ·Swap(SB), NOSPLIT, $0
```

> 注意: 函数是没有类型, 上面定义的 Swap 函数签名可以说下面任意一种格式:
```
func Swap(a, b, c int) int
func Swap(a, b, c, d int)
func Swap()(a, b, c, d int)
func Swap()(a []int, d int)
...
```

### 函数参数和返回值

对于函数而言, 最重要的是函数对外提供的API约定, 包含函数的名称, 参数和返回值. 当这些都确定之后, 如何精确计算参数和返回值
大小是第一个需要解决的问题.

例如, Swap函数签名如下:

```
func Swap(a, b int)(ret0, ret1 int)
```

对于这个函数, 参数和返回值大小是32字节:

```
TEXT ·Swap(SB), $0-32
```

如何在汇编中引用这4个参数呢? 为此 Go 汇编中引入了一个 FP 伪寄存器, 表示函数当前帧的地址. 也就是第一个参数的地址. 因此
可以通过 `+0(FP)`, `+8(FP)`, `+16(FP)`, `+24(FP)` 来分别表示 a, b, ret0, ret1 四个参数.

但是, 在汇编代码当中, 我们并不能直接以 `+0(FP)` 的方式来使用参数. 为了编写易于维护的汇编代码, Go 汇编语言要求, 任何通
过 FP 伪寄存器访问的遍历必和一个临时标识符前缀组合后才能有效, **一般使用参数对应的变量名作为前缀**.

![image](/images/develop_assembly_swap_mem.png)


汇编函数实现 Swap 函数:

```
TEXT ·Swap(SB), $0-32
    MOVQ a+0(FP), AX      // AX=a
    MOVQ b+8(FP), BX      // BX=b
    MOVQ BX, ret0+16(FP)  // ret0=BX
    MOVQ AX, ret1+24(FP)  // ret1=AX
    RET
```

### 参数返回和返回值的内存布局

如果参数返回值类型比较负载的状况下该怎样处理呢? 例如有以下函数:

```
func Foo(a bool, b init16) (c []byte)
```

如何计算上述函数每个参数的位置和总的大小呢?

其实函数参数和返回值的大小以及对齐问题和结构体的大小和成员对齐问题是一致的, 函数的第一个参数和第一个返回值会分别进行一次内
存地址内存对齐. 可以使用代替思路将全部的参数和返回值以相同的顺序分别放到两个结构体中, 将FP伪寄存器作为唯一的一个指针参数,
而每个成员的地址也就是对应原来参数的地址.

使用上述的策略很容易计算 Foo 函数的参数和返回值的地址和总大小. 

```
type Foo_args struct {
    a bool
    b int16
    c []byte
}

type Foo_returns struct {
    c []byte
}
```

使用 `Foo_args` 和 `Foo_returns` 临时结构体类型用于替代原始的参数和返回值.

然后将 Foo 原来的参数替换为结构体形式, 并且只保留唯一的 FP 作为参数:

```
func Foo(FP *SomeFun_args, FP_ret *SomeFunc_returns) {
    // a = FP+offset(&args.a)
    _ = uintptr(FP) + unsafe.Offset(FP.a) // a
    // b = FP+offset(&args.b)
    _ = uintptr(FP) + unsafe.Offset(FP.b) // b
    
    
    // argsize = sizeof(args)
    argsize = unsafe.Offset(FP)
    
    // c = FP + argsize + offset(&return.c)
    _ = uintptr(FP) + argsize + unsafe.Offset(&FP_ret.c) // c
    
    // framesize = sizeof(args) + sizeof(returns)
    _ = unsafe.Offset(FP) + unsafe.Offset(FP_ret)
}
```

代码完全和Foo函数参数方式类型. 唯一的差异是每个函数的偏移量. 通过 `unsafe.Offset()` 函数自动计算生成. 因为 Go 结构体
中的每个成员已经满足了对齐的要求, 因此采用通用的方式得到每个参数的偏移量也是满足对齐要求的.

Foo 函数的参数和返回值的大小和内存布局:

![image](/images/develop_assembly_foo_mem.png)

Foo 汇编函数参数和返回值的定位:

```
TEXT ·Foo(SB), $0
    MOVQ a+0(FP), AX         // a
    MOVQ b+2(FP), BX         // b
    MOVQ c_dat+8*1(FP), CX   // c.Data
    MOVQ c_dat+8*2(FP), DX   // c.Len
    MOVQ c_dat+8*3(FP), DI   // c.Cap
    RET
```

其中 a 和 b 参数之间出现了 1 个字节空洞. b 和 c 出现了 4 个字节的空洞.

### 函数中的局部变量

从 Go 语言函数角度讲, 局部变量是函数内明确定义的变量, 同时包含函数的参数和返回值变量. 但是 **从 Go 汇编角度来看, 局部变
量是指函数运行时, 在当前函数栈所对应的内存内的变量, 不包括函数的参数和返回值(访问差异)**. 函数栈帧的空间主要由函数参数和
返回值, 局部变量和被调用其他函数的参数和返回值空间组成.

为了便于访问局部变量, Go汇编引入了 `伪SP寄存器`, 对应当前栈帧的底部(高地址). 因为当前栈帧的底部是固定不变的, 因此访问局
部变量的相对于伪SP寄存器也就是固定的(简化局部变量维护). SP真寄存器对应当前栈帧的顶部(低地址). 


SP真伪寄存器的区分标准是啥?, 如果使用 SP 时有一个临时标识符前缀就是伪SP寄存器, 否则就是SP真寄存器. 比如 `a(SP)`, 
`b+8(SP)` 这是伪SP寄存器. 比如 `0(SP)`, `8(SP)`, `+16(SP)` 这是真SP寄存器. 

**真SP寄存器一般用于表示调用其他函数时的参数和返回值**, 真SP寄存器对应内存低地址, 所以被访问的变量的偏移量是正数. 而伪
SP寄存器对应内存的高地址, 对应局部变量的偏移量都是负数.

```
func Foo() {
    var c []byte
    var b int16
    var a bool
}
```

汇编实现, 并通过伪SP寄存器来定位局部变量.

```
TEXT ·Foo(SB), $32
    MOVQ a-32(SP), AX // a
    MOVQ b-30(SP), BX // b
    MOVQ c_data-24(SP), CX // c.Data
    MOVQ c_len-16(SP), DX  // c.Len
    MOVQ c_cap-8(SP), DI   // c.Cap
    
    RET
```

### 控制流

对于顺序执行, 这个不强调了.
对于跳转(if/goto), 这个参考 goto_amd64.s
对于循环, 这个参考 loop_amd64.s


### 函数调用规范

Go汇编当中 CALL 指令用于调用函数. RET指令用于从调用函数返回. 但是 CALL 和 RET 指令并没有处理函数调用时输入参数和返回
值的问题. 

CALL 指令类似 `PUSH IP` 和　`JMP somefunc` 两个指令的组合. 首先, 将当前的IP指令寄存器的值压入栈中, 然后通过JMP指令
将调用函数的地址写入到 IP 寄存器实现跳转. 

RET指令则是和 CALL 相反的操作, 基本上和 `POP IP` 指令等价, 也就是将执行 CALL 指令时的保存在 SP 中的返回地址重新载入
IP寄存器, 实现函数的返回.

和 C 函数不同, Go 函数的参数和返回值完全通过栈传递. 下面是 Go 函数调用时栈的布局图:

![image](/images/develop_assembly_func_stack.png)

### 方法函数

Go 语言中方法函数和全局函数非常相似, 比如:

```
type Int int

func (i Int) Twice() int {
    return int(i)*2
}

func Int_Twice(i Int) int {
    return int(i)*2
}
```

其中 Int 类型的 Twice 方法和 Int_Twice 函数的类型是完全一样的, 只不过 Twice 的目标文件中被修饰为 `main.Int.Twice`
名称. 可以使用汇编实现该方法:

```
TEXT ·Int·Twice(SB), NOSPLIT, $0-16
    MOVQ a+0(FP), AX   // i
    ADDQ AX, AX
    MOVQ AX, ret+8(FP) // return 
    RET
```

不过这只是接收非指针的方法函数. 现在增加一个接收参数为指针类型的 `Ptr` 方法, 函数返回指针.

```
func (i *Int) Ptr() *Int {
    return i
}
```

在目标文件中, Ptr 方法名被修饰为 `main.(*Int).Ptr`, 也就是对应汇编中的 `·(*Int)·Ptr`. **不过 Go 汇编语言中, 星号,
小括号都无法用作函数名字, 也就是无法使用汇编直接实现接收参数为指针类型的方法.**

### 递归函数

参考代码 sum_amd64.s

### 闭包函数

闭包函数是最强大的函数, 因为闭包函数可以捕获外层局部作用域的局部变量, 因此闭包函数本身具有了状态.  先看一个例子:

```
func NewTwiceClosure(x int) func() int {
    return func() int {
        x *= 2
        return x
    }
}


func main() {
    fn := NewTwiceClosure(10)
    
    fmt.Println(fn()) // 20
    fmt.Println(fn()) // 40
    fmt.Println(fn()) // 80
}
```

其中 `NewTwiceClosure` 函数返回一个闭包函数对象, 返回的闭包函数对象捕获了外层的 x 参数. 返回的闭包函数对象在执行时, 
每次将捕获的外层变量乘以2之后再返回. 在 `main` 当中, 首先以 10 作为参数调用 `NewTwiceClosure` 函数构造一个闭包函数,
返回的闭包函数保存在 `fn` 闭包函数类型的变量中. 然后每次调用 `fn` 闭包函数将返回翻倍后的结果.

Go语言层面的很容易理解. 但是闭包函数如何使用汇编实现呢? 首先构造一个 `FunTwiceClosure` 结构体类型, 用来表示闭包对象:

```
//go:noescape
func ptrToFunc(p unsafe.Pointer) func() int

//go:noescape
func asmFunTwiceClosureAddr() uintptr

//go:noescape
func asmFunTwiceClosureBody() int

type TwiceClosure struct {
	F uintptr
	X int
}

func NewTwiceClosure(x int) func() int {
	var p = TwiceClosure{
		F: asmFunTwiceClosureAddr(),
		X: x,
	}

	return ptrToFunc(unsafe.Pointer(&p))
}
```

`NewTwiceClosure` 当中的 F 表示闭包函数的函数指令的地址, X 表示闭包捕获的外部变量. 如果捕获多个外部变量, 那么结构体
也是需要做相应的调整. 

构造 `NewTwiceClosure` 结构体对象, 其中 `asmFunTwiceClosureAddr` 函数用于辅助获取必报函数的函数指令的地址, 采用
汇编实现. `ptrToFunc` 辅助函数将结构体转为闭包函数对象返回.


```
#include "textflag.h"

TEXT ·ptrToFunc(SB), NOSPLIT, $0
    MOVQ ptr+0(FP), AX // AX = ptr
    MOVQ AX, ret+8(FP) // return AX
    RET

TEXT ·asmFunTwiceClosureAddr(SB), NOSPLIT, $0
    LEAQ ·asmFunTwiceClosureBody(SB), AX // AX=·asmFunTwiceClosureAddr(SB)
    MOVQ AX, ret+0(FP)
    RET

TEXT ·asmFunTwiceClosureBody(SB), NOSPLIT|NEEDCTXT, $0
    MOVQ 8(DX), AX     // 获取 X 的值
    ADDQ AX, AX        // AX += AX
    MOVQ AX, 8(DX)     // 将 X 保存到 DX 当中
    MOVQ AX, ret+0(FP) // return
    RET
```

其中 `·asmFunTwiceClosureAddr` 的当中 `·asmFunTwiceClosureBody(SB)` 代表了 `asmFunTwiceClosureBody` 函数
的地址, 注意操作当中要把 `(SB)` 部分带上. 最重要的是 `·asmFunTwiceClosureBody` 函数的实现: **它带有一个 NEEDCTXT
标志. 采用 `NEEDCTXT` 标志定义的汇编函数表示需要一个上下文环境, 在 AMD64 环境中通过 DX 寄存器来传递这个上下文环境指针,
也就是 `TwiceClosure` 结构体指针`. 
