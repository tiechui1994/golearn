# net 包当中的异步 IO 实现原理

与其他的语言的网络 IO 强调异步非阻塞不同, Go 里的网络 IO 模型是: 创建多个 goroutine, 每个 goroutine 的网络 IO 都
是阻塞的.

但底层, 所有的网络 IO 实际上都是 `非阻塞的`, 以 net.Dial 为例子, 其他的 Read/Write 机制类似.

```
 DialTCP()           DialTCP()           DialTCP()
    |                   |                   |
  socket()           socket()            socket()
    |                   |                   |
  newFD()             newFD()             newFD()
    |                   |                   |
SetNonblock()       SetNonlock()        SetNonblock()
    |                   |                   |
    +-------------------+-------------------+
                        |
                        | once(在一个进程内, 只运行一次)
                        |
                  newPollServer()
                        |
                go pollServer.Run()
                        |
                   for {
                     epoll_wait()
                     wake FD via chan ------->----------+
                   }                                    |
                        |                               |
    +-------------------+--------------------+          |
    |                   |                    |          |
 connect()           connect()            connect()     | 把底层非阻塞变成上层阻塞的同步机制
    |                                                   |
    |                                                   |
if EINPROGRES {                                         |
    EPOLL_CTL_ADD to register to pollServer             |
    get response from chan ------------------->---------+
}
```

Read 的原理:

```
for {
    fd.read()
    if err == EAGAIN {
        pollserver.WaitRead(fd)
        continue
    }
    break
}
```

网络关键 API 的实现, 主要包括 Listen, Accept, Read, Write 等.

## Listen

`Listen()` 方法的调用链路.

`Listen()` -> `ListenConfig.listenTCP()` -> `internetSocket()` -> `socket()` -> `FD.listenStream()`

```cgo
// network 的值是 "tcp", "tcp4", "tcp6", "unix" or "unixpacket" 之一.
// 
// 针对 network 是 "tcp":
// 如果 address 参数中的主机 "为空" 或 "未指定IP地址", 则 Listen 侦听本地系统的所有可用 `单播` 和 `任播` IP地址.
// 如果 address 使用主机名(不建议使用), 那么它会为该主机的最多一个IP地址创建侦听器.
// 如果 address 参数中的端口 "为空" 或 "0", 例如, "127.0.0.1:" 或 "[::1]:0", 则会自动选择端口号.
//
// 如果要仅使用 IPv4, 则 network 参数使用 "tcp4".
// Listener 的 Addr() 方法可用于发现所选端口.
func Listen(network, address string) (Listener, error) {
    var lc ListenConfig
    return lc.Listen(context.Background(), network, address)
}


func (lc *ListenConfig) Listen(ctx context.Context, network, address string) (Listener, error) {
  // 根据 network 和 address 解析统一的 IP 列表
    addrs, err := DefaultResolver.resolveAddrList(ctx, "listen", network, address, nil)
    if err != nil {
        return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: nil, Err: err}
    }
    sl := &sysListener{
        ListenConfig: *lc,
        network:      network,
        address:      address,
    }
    var l Listener
    la := addrs.first(isIPv4) // 返回第一个 ipv4 地址, 如果没找到, 则默认是第一个地址
	
    // 通过下面的逻辑可以确定: Listen() 函数只能监听 TCP 和 Unix 协议端口 
    switch la := la.(type) {
    case *TCPAddr:
        l, err = sl.listenTCP(ctx, la)
    case *UnixAddr:
        l, err = sl.listenUnix(ctx, la)
    default:
        return nil, &OpError{
            Op: "listen", Net: sl.network, 
            Source: nil, Addr: la, 
            Err: &AddrError{
                Err: "unexpected address type", 
                Addr: address,
            },
        }
    }
    if err != nil {
        return nil, &OpError{
            Op: "listen", Net: sl.network, 
            Source: nil, Addr: la, 
            Err: err,
        } // l is non-nil interface containing nil pointer
    }
    return l, nil
}
```

- listenTCP

```cgo
func (sl *sysListener) listenTCP(ctx context.Context, laddr *TCPAddr) (*TCPListener, error) {
    // 创建 socket
    fd, err := internetSocket(ctx, sl.network, laddr, nil, syscall.SOCK_STREAM, 0, "listen", 
        sl.ListenConfig.Control)
    if err != nil {
        return nil, err
    }
    return &TCPListener{fd: fd, lc: sl.ListenConfig}, nil
}
```

