package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"unsafe"
)

type eface struct {
	_type *_type
	data  unsafe.Pointer
}

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
	equal        func(unsafe.Pointer, unsafe.Pointer) bool
	gcdata     *byte
	str        int32
	ptrToThis  int32
}

func escapes(x interface{}) {
	if dummy.b {
		dummy.x = x
	}
}

var dummy struct {
	b bool
	x interface{}
}

type Sayer interface {
	Bye()
	Good()
	Say()
}

type A struct {
	AA string
	BB int
}

func (a *A) Bye() {
}

func (a *A) Say() {
	fmt.Println("Say")
}

func (a *A) Good() {
}

func Test_Interface() {
	var r io.Reader
	fmt.Printf("init r: %T, %v\n", r, r)

	tty, _ := os.OpenFile("/home/user/go/src/golearn/develop/reflect/reflect.md", os.O_RDONLY, 0)
	fmt.Printf("tty: %T, %v\n", tty, tty)

	r = tty
	fmt.Printf("r: %T, %v\n", r, r)

	riface := (*iface)(unsafe.Pointer(&r))
	fmt.Printf("r:iface.tab._type = %#x, iface.data = %#x\n", riface.tab._type, riface.data)

	var w io.Writer
	w = r.(io.Writer)
	fmt.Printf("w: %T, %v\n", w, w)

	wiface := (*iface)(unsafe.Pointer(&w))
	fmt.Printf("w: iface.tab._type = %#x, iface.data = %#x\n", wiface.tab._type, wiface.data)

	var empty interface{}
	empty = w
	fmt.Printf("empty: %T, %v\n", empty, empty)

	eeface := (*eface)(unsafe.Pointer(&empty))
	fmt.Printf("empty: eface._type = %#x, eface.data = %#x\n", eeface._type, eeface.data)
}

func Test_Structure() {
	// ptrdata
	var a interface{} = A{}
	eface := (*eface)(unsafe.Pointer(&a))
	fmt.Println("data:", eface.data)
	if eface._type != nil {
		fmt.Printf("_type: %+v\n", eface._type)
	}

	bb := A{}
	var b Sayer = &bb
	iface := (*iface)(unsafe.Pointer(&b))
	fmt.Println("data:", iface.data)
	if iface.tab != nil {
		fmt.Printf("_type: %+v\n", iface.tab._type)
		fmt.Printf("inter: %+v\n", iface.tab.inter)

		fmt.Printf("hash: %v\n", iface.tab.hash)

		fun := runtime.FuncForPC(iface.tab.fun[0])
		fmt.Printf("func addr: %v\n", iface.tab.fun[0])
		fmt.Printf("func: %v\n", fun.Name())

		fun = runtime.FuncForPC(iface.tab.fun[0] + 16)
		fmt.Printf("func: %v\n", fun.Name())

		fun = runtime.FuncForPC(iface.tab.fun[0] - 16)
		fmt.Printf("func: %v\n", fun.Name())

		v := reflect.ValueOf(&bb)
		vv := v.MethodByName("Say")
		fmt.Printf("func addr: %v \n", v.MethodByName("Say").Call([]reflect.Value{}))

		fmt.Printf("%+v\n", vv)
	}

}

