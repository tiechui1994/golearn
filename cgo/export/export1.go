package main

import "C"
import "fmt"

type Int int

// maybe build error
//export String
func (i Int) String() string {
	return fmt.Sprintf("%d", i)
}

//export Inc
func (i *Int) Inc(j int) {
	*i = *i + Int(j)
}

//export Add
func Add(i, j C.int) C.int {
	return i+j
}

// in no buffer chan, maybe success
//export Chan
func Chan(c chan string)  {
	c <- "AA2"
	c <- "BB3"
}

func main() {
}
