package ext

import (
	"sync"
	"container/list"
	"context"
)

// 信号量, 链表记录等待者 + context的方式实现阻塞方式(ctx+channel)的获取
type Semaphore struct {
	cur     int64
	size    int64
	m       sync.Mutex
	waiters list.List
}

type waiter struct {
	ready chan struct{}
	n     int64
}

func NewSemaphore(size int64) *Semaphore {
	return &Semaphore{size: size}
}

func (s *Semaphore) Acquire(ctx context.Context, n int64) error {
	s.m.Lock()
	if s.size-s.cur >= n && s.waiters.Len() == 0 {
		s.cur += n
		s.m.Unlock()
		return nil
	}

	// 无法完成, 只能等待ctx的结束了
	if s.size < n {
		s.m.Unlock()
		<-ctx.Done()
		return ctx.Err()
	}

	ready := make(chan struct{})
	w := waiter{n: n, ready: ready}
	ele := s.waiters.PushBack(w)
	s.m.Unlock()

	select {
	case <-ctx.Done():
		s.m.Lock()
		err := ctx.Err()
		select {
		case <-ready:
			err = nil
		default:
			s.waiters.Remove(ele)
		}
		s.m.Unlock()
		return err
	case <-ready:
		return nil
	}
}

func (s *Semaphore) TryAcquire(n int64) bool {
	s.m.Lock()
	result := s.size-s.cur >= n && s.waiters.Len() == 0
	if result {
		s.cur += n
		s.m.Unlock()
		return true
	}

	s.m.Unlock()
	return false
}

func (s *Semaphore) Release(n int64) {
	s.m.Lock()
	s.cur -= n
	if s.cur < 0 {
		s.m.Unlock()
		panic("~~~~~~~~~~")
	}

	for {
		next := s.waiters.Front()
		if next == nil {
			break
		}

		w := next.Value.(waiter)

		if s.size-s.cur < w.n {
			break
		}

		s.cur += w.n
		close(w.ready)
		s.waiters.Remove(next)
	}
	s.m.Unlock()
}
