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
假设环入口距离链表头的长度为L, 快慢指针相遇的位置是 cross, 且该位置距离环入口的长度是S. R是环的长度

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
	for slow != nil && quick != nil && quick.Next != nil {
		slow = slow.Next
		quick = quick.Next.Next
		if slow == quick {
			exist = true
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

/*
两个链表的相交位置
*/
// 方式一: 先计算长度, 然后根据长度差来同时走 O(N)
func getIntersectionNodeI(a, b *ListNode) *ListNode {
	if a == nil || b == nil {
		return nil
	}

	lenA, lenB := 0, 0
	curA, curB := a, b
	for curA != nil || curB != nil {
		if curA != nil {
			lenA++
			curA = curA.Next
		}

		if curB != nil {
			lenB++
			curB = curB.Next
		}
	}

	diff := 0
	if lenA >= lenB {
		curA = a
		curB = b
		diff = lenA - lenB
	} else {
		curA = b
		curB = a
		diff = lenB - lenA
	}

	for curA != nil && curB != nil {
		if diff > 0 {
			curA = curA.Next // A 先走 diff 步
			diff -= 1
		}

		if curA == curB {
			return curA
		}

		curA = curA.Next
		curB = curB.Next
	}

	return nil
}

// 方式二: 同时走, A->B, B->A, 相遇的地方必定是首个公共点
func getIntersectionNodeII(a, b *ListNode) *ListNode {
	if a == nil || b == nil {
		return nil
	}

	curA, curB := a, b
	for curA != nil || curB != nil {
		if curA == curB {
			return curA
		}

		if curA == nil || curB == nil {
			if curA == nil {
				curA = b
			}
			if curB == nil {
				curB = a
			}
			continue
		}

		curA = curA.Next
		curB = curB.Next
	}

	return nil
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

/*
92. 反转链表 II

反转从位置 m 到 n 的链表. 请使用一趟扫描完成反转.

记录以下的内容:
	反转链表前缀(prev)节点, 反转链表后缀(next)节点.
    反转链表头节点(start), 反转链表尾节点(end)

链接操作:
	prev.Next = start
	end.Next = next
*/
func reverseBetween(head *ListNode, m int, n int) *ListNode {
	if head == nil || head.Next == nil || m == n {
		return head
	}

	idx := 0
	cur := head
	var start, end *ListNode // 链表的头,尾
	var prev, next *ListNode

	for cur != nil {
		idx += 1

		// 在 m-1 位置捕捉 prev, end, 在n+1 位置捕捉 next
		if idx == m-1 {
			prev = cur
			end = cur.Next
		}
		if idx == n+1 {
			next = cur
			break
		}

		if idx >= m && idx <= n {
			next := cur.Next

			cur.Next = start
			start = cur

			cur = next
			continue
		}

		cur = cur.Next
	}

	// 特殊情况: 从头节点开始反转, 此时链表只有两部分.
	if prev == nil {
		head.Next = next
		return start
	}

	// 正常的逻辑, 三部分处理
	prev.Next = start
	end.Next = next

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

思路: 首先统计 head 长度, 计算结束的 index

翻转过程(每 k 一个一翻转, idx % k == 0), 翻转过程记录变量:

pre, end 当前翻转的链表的头和尾.
pend 上一次翻转的尾巴.

初始化: end 为 head, pend, pre 为nil

第一次迭代完成:
	root=pre // root是最终的返回结果

	pend=end
	end=next
	pre=nil

非第一次迭代完成:
	pend.Next=pre // 链接

	pend=end
	end=next
	pre=nil
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

	cur, end = head, head
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

//==================================================================================================

// 两个链表, 逆序存储数字, 求和
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	res := new(ListNode)
	n1 := l1
	n2 := l2

	var node = res
	var carry int
	for n1 != nil && n2 != nil {
		sum := n1.Val + n2.Val + carry
		carry = sum / 10
		node.Val = sum % 10

		n1 = n1.Next
		n2 = n2.Next

		// 存在n1或者n2
		if n1 != nil || n2 != nil {
			node.Next = new(ListNode)
			node = node.Next
		}
	}

	for n1 != nil {
		sum := n1.Val + carry
		carry = sum / 10
		node.Val = sum % 10

		n1 = n1.Next
		if n1 != nil {
			node.Next = new(ListNode)
			node = node.Next
		}
	}

	for n2 != nil {
		sum := n2.Val + carry
		carry = sum / 10
		node.Val = sum % 10

		n2 = n2.Next
		if n2 != nil {
			node.Next = new(ListNode)
			node = node.Next
		}
	}

	if carry > 0 {
		node.Next = &ListNode{
			Val:  carry,
			Next: nil,
		}
	}

	return res
}

// 两个链表, 顺序存储数字, 求和 ==> (栈的先进后出, 变为逆序存储)
func addTwoNumbersMore(l1 *ListNode, l2 *ListNode) *ListNode {
	var s1, s2 []int
	n1 := l1
	for n1 != nil {
		s1 = append(s1, n1.Val)
		n1 = n1.Next
	}
	n2 := l2
	for n2 != nil {
		s2 = append(s2, n1.Val)
		n2 = n2.Next
	}
	var node = new(ListNode)
	var carry int

	for len(s1) > 0 && len(s2) > 0 {
		e1 := s1[len(s1)-1]
		s1 = s1[:len(s1)-1]

		e2 := s2[len(s2)-1]
		s2 = s2[:len(s2)-1]

		sum := e1 + e2 + carry
		carry = sum / 10
		node.Val = sum % 10

		if len(s1) > 0 || len(s2) > 0 {
			node = &ListNode{
				Next: node,
			}
		}
	}

	for len(s1) > 0 {
		e1 := s1[len(s1)-1]
		s1 = s1[:len(s1)-1]

		sum := e1 + carry
		carry = sum / 10
		node.Val = sum % 10

		if len(s1) > 0 {
			node = &ListNode{
				Next: node,
			}
		}
	}

	for len(s2) > 0 {
		e2 := s2[len(s2)-1]
		s2 = s2[:len(s2)-1]

		sum := e2 + carry
		carry = sum / 10
		node.Val = sum % 10

		if len(s2) > 0 {
			node = &ListNode{
				Next: node,
			}
		}
	}

	if carry > 0 {
		node = &ListNode{
			Val:  carry,
			Next: node,
		}
	}

	return node
}
