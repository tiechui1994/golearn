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

- listenTCP, 监听 TCP 协议的端口

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
	
	// 设置 socket 的选项, 系统调用 setsockopt
	if err = setDefaultSockopts(s, family, sotype, ipv6only); err != nil {
		poll.CloseFunc(s)
		return nil, err
	}
	
	// 创建文件描述符 FD (通用), 并且绑定 socket 
	if fd, err = newFD(s, family, sotype, net); err != nil {
		poll.CloseFunc(s)
		return nil, err
	}

	// 此函数为以下应用程序创建 "网络文件描述符" :
    //
    //- 打开一个被动 stream connection 的 endpoint 持有者, 称为stream listener
    //
    //- 打开一个目标无关的 datagram connection 的 endpoint 持有者, 称为datagram listener
    //
    //- 打开 active stream 或特定于目标的 datagram connection 的 endpoint 持有者, 称为 dialer.
    //
    //- 打开另一个连接的 endpoint 持有者, 例如: 与内核内部的协议栈通信
    //
    // 对于 stream 和 datagram listener, 它们仅需要named socket, 因此当laddr不是nil, 而raddr为nil时, 
    // 我们可以假设这只是来自 stream 或 datagram listener 的 request. 否则, 我们假定它仅适用于 dialer 或其他连接持有者.
    
    // stream, datagram
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
	if err := fd.dial(ctx, laddr, raddr, ctrlFn); err != nil {
		fd.Close()
		return nil, err
	}
	
	// other
	return fd, nil
}
```


`sysSocket()` 进行系统调用创建 socket, 设置 socket 为非阻塞.

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

	// 不支持 SOCK_NONBLOCK, SOCK_CLOEXEC 选项的的 socket, 年代久远的 Linux 内核
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

`listenStream()` 设置 socket 的属性 SO_REUSEADDR. 

> 系统调用 bind() 和 listen() 函数

```cgo
func (fd *netFD) listenStream(laddr sockaddr, backlog int, ctrlFn func(string, string, syscall.RawConn) error) error {
	// 设置 socket 的 SO_REUSEADDR 属性. setsockopt
	var err error
	if err = setDefaultListenerSockopts(fd.pfd.Sysfd); err != nil {
		return err
	}
	
	// 获取 laddr(本地) 的地址
	var lsa syscall.Sockaddr
	if lsa, err = laddr.sockaddr(fd.family); err != nil {
		return err
	}
	
	// 调用控制函数
	if ctrlFn != nil {
		c, err := newRawConn(fd)
		if err != nil {
			return err
		}
		if err := ctrlFn(fd.ctrlNetwork(), laddr.String(), c); err != nil {
			return err
		}
	}
	
	// 调用底层 bind()
	if err = syscall.Bind(fd.pfd.Sysfd, lsa); err != nil {
		return os.NewSyscallError("bind", err)
	}
	
	// 调用底层 listen() 
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

netFD.init() => FD.Init(net string, pollable bool) => pollDesc.init(*FD) 

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
    // runtime_pollServerInit => runtime.netpollServerInit
	serverInit.Do(runtime_pollServerInit)
	
	// runtime_pollOpen => runtime.netpollOpen
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


## Accept

TCPListener.Accept() => TCPListener.accept() => netFD.accept() => FD.Accept()

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


```cgo
// Accept wraps the accept network call.
func (fd *FD) Accept() (int, syscall.Sockaddr, string, error) {
	if err := fd.readLock(); err != nil {
		return -1, nil, "", err
	}
	defer fd.readUnlock()

	if err := fd.pd.prepareRead(fd.isFile); err != nil {
		return -1, nil, "", err
	}
	for {
		s, rsa, errcall, err := accept(fd.Sysfd)
		if err == nil {
			return s, rsa, "", err
		}
		switch err {
		case syscall.EAGAIN:
			if fd.pd.pollable() {
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