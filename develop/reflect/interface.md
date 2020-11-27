# interface


## `eface` vs `iface`

`eface` 和 `iface` 都是 Go 中描述接口的底层数据结构, 区别在于 `iface` 描述的接口包含方法, 而 `eface` 则是不包含
任何方法的空接口: `interface{}`. 

接口的底层数据结构:

```cgo
// 空接口, interface{} 的底层结构
type eface struct {
	_type *_type
	data  unsafe.Pointer
}

// 方法接口, 自定义的接口的数据结构
type iface struct {
	tab  *itab
	data unsafe.Pointer
}

type itab struct {
	inter *interfacetype // 接口类型
	_type *_type         // 值类型
	hash  uint32         // 值类型 _type 当中的 hash 的拷贝
	_     [4]byte
	fun   [1]uintptr
}

type interfacetype struct {
	typ     _type
	pkgpath name
	mhdr    []imethod
}

type name struct {
	bytes *byte
}
type imethod struct {
	name int32
	ityp int32
}

type _type struct {
	size       uintptr
	ptrdata    uintptr // 包含所有指针的内存前缀的大小. 如果为0, 表示的是一个值, 而非指针
	hash       uint32
	tflag      uint8
	align      uint8
	fieldalign uint8
	kind       uint8
	alg        uintptr
	gcdata     *byte
	str        int32
	ptrToThis  int32
}
```

`iface` 维护两个指针, `tab` 指向一个 `itab` 实体, 它表示接口的类型 以及赋值给这个接口的实体类型. `data` 则指向接口
具体的值, 一般情况下是一个指向堆内存的指针.

`itab` 结构: `inter` 字段描述了接口的类型(静态), `_type` 字段描述了接口值的类型(动态), 包括内存对齐方式, 大小等.
`fun` 字段放置接口类型的方法地址 (*与接口方法对于的具体数据类型的方法地址??*), 实现接口调用方法的动态分派.

为何 `fun` 数组的大小为1, 如果接口定义了多个方法怎么办? 实际上, 这里存储的是第一个方法的函数指针, 如果有更多的方法, 在
它之后的内存空间里继续存在. 从汇编角度来看, 通过增加地址就能获取到这些函数指针, 没什么影响. 方法是按照函数名称的字典顺序
进行排列的.

`iterfacetype` 类型, 它描述了接口的类型(静态). 它包装了 `_type` 类型, `_type` 是描述 Go 语言各种数据类型的结构体.
`mhdr` 字段, 表示接口定义的函数列表, 通过这里的内容, 在反射的时候可以获取实际函数的指针. `pkgpath` 描述了接口的包名.


![image](/images/develop_interface_iface.png)


相比 `iface`, `eface` 就比较简单, 只维护了一个 `_type` 字段, 表示空接口所承载的具体的实体类型. `data` 描述了其值.

![image](/images/develop_interface_eface.png)


## 接口的动态类型和动态值

接口类型和 `nil` 做比较:

接口值的零值指 `动态类型` 和 `动态值` 都为 nil. 当且仅当这两部分的值都为 `nil` 的状况下, 这个接口值才被认为 `接口值 == nil`.

例子1:

```cgo
type Coder interface {
    code()
}

type Gopher struct {
    name string
}

func (g Gopher) code() {
    fmt.Printf("%s is coding\n", g.name)
}

func main() {
    var c Coder
    fmt.Println(c == nil) // true
    fmt.Printf("c: %T, %v\n", c, c)

    var g *Gopher
    fmt.Println(g == nil) // true

    c = g
    fmt.Println(c == nil) // fasle, 值类型为 *Gopher, 值是 nil
    fmt.Printf("c: %T, %v\n", c, c)
}
```

例子2:

```cgo
type MyError struct {}

func (i MyError) Error() string {
    return "MyError"
}

func Process() error {
    var err *MyError = nil
    return err // 值类型为 *MyError, 值是 nil
}

func main() {
    err := Process()
    fmt.Println(err) // nil

    fmt.Println(err == nil) // false
}
```

例子3:

