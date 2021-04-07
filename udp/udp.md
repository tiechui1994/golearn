## UDP

### `ListenUDP` vs `DialUDP`

客户端使用 `DailUDP` 建立数据报的源对象和目标对象(IP+Port), 它会创建 UDP Socket文件描述符, 然后调用内部的 `connect`
为这个文件描述符设置源地址和目标地址, 这时Go将它称之为 `connected`. 这个方法返回 `*UDPConn`.

服务器使用 `ListenUDP` 返回的 `*UDPConn` 直接往某个目标地址发送数据报, 而不是通过 `DailUDP` 方式发送, 原因在于两者
返回的 `*UDPConn` 是不同的. 前者是 `connected`, 后者是 `unconnected`.


必须要清楚知道UDP是连接的(connected)还是未连接的(unconnected)的, 这样才能正确的选择读写方法.

如果 `*UDPConn` 是 `connected`, 读写方法是 `Read` 和 `Write`.

如果 `*UDPConn` 是 `unconnected`, 读写方法是 `ReadFromUDP` 和 `WriteToUDP` (以及 `ReadFrom` 和 `WriteTo`)


下面是 Linux 关于 UDP 的文档:

> When a UDP socket is created, its local and remote addresses are unspecified. Datagrams can be 
sent immediately using `sendto` or `sendmsg` with a valid destination address as an argument. When 
`connect` is called on the socket, the default destination address is set and datagrams can now be 
sent using `send` or `write` without specifying a destination address. It is still possible to send 
to other destinations by passing an address to `sendto` or `sendmsg`. In order to receive packets, 
the socket can be bound to a local address first by using `bind`. Otherwise, the socket layer will 
automatically assign a free local port out of the range defined by 
`/proc/sys/net/ipv4/ip_local_port_range` and bind the socket to INADDR_ANY.


`ReadFrom` 和 `WriteTo` 是为了实现 `PacketConn` 接口而实现的方法, 它们基本上和 `ReadFromUDP` 和 `WriteToDUP`
一样, 只不过地址换成了更为通用的 `Addr`, 而不是具体化的 `UDPAddr`.


几种特殊情况:

1. 因为 `unconnected` 的 `*UDPConn` 还没有目标地址, 所以需要把目标地址当作参数传入到 `WriteToUDP` 的方法中, 但是
`unconnected` 的 `*UDPConn` 可以调用 `Read` 方法吗?

**可以**, 但是在这种状况下, 客户端的地址信息被忽略了.(也就是说客户端的地址信息提前确定, 才能往客户端写入)

2. `unconnected` 的 `*UDPConn` 可以调用 `Write` 方法吗?

**不可以**, 因为不知道目标地址.

3. **`connected` 的 `*UDPConn` 可以调用 `WriteToUDP` 方法吗?**

**不可以**, 因为目标地址已经设置. 即使是相同的目标地址也不可以.

4. `connected` 的 `*UDPConn` 如果调用 `Closed` 以后可以调用 `WriteToUDP` 方法吗?

**不可以**, 调用 `Closed` 之后, 

5. `connected` 的 `*UDPConn` 可以调用 `ReadFromUDP` 方法吗?

**可以**, 但是它的功能基本和 `Read` 一样, 只能和它 `connected` 的对端通信.

6. `*UDPConn` 还有一个通用的 `WriteMsgUDP(b, oob []byte, addr *UDPAddr)`, 同时支持 `connected` 和 
`unconnected` 的 UDPConn:

a) 如果 `UDPConn` 还未连接, 那么会发送数据到addr

b) 如果 `UDPConn` 已连接, 那么它会发送数据给连接的对端, 这种状况下会忽略 addr

### 广播


### socket 编程的相关选项

- `SOCK_NONBLOCK`(非阻塞IO), `SOCK_CLOEXEC`(fork子进程之后, 关闭父进程的文件描述符), 这两个是fd的属性

