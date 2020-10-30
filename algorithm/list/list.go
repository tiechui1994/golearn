package list

/**

存在环理论证明:

假设环的长度是 R, 当慢指针slow走到环入口时, 快指针quick处于环中的某个位置, 且两者之间的距离是S

在慢指针进入环后的t时间内, 快指针从距离环入口S处走了2t个节点, 相当于从环入口走了S+2t个节点.

假设快慢指针移动可以相遇:
	S+2t - t = nR  => S+t = nR  如果对于任意的S, R, n 总可以找到一个合适的t满足公式, 那么
说明快慢指针一定可以相遇.

	实际上 S < R, 所以在慢指针走过一圈之前就可以相遇.



环入口位置:
假设环入口距离链表头的长度为L, 快慢指针相遇的位置是cross, 且该位置距离环入口的长度是S.

慢指针: L+S
快指针: L+S+nR

2(L+S) = L+S+nR => L+S=nR

当R=1, 快指针在相遇之前多走了1圈, 即L+S=R, L=R-S. L表示链表头距离环入口的距离, R-S表示从cross
继续移动到达环入口的距离, 二者是相等的, 可以采用两个指针, 一个从表头出发, 一个从cross出发

当R>1, L=nR-S L表示链表头距离环入口的距离. nR-S可以看成从cross出发移动nR步后再倒退S步. 从cross
移动nR步后回到cross位置, 倒退S步后是环入口.
**/

type Node struct {
	Next *Node
	Val  int
}

func findCycleNode(root *Node) *Node {
	slow := root
	quick := root
	// 相遇
	var exist bool
	for slow != nil && quick.Next != nil {
		slow = slow.Next
		quick = quick.Next.Next
		if slow == quick {
			exist = true
			break
		}
		if slow == nil || quick == nil {
			break
		}
	}
	// 非break
	if !exist {
		return nil
	}

	// 走L步
	slow = root
	for slow != quick {
		slow = slow.Next
		quick = quick.Next
	}

	return slow
}

// 两个链表相交的节点
type liststack []*Node

func (l *liststack) push(node *Node) {
	*l = append(*l, node)
}
func (l *liststack) pop() *Node {
	node := (*l)[l.len()-1]
	*l = (*l)[0 : l.len()-1]
	return node
}
func (l *liststack) len() int {
	return len(*l)
}

// 使用栈方法
func findTwoListIntersectionNodeI(a, b *Node) *Node {
	if a == nil || b == nil {
		return nil
	}
	var stackA, stackB liststack
	var rA, rB = a, b

	for rA != nil || rB != nil {
		if rA != nil {
			stackA.push(rA)
			rA = rA.Next
		}

		if rB != nil {
			stackB.push(rA)
			rB = rB.Next
		}
	}

	nA := stackA.pop()
	nB := stackB.pop()
	if nA != nB {
		return nil
	}

	var res = nA
	for stackA.len() != 0 && stackA.len() != 0 {
		nA = stackA.pop()
		nB = stackB.pop()
		if nA != nB {
			break
		}
		res = nA
	}

	return res
}

// 不使用栈的方法, O(N)
func findTwoListIntersectionNodeII(a, b *Node) *Node {
	if a == nil || b == nil {
		return nil
	}

	var aLast, bLast *Node
	var aLen, bLen int
	var rA, rB = a, b

	for rA != nil || rB != nil {
		if rA != nil {
			aLen++
			if rA.Next == nil {
				aLast = rA
			}
			rA = rA.Next
		}

		if rB != nil {
			bLen++
			if rB.Next == nil {
				bLast = rB
			}
			rB = rB.Next
		}
	}

	if aLast != bLast {
		return nil
	}

	var long, short *Node
	var diff int

	if aLen > bLen {
		long, short = a, b
		diff = aLen - bLen
	} else {
		long, short = b, a
		diff = bLen - aLen
	}

	for diff > 0 {
		long = long.Next
	}

	for long != short {
		long = long.Next
		short = short.Next
	}

	return long
}

// 删除倒数第N个Node
// 双指针, 第一次, first 从头遍历 n 次
//        第二次, second 从头开始, first 遍历到结尾. 则second也到达删除的位置
func removeNthFromEnd(head *ListNode, n int) *ListNode {
	if n <= 0 {
		return head
	}

	first := head
	for n > 0 {
		first = first.Next
		n--
	}

	if first == nil {
		return head.Next
	}

	second := head
	for first.Next != nil {
		first = first.Next
		second = second.Next
	}

	second.Next = second.Next.Next

	return head
}

func reverseBetween(head *ListNode, m int, n int) *ListNode {
	var i = 0
	var node = head

	var prev, start *ListNode
	for node != nil {
		i++
		next := node.Next
		if i >= m && i <= n {
			if i == m {
				start = node
				prev = node
			} else {
				node.Next = prev
				prev = node
			}
		}
		if i == n {
			start.Next = prev
			break
		}
		node = next
	}

	return head
}
