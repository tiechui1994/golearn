package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"time"
)

type Human interface {
	Say() string
}

type Man struct{}

func (m *Man) Say() string {
	return "man"
}

func IsNil(h interface{}) bool {
	return h == nil
}

// 可以进行比较的
// array, struct, pointer, string, int, interface, 匿名结构体, channel
//
// struct, interface, channel, 类型相同("type Y=X" , "type Z X", X,Y的类型相同, X,Z类型不一样), 且值相等, 才相等.
// 注: interface 的类型是指的是静态类型.
//
// 匿名结构体, 内存布局一样, 且值相同, 才相等.
//
// interface == nil, 当且仅当动态类型为 nil, 且动态值为nil

// 不可以比较的
// slice, map

func NilTestCase() {

	var a1 [10]int
	var b1 [10]int
	fmt.Println(a1 == b1)

	type X struct {
		A string
	}
	type Y = X

	var xx X
	var yy Y
	fmt.Println(xx == yy)

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

	var x1 chan int
	var x2 chan int
	fmt.Println(x1 == x2)

	var a interface{}
	var b *Man
	var c *Man
	var d Human
	var e interface{}
	a = b // eface (*Man, nil)
	e = a // eface (*Man, nil)

	fmt.Println("a==nil", a == nil)
	// (1) false
	// a是eface类型, _type指向的是*Man类型，
	// data指向的是nil，所以此题为false

	fmt.Println("e==nil", e == nil)
	// (2) false
	// 同理，e为eface类型,_type也是指向的*Man类型

	fmt.Println("a==c", a == c)
	// (3) true
	// a的_type是*Man类型, data是nil
	// c的data也是nil

	fmt.Println("a==d", a == d)
	// (4) false
	// a为eface类型, d为iface类型，而且d的itab指向的是nil, data也是nil
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

//==================================================================================================

type Language interface {
	grow()
}

type Java struct {
	Age  int    `json:"age"`
	Name string `json:"name"`
}

func (p Java) grow() {
	p.Age += 1
	return
}

func ConvertCase() {
	// 如果接口的类型一致, 那么接口转换是不产生额外的操作的
	// var x Interface
	// _ = Interface(x)
	// 其中 Interface 是 iface
	// 接口转换
	v1 := Language(Java{Age: 108, Name: "abc"}) // runtime.convT2I
	val := io.ByteScanner(&bytes.Buffer{})
	v6 := io.ByteReader(val) // runtime.convI2I
	fmt.Println(v1, v6)

	// 如果接口的值类型和断言的类型一致, 则不会产生额外的操作
	// var x Interface = T{}
	// _, _ = x.(T)
	// 其中 Interface 可以是 eface, 也可以是iface
	// 接口断言
	var eface interface{} = Java{}
	v2, _ := eface.(Language) // runtime.assertE2I2
	v3 := eface.(Language)    // runtime.assertE2I

	var iface Language = Java{} // runtime.convT2E
	v4, _ := iface.(Language)      // runtime.assertI2I2
	v5 := iface.(Language)         // runtime.assertI2I

	fmt.Println(v2, v3, v4, v5)
}

//==================================================================================================

// hash - 二次寻址

func hash() {
	rand.Seed(time.Now().UnixNano())
	for k := uintptr(1); k < 68; k++ {
		var mask uintptr = 1<<k - 1
		var h0 = uintptr(rand.Uint64())

		h := h0
		for i := uintptr(1); ; i++ {
			h += i
			h &= mask

			fmt.Println(h, (h0+i*(i+1)/2)&mask)

			if i >= 100 {
				break
			}
		}
	}
}

func main() {

}
