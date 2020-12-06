package main

import "golearn/develop/assembly/pkg"

func F(ok bool, a, b int) int {
	if ok {
		println(a)
		return a
	}

	println(b)
	return b
}

func main() {
	pkg.If(0, 10, 20)
}
