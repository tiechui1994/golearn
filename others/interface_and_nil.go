package main

import "fmt"

type Human interface {
	Say() string
}

type Man struct {
}

func (m *Man) Say() string {
	return "man"
}

func IsNil(h interface{}) bool {
	return h == nil
}

// golang当中 == 比较, 前提是 _type 是一致的, 可以比较的类型
// 数组(大小,类型), 结构体[同一类型结构体(结构体名,结构体内容), 匿名结构体(结构体内容)],
// map, channel, pointer, interface, func

// interface相等的条件,(_type,itab一样 + 字面值一样)

// golang == nil
// interface, iface(itab为nil + data为nil) eface(_type为nil + data为nil)
// pointer, channel, slice, map, func 只是声明未初始化

func main() {
	var a1 [10]int
	var b1 [10]int
	fmt.Println(a1 == b1)

	type X struct {
		A string
	}
	type Y X

	var x *int
	var y *int
	fmt.Println(x == y)

	var aa struct {
		A string
	}
	var bb struct {
		A string
	}

	fmt.Println(aa == bb)

	var a interface{}
	var b *Man
	var c *Man
	var d Human
	var e interface{}
	a = b
	e = a

	// a eface+*Man
	// b *Man
	// c *Man        --> h: eface+*Man
	// d iface+nil   --> h: eface+nil
	// e eface+*Man
	fmt.Println(b == c, a == e, a == b)
	fmt.Println("a==nil", a == nil)
	// (1) false
	// a是eface类型，_type指向的是*Man类型，
	// data指向的是nil，所以此题为false

	fmt.Println("e==nil", e == nil)
	// (2) false
	// 同理，e为eface类型，_type也是指向的*Man类型

	fmt.Println("a==c", a == c)
	// (3) true
	// a的_type是*Man类型，data是nil
	// c的data也是nil

	fmt.Println("a==d", a == d)
	// (4) false
	// a为eface类型，d为iface类型，而且d的itab指向的是nil，data也是nil
	// 因为d没有具体到哪种数据类型

	fmt.Println("c==d", c == d)
	// (5) false
	// c和d其实是两种不同的数据类型

	fmt.Println("e==b", e == b)
	// (6) true
	// 分析见(4)

	fmt.Println(IsNil(c))
	// (7) false
	// c是*Man类型，以参数的形式传入IsNil方法
	// 虽然c指向的是nil，但是参数i的_type指向的是*Man，所以i不为nil

	fmt.Println(IsNil(d))
	// (8) true
	// d没有指定具体的类型，所以d的itab指向的是nil，data也是nil

	fmt.Println("+++++++++++++++++++++++++++++++++")

	var m1 Man
	var m2 Man

	var h1 Human = &m1
	var h2 interface{} = m2
	fmt.Println(h1 == h2)
}
