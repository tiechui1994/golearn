package algorithm

import (
	"fmt"
)

type BTreeNode struct {
	n        int          // number of keys
	degree   int          // minimum degree(minimum children)
	leaf     bool         // true for leaf node
	keys     []int        // key slice
	children []*BTreeNode // child node slice
}

func NewBTreeNode(degree int, leaf bool) *BTreeNode {
	return &BTreeNode{
		n:        0,
		degree:   degree,
		leaf:     leaf,
		keys:     make([]int, 2*degree-1),
		children: make([]*BTreeNode, 2*degree),
	}
}

// 遍历所有以本节点为根的所有结点，打印它们的key
func (node *BTreeNode) Traverse() {
	var i int
	for i = 0; i < node.n; i++ {
		if !node.leaf {
			node.children[i].Traverse()
		}
		fmt.Printf("%d ", node.keys[i])
	}

	if node.leaf {
		return
	}
	// 遍历最后一个子节点
	node.children[i].Traverse()
}

// 在以本节点为根的结点中搜索包含指定关键字的结点以及关键字所在位置
func (node *BTreeNode) Search(k int) (*BTreeNode, int) {
	var i int
	// 找到第一个不大于key的关键字所在位置
	// 结束条件，i = node.n || node.keys[i] >= k
	for i = 0; i < node.n && node.keys[i] < k; i++ {
	}
	if i < node.n && node.keys[i] == k {
		return node, i
	}
	if node.leaf {
		return nil, 0
	}
	// 在孩子中查找
	// 包含 i == node.n 或 node.keys[i] > k
	return node.children[i].Search(k)
}
func (node *BTreeNode) insertNonFull(key int) {
	// 找到要插入的位置
	var i int
	for i < node.n && node.keys[i] < key {
		i++
	}
	// k 已经在本节点里存在
	if i < node.n && node.keys[i] == key {
		return
	}
	// 如果是叶节点则插入
	if node.leaf {
		// 将i之后的key右移
		copy(node.keys[i+1:], node.keys[i:node.n])
		// k放到位置i
		node.keys[i] = key
		node.n++
		return
	}
	var c = node.children[i]
	// 如果子节点已满，则分裂子节点
	if c.isFull() {
		node.splitChild(c, i)
		// 如果提升上来的key等于k则不插入
		if node.keys[i] == key {
			return
		}
		// 如果提升上来的新的key小于k则插入到它的右孩子
		if node.keys[i] < key {
			c = node.children[i+1]
		}
	}
	c.insertNonFull(key)
}

// c 为要分裂的子节点
// i 为父节点保存要提升的子节点key的位置
// 关键步骤：
// 	1. 生成新的结点将分裂结点后t个key和child复制过去
//  2. 父节点key和child数组右移
// 	3. 新节点和上升的key复制到父节点
func (node *BTreeNode) splitChild(c *BTreeNode, i int) {
	// 生成新的结点
	z := NewBTreeNode(node.degree, c.leaf)
	z.n, c.n = c.degree-1, c.degree-1
	// 将c的后t-1个key复制到新的结点
	// c中keys索引: 0..degree-2, degree-1, degree..2t-2
	for j := c.degree; j < 2*c.degree-1; j++ {
		z.keys[j-c.degree] = c.keys[j]
		c.keys[j] = 0
	}
	// 将s的keys里i位之后的元素右移
	for j := node.n - 1; j >= i; j-- {
		node.keys[j+1] = node.keys[j]
	}
	// 将中间的key复制到s的keys列表的位置i
	node.keys[i] = c.keys[c.degree-1]
	c.keys[c.degree-1] = 0
	// 将c的后t个孩子指针复制到新的结点
	// c中children索引: 0..degree-1, degree..2t-1
	if !c.leaf {
		for j := c.degree; j < 2*c.degree; j++ {
			z.children[j-c.degree] = c.children[j]
			c.children[j] = nil
		}
	}
	// 将s的i+1之后的孩子指针在列表中后移
	for j := node.n; j >= i+1; j-- {
		node.children[j+1] = node.children[j]
	}
	// 将新的结点指针放到s的孩子列表位置i
	node.children[i+1] = z
	node.n++
}

