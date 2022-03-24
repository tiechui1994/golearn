package main

/*
#include "test.h"
#include <stdlib.h>
typedef void(*cb)(void);
#cgo LDFLAGS: -L. -lstdc++ -ltest
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	Sigsetup2()
	//SafeCall(C.cb(C.test_crash2))

	var x func()
	x = gopanic
	SafeCall(C.cb(unsafe.Pointer(&x)))
	select {}
}

func gopanic() {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println(x)
		}
	}()
	//var a *int
	//fmt.Println(*a)
	//C.test_crash2()
	go func() {
		panic("本协程请求出错")
	}()
}
