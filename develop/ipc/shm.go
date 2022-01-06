package ipc

import (
	"golang.org/x/sys/unix"
	"syscall"
)

const (
	SHM_HUGETLB = 04000
)

type ShareMemory struct {
	shid uintptr
}

func ShareM(pathname string, projectid int, size uint) (*ShareMemory, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(pathname, &stat); err != nil {
		return nil, err
	}

	key := int(uint(projectid&0xff)<<24 | uint((stat.Dev&0xff)<<16) | (uint(stat.Ino) & 0xffff))
	flag := unix.IPC_CREAT | SHM_HUGETLB
	shid, _, errno := syscall.RawSyscall(syscall.SYS_SHMGET, uintptr(key), uintptr(size), uintptr(flag))
	if errno != 0 && errno != syscall.EEXIST {
		return nil, errnoErr(errno)
	}

	return &ShareMemory{shid: shid}, nil
}

func (shm *ShareMemory) shmat()  {
	
}
