package leetcode

import (
	"math"
	"fmt"
)

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

// 最大深度
func maxDepth(root *TreeNode) int {
	if root == nil {
		return 0
	}

	var maxdep int
	var vist func(node *TreeNode, parentDep int)
	vist = func(node *TreeNode, parentDep int) {
		if maxdep < parentDep+1 {
			maxdep = parentDep + 1
		}

		if node.Left != nil {
			vist(node.Left, parentDep+1)
		}
		if node.Right != nil {
			vist(node.Right, parentDep+1)
		}
	}

	vist(root, 0)
	return maxdep
}

func minDepth(root *TreeNode) int {
	if root == nil {
		return 0
	}

	var maxdep = math.MaxInt64
	var vist func(node *TreeNode, parentDep int)
	vist = func(node *TreeNode, parentDep int) {
		if node.Left == nil && node.Right == nil {
			if maxdep > parentDep+1 {
				maxdep = parentDep + 1
			}
		}

		if node.Left != nil {
			vist(node.Left, parentDep+1)
		}
		if node.Right != nil {
			vist(node.Right, parentDep+1)
		}
	}

	vist(root, 0)
	return maxdep
}

func pathInZigZagTree(label int) []int {
	if label == 1 {
		return []int{1}
	}

	h := int(math.Log2(float64(label))) + 1
	var nodes = []int{label}

	var findparent func(h int, child int)
	findparent = func(h int, child int) {
		if h == 2 {
			nodes = append([]int{1}, nodes...)
			return
		}

		ps := int(math.Pow(2, float64(h-2)))
		pe := int(math.Pow(2, float64(h-1)) - 1)

		parent := ps + pe - child/2
		nodes = append([]int{parent}, nodes...)
		findparent(h-1, parent)
	}

	findparent(h, label)
	fmt.Println(nodes)
	return nodes
}

type queue []*TreeNode

func (q *queue) add(ele ...*TreeNode) {
	*q = append(*q, ele...)
}

func (q *queue) poll() *TreeNode {
	val := (*q)[0]
	*q = (*q)[1:]
	return val
}

func (q *queue) len() int {
	return len(*q)
}

// 到target的距离是k的节点.
// 思路1: 深度优先遍历, 获取node->parent的map关系.
// 利用queue先进先出的特性, target为头元素, 加入其 "元素的孩子和父亲". nil为特殊元素, 遇到nil的时候就需要
// 更新深度. 还有一个全局的已经加入的节点seen, 防止多次添加
func distanceK(root, target *TreeNode, k int) []int {
	// 1. 深度优先遍历, 记录node -> parent
	var parent = make(map[*TreeNode]*TreeNode) // node -> parent
	var dfs func(node, parent *TreeNode)
	dfs = func(node, par *TreeNode) {
		if node != nil {
			parent[node] = par
			dfs(node.Left, node)
			dfs(node.Right, node)
		}
	}

	dfs(root, nil)

	// 队列
	var queue = new(queue)
	queue.add(nil, target) // nil和target被加入到队列当中. 最终存放节点
	var seen = map[*TreeNode]bool{
		target: true,
		nil:    true,
	}

	dist := 0
	for queue.len() > 0 {
		node := queue.poll()

		if node == nil {
			if dist == k {
				var res []int
				for i := range *queue {
					res = append(res, (*queue)[i].Val)
				}
				return res
			}

			queue.add(nil)
			dist++
		} else {
			if _, ok := seen[node.Left]; !ok {
				seen[node.Left] = true
				queue.add(node.Left)
			}
			if _, ok := seen[node.Right]; !ok {
				seen[node.Right] = true
				queue.add(node.Right)
			}

			par := parent[node] // 父节点
			if _, ok := seen[par]; !ok {
				seen[par] = true
				queue.add(par)
			}
		}
	}

	return nil
}

// 最大直径
func diameterOfBinaryTree(root *TreeNode) int {
	var res = 1
	max := func(i, j int) int {
		if i > j {
			return i
		}
		return j
	}
	var dfs func(node *TreeNode) int
	dfs = func(node *TreeNode) int {
		if node == nil {
			return 0
		}

		L := dfs(node.Left)
		R := dfs(node.Right)

		res = max(res, L+R+1)

		return max(L, R) + 1
	}

	dfs(root)
	return res - 1
}
