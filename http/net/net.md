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
		return nil, &OpError{Op: "listen", Net: sl.network, Source: nil, Addr: la, Err: &AddrError{Err: "unexpected address type", Addr: address}}
	}
	if err != nil {
		return nil, &OpError{Op: "listen", Net: sl.network, Source: nil, Addr: la, Err: err} // l is non-nil interface containing nil pointer
	}
	return l, nil
}
```


- resolveAddrList, 解析 IP 列表, 返回结果至少包含一个可用的 IP 地址列表

```cgo
func (r *Resolver) resolveAddrList(ctx context.Context, op, network, addr string, hint Addr) (addrList, error) {
    // 根据 network 获取 afnet(协议), proto(协议编号[私有协议编号和公有协议编号], 公有协议编号: ip是0, ipv6是41, udp是17, tcp是6)
	afnet, _, err := parseNetwork(ctx, network, true)
	if err != nil {
		return nil, err
	}
	if op == "dial" && addr == "" {
		return nil, errMissingAddress
	}
	
	// 解析 "类unix域协议" 的地址列表.
	// Unix域协议使用套接字API, 支持同一台主机的不同进程之间进行通信. (本地)
	switch afnet {
	case "unix", "unixgram", "unixpacket":
		addr, err := ResolveUnixAddr(afnet, addr) // 解析协议的地址
		if err != nil {
			return nil, err
		}
		if op == "dial" && hint != nil && addr.Network() != hint.Network() {
			return nil, &AddrError{Err: "mismatched local address type", Addr: hint.String()}
		}
		return addrList{addr}, nil // unix的地址
	}
	
	// 解析 "网络协议", 两大类:
	// 一种是 tcp, udp, 传输层协议. 地址格式: "IP:PORT"
	// 另一种是 ip, 网络层协议. 地址格式: "IP"
	// 涉及到的逻辑是: 端口号检查, DNS解析主机名
	addrs, err := r.internetAddrList(ctx, afnet, addr)
	if err != nil || op != "dial" || hint == nil {
		return addrs, err // 针对 op = "listen" 到此处就结束了. 
	}
	
	// 针对 op = "dial" 的解析, 通配符解析 
	var (
		tcp      *TCPAddr
		udp      *UDPAddr
		ip       *IPAddr
		wildcard bool
	)
	switch hint := hint.(type) {
	case *TCPAddr:
		tcp = hint
		wildcard = tcp.isWildcard()
	case *UDPAddr:
		udp = hint
		wildcard = udp.isWildcard()
	case *IPAddr:
		ip = hint
		wildcard = ip.isWildcard()
	}
	naddrs := addrs[:0]
	for _, addr := range addrs {
		if addr.Network() != hint.Network() {
			return nil, &AddrError{Err: "mismatched local address type", Addr: hint.String()}
		}
		switch addr := addr.(type) {
		case *TCPAddr:
			if !wildcard && !addr.isWildcard() && !addr.IP.matchAddrFamily(tcp.IP) {
				continue
			}
			naddrs = append(naddrs, addr)
		case *UDPAddr:
			if !wildcard && !addr.isWildcard() && !addr.IP.matchAddrFamily(udp.IP) {
				continue
			}
			naddrs = append(naddrs, addr)
		case *IPAddr:
			if !wildcard && !addr.isWildcard() && !addr.IP.matchAddrFamily(ip.IP) {
				continue
			}
			naddrs = append(naddrs, addr)
		}
	}
	if len(naddrs) == 0 {
		return nil, &AddrError{Err: errNoSuitableAddress.Error(), Addr: hint.String()}
	}
	return naddrs, nil
}

```


- listenTCP, 监听 TCP 协议的端口

```cgo
func (sl *sysListener) listenTCP(ctx context.Context, laddr *TCPAddr) (*TCPListener, error) {
	// 创建 socket
	fd, err := internetSocket(ctx, sl.network, laddr, nil, syscall.SOCK_STREAM, 0, "listen", sl.ListenConfig.Control)
	if err != nil {
		return nil, err
	}
	return &TCPListener{fd: fd, lc: sl.ListenConfig}, nil
}


