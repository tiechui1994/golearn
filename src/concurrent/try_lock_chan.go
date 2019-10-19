package concurrent

import "time"

type ChanLock struct {
	ch chan struct{}
}

func NewChanLock() *ChanLock {
	return &ChanLock{
		ch: make(chan struct{}, 1),
	}
}

func (c *ChanLock) Lock() {
	c.ch <- struct{}{}
}

func (c *ChanLock) Unlock() {
	select {
	case <-c.ch:
	default:
		panic("")
	}
}

func (c *ChanLock) Trylock() bool {
	select {
	case c.ch <- struct{}{}:
		return true
	default:
	}
	return false
}

func (c *ChanLock) TryLockTimeOut(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	select {
	case c.ch <- struct{}{}:
		timer.Stop()
		return true
	case <-time.After(timeout):
		return false
	}
}
