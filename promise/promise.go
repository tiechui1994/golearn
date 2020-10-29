package promise

import (
	"sync/atomic"
	"unsafe"
	"math/rand"
)

type anyPromiseResult struct {
	result interface{}
	i      int
}

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
	promise := NewPromise()
	if f, ok := action.(*Future); ok {
		return f
	}

	if proxy := getAction(promise, action); proxy != nil {
		if syncs != nil && len(syncs) > 0 && !syncs[0] {
			// 同步调用
			result, err := proxy()
			if promise.IsCancelled() {
				promise.Cancel()
			} else {
				if err == nil {
					promise.Resolve(result)
				} else {
					promise.Reject(err)
				}
			}
		} else {
			// 异步调用
			go func() {
				r, err := proxy()
				if promise.IsCancelled() {
					promise.Cancel()
				} else {
					if err == nil {
						promise.Resolve(r)
					} else {
						promise.Reject(err)
					}
				}
			}()
		}
	}

	return promise.Future
}

// 包装Future
func Wrap(value interface{}) *Future {
	promise := NewPromise()
	if e, ok := value.(error); !ok {
		promise.Resolve(value)
	} else {
		promise.Reject(e)
	}

	return promise.Future
}

// 返回一个Future
// 如果任何一个Future执行成功, 当前的Future也将会执行成功,并且返回已经成功执行的Future的值; 否则,
// 当前的Future将会执行失败, 并且返回所有Future的执行结果.
func WhenAny(actions ...interface{}) *Future {
	return WhenAnyMatched(nil, actions...)
}

// 返回一个Future
// 如果任何一个Future执行成功并且predicate()函数执返回true, 当前的Future也将会执行成功,并且返回已经成功执行的Future的值.
// 如果所有的Future都被取消, 当前的Future也会被取消; 否则, 当前的Future将会执行失败NoMatchedError, 并且返回所有Future的执行结果.
func WhenAnyMatched(predicate func(interface{}) bool, actions ...interface{}) *Future {
	if predicate == nil {
		predicate = func(v interface{}) bool { return true }
	}

	// todo: action包装成Future
	functions := make([]*Future, len(actions))
	for i, act := range actions {
		functions[i] = Start(act)
	}

	// todo: 构建 Promise 和 返回结果集合
	promise, results := NewPromise(), make([]interface{}, len(functions))
	if len(actions) == 0 {
		promise.Resolve(nil)
	}

	// todo: 设置channel
	chFails, chDones := make(chan anyPromiseResult), make(chan anyPromiseResult)
	go func() {
		for i, function := range functions {
			k := i
			function.OnSuccess(func(v interface{}) {
				defer func() { _ = recover() }()
				chDones <- anyPromiseResult{v, k}
			}).OnFailure(func(v interface{}) {
				defer func() { _ = recover() }()
				chFails <- anyPromiseResult{v, k}
			}).OnCancel(func() {
				defer func() { _ = recover() }()
				chFails <- anyPromiseResult{CANCELLED, k}
			})
		}
	}()

	// todo: 根据预定的规则执行(阻塞)
	if len(functions) == 1 {
		select {
		case fail := <-chFails:
			if _, ok := fail.result.(CancelledError); ok {
				promise.Cancel()
			} else {
				promise.Reject(newNoMatchedError(fail.result))
			}
		case done := <-chDones:
			if predicate(done.result) {
				promise.Resolve(done.result)
			} else {
				promise.Reject(newNoMatchedError(done.result))
			}
		}
	} else {
		go func() {
			defer func() {
				if e := recover(); e != nil {
					promise.Reject(newErrorWithStacks(e))
				}
			}()

			j := 0
			for {
				// todo: 有一个执行结果返回
				select {
				case fail := <-chFails:
					results[fail.i] = getError(fail.result)
				case done := <-chDones:
					if predicate(done.result) {
						// 任何一个Future成功返回, 当前的Future也需要成功返回. 此时需要取消其他的Future的执行
						for _, function := range functions {
							function.Cancel()
						}

						// 关闭channel以避免 `发送方` 被阻塞
						closeChan := func(c chan anyPromiseResult) {
							defer func() { _ = recover() }()
							close(c)
						}
						closeChan(chDones)
						closeChan(chFails)

						// 成功执行并且返回
						promise.Resolve(done.result) // 成功执行并返回
						return
					} else {
						results[done.i] = done.result
					}
				}

				// todo: 执行的次数和functions的长度一致, 需要退出循环
				if j++; j == len(functions) {
					m := 0
					for _, result := range results {
						switch val := result.(type) {
						case CancelledError:
						default:
							m++
							_ = val
						}
					}
					if m > 0 {
						promise.Reject(newNoMatchedError(results)) // 存在取消的Future
					} else {
						promise.Cancel() // 所有的Future都已经被执行(没有取消), 这个时候可以取消当前Promise的执行
					}
					break
				}
			}
		}()
	}

	return promise.Future
}

// 返回一个Future
// 如果所有的Future都成功执行, 当前的Future也会成功执行并且返回相应的结果数组(成功执行的Future的结果);
// 否则, 当前的Future将会执行失败, 并且返回所有Future的执行结果.
func WhenAll(actions ...interface{}) (future *Future) {
	promise := NewPromise()
	future = promise.Future

	if len(actions) == 0 {
		promise.Resolve([]interface{}{})
		return
	}

	// todo: function封装成Future
	functions := make([]*Future, len(actions))
	for i, act := range actions {
		functions[i] = Start(act)
	}

	future = whenAllFuture(functions...)

	return
}

