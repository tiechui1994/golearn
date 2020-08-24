package tree

func PreOrder(root *Node) []int {
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

func MidleOrder(root *Node) []int {
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

func LastOrder(root *Node) []int {
	if root == nil {
		return nil
	}
	if root.Left == nil && root.Right == nil {
		return []int{root.Val}
	}

	help := Stack{}
	visit := Stack{}
	cur := root
	for !help.empty() || cur != nil {
		for cur != nil {
			help.push(cur)
			visit.push(cur)
			cur = cur.Right
		}

		if !help.empty() {
			cur = help.pop()
			cur = cur.Left
		}
	}

	m := make([]int, 0)
	for !visit.empty() {
		cur = visit.pop()
		if cur != nil {
			m = append(m, cur.Val)
		}
	}

	return m
}

func LevelOrder(root *Node) [][]int {
	if root == nil {
		return nil
	}
	if root.Left == nil && root.Right == nil {
		return [][]int{{root.Val}}
	}

	var (
		cur    *Node
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
