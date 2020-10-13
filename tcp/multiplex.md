## IO 复用

[深入理解IO复用之epoll](http://r6d.cn/sLAp)

[深入分析select&poll&epoll原理](http://r6d.cn/sLAF)

### Linux 内核事件机制

在 Linux 内核中存在着等待队列的数据结构, 该数据结构是基于双端链表实现. Linux内核通过将阻塞的进程任务添加到等待队列中, 
而进程任务被唤醒则是在队列轮训遍历检测是否处于就绪状态, 如果是那么在等待队列中删除等待节点并通过节点上的回调函数进行通知,
然后加入到CPU就绪队列中, 等待CPU调度执行.

其具体流程包含 "休眠逻辑" 和 "唤醒逻辑":

### select 

- select 技术的实现也是基于Linux内核的等待与唤醒机制诗选.

- 在Linux中基于POSIX协议定义的select技术最大可支持的描述符个数为1024个, 显然对于互联网的高并发应用是远远不够的. 虽然
现代操作系统支持更多的描述符, 但是对于select技术增加描述符的话, 需要更改POSIX协议这描述符个数的定义, 但是此时需要重新
编译内核, 不切实际.

- 另外一个是用户进程调用select的时候需要将一整个fd集合的大块内存从用户拷贝到内核中, 期间用户空间与内核空间来回切换开销比
较大, 再加上调用select的频率本身非常平凡, 这样导致高频率调用且大内存数据的拷贝, 严重影响性能.

- 最后唤醒逻辑的处理, select技术在等待过程中如果监控到至少有一个socket事件是可读的时候将会唤醒整个等待队列, 告知当前等
待队列中存在就绪事件的socket, 但是具体是哪个socket不知道, 必须通过轮训的方式逐个遍历进行回调通知, 也就是唤醒轮训节点包
含了就绪和等待通知的socket事件, 如果每次只有一个socket事件可读, 那么每次轮训遍历的事件复杂度是O(n), 影响性能.


### poll

poll技术使用链表结构的方式来存储fdset的集合, 相比select而言, 链表不受限于FD_SIZE的个数限制, 但是对于select存在的性能
并没有解决. 即一个存在大内存数据拷贝的问题, 一个是轮训遍历整个等待队列的每个节点并逐个通过回调函数来实现读取任务的唤醒.


### epoll

为了解决 select&poll 技术存在的两个性能问题, 对于大内存数据拷贝的问题, epoll 通过 epoll_create 函数创建 epoll 空
间(相当于一个容器管理), 在内核中只存储一份数据来维护N个socket事件的变化. 

通过 epoll_ctl 函数来实现对 socket 事件的增删改操作, 保证用户空间与内核空间对内核是具备可见性, 直接通过指针引用的方式
进行操作, 避免大内存数据的拷贝导致的空间切换性能问题. 

对于轮训等待事件通过 epoll_wait 的方式来实现对 socket 事假的监听, 将不断轮训等待高频事假wait与低频socket注册两个操作
分开, 同时会对监听就绪的 socket 事件添加到就绪队列中, 也就保证唤醒轮训的事件都是具备可读的.


```cgo
// 创建保存 epoll 文件描述符的空间
int epoll_create(int size); // 使用链表
int epoll_create1(int flag); // 使用红黑树


long epoll_ctl(int epfd,                    // epoll fd 
               int op,                      // 操作, EPOLL_CTL_ADD|EPOLL_CTL_MOD|EPOLL_CTL_DEL
               int fd,                      // 注册 fd
               struct epoll_event *event);  // epoll 监听事件

struct epoll_event {
    __poll_t events;
    _u64 data;
}


int epoll_wait(int epfd,                    // epoll fd
               struct epoll_event *events,  // epoll监听的事件
               int maxevents,               // epoll可以保存的最大事件数量
               int timeout);                // 超时时间
```
