package main

/*
#cgo CXXFLAGS: -std=c++11 -I.
#cgo LDFLAGS: -L. -ldate -lstdc++ -Wl,-rpath -Wl,.
#include "bridge.h"
*/
import "C"
import "fmt"

func main() {
	val := C.NewDate(C.int(2022), C.int(11), C.int(22))

	fmt.Printf("before: %d, %d, %d\n", C.getYear(val), C.getMonth(val), C.getDay(val))
	C.SetDate(val, C.int(2022), C.int(11), C.int(11))
	fmt.Printf("after: %d, %d, %d\n", C.getYear(val), C.getMonth(val), C.getDay(val))
}
