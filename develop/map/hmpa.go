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
	var x int
	fmt.Println(unsafe.Sizeof(x))

	var y uintptr
	fmt.Println(unsafe.Sizeof(y))

	fmt.Println(4 << (^uintptr(0) >> 63))

	fmt.Println(^uintptr(0), ^uint(0), ^uint64(0))

	fmt.Println(uintptr(1<<(8*8)-1), uintptr(1<<(8*4)-1))

	fmt.Println(uintptr(1) << (63 & (8*8 - 1)))

	var p = 5
	if p&hashWriting != 0 {
		fmt.Println("haswrite")
		p ^= hashWriting
		fmt.Println(p)
	}

	if p&hashWriting == 0 {
		fmt.Println("success")
	}
}


