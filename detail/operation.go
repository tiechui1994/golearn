package main

import "fmt"

/**
Go 当中的运算符:



|   OR运算
&   AND运算

^   一元运算符,表示取反; 二元运算符, 表示异或运算
&^  AND NOT, 位清空

**/

func main() {
	fmt.Println(1 | 2)
	fmt.Println(1 & 2)
	fmt.Println(1 ^ 2)
	fmt.Println(^(1 ^ 2))
	fmt.Println(1 &^ 2)
}
