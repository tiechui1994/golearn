## IO 复用

[深入理解IO复用之epoll](http://r6d.cn/sLAp)

[深入分析select&poll&epoll原理](http://r6d.cn/sLAF)

### Linux 进程的睡眠和唤醒

**Linux 进程的睡眠和唤醒**

> 进程调度(主动+被动), 即进程获得CPU, 执行代码 

在Linux中, 仅等待CPU时间的进程称为就绪进程, 它们被设置在一个运行队列中, 一个就绪进程的状态标记为 TASK_RUNNING. 一旦
一个运行中的进程时间片用完, Linux内核的调度器会剥夺这个进程对CPU的控制权, 并且从运行队列中选择一个合适的进程投入运行.


一个进程也可以主动释放CPU控制权. 函数 schedule() 是一个调度函数, 它可以被一个进程主动调用, 从而调度其他进程占用CPU. 
一旦这个主动放弃CPU的进程被再次被调度, 那么它将从上次停止执行的位置开始执行, 即它将从调用 schedule() 的下一行代码处开
始执行.

> 睡眠, 及其状态标志位

有的时候, 进程需要等待直到某个特定的事件发生, 例如设备初始化完成, I/O操作完成或定时器到时等. 在这种情况下, 进程则必须从
运行队列移除, 加入到一个等待队列中, 这个时候进程就进入了睡眠状态.

Linux中的进程睡眠状态有两种: 一种是可中断的睡眠状态, 其状态标志位为 TASK_INTERRUPTIBLE, 另一种是不可中断的睡眠状态,
其状态标志位为 TASK_UNINTERRUPTIBLE. 可中断的睡眠状态的进程会睡眠直到"某个条件为真", 比如,产生一个硬件中断, 释放进程
正在等待的系统资源或传递一个信号都可以是唤醒进程的条件. 不可中断睡眠状态与可中断睡眠状态类似, 但是它有一个例外, 那就是把
信号传递到这种睡眠状态的进程不能改变它的状态, 也就说它不响应信号的唤醒. 不可中断睡眠状态一般很少用到, 但在一些特定情况下
这种状态还是很有用的, 比如说: 进程必须等待, 不能被中断, 直到某个特定的事件发生.


`现代的 Linux 操作系统中, 进程一般是调用 schedule() 的方法自身会进入睡眠状态. [需要提前设置好相应的状态标记位]` 

example:

```cgo
sleeping_task = current;
set_current_state(TASK_INTERRUPTIBLE);
schedule();

...
```

> 首先保存一份进程结构指针 sleeping_task. current是一个宏, 指向正在指向的进程结构.
> set_current_state() 设置当前进程的状态位 TASk_INTERRUPTIBLE(注意, 此时的进程的状态还是 TASK_RUNNING)
> 如果 schedule() 是被一个状态为 TASK_RUNNING 的进程调度, 那么 schedule() 将调度另外一个进程占用CPU; 
> 如果 schedule() 是被一个状态为 TASk_INTERRUPTIBLE 或 TASk_UNINTERRUPTIBLE 的进程调度, 那么还有一个附加的步骤
> 需要被执行: 当前执行的进程在另外一个进程被调度之前会被从运行队列中移除, 这将导致正在运行的那个进程进入睡眠, 因为它已经
> 不在运行队列中了.


```cgo
wake_up_process(sleeping_task);
```

在调用 wake_up_process() 以后, 这个睡眠进程的状态会被设置为 TASK_RUNNING, 而且调度器会把它加入到运行队列中去. 

---

**无效唤醒**

几乎在所有的情况下, 进程都会在检查了某些条件之后, 发现条件不满足才进入睡眠. 可是有的时候进程却会在判断条件为真后开始睡眠,
如果这样的话, 进程就会无限期地休眠下去, 这就是所谓的无效唤醒问题. 

在操作系统中, 当多个进程都企图对共享数据进行某种处理, 而最后的结果又取决于运行的顺序时, 就会发生竞争条件, 这是操作系统中
一个典型的问题, 无效唤醒恰恰就是由于竞争条件导致的.


设置两个进程A, B. A正在处理一个链表, 它需要检查这个链表是否为空, 如果非空就对链表里面的数据进行一些操作, 同时B也在往这个
链表添加节点. 当这个链表为空的时候, 由于无数据可操作, A进入睡眠. 当B进程向链表里面添加了节点之后就唤醒A进程:

```cgo
A:

spin_lock(&list_lock);
if (list_empty(&list_head)) {
    spin_unlock(&list_lock);
    set_current_state(TASK_INTERRUPTIBLE);
    schedule();
    spin_lock(&list_lock);
}

... 

spin_unlock(&list_lock);


B:

spin_lock(&list_lock);
list_add_tail(&list_head, new_node);
spin_unlock(&list_lock);
wake_up_process(process_a_task);
```

> 假设 A 进程执行到 `spin_unlock(&list_lock);` 之后, `set_current_state(TASk_INTERRUPTIBLE);` 之前的时候,
B进程被另外一个处理器调度投入运行. 在这个时间片内, B进程执行完了它所有的指令, 试图唤醒A进程, 而此时的A进程还没有进入睡眠,
所以唤醒操作无效. 在这之后, A进程继续执行, 它会错误的认为这个时候链表仍然是空的, 于是将自己的状态设置为 TASK_INTERRUPTIBLE
然后调用 schedule() 进入睡眠. 由于错过了B进程的唤醒, 它将无限期的睡眠下去, 这就是无效唤醒问题. 

---

**避免无效唤醒**

```cgo
A:

set_current_state(TASK_INTERRUPTIBLE); // 设置为睡眠状态
spin_lock(&list_lock);
if (list_empty(&list_head)) {
    spin_unlock(&list_lock);
    schedule(); // 当前进程休眠, 调度其他进程执行. 程序一旦唤醒, 则会从后面的部分开始执行
    spin_lock(&list_lock);
}

set_current_state(TASK_RUNNING); // 设置为运行状态

.... 

spin_unlock(&list_lock);
```

**Linux 内核进程睡眠**

```cgo
DECLARE_WAITQUEUE(wait, current); // 创建一个等待队列的项
add_wait_queue(q, &wait); // 将wait加入到等待队列q当中
set_current_state(TASK_INTERRUPTIBLE); // 设置为睡眠状态

// condition 是等待条件
while (!condition) {
    schedule(); // 当前进程休眠, 调度其他进程执行. 程序一旦唤醒, 则会从后面的部分开始执行
    set_current_state(TASK_INTERRUPTIBLE);
}

set_current_state(TASK_RUNNING); // 设置为运行状态
remove_wait_queue(q, &wait); // 从等待队列q当中移除wait
```


### Linux 内核事件机制

在 Linux 内核中存在着等待队列的数据结构, 该数据结构是基于双端链表实现. Linux内核通过将阻塞的进程任务添加到等待队列中, 
而进程任务被唤醒则是在队列轮询遍历检测是否处于就绪状态, 如果是那么在等待队列中删除等待节点并通过节点上的回调函数进行通知,
然后加入到CPU就绪队列中, 等待CPU调度执行.

其具体流程包含 "休眠逻辑" 和 "唤醒逻辑":

- 休眠(进程将自己标记为休眠状态)

> 休眠在Linux中有两种状态, 一种会忽略信息, 一种则会在收到信号的时候被唤醒并响应. 不过这两种状态的进程是处于同一个等待队
列上的.

1.在Linux内核中某一个进程任务task执行需要等待某个条件condition被触发执行之前, 首先会在内核中创建一个等待节点 entry, 
然后初始化entry相关的属性信息, 其中将进程任务存放在entry节点并同时存储一个wake_callback函数并挂起当前进程.

2.其次不断轮询检查当前进程任务task执行的condition是否满足,如果不满足则调用schedule()进入休眠状态.
> schedule(), 选择和执行一个其他进程.

3.最后如果满足condition, 就会将entry从队列当中移除, 也就是说这个时候事件已经被唤醒, 进程处于就绪状态.

- 唤醒(进程被设置为可执行状态)

1.在等待队列中循环遍历所有的entry节点, 并执行回调函数, 直到当前 entry 为排他节点是时候退出循环遍历.

2.执行的回调函数中, 存在私有逻辑和公用逻辑, 类似模板方法设计模式

3.对于default_wake_function的唤醒回调函数主要是将entry的进程任务task添加到cpu就绪队列中等待cpu调度执行任务task


### select 

select 的关键在于 fd_set. fd_set 的每一个 bit 可以对应一个文件描述符, 也就是说 1 byte 长的 fd_set 最大可以对应 8 
个 fd. 

```
int select(int maxfdp, fd_set *rfds, fd_set *wfds, fd_set *efds, struct timeval *timeout);
```

在调用 select 的过程当中, fd_set 既是参数(告知内核自己所关心的 fd 集合), 也是返回值(内核响应用户准备就绪的 fd 集合).
因为上述的特点, 将 fd 加入 select 监控集合时, 还要再使用一个 array 保存放到 select 监控集中的 fd, 一是用于在 select
返回后, array 作为源数据和 fd_set 进行 FD_ISSET 判断, 二是 select 返回后要将以前加入的但并未就绪的 fd 清空.


- select 技术的实现也是基于Linux内核的等待与唤醒机制.

- 在Linux中基于POSIX协议定义的select技术最大可支持的描述符个数为1024个, 显然对于互联网的高并发应用是远远不够的. 虽然
现代操作系统支持更多的描述符, 但是对于select技术增加描述符的话, 需要更改POSIX协议这描述符个数的定义, 但是此时需要重新
编译内核, 不切实际.

- 用户进程调用 select 的时候需要 "将整个 fd 集合的大块内存从用户拷贝到内核中", 在这期间用户空间与内核空间来回切换开销比
较大, 再加上 "调用select的频率" 本身非常平凡, 这样导致高频率调用且大内存数据的拷贝, 严重影响性能.

- 最后唤醒逻辑的处理, select技术在等待过程中如果监控到至少有一个socket事件是可读的时候将会唤醒整个等待队列, 告知当前等
待队列中存在就绪事件的socket, 但是具体是哪个socket不知道, 必须通过 "轮询" 的方式逐个遍历进行回调通知, 也就是轮询唤醒节
点包含了就绪和等待通知的socket事件, 如果每次只有一个socket事件可读, 那么每次轮询遍历的事件复杂度是O(n), 影响性能.


### poll

poll 技术使用链表结构的方式来存储 fdset 的集合, 相比 select 而言, 链表不受限于FD_SIZE的个数限制, 但是对于select存
在的性能并没有解决. 即一个存在大内存数据拷贝的问题, 一个是轮询遍历整个等待队列的每个节点并逐个通过回调函数来实现读取任务的
唤醒.

```
struct pollfd {
  int fd;         /* File descriptor to poll. */
  short events;   /* Types of events poller cares about. */
  short revents;  /* Types of events that actually occurred. */
};
```

### epoll

为了解决 select&poll 技术存在的两个性能问题, 对于大内存数据拷贝的问题, epoll 通过 epoll_create 函数创建 epoll 空
间(相当于一个容器管理), 在内核中只存储一份数据来维护N个socket事件的变化. 

通过 epoll_ctl 函数来实现对 socket 事件的增删改操作, 保证用户空间与内核空间对内核是具备可见性, 直接通过指针引用的方式
进行操作, 避免大内存数据的拷贝导致的空间切换性能问题. 

对于轮询等待事件通过 epoll_wait 的方式来实现对 socket 事假的监听, 将不断轮询等待高频事假wait与低频socket注册两个操作
分开, 同时会对监听就绪的 socket 事件添加到就绪队列中, 也就保证唤醒轮询的事件都是具备可读的.


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


epoll 工作流程:

![image](/images/net_epoll.png)

epoll 特点:

1. 与 select 相比, epoll 分清了高频率调用和低频调用. epoll_ctl 相对来说是非高频调用的, 而 epoll_wait 则是会被高频
调用的. epoll 利用 epoll_ctl 来插入或删除一个 fd, 实现用户态到内核态的数据拷贝, 确保了每一个 fd 在其生命周期只需要被
拷贝一次, 而不是每次调用 epoll_wait 的时候拷贝一次. epoll_wait 则被设计成几乎没有入参的调用.

2. 在实现上 epoll 采用红黑树存储监听的 fd, 而红黑树本身的插入和删除性能比较稳定. 使用红黑树, 保证了 fd 不会被重复添加.
fd 添加时, 会将 fd 与相应的设备驱动程序建立回调关系, 也就是在内核中断处理程序为它注册一个回调函数, 在 fd 相应的事件触发
(中断)之后, 内核就会调用这个回调函数.

3. 相比 select 调用时会将全部监听的 fd 从用户空间拷贝到内核空间并线性扫描一遍后找到就绪的 fd 再返回到用户空间. epoll_wait
则直接返回已就绪的 fd.



