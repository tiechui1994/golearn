package leetcode

import (
	"fmt"
	"testing"
)

func TestThreeSum(t *testing.T) {
	fmt.Println(threeSum([]int{-4, -1, -1, 0, 1, 2}))
}

func TestListSum(t *testing.T) {
	l1 := &ListNode{
		Val: 2,
		Next: &ListNode{
			Val: 4,
			Next: &ListNode{
				Val:  9,
				Next: nil,
			},
		},
	}

	l2 := &ListNode{
		Val: 5,
		Next: &ListNode{
			Val: 6,
			Next: &ListNode{
				Val:  4,
				Next: nil,
			},
		},
	}

	res := addTwoNumbers(l1, l2)
	for res != nil {
		fmt.Printf("%+v\n", res.Val)
		res = res.Next
	}

}

func TestListSumMore(t *testing.T) {
	l1 := &ListNode{
		Val: 7,
		Next: &ListNode{
			Val: 2,
			Next: &ListNode{
				Val: 4,
				Next: &ListNode{
					Val:  3,
					Next: nil,
				},
			},
		},
	}

	l2 := &ListNode{
		Val: 5,
		Next: &ListNode{
			Val: 6,
			Next: &ListNode{
				Val:  4,
				Next: nil,
			},
		},
	}

	//l1 = &ListNode{
	//	Val:0,
	//}
	//l2 = &ListNode{
	//	Val:0,
	//}
	res := addTwoNumbersMore(l1, l2)
	for res != nil {
		fmt.Printf("==> %+v  ", res.Val)
		res = res.Next
	}
}

func TestRemoveDuplicates(t *testing.T) {
	res := removeDuplicates([]int{0, 0, 1, 1, 1, 2, 2, 3, 3, 4})
	fmt.Println(res)
}

func TestFindPeek(t *testing.T) {
	ok, res := findPeek([]int{-200, -100, 1, 4, 0})
	fmt.Println(ok, res)
}

func TestKNumSum(t *testing.T) {
	nums := []int{1, 0, -1, 0, -2, 2}
	res := KNumSum(nums, 4, 0)
	for i := range res {
		fmt.Println(res[i])
	}
}

func TestSearch(t *testing.T) {
	nums := []int{4, 5, 6, 7, 0, 1, 2}
	search(nums, 10)
	fmt.Println(0&^0, 0&^1, 1&^0, 1&^1)
	fmt.Printf("%b, %b",^0, ^1)
}
