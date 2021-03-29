package main

import (
	"fmt"
	"unsafe"
)

const (
	iterator     = 1 // 可能有迭代器在使用buckets
	oldIterator  = 2 // 可能有迭代器在使用oldbuckets
	hashWriting  = 4 // 有协程正在向map写人key
	sameSizeGrow = 8 // 等量扩容
)

func main() {
	Pointer()
}

func Num() {
	var x int
	fmt.Println(unsafe.Sizeof(x))

	var y uintptr
	fmt.Println(unsafe.Sizeof(y))

	fmt.Println(4 << (^uintptr(0) >> 63))

	fmt.Println(^uintptr(0), ^uint(0), ^uint64(0))

	fmt.Println(uintptr(1<<(8*8)-1), uintptr(1<<(8*4)-1))

	fmt.Println(uintptr(1) << (63 & (8*8 - 1)))
}

func BitOp() {
	// x | 1 << i 将第 i 位设置为 1
	// x & ^(1<<i) 将第 i 位设置为 0

	x := 110
	fmt.Printf("%0.7b, %0.7b\n", x, x|1<<4)
	fmt.Printf("%0.7b, %0.7b\n", x, x&^(1<<3))

	// x ^ y 异或
	fmt.Printf("  x:%0.7b\n", 10, )
	fmt.Printf("  y:%0.7b\n", 20)
	fmt.Printf("x^y:%0.7b\n", 10^20)
	var p struct {
		A string
		B [2]int64
	}
	fmt.Println(unsafe.Alignof(p))
	fmt.Println(unsafe.Sizeof(p))
}

type bmap struct {
	tophash  [8]uint
	overflow *bmap
}

//go:nosplit
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}
func (b *bmap) setoverflow(ovf *bmap) {
	*(**bmap)(add(unsafe.Pointer(b), uintptr(16)-8)) = ovf
}
func Pointer() {
	b := &bmap{}
	b.setoverflow(&bmap{})
}
