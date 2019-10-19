package ext

import "sync"

// 单飞, 使用map维护调用的call, 当存在call,就会直接返回
// Do(), call使用的是wg的方式, 主要是利用了wg.Wait()的通知
// DoChan(), call使用的是chan的异步通知的特性
type SigleFlight struct {
	m  map[string]*call
	mu sync.Mutex
}

type call struct {
	wg sync.WaitGroup

	err error
	val interface{}

	forgotten bool //
	dumps     int  // shared

	chs []chan<- Result
}

type Result struct {
	Err    error
	Val    interface{}
	Shared bool
}

func (s *SigleFlight) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	s.mu.Lock()
	if s.m == nil {
		s.m = make(map[string]*call)
	}

	if val, ok := s.m[key]; ok {
		val.dumps++
		s.mu.Unlock()
		val.wg.Wait()
		return val.val, val.err
	}

	call := new(call)
	call.wg.Add(1)
	s.m[key] = call
	s.mu.Unlock()

	s.doCall(call, key, fn)
	return call.val, call.err
}

func (s *SigleFlight) doCall(call *call, key string, fn func() (interface{}, error)) {
	call.val, call.err = fn()
	call.wg.Done()

	s.mu.Lock()
	if !call.forgotten {
		delete(s.m, key)
	}
	for _, ch := range call.chs {
		ch <- Result{Val: call.val, Err: call.err, Shared: call.dumps > 0}
	}
	s.mu.Unlock()
}

func (s *SigleFlight) DoChan(key string, fn func() (interface{}, error)) <-chan Result {
	result := make(chan Result, 1) // 异步

	s.mu.Lock()
	if s.m == nil {
		s.m = make(map[string]*call)
	}

	if val, ok := s.m[key]; ok {
		val.dumps++
		val.chs = append(val.chs, result)
		s.mu.Unlock()
		return result
	}

	call := &call{chs: []chan<- Result{result}}
	call.wg.Add(1)
	s.m[key] = call
	s.mu.Unlock()

	go s.doCall(call, key, fn)

	return result
}

func (s *SigleFlight) Forget(key string) {
	s.mu.Lock()
	if c, ok := s.m[key]; ok {
		c.forgotten = true
	}
	delete(s.m, key)
	s.mu.Unlock()
}
