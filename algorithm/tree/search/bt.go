package search

import "fmt"

const m = 4         // M阶B树
const min = m/2 - 1 // 每个节点至少有的关键字个数

type BT struct {
	parent *BT        //指向父节点的指针
	keyNum int        //关键字个数
	key    [m + 1]int //关键字向量,key[0]未用
	child  [m + 1]*BT //子树指针向量
}

//在B树中查找关键字为value的节点，查找成功，返回在节点的位置和该节点
func (t *BT) Search(value int) (bool, int, *BT) {
	node := &BT{}
	var i int
	if t == nil {
		return false, 0, nil
	}
	for t != nil {
		for i = t.keyNum; i > 0 && value <= t.key[i]; i-- {
			if value == t.key[i] {
				return true, i, t
			}
		}
		if t.child[i] == nil {
			node = t
		}
		t = t.child[i]
	}
	return false, i, node
}

//分裂节点t
func (t *BT) Split() *BT {
	newNode := BT{}
	parent := t.parent
	if parent == nil {
		parent = &BT{}
	}
	mid := t.keyNum/2 + 1
	newNode.keyNum = m - mid
	t.keyNum = mid - 1
	j := 1
	k := mid + 1
	for ; k <= m; k++ { //新生成的右节点
		newNode.key[j] = t.key[k]
		newNode.child[j-1] = t.child[k-1]
		j = j + 1
	}
	newNode.child[j-1] = t.child[k-1]
	newNode.parent = parent
	t.parent = parent
	//将该节点中间节点插入到父节点
	k = parent.keyNum
	for ; t.key[mid] < parent.key[k]; k-- {
		parent.key[k+1] = parent.key[k]
		parent.child[k+1] = parent.child[k]
	}
	parent.key[k+1] = t.key[mid]
	parent.child[k] = t
	parent.child[k+1] = &newNode
	parent.keyNum++

	if parent.keyNum >= m {
		return parent.Split()
	}
	return parent
}

/*
在树中插入关键字value
1、先查看树中是否有此关键字
2、通过第1步也能确定会插入到哪个节点中
2、查看插入节点后是否需要分裂节点
*/
func (t *BT) Insert(value int) *BT {
	var i int
	ok, _, node := t.Search(value) //1、先在整颗树中找到要插入到哪个节点中
	if !ok {                       //树中不存在此节点
		node.key[0] = value
		for i = node.keyNum; i > 0 && value < node.key[i]; i-- {
			node.key[i+1] = node.key[i]
		}
		node.key[i+1] = value
		node.keyNum++
		if node.keyNum < m { //没有超过节点最大数，结束插入
			return t
		} else { //否则，分裂该节点
			parent := node.Split()
			for parent.parent != nil {
				parent = parent.parent
			}
			return parent
		}
	}
	return t
}

/*删除关键字
1）要删除的key位于非叶子结点上，则用后继key覆盖要删除的key，然后删除该后继key。
2）该结点key个数大于等于Math.ceil(m/2)-1，结束删除操作，否则执行第3步。
3）如果兄弟结点key个数大于Math.ceil(m/2)-1，则父结点中的key下移到该结点，
3）兄弟结点中的一个key上移，删除操作结束。
否则，将父结点中的key下移与当前结点及它的兄弟结点中的key合并，形成一个新的结点。
然后当前结点的指针指向父结点，重复上第2步。
*/

//在整颗树中找到要删除的关键字在哪个节点中
func (t *BT) Delete(value int) *BT {
	ok, i, node := t.Search(value) //先
	if ok {
		t = node.DeleteNode(value, i) //在这个节点中删除这个节点
	}
	return t
}

