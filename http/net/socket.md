# socket 

- socket

头文件: `#include <sys/types.h>`, `#include <sys/socket.h>`

```cgo
int socket(int domain, int type, int protocol);
```

> 说明:
>
>



- setsockopt, 设置 socket 状态

头文件: `#include <sys/types.h>`, `#include <sys/socket.h>`


函数定义:

```cgo
int setsockopt(int s, int level, int optname, const void* optval, socklen_t optlen);
int getsockopt(int s, int level, int optname, const void* optval, socklen_t optlen);
``` 

> 说明: 设置参数 s 所指定的 socket 状态. 参数 level 代表欲设置的网络层, 一般设为 SOL_SOCKET 以存取 socket 层. 参数 optname 代表设置的
> 选项, 有以下几种值:
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