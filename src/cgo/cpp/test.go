package main

// #include <stdio.h>
import "C"
import "unsafe"

func main() {
	buf := NewBuffer(1024)
	defer buf.Delete()

	copy(buf.Data(), []byte("Hello\x00"))
	C.puts((*C.char)(unsafe.Pointer(&(buf.Data()[0]))))
}
