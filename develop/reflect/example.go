package main

import (
	"fmt"
	"io"
	"bytes"
	"math/rand"
	"time"
)

type Person interface {
	grow()
}

type Student struct {
	Age  int    `json:"age"`
	Name string `json:"name"`
}

func (p Student) grow() {
	p.Age += 1
	return
}

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
	hash()

	// 如果接口的类型一致, 那么接口转换是不产生额外的操作的
	// var x Interface
	// _ = Interface(x)
	// 其中 Interface 是 iface
	// 接口转换
	v1 := Person(Student{Age: 108, Name: "abc"}) // runtime.convT2I
	val := io.ByteScanner(&bytes.Buffer{})
	v6 := io.ByteReader(val) // runtime.convI2I
	fmt.Println(v1, v6)

	// 如果接口的值类型和断言的类型一致, 则不会产生额外的操作
	// var x Interface = T{}
	// _, _ = x.(T)
	// 其中 Interface 可以是 eface, 也可以是iface
	// 接口断言
	var eface interface{} = Student{}
	v2, _ := eface.(Person) // runtime.assertE2I2
	v3 := eface.(Person)    // runtime.assertE2I

	var iface Person = Student{} // runtime.convT2E
	v4, _ := iface.(Person)      // runtime.assertI2I2
	v5 := iface.(Person)         // runtime.assertI2I

	fmt.Println(v2, v3, v4, v5)
}
