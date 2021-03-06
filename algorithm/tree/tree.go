package tree

import (
	"log"
	"math"
	"strconv"
	"strings"
)

// 对称二叉树
// 性质: 1. 两个根节点具有相同的值; 2. 每个树的右子树与另一个树的左子树对称
// 迭代: 利用队列, 两棵树同时入队, 比较 (u.Left,v.Right), (u.Right,v.Left)
func IsSymmetric(root *TreeNode) bool {
	if root == nil {
		return true
	}

	q := Queue{}
	q.push(root)
	q.push(root)

	for !q.empty() {
		u := q.pop()
		v := q.pop()
		if u == nil && v == nil {
			continue
		}

		if u == nil || v == nil || u.Val != v.Val {
			return false
		}

		q.push(u.Left)
		q.push(v.Right)

		q.push(u.Right)
		q.push(v.Left)

	}

	return true
}

// 构建二叉树
func BuildTree(inorder []int, postorder []int) *TreeNode {
	if len(inorder) == 0 || len(postorder) == 0 {
		return nil
	}
	if len(inorder) != len(postorder) {
		return nil
	}
	index := func(arr []int, el int) int {
		for i := 0; i < len(arr); i++ {
			if arr[i] == el {
				return i
			}
		}

		return -1
	}

	var build func(inorder []int, postorder []int) *TreeNode
	build = func(inorder []int, postorder []int) *TreeNode {
		// 当只有一个节点或者没有节点
		if len(inorder) == 0 {
			return nil
		}
		if len(inorder) == 1 {
			return &TreeNode{Val: inorder[0]}
		}
		// inorder: 左根右
		// postorder: 左右根
		//
		// 思路:
		// postorder 的最后一个元素就是当前的根, 以此根为划分 inorder, 划分出左右子树;
		// 与此同时, 也需要划分 postorder, 因为是 "左右根", 左子树,[0:左子树长度) 右子树, [左子树长度:n-1)
		n := len(postorder)
		root := &TreeNode{Val: postorder[n-1]}

		idx := index(inorder, postorder[n-1])
		root.Left = build(inorder[:idx], postorder[0:idx])
		root.Right = build(inorder[idx+1:], postorder[idx:n-1])

		return root
	}

	return build(inorder, postorder)
}

// 最近公共祖先
// 思路: 自底向上递归, 左右根, 即后续遍历算法
// 1. 如果当前节点为 null, 则返回 null
// 2. 如果当前节点为 p 或 q, 或者包含了 p 或者 q 则返回当前节点
// 3. 如果当前包含了p和q, 则第一次到达的节点为最近的祖先
func NearestAncestor(root, p, q *TreeNode) *TreeNode {
	if root == nil {
		return nil
	}
	if root == p || root == q {
		return root
	}

	left := NearestAncestor(root.Left, p, q)
	right := NearestAncestor(root.Right, p, q)
	if left != nil && right != nil {
		return root
	}

	if left != nil {
		return left
	}
	if right != nil {
		return right
	}
	return nil
}

// 搜索树的第k小数, 中序遍历(左根右) -> 有序
func KthSmallest(root *TreeNode, k int) int {
	if root == nil {
		return 0
	}

	var res int
	var count int
	var visit func(node *TreeNode)
	visit = func(node *TreeNode) {
		if node.Left != nil {
			visit(node.Left)
		}

		count++
		if count == k {
			res = node.Val
			return
		}

		if node.Right != nil {
			visit(node.Right)
		}
	}

	visit(root)
	return res
}

func PathInZigZagTree(label int) []int {
	if label == 1 {
		return []int{1}
	}

	h := int(math.Log2(float64(label))) + 1
	var nodes = []int{label}

	var findparent func(h int, child int)
	findparent = func(h int, child int) {
		if h == 2 {
			nodes = append([]int{1}, nodes...)
			return
		}

		ps := int(math.Pow(2, float64(h-2)))
		pe := int(math.Pow(2, float64(h-1)) - 1)

		parent := ps + pe - child/2
		nodes = append([]int{parent}, nodes...)
		findparent(h-1, parent)
	}

	findparent(h, label)
	return nodes
}

