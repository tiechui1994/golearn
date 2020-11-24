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
    inter  *interfacetype
    _type  *_type
    link   *itab
    hash   unit32
    bad    bool
    inhash bool
    unused [2]byte
    fun    [1]unitptr
}
```

其中 itab 由具体类型 `_type` 以及 `interfacetype` 组成. `_type` 表示具体类型, `interfacetype` 表示具体类型实
现的接口类型. 

![image](/images/develop_reflect_iface.jpeg)

实际上, iface 描述的是非空接口, 它包含方法; 与之相对的是 eface, 描述的是空接口, 不包含任何方法. Go 语言里所有的类型都
"实现了" 空接口.

```cgo
type eface struct {
    _type *_type
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