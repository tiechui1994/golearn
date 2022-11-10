package main

/*
#include <stdio.h>

typedef int (*callback)(int*, int);

static int invoke(int (*fun)(int*, int), int* arg1, int arg2) {
	printf("args: %p, %d\n", arg1, arg2);
	return fun(arg1, arg2);
}

extern int exportGoFun(int*, int);
*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"
)

//export exportGoFun
func exportGoFun(x *C.int, n C.int) C.int {
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
	// 回调函数的参数, 相当于 C 客户端
	args := []C.int{1, 3, 4, 5, 8, 12, 34, 818, 1219}
	argsp := (*C.int)(unsafe.Pointer(&args[0]))

	// 方式一: 将函数指针转换为 *[0]byte 类型
	ans := C.invoke((*[0]byte)(unsafe.Pointer(C.exportGoFun)), argsp, C.int(len(args)))
	fmt.Println(ans)

	// 方式二: 增加函数指针的方式
	ans = C.invoke(C.callback(C.exportGoFun), argsp, C.int(len(args)))
	fmt.Println(ans)
}
