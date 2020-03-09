package algorithm

/**
@ 字符串查询算法
*/

// BF算法, 暴力查询
func BruteForce(main string, sub string) int {
	if main == "" || sub == "" {
		return -1
	}

	for i := 0; i < len(main); {
		for j := 0; j < len(sub); {
			if main[i] == sub[j] {
				i++
				j++
			} else {
				i = i - (j - 1)
				j = 0
			}

			if j == len(sub)-1 {
				return i - j
			}
		}
	}

	return -1
}

// KMP算法
//
// BF 算法效率低在于每轮主串从pos位置开始比较过程中, 若字符不相等, 主串会回退到 pos+1 位置, 然后与匹配串第一个位置字符开始比较.
//
// KMP 算法首先算出一个next数组, 匹配串每轮匹配在 j 位置失配时, 匹配串向右滑动的距离为 j - next[j].
// next 数组的计算规则如下. 其中 j 为匹配串索引：
//
// 1. j=0 时, next[j] = -1;
//
// 2. j>0, 匹配串 0 至 j-1 位置的子串前缀和后缀数组中最长相同元素的长度为maxL, 则next[j] = maxL.
// 特殊的, 当j=1时,前缀和后缀都为空, 所以maxL=0, next[1] = 0:
//
func KMP(main string, sub string) int {

	//getNext := func(sub string) []int {
	//	length := len(sub)
	//	next := make([]int, length)
	//	next[0] = -1
	//
	//	// 保存前缀和后缀数组最长相同子串
	//	k := 0
	//	for j := 1; j < length-1; j++ {
	//		for (k > 0) && (sub[j-1] != sub[k]) {
	//			k = next[j-1]
	//		}
	//		if (j > 1) && (sub[j-1] == sub[k]) {
	//			k++
	//		}
	//		next[j] = k
	//	}
	//
	//	return next
	//}

	next := getnext(sub)
	j := 0
	i := 0
	for i < len(main) && j < len(sub) {
		if j == -1 || main[i] == sub[j] {
			i++
			j++
		} else {
			j = next[j]
		}

		if j == len(sub)-1 {
			return i - j
		}
	}

	return -1
}

func getnext(p string) []int {
	length := len(p)
	next := make([]int, length)
	next[0] = -1

	k := -1
	for j := 0; j < length-1; {
		if k == -1 || p[j] == p[k] {
			k++
			j++
			next[j] = k
		} else {
			k = next[k]
		}
	}

	return next
}
