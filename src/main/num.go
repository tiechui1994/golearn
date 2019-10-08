package main

import (
	"fmt"
	"time"
)

func Processor(done chan<- interface{}, seq <-chan int, level int) {
	go func() {
		if level == 5001 {
			done <- "done"
			return
		}
		prime, ok := <-seq
		if !ok {
			done <- "done"
			return
		}
		fmt.Printf("[%d]: %d\n", level, prime)
		out := make(chan int)
		defer close(out)
		Processor(done, out, level+1)
		for num := range seq {
			if num%prime != 0 {
				out <- num
			}
		}
	}()
}

func main() {
	now := time.Now()
	origin, done := make(chan int), make(chan interface{})
	Processor(done, origin, 1)
	for num := 2; num < 50000; num++ {
		origin <- num
	}
	close(origin)
	<-done
	fmt.Println(time.Now().Sub(now).Seconds())
}