- internetSocket

```cgo
func internetSocket(ctx context.Context, net string, laddr, raddr sockaddr, sotype, proto int, 
    mode string, ctrlFn func(string, string, syscall.RawConn) error) (fd *netFD, err error) {
    // favoriteAddrFamily 返回给定 net, laddr, raddr 和 mode 的适配的地址族.
    // listenTCP 的 mode 是 "listen"
    //
    // 如果mode是 "listen", 而laddr是通配符, 则假设用户希望使用通配符地址族(AF_INET和AF_INET6)建立被动开放连接, 
    // 通配符地址如下:
    //
    // - 使用通配符地址监听通配符通信域 "tcp" 或 "udp": 如果平台同时支持 "IPv6" 和 "IPv4-mapped IPv6" 通信功能, 
    // 或者不支持IPv4, 则使用双栈(dual stack), AF_INET6和IPV6_V6ONLY=0, 通配符地址侦听. 双栈通配符地址侦听
    // 可能会退回到 "IPv6-only", AF_INET6 和 IPV6_V6ONLY=1, 通配符地址侦听. 否则, 使用 "IPv4-only" 和 AF_INET
    // 通配符地址侦听.
    //
    // - 使用IPv4通配符地址侦听通配符通信域 "tcp" 或 "udp": 与上述相同.
    //
    // - 使用IPv6通配符地址侦听通配符通信域 "tcp" 或 "udp": 与上面相同.
    //
    // - 使用IPv4通配符地址监听IPv4通信域 "tcp4" 或 "udp4": 使用 "IPv4-only" 的AF_INET通配符地址监听.
    //
    // - 使用IPv6通配符地址监听IPv6通信域 "tcp6" 或 "udp6": 使用 AF_INET6和IPV6_V6ONLY=1 通配符地址监听.
    //
    // 其他状况: 如果地址为IPv4, 则返回AF_INET, 否则返回AF_INET6. 
    // 
    // 函数还返回一个布尔值, 该布尔值指定IPV6_V6ONLY选项.
    //
    family, ipv6only := favoriteAddrFamily(net, laddr, raddr, mode)
	
    // net 是 tcp, tcp4, tcp6
    // family 是 AF_INET 或 AF_INET6
    // sotype 是 syscall.SOCK_STREAM
    // proto 是 0
    return socket(ctx, net, family, sotype, proto, ipv6only, laddr, raddr, ctrlFn)
}
```


- socket

socket() 创建 socket, 并设置 socket 的属性.

```cgo
// socket() 返回一个网络文件描述符, 该描述符已准备好使用网络轮询器进行异步I/O. 
func socket(ctx context.Context, net string, family, sotype, proto int, ipv6only bool, 
    laddr, raddr sockaddr, ctrlFn func(string, string, syscall.RawConn) error) (fd *netFD, err error) {
    // 创建 socket
    s, err := sysSocket(family, sotype, proto)
    if err != nil {
        return nil, err
    }
	
    // 设置 IPV6_V6ONLY, SO_BROADCAST(UDP), 系统调用 setsockopt
    if err = setDefaultSockopts(s, family, sotype, ipv6only); err != nil {
        poll.CloseFunc(s)
        return nil, err
    }
	
    // 创建文件描述符 FD (通用)
    if fd, err = newFD(s, family, sotype, net); err != nil {
        poll.CloseFunc(s)
        return nil, err
    }
	
    // stream, datagram(TCP和UDP的listen)
    // listen的操作: 设置默认属性, 执行 ctrlFn 回调, 系统调用 bind, 系统调用listen(只有TCP才有), netFD初始化
    if laddr != nil && raddr == nil {
        switch sotype {
        case syscall.SOCK_STREAM, syscall.SOCK_SEQPACKET:
            if err := fd.listenStream(laddr, listenerBacklog(), ctrlFn); err != nil {
                fd.Close()
                return nil, err
            }
            return fd, nil
        case syscall.SOCK_DGRAM:
            if err := fd.listenDatagram(laddr, ctrlFn); err != nil {
                fd.Close()
                return nil, err
            }
            return fd, nil
        }
    }
	
    // dialer
    // dial的操作: 执行 ctrlFn 回调, 系统调用 bind(可选操作, 仅当laddr 存在), 系统调用 connect
    if err := fd.dial(ctx, laddr, raddr, ctrlFn); err != nil {
        fd.Close()
        return nil, err
    }
	
    // other
    return fd, nil
}
```


