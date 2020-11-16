package memory

import (
	"testing"
	"time"
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
