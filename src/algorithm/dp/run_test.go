package dp

import "testing"

func TestLongIncrSeq(t *testing.T) {
	testcase := []int{10, 22, 9, 33, 21, 50, 41, 60, 80}

	testcase = []int{1, 73, 10, 33, 44, 88, 22, 9, 33, 21, 50}
	testcase = []int{10, 22, 9, 33, 21, 50, 88, 55, 41, 22, 60, 80}

	res := LongIncrSeq2(testcase)
	t.Log("res", res)
}

func TestCombineStone(t *testing.T) {
	testcase := []int{3, 1, 2, 4, 5}
	testcase = []int{3, 2, 4, 1}
	res := CombineStone(testcase)
	t.Log("res", res)
}
