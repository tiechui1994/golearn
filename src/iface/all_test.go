package iface

import (
	"testing"
	"log"
	"fmt"
)

func TestIface(t *testing.T) {
	var pig Dulk = &Pig{}
	pig.Quack("Pointer-Pointer")
	log.Printf("pig:%+v", pig)
	//var pig1 Dulk = Pig{}
	//pig1.Quack("Pointer-Struct")

	var cat Dulk = &Cat{}
	cat.Quack("Struct-Pointer")
	log.Printf("cat:%+v", cat)

	var cat1 Dulk = Cat{}
	cat1.Quack("Struct-Struct")
	log.Printf("cat1:%+v", cat1)
}
