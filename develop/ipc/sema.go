package ipc

import (
	"syscall"
	"unsafe"
)

type sembuf struct {
	semnum  int16 // 信号量组的序号: 0~semnums-1
	semop   int16 // 信号量值在一次操作当中改变的量
	semflag int16 // IPC_NOWAIT, SEM_UNDO
}

type Sema struct {
	semaid uintptr
}

func Semaphore(pathname string, projectid, nums int) (*Sema, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(pathname, &stat); err != nil {
		return nil, err
	}

	key := int(uint(projectid&0xff)<<24 | uint((stat.Dev&0xff)<<16) | (uint(stat.Ino) & 0xffff))
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

func (sm *Sema) Close() error {
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
	case IPC_SET, IPC_STAT, IPC_RMID:
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
