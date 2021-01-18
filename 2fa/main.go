package main

/*
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define ULONG_MAX 0xFFFFFFFF

static int ISSPACE(int ch) {
	return (ch == '\t' || ch == '\n' || ch == '\v' || ch == '\f' || ch == '\r' || ch == ' ') ? 1 : 0;
}
static int ISDIGIT(int ch) {
	return  (ch >= '0' && ch <= '9') ? 1 : 0;
}
static int ISALPHA(int ch) {
	return ((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) ? 1 : 0;
}
static int ISUPPER(int ch) {
	return ('A' <= ch && ch <= 'Z') ? 1 : 0;
}

unsigned long strtou(const char *nptr, char **endptr, register int base)
{
        register const char *s = nptr;
        register unsigned long acc;
        register int c;
        register unsigned long cutoff;
        register int neg = 0, any, cutlim;
        do {
                c = *s++;
        } while (ISSPACE(c));
        if (c == '-') {
                neg = 1;
                c = *s++;
        } else if (c == '+')
                c = *s++;
        if ((base == 0 || base == 16) &&
            c == '0' && (*s == 'x' || *s == 'X')) {
                c = s[1];
                s += 2;
                base = 16;
        }
        if (base == 0)
              base = c == '0' ? 8 : 10;

        printf("cur1:[%s]\n",s);

        cutoff = (unsigned long)ULONG_MAX / (unsigned long)base;
        cutlim = (unsigned long)ULONG_MAX % (unsigned long)base;
        int idx = 0;
        for (acc = 0, any = 0;; c = *s++) {
        		idx++;
                if (ISDIGIT(c))
                        c -= '0';
                else if (ISALPHA(c))
                        c -= ISUPPER(c) ? 'A' - 10 : 'a' - 10;
                else
                        break;
                if (c >= base)
                        break;
                if (any < 0 || acc > cutoff || (acc == cutoff && c > cutlim))
                        any = -1;
                else {
                        any = 1;
                        acc *= base;
                        acc += c;
                }
        }

        printf("cur2:[%s]\n", s);
        printf("idx:%d\n", idx);

        if (any < 0) {
                acc = ULONG_MAX;
                errno = ERANGE;
        } else if (neg)
                acc = -acc;

        if (endptr != 0)
        	printf("endptr: [%s]\n", s-1);
        	printf("any:%d\n", any);
            *endptr = (char *) (any ? s - 1 : nptr);
        return (acc);
}

static void plus(const char* desc, char* endptr) {
	if (*endptr) {
  		printf("%s (*var) is TRUE \n", desc);
  	}
}


*/
import "C"
import "fmt"

func main() {
	C.plus(C.CString("empty str"), C.CString(""))
	C.plus(C.CString("zero array"), C.CString(string([]byte{0, 0, 0, 0, 0})))
	fmt.Println("len", len(string([]byte{0, 0, 0, 0, 0})))
	C.plus(C.CString("'0' str"), C.CString("0"))
	C.plus(C.CString("' ' str"), C.CString(" "))
}
