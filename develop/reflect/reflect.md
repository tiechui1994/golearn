# Go 反射原理

反射是如何实现的?

答案是 `interface`, 它是 Go 语言实现抽象的一个非常强大的工具. 当像接口变量赋予一个实体类型的时候, 接口会存储实体的类型
信息, 反射就是通过接口的类型信息实现的, 反射建立在类型的基础上的,

Go 在 reflect 包里定义了各种类型, 时序反射的各种函数, 通过它们可以在运行时检测类型的信息, 改变类型的值.


types 与 interface

Go 当中, 每个变量都有一个静态类型, 在编译阶段就确定了的. 注意: 这个类型是声明时候的类型, 不是底层数据类型.

举个例子:

```cgo
type MyInt int 

var i int
var j MyInt
```

尽管 i, j的底层类型都是 int, 但是, 他们是不同的静态类型, 除非类型转换, 否则, i和j 不能同时出现在等号两侧. j 的静态类
型的 MyInt. 

反射主要和 `interface{}` 类型相关. 其底层数据结构:

```cgo
type iface struct {
    tab *itab
    data unsafe.Pointer
}

type itab struct {
    inter  *interfacetype // 静态类型
    _type  *_type         // 动态类型
    hash   unit32         // copy _type.hash
    _     [4]byte
    fun    [1]unitptr     // method table
}

type interfacetype struct {
	typ     _type
	pkgpath name
	mhdr    []imethod
}


type _type struct {
	size       uintptr
	ptrdata    uintptr // 包含所有指针的内存前缀的大小. 如果为0, 表示的是一个值, 而非指针
	hash       uint32
	tflag      tflag
	align      uint8
	fieldalign uint8
	kind       uint8
	alg        *typeAlg
	// gcdata存储垃圾回收器的GC类型数据.
    // 如果 KindGCProg 位设置为 kind, 则gcdata是GC程序. 否则为 ptrmask 位图. 
    // 有关详细信息, 请参见 mbitmap.go.
	gcdata    *byte
	str       nameOff
	ptrToThis typeOff
}
```

其中 itab 由具体类型 `_type` 以及 `interfacetype` 组成. `_type` 表示具体类型, `interfacetype` 表示具体类型实
现的接口类型. 

![image](/images/develop_reflect_iface.jpeg)

实际上, iface 描述的是非空接口, 它包含方法; 与之相对的是 eface, 描述的是空接口, 不包含任何方法. Go 语言里所有的类型都
"实现了" 空接口.

```cgo
type eface struct {
    _type *_type        // 动态类型
    data unsafe.Pointer
}
```

与 `iface` 相比, `eface` 就比较简单了. 只维护了一个 `_type` 字段, 表示空接口所承载的具体的实体类型. `data` 描述了
具体的值.

![image](/images/develop_reflect_eface.jpeg)

举个例子:

> **接口变量可以存储任何实现了接口定义的所有方法的变量**

Go 语言中最常见的就是 `Reader` 和 `Writer` 接口:

```cgo
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}
```

接口之间的各种转换和赋值:

```cgo
var r io.Reader
tty, err := os.OpenFile("~/Desktop/test", os.O_RDWR, 0)
if err != nil {
    return nil, err
}

r = tty
```

说明: 首先声明 `r` 的类型是 `io.Reader`, 注意, 这是 r 的静态类型, 此时它的动态类型为 `nil`, 并且它的动态值也是 `nil`.

之后, `r = tty`, 将 `r` 的动态类型变成 `*os.File`, 动态值变为 tty (非空), 表示打开的文件对象. 此时, r 可以用 
`<value, type>` 对来表示为: `<tty, *os.File>`


![image](/images/develop_reflect_reader.png)

> 注意: 此时虽然 `fun` 所指向的函数只要一个 `Read` 函数, 其实 `*os.File` 还包含 `Write` 函数. 也就是说. *os.File
其实还实现了 `io.Writer` 接口. 因此下面的断言语句可以执行:

```
var w io.Writer
w = r.(io.Writer)
```

之所以使用断言, 而不是直接赋值, 是因为 `r` 的静态类型是 `io.Reader`, 并没有实现 `io.Writer` 接口. 断言能否成功,
看 `r` 的动态类型是否符合要求.

这样, w 也可以表示成 `<tty, *os.File>`, 尽管它和 `r` 是一样的, 但是 w 可调用的函数取决于它的静态类型 `io.Writer`,
也就是说它只能有这样的调用形式: `w.Write()`.

![image](/images/develop_reflect_writer.png)


最后, 再来一个赋值:

```cgo
var empty interface{}
empty = w
```

由于 `empty` 是一个空接口, 因此所有的类型都实现了它, `w` 可以直接赋值给它, 不需要执行断言操作.

![image](/images/develop_reflect_empty.png)


从上面的三张图看出, interface 包含三部分信息: `_type` 是类型信息, `data` 指向实际类型的实际值, `itab` 包含实际类型
的信息, 包含大小, 包路径, 还包含绑定在类型上各种方法. 补充一些 `*os.File` 结构体的图:

![image](/images/develop_reflect_empty_detail.png)

参考案例 `reflect.go` 当中的代码.


## 反射的基本函数

`reflect` 包里定义了一个接口和一个结构体, 即 `reflect.Type` 和 `reflect.Value`, 它们提供很多函数来获取存储接口里
的类型信息.

`reflect.Type` 主要提供关于类型相关的信息, 所以它和 `_type` 关联比较紧密;

`reflect.Value` 则结合 `_type` 和 `data` 两者, 因此程序可获取甚至改变类型的值.

`reflect.Typeof` 函数用于提取一个接口中值的类型信息.  


`Type()`, `Interface()` 可以打通 `interface`, `Type`, `Value` 三者.

![image](/images/develop_reflect_convert.png) 

小结: `TypeOf()` 返回一个接口, 这个接口定义了一系列的方法, 利用这些方法可以获取关于类型的所有信息; `ValueOf()` 返回
一个结构体变量, 包含类型信息以及实际的值.

![image](/images/develop_reflect_sum.png)

> `rtype` 实现了 `Type` 接口, 是所有类型的公共部分, `emptyface` 结构体和 `eface` 其实是一个东西, 而 `rtype` 和
`_type` 也是一个东西, 只是一些字段上稍微有点差别. 


## 反射三大定律

1. Reflection goes from interface value to reflection object.
2. Reflection goes from reflection object to interface value.
3. To modify a reflection object, the value must be settable.

- 第一条: 反射是一种检测存储在 `interface` 中的类型和值机制. 这可以通过 `TypeOf()` 和 `ValueOf()` 得到.

- 第二条和第一条是相反的机制, 它将 `Value` 通过 `Interface()` 函数反向转变成 `interface` 变量

- 第三条, 如果需要修改一个反射变量的值, 那么它必须是可设置的. 反射变量可设置的本质是它存储了原变量本身, 这样对反射变量的
操作, 就会反映到原变量本身; 反之, 如果反射变量不能代表原变量, 那么操作了反射变量, 不会对原变量产生任何影响.