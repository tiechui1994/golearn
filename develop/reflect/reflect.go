package main

import (
	"unsafe"
	"io"
	"fmt"
	"os"
	"runtime"
	"reflect"
	"log"
)

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

type eface struct {
	_type *_type
	data  unsafe.Pointer
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
	log.Println("++++++++++=")
}

func (a *A) Good() {

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

func main() {
	Test_Structure()
}
