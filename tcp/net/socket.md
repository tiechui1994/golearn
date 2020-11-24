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
> - SO_REUSEADDR, 允许在 bind() 过程中本地地址可重复使用
> - SO_TYPE, 返回 socket 的类型
> - SO_ERROR, 返回 socket 已发生的错误
> - SO_DONTROUTE, 发送的数据包不要使用路由设备来传输.
> - SO_BROADCAST, 使用广播方式发送
> - SO_SNDBUF, 设置发送缓存区大小
> - SO_RCVBUF, 设置接收缓存区大小
> - SO_KEEPALIVE, 定期确定连接是否已经终止
> - SO_OOBINLINE, 当收到 OOB 数据时, 马上送至输入设备
> - SO_LINGER, 确保数据安全且可靠的发生出去
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
