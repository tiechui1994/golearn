package main

import (
	"strings"
	"crypto/hmac"
	"crypto/sha1"
)

const (
	char_zero = '\000'
	char_line = "\n"
)

const (
	SECRET_BITS               = 128
	VERIFICATION_CODE_MODULUS = 1000 * 1000
	SCRATCHCODES              = 5
	MAX_SCRATCHCODES          = 10
	SCRATCHCODE_LENGTH        = 8
	BYTES_PER_SCRATCHCODE     = 4
	BITS_PER_BASE32_CHAR      = 5 // Base32 expands space by 8/5

	SHA1_DIGEST_LENGTH = 20
)

// 返回 str1 中第一个不在字符串 str2 中出现的字符下标
func strspn(str1, str2 string) int {
	if strings.Contains(str2, str1) {
		return len(str1)
	}

	set := make(map[rune]int)
	for i, b := range str2 {
		set[b] = i
	}

	for i, b := range str1 {
		if _, ok := set[b]; !ok {
			return i
		}
	}

	return len(str1)
}

// 检索字符串 str1 开头连续有几个字符都不含字符串 str2 中的字符
func strcspn(str1, str2 string) int {
	set := make(map[rune]int)
	for i, b := range str2 {
		set[b] = i
	}

	var idx int
	for _, b := range str1 {
		if _, ok := set[b]; !ok {
			idx += 1
		} else {
			break
		}
	}

	return idx
}

func lower(c byte) byte {
	return c | ('x' - 'X')
}

func strtouint(s string, endptr *string, base int) uint64 {
	if endptr == nil {
		endptr = new(string)
	}

	*endptr = ""
	base0 := base == 0

	switch {
	case 2 <= base && base <= 36:
	case base == 0:
		base = 10
		if s[0] == '0' {
			switch {
			case len(s) >= 3 && lower(s[1]) == 'b':
				base = 2
				s = s[2:]
			case len(s) >= 3 && lower(s[1]) == 'o':
				base = 8
				s = s[2:]
			case len(s) >= 3 && lower(s[1]) == 'x':
				base = 16
				s = s[2:]
			default:
				base = 8
				s = s[1:]
			}
		}
	default:
		return 0
	}

	const intSize = 32 << (^uint(0) >> 63)
	const maxUint64 = 1<<64 - 1

	var cutoff uint64
	switch base {
	case 10:
		cutoff = maxUint64/10 + 1
	case 16:
		cutoff = maxUint64/16 + 1
	default:
		cutoff = maxUint64/uint64(base) + 1
	}

	bitSize := int(intSize)
	maxVal := uint64(1)<<uint(bitSize) - 1
	var n uint64
	for i, c := range []byte(s) {
		var d byte
		switch {
		case c == '_' && base0:
			// underscoreOK already called
			continue
		case '0' <= c && c <= '9':
			d = c - '0'
		case 'a' <= lower(c) && lower(c) <= 'z':
			d = lower(c) - 'a' + 10
		default:
			*endptr = s[i:]
			return n
		}

		if d >= byte(base) {
			*endptr = s[i:]
			return n
		}

		if n >= cutoff {
			*endptr = s[i:]
			// n*base overflows
			return maxVal
		}

		n *= uint64(base)
		n1 := n + uint64(d)
		if n1 < n || n1 > maxVal {
			*endptr = s[i:]
			// n+v overflows
			return maxVal
		}

		n = n1
	}

	return n
}

func GenerateCode(key []byte, tm int) int {
	var challenge [8]uint8
	for i := len(challenge) - 1; i >= 0; tm >>= 8 {
		challenge[i] = uint8(tm)
		i--
	}

	secretLen := (strlen(key) + 7) / 8 * BITS_PER_BASE32_CHAR

	if secretLen <= 0 || secretLen > 100 {
		return -1
	}

	var secret [100]uint8
	n := base32Decode(secret[:], []byte(key))
	if n == -1 {
		return -1
	}

	hash := HmacSha1(secret[:n], challenge[:])

	offset := int(hash[SHA1_DIGEST_LENGTH-1] & 0xF)
	var truncatedHash uint
	for i := 0; i < 4; i++ {
		truncatedHash <<= 8
		truncatedHash |= uint(hash[offset+i])
	}

	truncatedHash &= 0x7FFFFFFF
	truncatedHash %= VERIFICATION_CODE_MODULUS

	return int(truncatedHash)
}

func HmacSha1(key []byte, data []byte) (result [SHA1_DIGEST_LENGTH]uint8) {
	mac := hmac.New(sha1.New, key)
	mac.Write(data)
	copy(result[:], mac.Sum(nil))
	return result
}

// 字符串长度
func strlen(data []byte) int {
	var i = len(data) - 1
	for i >= 0 {
		if data[i] == char_zero {
			i--
		} else {
			return i + 1
		}
	}

	return 0
}

// 字符串追加
func strcat(data []byte, str string) {
	sl := strlen(data)
	copy(data[sl:], str)
}

func base32Encode(dst []byte, src []byte) int {
	legth := len(src)
	if legth < 0 || legth > (1<<28) {
		return -1
	}

	var count = 0
	if legth > 0 {
		encodeStd := "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
		buffer := int(src[0])
		next := 1
		bitsLeft := 8
		for count < len(dst) && (bitsLeft > 0 || next < legth) {
			if bitsLeft < 5 {
				if next < legth {
					buffer <<= 8
					buffer |= int(src[next]) & int(0xFF)
					next += 1
					bitsLeft += 8
				} else {
					pad := 5 - bitsLeft
					buffer <<= uint(pad)
					bitsLeft += pad
				}
			}

			index := 0x1F & (buffer >> uint(bitsLeft-5))
			bitsLeft -= 5
			dst[count] = encodeStd[index]
			count += 1
		}
	}

	if count < len(dst) {
		dst[count] = char_zero
	}

	return count
}

func base32Decode(dst []byte, src []byte) int {
	var (
		buffer   = 0
		bitsLeft = 0
		count    = 0
	)

	for i := 0; count < len(dst) && i < len(src); i++ {
		ch := int(src[i])
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' || ch == '-' {
			continue
		}

		buffer <<= 5

		if ch == '0' {
			ch = 'O'
		} else if ch == '1' {
			ch = 'L'
		} else if ch == '8' {
			ch = 'B'
		}

		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
			ch = (ch & 0x1F) - 1
		} else if ch >= '2' && ch <= '7' {
			ch -= '2' - 26
		} else {
			return -1
		}

		buffer |= ch
		bitsLeft += 5

		if bitsLeft >= 8 {
			dst[count] = uint8(buffer >> uint(bitsLeft-8))
			count += 1
			bitsLeft -= 8
		}
	}

	if count < len(dst) {
		dst[count] = char_zero
	}

	return count
}
