## Go 导出库

- 导出的 Go 函数满足的条件:

1. 必须在 `main` 包当中导出.

2. 必须包含 `main()` 函数.

3. `//export Function` 导出函数标记. 注: `//` 和 `export` 之间没有任何空格. 导出的 `Function` 名称和 Go函数
名称必须一致.

4. 导出的 Go 函数参数或返回值当中不能包含自定义的struct(不包含 string, int, slice, array, map, chan 的别名). 
对于自定义的结构体, 可以使用 interface{} 或 unsafe.Pointer 替换.

5. 可以导出struct的方法, 但是这些 struct 必须是 string, int, slice, array, map, chan 的别名. 即:

```go 
type String string
type Int int
type Slice []int
type Array [10]int
type Map map[string]string
type Chan chan struct{}
```

除此之外, 其他的结构体方法不能导出.

案例1: 导出 C 类型参数函数

```cgo
package main

import "C"

//export Concat
func Concat(a, b *C.char) *C.char {
	return C.CString(C.GoString(a) + C.GoString(b))
}

func main() {}
```

案例2: 导出 Go 类型方法

```cgo 
package main

//export Add
func Add(i, j int) int {
	return i + j
}

func main() {}
```

案例3: 导出 struct 方法

```cgo
package main

import "fmt"

type Int int

//export String
func (i Int) String() string {
	return fmt.Sprintf("%d", i)
}

//export Inc
func (i *Int) Inc(j int) {
	*i = *i + Int(j)
}

func main() {}
```

- 导出命令:

动态库:

```bash
go build -buildmode=c-shared -o libxxx.so xxx.go

go build -ldflags='-buildmode=c-shared' -o libxxx.a xxx.go
```

静态库:

```bash
go build -buildmode=c-archive -o libxxx.a xxx.go

go build -ldflags='-buildmode=c-archive' -o libxxx.a xxx.go
```

