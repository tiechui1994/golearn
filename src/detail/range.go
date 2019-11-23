package detail

import "fmt"

/*
golang range 揭秘:

	range 支持的数据类型的底层结构:
		array	数组(或者数组指针)
		string	一个结构体:拥有一个变量 len 和一个指针指向背后的数组
		slice	一个结构体:拥有一个变量 len 一个变量 cap 和一个指针指向背后的数组
		map	    指向一个结构体的指针
		channel	指向一个结构体的指针

	复制操作:
		// 复制整个数组
		var a [10]int
		acopy := a

		// 只复制了 slice 的结构体, 并没有复制成员指针指向的数组
		s := make([]int, 10)
		scopy := s

		// 只复制了 map 的指针
		m := make(map[string]int)
		mcopy := m

	如果要在range循环开始前把一个数组表达式赋值给一个变量(保证表达式只evaluate一次),
	就会复制整个数组.

	range 编译后的伪代码:
		for INIT; COND; POST {
			ITER_INIT
			INDEX = INDEX_TEMP
			VALUE = VALUE_TEMP
			origin statements
		}
	也就是说, range的内部实现是C风格for循环, 编译器会针对每一种range支持的类型做专门
	的处理还原.
        数组:
		len_temp := len(range)
		range_temp := range
		for index_temp = 0; index_temp < len_temp; index_temp++ {
			value_temp = range_temp[index_temp]
			index = index_temp
			value = value_temp
			original body
		}

		切片:
		len_temp := len(for_temp)
		range_temp := range
		for index_temp = 0; index_temp < len_temp; index_temp++ {
			value_temp = range_temp[index_temp]
			index = index_temp
			value = value_temp
			original body
		}

	结论:
		1. 如果要在range循环开始前把一个表达式赋值给一个变量(保证表达式只evaluate一次),
		循环过程中的index, value始终是同一个变量(地址不发生改变)

		2. 可以在迭代过程中移除一个map元素, 或者向map里添加元素. 添加的元素并不一定在后续
		迭代中遍历到

*/

// For循环当中, 循环变量只会创建一次, 后面的更新是指针方式的更新. 千万不要操作循环变量的指针
func ForRange() {
	// 只会创建一次
	loops := []int{1, 2}
	for i, v := range loops {
		fmt.Printf("%+v, %+v\n", &i, &v)
	}

	fmt.Println("====================")

	mp := map[string]string{"A": "1", "B": "2"}
	for i, v := range mp {
		fmt.Printf("%+v, %+v\n", &i, &v)
	}

	fmt.Println("====================")

	ints := []*int{new(int), new(int)}
	for i, v := range ints {
		fmt.Printf("%+v, %+v\n", &i, &v)
	}
}

func MapRange() {
	m := make(map[string]string)
	m["1"] = "1"
	m["2"] = "2"

	for k, v := range m { // 只是复制了指针
		fmt.Println(k, v)
		m[k+v] = k + v
	}

	fmt.Println(len(m)) // 结果不确定
}
