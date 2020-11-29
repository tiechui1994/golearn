package list

import (
	"fmt"
)

type ListNode struct {
	Next *ListNode
	Val  int
}

func (l *ListNode) String() string {
	res := "["
	cur := l
	for cur != nil {
		res += fmt.Sprintf("%v,", cur.Val)
		cur = cur.Next
	}

	if res == "[" {
		return res + "]"
	}

	return res[:len(res)-1] + "]"
}

//==================================================================================================

/**
存在环理论证明:

假设环的长度是 R, 当慢指针slow走到环入口时, 快指针quick处于环中的某个位置, 且两者之间的距离是S

在慢指针进入环后的t时间内, 快指针从距离环入口S处走了2t个节点, 相当于从环入口走了S+2t个节点.

假设快慢指针移动可以相遇:
	S+2t - t = nR  => S+t = nR  如果对于任意的S, R, n 总可以找到一个合适的t满足公式, 那么
说明快慢指针一定可以相遇.

	实际上 S < R, 所以在慢指针走过一圈之前就可以相遇.



环入口位置:
假设环入口距离链表头的长度为L, 快慢指针相遇的位置是cross, 且该位置距离环入口的长度是S. R是环的长度

慢指针: L+S
快指针: L+S+nR

2(L+S) = L+S+nR => L+S=nR

当R=1, 快指针在相遇之前多走了1圈, 即L+S=R, L=R-S. L表示链表头距离环入口的距离, R-S表示从cross
继续移动到达环入口的距离, 二者是相等的, 可以采用两个指针, 一个从表头出发, 一个从cross出发

当R>1, L=nR-S L表示链表头距离环入口的距离. nR-S可以看成从cross出发移动nR步后再倒退S步. 从cross
移动nR步后回到cross位置, 倒退S步后是环入口.
**/

func findCycleNode(root *ListNode) *ListNode {
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

//==================================================================================================

/*
141. 环形链表(判断是否存在环)
*/

// 方式一: 快慢指针(需要环的理论支持)
func hasCycleI(head *ListNode) bool {
	if head == nil || head.Next == nil {
		return false
	}

	slow := head
	quick := head

	for slow != nil && quick != nil && quick.Next != nil {
		slow = slow.Next
		quick = quick.Next.Next

		if slow == quick {
			return true
		}
	}

	return false
}

// 方式二: 删除头部指针(巧妙删除).
// 删除的方法: cur.Next=cur 断开与后面的联系, 如果存在环, 必然会指向到已经删除的节点, 这个时候 cur.Next == cur
func hasCycleII(head *ListNode) bool {
	if head == nil || head.Next == nil {
		return false
	}

	if head.Next == head {
		return true
	}

	next := head.Next
	head.Next = head

	return hasCycleII(next)
}

//==================================================================================================

// 两个链表相交的节点
type liststack []*ListNode

func (l *liststack) push(node *ListNode) {
	*l = append(*l, node)
}
func (l *liststack) pop() *ListNode {
	node := (*l)[l.len()-1]
	*l = (*l)[0: l.len()-1]
	return node
}
func (l *liststack) len() int {
	return len(*l)
}

// 使用栈方法
func findTwoListIntersectionNodeI(a, b *ListNode) *ListNode {
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
func findTwoListIntersectionNodeII(a, b *ListNode) *ListNode {
	if a == nil || b == nil {
		return nil
	}

	var aLast, bLast *ListNode
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

	var long, short *ListNode
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

//==================================================================================================

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

//==================================================================================================

/*
234. 回文链表(进阶方法), O(N)+O(1)
*/

/*
快慢指针:

所谓的快慢指针, 开始的位置都是 root, 只不过 slow 指针每次迭代一次, quick 指针每次迭代两次.
循环条件: slow != nil && quck != nil && quck.Next != nil (也是可以改进的)

最终出现的结果是: slow指针的索引位置(中间位置后面的第一个): N/2 向上取整.

反转链表(head):

pre = nil
cur = head

loop:
	next := cur.Next

	// 反转
	cur.Next = pre
	pre = cur

	cur = next

最终的结果是 pre
*/

// 方式一: 快慢指针找到中点, 然后翻转重点指针后面的部分
// 注意事项: 一定要将 slow 指针的前面那个指针断开
func isPalindromeI(head *ListNode) bool {
	if head == nil || head.Next == nil {
		return true
	}

	if head.Next.Next == nil {
		return head.Val == head.Next.Val
	}

	reverse := func(node *ListNode) *ListNode {
		var pre *ListNode
		cur := node
		for cur != nil {
			next := cur.Next

			// 翻转
			cur.Next = pre
			pre = cur

			cur = next
		}

		return pre
	}

	pre := head
	slow := head
	quick := head
	for slow != nil && quick.Next != nil {
		pre = slow
		slow = slow.Next
		quick = quick.Next.Next
	}
	pre.Next = nil // 断开链接, 重点

	c1 := reverse(slow)
	c2 := head
	for c1 != nil && c2 != nil {
		if c1.Val != c2.Val {
			return false
		}

		c1 = c1.Next
		c2 = c2.Next
	}

	return true
}

// 方式二: hash 方法(需要理论支持), 这个思路比较独特, 数学要好才可以.
//
// hash = hash * seed + val, seed 是一个质数, val 是节点的值, 哈希结果为:
// hash1 = a[1]*seed^(n-1) + a[2]*seed^(n-2) + ... + a[n-1]*seed^(1) + a[n]*seed^(0)
//
// 逆序的哈希结果:
// hash2 = a[1]*seed^(0) + a[2]*seed^(1) + ... + a[n-1]*seed^(n-2) + a[n]*seed^(n-1)
//
// 如果 hash1 == hash2, 可以保证是回文的
func isPalindromeII(head *ListNode) bool {
	if head == nil || head.Next == nil {
		return true
	}

	var (
		hash1, hash2 int64
		seed         int64 = 0x1000193
		h            int64 = 1
	)

	cur := head
	for cur != nil {
		hash1 = hash1*seed + int64(cur.Val)
		hash2 = hash2 + int64(cur.Val)*h

		h *= seed
		cur = cur.Next
	}

	return hash2 == hash1
}

//==================================================================================================
/*
25. K个一组翻转链表(主要是逻辑比较复杂而已)
*/
func reverseKGroup(head *ListNode, k int) *ListNode {
	count := 0
	cur := head
	for cur != nil {
		count += 1
		cur = cur.Next
	}

	var root *ListNode
	var pre, end, pend *ListNode

	idx := 0
	stop := count - count%k

	cur = head
	pend, end = head, head
	for cur != nil {
		next := cur.Next

		if idx == stop {
			pend.Next = cur
			break
		}

		cur.Next = pre
		pre = cur

		cur = next
		idx += 1
		if idx%k == 0 {
			if root == nil {
				root = pre
				pend = end
				pre = nil
				end = next
				continue
			}

			pend.Next = pre
			pend = end
			pre = nil
			end = next
		}
	}

	return root
}
