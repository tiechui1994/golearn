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

	//go func() {
	//	fmt.Println(pkg.GetGoid())
	//}()
	//time.Sleep(time.Millisecond)
	//f := pkg.NewTwiceClosure(10)
	//fmt.Println("Closure", f(), f())

	pkg.AsmAdd(11, 22)
	pkg.Add(30, 20)
	rsp, sp, fp := pkg.GetRegister()
	fmt.Println(" FP ", fp)
	fmt.Println("(SP)", rsp)
	fmt.Println(" SP ", sp)
}
