package tree

import "strconv"

type Node struct {
	Left  *Node
	Right *Node
	Val   int
}

type Stack []*Node

func (s *Stack) empty() bool {
	return len(*s) == 0
}

func (s *Stack) push(node *Node) {
	*s = append(*s, node)
}

func (s *Stack) pop() *Node {
	if s == nil || len(*s) == 0 {
		return nil
	}

	node := (*s)[len(*s)-1]
	*s = (*s)[0:len(*s)-1]
	return node
}

type Queue []*Node

func (q *Queue) empty() bool {
	return len(*q) == 0
}

func (q *Queue) push(node *Node) {
	*q = append(*q, node)
}

func (q *Queue) pop() *Node {
	if q == nil || len(*q) == 0 {
		return nil
	}

	node := (*q)[0]
	*q = (*q)[1:]
	return node
}

func SliceToNode(strs []string) ([]*Node) {
	if strs == nil || len(strs) == 0 {
		return nil
	}

	nodes := make([]*Node, 0, len(strs))
	for _, s := range strs {
		if s == "null" {
			nodes = append(nodes, nil)
		} else {
			val, _ := strconv.ParseInt(s, 10, 64)
			nodes = append(nodes, &Node{
				Val: int(val),
			})
		}
	}

	return nodes
}

func SliceToTree(st []*Node) (root *Node) {
	if len(st) == 0 {
		return nil
	}
	if len(st) == 1 {
		return st[0]
	}

	root, st = fillTree(st[0], st[1:])
	node := root
	for len(st) > 0 {
		if node.Left != nil && node.Right != nil {
			_, st = fillTree(node.Left, st)
			_, st = fillTree(node.Right, st)
			node = node.Left
			continue
		}

		if node.Left != nil {
			_, st = fillTree(node.Left, st)
			node = node.Left
			continue
		}

		if node.Right != nil {
			_, st = fillTree(node.Right, st)
			node = node.Right
			continue
		}
	}

	return root
}

func fillTree(h *Node, st []*Node) (*Node, []*Node) {
	if len(st) < 1 {
		return h, nil
	}
	if len(st) == 1 {
		h.Left = st[0]
		return h, nil
	}
	if len(st) == 2 {
		h.Left = st[0]
		h.Right = st[1]
		return h, nil
	}

	h.Left = st[0]
	h.Right = st[1]
	return h, st[2:]
}

func Max(slice []int) int {
	if len(slice) == 1 {
		return slice[0]
	}

	max := slice[0]
	for i := 1; i < len(slice); i++ {
		if max < slice[i] {
			max = slice[i]
		}
	}

	return max
}
