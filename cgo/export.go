package main

/*
extern void SayHello(_GoString_ s);
*/
import "C"
import "fmt"

// 文档: https://chai2010.cn/gopherchina2018-cgo-talk/#/7/11
// https://bbs.huaweicloud.com/blogs/117132

//export SayHello
func SayHello(s string) {
	fmt.Println(s)
}

//// Go => C => Go => C
//func main1() {
//	fmt.Println(C.c_add(1, 7))
//}
//
//// Go => C
//func main2() {
//	fmt.Println(C.c_add(1, 10))
//}

//// Go => C => Go
//func main3()  {
//	var x interface{} = "java"
//	C.SayHello(x)
//}

func main() {
	C.SayHello("AA")
}
