## golang io

- io.CopyBuffer, io.Copy, io.CopyN

```
io.CopyBuffer() // 使用用户提供的 buffer 作为缓冲区
io.Copy()       // 没有缓冲区, 每次调用都会创建 32KB 的缓冲区
io.CopyN()      // 只拷贝N个字节. 创建的缓冲区大小为 min(max(N,1), 32K)
```

> 相同点: 底层的实现都是 io.copyBuffer()

`io.copyBuffer()` 的实现逻辑:

1. 如果 Reader 实现了 WriterTo() 接口, 则直接调用 WriterTo() 接口将数据写入到 writer
   如果 Writer 实现了 ReadFrom() 接口, 则直接调用 ReadFrom() 接口从 reader 当中读取数据

2. 数据缓冲区创建

3. for 循环读取(Reader)和写入(Writer)操作. 在 read 出现错误(非EOF), write 出现错误, 或者 读取与写入长度不一致的
状况下, 循环都会结束返回.


- io.ReadFull, io.ReadAtLeast 

```
io.ReadFull(r Reader, buf []byte) // 读取 buf 长度个字节的内容
io.ReadAtLeast(r Reader, buf []byte, min int) // 读取至少 mim 个字节的内容. 必须满足 len(buf) >= min
```

- io.Pipe

`io.Pipe` 创建一个同步的内存管道. 该管道可用于将期望io.Reader的代码与期望io.Writer的代码连接.

实现思路: 使用一个`chan []byte`作为数据管道(可以读取和写入), 由于每次往管道当中写入的数据大小不一, 每次从管道当中读取
的数据大小也不一. 为了平衡这两个不一样, 在写入端(Write)采用了序列化的方式, 保证并发状况下写入是有序的, 并且读取的顺序和
写入的顺序也是一致的, 但是不能保证每个读取端能读取到完整的数据, 因为有可能存在多个读取端(所有读取端读到的数据加起来就是写
入的数据). 

那么如何保证写入的顺序和读取的顺序是一致的呢? 

首先, 在每次写入数据的时候, 加锁处理, 保证任何时刻, 只能有一个写入端. 

其次, 写入端往`chan []byte`写入数据以后, 会从`chan int`当中读取读取端已经读取的字节数, 如果读取的数据总长度等于写入
的长度, 那么此次写入就结束了, 否则, 再次将剩余数据再次写入到`chan []byte`当中.

最后, 读取端, 每次获取不超过读取缓存区(Read()的参数b)大小的数据, 从`chan []byte`当中获取数据, 然后将数据拷贝到读取
缓冲区当中, 然后向`chan int`写入当前已经读取的数据量, 反馈给写入端, 然后返回.

通过上述的三步, 形成了一个反馈系统, 这样保证写入数据之后, 立马就可以被读取.

在所有的数据写入和读取完成之后, 此时


```cgo
type pipe struct {
	wrMu sync.Mutex // 序列化写入操作
	wrCh chan []byte // 管道数据
	rdCh chan int // 读取管道

	once sync.Once // 保护 done, 只能调用一次
	done chan struct{}
	rerr atomicError 
	werr atomicError
}
```

Write: 序列化向`chan wrCh`当中写入内容. 使用`sync.Mutex`加锁是保证序列化的.

```cgo
func (p *pipe) Write(b []byte) (n int, err error) {
	select {
	case <-p.done:
		return 0, p.writeCloseError()
	default:
		p.wrMu.Lock()
		defer p.wrMu.Unlock()
	}

	for once := true; once || len(b) > 0; once = false {
		select {
		case p.wrCh <- b:
			nw := <-p.rdCh
			b = b[nw:]
			n += nw
		case <-p.done:
			return n, p.writeCloseError()
		}
	}
	return n, nil
}
```


Read: 从管道当中读取.

```cgo
func (p *pipe) Read(b []byte) (n int, err error) {
	select {
	case <-p.done:
		return 0, p.readCloseError()
	default:
	}

	select {
	case bw := <-p.wrCh:
		nr := copy(b, bw)
		p.rdCh <- nr
		return nr, nil
	case <-p.done:
		return 0, p.readCloseError()
	}
}
```

> 使用注意:
>
> 1. reader 和 writer 必须在不同的 goroutine 当中使用. 否则, 会产生死锁状况.
> 2. 为了保证数据安全性, 当数据写完之后, writer/reader 需要关闭.


应用:

1.从JSON到HTTP请求. 

将某些数据编码为 JSON, 并希望通过 http.Post 将其发送到 Web 端点. JSON的Encode采用的是io.Writer, 而http请求使用
的io.Reader作为输入. 这样刚好可以使用 pipe 将其连接在一起.


## golang tcp

首先介绍下 golang 系统调用接口:

```cgo
func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)

func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)

func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)

func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)
```

trap 是 go 定义的系统调用类型. a1~a6是系统调用参数, r1,r2 是系统调用返回值. err是系统调用错误码.

最常见的 trap:

```
# read, write, open, close
SYS_READ                   = 0
SYS_WRITE                  = 1
SYS_OPEN                   = 2
SYS_CLOSE                  = 3

# select
SYS_SELECT                 = 23

# poll
SYS_POLL                   = 7

# epoll
SYS_EPOLL_CREATE           = 213
SYS_EPOLL_WAIT             = 232
SYS_EPOLL_CTL              = 233

# socket, connect, accept, bind, listen, sendto, recvfrom
SYS_SOCKET                 = 41
SYS_CONNECT                = 42
SYS_ACCEPT                 = 43
SYS_SENDTO                 = 44
SYS_RECVFROM               = 45
SYS_SENDMSG                = 46
SYS_RECVMSG                = 47
SYS_SHUTDOWN               = 48
SYS_BIND                   = 49
SYS_LISTEN                 = 50
SYS_GETSOCKNAME            = 51
SYS_GETPEERNAME            = 52
SYS_SETSOCKOPT             = 54
SYS_GETSOCKOPT             = 55

# mmap
SYS_MMAP                   = 9

# sendfile
SYS_SENDFILE               = 40

# splice
SYS_SPLICE                 = 275
```


首先介绍两个系统调用函数: `sendfile`和`splice`

```cgo
#include <sys/sendfile.h>
ssize_t sendfile(int out_fd, int in_fd, off_t *offset, size_t count);
```

`in_fd`是代表输入文件的文件描述符, `out_fd`是代表输出文件的文件描述符. 

> out_id 必须为 socket (linux 2.6.33 开始可以是任何文件).
> in_fd 指向的文件必须为可以进行 mmap() 操作的, 通常为普通文件.

```cgo
#define _GNU_SOURCE 
#include <fcntl.h>
ssize_t splice(int fd_in, loff_t *off_in, int fd_out, loff_t *off_out, 
               size_t len, unsigned int flags);
```



