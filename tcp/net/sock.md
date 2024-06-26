# socket 

##  socket 编程的流程

![image](/images/tcp_net_socket.png)


## socket 相关的函数

- socket

头文件: `#include <sys/types.h>`, `#include <sys/socket.h>`

```cgo
int socket(int domain, int type, int protocol);
```

> 说明: socket() 用来建立一个新的socket. 也就是系统注册, 通知系统建立一通信端口. 参数 domain 指定使用何种的地址类
> 型, 完整的定义在 `/usr/include/bits/socket.h` 内. 常见的地址类型:
>
> PF_UNIX/PF_LOCAL/AF_UNIX/AF_LOCAL, unix进程通信协议
> PF_INET/AF_INET, ipv4网络协议
> PF_INET6/AF_INET6, ipv6网络协议
> PF_PACKET/AF_PACKET, 初级封包接口
>
> 参数 type 有以下几种值:
> 
> 1. SOCK_STREAM 提供双相连接且可信赖的数据流, 即 TCP. 支持 OOB 机制, 在所有数据传送前必须使用connect()来建立连
> 线状态. 
> 2. SOCK_DGRAM 使用不连续不可信赖的数据包连接
> 3. SOCK_SEQPACKET 提供连续可信赖的数据包连接
> 4. SOCK_RAW 提供原始网络协议存取
> 5. SOCK_RDM 提供可信赖的数据包连接
> 6. SOCK_PACKET 提供和网络驱动程序直接通信. 
> 
> protocol 用来指定 socket 所使用的传输协议编号, 通常此参数不用管, 设为 0 即可.
>
> 
> 返回值: 0表示成功, -1表示失败
> 错误代码:
> 1. EPROTONOSUPPORT 参数 domain 指定的类型不支持参数 type 或 protocol 指定的协议
> 2. ENFILE 核心内存不足, 无法建立新的 socket 结构
> 3. EMFILE 进程文件表溢出, 无法再建立新的socket
> 4. EACCESS 权限不足, 无法建立 type 或 protocol 指定的协议
> 5. ENOBUFS/ENOMEM 内存不足
> 6. EINVAL 参数domain/type/protocol不合法


- setsockopt, 设置 socket 状态

头文件: `#include <sys/types.h>`, `#include <sys/socket.h>`


函数定义:

```cgo
int setsockopt(int s, int level, int optname, const void* optval, socklen_t optlen);
int getsockopt(int s, int level, int optname, const void* optval, socklen_t optlen);
``` 

> 说明: 设置参数 s 所指定的 socket 状态. 参数 level 代表欲设置的网络层, 一般设为 SOL_SOCKET 以存取 socket 层. 
> 参数 optname 代表设置的选项, 有以下几种值:
>
> - SO_DEBUG, 打开或关闭debug模式
> - SO_REUSEADDR, 打开或关闭地址复用功能. 允许在 bind() 过程中本地地址可重复使用
> - SO_DONTROUTE, 打开或关闭路由查询功能. 发送的数据包不要使用路由设备来传输.
> - SO_BROADCAST, 允许或禁止发送广播数据
> - SO_TYPE, 返回 socket 的类型
> - SO_ERROR, 返回 socket 已发生的错误
> - SO_SNDBUF, 设置发送缓存区大小.
> - SO_RCVBUF, 设置接收缓存区大小.
> - SO_KEEPALIVE, 定期确定连接是否已经终止
> - SO_OOBINLINE, 当收到 OOB 数据时, 马上送至输入设备
> - SO_LINGER, 如果设置此选项, close或shutdown将等到所有套接字里排队的消息成功发送或到达延迟时间后才返回, 否则立即返回.
> - SO_NO_CHECK, 打开或关闭校验和
>
> optval 代表设置的值, optlen 是 optval 的长度.
>
> 返回值: 0表示成功, -1表示错误,错误原因存在于 errno
>
>> 附加说明:
>> 1. EBADF 参数 s 并非合法的 socket
>> 2. ENOTSOCKET 参数 s 为文件描述符, 非 socket
>> 3. ENOPROTOOPT 参数 optname 指定选项不正确
>> 4. EFAULT 参数 optval 指针指向无法存取的内存空间


