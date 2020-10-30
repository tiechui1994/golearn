package concurrent

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	mutexLocked = 1 << iota // mutex is locked
	mutexWoken
	mutexStarving
	mutexWaiterShift = iota
)

type Mutex struct {
	mu sync.Mutex
}

func (m *Mutex) Lock() {
	m.mu.Lock()
}

func (m *Mutex) Unlock() {
	m.mu.Unlock()
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.mu)), 0, mutexLocked)
}

func (m *Mutex) IsLocked() bool {
	return atomic.LoadInt32((*int32)(unsafe.Pointer(&m.mu)))&mutexLocked == mutexLocked
}

func (m *Mutex) IsStarving() bool {
	return atomic.LoadInt32((*int32)(unsafe.Pointer(&m.mu)))&mutexStarving == mutexStarving
}

func (m *Mutex) Counter() int32 {
	state := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.mu)))
	v := state >> mutexWaiterShift
	v = v + (v & mutexLocked)
	return v
}
