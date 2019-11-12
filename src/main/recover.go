package main

import (
	"sort"
	"fmt"
)

func arrayPairSum(nums []int) int {
	if len(nums) == 0 {
		return 0
	}

	sort.SliceStable(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})

	sum := 0
	for i := 0; i < len(nums); i += 2 {
		sum += nums[i]
	}
	return sum
}

func twoSum(numbers []int, target int) []int {
	if len(numbers) <= 1 {
		return nil
	}

	low := 0
	high := len(numbers) - 1
	for low < high {
		sum := numbers[low] + numbers[high]
		if sum == target {
			return []int{low + 1, high + 1}
		}
		if sum > target {
			high --
		} else {
			low++
		}
	}
	return nil
}

// 最多的0
func findMaxConsecutiveOnes(nums []int) int {
	if nums == nil {
		return 0
	}
	var max int
	var k = -1
	for i := 0; i < len(nums); i++ {
		if nums[i] == 1 && k == -1 {
			k = i
		}

		if nums[i] != 1 && k != -1 {
			if i-k > max {
				max = i - k
			}
			k = -1
		}

		if i == len(nums)-1 && k != -1 {
			if i-k+1 > max {
				max = i - k + 1
			}
		}
	}

	return max
}

// 最小长度和
func minSubArrayLen(s int, nums []int) int {
	if nums == nil {
		return 0
	}

	l := 0
	r := -1
	sum := 0
	min := len(nums) + 1
	for l < len(nums) {
		if r+1 < len(nums) && sum < s {
			r++
			sum += nums[r]
		} else {
			sum -= nums[l]
			l++
		}

		if sum >= s {
			if r-l+1 < min {
				min = r - l + 1
			}
		}
	}

	if min == len(nums)+1 {
		min = 0
	}

	return min
}

type ListNode struct {
	Next *ListNode
	Val  int
}

func deleteDuplicates(head *ListNode) *ListNode {
	if head == nil || head.Next == nil {
		return head
	}

	val := head.Val
	root := head
	for root.Next != nil {
		if root.Next.Val == val {
			root.Next = root.Next.Next
		} else {
			val = root.Next.Val
			root = root.Next
		}
	}

	return head
}

func splitListToParts(root *ListNode, k int) []*ListNode {
	if root == nil {
		return nil
	}

	var count = 0
	head := root
	for head != nil {
		count++
		head = head.Next
	}

	nodes := make([]*ListNode, k)
	if k >= count {
		i := 0
		for root != nil {
			node := root
			nodes[i] = node
			i++
			root = root.Next
			node.Next = nil
		}

		return nodes
	}

	avg := count / k
	remain := count % k

	var i, isum int
	var inode, temp *ListNode
	for root != nil {
		if isum == 0 {
			inode = root
			temp = root
			isum++
		}

		if i < remain && isum == avg+1 || i>=remain && isum == avg {
			root = root.Next
			nodes[i] = inode
			temp.Next = nil
			isum = 0
			i++
			continue
		}

		temp = temp.Next
		root = root.Next
	}

	return nodes
}

func main() {
	head := &ListNode{
		Next: &ListNode{
			Next: &ListNode{
				Next: &ListNode{
					Next: &ListNode{
						Val: 3,
					},
					Val: 3,
				},
				Val: 2,
			},
			Val: 1,
		},
		Val: 1,
	}

	node := deleteDuplicates(head)

	for node != nil {
		fmt.Println(node.Val)
		node = node.Next
	}
}
