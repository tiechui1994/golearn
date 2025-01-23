## 监听文件变化的实现

### Linux

Linux下 inotify 特性:

inotify是内核一个特性, 可以用来监控目录, 文件的读写等事件. 当监控目标是目录时, inotify 除了会监控目录本身, 还会监
控目录中的文件.

inotify 的监控功能由如下几个 syscall 组成: `inotify_init1`, `inotify_add_watch`, `inotify_rm_watch`, `read` 和 `close`.

inotify的主要操作基于 `inotify_init1` 返回的 inotify 文件描述符, 该描述符的作用类似于 epoll 的 epoll_fd. 

注: inotify 在监控目录的时候, 不支持对目录的递归监控, 即只能监控一层目录, 如果需要递归监控, 就需要将这些目录通过 
`inotify_add_watch` 添加.

伪代码:

```

ifd := InotifyInit1()

wfd := InotifyAddWatch(ifd, path, IN_OPEN|IN_CLOSE|IN_CREATE|IN_DELETE|IN_MODIFY|IN_ATTRIB|...)

for (;;;) { 
    data, ret = read(wfd)
    if ret == EAGAIN {
       continue
    }

    // parse data loop
    // header: struct {
    //   	wd     int32   
    //   	mask   uint32  // contains event mask
    //   	cookie uint32  // unused
    //   	len    uint32
    // }
    // data: length is header len. is name
}
```

### Windows

Windows 下 `ReadDirectoryChangesW` 可以监控目录的变化


### fsnotify

核心: inotify系统调用产生的 fd 可以使用 epoll 进行去监听(类似于网络fd). 对于 inotify 当中监听的文件的变更都会使得
inotifyfd 就绪, 从而可以从 inotifyfd 当中读取就绪的内容(文件变化的情况)

fsnotify 的工作原理如下:

```
1. 创建 inotify 文件描述符. (即 SYS_INOTIFY_INIT1 系统调用)

2. 创建 pipe 管道. (程序唤醒退出的作用, 只是辅助作用.)

3. 创建 epoller. (只是为了高效率呗, 其实 select, polle 也可以的).

4. 将 inotifyfd, pipefd[0] 的 EPOLLIN 事件添加到 epoll 当中. (就是 epoll 的添加注册事件呗, 为后续的监听等待准备)

5. 新开一个 goroutine, 从 epoll 当中获取就绪的事件(阻塞调用), 如果文件有变化时, 就可以从 inotifyfd 当中读取件内容 
InotifyEvent, 里面包含了事件, 文件描述符fd, 变更文件的临时名称. 存储格式:

InotifyEvent:
[
    Wd     int32  // 变化的文件fd
    Mask   uint32 
    Cookie uint32
    Len    uint32 // Data 长度
    Data   []byte // 文件的路径
]

注意: 一次性可能存在多个事件发生.

如果是 `pipefd[0]` 有事件发生, 那么就是唤醒程序退出.

6. 将文件/目录添加到 inotifyfd 监控当中(相对于监听是异步的). 主要监控的事件有: IN_MOVED_TO, IN_CREATE, IN_MOVED_FROM, 
IN_ATTRIB, IN_MODIFY, IN_MOVE_SELF, IN_MOVE_SELF, IN_DELETE_SELF.

IN_MOVED_TO, IN_CREATE 是新增文件
IN_DELETE_SELF, IN_MODIFY 是删除文件
IN_MODIFY 是修改文件
IN_MOVE_SELF, IN_MOVED_FROM 是文件重命名
IN_ATTRIB 是文件权限
```

- 创建 inotify 监听

```cgo
func NewWatcher() (*Watcher, error) {
    poller, err := newFdPoller(fd) // epoll
    if err != nil {
        unix.Close(fd)
        return nil, err
    }
    w := &Watcher{
        fd:       poller.fd,
        poller:   poller,
        watches:  make(map[string]*watch), // 监听的文件/目录
        paths:    make(map[int]string),    // fd <-> path 映射
        Events:   make(chan Event), // 事件
        Errors:   make(chan error), // 错误
        done:     make(chan struct{}), // 结束的标记
        doneResp: make(chan struct{}), // 结束唤醒的标记
    }
    
    go w.readEvents()
    return w, nil
}
```


// epoll 构建
// epoll { `pipefd[0]`, `inotifyfd` }, 在这里 `inotifyfd` 是核心, 文件所有的变动都得通过它来告知. `pipefd[0]`
// 只是一个辅助, 用于程序唤醒退出的作用.

