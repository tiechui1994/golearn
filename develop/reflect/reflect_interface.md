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
| 指针类型 | 指针类型`被解引用`为值, 上例中的 `ptr.howOld()` 实际上是 `(*ptr).howOld()` | 实际上也是"传值", 方法里的操作会影响到调用者, 类似与指针传参, 拷贝了一份指针 |

## 值接收者和指针接收者

结论: **实现了接收者是值类型的方法, 相当于自动实现了接收者是指针类型的方法; 而实现了接收者是指针类型的方法, 不会自动生成
对应接收者是值类型的方法**.


举个例子:

```
type Animal interface {
    howOld() int
    growUp()
}

type Person struct {
    age int
}

func (p Person) howOld() int {
    return p.age
}

func (p *Person) growUp() {
    p.age += 1
}
```

上述代码定义了一个接口 `Animal`, 接口定义了两个函数. 接着定义了一个结构体 `Person`, 它实现了两个方法, 一个值接收者,
一个是指针接收者.

下面我们看两个对上述代码的两个测试:

测试一:

```
func main() {
    var c Animal = &Person{"Go"}
    c.howOld()
    c.growUp()
}
```

测试二:

```
func main() {
    var c Animal = Person{"Go"}
    c.howOld()
    c.growUp()
}
```

分别对上述两个代码进行编译运行, 会出现什么情况?

测试一, 可以正常编译运行.

测试二, 编译失败, 出现错误, `Person does not implement Animal (growUp method has pointer receiver)`. 错误
的意思是 `Person没有实现Animal接口(growUp方法只有一个指针接收者)`

> 分析: 上述的两个测试唯一的不同的, 测试一是将 `&Person` 赋值给 `c`, 而测试二是将 `Person` 赋值给了 `c`. 根据错误
> 的提示来看, 说 `growUp方法只有一个指针接收者`, 也就意味着, `growUp方法没有一个值接收者` (出错的时候 c 是一个值). 
> 进一步看方法 `howOld`, 可以推到得到 `howOld方法既拥有值接收者(Person真实定义就是值接收者), 也拥有指针接收者(编译
> 器隐式的转换).`, 从而得出前面刚开始提到的结论. **实现了接收者是值类型的方法, 相当于自动实现了接收者是指针类型的方法; 
> 而实现了接收者是指针类型的方法, 不会自动生成对应接收者是值类型的方法**

上述现象的一个解释: 接收者是指针类型, 很可能在方法中对接收者的属性进行更改操作, 从而影响接收者; 而对于接收者是值类型的方
法, 在方法中不会对接收者本身产生影响. 因此, 当实现了一个接收者是值类型的方法, 可以自动生成一个接收者是对应指针类型的方法, 
因为两者都不影响接收者. 


记住下面的结论:

> 如果实现了接收者是值类型的方法, 会隐含地实现了接收者是指针类型的方法.

在实际的汇编当中, 对于接收者是值类型的方法, go 会自动添加接收者是指针类型的方法, 但是对于接收者是指针类型的方法, 却只添
加其本身. 参考 `type.*"".X` 表示 X 指针类型, `type."".X` 表示 X 类型(这里的 X 可以为结构体, 函数等)

## `eface` vs `iface` (反射基础)

`eface` 和 `iface` 都是 Go 中描述接口的底层数据结构, 区别在于 `iface` 描述的接口包含方法, 而 `eface` 则是不包含
任何方法的空接口: `interface{}`. 

接口的底层数据结构:

