package tree

import (
	"log"
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
	res := PathSumI(node, 22)
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

func TestBuildTree(t *testing.T) {
	res := FindMinHeightTrees(6, [][]int{
		{0, 3}, {1, 3}, {2, 3}, {4, 3}, {5, 4},
	})
	t.Log("expect:[3,4], real:", res)

	res = FindMinHeightTrees(4, [][]int{
		{1, 0}, {1, 2}, {1, 3},
	})
	t.Log("expect: [1], real:", res)

	res = FindMinHeightTrees(6, [][]int{
		{0, 1}, {0, 2}, {0, 3}, {3, 4}, {4, 5},
	})
	t.Log("expect:[3], real:", res)
}

func TestPathSum2(t *testing.T) {
	nodes := []string{"5", "4", "8", "11", "null", "13", "4", "7", "2", "null", "null", "5", "1"}
	node := SliceToTree(SliceToNode(nodes))
	log.Println(PathSumII(node, 17))

	//nodes := []string{"1", "null", "2", "null", "3", "null", "4", "null", "5"}
	//node := SliceToTree(SliceToNode(nodes))
	//PathSumII(node, 3)

	//nodes := []string{"0", "1","1"}
	//node := SliceToTree(SliceToNode(nodes))
	//pathSum(node, 1)
}

func TestDelNodes(t *testing.T) {
	nodes := []string{"1", "2", "null", "4", "3"}
	node := SliceToTree(SliceToNode(nodes))
	DelNodes(node, []int{2, 3})
}

func TestMinTime(t *testing.T) {
	res := MinTime(7, [][]int{{0, 1}, {0, 2}, {1, 4}, {1, 5}, {2, 3}, {2, 6}}, []bool{false, false, true, false, true, true, false})
	log.Println(res)
}

func TestFindMode(t *testing.T) {
	nodes := []string{"1", "null", "2", "2"}
	node := SliceToTree(SliceToNode(nodes))
	FindMode(node)
}

func TestLongestZigZag(t *testing.T) {
	nodes1 := []string{"1", "null", "1", "1", "1", "null", "null", "1", "1", "null", "1", "null", "null", "null", "1", "null", "1"}
	nodes2 := []string{"1", "1", "1", "null", "1", "null", "null", "1", "1", "null", "1"}
	node1 := SliceToTree(SliceToNode(nodes1))
	t.Log("expect:3, real:", LongestZigZag(node1))

	node2 := SliceToTree(SliceToNode(nodes2))
	t.Log("expect:4, real:", LongestZigZag(node2))
}

func TestBtreeGameWinningMove(t *testing.T) {
	RecoverFromPreorder("1-2--3---4-5--6---7")
}
