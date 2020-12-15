package pkg

import (
	_ "unsafe"
	"sync"
	"time"
	"fmt"
)

func AddWithoutLock() {
	var count = 0
	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			count++
			time.Sleep(200 * time.Millisecond)
		}()
	}

	wg.Wait()
	fmt.Println(count)
}

type mutex struct {
	key uintptr
}

//go:linkname lock runtime.lock
func lock(l *mutex)

//go:linkname unlock runtime.unlock
func unlock(l *mutex)

func AddWithLock() {
	l := &mutex{}
	var count = 0
	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			lock(l)
			count++
			unlock(l)
			time.Sleep(200 * time.Millisecond)
		}()
	}

	wg.Wait()
	fmt.Println(count)
}