```cgo
// 不带方法, 就是 interface{}
type eface struct {
	_type *_type
	data  unsafe.Pointer
}

// 带方法
type iface struct {
	tab  *itab
	data unsafe.Pointer
}
type itab struct {
	inter *interfacetype // 接口的静态类型
	_type *_type         // 值类型(动态混合类型)
	hash  uint32         // 值类型 _type 当中的 hash 的拷贝
	_     [4]byte
	fun   [1]uintptr
}
type interfacetype struct {
	typ     _type
	pkgpath name      // 汇编当中对应为: type..importpath."".+0
	mhdr    []imethod // 比较复杂
}
type name struct {
	bytes *byte
}
type imethod struct {
	name int32
	ityp int32
}



// go1.13 之前的版本
type _type struct {
	size       uintptr
	ptrdata    uintptr // 包含所有指针的内存前缀的大小. 如果为0, 表示的是一个值, 而非指针
	hash       uint32
	tflag      uint8
	align      uint8
	fieldAlign uint8
	kind       uint8
	alg        *typeAlg // 包含了 hash 和 hash 两个方法.
	gcdata     *byte
	str        int32 // nameOff
    	ptrToThis  int32 // typeOff
}
type typeAlg struct {
	// hashing objects of this type (ptr to object, seed) -> hash
	hash func(unsafe.Pointer, uintptr) uintptr
	// comparing objects of this type (ptr to object A, ptr to object B) -> ==?
	equal func(unsafe.Pointer, unsafe.Pointer) bool
}

// go1.14 版本之后
type _type struct {
	size       uintptr
	ptrdata    uintptr // 包含所有指针的内存前缀的大小. 如果为0, 表示的是一个值, 而非指针
	hash       uint32
	tflag      uint8
	align      uint8
	fieldAlign uint8
	kind       uint8
	// equal: comparing objects of this type (ptr to object A, ptr to object B) -> ==?
	// 可能的值: 
	// runtime.interequal, iface
	// runtime.nilinterequal, eface
	// runtime.memequal64, 结构体, 指针,  
	// runtime.strequal, 专门针对 str 的优化
	// runtime.interequal, 专门针对 int 的优化
	equal func(unsafe.Pointer, unsafe.Pointer) bool 
	// gcdata: 存储垃圾收集器的 GC 类型数据. 如果 KindGCProg 位在 kind 中设置, 则 gcdata 是一个 GC 程序.
	// 否则它是一个 ptrmask 位图.
	gcdata     *byte
	str        int32 // nameOff
	ptrToThis  int32 // typeOff
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


为了 `iface` 与 `eface` 的统一,  `eface` 也维护两个指针. `_type`, 表示空接口所承载的具体的实体类型. `data` 描
述了其值.

![image](/images/develop_interface_eface.png)


顺便说一下: `_type` 是所有类型原始信息的元信息. 例如:

```cgo
type arraytype struct {
	typ   _type
	elem  *_type
	slice *_type
	len   uintptr
}

type chantype struct {
	typ  _type
	elem *_type
	dir  uintptr
}

type functype struct {
	typ      _type
	inCount  uint16
	outCount uint16 
}

// 48+8
type ptrtype struct {
	typ   _type
	elem *_type 
}

// 48+8+24*N
type structtype struct {
	typ     _type
	pkgPath name
	fields  []structField // sorted by offset
}
type structField struct {
	name        name    // name is always non-empty
	typ         *_type  // type of field
	offsetEmbed uintptr // byte offset of field<<1 | isEmbedded
}

// 48+40
type maptype struct {
	typ     _type
	key    *_type // map key type
	elem   *_type // map element (value) type
	bucket *_type // internal bucket structure
	// function for hashing keys (ptr to key, seed) -> hash
	hasher     func(unsafe.Pointer, uintptr) uintptr
	keysize    uint8  // size of key slot
	valuesize  uint8  // size of value slot
	bucketsize uint16 // size of bucket
	flags      uint32
}

