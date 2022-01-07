package ipc

/*
#include <sys/syscall.h>
#include <unistd.h>
#include <errno.h>
*/
import "C"

import (
	"syscall"
)

const (
	// msgget, semget, shmget modes
	IPC_CREAT  = 01000 // create key if key not exist
	IPC_EXCL   = 02000 // fail if key exists
	IPC_NOWAIT = 04000 // return error on wait

	// msgctl, semctl, shmctl commands
	IPC_RMID = 0x0 // remove identififer
	IPC_SET  = 0x1 // set ipc_perm options
	IPC_STAT = 0x2 // get ipc_perm options

	// special key value
	IPC_PRIVATE = 0
)

const (
	// shm_mode flags
	SHM_DEST    = 01000 // segment will be destoroyed on last detach
	SHM_LOCKED  = 02000 // segment will not be swapped
	SHM_HUGETLB = 04000 // segment is mapped via hugetlb

	// shmat flags
	SHM_RDONLY = 010000 // attach read-only else read-write
	SHM_RND    = 020000 // round attach address SHMLBA
	SHM_REMAP  = 040000
	SHM_EXEC   = 0100000 // execution access

	// shmctl commands
	SHM_LOCK   = 11 // lock segment (root only)
	SHM_UNLOCK = 12 // unlock segment (root only)
)

const (
	// semop flags
	SEM_UNDO = 0x1000 // undo the operation on exit

	// sem ctl cmd
	SETALL  = 13
	GETNCNT = 14 // semncnt
	GETNZNT = 15 // semzcnt
	SETVAL  = 16
	GETALL  = 17
)

type ipc_perm struct {
	key int32

	uid  uint32
	gid  uint32
	cuid uint32
	cgid uint32

	mode uint16
	_    uint16
	seq  uint16
	_    uint16
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

type shmid_ds struct {
	shm_perm  ipc_perm
	shm_segsz uint64 // size of segment in bytes

	shm_atime int64 // time of last shmat
	shm_dtime int64 // time of last shmdt
	shm_ctime int64 // time of last shmctl

	shm_cpid   int32  // pid of creator
	shm_lpid   int32  // pid of last shmop
	shm_nattch uint64 // numbers of cur attaches
}

func ftok(pathname string, projectid uint) (key uint, err error) {
	var stat syscall.Stat_t
	err = syscall.Stat(pathname, &stat)
	if err != nil {
		return key, err
	}

	key = uint(projectid&0xff)<<24 | uint((stat.Dev&0xff)<<16) | (uint(stat.Ino) & 0xffff)
	return key, nil
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

func Syscall(trap, a1, a2, a3 uintptr) (ret uintptr, err error) {
	ret = C.syscall(C.long(trap), a1, a2, a3)
	if ret != 0 {
		return ret, syscall.Errno(C.errno)
	}

	return ret, nil
}