> 创建一个 pipe(匿名管道), 会打开两个文件描述符, `fd[0]` 是读, `fd[1]` 是写.

```cgo
// epoll 创建, 添加关注的事件
func newFdPoller() (*fdPoller, error) {
    // 创建 inotify_fd
    fd, errno := unix.InotifyInit1(unix.IN_CLOEXEC)
    if fd == -1 {
        return nil, errno
    }
    
    var errno error
    poller := new(fdPoller)
    poller.fd = fd
    poller.epfd = -1
    poller.pipe[0] = -1
    poller.pipe[1] = -1
    defer func() {
        if errno != nil {
            poller.close()
        }
    }()
    
    // 匿名管道 pipe, 其中 pipe[0] 是读, pipe[1]是写.
    errno = unix.Pipe2(poller.pipe[:], unix.O_NONBLOCK|unix.O_CLOEXEC)
    if errno != nil {
        return nil, errno
    }
    
    // epoll
    poller.epfd, errno = unix.EpollCreate1(unix.EPOLL_CLOEXEC)
    if poller.epfd == -1 {
        return nil, errno
    }
    
    // 注册 inotifyfd "读" 的事件到 epoll
    event := unix.EpollEvent{
        Fd:     int32(poller.fd),
        Events: unix.EPOLLIN,
    }
    errno = unix.EpollCtl(poller.epfd, unix.EPOLL_CTL_ADD, poller.fd, &event)
    
    // 注册 pipfd[0] "读" 的事件到 epoll
    event = unix.EpollEvent{
        Fd:     int32(poller.pipe[0]),
        Events: unix.EPOLLIN,
    }
    errno = unix.EpollCtl(poller.epfd, unix.EPOLL_CTL_ADD, poller.pipe[0], &event)
    
    return poller, nil
}
```


// epoll 监听(EPOLLIN 事件)

```cgo
func (poller *fdPoller) wait() (bool, error) {
    // 监听 EPOLLIN, 同时 EPOLLHUP, EPOLLERR 是无须监听, 任何时候都会发生
    // 2*3+1
    events := make([]unix.EpollEvent, 7)
    for {
        n, errno := unix.EpollWait(poller.epfd, events, -1) 
        if n == -1 {
            if errno == unix.EINTR {
                continue
            }
            return false, errno
        }
        if n == 0 {
            continue
        }
        if n > 6 {
            return false, errors.New("epoll_wait returned more events than I know what to do with")
        }
        
        // n个就绪事件
        ready := events[:n]
        var epollhup, epollerr, epollin bool
        for _, event := range ready {
            // inotifyfd, 真正关心的事件
            if event.Fd == int32(poller.fd) {
                if event.Events&unix.EPOLLHUP != 0 {
                    epollhup = true
                }
                if event.Events&unix.EPOLLERR != 0 {
                    epollerr = true
                }
                if event.Events&unix.EPOLLIN != 0 {
                    epollin = true
                }
            }
            
            // pipefd[0], 一般是错误事件.
            if event.Fd == int32(poller.pipe[0]) {
                if event.Events&unix.EPOLLHUP != 0 {
                    // Write pipe descriptor was closed, by us. This means we're closing down the
                    // watcher, and we should wake up.
                }
                if event.Events&unix.EPOLLERR != 0 {
                    // If an error is waiting on the pipe file descriptor.
                    // This is an absolute mystery, and should never ever happen.
                    return false, errors.New("Error on the pipe descriptor.")
                }
                if event.Events&unix.EPOLLIN != 0 {
                    // 常规的唤醒操作, 唤醒之后需要立即清理写入的缓存(因为写入缓存, 导致唤醒事件的发生).
                    // 从 pipefd[0] 当中读出缓存. 因为 wake() 就是向 pipfd[1] 当中写入了缓存.
                    // 这样做的目的: 当需要 close 的时候, 程序创建调用 wake() 从而使 epoll EpollWait() 快速
                    // 唤醒, 从而程序退出.
                    err := poller.clearWake() 
                    if err != nil {
                        return false, err
                    }
                }
            }
        }
    
        if epollhup || epollerr || epollin {
            return true, nil
        }
        return false, nil
    }
}
```

- 监听(异步)

// 将文件的变化通过 chan 发送给客户端. 

