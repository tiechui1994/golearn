## Go context 包

什么是 context ?

context, 中文翻译为 "上下文", 准确的说它是 goroutine 的上下文, 包含 goroutine 的运行状态, 环境, 现场等信息.

context 主要用来在 goroutine 之间传递上下文信息, 包括: "取消信号", "超时时间", "截止时间", "k-v" 等.

context 包的引入, 标准库很多接口因此都加上了 context 参数, 例如 database/sql 包. context 几乎成为了并发控制和超
时控制的标准做法.

> context.Context 类型的值可以协调多个 goroutine 中的代码执行 "取消" 操作, 并且可以存储键值对. 最为重要的是它是并
发安全的.

为什么需要 context ?

先看一个例子吧.

![image](/images/develop_context_nouse.png)

一个请求当中启动了若干个 goroutine, 这些 goroutine 去做一些不同的工作: 数据库查询, 调用下游接口等等. 现在要对这些
goroutine 进行状态同步控制, 比如当其中有一个 goroutine 操作失败了, 需要通知其下游的 goroutine 也都取消操作, 快速返
回, 以减少资源占用. 

上述的案例该怎么实现呢? 如果 goroutine 调用链只有两层, 那么可以创建一个无缓存 channel, 然后使用 channel 进行同步控
制, 但是当调用链拉长以后, 简单的使用 channel 去处理, 代码的逻辑将变得异常复杂. 这个时候就需要一种通用的技术, 来在协程
之间进行状态的同步. 这就是 context 产生的原因.

上述例子, 使用了 context 之后, 调用链是这样的:

![image](/images/develop_context_use.png)

看起来是不是挺简单的.


### Contex 接口

**Context** 是一个接口, 它定义了 context 协议所支持的行为. 

```cgo
type Context interface {
    // 当 context 被取消或者到达deadline, 返回一个被关闭的 chan
    Done() <-chan struct{}
    
    // 在 chan Done 被关闭后, 返回 context 被取消的愿意
    Err() error
    
    // 返回 context 是否会被取消, 以及自动取消的时间
    Deadline() (deadline time.Time, ok bool)
    
    // 获取 key 对应的 value
    Value(key interface{}) interface{}
}
```

`Context` 定义了 4 个方法, 上面的方法可以多次调用.

- `Done()` 返回一个 chan, 可以表示 context 被取消的信号: 当这个 chan 被关闭时, 说明 context 被取消了. 注意, 这 
是一个只读的 chan, 读一个关闭的 chan 会读出相应类型的零值. 关闭 chan 的操作是由 `取消`(用户) 或者 `超时`(系统) 触发
的, 使用者不能直接区操作这个 chan. 当子协程从 chan 当中读取到值后, 就可以做一些收尾工作, 尽快退出.

- `Err()` 返回一个 error, 表示 chan 被关闭的原因. 例如: `取消`, `超时`. 在 chan 未被关闭之前返回结果都是 nil.

- `Deadline()` 返回 context 的截止时间, 通过这个时间, 函数可以决定是否进行接下来的操作, 当时间太短, 就可以不往下做了,
以免浪费资源. 

- `Value()` 获取父协程透传的值.


**canceler** 是一个专注 `取消` 操作的接口:

```cgo
type canceler interface {
    cancel(removeFromParent bool, err error)
    Done() <-chan struct{}
}
```

实现了 `canceler` 接口的 Context, 就表面该 Context 是可取消的(需要同时实现Context接口和canceler接口). 在 go 的
`context` 包当中 `cancelCtx` 和 `timerCtx` 就是可取消的 Context.

Context 接口这样设计的原因:

- "取消" 操作应该是建议性, 而非强制性. caller 不应该去关心, 干涉 callee 的情况, 决定如何以及何时 return 是 callee
的责任. caller 只需要发送 "取消" 消息, callee 根据收到的消息来做进一步的决策, 因此并没有定义 Cancel 方法.

- "取消" 操作应该可传递. "取消" 某个函数时, 和它相关联的其他函数也应该 "取消". 因此 `Done()` 方法返回一个只读的 chan,
所有相关函数监听此 chan. 一旦 chan 关闭, 通过 chan 的 "广播机制", 所有监听者都能收到.

