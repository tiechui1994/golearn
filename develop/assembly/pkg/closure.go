package pkg

import "unsafe"

//go:noescape
func ptrToFunc(p unsafe.Pointer) func() int

//go:noescape
func asmFunTwiceClosureAddr() uintptr

//go:noescape
func asmFunTwiceClosureBody() int

type TwiceClosure struct {
	F uintptr
	X int
}

func NewTwiceClosure(x int) func() int {
	var p = TwiceClosure{
		F: asmFunTwiceClosureAddr(),
		X: x,
	}

	return ptrToFunc(unsafe.Pointer(&p))
}
