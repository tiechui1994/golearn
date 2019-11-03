package detail

import "fmt"

// 字面量在初始化切片的时候,可以指定索引, 没有指定的索引的元素会在前一个索引基础上加1
// 如果切边当中,开始位置没有指定索引, 则默认是0. 切片当中的索引不能重复(隐式和显式)
func ArrayInit() {
	var x = []int{2: 3, 2, 0: 1}
	fmt.Println(x) // [1 0 2 3]

	//var _ = []int{1, 2, 1: 3} // 编译失败,duplicate index in array literal: 0
	//var _ = []int{1, 2, 2: 3} // 编译成功
}

func MapAndPointer() {
	type Param map[string]interface{}

	type Show struct {
		*Param
	}

	// s := new(Show)    // 只是创建一个Show的指针, 实例的*Param为nil, [只有通过make创建map,然后引用赋值有效]
	// s.Param["A"] = 100 // 指针不支持引用, s.Param是一个指针, 无法进行引用
	// (*s.Param)["A"] = 100 // 合法
}