// 48+8+8*N
type interfacetype struct {
	typ     _type     // 类型元信息
	pkgpath name      // 包路径和描述信息等等
	mhdr    []imethod // 方法
}
```

在 arraytype, chantype 等中保存类型的元信息靠的是 _type. interfacetype(interface)的定义也是类似的. 由于 Go 当
中函数方法是以包为单位隔离的. 因此 interfacetype 除了保存 _type 还需要保存包路径等描述信息. mhdr 保存的是 interface
函数方法在段内的偏移 offset.

这些类型的 _type 信息是在编译的时候确定的. 可以通过汇编代码查找到相关类型的 _type 的详情.

## 接口的动态类型和动态值

接口类型和 `nil` 做比较:

接口值的零值指 `动态类型` 和 `动态值` 都为 nil. 当且仅当这两部分的值都为 `nil` 的状况下, 这个接口值才被认为是nil, 即
`接口值 == nil`.

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
"".main STEXT size=331 args=0x0 locals=0xa8
    00000 (main.go:19)	TEXT	"".main(SB), ABIInternal, $168-0
    00000 (main.go:19)	MOVQ	(TLS), CX
    00009 (main.go:19)	LEAQ	-40(SP), AX
    00014 (main.go:19)	CMPQ	AX, 16(CX)
    00018 (main.go:19)	JLS	321
    00024 (main.go:19)	SUBQ	$168, SP
    00031 (main.go:19)	MOVQ	BP, 160(SP)
    00039 (main.go:19)	LEAQ	160(SP), BP
    00047 (main.go:20)	XORPS	X0, X0
    00050 (main.go:20)	MOVUPS	X0, ""..autotmp_1+136(SP)
    00058 (main.go:20)	MOVQ	$0, ""..autotmp_1+152(SP)
    00070 (main.go:20)	MOVQ	$18, ""..autotmp_1+136(SP)
    00082 (main.go:20)	LEAQ	go.string."san"(SB), AX
    00089 (main.go:20)	MOVQ	AX, ""..autotmp_1+144(SP)
    00097 (main.go:20)	MOVQ	$3, ""..autotmp_1+152(SP)
    00109 (main.go:20)	LEAQ	go.itab."".Student,"".Person(SB), AX
    00116 (main.go:20)	MOVQ	AX, (SP)
    00120 (main.go:20)	LEAQ	""..autotmp_1+136(SP), AX
    00128 (main.go:20)	MOVQ	AX, 8(SP)
    00133 (main.go:20)	CALL	runtime.convT2I(SB)
    00138 (main.go:20)	MOVQ	24(SP), AX
    00143 (main.go:20)	MOVQ	16(SP), CX
    00148 (main.go:20)	MOVQ	CX, "".qcrao+64(SP)
    00153 (main.go:20)	MOVQ	AX, "".qcrao+72(SP)
    00158 (main.go:22)	MOVQ	"".qcrao+72(SP), AX
    00163 (main.go:22)	MOVQ	"".qcrao+64(SP), CX
    00168 (main.go:22)	MOVQ	CX, ""..autotmp_3+80(SP)
    00173 (main.go:22)	MOVQ	AX, ""..autotmp_3+88(SP)
    00178 (main.go:22)	MOVQ	CX, ""..autotmp_4+56(SP)
    00183 (main.go:22)	CMPQ	""..autotmp_4+56(SP), $0
    00189 (main.go:22)	JNE	193
    00191 (main.go:22)	JMP	319
    00193 (main.go:22)	TESTB	AL, (CX)
    00195 (main.go:22)	MOVQ	8(CX), AX
    00199 (main.go:22)	MOVQ	AX, ""..autotmp_4+56(SP)
    00204 (main.go:22)	JMP	206
    00206 (main.go:22)	PCDATA	$1, $5
    00206 (main.go:22)	XORPS	X0, X0
    00209 (main.go:22)	MOVUPS	X0, ""..autotmp_2+96(SP)
    00214 (main.go:22)	LEAQ	""..autotmp_2+96(SP), AX
    00219 (main.go:22)	MOVQ	AX, ""..autotmp_6+48(SP)
    00224 (main.go:22)	TESTB	AL, (AX)
    00226 (main.go:22)	MOVQ	""..autotmp_4+56(SP), CX
    00231 (main.go:22)	MOVQ	""..autotmp_3+88(SP), DX
    00236 (main.go:22)	MOVQ	CX, ""..autotmp_2+96(SP)
    00241 (main.go:22)	MOVQ	DX, ""..autotmp_2+104(SP)
    00246 (main.go:22)	TESTB	AL, (AX)
    00248 (main.go:22)	JMP	250
    00250 (main.go:22)	MOVQ	AX, ""..autotmp_5+112(SP)
    00255 (main.go:22)	MOVQ	$1, ""..autotmp_5+120(SP)
    00264 (main.go:22)	MOVQ	$1, ""..autotmp_5+128(SP)
    00276 (main.go:22)	MOVQ	AX, (SP)
    00280 (main.go:22)	MOVQ	$1, 8(SP)
    00289 (main.go:22)	MOVQ	$1, 16(SP)
    00298 (main.go:22)	CALL	fmt.Println(SB)
    00303 (main.go:23)	MOVQ	160(SP), BP
    00311 (main.go:23)	ADDQ	$168, SP
    00318 (main.go:23)	RET
    00319 (main.go:22)	JMP	206
    00321 (main.go:22)	NOP
    00321 (main.go:19)	CALL	runtime.morestack_noctxt(SB)
    00326 (main.go:19)	JMP	0
```

