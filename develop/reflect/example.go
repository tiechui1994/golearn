package main

import "fmt"

type Person interface {
	grow()
}

type Student struct {
	age int
	name string
}

func (p Student) grow() {
	p.age += 1
	return
}

func main() {
	var qcrao = Person(Student{age: 18, name:"san"})

	fmt.Println(qcrao)
}
