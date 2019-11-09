package bitset

import (
	"testing"
	"fmt"
	"math/bits"
)

func TestNewBitSet(t *testing.T) {
	bs := NewBitSet(1, 2, 3, 33, 282, 211, 21)
	bs.Visit(func(i int) bool {
		fmt.Printf("%v  ", i)
		return false
	})
}

func TestIndex(t *testing.T) {
	fmt.Println(index(10))
	fmt.Println(pos(10))
	// 1011
	w := 11
	for w != 0 {
		// 获取首个非0的位置
		val := bits.TrailingZeros64(uint64(w))
		fmt.Println(val)
		// 将首个非0的位置设置为0
		fmt.Println("spec", 1<<uint64(val), w&^0 == w)
		w &^= 1 << uint64(val)
	}
}
