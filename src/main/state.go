package main

import (
	"unsafe"
	"sync"
	"time"
	"fmt"
)

func statewc(state [3]uint32) (statep *uint64, semap *uint32) {
	if uintptr(unsafe.Pointer(&state))%8 == 0 {
		return (*uint64)(unsafe.Pointer(&state)), &state[2]
	} else {
		return (*uint64)(unsafe.Pointer(&state[1])), &state[0]
	}
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			time.Sleep(10 * time.Millisecond)
			fmt.Println("Hello")
			wg.Done()
			wg.Wait()
		}()
		time.Sleep(500 * time.Millisecond)
	}

	wg.Wait()
}
