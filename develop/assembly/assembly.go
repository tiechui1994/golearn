package main

import (
	"golearn/develop/assembly/pkg"
	"fmt"
)

func F() {
	var c = "aa"
	var b int16 = 222
	var a = true
	println(a, b, c)
}

func NewTwiceFunClosure(x int) func() int {
	return func() int {
		x *= 2
		return x
	}
}

func main() {
	//pkg.If(false, 10, 20)
	//fmt.Println(pkg.Sum(10))
	//pkg.Add(10, 20)
	//go func() {
	//	fmt.Println(pkg.GetGoid())
	//}()
	//time.Sleep(time.Millisecond)

	f := pkg.NewTwiceClosure(10)
	fmt.Println(f(), f())
}
