package structs

import "fmt"

/*
Binary Indexed Tree(树状数组)

所谓的树状数组, 是一种用于高效处理对一个存储数字列表进行 "更新" 以及 "求前缀和" 的数据结构.

Binary Indexed Tree 事实上是将数字按照二进制表示来对数组中的元素进行逻辑上分层存储.

基本思想在于, 给定需要求和的位置 i, 比如 13, 利用二进制表示进行分段(或者分层)求和:

13 = 2^3 + 2^2 + 2^0 则, prefix(13) = RANGE(1,8) + RANGE(9,12) + RANGE(13,13)


Binary Indexed Tree:

              0000
   /     /      |         \
0001   0010   0100      1000
	    /       |     /   |   \
      0011     0101 1001 1010  1100
                |         |
               0111      1011

x & (-x) 刚好是最右边的第一个1的二进制.

x - x & (-x) 表示 child 寻找 parent 的过程

x + x & (-x) 当前元素的增加最右边的第一个1的二进制

parent -> child  0101 -> 0111 (最右边的第一个0变1)

child -> parent  1011 -> 1010 (最右边的第一个1变0)  x - x&(-x)

data[i] 存储的内容:
*/

type bit struct {
	data []int
}

func (b *bit) add(i int, inc int) {
	for i += 1; i < len(b.data); i = i + i&(-i) {
		b.data[i] += inc
	}
}

func (b *bit) sum(i int) int {
	sum := 0
	// i 是索引, i+=1 就是实际的二进制位置.
	// data 的实际长度是 N+1, 其中位置 0 不存储内容
	// i - i&(-i) child -> parent
	for i += 1; i > 0; i = i - i&(-i) {
		sum += b.data[i]
	}

	return sum
}

func main() {
	for i := 1; i < 16; i++ {
		fmt.Printf("%0.4b, %0.4b, %0.4b \n", i, i-i&(-i), i+i&(-i))
	}
}
