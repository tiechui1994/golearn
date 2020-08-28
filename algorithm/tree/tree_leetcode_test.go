package tree

import (
	"testing"
)

func TestMaxSumPath(t *testing.T) {
	nodes := []string{"5", "4", "8", "11", "null", "13", "4", "7", "2", "null", "null", "null", "1"}
	node := SliceToTree(SliceToNode(nodes))
	result := MaxSumPath(node)
	t.Log("expect: 48, real:", result)
}

func TestNearestAncestor(t *testing.T) {
	strs := []string{"5", "4", "8", "11", "null", "13", "4", "7", "2", "null", "null", "null", "1"}
	nodes := SliceToNode(strs)
	node := SliceToTree(nodes)
	PrintTree(node)
	p := nodes[5]
	q := nodes[12]
	result := NearestAncestor(node, p, q)
	t.Log("expect: 8, real:", result.Val)
}

func TestPrintTree(t *testing.T) {
	strs := []string{"5", "4", "8", "11", "null", "13", "4", "7", "2", "null", "null", "null", "1"}
	nodes := SliceToNode(strs)
	_ = SliceToTree(nodes)
	Print()
}

func TestPathSum(t *testing.T) {
	nodes := []string{"5", "4", "8", "11", "null", "13", "4", "7", "2", "null", "null", "5", "1"}
	node := SliceToTree(SliceToNode(nodes))
	res := PathSum(node, 22)
	t.Log("expect: [[5 4 11 2] [5 8 4 5]]")
	t.Log("real:", res)
}

func TestVerifyPostorder(t *testing.T) {
	res := VerifyPostorder([]int{1, 2, 3, 4, 5})
	t.Log("expect: true, real:", res)
	res = VerifyPostorder([]int{5, 4, 3, 2, 1})
	t.Log("expect: true, real:", res)
	res = VerifyPostorder([]int{4, 8, 6, 12, 16, 14, 10})
	t.Log("expect: true, real:", res)
}
