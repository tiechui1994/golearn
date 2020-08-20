## API 限流

限流算法: 信号量, 漏桶算法 和 令牌桶算法.


### 信号量

信号量两个重要算法 `Acquire()` 和 `Release()`, 通过  `Acquire()` 获取许可, 该方法会阻塞, 直到获取许可为止. 
通过 `Release()` 方法释放许可.

阻塞方式实现: 使用 `channel`, `channel` 的大小就是信号量数值. 

获取信号量, 本质上是向 `channel` 当中放入元素, 如果多个协程并发获取信号量, 则 `channel` 会 `full` 从而阻塞, 从而
达到控制并发协程的目的. 

释放信号量, 本质上是从 `channel`当中获取元素, 由于有 `Acquire()` 的放入元素, 因此此处一定能获取到元素, 即能释放成
功, 注意: 处于编程安全, 需要添加 `default` 部分.


非阻塞方式: 使用并发安全计数方式, 比如原子 `atomic` 加减操作.

---

### 令牌桶算法 和 漏桶算法

`令牌桶算法` 和 `漏桶算法` 是相反的, 一个是 *进水*, 一个是 *漏水*.

- 漏桶(Leaky Bucket)

思路: 水(请求)先进入到漏桶里, 漏桶以一定的速度出水(接口有响应速率), 当水流入速度过大, 则会直接溢出(访问频率超过接口响应
速率), 然后就拒绝请求. 

![image](/images/develop_limit_bucket.jpeg)

这里重要的两个变量: 一个是桶的大小, 支持流量突发增多时可以存多少水(burst), 另一个是水桶漏洞的大小(rate).

漏桶算法可以使用 `redis` 队列实现, 生产者发送消息前检查 *队列* 长度是否超过阈值, 超过阈值则丢弃消息, 否则发送消息到
`redis` 队列中; 消费者以固定速率从 `redis` 队列中获消息. `redis` 队列在这里起到一个缓存池的作用.

漏桶算法:

```cgo
func (sp *servicePanel) incLimit() error {
    if sp.currentLimitCount.Load() > sp.currentLimitFunc(nil) {
        return ErrCurrentLimit
    }
    
    sp.currentLimitCount.Inc()
    
    return nil
}

func (sp *servicePanel) clearLimit() {
    // 定期每秒重置计数器, 从而达到每秒限制的并发数
    t := time.NewTicker(time.Second)
    for {
        select {
            case <-t.C:
                sp.currentLimitCount.Store(0)
        }
    }
}
```

改进:

```cgo
// 严格按照每个请求按照某个固定数值进行, 改进时间的粗粒度
func (sp *servicePanel) incLimit() error {
    if sp.currentLimitCount.Load() > 1 {
        return ErrCurrentLimit
    }
    
    sp.currentLimitCount.Inc()
    
    return nil
}

func (sp *servicePanel) clearLimit() {
    t := time.NewTicker(time.Second/time.Duration(sp.currentLimitFull(nil)))
    for {
        select {
            case <-t.C:
                sp.currentLimitCount.Store(0)
        }
    }
}
```

uber.go 实现的思路:

1. 创建令牌桶的时候, 计算两个连续令牌产生的时间间隔 `perRequest`, 同时产生一个松弛度 `maxSlack`, 其值为 *`-10*perRequest`*
松弛度含义是, 当请求的最大瞬时速率不能超过令牌桶的10倍. 还有就是存储当前的一个状态 state, 保存了最近一次获取令牌的时间
和上一次获取令牌需要等待的时间. 

2. 获取令牌桶, 先加载令牌桶的状态, 即上一次获取令牌的时间. 如果是第一次获取令牌, 直接允许 (CMP), 否则需要计算公式:
`newState.sleepFor += t.perRequest - now.Sub(oldState.last)`

- 如果 newState.sleepFor <= 0, 当 newState.sleepFor < maxSlack, 说明上一次令牌离当前时间很久, 则 
newState.sleepFor 设定为 maxSlack, 即接下来可以瞬时承担 10 倍于令牌桶的速率. 在当前状况下, 令牌可以立马返回, 不需
要等待.

- 如果 newState.sleepFor > 0, 则需要等待 newState.sleepFor 的时间, 然后才可以返回令牌.

> 注: 因为是并发请求, 需要先更新令牌桶的 state, 然后进行 sleep 等待令牌.
