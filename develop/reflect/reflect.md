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

