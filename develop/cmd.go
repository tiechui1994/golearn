package main

/*
#include <stdio.h>
#include <getopt.h>
*/
import "C"
import (
	"fmt"
	"os"
	"strings"
	"unsafe"
)

type arg = int

const (
	no_argument       arg = 0
	required_argument arg = 1
	optional_argument arg = 2
)

type option struct {
	name    *C.char
	has_arg C.int
	flag    *C.int
	val     C.int
}

/*
使用 C 语言 getopt_long 函数解析命令行参数
 */

func main() {
	// t 表示选项 t 没有参数. 合法格式: -t
	// t: 表示选项 t 必须有参数. 合法格式: -t 100 或 -t100 或 --tlong 100 或 --tlong=100
	// t:: 表示选项 t 可以有参数, 也可以没有参数. 合法格式: -t100 或 --tlong=100
	// 上面的 "--tlong" 是 "-t" 的长参数
	// 上面的三种类型分别对应 has_arg 的值是0, 1, 2
	optstring := "t:c::h"

	var method int
	options := [][]interface{}{
		{"time", required_argument, 0, 't'},
		{"help", no_argument, 0, 'h'},
		{"client", optional_argument, 0, 'c'},
		{"trace", required_argument, &method, int32(2)},
		{0, 0, 0, int32(0)},
	}

	N := len(options)
	size := N * int(unsafe.Sizeof(option{}))
	opts := (*C.struct_option)(C.malloc(C.size_t(size)))
	optsptr := (*[1024]C.struct_option)(unsafe.Pointer(opts))[:N:N]
	for i := range options {
		arg1 := options[i][1].(arg)

		arg3 := options[i][3].(int32)
		opt := option{
			has_arg: C.int(arg1),
			val:     C.int(arg3),
		}

		if arg2, ok := options[i][2].(*int); ok && arg2 != nil {
			opt.flag = (*C.int)(unsafe.Pointer(arg2))
		}

		if arg0, ok := options[i][0].(string); ok {
			opt.name = C.CString(arg0)
		}

		optsptr[i] = *(*C.struct_option)(unsafe.Pointer(&opt))
	}

	N = len(os.Args)
	size = N * 8
	argv := (**C.char)(C.malloc(C.size_t(size)))
	argvptr := (*[1024]*C.char)(unsafe.Pointer(argv))[:N:N]
	for i := range os.Args {
		argvptr[i] = C.CString(os.Args[i])
	}

	argc := C.int(len(os.Args))

	fmt.Println("args", strings.Join(os.Args, "|"))

	var idx int32
	for {
		var c C.int
		c = C.getopt_long(
			argc,
			argv,
			C.CString(optstring),
			opts,
			(*C.int)(unsafe.Pointer(&idx)),
		)

		if c > 0 {
			fmt.Println("index", idx)
			fmt.Println("args", C.GoString(C.optarg))
		} else if c < 0 {
			break
		}
	}
}
