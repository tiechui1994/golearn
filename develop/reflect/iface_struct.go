package main

import "fmt"

func main() {
	var s Person = Student{First: "halfrost"}
	s.sayHello("everyone")
}

type Person interface {
	sayHello(name string) string
	sayGoodbye(name string) string
}

type Student struct {
	First string
}

//go:noinline
func (s Student) sayHello(name string) string {
	return fmt.Sprintf("%v: Hello %v, nice to meet you.\n", s.First, name)
}

//go:noinline
func (s Student) sayGoodbye(name string) string {
	return fmt.Sprintf("%v: Hi %v, see you next time.\n", s.First, name)
}
