package pool

import (
	"sync"
	"io"
	"errors"
)

var (
	ErrPoolClosed = errors.New("")
)

type Pool struct {
	m       sync.Mutex                // 多个goroutine访问安全, closed的线程安全
	res     chan io.Closer            // 连接存储
	factory func() (io.Closer, error) // 工厂方法
	closed  bool                      // 连接池关闭
}

func New(f func() (io.Closer, error), size uint) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("size value too small")
	}

	return &Pool{
		factory: f,
		res:     make(chan io.Closer, size),
	}, nil
}

func (p *Pool) Acquire() (io.Closer, error) {
	select {
	case r, ok := <-p.res:
		if !ok {
			return nil, ErrPoolClosed
		}
		return r, nil
	default:
		return p.factory()
	}
}

func (p *Pool) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	// close chan
	close(p.res)

	// close chan resource
	for r:= range p.res {
		r.Close()
	}
}

