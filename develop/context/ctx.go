package main

import (
	"context"
	"fmt"
	"time"
)

type privateCtx struct {
	context.Context
}

func main() {
	parentCtx, parentFunc := context.WithCancel(context.Background())

	ctx := privateCtx{parentCtx}

	childCtx, childFunc := context.WithCancel(ctx)

	childCancel := false

	if childCancel {
		childFunc()
	} else {
		parentFunc()
	}

	fmt.Println(parentCtx)
	fmt.Println(ctx)
	fmt.Println(childCtx)

	time.Sleep(time.Second)
}
