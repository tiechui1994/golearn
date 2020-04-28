package dp

/**

对于一个数字序列, 请设计一个复杂度为O(nlogn)的算法, 返回该序列的最长上升子序列的长度,
这里的子序列定义为这样一个序列U1，U2…，其中Ui < Ui+1, 且A[Ui] < A[Ui+1].
给定一个数字序列A及序列的长度n，请返回最长上升子序列的长度.

**/

// remain[i] 以i为结尾的最长的上升子序列的长度
// remain[i] = max(remain[k]+1) remain[k]<arr[i]  0 <= k <i
func LongIncrSeq1(arr []int) int {
	n := len(arr)
	remain := make([]int, n)
	remain[0] = 1

	max := func(i, j int) int {
		if i > j {
			return i
		}
		return j
	}

	for i := 1; i < n; i++ {
		val := 1
		for k := i - 1; k >= 0; k-- {
			if arr[k] < arr[i] {
				val = max(val, remain[k]+1)
			}
		}
		remain[i] = val
	}

	return remain[n-1]
}

// ends[i] 表示长度为i的最长上升子序列, 结尾的最小值. ends始终是递增的
func LongIncrSeq2(arr []int) int {
	var n = len(arr)
	var ends = make([]int, n)
	ends[0] = arr[0]
	length := 0 // 当前ends递增序列的最大长度

	max := func(i, j int) int {
		if i > j {
			return i
		}
		return j
	}

	for i := 1; i < n; i++ {
		left := 0
		right := length
		// 查找ends当中第一个大于arr[i]的位置
		for left <= right {
			mid := (left + right) / 2
			if ends[mid] > arr[i] {
				right = mid - 1
			} else {
				left = mid + 1
			}
		}
		length = max(length, left)
		ends[left] = arr[i]
	}

	return length + 1
}
