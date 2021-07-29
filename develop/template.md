## Go 模板

### pipeline

在 `{{}}` 內的操作将其称为 pipeline. 

pipeline 通常是将一个 command 序列分割开, 再使用管道符`|`连接起来. (不使用管道符的command序列视为一个管道).

变量:

```
{{ $name := pipeline }}
```

1. **一个变量的作用域直到声明它的控制结构(`if`, `with`, `range`) 的 "end" 为止.** 如果不是在控制结构里声明会直到模板
结束为止. 子模板的调用不会从调用它的位置(作用域)继承变量.

2. **有一个特殊的边框 `$`, 它代表模板的最顶级作用域对象(模板全局作用域的根对象). 在 Execute() 时候进行赋值, 且一直不变.**

循环:

```
{{ range $index, $element := pipeline }} 
```

> 其中 $index 和 $element 分别设置为数组/切片的索引或map的键, 以及对应成员元素. 注: 这与 go range 当中只有一个参数时
设置为索引/键不同!


在一个链式的 pipeline 里, 每个 command 的结果都作为下一个 command 的最后一个参数. pipeline 最后一个 command 的输出
作为整个管道执行的结果.

command 的输出可以是1到2个值, 如果是2个, 后一个必须是 error 接口类型. 如果 error 类型返回值非 nil, 模板执行会终止并将
错误返回给执行模板的调用者.


下面是一个 action 列表:

```
{{ /* a comment */ }}, 注释, 执行时会被忽略. 可以是多行. 


{{ pipeline }},  pipeline 的值的默认文本会拷贝到输出结果.


{{ if pipeline }} T1 {{ end }}, 如果 pipeline 的值为 empty, 不产生输出, 否则输出T1执行结果. 不改变 dot 的值. 
empty 值包括 false, 0, nil指针或接口, 长度为0的数组, 切片, 字典.

{{ if pipeline }} T1 {{ else }} T0 {{ end }}

{{ if pipeline }} T1 {{ else if pipeline }} T0 {{ end }}


{{ range  pipeline }} T1 {{ end }}, pipeline的值必须是array, slice, map, 或chan. 

{{ range  pipeline }} T1 {{ else }} T0 {{ end }}, 如果 pipeline 的长度为0. 不改变 dot 的值, 并执行 T0


{{ with pipeline }} T1 {{ end }}, 如果 pipeline 为 empty 不产生输出, 否则将 dot 设为 pipeline 的值并执行 T1.
不修改外部的 dot 的值.

{{ with pipeline }} T1 {{ else }} T0 {{ end }}


{{ template "name" }}, 执行名为 name 的模板, 提供给模板的参数为 nil, 如果模板不存在输出 ""

{{ template "name" pipeline }}, 执行名为 name 的模板, 提供给模板的参数为 pipeline 的值
```


### 函数

template 内置的函数. 

```
and, 返回第一个为空的参数或最后一个参数. 可以有任意多个参数.
     and x y <=> if x then y else x

not, 布尔取反. 只能是一个参数.

or, 返回第一个不为空的参数或最后一个参数. 可以有任意多个参数.
    or x y <=> if x then x else y


print
printf
println, 分别等价于 fmt 当中的 Sprint, Sprintf, Sprintln


len, 返回参数的长度.


index, 对可索引对象进行索引取值. 第一个参数是索引对象, 后面的参数是索引位. 
       index x 1 2 3 代表 x[1][2][3]. 可索引对象包括 map, slice, array
```

常用的比较函数:

```
eq arg1 arg2 <=> arg1 == arg2

ne arg1 arg2 <=> arg1 != arg2

lt arg1 arg2 <=> arg1 < arg2

gt arg1 arg2 <=> arg1 > arg2
```


案例1: 比较
```
{{ if (gt $x 33) }}
{{ println $x}}
{{ end }}
```

案例2: 函数
```
{{ $x := (len $) }}
{{ println $x }}
```

案例3: 文件列表
```
total {{ (len $) }}
{{- range $ }}
{{- $x := "f" }}
{{- if (eq .Type 1) }} 
	{{- $x = "d" }}
{{- end }}
{{printf "%s  %24s  %s" $x .Updated .Name}}
{{- end }}
```

> 在 `{{}}` 当中添加 `-` 表示去掉空格(包括换行). 

上述的输出的效果:
```
total 3
d  2021-04-24T07:36:48.591Z  111
f  2021-04-24T07:36:48.591Z  222
f  2021-04-24T07:36:48.591Z  333333
```