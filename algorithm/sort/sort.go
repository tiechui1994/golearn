package sort

/*
快速排序: (冒泡排序的改进版本)

通过一趟排序将待排的数据分割成两个序列, 其中左边子序列的元素都比右边的子序列元素小.
然后, 分别对左边的子序列和右边的子序列继续进行排序, 达到整个序列有序的目的.

主要做两件事情:
第一: 从什么位置划分原始序列, 将其变成两个左右序列, 即确定 pivot(哨兵, 存在优化点) 的位置
第二: 分别堆左右两个子序列进行递归操作, 继续分割使其变得有序

关于分区需要注意的点:

一般状况下, 选择的是 left 位置的元素作为 pivot(哨兵); 那么需要先 "从 right 位置查找一个位置, 此位置的元素 < pivot",
然后交换left, right, 接着 "从 left 位置查找一个位置, 此位置的元素 > pivot", 然后交换left, right.

注意: 上述的条件都是在 left < right 的前提下进行的. 上面的顺序不能乱了(乱了就出问题了)
*/

func QuickSort(arr []int) {
	partion := func(left, right int, arr []int) int {
		privot := arr[left]
		for left < right {
			for left < right && arr[right] >= privot {
				right--
			}
			arr[right], arr[left] = arr[left], arr[right]

			for left < right && arr[left] <= privot {
				left++
			}
			arr[right], arr[left] = arr[left], arr[right]
		}

		return left
	}

	var sort func(left, right int, arr []int)
	sort = func(left, right int, arr []int) {
		if left < right {
			idx := partion(left, right, arr)
			sort(left, idx-1, arr)
			sort(idx+1, right, arr)
		}
	}

	sort(0, len(arr)-1, arr)
}

/*
归并排序; 递归拆分+合并.

以 mid 节点进行拆分(当数组的长度为1的时候就不能拆分了), 拆分的返回的还是一个 arr.

合并注意的事项:
1. 记录 left, right 当前已经合并的索引
2. 合并剩下的 left 或者 right
**/
func MergeSort(arr []int) {
	merge := func(left, right []int) []int {
		result := make([]int, 0, len(left)+len(right))

		var i, j int
		for i < len(left) && j < len(right) {
			if left[i] <= right[j] {
				result = append(result, left[i])
				i++
			} else {
				result = append(result, right[j])
				j++
			}
		}

		for k := i; k < len(left); k++ {
			result = append(result, left[k])
		}
		for k := j; k < len(right); k++ {
			result = append(result, right[k])
		}

		return result
	}

	var sort func(arr []int) []int
	sort = func(arr []int) []int {
		if len(arr) <= 1 {
			return arr
		}
		mid := len(arr) / 2
		return merge(sort(arr[:mid]), sort(arr[mid:]))
	}

	result := sort(arr)
	for i, v := range result {
		arr[i] = v
	}
}

/*
堆排序(关键是思维和思路,一定要清晰)

1. 堆调整, 自顶向下递归调整方式. 注意就是根据堆的规则调整. idx是要调整的位置, length是当前堆的大小, 索引不能超过该大小

2. 构造堆, 初始化的堆该从什么位置调整? idx = N/2-1
          调整的顺序是什么? "自底向上", 从第一个非叶子节点开始调整, 直到 0 好节点

3. 排序过程, 每次拿走第一个位置的元素(最大元素)放置到对应位置后, 接下来将之前的最低层的位置的元素移动到首位, 在这之后,
堆该怎样调整? "自顶向下", 此时的堆底层是稳定的, 但是上层不稳定, 只能从上层不稳定的位置调整, 调整的时候只会影响一个分支,
另一个分支依旧是稳定的.
**/
func HeapSort(arr []int) {
	var adjust func(idx, length int, arr []int)
	adjust = func(idx, length int, arr []int) {
		left := 2*idx + 1
		right := 2*idx + 2
		index := idx
		max := arr[idx]
		if left < length && max < arr[left] {
			max = arr[left]
			index = left
		}

		if right < length && max < arr[right] {
			max = arr[right]
			index = right
		}

		if idx != index {
			arr[index], arr[idx] = arr[idx], arr[index]
			adjust(index, length, arr)
		}
	}

	build := func(arr []int) {
		for i := len(arr)/2 - 1; i >= 0; i-- {
			adjust(i, len(arr), arr)
		}
	}

	build(arr)

	n := len(arr)
	arr[n-1], arr[0] = arr[0], arr[n-1]
	for i := n - 1; i >= 1; i-- {
		adjust(0, i, arr)
		arr[i-1], arr[0] = arr[0], arr[i-1]
	}
}
