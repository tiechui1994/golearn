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
