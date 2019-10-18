package main

import "fmt"

func main() {
	defer fmt.Println("in main")
	defer func() {
		panic("panic again")
	}()

	panic("panic once")
}
