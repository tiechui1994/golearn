package algorithm

import (
	"fmt"
	"math/rand"
)

const SKIPLIST_MAXLEVEL = 32 //8
const SKIPLIST_P = 4

type Node struct {
	Next  []Node
	Value interface{}
}

func NewNode(v interface{}, level int) *Node {
	return &Node{Value: v, Next: make([]Node, level)}
}

type SkipList struct {
	Header *Node
	Level  int
}

func NewSkipList() *SkipList {
	return &SkipList{Level: 1, Header: NewNode(0, SKIPLIST_MAXLEVEL)}
}

/*
数据结构:

       +---+                      +---+
level2 | *-|--------------------->| * |
       +---+             +---+    +---+
level1 | *-|------------>| *-|--->| * |
       +---+    +---+    +---+    +---+    +---+
level0 | *-|--->| *-|--->| * |--->| *-|--->| * |
       +---+    +---+    +---+    +---+    +---+
       |   |    | 3 |    | 9 |    | 10|    | 12|
       +---+    +---+    +---+    +---+    +---+
       header
*/

func (skipList *SkipList) Insert(key int) {
	update := make(map[int]*Node) // level - 记录当前key前面的那个节点
	node := skipList.Header

	for i := skipList.Level - 1; i >= 0; i-- {
		for {
			if node.Next[i].Value != nil && node.Next[i].Value.(int) < key {
				node = &node.Next[i]
			} else {
				break
			}
		}

		update[i] = node
	}

	// 更新SkipList的level
	level := skipList.Random_level()
	if level > skipList.Level {
		for i := skipList.Level; i < level; i++ {
			update[i] = skipList.Header
		}
		skipList.Level = level
	}

	// 构建新的SkipList
	newNode := NewNode(key, level)
	for i := 0; i < level; i++ {
		newNode.Next[i] = update[i].Next[i]
		update[i].Next[i] = *newNode
	}

}

func (skipList *SkipList) Random_level() int {
	level := 1
	for (rand.Int31()&0xFFFF)%SKIPLIST_P == 0 {
		level += 1
	}
	if level < SKIPLIST_MAXLEVEL {
		return level
	} else {
		return SKIPLIST_MAXLEVEL
	}
}

func (skipList *SkipList) PrintSkipList() {

	for i := skipList.Level - 1; i >= 0; i-- {
		fmt.Println("level:", i+1)
		node := skipList.Header.Next[i]
		for {
			if node.Value != nil {
				fmt.Printf("%d ", node.Value.(int))
				node = node.Next[i]
			} else {
				break
			}
		}
		fmt.Println("\n--------------------------------------------------------")
	}

	fmt.Println("Current MaxLevel:", skipList.Level)
}

func (skipList *SkipList) Search(key int) *Node {
	node := skipList.Header
	for i := skipList.Level - 1; i >= 0; i-- {
		for {
			if node.Next[i].Value == nil {
				break
			}
			if node.Next[i].Value.(int) == key {
				return &node.Next[i]
			} else if node.Next[i].Value.(int) < key {
				node = &node.Next[i]
			} else {
				break
			}
		}
	}

	return nil
}

func (skipList *SkipList) Remove(key int) {
	update := make(map[int]*Node) // level - key的前一个节点
	node := skipList.Header
	for i := skipList.Level - 1; i >= 0; i-- {
		for {
			if node.Next[i].Value == nil {
				break
			}

			if node.Next[i].Value.(int) == key {
				update[i] = node
				break
			} else if node.Next[i].Value.(int) < key {
				node = &node.Next[i]
			} else {
				break
			}
		}
	}

	for i, v := range update {
		if v == skipList.Header {
			skipList.Level--
		}
		v.Next[i] = v.Next[i].Next[i]
	}
}
