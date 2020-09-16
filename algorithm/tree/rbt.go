package tree

import (
	"fmt"
	"bytes"
	"log"
)

/*
红黑树:
1. 根节点永远是黑色的；
2. 所有的叶节点都是是黑色的(指的是nil节点)
3. 任何相邻的两个节点不能同时为红色; (每个红色节点的两个子节点一定都是黑色)
4. 从任一节点到其子树中每个叶子节点的路径都包含相同数量的"黑色节点";
*/

type Color string

const (
	Red   = Color("Red")
	Black = Color("Black")
)

type rbNode struct {
	val    int
	color  Color
	left   *rbNode
	right  *rbNode
	parent *rbNode
}

func (n *rbNode) getUncle() *rbNode {
	if n.parent == nil || n.parent.parent == nil {
		return nil
	}

	if n.parent.parent.left == n.parent {
		return n.parent.parent.right
	}

	return n.parent.parent.left
}

func (n *rbNode) String() string {
	return fmt.Sprintf("%v", n.val)
}

/*
插入: 颜色变换 -> 旋转

注: 插入节点的颜色一定是红色

颜色变换:

1. 新插入节点的父节点是黑色, 则修复完成.

2. 新插入节点的父节点是红色
2.1 叔叔是红色.
a) 父节点和叔叔变黑, 祖父变红
b) 祖父变成新节点

2.2 叔叔不存在, 祖孙三代处于同一条斜线上.
a) 父节点变黑色, 祖父节点变红
b) 根据"父节点"向相反的方向旋转(斜线)

2.2 叔叔不存在, 祖孙三代没有在统一斜线上.
a) 根据"当前节点"向相反的方向旋转(父子斜线), 情况转变为状况2.2
注意: 此种状况下,没有改变颜色
*/

type RBTree struct {
	root *rbNode
}

func (r *RBTree) insert(val int) bool {
	node := &rbNode{val: val, color: Red}
	defer func() {
		log.Println(node.parent, node.left, node.right, node, node.color)
	}()

	if r.root == nil {
		r.root = node
		node.color = Black
	} else {
		parent := r.findParent(node)
		if node.val == parent.val {
			return true
		}

		node.parent = parent
		if node.val < parent.val {
			parent.left = node
		} else {
			parent.right = node
		}

		log.Println("fix", node, node.parent, node.parent.color)
		r.fixInsert(node)
	}

	return false
}

func (r *RBTree) findParent(node *rbNode) *rbNode {
	parent := r.root
	cur := r.root

	for cur != nil {
		if cur.val == node.val {
			return cur
		}

		if cur.val > node.val {
			parent = cur
			cur = cur.left
		} else if cur.val < node.val {
			parent = cur
			cur = cur.right
		}
	}

	return parent
}

func (r *RBTree) fixInsert(node *rbNode) {
	parent := node.parent

	for parent != nil && parent.color == Red {
		uncle := node.getUncle()

		// 若 uncle 存在, uncle 颜色变成黑色(必然结果). 如果是黑色(正好), 如果是红色(当前需要变色)
		// 直接变色, 然后祖先节点变为当前的节点
		if uncle != nil {
			uncle.color = Black       // 叔叔
			parent.color = Black      // 父亲
			parent.parent.color = Red // 祖先
			node = parent.parent      // 更改当前节点
			parent = node.parent
			continue
		}

		// todo: 线条, 先看祖先, 再看父子
		// 若 uncle 不存在, 需要旋转
		ancestor := parent.parent
		if parent == ancestor.left {
			inLine := node == parent.left // 祖先, 父亲, 孩子 是三者一线

			// 当前节点是右孩子, 即祖孙三代不同线, "当前节点"为旋转点, 相同的方向旋转(左旋)
			if !inLine {
				r.leftRoate(parent)
			}

			// 当前节点是左孩子, 即祖孙三代一条线, "父节点"为旋转点, 相反的方向旋转(右旋)
			r.rightRoate(ancestor)

			// 变色完成, 结束(祖先变为红色, 父亲变为黑色)
			if !inLine {
				node.color = Black
				parent = nil
			} else {
				parent.color = Black
			}

			ancestor.color = Red
		} else {
			inLine := node == parent.right // 父亲是否是左孩子

			// 当前节点是左孩子, 即祖孙三代不同线, 当前节点为旋转点, 相同的方向旋转(右旋)
			if !inLine {
				r.rightRoate(parent)
			}

			// 当前节点是右孩子, 即祖孙三代一条线, 父节点为旋转点, 相反的方向旋转(左旋)
			r.leftRoate(ancestor)

			// 变色完成, 结束
			if !inLine {
				node.color = Black
				parent = nil
			} else {
				parent.color = Black
			}
			ancestor.color = Red
		}
	}

	// 每次都设置 root 节点
	r.root.color = Black
	r.root.parent = nil
}

