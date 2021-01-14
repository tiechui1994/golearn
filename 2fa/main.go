package main

/*
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static int convert() {
	const char* value = "3 30";
	const char *endptr = value, *ptr;
	int attempts, interval;
	errno = 0;

	attempts = (int)strtoul(ptr = endptr, (char **)&endptr, 10);
	printf("result:%d\n", attempts < 1 || ptr == endptr || attempts > 100 || errno || (*endptr != ' ' && *endptr != '\t'));
	printf("endptr:[%s]\n", endptr);
	interval = (int)strtoul(ptr = endptr, (char **)&endptr, 10);
	printf("result:%d\n", interval<1 || ptr == endptr || interval>3600 || errno );
	printf("endptr:[%s]\n", endptr);
	return 0;
}


static int plus() {
	char d[] = {'0','1','2','3','4','5','6'};
	char* s = d;
	printf("d: %c\n", *s);
	char c = *s++; // c=*s; s++
	printf("d: %c\n", *s);
	printf("c: %c\n", c);
	printf("%s\n", s-1);
	// *s, s++
	return 0;
}
*/
import "C"

func main() {
	C.plus()
}
