package tree

/*
树的遍历
*/

// 先序遍历: 根左右
func PreOrder(root *TreeNode) []int {
	if root == nil {
		return nil
	}
	if root.Left == nil && root.Right == nil {
		return []int{root.Val}
	}

	m := make([]int, 0)
	s := Stack{}

	cur := root
	for !s.empty() || cur != nil {
		for cur != nil {
			m = append(m, cur.Val)
			s.push(cur)
			cur = cur.Left
		}

		if !s.empty() {
			cur = s.pop()
			cur = cur.Right
		}
	}

	return m
}

// 中序遍历: 左根右
func MidleOrder(root *TreeNode) []int {
	if root == nil {
		return nil
	}
	if root.Left == nil && root.Right == nil {
		return []int{root.Val}
	}

	m := make([]int, 0)
	s := Stack{}
	cur := root
	for !s.empty() || cur != nil {
		for cur != nil {
			s.push(cur)
			cur = cur.Left
		}
		if !s.empty() {
			cur = s.pop()
			m = append(m, cur.Val)
			cur = cur.Right
		}
	}

	return m
}

// 后序遍历: 左右根 => 根右左
func LastOrder(root *TreeNode) []int {
	if root == nil {
		return nil
	}
	if root.Left == nil && root.Right == nil {
		return []int{root.Val}
	}

	s := Stack{}
	data := Stack{}
	cur := root
	for !s.empty() || cur != nil {
		for cur != nil {
			s.push(cur)
			data.push(cur)
			cur = cur.Right
		}

		if !s.empty() {
			cur = s.pop()
			cur = cur.Left
		}
	}

	m := make([]int, 0)
	for !data.empty() {
		cur = data.pop()
		if cur != nil {
			m = append(m, cur.Val)
		}
	}

	return m
}

// 层序遍历
func LevelOrder(root *TreeNode) [][]int {
	if root == nil {
		return nil
	}
	if root.Left == nil && root.Right == nil {
		return [][]int{{root.Val}}
	}

	var (
		cur    *TreeNode
		values [][]int
	)
	q := Queue{}
	q.push(root)
	for !q.empty() {
		value := make([]int, 0)
		p := Queue{}
		for !q.empty() {
			cur = q.pop()
			value = append(value, cur.Val)
			if cur.Left != nil {
				p.push(cur.Left)
			}
			if cur.Right != nil {
				p.push(cur.Right)
			}
		}
		values = append(values, value)
		q = p
	}

	return values
}
