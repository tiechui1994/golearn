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
	seeds := []int{
		100, 810, 295, 131, 529, 225, 263, 481, 80, 406,
		502, 210, 51, 350, 815, 110, 788, 424, 510, 592,
		256, 420, 99, 458, 185, 974, 956, 894, 518, 0,
		151, 35, 773, 186, 973, 476, 931, 466, 652, 8,
		622, 591, 362, 669, 717, 590, 928, 942, 928, 950,
		534, 532, 615, 812, 361, 728, 845, 330, 317, 756,
		658, 633, 917, 926, 617, 479, 948, 574, 987, 799,
		752, 21, 292, 802, 842, 789, 587, 208, 98, 316,
		515, 489, 958, 287, 88, 702, 886, 468, 197, 590,
		860, 652, 42, 194, 820, 892, 47, 95, 788, 731,
	}
	_ = seeds
	for i := 0; i < 999; i++ {
		rnd := int(rand.Intn(1000000) + 1)
		nodes = append(nodes, rnd)
		rbt.insert(rnd)
		if !rbt.Valid() {
			log.Println("invalid insert", rbt.Println())
			panic("invalid")
		}
	}

	for len(nodes) > 0 {
		log.Println("cur size", len(nodes))
		idx := int(len(nodes) % 11)
		if idx == len(nodes) {
			idx -= 1
		}
		rbt.remove(nodes[idx])
		if !rbt.Valid() {
			log.Println("invalid remove", rbt.Println())
			panic("invalid")
		}
		nodes = append(nodes[0:idx], nodes[idx+1:]...)
	}
}
