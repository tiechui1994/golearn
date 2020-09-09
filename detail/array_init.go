package main

import (
	"fmt"
)

/**
字面量在初始化切片的时候, 可以显示指定索引, 没有指定的索引的元素会在前一个索引基础上加1. 例如 {"x","y"}, 那么针对 "y"
的索引就是1. {1:"x", "y"}, 那么此时 "y" 的索引就是2.

如果切片当中, 开始位置没有指定索引, 则默认是0. 切片当中的索引不能重复(隐式推断和显式指定).

索引推断的规则是从左到右, 如果有显示指定, 则使用显示指定, 否则使用隐式推断. 注意: 在这个过程中任何一个位置的索引不能重复.
**/
func ArrayInit() {
	var x = []int{2: 3, 2, 0: 1}
	fmt.Println(x) // [1 0 2 3]

	//var _ = []int{1, 2, 1: 3} // 编译失败, duplicate index in array literal: 0
	//var _ = []int{1, 2, 2: 3} // 编译成功
}

/**
如果 map 的值是指针类型, 那么每次获取 map 的值的时候都是获得一个指向底层值的指针, 在指针的基础上可以直接修改值.

如果 map 的值是非指针类型, 每次获取 map 的值的时候都会将原来的值复制一份, 这样的在修改的时候, 只能是修改其原有值的副本,
一定要记得将修改后的值进行进行重新写入.
**/
func Value() {
	type T struct {
		Val int
	}

	m1 := map[string]T{}
	m1["A"] = T{Val: 100}
	fmt.Printf("before:%+v\n", m1["A"])
	// m1["A"].Val = 200 // cannot assign to struct field m1["A"].Val in map
	val := m1["A"]
	val.Val = 300 // no uesed
	fmt.Printf("after:%+v\n", m1["A"])

	m2 := map[string]*T{}
	m2["B"] = &T{Val: 100}
	fmt.Printf("before:%+v\n", m2["B"])
	m2["B"].Val = 200
	fmt.Printf("after: %+v\n", m2["B"])

	s1 := make([]T, 2)
	s1[0] = T{Val: 20}
	fmt.Printf("before:%+v\n", s1[0])
	s1[0].Val = 200
	fmt.Printf("before:%+v\n", s1[0])

}

func main() {
	Value()
}