// 找到第一个大于key的关键字的位置
func (node *BTreeNode) findLargerKeyIndex(key int) (idx int) {
	for idx < node.n && node.keys[idx] < key {
		idx++
	}

	return idx
}

// 1. 如果key在节点中，并且该结点是leaf，则直接删除
// 2. 如果key在节点中，并且该结点是内结点，分如下情况
// 	2.1 若key所在的位置的左子树是满足最少t个关键字，则将其子树下最接近key的关键字key1复制到key然后在左子树递归删掉key1，否则
// 	2.2 若key所在的位置的右子树是满足最少t个关键字，则将其子树下最接近key的关键字key1复制到key然后在左子树递归删掉key1，否则
// 	2.3 若key的左右子树都少于t个关键字，则合并两子树合并并删掉key
// 3. 若key不在结点中，找到key最有可能在的子结点，并删除，删除之前判断如果子结点的key个数为t-1个
// 	3.1 如果其左兄弟有t个key则借一个key，如果右兄弟有t个key则从其借一个key
//  3.2 如果左或右兄弟都只有t-1个key，则从合并左或右兄弟
func (node *BTreeNode) delete(key int) {
	idx := node.findLargerKeyIndex(key)
	// key 在本结点内
	if idx < node.n && node.keys[idx] == key {
		switch {
		case node.leaf: // 情况1
			node.deleteFromLeaf(idx)
		case node.children[idx].n >= node.degree: // 情况2.1
			pre := node.getFrontSubTreeMax(idx)
			node.keys[idx] = pre
			node.children[idx].delete(pre)
		case node.children[idx+1].n >= node.degree: // 情况2.2
			succ := node.getBackSubTreeMin(idx)
			node.keys[idx] = succ
			node.children[idx+1].delete(succ)
		default: // 情况2.3
			node.merge(idx)
			node.children[idx].delete(key)
		}
		return
	}
	// 未在本节点并是leaf
	if node.leaf {
		return
	}
	// key在最后一个子节点子树中
	flag := idx == node.n
	// 情况3
	if node.children[idx].n < node.degree {
		node.fill(idx)
	}
	// fill时如果有merge操作，关键词个数减少
	// 如果关键字在最后一个子节点，但其被merge
	// 则需要在前一个子节点删除
	if flag && idx > node.n {
		node.children[idx-1].delete(key)
	} else {
		node.children[idx].delete(key)
	}
}

// 将第idx的子树从其兄弟中借一个key
func (node *BTreeNode) fill(idx int) {
	switch {
	// 有左子树且其关键字个数>=degree
	case idx != 0 && node.children[idx-1].n >= node.degree: // 3.1
		node.borrowFromFrontBrother(idx)
		// 有右子树且其关键字个数>=degree
	case idx != node.n && node.children[idx+1].n >= node.degree: // 3.1
		node.borrowFromBackBrother(idx)
		// idx在最后一个key位置
	case idx != node.n:
		node.merge(idx)
		// idx为key的最后一个
	default:
		node.merge(idx - 1)
	}
}

// 从左兄弟借一个key
func (node *BTreeNode) borrowFromFrontBrother(idx int) {
	cur := node.children[idx]
	left := node.children[idx-1]
	// 右移cur关键字为借的key留出空间
	copy(cur.keys[1:], cur.keys[0:cur.n])
	// s的关键字下移
	cur.keys[0] = node.keys[idx-1]
	// 左兄弟的最后一个key上升
	node.keys[idx-1] = left.keys[left.n-1]
	// 右移cur的子指针，为借的指针留出空间
	if !cur.leaf {
		copy(cur.children[1:], cur.children[0:cur.n+1])
		// 左兄弟最后一个child移到cur
		cur.children[0] = left.children[left.n]
	}
	left.n--
	cur.n++
}

