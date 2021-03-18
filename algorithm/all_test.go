package algorithm

import (
	"math/rand"
	"testing"
)

func TestNewSkipList(t *testing.T) {
	sk := NewSkipList()
	seed := rand.New(rand.NewSource(100))
	for i := 10; i > 0; i-- {
		sk.Insert(seed.Intn(100))
	}

	sk.PrintSkipList()

	sk.Remove(60)

	sk.PrintSkipList()
}
