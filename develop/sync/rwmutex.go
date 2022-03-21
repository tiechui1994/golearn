package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var rw sync.RWMutex

	go func() {
		defer rw.RUnlock()
		rw.RLock()
		time.Sleep(time.Second)
		fmt.Println("B")
	}()

	time.Sleep(time.Millisecond)
	rw.Lock() // 等待 RLock 的释放.
}
