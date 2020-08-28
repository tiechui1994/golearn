package tree

import (
	"math"
)

// 给定一个非空二叉树, 返回其最大路径和
// 路径: 一条从树中任意节点出发, 到达任意节点的序列. 该路径至少包含一个节点, 且不一定经过根节点

// 思路: 递归 + 动态规划
// 需要思考两个问题, 左右分支含根+[当前根], 左右分支含根+当前根, 左(右)分支不含根
// 左右分支含根: max(根, 左含根最大值+根, 右含根最大值+根)
func MaxSumPath(root *Node) int {
	if root == nil {
		return 0
	}

	if root.Left != nil && root.Right != nil {
		ll := maxSumPathRoot(root.Left)
		rr := maxSumPathRoot(root.Right)

		l := MaxSumPath(root.Left)
		r := MaxSumPath(root.Right)
		arr := []int{root.Val, ll + root.Val, rr + root.Val, ll + rr + root.Val, ll, rr, l, r}
		return Max(arr)
	} else if root.Left != nil {
		ll := maxSumPathRoot(root.Left)
		l := MaxSumPath(root.Left)
		arr := []int{root.Val, root.Val + ll, ll, l}
		return Max(arr)
	} else if root.Right != nil {
		rr := maxSumPathRoot(root.Right)
		r := MaxSumPath(root.Right)
		arr := []int{root.Val, root.Val + rr, rr, r}
		return Max(arr)
	}

	return root.Val
}

func maxSumPathRoot(root *Node) int {
	if root == nil {
		return 0
	}

	l := maxSumPathRoot(root.Left)
	r := maxSumPathRoot(root.Right)
	arr := []int{l + root.Val, r + root.Val, root.Val}
	return Max(arr)
}

// 对称二叉树
// 性质: 1. 两个根节点具有相同的值; 2. 每个树的右子树与另一个树的左子树对称
// 迭代: 利用队列, 两棵树同时入队, 比较 (u.Left,v.Right), (u.Right,v.Left)
func IsSymmetric(root *Node) bool {
	if root == nil {
		return true
	}

	q := Queue{}
	q.push(root)
	q.push(root)

	for !q.empty() {
		u := q.pop()
		v := q.pop()
		if u == nil && v == nil {
			continue
		}

		if u == nil || v == nil || u.Val != v.Val {
			return false
		}

		q.push(u.Left)
		q.push(v.Right)

		q.push(u.Right)
		q.push(v.Left)

	}

	return true
}

func BuildTree(inorder []int, postorder []int) *Node {
	if len(inorder) == 0 || len(postorder) == 0 {
		return nil
	}
	if len(inorder) != len(postorder) {
		return nil
	}
	var build func(inorder []int, postorder []int) *Node
	build = func(inorder []int, postorder []int) *Node {
		// 当只有一个节点或者没有节点
		if len(inorder) == 0 {
			return nil
		}
		if len(inorder) == 1 {
			return &Node{Val: inorder[0]}
		}
		// inorder: 左根右
		// postorder: 左右根
		//
		// 思路:
		// postorder 的最后一个元素就是当前的根, 以此根为划分 inorder, 划分出左右子树;
		// 与此同时, 也需要划分 postorder, 因为是 "左右根", 左子树,[0:左子树长度) 右子树, [左子树长度:n-1)
		n := len(postorder)
		root := &Node{Val: postorder[n-1]}

		idx := index(inorder, postorder[n-1])
		root.Left = build(inorder[:idx], postorder[0:idx])
		root.Right = build(inorder[idx+1:], postorder[idx:n-1])

		return root
	}

	return build(inorder, postorder)
}

func index(arr []int, el int) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == el {
			return i
		}
	}

	return -1
}

// 最近公共祖先
// 思路: 自底向上递归, 左右根, 即后续遍历算法
// 1. 如果当前节点为 null, 则返回 null
// 2. 如果当前节点为 p 或 q, 或者包含了 p 或者 q 则返回当前节点
// 3. 如果当前包含了p和q, 则第一次到达的节点为最近的祖先
func NearestAncestor(root, p, q *Node) *Node {
	if root == nil {
		return nil
	}
	if root == p || root == q {
		return root
	}

	left := NearestAncestor(root.Left, p, q)
	right := NearestAncestor(root.Right, p, q)
	if left != nil && right != nil {
		return root
	}

	if left != nil {
		return left
	}
	if right != nil {
		return right
	}
	return nil
}

// 二叉树中和为某一值的路径
// 打印出二叉树中节点值的和为输入整数的所有路径. 从树的 "根节点" 开始往下一直到 "叶节点" 所经过的节点形成一条路径.
// 思路: 自顶向下, 根左右, 直到根节点
func PathSum(root *Node, sum int) [][]int {
	if root == nil {
		return nil
	}

	if root.Left == nil && root.Right == nil && root.Val == sum {
		return [][]int{{root.Val}}
	}

	parent := make([]int, 0)
	return pathsum(root, sum, parent)
}

func pathsum(root *Node, sum int, parent []int) [][]int {
	if root.Left == nil && root.Right == nil {
		if root.Val == sum {
			parent = append(parent, root.Val)
			return [][]int{parent}
		}
		return nil
	}

	// 分别计算左右子树, 和为 sum-root.Val 的路径
	var left, right [][]int
	if root.Left != nil {
		// 注意: 由于 slice, 必须使用深度复制
		t := make([]int, len(parent))
		copy(t, parent)
		t = append(t, root.Val)
		left = pathsum(root.Left, sum-root.Val, t)
	}
	if root.Right != nil {
		t := make([]int, len(parent))
		copy(t, parent)
		t = append(t, root.Val)
		right = pathsum(root.Right, sum-root.Val, t)
	}

	return append(left, right...)
}

