package list

import (
	"fmt"
	"testing"
	"time"
)

func List(arr []int) *ListNode {
	if arr == nil {
		return nil
	}
	head := &ListNode{
		Val: arr[0],
	}
	root := head
	for i := 1; i < len(arr); i++ {
		head.Next = &ListNode{Val: arr[i]}
		head = head.Next
	}

	return root
}

func TestReverseBetween(t *testing.T) {
	s := time.Now()
	head := &ListNode{
		Next: &ListNode{
			Next: &ListNode{
				Next: &ListNode{
					Next: &ListNode{
						Val: 5,
					},
					Val: 4,
				},
				Val: 3,
			},
			Val: 2,
		},
		Val: 1,
	}
	reverseBetween(head, 2, 4)
	fmt.Printf("%+v", time.Now().Sub(s).Nanoseconds())
}

func TestIsPalindrome(t *testing.T) {
	var isPalindrome = isPalindromeII
	case1 := List(nil)
	fmt.Println(case1, isPalindrome(case1))

	case2 := List([]int{1, 2})
	fmt.Println(case2, isPalindrome(case2))

	case3 := List([]int{1, 2, 2, 1})
	fmt.Println(case3.String(), isPalindrome(case3))

	case4 := List([]int{1, 2, 3, 2, 1})
	fmt.Println(case4.String(), isPalindrome(case4))
}

func TestHasCycle(t *testing.T) {
	var hasCycle = hasCycleI

	case1 := List([]int{3, 2, 0, -4})
	case1.Next.Next.Next.Next = case1.Next
	fmt.Println(hasCycle(case1))

	case2 := List([]int{1, 2})
	fmt.Println(hasCycle(case2))
}

func TestReverseKGroup(t *testing.T) {
	reverseKGroup := reverseKGroup
	case1 := List([]int{1, 2, 3, 4, 5})
	fmt.Println(reverseKGroup(case1, 2))

}
