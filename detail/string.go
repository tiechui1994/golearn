package main

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

// primeRK is the prime base used in Rabin-Karp algorithm.
const primeRK = 16777619

// hashStr returns the hash and the appropriate multiplicative
// factor for use in Rabin-Karp algorithm.
func hashStr(sep string) (uint32, uint32) {
	// hash 算法
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*primeRK + uint32(sep[i])
	}

	// Rabin 当中的乘数
	var pow, sq uint32 = 1, primeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

/**
核心: 利用了hash算法的递归计算效应
**/

func IndexRabinKarp(s, substr string) int {
	if len(s) < len(substr) {
		return -1
	}
	// Rabin-Karp search
	hashss, pow := hashStr(substr)

	// 快速比较
	n := len(substr)
	var h uint32
	for i := 0; i < n; i++ {
		h = h*primeRK + uint32(s[i])
	}
	if h == hashss && s[:n] == substr {
		return 0
	}

	// 慢比较
	for i := n; i < len(s); {
		h = h*primeRK + uint32(s[i])
		h -= pow * uint32(s[i-n])
		i++
		if h == hashss && s[i-n:i] == substr {
			return i - n
		}
	}
	return -1
}

func main() {
	println(IndexRabinKarp("Rabin–Karp string search algorithm: Rabin-Karp", "string"))
}
