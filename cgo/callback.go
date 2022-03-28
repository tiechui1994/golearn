package main

import "C"

/*
#include <stdio.h>

static int foo(int (*fun)(int*, int), int* arg1, int arg2) {
	printf("args: %p, %d\n", arg1, arg2);
	return fun(arg1, arg2);
}

extern int xyz(int*, int);

typedef int (*call)(int*, int);
*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"
)

//export xyz
func xyz(x *C.int, n C.int) C.int {
	sh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(x)),
		Len:  int(n),
		Cap:  int(n),
	}

	values := *(*[]int32)(unsafe.Pointer(&sh))

	var sum int32
	for i := 0; i < len(values); i++ {
		sum += values[i]
	}
	return C.int(sum)
}

func main() {
	// 回调函数的参数
	args := []C.int{1, 3, 4, 5, 8, 12, 34, 818, 1219}
	argsp := (*C.int)(unsafe.Pointer(&args[0]))

	// 方式一: 将函数指针转换为 *[0]byte 类型
	ans := C.foo((*[0]byte)(unsafe.Pointer(C.xyz)), argsp, C.int(len(args)))
	fmt.Println(ans)

	// 方式二: 增加函数指针的方式
	ans = C.foo(C.call(C.xyz), argsp, C.int(len(args)))
	fmt.Println(ans)
}
