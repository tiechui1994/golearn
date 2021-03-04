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
	统计字母最后一次出现的位置.
	单调递减栈. 维持栈递减的条件 (当前的元素不是最后一次出现) => 元素出栈
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

		// 统计加入栈的元素
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
