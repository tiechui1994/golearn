package tree

import "testing"

func TestMaxSumPath(t *testing.T) {
	nodes := []string{"5", "4", "8", "11", "null", "13", "4", "7", "2", "null", "null", "null", "1"}
	node := SliceToTree(SliceToNode(nodes))
	result := MaxSumPath(node)
	t.Log("expect: 42, real:", result)
}
