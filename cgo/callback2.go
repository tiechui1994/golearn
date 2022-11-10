package main

/*
#include <stdio.h>

typedef int (*callback)(int*, int);

int invoke(int (*fun)(int*, int), int* arg1, int arg2) {
	printf("args: %p, %d\n", arg1, arg2);
	return fun(arg1, arg2);
}

int exportCfun(int* args, int l) {
	int ans = 0, i = 0;
	for (i=0; i<l; i++) {
		ans += *(args+i);
	}

	return ans;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func main() {
	// 回调函数的参数, 相当于 C 客户端
	args := []C.int{1, 3, 4, 5, 8, 12, 34, 818, 1219}
	argsp := (*C.int)(unsafe.Pointer(&args[0]))

	// 方式一: 将函数指针转换为 *[0]byte 类型
	ans := C.invoke((*[0]byte)(unsafe.Pointer(C.exportCfun)), argsp, C.int(len(args)))
	fmt.Println(ans)

	// 方式二: 增加函数指针的方式
	ans = C.invoke(C.callback(C.exportCfun), argsp, C.int(len(args)))
	fmt.Println(ans)
}
