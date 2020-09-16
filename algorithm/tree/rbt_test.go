package tree

import (
	"testing"
	"log"
)

func TestRBT(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	rbt := RBTree{}

	rbt.insert(4)
	rbt.insert(3)
	rbt.insert(2)
	rbt.insert(6)
	rbt.insert(7)

	rbt.Println()

}
