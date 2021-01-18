package category

import (
	"sort"
	"container/heap"
)

/**
滑动窗口:

**/

/*
53: 给定一个整数数组 nums, 找到一个具有最大和的连续子数组(子数组最少包含一个元素), 返回其最大和.

1) 动态规划: f(i) 表示以第 i 个数结尾的 "连续子数组的最大和", 那么 result = max{ f(i) } 0<=i<=n-1
因此, 只需要求出每个位置的 f(i), 然后返回 f 数组中的最大值即可. 状态转移方程:

	f(i) = max{ f(i-1)+a[i], ai }

2) 分治: 定义一个操作 get(a, l, r) 表示查询 a 序列 [l,r] 区间内的最大子段和, 那么最终的答案是 get(nums, 0, len-1)
对于一个区间 [l, r], 取 m = (l+r)/2, 对于区间 [l, m] 和 [m+1,r] 分治求解. 当递归逐层深入直到区间长度缩小为1时,
递归开始回升. 下来考虑如何通过 [l,m] 和 [m+1, r] 合并成区间 [l, r] 的信息.

关键问题: 维护区间的哪些信息? 如何合并这些信息?

对于区间 [l,r], 维护四个变量:
- lSum 表示 [l, r] 内以 l 为左端点的最大子段和
- rSum 表示 [l, r] 内以 r 为右端点的最大子段和
- mSum 表示 [l, r] 内的最大子段和
- iSum 表示 [l, r] 的区间和

区间合并: 对于长度大于1的区间:

- [l, r] 的 iSum, "左子区间" 的 iSum 与 "右子区间" 的 iSum 的和

- [l, r] 的 lSum, 存在两种可能, 要么等于 "左子区间" 的 lSum, 要么等于 "左子区间" 的 iSum 加上 "右子区间" 的 lSum,
二者取大.

- [l, r] 的 rSum, 存在两种可能, 要么等于 "右子区间" 的 rSum, 要么等于 "右子区间" 的 iSum 加上 "左子区间" 的 rSum,
二者取大.

- [l, r] 的 mSum, 考虑 [l, r] 的 mSum 是否跨越 m, 如果没有跨越, mSum 是 "左子区间" mSum 和 "右子区间" mSum 的
最大值. 跨越 m, 可能是 "左子区间" rSum 和 "右子区间" lSum 的和, 三者取大

*/

func maxSubArrayI(nums []int) int {
	f := nums[0]
	max := f
	for i := 1; i < len(nums); i++ {
		if f+nums[i] > nums[i] {
			f = f + nums[i]
		} else {
			f = nums[i]
		}

		if max > f {
			max = f
		}
	}

	return max
}

func maxSubArrayII(nums []int) int {
	var get func(nums []int, l, r int) (iSum, lSum, rSum, mSum int)
	get = func(nums []int, l, r int) (iSum, lSum, rSum, mSum int) {
		if l == r {
			return nums[l], nums[l], nums[l], nums[l]
		}

		m := (l + r) / 2
		LiSum, LlSum, LrSum, LmSum := get(nums, l, m)
		RiSum, RlSum, RrSum, RmSum := get(nums, m+1, r)

		iSum = LiSum + RiSum
		lSum = max(LiSum+RlSum, LlSum)
		rSum = max(RiSum+LrSum, RrSum)
		mSum = max(LmSum, RmSum, LrSum+RlSum)

		return iSum, lSum, rSum, mSum
	}

	_, _, _, result := get(nums, 0, len(nums)-1)
	return result
}

var a []int

type hp struct{ sort.IntSlice }

func (h hp) Less(i, j int) bool  { return a[h.IntSlice[i]] > a[h.IntSlice[j]] }
func (h *hp) Push(v interface{}) { h.IntSlice = append(h.IntSlice, v.(int)) }
func (h *hp) Pop() interface{}   { a := h.IntSlice; v := a[len(a)-1]; h.IntSlice = a[:len(a)-1]; return v }

func maxSlidingWindow(nums []int, k int) []int {
	a = nums
	q := &hp{make([]int, k)}
	for i := 0; i < k; i++ {
		q.IntSlice[i] = i
	}
	heap.Init(q)

	n := len(nums)
	ans := make([]int, 1, n-k+1)
	ans[0] = nums[q.IntSlice[0]]
	for i := k; i < n; i++ {
		heap.Push(q, i)
		for q.IntSlice[0] <= i-k {
			heap.Pop(q)
		}
		ans = append(ans, nums[q.IntSlice[0]])
	}
	return ans
}
