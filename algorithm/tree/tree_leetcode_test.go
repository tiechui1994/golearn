package tree

import "testing"

func TestMaxSumPath(t *testing.T) {
	nodes := []string{"-10", "9", "20", "null", "null", "15", "7"}
	node := SliceToTree(SliceToNode(nodes))
	result := MaxSumPath(node)
	t.Log("expect: 42, real:", result)
}

func TestBuild(t *testing.T) {
	mid := []int{9, 3, 15, 20, 7}
	last := []int{9, 15, 7, 20, 3}
	buildTree(mid, last)
}
