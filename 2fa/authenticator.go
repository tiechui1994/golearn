package main

import (
	"crypto/sha1"
	"crypto/hmac"
	"os"
	"log"
	"os/user"
	"fmt"
	"net/url"
	"time"
	"unsafe"
	"strconv"
)
/*
#include <getopt.h>
#include <stdio.h>

int callparam(struct option *longopts, char* const argv[]) {
	printf("argv %s\n", argv[0]);
	int i = 0;
	for (struct option* ptr=longopts; ptr; ++ptr) {
		struct option val = *ptr;
		printf("has_arg %d\n", val.has_arg);
		printf("flag %d\n", *val.flag);
		printf("val %d\n", val.val);
		printf("name %s\n", val.name);
		i++;
		if (i==21) {
			return val.val;
		}
	}

	return sizeof(longopts) ;
}
*/
import "C"

const (
	ASK_MODE  = 0
	HOTP_MODE = 1
	TOTP_MODE = 2
)

const (
	ASK_REUSE      = 0
	DISALLOW_REUSE = 1
	ALLOW_REUSE    = 2
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

func GenerateCode(key []byte, tm int) int {
	var challenge [8]uint8
	for i := len(challenge) - 1; i >= 0; tm >>= 8 {
		challenge[i] = uint8(tm)
		i--
	}

	secretLen := (len(key) + 7) / 8 * BITS_PER_BASE32_CHAR

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
		dst[count] = '\000'
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
		dst[count] = '\000'
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

func getLabel() string {
	userinfo, err := user.Current()
	if err == nil {
		return userinfo.Username + "@" + hostname()
	}

	return ""
}

func display(secret []byte, label string, use_totop int, issuer string) {
	totop := "t"
	if use_totop == HOTP_MODE {
		totop = "h"
	}

	log.Println(issuer, len(secret))

	value := "secret=" + string(secret)

	log.Println(issuer != "")
	if issuer != "" {
		value = fmt.Sprintf("%v&issuer=%v", value, issuer)
	}

	u := fmt.Sprintf("otpauth://%votp/%s?%s", totop, string(label), value)
	u = "https://www.google.com/chart?chs=300x300&chld=M|0&cht=qr&chl=" + url.PathEscape(u)
	log.Println(u)
}

/*
struct option {
    const char *name;
    int         has_arg;
    int        *flag;
    int         val;
}
*/
type option struct {
	name    *C.char
	has_arg C.int
	flag    *C.int
	val     C.int
}

func usage() {
	fmt.Println(
		`google-authenticator [<options>]
 -h, --help                     Print this message
     --version                  Print version
 -c, --counter-based            Set up counter-based (HOTP) verification
 -C, --no-confirm               Don't confirm code. For non-interactive setups
 -t, --time-based               Set up time-based (TOTP) verification
 -d, --disallow-reuse           Disallow reuse of previously used TOTP tokens
 -D, --allow-reuse              Allow reuse of previously used TOTP tokens
 -f, --force                    Write file without first confirming with user
 -l, --label=<label>            Override the default label in "otpauth://" URL
 -i, --issuer=<issuer>          Override the default issuer in "otpauth://" URL
 -q, --quiet                    Quiet mode
 -Q, --qr-mode={NONE,ANSI,UTF8} QRCode output mode
 -r, --rate-limit=N             Limit logins to N per every M seconds
 -R, --rate-time=M              Limit logins to N per every M seconds
 -u, --no-rate-limit            Disable rate-limiting
 -s, --secret=<file>            Specify a non-standard file location
 -S, --step-size=S              Set interval between token refreshes
 -w, --window-size=W            Set window of concurrently valid codes
 -W, --minimal-window           Disable window of concurrently valid codes
 -e, --emergency-codes=N        Number of emergency codes to generate`)
}

func dec(val *int32) int32 {
	res := *val
	*val -= 1
	return res
}

func maybe(msg string) int {
	fmt.Println()
	for {
		fmt.Printf("%s (y/n) ", msg)

		var read string
		fmt.Scanln(&read)
		switch read {
		case "Y", "y":
			return 1
		case "N", "n":
			return 0
		}
	}
}

func ask(msg string) string {
	fmt.Printf("%s ", msg)
	var read string
	fmt.Scanln(&read)
	return read
}

func getParams() {
	optstring := "+hcCtdDfl:i:r:R:us:S:w:We:"

	options := [][]interface{}{
		{"help", 0, 0, 'h'},
		{"version", 0, 0, int32(0)},
		{"counter-based", 0, 0, 'c'},
		{"no-confirm", 0, 0, 'C'},
		{"time-based", 0, 0, 't'},
		{"disallow-reuse", 0, 0, 'd'},
		{"allow-reuse", 0, 0, 'D'},
		{"force", 0, 0, 'f'},
		{"label", 1, 0, 'l'},
		{"issuer", 1, 0, 'i'},
		{"quiet", 0, 0, 'q'},
		{"rate-limit", 1, 0, 'r'},
		{"rate-time", 1, 0, 'R'},
		{"no-rate-limit", 0, 0, 'u'},
		{"secret", 1, 0, 's'},
		{"step-size", 1, 0, 'S'},
		{"window-size", 1, 0, 'w'},
		{"minimal-window", 0, 0, 'W'},
		{"emergency-codes", 1, 0, 'e'},
		{0, 0, 0, int32(0)},
	}

	/*
	int getopt_long(int argc, char* const argv[],
           const char *optstring,
           const struct option *longopts, int *longindex);
	*/
	N := len(options)
	size := N * int(unsafe.Sizeof(option{}))
	opts := (*C.struct_option)(C.malloc(C.size_t(size)))
	optsptr := (*[1024]C.struct_option)(unsafe.Pointer(opts))[:N:N]
	for i := range options {
		name, ok := options[i][0].(string)
		flag := options[i][2].(int)
		opt := option{
			has_arg: C.int(options[i][1].(int)),
			flag:    (*C.int)(unsafe.Pointer(&flag)),
			val:     C.int(options[i][3].(int32)),
		}
		if ok {
			opt.name = C.CString(name)
		}
		optsptr[i] = *(*C.struct_option)(unsafe.Pointer(&opt))
	}

	N = len(os.Args)
	size = N * 8
	argv := (**C.char)(C.malloc(C.size_t(size)))
	argptr := (*[1024]*C.char)(unsafe.Pointer(argv))[:N:N]
	for i := range os.Args {
		argptr[i] = C.CString(os.Args[i])
	}

	var idx int32 = -1

	mode := ASK_MODE
	reuse := ALLOW_REUSE
	confirm := 1
	force := 0
	label := ""
	issuer := ""
	quite := 0
	r_limit := 0
	r_time := 0
	step_size := 0
	window_size := 0
	emergency_codes := 0
	secret_fn := ""

	version := func() {
		fmt.Println("google-authenticator 1.0.0")
	}

	success := func() {
		os.Exit(0)
	}
	failed := func(msg string) {
		fmt.Println(msg)
		os.Exit(1)
	}
	dump_err := func(args ...string) {
		switch len(args) {
		case 1:
			fmt.Printf("Duplicate %v option detected\n", args[0])
		case 2:
			fmt.Printf("Duplicate %v and/or %v option detected\n", args[0], args[1])
		}

		os.Exit(1)
	}
	exclusive_err := func(arg1, arg2 string) {
		fmt.Println(arg1 + " is mutually exclusive with " + arg2)
		os.Exit(1)
	}
	range_err := func(arg, ranges string) {
		fmt.Println(arg + " requires an argument in the range " + ranges)
		os.Exit(1)
	}

	reuse_err := func() {
		fmt.Println("Reuse of tokens is not a meaningful parameter in counter-based mode")
		os.Exit(1)
	}
	for {
		var c C.int
		c = C.getopt_long(
			C.int(len(os.Args)),
			argv,
			C.CString(optstring),
			opts,
			(*C.int)(unsafe.Pointer(&idx)),
		)

		if c > 0 {
			for i := 0; i < len(options)-1; i++ {
				if options[i][3].(int32) == int32(c) {
					idx = int32(i)
					break
				}
			}
		} else if c < 0 {
			break
		}

		log.Println("c", string(c), "idx", idx)

		if dec(&idx) <= 0 {
			// --help
			usage()
			if idx < -1 {
				failed("Failed to parse command line")
			}
			success()
		} else if dec(&idx) == 0 {
			// --version
			version()
			success()
		} else if dec(&idx) == 0 {
			// --counter-based, -c
			if mode != ASK_MODE {
				dump_err("-c", "-t")
			}
			if reuse != ASK_REUSE {
				reuse_err()
			}
			mode = HOTP_MODE
		} else if dec(&idx) == 0 {
			// --no-confirm, -C
			confirm = 0
		} else if dec(&idx) == 0 {
			// --time-based, -t
			if mode != ASK_MODE {
				dump_err("-c", "-t")
			}
			mode = TOTP_MODE
		} else if dec(&idx) == 0 {
			// --disallow-reuse, -d
			if reuse != ASK_REUSE {
				dump_err("-d", "-D")
			}
			if mode == HOTP_MODE {
				reuse_err()
			}
			reuse = DISALLOW_REUSE
		} else if dec(&idx) == 0 {
			// --allow-reuse, -D
			if reuse != ASK_REUSE {
				dump_err("-d", "-D")
			}
			if mode == HOTP_MODE {
				reuse_err()
			}
			reuse = ALLOW_REUSE
		} else if dec(&idx) == 0 {
			// --force, -f
			if reuse != ASK_REUSE {
				dump_err("-f")
			}
			force = 1
		} else if dec(&idx) == 0 {
			// --label, -l
			if reuse != ASK_REUSE {
				dump_err("-l")
			}
			label = C.GoString(C.optarg)
		} else if dec(&idx) == 0 {
			// --issuer, -i
			if issuer != "" {
				dump_err("-i")
			}
			issuer = C.GoString(C.optarg)
		} else if dec(&idx) == 0 {
			// --quiet, -q
			if quite != 1 {
				dump_err("-q")
			}
			quite = 1
		} else if dec(&idx) == 0 {
			// --rate-limit, -r
			if r_limit > 0 {
				dump_err("-r")
			} else if r_limit < 0 {
				exclusive_err("-u", "-r")
			}

			val, err := strconv.ParseInt(C.GoString(C.optarg), 10, 64)
			if err != nil || val < 1 || val > 10 {
				range_err("-r", "1..10")
			}

			r_limit = int(val)
		} else if dec(&idx) == 0 {
			// --rate-time, -R
			if r_limit > 0 {
				dump_err("-R")
			} else if r_limit < 0 {
				exclusive_err("-u", "-R")
			}

			val, err := strconv.ParseInt(C.GoString(C.optarg), 10, 64)
			if err != nil || val < 15 || val > 600 {
				range_err("-R", "15..600")
			}

			r_time = int(val)
		} else if dec(&idx) == 0 {
			// --no-rate-limit, -u
			if r_limit > 0 || r_time > 0 {
				dump_err("-r")
			} else if r_limit < 0 {
				dump_err("-u")
			}
			r_time = -1
			r_limit = -1
		} else if dec(&idx) == 0 {
			// --secret, -s
			if secret_fn != "" {
				dump_err("-s")
			}
			if C.GoString(C.optarg) == "" {
				fmt.Println("-s must be followed by a filename")
				os.Exit(1)
			}

			secret_fn = C.GoString(C.optarg)
		} else if dec(&idx) == 0 {
			// --step-size, -S
			if step_size != 0 {
				dump_err("-S")
			}

			val, err := strconv.ParseInt(C.GoString(C.optarg), 10, 64)
			if err != nil || val < 1 || val > 60 {
				range_err("-S", "1..60")
			}

			step_size = int(val)
		} else if dec(&idx) == 0 {
			// --window-size, -w
			if window_size != 0 {
				dump_err("-w/-W")
			}

			val, err := strconv.ParseInt(C.GoString(C.optarg), 10, 64)
			if err != nil || val < 1 || val > 21 {
				range_err("-w", "1..21")
			}

			window_size = int(val)
		} else if dec(&idx) == 0 {
			// --minimal-window, -W
			if window_size != 0 {
				dump_err("-w/-W")
			}
			window_size = -1
		} else if dec(&idx) == 0 {
			// --emergency-codes, -e
			if emergency_codes >= 0 {
				dump_err("-e")
			}

			val, err := strconv.ParseInt(C.GoString(C.optarg), 10, 64)
			if err != nil || val < 0 || val > 21 {
				range_err("-e", fmt.Sprintf("0..%v", MAX_SCRATCHCODES))
			}

			emergency_codes = int(val)
		} else {
			fmt.Println("Error")
			os.Exit(1)
		}
	}

	_, _, _ = confirm, force, label

	optind := *(*int)(unsafe.Pointer(&C.optind))
	if optind != len(os.Args) {
		usage()
		if idx < -1 {
			failed("Failed to parse command line")
		}
		success()
	}

	if reuse != ASK_REUSE && mode != TOTP_MODE {
		failed("Must select time-based mode, when using -d or -D")
	}

	if (r_time != 0 && r_limit == 0) || (r_time == 0 && r_limit != 0) {
		failed("Must set -r when setting -R, and vice versa")
	}

	if emergency_codes < 0 {
		emergency_codes = SCRATCHCODES
	}

	if label == "" {
		label = getLabel()
	}
	if issuer == "" {
		issuer = hostname()
	}

	secret := Secret()

	use_totp := 0
	if mode == ASK_MODE {
		use_totp = maybe("Do you want authentication tokens to be time-based")
	} else {
		use_totp, mode = TOTP_MODE, TOTP_MODE
	}

	if quite == 0 {
		display(secret, label, use_totp, issuer)
		fmt.Printf("Your new secret key is: %s", secret_fn)

		if confirm != 0 && use_totp != 0 {
			for {
				test_code := ask("Enter code from app (-1 to skip):")
				val, err := strconv.ParseInt(test_code, 10, 64)
				if err != nil || val < 0 {
					fmt.Println("Code confirmation skipped")
					break
				}

				if step_size != 0 {
					step_size = 30
				}

				tm := int(time.Now().Unix()) / step_size
				correct_code := GenerateCode(nil, tm)
				if correct_code == int(val) {
					fmt.Println("Code confirmed")
					break
				}

				fmt.Println("Code incorrect (correct code %06d). Try again.")
				fmt.Println(correct_code)
			}
		} else {
			tm := 1
			fmt.Printf("Your verification code for code %d is %06d\n", tm,
				GenerateCode(nil, tm))
		}

		fmt.Println("Your emergency scratch codes are:")
	}

	_ = use_totp

}

func Secret() []byte {
	var (
		hotp      = `" HOTP_COUNTER 1\n`
		_         = `" TOTP_AUTH\n`
		disallow  = `" DISALLOW_REUSE\n`
		step      = `" STEP_SIZE 30\n`
		window    = `" WINDOW_SIZE 17\n`
		ratelimit = `" RATE_LIMIT 3 30\n`
	)
	log.Println(len(hotp))

	secretLen := (SECRET_BITS+BITS_PER_BASE32_CHAR-1)/BITS_PER_BASE32_CHAR +
		1 /*newline*/ +
		len(hotp) + // hottop and totp are mutually exclusive
		len(disallow) +
		len(step) +
		len(window) +
		len(ratelimit) + 5 /* NN MMM (total of five digits)*/ +
		SCRATCHCODE_LENGTH*(MAX_SCRATCHCODES+1 /*newline*/) +
		1 /* NUL termination character */

	var secret = make([]byte, secretLen, secretLen)

	dataLen := SECRET_BITS/8 + MAX_SCRATCHCODES*BYTES_PER_SCRATCHCODE
	var data = make([]byte, dataLen, dataLen)

	fd, _ := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	n, _ := fd.Read(data)
	if n != len(data) {
		os.Exit(1)
		return nil
	}

	n = base32Encode(secret, data[:SECRET_BITS/8])
	return secret[:n]
}

func main() {
	var (
		hotp      = `" HOTP_COUNTER 1\n`
		_         = `" TOTP_AUTH\n`
		disallow  = `" DISALLOW_REUSE\n`
		step      = `" STEP_SIZE 30\n`
		window    = `" WINDOW_SIZE 17\n`
		ratelimit = `" RATE_LIMIT 3 30\n`
	)
	log.Println(len(hotp))

	secretLen := (SECRET_BITS+BITS_PER_BASE32_CHAR-1)/BITS_PER_BASE32_CHAR +
		1 /*newline*/ +
		len(hotp) + // hottop and totp are mutually exclusive
		len(disallow) +
		len(step) +
		len(window) +
		len(ratelimit) + 5 /* NN MMM (total of five digits)*/ +
		SCRATCHCODE_LENGTH*(MAX_SCRATCHCODES+1 /*newline*/) +
		1 /* NUL termination character */

	var secret = make([]byte, secretLen, secretLen)

	dataLen := SECRET_BITS/8 + MAX_SCRATCHCODES*BYTES_PER_SCRATCHCODE
	var data = make([]byte, dataLen, dataLen)

	fd, _ := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	n, _ := fd.Read(data)
	if n != len(data) {
		return
	}

	data = []byte("12345678901234567890123456789012345678901234567890123456")
	n = base32Encode(secret, data[:SECRET_BITS/8])

	label := getLabel()
	issuer := hostname()

	display(secret[:n], label, TOTP_MODE, issuer)

	tm := time.Now().Unix() / 30
	code := GenerateCode(secret[:n], int(tm))
	log.Println("code", code)

	//getParams()

	maybe("Do you want authentication tokens to be time-based")
}
