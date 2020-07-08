package main

// #cgo CFLAGS: -I ./
// #cgo LDFLAGS: -l header -L ./ -Wl,-rpath -Wl,./
// #include "header.h"
import "C"

func Hello() {
	C.SayHello()
}

func main() {
	Hello()
}
