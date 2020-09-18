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

1. 待删除的节点的兄弟节点是红色.
思路: 兄弟节点是红色, 无法借调黑色节点, 因此需要将兄弟提升为父节点, 由于兄弟是红色的, 根据定义, 兄弟节点的子节点是黑色的
就可以从它的子节点借调了. (兄弟相反的方向旋转, 旋转点是兄弟. 父节点和兄弟节点交换颜色)

这个时候, 情况就会转换为 2, 3 情况了.


2. 待删除的节点的兄弟节点是黑色
2.1 兄弟节点中的子节点都是黑色.

思路: 兄弟可以消除一个黑色节点, 因为兄弟子节点都是黑色的, 所以可以将兄弟节点变红, 从而保证局部树的局部颜色符合定义. 这时
候将父节点变成新的节点, 继续向上调整. (兄弟变成红色)


3. 待删除的节点的兄弟节点是黑色(相反)
3.1 兄弟在右边, 兄弟的左子节点为红色
3.2 兄弟在左边, 兄弟的右子节点为红色

思路: 将左边的红色节点借调过来, 从而达到状态的转换. 使得转换为状态 4.
以红色节点为点, 向相反的方向旋转. 红色变黑色, 兄弟变红色

4. 待删除的节点的兄弟节点是黑色(相同)
3.3 兄弟在右边, 兄弟的右子节点为红色
3.4 兄弟在左边, 兄弟的左子节点为红色

思路: 真正的节点借调操作, 将"兄弟节点"以及"兄弟的红色节点"借调过来.
*/

/*
删除的思路: 先寻找到删除的节点; 然后使用 "删除节点右子树当中最小的节点" 或者 "删除节点的左孩子节点" 替换当前删除的节点;
替换完成之后, 如果被替换的节点是黑色, 需要从替换的节点或者其其父亲向上递归修复. 修复就是上述提到的 4 种大情况.
*/

func (r *RBTree) remove(val int) int {
	parent := r.root
	node := r.root

	for node != nil {
		if node.val > val {
			parent = node
			node = node.left
		} else if node.val < val {
			parent = node
			node = node.right
		} else {
			if node.right != nil {
				// 右孩子存在, 寻找右子树最小的节点
				min := r.removeMin(node.right)

				var (
					isParent = false
					x        = min.left
				)
				if x == nil {
					isParent = true
					x = min.parent
				}

				// 使用 min 替换当前节点 node, 分别是left, right, parent, color 替换
				min.left = node.left
				node.left.parent = min

				if min != node.right {
					min.right = node.right
					node.right.parent = min
				}

				if parent.left == node {
					parent.left = min
				} else {
					parent.right = min
				}
				min.parent = parent

				minColor := min.color
				min.color = node.color

				// 条件是被替换的 min 颜色是黑色(因为删除红色, 原来的关系依然是平衡的) 开始调整
				if minColor == Black {
					if min != node.right {
						r.fixRemove(x, isParent)
					} else if min.right != nil {
						r.fixRemove(min.right, false)
					} else {
						r.fixRemove(min, true)
					}
				}
			} else {
				// 右孩子不存在, 使用左孩子接替当前的节点

				// 删除当前node节点, 使用左孩子代替
				if node.left != nil {
					node.left.parent = parent
				}
				if parent.left == node {
					parent.left = node.left
				} else {
					parent.right = node.right
				}

				// 开始调整(删除节点是黑色, 红色无需调整), 要么是替换者, 要么是其父亲
				if node.color == Black && r.root != nil {
					var (
						isParent = false
						x        = node.left
					)
					if x == nil {
						isParent = true
						x = node.parent
					}

					r.fixRemove(x, isParent)
				}
			}

			// 彻底移除被删除的节点 node
			node.parent = nil
			node.left = nil
			node.right = nil

			if r.root != nil {
				r.root.color = Black
				r.root.parent = nil
			}

			return node.val
		}
	}

	return 0
}

func (r *RBTree) getBrother(node *rbNode, parent *rbNode) *rbNode {
	if node != nil {
		parent = node.parent
	}
	if parent.left == node {
		return parent.right
	} else {
		return parent.left
	}
}

func (r *RBTree) fixRemove(node *rbNode, isParent bool) {
	var (
		cur, parent *rbNode
		isred       bool
	)
	if !isParent {
		cur = node
		isred = cur.color == Red
	} else {
		parent = node
	}

	// 遇到红色节点或者根节点, 则停止
	for !isred && cur != r.root {
		log.Println(cur, parent)
		brother := r.getBrother(cur, parent) // 兄弟
		isLeft := parent.left == brother     // 兄弟是否是左孩子

		if brother.color == Red && isLeft { // 状况1, 兄弟是红色
			r.leftRoate(parent)
			parent.color, brother.color = brother.color, parent.color
		} else if brother.color == Red && !isLeft {
			r.rightRoate(parent)
			parent.color, brother.color = brother.color, parent.color
		} else if r.isBlack(brother.left) && r.isBlack(brother.right) { // 状况2, 兄弟孩子全是黑色
			brother.color = Red
			cur = parent
			parent = cur.parent
			isred = cur.color == Red
		} else if isLeft && !r.isBlack(brother.left) && r.isBlack(brother.right) { // 状况3, 兄弟反方向的孩子是红色
			brother.color = Red
			if brother.left != nil {
				brother.left.color = Black
			}
			r.rightRoate(brother)
		} else if !isLeft && !r.isBlack(brother.right) && r.isBlack(brother.left) {
			brother.color = Red
			if brother.right != nil {
				brother.right.color = Black
			}
			r.leftRoate(brother)
		} else if isLeft && r.isRed(brother.right) { // 状况4, 兄弟同方向的还是是红色
			brother.color, parent.color = parent.color, Black
			if brother.right != nil {
				brother.right.color = Black
			}
			r.rightRoate(parent)
			cur = r.root
		} else if !isLeft && r.isRed(brother.left) {
			brother.color, parent.color = parent.color, Black
			if brother.left != nil {
				brother.left.color = Black
			}
			r.leftRoate(parent)
			cur = r.root
		}
	}

	if isred {
		cur.color = Black
	}

	if r.root != nil {
		r.root.color = Black
		r.root.parent = nil
	}
}

func (r *RBTree) removeMin(min *rbNode) *rbNode {
	parent := min
	for min != nil && min.left != nil {
		parent = min
		min = parent.left
	}

	// 最小节点就是当前传入的节点
	if parent == min {
		return min
	}

	// 删除最小节点, 注意: 此时的 min 的左孩子是 nil, parent的左孩子是min
	parent.left = min.right
	if min.right != nil {
		min.right.parent = parent
	}

	return min
}

func (r *RBTree) isBlack(node *rbNode) bool {
	return node == nil || node.color == Black
}
func (r *RBTree) isRed(node *rbNode) bool {
	return node != nil && node.color == Red
}
