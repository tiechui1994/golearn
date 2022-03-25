package main

/*
#include "test.h"
#include <stdlib.h>
#cgo LDFLAGS: -L. -lstdc++ -ltest
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"
)

func main() {
	Sigsetup()

	gopanic()
	gopanic()

	fmt.Println("++++++++++++++++++++++")

	time.Sleep(1 * time.Second)
	str := "From Golang"
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	C.test_crash2()
	C.test_crash2()
	select {}
}

func gopanic() {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println(x)
		}
	}()
	var a *int
	fmt.Println(*a)
}
