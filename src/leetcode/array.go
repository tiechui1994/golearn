package leetcode

import (
	"fmt"
	"sort"
)

/**
输入：(2 -> 4 -> 3) + (5 -> 6 -> 4)
输出：7 -> 0 -> 8
原因：342 + 465 = 807
 */

/**
* Definition for singly-linked list.
* type ListNode struct {
*     Val int
*     Next *ListNode
* }
*/

type ListNode struct {
	Val  int
	Next *ListNode
}

// 逆序存储
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	res := new(ListNode)
	n1 := l1
	n2 := l2

	var node = res
	var carry int
	for n1 != nil && n2 != nil {
		sum := n1.Val + n2.Val + carry
		carry = sum / 10
		node.Val = sum % 10

		n1 = n1.Next
		n2 = n2.Next

		// 存在n1或者n2
		if n1 != nil || n2 != nil {
			node.Next = new(ListNode)
			node = node.Next
		}
	}

	for n1 != nil {
		sum := n1.Val + carry
		carry = sum / 10
		node.Val = sum % 10

		n1 = n1.Next
		if n1 != nil {
			node.Next = new(ListNode)
			node = node.Next
		}
	}

	for n2 != nil {
		sum := n2.Val + carry
		carry = sum / 10
		node.Val = sum % 10

		n2 = n2.Next
		if n2 != nil {
			node.Next = new(ListNode)
			node = node.Next
		}
	}

	if carry > 0 {
		node.Next = &ListNode{
			Val:  carry,
			Next: nil,
		}
	}

	return res
}

type stack []int

func (s *stack) Push(key int) {
	*s = append([]int{key}, *s...)
}

func (s *stack) Pop() int {
	val := (*s)[0]
	*s = (*s)[1:]
	return val
}

func (s *stack) Len() int {
	return len(*s)
}

// 顺序存储(栈的先进后出, 变为逆序存储)
func addTwoNumbersMore(l1 *ListNode, l2 *ListNode) *ListNode {
	var s1, s2 = new(stack), new(stack)
	n1 := l1
	for n1 != nil {
		s1.Push(n1.Val)
		n1 = n1.Next
	}
	n2 := l2
	for n2 != nil {
		s2.Push(n2.Val)
		n2 = n2.Next
	}
	var node = new(ListNode)
	var carry int

	for s1.Len() > 0 && s2.Len() > 0 {
		e1 := s1.Pop()
		e2 := s2.Pop()
		sum := e1 + e2 + carry
		carry = sum / 10
		node.Val = sum % 10

		if s1.Len() > 0 || s2.Len() > 0 {
			node = &ListNode{
				Next: node,
			}
		}
	}

	for s1.Len() > 0 {
		e1 := s1.Pop()
		sum := e1 + carry
		carry = sum / 10
		node.Val = sum % 10

		if s1.Len() > 0 {
			node = &ListNode{
				Next: node,
			}
		}
	}

	for s2.Len() > 0 {
		e2 := s2.Pop()
		sum := e2 + carry
		carry = sum / 10
		node.Val = sum % 10

		if s2.Len() > 0 {
			node = &ListNode{
				Next: node,
			}
		}
	}

	if carry > 0 {
		node = &ListNode{
			Val:  carry,
			Next: node,
		}
	}

	return node
}

// 判断是否有环
func hasCycle(head *ListNode) bool {
	m := make(map[*ListNode]bool)
	for n := head; n != nil; n = n.Next {
		if _, ok := m[n]; ok {
			return true
		}
		m[n] = true
	}

	return false
}

// 赛跑
func hasCyclePointer(head *ListNode) bool {
	if head == nil || head.Next == nil {
		return false
	}

	slow := head
	fast := head.Next
	for slow != fast {
		if fast == nil || fast.Next == nil {
			return false
		}

		slow = slow.Next
		fast = fast.Next.Next
	}

	return true
}

// 三个数字的和是0的所有组合, 空间:O(N) 时间:O(N^2)
func threeSum(nums []int) [][]int {
	sort.SliceStable(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})

	var res [][]int
	size := len(nums)
	if nums[0] > 0 && nums[size-1] < 0 {
		return res
	}

	for k := 0; k < size-2; k++ {
		if nums[k] > 0 {
			break
		}
		// 重复
		if k > 0 && nums[k] == nums[k-1] {
			continue
		}

		i := k + 1
		j := size - 1
		for i < j {
			sum := nums[k] + nums[i] + nums[j]
			if sum == 0 {
				res = append(res, []int{nums[i], nums[j], nums[k]})
				i++
				for i < j && nums[i] == nums[i-1] {
					i++
				}

				j--
				for i < j && nums[j] == nums[j+1] {
					j--
				}
			} else if sum < 0 {
				i++
				for i < j && nums[i] == nums[i-1] {
					i++
				}
			} else {
				j--
				for i < j && nums[j] == nums[j+1] {
					j--
				}
			}
		}
	}

	return res
}

//-----------------------------------------------------------------

// 有序的数组, 去除重复
func removeDuplicates(nums []int) int {
	fmt.Println(nums)
	if len(nums) == 0 {
		return 0
	}

	var k int
	for i := 1; i < len(nums); i++ {
		if nums[i] != nums[k] {
			k++
			nums[k] = nums[i]
		}
	}

	return k + 1
}

// 数组, 移除重复元素
func removeElement(nums []int, val int) int {
	if len(nums) == 0 {
		return 0
	}

	var k int
	for i := 0; i < len(nums); i++ {
		if nums[i] != val {
			nums[k] = nums[i]
			k++
		}
	}

	return k
}
