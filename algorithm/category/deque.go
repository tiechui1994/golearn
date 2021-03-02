package category

/*
单调队列:

所有的元素都是单调的(递增或者递减), 承载的基础数据结构是队列, 实现是双端队列, 队列中存入的元素为数组索引, 队头元素为
窗口的的最大(最小)元素.

"对头删除不符合有效窗口的元素, 队尾删除不符合最值的候选元素"

单调队列不是真正的队列. (队列都是FIFO的) 但单调队列是从队尾入列, 从队尾或队首出列.

- 去尾操作: 对尾元素出队列. 当队列有新元素等待入队列, 需要从对尾开始, 删除影响队列单调性的元素, 维护队列的单调性.

"去尾操作之后, 新的元素就要入队列了."

- 删头操作: 队头元素出队列. 判断队头元素是否在待求解的区间之内, 如果不在, 就将其删除.

经过上述两步操作之后, 队头元素就是当前区间的极值.

一般是先维护 k-1 个 size, 然后 offerLast.
*/

/*
239: 滑动窗口最大值.

862: 和至少为 K 的最短子数组(前缀和)

思路: 1) 前缀和转换. sum[i]-sum[k] >= k, j-i的最小值
	 2) 前缀和从1开始计算, sum[0] = 0, 应对极端特殊状况(所有的数组和 == k)
	 3) 单调递增队列

1425: 带限制的子序列和 (SubSequence最多可以跳K)

给你一个整数数组 nums 和一个整数 k, 返回"非空"子序列元素和的最大值, 子序列需要满足: 子序列中每两个"相邻"的整数 
nums[i] 和 nums[j], 它们在原数组中的下标 i 和 j 满足 i < j 且 j - i <= k.

数组的子序列定义为: 将数组中的若干个数字删除(可以删除 0 个数字), 剩下的数字按照原本的顺序排布.

思路: 1) 单调递减队列. dp[i] 作为单调递减的对象.
     2) 动态规划 dp[i] = max(nums[i], dp[i-k]+nums[i]) 1 <= k <= i-1


1438: 绝对差不超过限制的最长连续子数组

思路1: 滑动窗口 + 排序set

| set[0] - set[n-1] | <= limit => window = max(window, i-left+1)

否则, left++


思路2: 滑动窗口 + 双递增队列.

递增队列(iq) + 递减队列(dq)

iq.offerLast(i)
dq.offerLast(i)

nums[dq.0] - nums[iq.0] > limit, 需要修改区间了.(元素需要出队列了)


1696: 跳跃游戏, 思路和 1425是一样的
*/

func MaxSlidingWindow(nums []int, k int) []int {
	q := Deque{}
	ans := make([]int, 0)

	// 单调递减
	for i := 0; i < len(nums); i++ {
		index := i - q.peekFirst() - k
		// 去头, 保证窗口大小 k-1
		for !q.isEmpty() && i-q.peekFirst() >= k {
			q.pollFirst()
		}

		// 去尾, 保证递减队列
		for !q.isEmpty() && nums[q.peekLast()] <= nums[i] {
			q.pollLast()
		}

		// 进 q, 此时 q.size == k
		q.offerLast(i)

		if index >= 0 {
			ans = append(ans, nums[q.peekFirst()])
		}
	}

	return ans
}

func ShortestSubarray(nums []int, k int) int {
	sum := make([]int, len(nums)+1)
	sum[0] = 0

	for i := 0; i < len(nums); i++ {
		sum[i+1] = sum[i] + nums[i]
	}

	q := Deque{}
	ans := len(sum) + 1

	// 单调递增
	for i := 0; i < len(sum); i++ {
		// 满足条件
		for !q.isEmpty() && sum[i]-sum[q.peekFirst()] >= k {
			t := i - q.pollFirst()
			if t < ans {
				ans = t
			}
		}

		// 单调递增. 正常状况: sum[last] < sum[i]
		for !q.isEmpty() && sum[q.pollLast()] >= sum[i] {
			q.pollLast()
		}

		q.offerLast(i)
	}

	if ans == len(sum)+1 {
		ans = -1
	}

	return ans
}

func ConstrainedSubsetSum(nums []int, k int) int {
	q := Deque{}

	sum := make([]int, len(nums))

	res := nums[0]
	for i := 0; i < len(nums); i++ {
		sum[i] = nums[i]

		if !q.isEmpty() {
			sum[i] += sum[q.peekFirst()] // 计算当前的 sum[i] = max(nums[i], sum[j]+nums[i])
		}

		// 计算最大值.
		res = max(res, sum[i])

		// 单调递减 nums[last] > nums[i]
		for !q.isEmpty() && sum[q.peekLast()] <= sum[i] {
			q.pollLast()
		}

		for !q.isEmpty() && i-q.peekFirst() >= k {
			q.pollFirst()
		}

		q.offerLast(i)
	}

	return res
}

func LongestSubarray(nums []int, limit int) int {
	maxq := Deque{}
	minq := Deque{}

	ans := 0
	left := 0
	for i := 0; i < len(nums); i++ {
		// nums[last] > nums[i]
		for !maxq.isEmpty() && nums[minq.peekLast()] <= nums[i] {
			maxq.pollLast()
		}
		// nums[last] < nums[i]
		for !minq.isEmpty() && nums[minq.peekLast()] >= nums[i] {
			minq.pollLast()
		}

		maxq.offerLast(i)
		minq.offerLast(i)

		for nums[maxq.peekFirst()]-nums[minq.peekFirst()] > limit {
			if nums[i] == nums[maxq.peekFirst()] {
				maxq.pollFirst()
			}
			if nums[i] == nums[minq.peekFirst()] {
				minq.peekFirst()
			}
			left++
		}

		ans = max(ans, i-left+1)
	}

	return ans
}
