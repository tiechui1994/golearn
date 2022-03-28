## Go 导出库

### 导出的 Go 函数满足的条件

1. 需要引入 "C" 包, 这样导出包才能生成 `libxxx.h` 文件.

2. 必须在 `main` 包当中导出.

3. 必须包含 `main()` 函数.

4. `//export Function` 导出函数标记. 注: `//` 和 `export` 之间没有任何空格. 导出的 `Function` 名称和 Go 函数
名称必须一致.

5. 导出的 Go 函数参数或返回值当中不能包含自定义的struct(不包含 string, int, slice, array, map, chan 的别名). 
对于自定义的结构体, 可以使用 interface{} 或 unsafe.Pointer 替换.

6. 可以导出struct的方法, 但是这些 struct 必须是 string, int, slice, array, map, chan 的别名. 即:

```go 
type String string
type Int int
type Slice []int
type Array [10]int
type Map map[string]string
type Chan chan struct{}
```

除此之外, 其他的结构体方法不能导出.

> 关于 Go 导出函数的看法:
1. 导出的函数建议使用 C 类型的参数, 这样方便外部函数进行调用.
2. 如果导出函数的类型是 Go 内置的类型(int,float,map,interface,slice,chan), `函数导出` 和 `函数调用` 要分离, 否
则会存在类型未定义的错误. 对于 string 类型, 建议最好使用 *C.char 类型进行替换, 否则存在一些问题.
3. 在导出的函数中, 不建议将(map,interface,chan)类型作为参数进行传递, 因为它们会失去 Go 本身自带的一些属性, 导致一些
异常的行为发生.


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

go build -ldflags='-buildmode=c-shared' -o libxxx.so xxx.go
```

静态库:

```bash
go build -buildmode=c-archive -o libxxx.a xxx.go

go build -ldflags='-buildmode=c-archive' -o libxxx.a xxx.go
```

### 常用的调用模式 

- Go => C => Go => C 模式

案例:

//add.h
```cgo
extern int add(int i, int j);
```

//add.c
```cgo
int add(int i, int j) {
    return i+j;
}
```

//export.go
```cgo
package main

/*
#cgo CFLAGS:-I.
#cgo LDFLAGS:-l add -L. -Wl,-rpath -Wl,.

#include "add.h"
*/
import "C"

//export Add
func Add(a, b C.int) C.int {
	return C.add(a, b)
}

func main() {}
```

//main.go
```cgo
package main

/*
#cgo CFLAGS:-I.
#cgo LDFLAGS:-l goadd -L. -Wl,-rpath -Wl,.

#include "libgoadd.h"
*/
import "C"
import "fmt"

// Go => C => Go => C
func main() {
	fmt.Println(C.Add(1, 7))
}
```

编译:

```
gcc -shared -o libadd.so add.c
go build -o libgoadd.so -buildmode=c-shared export.go
go build -o main main.go
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

### 常用的内置函数实现

// CString() 和 GoString()

```cgo
func _Cfunc_CString(s string) *_Ctype_char {
    p := _cgo_cmalloc(uint64(len(s)+1)) // 调用 C 的 malloc 进行内存分配, 在 stdlib.h 当中
    pp := (*[1<<30]byte)(p) // 转换为 byte 数组的一个指针.
    copy(pp[:], s) // 内容拷贝
    pp[len(s)] = 0 // 设置尾端为 NULL, 这是 C 当中对于字符串的定义.
    return (*_Ctype_char)(p)
}

//go:cgo_import_static _cgo_212663114200_Cfunc__Cmalloc
//go:linkname __cgofn__cgo_212663114200_Cfunc__Cmalloc _cgo_212663114200_Cfunc__Cmalloc
var __cgofn__cgo_212663114200_Cfunc__Cmalloc byte
var _cgo_212663114200_Cfunc__Cmalloc = unsafe.Pointer(&__cgofn__cgo_212663114200_Cfunc__Cmalloc)

//go:cgo_unsafe_args
func _cgo_cmalloc(p0 uint64) (r1 unsafe.Pointer) {
    // 使用 runtime.cgocall() 调用 malloc 函数
    _cgo_runtime_cgocall(_cgo_212663114200_Cfunc__Cmalloc, uintptr(unsafe.Pointer(&p0)))
    if r1 == nil {
        runtime_throw("runtime: C malloc failed")
    }
    return
}

//go:linkname _cgo_runtime_gostring runtime.gostring
func _cgo_runtime_gostring(*_Ctype_char) string

func _Cfunc_GoString(p *_Ctype_char) string {
    // 直接调用 runtime.gostring() 进行转换
    return _cgo_runtime_gostring(p)
}

// runtime.gostring
func gostring(p *byte) string {
	l := findnull(p) // 查找 NULL 的位置
	if l == 0 {
		return ""
	}
	s, b := rawstring(l) // 在 go 内存上分配一个 l 长度的内存空间
	memmove(unsafe.Pointer(&b[0]), unsafe.Pointer(p), uintptr(l)) // 内存拷贝
	return s
}
```


