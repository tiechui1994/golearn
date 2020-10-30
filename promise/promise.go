package promise

import (
	"math/rand"
	"sync/atomic"
	"unsafe"
)

/**************************************************************
开启一个goroutine用来执行一个act函数并返回一个Future(act执行的结果).
如果option参数是true, act函数会被异步调用.
act的函数类型可以是以下4种:
  func() (r interface{}, err error)
  func()
  func(c promise.Canceller) (r interface{}, err error)
     c可以调用c.IsCancelled()方法退出函数执行
  func(promise.Canceller)
***************************************************************/
func Start(action interface{}, syncs ...bool) *Future {
	if f, ok := action.(*Future); ok {
		return f
	}

	promise := NewPromise()
	if proxy := getProxy(promise, action); proxy != nil {
		if syncs != nil && len(syncs) > 0 && syncs[0] {
			// sync
			result, err := proxy()
			if promise.IsCancelled() {
				promise.Cancel()
				return promise.Future
			}

			if err == nil {
				promise.Resolve(result)
			} else {
				promise.Reject(err)
			}

		} else {
			// async
			go func() {
				r, err := proxy()
				if promise.IsCancelled() {
					promise.Cancel()
					return
				}

				if err == nil {
					promise.Resolve(r)
				} else {
					promise.Reject(err)
				}
			}()
		}
	}

	return promise.Future
}

func getProxy(promise *Promise, action interface{}) (proxy func() (interface{}, error)) {
	var (
		canCancel bool
		func1     func() (interface{}, error)
		func2     func(Canceller) (interface{}, error)
	)

	switch v := action.(type) {
	case func() (interface{}, error):
		canCancel = false
		func1 = v

	case func(Canceller) (interface{}, error):
		canCancel = true
		func2 = v

	case func():
		canCancel = false
		func1 = func() (interface{}, error) {
			v()
			return nil, nil
		}

	case func(Canceller):
		canCancel = true
		func2 = func(canceller Canceller) (interface{}, error) {
			v(canceller)
			return nil, nil
		}

	default:
		if e, ok := v.(error); !ok {
			promise.Resolve(v)
		} else {
			promise.Reject(e)
		}
		return nil
	}

	// 当action函数带有参数Canceller, 则Future将来可以被取消
	var canceller Canceller = nil
	if promise != nil && canCancel {
		canceller = promise.Canceller()
	}

	// 返回代理action的函数
	proxy = func() (result interface{}, err error) {
		defer func() {
			if e := recover(); e != nil {
				err = newErrorWithStacks(e)
			}
		}()

		if canCancel {
			result, err = func2(canceller)
		} else {
			result, err = func1()
		}

		return result, err
	}

	return proxy
}

func Wrap(value interface{}) *Future {
	promise := NewPromise()
	if e, ok := value.(error); !ok {
		promise.Resolve(value)
	} else {
		promise.Reject(e)
	}

	return promise.Future
}

// 如果任何一个Future执行成功, 当前的Future也将会执行成功,并且返回已经成功执行的Future的值;
// 否则, 当前的Future将会执行失败, 并且返回所有Future的执行结果.
func WhenAny(actions ...interface{}) *Future {
	return WhenAnyMatched(nil, actions...)
}

// 如果任何一个Future执行成功并且predicate()函数执返回true, 当前的Future也将会执行成功,并且返回已经成功执行的Future的值.
// 如果所有的Future都被取消, 当前的Future也会被取消; 否则, 当前的Future将会执行失败NoMatchedError, 并且返回所有Future的执行结果.
func WhenAnyMatched(predicate func(interface{}) bool, actions ...interface{}) *Future {
	if predicate == nil {
		predicate = func(v interface{}) bool { return true }
	}

	// build to Future
	futures := make([]*Future, len(actions))
	for i, act := range actions {
		futures[i] = Start(act)
	}

	// results
	promise := NewPromise()
	if len(actions) == 0 {
		promise.Resolve(nil)
		return promise.Future
	}

	fails, dones := make(chan anyPromiseResult), make(chan anyPromiseResult)
	go func() {
		// register callback for everyone
		for i, future := range futures {
			k := i
			future.OnSuccess(func(v interface{}) {
				defer func() { _ = recover() }()
				dones <- anyPromiseResult{v, k}
			}).OnFailure(func(v interface{}) {
				defer func() { _ = recover() }()
				fails <- anyPromiseResult{v, k}
			}).OnCancel(func() {
				defer func() { _ = recover() }()
				fails <- anyPromiseResult{CANCELLED, k}
			})
		}
	}()

	// sync exec
	if len(futures) == 1 {
		select {
		case fail := <-fails:
			if _, ok := fail.value.(CancelledError); ok {
				promise.Cancel()
			} else {
				promise.Reject(newNoMatchedError(fail.value))
			}
		case done := <-dones:
			if predicate(done.value) {
				promise.Resolve(done.value)
			} else {
				promise.Reject(newNoMatchedError(done.value))
			}
		}

		return promise.Future
	}

	// async exec
	results := make([]interface{}, len(futures))
	go func() {
		defer func() {
			if e := recover(); e != nil {
				promise.Reject(newErrorWithStacks(e))
			}
		}()

	loop:
		for j := 0; ; {
			select {
			case fail := <-fails:
				results[fail.index] = getError(fail.value)

			case done := <-dones:
				if predicate(done.value) {
					// meet expect result, cancel other task and return
					for _, future := range futures {
						future.Cancel()
					}

					// close chan
					closeChan := func(c chan anyPromiseResult) {
						defer func() { _ = recover() }()
						close(c)
					}
					closeChan(dones)
					closeChan(fails)

					// 成功执行并且返回
					promise.Resolve(done.value)
					break loop

				} else {
					results[done.index] = done.value
				}
			}

			// exec n times
			if j++; j == len(futures) {
				m := 0 // count not cancel result
				for _, result := range results {
					switch result.(type) {
					case CancelledError:
					default:
						m++
					}
				}

				if m > 0 {
					// no cancel result
					promise.Reject(newNoMatchedError(results))
				} else {
					// all canceled
					promise.Cancel()
				}

				break loop
			}
		}
	}()

	return promise.Future
}

