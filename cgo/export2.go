package main

/*
#cgo CFLAGS: -I ./
#cgo LDFLAGS: -l wx -L ./ -Wl,-rpath -Wl,. -shared

#include "libwx.h"
*/
import "C"

import "fmt"

func main() {
	ans := C.Concat(C.CString("a"), C.CString("b"))
	fmt.Println(C.GoString(ans))
}
