package main

import (
	"os"
	"os/user"
	"fmt"
	"time"
	"unsafe"
	"strconv"
	"bytes"
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
	TRUE  = 1
	FALSE = 0
)

const (
	ASK_MODE  = 0
	HOTP_MODE = 1
	TOTP_MODE = 2

	ASK_REUSE      = 0
	DISALLOW_REUSE = 1
	ALLOW_REUSE    = 2
)

func getLabel() string {
	userinfo, err := user.Current()
	if err == nil {
		return userinfo.Username + "@" + hostname()
	}

	return ""
}

func display(secret []byte, label string, use_totop int, issuer string) {
	totop := "h"
	if use_totop != 0 {
		totop = "t"
	}

	sl := strlen(secret)
	value := "secret=" + string(secret[:sl])

	if issuer != "" {
		value = fmt.Sprintf("%v&issuer=%v", value, urlencode(issuer))
	}
	u := fmt.Sprintf("otpauth://%votp/%s?%s", totop, urlencode(string(label)), value)
	u = "https://www.google.com/chart?chs=200x200&chld=M|0&cht=qr&chl=" + urlencode(u)
	fmt.Println("Warning: pasting the following URL into your browser exposes the OTP secret to Google:")
	fmt.Println("  " + u)
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

func maybe(msg string) int {
	fmt.Println()
	for {
		fmt.Printf("%s (y/n) ", msg)

		var read string
		fmt.Scanln(&read)
		switch read {
		case "Y", "y":
			return TRUE
		case "N", "n":
			return FALSE
		}
	}
}

func ask(msg string) string {
	fmt.Printf("%s ", msg)
	var read string
	fmt.Scanln(&read)
	return read
}

/*
头文件: #include <string.h>
声明: void* memcpy(void *dst, const void *src, size_t n);
功能: 将 src 所指向的内存内容的前 n 个字节拷贝到 dst 所指的内存地址上.
返回值: dst 的首地址

头文件: #include <string.h>
声明: void* memmove(void *dst, const void *src, size_t n);
功能: 将 src 所指向的内存内容的前 n 个字节拷贝到 dst 所指的内存地址上.
返回值: dst 的首地址

memmove() 和 memcpy() 的功能是一样的.
memcpy() 函数有一个限制, 就是目的地址(dst)和原地址(src)不能重叠.
memmove() 函数没有 memcpy() 的缺点, 就是执行的效率比 memcpy() 函数低一些.
*/

func addOption(buf []byte, opt string) {
	if strlen(buf)+strlen([]byte(opt)) >= len(buf) {
		panic("assert")
	}

	idx := bytes.IndexByte(buf, '\n') + 1
	if idx == 0 {
		panic("assert")
	}

	N := strlen([]byte(opt))
	L := strlen(buf[N:]) + 1
	copy(buf[idx+N:idx+N+L], buf[idx:])
	copy(buf[idx:idx+N], opt)
}

const (
	option_1 = `Do you want to disallow multiple uses of the same authentication
token? This restricts you to one login about every 30s, but it increases
your chances to notice or even prevent man-in-the-middle attacks`

	option_2 = `By default, a new token is generated every 30 seconds by the mobile app.
In order to compensate for possible time-skew between the client and the server,
we allow an extra token before and after the current time. This allows for a
time skew of up to 30 seconds between authentication server and client. If you
experience problems with poor time synchronization, you can increase the window
from its default size of 3 permitted codes (one previous code, the current
code, the next code) to 17 permitted codes (the 8 previous codes, the current
code, and the 8 next codes). This will permit for a time skew of up to 4 minutes
between client and server.
Do you want to do so?`

	option_4 = `If the computer that you are logging into isn't hardened against brute-force
login attempts, you can enable rate-limiting for the authentication module.
By default, this limits attackers to no more than 3 login attempts every 30s.
Do you want to enable rate-limiting?`

	option_3 = `By default, three tokens are valid at any one time. This accounts for
generated-but-not-used tokens and failed login attempts. In order to
decrease the likelihood of synchronization problems, this window can be 
increased from its default size of 3 to 17. Do you want to do so?`
)

func maybeAddOption(msg string, buf []byte, opt string) {
	if maybe(msg) == TRUE {
		addOption(buf, opt)
	}
}

/*
文档: https://nanxiao.me/getopt-vs-getopt_long/

int getopt_long(int argc, char* const argv[], const char *optstring,
				const struct option *longopts, int *longindex);
*/
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
	reuse := ASK_REUSE

	force, quite := 0, 0
	r_limit, r_time := 0, 0

	secret_fd := ""
	label := ""
	issuer := ""

	confirm := 1
	step_size := 0
	window_size := 0
	emergency_codes := -1

	// error func
	version := func() {
		fmt.Println("google-authenticator 1.0.0")
	}
	success := func() {
		os.Exit(0)
	}
	reuse_err := func() {
		fmt.Println("Reuse of tokens is not a meaningful parameter in counter-based mode")
		os.Exit(1)
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
			if secret_fd != "" {
				dump_err("-s")
			}
			if C.GoString(C.optarg) == "" {
				fmt.Println("-s must be followed by a filename")
				os.Exit(1)
			}

			secret_fd = C.GoString(C.optarg)
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

	optind := int(C.optind)
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

	buf, secret := Secret()

	use_totp := 0
	if mode == ASK_MODE {
		use_totp = maybe("Do you want authentication tokens to be time-based")
	} else {
		use_totp, mode = TOTP_MODE, TOTP_MODE
	}

	// quite
	if quite == 0 {
		display(secret, label, use_totp, issuer)
		fmt.Printf("Your new secret key is: %s\n", string(secret))

		if confirm != 0 && use_totp != 0 {
			for {
				test_code := ask("Enter code from app (-1 to skip):")
				val, err := strconv.ParseInt(test_code, 10, 64)
				if err != nil || val < 0 {
					fmt.Println("Code confirmation skipped")
					break
				}

				if step_size == 0 {
					step_size = 30
				}
				tm := int(time.Now().Unix()) / step_size
				correct_code := GenerateCode(string(secret), tm)
				if correct_code == int(val) {
					fmt.Println("Code confirmed")
					break
				}

				fmt.Printf("Code incorrect (correct code %06d). Try again.\n", correct_code)
			}
		} else {
			tm := 1
			fmt.Printf("Your verification code for code %d is %06d\n", tm,
				GenerateCode(string(secret), tm))
		}

		fmt.Println("Your emergency scratch codes are:")
	}

	strcat(secret, "\n")
	if use_totp != 0 {
		strcat(secret, totp)
	} else {
		strcat(secret, hotp)
	}

	for i := 0; i < emergency_codes; i++ {
	scratch_code:
		scratch := 0
		for j := 0; j < BYTES_PER_SCRATCHCODE; j++ {
			scratch = 256*scratch + int(buf[SECRET_BITS/8+BYTES_PER_SCRATCHCODE*i+j])
		}

		modules := 1
		for j := 0; j < SCRATCHCODE_LENGTH; j++ {
			modules *= 10
		}

		scratch = (scratch & 0x7FFFFFFF) % modules
		if scratch < modules/10 {
			idx := SECRET_BITS/8 + BYTES_PER_SCRATCHCODE*i
			urandom(buf[idx:idx+BYTES_PER_SCRATCHCODE])
			goto scratch_code
		}

		if quite == 0 {
			fmt.Printf("  %08d\n", scratch)
		}

		idx := bytes.IndexByte(secret, char_zero)
		strcat(secret[idx:], fmt.Sprintf("%08d\n", scratch))
	}

	// secret file
	if secret_fd == "" {
		home := os.Getenv("HOME")
		if home == "" {
			fmt.Println("Cannot determine home directory")
			return
		}

		secret_fd = home + "/.google_authenticator"
	}

	// force
	if force == 0 {
		s := fmt.Sprintf(`Do you want me to update your "%s" file?`, secret_fd)
		if maybe(s) == FALSE {
			os.Exit(0)
		}
	}

	// Add optional flags
	if use_totp != 0 {
		if reuse == ASK_REUSE {
			maybeAddOption(option_1, secret, disallow)
		} else if reuse == DISALLOW_REUSE {
			addOption(secret, disallow)
		}

		if step_size != 0 {
			s := fmt.Sprintf(`" STEP_SIZE %d`+char_line, step_size)
			addOption(secret, s)
		}

		if window_size == 0 {
			maybeAddOption(option_2, secret, "")
		} else {
			if window_size <= 0 {
				window_size = 3
			}
			s := fmt.Sprintf(`" WINDOW_SIZE %d`+char_line, window_size)
			addOption(secret, s)
		}
	} else {
		// Counter based
		if window_size == 0 {
			maybeAddOption(option_3, secret, window)
		} else {
			if window_size <= 0 {
				window_size = 1
			}
			s := fmt.Sprintf(`" WINDOW_SIZE %d`+char_line, window_size)
			addOption(secret, s)
		}
	}

	if r_limit == 0 && r_time == 0 {
		maybeAddOption(option_4, secret, ratelimit)
	} else if r_limit > 0 && r_time > 0 {
		s := fmt.Sprintf(`" RATE_LIMIT %d %d`+char_line, r_limit, r_time)
		addOption(secret, s)
	}

	// Save file.
	tmp_fd := fmt.Sprintf("%s~", secret_fd)
	fd, err := os.OpenFile(tmp_fd, os.O_WRONLY|os.O_EXCL|os.O_CREATE|os.O_TRUNC, 0400)
	if err != nil {
		fmt.Fprintf(os.Stderr, `Failed to create "%s" (%s)`, secret_fd, err.Error())
		return
	}

	defer func() {
		os.Rename(tmp_fd, secret_fd)
		fd.Close()
	}()

	_, err = fd.Write(secret[:strlen(secret)])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to write new secret")
		return
	}
}

