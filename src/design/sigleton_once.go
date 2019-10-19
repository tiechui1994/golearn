package design

import "sync"

type sigleton struct {
}

var (
	once     sync.Once
	instance *sigleton
)

func GetInstance() *sigleton {
	if instance == nil {
		once.Do(func() {
			instance = &sigleton{}
		})
	}

	return instance
}