`_Cfunc_CString` 是 cgo 内置的 `Go String` 到 `C char*` 类型转换函数:

- 使用 `_cgo_cmalloc` 在 C 空间申请内存

- 使用该段 C 内存转换成一个 `*[1<<30]byte` 对象

- 将 string 拷贝到 byte 数组当中.

- 最后是返回 C 内存地址.


与 `_Cfunc_CString` 对于的是 `_Cfunc_GoString`, 将`C char*`转换为`Go String`, 它是直接调用了 runtime.gostring() 
函数, 转换过程与 `_Cfunc_CString` 类似.

C.CString() 函数简单安全, 但是它涉及了一次 Go 到 C 空间的内存拷贝, 对于长字符串这将是不可忽视的开销.

Go 当中 string 类型是 "不可变的", 实际当中可以发现, 除了常量字符串会在编译期被分配到只读段, 其他动态生成的字符串实际上
都是在堆上.

因此, 如果能够获取到 string 的内存缓存区地址, 就可以使用类似数组的方式将字符串指针和长度直接传递给 C 使用. 这其实就使用
到了 string 底层数据结构:

```cgo
type StringHeader struct{
    Data unitptr
    Len int
}
```

### Go => C 原理

// `export/export2_go_c.go`
```cgo
package main

/*
int sum(int a, int b) {
	return a+b;
}
*/
import "C"

func main() {
	println(C.sum(11, 12))
}
```

按照上面的 CGO 编译过程, 会生成 `test.cgo1.go`, `test.cgo2.c`, `_cgo_gotypes.go`, `_cgo_export.c`, `_cgo_export.h`
`cgo_flags` 文件. 这里重点介绍 `test.cgo2.c` 和 `_cgo_gotypes.go`, 它们是核心.

在 `test.cgo2.c` 当中, 这里只有一个方法 `_cgo_9a439e687ff9_Cfunc_sum`, 它直接进行了 `sum()` (C方法) 的调用.

// `test.cgo2.c`
```cgo

// 声明了 _cgo_topofstack 函数
extern char* _cgo_topofstack(void);

void _cgo_9a439e687ff9_Cfunc_sum(void *v)
{
        struct {
                int p0;
                int p1;
                int r;
                char __pad12[4];
        } __attribute__((__packed__, __gcc_struct__)) *_cgo_a = v;
        char *_cgo_stktop = _cgo_topofstack(); // 获取栈顶
        __typeof__(_cgo_a->r) _cgo_r;
        _cgo_tsan_acquire();
        _cgo_r = sum(_cgo_a->p0, _cgo_a->p1); // 调用 sum 函数
        _cgo_tsan_release();
        _cgo_a = (void*)((char*)_cgo_a + (_cgo_topofstack() - _cgo_stktop));
        _cgo_a->r = _cgo_r;
        _cgo_msan_write(&_cgo_a->r, sizeof(_cgo_a->r));
}
```

上述的 `_cgo_9a439e687ff9_Cfunc_sum` 函數是一个通用的模板实例, 所有的 Go 调用 C 都会有这样的一个模板实例. 模板函数
做的工作:

- 准备调用 C 的参数, 该参数包含了三部分, 函数参数, 函数返回值, padding.


// `_cgo_gotypes.c`

```cgo
//go:linkname _Cgo_always_false runtime.cgoAlwaysFalse
var _Cgo_always_false bool // 一个常亮, 正常状况下值永远是 false

//go:linkname _Cgo_use runtime.cgoUse
func _Cgo_use(interface{}) // 抛出一个 Error

type _Ctype_int int32    // int
type _Ctype_void [0]byte // void 


//go:linkname _cgo_runtime_cgocall runtime.cgocall
func _cgo_runtime_cgocall(unsafe.Pointer, uintptr) int32

//go:linkname _cgo_runtime_cgocallback runtime.cgocallback
func _cgo_runtime_cgocallback(unsafe.Pointer, unsafe.Pointer, uintptr, uintptr)


//go:linkname _cgoCheckPointer runtime.cgoCheckPointer
func _cgoCheckPointer(interface{}, interface{}) // 检查传入C的指针,防止传入了指向Go指针的Go指针

//go:linkname _cgoCheckResult runtime.cgoCheckResult
func _cgoCheckResult(interface{}) // 检查返回值,防止返回一个Go指针.

//go:cgo_import_static _cgo_9a439e687ff9_Cfunc_sum
//go:linkname __cgofn__cgo_9a439e687ff9_Cfunc_sum _cgo_9a439e687ff9_Cfunc_sum
var __cgofn__cgo_9a439e687ff9_Cfunc_sum byte
var _cgo_9a439e687ff9_Cfunc_sum = unsafe.Pointer(&__cgofn__cgo_9a439e687ff9_Cfunc_sum)

//go:cgo_unsafe_args
func _Cfunc_sum(p0 _Ctype_int, p1 _Ctype_int) (r1 _Ctype_int) {
    _cgo_runtime_cgocall(_cgo_9a439e687ff9_Cfunc_sum, uintptr(unsafe.Pointer(&p0)))
    if _Cgo_always_false {
        _Cgo_use(p0) // 逃逸 p1
        _Cgo_use(p1) // 逃逸 p2
    }
    return
}
```

Go 当中变量可以分配在栈上或堆上. 栈中变量的地址随着 go 调度, 发生变化. 堆当中的变量则不会.

当程序进入 C 空间后, 会脱离 Go 的调度机制, 所以必须保证 C 函数的参数分配在堆上. Go 通过编译器里做逃逸分析来决定一个对象
放在栈上还是堆上, 不逃逸的对象放在栈上, 可能逃逸的放在堆上. 

`_Cgo_use` 以 interface 类型为参数, 编译器很难在编译器知道, 变量最好会是什么类型, 因此它的参数都会被分配在堆上. 这是
在 `_Cfunc_sum` 当中调用 `_Cgo_use` 的原因.

`_cgo_runtime_cgocall` 是从 Go 调用 C 的关键函数, 该函数通过 `go:linkname` 从 runtime.cgocall 链接而来.

// runtime/cgocall.go