// 如果所有的Future都成功执行, 当前的Future也会成功执行并且返回相应的结果数组(成功执行的Future的结果);
// 否则, 当前的Future将会执行失败, 并且返回所有Future的执行结果.
func WhenAll(actions ...interface{}) *Future {
	promise := NewPromise()

	if len(actions) == 0 {
		promise.Resolve([]interface{}{})
		return promise.Future
	}

	// build future
	futures := make([]*Future, len(actions))
	for i, act := range actions {
		futures[i] = Start(act)
	}

	// can execute once
	n := int32(len(futures))
	cancel := func(j int) {
		for k, future := range futures {
			if k != j {
				future.Cancel()
			}
		}
	}

	// exec
	results := make([]interface{}, len(futures))
	go func() {
		isCancelled := int32(0)
		for i, future := range futures {
			j := i
			// register call back for evenry one
			future.OnSuccess(func(v interface{}) {
				results[j] = v
				if atomic.AddInt32(&n, -1) == 0 {
					promise.Resolve(results)
				}
			}).OnFailure(func(v interface{}) {
				// CAS
				if atomic.CompareAndSwapInt32(&isCancelled, 0, 1) {
					cancel(j)
					promise.Reject(newAggregateError("Error appears in WhenAll:", v))
				}
			}).OnCancel(func() {
				// CAS
				if atomic.CompareAndSwapInt32(&isCancelled, 0, 1) {
					cancel(j)
					promise.Cancel()
				}
			})
		}
	}()

	return promise.Future
}

// only used in WhenAnyMatched()/WhenAny() and
// store intermediate results
type anyPromiseResult struct {
	value interface{}
	index int
}

// Future最终的状态
type resultType int

const (
	RESULT_SUCCESS resultType = iota
	RESULT_FAILURE
	RESULT_CANCELLED
)

type PromiseResult struct {
	Value interface{}
	Type  resultType
}

/*********************************************************************
1. Promise提供一个对象作为结果的代理. 这个结果最初是未知的, 通常是因为其值尚未被计算出.
2. 可以使用Resolve() | Reject() | Cancel() 来设置Promise的最终结果.
3. Future只返回一个带有只读占位符视图.
*********************************************************************/
type Promise struct {
	*Future
}

// cancel exeute
func (promise *Promise) Cancel() (e error) {
	return promise.setResult(&PromiseResult{CANCELLED, RESULT_CANCELLED})
}

// set result to success
func (promise *Promise) Resolve(v interface{}) (e error) {
	return promise.setResult(&PromiseResult{v, RESULT_SUCCESS})
}

// set result to failure
func (promise *Promise) Reject(err error) (e error) {
	return promise.setResult(&PromiseResult{err, RESULT_FAILURE})
}

// Callback when success
func (promise *Promise) OnSuccess(callback func(v interface{})) *Promise {
	promise.Future.OnSuccess(callback)
	return promise
}

// Callback when failure
func (promise *Promise) OnFailure(callback func(v interface{})) *Promise {
	promise.Future.OnFailure(callback)
	return promise
}

// Callback when success or failure
func (promise *Promise) OnComplete(callback func(v interface{})) *Promise {
	promise.Future.OnComplete(callback)
	return promise
}

// Callback when cancel
func (promise *Promise) OnCancel(callback func()) *Promise {
	promise.Future.OnCancel(callback)
	return promise
}

func NewPromise() *Promise {
	value := &value{
		dones:   make([]func(v interface{}), 0, 8),
		fails:   make([]func(v interface{}), 0, 8),
		always:  make([]func(v interface{}), 0, 4),
		cancels: make([]func(), 0, 2),
		pipes:   make([]*pipe, 0, 4),
		result:  nil,
	}

	promise := &Promise{
		Future: &Future{
			ID:    rand.Int(),
			final: make(chan struct{}),
			val:   unsafe.Pointer(value),
		},
	}

	return promise
}
