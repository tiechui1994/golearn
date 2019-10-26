package main

func escapes(x interface{}) {
	if dummy.b {
		dummy.x = x
	}
}

var dummy struct {
	b bool
	x interface{}
}

func Update() {
	var x int
	escapes(x)
}

func main() {

}
