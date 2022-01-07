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
type MsqQ struct {
	msgid uintptr
}

type msgbuf struct {
	mtype int64
	mtext [256]uint8
}

func Queue(pathname string, projectid, flag uint) (*MsqQ, error) {
	key, err := ftok(pathname, projectid)
	if err != nil {
		return nil, err
	}

	flags := uint32(flag | IPC_CREAT)
	msgid, _, errno := syscall.Syscall(syscall.SYS_MSGGET, uintptr(key), uintptr(flags), 0)
	if errno != 0 && errno != syscall.EEXIST {
		return nil, errnoErr(errno)
	}

	return &MsqQ{msgid: msgid}, nil
}

func (mq *MsqQ) Send(mtype int, mtext [256]uint8) error {
	var msg msgbuf
	msg.mtype = int64(mtype)
	msg.mtext = mtext

	ret, _, errno := syscall.Syscall6(syscall.SYS_MSGSND,
		mq.msgid, uintptr(unsafe.Pointer(&msg)), uintptr(unsafe.Sizeof(msgbuf{})), 0, 0, 0)
	if int(ret) != 0 {
		return errnoErr(errno)
	}
	return nil
}

func (mq *MsqQ) Recv(mtype uint) ([]byte, error) {
	var msg msgbuf
	_, _, errno := syscall.Syscall6(syscall.SYS_MSGRCV,
		mq.msgid, uintptr(unsafe.Pointer(&msg)), uintptr(unsafe.Sizeof(msgbuf{})), uintptr(mtype), 0, 0)
	if errno == syscall.EINTR {
		return nil, ErrEmpty
	}
	if errno != 0 {
		return nil, errnoErr(errno)
	}

	return msg.mtext[:], nil
}

func (mq *MsqQ) Remove() error {
	return mq.Ctrl(IPC_RMID, &msqid_ds{})
}

func (mq *MsqQ) Ctrl(cmd int32, buf *msqid_ds) error {
	if buf == nil {
		return errors.New("buf is nil")
	}

	_, _, errno := syscall.Syscall(syscall.SYS_MSGCTL,
		mq.msgid, uintptr(cmd), uintptr(unsafe.Pointer(buf)))
	if errno != 0 {
		return errnoErr(errno)
	}

	return nil
}
