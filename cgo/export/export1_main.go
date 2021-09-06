package main

/*
#cgo CFLAGS:-I.
#cgo LDFLAGS:-L. -lgomodule -Wl,-rpath -Wl,.

#include "libgomodule.h"
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

func main() {
	fmt.Println(C.Add(C.int(1), C.int(111)))

	var ch = make(chan string)
	go func() {
		for x := range ch {
			fmt.Println(x)
		}
		time.Sleep(1000)
	}()
	C.Chan(*(*C.GoChan)(unsafe.Pointer(&ch)))
	time.Sleep(1000)
}