// 到target的距离是k的节点.
// 思路1: 深度优先遍历, 获取node->parent的map关系.
// 利用queue先进先出的特性, target为头元素, 加入其 "元素的孩子和父亲". nil为特殊元素, 遇到nil的时候就需要
// 更新深度. 还有一个全局的已经加入的节点seen, 防止多次添加.
func DistanceK(root, target *TreeNode, k int) []int {
	// 1. 深度优先遍历, 记录node -> parent
	var parent = make(map[*TreeNode]*TreeNode) // node -> parent
	var dfs func(node, parent *TreeNode)
	dfs = func(node, par *TreeNode) {
		if node != nil {
			parent[node] = par
			dfs(node.Left, node)
			dfs(node.Right, node)
		}
	}

	dfs(root, nil)

	// 队列
	var queue = Queue{}
	queue.push(nil)
	queue.push(target) // nil和target被加入到队列当中. 最终存放节点
	var seen = map[*TreeNode]bool{
		target: true,
		nil:    true,
	}

	dist := 0
	for !queue.empty() {
		node := queue.pop()

		if node == nil {
			if dist == k {
				var res []int
				for i := range queue {
					res = append(res, queue[i].Val)
				}
				return res
			}

			queue.push(nil)
			dist++
		} else {
			if _, ok := seen[node.Left]; !ok {
				seen[node.Left] = true
				queue.push(node.Left)
			}
			if _, ok := seen[node.Right]; !ok {
				seen[node.Right] = true
				queue.push(node.Right)
			}

			p := parent[node] // 父节点
			if _, ok := seen[p]; !ok {
				seen[p] = true
				queue.push(p)
			}
		}
	}

	return nil
}

// 输入一个整数数组, 判断该数组是不是某二叉搜索树的后序遍历结果.
// 思路: 对于长度是0, 1, 2 的 postorder, 无论什么样的顺序, 都返回 true
//      对于长度为大于等于3的 postorder, 最后一个元素是 root, 然后在前面查找一个位置, 该位置满足条件:
//		该位置之前的元素, postorder[i] < root
//		该位置之后的元素, postorder[i] > root
// 如果存在这样的位置, 则说明当前的根元素满足要求, 依据找到的位置, 将 postorder 一分为2, 左子树和右子树,
// 当左右子树都满足要求, 则最终满足要求.
func VerifyPostorder(postorder []int) bool {
	index := func(arr []int, ele int) int {
		isLess := true
		index := -1
		for i, val := range arr {
			if isLess && i+1 < len(arr) {
				if arr[i+1] > ele {
					index = i
					isLess = false
					continue
				}
			}

			// 所有元素都小于 ele
			if isLess && i+1 == len(arr) {
				index = i
				isLess = false
				continue
			}

			// 条件检测
			if isLess && val > ele {
				return -1
			}

			if !isLess && val < ele {
				return -1
			}
		}

		return index
	}

	if len(postorder) <= 2 {
		return true
	}

	n := len(postorder)
	root := postorder[n-1]
	idx := index(postorder[:n-1], root)
	if idx == -1 {
		return false
	}

	return VerifyPostorder(postorder[0:idx+1]) && VerifyPostorder(postorder[idx+1:n-1])
}

// 树的子结构 | 检查子树
// 思路: 首先需要定义一个相同树结构比较的先序遍历函数, recur(root, cur *TreeNode), 用来返回 root 和 cur 是否具有相同
// 的结构:
// 1. 如果 cur 为 nil, 说明子树已经遍历完成, 则返回 true
// 2. 如果 root 为 nil 或者 当前的 root 和 cur 值不一样, 则说明两者在当前节点不一样, 则返回 false
// 3. 递归遍历 root 和 cur 的 Left 和 Right
//
// t1 和 t2 具有相同的子结构, 只有三种状况:
// 1. 当前节点的 t1 和 t2 就具有相同结构
// 2. t1.Left 当中包含 t2 子结构(注意: 是包含, 可能包含 t1.Left节点)
// 3. t1.Right 当中包含 t2 子结构(注意: 是包含, 可能包含 t1.Right节点)
func IsSubStructure(t1 *TreeNode, t2 *TreeNode) bool {
	return (t1 != nil && t2 != nil) && (recur(t1, t2) || IsSubStructure(t1.Left, t2) || IsSubStructure(t1.Right, t2))
}

func recur(root *TreeNode, cur *TreeNode) bool {
	// 检查子树
	//if cur == nil && root == nil {
	//	return true
	//}
	//if root == nil || cur == nil || root.Val != cur.Val {
	//	return false
	//}

	// 子结构
	if cur == nil {
		return true
	}

	if root == nil || root.Val != cur.Val {
		return false
	}

	return recur(root.Left, cur.Left) && recur(root.Right, cur.Right)
}