listen 时, socket 设置的 options 和 sotype

options:
```
# 通用. sysSocket
O_NONBLOCK # fcntl 设置

# 通用. socket
IPPROTO_IPV6 - IPV6_V6ONLY  # family是AF_INET6
SOL_SOCKET - SO_BROADCAST   # UDP, RAW

# TCP
SOL_SOCKET - SO_REUSEADDR 

# UDP, 多播地址
SOL_SOCKET - SO_REUSEADDR
SOL_SOCKET - SO_REUSEPORT

# 自定义设置
ListenConfig.Control # 该函数在 socket() 之后, bind 之前调用

# 系统提供的其他 TCP 选项
SOL_SOCKET - SO_RCVBUF     # SetReadBuffer
SOL_SOCKET - SO_SNDBUF     # SetWriteBuffer
SOL_SOCKET - SO_LINGER     # SetLinger 
SOL_SOCKET - SO_KEEPALIVE  # SetKeepAlive
IPPROTO_TCP -  TCP_NODELAY # SetNoDelay

# 系统提供的其他 UDP 选项
SOL_SOCKET - SO_RCVBUF # SetReadBuffer
SOL_SOCKET - SO_SNDBUF # SetWriteBuffer
```

sotype:
```
# 通用, sysSocket
SOCK_NONBLOCK | SOCK_CLOEXEC

# TCP
SOCK_STREAM

# UDP
SOCK_DGRAM
```

设置的 socket 选项的 level 包含:

```
protocol:
IPPROTO_IP (IP层面选项)

IPPROTO_TCP (TCP层面选项)
IPPROTO_UDP (UDP层面选项)

socket:
SOL_SOCKET (SOCKET层面选项)
```


`sysSocket()` 进行系统调用创建 socket, socket属性为: SOCK_NONBLOCK | SOCK_CLOEXEC

> SOCK_NONBLOCK 本质是文件fd属性 O_NONBLOCK
> SOCK_CLOEXEC 本质上是文件fd属性 FD_CLOEXEC

```cgo
func sysSocket(family, sotype, proto int) (int, error) {
    // 创建 socket. 系统调用 socket
    s, err := socketFunc(family, sotype|syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC, proto)
    // On Linux the SOCK_NONBLOCK and SOCK_CLOEXEC flags were
    // introduced in 2.6.27 kernel and on FreeBSD both flags were
    // introduced in 10 kernel. 
    // If we get an EINVAL error on Linux or EPROTONOSUPPORT error on FreeBSD, 
    // fall back to using socket without them.
    switch err {
    case nil:
        return s, nil
    default:
        return -1, os.NewSyscallError("socket", err)
    case syscall.EPROTONOSUPPORT, syscall.EINVAL:
    }

    // 不支持 SOCK_NONBLOCK, SOCK_CLOEXEC 选项的的 socket.
    syscall.ForkLock.RLock()
    s, err = socketFunc(family, sotype, proto)
    if err == nil {
        syscall.CloseOnExec(s) //  socket 属性设置为 FD_CLOEXEC
    }
    syscall.ForkLock.RUnlock()
    if err != nil {
        return -1, os.NewSyscallError("socket", err)
    }
	
    // 设置 socket 为 O_NONBLOCK, 系统调用 fcntl
    if err = syscall.SetNonblock(s, true); err != nil {
        poll.CloseFunc(s)
        return -1, os.NewSyscallError("setnonblock", err)
    }
    return s, nil
}
```


- listenStream

> 系统调用 bind() 和 listen() 函数

