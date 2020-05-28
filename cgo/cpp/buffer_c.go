package main

/*
#cgo CXXFLAGS: -std=c++11

#include "buffer_c.h"
*/
import "C"

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
