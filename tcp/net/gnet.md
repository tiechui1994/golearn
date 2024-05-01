# gnet 网络模型分析

gent 的运行模式有两种, 分别是 Reactor 和 EventLoop, 它两个的区别在于:

Reactor 模式, 设计模式(design pattern). Reactor 模式(Reactor Pattern) 比较生硬的翻译就是反应器模式. 它是一种
事件处理模式(event handing pattern), 用于处理一个或多个并发输入的服务请求给服务处理器(service handler). 服务处
理器(service handler)将请求多路分解同步派发给业务关联的请求处理器(request handler).


EventLoop 是一个程序结构或者设计模式, 它在程序中是用来等待调度事件或者消息. EventLoop 处理内部或者外部的请求(通
常这些请求都会一直等待, 直到有返回给他们), 当请求符合 EventLoop 所监听的事件, 就会分配给对应的事件处理器.

事件循环常常用于 Reactor 模式的连接器(用于连接请求与处理器, 也就是 Reactor 模式中的 Synchronous Event Demultiplexer).
如果事件提供对于状态接口, 事件循环可以 seleted(符合条件) 事件把这个事件 Polled(抛出去) selected-polled 模式.

在开启了端口重用或 UDP 服务的状况下, gnet 采用的是 EventLoop 模型, 否则采用的是 Reactor 模型.

在 EventLoop 模式下, 创建 N 个 EventLoop, 每个 EventLoop 创建一个 Poller, 每个 EventLoop 启动一个监听器与之绑定,  
当新连接建立时, 会选择启用一个 EventLoop 去 accept, 以及后续的读写事件处理都会在这个 Poller 上进行.

在 EventLoop 模式下, 创建 N+1 个 EventLoop, 每个 EventLoop 创建一个 Poller, 启动 1 个监听器, 所有的 EventLoop 都与
之绑定. 对于 1 个 Main EventLoop, 由其 Poller 只处理 Accept, 然后从 N 个 Sub EventLoop 当中选择一个, 将连接放入其中, 
后续由该EventLoop Poller 处理该连接的读写事件.

EventLoop 实现:

![image](/images/tcp_gnet_el.png)

Reactor 实现:

![image](/images/tcp_gnet_reactor.png)


### EventLoop 模型(Linux)

在这种模型下, 会创建 N (取决于 CPU 核心数, 配置的 NumEventLoop, 最大值不超过 256) 个 eventLoop 对象, 每个 eventLoop
绑定一个 Listener. 每个 eventLoop 都会有一个 Epoll 与之绑定(用于处理请求).
 
初始化时 eventLoop.accept 方法注册到 Epoll 中 (Read Event), 用于接收请求. (polling 当中用于处理 accept, 
read, write)

create N eventLoop:
```cgo
func (eng *engine) activateEventLoops(numEventLoop int) (err error) 
    ...
    ln := eng.ln
    for i := 0; i < numEventLoop; i++ {
        if i > 0 {
            if ln, err = initListener(network, address, eng.opts); err != nil {
                return
            }
        }
        var p *netpoll.Poller
        if p, err = netpoll.OpenPoller(); err == nil {
            el := new(eventloop)
            el.ln = ln
            el.engine = eng
            el.poller = p
            el.buffer = make([]byte, eng.opts.ReadBufferCap)
            el.connections.init()
            el.eventHandler = eng.eventHandler
            // poller 添加 Read 监听事件, Read 触发是在 Poller 的 Polling 当中
            if err = el.poller.AddRead(el.ln.packPollAttachment(el.accept)); err != nil {
                return
            }
            eng.eventLoops.register(el)

            // Start the ticker.
            if el.idx == 0 && eng.opts.Ticker {
                striker = el
            }
        } else {
            return
        }
    }

    ...
}
```

