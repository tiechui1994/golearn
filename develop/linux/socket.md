# socket

## scoket options

可以使用 setsockopt(2) 设置这些套接字选项, 并使用 getsockopt(2) 读取所有套接字的套接字选项.

- SO_ACCEPTCONN

- SO_BINDTODEVICE

- SO_BROADCAST

- SO_BSDCOMPAT

- SO_DEBUG

- SO_DOMAIN

- SO_ERROR

- SO_DONTROUTE

- SO_KEEPALIVE, 启用 TCP keeplive. 设置发送 keepalive 周期.

- SO_LINGER

- SO_MARK

- SO_OOBINLINE, 如果启用此选项, 则 out-of-data 将直接放入接收数据流中. 否则, 只有在接收期间设置了 MSG_OOB 标志时
才会传递 out-of-data

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



- SO_REUSEADDR


- SO_TIMESTAMP, 启用或禁用 SO_TIMESTAMP 控制消息的接收. 时间戳控制消息以 SOL_SOCKET 级别发送, cmsg_data 字段
是一个 struct timeval, 指示在此调用中传递给用户的最后一个数据包的接收时间. 

- SO_TYPE




























