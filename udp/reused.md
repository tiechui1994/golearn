# Go UDP 端口重用

一般来说, TCP/UDP 的端口只能绑定在一个套接字上. 当我们尝试监听一个已经被其他进程监听的端口时, bind 调用就会失败, 
errno 置为 98 EADDRINUSE. 也就是所谓的端口占用.

但是一个端口只能被一个进程监听吗? 显然不是的. 比如说我们可以先 bind 一个套接字再 fork, 这样两个子进程就监听了同一个
端口. Nginx 就是这样做的, 它的所有 worker 进程都监听着同一个端口. 我们还可以使用 UNIX domain socket 传递文件, 
将一个 fd "发送" 到另一个进程中, 实现同样的效果.

根据 TCP/IP 标准, 端口本身是允许服用的, 绑定端口的本质就是当系统收到一个 TCP 报文段或 UDP 报文段时, 可以根据其头部
的端口字段找到对应的进程, 并将数据传递给对应的进程. 另外对于广播和组播, 端口复用是必须的, 因为它们本身就是多重交付的.

Linux 当中可以通过设置 socket 的 SO_REUSEADDR 和 SO_REUSEPORT 来启用地址复用和端口复用.

注: SO_REUSEADDR 是在 Linux kernel 2.4 版本以后开始支持. SO_REUSEPORT 是 Linux kernel 3.9 版本以后开始支持.
使用时需要注意下.

## Go 当中端口重用实现

```
func Control(network, address string, c syscall.RawConn) error {
    var err error
    c.Control(func(fd uintptr) {
        err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
        if err != nil {
            return
        }
    
        err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
        if err != nil {
            return
        }
    })
    return err
}

var listenConfig = net.ListenConfig{
    Control: Control,
}

// 可以 Listen UDP, TCP. 
// Control 是在创建完系统 socket(调用 sysSocket) 后, 调用 netFD.listenDatagram 当中进行回调执行的.
// 整个过程发生在 socket() 函数当中
func ListenPacket(network, address string) (net.PacketConn, error) {
    return listenConfig.ListenPacket(context.Background(), network, address)
}

// 可以 Dail UDP, TCP
// Control 的调用时机与 ListenPacket 是一致的, 但是其调用链更长.
// 1. Dialer.DialContext
// 2. 构建 sysDialer, 调用 sysDialer.dialSerial
// 3. 调用 sysDialer.dialSingle
// 4. 调用 sysDialer.dialUDP
// 5. internetSocket() -> socket()
func Dial(network, laddr, raddr string) (net.Conn, error) {
    nla, err := ResolveAddr(network, laddr)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve local addr: %w", err)
    }
    d := net.Dialer{
        Control:   Control,
        LocalAddr: nla,
    }
    return d.Dial(network, raddr)
}

func ResolveAddr(network, address string) (net.Addr, error) {
    switch network {
    default:
        return nil, net.UnknownNetworkError(network)
    case "ip", "ip4", "ip6":
        return net.ResolveIPAddr(network, address)
    case "tcp", "tcp4", "tcp6":
        return net.ResolveTCPAddr(network, address)
    case "udp", "udp4", "udp6":
        return net.ResolveUDPAddr(network, address)
    case "unix", "unixgram", "unixpacket":
        return net.ResolveUnixAddr(network, address)
    }
}
```

