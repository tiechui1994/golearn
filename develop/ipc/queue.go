package ipc

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	ErrEmpty = errors.New("queue is empty")
)
// 消息队列
type msgbuf struct {
	mtype int64
	mtext [256]uint8
}


type MsqQ struct {
	msgid uintptr
}

func Queue(pathname string, projectid int) (*MsqQ, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(pathname, &stat); err != nil {
		return nil, err
	}

	key := int(uint(projectid&0xff)<<24 | uint((stat.Dev&0xff)<<16) | (uint(stat.Ino) & 0xffff))
	flags := IPC_CREAT | 0664
	msgid, _, errno := syscall.RawSyscall(syscall.SYS_MSGGET, uintptr(key), uintptr(flags), 0)
	if errno != 0 && errno != syscall.EEXIST {
		return nil, errnoErr(errno)
	}

	return &MsqQ{msgid: msgid}, nil
}

func (mq *MsqQ) Send(mtype int, mtext [256]uint8) error {
	var msg msgbuf
	msg.mtype = int64(mtype)
	msg.mtext = mtext

	ret, _, errno := syscall.RawSyscall6(syscall.SYS_MSGSND,
		uintptr(mq.msgid), uintptr(unsafe.Pointer(&msg)), uintptr(unsafe.Sizeof(msgbuf{})), 0, 0, 0)
	if int(ret) != 0 {
		return errnoErr(errno)
	}
	return nil
}

func (mq *MsqQ) Recv(mtype uint) ([]byte, error) {
	var msg msgbuf
	_, _, errno := syscall.RawSyscall6(syscall.SYS_MSGRCV,
		uintptr(mq.msgid), uintptr(unsafe.Pointer(&msg)), uintptr(unsafe.Sizeof(msgbuf{})), uintptr(mtype), 0, 0)
	if errno == syscall.EINTR {
		return nil, ErrEmpty
	}
	if errno != 0 {
		return nil, errnoErr(errno)
	}

	return msg.mtext[:], nil
}

func (mq *MsqQ) Close() error {
	return mq.Ctrl(IPC_RMID, &msqid_ds{})
}

func (mq *MsqQ) Ctrl(cmd int32, buf *msqid_ds) error {
	if buf == nil {
		return errors.New("buf is nil")
	}

	_, _, errno := syscall.RawSyscall(syscall.SYS_MSGCTL,
		uintptr(mq.msgid), uintptr(cmd), uintptr(unsafe.Pointer(buf)))
	if errno != 0 {
		return errnoErr(errno)
	}

	return nil
}