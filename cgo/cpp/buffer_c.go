package main

/*
#cgo CXXFLAGS: -std=c++11 -I .
#cgo LDFLAGS: -L . -l buffer -l stdc++

#include "buffer_c.h"
*/
import "C"

/**
这里采用了静态库链接的方式, 链接 libbuffer.a 文件

在使用 gcc 编译包装了 C++ 库的时候, 一定要链接 "stdc++" 库.
**/

type cgo_Buffer_T C.Buffer_T

func cgo_NewBuffer(size int) *cgo_Buffer_T {
	p := C.NewBuffer(C.int(size))
	return (*cgo_Buffer_T)(p)
}

func cgo_DeleteBuffer(p *cgo_Buffer_T) {
	C.DeleteBuffer((*C.Buffer_T)(p))
}

func cgo_Buffer_Data(p *cgo_Buffer_T) *C.char {
	return C.Buffer_Data((*C.Buffer_T)(p))
}

func cgo_Buffer_Size(p *cgo_Buffer_T) C.int {
	return C.Buffer_Size((*C.Buffer_T)(p))
}
