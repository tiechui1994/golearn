package main

/*
#include "test.h"
#include <stdlib.h>
#include <stdio.h>
typedef void(*cb)(void);
#cgo LDFLAGS: -L. -lstdc++ -ltest
*/
import "C"

import (
	"fmt"
)

func main() {
	Sigsetup2()

	for i := 0; i < 3; i++ {
		SafeCall(C.cb(C.test_crash2))
		SafeCall(C.cb(C.test_safe2))
	}

	select {}
}

//export Gopanic
func Gopanic() {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println(x)
		}
	}()

	var a *int
	fmt.Println(*a)
}
