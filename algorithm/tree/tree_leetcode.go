package tree

// 给定一个非空二叉树, 返回其最大路径和
// 路径: 一条从树中任意节点出发, 到达任意节点的序列. 该路径至少包含一个节点, 且不一定经过根节点

func MaxSumPath(root *Node) int {
	if root == nil {
		return 0
	}

	if root.Left != nil && root.Right != nil {
		ll := maxSumPathRoot(root.Left)
		rr := maxSumPathRoot(root.Right)

		l := MaxSumPath(root.Left)
		r := MaxSumPath(root.Right)
		arr := []int{root.Val, ll + root.Val, rr + root.Val, ll, rr, l + r + root.Val, l, r}
		return Max(arr)
	} else if root.Left != nil {
		ll := maxSumPathRoot(root.Left)
		l := MaxSumPath(root.Left)
		arr := []int{root.Val, root.Val + ll, ll, l + root.Val, l}
		return Max(arr)
	} else if root.Right != nil {
		rr := maxSumPathRoot(root.Right)
		r := MaxSumPath(root.Right)
		arr := []int{root.Val, root.Val + rr, rr, r + root.Val, r}
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
