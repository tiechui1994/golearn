package tree

func Pre(root *Node) []int {
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

func Midle(root *Node) []int {
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

func Last(root *Node) []int {
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
