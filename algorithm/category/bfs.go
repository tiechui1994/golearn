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

由空地和墙组成的迷宫中有一个球.
球可以向上下左右四个方向滚动, 但在遇到墙壁前不会停止滚动.
当球停下时, 可以选择下一个方向.

给定球的起始位置, 目的地和迷宫, 判断球能否在目的地停下.

505: 迷宫II. 最短路径.

每次保存除了坐标之外, 还要保存到当前位置的最短路径.

queue 使用优先级(根据最短路径进行排序).

dp[x][y] = min(dp[i][j]+count)  表示从 i, j 到 x, y 的最小距离

i, j 的开始起点是 start, 然后按照四个方向走, 之后都会到达 x,y (终点).
*/

func HasPath(maze [][]int, start []int, end []int) bool {
	dirs := [][2]int{{0, 1}, {0, -1}, {-1, 0}, {1, 0}}
	queue := make([][2]int, 0)
	visit := make(map[[2]int]bool)

	queue = append(queue, [2]int{start[0], start[1]})
	for len(queue) > 0 {
		size := len(queue)

		for i := 0; i < size; i++ {
			point := queue[0]
			queue = queue[1:]

			if point[0] == end[0] && point[1] == end[1] {
				return true
			}

			// 开始行走
			for _, dir := range dirs {
				point[0] += dir[0] // x
				point[1] += dir[1] // y
				// 开始行走, 直到撞墙
				for point[0] >= 0 && point[1] >= 0 &&
					point[0] < len(maze) && point[1] < len(maze[0]) &&
					maze[point[0]][point[1]] == 0 {
					point[0] += dir[0]
					point[1] += dir[1]
				}
				// 撞墙回退
				point[0] -= dir[0]
				point[1] -= dir[1]

				if !visit[point] {
					visit[point] = true
					queue = append(queue, point) // 进入队列
				}
			}
		}

	}

	return false
}

/*
拓扑排序:

每次选择入度为0的节点, 加入当前的 queue. 在遍历的时候, 会将下一个节点的入度减1.

indegree []int 节点的入度

grap map[node][]node // 节点node, node后的node

207. 课程表.(上课顺序)
*/

func CanFinish(numCourses int, prerequisites [][]int) {
	graph := make(map[int][]int)
	indgree := make([]int, numCourses)
	for i := 0; i < len(prerequisites); i++ {
		start := prerequisites[i][1] // 开始节点
		end := prerequisites[i][0]   // 结束节点
		if val, ok := graph[start]; ok {
			graph[start] = append(val, end)
		} else {
			graph[start] = []int{end}
		}

		indgree[end]++ // 节点入度
	}

	// 注: 如果是环, 则不存在入度为0的点
	queue := make([]int, 0)
	// 入度为0的点, 进入队列
	for i := 0; i < len(indgree); i++ {
		if indgree[i] == 0 {
			queue = append(queue, i)
		}
	}

	count := 0
	for len(queue) > 0 {
		size := len(queue)
		for i := 0; i < size; i++ {
			cur := queue[0]
			queue = queue[1:]
			count++

			nexts := graph[cur] // 遍历当前节点的下一个节点
			for _, next := range nexts {
				indgree[next]--
				if indgree[next] == 0 {
					queue = append(queue, next)
				}
			}
		}
	}
}
