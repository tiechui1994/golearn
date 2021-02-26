package category

import (
	"strconv"
	"strings"
)

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
	const Max = int(1<<63 - 1)

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
	const Max = int(1<<63 - 1)

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

	N := len(piles)
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

//==================================================================================================

/*
动态规划二: 单序列类

基础: 斐波那契数列
*/

/**
91: 解码方法

'A' -> 1
'B' -> 2
...
'Z' -> 26

已知编码消息, 求解解码的方法总数.

dp[i] = dp[i-1] + dp[i-2]
*/

func NumDecodeings(s string) int {
	if len(s) == 0 {
		return 0
	}

	dp := make([]int, len(s)+1)

	dp[0] = 1
	if s[0] != '0' {
		dp[1] = 1
	}

	for i := 2; i <= len(s); i++ {
		two, _ := strconv.ParseInt(s[i-2:i], 10, 32)
		one, _ := strconv.ParseInt(s[i-1:i], 10, 32)
		if 10 <= two && two <= 26 {
			dp[i] += dp[i-2]
		}
		if one != 0 {
			dp[i] += dp[i-1]
		}
	}

	return dp[len(s)]
}

func NumDecodeingsI(s string) int {
	var dfs func(i int, memo []int) int

	dfs = func(i int, memo []int) int {
		if i == 0 {
			return 1
		}

		if memo[i] > 0 {
			return memo[i]
		}

		if s[i] == '0' {
			memo[i] = 0
		} else if i+1 < len(s) && (s[i] == '1' || s[i] == '2' && s[i+1] <= '6') {
			memo[i] = dfs(i+1, memo) + dfs(i+2, memo)
		} else {
			memo[i] = dfs(i+1, memo)
		}

		return memo[i]
	}

	memo := make([]int, len(s))
	return dfs(0, memo)
}

/*
139: 单词拆分.

给定一个非空字符串s和一个包含非空单词的列表 wordDict, 判断s是否可以被空格拆分为一个或多个在字典都出现的单词.

注: 拆分后的单词列表当中包含所有的 wordDict 当中的单词.

dp[j] = dp[i] && dict.has(s[i:j])  0 <= i < j

dp[0] = true 因为 s[0:0] 是空字符串. 最终看 dp[N].
*/

func WordBreak(s string, wordDict []string) bool {
	N := len(s)
	dict := make(map[string]bool)
	for _, w := range wordDict {
		dict[w] = true
	}

	dp := make([]bool, N+1)
	dp[0] = true

	for i := 1; i <= N; i++ {
		for j := 0; j < i; j++ {
			if dp[j] && dict[s[j:i]] {
				dp[i] = true
			}
		}
	}

	return dp[N]
}

/*
140: 单词拆分II.

给定一个非空字符串s和一个包含非空单词的列表 wordDict. 在字符串中增加空格来构建一个句子, 使得句子中所有的单词都在词典当
中, 返回所有可能的句子.

dp[i] = { j }; 其中 j 满足 dp[j].length > 0 && 0 <= j < i && s[j:i] 存在于 wordDict 当中

这样组成一个二维数组 dp[][]

从最后 dp[N] 开始倒退计算, 当前 cur 为"", 遍历 dp[N], s[k:N] + " " + cur 便是当前的 cur,
继续上述的过程, 直到 N 变为 0, 遍历结束, 此时的 cur 就是一个截断结果.
*/

func WordBreakII(s string, wordDict []string) []string {
	N := len(s)
	dict := make(map[string]bool)
	for _, w := range wordDict {
		dict[w] = true
	}

	var dp = make([][]int, N+1)
	for i := 0; i <= N; i++ {
		dp[i] = []int{}
	}

	dp[0] = append(dp[0], 0)
	for i := 1; i <= N; i++ {
		for j := 0; j < i; j++ {
			if len(dp[j]) > 0 && dict[s[j:i]] {
				dp[i] = append(dp[i], j)
			}
		}
	}

	var ans []string
	var result func(cur string, index int, str string)
	result = func(cur string, index int, str string) {
		if index == 0 {
			ans = append(ans, strings.TrimSpace(cur))
			return
		}

		for _, idx := range dp[index] {
			result(str[idx:index]+" "+cur, idx, str)
		}
	}

	result("", N, s)

	return ans
}

/*
53: 最大子序(Sequence)和

dp[i] = max(dp[i-1]+nums[i], nums[i])
*/

/*
152: 最大子序(Sequence)乘积

dpmax[i] = max(dp[i-1]*nums[i], nums[i])

dpmin[i] = min(dp[i-1]*nums[i], nums[i])
*/

