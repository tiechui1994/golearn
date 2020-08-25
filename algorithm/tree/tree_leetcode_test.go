package tree

import "testing"

func TestMaxSumPath(t *testing.T) {
	nodes := []string{"-10", "9", "20", "null", "null", "15", "7"}
	node := SliceToTree(SliceToNode(nodes))
	result := MaxSumPath(node)
	t.Log("expect: 42, real:", result)
}
