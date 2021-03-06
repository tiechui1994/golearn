package main

/*
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

typedef struct option {
    int iarg;
    float farg;
    const char* carg;
    int* iptr;
} option;


int call(const option arg1, const option* arg2, unsigned int arg2len) {
    printf("arg1 iarg: %d\n", arg1.iarg);
    printf("arg1 farg: %0.2f\n", arg1.farg);
    printf("arg1 carg: %s\n", arg1.carg);
    printf("arg1 iptr: %d\n", *arg1.iptr);

	printf("%ld\n", sizeof(option));

	printf("\n==========================\n\n");

	for (unsigned int i=0; i<arg2len; i++) {
		printf("arg2 iarg: %d\n", (arg2[i]).iarg);
		printf("arg2 farg: %0.2f\n", (arg2[i]).farg);
		printf("arg2 carg: %s\n", (arg2[i]).carg);
		printf("arg2 iptr: %d\n", *(arg2[i]).iptr);
		printf("\n--------------------\n\n");
    }

    return 0;
}

char arr[10]={'h','e','l','l','o'};
char *s = "Hello";
*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"
)

func main() {
	type option struct {
		iarg C.int
		farg C.float
		carg *C.char
		iptr *C.int
	}

	val := 100
	opt := option{
		iarg: C.int(10),
		farg: C.float(100.00),
		carg: C.CString("Hello World"),
		iptr: (*C.int)(unsafe.Pointer(&val)),
	}

	arg1 := *(*C.struct_option)(unsafe.Pointer(&opt))

	N := 2
	size := N * int(unsafe.Sizeof(option{}))
	arg2 := (*C.struct_option)(C.malloc(C.size_t(size)))
	arg2ptr := (*[1024]C.struct_option)(unsafe.Pointer(arg2))[:N:N]

	fmt.Println("size", size, "len", len(arg2ptr), size)

	arg2ptr[0] = *(*C.struct_option)(unsafe.Pointer(&opt))
	arg2ptr[1] = *(*C.struct_option)(unsafe.Pointer(&opt))

	var res C.int
	res = C.call(arg1, arg2, C.uint(N))

	fmt.Println(res)

	// 通过 reflect.SliceHeader 转换
	var arr []byte
	array := (*reflect.SliceHeader)(unsafe.Pointer(&arr))
	array.Data = uintptr(unsafe.Pointer(&C.arr[0]))
	array.Len = 10
	array.Cap = 10

	// 切片
	arr1 := (*[31]byte)(unsafe.Pointer(&C.arr[0]))[:10:10]

	// 通过 reflect.StringHeader 转换
	var s string
	str := (*reflect.StringHeader)(unsafe.Pointer(&s))
	str.Data = uintptr(unsafe.Pointer(C.s))
	str.Len = int(C.strlen(C.s))

	// 切片
	length := int(C.strlen(C.s))
	s1 := string((*[31]byte)(unsafe.Pointer(C.s))[:length:length])

	fmt.Println("arr:", string(arr), "arr1:", string(arr1), "s:", s, "s1:", s1)
}
