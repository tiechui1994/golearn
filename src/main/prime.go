package main

import (
	"fmt"
	"time"
)

func Seq() chan int {
	seq := make(chan int)
	go func() {
		for i := 2; ; i++ {
			seq <- i
		}
	}()
	return seq
}
func Prime(in <-chan int, prime int) chan int {
	out := make(chan int)
	go func() {
		for {
			if i := <-in; i%prime != 0 {
				out <- i
			}
		}
	}()
	return out
}

func main() {
	now := time.Now()
	seq := Seq()

	for i := 0; i < 5000; i++ {
		prime := <-seq
		fmt.Printf("[%d]: %d \n", i+1, prime)
		seq = Prime(seq, prime)
	}

	fmt.Println(time.Now().Sub(now).Seconds())
}
