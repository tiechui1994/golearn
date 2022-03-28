package main

/*
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

typedef struct sub {
	int* p;
} sub;

typedef struct main {
    int  a;
    sub* p;
} main;

void plus(main* f) {
	printf("before: %d, %d\n", f->a, *((f->p)->p));
    f->a+=1;
    *((f->p)->p)+=1;
    printf("after: %d, %d\n", f->a, *((f->p)->p));
}

typedef struct xyz {
	int  a;
	int* p;
} xyz;

void pplus(xyz* f) {
	printf("before: %d, %d\n", f->a, *(f->p));
    f->a+=1;
    *(f->p)+=1;
    printf("after: %d, %d\n", f->a, *(f->p));
}

*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"
)

// test parent memory
func case1()  {
	// malloc 
	size := int(unsafe.Sizeof(C.struct_main{}))
	args := (*C.struct_main)(unsafe.Pointer(C.malloc(C.size_t(size))))

	// Go
	p := (*[10]C.struct_main)(unsafe.Pointer(args))[:1:1]
	sub := C.struct_sub{
		p: (*C.int)(unsafe.Pointer(new(int))),
	}
	main := C.struct_main{
		a: 5,
		p: (*C.struct_sub)(unsafe.Pointer(&sub)),
	}
	p[0] = main

	C.plus(args)
	fmt.Println("p[0]", *(p[0].p.p), p[0].a)
	fmt.Println("main", *(main.p.p), main.a)
}

// case2, Resolve parnet memory
func case2()  {
	// malloc
	size := int(unsafe.Sizeof(C.struct_xyz{}))
	args := (*C.struct_xyz)(unsafe.Pointer(C.malloc(C.size_t(size))))

	// Go, Use Array
	p := (*[10]C.struct_xyz)(unsafe.Pointer(args))[:1:1]
	xyz := C.struct_xyz{
		a: 5,
		p: (*C.int)(unsafe.Pointer(new(int))),
	}
	p[0] = xyz

	C.pplus(args)
	fmt.Println("p[0]", *(p[0].p), p[0].a)
	fmt.Println("xyz", *(xyz.p), xyz.a)

	// Go, Use reflect.SliceHeader
	sh := reflect.SliceHeader{
		Data:uintptr(unsafe.Pointer(args)),
		Len:1,
		Cap:1,
	}
	px := *(*[]C.struct_xyz)(unsafe.Pointer(&sh))
	xyz = C.struct_xyz{
		a: 5,
		p: (*C.int)(unsafe.Pointer(new(int))),
	}
	px[0] = xyz
	C.pplus(args)
	fmt.Println("px[0]", *(px[0].p), px[0].a)
	fmt.Println("xyz", *(xyz.p), xyz.a)
}

// case3, Resolve Leaf Memory
func case3()  {
	slice := make([]C.struct_xyz, 1)
	xyz := C.struct_xyz{
		a: 5,
		p: (*C.int)(unsafe.Pointer(C.malloc(C.size_t(4)))),
	}
	slice[0] = xyz

	args := (*C.struct_xyz)(unsafe.Pointer(&slice[0]))
	C.pplus(args)
	fmt.Println("xyz", *(xyz.p), xyz.a)
}

func main() {
	case1()
	case2()
	case3()
}
