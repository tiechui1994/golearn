package concurrent

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	mutexLocked      = 1 << iota // mutex is locked
	mutexWoken
	mutexStarving
	mutexWaiterShift = iota
)

type Mutext struct {
	mu sync.Mutex
}

func (m *Mutext) Lock() {
	m.mu.Lock()
}

func (m *Mutext) Unlock() {
	m.mu.Unlock()
}

func (m *Mutext) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.mu)), 0, mutexLocked)
}

func (m *Mutext) IsLocked() bool {
	return atomic.LoadInt32((*int32)(unsafe.Pointer(&m.mu)))&mutexLocked == mutexLocked
}

func (m *Mutext) IsStarving() bool {
	return atomic.LoadInt32((*int32)(unsafe.Pointer(&m.mu)))&mutexStarving == mutexStarving
}

func (m *Mutext) Counter() uint32 {
	state := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.mu)))
	v := state >> mutexWaiterShift
	v = v + (v & mutexLocked)
	return v
}
