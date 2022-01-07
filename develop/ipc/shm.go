package ipc

import (
	"errors"
	"syscall"
	"unsafe"
)

// 共享内存
type ShareMemory struct {
	shid uintptr
	addr uintptr
}

func ShareM(pathname string, projectid, size, flag uint) (*ShareMemory, error) {
	key, err := ftok(pathname, projectid)
	if err != nil {
		return nil, err
	}

	flags := uint32(flag | IPC_CREAT | SHM_DEST)
	shid, _, errno := syscall.Syscall(syscall.SYS_SHMGET, uintptr(key), uintptr(size), uintptr(flags))
	if errno != 0 && errno != syscall.EEXIST {
		return nil, errnoErr(errno)
	}

	return &ShareMemory{shid: shid}, nil
}

func (shm *ShareMemory) Address() uintptr {
	if shm.addr == 0 {
		panic("must be link before")
	}
	return shm.addr
}

func (shm *ShareMemory) Link() error {
	addr, _, errno := syscall.Syscall(syscall.SYS_SHMAT,
		shm.shid, 0, 0)
	if errno != 0 {
		return errnoErr(errno)
	}

	shm.addr = addr
	return nil
}

func (shm *ShareMemory) UnLink() error {
	_, _, errno := syscall.Syscall(syscall.SYS_SHMDT,
		uintptr(shm.addr), 0, 0)
	if errno != 0 {
		return errnoErr(errno)
	}

	shm.addr = 0
	return nil
}

func (shm *ShareMemory) Remove() error {
	return shm.Ctrl(IPC_RMID, &shmid_ds{})
}

func (shm *ShareMemory) Ctrl(cmd int32, buf *shmid_ds) error {
	if buf == nil {
		return errors.New("buf is nil")
	}

	_, _, errno := syscall.Syscall(syscall.SYS_SHMCTL,
		shm.shid, uintptr(cmd), uintptr(unsafe.Pointer(buf)))
	if errno != 0 {
		return errnoErr(errno)
	}

	return nil
}
