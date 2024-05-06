package main

func main() {
	var a int
	defer func(i int) {
		println("defer=", i)
	}(a)

	println("main=", a)

	defer func() {
		println(1)
	}()
}