// 左旋, 旋转杆 parent - right
func (r *RBTree) leftRoate(parent *rbNode) {
	if parent.right == nil {
		panic("right is null")
	}

	ancestor := parent.parent // 祖先节点
	roate := parent.right     // 旋转点

	// 调整父节点的右孩子
	parent.right = roate.left
	if roate.left != nil {
		roate.left.parent = parent
	}

	// 调整旋转点的左孩子
	roate.left = parent
	parent.parent = roate

	// 调整祖先节点
	if ancestor == nil {
		r.root = roate
		roate.parent = nil
	} else {
		if ancestor.left == parent {
			ancestor.left = roate
			roate.parent = ancestor
		} else {
			ancestor.right = roate
			roate.parent = ancestor
		}
	}
}

// 左旋, 旋转杆 parent - left.
func (r *RBTree) rightRoate(parent *rbNode) {
	if parent.left == nil {
		panic("left is nil")
	}

	ancestor := parent.parent
	roate := parent.left

	parent.left = roate.right
	if roate.right != nil {
		roate.right.parent = parent
	}

	roate.right = parent
	parent.parent = roate

	if ancestor == nil {
		r.root = roate
		roate.parent = nil
	} else {
		if ancestor.left == parent {
			ancestor.left = roate
			roate.parent = ancestor
		} else {
			ancestor.right = roate
			roate.parent = ancestor
		}
	}
}

func (r *RBTree) Println() {
	if r.root == nil {
		return
	}

	var buf bytes.Buffer
	queue := []*rbNode{r.root}
	for len(queue) != 0 {
		var temp []*rbNode

		for len(queue) != 0 {
			cur := queue[0]
			queue = queue[1:]

			if cur == nil {
				buf.WriteString("\n")
				continue
			}

			var pos string
			if cur.parent != nil {
				if cur.parent.left == cur {
					pos = " LE"
				} else {
					pos = " RI"
				}
			}

			var parent string
			if cur.parent != nil {
				parent = cur.parent.String()
			}

			var color = "R"
			if cur.color == Black {
				color = "B"
			}
			if cur.parent != nil {
				color += " "
			}

			buf.WriteString(fmt.Sprintf("%v(%v)\t", cur.String(), color+parent+pos))

			if cur.left != nil {
				temp = append(temp, cur.left)
			}
			if cur.right != nil {
				temp = append(temp, cur.right)
			}
		}

		queue = temp
	}

	log.Println(buf.String())
}


/*
[doc](https://tech.meituan.com/2016/12/02/redblack-tree.html)

删除:

如果是叶子节点就直接删除, 如果是非叶子节点, 要使用中序遍历的后续节点顶替要删除的位置. 删除后需要做修复操作.

删除修复操作在遇到被删除的节点是红色节点或者到达root节点, 修复完毕.

删除修复操作是针对删除黑色节点才有的. 处理思想是从兄弟节点上借调黑色节点过来, 如果兄弟节点没有黑色节点可以借调, 只能往上
追溯, 将每一级的黑节点数减去一个, 使得整棵树符合定义.

1. 待删除的节点的兄弟节点是红色

2. 待删除的节点的兄弟节点是黑色

2.1 兄弟节点的子节点都是黑色.

2.2 兄弟节点的子节点中左子为红色, 右子是黑色

2.3 兄弟节点的子节点中左子为红色, 右子为红色

*/