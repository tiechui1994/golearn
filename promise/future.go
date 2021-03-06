package promise

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"
)

type callbackType int

const (
	CALLBACK_DONE callbackType = iota
	CALLBACK_FAIL
	CALLBACK_ALWAYS
	CALLBACK_CANCEL
)

type Task func(interface{}) *Future

// pipe 是 Promise的链式执行
type pipe struct {
	doneTask, failTask Task
	promise            *Promise
}

// getPipe returns piped Future task function and pipe Promise by the status of current Promise.
func (pipe *pipe) getPipe(isResolved bool) (Task, *Promise) {
	if isResolved {
		return pipe.doneTask, pipe.promise
	} else {
		return pipe.failTask, pipe.promise
	}
}

//--------------------------------------------------------------------------------------------------

type Canceller interface {
	IsCancelled() bool
	Cancel()
}

type canceller struct {
	future *Future
}

func (cancel *canceller) Cancel() {
	cancel.future.Cancel()
}
func (cancel *canceller) IsCancelled() bool {
	return cancel.future.IsCancelled()
}

//--------------------------------------------------------------------------------------------------

// Future val
type value struct {
	dones, fails, always []func(v interface{})
	cancels              []func()
	pipes                []*pipe
	result               *PromiseResult
}

// Future 提供的是一个只读的Promise的视图. 它的值在调用Promise的 Resolve | Reject | Cancel 方法之后被确定
type Future struct {
	ID    int // Future 的唯一标识
	final chan struct{}
	val   unsafe.Pointer // val 指向 value 的指针地址, 使用CAS方法原子修改它的值
}

// return to user need value
func (future *Future) getResultValue(r *PromiseResult) (interface{}, error) {
	if r.Type == RESULT_SUCCESS {
		return r.Value, nil
	} else if r.Type == RESULT_FAILURE {
		return nil, getError(r.Value)
	} else {
		return nil, getError(r.Value)
	}
}

func (future *Future) loadResult() *PromiseResult {
	value := future.loadVal()
	return value.result
}

func (future *Future) loadVal() *value {
	return (*value)(atomic.LoadPointer(&future.val))
}

// shared future
func (future *Future) Canceller() Canceller {
	return &canceller{future}
}

func (future *Future) IsCancelled() bool {
	val := future.loadVal()
	return val != nil && val.result != nil && val.result.Type == RESULT_CANCELLED
}

// timeout unit is ms,
func (future *Future) SetTimeout(timeout int) *Future {
	if timeout == 0 {
		timeout = 10
	} else {
		timeout = timeout * 1000 * 1000
	}

	go func() {
		<-time.After((time.Duration)(timeout) * time.Nanosecond)
		future.Cancel()
	}()

	return future
}

func (future *Future) GetChan() <-chan *PromiseResult {
	c := make(chan *PromiseResult, 1)

	future.OnComplete(func(interface{}) {
		c <- future.loadResult()
	}).OnCancel(func() {
		c <- future.loadResult()
	})

	return c
}

// get result
func (future *Future) Get() (interface{}, error) {
	<-future.final
	return future.getResultValue(future.loadResult())
}

// in timeout time get result
func (future *Future) GetOrTimeout(timeout uint) (interface{}, error, bool) {
	if timeout == 0 {
		timeout = 10
	} else {
		timeout = timeout * 1000 * 1000
	}

	select {
	case <-time.After((time.Duration)(timeout) * time.Nanosecond):
		return nil, nil, true
	case <-future.final:
		result, err := future.getResultValue(future.loadResult())
		return result, err, false
	}
}

// cancel execute
func (future *Future) Cancel() (e error) {
	return future.setResult(&PromiseResult{CANCELLED, RESULT_CANCELLED})
}

// register callback
func (future *Future) OnSuccess(callback func(v interface{})) *Future {
	future.callback(callback, CALLBACK_DONE)
	return future
}
func (future *Future) OnFailure(callback func(v interface{})) *Future {
	future.callback(callback, CALLBACK_FAIL)
	return future
}
func (future *Future) OnComplete(callback func(v interface{})) *Future {
	future.callback(callback, CALLBACK_ALWAYS)
	return future
}
func (future *Future) OnCancel(callback func()) *Future {
	future.callback(callback, CALLBACK_CANCEL)
	return future
}

// callbacks: onSuccess, onFailure
func (future *Future) Pipe(callbacks ...interface{}) (new *Future, ok bool) {
	if len(callbacks) == 0 ||
		(len(callbacks) == 1 && callbacks[0] == nil) ||
		(len(callbacks) > 1 && callbacks[0] == nil && callbacks[1] == nil) {
		new = future
		return
	}

	// 验证回调函数的格式 => func(interface{}) *Future
	// 6 种情况, 分别是: 参数[interface, empty], 结果[Future, empty, withError]
	tasks := make([]Task, len(callbacks))
	for i, callback := range callbacks {
		if function, ok := callback.(Task); ok {
			// func(interface{}) *Future
			tasks[i] = function

		} else if function, ok := callback.(func() *Future); ok {
			// func() *Future
			tasks[i] = func(interface{}) *Future {
				return function()
			}

		} else if function, ok := callback.(func(interface{})); ok {
			// func(interface{})
			tasks[i] = func(v interface{}) *Future {
				return Start(func() { function(v) })
			}

		} else if function, ok := callback.(func(interface{}) (interface{}, error)); ok {
			// func(interface{}) (interface{}, error)
			tasks[i] = func(v interface{}) *Future {
				return Start(func() (interface{}, error) { return function(v) })
			}

		} else if function, ok := callback.(func()); ok {
			// func()
			tasks[i] = func(v interface{}) *Future {
				return Start(func() { function() })
			}

		} else if function, ok := callback.(func() (interface{}, error)); ok {
			// func() (interface{}, error)
			tasks[i] = func(v interface{}) *Future {
				return Start(func() (interface{}, error) { return function() })
			}

		} else {
			return nil, false
		}
	}

	for {
		val := future.loadVal()
		result := val.result // 对于初始化的Promise,其result值是nil

		if result != nil {
			// 执行完成
			if result.Type == RESULT_SUCCESS && tasks[0] != nil {
				new = tasks[0](result.Value)
			} else if result.Type == RESULT_FAILURE && len(tasks) > 1 && tasks[1] != nil {
				new = tasks[1](result.Value)
			} else {
				new = future
			}

		} else {
			// 执行中...
			p := &pipe{promise: NewPromise()}
			p.doneTask = tasks[0]
			if len(tasks) > 1 {
				p.failTask = tasks[1]
			}

			newval := *val
			newval.pipes = append(val.pipes, p)

			// CAS
			if atomic.CompareAndSwapPointer(&future.val, unsafe.Pointer(val), unsafe.Pointer(&newval)) {
				new = p.promise.Future
				break
			}
		}
	}

	return new, true
}

