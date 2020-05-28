package leetcode

import (
	"testing"
	"fmt"
)

func TestInorderTraversal(t *testing.T) {
	tree := &TreeNode{
		Val:  1,
		Left: nil,
		Right: &TreeNode{
			Val: 2,
			Left: &TreeNode{
				Val: 3,
			},
		},
	}

	res := inorderTraversal(tree)
	fmt.Println(res)
}

func TestKthSmallest(t *testing.T) {
	tree := &TreeNode{
		Val: 3,
		Left: &TreeNode{
			Val: 1,
			Right: &TreeNode{
				Val: 2,
			},
		},
		Right: &TreeNode{
			Val: 4,
		},
	}

	res := kthSmallest(tree, 1)
	fmt.Println(res)
}

func TestPathInZigZagTree(t *testing.T) {
	pathInZigZagTree(26)
}
