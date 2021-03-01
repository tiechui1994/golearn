package category

/*
单调队列:

所有的元素都是单调的(递增或者递减), 承载的基础数据结构是队列, 实现是双端队列, 队列中存入的元素为数组索引, 队头元素为
窗口的的最大(最小)元素.

"对头删除不符合有效窗口的元素, 队尾删除不符合最值的候选元素"

单调队列不是真正的队列. (队列都是FIFO的) 但单调队列是从队尾入列, 从队尾或队首出列.
*/

/*
862: 子数组(SubArray)和至少的K的最小数组(转换为前缀和)

1425: Max Sum of SubSequence (SubSequence最多可以跳K)

1438:

1696: u
*/

func MonotonicQueue(nums []int, k int) []int {
	q := Deque{}
	ans := make([]int, 0)

	for i := 0; i < len(nums); i++ {
		// 头出, 保证窗口大小 k-1
		for !q.isEmpty() && i-q.peekFirst() >= k {
			q.pollFirst()
		}

		// 尾出, 保证递减队列
		for !q.isEmpty() && nums[q.peekLast()] <= nums[i] {
			q.pollLast()
		}

		// 进 q, 此时 q.size == k
		q.offerLast(i)
		ans = append(ans, nums[q.peekFirst()])
	}

	return ans
}
