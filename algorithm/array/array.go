package array

import (
	"fmt"
	"sort"
)

/*
15. 两数之和, 三数之和, K数之和

需要从两数之和开始说起, 首先进行排序, 排序的目的是可以从高低位置开始查找, 以达到扫描一次的效果.

sum = nums[low]+nums[high]

三种情况:

sum == target, 当前的 low 和 high 是满足条件的. 接下来只能同时增大low和减小high才能找到下一对. 注意去重操作.
sum > target, high--
sum < target, low++

三数之和, 是在两数之和的基础上增加了一个数. 先可以固定一个数(索引idx, 0<=idx<N-2), 这里要特别注意这个数必须每次不一样,
由于idx是从0开始, 条件为 idx>0 && nums[idx] != nums[idx-1]; 然后在剩下的 [idx+1,N-1] 范围内进行查找两数之和为
target-nums[idx].

K数之和, 思路基本上三数之和是一样的, 只不过要使用递归来求解

**/
// 三个数字的和是0的所有组合, 空间: O(N) 时间:O(N^2)
func threeSum(nums []int) [][]int {
	sort.SliceStable(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})

	var result [][]int
	N := len(nums)
	if nums[0] > 0 && nums[N-1] < 0 {
		return result
	}

	for k := 0; k < N-2; k++ {
		if nums[k] > 0 {
			break
		}

		// 重复的数字(向下去重)
		if k > 0 && nums[k] == nums[k-1] {
			continue
		}

		// k 锁定一个数, 剩下数的范围是 [k+1, N-1], 在这个范围内可以高低指针双向查找.
		i := k + 1
		j := N - 1
		for i < j {
			sum := nums[k] + nums[i] + nums[j]
			if sum == 0 {
				result = append(result, []int{nums[i], nums[j], nums[k]})
				for i < j && nums[i] == nums[i+1] {
					i++
				}
				for i < j && nums[j] == nums[j-1] {
					j--
				}

				i++
				j--
			} else if sum < 0 {
				i++
			} else {
				j--
			}
		}
	}

	return result
}

func kSum(nums []int, k, target int) [][]int {
	sort.SliceStable(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})

	var ksum func(nums []int, start, k, target int) [][]int
	ksum = func(nums []int, startIndex, k, target int) [][]int {
		N := len(nums)
		var result [][]int

		// 两数之和为特定值, 一次遍历, O(N)
		if k == 2 {
			var left, right = startIndex, N - 1
			for left < right {
				sum := nums[left] + nums[right]
				if sum == target {
					result = append(result, []int{nums[left], nums[right]})
					// 跳过重复的数
					for left < right && nums[left] == nums[left+1] {
						left++
					}
					for left < right && nums[right] == nums[right-1] {
						right--
					}

					// 非重复的数字
					left++
					right--
				} else if sum < target {
					left++
				} else {
					right--
				}
			}

			return result
		}

		// 多数之和.
		end := N - (k - 1)
		for i := startIndex; i < end; i++ {
			// 相同的数, 跳跃之. 需要注意的事项是向前看
			if i > startIndex && nums[i] == nums[i-1] {
				continue
			}

			// 锁定一个数 i, 在剩下的范围内查找 [i+1,N-1] 查找剩下的 k-1 个数的和是 target-nums[i]
			// 查找成功之后, 在每个结果当中追加当前的数字.
			temp := ksum(nums, i+1, k-1, target-nums[i])
			for j := range temp {
				temp[j] = append([]int{nums[i]}, temp[j]...)
			}
			result = append(result, temp...)
		}

		return result
	}

	return ksum(nums, 0, k, target)
}

//==================================================================================================

// 有序的数组, 去除重复的数字 ==> 双指针
func removeDuplicates(nums []int) int {
	fmt.Println(nums)
	if len(nums) == 0 {
		return 0
	}

	var k int
	for i := 1; i < len(nums); i++ {
		if nums[i] != nums[k] {
			k++
			nums[k] = nums[i]
		}
	}

	return k + 1
}

// 无序的数组, 去除某个所有的元素 ==> 双指针
func removeElement(nums []int, val int) int {
	if len(nums) == 0 {
		return 0
	}

	var k int
	for i := 0; i < len(nums); i++ {
		if nums[i] != val {
			nums[k] = nums[i]
			k++
		}
	}

	return k
}