最重要核心的代码是 `runtime.convT2I(SB)`, 该函数位于 `src/runtime/iface.go` 当中. `convT2I()` 函数是将一个 
`struct` 对象转换为 `iface`. 类似的方法还有, `convT64()`, `convTstring()` 等.这些方法都涉及到了接口的动态值.

下面看一下其中几个代表性的函数源码:

```cgo
// src/runtime/iface.go
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
```

> func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer
>
> 分配一个 size 大小字节的 object.
> 从per-P缓存的空闲列表中分配小对象; 从堆直接分配大对象( >32 kB)
> 
> 参数 needzero 描述分配的对象是否是指针
> 参数 typ 描述当前分配对象的数据类型

上述的代码逻辑都比较简单, 这里就不再解释了.

## 类型转换与断言的区别

Go 当中不允许饮食类型转换, 也就是说在 `=` 两边, 变量的类型必须相同.

`类型转换`, `类型断言` 本质都是把一个类型转换为另外一个类型. 不同之处在于, 类型断言是`接口变量`进行的操作.

类型转换(任何类型变量):

`<结果类型值> := <目标类型>(<表达式>)`

```
var x = 15.21
y = int(x)
```


类型断言(必须是接口变量):

`<目标类型值>, <布尔参数> := <表达式>.(目标类型)`

`<目标类型值> := <表达式>.(目标类型)`

```
var x interface{} = &bytes.Buffer{}
y, ok := x.(Writer)

var x Reader = &bytes.Buffer{}
y, ok := x.(Writer)
```


### 类型转换

对于类型转换而言, 需要转换前后的两个类型是兼容的才可以.

### 断言

引申1: `fmt.Println` 函数的参数是 `interface`. 对于内置类型, 函数内部会用穷举法, 得出它的真实类型, 然后转换为字符
串打印. 而对于自定义类型, 首先确定该类型是否实现了 `String()` 方法, 如果实现了, 则直接打印输出 `String()` 方法的结
果; 否则, 会通过反射来遍历对象的成员进行打印.

```
type Student struct{
    Name string
    Age int
}

func main() {
    var s = Student{Name:"www", Age:18}
    fmt.Println(s)
}
```

输出结果为 `{www 18}`.

由于 `Student` 没有实现 `String()` 方法, 所以 `fmt.Println` 会利用反射获取成员变量.

如果增加一个 `String()` 方法:

```
func (s String) String() string {
    return fmt.Sprintf("[Name:%s], [Age:%d]", s.Name, s.Age)
}
```

输出结果为 `[Name:www], [Age:18]`


如果 `String()` 方法是下面这样的:

```
func (s *String) String() string {
    return fmt.Sprintf("[Name:%s], [Age:%d]", s.Name, s.Age)
}
```

输出结果为 `{www 18}`. 是不是有点意外? 其实仔细思考一下, 因为 `String()` 方法是指针接收者, 不会隐式生成值类型接收者
的 `String()` 方法, 那么在断言的时候, 会将 `Student{}` 判断为没有实现 `String()` 方法的接口. 因此按照一般的反射
方法进行打印喽了.

> 一般情况下, 实现 `String()` 方法最好是值接收者, 这样无论是值还是指针在打印的时候都会用到此方法.

## 接口间转换原理

原理:

`<interface类型, 实体类型> -> itab`

当判断一种类型是否满足某个接口时, Go 使用类型的方法集和接口所需要的方法集进行匹配. 如果类型的方法集完全包含接口的方法集,
则认为该类型实现可该接口.

例如某类型有 `m` 个方法, 某接口有 `n` 个方法, 则很容易知道这种判定的时间复杂度为 `O(mn)`, Go会对方法集的函数按照函数
名的字典序进行排序, 所以实际的时间复杂度为 `O(m+n)`.

接口转换另外一个背后的原理: 类型兼容. 

一个例子:

```
type coder interface {
    code()
    run()
}

type runner interface {
    run()
}

type Language string

func (l Language) code() {
}

func (l Language) run() {
}

func main() {
    var c coder = Language{}
    
    var r runner
    r = c
    fmt.Println(c, r)
}
```

