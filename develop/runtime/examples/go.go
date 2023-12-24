package main

func f(arg1 int, arg2 int) {
	_ = arg1 + arg2
}

func main() {
	go f(100, 200)
}
