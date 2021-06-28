## Go 导出库

### 导出的 Go 函数满足的条件

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

需要做一点补充说明: 对于导出的 Go 函数的参数或返回值是 Go 类型的, 它们在 C 头文件当中对应的类型的关系如下:

| Go 函数参数的类型 | C 头文件定义的类型 |
| --- | --- |
| int | `GoInt` |
| int8 | `GoInt8` |
| float32 | `GoFloat32` |
| unitptr | `GoUintptr` |
| string | `_GoString_` 或 `GoString` |
| map | `GoMap` |
| chan | `GoChan` |
| interface{} | `GoInterface` |
| slice | `GoSlice` |

### 导出命令:

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

### 常用的调用模式 

- Go => C => Go => C 模式

案例:

```cgo
package main

/*
static int c_add(int a, int b) {
	return a+b;
}

static int go_add_proxy(int a, int b) {
	extern int GoAdd(int a, int b);
	return GoAdd(a,b);
}
*/
import "C"
import "fmt"


//export GoAdd
func GoAdd(a, b C.int) C.int {
	return C.c_add(a, b)
}

// Go => C => Go => C
func main() {
	fmt.Println(C.go_add_proxy(1, 7))
}
```


### go 编译过程

go 导入/导出符号表:

> go:cgo_export_dynamic <local> <remote>
> 
> 在 internal 链接模式下, 将名为 <local> 的 Go 符号作为 <remote> 放入程序导出的符号表中. 以便 C 代码可以通过该名称引
用它. 这种机制使得 C 代码可以调用 Go 或共享 Go 的数据.

> go:cgo_export_static <local> <remote>
>
> 在 external 链接模式下. 将名为 <local> 的 Go 符号作为 <remote> 放入程序导出的符号表中. 以便 C 代码可以通过该名称引
用它. 这种机制使得 C 代码可以调用 Go 或共享 Go 的数据.

> //go:cgo_import_static <local>
>
> 在 external 链接模式下, 在为主机链接器准备的 go.o 目标文件中允许对 <local> 的未解析引用, 假设 <local> 将由将与 go.o 
> 链接的其他目标文件提供. 
>
> Example:
> //go:cgo_import_static puts_wrapper

cgo 编译过程: (cgo预处理 -> gcc编译 -> compile编译 -> pack打包 -> link链接)

![image](/images/cgo_call.png)

- cgo 预处理 main.go, 生成 `_cgo_flags`, `main.cgo1.go`, `main.cgo2.c`, `_cgo_gotypes.go`, `_cgo_main.c`
`_cgo_export.c`, `_cgo_export.h`.

> `_cgo_flags`, gcc 编译的参数.
> `_cgo_main.c`, main 函数, cgo 相关函数的空实现.
> `_cgo_export.h` Go导出的函数声明, Go类型的声明.
> `_cgo_export.c`, 对于 `_cgo_export.h` 的"实现"

> `main.cgo2.c`, 内置 cgo 函数的定义, `GoString()`, `GoBytes()`, `CSting()`, `CBytes()` 等. 用户导出函数的
具体真实调用.

> `main.cgo1.go`, 用户 Go 代码部分, 使用 `_Cfunc_xxx` 或 `_Ctype_xxx` 替换了 `C.xxx`(xxx部分是函数或变量). 而
且 `Cfunc_xxx` 或 `Ctype_xxx` 是在 `_cgo_gotypes.go` 当中定义或实现.

> `_cgo_gotypes.go`, CGO 交互的桥梁. 导入(C)函数和导出(Go)函数的入口. 对于导入函数, 使用 `go:cgo_import_static`
将 C 符号表导入到 Go 当中, 以供 Go 使用 runtime.cgocall() 调用 C 函数; 对于导出函数, 使用 `go:cgo_export_dynamic`
和 `go:cgo_export_static` 将 Go 符号表导出, 然后供外部 C 调用 (runtime.cgocallback()).


- gcc 编译, 先将 `_cgo_export.c`, `main.cgo2.c`, `xxx.c`, `_cgo_main.c` 分别编译成 `*.o`, 然后 `*.o` 文件
合并生成 `_cgo_.o`, 其中 `xxx.c` 文件是外部依赖的 C 源文件.

> 注: 如果是库依赖, 则不需要编译 `xxx.c` 文件, 直接依赖库即可.

- cgo 根据 `_cgo_.o` 生成 `_cgo_import.go` 文件(文件当中包含了 cgo 需要导入的系统动态库), 这是 `_cgo_.o` 文件的唯
一用途. 

- compile 编译 `importcfg` 和 `_cgo_gotypes.go`, `main.cgo1.go`, `_cgo_import.go`, 生成 `_pkg_.a` 文件.

> 注: 这里的 importcfg 是 main 函数所直接依赖的包

- pack 打包 `*.o` (这里是 `_cgo_export.c`, `main.cgo2.c`, `xxx.c` 编译的文件) 文件到 `_pkg_.a` 当中(这一步很
关键).

- link 链接 `importcfg.link` 到 `_pkg_.a` 文件, 生成 `a.out`

> 注: importcfg.link = `importcfg 包` + `importcfg依赖包` + `_pkg_.a`

- 对 `a.out` 重命名输出.


---


go 编译过程:

- compile 编译 `importcfg` 生成 `_pkg_.a` 文件.

> 注: 这里的 importcfg 是 main 函数所直接依赖的包

- link 链接 `importcfg.link` 到 `_pkg_.a` 文件, 生成 `a.out`

> 注: importcfg.link = `importcfg 包` + `importcfg依赖包` + `_pkg_.a`

- 对 `a.out` 重命名输出.



cgo 预处理:

```
src="/tmp/cgo"
WORK="/tmp/out"
mkdir -p ${WORK}/b001
cd ${src}
CGO_LDFLAGS='"-g" "-O2"' /opt/share/local/go/pkg/tool/linux_amd64/cgo -objdir $WORK/b001/ -importpath _${src} -- -I $WORK/b001/ -g -O2 ./main.go
```
