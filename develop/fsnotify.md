## 监听文件变化的实现

> Linux下inotify特性:
>
>inotify是内核一个特性, 可以用来监控目录, 文件的读写等事件. 当监控目标是目录时, inotify除了会监控目录本身, 还会监控目
录中的文件. inotify的监控功能由如下几个系统调用组成: inotify_init1, inotify_add_watch, inotify_rm_watch,
read 和 close.
>
>inotify的主要操作基于inotify_init1返回的 inotify 文件描述符, 该描述符的作用类似于 epoll 的 epoll_fd. inotify 
在监控目录的时候, 不支持对目录的地柜监控, 即只能监控一层目录, 如果需要地柜监控, 就需要将这些目录通过 inotify_add_watch
添加进来.


fsnotify 的工作原理如下:

```
1. 创建 inotify 文件描述符.(即 SYS_INOTIFY_INIT1 系统调用)

2. 创建 pipe 管道. (非阻塞, 主要的作用是唤醒 goroutine 的作用)

3. 创建 epoll 文件描述符, 通过 epoll 进行事件监听.

4. 将 inotify_fd, pipe_fd 的 EPOLLIN 事件添加到 epoll 当中. (SYS_EPOLL_CTL), 也就说当 inotify_fd 或 pipe_fd 
有事件就绪(EPOLLIN, EPOLLHUP, EPOLLERR)之后可以从 epoll 当中获取就绪的事件(SYS_EPOLL_WAIT)

5. 将文件/目录添加到 inotify_fd 监控当中 (SYS_INOTIFY_ADD_WATCH). 主要监控的事件有: IN_MOVED_TO, IN_CREATE,
IN_MOVED_FROM, IN_ATTRIB, IN_MODIFY, IN_MOVE_SELF, IN_MOVE_SELF, IN_DELETE_SELF.

IN_MOVED_TO, IN_CREATE 是新增文件
IN_DELETE_SELF, IN_MODIFY 是删除文件
IN_MODIFY 是修改文件
IN_MOVE_SELF, IN_MOVED_FROM 是文件重命名
IN_ATTRIB 是文件权限

6. 开启一个 goroutine, 不断从 epoll 当中读取就绪的事件(SYS_EPOLL_WAIT), 当文件有变化时, 从 inotify_fd 当中读取
事件内容 InotifyEvent, 里面包含了事件, 文件描述符fd, 变更文件的临时名称. 存储格式如下:

InotifyEvent: (头部)
{
    Wd     int32
	Mask   uint32
	Cookie uint32
	Len    uint32
}

Data: 长度是 Len

注意: 一次性可能存在多个事件发生.
```

- 创建 inotify 监听

```cgo
func NewWatcher() (*Watcher, error) {
	// 创建 inotify_fd
	fd, errno := unix.InotifyInit1(unix.IN_CLOEXEC)
	if fd == -1 {
		return nil, errno
	}
	// 创建 epoll 
	poller, err := newFdPoller(fd)
	if err != nil {
		unix.Close(fd)
		return nil, err
	}
	w := &Watcher{
		fd:       fd,
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


```cgo
func newFdPoller(fd int) (*fdPoller, error) {
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

	// 创建 epoll_fd
	poller.epfd, errno = unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if poller.epfd == -1 {
		return nil, errno
	}
	// 创建 pip_fd, 其中 pipe[0] 是读取, pipe[1]是写入. O_NONBLOCK 表示非阻塞
	errno = unix.Pipe2(poller.pipe[:], unix.O_NONBLOCK|unix.O_CLOEXEC)
	if errno != nil {
		return nil, errno
	}

	// 注册 inotify_fd "读" 的事件到 epoll_fd
	event := unix.EpollEvent{
		Fd:     int32(poller.fd),
		Events: unix.EPOLLIN,
	}
	errno = unix.EpollCtl(poller.epfd, unix.EPOLL_CTL_ADD, poller.fd, &event)
	

	// 注册 pip_fd "读" 的事件到 epoll_fd
	event = unix.EpollEvent{
		Fd:     int32(poller.pipe[0]),
		Events: unix.EPOLLIN,
	}
	errno = unix.EpollCtl(poller.epfd, unix.EPOLL_CTL_ADD, poller.pipe[0], &event)

	return poller, nil
}
```

- 监控事件

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
        
        // epoll 的 SYS_EPOLL_WAIT 系统调用, 判断是否有读取事件就绪
		ok, errno = w.poller.wait()
		if errno != nil {
		
		    // Send Error
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
        
        // inotify_fd 就绪, 读取就绪的事件 InotifyEvent 
		n, errno = unix.Read(w.fd, buf[:])
		if errno == unix.EINTR {
			continue
		}
        
        // 再次判断, 双保险
		if w.isClosed() {
			return
		}
        
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

- epoll wait

```cgo
func (poller *fdPoller) wait() (bool, error) {
	// 2个fd, fd监听 EPOLLIN, 同时 EPOLLHUP, EPOLLERR 是无须监听, 任何时候都会发生
	// 取最大值 2*3+1
	events := make([]unix.EpollEvent, 7)
	for {
		n, errno := unix.EpollWait(poller.epfd, events, -1) // n是就绪的事件个数
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
		
		// 就绪的事件的判断
		ready := events[:n]
		epollhup := false
		epollerr := false
		epollin := false
		for _, event := range ready {
			if event.Fd == int32(poller.fd) {
				if event.Events&unix.EPOLLHUP != 0 {
					// This should not happen, but if it does, treat it as a wakeup.
					epollhup = true
				}
				if event.Events&unix.EPOLLERR != 0 {
					// If an error is waiting on the file descriptor, we should pretend
					// something is ready to read, and let unix.Read pick up the error.
					epollerr = true
				}
				if event.Events&unix.EPOLLIN != 0 {
					// There is data to read.
					epollin = true
				}
			}
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
					// This is a regular wakeup, so we have to clear the buffer.
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

- 添加监控的目录/文件

```cgo
func (w *Watcher) Add(name string) error {
	name = filepath.Clean(name)
	if w.isClosed() {
		return errors.New("inotify instance already closed")
	}
    
    // 文件/目录需要通知的事件
	const agnosticEvents = unix.IN_MOVED_TO | unix.IN_MOVED_FROM |
		unix.IN_CREATE | unix.IN_ATTRIB | unix.IN_MODIFY |
		unix.IN_MOVE_SELF | unix.IN_DELETE | unix.IN_DELETE_SELF

	var flags uint32 = agnosticEvents

	w.mu.Lock()
	defer w.mu.Unlock()
	watchEntry := w.watches[name]
	if watchEntry != nil {
		flags |= watchEntry.flags | unix.IN_MASK_ADD
	}
	wd, errno := unix.InotifyAddWatch(w.fd, name, flags) // 添加
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