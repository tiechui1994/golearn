package main

//go build -buildmode=c-archive -o libwx.a export1.go
//go build -buildmode=c-shared -o libwx.so export1.go

import "C"

//export Concat
func Concat(a, b *C.char) *C.char {
	return C.CString(C.GoString(a) + C.GoString(b))
}

//export Add
func Add(i, j int) int {
	return i + j
}

func main() {}