```cgo
type iface struct {
    itab, data uintptr
}

func main() {
    var a interface{} = nil
    
    /* 
     * 等价形式:
     * var t *int
     * var b interface{} = t
     */
    var b interface{} = (*int)(nil) 

    x := 5
    var c interface{} = (*int)(&x)

    ia := *(*iface)(unsafe.Pointer(&a)) // 值类型为 nil, 值为 nil
    ib := *(*iface)(unsafe.Pointer(&b)) // 值类型为 *int, 值为 nil
    ic := *(*iface)(unsafe.Pointer(&c)) // 值类型为 *int, 值为 5

    fmt.Println(ia, ib, ic) 

    fmt.Println(*(*int)(unsafe.Pointer(ic.data)))
}
```


## 编译器自动检测类型是否实现接口

奇怪的代码:

```cgo
var _ io.Writer = (*myWriter)(nil)
```

为啥有上面的代码, 上面的代码究竟是要干什么? 编译器会通过上述的代码检查 `*myWriter` 类型是否实现了 `io.Writer` 接口.

例子:

```cgo
type myWriter struct{
}

/*
func (w myWriter) Write(p []byte) (n int, err error) {
    return
}
*/

func main() {
    // 检查 *myWriter 是否实现了 io.Writer 接口
    var _ io.Writer = (*myWriter)(nil)
    
    // 检查 myWriter 是否实现了 io.Writer 接口
    var _ io.Writer = myWriter{}
}
```

> 注释掉为 myWriter 定义的 Writer 函数后, 运行程序, 报错信息: `*myWriter/myWriter 未实现 io.Writer 接口`, 也
就是未实现 Write 方法. 解除注释后, 运行程序不报错.


实际上, 上述赋值语句会发生隐式转换, 在转换的过程中, 编译器会检测等号右边的类型是否实现了等号左边接口所规定的函数.

### 接口的构造过程

先从一个例子开始:

```cgo
package main

import "fmt"

type Person interface {
	growUp()
}

type Student struct {
	age int
}

func (p Student) growUp() {
	p.age += 1
	return
}

func main() {
	var qcrao = Person(Student{age: 18})

	fmt.Println(qcrao)
}
```

使用 `go tool compile -S example.go` 打印汇编代码. go 版本是 `1.13`