// 波峰问题. 二分查找的特性
// 条件: arr[0]<arr[1] a[m] < a[m-1]
// 波峰: arr[m-1] < arr[m] && arr[m] > arr[m+1]
// a[m-1] < a[m] < a[m+1], 增长
// a[m-1] > a[m] > a[m+1], 减少
func findPeek(nums []int) (bool, int) {
	// 单调递增
	isupper := func(nums []int, m int) bool {
		return nums[m-1] < nums[m] && nums[m] < nums[m+1]
	}

	if nums == nil || len(nums) <= 2 {
		return false, -1
	}

	start := 0
	end := len(nums) - 1
	for start <= end {
		m := (start + end) / 2
		if m == 0 || m == len(nums)-1 {
			return false, -1
		}

		if nums[m-1] < nums[m] && nums[m] > nums[m+1] {
			return true, nums[m]
		}

		if isupper(nums, m) {
			start = m + 1
		} else {
			end = m - 1
		}
	}

	return false, -1
}

//==================================================================================================

// 旋转矩阵
func spiralOrder(matrix [][]int) []int {
	tR := 0
	tC := 0
	dR := len(matrix)
	dC := len(matrix[0])

	var res []int
	print := func(tR, tC, dR, dC int) {
		if tR == dR {
			for i := tC; i < dC; i++ {
				res = append(res, matrix[i][tR])
			}
		} else if tC == dC {
			for i := tR; i < dR; i++ {
				res = append(res, matrix[tC][i])
			}
		} else {
			curR := tR
			curC := tC
			for curC != dC {
				res = append(res, matrix[tR][curC])
				curC++
			}
			for curR != dR {
				res = append(res, matrix[curR][dC])
				curR++
			}
			for curC != tC {
				res = append(res, matrix[dR][curC])
				curC--
			}
			for curR != tR {
				res = append(res, matrix[curR][tC])
				curR--
			}
		}
	}

	for tR <= dR && tC <= dC {
		print(tR, tC, dR, dC)
		tR++
		tC++
		dR--
		dC--
	}

	return res
}

//==================================================================================================

//原题: 传送带上的第 i 个包裹的重量为 weights[i]. 每一天, 我们都会按给出重量的顺序往传送带上装载包裹. 我们装载的重量
//不会超过船的最大运载重量.
//返回能在 D 天内将传送带上的所有包裹送达的船的最低运载能力.
//
//2. 一个必须依照数组顺序完成的工作, 数字代表工作难易度, 分成K天完工, 尽可能把困难度最高的一天变得比较不累, 求最累的一天
// 的困难度
//3. 一个数组代表一排的书, 数字代表页数, 现在必须把相邻的书分成K组放置到K台打印机, 设置一个配置方法使得需要打印最多页数
// 的机器打印最少页, 求工作量最多的打印机需要打印的页数.
//4. 一个包裹数组, 数字代表重量, 依包裹排列顺序分成K批寄送, 使得最重的一批重量最小, 求最重一批重量.
func shipWithinDays(weights []int, D int) int {
	max := 0
	sum := 0
	for _, w := range weights {
		if w > max {
			max = w
		}
		sum += w
	}

	low, high := max, sum
	for low <= high {
		mid := (low + high) / 2

		sum := 0
		day := 1
		for _, w := range weights {
			sum += w
			if sum == mid {
				day += 1
				sum = 0
			}
			if sum > mid {
				day += 1
				sum = w
			}
		}

		if day > D {
			low = mid + 1
		} else if day <= D {
			high = mid - 1
		}
	}

	return low
}

//==================================================================================================

/*
33. 搜索旋转排序数组

该整数数组原本是按升序排列, 但输入时在预先未知的某个点上进行了旋转. (例如, 数组 [0,1,2,4,5,6,7] 可能变为
[4,5,6,7,0,1,2] )

*/

func search(nums []int, target int) int {
	binarySearch := func(left, right int) int {
		for left <= right {
			mid := (left + right) / 2
			if nums[mid] == target {
				return mid
			}

			if nums[mid] < target {
				left = mid + 1
			} else {
				right = mid - 1
			}
		}

		return -1
	}

	findRoateIndex := func(left, right int) int {
		if nums[left] < nums[right] {
			return 0
		}
		for left <= right {
			mid := (left + right) / 2
			if nums[mid] > nums[mid+1] {
				return mid + 1
			}
			if nums[mid] < nums[left] {
				right = mid - 1
			} else {
				left = mid + 1
			}
		}

		return 0
	}

	index := findRoateIndex(0, len(nums)-1)
	if nums[index] == target {
		return index
	}

	if index == 0 {
		return binarySearch(0, len(nums)-1)
	}
	if target < nums[0] {
		return binarySearch(index, len(nums)-1)
	}

	return binarySearch(0, index)
}

//==================================================================================================