```cgo
func (fd *netFD) listenStream(laddr sockaddr, backlog int, ctrlFn func(string, string, syscall.RawConn) error) error {
    // 设置 SO_REUSEADDR.
    var err error
    if err = setDefaultListenerSockopts(fd.pfd.Sysfd); err != nil {
        return err
    }
	
    // 获取 laddr(本地) 的地址
    var lsa syscall.Sockaddr
    if lsa, err = laddr.sockaddr(fd.family); err != nil {
        return err
    }
	
    // 回调函数
    if ctrlFn != nil {
        c, err := newRawConn(fd)
        if err != nil {
            return err
        }
        if err := ctrlFn(fd.ctrlNetwork(), laddr.String(), c); err != nil {
            return err
        }
    }
	
    // 系统调用 bind()
    if err = syscall.Bind(fd.pfd.Sysfd, lsa); err != nil {
        return os.NewSyscallError("bind", err)
    }
	
    // 系统调用 listen() 
    if err = listenFunc(fd.pfd.Sysfd, backlog); err != nil {
        return os.NewSyscallError("listen", err)
    }
	
    // 关键: 初始化socket与异步IO相关的内容
    if err = fd.init(); err != nil {
        return err
    }
    lsa, _ = syscall.Getsockname(fd.pfd.Sysfd)
    fd.setAddr(fd.addrFunc()(lsa), nil)
    return nil
}
```

- netFD.init

netFD.init() => FD.Init(net string, pollable bool) => pollDesc.init(*FD) 

// netFD, net
// FD, internal/poll
// pollDesc, internal/poll

```cgo
func (fd *netFD) init() error {
    return fd.pfd.Init(fd.net, true)
}
            ||
            \/
    
func (fd *FD) Init(net string, pollable bool) error {
    // We don't actually care about the various network types.
    if net == "file" {
        fd.isFile = true
    }
    if !pollable {
        fd.isBlocking = 1
        return nil
    }
    err := fd.pd.init(fd)
    if err != nil {
        // If we could not initialize the runtime poller,
        // assume we are using blocking mode.
        fd.isBlocking = 1
    }
    return err
}
            ||
            \/

func (pd *pollDesc) init(fd *FD) error {
    // runtime_pollServerInit => runtime.poll_runtime_pollServerInit
    serverInit.Do(runtime_pollServerInit)
	
    // runtime_pollOpen => runtime.poll_runtime_pollOpen
    ctx, errno := runtime_pollOpen(uintptr(fd.Sysfd))
    if errno != 0 {
        if ctx != 0 {
            // runtime_pollUnblock => runtime.netpollUnblock
            runtime_pollUnblock(ctx)
            // runtime_pollClose => runtime.netpollClose
            runtime_pollClose(ctx)
        }
        return errnoErr(syscall.Errno(errno))
    }
    pd.runtimeCtx = ctx
    return nil
}
```

关于 runtime_pollServerInit 调用链:

poll.runtime_pollServerInit => runtime.poll_runtime_pollServerInit => runtime.netpollGenericInit =>
netpollinit => epollcreate(先尝试系统调用 epoll_create1, 若失败, 尝试 epoll_create)

> 1.这里使用了 `go:linkname` 魔法, 将 `poll.runtime_pollServerInit` 和 `runtime.runtime_pollServerInit`
进行链接绑定. `//go:linkname poll_runtime_pollServerInit internal/poll.runtime_pollServerInit`.
> 2.在 linux 当中 netpollinit 是在 `runtime/netpoll_epoll.go` 文件.
> 3.epoll_create1 是使用 flags 为参数, epoll_create 是使用 size 为参数.

## Accept

TCPListener.Accept() => TCPListener.accept() => netFD.accept() => FD.Accept()

- Accept

```cgo
// Accept在侦听器接口中实现Accept方法; 它等待下一个调用并返回通用Conn.
func (l *TCPListener) Accept() (Conn, error) {
    if !l.ok() {
        return nil, syscall.EINVAL
    }
    c, err := l.accept()
    if err != nil {
        return nil, &OpError{Op: "accept", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
    }
    return c, nil
}
```

- TCPListener.accept

TCPListener.accept(), 获取接收请求socket, 并将socket包装成一个 TCPConn

```cgo
func (ln *TCPListener) accept() (*TCPConn, error) {
    fd, err := ln.fd.accept()
    if err != nil {
        return nil, err
    }
    tc := newTCPConn(fd)
    if ln.lc.KeepAlive >= 0 {
        setKeepAlive(fd, true)
        ka := ln.lc.KeepAlive
        if ln.lc.KeepAlive == 0 {
            ka = defaultTCPKeepAlive
        }
        setKeepAlivePeriod(fd, ka)
    }
    return tc, nil
}
```