// 返回一个Future
// 如果所有的Future都成功执行, 当前的Future也会成功执行并且返回相应的结果数组(成功执行的Future的结果).
// 如果任何一个Future被取消, 当前的Future也会被取消; 否则, 当前的Future将会执行失败, 并且返回所有Future的执行结果.
func whenAllFuture(futures ...*Future) *Future {
	promise := NewPromise()
	results := make([]interface{}, len(futures))

	if len(futures) == 0 {
		promise.Resolve([]interface{}{})
	} else {
		n := int32(len(futures))
		cancelOthers := func(j int) {
			for k, future := range futures {
				if k != j {
					future.Cancel()
				}
			}
		}

		// todo 逻辑: 全部成功->成功, 任何一下取消->取消, 任何一个失败 -> 失败
		go func() {
			isCancelled := int32(0) // 只能设置一次
			for i, future := range futures {
				j := i
				// 注册函数
				future.OnSuccess(func(v interface{}) {
					results[j] = v
					if atomic.AddInt32(&n, -1) == 0 {
						promise.Resolve(results)
					}
				})

				future.OnFailure(func(v interface{}) {
					// 任何一个失败, 当前Future失败
					if atomic.CompareAndSwapInt32(&isCancelled, 0, 1) {

						cancelOthers(j)

						e := newAggregateError("Error appears in WhenAll:", v)
						promise.Reject(e) // 失败
					}
				})

				future.OnCancel(func() {
					if atomic.CompareAndSwapInt32(&isCancelled, 0, 1) {

						cancelOthers(j)

						promise.Cancel() // 取消
					}
				})
			}
		}()
	}

	return promise.Future
}

var (
	CANCELLED error = &CancelledError{}
)

// future退出时的错误
type CancelledError struct{}

func (e *CancelledError) Error() string {
	return "Task be cancelled"
}

// Future最终的状态
type resultType int

const (
	RESULT_SUCCESS   resultType = iota
	RESULT_FAILURE
	RESULT_CANCELLED
)

// Promise的结果
// Type: 0, Result是Future的返回结果
// Type: 1, Result是Future的返回的错误
// Type: 2, Result是null
type PromiseResult struct {
	Result interface{}
	Type   resultType // success, failure, or cancelled?
}

/*********************************************************************
1. Promise提供一个对象作为结果的代理. 这个结果最初是未知的, 通常是因为其值尚未被计算出.
2. 可以使用Resolve() | Reject() | Cancel() 来设置Promise的最终结果.
3. Future只返回一个带有只读占位符视图.
*********************************************************************/
type Promise struct {
	*Future
}

/*********************************************************************
 方法总体说明:
	1. Cancel() Resolve() Reject(), 这些方法的调用会导致Promise任务执行完毕
	2. OnXxx() 此类型的方法是设置回调函数, 应当在Promise的任务执行完毕前调用添加
*********************************************************************/

// Cancel() 会将 Promise 的结果的 Type 设置为RESULT_CANCELLED。
// 如果promise被取消了, 调用Get()将返回nil和CANCELED错误. 并且所有的回调函数将不会被执行
func (promise *Promise) Cancel() (e error) {
	return promise.setResult(&PromiseResult{CANCELLED, RESULT_CANCELLED})
}

// Resolve() 会将 Promise 的结果的 Type 设置为RESULT_SUCCESS. Result设置为特定值
// 如果promise被取消了, 调用Get()将返回相应的值和nil
func (promise *Promise) Resolve(v interface{}) (e error) {
	return promise.setResult(&PromiseResult{v, RESULT_SUCCESS})
}

// Resolve() 会将 Promise 的结果的 Type 设置为RESULT_FAILURE.
func (promise *Promise) Reject(err error) (e error) {
	return promise.setResult(&PromiseResult{err, RESULT_FAILURE})
}

// OnSuccess注册一个回调函数, 该函数将在Promise有成功返回的时候调用. Promise的值将是Done回调函数的参数.
func (promise *Promise) OnSuccess(callback func(v interface{})) *Promise {
	promise.Future.OnSuccess(callback)
	return promise
}

// OnSuccess注册一个回调函数, 该函数将在Promise有失败返回的时候调用. Promise的error将是Done回调函数的参数.
func (promise *Promise) OnFailure(callback func(v interface{})) *Promise {
	promise.Future.OnFailure(callback)
	return promise
}

// OnComplete注册一个回调函数，该函数将在Promise成功或者失败返回的时候被调用.
// 根据Promise的状态，值或错误将是Always回调函数的参数.
// 如果Promise被调用, 则不会调用回调函数.
func (promise *Promise) OnComplete(callback func(v interface{})) *Promise {
	promise.Future.OnComplete(callback)
	return promise
}

// OnSuccess注册一个回调函数, 该函数将在Promise被取消的时候调用
func (promise *Promise) OnCancel(callback func()) *Promise {
	promise.Future.OnCancel(callback)
	return promise
}

func NewPromise() *Promise {
	value := &futureValue{
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
			value: unsafe.Pointer(value),
		},
	}

	return promise
}
