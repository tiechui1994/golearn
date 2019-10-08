package main

type Duck interface {
	Quack()
	Call()
}

type Cat struct {
	Name string
}

//go:noinline
func (c Cat) Quack() {
	println(c.Name + " meow")
}

func (c Cat) Call()  {

}

func main() {
	var c Duck = Cat{Name: "grooming"}
	c.Call()
	c.Quack()
}
