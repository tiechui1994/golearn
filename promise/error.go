package promise

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

var (
	CANCELLED error = &CancelledError{}
)

// Cancel
type CancelledError struct{}

func (e *CancelledError) Error() string {
	return "Task be cancelled"
}

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
		writeStrings(buf, []string{name, " ", file, ":", strconv.Itoa(line), "\n"})
	}
	return errors.New(buf.String())
}

func writeStrings(buf *bytes.Buffer, strings []string) {
	for _, s := range strings {
		buf.WriteString(s)
	}
}

// error handling struct and functions
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
