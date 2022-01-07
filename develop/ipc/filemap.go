package ipc

import (
	"os"
	"syscall"
)

type options struct {
	offset uint64
	prot   uint32
	flags  uint32
}

type Option interface {
	apply(*options)
}

type funcOption struct {
	f func(*options)
}

func (f *funcOption) apply(do *options) {
	f.f(do)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{f: f}
}

func WithWrite() Option {
	return newFuncOption(func(o *options) {
		o.prot |= syscall.PROT_WRITE
	})
}

func WithAnonymous() Option {
	return newFuncOption(func(o *options) {
		o.flags |= syscall.MAP_ANONYMOUS
	})
}

func WithPrivate() Option {
	return newFuncOption(func(o *options) {
		o.flags |= syscall.MAP_PRIVATE
	})
}

func WithOffset(offset uint64) Option {
	return newFuncOption(func(o *options) {
		o.offset = offset
	})
}

type FileMapping struct {
	data []byte
}

func OpenFileMapping(path string, size int, opt ...Option) (*FileMapping, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	opts := &options{
		prot:  syscall.PROT_READ,
		flags: syscall.MAP_SHARED,
	}
	for _, o := range opt {
		o.apply(opts)
	}

	data, err := syscall.Mmap(int(file.Fd()), int64(opts.offset), size, int(opts.prot), int(opts.flags))
	if err != nil {
		return nil, err
	}

	return &FileMapping{data: data}, nil
}

func (f *FileMapping) Close() error {
	return syscall.Munmap(f.data)
}
