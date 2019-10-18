package main

import (
	"fmt"
	"unsafe"
	"reflect"
	"math/rand"
	"io"
	"os"
)

/*
1. channel 发送的值是 "指针" 或 "其包含指针", 由于在编译阶段无法确定其作用域与传递的路径, 所以一般都会逃逸到
堆上分配(指针部分). 注意, channel本身的底层存储结构没有逃逸.

2. slices 中的值是指针的指针或包含指针字段. 一个例子是类似[]*string 的类型. 这总是导致slice的逃逸. 即使切
片的底层存储数组仍可能位于堆栈上, 数据的引用也会转移到堆中.

3. slice 由于 append 操作超出其容量, 因此会导致 slice 重新分配. 这种情况下, 由于在编译时 slice 的初始大小
的已知情况下, 将会在栈上分配. 如果 slice 的底层存储必须基于仅在运行时数据进行扩展, 则它将分配在堆上.

4. 调用接口类型的方法. 接口类型的方法调用是动态调度 - 实际使用的具体实现只能在运行时确定. 考虑一个接口类型为
io.Reader 的变量 r. 对 r.Read(b) 的调用将导致 r 的值和字节片 b 的后续转义并因此分配到堆上.
参考 http://npat-efault.github.io/programming/2016/10/10/escape-analysis-and-interfaces.html

5. 尽管能够符合分配到栈的场景, 但是其大小不能够在在编译时候确定的情况, 也会分配到堆上
 */

type Name struct {
	A *string
	B int
}

func sendchanl() {
	channel := make(chan int, 1)
	i := 1
	channel <- i

	arraychan := make(chan []int, 1)
	a := []int{1, 2, 3}
	arraychan <- a

	pchan := make(chan *string, 1)
	str := "Java"
	pchan <- &str

	nchan := make(chan Name, 1)
	aa := "~~~~~~~~~~~"
	n := Name{
		A: &aa,
		B: 100,
	}
	nchan <- n
}

func unknownsizeslice() {
	sizes := make([]string, 0, rand.Int()) // 编译期间无法确定具体的数字,会逃逸的heap
	str := "hello"
	sizes = append(sizes, str)
}

func bigsizeslice() {
	strings := make([]byte, 64*1024) // 超过64k大小会逃逸到heap
	strings = append(strings, 'a')
}

func pointervalueslice() {
	pointers := make([]*byte, 0, 100) // slice结构没有逃逸, 依然存在于stack上
	var b byte = 'b'
	pointers = append(pointers, &b) // &b逃逸到heap, 因为append操作, 将&b的作用域扩大
}

func iface(reader io.Reader) {
	data := make([]byte, 1024)
	reader.Read(data)
}

func pointer() interface{} {
	type Name struct {
		A, B int
		C    *string
	}

	var str = "JAVA"
	var p = &Name{A: 100, C: &str}
	p.B = 1000

	return p
}

func closeure() func() error {
	var param = 10
	var value int
	value += 1
	return func() error {
		param += 1
		return nil
	}
}

func call() {
	reader, _ := os.Open("/dev/null")
	iface(reader)
}

func main() {
	s := []byte("")
	fmt.Printf("s origin:%v, %v\n", len(s), cap(s))

	s1 := append(s, 'a')
	x1 := (*reflect.SliceHeader)(unsafe.Pointer(&s1))
	fmt.Println("unsafe len s1: ", x1.Len, "-----", x1.Cap)
	fmt.Printf("unsfate data s1 : %v\n", *(*byte)(unsafe.Pointer(x1.Data)))

	s2 := append(s, 'b')
	fmt.Printf("s2 origin:%v, %v\n", len(s2), cap(s2))
	x2 := (*reflect.SliceHeader)(unsafe.Pointer(&s2))
	fmt.Println("unsafe len s2: ", x2.Len, "-----", x2.Cap)
	fmt.Printf("unsafe data s2 : %v\n", *(*byte)(unsafe.Pointer(x2.Data)))
	// 如果有此行，打印的结果是 a b，否则打印的结果是b b
	//fmt.Printf("%s, %s \n", s1, s2)
	fmt.Printf("result: %s, %s \n", string(s1), string(s2))

	fmt.Println("\n+++++++++++++++++++++++\n")
	var o []byte
	fmt.Println("orgin: ", len(o), cap(o))
	o1 := append(o, 'a')
	fmt.Println("orgin1: ", len(o), cap(o))
	o2 := append(o, 'b')
	fmt.Println("orgin2: ", len(o), cap(o))
	fmt.Printf("origin  %s, %s\n", o1, o2)
}
