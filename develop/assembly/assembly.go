package main

import (
	"golearn/develop/assembly/pkg"
	"fmt"
)

func F(ok bool, a, b int) int {
	if ok {
		println(a)
		return a
	}

	println(b)
	return b
}

func main() {
	pkg.If(false, 10, 20)
	fmt.Println(pkg.Sum(10))
	pkg.Add(10, 20)
}