// 从右兄弟借一个key
func (node *BTreeNode) borrowFromBackBrother(idx int) {
	cur := node.children[idx]
	right := node.children[idx+1]
	// s的关键字下移到cur的关键字尾部
	cur.keys[cur.n] = node.keys[idx]
	// 右兄弟的第一个关键词上移
	node.keys[idx] = right.keys[0]
	// 右兄弟的关键字列表左移
	copy(right.keys[0:], right.keys[1:right.n])
	// 右兄弟的第一个子指针借到cur的最后一个child
	if !cur.leaf {
		cur.children[cur.n+1] = right.children[0]
		// 有兄弟的子指针列表左移
		copy(right.children[:], right.children[1:right.n+1])
	}
	right.n--
	cur.n++
}

// 合并key的左右子树
func (node *BTreeNode) merge(idx int) {
	left := node.children[idx]
	right := node.children[idx+1]
	// s的关键字下降
	left.keys[left.degree-1] = node.keys[idx]
	// 复制右子t-1个key
	for i := 0; i < right.degree-1; i++ {
		left.keys[left.degree+i] = right.keys[i]
	}
	if !left.leaf {
		// 复制右子t个child
		for i := 0; i < right.degree; i++ {
			left.children[left.n+i+1] = right.children[i]
		}
	}
	left.n += right.degree
	// s的key复制
	for i := idx; i < node.n-1; i++ {
		node.keys[i] = node.keys[i+1]
	}
	// s的child复制
	for i := idx + 1; i < node.n; i++ {
		node.children[i] = node.children[i+1]
	}
	node.n--
}

// 从左子树取最大值
func (node *BTreeNode) getFrontSubTreeMax(idx int) int {
	cur := node.children[idx]
	for !cur.leaf {
		cur = cur.children[cur.n]
	}
	return cur.keys[cur.n-1]
}

// 从右子树取最小值
func (node *BTreeNode) getBackSubTreeMin(idx int) int {
	cur := node.children[idx+1]
	for !cur.leaf {
		cur = cur.children[0]
	}
	return cur.keys[0]
}
func (node *BTreeNode) deleteFromLeaf(idx int) {
	copy(node.keys[idx:], node.keys[idx+1:node.n])
	node.n--
}
func (node *BTreeNode) isFull() bool {
	return node.n == 2*node.degree-1
}

type BTree struct {
	root   *BTreeNode
	degree int
}

func NewBTree(degree int) *BTree {
	return &BTree{degree: degree}
}

// 单程下行方式遍历树插入关键字。
// 关键方法是"主动分裂", 即，在遍历一个子节点前，如果子节点已满则对其进行分裂。
// 相反的"被动分裂"则是在要插入的时候遇满才分裂，会出现重复遍历的情况。
// 比如，从根节点到叶节点都是满的，当到达叶节点要发现其已满需进行分裂，
// 提升一个关键字到父节点，发现父节点也已满需要分裂父节点，重复下去一直到根节点。
// 这样就出现了从根节点到叶节点的重复遍历。而"主动分裂"则不出有出现这种情况，因为
// 在分裂一个子节点时候父节点是已经有足够空间容纳要提升的新的key了
//
// B数的增高依赖的是root结点分裂
// B数新增的关键字都增加到了叶节点上
func (tree *BTree) Insert(key int) {
	if tree.root == nil {
		tree.root = NewBTreeNode(tree.degree, true)
	}

	var root = tree.root

	// root已满则分裂root
	if root.isFull() {
		newroot := NewBTreeNode(tree.degree, false)

		newroot.children[0] = root  // 老的root称为新结点的孩子
		newroot.splitChild(root, 0) // 分裂老的root，并将一个key提升到新的root

		// 新的root有两个孩子，决定将key插入到哪个孩子
		var i int
		if key > newroot.keys[0] {
			i = 1
		}
		newroot.children[i].insertNonFull(key)

		tree.root = newroot

		return
	}

	tree.root.insertNonFull(key)
}
func (tree *BTree) Delete(k int) {
	if tree.root == nil {
		return
	}
	tree.root.delete(k)
}
func (tree *BTree) Traverse() {
	if tree.root == nil {
		return
	}
	tree.root.Traverse()
	fmt.Printf("\n")
}
