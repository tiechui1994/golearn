package main

// #cgo CFLAGS: -I ./
// #cgo LDFLAGS: -l header -L ./ -Wl,-rpath -Wl,./
// #include "header.h"
import "C"

func main() {
	_ = C.SayHello(C.CString("abcd"))
}
