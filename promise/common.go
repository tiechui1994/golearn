package promise

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

// NoMatchedError presents no future that returns matched result in WhenAnyTrue function.
type NoMatchedError struct {
	Results []interface{}
}

func (e *NoMatchedError) Error() string {
	return "No matched future"
}

func (e *NoMatchedError) HasError() bool {
	for _, ie := range e.Results {
		if _, ok1 := ie.(error); ok1 {
			return true
		}
	}
	return false
}

func newNoMatchedError(results ...interface{}) *NoMatchedError {
	return &NoMatchedError{results}
}

// AggregateError aggregate multi errors into an error
type AggregateError struct {
	s         string
	InnerErrs []error
}

func (e *AggregateError) Error() string {
	if e.InnerErrs == nil {
		return e.s
	} else {
		buf := bytes.NewBufferString(e.s)
		buf.WriteString("\n\n")
		for i, ie := range e.InnerErrs {
			if ie == nil {
				continue
			}
			buf.WriteString("error appears in Future ")
			buf.WriteString(strconv.Itoa(i))
			buf.WriteString(": ")
			buf.WriteString(ie.Error())
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
		return buf.String()
	}
}

func newAggregateError(s string, e interface{}) *AggregateError {
	return &AggregateError{newErrorWithStacks(s).Error(), []error{getError(e)}}
}

func newErrorWithStacks(i interface{}) (e error) {
	err := getError(i)
	buf := bytes.NewBufferString(err.Error())
	buf.WriteString("\n")

	pcs := make([]uintptr, 50)
	num := runtime.Callers(2, pcs)
	for _, v := range pcs[0:num] {
		fun := runtime.FuncForPC(v)
		file, line := fun.FileLine(v)
		name := fun.Name()
		//fmt.Println(name, file + ":", line)
		writeStrings(buf, []string{name, " ", file, ":", strconv.Itoa(line), "\n"})
	}
	return errors.New(buf.String())
}

// 对action进行代理包装

func startPipe(r *PromiseResult, pipeTask func(v interface{}) *Future, pipePromise *Promise) {
	//处理链式异步任务
	if pipeTask != nil {
		f := pipeTask(r.Result)
		f.OnSuccess(func(v interface{}) {
			pipePromise.Resolve(v)
		}).OnFailure(func(v interface{}) {
			pipePromise.Reject(getError(v))
		})
	}

}

func getFutureReturnVal(r *PromiseResult) (interface{}, error) {
	if r.Type == RESULT_SUCCESS {
		return r.Result, nil
	} else if r.Type == RESULT_FAILURE {
		return nil, getError(r.Result)
	} else {
		return nil, getError(r.Result) //&CancelledError{}
	}
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
		forSlice(s, func(f func(interface{})) { f(r.Result) })
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

// Error handling struct and functions
type stringer interface {
	String() string
}

func getError(i interface{}) (e error) {
	if i != nil {
		switch v := i.(type) {
		case error:
			e = v
		case string:
			e = errors.New(v)
		default:
			if s, ok := i.(stringer); ok {
				e = errors.New(s.String())
			} else {
				e = errors.New(fmt.Sprintf("%v", i))
			}
		}
	}
	return
}

func writeStrings(buf *bytes.Buffer, strings []string) {
	for _, s := range strings {
		buf.WriteString(s)
	}
}