- bind

头文件: `#include <sys/types.h>`, `#include <sys/socket.h>`

函数定义:

```cgo
int bind(int sockfd, struct sockaddr* addr, int addrlen);
```

> 函数说明: bind() 用来设置参数 sockfd 的 socket 一个名称. 此名称由参数 addr 指向一 sockaddr 结构, 对于不同的
> socket domain 定义了一个通用的数据结构.
>
> ```
> struct sockaddr {
>   unsigned short int sa_family;
>   char sa_data[14];
> }
> ```
>
> sa_family 为调用 socket() 时的 domain 参数, 即 AF_xxx; sa_data 最多使用 14 个字符长度.
>
> sockaddr 结构会因使用不同的socket domain 而有不同结构定义, 例如 AF_INET, 其sockadr 定义为:
> 
> ```
> struct socketaddr_in {
>   unsigned short int sin_family;
>   uint16_t sin_port;
>   struct in_addr sin_addr;
>   unsigned char sin_zero[8];
> }
> 
> struct in_addr {
>   uint32_t s_addr;
> } 
> ```
>
> sin_family 为 sa_family; sin_port 为使用的端口号; sin_addr.s_addr 为IP地址; sin_zero未使用;
>
>
> 返回值: 0表示成功, -1表示失败
> 错误码:
> 1. EBADF 参数 sockfd 非合法 socket 处理代码
> 2. EACCESS 权限不足
> 3. ENOTSOCK 参数 sockfd 为一文件描述符, 非 socket


- listen

头文件: `#include <sys/socket.h>`

函数定义:

```cgo
int listen(int sockfd, int backlog);
```

> 函数说明: listen() 用来等待参数 sockfd 的 socket 连接. 参数 backlog 指定同时处理的最大连接请求, 如果连接数达
> 此上限, 则 client 端将收到 ECONNREFUSED 的错误. listen()并未开始接受连接, 只是设置 socket 为 listen 模式,
> 真正接收 client 端连接的是 accept(). 通常 listen() 会在 socket(), bind() 之后调用, 接着才调用 accept().
>
> 返回值: 0是成功, -1表示失败
> 错误代码:
> 1. EBADF 参数 sockfd 非合法 socket 处理代码
> 2. EACCESS 权限不足
> 3. EOPNOTSUPP指定的 socket 未支持 listen 模式
>
>> 说明: listen() 只适用于 SOCK_STREAM, SOCK_SEQPACKET 的 socket 类型. 如果 socket 为 AF_INET 则参数 
>> backlog 最大值可设至 128


- accpet

头文件: `#include <sys/types.h>`, `#include <sys/socket.h>`

定义函数:

```cgo
int accpet(int sockfd, struct sockaddr* addr, int addrlen);
```

> 函数说明: accpet() 用来接收参数 sockfd 的 socket 连接. 参数 sockfd 必须先经过 bind(), listen() 函数处理过,
> 当有连接进来时 accpet() 会返回一个新的 socket, 往后的数据传送和读取都是经由此socket处理, 而原来的参数 sockfd 的
> socket 继续使用 accpet() 来接收新的连接请求. 连接成功时, 参数 addr 所指的结构会被系统填入远程主机的地址数据, 参数
> addrlen 为 sockaddr 的结构长度.
>
> 返回值: 成功则返回新的 socket, 失败返回 -1
> 错误码:
> 1. EBADF 参数 sockfd 非合法的 socket
> 2. EFAULT 参数 addr 指针指向无法读取的内存空间
> 3. ENOTSOCK 参数 sockfd 为文件描述符, 非 socket
> 4. EOPNOTSUPP 指定的socket并非 SOCK_STREAM
> 5. EPERM 防火墙拒绝此连接
> 6. ENOBUFS 系统缓存内存不足
> 7. ENOMEM 内存不足


