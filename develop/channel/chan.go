package main

func main() {
	_ = make(chan int64)
	println("+++", uintptr(-int(48)&(8-1)))

	println("---", uintptr(-int(9)&(3)))

	var x chan int
	_ = <-x
}
