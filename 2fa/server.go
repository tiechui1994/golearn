package main

import (
	"strings"
	"fmt"
	"bytes"
	"time"
	"sort"
	"unsafe"
	"os"
	"io"
	"os/user"
)

const (
	NULLERR        = 0
	NULLOK         = 1
	SECRETNOTFOUND = 2

	PROMPT         = 0
	TRY_FIRST_PASS = 1
	USE_FIRST_PASS = 2
)

const (
	SECRET        = "/home/user/.google_authenticator"
	CODE_PROMPT   = "Verification code: "
	PWCODE_PROMPT = "Password & verification code: "
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

func check_scratch_codes(updated *int, buf []byte, code int, params Params) int {
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

		var errno int
		scratchcode := strtoul(ptr, &endptr, 10, &errno)

		if errno != 0 || ptr == endptr ||
			(endptr[0] != '\r' && endptr[0] != '\n' && endptr != "") ||
			scratchcode < 10*1000*1000 || scratchcode >= 100*1000*1000 {
			break
		}

		if int(scratchcode) == code {
			for endptr[0] == '\n' || endptr[0] == '\r' {
				endptr = endptr[1:]
			}

			ptr = endptr[:strlen([]byte(endptr))+1]
			*updated = 1

			if params.debug {
				fmt.Printf("debug: scratch code %d used and removed.\n", code)
			}

			return 0
		}

		ptr = endptr
	}

	if params.debug {
		fmt.Println("debug: no scratch code used")
	}

	return 1
}

func check_counterbased_code(updated *int, buf *[]byte, secret []byte, seclen, code, hotpcounter int,
	mustadvancecounter *int, param Params) int {
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

	if param.debug {
		fmt.Printf("debug: counter window:%d\n", window)
	}

	for i := 0; i < window; i++ {
		hash := compute_code(secret, seclen, hotpcounter+i)
		if param.debug {
			fmt.Printf("debug window counter code:%v\n", hash)
		}

		if hash == code {
			counter_str := fmt.Sprintf("%d", hotpcounter+i+1)
			set("HOTP_COUNTER", counter_str, buf)
			*updated = 1
			*mustadvancecounter = 0
			return 0
		}
	}

	*mustadvancecounter = 1
	return 1
}