```cgo
func (w *Watcher) readEvents() {
    var (
        buf   [unix.SizeofInotifyEvent * 4096]byte // Buffer for a maximum of 4096 raw events
        n     int                                  // Number of bytes read with read()
        errno error                                // Syscall errno
        ok    bool                                 // For poller.wait
    )
    
    defer close(w.doneResp)
    defer close(w.Errors)
    defer close(w.Events)
    defer unix.Close(w.fd)
    defer w.poller.close()

    for {
        if w.isClosed() {
            return
        }
        
        // 监听
        ok, errno = w.poller.wait()
        if errno != nil {
            select {
            case w.Errors <- errno:
            case <-w.done:
                return
            }
            continue
        }
        if !ok {
            continue
        }
        
        // 读取就绪的事件 InotifyEvent 
        n, errno = unix.Read(w.fd, buf[:])
        if errno == unix.EINTR {
            continue
        }
        
        // 再次判断, 双保险
        if w.isClosed() {
            return
        }
        
        // 接下来就是一个事件协议的处理, 中规中矩.
        // 读取到的内容不足一个事件头
        if n < unix.SizeofInotifyEvent {
            var err error
            if n == 0 {
                err = io.EOF // EOF
            } else if n < 0 {
                err = errno // reading error
            } else {
                err = errors.New("notify: short read in readEvents()") // too short
            }
            
            // Send Error
            select {
            case w.Errors <- err:
            case <-w.done:
                return
            }
            continue
        }
        
        // 处理读取的事件, 可能是多个
        var offset uint32
        for offset <= uint32(n-unix.SizeofInotifyEvent) {
            raw := (*unix.InotifyEvent)(unsafe.Pointer(&buf[offset])) // 直接转换为 InotifyEvent
    
            mask := uint32(raw.Mask) // 事件
            nameLen := uint32(raw.Len) // 发生事件的文件的名称(注意是文件, 不是目录)
            
            // Send Error
            if mask&unix.IN_Q_OVERFLOW != 0 {
                select {
                case w.Errors <- ErrEventOverflow:
                case <-w.done:
                    return
                }
            }
    
            w.mu.Lock()
            name, ok := w.paths[int(raw.Wd)]
            // 监听的文件/目录被删除了, 则移除相应的内容
            if ok && mask&unix.IN_DELETE_SELF == unix.IN_DELETE_SELF {
                delete(w.paths, int(raw.Wd))
                delete(w.watches, name)
            }
            w.mu.Unlock()
            
            // 文件名称(注意: 不是目录)
            if nameLen > 0 {
                bytes := (*[unix.PathMax]byte)(unsafe.Pointer(&buf[offset+unix.SizeofInotifyEvent]))[:nameLen:nameLen]
                name += "/" + strings.TrimRight(string(bytes[0:nameLen]), "\000")
            }
    
            event := newEvent(name, mask)
            
            // Send Event
            if !event.ignoreLinux(mask) {
                select {
                case w.Events <- event:
                case <-w.done:
                    return
                }
            }
    
            offset += unix.SizeofInotifyEvent + nameLen
        }
    }
}

```

- 添加监控的目录/文件(相对于监听是异步的)

// 核心在于 InotifyAddWatch 系统调用. 与 epoll 添加关注的 fd 类似, 只是格式不大一样.

```cgo
func (w *Watcher) Add(name string) error {
    name = filepath.Clean(name)
    if w.isClosed() {
        return errors.New("inotify instance already closed")
    }
    
    // 文件/目录需要通知的事件
    var flags uint32 = unix.IN_MOVED_TO | unix.IN_MOVED_FROM |
        unix.IN_CREATE | unix.IN_ATTRIB | unix.IN_MODIFY |
        unix.IN_MOVE_SELF | unix.IN_DELETE | unix.IN_DELETE_SELF
    
    w.mu.Lock()
    defer w.mu.Unlock()
    
    watchEntry := w.watches[name]
    if watchEntry != nil {
        flags |= watchEntry.flags | unix.IN_MASK_ADD // 对于已存在的条目.
    }
    wd, errno := unix.InotifyAddWatch(w.fd, name, flags) // 系统调用
    if wd == -1 {
        return errno
    }
    
    if watchEntry == nil {
        w.watches[name] = &watch{wd: uint32(wd), flags: flags}
        w.paths[wd] = name // 注册
    } else {
        watchEntry.wd = uint32(wd)
        watchEntry.flags = flags
    }
    
    return nil
}
```