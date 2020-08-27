package main

// 使用外部库 (动态库或者静态库)

/*
#cgo CFLAGS: -I ${SRCDIR}/lib
#cgo LDFLAGS: -L ${SRCDIR} -l math
#include "math.h"
*/
import "C"

import "fmt"

func main() {
	value := C.add(C.int(1), C.int(2))
	fmt.Println(value)

	value = C.sub(C.int(10), C.int(5))
	fmt.Println(value)
}