代码定义了两个接口, `coder` 和 `runner`. 定义了一个实体类 `Language`. 类型  `Language` 实现了两个方法 `code()` 
和 `run()`. main 函数里定义了一个接口变量 `c`, 绑定了一个 `Language` 对象, 之后将 `c` 赋值给另一个接口变量 `r`.
赋值成功的原因在于 `c` 中包含了 `run()` 方法. 这样这两个接口变量完成了转换.

通过汇编可以得到, 上述的 `r=c` 背后实际上是调用了 `runtime.convI2I(SB)`, 也就是 `convI2I` 函数. 

```cgo
// iface -> iface
// inter 是接口类型, i 是源 iface, r 是最终转换的 iface
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

代码非常简单. 最重要的代码是 `getitab` 函数, 根据 `interfacetype`(接口类型) 和 `_type`(动态值的类型) 获取到 r 的
`itab` 值.

```cgo
// inter 接口类型, typ 值类型, canfail表示转换是否可接受失败.
// 当canfail为false, 在获取不到 itab 状况下会 panic
func getitab(inter *interfacetype, typ *_type, canfail bool) *itab {
	if len(inter.mhdr) == 0 {
		throw("internal error - misuse of itab")
	}
    
    // 也就是 typ 的最低位是 0 状况下, 直接快速失败. 至于为什么, 不太清楚.
	// tflagUncommon=1
	if typ.tflag&tflagUncommon == 0 {
		if canfail {
			return nil
		}
		name := inter.typ.nameOff(inter.mhdr[0].name)
		panic(&TypeAssertionError{nil, typ, &inter.typ, name.name()})
	}

	var m *itab
    // 首先, 查看现有表(itabTable, 一个保存了 itab 的全局未导出的变量) 以查看是否可以找到所需的itab.
    // 在这种状况下, 不要使用锁.(常识)
    // 
    // 使用 atomic 确保我们看到该线程完成的所有先前写操作. (下面使用 atomic 的原因解释)
    // 如果未找到所要的 itab, 则只能创建一个, 然后更新itabTable字段(在itabAdd中使用atomic.Storep)
    t := (*itabTableType)(atomic.Loadp(unsafe.Pointer(&itabTable)))
	if m = t.find(inter, typ); m != nil {
		goto finish
	}

	// 没有找到所需的 itab, 这种状况下, 在加锁的状况下再次查找. 这就是所谓的 dobule-checking
	lock(&itabLock)
	if m = itabTable.find(inter, typ); m != nil {
		unlock(&itabLock)
		goto finish
	}

	// 没有查找到. 只能先 create 一个 itab, 然后 update itabTable
	// 这个是在当前的线程栈上去操作分配内存.
	m = (*itab)(persistentalloc(unsafe.Sizeof(itab{})+uintptr(len(inter.mhdr)-1)*sys.PtrSize, 0, &memstats.other_sys))
	m.inter = inter
	m._type = typ
	m.init() // itab 初始化
	
	// 将 itab 添加到 itabTable 当中. 
	// 这当中可能发生 itabTable 的扩容(默认存储512个, 长度超过 75% 就会发生扩容.
	itabAdd(m) 
	unlock(&itabLock)
finish:
	if m.fun[0] != 0 {
		return m
	}
	if canfail {
		return nil
	}
	
	// 如果是断言, 当没有使用 _, ok := x.(X) 的状况下, 转换失败就 panic
	// 如果是转换, 在 T -> I, 则允许失败; 如果是 I -> I, 则不允许失败
	panic(&TypeAssertionError{concrete: typ, asserted: &inter.typ, missingMethod: m.init()})
}
```

接下来看一下, itabTable 是怎么缓存 itab 的.

首先, 全局变量 `itabTable` 的结构是长啥样子的?

```cgo
type itabTableType struct {
	size    uintptr             // entries 的长度. 大小必须是 2 的指数
	count   uintptr             // 当前已经填充了的 entries 的长度
	entries [itabInitSize]*itab // 真实的长度是 size, itabInitSize=512
}
```

数据结构比较简单, 但需要注意的点是:

1. `itabTableType` 的 entries 真实长度是 size, 不是 entries 数组的大小. 这个数组会在适当的状况下进行扩容的.

2. `itabTableType` 的 entries 使用的内存是连续的, 它类似一个 "切片" ("切片"只是整体结构上类似, 但并不是真的切片),
可以使用内存偏移的方法来获取 `entries` 当中存储的值. 也正式因为这一点, 它又特别像一个 table, 我猜测这也是称为 `itabTable`
的原因吧.

在 itabTableType 当中是怎样去查找 itab 的呢? 思路的使用了类似 hashtable 的技术. 接下来看看它是怎样去实现的.

> 这里 `哈希函数` 使用的是 `二次探测` 实现的. h(i) = (h0 + i*(i+1)/2) mod 2^k.

> 哈希函数冲突的解决方法:
>
> - 线性探测法: h(i) = (h0 + i) mod k
> - 二次探测法: h(i) = (h0 + i*(i+1)/2) mod k
> - 链地址法: 数组 + 链表

```cgo
func (t *itabTableType) find(inter *interfacetype, typ *_type) *itab {
    // 使用二次探测实现.
    // 探测顺序为 h(i) = (h0 + i*(i+1)/2) mod 2^k. i 表示第i次探测
    //          h(i) = ( h(i-1) + 1 ) mod 2^k. i 表示第i次探测 
    // 我们保证使用此探测序列击中所有表条目.
	mask := t.size - 1
	
	// itabHashFunc => inter.typ.hash ^ typ.hash
	// h 是初始化的 h0 的值, 也可以认为它是 itabTableType 当中 entries 的 index
	h := itabHashFunc(inter, typ) & mask
	for i := uintptr(1); ; i++ {
		p := (**itab)(add(unsafe.Pointer(&t.entries), h*sys.PtrSize))
		
		// 在这里使用atomic read, 因此如果我们看到 m != nil, 我们还将看到m字段的初始化.
		// m := *p
		m := (*itab)(atomic.Loadp(unsafe.Pointer(p)))
		if m == nil {
			return nil
		}
		if m.inter == inter && m._type == typ {
			return m
		}
		h += i
		h &= mask
	}
}
```

主要的逻辑就是进行二次探测, 看是否能找到合适的 itab. 你可能存在这样疑问, 当 entries 全部填满的话, 如果没有合适的itab,
上述的代码就会陷入死循环?  从逻辑角度考虑, 确实是这样的, 但是, entries 是永远不会被填满的, 最多在填充 75% 之后, 就会
发生扩容. 那么扩容发生在哪里呢, 前面的 `getitab` 函数当中的 `itabAdd` 就会导致扩容.

```cgo
func itabAdd(m *itab) {
	t := itabTable
	
	// 75% load factor, 发生扩容
	if t.count >= 3*(t.size/4) { 
	    // itabTable进行扩容.
        // t2的内存大小 = (2+2*t.size)*sys.PtrSize, sys.PtrSize是指针大小, 多出来的2表示的是size, count字段
        // 我们撒谎并告诉 malloc 我们想要无指针的内存, 因为所有指向的值都不在堆中.
		t2 := (*itabTableType)(mallocgc((2+2*t.size)*sys.PtrSize, nil, true))
		t2.size = t.size * 2

		// copy
        // 注意: 在复制时, 其他线程可能会寻找itab并找不到它. 没关系, 然后他们将尝试获取 itab 锁, 结果请等到复制完成.
        // 这里调用的是 t2.add 表示向 t2 当中添加 *itab. 遍历的是 itabTable 列表.
		iterate_itabs(t2.add)
		if t2.count != t.count {
			throw("mismatched count during itab table copy")
		}
		// 发布新的哈希表. 使用atomic write
		atomicstorep(unsafe.Pointer(&itabTable), unsafe.Pointer(t2))
		// Adopt the new table as our own.
		t = itabTable
		// Note: the old table can be GC'ed here.
	}
	t.add(m) // 将当前的 m 添加到 itabTable 当中.
}
```

`iterate_itabs` 就是一个迭代拷贝. `add`的逻辑和`find`的逻辑是十分类似的.
了.


还有一个函数, 就是 `itab` 的初始化函数 `init`, 该函数只在新添加 itab 的时候调用, 稍微有点复杂, 下面看看吧.

```cgo
// init用 m.inter/m._type 对的所有代码指针填充 m.fun 数组. 
// 如果该类型未实现该接口, 将 m.fun[0]设置为0, 并返回缺少的接口函数的名称.
// 可以在同一 m 上多次调用此函数, 甚至可以同时调用.
func (m *itab) init() string {
    // 注: 调用此函数的时候, inter, _type, hash 已经赋值.
	inter := m.inter
	typ := m._type
	x := typ.uncommon() // 根据 kind 生成相应的类型.

	// both inter and typ have method sorted by name,
	// and interface names are unique,
	// so can iterate over both in lock step;
	// the loop is O(ni+nt) not O(ni*nt).
	ni := len(inter.mhdr)
	nt := int(x.mcount)
	
	// 接口方法(用于绑定对应于值当中方法)
	methods := (*[1 << 16]unsafe.Pointer)(unsafe.Pointer(&m.fun[0]))[:ni:ni] 
	// 值方法
	xmhdr := (*[1 << 16]method)(add(unsafe.Pointer(x), uintptr(x.moff)))[:nt:nt] 
	var fun0 unsafe.Pointer
	j := 0
	
imethods:
	for k := 0; k < ni; k++ {
	    // 获取接口当中方法 i, itype, iname, ipkg
		i := &inter.mhdr[k]
		itype := inter.typ.typeOff(i.ityp) // 解析接口方法类型
		name := inter.typ.nameOff(i.name)  // 解析接口方法名称
		iname := name.name()
		ipkg := name.pkgPath()
		if ipkg == "" {
			ipkg = inter.pkgpath.name()
		}
		for ; j < nt; j++ {
		    // 获取值当中的方法 t, ttype, tname, tpkg
			t := &xmhdr[j]
			tname := typ.nameOff(t.name)
			if typ.typeOff(t.mtyp) == itype && tname.name() == iname {
				pkgPath := tname.pkgPath()
				if pkgPath == "" {
					pkgPath = typ.nameOff(x.pkgpath).name()
				}
				
				if tname.isExported() || pkgPath == ipkg {
					if m != nil {
						ifn := typ.textOff(t.ifn)
						if k == 0 {
							fun0 = ifn // we'll set m.fun[0] at the end
						} else {
							methods[k] = ifn
						}
					}
					continue imethods
				}
			}
			
		}
		
		// 只要有一个方法不匹配, 则都会走到这里. 最终会失败的. 正常的跳出循环才是匹配成功的.
		m.fun[0] = 0
		return iname
	}
	m.fun[0] = uintptr(fun0)
	m.hash = typ.hash
	return ""
}
```

断言函数:

```
// eface => iface
func assertE2I(inter *interfacetype, e eface) (r iface) 
func assertE2I2(inter *interfacetype, e eface) (r iface, b bool)

