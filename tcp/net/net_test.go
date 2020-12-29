package net

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestListen(t *testing.T) {
	Listen()
}

type Addr interface {
	Addr() string
}

func IsNil(addr Addr) {
	if addr == nil {
		fmt.Println("is nil")
	}
}

func TestNil(t *testing.T) {
	IsNil(nil)
}

func TestPollDesc(t *testing.T) {
	type timer struct {
		tb     uintptr // the bucket the timer lives in
		i      int     // heap index
		when   int64
		period int64
		f      func(interface{}, uintptr)
		arg    interface{}
		seq    uintptr
	}

	type mutex struct {
		key uintptr
	}

	type pollDesc struct {
		link *pollDesc // in pollcache, protected by pollcache.lock

		lock    mutex // protects the following fields
		fd      uintptr
		closing bool
		everr   bool    // marks event scanning error happened
		user    uint32  // user settable cookie
		rseq    uintptr // protects from stale read timers
		rg      uintptr // pdReady, pdWait, G waiting for read or nil
		rt      timer   // read deadline timer (set if rt.f != nil)
		rd      int64   // read deadline
		wseq    uintptr // protects from stale write timers
		wg      uintptr // pdReady, pdWait, G waiting for write or nil
		wt      timer   // write deadline timer
		wd      int64   // write deadline
	}

	t.Log("pollDesc Size", unsafe.Sizeof(pollDesc{}))
	t.Log(4096 / unsafe.Sizeof(pollDesc{}))

	t.Log("timer Size", unsafe.Sizeof(timer{}))
	t.Log("timer Size", unsafe.Sizeof(timer{}.arg))
}