// add callback
func (future *Future) callback(callback interface{}, t callbackType) {
	if callback == nil {
		return
	}

	// callback match type
	if (t == CALLBACK_DONE) || (t == CALLBACK_FAIL) || (t == CALLBACK_ALWAYS) {
		if _, ok := callback.(func(interface{})); !ok {
			panic(errors.New("callback function spec must be func(v interface{})"))
		}
	} else if t == CALLBACK_CANCEL {
		if _, ok := callback.(func()); !ok {
			panic(errors.New("callback function spec must be func()"))
		}
	}

	for {
		val := future.loadVal()
		result := val.result
		if result == nil {
			// 执行中...
			newval := *val
			switch t {
			case CALLBACK_DONE:
				val.dones = append(val.dones, callback.(func(interface{})))
			case CALLBACK_FAIL:
				val.fails = append(val.fails, callback.(func(interface{})))
			case CALLBACK_ALWAYS:
				val.always = append(val.always, callback.(func(interface{})))
			case CALLBACK_CANCEL:
				val.cancels = append(val.cancels, callback.(func()))
			}

			// CAS
			if atomic.CompareAndSwapPointer(&future.val, unsafe.Pointer(val), unsafe.Pointer(&newval)) {
				break
			}
		} else {
			// 执行结束
			if (t == CALLBACK_DONE && result.Type == RESULT_SUCCESS) ||
				(t == CALLBACK_FAIL && result.Type == RESULT_FAILURE) ||
				(t == CALLBACK_ALWAYS && result.Type != RESULT_CANCELLED) {
				callback.(func(interface{}))(result.Value)
			} else if t == CALLBACK_CANCEL && result.Type == RESULT_CANCELLED {
				callback.(func())()
			}

			return
		}
	}
}

// call can do once
func (future *Future) setResult(result *PromiseResult) (e error) {
	defer func() {
		if err := getError(recover()); err != nil {
			e = err
		}
	}()

	e = errors.New("cannot resolve/reject/cancel more than once")
	for {
		val := future.loadVal()
		if val.result != nil {
			return
		}

		newval := *val
		newval.result = result

		// CAS
		if atomic.CompareAndSwapPointer(&future.val, unsafe.Pointer(val), unsafe.Pointer(&newval)) {
			// close final make sure Get() and GetOrTimeout() return
			close(future.final)

			// call callback functions
			if len(val.dones) > 0 || len(val.fails) > 0 || len(val.always) > 0 || len(val.cancels) > 0 {
				go func() {
					execCallback(result, val.dones, val.fails, val.always, val.cancels)
				}()
			}

			// start the pipeline
			if len(val.pipes) > 0 {
				go func() {
					for _, pipe := range val.pipes {
						task, promise := pipe.getPipe(result.Type == RESULT_SUCCESS)
						execPipe(result, task, promise)
					}
				}()
			}

			return nil
		}
	}

	return e
}

// 执行回调函数
func execCallback(r *PromiseResult,
	dones []func(v interface{}),
	fails []func(v interface{}),
	always []func(v interface{}),
	cancels []func()) {

	if r.Type == RESULT_CANCELLED {
		for _, f := range cancels {
			func() {
				defer func() {
					if e := recover(); e != nil {
						err := newErrorWithStacks(e)
						fmt.Println("error happens:\n ", err)
					}
				}()
				f()
			}()
		}
		return
	}

	var callbacks []func(v interface{})
	if r.Type == RESULT_SUCCESS {
		callbacks = dones
	} else {
		callbacks = fails
	}

	forFs := func(s []func(interface{})) {
		forSlice(s, func(f func(interface{})) { f(r.Value) })
	}

	forFs(callbacks)
	forFs(always)
}

func forSlice(s []func(v interface{}), f func(func(v interface{}))) {
	for _, e := range s {
		func() {
			defer func() {
				if e := recover(); e != nil {
					err := newErrorWithStacks(e)
					fmt.Println("error happens:\n ", err)
				}
			}()
			f(e)
		}()
	}
}

func execPipe(r *PromiseResult, task func(interface{}) *Future, promise *Promise) {
	if task != nil {
		task(r.Value).OnSuccess(func(v interface{}) {
			promise.Resolve(v)
		}).OnFailure(func(v interface{}) {
			promise.Reject(getError(v))
		})
	}
}
