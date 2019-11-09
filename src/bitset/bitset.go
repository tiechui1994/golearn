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

// 使用 & 操作, 长度是最小值
func (set *BitSet) Intersect(other *BitSet) *BitSet {
	n := len(set.data)
	if len(set.data) < len(other.data) {
		n = len(other.data)
	}

	if n == 0 {
		return new(BitSet)
	}

	intersect := &BitSet{
		data: make([]int64, n),
	}

	for i := 0; i < n; i++ {
		intersect.data[i] = set.data[i] & other.data[i]
	}
	intersect.size = intersect.computeSize()
	return intersect
}

// 使用 | 操作, 长度是最大值
func (set *BitSet) Union(other *BitSet) *BitSet {
	return nil
}

// 使用 &^ 操作, 长度是被减结合决定
func (set *BitSet) Difference(other *BitSet) *BitSet {
	return nil
}

func (set *BitSet) Visit(skip func(int) bool) (abort bool) {
	d := set.data
	for i := 0; i < len(d); i++ {
		w := d[i]
		if w == 0 {
			continue
		}

		base := i << shift
		for w != 0 {
			// 获取首个非0的位置, 注意: 1011,其中1的位置是0,1,3
			val := bits.TrailingZeros64(uint64(w))
			if skip(base + val) {
				return true
			}

			// 将首个非0的位置设置为0
			w &^= 1 << uint64(val)
		}
	}

	return false
}
