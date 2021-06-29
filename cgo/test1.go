package main

/*
#cgo LDFLAGS:-L. -lwx -Wl,-rpath -Wl,.
#cgo CFLAGS:-I.

extern void Link(char* a, char* b);

#include "libwx.h"
*/
import "C"

//export Link
func Link(a, b *C.char) {
	ans := C.GoString(a) + C.GoString(b)
	_ = C.CString(ans)
}

func main() {
	ans := C.Concat(C.CString("aa"), C.CString("bb"))
	println(ans)

	C.Link(C.CString("aa"), C.CString("xx"))
}
