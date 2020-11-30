package tree

import "testing"

var nodes = []string{
	"1", "2", "4", "7", "8", "11", "null", "22",
}

func TestPre(t *testing.T) {
	node := SliceToTree(SliceToNode(nodes))

	m := PreOrder(node)
	t.Log(m)
}

func TestMidle(t *testing.T) {
	node := SliceToTree(SliceToNode(nodes))
	m := MidleOrder(node)
	t.Log(m)
}

func TestLast(t *testing.T) {
	node := SliceToTree(SliceToNode(nodes))
	m := LastOrder(node)
	t.Log(m)
}

func TestLevel(t *testing.T) {
	node := SliceToTree(SliceToNode(nodes))
	m := LevelOrder(node)
	t.Log(m)
}