/**
关于图,或者是多叉树的深度优先遍历和广度优先遍历:

深度优先遍历的本质是:  root-childs. 因此首先确定每一个 node 的所有孩子节点, 对于二叉树, 三叉树, 以及一些特殊的结构体,
结构体本身以及包括了孩子节点. 但是, 针对一些不能直接获取孩子节点的结构体, 需要先建立起 root 和 child 的结构体关系. 最
常用的结构体是 [root]Set, map[root]Set.

接着就是按照固定的套路:

func dfs(node) {
	for child in getChild(node) {
		... // top-down
		dfs(child)
		... // down-top
	}
}

特别的(二叉树):

func dfs(node) {
	... // top-down
	dfs(node.Left)
	dfs(node.Right)
	... // down-top
}


广度优先遍历的本质是: 层序遍历, 即, 先遍历每一个节点, 然后将节点的孩子加入到队列当中, 不断循环, 直到队列为空.

**/

// 树每个节点到其他节点的和
//
// 深度优先搜索 + 子树计数
//
// 分析
// 用 ans 数组表示答案, 其中 ans[x] 表示节点 x 距离树中其它节点的距离之和. 直接对于每个 x 并通过搜索的方式计算
// ans[x] 的总时间复杂度是 O(N^2), 其中 N 是树中的节点个数, 这会超出时间限制. 因此我们需要找到 ans[x] 之间的关系,
// 减少重复搜索.
//
// 假设两个节点为 x 和 y, 它们在树中有一条边直接相连.
// 如果将这条边删除, 会得到一棵以 x 为根节点的子树 X 以及一棵以 y 为根节点的子树 Y.
//
// ans[x] 的值由三部分组成:
// 第一部分是子树 X 中所有节点到 x 的总距离, 记为 x@X;
// 第二部分是子树 Y 中所有节点到 y 的总距离, 记为 y@Y;
// 第三部分是子树 Y 中所有节点从 y 到达 x 的总距离, 它等于 Y 中的节点个数, 记为 #(Y).
// 将这三部分进行累加, 得到 ans[x] = x@X + y@Y + #(Y). 同理我们有 ans[y] = y@Y + x@X + #(X),
// 因此 ans[x] - ans[y] = #(Y) - #(X).
//
// 算法:
//
// 指定 0 号节点为树的根节点, 对于每个节点 node, 设 S(node) 为以 node 为根的子树 (包括 node 本身).
//
// 用 count[node] 表示 S(node) 中节点的个数, 并用 stsum[node](子树和)表示 S(node) 中所有节点到 node 的总距离.
//
// 可以使用 "深度优先搜索" 计算出所有的 count 和 stsum.
//
// 对于节点 node, 我们计算出它的每个子节点的 count 和 stsum 值, 那么就有:
// count[node] = sum(count[child]) + 1
// stsum[node] = sum(stsum[child] + count[child]), 其中 sum() 表示对子节点进行累加.
//
// 当所有节点计算完之后, 对于根节点, 它的答案 ans[root] 即为 stsum[root].
//
// 再进行一次"深度优先搜索", 并且根据上文推出的两个相邻节点 ans 值的关系, 得到其它节点的 ans 值, 即: 对于节点
// parent(父节点)以及节点 child(子节点), 有:
// ans[child] = ans[parent] - count[child] + (N - count[child]), 与上文的
// ans[y] = ans[x] - #(Y) + #(X) 相对应.
//
// 当第二次深度优先搜索结束后, 我们就得到了所有节点对应的 ans 值.

func SumOfDistancesInTree(N int, edges [][]int) []int {
	var dfsCount, dfsAns func(node, parent int, graph []Set, ans []int, count []int)
	dfsCount = func(node, parent int, graph []Set, ans []int, count []int) {
		for _, child := range graph[node].set() {
			if child != parent {
				dfsCount(child, node, graph, ans, count)
				count[node] += count[child]
				ans[node] += ans[child] + count[child]
			}
		}
	}

	dfsAns = func(node, parent int, graph []Set, ans []int, count []int) {
		for _, child := range graph[node].set() {
			if child != parent {
				ans[child] = ans[node] - count[child] + len(count) - count[child]
				dfsAns(child, node, graph, ans, count)
			}
		}
	}

	ans := make([]int, N)
	// 首先, 对于每一个节点, 肯定包含其自身
	count := make([]int, N)
	for i := range count {
		count[i] = 1
	}
	// 构建, 父子关系, 这是为深度优先遍历做准备工作
	graph := make([]Set, N)
	for i := 0; i < N; i++ {
		graph[i] = Set{}
	}
	for _, edge := range edges {
		graph[edge[0]].add(edge[1])
		graph[edge[1]].add(edge[0])
	}

	// 深度优先遍历, 计算 count
	dfsCount(0, -1, graph, ans, count)
	// 深度优先遍历, 计算 ans
	dfsAns(0, -1, graph, ans, count)

	return ans
}

