package ipc

import (
	"errors"
	"syscall"
	"unsafe"
)

type ShareMemory struct {
	shid uintptr
	addr *uint8
}

func ShareM(pathname string, projectid int, size uint) (*ShareMemory, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(pathname, &stat); err != nil {
		return nil, err
	}

	key := int(uint(projectid&0xff)<<24 | uint((stat.Dev&0xff)<<16) | (uint(stat.Ino) & 0xffff))
	flag := IPC_CREAT | SHM_HUGETLB
	shid, _, errno := syscall.RawSyscall(syscall.SYS_SHMGET, uintptr(key), uintptr(size), uintptr(flag))
	if errno != 0 && errno != syscall.EEXIST {
		return nil, errnoErr(errno)
	}

	return &ShareMemory{shid: shid}, nil
}

func (shm *ShareMemory) Link() error {
	addr, _, errno := syscall.RawSyscall(syscall.SYS_SHMAT,
		shm.shid, 0, 0)
	if errno != 0 {
		return errnoErr(errno)
	}

	shm.addr = (*uint8)(unsafe.Pointer(addr))
	return nil
}

func (shm *ShareMemory) UnLink() error {
	_, _, errno := syscall.RawSyscall(syscall.SYS_SHMDT,
		uintptr(unsafe.Pointer(shm.addr)), 0, 0)
	if errno != 0 {
		return errnoErr(errno)
	}

	shm.addr = nil
	return nil
}

func (shm *ShareMemory) Close() error {
	return shm.Ctrl(IPC_RMID, &msqid_ds{})
}

func (shm *ShareMemory) Ctrl(cmd int32, buf *msqid_ds) error {
	if buf == nil {
		return errors.New("buf is nil")
	}

	_, _, errno := syscall.RawSyscall(syscall.SYS_SHMCTL,
		shm.shid, uintptr(cmd), uintptr(unsafe.Pointer(buf)))
	if errno != 0 {
		return errnoErr(errno)
	}

	return nil
}
