package main

import (
	"strings"
	"fmt"
	"bytes"
	"io/ioutil"
	"time"
	"sort"
	"unsafe"
	"os"
)

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static char oom;

int test() {
	int reset_size = 20;
	char reset[reset_size];
    *reset = '\000';

    char* pos = strrchr(reset, '\000');
	snprintf(pos, reset_size-(pos-reset), " %d%+d" + !*reset, 100, 200);

	printf("[%s]\n",reset);
	printf("[%s]\n",reset);
	printf("[%+d]\n",0);
	printf("[%+d]\n",1);
	printf("[%+d]\n",-1);
	return 0;
}
//static char *get_cfg_value(const char *key, const char *buf) {
//  const size_t key_len = strlen(key);
//  for (const char *line = buf; *line; ) {
//    const char *ptr;
//    if (line[0] == '"' && line[1] == ' ' && !strncmp(line+2, key, key_len) &&
//        (!*(ptr = line+2+key_len) || *ptr == ' ' || *ptr == '\t' ||
//         *ptr == '\r' || *ptr == '\n')) {
//      ptr += strspn(ptr, " \t");
//      size_t val_len = strcspn(ptr, "\r\n");
//      char *val = malloc(val_len + 1);
//      if (!val) {
//        printf("Out of memory\n");
//        return &oom;
//      } else {
//        memcpy(val, ptr, val_len);
//        val[val_len] = '\000';
//        return val;
//      }
//    } else {
//      line += strcspn(line, "\r\n");
//      line += strspn(line, "\r\n");
//    }
//  }
//  return NULL;
//}
*/
import "C"

const (
	NULLERR        = 0
	NULLOK         = 1
	SECRETNOTFOUND = 2

	PROMPT         = 0
	TRY_FIRST_PASS = 1
	USE_FIRST_PASS = 2

	SECRET = "/home/user/.google_authenticator"
)

