package ipc

import (
	"syscall"
	"unsafe"
)

// 信号量
type Sema struct {
	semaid uintptr
}

type sembuf struct {
	semnum  int16 // 信号量组的序号: 0~semnums-1
	semop   int16 // 信号量值在一次操作当中改变的量
	semflag int16 // IPC_NOWAIT, SEM_UNDO
}

func Semaphore(pathname string, projectid, nums int) (*Sema, error) {
	key, err := ftok(pathname, projectid)
	if err != nil {
		return nil, err
	}

	flag := IPC_CREAT | 0666
	semaid, _, errno := syscall.RawSyscall(syscall.SYS_SEMGET, uintptr(key),
		uintptr(nums), uintptr(flag))
	if errno != 0 && errno != syscall.EEXIST {
		return nil, errnoErr(errno)
	}

	return &Sema{semaid: semaid}, nil
}

func (sm *Sema) Init(v int) error {
	return sm.Ctrl(SETVAL, 0, v)
}

func (sm *Sema) Remove() error {
	return sm.Ctrl(IPC_RMID, 0, semid_ds{})
}

func (sm *Sema) P() error {
	var buf sembuf
	buf.semnum = 0
	buf.semop = -1
	buf.semflag = SEM_UNDO

	ret, _, errno := syscall.RawSyscall(syscall.SYS_SEMOP, sm.semaid,
		uintptr(unsafe.Pointer(&buf)), uintptr(1))
	if ret == -1 {
		return errnoErr(errno)
	}

	return nil
}

func (sm *Sema) V() error {
	var buf sembuf
	buf.semnum = 0
	buf.semop = 1
	buf.semflag = SEM_UNDO

	ret, _, errno := syscall.RawSyscall(syscall.SYS_SEMOP, sm.semaid,
		uintptr(unsafe.Pointer(&buf)), uintptr(1))
	if ret == -1 {
		return errnoErr(errno)
	}

	return nil
}

func (sm *Sema) Ctrl(cmd, semnum int, buf interface{}) error {
	switch cmd {
	case SETVAL:
		val := buf.(int)
		_, _, errno := syscall.RawSyscall6(syscall.SYS_SEMCTL, sm.semaid,
			uintptr(semnum), uintptr(cmd), uintptr(val), 0, 0)
		if errno != 0 {
			return errnoErr(errno)
		}
	case IPC_SET, IPC_STAT:
		val := buf.(semid_ds)
		_, _, errno := syscall.RawSyscall6(syscall.SYS_SEMCTL, sm.semaid,
			uintptr(semnum), uintptr(cmd), uintptr(unsafe.Pointer(&val)), 0, 0)
		if errno != 0 {
			return errnoErr(errno)
		}
	case SETALL, GETALL:
		val := buf.([]uint16)
		_, _, errno := syscall.RawSyscall6(syscall.SYS_SEMCTL, sm.semaid,
			uintptr(semnum), uintptr(cmd), uintptr(unsafe.Pointer(&val[0])), 0, 0)
		if errno != 0 {
			return errnoErr(errno)
		}
	}

	return nil
}
