package main

import (
	"fmt"
)

type Person interface {
	grow()
}

type Student struct {
	Age  int    `json:"age"`
	Name string `json:"name"`
}

func (p Student) grow() {
	p.Age += 1
	return
}

func main() {
	var s = Person(Student{Age: 108, Name: "abcdefg"})

	fmt.Println(s)
}