```
"".main STEXT size=148 args=0x0 locals=0x58
        0x0000 00000 (example.go:18)    TEXT    "".main(SB), ABIInternal, $88-0
        0x0000 00000 (example.go:18)    MOVQ    (TLS), CX
        0x0009 00009 (example.go:18)    CMPQ    SP, 16(CX)
        0x000d 00013 (example.go:18)    JLS     138
        0x000f 00015 (example.go:18)    SUBQ    $88, SP
        0x0013 00019 (example.go:18)    MOVQ    BP, 80(SP)
        0x0018 00024 (example.go:18)    LEAQ    80(SP), BP
        0x001d 00029 (example.go:18)    FUNCDATA        $0, gclocals·69c1753bd5f81501d95132d08af04464(SB)
        0x001d 00029 (example.go:18)    FUNCDATA        $1, gclocals·568470801006e5c0dc3947ea998fe279(SB)
        0x001d 00029 (example.go:18)    FUNCDATA        $2, gclocals·bfec7e55b3f043d1941c093912808913(SB)
        0x001d 00029 (example.go:18)    FUNCDATA        $3, "".main.stkobj(SB)
        0x001d 00029 (example.go:19)    PCDATA  $0, $0
        0x001d 00029 (example.go:19)    PCDATA  $1, $0
        0x001d 00029 (example.go:19)    MOVQ    $18, (SP)
        0x0025 00037 (example.go:19)    CALL    runtime.convT64(SB)
        0x002a 00042 (example.go:19)    PCDATA  $0, $1
        0x002a 00042 (example.go:19)    MOVQ    8(SP), AX
        0x002f 00047 (example.go:21)    PCDATA  $0, $2
        0x002f 00047 (example.go:21)    MOVQ    go.itab."".Student,"".Person+8(SB), CX
        0x0036 00054 (example.go:21)    PCDATA  $1, $1
        0x0036 00054 (example.go:21)    XORPS   X0, X0
        0x0039 00057 (example.go:21)    MOVUPS  X0, ""..autotmp_16+64(SP)
        0x003e 00062 (example.go:21)    PCDATA  $0, $1
        0x003e 00062 (example.go:21)    MOVQ    CX, ""..autotmp_16+64(SP)
        0x0043 00067 (example.go:21)    PCDATA  $0, $0
        0x0043 00067 (example.go:21)    MOVQ    AX, ""..autotmp_16+72(SP)
        0x0048 00072 (<unknown line number>)    NOP
        0x0048 00072 ($GOROOT/src/fmt/print.go:274)     PCDATA  $0, $1
        0x0048 00072 ($GOROOT/src/fmt/print.go:274)     MOVQ    os.Stdout(SB), AX
        0x004f 00079 ($GOROOT/src/fmt/print.go:274)     PCDATA  $0, $2
        0x004f 00079 ($GOROOT/src/fmt/print.go:274)     LEAQ    go.itab.*os.File,io.Writer(SB), CX
        0x0056 00086 ($GOROOT/src/fmt/print.go:274)     PCDATA  $0, $1
        0x0056 00086 ($GOROOT/src/fmt/print.go:274)     MOVQ    CX, (SP)
        0x005a 00090 ($GOROOT/src/fmt/print.go:274)     PCDATA  $0, $0
        0x005a 00090 ($GOROOT/src/fmt/print.go:274)     MOVQ    AX, 8(SP)
        0x005f 00095 ($GOROOT/src/fmt/print.go:274)     PCDATA  $0, $1
        0x005f 00095 ($GOROOT/src/fmt/print.go:274)     PCDATA  $1, $0
        0x005f 00095 ($GOROOT/src/fmt/print.go:274)     LEAQ    ""..autotmp_16+64(SP), AX
        0x0064 00100 ($GOROOT/src/fmt/print.go:274)     PCDATA  $0, $0
        0x0064 00100 ($GOROOT/src/fmt/print.go:274)     MOVQ    AX, 16(SP)
        0x0069 00105 ($GOROOT/src/fmt/print.go:274)     MOVQ    $1, 24(SP)
        0x0072 00114 ($GOROOT/src/fmt/print.go:274)     MOVQ    $1, 32(SP)
        0x007b 00123 ($GOROOT/src/fmt/print.go:274)     CALL    fmt.Fprintln(SB)
        0x0080 00128 (<unknown line number>)    MOVQ    80(SP), BP
        0x0085 00133 (<unknown line number>)    ADDQ    $88, SP
        0x0089 00137 (<unknown line number>)    RET
        0x008a 00138 (<unknown line number>)    NOP
        0x008a 00138 (example.go:18)    PCDATA  $1, $-1
        0x008a 00138 (example.go:18)    PCDATA  $0, $-1
        0x008a 00138 (example.go:18)    CALL    runtime.morestack_noctxt(SB)
        0x008f 00143 (example.go:18)    JMP     0
```

从第15行开始看, 前面的对当前的分析不重要. 可以忽略.



```cgo

var (
	uint64Eface interface{} = uint64InterfacePtr(0)
	stringEface interface{} = stringInterfacePtr("")
	sliceEface  interface{} = sliceInterfacePtr(nil)

	uint64Type *_type = (*eface)(unsafe.Pointer(&uint64Eface))._type
	stringType *_type = (*eface)(unsafe.Pointer(&stringEface))._type
	sliceType  *_type = (*eface)(unsafe.Pointer(&sliceEface))._type
)

func convT64(val uint64) (x unsafe.Pointer) {
	if val == 0 {
		x = unsafe.Pointer(&zeroVal[0])
	} else {
		x = mallocgc(8, uint64Type, false)
		*(*uint64)(x) = val
	}
	return
}

func convTstring(val string) (x unsafe.Pointer) {
	if val == "" {
		x = unsafe.Pointer(&zeroVal[0])
	} else {
		x = mallocgc(unsafe.Sizeof(val), stringType, true)
		*(*string)(x) = val
	}
	return
}
```

> func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer
>
> 分配一个 size 大小字节的 object.
> 从per-P缓存的空闲列表中分配小对象; 从堆直接分配大对象( >32 kB)
> 
> 参数 needzero 描述分配的对象是否是指针
> 参数 typ 描述当前分配对象的数据类型



[参考文档](https://mp.weixin.qq.com/s/EbxkBokYBajkCR-MazL0ZA)