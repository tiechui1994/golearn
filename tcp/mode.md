## Linux IO

Linux 下主要的 5 种IO模型:

- 阻塞I/O (blocking I/O)
- 非阻塞I/O (nonblocking I/O)
- I/O 复用 (I/O multiplexing)
- 信号驱动I/O (signal driven I/O (SIGIO))
- 异步I/O (asynchronous I/O)

#### 阻塞I/O 

进程一直阻塞, 直到数据拷贝完成. 应用程序一个 IO 函数, 导致应用程序阻塞, 等待数据准备好. 数据准备好以后(从设备拷贝到内核
空间), 从内核空间拷贝到用户空间, IO函数返回成功

![image](/images/linux_io_block.png)

#### 非阻塞I/O

通过进程反复调用 IO 函数, 在数据拷贝过程中, 进程是阻塞的.

![image](/images/linux_io_nonblock.png)'


#### I/O 复用

主要是 epoll 和 select. 一个线程可以对多个IO端口进行监听, 当 socket 有读写事件时分发到具体的线程进行处理.

![image](/images/linux_io_multiplex.png)


#### 信号驱动I/O

首先允许 socket 进行信号驱动 IO, 并设置一个信号处理函数, 进程继续运行并不阻塞. 当数据准备好时, 进程会收到一个 SIGIO
信号, 可以在信号处理函数中调用 IO 操作函数处理数据. (类似回调函数)

![image](/images/linux_io_sigle.png)

#### 异步I/O

异步IO不是顺序指向. 用户进程进行 aio_read 系统调用之后,无论内核数据是否准备好, 都会直接返回给用户进程. 然后用户态进程
可以去做别的事情. 等到 socket 数据准备好了, 内核直接拷贝数据到用户空空, 然后从内核向进程发送通知. IO两个阶段, 进程都是
非阻塞的.

![image](/images/linux_io_async.png)

#### 区别

![image](/images/linux_io_diff.png)

