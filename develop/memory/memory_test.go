package memory

import (
	"fmt"
	"testing"
	"time"
	"unsafe"
)

func AllocateLargeMem(size int64) {
	arr := make([]int64, size)
	arr[len(arr)-1] = 100102
}

func TestAllocate(t *testing.T) {
	s := time.Now()
	t.Log("size", 1<<20*64)
	AllocateLargeMem(1<<20*64 + 100)
	t.Logf("times: %v", time.Now().Sub(s).String())
}

func TestAligh(t *testing.T) {
	type Complex struct {
		a bool
		b int16
		c []byte
		d int64
	}

	var x Complex
	fmt.Println("c", unsafe.Offsetof(x.c))
	fmt.Println("d", unsafe.Offsetof(x.d))
	fmt.Println("aligin", unsafe.Alignof(x))
	fmt.Println("size", unsafe.Sizeof(x))
}