// 二叉树着色问题
//
// 取胜的关键是取得一半以上的节点数量.
// 一个节点, 如果取它的左(右)节点, 则(左)右节点的所有节点都是你的. 取父节点, 则父节点以上的所有节点都是你的.
// 一号玩家选取了 x 节点, 其左节点的总数是 left, 右节点的总数是 right:
// 1.left > m 或者 right > m, 你只要选取大于一半的那个节点, 则必赢
// 2. left == m 或者 right == m, 无论选取哪个节点, 其他另一半和父节点都是一号玩家的, 则必输
// 3. left + right + 1 <= m, 选取父节点, 即必赢
// 4. left + right + 1 > m, 无论选取哪个节点, 其他另一半节点和父节点都是一号玩家的, 则必输
func BtreeGameWinningMove(root *TreeNode, n int, x int) bool {
	// 查找节点
	var search func(*TreeNode, int) *TreeNode
	search = func(node *TreeNode, val int) *TreeNode {
		if node == nil {
			return nil
		}
		if node.Val == val {
			return node
		}

		left := search(node.Left, val)
		right := search(node.Right, val)
		if left != nil {
			return left
		}

		return right
	}

	// 统计节点子树的个数, 包括节点
	var count func(*TreeNode) int
	count = func(root *TreeNode) int {
		if root == nil {
			return 0
		}
		if root.Left == nil && root.Right == nil {
			return 1
		}

		left := count(root.Left)
		right := count(root.Right)
		return left + right + 1
	}

	if n == 1 || root == nil {
		return false
	}

	m := n / 2
	node := search(root, x)
	left := count(node.Left)
	right := count(node.Right)
	if left > m || right > m || left+right < m {
		return true
	}

	return false
}

// 树的最短高度问题, 给定一个树, 找到 root 节点, 到达各个子节点的高度最小
//
// 思路: 首先分析结果, 最多有两个节点. 原因很简单, 三个节点形成一个平面, 但这只是一个点或者线的问题
// 整体的思路就是剪除叶子节点的思想.
// 叶子节点的特点是度为1.
// 当剪除叶子节点的时候, 可以根据叶子节点找到邻居节点, 如果剪除之后, 邻居节点的度是1, 则又加入到剪除的队列,
// 从而形成了一个广度遍历.
// 依照上述的思想, 最终所有的节点都会被剪除掉. 那么最后一次剪除前保存的元素就是最终的结果.
//
// 叶子节点到根的聚拢算法.
// a-b 或 a
func FindMinHeightTrees(n int, edges [][]int) []int {
	if n == 0 {
		return nil
	}
	if n == 1 {
		return []int{0}
	}
	if n == 2 {
		return []int{0, 1}
	}

	degree := make([]int, n)
	graph := make([]Set, n)
	for _, v := range edges {
		degree[v[0]]++
		degree[v[1]]++
		graph[v[0]].add(v[1])
		graph[v[1]].add(v[0])
	}

	q := make([]int, 0)
	for i := 0; i < n; i++ {
		if degree[i] == 1 {
			q = append(q, i)
		}
	}

	var res []int
	for len(q) != 0 {
		res = nil
		size := len(q)
		for i := 0; i < size; i++ {
			cur := q[0]
			q = q[1:]
			res = append(res, cur)
			members := graph[cur].set()
			for _, member := range members {
				degree[member]--
				if degree[member] == 1 {
					q = append(q, member)
				}
			}
		}
	}

	return res
}

