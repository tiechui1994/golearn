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

3. `connected` 的 `*UDPConn` 可以调用 `WriteToUDP` 方法吗?

**不可以**, 因为目标地址已经设置. 即使是相同的目标地址也不可以.

4. `connected` 的 `*UDPConn` 如果调用 `Closed` 以后可以调用 `WriteToUDP` 方法吗?

**不可以**

5. `connected` 的 `*UDPConn` 可以调用 `ReadFromUDP` 方法吗?

**可以**, 但是它的功能基本和 `Read` 一样, 只能和它 `connected` 的对端通信.

6. `*UDPConn` 还有一个通用的 `WriteMsgUDP(b, oob []byte, addr *UDPAddr)`, 同时支持 `connected` 和 
`unconnected` 的 UDPConn:

a) 如果 `UDPConn` 还未连接, 那么会发送数据到addr

b) 如果 `UDPConn` 已连接, 那么它会发送数据给连接的对端, 这种状况下会忽略 addr


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