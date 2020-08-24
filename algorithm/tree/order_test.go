package tree

import "testing"

var nodes = []string{
	"1", "2", "4", "7", "8", "11", "null", "22",
}

func TestPre(t *testing.T) {
	node := SliceToTree(SliceToNode(nodes))

	m := Pre(node)
	t.Log(m)
}

func TestMidle(t *testing.T) {
	node := SliceToTree(SliceToNode(nodes))
	m := Midle(node)
	t.Log(m)
}

func TestLast(t *testing.T) {
	node := SliceToTree(SliceToNode(nodes))
	m := Last(node)
	t.Log(m)
}