// 具有所有最深结点的最小子树
//
// 两次深度优先遍历
// 1. 第一次深度优先遍历, 计算出每个节点的深度
// 2. 第二次深度优先遍历, 获取具有最大深度的子树
//	  - 如果当前节点具有最大深度, 返回当前节点
//	  - 如果当前节点的左右孩子具有最大深度, 返回当前节点
//	  - 如果当前节点的左(右)孩子具有最大深度, 返回左(右)孩子
func SubtreeWithAllDeepest(root *TreeNode) *TreeNode {
	if root == nil || root.Left == nil && root.Right == nil {
		return root
	}

	// 计算节点的深度
	var dfsdepth func(cur *TreeNode, dep int, depmap map[*TreeNode]int)
	dfsdepth = func(cur *TreeNode, dep int, depmap map[*TreeNode]int) {
		if cur == nil {
			return
		}
		depmap[cur] = dep
		if cur.Left != nil {
			dfsdepth(cur.Left, dep+1, depmap)
		}
		if cur.Right != nil {
			dfsdepth(cur.Right, dep+1, depmap)
		}
	}

	var mintree func(cur *TreeNode, max int, depmap map[*TreeNode]int) *TreeNode
	mintree = func(cur *TreeNode, max int, depmap map[*TreeNode]int) *TreeNode {
		if cur == nil {
			return nil
		}
		if depmap[cur] == max {
			return cur
		}

		left := mintree(cur.Left, max, depmap)
		right := mintree(cur.Right, max, depmap)
		if left != nil && right != nil {
			return cur
		}
		if left != nil {
			return left
		}
		if right != nil {
			return right
		}

		return nil
	}

	// 计算
	depmap := make(map[*TreeNode]int)
	dfsdepth(root, 0, depmap)
	max := -1
	for _, v := range depmap {
		if max < v {
			max = v
		}
	}
	return mintree(root, max, depmap)
}

// 好叶子节点对的数量
//
// 给你二叉树的根节点 root 和一个整数 distance.
//
// 如果二叉树中两个叶节点之间的最短路径长度小于或者等于 distance, 那它们就可以构成一组好叶子节点对.
//
// 思路: 深度优先遍历, 计算当前节点到叶子节点的距离数组(长度是distance+1).
// 1. 如果当前节点为叶子节点, 则 distance[0]=1, 即到叶子节点的距离是0的有一个
// 2. 计算完左右节点之后, 当前节点 distance[i] = left[i-1] + right[i-1], i>=1, 即当前节点到叶子节点距离为i的个数
//
// 计算: left, right 表示左右孩子节点到叶子节点距离数组
// ans += left[i] * left[j]; i+j+2 <= distance, i 左孩子距离叶子节点值为 i 的个数, j 是右孩子距离叶子节点的个数
// i+1+j+1 则表示连接当前节点的孩子的距离.
func CountPairs(root *TreeNode, distance int) int {
	var dfs func(root *TreeNode, distance int, ans *int) []int
	dfs = func(root *TreeNode, distance int, ans *int) []int {
		if root == nil {
			return make([]int, distance+1)
		}

		ret := make([]int, distance+1)
		if root.Left == nil && root.Right == nil {
			ret[0] = 1
			return ret
		}

		left := dfs(root.Left, distance, ans)
		right := dfs(root.Right, distance, ans)
		for i := 0; i < distance; i++ {
			for j := 0; i+j+2 <= distance; j++ {
				*ans += left[i] * right[j]
			}
		}

		for i := 1; i <= distance; i++ {
			ret[i] = left[i-1] + right[i-1]
		}

		return ret
	}

	var ans = 0
	dfs(root, distance, &ans)
	return ans
}

// 二叉树中和为某一值的路径, (从根到某个节点, 限定了开始的位置)
//
// 打印出二叉树中节点值的和为输入整数的所有路径. 从树的 "根节点" 开始往下一直到 "叶节点" 所经过的节点形成一条路径.
// 思路: 自顶向下, 根左右, 直到根节点
func PathSumI(root *TreeNode, sum int) [][]int {
	var pathsum func(root *TreeNode, sum int, parent []int) [][]int
	pathsum = func(root *TreeNode, sum int, parent []int) [][]int {
		if root.Left == nil && root.Right == nil {
			if root.Val == sum {
				parent = append(parent, root.Val)
				return [][]int{parent}
			}
			return nil
		}

		// 分别计算左右子树, 和为 sum-root.Val 的路径
		var left, right [][]int
		if root.Left != nil {
			// 注意: 由于 slice, 必须使用深度复制
			t := make([]int, len(parent))
			copy(t, parent)
			t = append(t, root.Val)
			left = pathsum(root.Left, sum-root.Val, t)
		}
		if root.Right != nil {
			t := make([]int, len(parent))
			copy(t, parent)
			t = append(t, root.Val)
			right = pathsum(root.Right, sum-root.Val, t)
		}

		return append(left, right...)
	}

	if root == nil {
		return nil
	}
	if root.Left == nil && root.Right == nil && root.Val == sum {
		return [][]int{{root.Val}}
	}

	parent := make([]int, 0)
	return pathsum(root, sum, parent)
}

