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

	type bmap struct {
		tophash [8]uint8
	}

	const dataOffset = unsafe.Offsetof(struct {
		b bmap
		v int64
	}{}.v)

	fmt.Println(dataOffset)
	call()
}

func call() {
	fmt.Println(" 0^1", 0^1)
	fmt.Println("0&^1", 0&^1)
	fmt.Println("1&^0", 1&^0)

	fmt.Println(" 0^0", 0^0)
	fmt.Println("0&^0", 0&^0)

	fmt.Println(" 1^1", 1^1)
	fmt.Println("1&^1", 1&^1)

	fmt.Println("0|^0", 0|^uint8(0))
	fmt.Println("0|^1", 0|^1)
	fmt.Println("1|^0", 1|^0)
	fmt.Println("1|^1", 1|^1)

	fmt.Println(1 &^ 1)
	const (
		evacuatedX = 2
		evacuatedY = 3
	)
	if evacuatedX+1 != evacuatedY || evacuatedX^1 != evacuatedY {
		panic("bad evacuatedN")
	}

	x := 0
	for ; x < 10; x++ {
		if x%5 == 0 {
			fmt.Println("----", &x)
			x := 10
			fmt.Println("current", &x)
		}
	}

	var p struct {
		A string
		B [2]int64
	}
	fmt.Println(unsafe.Alignof(p))
	fmt.Println(unsafe.Sizeof(p))
}
