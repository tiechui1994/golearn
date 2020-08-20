package main

import (
	"fmt"
	"time"

	"go.uber.org/ratelimit"
)

func main() {
	fmt.Println(time.Time{}.IsZero())
	rl := ratelimit.New(1000000000000000000) // per second

	prev := time.Now()
	for i := 0; i < 10; i++ {
		now := rl.Take()
		fmt.Println(i, now.Sub(prev))
		prev = now
	}
}
