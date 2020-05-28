package main

/*
#cgo windows CFLAGS: -D CGO_OS_WINDOWS=1
#cgo darwin  CFLAGS: -D CGO_OS_DRWIN=1
#cgo linux CFLAGS: -D CGO_OS_LINUX=1

#if defined(CGO_OS_WINDOWS)
	const char* os = "windows";
#elif defined(CGO_OS_DARWIN)
	const char* os = "darwin";
#elif defined(CGO_OS_LINUX)
	const char* os = "linux";
#else
	error(unknown os)
#endif


#include <stdio.h>
#include <stdint.h>

union B {
    int i;
    float f;
};

union C {
    int8_t i8;
    int64_t i64;
};
*/
import "C"
import (
	"fmt"
	"unsafe"
	"encoding/binary"
)

func unpack() {
	var ub C.union_B
	fmt.Println("b.i:", binary.LittleEndian.Uint32(ub[:]))
	fmt.Println("b.f:", binary.LittleEndian.Uint32(ub[:]))

	fmt.Println("b.i:", *(*C.int)(unsafe.Pointer(&ub)))
	fmt.Println("b.f:", *(*C.float)(unsafe.Pointer(&ub)))
}

func main() {
	unpack()
	return
	fmt.Println(C.GoString(C.os))
	var b C.union_B
	fmt.Printf("%T \n", b)
	fmt.Println("b.i:", *(*C.int)(unsafe.Pointer(&b)))
	fmt.Println("b.f:", *(*C.float)(unsafe.Pointer(&b)))

	var c C.union_C
	fmt.Printf("%T \n", c)
}
