package leetcode

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

// 中序遍历
func inorderTraversal(root *TreeNode) []int {
	if root == nil {
		return nil
	}

	var visit func(node *TreeNode, arr *[]int)
	visit = func(node *TreeNode, arr *[]int) {
		if node.Left != nil {
			visit(node.Left, arr)
		}

		*arr = append(*arr, node.Val)

		if node.Right != nil {
			visit(node.Right, arr)
		}

	}

	var arr = make([]int, 0, 100)
	visit(root, &arr)
	return arr
}

// 搜索树的第k小数, 中序遍历(左根右) -> 有序
func kthSmallest(root *TreeNode, k int) int {
	if root == nil {
		return 0
	}

	var res int
	var count int
	var visit func(node *TreeNode)
	visit = func(node *TreeNode) {
		if node.Left != nil {
			visit(node.Left)
		}

		count++
		if count == k {
			res = node.Val
			return
		}

		if node.Right != nil {
			visit(node.Right)
		}
	}

	visit(root)
	return res
}
