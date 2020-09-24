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
	RED   = Color("RED")
	BLACK = Color("BLACK")
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

2.2 叔叔不存在或者叔叔是黑色, 祖孙三代处于同一条斜线上.
a) 父节点变黑色, 祖父节点变红
b) 根据"父节点"向相反的方向旋转(斜线)

2.2 叔叔不存在或者叔叔是黑色, 祖孙三代没有在统一斜线上.
a) 根据"当前节点"向相反的方向旋转(父子斜线), 情况转变为状况2.2
注意: 此种状况下,没有改变颜色
*/

type RBTree struct {
	root *rbNode
}

func (r *RBTree) insert(val int) bool {
	node := &rbNode{val: val}

	if r.root == nil {
		r.root = node
		node.color = BLACK
		return false
	}

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

	r.fixAfterInsertion(node)
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

func (r *RBTree) Println() string {
	if r.root == nil {
		return ""
	}

	var buf bytes.Buffer
	queue := []*rbNode{r.root}
	for len(queue) != 0 {
		var temp []*rbNode

		for len(queue) != 0 {
			cur := queue[0]
			queue = queue[1:]

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
			if cur.color == BLACK {
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
		buf.WriteString("\n")
	}

	return buf.String()
}

func colorOf(node *rbNode) Color {
	if node == nil {
		return BLACK
	}
	return node.color
}

func leftOf(node *rbNode) *rbNode {
	if node == nil {
		return nil
	}
	return node.left
}

func rightOf(node *rbNode) *rbNode {
	if node == nil {
		return nil
	}
	return node.right
}

func parentOf(node *rbNode) *rbNode {
	if node == nil {
		return nil
	}
	return node.parent
}

func setColor(node *rbNode, color Color) {
	if node != nil {
		node.color = color
	}
}

func (r *RBTree) find(val int) *rbNode {
	node := r.root

	for node != nil {
		if node.val > val {
			node = node.left
		} else if node.val < val {
			node = node.right
		} else {
			return node
		}
	}

	return nil
}

// 左旋, p是父节点, 旋转点是 p - p.right
func (r *RBTree) rotateLeft(p *rbNode) {
	if p != nil {
		right := p.right

		p.right = right.left
		if right.left != nil {
			right.left.parent = p
		}

		right.parent = p.parent
		if p.parent == nil {
			r.root = right
		} else if p.parent.left == p {
			p.parent.left = right
		} else {
			p.parent.right = right
		}

		right.left = p
		p.parent = right
	}
}

// 右旋, p是父节点, 旋转点是 p - p.left
func (r *RBTree) rotateRight(p *rbNode) {
	if p != nil {
		left := p.left

		p.left = left.right
		if left.right != nil {
			left.right.parent = p
		}

		left.parent = p.parent
		if p.parent == nil {
			r.root = left
		} else if p.parent.right == p {
			p.parent.right = left
		} else {
			p.parent.left = left
		}

		left.right = p
		p.parent = left
	}
}

func (r *RBTree) remove(val int) int {
	node := r.find(val)
	if node == nil {
		return 0
	}

	// 在左右不为空的状况下, 寻找 node 的后继者
	if node.left != nil && node.right != nil {
		node = r.successor(node)
	}

	var replace *rbNode
	if node.left != nil {
		replace = node.left
	} else {
		replace = node.right
	}

	// 分为三种情况讨论:
	// 1. node 的左孩子或者右孩子不为空
	// 2. node 的父节点为空(根节点)
	// 3. node 的父节点不为空且左右孩子都为空
	if replace != nil {
		// 删除 node 节点
		replace.parent = node.parent
		if node.parent == nil {
			r.root = replace
		} else if node == node.parent.left {
			node.parent.left = replace
		} else {
			node.parent.right = replace
		}

		node.left = nil
		node.right = nil
		node.parent = nil

		// 从替换 node 的节点开始修复
		if node.color == BLACK {
			r.fixAfterDeletion(replace)
		}
	} else if node.parent == nil {
		r.root = nil
	} else {
		// 从 node 节点开始修复
		if node.color == BLACK {
			r.fixAfterDeletion(node)
		}

		// 删除 node 节点
		if node.parent != nil {
			if node == node.parent.left {
				node.parent.left = nil
			} else if node == node.parent.right {
				node.parent.right = nil
			}

			node.parent = nil
		}
	}

	return node.val
}

func (r *RBTree) fixAfterInsertion(x *rbNode) {
	x.color = RED

	for x != nil && x != r.root && x.parent.color == RED {
		if parentOf(x) == leftOf(parentOf(parentOf(x))) {
			y := rightOf(parentOf(parentOf(x)))
			if colorOf(y) == RED {
				setColor(parentOf(x), BLACK)
				setColor(y, BLACK)
				setColor(parentOf(parentOf(x)), RED)
				x = parentOf(parentOf(x))
			} else {
				if x == rightOf(parentOf(x)) {
					x = parentOf(x)
					r.rotateLeft(x)
				}

				setColor(parentOf(x), BLACK)
				setColor(parentOf(parentOf(x)), RED)
				r.rotateRight(parentOf(parentOf(x)))
			}
		} else {
			y := leftOf(parentOf(parentOf(x)))
			if colorOf(y) == RED {
				setColor(parentOf(x), BLACK)
				setColor(y, BLACK)
				setColor(parentOf(parentOf(x)), RED)
				x = parentOf(parentOf(x))
			} else {
				if x == leftOf(parentOf(x)) {
					x = parentOf(x)
					r.rotateRight(x)
				}

				setColor(parentOf(x), BLACK)
				setColor(parentOf(parentOf(x)), RED)
				r.rotateLeft(parentOf(parentOf(x)))
			}
		}
	}

	r.root.color = BLACK
}

/*
[doc](https://tech.meituan.com/2016/12/02/redblack-tree.html)

删除:

首先查找要删除节点的后面的节点(中序遍历)


删除修复:

删除修复操作在遇到被修复的节点是 "红色节点" 或 "root节点", 修复完毕(核心).

删除修复操作是针对 "非根黑色" 节点才有的. 处理思想是从兄弟节点上借调黑色节点过来, 如果兄弟节点没有黑色节点可以借调, 只能
往上追溯, 将每一级的黑节点数减去一个, 使得整棵树符合定义.

1. 待删除的节点的兄弟节点是红色.
思路: 兄弟节点是红色, 无法借调黑色节点, 因此需要将兄弟提升为父节点, 由于兄弟是红色的, 根据定义, 兄弟节点的子节点是黑色的
就可以从它的子节点借调了.

调整: 向 "父亲-兄弟" 向相反的方向进行旋转. 父节点变为红色, 兄弟节点变为黑色

情况就会转换为 2, 3, 4 情况了.

2. 待删除的节点的兄弟节点是黑色
2.1 兄弟节点中的子节点都是黑色.

思路: 兄弟可以消除一个黑色节点, 因为兄弟子节点都是黑色的, 所以可以将兄弟节点变红, 从而保证局部树的局部颜色符合定义. 这时
候将父节点变成新的节点, 继续向上调整.

调整: 兄弟变成红色, 使用父节点向上调整.


3. 待删除的节点的兄弟节点是黑色(相反)
3.1 兄弟在右边, 兄弟的左子节点为红色
3.2 兄弟在左边, 兄弟的右子节点为红色

思路: 将红色节点借调过来, 从而达到状态的转换. 使得转换为状态 4.

调整: 向 "兄弟-兄弟红色孩子" 相反的方向进行进行调整. 兄弟红色子节点与兄弟节点颜色交换.


4. 待删除的节点的兄弟节点是黑色(相同)
4.1 兄弟在右边, 兄弟的右子节点为红色
4.2 兄弟在左边, 兄弟的左子节点为红色

思路: 真正的节点借调操作, 将"兄弟的黑色孩子"以及"父亲节点"借调过来. 从而实现平衡(调整完毕)

调整: 向 "父亲-兄弟" 相反的方向进行调整. 兄弟的颜色就是父亲的颜色, 父亲, 兄弟的红色孩子节点变为黑色
*/
func (r *RBTree) fixAfterDeletion(x *rbNode) {
	for x != r.root && colorOf(x) == BLACK {
		// x 是左孩子
		if x == leftOf(parentOf(x)) {
			// 兄弟
			sib := rightOf(parentOf(x))

			// 兄弟是颜色是红色
			if colorOf(sib) == RED {
				setColor(sib, BLACK)
				setColor(parentOf(x), RED)
				r.rotateLeft(parentOf(x))
				sib = rightOf(parentOf(x))
			}

			// 兄弟的孩子都是黑色(不存在或者确实是黑色)
			if colorOf(leftOf(sib)) == BLACK && colorOf(rightOf(sib)) == BLACK {
				setColor(sib, RED)
				x = parentOf(x)
			} else {
				// 兄弟的左孩子是红色
				if colorOf(leftOf(sib)) == RED {
					setColor(leftOf(sib), BLACK)
					setColor(sib, RED)
					r.rotateRight(sib)
					sib = rightOf(parentOf(x)) // 重新获取兄弟节点(兄弟节点的左孩子)
				}

				// 兄弟的右孩子是红色
				setColor(sib, colorOf(parentOf(x)))
				setColor(parentOf(x), BLACK)
				setColor(rightOf(sib), BLACK)
				r.rotateLeft(parentOf(x))
				x = r.root
			}
		} else {
			sib := leftOf(parentOf(x))
			if colorOf(sib) == RED {
				setColor(sib, BLACK)
				setColor(parentOf(x), RED)
				r.rotateRight(parentOf(x))
				sib = leftOf(parentOf(x))
			}

			if colorOf(rightOf(sib)) == BLACK && colorOf(leftOf(sib)) == BLACK {
				setColor(sib, RED)
				x = parentOf(x)
			} else {
				if colorOf(rightOf(sib)) == RED {
					setColor(rightOf(sib), BLACK)
					setColor(sib, RED)
					r.rotateLeft(sib)
					sib = leftOf(parentOf(x))
				}

				setColor(sib, colorOf(parentOf(x)))
				setColor(parentOf(x), BLACK)
				setColor(leftOf(sib), BLACK)
				r.rotateRight(parentOf(x))
				x = r.root
			}
		}
	}

	setColor(x, BLACK)
}

// 寻找 node 的后继节点(顺序上);
// 如果 node 存在右子树, 向下查找"右子树的最左的孩子";
// 否则, 向上查找到 "右子树上的第一个左子节点";
func (r *RBTree) successor(node *rbNode) *rbNode {
	if node == nil {
		return node
	} else if node.right != nil {
		p := node.right
		for p.left != nil {
			p = p.left
		}
		return p
	} else {
		// 向上查找 "右子树上的第一个左子节点"
		p := node.parent
		ch := node
		for p != nil && ch == p.right {
			ch = p
			p = p.parent
		}
		log.Println("右子树上的第一个左子节点")
		return p
	}
}

func (r *RBTree) Valid() bool {
	root := r.root
	if root == nil {
		return true
	}

	if root.color == RED {
		return false
	}

	ans := true
	dfs(root, &ans)
	return ans
}

func dfs(node *rbNode, ans *bool) int {
	if node == nil {
		return 0
	}

	if node.left == nil && node.right == nil {
		return ifelse(node.color == BLACK, 1, 0)
	}

	if node.left != nil && node.color == RED {
		if node.color == node.left.color {
			log.Println("color left", node)
			*ans = false
		}
	}
	if node.right != nil && node.color == RED {
		if node.color == node.right.color {
			log.Println("color right", node)
			*ans = false
		}
	}

	L := dfs(node.left, ans)
	R := dfs(node.right, ans)

	if L != R {
		log.Println("len", node, L, R)
		*ans = false
	}

	return L + ifelse(node.color == BLACK, 1, 0)
}

func ifelse(conditio bool, ifresult, elseresult int) int {
	if conditio {
		return ifresult
	} else {
		return elseresult
	}
}
