package leetcode

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
