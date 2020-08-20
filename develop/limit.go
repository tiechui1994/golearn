package main

import (
	"fmt"
	"time"

	"go.uber.org/ratelimit"
)

func main() {
	fmt.Println(time.Time{}.IsZero())
	rl := ratelimit.New(10000) // per second

	prev := time.Now()
	for i := 0; i < 10; i++ {
		now := rl.Take()
		println(i, now.Sub(prev))
		prev = now
	}
}