// 任意节点之间的和为某一值. 其方向必须向下(只能从父节点指向子节点方向)
//
// 需要注意的条件: 满足条件的路径的方向, 否则可能陷入误区.
//
// 思路: 前缀和计数 + 深度优先遍历
//
// 前缀和: 到达当前元素的路径上, 之前所有元素的和.
//
// 如果两个节点的前缀和是相同的, 那么这两个节点之间的元素和为零.
func PathSumII(root *TreeNode, sum int) int {
	var recurPathSum func(root *TreeNode, prefixSumCount map[int]int, target, cursum int) int
	recurPathSum = func(root *TreeNode, prefixSumCount map[int]int, target, cursum int) int {
		if root == nil {
			return 0
		}

		res := 0
		cursum += root.Val
		res += prefixSumCount[cursum-target]
		prefixSumCount[cursum] = prefixSumCount[cursum] + 1

		res += recurPathSum(root.Left, prefixSumCount, target, cursum)
		res += recurPathSum(root.Right, prefixSumCount, target, cursum)

		// 由于当前节点已经遍历完成, 对于上一层而言, 其已经失去作用.
		// 前缀和只对当前节点的子孙节点有作用, 对于父辈节点已经无影响.
		prefixSumCount[cursum] = prefixSumCount[cursum] - 1

		return res
	}
	if root == nil {
		return 0
	}
	prefixSumCount := map[int]int{0: 1}
	return recurPathSum(root, prefixSumCount, sum, 0)
}

// 给定一个非空二叉树, 返回其最大路径和(比较任意)
// 路径: 一条从树中任意节点出发, 到达任意节点的序列. 该路径至少包含一个节点, 且不一定经过根节点

// 思路: 递归 + 动态规划
// 需要思考两个问题, 左右分支含根+[当前根], 左右分支含根+当前根, 左(右)分支不含根
// 左右分支含根: max(根, 左含根最大值+根, 右含根最大值+根)
func MaxSumPath(root *TreeNode) int {
	if root == nil {
		return 0
	}

	if root.Left != nil && root.Right != nil {
		ll := maxSumPathRoot(root.Left)
		rr := maxSumPathRoot(root.Right)

		l := MaxSumPath(root.Left)
		r := MaxSumPath(root.Right)
		return Max(root.Val, ll+root.Val, rr+root.Val, ll+rr+root.Val, ll, rr, l, r)
	} else if root.Left != nil {
		ll := maxSumPathRoot(root.Left)
		l := MaxSumPath(root.Left)
		return Max(root.Val, root.Val+ll, ll, l)
	} else if root.Right != nil {
		rr := maxSumPathRoot(root.Right)
		r := MaxSumPath(root.Right)
		return Max(root.Val, root.Val+rr, rr, r)
	}

	return root.Val
}

func maxSumPathRoot(root *TreeNode) int {
	if root == nil {
		return 0
	}

	l := maxSumPathRoot(root.Left)
	r := maxSumPathRoot(root.Right)
	return Max(l+root.Val, r+root.Val, root.Val)
}

// 删点成林
//
// 一棵树, 删除一些节点, 形成林
//
// 思路: 深度优先遍历获取父子关系, 删除节点之后, 寻找父节点为 nil 的树则是最终的结果
func DelNodes(root *TreeNode, to_delete []int) []*TreeNode {
	// child -> parent 关系
	var dfs func(root *TreeNode, parents map[*TreeNode]*TreeNode)
	dfs = func(root *TreeNode, parents map[*TreeNode]*TreeNode) {
		if root == nil {
			return
		}
		if root.Left != nil {
			parents[root.Left] = root
		}
		if root.Right != nil {
			parents[root.Right] = root
		}

		dfs(root.Left, parents)
		dfs(root.Right, parents)
	}

	if root == nil {
		return nil
	}
	if to_delete == nil {
		return []*TreeNode{root}
	}

	var parents = make(map[*TreeNode]*TreeNode)
	parents[root] = nil
	dfs(root, parents)

	// 删除
	for child, parent := range parents {
		var exist bool
		if len(to_delete) > 0 {
			for i, v := range to_delete {
				if child.Val == v {
					exist = true
					to_delete = append(to_delete[0:i], to_delete[i+1:]...)
					break
				}
			}
		}

		if !exist {
			continue
		}

		log.Println(child.Val, root.Val)

		if parent != nil {
			if parent.Left == child {
				parent.Left = nil
			}

			if parent.Right == child {
				parent.Right = nil
			}
		}

		delete(parents, child)
		if child.Left != nil {
			parents[child.Left] = nil
		}
		if child.Right != nil {
			parents[child.Right] = nil
		}

	}

	var res []*TreeNode
	for child, parent := range parents {
		if parent == nil {
			res = append(res, child)
			log.Println(child)
		}
	}

	return res
}