run: 后台执行, Epoll 轮询(核心)
```cgo
func (el *eventloop) run() error {
    if el.engine.opts.LockOSThread {
        runtime.LockOSThread()
        defer runtime.UnlockOSThread()
    }

    // Polling 本身是死循环, 不断接收 poller 当中就绪的 Event, 并调用回调函数处理.
    // 在这里的将事件分发给不同的 handler. 如果 fd 已存在, 则是对 Conn 读写事件, 如果 fd 不存在, 则是新连接事件.
    err := el.poller.Polling(func(fd int, ev uint32) error {
        if c := el.connections.getConn(fd); c != nil {
            // Don't change the ordering of processing EPOLLOUT | EPOLLRDHUP / EPOLLIN unless you're 100%
            // sure what you're doing!
            // Re-ordering can easily introduce bugs and bad side-effects, as I found out painfully in the past.

            // We should always check for the EPOLLOUT event first, as we must try to send the leftover data back to
            // the peer when any error occurs on a connection.
            //
            // Either an EPOLLOUT or EPOLLERR event may be fired when a connection is refused.
            // In either case write() should take care of it properly:
            // 1) writing data back,
            // 2) closing the connection.
            if ev&netpoll.OutEvents != 0 && !c.outboundBuffer.IsEmpty() {
                // 从 outboundBuffer 当中获取数据, 将获取的数据写入到 fd 当中
                // 注: 当 outboundBuffer 为空时, 将 fd 监听的 Write Event 修改为 Read Event.
                if err := el.write(c); err != nil {
                    return err
                }
            }

            // read, 即从 fd 当中读取数据到 buffer 当中, 回调 OnTraffic 进行处理.
            // 同时将数据写入到 inboundBuffer 当中
            if ev&netpoll.InEvents != 0 {
                return el.read(c)
            }
            return nil
        }
        
        // 新的连接事件, 与前面的 activateEventLoops 当中设置回调对应.
        return el.accept(fd, ev)
    })

    if err == errors.ErrEngineShutdown {
        el.engine.opts.Logger.Debugf("event-loop(%d) is exiting in terms of the demand from user, %v", el.idx, err)
        err = nil
    } else if err != nil {
        el.engine.opts.Logger.Errorf("event-loop(%d) is exiting due to error: %v", el.idx, err)
    }

    el.closeConns()
    el.engine.shutdown(err)

    return err
}
```

Polling 逻辑的实现: blocking waiting + infinite loop

```cgo
// Polling blocks the current goroutine, waiting for network-events.
func (p *Poller) Polling(callback func(fd int, ev uint32) error) error {
    el := newEventList(InitPollEventsCap)
    var doChores bool

    msec := -1
    for {
        // blocking 
        n, err := unix.EpollWait(p.fd, el.events, msec)
        if n == 0 || (n < 0 && err == unix.EINTR) {
            msec = -1
            runtime.Gosched() // 手动去调度
            continue
        } else if err != nil {
            logging.Errorf("error occurs in epoll: %v", os.NewSyscallError("epoll_wait", err))
            return err
        }
        msec = 0

        // event ready
        for i := 0; i < n; i++ {
            ev := &el.events[i]

            // notice: poller.fd 是 epoll 创建时的 fd
            //         poller.efd 是 epoll 自己关注的事件 fd(只关注 Read Event), 这个 efd 会注册到 poller 轮训当中
            if fd := int(ev.Fd); fd != p.efd {
                // 非 poller 关注的 event, 这里的 callback 就是 run 当中注册的函数, 事件分发.
                switch err = callback(fd, ev.Events); err {
                case nil:
                case errors.ErrAcceptSocket, errors.ErrEngineShutdown:
                    return err
                default:
                    logging.Warnf("error occurs in event-loop: %v", err)
                }
            } else { 
                // poller 关注的 event, poller 被唤醒, 执行 Queue 当中的 task
                doChores = true
                _, _ = unix.Read(p.efd, p.efdBuf)
            }
        }

        if doChores {
            doChores = false
            // 高优先级队列
            task := p.urgentAsyncTaskQueue.Dequeue()
            for ; task != nil; task = p.urgentAsyncTaskQueue.Dequeue() {
                switch err = task.Run(task.Arg); err {
                case nil:
                case errors.ErrEngineShutdown:
                    return err
                default:
                    logging.Warnf("error occurs in user-defined function, %v", err)
                }
                queue.PutTask(task)
            }
            // 低优先级队列, 一次最多执行256个
            for i := 0; i < MaxAsyncTasksAtOneTime; i++ {
                if task = p.asyncTaskQueue.Dequeue(); task == nil {
                    break
                }
                switch err = task.Run(task.Arg); err {
                case nil:
                case errors.ErrEngineShutdown:
                    return err
                default:
                    logging.Warnf("error occurs in user-defined function, %v", err)
                }
                queue.PutTask(task)
            }

            // 这里的逻辑比较有意思:
            // poller.wakeupCall 是 poller 的一个唤醒调用的标记. 0 表示可被唤醒, 1 表示已经被唤醒.
            // poller.efd 是 poller 自身关注的 Read Event, 只有向其写入数据, 才能触发 Read Event.
            // 这里使用 CAS 无锁方式进行安全操作数据.
            atomic.StoreInt32(&p.wakeupCall, 0)
            if (!p.asyncTaskQueue.IsEmpty() || !p.urgentAsyncTaskQueue.IsEmpty()) && atomic.CompareAndSwapInt32(&p.wakeupCall, 0, 1) {
                switch _, err = unix.Write(p.efd, b); err {
                case nil, unix.EAGAIN:
                default:
                    doChores = true
                }
            }
        }

        // 调整 list 大小
        if n == el.size {
            el.expand()
        } else if n < el.size>>1 {
            el.shrink()
        }
    }
}
```

