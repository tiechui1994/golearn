package main

import (
	"unsafe"
	"io"
	"fmt"
	"os"
)

type iface struct {
	tab  *itab
	data unsafe.Pointer
}

type itab struct {
	inter uintptr
	_type uintptr
	link  uintptr
	hash  uint32
	_     [4]byte
	fun   [1]uintptr
}

type eface struct {
	_type uintptr
	data  unsafe.Pointer
}

func main() {
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