```cgo
// Call from Go to C.
//
// 这里使用 nosplit, 是因为它用于某些平台上的系统调用. 系统调用可能在栈上有 untyped 参数, 因此 grow 或 scan
// 只能都是不安全的. 
//
//go:nosplit
func cgocall(fn, arg unsafe.Pointer) int32 {
	if !iscgo && GOOS != "solaris" && GOOS != "illumos" && GOOS != "windows" {
		throw("cgocall unavailable")
	}
    
    // 函数检查
	if fn == nil {
		throw("cgocall nil")
	}

	if raceenabled {
		racereleasemerge(unsafe.Pointer(&racecgosync))
	}

	mp := getg().m // 获取当前的 m
	mp.ncgocall++  // 统计 m 调用 CGO 的次数
	mp.ncgo++      // 周期内调用的次数

	// Reset traceback.
	mp.cgoCallers[0] = 0 // 如果在 cgo 中 crash. 记录 CGO 的 traceback

	// 进入系统调用, 以便调度程序创建新的 M 来执行 G
	// 对于 asmcgocall 的调用保证不会增加 stack 并且不会分配内存, 因此在超出 $GOMAXPROCS
	// 之外的 "系统调用中" 中调用是安全的.
	// fn 可能会会掉到 Go 代码中, 这种情况下, 将退出 "system call", 运行 Go 代码(可能会增加堆栈),
	// 然后重新进入 "system call". PC和SP在这里被保存.
	entersyscall() // 进入系统调用前期准备工作. M, P 分离, 防止系统调用阻塞 P 的调度, 保存上下文.

	// 告诉异步抢占我们正在进入外部代码. 在 entersyscall 之后这样做, 因为这可能会阻塞并导致异步抢占失败,
	// 但此时同步抢占会成功(尽管这不是正确性的问题)
	osPreemptExtEnter(mp) // 在 linux 当中是空函数

	mp.incgo = true
	errno := asmcgocall(fn, arg) // 切换到 g0, 调用 C 函数 fn, 汇编实现

	// Update accounting before exitsyscall because exitsyscall may
	// reschedule us on to a different M.
	mp.incgo = false
	mp.ncgo--

	osPreemptExtExit(mp) // 在 linux 当中是空函数

	exitsyscall() // 退出系统调用, 寻找 P 来绑定 M

	// Note that raceacquire must be called only after exitsyscall has
	// wired this M to a P.
	if raceenabled {
		raceacquire(unsafe.Pointer(&racecgosync))
	}
    
    // 防止 Go 的 gc, 在 C 函数执行期间回收相关参数, 用法与前述_Cgo_use类似
	KeepAlive(fn) 
	KeepAlive(arg)
	KeepAlive(mp)

	return errno
}
```

Go 调入 C 之后, 程序的运行将不受 Go 的 runtime 的管控. 一个正常的 Go 函数是需要 runtime 的管控的, 即函数的运行时间
过长导致 goroutine 的抢占, 以及 GC 的执行会导致所有的 goroutine 被挂起.

C 程序的执行, 限制了 Go 的 runtime 的调度行为. 为此, Go 的 runtime 会在进入 C 程序之前, 标记这个运行 C 的线程 M,
将其排除在调度之外.

由于正常的 Go 程序运行在一个 2k 的栈上, 而 C 程序需要一个无穷大的栈, 因此在进入 C 函数之前需要把当前线程的栈从 2k 切换
到线程本身的系统栈上, 即切换到 g0.

asmcgocall 采用汇编实现:

// runtime/asm_amd64.s

