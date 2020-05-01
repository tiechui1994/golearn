package main

import "fmt"

type Duck interface {
	Quack()
}

type StructCat struct {
	Name string
}

//go:noinline
func (s StructCat) Quack() {
	fmt.Println(s.Name + " meow")
}

type PointerCat struct {
	Name string
}

//go:noinline
func (p *PointerCat) Quack() {
	fmt.Println(p.Name + " meow")
}

// multiple inheritance
// when call Quack() will happen error: `ambiguous selector m.Quack`
type Multiple struct {
	PointerCat
	StructCat
	name string
}

func main() {
	var p Duck = &PointerCat{Name: "Pointer"}
	p.Quack()

	//var ps Duck = PointerCat{Name:"AA"}

	var s Duck = StructCat{Name: "Struct"}
	s.Quack()

	var sp Duck = &StructCat{Name: "Struct Pointer"}
	sp.Quack()

	var _ = Multiple{
		PointerCat: PointerCat{Name: "--------"},
		StructCat:  StructCat{Name: "++++++++"},
		name:       "MutliStruct",
	}

	//m.Quack()
}
