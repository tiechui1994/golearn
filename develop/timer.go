package main

import (
	"container/heap"
	"sync"
	"time"
)

// 模拟代表仅通过编程方式向前移动的模拟时钟.
// 测试基于时间的功能时, 它可能比实时时钟更好.
type Mock struct {
	sync.Mutex
	now    time.Time // current time
	timers Timers    // timers
}

func NewMock() *Mock {
	return &Mock{now: time.Unix(0, 0)}
}

// Add将 "模拟时钟" 的当前时间向前移动持续时间.
func (m *Mock) Add(d time.Duration) {
	m.Lock()
	end := m.now.Add(d) // 计算最终的时间

	// 对于在 end 之前的所有的时间将被移除, 同时触发相应的事件
	for len(m.timers) > 0 && m.now.Before(end) {
		t := heap.Pop(&m.timers).(*Timer)
		m.now = t.next
		m.Unlock()
		t.Tick() // 触发 ticket 返回
		m.Lock()
	}

	m.Unlock()
	nap() // 给一个小的缓冲区, 以确保其他goroutines得到处理.
}

// Timer produces a timer that will emit a time some duration after now.
// After, AfterFunc, Sleep 都是基于此实现的
func (m *Mock) Timer(d time.Duration) *Timer {
	ch := make(chan time.Time)
	t := &Timer{
		C:    ch,
		c:    ch,
		next: m.now.Add(d),
	}
	m.addTimer(t) // 添加 timer, 即将新创建的 timer 加入到 heap 当中
	return t
}

func (m *Mock) addTimer(t *Timer) {
	m.Lock()
	defer m.Unlock()
	heap.Push(&m.timers, t)
}

func (m *Mock) After(d time.Duration) <-chan time.Time {
	return m.Timer(d).C
}

func (m *Mock) AfterFunc(d time.Duration, f func()) *Timer {
	t := m.Timer(d)
	go func() {
		<-t.c
		f()
	}()
	nap()
	return t
}

func (m *Mock) Sleep(d time.Duration) {
	<-m.After(d)
}

// Now returns the current wall time on the mock clock.
func (m *Mock) Now() time.Time {
	m.Lock()
	defer m.Unlock()
	return m.now
}

// Timer represents a single event.
type Timer struct {
	C    <-chan time.Time
	c    chan time.Time
	next time.Time // next tick time
}

func (t *Timer) Next() time.Time { return t.next }

func (t *Timer) Tick() {
	select {
	case t.c <- t.next:
	default:
	}
	nap()
}

// Sleep momentarily so that other goroutines can process.
func nap() { time.Sleep(1 * time.Millisecond) }

// timers represents a list of sortable timers.
type Timers []*Timer

func (ts Timers) Len() int { return len(ts) }

func (ts Timers) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

func (ts Timers) Less(i, j int) bool {
	return ts[i].Next().Before(ts[j].Next())
}

func (ts *Timers) Push(t interface{}) {
	*ts = append(*ts, t.(*Timer))
}

func (ts *Timers) Pop() interface{} {
	t := (*ts)[len(*ts)-1]
	*ts = (*ts)[:len(*ts)-1]
	return t
}
