package category

import (
	"container/heap"
	"sort"
)

/**
滑动窗口:

解决数组/字符串的子元素问题, 它可以将嵌套的循环问题转换为单循环问题, 降低时间复杂度.

Easy: size fixed
eg: max sum of size = k

Median: size 可变, 单限制条件
eg: 找到 subarray sum 比目标值大一点点

Median: size 可变, 双限制条件
eg: longest substring with K distinct char (159. 至多包含两个不同字符的最长子串)

Hard: size fixed, 单限制条件
eg: sliding window maxinum, 考察单调队列
**/

/*
在字符串当中, 出现 K 个 distinct 字符的 SubString 的最长长度. [Mid]

要求: SubString 包含了 K 个字母, 然后求解最大长度.
*/
func LongSubstringKDistinct(s string, k int) int {
	m := make(map[byte]int)

	left := 0
	res := 0

	for i := 0; i < len(s); i++ {
		cur := s[i]
		m[cur] = m[cur] + 1 // 进: 当前遍历的i进入窗口

		for len(m) > k { // 出: 当前窗口不符合条件时, left 持续退出窗口
			c := s[left]
			m[c] = m[c] - 1
			if m[c] == 0 {
				delete(m, c)
			}

			left++
		}

		if res < i-left+1 { // 算: 现在的窗口 valid, 计算结果
			res = i - left + i
		}
	}

	return res
}

/*
76: 最小覆盖子串 [Hard]

给你一个字符串 s, 一个字符串 t. 返回 s 中涵盖 t 所有字符的最小子串.

如果 s 中不存在涵盖 t 所有字符的子串, 则返回空字符串 "".

类型: 长度固定, 单个条件限制.

map[char,int] (表示当前窗口, 字符的个数统计) + validcount(当前窗口合法字符串长度)

进, 出必须得"保持统一".
*/

func MinWindow(s, t string) string {
	var minLeft, minLen int

	chars := make(map[byte]int)
	for i := 0; i < len(t); i++ {
		chars[t[i]] += 1
	}

	minLen = len(s) + 1
	left := 0
	validCount := 0
	for i := 0; i < len(s); i++ {
		cur := s[i]
		if val, ok := chars[cur]; ok {
			// valid++, chars--
			if val > 0 {
				validCount++
			}
			chars[cur] -= 1
		}

		// 当 validCount 和 t 的长度相等时, 说明此时是满足要求的, 要开始移动 right
		for validCount == len(t) {
			if minLen > i-left+1 {
				minLeft = left
				minLen = i - left + 1
			}

			ch := s[left]
			if _, ok := chars[ch]; ok {
				// valid--, chars++
				chars[ch] += 1
				if chars[ch] > 0 {
					validCount--
				}
			}

			left++
		}
	}

	if minLen == len(s)+1 {
		minLen = 0
	}

	return s[minLeft : minLeft+minLen]
}

/*
395. 至少有 K 个重复字符的最长子串 [Hard]

给你一个字符串 s 和一个整数 k, 找出 s 中的最长子串, 要求该子串中的每一字符出现次数都不少于 k. 返回这一子串的长度.

注: 所给的字符串全都是小写英文字母.

要求: SubString 的每个字母出现最少是 K 次, 最长的长度.


map: 当前窗口字符出现次数统计(该值取值范围是 1..26)
vaildcount: 统计在当前的窗口, 字符出现次数>=k 的个数

分析: 给定的字符串都行小写字母, 那么该字符串的 UNIQUE 长度最大是 26. 那么是否可以每次去统计 unique 个字符, 在这个
unique 个字符当中, 计算每个字符的出现的次数, 当出现的次数==k, 则 validcount++, 该过程中使用 map 保存.

当 map.length > unique (可能存在的误区是 map.length 和 validcount 进行比较.), 说明 mp 当中存在某些字符的次数未
达到 k 次. 这个时候需要 left 移动了.

理想状况: map.length == unique && map.length == validcount, 这说明当前窗口是满足要求的.
*/
func LongestSubstringKRepeat(s string, k int) int {
	max := 0
	for unique := 1; unique <= 26; unique++ {
		validCount := 0
		left := 0
		mp := make(map[byte]int)
		for i := 0; i < len(s); i++ {
			cur := s[i]
			mp[cur] += 1 // 进
			if mp[cur] == k {
				validCount++
			}

			for len(mp) > unique {
				ch := s[left]
				mp[ch] -= 1 // 出
				if mp[ch] == k {
					validCount--
				}

				// 移除那些次数为0的字符
				if mp[ch] == 0 {
					delete(mp, ch)
				}

				left++
			}

			if unique == len(mp) && validCount == len(mp) {
				if max < i-left+1 {
					max = i - left + 1
				}
			}
		}
	}

	return max
}

//==================================================================================================

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

