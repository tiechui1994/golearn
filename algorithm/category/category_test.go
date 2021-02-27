package category

import (
	"fmt"
	"strings"
	"testing"
)

func TestMaxSlidingWindow(t *testing.T) {
	maxSlidingWindow := MaxSlidingWindowII
	case1 := []int{1, 3, -1, -3, 5, 3, 6, 7}
	casek := 3
	ans := maxSlidingWindow(case1, casek)
	t.Logf("ans:%v, real:%v", ans, []int{3, 3, 5, 5, 6, 7})
}

func TestPriorityQueue_Less(t *testing.T) {
	fmt.Println(2 ^ 2 ^ 2)
}

func TestCanIWin(t *testing.T) {
	N := 4
	t.Logf("N: %v, CanWin:%v", N, CanIWin(N))
}

func TestWordBreakII(t *testing.T) {
	s1 := "catsanddog"
	dict1 := []string{"cat", "cats", "and", "sand", "dog"}
	res1 := WordBreakII(s1, dict1)
	t.Log("\n" + strings.Join(res1, "\n"))
}

func TestLengthLIS(t *testing.T) {
	case1 := []int{10, 9, 2, 5, 7, 101, 18}
	t.Log(case1, "LIS", LengthLIS2(case1))
}

func TestMinCutCost(t *testing.T) {
	n := 7
	cuts := []int{1, 3, 4, 5}
	t.Log("cost", MinCutCost(n, cuts), "real", 16)

	t.Log("LPS", LongestPalindromeSubseq("bbbab"), "real", 4)
}

func TestMinDamage(t *testing.T) {
	t.Log(MinDamage([]int{3, 5, 7}, []int{8, 2, 0}))
}

// 最小覆盖子串
func TestMinWindow(t *testing.T) {
	//t.Log("MinWindow", MinWindow("ADOBECODEBANC", "ABC"))
	t.Log("MinWindow", MinWindow("BBA", "AB"))
}

func TestLongestSubstring(t *testing.T) {
	t.Log("LongestSubstring", LongestSubstring("ababbc", 2))
}