accept: 针对 listener fd 的 Read Event 被调用的函数
```cgo
func (el *eventloop) accept(fd int, ev netpoll.IOEvent) error {
    if el.ln.network == "udp" {
        return el.readUDP(fd, ev)
    }

    nfd, sa, err := unix.Accept(el.ln.fd)
    if err != nil {
        switch err {
        case unix.EINTR, unix.EAGAIN, unix.ECONNABORTED:
            // ECONNABORTED means that a socket on the listen
            // queue was closed before we Accept()ed it;
            // it's a silly error, so try again.
            return nil
        default:
            el.getLogger().Errorf("Accept() failed due to error: %v", err)
            return errors.ErrAcceptSocket
        }
    }

    // conn 的属性设置. 
    if err = os.NewSyscallError("fcntl nonblock", setNonBlock(nfd, true)); err != nil {
        return err
    }
    remoteAddr := socket.SockaddrToTCPOrUnixAddr(sa)
    if el.engine.opts.TCPKeepAlive > 0 && el.ln.network == "tcp" {
        err = socket.SetKeepAlivePeriod(nfd, int(el.engine.opts.TCPKeepAlive/time.Second))
        logging.Error(err)
    }

    // 注册 conn 的 Read Event 到 poller  
    c := newTCPConn(nfd, el, sa, el.ln.addr, remoteAddr)
    if err = el.poller.AddRead(&c.pollAttachment); err != nil {
        return err
    }
    el.connections.addConn(c, el.idx)
    return el.open(c)
}
```

open: TCP 刚建立, 需要触发 OnOpen 回调, 如果用户需要在 TCP 连接刚建立发送数据, 注册 conn 的 Write Event 到 poller 

```cgo
func (el *eventloop) open(c *conn) error {
    c.opened = true

    out, action := el.eventHandler.OnOpen(c)
    if out != nil {
        if err := c.open(out); err != nil {
            return err
        }
    }

    // outboundBuffer 是 conn 的发送数据缓冲区. 一旦该缓冲区当中不为空, 就需要触发 Write Event
    if !c.outboundBuffer.IsEmpty() {
        if err := el.poller.AddWrite(&c.pollAttachment); err != nil {
            return err
        }
    }

    return el.handleAction(c, action)
}
```

### Reactor 模型(Linux)

(1 Main)EventLoop( polling 只处理 accept ) + (N Sub)EventLoop (  polling 处理 read, write)

在这种模型下, 会创建 N (取决于 CPU 核心数, 配置的 NumEventLoop, 最大值不超过 256) 个 Sub eventLoop 对象. 每个
eventLoop 都会与一个 Epoll 与之绑定 (只处理 read, write). 

创建一个 Main eventLoop, 绑定一个 Epoll, 并将 accept 方法注册到 Epoll 中(Read Event), 只用于接收请求.

accept: 绝大多数逻辑与 eventLoop 的 accept 类似. 只是在连接建立后, 通过 LSB 选择一个 EventLoop, 将这个请求交给
该 EventLoop (触发了一个 UrgentTrigger 事件, 添加 Read Event), 后续由该 EventLoop 处理.

// create N Sub EventLoop, 1 Main EventLoop  
```cgo
func (eng *engine) activateReactors(numEventLoop int) error {
    // N 个 Sub EventLoop, 每个 EventLoop 都绑定一个 Poller
    // 注: OpenPoller() 当中会监听一个用于处理优先级任务的 Read Event (efd)
    for i := 0; i < numEventLoop; i++ {
        if p, err := netpoll.OpenPoller(); err == nil {
            el := new(eventloop)
            el.ln = eng.ln
            el.engine = eng
            el.poller = p
            el.buffer = make([]byte, eng.opts.ReadBufferCap)
            el.connections.init()
            el.eventHandler = eng.eventHandler
            eng.eventLoops.register(el)
        } else {
            return err
        }
    }
    
    // 后台运行 N 个 Sub EventLoop 的 Polling, 主要用于处理 Read, Write Event
    eng.startSubReactors()

    // 启动 Main EventLoop, 添加 listenr 的 fd 的 Read Event 到该 poller 中
    if p, err := netpoll.OpenPoller(); err == nil {
        el := new(eventloop)
        el.ln = eng.ln
        el.idx = -1
        el.engine = eng
        el.poller = p
        el.eventHandler = eng.eventHandler
        if err = el.poller.AddRead(eng.ln.packPollAttachment(eng.accept)); err != nil {
            return err
        }
        eng.acceptor = el

        // 后台运行 Main EventLoop, Polling, 主要用于处理 Accept 
        eng.workerPool.Go(el.activateMainReactor)
    } else {
        return err
    }

    // Start the ticker.
    if eng.opts.Ticker {
        eng.workerPool.Go(func() error {
            eng.acceptor.ticker(eng.ticker.ctx)
            return nil
        })
    }
    return nil
}
```