// 螺旋矩阵I
// 给定一个包含 m * n 个元素的矩阵(m行, n列), 请按照顺时针螺旋顺序, 返回矩阵中的所有元素.
func spiralMatrixI(matrix [][]int) []int {
	if len(matrix) == 0 {
		return nil
	}

	if len(matrix) == 1 {
		return matrix[0]
	}

	if len(matrix[0]) == 0 {
		return nil
	}

	tR := 0
	tC := 0
	dR := len(matrix) - 1
	dC := len(matrix[0]) - 1

	var res []int
	print := func(tR, tC, dR, dC int) {
		if tR == dR {
			for i := tC; i <= dC; i++ {
				res = append(res, matrix[tR][i])
			}
		} else if tC == dC {
			for i := tR; i <= dR; i++ {
				res = append(res, matrix[i][tC])
			}
		} else {
			curR := tR
			curC := tC
			for curC != dC {
				res = append(res, matrix[tR][curC])
				curC++
			}
			for curR != dR {
				res = append(res, matrix[curR][dC])
				curR++
			}
			for curC != tC {
				res = append(res, matrix[dR][curC])
				curC--
			}
			for curR != tR {
				res = append(res, matrix[curR][tC])
				curR--
			}
		}
	}

	for tR <= dR && tC <= dC {
		print(tR, tC, dR, dC)
		tR++
		tC++
		dR--
		dC--
	}

	return res
}

func spiralMatrixII(n int) [][]int {
	if n == 0 {
		return nil
	}

	if n == 1 {
		return [][]int{{1}}
	}

	var res = make([][]int, n)
	var num = 1
	for i := 0; i < n; i++ {
		res[i] = make([]int, n)
	}

	tR := 0
	tC := 0
	dR := n - 1
	dC := n - 1

	print := func(tR, tC, dR, dC int) {
		if tR == dR {
			for i := tC; i <= dC; i++ {
				res[tR][i] = num
				num++
			}
		} else if tC == dC {
			for i := tR; i <= dR; i++ {
				res[i][tC] = num
				num++
			}
		} else {
			curR := tR
			curC := tC
			for curC != dC {
				res[tR][curC] = num
				num++
				curC++
			}
			for curR != dR {
				res[curR][dC] = num
				num++
				curR++
			}
			for curC != tC {
				res[dR][curC] = num
				num++
				curC--
			}
			for curR != tR {
				res[curR][tC] = num
				num++
				curR--
			}
		}
	}

	for tR <= dR && tC <= dC {
		print(tR, tC, dR, dC)
		tR++
		tC++
		dR--
		dC--
	}

	return res
}

// 螺旋行走, 最直观的解法
// R,C是范围, r0,c0是开始的位置
// 在每个方向的行走长度, 发现如下模式: 1,1,2,2,3,3,4,4,... 即先向东走1单位, 然后向南走1单位,
// 再向西走2单位, 再向北走2单位,
// 再向东走3单位, 等等. 因为我们的行走方式是自相似的, 所以这种模式会以我们期望的方式重复.
func spiralMatrixIII(R int, C int, r0 int, c0 int) [][]int {
	var res = make([][]int, R*C)
	var size int
	res[size] = []int{r0, c0}
	size++
	if size == R*C {
		return res
	}

	var step = 0
	for size < R*C {
		step += 1
		for i := 0; i < step; i++ {
			c0++
			if 0 <= c0 && c0 < C && 0 <= r0 && r0 < R {
				res[size] = []int{r0, c0}
				size++
			}
		}

		for i := 0; i < step; i++ {
			r0++
			if 0 <= c0 && c0 < C && 0 <= r0 && r0 < R {
				res[size] = []int{r0, c0}
				size++
			}
		}

		step++

		for i := 0; i < step; i++ {
			c0--
			if 0 <= c0 && c0 < C && 0 <= r0 && r0 < R {
				res[size] = []int{r0, c0}
				size++
			}
		}

		for i := 0; i < step; i++ {
			r0--
			if 0 <= c0 && c0 < C && 0 <= r0 && r0 < R {
				res[size] = []int{r0, c0}
				size++
			}
		}

	}

	return res
}

//==================================================================================================

func findRoateIndex(nums []int) int {
	var left, right = 0, len(nums) - 1
	if nums[left] < nums[right] {
		return 0
	}

	for left <= right {
		mid := (left + right) / 2
		if nums[left] <= nums[mid] && nums[mid] > nums[right] {
			left = mid + 1
		}
		if nums[left] > nums[mid] && nums[mid] <= nums[right] {
			right = mid
		}
		if nums[left] <= nums[mid] && nums[mid] <= nums[right] {
			right--
		}
	}

	return left
}

//==================================================================================================

func searchRange(nums []int, target int) []int {
	if nums == nil {
		return []int{-1, -1}
	}

	find := func(isleft bool) int {
		low, high := 0, len(nums)
		for low < high {
			mid := (low + high) / 2
			if nums[mid] > target || nums[mid] == target && isleft {
				high = mid
			} else {
				low = mid + 1
			}
		}

		return low
	}

	res := []int{-1, -1}
	leftIndex := find(true)
	if leftIndex == len(nums) || nums[leftIndex] != target {
		return res
	}
	res[0] = leftIndex
	res[1] = find(false) - 1
	return res
}