func Test_Kind() {
	type flag uintptr
	type value struct {
		typ *_type
		ptr unsafe.Pointer
		flag
	}

	var buf bytes.Buffer
	t := reflect.ValueOf(buf.Read)
	v := (*value)(unsafe.Pointer(&t))
	fmt.Printf("%b\n", v.flag)

	fmt.Printf("Int      : %06b\n", reflect.Int)
	fmt.Printf("UINT64   : %06b\n", reflect.Uint64)
	fmt.Printf("String   : %06b\n", reflect.String)
	fmt.Printf("Struct   : %06b\n", reflect.Struct)
	fmt.Printf("Slice    : %06b\n", reflect.Slice)

	fmt.Printf("FUNC     : %06b\n", reflect.Func)
	fmt.Printf("PTR      : %06b\n", reflect.Ptr)
	fmt.Printf("Chan     : %06b\n", reflect.Chan)
	fmt.Printf("Map      : %06b\n", reflect.Map)
	fmt.Printf("Interface: %06b\n", reflect.Interface)

	type eface struct {
		typ *_type
		ptr unsafe.Pointer
	}
	var x1 = 10
	var x2 = "aaaa"
	var x3 = fmt.Println
	var x4 = eface{}
	var x5 = &eface{}
	var x6 chan int
	var x7 []string

	var xx interface{} = x1
	vv := (*eface)(unsafe.Pointer(&xx))
	fmt.Printf("%20s : %06b\n", `var x1 = 10`, vv.typ.kind)

	xx = x2
	escapes(xx)
	vv = (*eface)(unsafe.Pointer(&xx))
	fmt.Printf("%20s : %06b\n", `var x2 = "aaaa"`, vv.typ.kind)

	xx = x3
	escapes(xx)
	vv = (*eface)(unsafe.Pointer(&xx))
	fmt.Printf("%20s : %06b\n", `var x3 = fmt.Println`, vv.typ.kind)

	xx = x4
	escapes(xx)
	vv = (*eface)(unsafe.Pointer(&xx))
	fmt.Printf("%20s : %06b\n", `var x4 = eface{}`, vv.typ.kind)

	xx = x5
	escapes(xx)
	vv = (*eface)(unsafe.Pointer(&xx))
	fmt.Printf("%20s : %06b\n", `var x5 = &eface{}`, vv.typ.kind)

	xx = x6
	escapes(xx)
	vv = (*eface)(unsafe.Pointer(&xx))
	fmt.Printf("%20s : %06b\n", `var x6 chan int`, vv.typ.kind)

	xx = x7
	escapes(xx)
	vv = (*eface)(unsafe.Pointer(&xx))
	fmt.Printf("%20s : %06b\n", `var x7 []string`, vv.typ.kind)
}

func Nil(a interface{}) {
	n := reflect.ValueOf(a).Field(0)
	if !n.IsNil() {
		fmt.Printf("%v should be nil\n", a)
	}
}

func NotNil(a interface{}) {
	n := reflect.ValueOf(a).Field(0)
	if n.IsNil() {
		fmt.Printf("value of type %v should not be nil\n", reflect.ValueOf(a).Type().String())
	}
}

func Test_IsNil() {
	// These implement IsNil.
	// Wrap in extra struct to hide interface type.
	doNil := []interface{}{
		struct{ x *int }{},
		struct{ x interface{} }{},
		struct{ x map[string]int }{},
		struct{ x func() bool }{},
		struct{ x chan int }{},
		struct{ x []string }{},
		struct{ x unsafe.Pointer }{},
	}
	for _, ts := range doNil {
		ty := reflect.TypeOf(ts).Field(0).Type
		v := reflect.Zero(ty)
		v.IsNil() // panics if not okay to call
		fmt.Println(v.IsNil())
	}

	// Check the implementations
	var pi struct {
		x *int
	}
	Nil(pi)
	pi.x = new(int)
	NotNil(pi)

	var si struct {
		x []int
	}
	Nil(si)
	si.x = make([]int, 10)
	NotNil(si)

	var ci struct {
		x chan int
	}
	Nil(ci)
	ci.x = make(chan int)
	NotNil(ci)

	var mi struct {
		x map[int]int
	}
	Nil(mi)
	mi.x = make(map[int]int)
	NotNil(mi)

	var ii struct {
		x interface{}
	}
	Nil(ii)
	ii.x = 2
	NotNil(ii)

	var fi struct {
		x func()
	}
	Nil(fi)
	fi.x = Test_IsNil
	NotNil(fi)

	var x interface{
		Say()
	}
	fmt.Println(reflect.ValueOf(x).Kind())
}


func main() {
	Test_Kind()
}
