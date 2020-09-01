package tree

import (
	"math"
)

// 给定一个非空二叉树, 返回其最大路径和
// 路径: 一条从树中任意节点出发, 到达任意节点的序列. 该路径至少包含一个节点, 且不一定经过根节点

// 思路: 递归 + 动态规划
// 需要思考两个问题, 左右分支含根+[当前根], 左右分支含根+当前根, 左(右)分支不含根
// 左右分支含根: max(根, 左含根最大值+根, 右含根最大值+根)
func MaxSumPath(root *Node) int {
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

func maxSumPathRoot(root *Node) int {
	if root == nil {
		return 0
	}

	l := maxSumPathRoot(root.Left)
	r := maxSumPathRoot(root.Right)
	return Max(l+root.Val, r+root.Val, root.Val)
}

// 对称二叉树
// 性质: 1. 两个根节点具有相同的值; 2. 每个树的右子树与另一个树的左子树对称
// 迭代: 利用队列, 两棵树同时入队, 比较 (u.Left,v.Right), (u.Right,v.Left)
func IsSymmetric(root *Node) bool {
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
func BuildTree(inorder []int, postorder []int) *Node {
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

	var build func(inorder []int, postorder []int) *Node
	build = func(inorder []int, postorder []int) *Node {
		// 当只有一个节点或者没有节点
		if len(inorder) == 0 {
			return nil
		}
		if len(inorder) == 1 {
			return &Node{Val: inorder[0]}
		}
		// inorder: 左根右
		// postorder: 左右根
		//
		// 思路:
		// postorder 的最后一个元素就是当前的根, 以此根为划分 inorder, 划分出左右子树;
		// 与此同时, 也需要划分 postorder, 因为是 "左右根", 左子树,[0:左子树长度) 右子树, [左子树长度:n-1)
		n := len(postorder)
		root := &Node{Val: postorder[n-1]}

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
func NearestAncestor(root, p, q *Node) *Node {
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

// 二叉树中和为某一值的路径
// 打印出二叉树中节点值的和为输入整数的所有路径. 从树的 "根节点" 开始往下一直到 "叶节点" 所经过的节点形成一条路径.
// 思路: 自顶向下, 根左右, 直到根节点
func PathSum(root *Node, sum int) [][]int {
	if root == nil {
		return nil
	}

	if root.Left == nil && root.Right == nil && root.Val == sum {
		return [][]int{{root.Val}}
	}

	parent := make([]int, 0)
	return pathsum(root, sum, parent)
}

func pathsum(root *Node, sum int, parent []int) [][]int {
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

// 搜索树的第k小数, 中序遍历(左根右) -> 有序
func KthSmallest(root *Node, k int) int {
	if root == nil {
		return 0
	}

	var res int
	var count int
	var visit func(node *Node)
	visit = func(node *Node) {
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
func DistanceK(root, target *Node, k int) []int {
	// 1. 深度优先遍历, 记录node -> parent
	var parent = make(map[*Node]*Node) // node -> parent
	var dfs func(node, parent *Node)
	dfs = func(node, par *Node) {
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
	var seen = map[*Node]bool{
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
// 思路: 首先需要定义一个相同树结构比较的先序遍历函数, recur(root, cur *Node), 用来返回 root 和 cur 是否具有相同
// 的结构:
// 1. 如果 cur 为 nil, 说明子树已经遍历完成, 则返回 true
// 2. 如果 root 为 nil 或者 当前的 root 和 cur 值不一样, 则说明两者在当前节点不一样, 则返回 false
// 3. 递归遍历 root 和 cur 的 Left 和 Right
//
// t1 和 t2 具有相同的子结构, 只有三种状况:
// 1. 当前节点的 t1 和 t2 就具有相同结构
// 2. t1.Left 当中包含 t2 子结构(注意: 是包含, 可能包含 t1.Left节点)
// 3. t1.Right 当中包含 t2 子结构(注意: 是包含, 可能包含 t1.Right节点)
func IsSubStructure(t1 *Node, t2 *Node) bool {
	return (t1 != nil && t2 != nil) && (recur(t1, t2) || IsSubStructure(t1.Left, t2) || IsSubStructure(t1.Right, t2))
}

func recur(root *Node, cur *Node) bool {
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
func BtreeGameWinningMove(root *Node, n int, x int) bool {
	// 查找节点
	var search func(*Node, int) *Node
	search = func(node *Node, val int) *Node {
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
	var count func(*Node) int
	count = func(root *Node) int {
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
func SubtreeWithAllDeepest(root *Node) *Node {
	if root == nil || root.Left == nil && root.Right == nil {
		return root
	}

	// 计算节点的深度
	var dfsdepth func(cur *Node, dep int, depmap map[*Node]int)
	dfsdepth = func(cur *Node, dep int, depmap map[*Node]int) {
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

	var mintree func(cur *Node, max int, depmap map[*Node]int) *Node
	mintree = func(cur *Node, max int, depmap map[*Node]int) *Node {
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
	depmap := make(map[*Node]int)
	dfsdepth(root, 0, depmap)
	max := -1
	for _, v := range depmap {
		if max < v {
			max = v
		}
	}
	return mintree(root, max, depmap)
}