// 一棵树, 摘苹果的最小时间
//
// 思路:  自顶向下, 深度优先遍历, 记录每个节点的父节点, 先需要一个每个节点的叶节点
//       自底而上, 深度优先遍历, 统计获取每个苹果父亲的节点(需要使用一个visit数组记录已经访问的父节点)
func MinTime(n int, edges [][]int, hasApple []bool) int {
	var ans int
	// 记录每个节点的父节点
	parents := make([]int, n)
	for i := range parents {
		parents[i] = -1
	}

	// 自根向叶子的深度优先遍历
	var buildParents func(nodemap [][]int, val int)
	buildParents = func(nodemap [][]int, val int) {
		for _, child := range nodemap[val] {
			if child != 0 && parents[child] == -1 {
				parents[child] = val
				buildParents(nodemap, child)
			}
		}
	}

	// 记录当前节点是否已经访问过, 自叶子向根的深度优先遍历
	visited := make([]bool, n)
	var dfsEdge func(to int)
	dfsEdge = func(to int) {
		if !visited[to] {
			visited[to] = true
			ans++
			dfsEdge(parents[to])
		}
	}

	// nodeMap => 树结构
	nodeMap := make([][]int, n)
	for _, edge := range edges {
		from, to := edge[0], edge[1]
		nodeMap[from] = append(nodeMap[from], to)
		nodeMap[to] = append(nodeMap[to], from)
	}

	// 从树根节点开始访问, 深度优先遍历. 访问完所有的节点
	buildParents(nodeMap, 0)

	visited[0] = true // 根节点总会被访问到的
	for i := 0; i < n; i++ {
		if hasApple[i] {
			dfsEdge(i)
		}
	}

	return ans * 2
}

// 二叉搜索树中的众数
//
// 空间复杂度为常数
// 两次中序列遍历: 计算当前数计数 curCount, 最大计数maxCount, 最大计数个数retCount
// a), 判断是否存在 pre, 比较 pre 的值和当前的值, 如果一致, curCount++, 否则 curCount=1
// b), 判断当前的 curCount > maxCount, 如果是, 则 maxCount = curCount, retCount=1
// c), 判断当前的 curCount == maxCount, 如果是, 则 retCount++ 并将值填入到结果集合当中.
//
// 注意: 第一次遍历和第二次遍历之间, 需要将 curCount, retCount 重置为0, pre = nil
func FindMode(root *TreeNode) []int {
	var (
		pre                          *TreeNode
		curCount, maxCount, retCount int
		ret                          []int
	)
	var inOrder func(cur *TreeNode)
	inOrder = func(cur *TreeNode) {
		if cur == nil {
			return
		}

		inOrder(cur.Left)

		if pre != nil && pre.Val == cur.Val {
			curCount++
		} else {
			curCount = 1
		}

		if curCount > maxCount {
			maxCount = curCount
			retCount = 1
		} else if curCount == maxCount {
			if ret != nil {
				ret = append(ret, cur.Val)
			}
			retCount++
		}

		pre = cur
		inOrder(cur.Right)
	}

	inOrder(root)

	// 非常关键, 否则第二次遍历就会无效.
	pre = nil
	retCount = 0
	curCount = 0
	ret = make([]int, 0, retCount)
	inOrder(root)

	return ret
}

