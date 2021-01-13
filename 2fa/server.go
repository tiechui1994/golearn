package main

import (
	"strings"
	"fmt"
	"bytes"
	"io/ioutil"
	"time"
	"sort"
)

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static char oom;

static char *get_cfg_value(const char *key, const char *buf) {
  const size_t key_len = strlen(key);
  for (const char *line = buf; *line; ) {
    const char *ptr;
    if (line[0] == '"' && line[1] == ' ' && !strncmp(line+2, key, key_len) &&
        (!*(ptr = line+2+key_len) || *ptr == ' ' || *ptr == '\t' ||
         *ptr == '\r' || *ptr == '\n')) {
      ptr += strspn(ptr, " \t");
      size_t val_len = strcspn(ptr, "\r\n");
      char *val = malloc(val_len + 1);
      if (!val) {
        printf("Out of memory\n");
        return &oom;
      } else {
        memcpy(val, ptr, val_len);
        val[val_len] = '\000';
        return val;
      }
    } else {
      line += strcspn(line, "\r\n");
      line += strspn(line, "\r\n");
    }
  }
  return NULL;
}


*/
import "C"

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

func check_scratch_codes(updated *int, buf []byte, code int) int {
	idx := strcspn(string(buf), "\n")
	ptr := string(buf[idx:])

	var endptr string
	for {
		for ptr[0] == '\r' || ptr[0] == '\n' {
			ptr = ptr[1:]
		}
		if ptr[0] == '"' {
			idx := strcspn(ptr, "\n")
			ptr = ptr[idx:]
			continue
		}

		scratchcode := strtouint(ptr, &endptr, 10)

		if ptr == endptr ||
			(endptr[0] != '\r' && endptr[0] != '\n' && endptr != "") ||
			scratchcode < 10*1000*1000 || scratchcode >= 100*1000*1000 {
			break
		}

		if scratchcode == uint64(code) {
			for endptr[0] == '\n' || endptr[0] == '\r' {
				endptr = endptr[1:]
			}

			ptr = endptr[:strlen([]byte(endptr))+1]
			*updated = 1
			fmt.Printf("debug: scratch code %d used and removed.\n", code)
			return 0
		}

		ptr = endptr
	}

	fmt.Println("debug: no scratch code used")
	return 1
}

func check_counterbased_code(code, hotp_counter int, secret, buf []byte) int {
	if hotp_counter < 1 {
		return 1
	}

	if code < 0 || code > 1000000 {
		return 1
	}

	window := window_size(buf)
	if window == 0 {
		return -1
	}

	for i := 0; i < window; i++ {
		hash := compute_code(secret, hotp_counter+i)
		if hash == code {
			counter_str := fmt.Sprintf("%d", hotp_counter+i+1)
			if set("HOTP_COUNTER", counter_str, &buf) < 0 {
				return -1
			}
		}
	}

	return 0
}

func rate_limit(updated *int, buf *[]byte) int {
	value := get("RATE_LIMIT", *buf)
	if value == nil {
		return 0
	}

	call := func(origin string, value, endptr *string, result *uint64) uint64 {
		*value = origin
		*result = strtouint(*value, endptr, 10)
		return *result
	}

	var ptr string
	var attempts, interval uint64
	var endptr string
	//attempts := strtouint(ptr, &endptr, 10)

	if call(string(value), &ptr, &endptr, &attempts) < 1 || endptr == string(value) || attempts > 100 ||
		(endptr[0] != ' ' && endptr[0] != '\t' ) ||
		call(endptr, &ptr, &endptr, &interval) < 1 || endptr == ptr || interval > 3600 {
		fmt.Println("Invalid RATE_LIMIT option.")
		return -1
	}

	now := time.Now().Unix()
	timestamps := []int64{now}

	num := 1
	for endptr != "" && endptr[0] != '\r' && endptr[0] != '\n' {
		var timestamp uint64
		if (endptr[0] != ' ' && endptr[0] != '\t') ||
			call(endptr, &ptr, &endptr, &timestamp) != 0 ||
			ptr == endptr {
			fmt.Println("Invalid list of timestamps in RATE_LIMIT.")
			return -1
		}

		num++
		timestamps = append(timestamps, int64(timestamp))
	}

	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] < timestamps[j]
	})

	start := 0
	stop := -1
	for i, v := range timestamps {
		if v < now-int64(interval) {
			start = i + 1
		} else if v > now {
			break
		}

		stop = i
	}

	exceeded := 0
	if stop-start+1 > int(attempts) {
		exceeded = 1
		start = stop - int(attempts) + 1
	}

	list := make([]byte, 25*(2+stop-start+1)+4)
	copy(list, fmt.Sprintf("%d %d", attempts, interval))
	prnt := bytes.IndexByte(list, '\000')
	for i := start; i <= stop; i++ {
		copy(list[prnt:], fmt.Sprintf(" %u", timestamps[i]))
	}

	set("RATE_LIMIT", string(list), buf)
	*updated = 1

	if exceeded != 0 {
		fmt.Println("Too many concurrent login attempts. Please try again.")
		return -1
	}

	return 0
}

