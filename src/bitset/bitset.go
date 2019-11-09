package bitset

import "math/bits"

type BitSet struct {
	data []int64
	size int
}

const (
	shift = 6
	mask  = 0x3f
)

func index(n int) int {
	return n >> shift
}

func pos(n int) int64 {
	// n & mask 等价于 n % 64
	// 1 << m, 将1向左移动m位, 其实质是就算所对应的值
	return 1 << uint(n&mask)
}

func NewBitSet(n ...int) *BitSet {
	if len(n) == 0 {
		return new(BitSet)
	}

	// 最大值确定数组的大小
	max := n[0]
	for _, v := range n {
		if v > max {
			max = v
		}
	}

	if max < 0 {
		return new(BitSet)
	}

	s := &BitSet{
		data: make([]int64, index(max)+1),
	}

	for _, v := range n {
		if v >= 0 {
			s.data[index(v)] |= pos(v)
			s.size++
		}
	}

	return s
}

func (set *BitSet) Contains(n int) bool {
	if n < 0 {
		return false
	}

	index := index(n)
	if index >= len(set.data) {
		return false
	}

	return set.data[index]&pos(n) != 0
}

func (set *BitSet) Clear(n int) *BitSet {
	if n < 0 {
		return set
	}

	index := index(n)
	if index >= len(set.data) {
		return set
	}

	if set.data[index]&pos(n) != 0 {
		set.data[index] &^= pos(n)
		set.size --
		set.trim()
	}

	return set
}

func (set *BitSet) trim() {
	d := set.data
	n := len(d) - 1
	for n >= 0 && d[n] == 0 {
		n--
	}
	set.data = d[:n+1]
}

func (set *BitSet) Add(n int) *BitSet {
	if n < 0 {
		return set
	}

	index := index(n)
	if index >= len(set.data) {
		data := make([]int64, index+1)
		copy(data, set.data)
		set.data = data
	}

	if set.data[index]&pos(n) == 0 {
		set.data[index] |= pos(n)
		set.size++
	}

	return set
}

func (set *BitSet) computeSize() int {
	d := set.data
	n := 0
	for i, l := 0, len(d); i < l; i++ {
		if w := uint64(d[i]); w != 0 {
			n += bits.OnesCount64(w)
		}
	}

	
	return n
}
