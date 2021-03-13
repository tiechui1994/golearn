package category

import (
	"sort"
)

/*
DFS=Depth First Search

深度优先遍历思想: 以一个未被访问的节点作为起始节点, 沿当定点的边走到未访问过的节点; 当没有未访问的节点时, 则回退到上一个
节点, 继续访问别的节点.

场景:

1. 模板 dfs
2. 外部空间 dfs(用 stack, eg: tree的遍历)
3. dfs+memo
4. 模拟流程, 寻找全排列解
*/

func Inorder(root *TreeNode) []int {
	var ans []int

	node := root
	stack := make([]*TreeNode, 0)

	for len(stack) > 0 || node != nil {
		for node != nil {
			stack = append(stack, node) // 根节点入栈
			node = node.left            // 转向左孩子
		}

		// 出栈
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// 访问, 回退到右孩子
		ans = append(ans, cur.val)
		node = cur.right
	}

	return ans
}

/*
78.子集(给定的nums是互不相同的)

思路一: 二进制表示 ( 0 - 2^n-1 ), 其中 n 是 nums 的长度. 使用二进制进行表示, 每次取1位置对应的nums 作为元素

思路二: 使用 dfs(cur, index), cur 表示当前取的元素, index是当前取到的位置, 后面从 index 开始, 每次添加一个元素到
cur数组当中, 遍历结束后, 将该元素删除掉.

子集II(给定的nums存在重复的)

思路一: 首先进行一个数组排序, 这样相同的元素会在相邻的位置. 然后按照上述的思路二, 只是在添加相同元素的时候, 只是添加第一
个, 后面的不在添加.

*/
func SubSets(nums []int) [][]int {
	var dfs func(tmp []int, index int, ans *[][]int)
	dfs = func(tmp []int, index int, ans *[][]int) {
		*ans = append(*ans, tmp[:]) // 将当前的 tmp 加入到结果

		for i := index; i < len(nums); i++ {
			// 这样保证对于重复的元素只是添加一次
			if i != index && nums[i] == nums[index] {
				continue
			}

			tmp = append(tmp, nums[i]) // 往 tmp 当中添加一个元素
			dfs(tmp, i+1, ans)         // 深度遍历
			tmp = tmp[:len(tmp)-1]     // 后退一步
		}
	}

	ans := make([][]int, 0)
	dfs([]int{}, 0, &ans)
	return ans
}

/*
46: 全排列
*/

func Permute(nums []int) [][]int {
	sort.SliceStable(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})
	var dfs func(cur []int, ans *[][]int, used []bool)
	dfs = func(cur []int, ans *[][]int, used []bool) {
		if len(cur) == len(nums) {
			*ans = append(*ans, cur[:])
		} else {
			for i := 0; i < len(nums); i++ {
				// 查重, 一个数字只能使用一次.
				// 如果当前的数字和前一个数字一样,并且前一个数字没有被访问到, 则跳过.
				// 保证只有一次前一个数字+当前数字数字(如果当前数字和前一个数字一致的话)
				// 代码逻辑有待商榷
				if used[i] || i > 0 && nums[i] == nums[i-1] && !used[i-1] {
					continue
				}

				used[i] = true
				cur = append(cur, nums[i])
				dfs(cur, ans, used)
				cur = cur[:len(cur)-1]
				used[i] = false
			}
		}
	}

	ans := make([][]int, 0)
	dfs([]int{}, &ans, make([]bool, len(nums)))
	return ans
}

/*
77: 组合
*/

/*
37: 猜数字

每一行必须是1-9
每一列必须是1-9
每个3x3的矩阵也必须是1-9

算法: 暴力破解. 针对每一个需要填充的数字, 尝试'1'-'9', 然后针对选择的数字进行合法化判断. 如果合法, 则继续下一次dfs,
否则, 进行回退, 继续尝试下一个数字. 当尝试完了所有的数字, 如果还是没有解, 则直接返回false

51: N-皇后问题

同一行, 同一列, 同一斜, 不能出现两个相同的 queue

算法: 玻璃破解. 和猜数字是类似的
*/
