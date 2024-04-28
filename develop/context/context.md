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

实现了 `canceler` 接口的 Context, 就表明该 Context 是可取消的(即同时实现Context接口和canceler接口). 在 go 的
`context` 包当中 `cancelCtx` 和 `timerCtx` 就是可取消的 Context.

Context 接口这样设计的原因:

- "取消" 操作应该是建议性, 而非强制性. caller 不应该去关心, 干涉 callee 的情况, 决定如何以及何时 return 是 callee
的责任. caller 只需要发送 "取消" 消息, callee 根据收到的消息来做进一步的决策, 因此并没有定义 Cancel 方法.

- "取消" 操作应该可传递. "取消" 某个函数时, 和它相关联的其他函数也应该 "取消". 因此 `Done()` 方法返回一个只读的 chan,
所有相关函数监听此 chan. 一旦 chan 关闭, 通过 chan 的 "广播机制", 所有监听者都能收到.


### 实现 Context 接口的类型

在 context 包当中提供了一些实现 `Context` 接口的类型. 它们分别是 `emptyCtx`, `cancelCtx`, `timerCtx`, `valueCtx`.

#### emptyCtx

`emptyCtx` 顾名思义, 就是一个空的 context, **永远不会被 cancel, 没有存储值, 也没有deadline**. 

它被包装成:

```cgo
var (
    background = new(emptyCtx)
    todo       = new(emptyCtx)
)

func Background() Context {
    return background
}

func TODO() Context {
    return todo
}
```

background 通常用在 main 函数当中, 作为所有 context 的根节点.

todo 通常用在并不知道传递什么 context 的情形. 例如, 调用一个需要传递 context 参数的函数, 但是并没有其他 context 可
以传递, 这时可以传递 todo. 这常常发生在重构进行中, 给一些函数添加一个 context 参数, 但不知道要传什么, 就用 todo "占
个位子", 最终要换成其他 context.


#### cancelCtx

`cancelCtx`, 顾名思义, 可以取消的 context, **可以被cancel, 没有存储值, 也没有deadline**

```cgo
type cancelCtx struct {
    Context
    
    // 保护之后的字段
    mu       sync.Mutex
    done     chan struct{}
    children map[canceler]struct{}
    err      error
}
```

`cancelCtx` 相比 `emptyCtx`, 它可以被 cancel(实现了 canceler 接口). 

注: `cancelCtx` 继承了 Context, 它本身只实现了 `Done()`, `Err()` 两个方法. 因此在创建 `cancelCtx` 的时必须要
有一个 `Context` 实例, 这样在调用 `Deadline()` 和 `Value()` 方法的时候, 直接调用 `Context` 实例的方法.


由 `cancelCtx` 产生的函数:

```cgo
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
    c := newCancelCtx(parent)
    propagateCancel(parent, &c)
    return &c, func() { c.cancel(true, Canceled) }
}
```

`propagateCancel()` 在 parent 被取消的时候, 取消 child. 主要的作用有两个:

1) 将当前的 child "挂载" 的 "可取消" 的 context(向上). 

2) 在 parent 取消的时候, 取消 child.

```cgo
func propagateCancel(parent Context, child canceler) {
    if parent.Done() == nil {
        return // parent is never canceled
    }
    
    // 找到可以取消的父 context
    if p, ok := parentCancelCtx(parent); ok {
        p.mu.Lock()
        if p.err != nil {
            // 父节点已经取消, 子节点也需要取消
            child.cancel(false, p.err)
        } else {
            // 父节点未取消, 将子节点加入到父节点当中.
            if p.children == nil {
                p.children = make(map[canceler]struct{})
            }
            p.children[child] = struct{}{}
        }
        p.mu.Unlock()
    } else {
        // 没有找到可取消的父 context, 新启动协程监控父节点/子节点取消信号
        go func() {
            select {
            case <-parent.Done():
                child.cancel(false, parent.Err())
            case <-child.Done():
            }
        }()
    }
}
```

上面函数的作用就是向上寻找可以 "挂载" 的 "可取消" 的 context, 并 "挂载" 上去. 这样, 在上层调用 cancel 方法的时候,
就可以层层传递, 将那些挂靠的子 context 一并 "取消".

函数当中的 `else`, 它是指当前节点context没有向上可以取消的父节点, 那么只能启动一个协程来监控父节点或子节点的取消动作.

> 这里存在一个疑问, 那就是在 `else` 当中, 既然没有找到可以取消的父节点, 那么 `case <-parent.Done()` 这个 case 将
> 永远不会发生, 所以可以忽略这个 case; 而 `case <-child.Done()` 这个 case 又啥事不干. 那这个 `else` 不就是多余的
> 吗?

