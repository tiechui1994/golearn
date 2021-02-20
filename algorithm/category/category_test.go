package category

import (
	"fmt"
	"strings"
	"testing"
)

func TestMaxSlidingWindow(t *testing.T) {
	maxSlidingWindow := maxSlidingWindowII
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