func MaxProduct(nums []int) int {
	N := len(nums)
	dpmax := make([]int, N)
	dpmin := make([]int, N)

	ans := nums[0]
	dpmax[0] = nums[0]
	dpmin[0] = nums[0]

	for i := 1; i < N; i++ {
		if nums[i] > 0 {
			dpmax[i] = max(nums[i], dpmax[i-1]*nums[i])
			dpmin[i] = min(nums[i], dpmin[i-1]*nums[i])
		} else {
			dpmax[i] = max(nums[i], dpmin[i-1]*nums[i])
			dpmin[i] = min(nums[i], dpmax[i-1]*nums[i])
		}

		ans = max(dpmax[i], ans)
	}

	return ans
}

/*
198. 打家劫舍

dp[i] = max(dp[i-1], dp[i-2]+nums[i]) 表示第 i 天最大的收益. 分别表示第i天不抢和抢
*/

/*
300: LIS (最长递增子序列) Longest Increasing Subsequence

dp[i] = max(dp[i], dp[k]+1) 其中 k 满足的条件 0 <= k < i && nums[k] < nums[i]
表示以 i 结尾的最长递增子序列
*/

// O(n^2)
func LengthLIS(nums []int) int {
	N := len(nums)
	dp := make([]int, N)
	for i := range dp {
		dp[i] = 1
	}

	var ans = 1
	for i := 1; i < N; i++ {
		for j := 0; j < i; j++ {
			if nums[i] > nums[j] {
				dp[i] = max(dp[i], dp[j]+1)
			}
		}

		ans = max(ans, dp[i])
	}

	return ans
}

// O(nlgn)
func LengthLIS2(nums []int) int {
	binarySearch := func(arr []int, size int, target int) int {
		left, right := 0, size-1
		for left <= right {
			mid := (left + right) / 2
			if arr[mid] < target {
				left = mid + 1
			} else {
				right = mid - 1
			}
		}
		return left
	}

	size, N := 0, len(nums)
	dp := make([]int, N)
	for i := 0; i < N; i++ {
		pos := binarySearch(dp, size, nums[i])
		dp[pos] = nums[i]

		if pos == size {
			size++
		}
	}

	return size
}

/*
1713. 得到子序列(Subsequence)的最少操作次数

有两个数组, target和nums, 其中 target 包含若干个"互不相同"的整数, nums可能包含重复元素.

每一次操作中, 可以在 nums 的任意位置插入一个整数. 返回最少操作次数, 使得 target 成为 nums 的一个子序列(Subsequence).

注: 数组的子序列指的是是删除数组的某些元素(也可能一个元素都不删除), 同时不改变其余元素的相对顺序得到的数组. 例如:
[2,7,4] 是 [4, "2", 3, "7", 2, 1, "4"] 的一个子序列, 但是 [2, 4, 2] 则不是
*/

func MinOperations(target []int, nums []int) int {
	dict := make(map[int]int) // target<->index
	for i, val := range target {
		dict[val] = i
	}

	indexs := make([]int, 0, len(nums))
	for _, val := range nums {
		if idx, ok := dict[val]; ok {
			indexs = append(indexs, idx)
		}
	}

	return len(target) - LengthLIS2(indexs)
}

//==================================================================================================

/*
动态规划三: 区间DP

区间类DP是单序列类DP的扩展, 它在分阶段地划分问题时, 与阶段中元素的顺序和由前一阶段的哪些元素合并而来有很大的关系. 令状态
f(i, j) 表示将下标位置 i 到 j 的所有元素合并能获得的价值的最大值, 那么 f(i,j) = max(f(i,k)+f(k+1,j)+cost}

区间DP特点:

- 合并: 即将两个或多个部分进行整合, 也可能是反过来
- 特征: 能将问题分解为能两两合并的形式
- 求解: 整个问题的最优值, 枚举合并点, 将问题分解为左右部分, 最后合并两个部分的最优值得到原问题的最优值
*/

/*
1)合并石头[链接](最小成本)

dp(i, j) = min( dp(i, k) + dp(k+1, j) + sum[i..j] )  i < k < j


2)合并石头[环](最小成本) => nums[0..N] + nums[0..N] 组成一堆链, 计算合并石头最小代价

dp(i, j) = min( dp(i, k) + dp(k+1, j) + sum[i..j] )  i < k < j

min( dp(i, i+N) ) 则是最终的代价


3) 矩阵乘法最小成本:
A[i] = nums[i]*nums[i+1]

dp(i, j) = min( dp(i, k) + dp(k+1, j) + nums[i] * nums[k+1] * nums[j+1])
*/

