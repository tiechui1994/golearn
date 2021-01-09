package main

import "fmt"

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

func set_cfg_value(key, val string, buf []byte) int {
	return -1
}
