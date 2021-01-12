package main

import (
	"fmt"
	"strings"
	"os"
)


const (
	NULLERR        = 0
	NULLOK         = 1
	SECRETNOTFOUND = 2

	PROMPT         = 0
	TRY_FIRST_PASS = 1
	USE_FIRST_PASS = 2
)

type Params struct {
	authtok_propt     []byte
	nullok            int
	noskewadj         int
	echocode          int
	fixed_uid         int
	no_increment_hotp int
	uid               int
	pass_mode         int
	forward_pass      int
	no_strict_owner   int
	allowed_perm      int
	grace_period      int64
	allow_readonly    int
}

func check_counterbased_code(secretfile string, code, hotp_counter int, secret, buf []byte) int {
	if hotp_counter < 1 {
		return 1
	}

	if code < 0 || code > 1000000 {
		return 1
	}

	window := window_size(secretfile, buf)
	if window == 0 {
		return -1
	}

	for i := 0; i < window; i++ {
		hash := compute_code(secret, hotp_counter+i)
		if hash == code {
			counter_str := fmt.Sprintf("%d", hotp_counter+i+1)
			if set_cfg_value("HOTP_COUNTER", counter_str, buf) < 0 {
				return -1
			}
		}
	}

	return 0
}

func window_size(secretfile string, buf []byte) int {
	return 1
}

func compute_code(secret []byte, value int) int {
	var val [8]byte

	for i := len(val) - 1; i >= 0; value >>= 8 {
		val[i] = byte(value)
	}

	var hash = HmacSha1(secret, val[:])
	offset := hash[SHA1_DIGEST_LENGTH-1] & 0xF

	truncatedHash := 0
	for i := 0; i < 4; i++ {
		truncatedHash <<= 8
		truncatedHash |= int(hash[int(offset)+i])
	}

	truncatedHash &= 0x7FFFFFFF
	truncatedHash %= 1000000
	return truncatedHash
}

func get_cfg_value(key string, buf []byte) []byte {
	key_len := strlen([]byte(key))
	line := buf

	for len(line) > 0 {
		if line[0] == '"' && line[1] == ' ' && strings.Compare(string(line[2:2+key_len]), key) == 0 {
			idx := 2 + key_len
			char := line[idx]
			if char == 0 || char == ' ' || char == '\t' || char == '\r' || char == '\n' {
				idx += strspn(string(line[idx:]), " \t")
				val_len := strcspn(string(line[idx:]), "\r\n")
				val := line[idx:idx+val_len]
				return val
			}
		} else {
			idx := strcspn(string(line), "\r\n")
			line = line[idx:]
			idx += strspn(string(line), "\r\n")
			line = line[idx:]
		}
	}

	return nil
}

func set_cfg_value(key, val string, buf []byte) int {
	key_len := strlen([]byte(key))
	line := buf

	var (
		start, stop int
	)
	for len(line) > 0 {
		if line[0] == '"' && line[1] == ' ' && strings.Compare(string(line[2:2+key_len]), key) == 0 {
			idx := 2 + key_len
			char := line[idx]
			if char == 0 || char == ' ' || char == '\t' || char == '\r' || char == '\n' {
				start = 0
				stop = start + strcspn(string(line[start:]), "\r\n")
				stop += strspn(string(line[stop:]), "\r\n")
				break
			}
		} else {
			idx := strcspn(string(line), "\r\n")
			line = line[idx:]
			idx += strspn(string(line), "\r\n")
			line = line[idx:]
		}
	}

	if start == 0 && stop == 0 {
		start = strcspn(string(buf), "\r\n")
		start += strspn(string(buf[start:]), "\r\n")
		stop = start
	}

	val_len := strlen([]byte(val))
	total_len := key_len + val_len + 4

	_ = total_len

	return -1
}

func main() {
	for _, v := range os.Args {
		fmt.Println(v)
	}
}