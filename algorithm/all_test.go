package algorithm

import (
	"testing"
	"fmt"
	"math/rand"
	"time"
)

func TestNewSkipList(t *testing.T) {
	skiplist := NewSkipList()
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 100; i > 0; i-- {
		skiplist.Insert(seed.Intn(100))
	}

	skiplist.PrintSkipList()
}

func TestInsertSort(t *testing.T) {
	nums := []int{1, 2, 3, 3, 4}
	InsertSort(nums)
	fmt.Println(nums)
}

func TestMergeSort(t *testing.T) {
	nums := []int{1, 2, 3, 32, 11, 4}
	MergeSort(nums)
	fmt.Println(nums)
}

func TestQuickSort(t *testing.T) {
	nums := []int{1, 2, 3, 32, 11, 4}
	QuickSort(nums)
	fmt.Println(nums)
}

func TestQuickSort2(t *testing.T) {
	nums := []int{1, 2, 3, 32, 11, 4}
	QuickSort2(nums)
	fmt.Println(nums)
}

func TestQuickSort3(t *testing.T) {
	const base = 10000000
	nums1, nums2, nums3 := make([]int, base), make([]int, base), make([]int, base)
	for i := 0; i < base; i++ {
		r := rand.Intn(base)
		nums1[i] = r
		nums2[i] = r
		nums3[i] = r
	}

	start := time.Now()
	QuickSort(nums1)
	t.Logf("QuickSort: %v, num:%v", time.Now().Sub(start).Seconds(), base)

	start = time.Now()
	QuickSort2(nums2)
	t.Logf("QuickSort2:%v, num:%v", time.Now().Sub(start).Seconds(), base)

	start = time.Now()
	HeapSort(nums3)
	t.Logf("HeapSort:  %v, num:%v", time.Now().Sub(start).Seconds(), base)

	t.Logf("\n++++++++++++++++\n")

}

func TestHeapSort(t *testing.T) {
	nums := []int{1, 2, 3, 32, 11, 4, 181, 221, 1127, 11, 2, 43, 11}
	HeapSort(nums)
	fmt.Println(nums)
}
