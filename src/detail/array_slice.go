package detail

import "fmt"

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
