package main

import (
	"fmt"
)

/**
slice vs array

array:
var a [5]int            // [0,0,0,0,0]
var a = [5]int{1,2,3}   // [1,2,3,0,0]
var a = [...]int{1,2,3} // [1,2,3]

slice:
var a []int             // []
var a = make([]int,1)   // [0]
var a = make([]int,1,3) // [0]

slice是一个指针, 值复制时, 是指针的复制, 指向的底层数据不做修改. (无法进行比较操作)
array是一个结构体, 值复制时, 是底层的数据的复制, 复制前后数据的指向发生改变. (可以进行比较操作)

range操作进行值复制的操作.

append(), 第一个参数必须是slice.
1) 如果操作前后未发生扩容, 则操作前后的slice的地址不发生改变.
2) 如果操作前后发生了扩容, slice地址发生改变(底层的数据也会发生移动).


func f(a ...int)
函数f当中的参数a是一个slice类型, 参数传递的时候是指针的复制.

**/

func ArrayRange() {
	var a1 = [5]int{1, 3, 4, 5, 6}
	var a2 [5]int

	for i, v := range a1 {
		if i == 0 {
			a1[1] = 11
			a1[2] = 22
		}

		a2[i] = v
	}

	fmt.Println("a1= ", a1) // [1,11,22,5,6]
	fmt.Println("a2= ", a2) // [1,3, 4, 5,6]

	/**
	for i, v := range &a1 {
		if i == 0 {
			a1[1] = 11
			a1[2] = 22
		}

		a2[i] = v
	}
	**/
}

func SliceRange() {
	var a1 = []int{1, 3, 4, 5, 6}
	var a2 [5]int

	for i, v := range a1 {
		if i == 0 {
			a1[1] = 11
			a1[2] = 22
		}

		a2[i] = v
	}

	fmt.Println("a1= ", a1) // [1,11,22,5,6]
	fmt.Println("a2= ", a2) // [1,11,22,5,6]
}

/**
go当中可以进行比较的类型: (在满足下面的情况下, 如果值是一样的, 则相等)

int, byte, bool, string            -- 值的比较
array(数组大小和值类型必须一致)        -- 数组当中的值是否一样
channel(值类型必须一致)              -- 指向底层数据的指针是否是同一个
pointer(值类型必须一致)              -- 指向底层数据的指针是否是同一个

struct
命名结构体: 类型一致(包括类型别名)      --- 值比较
匿名结构体: 成员变量的名称,类型,顺序完全一致  --- 值比较

type A struct{}
type B = A  // B是A的别名, A和B是同一种类型
type C A    // 创建新的结构体C, A和C是两种类型


注: slice是不能进行比较的.
    map是不能进行比较的
**/

func main1() {
	var c0 chan int
	var c1 chan int
	fmt.Println(c0 == c1) // true

	var c2 = make(chan int)
	var c3 = make(chan int)
	fmt.Println(c2 == c3) // false

	var a1 = [5]int{1, 2, 3, 4, 0}
	var a2 = [5]int{1, 2, 3, 4}
	fmt.Println(a1 == a2) // true

	type A struct {
		A string
		B int
	}
	var an1 struct {
		A string
		B int
	}
	var an2 struct {
		A string
		B int
	}
	var an3 A
	fmt.Println(an1 == an2, an1 == an3) // true, true
}

/**
slice 与 array 之间的转换:

array => slice

slice = array[M:N]  M <= N
len: N-M
cap: len(array) - M


detail:
1. 在slice没有达到其"容量"之前, array 与 slice 共享底层数据段 [M:N] 部分, 这部分内容的修改两者都会修改
2. 在slice超过其"容量"之后, array 与 slice 将不再共享底层数据段, 两者修改互不影响.

slice => array
**/

func main2() {
	a := [5]int{1, 2, 3, 4, 5}
	s1 := a[0:2]
	s2 := a[0:4]
	s3 := a[2:2]
	fmt.Println(len(s1), cap(s1))
	fmt.Println(len(s2), cap(s2))
	fmt.Println(len(s3), cap(s3))

	var a1 [5]int
	copy(a1[:], s1)

	fmt.Println("=======================================")

	array := [3]int{1, 2, 3}
	slice := array[1:2] // len:1 cap:2

	slice[0] = 11
	slice = append(slice, 12) // len:2 cap:2
	slice[0] = 100            // 导致slice和array同时修改
	fmt.Println("共享内存", "slice:", slice, "array:", array)

	slice = append(slice, 13) // len:3 cap:4
	slice[0] = 200            // 导致只有slice修改, array不修改
	fmt.Println("不共享内存", "slice:", slice, "array:", array)
}
