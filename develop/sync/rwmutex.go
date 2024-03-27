package main

import (
	"fmt"
	"sync"
	"time"
)

// Lock() 时, readerCount 值要变为0, 否则自身要休眠等待
func case1() {
	var rw sync.RWMutex
	_ = sync.Pool{}
	go func() {
		defer rw.RUnlock()
		rw.RLock()
		time.Sleep(time.Second)
		fmt.Println("B")
	}()

	time.Sleep(time.Millisecond)
	rw.Lock() // 等待 RLock 的释放.
	fmt.Println("A")
}

// Unlock() 时, readerCount 值要变为0, 否则要唤醒所有的等待的 goroutine
func case2() {
	var rw sync.RWMutex
	rw.Lock()

	go func() {
		rw.RLock() // 当已经存在写锁的时, 需要休眠等待
		fmt.Println("B")
	}()

	time.Sleep(time.Millisecond)
	rw.Unlock()
	fmt.Println("A")
	time.Sleep(time.Millisecond)
}

func main() {
	case1()
	case2()
}