- netFD.accept

netFD.accept(), 调用底层 socket 的 accpet() 方法, 获取 `客户端` 的 socket.

```cgo
func (fd *netFD) accept() (netfd *netFD, err error) {
    d, rsa, errcall, err := fd.pfd.Accept()
    if err != nil {
        if errcall != "" {
            err = wrapSyscallError(errcall, err)
        }
        return nil, err
    }

    if netfd, err = newFD(d, fd.family, fd.sotype, fd.net); err != nil {
        poll.CloseFunc(d)
        return nil, err
    }
    if err = netfd.init(); err != nil {
        fd.Close()
        return nil, err
    }
    lsa, _ := syscall.Getsockname(netfd.pfd.Sysfd)
    netfd.setAddr(netfd.addrFunc()(lsa), netfd.addrFunc()(rsa))
    return netfd, nil
}
```

- FD.Accept()

```cgo
// Accept wraps the accept network call.
func (fd *FD) Accept() (int, syscall.Sockaddr, string, error) {
    if err := fd.readLock(); err != nil {
        return -1, nil, "", err
    }
    defer fd.readUnlock()

    // runtime_pollReset => runtime.poll_runtime_pollReset
    // mode: 'r'
    if err := fd.pd.prepareRead(fd.isFile); err != nil {
        return -1, nil, "", err
    }
    for {
        // accept 系统调用, syscall.Syscall6
        s, rsa, errcall, err := accept(fd.Sysfd)
        if err == nil {
            return s, rsa, "", err
        }
        switch err {
        case syscall.EAGAIN:
            if fd.pd.pollable() {
                // runtime_pollWait => runtime.poll_runtime_pollWait
                // mode: 'r'
                if err = fd.pd.waitRead(fd.isFile); err == nil {
                    continue
                }
            }
        case syscall.ECONNABORTED:
            // This means that a socket on the listen
            // queue was closed before we Accept()ed it;
            // it's a silly error, so try again.
            continue
        }
        return -1, nil, errcall, err
    }
}
```

> poll_runtime_pollReset, 清除 pollDesc 的 rg 或 wg
> poll_runtime_pollWait, 调用 gopark 休眠.

> 常见的系统调用错误:
> EAGAIN, EWOULDBLOCK: resource temporarily unavailable.
> EINTR: interrupted system call.
> EINVAL: invalid argument.
> ECONNABORTED: software caused connection abort. (会出现在accept系统调用)
> ENOSYS: function not implemented.

## Read

conn.Read() => netFD.Read() => FD.Read() => syscall.Read()

其中 FD.Read() 与 FD.Accept() 在实现上是类似的.

```cgo
func (fd *FD) Read(p []byte) (int, error) {
    if err := fd.readLock(); err != nil {
        return 0, err
    }
    defer fd.readUnlock()
    if len(p) == 0 {
        // If the caller wanted a zero byte read, return immediately
        // without trying (but after acquiring the readLock).
        // Otherwise syscall.Read returns 0, nil which looks like
        // io.EOF.
        // TODO(bradfitz): make it wait for readability? (Issue 15735)
        return 0, nil
    }
    if err := fd.pd.prepareRead(fd.isFile); err != nil {
        return 0, err
    }
    if fd.IsStream && len(p) > maxRW {
        p = p[:maxRW]
    }
    for {
        // ignoringEINTR, 执行传入的函数, 如果出现错误 EINTR, 则一直执行, 直至非 EINTR 错误才算完成.
        n, err := ignoringEINTR(func() (int, error) { return syscall.Read(fd.Sysfd, p) })
        if err != nil {
            n = 0
            if err == syscall.EAGAIN && fd.pd.pollable() {
                if err = fd.pd.waitRead(fd.isFile); err == nil {
                    continue
                }
            }
        }
        err = fd.eofError(n, err)
        return n, err
    }
}
```

> Write 与 Read 在实现上是类似的.
> 在 FD 当中, Accept, Read, Write 实现原理是一致的, 它是通过类似于 pollDesc.waitRead 或 pollDesc.waitWrite
来 park 住 goroutine 直到期待的 I/O 事件发生才返回恢复.