//删除关键字
func (t *BT) DeleteNode(value int, i int) *BT {
	if t.child[i] != nil { //非终端节点
		valueTemp, nodeTemp := t.FindAfterMinNode(i) //找到后继最下层非终端结点的最小关键字和相应节点
		t.key[i] = *valueTemp                        //将最下层非终端结点中的最小关键字取代要删除的关键字
		nodeTemp.DeleteNode(*valueTemp, 1)           // 然后删除最下层非终端结点中的最小关键字
	} else {
		for k := i; k < t.keyNum; k++ {
			t.key[k] = t.key[k+1]
			t.child[k] = t.child[k+1]
		}
		t.keyNum--
		if t.keyNum < (m-1)/2 && t.parent != nil {
			ok, t := t.Restore()
			if !ok {
				t = t.MergeNode()
			}
		}
	}
	for t.parent != nil {
		t = t.parent
	}
	return t
}

//调整B树,该节点与父亲节点和兄弟节点之间的调整
func (t *BT) Restore() (bool, *BT) {
	parent := t.parent
	j := 0
	for ; parent.child[j] != t; j++ {
	}
	if j > 0 { //p有左邻兄弟节点
		b := parent.child[j-1]
		if b.keyNum > (m-1)/2 { //左兄弟有多余关键字
			for k := t.keyNum; k >= 0; k-- { //将t中关键字和指针都右移动,给父节点移动下来的关键字空出位置
				t.key[k+1] = t.key[k]
				//t.child[k+1] = t.child[k]
			}
			t.key[1] = parent.key[j]
			parent.key[j] = b.key[b.keyNum] //b中关键字上移至parent
			t.keyNum++
			b.keyNum--
			return true, parent
		}
	}
	if j < parent.keyNum { //p有右邻兄弟节点
		b := parent.child[j+1]
		if b.keyNum > (m-1)/2 { //右邻兄弟有多余关键字
			t.key[t.keyNum+1] = parent.key[j+1] //父节点关键字下移
			parent.key[j+1] = b.key[1]          //兄弟节点关键字上移
			for k := 1; k < b.keyNum; k++ {
				b.key[k] = b.key[k+1]
			}
			t.keyNum++
			b.keyNum--
			return true, parent
		}
	}
	return false, t //没有调整成功,需要合并
}

func (t *BT) MergeNode() *BT {
	b := &BT{}
	parent := t.parent
	j := 0
	for ; parent.child[j] != t; j++ {
	}
	if j > 0 { //t有左兄弟，与左兄弟合并
		b = parent.child[j-1]
		b.keyNum++
		b.key[b.keyNum] = parent.key[j]  //父结点中关键字下移
		for k := 1; k <= t.keyNum; k++ { //t节点中关键字移到左兄弟节点中
			b.keyNum++
			b.key[b.keyNum] = t.key[k]
		}
		parent.keyNum--
		for k := j; k < parent.keyNum; k++ { //改变父节点
			parent.key[k] = parent.key[k+1]
			parent.child[k] = parent.child[k+1]
		}
	} else { //与右兄弟合并
		b = parent.child[j+1]
		t.keyNum++
		t.key[t.keyNum] = parent.key[j]  //父结点中关键字下移
		for k := 1; k <= b.keyNum; k++ { //兄弟节点中关键字移到t节点中
			t.keyNum++
			t.key[t.keyNum] = b.key[k]
		}
		parent.keyNum--
		for k := j; k <= parent.keyNum; k++ { //改变父节点
			parent.key[k] = parent.key[k+1]
			parent.child[k] = parent.child[k+1]
		}
	}

	return parent
}

//查找后继最下层非终端结点的最小关键字
func (t *BT) FindAfterMinNode(i int) (*int, *BT) {
	leaf := t
	if leaf == nil {
		return nil, nil
	} else {
		leaf = leaf.child[i]
		for leaf.child[0] != nil {
			leaf = leaf.child[0]
		}
	}
	return &leaf.key[1], leaf
}

func (t *BT) BTreeTraverse() {
	queue := []*BT{}
	queue = append(queue, t)
	i := 0 //当前要出队列的下标s
	for i < len(queue) {
		current := queue[i]
		i = i + 1
		fmt.Print("[")
		for k := 1; k <= current.keyNum; k++ {
			fmt.Printf(" %d ", current.key[k])
		}
		fmt.Print("]\n")
		for k := 0; k <= current.keyNum; k++ {
			if current.child[k] != nil {
				queue = append(queue, current.child[k])
			}
		}
	}
}
