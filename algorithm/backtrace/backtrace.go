package backtrace

/*
N皇后问题
*/
func NQueue(n int) [][]string {
	var ans = make([][]string, n)
	isValid := func(cur [][]string, row, col int) bool {
		for i := row - 1; i >= 0; i-- {
			if cur[i][col] == "Q" {
				return false
			}
		}

		i := row - 1
		j := col + 1
		for i >= 0 && j <= n-1 {
			if cur[i][j] == "Q" {
				return false
			}
			i--
			j++
		}

		i = row - 1
		j = col - 1
		for i >= 0 && j >= 0 {
			if cur[i][j] == "Q" {
				return false
			}
			i--
			j--
		}

		return true
	}

	var traverse func(cur [][]string, row int) bool
	traverse = func(cur [][]string, row int) bool {
		if row == n {
			copy(ans, cur)
			return true
		}

		for i := 0; i < n; i++ {
			if !isValid(cur, row, i) {
				continue
			}

			cur[row][i] = "Q"
			if traverse(cur, row+1) {
				return true
			}
			cur[row][i] = ""
		}

		return false
	}

	result := make([][]string, n)
	for i := 0; i < n; i++ {
		result[i] = make([]string, n)
	}
	if traverse(result, 0) {
		return ans
	}

	return nil
}