// iface => iface
func assertI2I(inter *interfacetype, i iface) (r iface）
func assertI2I2(inter *interfacetype, i iface) (r iface, b bool)
```

转换函数:

```
// 内置类型转换
func convT16(val uint16) (x unsafe.Pointer) 
func convT32(val uint32) (x unsafe.Pointer)
func convT64(val uint64) (x unsafe.Pointer)
func convTstring(val string) (x unsafe.Pointer)
func convTslice(val []byte) (x unsafe.Pointer)

// _type => eface
func convT2E(t *_type, elem unsafe.Pointer) (e eface)
func convT2Enoptr(t *_type, elem unsafe.Pointer) (e eface)

// itab => iface
func convT2I(tab *itab, elem unsafe.Pointer) (i iface)
func convT2Inoptr(tab *itab, elem unsafe.Pointer) (i iface)

// interfacetype => iface
func convI2I(inter *interfacetype, i iface) (r iface)
```

总结:

getitab 函数的目的在缓存的 `itabTable` 当中查找一个合适的 `itab`, 如果没有查找到, 则向 `itabTable` 当中添加一个新
的 itab. `itabTable` 就是一个哈希表, 采用的是 `二次探测法` 来解决哈希冲突的, 并且哈希表的装载因子是 75%, 超过这个值
就会发生扩容.

还有就是 getitab 的参数 `canfail`, 在不同状况(断言, 接口转换)下 canfail 的参数是固定: 

- 带参数的断言(`runtime.assertI2I2` 和 `runtime.assertE2I2`) 值是 true
- 不带参数的断言 (`runtime.assertI2I` 和 `runtime.assertE2I`), 值是 false 
- 接口转换(`runtime.convI2I`) 值是 false

也就意味着, 当不带参数断言和接口转换失败之后, 程序会 panic 的. 这个在开发过程中一定要慎重使用.


参考:

[interface 10 问](https://mp.weixin.qq.com/s/EbxkBokYBajkCR-MazL0ZA)
