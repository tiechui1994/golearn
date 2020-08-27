package tree

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
