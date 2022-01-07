package ipc

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	AT_FDCWD = -0x64
)

// FIFO, 管道
// 1. mode 没有指定 O_NONBLOCK, 只读阻塞 open 到某个其他进程为写而打开此FIFO. 只写要阻塞 open 到某个进程为其他进
// 程为读而打开它.
// 2. mode 指定了 O_NONBLOCK, 则只读 open 立即返回. 对于只写 open, 如果没有进程已经为读而打开该 FIFO, 其 errno
// 为 ENXIO.

type Fifo struct {
	p  string
	fd int
}

func FIFO(filepath string, mode uint32) (*Fifo, error) {
	p0, err := syscall.BytePtrFromString(filepath)
	if err != nil {
		return nil, err
	}

	mode = mode | syscall.O_CREAT | syscall.S_IFIFO
	unknown := int(AT_FDCWD)
	_, _, errno := syscall.Syscall6(syscall.SYS_MKNODAT, uintptr(unknown), uintptr(unsafe.Pointer(p0)),
		uintptr(mode), uintptr(0), 0, 0)
	if errno != 0 && errno != syscall.EEXIST {
		return nil, errnoErr(errno)
	}

	fd, err := syscall.Open(filepath, int(mode), 0666)
	if err != nil && err != syscall.EEXIST {
		return nil, err
	}
	return &Fifo{fd: fd, p: filepath}, nil
}

func (io *Fifo) Read(buf []byte) (n int, err error) {
	return syscall.Read(io.fd, buf)
}

func (io *Fifo) Write(data []byte) (n int, err error) {
	return syscall.Write(io.fd, data)
}

func (io *Fifo) Remove() error {
	return os.Remove(io.p)
}
