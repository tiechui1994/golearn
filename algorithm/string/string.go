package string

import (
	"strconv"
)

// 求string组成的ip可能值

func stringIP(s string) []string {
	isNum := func(num string) bool {
		realNum, _ := strconv.ParseInt(num, 10, 64)
		realStr := strconv.FormatInt(realNum, 10)

		if 0 <= realNum && realNum <= 255 && realStr == num {
			return true
		}

		return false
	}

	var RestoreIP func(curstr string, remain, i int, ip string, list *[]string)
	RestoreIP = func(curstr string, remain, i int, ip string, list *[]string) {
		if remain == 0 {
			if len(curstr) == i {
				*list = append(*list, ip)
			}
			return
		}

		for k := i + 1; k <= len(curstr); k++ {
			if isNum(curstr[i:k]) {
				newip := curstr[i:k]
				if ip != "" {
					newip = ip + "." + newip
				}

				RestoreIP(curstr, remain-1, k, newip, list)
			} else {
				break
			}
		}
	}

	if len(s) < 4 || len(s) > 12 {
		return nil
	}

	res := make([]string, 0)
	RestoreIP(s, 4, 0, "", &res)
	return res
}

// 最长回文子串: 中心扩展算法
// O(n^2)

func longPalindromeI(str string) string {
	expandCenter := func(left, right int) int {
		L := left
		R := right
		for L >= 0 && R < len(str) && str[L] == str[R] {
			L--
			R++
		}

		return R - L - 1
	}

	var start, end = 0, 0
	for i := 0; i < len(str); i++ {
		L1 := expandCenter(i, i)
		L2 := expandCenter(i, i+1)
		length := L1 - L2
		if length < 0 {
			length = -length
		}
		if length > end-start {
			start = i - (length-1)/2
			end = i + length/2
		}
	}

	return str[start: end+1]
}

// 784 letter-case-permutation
// 字母大小写全排列

func letterCasePermutation(s string) []string {
	var runes = []rune(s)
	var res []string
	var dfs func(i int, temp []rune)
	dfs = func(i int, temp []rune) {
		if i == len(s) {
			res = append(res, string(temp))
			return
		}

		dfs(i+1, append(temp, runes[i]))
		if 'a' <= runes[i] && runes[i] <= 'z' {
			dfs(i+1, append(temp, runes[i]-'a'+'A'))
		}
		if 'A' <= runes[i] && runes[i] <= 'Z' {
			dfs(i+1, append(temp, runes[i]-'A'+'a'))
		}
	}

	dfs(0, nil)
	return res
}
