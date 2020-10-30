package array

import (
	"fmt"
	"math/rand"
	"strings"
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
	fmt.Printf("%b, %b", ^0, ^1)
}

func TestFindRoate(t *testing.T) {
	for i := 10; i < 1000; i++ {
		n := i
		nums := make([]int, 0, n)
		roate := rand.Intn(n)
		for k := roate; k < n; k++ {
			nums = append(nums, k+1)
		}
		for k := 0; k < roate; k++ {
			nums = append(nums, k+1)
		}
		r := findRoateIndex(nums)
		if nums[r] != 1 {
			t.Logf("%v, %v, %v", nums, r, nums[r])
		}

	}
}

func addBinary(a string, b string) string {
	if len(a) < len(b) {
		a = strings.Repeat("0", len(b)-len(a)) + a
	} else {
		b = strings.Repeat("0", len(a)-len(b)) + b
	}

	var ar, br = []rune(a), []rune(b)
	var remain rune
	for i := len(ar) - 1; i >= 0; i-- {
		sum := ar[i] + br[i] - 2*'0'
		if sum+remain == 3 {
			remain = 1
			ar[i] = '1'
		} else if sum+remain == 2 {
			remain = 1
			ar[i] = '0'
		} else if sum+remain == 1 {
			remain = 0
			ar[i] = '1'
		} else {
			remain = 0
			ar[i] = '0'
		}
	}
	if remain == 1 {
		ar = append([]rune("1"), ar...)
	}
	return string(ar)
}
func TestStack_Push(t *testing.T) {
	fmt.Println(addBinary("111", "111"))
}

func TestSearchRange(t *testing.T) {
	nums := []int{5, 7, 7, 8, 8, 10}
	fmt.Println(searchRange(nums, 8))
}

func TestRange(t *testing.T) {
	fmt.Println(letterCasePermutation("abc"))
}
