package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"os"
	"os/user"
	"strings"
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

const (
	ULONG_MAX = 0xFFFFFFFF
)

func strtoul(s string, endptr *string, base uint, errno ...*int) uint {
	isspace := func(ch byte) bool {
		return ch == '\t' || ch == '\n' || ch == '\v' || ch == '\f' || ch == '\r' || ch == ' '
	}
	isdigit := func(ch byte) bool {
		return ch >= '0' && ch <= '9'
	}
	isalpha := func(ch byte) bool {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
	}
	isupper := func(ch byte) bool {
		return 'A' <= ch && ch <= 'Z'
	}

	// char c = *s++; // c=*s; s++
	bptrplus := func(c *byte, idx *int, data []byte) byte {
		if *idx < len(data) {
			*c = data[*idx]
		} else {
			*c = char_zero
		}

		*idx += 1
		return *c
	}

	var (
		c   byte
		acc uint
		any int

		// 指针的替换
		idx  int
		data = []byte(s)
	)

	c = data[0]
	idx = 1
	for isspace(c) {
		c = bptrplus(&c, &idx, data)
	}

	neg := 0
	if c == '-' {
		neg = 1
		c = bptrplus(&c, &idx, data)
	} else if c == '+' {
		c = bptrplus(&c, &idx, data)
	}

	if (base == 0 || base == 16) && c == '0' && (data[idx] == 'x' || data[idx] == 'X') {
		c = data[idx+1]
		idx += 2
		base = 16
	}

	if base == 0 {
		base = 10
		if c == '0' {
			base = 8
		}
	}

	cutoff := ULONG_MAX / base
	cutlim := ULONG_MAX % base

	i := 0
	for acc, any = 0, 0; ; c = bptrplus(&c, &idx, data) {
		i++
		if isdigit(c) {
			c -= '0'
		} else if isalpha(c) {
			if isupper(c) {
				c -= 'A' - 10
			} else {
				c -= 'a' - 10
			}
		} else {
			break
		}

		if uint(c) >= base {
			break
		}

		if any < 0 || acc > cutoff || (acc == cutoff && uint(c) > cutlim) {
			any = -1
		} else {
			any = 1
			acc *= base
			acc += uint(c)
		}
	}

	if any < 0 {
		acc = ULONG_MAX
		if len(errno) > 0 {
			*(errno[0]) = 1
		}
	} else if neg != 0 {
		acc = -acc
	}

	if endptr != nil {
		if any != 0 {
			*endptr = string(data[idx-1:])
		} else {
			*endptr = string(s)
		}
	}

	return acc
}

func GenerateCode(key string, tm int) int {
	var challenge [8]byte
	for i := len(challenge) - 1; i >= 0; tm >>= 8 {
		challenge[i] = byte(tm)
		i--
	}

	secretLen := (strlen([]byte(key)) + 7) / 8 * BITS_PER_BASE32_CHAR
	if secretLen <= 0 || secretLen > 100 {
		return -1
	}

	var secret [100]byte
	secretLen = base32Decode([]byte(key), secret[:], secretLen)
	if secretLen < 1 {
		return -1
	}

	hash := HmacSha1(secret[:secretLen], challenge[:])

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

func base32Encode(data []byte, length int, result []byte, bufsize int) int {
	if length < 0 || length > (1<<28) {
		return -1
	}

	var count = 0
	if length > 0 {
		encodeStd := "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
		buffer := int(data[0])
		next := 1
		bitsLeft := 8
		for count < bufsize && (bitsLeft > 0 || next < length) {
			if bitsLeft < 5 {
				if next < length {
					buffer <<= 8
					buffer |= int(data[next]) & int(0xFF)
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
			result[count] = encodeStd[index]
			count += 1
		}
	}

	if count < bufsize {
		result[count] = char_zero
	}

	return count
}

func base32Decode(encoded []byte, result []byte, bufsize int) int {
	var (
		buffer   = 0
		bitsLeft = 0
		count    = 0
	)

	for i := 0; count < bufsize && encoded[i] != 0; i++ {
		ch := int(encoded[i])
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
			result[count] = byte(buffer >> uint(bitsLeft-8))
			count += 1
			bitsLeft -= 8
		}
	}

	if count < bufsize {
		result[count] = char_zero
	}

	return count
}

func hostname() string {
	hostname, err := os.Hostname()
	if err == nil {
		return hostname
	}
	return "unix"
}

func username() string {
	userinfo, err := user.Current()
	if err == nil {
		return userinfo.Username
	}

	return ""
}

func urlencode(s string) string {
	size := 3*strlen([]byte(s)) + 1
	if size > 10000 {
		fmt.Fprintln(os.Stderr, "Error: Generated URL would be unreasonably large.")
		os.Exit(1)
	}

	ret := make([]byte, size)
	idx := 0

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '%', '&', '?', '=':
			copy(ret[idx:], fmt.Sprintf("%%%02X", s[i]))
			idx += 3
		default:
			if (s[i] != 0 && s[i] <= ' ') || (s[i] >= '\x7F') {
				copy(ret[idx:], fmt.Sprintf("%%%02X", s[i]))
				idx += 3
				break
			}

			ret[idx] = s[i]
			idx += 1
		}
	}

	return string(ret[:strlen(ret)+1])
}

func memset(data []byte, value byte, length int) {
	for i := 0; i < length && i < len(data); i++ {
		data[i] = value
	}
}

func ask(msg string) string {
	fmt.Printf("%s ", msg)
	var read string
	fmt.Scanln(&read)
	return read
}
