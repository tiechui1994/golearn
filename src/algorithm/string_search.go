package algorithm

/**
@ 字符串查询算法
*/

// BF算法, 暴力查询
func BruteForce(main string, sub string) int {
	if main == "" || sub == "" {
		return -1
	}

	for i := 0; i < len(main); {
		for j := 0; j < len(sub); {
			if main[i] == sub[j] {
				i++
				j++
			} else {
				i = i - (j - 1)
				j = 0
			}

			if j == len(sub)-1 {
				return i - j
			}
		}
	}

	return -1
}

// 文档: https://blog.csdn.net/v_JULY_v/article/details/7041827
//
// KMP算法
//
// BF 算法效率低在于每轮主串从pos位置开始比较过程中, 若字符不相等, 主串会回退到 pos+1 位置, 然后与匹配串第一个位置字符开始比较.
//
// KMP 算法首先算出一个next数组, 匹配串每轮匹配在 j 位置失配时, 匹配串向右滑动的距离为 j - next[j].
// next 数组的计算规则如下. 其中 j 为匹配串索引：
//
// 1. j=0 时, next[j] = -1;
//
// 2. j>0, 匹配串 0 至 j-1 位置的子串前缀和后缀数组中最长相同元素的长度为maxL, 则next[j] = maxL.
// 特殊的, 当j=1时,前缀和后缀都为空, 所以maxL=0, next[1] = 0:
//
func KMP(main string, sub string) int {
	//getNext := func(sub string) []int {
	//	length := len(sub)
	//	next := make([]int, length)
	//	next[0] = -1
	//
	//	// 保存前缀和后缀数组最长相同子串
	//	k := 0
	//	for j := 1; j < length-1; j++ {
	//		for (k > 0) && (sub[j-1] != sub[k]) {
	//			k = next[j-1]
	//		}
	//		if (j > 1) && (sub[j-1] == sub[k]) {
	//			k++
	//		}
	//		next[j] = k
	//	}
	//
	//	return next
	//}

	next := getnext(sub)
	j := 0
	i := 0
	for i < len(main) && j < len(sub) {
		if j == -1 || main[i] == sub[j] {
			i++
			j++
		} else {
			j = next[j]
		}

		if j == len(sub)-1 {
			return i - j
		}
	}

	return -1
}

/**
1. 如果对于值k, 已有p0 p1, ..., pk-1 = pj-k pj-k+1, ..., pj-1, 相当于next[j] = k.
   此意味着什么呢? 究其本质, next[j] = k 代表 p[j] 之前的模式串子串中, 有长度为 k 的相同前缀和后缀.
   有了这个next 数组, 在KMP匹配中, 当模式串中j处的字符失配时, 下一步用next[j]处的字符继续跟文本串匹配, 相当于模式串
   向右移动 j-next[j] 位.

2. 下面的问题是: 已知next [0, ..., j], 如何求出 next[j+1] 呢?
   对于P的前j+1个序列字符:

   若p[k] == p[j], 则next[j+1] = next[j] + 1 = k + 1;
   注意: next[j] = k

   若p[k] ≠ p[j], 如果此时p[next[k]] == p[j], 则next[j+1]=next[k] + 1, 否则继续递归前缀索引k=next[k], 而后
   重复此过程. 

   相当于在字符p[j+1]之前不存在长度为k+1的前缀 "p0 p1, …, pk-1 pk" 跟后缀 "pj-k pj-k+1, …, pj-1 pj" 相等, 那
   么是否可能存在另一个值 t+1 < k+1, 使得长度更小的前缀 "p0 p1, …, pt-1 pt" 等于长度更小的后缀 "pj-t pj-t+1, …,
   pj-1 pj" 呢? 如果存在, 那么这个 t+1 便是 next[j+1] 的值, 此相当于利用已经求得的 next 数组 (next[0,...k,...,j])
   进行P串前缀跟P串后缀的匹配.

   算法内容:
   next[0] = -1

   k = -1
   j = 0
   while j<len(p) :
     if k != -1 && p[k] == p[j] :
        j++
        next[j] = k + 1
        k = next[j] 或者 k = k+1 因为此时j已经发生变化, 则k也需要发生变化(k的初始值就是k=next[j-1]
     else:
        k = next[k]
**/

/**
next改造分析:
  p[j] ≠ p[next[j]]

   问题出在不该出现 p[j] = p[next[j]].为什么呢? 理由是: 当p[j] != s[i] 时, 下次匹配必然是p[next[j]]
   跟s[i]匹配, 如果 p[j] = p[next[j]], 必然导致后一步匹配失败(因为p[j]已经跟s[i]失配, 然后你还用跟p[j]等同的值
   p[next[j]]去跟s[i]匹配, 很显然, 必然失配). 所以不能允许p[j] = p[next[j]].

   如果出现了p[j] = p[next[j]]咋办呢? 如果出现了, 则需要再次递归, 即令next[j] = next[next[j]]。

   算法内容:
   next[0] = -1

   k = -1
   j = 0
   while j<len(p) :
     if k != -1 && p[k] == p[j] :
        j++
        k++
        if p[j] != p[k]:
            next[j] = k
        else:
 			next[j] = next[k]
     else:
        k = next[k]
**/
func getnext(p string) []int {
	length := len(p)
	next := make([]int, length)
	next[0] = -1

	k := -1
	for j := 0; j < length-1; {
		if k == -1 || p[j] == p[k] {
			k++
			j++
			next[j] = k
		} else {
			k = next[k]
		}
	}

	return next
}

// 文档: https://www.cnblogs.com/lanxuezaipiao/p/3452579.html
//
// BM 算法
func BM(text string, pattern string) int {
	// 计算坏字符数组 bmbad
	bmbad := func() []int {
		var arr [256]int
		var length = len(pattern)
		for i := 0; i < 256; i++ {
			arr[i] = length
		}
		for i := 0; i < length-1; i++ {
			arr[pattern[i]] = length - 1 - i
		}
		return arr[:]
	}()

	//

	return -1
}
