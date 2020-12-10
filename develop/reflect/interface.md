# interface

## 值接受者和指针接受者的区别

在调用方法的时候, 值类型既可以调用 `值接收者` 的方法, 也可以调用 `指针接收者` 的方法; 指针类型既可以调用 `指针接收者` 
的方法, 也可以调用 `值接收者` 的方法.

换句话说, 不管方法的接收者是什么类型, 该类型的值和指针都可以调用, 不必严格符合接收者的类型.

```
type Person struct {
    age int
}

func (p Person) howOld() int {
    return p.age
}

func (p *Person) growUp() {
    p.age += 1
}

func main() {
    // val 是值类型
    val := Person{age: 18}

    // 值类型 调用接收者也是值类型的方法
    fmt.Println(val.howOld())

    // 值类型 调用接收者是指针类型的方法
    val.growUp()
    fmt.Println(val.howOld())


    // ptr 是指针类型
    ptr := &Person{age: 100}

    // 指针类型 调用接收者是值类型的方法
    fmt.Println(ptr.howOld())

    // 指针类型 调用接收者也是指针类型的方法
    ptr.growUp()
    fmt.Println(ptr.howOld())
}
```

输出的结果是:

```
18
19
100
101
```

调用了 `growUp` 函数之后, 不管调用者是值类型还是指针类型, 它的 `Age` 值都改变了.

实际上, 当类型和方法的接收者类型不同时, 其实是编译器在背后做了一些工作.

|     |  值接收者 | 指针接收者 |
| --- | ---- | ---- |
| 值类型 | 方法会使用调用者的一个副本, 类似于"传值" | 使用`值的引用`来调用方法, 上例 `val.growUp()` 实际上是 `(&val).groupUp` |
| 指针类型 | 指针类型`被解引用`为值, 上例中的 `ptr.howOld()` 实际上是 `(*ptr).howOld()` | 实际上也是"传值", 方法里的操作会影响到调用者, 类似与指针传参, 拷贝了一份指针 "


## 值接收者和指针接收者

结论: **实现了接收者是值类型的方法, 相当于自动实现了接收者是指针类型的方法; 而实现了接收者是指针类型的方法, 不会自动生成
对应接收者是值类型的方法**

解释: 接收者是指针类型, 很可能在方法中对接收者的属性进行更改操作, 从而影响接收者; 而对于接收者是值类型的放, 在方法中不会
对接收者本身产生影响. 因此, 当实现了一个接收者是值类型的方法, 可以自动生成一个接收者是对应指针类型的方法, 因为两者都不影
响接收者. 


记住下面的结论:

> 如果实现了接收者是值类型的方法, 会隐含地实现了接收者是指针类型的方法.

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
type Person interface {
	grow()
}

type Student struct {
	age int
	name string
}

func (p Student) grow() {
	p.age += 1
	return
}

