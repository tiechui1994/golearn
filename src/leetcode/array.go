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

// 两个链表, 逆序存储数字, 求和
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

// 两个链表, 顺序存储数字, 求和 ==> (栈的先进后出, 变为逆序存储)
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

// 三个数字的和是0的所有组合, 空间: O(N) 时间:O(N^2)
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

// 判断是否存在环 ==> map
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

// 判断是否存在环 ==> 双指针. next 和 next.next
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

//-----------------------------------------------------------------

// 有序的数组, 去除重复的数字 ==> 双指针
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

// 无序的数组, 去除某个所有的元素 ==> 双指针
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

// 波峰问题. 二分查找的特性
// 条件: arr[0]<arr[1] a[m] < a[m-1]
// 波峰: arr[m-1] < arr[m] && arr[m] > arr[m+1]
// a[m-1] < a[m] < a[m+1], 增长
// a[m-1] > a[m] > a[m+1], 减少
func findPeek(nums []int) (bool, int) {
	// 单调递增
	isupper := func(nums []int, m int) bool {
		return nums[m-1] < nums[m] && nums[m] < nums[m+1]
	}

	if nums == nil || len(nums) <= 2 {
		return false, -1
	}

	start := 0
	end := len(nums) - 1
	for start <= end {
		m := (start + end) / 2
		if m == 0 || m == len(nums)-1 {
			return false, -1
		}

		if nums[m-1] < nums[m] && nums[m] > nums[m+1] {
			return true, nums[m]
		}

		if isupper(nums, m) {
			start = m + 1
		} else {
			end = m - 1
		}
	}

	return false, -1
}

// ------------------------------------------------------------------------
// K个数的和是target的算法.
func KNumSum(nums []int, k, target int) [][]int {
	sort.SliceStable(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})
	return ksum(nums, 0, k, target)
}

func ksum(nums []int, start, k, target int) [][]int {
	n := len(nums)
	var res [][]int
	if k == 2 {
		var left, right = start, n-1
		for left < right {
			sum := nums[left] + nums[right]
			if sum == target {
				res = append(res, []int{nums[left], nums[right]})
				for left < right && nums[left] == nums[left+1] {
					left++
				}
				for left < right && nums[right] == nums[right-1] {
					right--
				}
				left++
				right--
			} else if sum < target {
				left++
			} else {
				right--
			}
		}

		return res
	}

	end := n - (k - 1)
	for i := start; i < end; i++ {
		if i > start && nums[i] == nums[i-1] {
			continue
		}
		temp := ksum(nums, i+1, k-1, target-nums[i])
		for j := range temp {
			temp[j] = append([]int{nums[i]}, temp[j]...)
		}
		res = append(res, temp...)
	}

	return res
}
