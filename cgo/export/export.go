package main


import "C"

//export Jdk
type Jdk struct {

}


//export Add
func Add(i, j C.int) C.int {
	return i+j
}

//export Chan
func Chan(c chan string)  {
	c <- "AA"
	c <- "BB"

	//fmt.Println(reflect.ValueOf(c).Type())
}

func main() {
}
