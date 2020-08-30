package tree

import (
	"strconv"
	"fmt"
)

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

type Set struct {
	data map[int]bool
}

func (s *Set) len() int {
	return len(s.data)
}
func (s *Set) add(i int) {
	if s.data == nil {
		s.data = make(map[int]bool)
	}
	s.data[i] = true
}

func (s *Set) remove(i int) {
	if s.data == nil {
		return
	}
	delete(s.data, i)
}

func (s *Set) contains(i int) bool {
	if s.data == nil {
		s.data = make(map[int]bool)
		return false
	}
	return s.data[i]
}

func (s *Set) set() []int {
	if s.data == nil {
		return []int{}
	}
	values := make([]int, 0, len(s.data))
	for k := range s.data {
		values = append(values, k)
	}
	return values
}

///////////////////////////////////////

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
	q := Queue{}
	q.push(root)

	for len(st) > 0 && !q.empty() {
		node := q.pop()
		if node.Left != nil && node.Right != nil {
			_, st = fillTree(node.Left, st)
			_, st = fillTree(node.Right, st)
			q.push(node.Left)
			q.push(node.Right)
			continue
		}

		if node.Left != nil {
			_, st = fillTree(node.Left, st)
			q.push(node.Left)
			continue
		}

		if node.Right != nil {
			_, st = fillTree(node.Right, st)
			q.push(node.Right)
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

func Max(slice ...int) int {
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

func Min(slice ...int) int {
	if len(slice) == 1 {
		return slice[0]
	}

	min := slice[0]
	for i := 1; i < len(slice); i++ {
		if min > slice[i] {
			min = slice[i]
		}
	}

	return min
}

func PrintTree(node *Node) {
	q := Queue{}
	q.push(node)
	for !q.empty() {
		p := Queue{}
		stoped := true
		for !q.empty() {
			cur := q.pop()
			if cur != nil {
				fmt.Printf("  %v  ", cur.Val)
				if cur.Left != nil || cur.Right != nil {
					stoped = false
				}
				p.push(cur.Left)
				p.push(cur.Right)
			} else {
				fmt.Printf(" null ")
			}
		}
		fmt.Println()
		if stoped {
			break
		}
		q = p
	}
}

func Print() {
	i, j, k, depth := 0, 0, 0, 4
	for j = 0; j < depth; j++ {
		w := 1 << uint(depth-j+1)
		if j == 0 {
			fmt.Printf("%*c\n", w, '_')
		} else {
			for i = 0; i < 1<<uint(j-1); i++ {
				fmt.Printf("%*c", w+1, ' ')
				for k = 0; k < w-3; k++ {
					fmt.Printf("_")
				}
				fmt.Printf("/ \\")
				for k = 0; k < w-3; k++ {
					fmt.Printf("_")
				}
				fmt.Printf("%*c", w+2, ' ')
			}
			fmt.Printf("\n")

			for i = 0; i < 1<<uint(j-1); i++ {
				fmt.Printf("%*c/%*c_%*c", w, '_', w*2-2, '\\', w, ' ')
			}
			fmt.Printf("\n")
		}

		for i = 0; i < 1<<uint(j); i++ {
			fmt.Printf("%*c_)%*c", w-1, '(', w-1, ' ')
		}
		fmt.Printf("\n")
	}
}
