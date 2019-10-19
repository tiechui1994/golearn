package ext

import (
	"context"
	"sync"
)

// 使用WaitGroup和Once(error), Context去实现ErrGroup
// 原则, 当第一个goroutine失败之后, wait立即返回
type Group struct {
	cancel func()

	wg sync.WaitGroup

	errOnce sync.Once
	err     error
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	child, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, child
}

func (g *Group) Go(fn func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := fn(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}
