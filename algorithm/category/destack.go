package category

/*
单调栈:

1. 维持递增/递减 [栈顶 -> 栈底]
2. 获取栈顶元素, 进行计算操作
3. 当前元素入栈
*/

/*
模板: 输入一个数组, 返回一个数组. 对于输入每个数组位置i, 找到在i之后首个元素, 作为返回数组位置i的值, 如果不存在, 则
当前的位置值为 -1
*/
func IncreseStacktempate(nums []int) []int {
	n := len(nums) - 1
	ans := make([]int, n)

	stack := Stack{}
	for i := n - 1; i >= 0; i++ {
		// 单调 [栈顶 -> 栈底] 递增, 则栈顶是元素最小.
		for !stack.isEmpty() && nums[i] >= stack.peek() {
			stack.pop()
		}
		ans[i] = -1
		if !stack.isEmpty() {
			ans[i] = stack.peek()
		}

		stack.push(nums[i])
	}

	return ans
}

/*
503. 下一个更大元素II

给定一个循环数组(最后一个元素的下一个元素是数组的第一个元素), 输出每个元素的下一个更大元素. 数字 x 的下一个更大的元素是
按数组遍历顺序, 这个数字之后的第一个比它更大的数, 这意味着你应该循环地搜索它的下一个更大的数. 如果不存在, 则输出 -1.

思路: 将原来的数组复制两份, 然后使用模板提供的方法进行计算. 简化方法是 idx % n 计算出真实数组当中的位置.


739: 时间温度.

思路: 使用存储 index 代替存储数字.
*/

/*
316. 去除重复字母

给你一个字符串 s, 去除字符串中重复的字母, 使得每个字母只出现一次. 需保证返回结果的字典序最小(要求不能打乱其他字符的相对
位置).

思路:
	计算字母最后一次出现的位置.
	单调递减栈. 维持栈递减的条件 (当前的元素不是最后一次出现 && 当前元素非递减) => 元素出栈
*/

func RemoveDuplicateLetters(s string) string {
	visit := make(map[int]bool)
	stack := Stack{}

	last := make([]int, 128)
	for i := 0; i < len(s); i++ {
		last[int(s[i])] = i
	}

	for i := 0; i < len(s); i++ {
		char := int(s[i])

		// 统计当前加入栈的元素. 保证栈中的元素唯一.
		if visit[char] {
			continue
		} else {
			visit[char] = true
		}

		// 递减顺序, 栈顶元素最大
		for !stack.isEmpty() && i < last[char] && char < stack.peek() {
			delete(visit, stack.pop())
		}

		stack.push(char)
	}

	n := stack.size()
	ans := make([]rune, n)
	for !stack.isEmpty() {
		ans[n-1] = rune(stack.pop())
		n -= 1
	}

	return string(ans)
}

/*
402. 移掉K位数字[Mid]

给定一个以字符串表示的非负整数 num, 移除这个数中的 k 位数字, 使得剩下的数字最小.

思路: 单调递减栈[可以元素相等]. 这样使得小数字在前面(栈底), 大数字在栈尾巴
*/

func RemoveKDights(num string, k int) string {
	n := len(num)
	stack := Stack{}

	for i := 0; i < n; i++ {
		char := int(num[i])
		for !stack.isEmpty() && k > 0 && char < stack.peek() {
			stack.pop()
			k--
		}

		stack.push(char)
	}

	for k > 0 {
		stack.pop()
		k--
	}

	zero := 0
	n = stack.size()
	ans := make([]rune, n)
	for !stack.isEmpty() {
		char := stack.pop()
		ans[n-1] = rune(char)
		n--
		if char == '0' {
			zero++
		} else {
			zero = 0
		}
	}

	return string(ans[zero:])
}

/*
42. 接雨水

给定 n 个非负整数表示每个宽度为 1 的柱子的高度图, 计算按此排列的柱子, 下雨之后能接多少雨水.

84: 条形图最大长方形面积.

思路: 当图形处于上升期(h[i] < h[i+]), 不用计算面积. 因为这种状况下, 再往后移动一格(i -> i+1), 将获得更大的面积.
     当图形处于下降期(h[i] > h[i+1]), 需要计算当前矩形的面积, 这个时候穷举 stack 里所有的高度, 由于 stack 是递增
的, 从最高的高度开始不断下降, 随着高度的下降, 更多的竖条可以加入到大长方形面积来, 保持所生成的大长方形的最大面积.
 */