> 子进程在 fork 出来的时候, 使用了写时复制(COW, Copy-On-Write)方式获得父进程的数据空间, 堆和栈副本, 这其中也包含文件
> 描述符. 
>
> 在 fork 成功时, 父子进程中相同的文件描述符指向系统文件表的同一项(共享同一文件偏移量). 
>
> 接着, 一般调用 exec 指向另一个程序, 此时会用全新的程序替换子进程的正文, 数据, 堆和栈等. 此时保存文件描述符的变量当然
> 就不存在了, 就无法关闭无用的文件描述符了. 所以通常会fork子进程后在子进程中直接close掉无用的文件描述符, 然后执行exec.
>
> 但是, 在 fork 子进程时无法准确获得打开的文件描述符(包括socket句柄等), 我们期望的是能在fork子进程前打开的文件描述符
> 都被指定好, 这样在 fork 后执行 exec 时关闭. 这就是所谓的 close-on-exec.

- `SO_REUSEPORT`(0x0F) 套接字地址(IP+Port)重用. 支持多个进程或线程绑定到同一个地址端口.

允许将多个 `AF_INET` 或 `AF_INET6` 套接字绑定到相同的套接字地址(socket address). 在套接字上调用bind之前, 必须在
每个套接字(包括第一个套接字) 上设置此选项.

为了防止端口劫持, 绑定到同一地址的所有进程都必须具有相同的有效UID. 此选项可与TCP和UDP套接字一起使用.


- `SO_REUSEADDR`(0x02)

设置用于验证bind调用中提供的地址的规则应允许重用本地地址. 对于AF_INET套接字, 这意味着套接字可以bind, 除非有活动的侦听
套接字绑定到该地址. 当 listen 套接字通过 "特定端口" 绑定到 INADDR_ANY 时, 则不能将其端口绑定到任何本地地址. 参数是
一个整数布尔值标志. 

- `SO_KEEPALIVE`(0x09), 开启TCP的KeepAlive机制

- `SO_LINGER` (0x0d), TCP 独有选项

启用后, close 或 shutdown将不会返回, 直到成功发送了套接字的所有排队消息或达到了超时为止. 否则, 该调用将立即返回, 并在
后台完成关闭. 当套接字作为exit的一部分关闭时, 它始终存在后台.

- `SO_BROADCAST` (0x06), 开启 UDP 广播选项

- `SO_BINDTODEVICE` (0x19), 将此套接字绑定到特定的设备. 例如 "eth0", 如果名称为空字符串或选项长度为零, 则删除套接
字设备绑定. 参数选项是网口的名称. 如果套接字绑定到网口, 则套接字仅处理从该特定网口接收的数据包. 注意: 这仅适用于某些套接
字类型, 尤其是AF_INET套接字. 数据包套接字不支持此功能.


> 使用 `recvmmsg` 代替 `recvmsg`, 调用 recvmsg 时会将收到的数据从内核空间拷贝到用户空间, 每调用一次就会产生一次内核
开销. linux 2.6.33 增加了 `recvmmsg`, 允许用户一次性接收多个数据包. (UDP)



### 通用多播编程

在广域网上广播的时候, 其中的交换机和路由器只向需要获取数据的主机复制并转发数据.

主机可以向路由器请求加入或退出某个组, 网络中的路由器和交换机有选择地复制并传输数据, 将数据仅仅传输给组内的主机. 

多播的这种功能, 可以一次将数据发送到多个主机, 又能保证不影响其他不需要(未加入组)的主机的其他通信.

广域网多播地址:

局部多播: 224.0.0.0 ~ 224.0.0.255, 路由协议和其他保留的地址, 路由器并不转发此范围的IP包.

预留多播: 224.0.1.0 ~ 238.255.255.255, 全球范围或网络协议.

管理权限多播: 239.0.0.0 ~ 239.255.255.255, 组织内部使用, 类似私有IP地址


多播(组播)的实现:

1) 找到要进行多播使用的网卡, 然后监听本机合适的地址和服务端口

2) 将应用(网卡)加入到多播组中, 它就可以从组中监听包信息, 当然还可以对包传送进行更多的控制设置.

3) 应用收到包后还可以检查包是否来自这个组的包.


同一个应用可以加入多个组中, 多个应用也可以加入到同一个组中.

