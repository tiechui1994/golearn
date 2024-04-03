# runtime mutex

Go 的 runtime 中封装了 mutex, 这个 mutex 在 channel, netpoll, 检测活跃的定时器等使用.

sync.Mutex vs runtime.mutex

sync.Mutex 是用户层的锁, Lock 失败的 goroutine 被阻塞(调用 gopark).

runtime.mutex 是 runtime 使用的锁, Lock 抢锁失败, 会造成 m 阻塞(线程阻塞, 底层调用 futex)

### 基础内容

futex(Fast Userspace Mutexes), Linux 下实现锁定和构建高级抽象锁如信号量和POSIX互斥的基本工具.

futex 由一块能够被多个进程共享的内存空间(一个对齐后的整型变量)组成; 这个整型变量的值能够通过汇编语言调用 CPU 提供的
原子操作指令来增加或减少, 并且一个进程可以等待直到那个值变成正数. futex 的操作几乎是在用户空间完成, 只有当操作结果不
一致需要仲裁时, 才需要进入操作系统内核空间执行. 

futex 的基本思想是竞争态总是很少发生的, 只有在竞争态才需要进入内核, 否则在用户态即可完成.


```cgo
/*
 * val: 0, unlock
 * val: 1, lock, no waiters
 * val: 2, one or more waiters
 */

int val = 0;

void lock() {
    int c;
    if ((c = cmpxh
}
```


