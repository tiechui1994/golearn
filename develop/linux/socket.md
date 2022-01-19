# socket

## scoket options

可以使用 setsockopt(2) 设置这些套接字选项, 并使用 getsockopt(2) 读取所有套接字的套接字选项. 

SOL_SOCKET 级别的选项: 

- SO_ACCEPTCONN

- SO_BINDTODEVICE

- SO_BROADCAST, 开启或禁止进程发送广播消息的能力. 只有数据报套接字支持广播, 并且还必须是在广播消息的网络上. 

- SO_DOMAIN

- SO_KEEPALIVE, 启用 TCP keeplive. 设置发送 keepalive 周期. 如果2小时内在该套接字的任一方向上没有数据交换, TCP
就自动给对端发送一个保持存活探测分节(keep-alive probe). 这是一个对端必须响应的TCP分节, 它会导致以下三种情况之一:

1) 对端以期望的ACK响应. 应用程序得不到通知, 在2个小时后, TCP将发出另一个探测分节.

2) 对端以RST响应, 它告知本端 TCP: 对端已崩溃且已重新启动.

3) 对端对保持存活探测分节没有任何响应.

- SO_LINGER, 该选项指定 close 函数对面向连接的协议(TCP, SCTP)如何操作. 默认操作是 close 立即返回, 但是如果有数据
残留在套接字发送缓冲区, 系统将试着把这些数据发送到对端. SO_LINGER 选项要求在用户进程和内核间传递如下结构:

```cgo
struct linger {
    int l_onoff;
    int l_linger;
}
```

如果 l_onoff 为0, 那么关闭该选项. l_linger 值被忽略, 执行上面的 close 默认操作.

如果 l_onoff 非0, 且 l_linger 为0, 那么当 close 关闭连接时, TCP将终止该连接. 也就是说 TCP 将丢弃保留在套接字缓
冲区中的数据, 并发送一个 RST 给对端, 而没有通常的四次挥手操作. (可以避免TCP的TIME_WAIT状态)

如果 l_onoff 非0, 且 l_linger 非0, 那么当 close 关闭连接时, 内核将拖延一段时间. 也就是说如果在套接字发送缓冲区中有
残留数据, 那么进程将进入休眠, 直到所有数据都已发送完且均被对方确认或延滞时间(l_linger)到. 如果套接字被设置为非阻塞, 那
么它将不等待 close 完成, 即使延滞时间非0. 当使用 SO_LINGER 选项时, 应用程序检查 close 的返回值是非常重要的, 因为如果
在数据发送完并被确认前延滞时间到的话, close 将返回 EWOULDBLOCK 错误, 且套接字发送缓冲区中的残留数据将被丢弃.


- SO_MARK

- SO_OOBINLINE, 如果启用此选项, 则 out-of-band data(带外数据) 将直接放入接收数据流中. 否则, 只有在接收期间设置了 
MSG_OOB 标志时才会接收 out-of-band data. TCP 当中的带外数据指的的具有较高优先级的紧急模式数据(带有 URG 标记).

> 默认情况下, out-of-band data 在接收时是放在一个独立的单字节缓冲区. 接收进程从这个单字节缓冲区读入数据的唯一方法是指
定带有 MSG_OOB 标志调用 recv, recvfrom 或 recvmsg. 如果新的 OOB 字节在旧的 OOB 字节被读取之前到达, 旧的 OOB 字节
就会丢弃. 接收进程 "接收到SIGURG信号" 或 "select异常" 下, 意味对端发送了一个带外字节.
> 如果开启了 SO_OOBINLINE 套接字选项, 那么 out-of-band data 在接收时通常放在套接字缓冲区. 这种情况下, 接收进程不能
指定 MSG_OOB 标志读入该字节. 相反, 接收进程通过检查该连接的 out-of-band mark 获取何时访问到这个数据字节. 
> sockatmark(sockfd) 函数获取是否处于带外标记.


- SO_PASSCRED

- SO_PEERCRED

- SO_PRIORITY

- SO_PROTOCOL

- SO_RCVBUF, 接收缓存区最大字节数. 当使用 setsockopt(2) 设置该值时, 内核将该值加倍, 并且该加倍值由 getsockopt(2)
返回. 默认值由 `/proc/sys/net/core/rmem_default` 文件设置, 最大允许值由 `/proc/sys/net/core/rmem_max` 文件
设置. 此选项的最小(加倍)值为256.

- SO_RCVBUFFORCE, 特权进程可以执行与 SO_RCVBUF 相同的效果, 并且可以覆盖 rmem_max 限制.

- SO_SNDBUF, 发送缓冲区最大字节数. 当使用 setsockopt(2) 设置该值时, 内核将该值加倍, 并且该加倍值由 getsockopt(2)
返回. 默认值由 `/proc/sys/net/core/wmem_default` 文件设置, 最大允许值由 `/proc/sys/net/core/wmem_max` 文件
设置. 此选项的最小(加倍)值为 2048.

- SO_SNDBUFFORCE, 特权进程可以执行与 SO_SNDBUF 相同的效果, 并且可以覆盖 wmem_max 限制.



- SO_RCVLOWAT, SO_SNDLOWAT

- SO_RCVTIMEO, SO_SNDTIMEO



- SO_REUSEADDR, SO_REUSEPORT

SO_REUSEADDR 选项具有的功能:

1) SO_REUSEADDR 允许启动一个监听服务器并绑定端口, 即使以前建立的该端口用作它们的本地端口的连接存在.

2) SO_REUSEADDR 允许在同一个端口上启动同一个服务器的多个实例, 只要每个实例绑定一个不同的本地IP地址即可.

3) SO_REUSEADDR 允许单个进程绑定同一个端口到多个套接字上, 只要每次绑定指定不同的本地IP地址即可.

4) SO_REUSEADDR 允许完全重复的绑定: 当一个IP地址和端口已绑定到某个套接字上时, 如果传输协议支持, 同样的 IP地址和端口
可以绑定到另一个套接字上. 一般来说这个特性仅支持 UDP 套接字. 该特效用于多播时, 允许在同一个主机上同时运行同一个应用程序的
多个副本.

> 当一个 UDP 数据报需要重复绑定套接字中的一个接收时, 所用的规则: 如果该数据报的目的地址是一个广播或多播地址, 那就给每个
匹配的套接字传递一个该数据报的副本; 但是如果该数据报的目的地址是一个单播地址, 那么它值传递给单个套接字.

SO_REUSEPORT 选项:

1) 允许完全重复的绑定, 不过只有在想要绑定同一IP地址和端口的每个套接字都指定了 SO_REUSEPORT 才可以.

2) 如果被绑定的IP地址是一个多播地址, 那么 SO_REUSEPORT 和 SO_REUSEADDR 被认为是等效的.

- SO_TIMESTAMP, 启用或禁用 SO_TIMESTAMP 控制消息的接收. 时间戳控制消息以 SOL_SOCKET 级别发送, cmsg_data 字段
是一个 struct timeval, 指示在此调用中传递给用户的最后一个数据包的接收时间. 





























