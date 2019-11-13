package leetcode

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
