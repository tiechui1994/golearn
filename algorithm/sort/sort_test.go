package sort

import (
	"testing"
	"fmt"
)

func TestQuickSort(t *testing.T) {
	sort := QuickSort
	case1 := []int{13, 2, 9, 3, 7, 6, 19, 22}

	fmt.Println("Before", case1)
	sort(case1)
	fmt.Println("After", case1)
}