- fcntl

头文件: `#include <sys/unistd.h>`, `#include <sys/fcntl.h>`

定义函数:

```cgo
int fcntl(int fd, int cmd);
int fcntl(int fd, int cmd, long arg); // Golang使用
```

> 函数说明: fcntl() 针对(文件)描述符提供控制. 参数 fd 是文件描述符, 针对参数 cmd 的值, fcntl 能够接受第三个参数
> (arg).

fcntl 有5种功能:

1) 复制一个现有的文件描述符 (cmd=F_DUPFD)
2) 获取/设置文件描述符 "标记" (cmd=F_GETFD, F_SETFD)
3) 获取/设置文件 "状态标记" (cmd=F_GETFL, F_SETFL)
4) 获取/设置 "异步I/O所有权" (cmd=F_GETOWN, F_SETOWN)
5) 获取/设置记录锁 (cmd=F_GETLK, F_SETLK)

> Go 当中设置网络 socket 为非阻塞, 使用的是第3种方式

cmd 与 arg 参数:

F_GETFD 获取与文件描述符 fd 的 close-on-exec 标志. 如果返回值与 FD_CLOEXEC 进行与运算结果是 0, 文件保持交叉式访问
exec(), 否则如果通过 exec 运行的话, 文件将被关闭. arg参数被忽略.

F_SETFD 设置 close-on-exec 标志. 该标志以参数 arg 的 FD_CLOEXEC 位决定.

F_GETFL 获取文件状态标记, arg参数被忽略.

F_SETFL 设置文件状态标记, arg是状态标记, 可以更改的几个标志是: O_APPEND, O_NONBLOCK, O_SYNC, O_ASYNC.

F_GETOWN 获取当前正在接收SIGIO 或 SIGURG 信号进程的id或者进程组id, 进程组id返回负值. arg参数被忽略.

F_SETOWN 设置将接收SIGIO和SIGURG 信号的进程id或进程组id. arg是进程id或进程组id(负值)

# socket 选项参数

可以使用 setsockopt(2) 设置这些套接字选项, 并使用 getsockopt(2) 读取所有套接字的套接字选项. 

## SOL_SOCKET 级别的选项

- SO_BROADCAST, 开启或禁止进程发送广播消息的能力. 只有数据报套接字支持广播, 并且还必须是在广播消息的网络上. 

- SO_KEEPALIVE, 启用 TCP keeplive. 设置发送 keepalive 周期. 如果2小时内在该套接字的任一方向上没有数据交换, TCP
就自动给对端发送一个保持存活探测分节(keep-alive probe). 这是一个对端必须响应的TCP分节, 它会导致以下三种情况之一:

1) 对端以期望的ACK响应. 应用程序得不到通知, 在2个小时后, TCP将发出一个探测分节.

2) 对端以RST响应, 它告知本端 TCP: 对端已崩溃且已重新启动. 该套接字的待处理错误被置为 ECONNRESET, 套接字本身被关闭.

3) 对端对保持存活探测分节没有任何响应. 源自Berkeley的TCP将另外发送8个探测字节, 试图得到一个响应. TCP在发出第一个探测
分节后 15 分钟左右后若仍没有收到任何响应则放弃. (Linux也是这样处理的)

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


- SO_RCVBUF, 接收缓存区最大字节数. 当使用 setsockopt(2) 设置该值时, 内核将该值加倍, 并且该加倍值由 getsockopt(2)
返回. 默认值由 `/proc/sys/net/core/rmem_default` 文件设置, 最大允许值由 `/proc/sys/net/core/rmem_max` 文件
设置. 此选项的最小(加倍)值为256.

