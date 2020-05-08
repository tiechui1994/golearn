package main

// 外部 C 代码

/*
#include "outer.c"
*/
import "C"
import "fmt"

func main() {
	a := C.int(1)
	b := C.int(2)
	value := C.add(a, b)
	fmt.Println(value)
}
