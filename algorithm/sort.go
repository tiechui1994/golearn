package algorithm

// 插入排序, 从当前位置开始向已排好序列交换, 更新当前位置(0), 直到合适的位置
func InsertSort(nums []int) {
	if len(nums) <= 1 {
		return
	}

	var n = len(nums)
	for i := 1; i < n; i++ {
		for j := i; j > 0; {
			if nums[j-1] > nums[j] {
				nums[j-1], nums[j] = nums[j], nums[j-1]
			}
			j = j - 1
		}
	}
}

// 选择排序. 每次从未排好中选择最小的值的索引, 交换之
func SelectSort(nums []int) {
	n := len(nums)
	for i := 0; i < n; i++ {
		var index = i
		for j := i; j < n; j++ {
			if nums[j] < nums[index] {
				index = j
			}
		}
		nums[i], nums[index] = nums[index], nums[i]
	}
}

// 快速排序, 核心是定位
func QuickSort(nums []int) {
	mid := func(left, right int, nums []int) int {
		low := left
		high := right
		pivot := nums[low]

		for low < high {
			for low < high && nums[high] >= pivot {
				high--
			}
			nums[low] = nums[high] // 比中轴小的记录移到低端

			for low < high && nums[low] <= pivot {
				low++
			}
			nums[high] = nums[low] // 比中轴大的记录移到最高端
		}

		nums[low] = pivot // 中轴记录到尾

		return low
	}

	var sort func(left, right int, nums []int)
	sort = func(left, right int, nums []int) {
		if left < right {
			mid := mid(left, right, nums)
			sort(left, mid-1, nums)
			sort(mid+1, right, nums)
		}

	}

	sort(0, len(nums)-1, nums)
}

func QuickSort2(nums []int) {
	var sort func(left, right int, nums []int)
	sort = func(left, right int, nums []int) {
		pivot := nums[(left+right)/2]
		i := left
		j := right

		for i <= j {
			for nums[i] < pivot {
				i++
			}

			for nums[j] > pivot {
				j--
			}

			if i <= j {
				nums[i], nums[j] = nums[j], nums[i]
				i++
				j--
			}
		}

		if left < j {
			sort(left, j, nums)
		}
		if i < right {
			sort(i, right, nums)
		}
	}

	sort(0, len(nums)-1, nums)
}

// 分解 + 归并
func MergeSort(nums []int) {
	if len(nums) <= 1 {
		return
	}

	merge := func(left, mid, right int, nums []int) {
		n1, n2 := mid-left+1, right-mid
		arr1, arr2 := make([]int, n1), make([]int, n2)

		copy(arr1, nums[left:mid+1])
		copy(arr2, nums[mid+1:right+1])

		i, j := 0, 0
		k := left

		for i < n1 && j < n2 {
			if arr1[i] <= arr2[j] {
				nums[k] = arr1[i]
				i++
			} else {
				nums[k] = arr2[j]
				j++
			}

			k++
		}

		for ; i < n1; i++ {
			nums[k] = arr1[i]
			k++
		}
		for ; j < n2; j++ {
			nums[k] = arr2[j]
			k++
		}
	}

	var sort func(left, right int, nums []int)
	sort = func(left, right int, nums []int) {
		if left < right {
			mid := (left + right) / 2
			sort(left, mid, nums)
			sort(mid+1, right, nums)
			merge(left, mid, right, nums)
		}
	}

	sort(0, len(nums)-1, nums)
}

func HeapSort(nums []int) {
	var adjust func(index, size int, nums []int)
	adjust = func(index, size int, nums []int) {
		if index < size {
			left := (index << 1) + 1
			right := left + 1
			max := index
			if left < size && nums[max] < nums[left] {
				max = left
			}
			if right < size && nums[max] < nums[right] {
				max = right
			}

			if max != index {
				nums[max], nums[index] = nums[index], nums[max]
				adjust(max, size, nums)
			}
		}
	}

	buildheap := func(size int, nums []int) {
		for i := (size - 1) / 2; i >= 0; i-- {
			adjust(i, size, nums)
		}
	}

	buildheap(len(nums), nums) // 构建大顶堆
	nums[0], nums[len(nums)-1] = nums[len(nums)-1], nums[0]

	// 排序
	for i := 1; i < len(nums); i++ {
		adjust(0, len(nums)-i, nums)
		nums[0], nums[len(nums)-1-i] = nums[len(nums)-1-i], nums[0]
	}
}
