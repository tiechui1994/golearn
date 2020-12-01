package list

import (
	"fmt"
	"testing"
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

func Tail(head *ListNode) *ListNode {
	if head == nil || head.Next == nil {
		return head
	}

	cur := head
	for cur != nil {
		if cur.Next == nil {
			return cur
		}
		cur = cur.Next
	}

	return nil
}

func Reverse(head *ListNode) *ListNode {
	if head == nil || head.Next == nil {
		return head
	}

	var pre *ListNode
	cur := head

	for cur != nil {
		next := cur.Next

		cur.Next = pre
		pre = cur

		cur = next
	}

	return pre
}

func TestReverseBetween(t *testing.T) {
	case1 := List([]int{1, 2, 3, 4, 5})
	reverseBetween(case1, 2, 4)
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

func TestGetIntersectionNode(t *testing.T) {
	getIntersectionNode := getIntersectionNodeII
	a := List([]int{1, 2, 4})
	b := List([]int{4, 5})
	common := List([]int{99, 22, 44, 11})
	Tail(a).Next = common
	Tail(b).Next = common

	node := getIntersectionNode(a, b)
	fmt.Println(node.Val)
}