- SO_RCVBUFFORCE, 特权进程可以执行与 SO_RCVBUF 相同的效果, 并且可以覆盖 rmem_max 限制.

- SO_SNDBUF, 发送缓冲区最大字节数. 当使用 setsockopt(2) 设置该值时, 内核将该值加倍, 并且该加倍值由 getsockopt(2)
返回. 默认值由 `/proc/sys/net/core/wmem_default` 文件设置, 最大允许值由 `/proc/sys/net/core/wmem_max` 文件
设置. 此选项的最小(加倍)值为 2048.

- SO_SNDBUFFORCE, 特权进程可以执行与 SO_SNDBUF 相同的效果, 并且可以覆盖 wmem_max 限制.


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


- SO_RCVTIMEO, SO_SNDTIMEO

这两个选项给套接字的接收和发送设置一个超时值. 可以使用秒和微秒来设置超时. 默认情况下这两个超时都是禁止的.

接收超时影响的函数: read, readv, recv, recvfrom, recvmsg. 发送超时影响的函数: write, writev, send, sendto,
sendmsg.

## IPPROTO_TCP 级别选项

- TCP_MAXSEG, 设置 TCP 连接的最大分节大小(MSS). 返回值是TCP可以发送给对端的最大数据量, 它通常是由对端使用 SYN 分节
告知的 MSS, 除非在TCP选择使用一个比对端通告的 MSSS 小些的值.

- TCP_NODELAY, 开启该选项将禁止 TCP 的 Nagle 算法. 默认情况下该算法是启动的. 

Nagle 算法的目的在于减少广域网上"小分组"的数目. 该算法指出: 如果某个给定连接上有待确认数据, 当用户写入小分组数据时, 写
入的数据不会立即发送出去, 直到 "现有数据被确认" 为止. 这里的 "小分组" 的定义是小于 MSS 的分组.

Nagle 算法常常与 ACK延滞算法(delay ACK alg) 一起使用. 该算法使得TCP在接收到数据后不立即发送ACK, 而是等待一小段时间(
典型值为 50~200ms), 然后发送ACK. TCP 期待在这一小段时间内自身有数据发送回对端, 被延滞的ACK就可以由这些数据捎带, 从而
省掉一个 TCP 分节.

> 注: 对于小分组数据通信, 不太适合使用 Nagle算法和 ACK延滞算法.

举个例子: 假设某个 client 向 server 发送一个 400 字节的请求, 该请求由一个 4 字节的请求类型 + 396字节请求数据组成.
如果 client 先执行一个 4 字节 write 调用, 再执行一个 396 字节的 write 调用, 那么第二个写操作将一直等到 server 的
TCP 确认了第一个写操作的 4 字节数据后才由 client 的 TCP 发送出去. 而且由于 server 应用程序难以在收到其余 396 字节之
前对先收到的 4 字节数据进行操作, server 的 TCP 将延迟该 4 字节数据的 ACK. 有三种方法解决:

1) 使用 writev 而不是两次调用 write. 单个 writev 调用最终导致调用 TCP 只产生一个 TCP分节.

2) 将前4字节和后396字节数据复制到单个缓冲区中, 然后对该缓冲区调用一次 write

3) 设置 TCP_NODELAY 选项. 

## fcntl 函数

fcntl 函数可执行各种描述符控制操作. 

| 操作 | fcntl | POSIX |
| ---- | ----- | --- |
| 设置套接字为非阻塞式 IO | F_SETFL, O_NONBLOCK | fcntl |
| 设置套接字为信号驱动式 IO | F_SETFL, O_ASYNC | fcntl |
| 设置套接字 owner | F_SETOWN | fcntl |
| 获取套接字 owner | F_GETOWN | fcntl |
| 获取套接字接收缓冲区中的字节数 | | |
| 测试套接字是否处于带外标准 | | sockatmark |


