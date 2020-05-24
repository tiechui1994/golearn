package main

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