type Params struct {
	secret_filename_spec string
	authtok_propt        string
	nullok               int
	noskewadj            int
	echocode             int
	fixed_uid            int
	no_increment_hotp    int
	uid                  int
	pass_mode            int
	forward_pass         int
	no_strict_owner      int
	allowed_perm         int
	grace_period         int64
	allow_readonly       int
	debug                bool
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

func check_counterbased_code(updated *int, buf *[]byte, secret []byte, code, hotpcounter int, mustadvancecounter *int) int {
	if hotpcounter < 1 {
		return 1
	}

	if code < 0 || code > 1000000 {
		return 1
	}

	window := window_size(*buf)
	if window == 0 {
		return -1
	}

	for i := 0; i < window; i++ {
		hash := compute_code(secret, hotpcounter+i)
		if hash == code {
			counter_str := fmt.Sprintf("%d", hotpcounter+i+1)
			set("HOTP_COUNTER", counter_str, buf)
			*mustadvancecounter = 0
			*updated = 1
			return 0
		}
	}

	*mustadvancecounter = 1
	return 1
}

func check_timebased_code(updated *int, buf *[]byte, secret []byte, code int, param Params) int {
	if is_top(*buf) {
		return 1
	}

	if code < 0 || code >= 1000000 {
		return 1
	}

	tm := timestamp(*buf)
	if tm == 0 {
		return -1
	}

	value := get("TIME_SKEW", *buf)
	var skew int
	if value != nil {
		skew = int(strtouint(string(value), nil, 10))
	}

	window := window_size(*buf)
	if window == 0 {
		return -1
	}

	for i := -((window - 1) / 2); i <= window/2; i++ {
		hash := compute_code(secret, tm+skew+i)
		if hash == code {
			return invalidate_timebased_code(tm+skew+i, updated, buf)
		}
	}

	if param.noskewadj != 0 {
		skew = 1000000

		for i := 0; i < 25*60; i++ {
			hash := compute_code(secret, tm-i)
			if hash == code && skew == 1000000 {
				skew = -i
			}

			hash = compute_code(secret, tm+i)
			if hash == code && skew == 1000000 {
				skew = i
			}
		}

		if skew != 1000000 {
			fmt.Println("debug: time skew adjusted")
			return check_time_skew(updated, buf, skew, tm)
		}
	}

	return 1
}

func invalidate_timebased_code(tm int, updated *int, buf *[]byte) int {
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

		if tm == int(blocked) {
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

	pos := bytes.LastIndexByte(disallow, char_zero)
	val := fmt.Sprintf(" %d", tm)
	copy(disallow[pos:], val)

	set("DISALLOW_REUSE", string(disallow), buf)
	*updated = 1

	return 0
}

func check_time_skew(updated *int, buf *[]byte, skew, tm int) int {
	rc := -1

	resetting := get("RESETTING_TIME_SKEW", *buf)
	var tms [3]uint
	const skewSize = int(unsafe.Sizeof(tms) / unsafe.Sizeof(int(0)))
	var skews [skewSize]int

	num_entries := 0
	if resetting != nil {
		ptr := string(resetting)

		for ptr != "" && ptr[0] != '\r' && ptr[0] != '\n' {
			var endptr string
			i := uint(strtouint(ptr, &endptr, 10))
			if ptr == endptr || (endptr[0] != '+' && endptr[0] != '-') {
				break
			}
			ptr = endptr
			j := int(strtouint(ptr[1:], &endptr, 10))
			if ptr == endptr ||
				(endptr[0] != ' ' && endptr[0] != '\t' && endptr[0] != '\r' && endptr[0] != '\n' && endptr != "") {
				break
			}

			if ptr[0] == '-' {
				j = -j
			}

			if num_entries == skewSize {
				offset := int(unsafe.Sizeof(tms) - unsafe.Sizeof(int(0)))
				copy(tms[:], tms[1:1+offset])
				offset = int(unsafe.Sizeof(skews) - unsafe.Sizeof(int(0)))
				copy(skews[:], skews[1:1+offset])
			} else {
				num_entries += 1
			}

			tms[num_entries-1] = i
			skews[num_entries-1] = j
			ptr = endptr
		}

		if num_entries != 0 && tm+skew == int(tms[num_entries-1])+skews[num_entries-1] {
			return -1
		}
	}

	if num_entries == skewSize {
		offset := int(unsafe.Sizeof(tms) - unsafe.Sizeof(int(0)))
		copy(tms[:], tms[1:1+offset])
		offset = int(unsafe.Sizeof(skews) - unsafe.Sizeof(int(0)))
		copy(skews[:], skews[1:1+offset])
	} else {
		num_entries += 1
	}

	tms[num_entries-1] = uint(tm)
	skews[num_entries-1] = skew

	if num_entries == skewSize {
		last_tm := tms[0]
		last_skew := skews[0]
		avg_skew := last_skew
		for i := 0; i < skewSize; i++ {
			if tms[i] <= last_tm || tms[i] > last_tm+2 ||
				last_skew-skew < -1 || last_skew-skew > 1 {
				goto keep_trying
			}
			last_tm = tms[i]
			last_skew = skews[i]
			avg_skew += last_skew

		}

		avg_skew /= skewSize
		time_skew := fmt.Sprintf("%d", avg_skew)
		set("TIME_SKEW", time_skew, buf)
		rc = 0
	}

keep_trying:
	const reset_size = 80 * skewSize
	var reset [reset_size]byte
	reset[0] = char_zero

	if rc != 0 {
		for i := 0; i < num_entries; i++ {
			pos := strings.LastIndexByte(string(reset[:]), char_zero)
			copy(reset[pos:], fmt.Sprintf(" %d%+d", tms[i], skews[i]))
		}
	}

	set("RESETTING_TIME_SKEW", string(reset[:]), buf)
	*updated = 1
	return rc
}

func rate_limit(updated *int, buf *[]byte) int {
	value := get("RATE_LIMIT", *buf)
	if value == nil {
		return 0
	}

	fmt.Println("value:", string(value))

	call := func(origin string, value, endptr *string, result *uint, errno *int) uint {
		*value = origin
		*result = strtoul(*value, endptr, 10, errno)
		fmt.Printf("result:%v, endptr:%v\n", *result, *endptr)
		return *result
	}

	var ptr string
	var attempts, interval uint
	var errno int
	var endptr string

	if call(string(value), &ptr, &endptr, &attempts, &errno) < 1 || endptr == string(value) ||
		attempts > 100 || errno != 0 ||
		(endptr[0] != ' ' && endptr[0] != '\t' ) ||
		call(endptr, &ptr, &endptr, &interval, &errno) < 1 || endptr == ptr ||
		interval > 3600 || errno != 0 {
		fmt.Println("Invalid RATE_LIMIT option.")
		return -1
	}

	now := time.Now().Unix()
	timestamps := []int64{now}

	num := 1
	for endptr != "" && endptr[0] != '\r' && endptr[0] != '\n' {
		var timestamp uint
		if (endptr[0] != ' ' && endptr[0] != '\t') ||
			call(endptr, &ptr, &endptr, &timestamp, &errno) != 0 ||
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
	prnt := bytes.IndexByte(list, char_zero)
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

func get_shared_secret(buf *[]byte, secretfile string, param Params) []byte {
	base32Len := strcspn(string(*buf), "\n")
	if base32Len > 100000 {
		return nil
	}

	secretLen := (base32Len*5 + 7) / 8
	secret := make([]byte, secretLen)
	copy(secret, (*buf)[:base32Len])
	secret[base32Len] = char_zero

	secretLen = base32Decode(secret, secret, base32Len)
	if secretLen < 1 {
		return nil
	}

	if param.debug {
		fmt.Printf("debug: shared secret in \"%s\" processed\n", secretfile)
	}

	return secret
}

func getSecretFile(username string, params Params) string {
	if username == "" {
		return ""
	}

	spec := SECRET
	if params.secret_filename_spec != "" {
		spec = params.secret_filename_spec
	}

	return spec
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

func is_top(buf []byte) bool {
	return strings.Index(string(buf), `" TOTP_AUTH`) == -1
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

const (
	PAM_AUTH_ERR = 1
	PAM_SUCCESS  = 2
	PAM_IGNORE   = 3
)

func run(argcode string) int {
	param := Params{
		nullok:       1,
		debug:        true,
		allowed_perm: 0600,
	}

	var (
		buf                   []byte
		secret                []byte
		earlyupdated, updated int
		stoppedbyratelimit    int

		rc         = PAM_AUTH_ERR
		username   = username()
		secretfile = getSecretFile(username, param)
	)

	if secretfile != "" {
		fd, err := os.OpenFile(secretfile, os.O_RDONLY, 0)
		if err != nil && os.IsNotExist(err) {
			if param.nullok != NULLERR {
				param.nullok = SECRETNOTFOUND
			} else {
				fmt.Printf("Failed to read \"%s\" for \"%s\": %s\n", secretfile,
					username, err.Error())
			}
		}

		if param.debug {
			stat, _ := fd.Stat()
			fmt.Printf("debug: Secret file permissions are %04o.\n"+
				" Allowed permissions are %04o\n", stat.Mode()&03777, param.allowed_perm)
		}

		buf, _ = ioutil.ReadAll(fd)
	}

	if buf != nil && len(buf) > 0 {
		if rate_limit(&earlyupdated, &buf) >= 0 {
			secret = get_shared_secret(&buf, secretfile, param)
		} else {
			stoppedbyratelimit = 1
		}
	}

	hotp_counter := get_hotp_counter(buf)
	if stoppedbyratelimit == 0 && (secret != nil || param.nullok != SECRETNOTFOUND) {
		if secret == nil {
			fmt.Printf("No secret configured for user %s, asking for code anyway.\n", username)
		}

		var (
			must_advance_counter = 0
			pw, savedpw          []byte
		)

		assign := func(arg *byte, value byte) byte {
			*arg = value
			return *arg
		}
		expr := func(cond bool, t, f byte) byte {
			if cond {
				return t
			}

			return f
		}

		for mode := 0; mode < 4; mode++ {
			if updated != 0 || pw != nil {
				if pw != nil {
					pw = nil
				}
				rc = PAM_AUTH_ERR
				break
			}

			switch mode {
			case 0:
			case 1:
				if param.pass_mode == USE_FIRST_PASS || param.pass_mode == TRY_FIRST_PASS {
					pw = []byte(argcode)
				}
			default:
				if mode != 2 && mode != 3 {
					rc = PAM_AUTH_ERR
					continue
				}

				if param.pass_mode == PROMPT || param.pass_mode == TRY_FIRST_PASS {
					if savedpw == nil {
						savedpw = []byte(argcode)
					}
					if savedpw != nil {
						pw = make([]byte, len(savedpw))
						copy(pw, savedpw)
					}
				}

				break
			}

			if pw == nil {
				continue
			}

			pwlen := strlen(pw)
			expectedlen := 6
			if mode&1 != 0 {
				expectedlen = 8
			}

			if pwlen > 0 && pw[0] == '\b' {
				fmt.Println("Dummy password supplied by PAM." +
					" Did OpenSSH 'PermitRootLogin <anything but yes>' or some" +
					" other config block this login?")
			}

			var ch byte
			if pwlen < expectedlen || assign(&ch, pw[pwlen-expectedlen]) > '9' ||
				ch < expr(expectedlen == 8, '1', '0') {
				pw = nil
				continue
			}

			var endptr string
			l := strtouint(string(pw[pwlen-expectedlen:]), &endptr, 10)
			if l < 0 || endptr != "" {
				pw = nil
				continue
			}

			code := int(l)

			if (mode == 2 || mode == 3) && param.forward_pass == 0 {
				if pw != nil {
					pw = nil
					continue
				}
			}

			if secret != nil {
				switch check_scratch_codes(&updated, buf, code) {
				case 1:
					if hotp_counter > 0 {
						switch check_counterbased_code(&updated, &buf, secret, code, hotp_counter, &must_advance_counter) {
						case 0:
							rc = PAM_SUCCESS
						case 1:
							pw = nil
							continue
						default:
						}
					} else {
						switch check_timebased_code(&updated, &buf, secret, code, param) {
						case 0:
							rc = PAM_SUCCESS
						case 1:
							pw = nil
							continue
						default:
						}
					}

				case 0:
					rc = PAM_SUCCESS
				default:
				}

				break
			}

			if rc == PAM_SUCCESS && param.forward_pass != 0 {
				if pw == nil {
					rc = PAM_AUTH_ERR
				}
			}

			if param.no_increment_hotp == 0 && must_advance_counter != 0 {
				counter := fmt.Sprintf("%ld", hotp_counter+1)
				set("HOTP_COUNTER", counter, &buf)
				updated = 1
			}

			if rc == PAM_SUCCESS {
				fmt.Println("Accepted google_authenticator for " + username)
				if param.grace_period != 0 {

				}
			} else {
				fmt.Println("Invalid verification code for " + username)
			}
		}

	}

	if param.nullok == SECRETNOTFOUND {
		rc = PAM_IGNORE
	}

	if earlyupdated != 0 || updated != 0 {
		// write all

		if param.allow_readonly != 0 {
			rc = PAM_AUTH_ERR
		}
	}

	if param.debug {
		fmt.Printf("debug: end of google_authenticator for \"%s\". Result: %d\n", username, rc)
	}

	return rc
}

func main() {
	for {
		fmt.Printf("%s ", "Verification code:")
		var read string
		fmt.Scanln(&read)
		if run(read) == PAM_SUCCESS {
			break
		}
	}

}
