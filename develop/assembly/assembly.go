package main

import (
	"golearn/develop/assembly/pkg"
	"fmt"
)

func NewTwiceFunClosure(x int) func() int {
	return func() int {
		x *= 2
		return x
	}
}

func info(name string) {
	fmt.Printf("\n============= %v =============\n", name)
}

func main() {
	info("declare.go")
	fmt.Println(pkg.INT, pkg.ARRAY, pkg.STRING, pkg.SLICE)

	info("goto.go")
	pkg.If(false, 10, 20)

	info("sum.go")
	fmt.Println(pkg.Sum(10))

	//go func() {
	//	fmt.Println(pkg.GetGoid())
	//}()
	//time.Sleep(time.Millisecond)
	//f := pkg.NewTwiceClosure(10)
	//fmt.Println("Closure", f(), f())

	info("add")
	pkg.AsmAdd(11, 22)
	pkg.Add(30, 20)

	info("register")
	rsp, sp, fp := pkg.GetRegister()
	fmt.Println(" FP ", fp)
	fmt.Println("(SP)", rsp)
	fmt.Println(" SP ", sp)

	info("linkname.go")
	pkg.AddWithLock()
	pkg.AddWithoutLock()
}
