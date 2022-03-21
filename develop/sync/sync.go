package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var lock sync.Mutex
	fmt.Print("a")
	lock.Lock()

	go func() {
		time.Sleep(200 * time.Millisecond)
		lock.Unlock()
	}()

	lock.Lock()
	fmt.Print("b")
}
