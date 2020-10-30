package main

import (
	"fmt"
	"os"
	"syscall"
)

/*
open vs opneat

定义:
int open(const char* pathname, int  flags, umode_t mode)
int openat(int  dirfd, const char* pathname, int  flags, umode_t mode)

如果 pathname 是一个相对路径, 并且 dirfd 包含 AT_FDCWD, 那么 pathname 是相对于"当前进程的工作目录"的. 此时的行为
和 open 是一致的. (在 go 的 runtime/sys_linux_amd64.s 当中定义为 -100). 也就说, 使用 openat 只使用 open 的功
能.

如果 pathname 是一个相对路径, 那么 pathname 相对于 dirfd 目录.

如果 pathname 是一个绝对路径, 那么 dirfd 将被忽略.

*/
func main() {
	pid, _, _ := syscall.Syscall(syscall.SYS_GETPID, 0, 0, 0) // 用不到的就补上 0
	fmt.Println("Process id: ", pid)

	os.Open("")
}
