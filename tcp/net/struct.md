
// netFD 是一个网络描述符, 类似于Linux的文件描述符的概念

```cgo
type netFD struct {
 pfd poll.FD

 // immutable until Close
 family      int
 sotype      int
 isConnected bool // handshake completed or use of association with peer
 net         string
 laddr       Addr
 raddr       Addr
}
```


// FD 包含了两个重要的数据结构 Sysfd 和 pollDesc, 前者是真正的系统文件描述符, 后者是对底层事件驱动的封装.

```cgo
type FD struct {
 // Lock sysfd and serialize access to Read and Write methods.
 fdmu fdMutex

 // System file descriptor. Immutable until Close.
 Sysfd int

 // I/O poller.
 pd pollDesc

 // Writev cache.
 iovecs *[]syscall.Iovec

 // Semaphore signaled when file is closed.
 csema uint32

 // Non-zero if this file has been set to blocking mode.
 isBlocking uint32

 // Whether this is a streaming descriptor, as opposed to a
 // packet-based descriptor like a UDP socket. Immutable.
 IsStream bool

 // Whether a zero byte read indicates EOF. This is false for a
 // message based socket connection.
 ZeroReadIsEOF bool

 // Whether this is a file rather than a network socket.
 isFile bool
}
```

// pollDesc, 底层事件驱动

```cgo
type pollDesc struct {
  runtimeCtx uintptr
}

// runtimeCtx 对应的底层指针是 src/runtime/netpoll.go 当中的 pollDesc

// pollDesc 网络poller描述符
//go:notinheap
type pollDesc struct {
	link *pollDesc // in pollcache, protected by pollcache.lock
	
	// lock 可以保护 pollOpen, polleSetDeadline, pollUnblock 和 durationimpl 操作. 
	// 这些操作完全涵盖了seq, rt 和 wt 变量. 
    // fd 在 PollDesc 的整个生命周期中都是不变的.
    // pollReset, pollWait, pollWaitCanceled 和 runtime·netpollready(IO就绪通知) 不带锁继续进行. 
    // 因此, closing, everr, rg, rd, wg和wd所有操作均以无锁方式进行.
    // NOTE: 以下代码使用 uintptr 来存储 *g (rg/wg), 当GC开始移动对象时, 该数据会爆炸.
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
```