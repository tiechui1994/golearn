package string

import (
	"fmt"
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

	return str[start : end+1]
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

/*
3. 无重复字符的最长子串

给定一个字符串，请你找出其中不含有重复字符的 "最长子串" 的长度.

思路: 滑动窗口(双指针).

letters 记录"字符"的位置表

初始化的时候, start=0, end=1, 此时 letters 只有一个字符, 逐步滑动 end (++end), 如果出现 end 位置的字符在 letters
当中, 就删除 start 到 letters[cur] 之间的所有字符, 同时将 start=letters[cur]+1; 否则, 将字符加入letters, 计算
一下大小.
*/
func lengthOfLongestSubstring(s string) int {
	if len(s) <= 1 {
		return len(s)
	}

	runes := []rune(s)
	n := len(runes)

	letters := make(map[rune]int)
	start, end := 0, 1
	max := 1
	letters[runes[0]] = 0

	for start < n && end < n {
		r := runes[end]
		if idx, ok := letters[r]; ok {
			for i := start; i <= idx; i++ {
				delete(letters, runes[i])
			}

			start = idx + 1
		} else {
			letters[r] = end
			if len(letters) > max {
				max = len(letters)
			}

			end++
		}
	}

	return max
}

func findSubstring(s string, words []string) []int {
	if len(words) == 0 || len(words[0]) == 0 {
		return nil
	}

	isExist := func(word string, words []string) bool {
		for _, v := range words {
			if v == word {
				return true
			}
		}

		return false
	}

	positions := make([]int, len(words))
	for i := range words {
		positions[i] = -1
	}

	result := make([]int, 0)

	step := len(words[0])
	start, end := 0, 0

	for start < len(s) && end < len(s) {
		word := s[end : end+step]
		exist := isExist(word, words)

		if ok := isExist(word, words); ok && val < start {
			letters[word] = end

			if isAll(letters, start, end) {
				fmt.Println(letters, start, end)
				result = append(result, start)
			}
			end += step
		} else if val >= start {
			start = val + step
			letters[word] = end
			end += step
		} else {
			start = end + step
			end = end + step
		}
	}

	return result
}

func isAll(words map[string]int, start, end int) bool {
	count := 0
	for _, v := range words {
		if start <= v && v <= end {
			count += 1
		}
	}

	return count == len(words)
}
