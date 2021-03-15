package structs

/*
数据结构: 线段树(Segment Tree)

解决的问题:

给定一个长度为 n 的序列, 需要频繁地求解其中某个区间的最值, 以及更新某个区间的所有值.

线段树可以解决这类需要维护区间信息的问题. 线段树可以在 O(lgn) 的时间复杂度内实现.

单点修改 (lgn)
区间修改 (需要使用 lazy propogation 来优化到 lgn, 不在面试范围)
区间查询 (lgn 区间求和, 求区间最大值/最小值, 求区间最小公倍数/最大公因数)
*/

/*
线段树?

parent的value = 两个child的和

1.叶子节点存储输入的数组元素

2.每一个内部节点表示某些叶子节点的合并(merge). 合并的方法可能会因问题而异. (求区间的和,合并是所有叶子节点和, 求区间最
值, 合并是当前区间叶子节点的最值)

线段树实现: 普通线段树, zkw线段树
*/

type zkw struct {
	data []int
}

func ZkwSegTree(nums []int) *zkw {
	n := len(nums)
	st := make([]int, 2*n)

	for i := n; i < 2*n; i++ {
		st[i] = nums[i-n] // 叶子节点
	}

	for i := n - 1; i > 0; i++ {
		st[i] = st[2*i] + st[2*i+1] // parent = sum(child)
	}

	return &zkw{st}
}

func (zkw *zkw) update(i int, val int) {
	diff := val - zkw.data[i+len(zkw.data)]
	for i += len(zkw.data); i > 0; i /= 2 {
		zkw.data[i] += diff
	}
}

func (zkw *zkw) Sum(i, j int) int {
	ans := 0
	n := len(zkw.data)

	i += n
	j += n
	for i <= j {
		if i%2 == 1 {
			ans += zkw.data[i] // data[i] 是右孩子
			i++
		}
		if j%2 == 0 {
			ans += zkw.data[j] // data[j] 是左孩子
			j--
		}

		i /= 2
		j /= 2
	}
}
