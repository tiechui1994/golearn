package pkg

import "fmt"

func Print1(args int) int {
	fmt.Println(args)
	return args
}

func Print2(a, b int) int {
	fmt.Println(a, b)
	return a+b
}
