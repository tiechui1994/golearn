package category

import "sort"

/*
扫描线:

基础例题: 给出飞机的起飞和降落时间列表, 用序列 interval 表示. 计算出天上同时最多有多少架飞机?

思路: 首先, 计算出飞起起飞和降落时间点, 统计当前这个时刻天上的飞机总数.(飞机起飞+1, 飞机降落-1)

253. 会议房间II


类似例题:
56(合并区间): 当前区间和最后一个区间做比较, 如果分离, 则加入当前区域; 如果需要合并, 修改最后一个区间的右边界值.
57(插入区间):
当 add == nil // add(cur)
当 add != nil && cur[1] < add[0] // add(cur), cur位于add的左边界
当 add != nil && add[1] < cur[0] // add(add, cur), add=nil, cur位于add的右边界
其他 // add[0]=min(add[0], cur[0]); add[1]=max(add[1], cur[1])


1272(删除区间), 左边界, 右边界(直接加入), 交接区域(加入只保留未被删除的左边界和右边界)

1288: 删除可以覆盖的区间. 排序+(当前面的可以覆盖后面的, 删除后面的那个)

352: 将数据流变为多个不相交区间

给定一个非负整数的数据流输入 a1,a2,…,an,…,将到目前为止看到的数字总结为不相交的区间列表.

数字流: 1, 3, 7, 2, 6

[1, 1]
[1, 1], [3, 3]
[1, 1], [3, 3], [7, 7]
[1, 3], [7, 7]
[1, 3], [6, 7]

思路: 左边紧挨着(2个合并), 右边紧挨着(2个合并), 两边都紧挨着(3个合并)

1229: 会议安排(区间交集)

986: 区间链表交集

759: 员工的空闲时间

218: 天气线. 思路: 打飞机(飞机个数是高度) + 优先级队列; 每次选择最大高度的高度.
*/

func NumOfAir(interval [][]int) int {
	list := make([][2]int, 0) // 记录当前时刻, 飞机在天上的数量
	for _, v := range interval {
		list = append(list, [2]int{v[0], 1})  // 飞起起飞
		list = append(list, [2]int{v[1], -1}) // 飞机降落
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i][0] != list[j][0] {
			return list[i][0] < list[j][0]
		} else {
			return list[i][1] < list[j][1]
		}
	})

	// 分别统计当前时刻天上的飞机总数和飞机数量的最大值
	count := 0
	ans := 0
	for _, val := range list {
		count += val[1]
		if ans < count {
			ans = count
		}
	}

	return ans
}

func DeleteInterval(interval [][]int, del []int) [][]int {
	var ans [][]int
	for _, cur := range interval {
		// 左边界, 右边界(这里使用 "=" 巧妙的解决了边界相等问题)
		if cur[1] <= del[0] {
			ans = append(ans, cur)
		} else if del[1] <= cur[0] {
			ans = append(ans, cur)
		} else {
			if cur[0] < del[0] {
				ans = append(ans, []int{cur[0], del[0]})
			}
			if del[1] < cur[1] {
				ans = append(ans, []int{del[1], cur[1]})
			}
		}
	}

	return ans
}