/*
1039. 多边形三角剖分的最低得分

dp[i][j] 表示 i..j 节点切割的最小成本

dp[i][j] = min( dp[i][k] + dp[k][j] + nums[i]*nums[k]*nums[j])  i < k < j

注: i+2 > j 这个时候无法切割.
*/

/*
1547. 切棍子的最小成本

有一根长度为 n 的棍子, 需要按照 cuts 当中的节点进行切割. 每次切割的成本是当前棍子的长度. 求解切割棍子的最小成本.

dp(i,j) 从 [i..j] 切割开的最小成本. (当前的长度就是 j-i)

dp(i,j) = min( dp(i,k) + dp(k,j) + j-i ) 其中 k 是切割点.

dp(0, n) 则是最终的结果.
*/

func MinCutCost(n int, cuts []int) int {
	var dfs func(i, j int, memo map[[2]int]int) int

	dfs = func(i, j int, memo map[[2]int]int) int {
		if i+1 == j {
			return 0
		}

		if val, ok := memo[[2]int{i, j}]; ok {
			return val
		}

		var find bool
		var res = int(1<<63 - 1)
		for _, cut := range cuts {
			if i < cut && cut < j {
				find = true
				res = min(dfs(i, cut, memo)+dfs(cut, j, memo)+j-i, res)
			}
		}

		if !find {
			return 0
		}

		memo[[...]int{i, j}] = res
		return res
	}

	return dfs(0, n, make(map[[2]int]int))
}

/*
312. 戳气球

有 n 个气球, 编号为0 到 n - 1, 每个气球上都标有一个数字, 这些数字存在数组 nums 中.

现在要求戳破所有的气球. 戳破第 i 个气球, 可以获得 nums[i-1] * nums[i] * nums[i+1] 积分. 如果 i-1 或 i+1 超出了
数组的边界, 那么就当它是一个数字为 1 的气球.

dp[i][j] 表示 [i..j] 之间获得的积分

dp[i][j] = max( dp[i][k-1] + dp[k+1][j] + nums[i-1]*nums[k]*nums[j+1] ) 其中 i <= k <= j, 代表戳中[i..j]
之间的一个气球.
*/

func MaxCoins(nums []int) int {
	get := func(i int) int {
		if i < 0 || i >= len(nums) {
			return 1
		}
		return nums[i]
	}

	var dfs func(i, j int, memo [][]int) int
	dfs = func(i, j int, memo [][]int) int {
		if i > j {
			return 0
		}
		if memo[i][j] != 0 {
			return memo[i][j]
		}

		var ans int
		for k := i; k <= j; k++ {
			left := dfs(i, k-1, memo)
			right := dfs(k+1, j, memo)
			ans = max(ans, left+right+get(i-1)*get(k)*get(j+1))
		}

		memo[i][j] = ans
		return ans
	}

	N := len(nums)
	memo := make([][]int, N)
	for i := 0; i < N; i++ {
		memo[i] = make([]int, N)
	}
	return dfs(0, N-1, memo)
}

/*
813. 最大平均值和的分组

将一个数组切割成 K 个子数组, 求解子数组平均值和的最大值.

dp[n][k] 表示长度为 n 的数组切割成 k 个部分, 最大的平均和
dp[n][k] = max(dp[i][k-1] + avg),  0 <= i < n
*/

// LSA
func LargestSumOfAverages(nums []int, K int) float64 {
	N := len(nums)
	dp := make([][]float64, N+1)
	for i := 0; i <= N; i++ {
		dp[i] = make([]float64, K+1)
	}

	// k=1 的状况
	sum := 0
	for i := 0; i < N; i++ {
		sum += nums[i]
		dp[i+1][1] = float64(sum) / float64(i+1)
	}

	max := func(i, j float64) float64 {
		if i > j {
			return i
		}
		return j
	}

	// dp[n][k] 表示长度为 n 的数组切割成 k 个部分, 最大的平均和
	// dp[n][k] = max(dp[i][k-1] + avg),  0 <= i < n

	var dfs func(n, k int, memo [][]float64) float64
	dfs = func(n, k int, memo [][]float64) float64 {
		if n < k {
			return 0
		}
		if memo[n][k] != 0 {
			return memo[n][k]
		}

		var ans float64
		var sum int
		for i := n - 1; i >= 0; i-- {
			sum += nums[i]
			ans = max(ans, dfs(i, k-1, memo)+float64(sum)/float64(n-i))
		}

		memo[n][k] = ans

		return ans
	}

	return dfs(N, K, dp)
}

