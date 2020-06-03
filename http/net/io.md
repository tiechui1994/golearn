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
