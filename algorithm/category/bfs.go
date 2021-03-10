package category

/*
BFS=Breadth First Search(广度优先搜索)

一般用来

1. 遍历树结构(level order)

2. 遍历图结构(BFS, Toplogical)

3.遍历二维数组(扩散思维)

BFS 最擅长搜索最短路径的解是比较合适的.

1. 比如最少步数, 最少交换次数的解

2. BFS是空间效率高. 一般是使用一个队列记录搜索的中间过程

3. BFS适合搜索全部解. 

4. 双向BFS优化搜索的速度.
*/

/*
102. 二叉树的层序遍历(分层, 每一层一个数组), 使用 Queue

变种: 奇偶层顺序顺序不一样(奇数层正序, 偶数层逆序)
	 树的左/右视图(思路: 当遍历当前层的第一个或最后一个的时候), 下面的 LevelOrder 代码
     树绕最外层圈打印(思路: 左视图+右视图 => 结果)
     树纵向打印(思路: DFS[node, height, x], 其中 height 表示哪一行, x 表示当前行偏离根节点的正负偏移量)
*/

type TreeNode struct {
	val   int
	left  *TreeNode
	right *TreeNode
}

func LevelOrder(root *TreeNode) {
	queue := make([]*TreeNode, 0)
	if root != nil {
		queue = append(queue, root)
	}

	// 左视图
	var leftvist []int

	for len(queue) > 0 {
		// 开始遍历一层
		size := len(queue)
		for i := 0; i < size; i++ {
			cur := queue[0]
			queue = queue[1:]

			if i == 0 {
				leftvist = append(leftvist, cur.val)
			}

			if cur.left != nil {
				queue = append(queue, cur.left)
			}
			if cur.right != nil {
				queue = append(queue, cur.right)
			}
		}

		// 一层遍历结束
	}
}

/*
127. 单词接龙

字典wordList中从单词beginWord和endWord的"转换序列"是一个按下述规格形成的序列:

- 序列中第一个单词是 beginWord.
- 序列中最后一个单词是 endWord.
- 每次转换只能改变一个字母.
- 转换过程中的中间单词必须是字典wordList中的单词.

给你两个单词beginWord和endWord和一个字典wordList, 找到从beginWord到endWord的 "最短转换序列" 中的 "单词数目".
如果不存在这样的转换序列,返回 0.

思路: 暴力BFS, 每一步, 修改当前单词的每个位置的字母, 进行搜索该单词在单词列表当中(如果存在, 则将该单词从单词列表当中移除,
同时将该单词加入到queue当中).

双向BFS:

传统 BFS 是从起点开始向四周扩散, 遇到终点停止;
双向 BFS 是从起点和终点同时扩散, 当两边有交集停止;(每次选择较小的集合进行扩展遍历)

*/

func LadderLength(begin string, end string, wordList []string) int {
	set := make(map[string]bool)
	for _, word := range wordList {
		set[word] = true
	}

	step := 1
	queue := []string{begin}
	for len(queue) > 0 {
		size := len(queue)
		for i := 0; i < size; i++ {
			cur := queue[0]
			queue = queue[1:]

			if cur == end {
				return step
			}

			N := len(cur)
			for j := 0; j < N; j++ {
				buf := []byte(cur)
				for k := byte('a'); k <= 'z'; k++ {
					if cur[j] == k {
						continue
					}

					buf[j] = k
					next := string(buf)
					if set[next] {
						if next == end {
							return step + 1
						}
						delete(set, next)
						queue = append(queue, next)
					}
				}
			}
		}

		step += 1
	}

	return 0
}

func LadderLength2(begin string, end string, wordList []string) int {
	set := make(map[string]bool)
	for _, word := range wordList {
		set[word] = true
	}

	beginSet, endSet := make(map[string]bool), make(map[string]bool)
	visited := make(map[string]bool)

	step := 1
	beginSet[begin] = true
	endSet[end] = true

	for len(beginSet) > 0 && len(endSet) > 0 {
		nextSet := make(map[string]bool)

		// 展开 beginset
		for cur := range beginSet {
			for j := 0; j < len(cur); j++ {
				buf := []byte(cur)
				for k := byte('a'); k <= 'z'; k++ {
					if cur[j] == k {
						continue
					}

					buf[j] = k
					next := string(buf)
					// 相遇, 有交集产生(即: 当前cur经过一步到达下一个集合)
					if endSet[next] {
						return step + 1
					}

					// 每个元素之遍历一次, 产生 nextSet
					if !visited[next] && set[next] {
						visited[next] = true
						nextSet[next] = true
					}
				}
			}
		}

		// 核心: 选择较小的一端进行展开.
		// 开始的时候从 beginSet 展开 => nextSet, 接下来就需要从 endSet (因为更小)
		if len(endSet) < len(nextSet) {
			beginSet = endSet
			endSet = nextSet
		} else {
			beginSet = nextSet
		}

		step += 1
	}

	return 0
}

/*
490: 迷宫. 小气球走迷宫(起点->终点)

505: 迷宫II. 最短路径(SSSP单源最短路径算法[一个节点到其他所有节点的最短路径, Dijkstra的PQ实现)
*/
