package category

/*
动态规划一: 博弈类
*/

/*
292: 捡石子游戏 (一次可以最多三个石子, 拿到最后一个的赢)

递推公式

dp[i] 表示 i 个石子, 结果

dp[i] = !(dp[i-1] || dp[i-2] || dp[i-3])

第 i 个石子, 取决于第 i-1, i-2, i-3 的状态(对手的状态), 如果任何一个对手状态都是true, 那么你必输.
*/

func CanIWin(n int) bool {
	const (
		Null = iota
		True
		False
	)

	var dfs func(n int, memo []int) bool
	dfs = func(n int, memo []int) bool {
		if n < 0 {
			return false
		}
		if memo[n] != Null {
			return memo[n] == True
		}

		var res uint

		for i := 1; i < 4; i++ {
			if n >= i {
				if dfs(n-i, memo) {
					res |= 0
				} else {
					res |= 1
				}
			}
		}

		memo[n] = int(res)
		return memo[n] == True
	}

	mem := make([]int, n+1)
	return dfs(n, mem)
}

/*
486 : 一个数组, 每次只能从数组的左边或者右边拿数, 最终拿到数字之和最大的赢. 如果你先拿, 是否能赢?
*/

func PredictTheWinner(arr []int) bool {
	const Max = int(1>>64 - 1)

	/**
	dfs 的作用就是计算 i, j 之间最大差距
	i, j 表示数组的索引范围
	memo[i][j] 表示在范围 i..j 之间最大结果最大差距(我和对手), 那么 memo[0][n-1] 则是结果

	memo[i][j] = max( arr[i] - memo[i+1][j], arr[j] - memo[i][j-1] )
	*/
	var dfs func(arr []int, i, j int, memo [][]int) int
	dfs = func(arr []int, i, j int, memo [][]int) int {
		if i > j {
			return 0
		}

		if memo[i][j] != Max {
			return memo[i][j]
		}

		memo[i][j] = max(arr[i]-dfs(arr, i+1, j, memo),
			arr[j]-dfs(arr, i, j-1, memo))
		return memo[i][j]
	}

	n := len(arr)
	var memo [][]int
	for i := 0; i < n; i++ {
		item := make([]int, n)
		for j := 0; j < n; j++ {
			item[j] = Max
		}
		memo = append(memo, item)
	}

	return dfs(arr, 0, n-1, memo) >= 0
}

/*
1025: 最初有一个数字 N, 然后选择一个数字 x, 数字 x 满足的要求: 0 < x < N, 并且 N % x == 0,
接着使用 N-x 代替原来的 N, 直到玩家无法执行上述的操作. 能拿到最后一个数字的玩家获胜.
*/

func DivisorGame(n int) bool {
	const (
		Null = iota
		True
		False
	)

	var dfs func(n int, memo []int) bool

	// 表示当数字是 n 时, 能赢否
	dfs = func(n int, memo []int) bool {
		if n == 1 {
			return false
		}
		if memo[n] != Null {
			return memo[n] == True
		}
		var canWin = False
		// 选择数字 x, 并计算对手结果
		for i := 1; i < n; i++ {
			if n%i == 0 && !dfs(n-i, memo) {
				canWin = True
				break
			}
		}

		memo[n] = canWin
		return memo[n] == True
	}

	memo := make([]int, n+1)
	return dfs(n, memo)
}

/*
1140: 石子游戏II

给定一个数字 piles, piles[i] 表示第 i 堆石子的个数.

最初 M = 1. 在每个回合中, 该玩家可以拿走剩下的 "前X堆的所有石子", 其中 1 <= X <= 2M, 然后, M = max(M, X).

游戏一直持续到所有 piles 当中的石子都被拿完. "返回先手拿到最大数量的石子".
**/

func StoneGameII(piles []int) {
	const Max = int(1>>64 - 1)
	var dfs func(arr []int, i, j int, m int, memo [][]int) int

	dfs = func(arr []int, i, j int, m int, memo [][]int) int {
		if i > j {
			return 0
		}

		if memo[i][j] != Max {
			return memo[i][j]
		}

		var (
			one
		)

		for k := 0; k < 2*m; k++ {
			me += arr[k+i]
			other := dfs(arr, i+k+1, j, max(m, k+1), memo)
			if me < other {

			}
		}
	}
}
