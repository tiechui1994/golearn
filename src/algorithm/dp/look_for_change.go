package dp

import "math"

// 假设有几种硬币, 如1, 3, 5, 并且数量无限. 请找出能够组成某个数目的找零所使用最少的硬币数.

// 记作dp[k]，其值为所需的最小硬币数.
// 对于不同的硬币面值 coin[0...n]
// dp[k] = min(dp[k-coin[0]], dp[k-coin[1]], ...)+1. 对于数目k, 最少的方案
// 对应于给定数目的找零total，需要求解sum[total]的值.

func CoinChange1(coins []int, total int) int {
	if total == 0 {
		return -1
	}

	dp := make([]int, total+1)
	for i := range dp {
		dp[i] = total + 1
	}

	for i := 1; i <= total; i++ {
		for _, coin := range coins {
			if coin <= i {
				dp[i] = Min(dp[i], dp[i-coin]+1)
			}
		}
	}

	if dp[total] == total+1 {
		return -1
	}

	return dp[total]
}

// 备忘录版本
func CoinChange2(coin []int, total int) int {
	if total == 0 {
		return -1
	}

	dp := make([]int, total+1)
	for i := 0; i < len(dp); i++ {
		dp[i] = math.MaxInt32
	}

	return helper(coin, total, dp)
}

func helper(coins []int, total int, dp []int) int {
	if total == 0 {
		return 0
	}
	if dp[total] != math.MaxInt32 {
		return dp[total]
	}

	var ans = int(math.MaxInt32)
	for _, v := range coins {
		// 不可达
		if total-v < 0 {
			continue
		}

		subans := helper(coins, total-v, dp)

		if subans != -1 {
			ans = Min(subans+1, ans)
		}
	}

	if ans == int(math.MaxInt32) {
		dp[total] = -1
	} else {
		dp[total] = ans
	}

	return dp[total]
}