func MaxSubArrayI(nums []int) int {
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

func MaxSubArrayII(nums []int) int {
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

//==================================================================================================

/*
237: 滑动窗口最大值. 一个整数数组 nums, 大小为 k 的滑动窗口, 从数组的最左侧移动到数组的最右侧. 每次看到在滑动窗口内的
k 个数字. 滑动窗口每次只向右移动一位. 返回滑动窗口中的最大值.

对于每个滑动窗口, 可以使用 O(k) 的时间遍历其中的每一个元素, 找到其中的最大值. 那么时间复杂度为 O(nk)

对于相邻的滑动窗口, 它们公用着 k-1 个元素, 而只有1个元素是变化的. 可以根据这个特点进行优化.

1) 优先队列

对于 [最大值] 很容易想到一种结构体, 优先级队列(堆), 其中的大根堆可以实时维护一系列元素中的最大值.

对于本题, 初始化时, 可以将数组 nums 前 k 个元素放入优先级队列中. 每当向右移动窗口时, 可以把一个新的元素放入优先级队列
中, 此时堆顶就是所有元素的最大值. 这个最大值可能不在滑动窗口中, 这种状况下, "这个值在数组 nums 中的位置出现在滑动窗口
左边界的左侧". 在后续向右移动窗口时, 这个值就永远不可能出现在滑动窗口中了, 可以永久将其从优先队列中移除.

优先队列的元素存储二元组(num,index)

2) 单调队列

需要求出的是滑动窗口的最大值, 如果当前的滑动窗口中有两个下标 i, j, 其中 i < j, 并且 i 对应的元素不大于 j 对应的元素
(nums[i] <= nums[j]), 会怎样呢?

当滑动窗口向右移动时, "只要 i 还在窗口中, 那么 j 一定还在窗口中", 这是 i < j 保证的. 由于 nums[j] 的存在, nums[i]
"一定不会是滑动窗口的最大值了", 因此可以将 nums[i] 永久移除.

可以使用一个队列存储所有还没有被移除的下标. 在队列当中, 这些下标按照从小达到的顺序存储, 并且它们在数组 nums 中对应的值
是严格单调递减的. 因此如果队列中两个相邻的下标, 它们对应的值是相等或者递增, 前者是 i, 后者为 j, 此时可以永久性的移除
nums[i]

当滑动窗口右移时, 需要将一个新的元素放入到队列中, 为了保证队列的性质, 需要不断将新元素与队尾元素比较, 如果前者大于等于后
者, 则队尾元素就可以被移除.

*/

type PriorityQueue struct {
	sort.IntSlice
	arr []int
}

func (p PriorityQueue) Less(i, j int) bool {
	return p.arr[p.IntSlice[i]] > p.arr[p.IntSlice[j]]
}

func (p *PriorityQueue) Push(v interface{}) {
	p.IntSlice = append(p.IntSlice, v.(int))
}

func (p *PriorityQueue) Pop() interface{} {
	n := len(p.IntSlice)
	v := p.IntSlice[n-1]
	p.IntSlice = p.IntSlice[:n-1]
	return v
}

func MaxSlidingWindowI(nums []int, k int) []int {
	// 注: PriorityQueue 存储的是 index, 实际比较的是 index 背后的值
	q := &PriorityQueue{IntSlice: make([]int, k), arr: nums}
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

func MaxSlidingWindowII(nums []int, k int) []int {
	queue := make([]int, 0, k)
	push := func(idx int) {
		// i < j && nums[i] <= nums[j], 剔除 i
		for len(queue) > 0 && nums[queue[len(queue)-1]] <= nums[idx] {
			queue = queue[:len(queue)-1]
		}
		queue = append(queue, idx)
	}
	var ans []int
	for idx := 0; idx < k; idx++ {
		push(idx)
	}

	ans = append(ans, nums[queue[0]])

	for idx := k; idx < len(nums); idx++ {
		push(idx)
		for queue[0] <= idx-k {
			queue = queue[1:]
		}
		ans = append(ans, nums[queue[0]])
	}

	return ans
}

/*
438. 找到字符串中所有字母异位词.

时间复杂度: O(N)
*/
func findAnagrams(s string, p string) []int {
	if s == "" || p == "" || len(s) < len(p) {
		return nil
	}

	hash := make([]byte, 26)
	for _, v := range p {
		hash[v-'a'] += 1
	}

	N := len(p)
	var ans []int
	var count int
	for l, r := 0, 0; r < len(s); r++ {
		rval := s[r] - 'a'

		// 当前的字符是其中之一, 统计数量
		if hash[rval] > 0 {
			hash[rval]--
			count++
		}

		// l需要进行移动
		if r >= N {
			lval := s[l] - 'a'
			hash[lval]++
			if hash[lval] > 0 {
				count--
			}
			l++
		}

		if count == N {
			ans = append(ans, l)
		}
	}

	return ans
}