// 搜索树的第k小数, 中序遍历(左根右) -> 有序
func KthSmallest(root *Node, k int) int {
	if root == nil {
		return 0
	}

	var res int
	var count int
	var visit func(node *Node)
	visit = func(node *Node) {
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

func PathInZigZagTree(label int) []int {
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
	return nodes
}

// 到target的距离是k的节点.
// 思路1: 深度优先遍历, 获取node->parent的map关系.
// 利用queue先进先出的特性, target为头元素, 加入其 "元素的孩子和父亲". nil为特殊元素, 遇到nil的时候就需要
// 更新深度. 还有一个全局的已经加入的节点seen, 防止多次添加.
func DistanceK(root, target *Node, k int) []int {
	// 1. 深度优先遍历, 记录node -> parent
	var parent = make(map[*Node]*Node) // node -> parent
	var dfs func(node, parent *Node)
	dfs = func(node, par *Node) {
		if node != nil {
			parent[node] = par
			dfs(node.Left, node)
			dfs(node.Right, node)
		}
	}

	dfs(root, nil)

	// 队列
	var queue = Queue{}
	queue.push(nil)
	queue.push(target) // nil和target被加入到队列当中. 最终存放节点
	var seen = map[*Node]bool{
		target: true,
		nil:    true,
	}

	dist := 0
	for !queue.empty() {
		node := queue.pop()

		if node == nil {
			if dist == k {
				var res []int
				for i := range queue {
					res = append(res, queue[i].Val)
				}
				return res
			}

			queue.push(nil)
			dist++
		} else {
			if _, ok := seen[node.Left]; !ok {
				seen[node.Left] = true
				queue.push(node.Left)
			}
			if _, ok := seen[node.Right]; !ok {
				seen[node.Right] = true
				queue.push(node.Right)
			}

			p := parent[node] // 父节点
			if _, ok := seen[p]; !ok {
				seen[p] = true
				queue.push(p)
			}
		}
	}

	return nil
}

// 输入一个整数数组, 判断该数组是不是某二叉搜索树的后序遍历结果.
// 思路: 对于长度是0, 1, 2 的 postorder, 无论什么样的顺序, 都返回 true
//      对于长度为大于等于3的 postorder, 最后一个元素是 root, 然后在前面查找一个位置, 该位置满足条件:
//		该位置之前的元素, postorder[i] < root
//		该位置之后的元素, postorder[i] > root
// 如果存在这样的位置, 则说明当前的根元素满足要求, 依据找到的位置, 将 postorder 一分为2, 左子树和右子树,
// 当左右子树都满足要求, 则最终满足要求.
func VerifyPostorder(postorder []int) bool {
	index := func(arr []int, ele int) int {
		isLess := true
		index := -1
		for i, val := range arr {
			if isLess && i+1 < len(arr) {
				if arr[i+1] > ele {
					index = i
					isLess = false
					continue
				}
			}

			// 所有元素都小于 ele
			if isLess && i+1 == len(arr) {
				index = i
				isLess = false
				continue
			}

			// 条件检测
			if isLess && val > ele {
				return -1
			}

			if !isLess && val < ele {
				return -1
			}
		}

		return index
	}

	if len(postorder) <= 2 {
		return true
	}

	n := len(postorder)
	root := postorder[n-1]
	idx := index(postorder[:n-1], root)
	if idx == -1 {
		return false
	}

	return VerifyPostorder(postorder[0:idx+1]) && VerifyPostorder(postorder[idx+1:n-1])
}

// 检查 t2 是否为 t1 的子树
// 思路: t1, t2 其中一个为 nil, 直接可以判断. 难点在于 t1, t2 都不为 nil
// 使用递归进行判断.
// 假设当前比较到了节点 root(t1), cur(t2), sub(t2)
// 1. 如果 root.Val == cur.Val, 则继续比较 [root.Left,cur.Left] && [root.Right, cur.Right],
// 2. 如果 root.Val == sub.Val, 则需要重新开始比较 [root.Left,sub.Left] && [root.Right, sub.Right],
// 3. 如果 root.Val != sub.Val, 则需要比较  [root.Left,sub] || [root.Right, sub],
// 结束的条件: root == nil && cur == nil, 某一个分支到达了根, 结果是 true
//			 root == nil || cur == nil, 某一个分支的 root 或者 cur 提前结束了, 则结果是 false

func CheckSubTree(t1 *Node, t2 *Node) bool {
	var checksub func(root, cur *Node, subtree *Node) bool
	checksub = func(root, cur *Node, subtree *Node) bool {
		if root == nil && cur == nil {
			return true
		}

		if root == nil || cur == nil {
			return false
		}

		if root.Val == cur.Val {
			return checksub(root.Left, cur.Left, subtree) && checksub(root.Right, cur.Right, subtree)
		} else if root.Val == subtree.Val {
			return checksub(root.Left, subtree.Left, subtree) && checksub(root.Right, subtree.Left, subtree)
		} else {
			return checksub(root.Left, subtree, subtree) || checksub(root.Right, subtree, subtree)
		}
	}

	if t1 == nil && t2 == nil {
		return true
	}
	if t1 == nil {
		return false
	}
	if t2 == nil {
		return true
	}

	return checksub(t1, t2, t2)
}
