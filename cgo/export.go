package main

/*
static int c_add(int a, int b) {
	return a+b;
}

static int go_add_proxy(int a, int b) {
	extern int GoAdd(int a, int b);
	return GoAdd(a,b);
}
*/
import "C"
import "fmt"

// 文档: https://chai2010.cn/gopherchina2018-cgo-talk/#/7/11

//export GoAdd
func GoAdd(a, b C.int) C.int {
	return C.c_add(a, b)
}

// Go => C => Go => C
func main1() {
	fmt.Println(C.go_add_proxy(1, 7))
}

// Go => C
func main2() {
	fmt.Println(C.c_add(1, 10))
}

func main() {
	main1()
	main2()
}
