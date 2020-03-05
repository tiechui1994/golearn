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
func KMP(main string, sub string) int {
	return -1
}