`else` 的逻辑代码是必须的, 下面的 `parentCancelCtx()` 函数的代码:

`parentCancelCtx()` 递归查找 ctx 最近的可以取消父节点

```cgo
func parentCancelCtx(parent Context) (*cancelCtx, bool) {
    for {
        switch c := parent.(type) {
        case *cancelCtx:
            return c, true
        case *timerCtx:
            return &c.cancelCtx, true
        case *valueCtx:
            parent = c.Context
        default:
            return nil, false
        }
    }
}
```

`parentCancelCtx()` 函数很明确, 只会识别三种类型: `cancelCtx`, `timerCtx`, `valueCtx`. 若是自定义的类型(比如
把Context内嵌到一个类型里), 这里就无法识别了. 但是, 自定义的类型可能就是 "可取消" 的 context. 那么前面的代码当中没有
`else`, 则会导致程序出现 bug.

再回过头来看 `else` 的协程代码:

第一个case (`<-parent.Done()`) 说明当 parent 节点取消, 则取消子节点. 如果去掉这个case, 那么父节点(自定义的可取消
的context)取消的信号就无法传递到子节点.

第二个case (`<-child.Done()`) 说明如果子节点自己取消了, 那么就退出当前的 select, 父节点的取消信号就不用管了. 如果
去掉这个 case, 那么很可能父节点一直不取消, 这个 goroutine 就泄露了.


接下来, 看一下取消操作:

```cgo
// removeFromParent 是否从 parent 移除的标记
func (c *cancelCtx) cancel(removeFromParent bool, err error) {
    if err == nil {
        panic("context: internal error: missing cancel error")
    }
    c.mu.Lock()
    if c.err != nil {
        c.mu.Unlock()
        return // already canceled
    }
    c.err = err
    if c.done == nil {
        c.done = closedchan // 已经关闭的 chan
    } else {
        close(c.done)
    }

    // 取消子节点, 递归调用
    for child := range c.children {
        child.cancel(false, err)
    }
    c.children = nil
    c.mu.Unlock()
    
    if removeFromParent {
        // 移除 c. 先使用 parentCancelCtx() 方法获取到 c.Context 的可取消的父节点 parent, 
        // 然后在 parent 加锁的状况下删除 c.
        // 如果 parent 是自定义的 context, 那么 child 将不会保存到 parent 当中. 
        removeChild(c.Context, c)
    }
}
```

```
// 从 parent 当中删除 child. 这里的 parent 是 cancelCtx.parent
func removeChild(parent Context, child canceler) {
    p, ok := parentCancelCtx(parent)
    if !ok {
        return
    }

    // 这里找到的 p 是 *cancelCtx, 里面已经增加了 child
    p.mu.Lock()
    if p.children != nil {
        delete(p.children, child)
    }
    p.mu.Unlock()
}
```

这里的 `removeChild()` 函数的代码需要仔细阅读, 要不然会觉得奇怪. 将 child 节点添加到 child 可取消的父节点的位置是在
函数 `propagateCancel()` 当中, 里面的 parent 就是 `c.Context`, child 是 `c`. 因此在 `removeChild()` 函数也
是需要相同的参数, 相同的查找方式, 这样才能得到相同的元素. 

```
// 重写 Done 与 Err
func (c *cancelCtx) Done() <-chan struct{} {
    c.mu.Lock()
    if c.done == nil {
        c.done = make(chan struct{})
    }
    d := c.done
    c.mu.Unlock()
    return d
}

func (c *cancelCtx) Err() error {
    c.mu.Lock()
    err := c.err
    c.mu.Unlock()
    return err
}
```


#### timerCtx

`timerCtx`, 与时间相关的 Context, **可以设置deadline, 可以被cancel, 没有存储值**.

timerCtx 直接继承了 cancelCtx, 因此, 它可以被cancel.

```cgo
type timerCtx struct {
    cancelCtx
    timer *time.Timer // Under cancelCtx.mu.
    
    deadline time.Time
}
```

注: `timerCtx` 继承了 `cancelCtx`, 间接的继承了 `Context` 接口. 本身只实现了 `Deadline()` 方法, 继承了 `cancelCtx`
实现的 `Done()` 与 `Err()` 的实现. 但是, 它重写了 `cancel()` 方法.

由 timerCtx 派生的函数:

```cgo
func WithDeadline(parent Context, d time.Time) (Context, CancelFunc) {
    // 检查 parent 的 deadline
    // 如果有 deadline, 并且 deadline < d, 这说明当前即将创建的节点不可能到达 deadline, 
    // 那么当前即将创建的节点只能以取消的方式退出.
    if cur, ok := parent.Deadline(); ok && cur.Before(d) {
        return WithCancel(parent)
    }
    
    c := &timerCtx{
        cancelCtx: newCancelCtx(parent),
        deadline:  d,
    }
    	
    // "挂载"
    propagateCancel(parent, c) 
    
    // 准备 cancel 函数
    dur := time.Until(d)
    if dur <= 0 {
        // deadline has already passed
        // 取消, 删除当前的节点
        c.cancel(true, DeadlineExceeded) 
        return c, func() { c.cancel(false, Canceled) }
    }
    
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.err == nil {
        // deadline 导致的取消操作. time.AfterFunc 是一个异步的任务操作.
        c.timer = time.AfterFunc(dur, func() {
            c.cancel(true, DeadlineExceeded)
        })
    }
    return c, func() { c.cancel(true, Canceled) }
}

func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
    return WithDeadline(parent, time.Now().Add(timeout))
}
```

虽然说是两个函数, 本质上就是一个函数 `WithDeadline()`.


前面提到了, `timerCtx` 重写了 `cancel()` 方法.

```cgo
// removeFromParent 表示是否从可取消的父节点当中删除当前的节点
func (c *timerCtx) cancel(removeFromParent bool, err error) {
    c.cancelCtx.cancel(false, err) // 注意, 这里的值是 false
    if removeFromParent {
        // 在进行添加的时候, parent 实际上就是 c.cancelCtx.Context,
        // 因此这里的移除的操作的 parent 才是 c.cancelCtx.Context
        removeChild(c.cancelCtx.Context, c)
    }
    
    // 停止掉 timer 的操作
    c.mu.Lock()
    if c.timer != nil {
        c.timer.Stop()
        c.timer = nil
    }
    c.mu.Unlock()
}
```

本质上就是 `cancelCtx` 的一些封装操作. 增加了额外的 `timer` 清除操作. 

注: 在调用 `cancelCtx` 的 `cancel()` 方法的时候,  removeFromParent 参数是 false, 也就是说, 并没有从可取消的父
节点当中移除当前节点, 而是在当前的方法当前去移除当前节点. 这一点十分巧妙.


#### valueCtx

`valueCtx`, 与存储相关的 Context, **只能存储值**.

```cgo
type valueCtx struct {
    Context
    key, val interface{}
}
```

`valueCtx` 继承了 Context, 本身只实现了 `Value()` 方法.

由 valueCtx 派生的函数:

```cgo
func WithValue(parent Context, key, val interface{}) Context {
    if key == nil {
        panic("nil key")
    }
    if !reflectlite.TypeOf(key).Comparable() {
        panic("key is not comparable")
    }
    return &valueCtx{parent, key, val}
}
```

非常的简单. 这里不做过多的解释.

`Value()` 方法, 则是递归向上查找 key 对应的 value 值. 

和链表有点像, 只是它的方向相反: Context 指向它的父节点, 链表则指向下一个节点. 通过 `WithValue()` 函数, 可以创建层层
的 valueCtx, 存储 goroutine 间共享的变量.

```cgo
func (c *valueCtx) Value(key interface{}) interface{} {
    if c.key == key {
        return c.val
    }
    return c.Context.Value(key)
}
```

它会顺着链路一直往上找, 比较当前节点的key是否是要找的 key, 如果是, 则直接返回 value. 否则, 一直沿着 context 往前, 
最终找到根节点(一般是emptyCtx), 直接返回一个 nil. 因此在使用 Value 方法的时候要判断结果是否为 nil.

注: 因为查找方向是往上走的, 所以, 父节点无法获取子节点存储的值, 子节点却可以获取父节点的值.

`WithValue()` 函数创建 context 节点的过程实际上就是创建链表节点的过程. 两个节点的 key 值是可以相等的, 但它们是两个
不同的 context 节点. 查找的时候, 会向上查找到最后一个挂载的 context 节点, 也就是离得比较近的一个父节点 context.

查找效率低效, 同时存在值被覆盖的可能, 这是 `context.Vlaue` 最受争议的地方.

### 如何使用 context

官方的建议:

1. 不要将 Context 塞到结构体里(前面提到了, 这样会创建更多的协程来处理取消操作). 直接将 Context 类型作为第一个参数, 而
且一般命名为 ctx. 

2. 不要向一个函数传入一个 nil 的 context, 如果实在不知道传什么, 就传 todo

3. 不要把本应该作为参数的类型塞到 context 中, context 存储的应该是一些共同的数据. 例如: 登录session, cookie等

4. 同一个 context 可能会被传递到多个协程, 别担心, context是并发安全的