// 二叉树中的最长交错路径
//
// 给你一棵以 root 为根的二叉树，二叉树中的交错路径定义如下：
//
// 选择二叉树中"任意"节点和一个方向(左或者右).
//
// 如果前进方向为右, 那么移动到当前节点的的右子节点, 否则移动到它的左子节点.
// 改变前进方向: 左变右或者右变左.
// 重复第二步和第三步, 直到你在树中无法继续移动.
//
// 交错路径的长度定义为: 访问过的节点数目 - 1 (单个节点的路径长度为 0)
//
// 深度优先遍历: 记录当前节点是处于的方向, 左边还是右边, 同时到当前节点的路径长度.
//             如果孩子方向与当前处于的方向一致, 则重新开启计算, 否则孩子的路径长度=当前路径长度+1
//
func LongestZigZag(root *TreeNode) int {
	if root == nil || root.Right == nil && root.Left == nil {
		return 0
	}

	// direct true: left  false: right
	var dfs func(node *TreeNode, parent int, direct bool, ans *int)
	dfs = func(node *TreeNode, parent int, direct bool, ans *int) {
		if node == nil {
			return
		}

		if *ans < parent {
			*ans = parent
		}

		if node.Left == nil && node.Right == nil {
			return
		}

		if direct {
			if node.Left != nil {
				dfs(node.Left, 1, direct, ans)
			}
			if node.Right != nil {
				dfs(node.Right, parent+1, !direct, ans)
			}
		} else {
			if node.Left != nil {
				dfs(node.Left, parent+1, !direct, ans)
			}
			if node.Right != nil {
				dfs(node.Right, 1, direct, ans)
			}
		}
	}

	var ans int
	dfs(root.Left, 1, true, &ans)
	dfs(root.Right, 1, false, &ans)

	return ans
}

// 先序遍历构建树, 分割点是路径长度
// 特点, 第一个节点是根, 后面的元素被当前节点路径个数个"-"分割成左右子树, 然后进行地柜遍历
func RecoverFromPreorder(s string) *TreeNode {
	return buildTree(s, 1)
}

func buildTree(s string, level int) *TreeNode {
	if len(s) == 0 {
		return nil
	}

	// 精准查找 sub 的位置
	indexOf := func(source string, sub string) (idx int) {
		l := len(sub)
		n := len(source)
		for i := 0; i < n; i++ {
			if i+l <= n && source[i:i+l] == sub {
				// 满足条件的下一个元素非 "-"
				if i == 0 && i+l < n && source[i+l] != '-' {
					return i
				}
				// 满足条件的前一个元素和下一个元素非 "-"
				if i > 0 && source[i-1] != '-' && i+l < n && source[i+l] != '-' {
					return i
				}
			}
		}
		return -1
	}

	// 当前 level 所形成的 "-"
	cut := strings.Repeat("-", level)
	idx := indexOf(s, cut)

	// 没有找到 cut, 说明当前元素是最后一个元素
	if idx == -1 {
		return &TreeNode{Val: str2int(s)}
	}

	// 找到 cut, 说明当前节点还存在孩子
	root := &TreeNode{Val: str2int(s[:idx])}
	newsource := s[idx+level:]

	// 查找分割点, 如果找到, 说明当前节点存在左右子树, 没有找到说明当前节点只有左子树
	idx = indexOf(newsource, cut)
	if idx == -1 {
		root.Left = buildTree(newsource, level+1)
	} else {
		root.Left = buildTree(newsource[:idx], level+1)
		root.Right = buildTree(newsource[idx+level:], level+1)
	}

	return root
}

func str2int(s string) int {
	i, _ := strconv.ParseInt(s, 10, 64)
	return int(i)
}

// 所有满二叉树
//
// 满二叉树是一类二叉树, 其中每个结点恰好有 0 或 2 个子结点.
//
// 返回包含 N 个结点的所有可能满二叉树的列表. 答案的每个元素都是一个可能树的根结点.
//
// 思路: 一棵树, 根, 左孩子, 右孩子 构成
// 针对 0, 1, 2, 3 这些满二叉树都是特殊的树
// 当 N > 3 时, 除去根节点, 分别取查找 i 个节点左子树 和 j 个节点的右子树, i+j+1=N
// 当查找到后, 也就是说存在这样的左子树和右子树. 使用笛卡尔集, 就可以获得最终的结果
func AllPossibleFBT(N int) []*TreeNode {
	if N == 0 {
		return nil
	}
	if N == 1 {
		root := &TreeNode{}
		return []*TreeNode{root}
	}
	if N == 2 {
		return nil
	}
	if N == 3 {
		root := &TreeNode{}
		root.Left = &TreeNode{}
		root.Right = &TreeNode{}
		return []*TreeNode{root}
	}

	var res []*TreeNode
	for i := 1; i < N-1; i++ {
		j := N - i - 1
		left := AllPossibleFBT(i)
		right := AllPossibleFBT(j)
		if len(left) == 0 || len(right) == 0 {
			continue
		}
		for k := 0; k < len(left); k++ {
			for v := 0; v < len(right); v++ {
				root := &TreeNode{}
				root.Left = left[k]
				root.Right = right[v]
				res = append(res, root)
			}
		}
	}

	return res
}