```cgo
// func asmcgocall(fn, arg unsafe.Pointer) int32
// fn 是函数地址, arg 是第一个参数地址
// 在 g0 上调用 fn(arg) 函数.
TEXT ·asmcgocall(SB),NOSPLIT,$0-20
	MOVQ	fn+0(FP), AX
	MOVQ	arg+8(FP), BX

	MOVQ	SP, DX // 保存当前的 SP 到 DX

	// Figure out if we need to switch to m->g0 stack.
	// We get called to create new OS threads too, and those
	// come in on the m->g0 stack already.
	// 切换 g 之前的检查
	get_tls(CX)
	MOVQ	g(CX), R8 // R8 = g
	CMPQ	R8, $0    // g == 0
	JEQ	nosave // 相等跳转, 则说明当前 g 为空
	MOVQ	g_m(R8), R8 // 当前 m
	MOVQ	m_g0(R8), SI // SI = m.g0
	MOVQ	g(CX), DI    // DI = g  
	CMPQ	SI, DI // m.g0 == g
	JEQ	nosave // 相等跳转, 当前在 g0 上
	MOVQ	m_gsignal(R8), SI // SI = m.gsignal
	CMPQ	SI, DI // m.gsignal == g
	JEQ	nosave // 相等跳转, 当前 m.gsignal 上

	// 切换到 g0 上
	MOVQ	m_g0(R8), SI // SI=m.g0
	CALL	gosave<>(SB) // 调用 gosave, 参数是 gobuf
	MOVQ	SI, g(CX) // 切换到 g0
	MOVQ	(g_sched+gobuf_sp)(SI), SP // 恢复 g0 的 SP 

	// Now on a scheduling stack (a pthread-created stack).
	// Make sure we have enough room for 4 stack-backed fast-call
	// registers as per windows amd64 calling convention.
	SUBQ	$64, SP     // SP=SP-64
	ANDQ	$~15, SP	// SP=SP+16, 偏移 gcc ABI
	MOVQ	DI, 48(SP)	// 保存 g 
	MOVQ	(g_stack+stack_hi)(DI), DI // DI=g.stack.hi
	SUBQ	DX, DI       // 计算 g 栈大小, 保存到 DI 当中
	MOVQ	DI, 40(SP)	// 保存 g 栈大小(这里不能保存 SP, 因为在回调时栈可能被拷贝)
	MOVQ	BX, DI		// DI = first argument in AMD64 ABI
	MOVQ	BX, CX		// CX = first argument in Win64
	CALL	AX          // 调用函数, 参数 DI, SI, CX, DX, R8

	// 函数调用完成, 恢复到 g, stack
	get_tls(CX)
	MOVQ	48(SP), DI // DI=g
	MOVQ	(g_stack+stack_hi)(DI), SI // SI=g.stack.hi
	SUBQ	40(SP), SI // SI=SI-size
	MOVQ	DI, g(CX)  // tls 保存, 恢复到 g 
	MOVQ	SI, SP     // 恢复 SP

	MOVL	AX, ret+16(FP) // 函数返回错误码
	RET

nosave:
	// 在系统栈上运行, 甚至可能没有 g.
    // 在线程创建或线程拆除期间可能没有 g 发生(例如, 参见 Solaris 上的 needm/dropm).
    // 这段代码和上面的代码作用是一样的, 但没有saving/restoring g, 并且不用担心 stack 移动(因为我们在系统栈上,
    // 而不是在 goroutine 堆栈上).
    // 如果上面的代码已经在系统栈上, 则可以直接使用, 但是通过此代码的唯一路径在 Solaris 上很少见.
	SUBQ	$64, SP
	ANDQ	$~15, SP
	MOVQ	$0, 48(SP)	// where above code stores g, in case someone looks during debugging
	MOVQ	DX, 40(SP)	// save original stack pointer
	MOVQ	BX, DI		// DI = first argument in AMD64 ABI
	MOVQ	BX, CX		// CX = first argument in Win64
	CALL	AX
	MOVQ	40(SP), SI	// restore original stack pointer
	MOVQ	SI, SP
	MOVL	AX, ret+16(FP)
	RET
```


**当Go调用C函数时, 会单独占用一个系统线程. 因此如果在 Go协程中并发调用C函数, 而C函数中又存在阻塞操作,就很可能会造成Go
程序不停的创建新的系统线程,而Go并不会回收系统线程,过多的线程会拖垮整个系统**



下面介绍一下将 C 符号导入 Go:

```cgo
//go:cgo_import_static _cgo_9a439e687ff9_Cfunc_sum
//go:linkname __cgofn__cgo_9a439e687ff9_Cfunc_sum _cgo_9a439e687ff9_Cfunc_sum
var __cgofn__cgo_9a439e687ff9_Cfunc_sum byte
var _cgo_9a439e687ff9_Cfunc_sum = unsafe.Pointer(&__cgofn__cgo_9a439e687ff9_Cfunc_sum)
```

- `go:cgo_import_static` 将 C 函数的 `_cgo_9a439e687ff9_Cfunc_sum` 加载到 Go 空间中.

- `go:linkname`, 将 Go 的 byte 对象 `__cgofn__cgo_9a439e687ff9_Cfunc_sum` 的内存地址链接到 C 函数的 `_cgo_9a439e687ff9_Cfunc_sum`
内存空间

- 创建 Go 对象的 `_cgo_9a439e687ff9_Cfunc_sum` 并赋值 C 函数地址.