- 信号驱动式IO. 通过使用 F_SETFL 命令设置 O_ASYNC 文件状态标记. 一旦套接字状态发生变化, 内核就产生一个SIGIO信号.

- F_SETOWN 命令用于指定接收 SIGIO 和 SIGURG 信号的套接字owner(进程ID或进程组ID). 其中 SIGIO信号是套接字被设置为
信号驱动式IO产生的. SIGURG信号是在新的带外数据到达套接字时产生的.

## 关于 TCP 连接终止

### accept 返回前连接终止

三次握手完成从而建立连接之后, 客户端发送了一个 RST. 在服务器看来, 就在该连接已由 TCP 排队, 等待服务进程调用 accept 的
时候 RST 到达. 稍后, 服务器进程调用 accept.

> 状况模拟: 启动服务器, 调用 socket, bind 和 listen, 然后在调用 accept 之前休眠一段时间. 在服务器进程休眠时, 启动
客户端, 调用socket 和 connect, 一旦 connect 返回, 设置 SO_LINGER 选项以产生一个 RST, 之后调用 close.

在这种情况下, 服务器返回错误码是 ECONNABORTED ("software caused connection abort", 软件引起的连接终止.). 对于
这种错误, 服务器是可以忽略的.

### 服务器进程终止

1) 当执行 kill 命令杀死服务器进程时, 进程当中所有打开的描述符都被关闭. 这就导致向客户端发送一个FIN, 而客户端则响应一个
ACK.

2) 客户端接着将数据发送给服务器. TCP允许这么做, 因为客户端接收到 FIN 只是表示服务器进程已关闭了连接的服务器, 从而不再接
往其中发送任何数据而已. FIN 的接收并没有告知客户端服务器进程已经终止.

3) 当客户端在收到服务器的RST之前, 调用 read 时, 会立即返回 EOF. 当客户端在收到服务器的 RST 之后, 调用 read, 会立即
返回 ECONNRESET("connection rest by peer" 对方复位连接错误)

### 服务器主机崩溃

> 当一个进程向某个已收到RST的套接字执行 write 操作时, 内核向该进程发送一个 SIGPIPE 信号. 该信号的默认行为是终止进程,
因此进程必须捕获它以避免意外的被终止.
> 
> 不论该进程是捕获了该信号并从信号处理函数返回, 还是简单的忽略该信号, write 操作都将返回 EPIPE 错误.
>
> 当向一个已经关闭的TCP发送数据时, 第一次写操作引发 RST, 第二次写引发 SIGPIPE 信号. 写一个已接收了 FIN 的套接字不成
问题, 但是写一个已接收了 RST 的套接字则是一个错误.


> 状况模拟: 在不同的主机上运行客户端和服务器. 先启动服务器, 再启动客户端, 接着客户端发送数据确认连接工作正常. 然后从网络
上断开服务器主机, 并在客户端上发送新的数据. 

当服务器主机崩溃时, 已有的网络连接上不发出任何数据. 此时客户端向服务器发送数据, 客户端TCP会持续重传数据分节, 试图从服务器
上收到一个 ACK. 

首先, 客户端尝试发送 keepalive 包(即使 keepalive 被禁止)

其次, 客户端尝试发送 ICMP 包.

最后, 客户端返回 ENETUNREACH 或 EHOSTUNREACH 错误.


### 服务器主机崩溃后重启

> 模拟方法: 先建立连接, 再从网络上断开服务器主机, 将它关机后重启, 最好把它重新连接到网络中.

当服务器主机崩溃后重启时, 它将丢失掉崩溃前所有的TCP连接信息, 因此服务器对于所有收到来自客户端发送的数据都响应 RST.

### 服务器主机关机

Unix 系统关机时, init 进程通常先给所有进程发送 SIGTERM 信号(该信号可被捕获), 等到一段固定的时间(往往是5到20秒之间),
然后给所有仍在运行的进程发送 SIGKILL 信号. 这么做留给所有运行的进程一小段时间来清除和终止.

