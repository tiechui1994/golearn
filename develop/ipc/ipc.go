package ipc

import (
	"syscall"
)

const (
	IPC_CREAT = 0x200

	IPC_RMID = 0x0
	IPC_SET  = 0x1
	IPC_STAT = 0x2
)

const (
	SHM_HUGETLB = 04000

	SHM_EXEC   = 0x0
	SHM_RDONLY = 0x1
	SHM_REMAP  = 0x2
)

const (
	SEM_UNDO = 0x0

	SETVAL = 0x04
	SETALL = 0x05
	GETALL = 0x06
)

type ipc_perm struct {
	key int32

	uid  uint32
	gid  uint32
	cuid uint32
	cgid uint32

	mode uint16
	seq  uint16
}

type msqid_ds struct {
	msg_perm  ipc_perm
	msg_stime int64
	msg_rtime int64
	msg_ctime int64

	msg_cbytes uint64
	msq_qnum   uint64
	msq_qbytes uint64

	msg_lspid int32
	msg_lrpid int32
}

type semid_ds struct {
	sem_perm  ipc_perm
	sem_otime int64
	sem_ctime int64
	sem_nsems uint64
}

func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case syscall.EAGAIN:
		return syscall.EAGAIN
	case syscall.EINVAL:
		return syscall.EINVAL
	case syscall.ENOENT:
		return syscall.ENOENT
	}

	return e
}