func main() {
	var qcrao = Person(Student{age: 18, name:"san"})

	fmt.Println(qcrao)
}
```

使用 `go tool compile -S example.go` 打印汇编代码. go 版本是 `1.13`

```
"".main STEXT size=219 args=0x0 locals=0x70
	0x0000 00000 (example.go:19)	TEXT	"".main(SB), ABIInternal, $112-0
	0x0000 00000 (example.go:19)	MOVQ	(TLS), CX
	0x0009 00009 (example.go:19)	CMPQ	SP, 16(CX)
	0x000d 00013 (example.go:19)	JLS	209
	0x0013 00019 (example.go:19)	SUBQ	$112, SP
	0x0017 00023 (example.go:19)	MOVQ	BP, 104(SP)
	0x001c 00028 (example.go:19)	LEAQ	104(SP), BP
	0x0021 00033 (example.go:19)	FUNCDATA	$0, gclocals·7d2d5fca80364273fb07d5820a76fef4(SB)
	0x0021 00033 (example.go:19)	FUNCDATA	$1, gclocals·ec3d218f521c2fd49f31b3bbe678b423(SB)
	0x0021 00033 (example.go:19)	FUNCDATA	$2, gclocals·bfec7e55b3f043d1941c093912808913(SB)
	0x0021 00033 (example.go:19)	FUNCDATA	$3, "".main.stkobj(SB)
	0x0021 00033 (example.go:20)	PCDATA	$0, $0
	0x0021 00033 (example.go:20)	PCDATA	$1, $1
	0x0021 00033 (example.go:20)	XORPS	X0, X0
	0x0024 00036 (example.go:20)	MOVUPS	X0, ""..autotmp_7+80(SP)
	0x0029 00041 (example.go:20)	MOVQ	$0, ""..autotmp_7+96(SP)
	0x0032 00050 (example.go:20)	MOVQ	$18, ""..autotmp_7+80(SP)
	0x003b 00059 (example.go:20)	PCDATA	$0, $1
	0x003b 00059 (example.go:20)	LEAQ	go.string."san"(SB), AX
	0x0042 00066 (example.go:20)	PCDATA	$0, $0
	0x0042 00066 (example.go:20)	MOVQ	AX, ""..autotmp_7+88(SP)
	0x0047 00071 (example.go:20)	MOVQ	$3, ""..autotmp_7+96(SP)
	0x0050 00080 (example.go:20)	PCDATA	$0, $1
	0x0050 00080 (example.go:20)	LEAQ	go.itab."".Student,"".Person(SB), AX
	0x0057 00087 (example.go:20)	PCDATA	$0, $0
	0x0057 00087 (example.go:20)	MOVQ	AX, (SP)
	0x005b 00091 (example.go:20)	PCDATA	$0, $1
	0x005b 00091 (example.go:20)	PCDATA	$1, $0
	0x005b 00091 (example.go:20)	LEAQ	""..autotmp_7+80(SP), AX
	0x0060 00096 (example.go:20)	PCDATA	$0, $0
	0x0060 00096 (example.go:20)	MOVQ	AX, 8(SP)
	0x0065 00101 (example.go:20)	CALL	runtime.convT2I(SB)
	0x006a 00106 (example.go:20)	PCDATA	$0, $1
	0x006a 00106 (example.go:20)	MOVQ	24(SP), AX
	0x006f 00111 (example.go:20)	PCDATA	$0, $2
	0x006f 00111 (example.go:20)	MOVQ	16(SP), CX
	0x0074 00116 (example.go:22)	TESTQ	CX, CX
	0x0077 00119 (example.go:22)	JEQ	125
	0x0079 00121 (example.go:22)	MOVQ	8(CX), CX
	0x007d 00125 (example.go:22)	PCDATA	$1, $2
	0x007d 00125 (example.go:22)	XORPS	X0, X0
	0x0080 00128 (example.go:22)	MOVUPS	X0, ""..autotmp_15+64(SP)
	0x0085 00133 (example.go:22)	PCDATA	$0, $1
	0x0085 00133 (example.go:22)	MOVQ	CX, ""..autotmp_15+64(SP)
	0x008a 00138 (example.go:22)	PCDATA	$0, $0
	0x008a 00138 (example.go:22)	MOVQ	AX, ""..autotmp_15+72(SP)
	0x008f 00143 (<unknown line number>)	NOP
	0x008f 00143 ($GOROOT/src/fmt/print.go:274)	PCDATA	$0, $1
	0x008f 00143 ($GOROOT/src/fmt/print.go:274)	MOVQ	os.Stdout(SB), AX
	0x0096 00150 ($GOROOT/src/fmt/print.go:274)	PCDATA	$0, $2
	0x0096 00150 ($GOROOT/src/fmt/print.go:274)	LEAQ	go.itab.*os.File,io.Writer(SB), CX
	0x009d 00157 ($GOROOT/src/fmt/print.go:274)	PCDATA	$0, $1
	0x009d 00157 ($GOROOT/src/fmt/print.go:274)	MOVQ	CX, (SP)
	0x00a1 00161 ($GOROOT/src/fmt/print.go:274)	PCDATA	$0, $0
	0x00a1 00161 ($GOROOT/src/fmt/print.go:274)	MOVQ	AX, 8(SP)
	0x00a6 00166 ($GOROOT/src/fmt/print.go:274)	PCDATA	$0, $1
	0x00a6 00166 ($GOROOT/src/fmt/print.go:274)	PCDATA	$1, $0
	0x00a6 00166 ($GOROOT/src/fmt/print.go:274)	LEAQ	""..autotmp_15+64(SP), AX
	0x00ab 00171 ($GOROOT/src/fmt/print.go:274)	PCDATA	$0, $0
	0x00ab 00171 ($GOROOT/src/fmt/print.go:274)	MOVQ	AX, 16(SP)
	0x00b0 00176 ($GOROOT/src/fmt/print.go:274)	MOVQ	$1, 24(SP)
	0x00b9 00185 ($GOROOT/src/fmt/print.go:274)	MOVQ	$1, 32(SP)
	0x00c2 00194 ($GOROOT/src/fmt/print.go:274)	CALL	fmt.Fprintln(SB)
	0x00c7 00199 (<unknown line number>)	MOVQ	104(SP), BP
	0x00cc 00204 (<unknown line number>)	ADDQ	$112, SP
	0x00d0 00208 (<unknown line number>)	RET
```

从第15行开始看, 前面的对当前的分析不重要. 可以忽略.



```cgo
// 源码位置 src/runtime/iface.go
var (
	uint64Eface interface{} = uint64InterfacePtr(0)
	stringEface interface{} = stringInterfacePtr("")
	sliceEface  interface{} = sliceInterfacePtr(nil)

	uint64Type *_type = (*eface)(unsafe.Pointer(&uint64Eface))._type
	stringType *_type = (*eface)(unsafe.Pointer(&stringEface))._type
	sliceType  *_type = (*eface)(unsafe.Pointer(&sliceEface))._type
)

// type(uint64) -> interface
func convT64(val uint64) (x unsafe.Pointer) {
	if val == 0 {
		x = unsafe.Pointer(&zeroVal[0])
	} else {
		x = mallocgc(8, uint64Type, false)
		*(*uint64)(x) = val
	}
	return
}

// type(uint64) -> interface
func convTstring(val string) (x unsafe.Pointer) {
	if val == "" {
		x = unsafe.Pointer(&zeroVal[0])
	} else {
		x = mallocgc(unsafe.Sizeof(val), stringType, true)
		*(*string)(x) = val
	}
	return
}

// type -> iface 
func convT2I(tab *itab, elem unsafe.Pointer) (i iface) {
	t := tab._type
	
	// 启用了 -race 选项
	if raceenabled {
		raceReadObjectPC(t, elem, getcallerpc(), funcPC(convT2I))
	}
	// 启用了 -msan 选项
	if msanenabled {
		msanread(elem, t.size)
	}
	
	// 生成 itab 当中动态类型的内存空间(指针), 并将 elem 的值拷贝相应位置
	x := mallocgc(t.size, t, true) // convT2I 和 convT2Enoptr 的区别
	typedmemmove(t, x, elem)
	i.tab = tab
	i.data = x
	return
}

// iface -> iface
func convI2I(inter *interfacetype, i iface) (r iface) {
	tab := i.tab
	if tab == nil {
		return
	}
	if tab.inter == inter {
		r.tab = tab
		r.data = i.data
		return
	}
	r.tab = getitab(inter, tab._type, false)
	r.data = i.data
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