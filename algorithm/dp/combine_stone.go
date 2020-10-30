package dp

import (
	"fmt"
	"math"
)

/**
在一个圆形操场的四周摆放N堆石子, 现要将石子有次序地合并成一堆. 规定每次只能选相邻的2堆合并成新的一堆,
并将新的一堆的石子数, 记为该次合并的得分. 试设计出1个算法,计算出将N堆石子合并成1堆的最小得分和最大得分.
**/

/**
f[i][j]意思是从堆i到堆j合并花费的最小值。
合并石子是一个合并类的动态规划问题，f[i][j]合并费用最小，则看是否存在k使得f[i][k]+f[k+1][j]+sum[i][j]更小
(其中sum[i][j]表示石子i堆到j堆得数量之和，它表示把已合并的[i][k]与[k+1][j]两堆再合并的花费)
这样的处理方法，就类似于最短路径问题了，欲求[i][j]之间的最短路，看是否能存在中间连接点k使得[i][k]+[k][j]路程更短的问题
动归方程: f[i][j]=min(f[i][j],(f[i][k]+f[k+1][j])+sum[i][j]), i=<k<j(f[k][k]=0)


环形DP问题，处理技巧
理论上线性序列按照上面的思路去处理合并的问题，必定会得到最优解的。但是面对环形的问题又会产生额外的许多麻烦。
因为1 2 3 1这样的序列最优的办法是先合并1 1，然后再合并2，最后去合并3。但是这是一个线性的序列前面的f[i][j]我们处理的
时候都是认为i<=j的。总不能去求f[4][1]的值吧。
对于这个棘手的环，我们最好的办法是将其打开，使其变成一个序列。将1 2 3 1展开变成1 2 3 1 1 2 3 1 这样的线性序列即可，
那么这个序列中连续4个堆合并最优的花费值，即为所求。弄明白以后问题就很简单了。

**/

func CombineStone(stones []int) int {
	n := len(stones)

	sums := make([][]int, n)
	dp := make([][]int, n)
	for i := 0; i < n; i++ {
		dp[i] = make([]int, n)
		sums[i] = make([]int, n)
	}

	for i := 0; i < n; i++ {
		sum := 0
		for j := i; j < n; j++ {
			sum += stones[j]
			sums[i][j] = sum
		}
	}

	// TODO: 对角线遍历的方法, d + [i,i+d}
	for d := 1; d < n; d++ {
		for i := 0; i < n-d; i++ {
			j := i + d
			dp[i][j] = math.MaxInt32
			for k := i; k < j; k++ {
				dp[i][j] = Min(dp[i][j], dp[i][k]+dp[k+1][j]+sums[i][j])
			}
		}
	}

	fmt.Println(dp)
	fmt.Println(sums)
	return dp[0][n-1]
}

/**

第一种方案是枚举[i...j]之间的k，找出能使得f[i][j]合并费用的最小的k使得f[i][j]=f[i][k]+f[k+1][j]+sum[i][j]取到最小值
换一个角度思考，递推的角度出发，首先肯定是计算序列中合并石子堆数少的时候的值，然后再一步步去计算合并更多堆石子的花费
例如：3 1 2 5 4这五堆石子，选择连续的两堆合并时的花费f[i][i+1]=w[i]+w[i+1]即可，但是合并连续的三堆的时候就麻烦了一些
假如是合并：3 1 2这三堆的时候，不得不去考虑是先合并1 2再合并3，还是先合并3 1再合并2.(f[2][3]和f[1][2]之前都计算出最优值)
事实上当合并3 1 2 5前面四堆石子的时候，最小花费考虑也只要考虑先合并3 1 2再合并5，还是先合并1 2 5再合并3的问题。
于是动归方程：f[i][j]=min(f[i+1][j],f[i][j-1])+sum[i][j]
细心的朋友可能发现第二种方案不就是第一种方案的特殊情况吗？只分析了当k=i和k=j-1的情况(f[i][i]=0),但是这种处理方法
符合动态规划的无后序性和最优性原则，得到的结果和方案一是一致的，事实上第一种方案使得结果最优的k值也肯定不只有一个的

**/
func CombineStone2(stones []int) int {
	n := len(stones)
	dp := make([][]int, n)
	sums := make([][]int, n)

	for i := 0; i < n; i++ {
		dp[i] = make([]int, n)
		sums[i] = make([]int, n)
	}

	for i := 0; i < n; i++ {
		sum := 0
		for j := i; j < n; j++ {
			sum += stones[j]
			sums[i][j] = sum
		}
	}

	for d := 1; d < n; d++ {
		for i := 0; i < n-d; i++ {
			j := i + d
			dp[i][j] = Min(dp[i+1][j], dp[i][j-1]) + sums[i][j]
		}
	}

	fmt.Println(dp)
	fmt.Println(sums)

	return dp[0][n-1]
}

//
func CombineStoneK(stones []int, k int) {

}
