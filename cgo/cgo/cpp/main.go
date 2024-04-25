package main

/*
#cgo CXXFLAGS: -std=c++11 -I .
#cgo LDFLAGS: -L . -lbuffer -lstdc++

#include <stdio.h>
#include "bridge.h"
*/
import "C"
import "unsafe"

/**
这里采用了静态库链接的方式, 链接 libbuffer.a 文件

在使用 gcc 编译包装了 C++ 库的时候, 一定要链接 "stdc++" 库, 这个非常重要.
**/

type Buffer struct {
	ptr *C.Buffer_T
}

func NewBuffer(size int) *Buffer {
	p := C.NewBuffer(C.int(size))

	return &Buffer{
		ptr: (*C.Buffer_T)(p),
	}
}

func (p *Buffer) Delete() {
	C.DeleteBuffer(p.ptr)
}

func (p *Buffer) Data() []byte {
	data := C.Buffer_Data(p.ptr)
	size := C.Buffer_Size(p.ptr)
	return ((*[1 << 31]byte)(unsafe.Pointer(data)))[0:int(size):int(size)]
}

func main() {
	buf := NewBuffer(1024)
	defer buf.Delete()

	copy(buf.Data(), []byte("Hello World. \x00"))
	C.puts((*C.char)(unsafe.Pointer(&(buf.Data()[0]))))
}
