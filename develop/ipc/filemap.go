package ipc

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	defaultMemMapSize = 8 * 1024
)

type options struct {
	offset uint64
	prot   uint32
	flags  uint32
	size   uint64
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

func WithSize(size uint64) Option {
	return newFuncOption(func(o *options) {
		if size < defaultMemMapSize {
			size = defaultMemMapSize
		}

		size = defaultMemMapSize - size&(defaultMemMapSize-1) + size // 8k
		o.size = size
	})
}

func WithOffset(offset uint64) Option {
	return newFuncOption(func(o *options) {
		o.offset = offset
	})
}

type FileMap struct {
	mapping []byte
	data    []byte
	fd      *os.File
}

func OpenFileMap(path string, opt ...Option) (*FileMap, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	opts := &options{
		prot:  syscall.PROT_READ,
		flags: syscall.MAP_SHARED,
		size:  defaultMemMapSize,
	}
	for _, o := range opt {
		o.apply(opts)
	}

	data, err := syscall.Mmap(int(file.Fd()), int64(opts.offset), int(opts.size), int(opts.prot), int(opts.flags))
	if err != nil {
		return nil, err
	}

	f := &FileMap{data: data, fd: file}

	err = f.Grow(defaultMemMapSize)
	if err != nil {
		return nil, err
	}

	f.mapping = (*(*[4096]uint8)(unsafe.Pointer(&data[0])))[:]
	return f, nil
}

func (f *FileMap) Grow(size uint64) error {
	if info, _ := f.fd.Stat(); info.Size() >= int64(size) {
		return nil
	}

	nsize := int64(defaultMemMapSize - size&(defaultMemMapSize-1) + size)
	return f.fd.Truncate(nsize)
}

func (f *FileMap) Close() error {
	return syscall.Munmap(f.data)
}
