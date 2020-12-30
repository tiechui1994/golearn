package pkg

import "fmt"

func Print1(a int) int {
	fmt.Println(a)
	return a
}

func Print2(a, b int) int {
	fmt.Println(a, b)
	return a + b
}

func Print3(a, b, c int) int {
	fmt.Println(a, b, c)
	return 0
}
