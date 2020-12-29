
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


type pollDesc struct {
  runtimeCtx uintptr
}
```