func check_timebased_code(updated *int, buf *[]byte, secret []byte, seclen, code int, param Params) int {
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

	if param.debug {
		fmt.Printf("debug: timestamp:%d\n", tm)
	}

	value := get("TIME_SKEW", *buf)
	var skew int
	if value != nil {
		skew = int(strtoul(string(value), nil, 10))
	}
	if param.debug {
		fmt.Printf("debug: skew:%d\n", skew)
	}

	window := window_size(*buf)
	if window == 0 {
		return -1
	}
	if param.debug {
		fmt.Printf("debug: timer window:%d\n", window)
	}

	for i := -((window - 1) / 2); i <= window/2; i++ {
		hash := compute_code(secret, seclen, tm+skew+i)
		if param.debug {
			fmt.Printf("debug window timer code:%v\n", hash)
		}

		if hash == code {
			return invalidate_timebased_code(tm+skew+i, updated, buf)
		}
	}

	if param.noskewadj != 0 {
		skew = 1000000

		for i := 0; i < 25*60; i++ {
			hash := compute_code(secret, seclen, tm-i)
			if hash == code && skew == 1000000 {
				skew = -i
			}

			hash = compute_code(secret, seclen, tm+i)
			if hash == code && skew == 1000000 {
				skew = i
			}
		}

		if skew != 1000000 {
			if param.debug {
				fmt.Println("debug: time skew adjusted")
			}

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

		if len(ptr) > 0 && ptr[0] == 0 {
			break
		}

		var endptr string
		var errno int
		blocked := strtoul(string(ptr), &endptr, 10, &errno)
		if errno != 0 || string(ptr) == endptr ||
			(endptr[0] != ' ' && endptr[0] != '\t' && endptr[0] != '\n' && endptr[0] != '\r' || endptr != "") {
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
			ptr = []byte(endptr[:strlen([]byte(endptr))+1])
		} else {
			ptr = []byte(endptr)
		}
	}

	data := make([]byte, strlen(disallow)+40)
	copy(data, disallow[:strlen(disallow)])
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
			var errno int
			i := strtoul(ptr, &endptr, 10, &errno)
			if errno != 0 || ptr == endptr || (endptr[0] != '+' && endptr[0] != '-') {
				break
			}
			ptr = endptr
			errno = 0
			j := int(strtoul(ptr[1:], &endptr, 10, &errno))
			if errno != 0 || ptr == endptr ||
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

func or(args ...bool) bool {
	switch len(args) {
	case 0:
		return true
	case 1:
		return args[0]
	default:
		result := args[0]
		for i := 1; i < len(args); i++ {
			result = result || args[i]
		}
		return result
	}
}

func rate_limit(updated *int, buf *[]byte) int {
	value := get("RATE_LIMIT", *buf)
	if value == nil {
		return 0
	}

	// cond || (result = f(value=origin, endptr, error))
	assign := func(origin string, value, endptr *string, result *uint, errno *int) uint {
		*value = origin
		*result = strtoul(*value, endptr, 10, errno)
		return *result
	}

	// cond || (t=exec, expect)
	other := func(exec interface{}, expect int) int {
		return expect
	}

	var ptr string
	var attempts, interval uint
	var errno int
	var endptr string

	if assign(string(value), &ptr, &endptr, &attempts, &errno) < 1 ||
		or(endptr == string(value), attempts > 100, errno != 0, endptr[0] != ' ' && endptr[0] != '\t') ||
		assign(endptr, &ptr, &endptr, &interval, &errno) < 1 ||
		or(endptr == ptr, interval > 3600, errno != 0) {
		fmt.Println("Invalid RATE_LIMIT option.")
		return -1
	}

	now := time.Now().Unix()
	timestamps := []int64{now}

	num := 1
	for endptr != "" && endptr[0] != '\r' && endptr[0] != '\n' {
		var timestamp uint
		if (endptr[0] != ' ' && endptr[0] != '\t') ||
			other(assign(endptr, &ptr, &endptr, &timestamp, &errno), errno) != 0 ||
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
		copy(list[prnt:], fmt.Sprintf(" %d", timestamps[i]))
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
	value := get("WINDOW_SIZE", buf)
	if value == nil {
		return 3
	}

	var endptr string
	window := strtoul(string(value), &endptr, 10)
	if string(value) == endptr ||
		(endptr != "" && endptr[0] != ' ' && endptr[0] != '\t' && endptr[0] != '\n' && endptr[0] != '\r') ||
		window < 1 || window > 100 {
		return 0
	}

	return int(window)
}

func step_size(buf []byte) int {
	value := get("STEP_SIZE", buf)
	if value == nil {
		return 30
	}

	var endptr string
	var errno int
	step := strtoul(string(value), &endptr, 10, &errno)
	if errno != 0 || string(value) == endptr ||
		(endptr != "" && endptr[0] != ' ' && endptr[0] != '\t' && endptr[0] != '\n' && endptr[0] != '\r') ||
		step < 1 || step > 60 {
		return 0
	}

	return int(step)
}

func get_hotp_counter(buf []byte) int {
	value := get("HOTP_COUNTER", buf)
	var counter uint
	if value != nil {
		counter = strtoul(string(value), nil, 10)
	}

	return int(counter)
}

func get_shared_secret(buf *[]byte, secretfile string, seclen *int, param Params) []byte {
	base32Len := strcspn(string(*buf), "\n")
	if base32Len > 100000 {
		return nil
	}

	*seclen = (base32Len*5 + 7) / 8
	secret := make([]byte, base32Len+1)
	copy(secret, (*buf)[:base32Len])
	secret[base32Len] = char_zero

	*seclen = base32Decode(secret, secret, base32Len)
	if *seclen < 1 {
		return nil
	}

	for i := *seclen; i < base32Len+1; i++ {
		secret[i] = 0
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

	var u *user.User
	var secfile string
	var allow_tilde int
	if params.fixed_uid == 0 {
		var err error
		u, err = user.Lookup(username)
		if err != nil {
			goto errout
		}

		if u.HomeDir == "" {
			fmt.Printf("user(\"%s\") has no home dir\n", username)
			goto errout
		}

		if (u.HomeDir)[0] != '/' {
			fmt.Printf("User \"%s\" hone dir not absolare\n", username)
			goto errout
		}
	}

	secfile = spec
	allow_tilde = 1

	for offset := 0; secfile[offset] != 0; {
		cur := secfile[offset:]
		vars := ""
		varslen := 0
		subst := ""
		if allow_tilde != 0 && cur[0] == '~' {
			varslen = 1
			if u == nil {
				fmt.Println("Home dir in  'secret' not  implemented when 'user' set")
				goto errout
			}

			subst = u.HomeDir
			vars = cur
		} else if secfile[offset] == '$' {
			if strings.HasPrefix(cur, "${HOME}") {
				varslen = 7
				if u == nil {
					fmt.Println("Home dir in  'secret' not  implemented when 'user' set")
					goto errout
				}

				subst = u.HomeDir
				vars = cur
			} else if strings.HasPrefix(cur, "${USER}") {
				varslen = 7
				subst = username
				vars = cur
			}

			if vars != "" {
				substlen := strlen([]byte(subst))
				varidx := strings.Index(secfile, vars)
				cp := make([]byte, strlen([]byte(secfile))+substlen)
				copy(cp[varidx+substlen:], vars[varslen:])
				copy(cp[:varidx+substlen], subst)
				offset = varidx + substlen
				allow_tilde = 0
				secfile = string(cp)
			} else {
				allow_tilde = 0

				if cur[0] == '/' {
					allow_tilde = 1
				}

				offset++
			}
		}
	}
	return secfile

errout:
	return ""
}

func timestamp(buf []byte) int {
	step := step_size(buf)
	if step == 0 {
		return 0
	}

	return int(time.Now().Unix()) / step
}

func compute_code(secret []byte, seclen, value int) int {
	var val [8]byte

	equal := func(arg *int, value int) int {
		*arg = value
		return value
	}

	for i := len(val); equal(&i, i-1) >= 0; value >>= 8 {
		val[i] = byte(value)
	}

	var hash = HmacSha1(secret[:seclen], val[:])
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
		buf    []byte
		secret []byte
		seclen int

		rc = PAM_AUTH_ERR
	)

	var prompt string
	if param.authtok_propt != "" {
		prompt = param.authtok_propt
	} else {
		prompt = CODE_PROMPT
		if param.forward_pass != 0 {
			prompt = PWCODE_PROMPT
		}
	}

	var earlyupdated, updated = 0, 0
	username := username()
	secretfile := getSecretFile(username, param)
	stoppedbyratelimit := 0

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
		defer fd.Close()

		stat, _ := fd.Stat()
		if param.debug {
			fmt.Printf("debug: Secret file permissions are %04o.\n"+
				" Allowed permissions are %04o\n", stat.Mode()&03777, param.allowed_perm)
		}

		buf = make([]byte, stat.Size()+1)
		var writer bytes.Buffer
		io.CopyN(&writer, fd, stat.Size())
		copy(buf, writer.Bytes())
		buf[stat.Size()] = char_zero
	}

	if buf != nil && len(buf) > 0 {
		if rate_limit(&earlyupdated, &buf) >= 0 {
			secret = get_shared_secret(&buf, secretfile, &seclen, param)
		} else {
			stoppedbyratelimit = 1
		}
	}

	hotp_counter := get_hotp_counter(buf)
	if stoppedbyratelimit == 0 && (secret != nil || param.nullok != SECRETNOTFOUND) {
		if secret == nil || len(secret) == 0 {
			fmt.Printf("No secret configured for user %s, asking for code anyway.\n", username)
		}

		var (
			mustadvancecounter = 0
			pw, savedpw        []byte
		)

		// cond || (arg=value) < xx
		equal := func(arg *byte, value byte) byte {
			*arg = value
			return *arg
		}
		// condition ? xxx : xxx
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
			case 0, 1:
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
						fmt.Printf("debug savedpw:[%s]\n", string(pw))
					}
					if savedpw != nil {
						pw = make([]byte, len(savedpw))
						copy(pw, savedpw)
					}
				}
			}

			if pw == nil {
				continue
			}

			pwlen := strlen(pw)
			expectedlen := 6
			if mode&1 != 0 {
				expectedlen = 8
			}

			if param.debug {
				fmt.Printf("debug pw:[%s] savedpw:[%s], mode:%d\n", string(pw),
					string(savedpw), mode)
			}

			if pwlen > 0 && pw[0] == '\b' {
				fmt.Println("Dummy password supplied by PAM." +
					" Did OpenSSH 'PermitRootLogin <anything but yes>' or some" +
					" other config block this login?")
			}

			var ch byte
			if pwlen < expectedlen || equal(&ch, pw[pwlen-expectedlen]) > '9' ||
				ch < expr(expectedlen == 8, '1', '0') {
				pw = nil
				continue
			}

			fmt.Printf("debug pwlen:%d expectedlen:%d\n", pwlen, expectedlen)

			var endptr string
			var errno int
			l := strtoul(string(pw[pwlen-expectedlen:]), &endptr, 10, &errno)
			if errno != 0 || l < 0 || endptr != "" {
				pw = nil
				continue
			}

			code := int(l)
			if param.debug {
				fmt.Printf("debug code: [%d]\n", code)
			}
			memset(pw[pwlen-expectedlen:], 0, expectedlen)

			if (mode == 2 || mode == 3) && param.forward_pass == 0 {
				if len(pw) > 0 && pw[0] != 0 {
					pw = nil
					continue
				}
			}

			if secret != nil {
				switch check_scratch_codes(&updated, buf, code, param) {
				case 1:
					if hotp_counter > 0 {
						switch check_counterbased_code(&updated, &buf, secret, seclen, code,
							hotp_counter, &mustadvancecounter, param) {
						case 0:
							rc = PAM_SUCCESS
						case 1:
							pw = nil
							continue
						default:
						}
					} else {
						switch check_timebased_code(&updated, &buf, secret, seclen, code, param) {
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
		}

		if rc == PAM_SUCCESS && param.forward_pass != 0 {
			if pw == nil {
				rc = PAM_AUTH_ERR
			}
		}

		if param.no_increment_hotp == 0 && mustadvancecounter != 0 {
			counter := fmt.Sprintf("%d", hotp_counter+1)
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

	if param.nullok == SECRETNOTFOUND {
		rc = PAM_IGNORE
	}

	if earlyupdated != 0 || updated != 0 {
		// write all
		fmt.Println("write", string(buf))

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
