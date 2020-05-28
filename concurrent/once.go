package concurrent

import (
	"sync"
	"sync/atomic"
)

type Once struct {
	done uint32
	m    sync.Mutex
}

func (o *Once) Do(f func()) {
	if atomic.LoadUint32(&o.done) == 0 {
		o.m.Lock()
		if o.done == 0 {
			f()
			atomic.StoreUint32(&o.done, 1)
		}
		o.m.Unlock()
	}
}
