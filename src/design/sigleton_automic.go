package design

import (
	"sync"
	"sync/atomic"
)

type sigleton struct {
}

var (
	done     int32
	lock     sync.Mutex
	instance *sigleton
)

func GetInstance() *sigleton {
	if atomic.LoadInt32(&done) == 0 {
		lock.Lock()
		if instance == nil {
			instance = &sigleton{}
		}
		atomic.StoreInt32(&done, 1)
		lock.Unlock()
	}

	return instance
}
