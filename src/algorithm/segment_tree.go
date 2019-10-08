package algorithm

import "math"

type SegNode struct {
	left  *SegNode
	right *SegNode

	start int
	end   int
	max   int
}

func NewSegTree(slice []int) *SegNode {
	return build(0, len(slice)-1, slice)
}

func build(left, right int, slice []int) *SegNode {
	if left > right {
		return nil
	}

	node := &SegNode{start: left, end: right}
	if left == right {
		node.max = slice[left]
		return node
	}

	mid := (left + right) / 2
	node.left = build(left, mid, slice)
	node.right = build(mid+1, right, slice)
	node.max = max(node.left.max, node.right.max)

	return node
}

func (seg *SegNode) Query(start, end int) int {
	if start < seg.start || seg.end > end {
		return -1
	}

	return seg.query(seg, start, end)
}

func (seg *SegNode) query(node *SegNode, start, end int) int {
	if start <= node.start && node.end <= end {
		return node.max
	}

	val := int(math.MaxInt64)
	mid := (node.start + node.end) / 2
	if start <= mid {
		val = max(val, seg.query(node.left, start, end))
	}
	if mid+1 <= end {
		val = max(val, seg.query(node.right, start, end))
	}

	return val
}

func max(x, y int) int {
	if x >= y {
		return x
	}
	return y
}

func (seg *SegNode) Update(index int, val int) {
	seg.update(seg, index, val)
}

func (seg *SegNode) update(node *SegNode, index, val int) {
	if node.start == node.end && node.start == index {
		node.max = val
		return
	}

	mid := (node.start + node.end) / 2
	if index <= mid {
		seg.update(node.left, index, val)
		node.max = max(node.left.max, node.max)
	} else {
		seg.update(node.right, index, val)
		node.max = max(node.right.max, node.max)
	}
}
