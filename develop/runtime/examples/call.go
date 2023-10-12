package main

import "fmt"

func sum(a, b int) int {
	a2 := a * a
	b2 := b * b
	c := a2 + b2

	return c
}

const (
	Build = "main"
)

func main() {
	sum(1, 2)
	fmt.Println(uintptr(0xfffffffffffff001))
	fmt.Println(1<<0)
}