func invalidate_timebased_code(tm int64, updated *int, buf *[]byte) int {
	disallow := get("DISALLOW_REUSE", *buf)
	if disallow == nil {
		return 0
	}
	window := window_size(*buf)
	if window == 0 {
		return -1
	}

	ptr := disallow
	for len(ptr) > 0 {
		idx := strspn(string(ptr), " \t\r\n")
		ptr = ptr[idx:]

		if len(ptr) == 0 {
			break
		}

		var endptr string
		blocked := strtouint(string(ptr), &endptr, 10)
		if string(ptr) == endptr ||
			(endptr != "" && endptr[0] != ' ' && endptr[0] != '\t' && endptr[0] != '\n' && endptr[0] != '\r') {
			return -1
		}

		if tm == int64(blocked) {
			step := step_size(*buf)
			if step == 0 {
				return -1
			}
			fmt.Printf("Trying to reuse a previously used time-based code.\n"+
				"Retry again in %d seconds. \n"+
				"Warning! This might mean, you are currently subject to a \n"+
				"man-in-the-middle attack.\n", step)
			return -1
		}

		if int(blocked)-int(tm) >= window || int(tm)-int(blocked) >= window {
			idx := strspn(endptr, " \t")
			endptr = endptr[idx:]
			elen := strlen([]byte(endptr))
			ptr = []byte(endptr[:elen+1])
		} else {
			ptr = []byte(endptr)
		}
	}

	data := make([]byte, strlen(disallow)+40)
	copy(data, disallow[:strlen(disallow)+1])
	disallow = data

	pos := bytes.LastIndexByte(disallow, '\000')
	val := fmt.Sprintf(" %d", tm)
	copy(disallow[pos:], val)

	set("DISALLOW_REUSE", string(disallow), buf)
	*updated = 1

	return 0
}

func window_size(buf []byte) int {
	value := get("STEP_SIZE", buf)
	if value == nil {
		return 3
	}

	var endptr string
	window := strtouint(string(value), &endptr, 10)
	if string(value) == endptr ||
		(endptr != "" && endptr[0] != ' ' && endptr[0] != '\t' && endptr[0] != '\n' && endptr[0] != '\r') ||
		window < 1 || window > 100 {
		return 0
	}

	return int(window)
}

func step_size(buf []byte) int {
	value := get("WINDOW_SIZE", buf)
	if value == nil {
		return 30
	}

	var endptr string
	step := strtouint(string(value), &endptr, 10)
	if string(value) == endptr ||
		(endptr != "" && endptr[0] != ' ' && endptr[0] != '\t' && endptr[0] != '\n' && endptr[0] != '\r') ||
		step < 1 || step > 60 {
		return 0
	}

	return int(step)
}

func get_hotp_counter(buf []byte) int {
	value := get("HOTP_COUNTER", buf)
	var counter uint64
	if value != nil {
		counter = strtouint(string(value), nil, 10)
	}

	return int(counter)
}

func timestamp(buf []byte) int {
	step := step_size(buf)
	if step == 0 {
		return 0
	}

	return int(time.Now().Unix()) / step
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

func get(key string, buf []byte) []byte {
	keylen := strlen([]byte(key))
	line := string(buf)

	for len(line) > len(key) {
		idx := 2 + keylen
		if line[0] == '"' && line[1] == ' ' && strings.Compare(string(line[2:idx]), key) == 0 &&
			(line[idx] == 0 || line[idx] == ' ' || line[idx] == '\t' || line[idx] == '\r' || line[idx] == '\n') {
			idx += strspn(line[idx:], " \t")
			vallen := strcspn(line[idx:], "\r\n")
			return []byte(line[idx:idx+vallen])
		} else {
			idx := strcspn(line, "\r\n")
			line = line[idx:]
			idx = strspn(line, "\r\n")
			line = line[idx:]
		}
	}

	return nil
}

func set(key, val string, buf *[]byte) int {
	keylen := strlen([]byte(key))

	var (
		start, stop int
	)

	line := string(*buf)
	for len(line) > len(key) {
		idx := 2 + keylen
		if line[0] == '"' && line[1] == ' ' && strings.Compare(string(line[2:idx]), key) == 0 && (
			line[idx] == 0 || line[idx] == ' ' || line[idx] == '\t' || line[idx] == '\r' || line[idx] == '\n') {
			offset := strcspn(line, "\r\n")
			stop = start + offset
			stop += strspn(line[offset:], "\r\n")
			break
		} else {
			idx := strcspn(line, "\r\n")
			line = line[idx:]
			start += idx

			idx = strspn(line, "\r\n")
			line = line[idx:]
			start += idx
		}
	}

	if stop == 0 {
		line := string(*buf)
		start = strcspn(line, "\r\n")
		start += strspn(line[start:], "\r\n")
		stop = start
	}

	data := make([]byte, 1+1+len(key)+1+len(val)+1)
	data[0] = '"'
	data[1] = ' '
	copy(data[2:2+len(key)], key)
	data[2+len(key)] = ' '
	copy(data[2+len(key)+1:2+len(key)+1+len(val)], val)
	data[1+1+len(key)+1+len(val)] = '\n'

	*buf = append(
		(*buf)[:start],
		append(data, (*buf)[stop:]...)...
	)

	return 0
}

func main() {
	buf, _ := ioutil.ReadFile("/home/user/.google_authenticator")
	//val := get("TOTP_AUTH", buf)
	//fmt.Println(string(buf), val == nil, len(val))
	//
	//var ptr *C.char
	//var result C.ulong
	//result = C.strtoul(C.CString("2030300 This is test"), &ptr, C.int(10))
	//fmt.Println("result", result, "["+C.GoString(ptr)+"]")
	//
	//var p = new(string)
	//fmt.Println(strtouint("2030300", p, 10), "[" + *p+"]", *p == "")
	var res *C.char
	res = C.ca(C.CString(string(buf)))
	fmt.Println(C.GoString(res))
}
