package main

/*
#include <signal.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <sys/types.h>
#include <unistd.h>
#include <setjmp.h>

jmp_buf env;
static void handle(int signum, siginfo_t *info, void *secret) {
	printf("crash signum:%d si_code:%d\n", signum, info->si_code);
	longjmp(env, 1);
}

static void sigsetup2(void) {
	struct sigaction act;
	memset(&act, 0, sizeof act);
	act.sa_flags = SA_ONSTACK | SA_SIGINFO;
	act.sa_sigaction = handle;
	sigaction(SIGSEGV, &act, 0);
	sigaction(SIGABRT, &act, 0);
}

typedef void (*cb)(void);

static int mysetjmp(cb f) {
	printf("size: %d, %d\n", sizeof(jmp_buf), env[10]);
	if (setjmp(env) == 0) {
		f();
	} else {
		printf("异常后恢复\n");
	}
}
*/
import "C"

/*
非局部跳转语句: setjmp 和 longjmp函数.

非局部指的是, 不是由普通C语言goto, 语句在一个函数内实施的跳转, 而是在栈上跳过若干调用帧, 返回到当前函数调用路径上的某一
个函数中.

#include <setjmp.h>
int setjmp(jmp_buf  env);
	返回值: 若直接调用则返回0, 若从 longjmp 调用返回则返回非0值的longjmp中的val值

void longjmp(jmp_buf env,int val);
	调用此函数则返回到语句 setjmp 所在的地方, 其中 env 就是 setjmp 中的 env, 而 val 则是使setjmp的返回值变为val.

当检查到一个错误时, 则以两个参数调用longjmp函数, 第一个就是在调用 setjmp 时所用的 env, 第二个参数是具有非0值的val,
它将成为从 setjmp 处返回的值. 使用第二个参数的原因是对于一个 setjmp 可以有多个longjmp.

1. setjmp 与 longjmp结合使用时, 它们必须有严格的先后执行顺序, 也即先调用 setjmp 函数, 之后再调用longjmp函数,
以恢复到先前被保存的 "程序执行点". 否则, 如果在 setjmp 调用之前, 执行longjmp函数, 将导致程序的执行流变的不可预测,
很容易导致程序崩溃而退出.

2.  longjmp必须在setjmp调用之后, 而且longjmp必须在setjmp的作用域之内. 具体来说, 在一个函数中使用setjmp来初始化一个
全局标号, 然后只要该函数未曾返回, 那么在其它任何地方都可以通过 longjmp 调用来跳转到 setjmp 的下一条语句执行.
实际上setjmp函数将发生调用处的局部环境保存在了一个 jmp_buf 的结构当中, 只要主调函数中对应的内存未曾释放(函数返回时局部
内存就失效了), 那么在调用 longjmp 的时候就可以根据已保存的 jmp_buf 参数恢复到 setjmp 的地方执行.
*/

// Sigsetup Sigsetup
func Sigsetup2() {
	C.sigsetup2()
}

func SafeCall(f C.cb) {
	C.mysetjmp(f)
}
