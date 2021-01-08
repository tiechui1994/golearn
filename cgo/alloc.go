package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// C 语言数组使用案例
typedef struct user {
	int age;
	const char* name;
} user;

void printString(const char* s, int n) {
	int length = 10;

	printf("\n============= **user ==================\n\n");
	user** p = (user**)malloc(length * sizeof(user*));
	memset(p, 0, length * sizeof(user*));

	for (int i=0; i<length; i++) {
		p[i] = (user*)malloc(sizeof(user));
		p[i]->age=i*1032;
		p[i]->name="name";
		printf("ptr[%d] %p\n", i, p[i]);
	}

	printf("0 name: %s\n", p[0]->name);
	printf("0 age: %d\n", p[0]->age);

	printf("\n============= *user ==================\n\n");
	user* ptr = (user*)malloc(length * sizeof(user));
	memset(ptr, 0, length * sizeof(user));

	for (int i=0; i<length; i++) {
		ptr[i].age=(i+1)*1032;
		ptr[i].name="name";
		printf("ptr[%d] %p\n", i, &(ptr[i]));
	}
	free(ptr);

	printf("0 name: %s\n", ptr->name);
	printf("0 age: %d\n", ptr->age);

	printf("size %ld\n", sizeof(const char*));
    printf("len is:%d, data: %s \n", n, s);
}
*/
import "C"

func printString(s string) {
	C.printString(C.CString(s), C.int(len(s)))
}

func main() {
	printString("Hello")
}
