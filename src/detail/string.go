package detail

import (
	"fmt"
	"unicode/utf8"
)

// 字符串不总是UTF8文本.
// 字符串的值不需要是UTF8的文本. 它可以包含任意的字节. 只有在string literal(字符串字面量)使用时, 字符串才会是UTF8.
// 即使之后它可以使用转义序列来包含其他数据.

func StringValue() {
	uft8str := "ABC"
	fmt.Printf("%v is UTF8: %v \n", uft8str, utf8.ValidString(uft8str))

	notuft8str := "A\xeeC"
	fmt.Printf("%v is UTF8: %v \n", notuft8str, utf8.ValidString(notuft8str))
}

// 字符串长度
// 内置的 len() 函数返回的是 byte 的数量.
// 对于 unicode 在 0-127 这之间的 char 的 byte 是1, 128以上的 char 的 byte 大于1
// utf8.RuneCountInString() 函数返回的是 rune 的数量. 即 unicode 的数量

func StringLen() {
	char := "♥"
	fmt.Println(char, "rune", utf8.RuneCountInString(char))
	fmt.Println(char, "rune len", len([]rune(char)))
	fmt.Println(char, "len", len(char))
}

// 字符和数字之间的转换
// char -> int            int(char)
// int -> char -> string  string(int)
func Convert() {
	for i := 0; i < 16535; i++ {
		fmt.Printf("%s	", string(i))
		if i%30 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()

	fmt.Println(int('✓'), int('✔'), int('✕'))
}
