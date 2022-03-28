package main

// #include <stdio.h>
import "C"
import "unsafe"

type Buffer struct {
	cptr *cgo_Buffer_T
}

func NewBuffer(size int) *Buffer {
	return &Buffer{
		cptr: cgo_NewBuffer(size),
	}
}

func (p *Buffer) Delete() {
	cgo_DeleteBuffer(p.cptr)
}

func (p *Buffer) Data() []byte {
	data := cgo_Buffer_Data(p.cptr)
	size := cgo_Buffer_Size(p.cptr)
	return ((*[1 << 31]byte)(unsafe.Pointer(data)))[0:int(size):int(size)]
}

func main() {
	buf := NewBuffer(1024)
	defer buf.Delete()

	copy(buf.Data(), []byte("Hello World. \x00"))
	C.puts((*C.char)(unsafe.Pointer(&(buf.Data()[0]))))
}