/*
516: 最长回文子序列(Sequence) [***]

dp[i][j] = max(dp[i+1][j], dp[i][j-1]), s[i] != s[j]

dp[i][j] = max(dp[i][j], dp[i+1][j-1] + 2), s[i] == s[j]
*/
func LongestPalindromeSubseq(s string) int {
	N := len(s)
	dp := make([][]int, N)
	for i := 0; i < N; i++ {
		dp[i] = make([]int, N)
		dp[i][i] = 1
	}

	for j := 1; j < N; j++ {
		for i := 0; i < j; i++ {
			if s[i] == s[j] {
				dp[i][j] = max(dp[i][j], dp[i+1][j-1]+2)
			} else {
				dp[i][j] = max(dp[i+1][j], dp[i][j-1])
			}
		}
	}

	return dp[0][N-1]
}

/*
删除回文子序列: 给定一个字符串, 每次可以删除一个回文子序列(SubSequence), 最少的删除次数

dp[i][j] = min( dp[i][k]+dp[k+1][j] ) s[i] != s[j] // 无法同时删除 s[i], s[j]

dp[i][j] = min( dp[i][j], dp[i+1][j-1]) s[i]==s[j] // 注: 这里不需要加1, 原因是删除倒数第2次就可以一次性删除


木板染色:

dp[i][j] 表示将 i..j 染成目标颜色最小次数

dp[i][j] = min( dp[i][k] + dp[k+1][j] ) // s[i] != s[j]

dp[i][j] = min( dp[i][j], dp[i][j-1], dp[l+1][r] ) // s[i] == s[j]



恐狼先锋:

攻击一头狼, 会受到该头狼的伤害和该头狼前后的余波伤害. nums 是每头狼的伤害, buf 是每头狼的余波伤害.

dp[i][j] 表示 i..j 之间被消灭掉的伤害最小值.

dp[i][j] = min( dp[i][k-1] + dp[k+1][j] + nums[k]+buf[i-1]+buf[j+1]) // 先干掉 [i..k-1] 狼, 然后干掉
[k+1..j] 狼, 最后干掉第 k 头狼, 所产生的最小值.
*/

func MinDamage(nums []int, buf []int) int {
	N := len(nums)
	nums = append([]int{0}, append(nums, 0)...)
	buf = append([]int{0}, append(buf, 0)...)

	dp := make([][]int, N+1)
	for i := 0; i <= N; i++ {
		dp[i] = make([]int, N+1)
	}

	var dfs func(i, j int, dp [][]int) int
	dfs = func(i, j int, dp [][]int) int {
		if i > j {
			return 0
		}

		if dp[i][j] != 0 {
			return dp[i][j]
		}

		var ans = int(1<<63 - 1)
		for k := i; k <= j; k++ {
			ans = min(ans, dfs(i, k-1, dp)+dfs(k+1, j, dp)+nums[k]+buf[i-1]+buf[j+1])
		}

		dp[i][j] = ans

		return dp[i][j]
	}

	return dfs(1, N, dp)
}


//==================================================================================================

/*
动态规划: 背包问题

0-1 背包

N 个物品, 第 i 个的代价是 cost[i], 价值是 value[i], 每种物品只有一个. 求解在大小为 V 的背包能获得的最大权益.

dp[i][v] = max( dp[i-1][j], dp[i-1][j-cost[i]]+value[i] ) // (不取, 取)第i个物品


完全背包

N 个物品, 第 i 个的代价是 cost[i], 价值是 value[i], 每种物品有无穷多个. 求解在大小为 V 的背包能获得的最大权益.

dp[i][j] 表示考虑前 i 种物品, 背包大小为j 获取的最大值.

dp[i][j] = max( dp[i][j], dp[i-1][j-n*cost[k]]+n*value[k] ) // 考虑第 i 个产品取了 k 个

dp[i][j] = max( dp[i-1][j], dp[i][j-cost[i]]+value[i] ) // 要不要获取一个第 i 个产品

多重背包:

N 个物品, 第 i 个的代价是 cost[i], 价值是 value[i], 每种物品 cnt[i]个. 求解在大小为 V 的背包能获得的最大权益.


分组背包:

N 个物品, 第 i 个的代价是 cost[i], 价值是 value[i], 每种物品1个. 这N个物品被分为 M 组, 每组最多取一个. 求解在大小
为 V 的背包能获得的最大权益.

dp[i][j] 表示考虑前 i 组物品. 背包大小为 j 的最大价值.

dp[i][j] = max( dp[i-1][j], dp[i-1][j-cost[k]]+value[k] ) // 不选 或 选择第 k 个.
*/