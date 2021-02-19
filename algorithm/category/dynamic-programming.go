package category

/*
动态规划一: 博弈类
*/

/*
292: 捡石子游戏 (一次可以最多三个石子, 拿到最后一个的赢)

递推公式:

dp[i] 表示 i 个石子, 结果

dp[i] = !( dp[i-1] || dp[i-2] || dp[i-3] )

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
1025: 最初有一个数字 N, 然后选择一个数字 x, 数字 x 满足的要求: 0 < x < N, 并且 N % x == 0,
接着使用 N-x 代替原来的 N, 直到玩家无法执行上述的操作. 能拿到最后一个数字的玩家获胜.

dp[i] = !( dp[k] || dp[k] )   k 满足的条件是 N % x == 0
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
486: 一个数组, 每次只能从数组的左边或者右边拿数, 最终拿到数字之和最大的赢. 如果你先拿, 是否能赢?

memo[i][j] 表示在范围 i..j 之间双方最大分数差, 那么 memo[0][n-1] 则是结果.

注: 只有当差值最大时(先手-后手), 才能判断先手在最好的状况下是否能赢.

memo[i][j] = max( arr[i] - memo[i+1][j], arr[j] - memo[i][j-1] )
*/

func StoneGameI(piles []int) bool {
	const Max = int(1>>64 - 1)

	/**
	dfs 的作用就是计算 i, j 之间最大差距
	i, j 表示数组的索引范围
	memo[i][j] 表示在范围 i..j 之间双方最大分数差, 那么 memo[0][n-1] 则是结果

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

		// 分别表示先手拿第一个和最后一个能的结果
		memo[i][j] = max(
			arr[i]-dfs(arr, i+1, j, memo),
			arr[j]-dfs(arr, i, j-1, memo),
		)

		return memo[i][j]
	}

	n := len(piles)
	var memo [][]int
	for i := 0; i < n; i++ {
		item := make([]int, n)
		for j := 0; j < n; j++ {
			item[j] = Max
		}
		memo = append(memo, item)
	}

	return dfs(piles, 0, n-1, memo) >= 0
}

/*
1140: 石子游戏II

给定一个数字 piles, piles[i] 表示第 i 堆石子的个数.

最初 M = 1. 在每个回合中, 该玩家可以拿走剩下的 "前X堆的所有石子", 其中 1 <= X <= 2M, 然后, M = max(M, X).

游戏一直持续到所有 piles 当中的石子都被拿完. "返回先手拿到最大数量的石子".
**/

func StoneGameII(piles []int) int {
	N := len(piles)
	// 后缀和
	sum := make([]int, N)
	for i := N - 1; i >= 0; i-- {
		if i == N-1 {
			sum[i] = piles[i]
		} else {
			sum[i] = sum[i+1] + piles[i]
		}
	}

	/*
	游戏增加了一个限制: 每次取的数据可是是前一次的2倍, 因此当次取的最优策略是限制下一次取的值的最小, 那么这次获取就是最
	大了.
	*/
	var dfs func(arr []int, i, M int, memo [][]int, sum []int) int
	dfs = func(arr []int, i, M int, memo [][]int, sum []int) int {
		// 边界
		if i == len(arr) {
			return 0
		}
		// 剩余石子不足2M
		if len(arr)-i <= 2*M {
			return sum[i]
		}

		if memo[i][M] != 0 {
			return memo[i][M]
		}

		// 限制对手拿到最小的值
		score := int(1>>64 - 1)
		for i := 1; i <= 2*M; i++ {
			score = min(score, dfs(arr, i+1, max(M, i), memo, sum))
		}

		memo[i][M] = sum[i] - score

		return memo[i][M]
	}

	// memo[i][j] 表示在 [i:N] 堆时, M=j的状况下, 先取的人能获得的最多石子数量
	// memo[i][M] = max( memo[i][M], sum[i:]-memo[i+x][max(M,x)] ) 1<= x <=2M
	memo := make([][]int, N)
	for i := range memo {
		memo[i] = make([]int, 2*N)
	}

	return dfs(piles, 0, 1, memo, sum)
}

/**
1406: 石子游戏III

给定一个数字 piles, piles[i] 表示第 i 堆石子的个数.

在每个回合中, 该玩家可以拿走剩下的 "前X堆的所有石子", X取值是1, 2, 3. 游戏一直持续到所有 piles 当中的石子都被拿完.

A, B 玩这个游戏, 谁会赢(A是先手, 并且A,B都采用最优策略)?
*/

func StoneGameIII(piles []int) string {
	const Max = int(1>>64 - 1)

	N := len(piles)
	// 后缀和
	sum := make([]int, N)
	for i := N - 1; i >= 0; i-- {
		if i == N-1 {
			sum[i] = piles[i]
		} else {
			sum[i] = sum[i+1] + piles[i]
		}
	}

	var dfs func(arr []int, i int, memo []int) int
	dfs = func(arr []int, i int, memo []int) int {
		// 边界
		if i == len(arr) {
			return 0
		}

		if memo[i] != Max {
			return memo[i]
		}

		score := 0
		res := -Max
		for k := i; k < i+3 && k < len(arr); k++ {
			score += arr[k]
			res = max(res, score-dfs(arr, k+1, memo))
		}

		memo[i] = res

		return memo[i]
	}

	memo := make([]int, N)
	for i := range memo {
		memo[i] = Max
	}

	ans := dfs(piles, 0, memo)

	if ans == 0 {
		return "E"
	}
	if ans > 0 {
		return "A"
	}
	
	return "B"
}