## select

select 返回套接字 "就绪" 的条件.

- 满足下列4个条件中的任何一个时, 一个套接字准备好读.

1) 该套接字 "接收缓冲区中的数据字节数" 大于等于 "套接字接收缓冲区低水位标记的当前大小". 对这样的套接字执行读操作不会阻塞
并将返回一个大于0的值(也就是准备好读入的数据). 可以使用 SO_RECVLOWAT 选项设置该套接字的低水位标记. 对于 TCP 和 UDP 
套接字, 其默认值是1.

2) 该连接的读部分关闭(也就是接收了FIN的TCP连接). 对这样的套接字将不阻塞并返回0(也就是EOF).

3) 改套接字是一个"监听套接字"且已完成的连接数不为0. 对这样的套接字的 accept 通常不会阻塞.

4) select 其上有一个套接字错误待处理. 对于这种的套接字的读操作将不阻塞并返回-1, 同时把 errno 设置成准确的错误条件. 这
些待处理错误也可以通过指定 SO_ERROR 套接字选项调用 getsockopt 获取并清除.

- 满足下列4个条件中的任何一个时, 一个套接字准备好写.

1) 该套接字 "发送缓冲区的可用空间字节数" 大于等于 "套接字发送缓冲区低水位标记的当前大小", 并且或者该套接字已连接, 或者该
套接字不需要连接(UDP). 这意味着如果将这样的套接字设置为非阻塞, 写操作将不阻塞并返回一个正值. 可用使用 SO_SNDLOWAT 套接
字选项设置该套接字的发送缓冲区低水位值. 对于 TCP 和 UDP 而言, 默认值通常是 2048.

2) 该连接的写部分关闭. 对这样的套接字写操作将产生 SIGPIPE 信号.

3) 使用非阻塞式 connect 的套接字已建立连接, 或者 connect 以失败告终.

4) select 其上有一个套接字错误待处理. 对于这种的套接字的写操作将不阻塞并返回-1, 同时把 errno 设置成准确的错误条件. 这
些待处理错误也可以通过指定 SO_ERROR 套接字选项调用 getsockopt 获取并清除.

- 如果一个套接字存在带外数据或者处于带外标记, 那么它有异常条件待处理.

> 注: 当某个套接字上发生错误时, 它将由 select 标记为即可读又可写.

## TCP shutdown 函数

终止网络连接的通用方法是调用 close() 函数. 不过 close 有两个限制, 却可以使用 shutdown 来避免.

1. close 把描述符的引用计数减1, 仅在该计数变为0时才关闭套接字. 使用 shutdown 可以不管引用计数就激发 TCP 的正常连接终
止序列.

2. close 终止读和写两个方向的数据传送. 既然 TCP 连接是全双工的, 有时候需要告知对端已经完成了数据发送, 即使对端仍有数据
要发送. 

```cgo
#include <sys/socket.h>

int shutdown(int sockfd, int howto);
```

该函数的行为依赖于 howto 参数的值:

SHUT_RD, 关闭连接读这一半. 套接字中不再有数据可接收, 而且套接字接收缓冲区中现有的数据都被丢弃. 进程不再对这样的套接字调
用任何读函数. 对于一个 TCP 套接字这样调用 shutdown() 函数后, 由该套接字接收的来自对端的任何数据都被确认, 然后丢弃.

SHUT_WR, 关闭连接写这一半. 对于TCP套接字, 称为半关闭.  当前留在套接字发送缓冲区的数据将被丢弃掉, 接着执行 TCP 的正常
连接终止序列. 不管套接字描述符引用计数是否为0, 都会执行. 进程不能再对这样的套接字调用任何写函数.

SHUT_RDWR, 连接的读半部和写半部都关闭. 这与调用 shutdown 两次等效; 第一次调用指定 SHUT_RD, 第二次调用指定 SHUT_WR.
