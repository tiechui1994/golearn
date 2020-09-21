package tree

import (
	"testing"
	"log"
)

func TestRBT(t *testing.T) {
	log.SetFlags(log.Lshortfile|log.Lmicroseconds)


	rbt := RBTree{}
	rbt.insert(4)
	rbt.insert(3)
	rbt.insert(2)
	rbt.insert(6)
	rbt.insert(7)
	rbt.insert(5)

	rbt.insert(1)
	rbt.insert(8)
	rbt.insert(11)


	rbt.insert(9)


	rbt.insert(16)
	rbt.insert(10)
	rbt.insert(12)
	rbt.insert(13)

	log.Println(rbt.Println())

	rbt.remove(7)
	log.Println(rbt.Println())

	rbt.remove(3)
	log.Println(rbt.Println())


	rbt.remove(5)
	log.Println(rbt.Println())
}
