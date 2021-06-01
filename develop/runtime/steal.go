package main

import "fmt"

var stealOrder randomOrder


// count 为总数, coprimes 保存所有的互质数
type randomOrder struct {
	count    uint32
	coprimes []uint32
}

func (ord *randomOrder) reset(count uint32) {
	ord.count = count
	ord.coprimes = ord.coprimes[:0]
	for i := uint32(1); i <= count; i++ {
		if gcd(i, count) == 1 {
			ord.coprimes = append(ord.coprimes, i)
		}
	}
}
func gcd(a, b uint32) uint32 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

type randomEnum struct {
	i     uint32 // 计数器, 统计多少次
	count uint32
	pos   uint32
	inc   uint32
}

// i 是随机数
func (ord *randomOrder) start(i uint32) randomEnum {
	return randomEnum{
		count: ord.count,
		pos:   i % ord.count, // 随机的开始的位置
		inc:   ord.coprimes[i%uint32(len(ord.coprimes))],  // 随机的互质数
	}
}


func main() {
	stealOrder.reset(15)
	fmt.Println(stealOrder.coprimes)
}
