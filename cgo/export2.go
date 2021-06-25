package main

/*
#cgo CFLAGS: -I ./
#cgo LDFLAGS: -shared -L. -lwx -Wl,-rpath -Wl,.

#include "libwx.h"
*/
import "C"

import "fmt"

func main() {
	ans := C.Concat(C.CString("a"), C.CString("b"))
	fmt.Println(C.GoString(ans))
}
