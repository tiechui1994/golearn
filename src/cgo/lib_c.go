package main

/**
#cgo linux CFLAGS: -I ./lib/math
#cgo LDFLAGS: -L ./lib/math -lmath
#include "lib/math.h"
*/
import "C"
import "fmt"

func main() {
	value := C.add(C.int(1), C.int(2))
	fmt.Println(value)

	value = C.sub(C.int(10), C.int(5))
	fmt.Println(value)
}
