package main

import "fmt"

func main() {
	var a int
	defer func(i int) {
		println(i)
	}(a)

	fmt.Println("==>", a)
}