func internetSocket(ctx context.Context, net string, laddr, raddr sockaddr, sotype, proto int, mode string, ctrlFn func(string, string, syscall.RawConn) error) (fd *netFD, err error) {
	// favoriteAddrFamily 返回给定net, laddr, raddr 和 mode的适配的地址族.
	// listenTCP 的 mode 是 "listen"
	//
    // 如果mode表示 "listen", 而laddr是通配符, 则假设用户希望使用通配符地址族(AF_INET和AF_INET6)建立被动开放连接, 通配符地址如下:
    //
    // - 使用通配符地址监听通配符通信域 "tcp"或 "udp": 如果平台同时支持IPv6 和 "IPv4映射的IPv6" 通信功能, 或者不支持IPv4, 则我们使用
    // 双栈(dual stack), AF_INET6和IPV6_V6ONLY=0, 通配符地址侦听. 双栈通配符地址侦听可能会退回到 "仅IPv6", AF_INET6 和 IPV6_V6ONLY=1,
    // 通配符地址侦听. 否则, 我们更喜欢仅使用IPv4的AF_INET通配符地址侦听.
    //
    // - 使用IPv4通配符地址侦听通配符通信域 "tcp" 或 "udp": 与上述相同.
    //
    // - 使用IPv6通配符地址侦听通配符通信域 "tcp" 或 "udp": 与上面相同.
    //
    //- 使用IPv4通配符地址监听IPv4通信域 "tcp4" 或 "udp4": 我们使用仅IPv4的AF_INET通配符地址监听.
    //
    //- 使用IPv6通配符地址监听IPv6通信域 "tcp6" 或 "udp6": 我们使用仅IPv6的AF_INET6和IPV6_V6ONLY=1 通配符地址监听.
    //
    // 其他状况: 如果地址为IPv4, 则返回AF_INET, 否则返回AF_INET6. 它还返回一个布尔值, 该布尔值指定IPV6_V6ONLY选项.
    //
	family, ipv6only := favoriteAddrFamily(net, laddr, raddr, mode)
	return socket(ctx, net, family, sotype, proto, ipv6only, laddr, raddr, ctrlFn)
}
```


底层 socket 的创建:

```cgo
// socket returns a network file descriptor that is ready for
// asynchronous I/O using the network poller.
func socket(ctx context.Context, net string, family, sotype, proto int, ipv6only bool, laddr, raddr sockaddr, ctrlFn func(string, string, syscall.RawConn) error) (fd *netFD, err error) {
	s, err := sysSocket(family, sotype, proto)
	if err != nil {
		return nil, err
	}
	if err = setDefaultSockopts(s, family, sotype, ipv6only); err != nil {
		poll.CloseFunc(s)
		return nil, err
	}
	if fd, err = newFD(s, family, sotype, net); err != nil {
		poll.CloseFunc(s)
		return nil, err
	}

	// This function makes a network file descriptor for the
	// following applications:
	//
	// - An endpoint holder that opens a passive stream
	//   connection, known as a stream listener
	//
	// - An endpoint holder that opens a destination-unspecific
	//   datagram connection, known as a datagram listener
	//
	// - An endpoint holder that opens an active stream or a
	//   destination-specific datagram connection, known as a
	//   dialer
	//
	// - An endpoint holder that opens the other connection, such
	//   as talking to the protocol stack inside the kernel
	//
	// For stream and datagram listeners, they will only require
	// named sockets, so we can assume that it's just a request
	// from stream or datagram listeners when laddr is not nil but
	// raddr is nil. Otherwise we assume it's just for dialers or
	// the other connection holders.

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
	if err := fd.dial(ctx, laddr, raddr, ctrlFn); err != nil {
		fd.Close()
		return nil, err
	}
	return fd, nil
}
```