const (
	hotp      = `" HOTP_COUNTER 1` + char_line
	totp      = `" TOTP_AUTH` + char_line
	disallow  = `" DISALLOW_REUSE` + char_line
	step      = `" STEP_SIZE 30` + char_line
	window    = `" WINDOW_SIZE 17` + char_line
	ratelimit = `" RATE_LIMIT 3 30` + char_line
)

var ufd *os.File

func urandom(data []byte) {
	for i := 0; i < len(data); i++ {
		data[i] = 10
	}

	return
	if ufd == nil {
		var err error
		ufd, err = os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
		if err != nil {
			os.Exit(1)
		}
	}

	var count int
retry:
	n, _ := ufd.Read(data)
	if n != len(data) && count < 3 {
		count++
		goto retry
	}
}

func Secret() (origin []byte, secret []byte) {
	secretLen := (SECRET_BITS+BITS_PER_BASE32_CHAR-1)/BITS_PER_BASE32_CHAR +
		1 + // newline
		len(hotp) + // hottop and totp are mutually exclusive
		len(disallow) +
		len(step) +
		len(window) +
		len(ratelimit) + 5 + // NN MMM (total of five digits)
		SCRATCHCODE_LENGTH*(MAX_SCRATCHCODES+1 ) + // newline
		1 // NUL termination character

	secret = make([]byte, secretLen, secretLen)

	originLen := SECRET_BITS/8 + MAX_SCRATCHCODES*BYTES_PER_SCRATCHCODE
	origin = make([]byte, originLen, originLen)

	urandom(origin)

	base32Encode(origin, SECRET_BITS/8, secret, len(secret))
	return origin, secret
}

// i-- 操作
func dec(val *int32) int32 {
	res := *val
	*val -= 1
	return res
}

func main() {
	getParams()
}
