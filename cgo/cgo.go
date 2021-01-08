package main

/*
#include <stdio.h>

typedef struct option {
    int iarg;
    float farg;
    const char* carg;
    int* iptr;
} option;

int call(const option arg1, const option* arg2) {
    printf("arg1 iarg: %d\n", arg1.iarg);
    printf("arg1 farg: %0.2f\n", arg1.farg);
    printf("arg1 carg: %s\n", arg1.carg);
    printf("arg1 iptr: %d\n", *arg1.iptr);

	printf("\n==========================\n\n");

    printf("arg2 iarg: %d\n", (*arg2).iarg);
    printf("arg2 farg: %0.2f\n", (*arg2).farg);
    printf("arg2 carg: %s\n", (*arg2).carg);
    printf("arg2 iptr: %d\n", *(*arg2).iptr);

    return 0;
}
*/
import "C"
import (
	"fmt"
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

	size := 1 * int(unsafe.Sizeof(option{}))
	arg2 := (*C.struct_option)(C.malloc(C.size_t(size)))
	arg2ptr := (*[1024]C.struct_option)(unsafe.Pointer(arg2))[:size:size]
	arg2ptr[0] = *(*C.struct_option)(unsafe.Pointer(&opt))

	var res C.int
	res = C.call(arg1, arg2)

	fmt.Println(res)
}