Main EventLoop 后台任务:
```cgo
func (el *eventloop) activateMainReactor() error {
    if el.engine.opts.LockOSThread {
        runtime.LockOSThread()
        defer runtime.UnlockOSThread()
    }
    
    // Polling 回调  engine.accept()
    // 相较于 EventLoop.run, 这里只处理了 Accept 
    err := el.poller.Polling(func(fd int, ev uint32) error { return el.engine.accept(fd, ev) })
    if err == errors.ErrEngineShutdown {
        el.engine.opts.Logger.Debugf("main reactor is exiting in terms of the demand from user, %v", err)
        err = nil
    } else if err != nil {
        el.engine.opts.Logger.Errorf("main reactor is exiting due to error: %v", err)
    }

    el.engine.shutdown(err)

    return err
}

func (eng *engine) accept(fd int, _ netpoll.IOEvent) error {
    nfd, sa, err := unix.Accept(fd)
    if err != nil {
        switch err {
        case unix.EINTR, unix.EAGAIN, unix.ECONNABORTED:
            // ECONNABORTED means that a socket on the listen
            // queue was closed before we Accept()ed it;
            // it's a silly error, so try again.
            return nil
        default:
            eng.opts.Logger.Errorf("Accept() failed due to error: %v", err)
            return errors.ErrAcceptSocket
        }
    }

    if err = os.NewSyscallError("fcntl nonblock", setNonBlock(nfd, true)); err != nil {
        return err
    }
    remoteAddr := socket.SockaddrToTCPOrUnixAddr(sa)
    if eng.opts.TCPKeepAlive > 0 && eng.ln.network == "tcp" {
        err = socket.SetKeepAlivePeriod(nfd, int(eng.opts.TCPKeepAlive.Seconds()))
        logging.Error(err)
    }

    // 选择一个 EventLoop, 将当前的 conn 添加到其中, 后续由该 EventLoop 的 Poller 处理该 conn 
    // 的 Read, Write Event
    el := eng.eventLoops.next(remoteAddr)
    c := newTCPConn(nfd, el, sa, el.ln.addr, remoteAddr)
    // 这里会触发选择的 EventLoop 的 Poller 的 efd 的 Read Event
    err = el.poller.UrgentTrigger(el.register, c)
    if err != nil {
        eng.opts.Logger.Errorf("UrgentTrigger() failed due to error: %v", err)
        _ = unix.Close(nfd)
        c.release()
    }
    return nil
}

// 添加 conn 到 EventLoop 当中
func (el *eventloop) register(itf interface{}) error {
    c := itf.(*conn)
    if err := el.poller.AddRead(&c.pollAttachment); err != nil {
        _ = unix.Close(c.fd)
        c.release()
        return err
    }

    el.connections.addConn(c, el.idx)

    if c.isDatagram {
        return nil
    }
    return el.open(c)
}
```

Sub EventLoop 后台任务:
```cgo
func (el *eventloop) activateSubReactor() error {
    if el.engine.opts.LockOSThread {
        runtime.LockOSThread()
        defer runtime.UnlockOSThread()
    }

    // 相较于 EventLoop.run, 这里只处理了 Read, Write Event
    err := el.poller.Polling(func(fd int, ev uint32) error {
        if c := el.connections.getConn(fd); c != nil {
            // Don't change the ordering of processing EPOLLOUT | EPOLLRDHUP / EPOLLIN unless you're 100%
            // sure what you're doing!
            // Re-ordering can easily introduce bugs and bad side-effects, as I found out painfully in the past.

            // We should always check for the EPOLLOUT event first, as we must try to send the leftover data back to
            // the peer when any error occurs on a connection.
            //
            // Either an EPOLLOUT or EPOLLERR event may be fired when a connection is refused.
            // In either case write() should take care of it properly:
            // 1) writing data back,
            // 2) closing the connection.
            if ev&netpoll.OutEvents != 0 && !c.outboundBuffer.IsEmpty() {
                if err := el.write(c); err != nil {
                    return err
                }
            }
            if ev&netpoll.InEvents != 0 {
                return el.read(c)
            }
            return nil
        }
        return nil
    })

    if err == errors.ErrEngineShutdown {
        el.engine.opts.Logger.Debugf("event-loop(%d) is exiting in terms of the demand from user, %v", el.idx, err)
        err = nil
    } else if err != nil {
        el.engine.opts.Logger.Errorf("event-loop(%d) is exiting due to error: %v", el.idx, err)
    }

    el.closeConns()
    el.engine.shutdown(err)

    return err
}
```
