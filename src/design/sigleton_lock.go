package design

import "sync"

type sigleton struct {
}

var (
	instance *sigleton
	lock     sync.Mutex
)

func GetInstance() *sigleton {
	if instance == nil {
		lock.Lock()
		if instance == nil {
			instance = &sigleton{}
		}
		lock.Unlock()
	}

	return instance
}
