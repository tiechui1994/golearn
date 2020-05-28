package iface

type Dulk interface {
	Walk()
	Quack(name string)
}

type Cat struct {
	Name string
}

func (c Cat) Walk() {

}
func (c Cat) Quack(name string) {
	c.Name = name
}

type Pig struct {
	Name string
}

func (c *Pig) Walk() {

}
func (c *Pig) Quack(name string) {
	c.Name = name
}
