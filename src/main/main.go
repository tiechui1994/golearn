package main

import (
	"sync"
	"fmt"
	"time"
	"runtime"
)

var wg sync.WaitGroup

func main() {
	a := make(chan int, 1)
	runtime.GOMAXPROCS(1)
	a <- 1

	for i := 1; i <= 10; i++ {
		go func(i int) {
			<-a
			fmt.Println(2*i - 1)
		}(i)
	}
	for i := 1; i <= 10; i++ {
		go func(i int) {
			a <- i
			fmt.Println(2 * i)
		}(i)
	}

	time.Sleep(1 * time.Second)
}
