package main

/*
int sum(int a, int b) {
	return a+b;
}
*/
import "C"

func main() {
	// Go => C
	println(C.sum(11, 12))
}
