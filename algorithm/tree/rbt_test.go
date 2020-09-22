package tree

import (
	"testing"
	"log"
	"math/rand"
)

func TestRBT(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)

	rbt := RBTree{}
	nodes := make([]int, 0, 100)
	for i := 0; i < 20; i++ {
		rnd := int(rand.Int31n(8000) + 1)
		nodes = append(nodes, rnd)
		rbt.insert(rnd)
	}

	for len(nodes) > 5 {
		idx := int(rand.Int31n(int32(len(nodes))))
		log.Println("Before", nodes[idx])
		rbt.remove(nodes[idx])
		log.Println("After", rbt.Println())
		nodes = append(nodes[0:idx], nodes[idx+1:]...)
	}
}
