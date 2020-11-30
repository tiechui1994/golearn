package algorithm

import (
	"math/rand"
	"testing"
	"time"
)

func TestNewSkipList(t *testing.T) {
	skiplist := NewSkipList()
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 100; i > 0; i-- {
		skiplist.Insert(seed.Intn(100))
	}

	skiplist.PrintSkipList()
}
