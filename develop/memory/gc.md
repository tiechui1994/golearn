## Go 垃圾回收

### 标记清除

标记清除(Mark-Sweep) 算法是最常见的垃圾收集算法. 标记清除搜集器是跟踪式垃圾收集器, 其执行过程可以分为标记 (Mark), 和
清除(Sweep) 两个阶段:

1. 标记阶段 -- 从根对象出发查找并并发标记所有存活的对象;

2. 清除阶段 -- 遍历队中的全部对象, 回收未被标记的垃圾对象并将回收的内存加入空闲链表.

如图所示, 内存空间中包含多个对象, 我们从根对象出发依次遍历对象的子对象并将从根节点可达的对象编辑成存活状态, 即 A, C 和
D 三个对象, 剩余的 B, E 和 F 三个对象因为从根节点不可达, 所以会被当做垃圾:

![image](/images/gc_mark_sweep_1.png)


标记阶段结束后会进入清除阶段, 在该阶段中收集器会依次遍历堆中的所有对象, 释放其中没有被标记的 B, E 和 F 三个对象并将新的
空间以链表的结构串联起来, 方便内存分配器的使用.

![image](/images/gc_mark_sweep_2.png)


上述介绍的是最传统的标记清除算法, 垃圾收集器从垃圾收集的根对象出发, 地柜遍历这些对象指向的子对象并将所有可达的对象标记成
存活; 标记阶段结束后, 垃圾收集器会依次遍历堆中的对象并清除其中的垃圾, 整个过程需要标记标记对象的存活状态, 用户程序在垃圾
收集的过程中不能指向, 我们需要更复杂的机制来解决STW的问题.

### 三色抽象

为了解决原始标记清除算法带来的长时间 STW, 多数现代的追踪式垃圾收集器都会实现三色标记算法的变种以缩短 STW 的时间. 三色
标记算法将程序中的对象分为白色, 黑色和灰色三类:

- 白色对象 -- 潜在的垃圾, 其内存可能会被垃圾收集器回收;

- 黑色对象 -- 活跃的对象, 包含**不存在任何引用外部指针的对象** 以及 **从跟对象可达的对象**;

- 灰色对象 -- 活跃的对象, 因为存在指向白色对象的外部指针, 垃圾收集器会扫描这些对象的子对象;


在垃圾收集器开始工作时, 程序中不存在任何的黑色对象, 垃圾收集的根对象会被标记为灰色, 来讲收集器只会从灰色对象结婚中取出
对象开始扫描, 当灰色集合中不存在任何对象时, 标记阶段就会结束.

![image](/images/gc_color_1.jpeg)


三色标记垃圾收集器的工作原理很简单, 分为以下结构步骤:

1.从灰色对象的集合中选择一个灰色对象并将其标记成黑色;

2.将黑色对象指向的所有对象都标记成灰色, 保证该对象和被该对象引用的对象都不会被回收;

3.重复上述的两个步骤直到对象图中不存在灰色对象;


当三个标记清除的 `标记阶段` 结束后, 应用程序的堆值就不存在任何的灰色的对象, 只能看到黑色的存活对象和白色的垃圾对象, 垃圾
收集器可以回收这些白色的垃圾. 下面是使用三色标记垃圾回收器执行标记后的堆内存, 堆中只有对象 D 为待回收的垃圾.

![image](/images/gc_color_marked.png)


`因为用户程序可能在标记执行的过程中修改对象的指针, 所以三色标记清除算法本身是不可以并发或者增量执行的, 它仍然需要 STW`.
在如下的三色标记过程中, `用户程序建立了从 A 对象到 D 对象的引用`, 但是因为程序中已经不存在灰色对象了, 所以 D 对象会被
垃圾收集器错误的回收.

![image](/images/gc_color_marking_ptr.png)


本来不该被回收的对象却被回收了, 这在内存管理中是非常严重的错误. 我们将这种错误称为悬挂指针, 即指针没有指向特定类型的合法
对象, 影响了内存的安全性. **想要并发或者增量地标记对象还是想要屏障技术**.

### 并发垃圾收集

Go 语言的并发垃圾收集器会在扫描对象之前暂停程序做一些标记对象的准备工作, 其中包括后台标记的垃圾收集器以及开启写屏障,如果
在后台执行的垃圾收集器不够快, 应用程序申请内存的速度超过预期, 运行时就会让申请内存的应用程序辅助完成垃圾收集的扫描解阶段,
在标记和标记终止阶段结束中华就会进入异步的清理阶段, 将不用的内存增量回收.

v1.5 版本实现的并发垃圾收集策略由专门的 Goroutine 负责在处理器之间同步和协调垃圾收集的状态. 当其他的 Goroutine 发现
需要触发垃圾收集时, 它们需要将该信息通知给负责修改状态的主 Goroutine, 然而这个通知的过程会带来一定的延迟, 这个延迟的时
间窗口很可能是不可控的, 用户程序会在这段时间分配很多内存空间.

v1.6 引入去中心化的垃圾收集协调机制, 将垃圾收集器变成一个显示的状态机, 任意的 Goroutine 都可以调用方法触发状态的迁移,
常见的迁移方法包括以下几个:

- runtime.gcStart -- 从 `_GCoff` 转换至 `_GCmark` 阶段, 进入并发标记阶段并打开写屏障

- runtime.gcMarkDone -- 如果所有可达对象都已经完成扫描, 调用 `runtime.gcMarkTermination`

- runtime.gcMarkTermination -- 从 `_GCmark` 转换 `_GCmarktermination` 阶段, 进入标记终止阶段并在完成后进入
`_GCoff`


### 混合写屏障

在 Go 语言 v1.7 版本之前, 运行时会使用 `插入写屏障` 保证强三色不变性, 但是 `运行时并没有在所有的垃圾收集根对象上开启
插入写屏障`, 因为 Go 语言的应用程序可能包含成百上千的 Goroutine, 而垃圾收集的根对象一般是包括全局变量和栈对象, 如果运
行时需要在几百个 Goroutine 的栈上都开启写屏障, 会带来巨大的开销, 所以实现上选择了 **在标记阶段完成时暂停程序,将所有栈
对象标记为灰色并重新扫描, 在活跃 Goroutine非常多的程序中, 重新扫描的过程需要占用 10 ~ 100ms 的时间**.

v1.8 中组合 `插入写屏障` 和 `删除写屏障`, 构成了如下所示的混合写屏障, 该写屏障会 **将覆盖的对象标记成灰色, 并在当前栈
没有扫描时将新对象也标记成灰色**.

```
writePointer(slot, ptr):
    shade(*slot)
    if current stack is grey:
        shade(ptr)
    *slot = ptr
```

为了移除栈的重新扫描过程, 除了引入混合写屏障之外, 在垃圾收集的标记阶段, 还需要**将场景的所有的新对象都标记成黑色**, 防
止新分配的栈内存和堆内存中的对象被错误的回收, 因为栈内存在标记阶段最终都会变为黑色, 所以不需要重新扫描占空间.


## 实现原理

Go 语言的垃圾收集可以分为`清除终止`, `标记`, `标记终止`, 和 `清除` 四个不同阶段, 它们分别完成了不同的工作:

1. 清理终止阶段;

- **暂停程序**, 所有的处理器在这时会进入安全点(Safe point);

- 如果当前垃圾收集循环是强制触发的, 我们还需要处理还未被清理的内存管理单元;

2. 标记阶段;

- 将状态切换至 `_GCmark`, 开启写屏障, 用户程序协助(Mutator Assiste)并将根对象入队;

- 恢复执行程序, 标记进场和用户协助的用户程序会开始并发标记内存中的对象, **写屏障会将被覆盖的指针和新指针都标记成灰色, 而
所有新创建的对象都会直接标记成黑色**;

- 开始扫描根对象, 包括所有的 Goroutine 的栈, 全局对象以及不在堆中的运行时数据结构, **扫描 Goroutine 栈期间会暂停当
前处理器**;

- 依次处理灰色队列中的对象, 将对象标记成黑色并将它们指向的对象标记成灰色;

- 使用分布式的终止算法检查剩余的工作, 发现标记阶段完成后进入标记终止阶段; 


3. 标记终止阶段;

- **暂停程序**, 将状态切换至 `_GCmarktermination` 并关闭辅助标记的用户程序;

- 清理处理器上的线程缓存;

4. 清理阶段;

- 将状态切换至 `_GCoff` 开始清理阶段, 初始化清理状态并关闭写屏障;

- 恢复用户程序, 所有新创建的对象会标记成白色;

- 后台并发清理所有的内存管理单元, 当 Goroutine 申请新的内存管理单元时就会触发清理;


### 全局变量

- runtime.gcphase 是垃圾收集器当前处于的阶段, 可能处于 `_GCoff`, `_GCmark`, `_GCmarktermination`, Goroutine
在读取或者修改该阶段时需要保证原子性;

- runtime.gcBackenEnabled 是一个布尔值, 当垃圾收集处于 `标记阶段` 时, 该变量会被置为1, 在这里 **辅助垃圾收集的用
户程序** 和 **后台标记的任务** 可以将对象涂黑.

- runtime.gcController, 实现了垃圾收集的调度算法, 它能够决定触发垃圾收集的时间和待处理的工作.

- runtime.gcpercent, 是触发垃圾收集的内存增长百分比, 默认情况下为 100, 即堆内存相比上次垃圾收集增长 100% 时应该触发
GC, 并行的垃圾收集器会在到达该目标前完成垃圾收集;

- runtime.writeBarrier 是一个包含写屏障的结构体, 其中的 `enabled` 字段表示写屏障的开启与关闭

- runtime.worldsema 全局的信号量, 获取该信号量的线程有权利暂停当前应用程序;


### 触发时机

运行时会通知如下所示的 `runtime.gcTrigger.test` 方法决定是否需要触发垃圾收集, 当满足触发垃圾收集的基本条件时 --`允
许垃圾收集`, `程序没有崩溃` 并且 `处于垃圾收集循环`, 该方法会根据三种不同的方式触发进行不同的检查:

```cgo
// 测试报告是否满足触发条件, 表示已满足 _GCoff 阶段的退出条件. 分配时应测试退出条件。
func (t gcTrigger) test() bool {
    if !memstats.enablegc || panicking != 0 || gcphase != _GCoff {
        return false
    }
	
    switch t.kind {
    case gcTriggerHeap:
        return memstats.heap_live >= memstats.gc_trigger
    case gcTriggerTime:
        if gcpercent < 0 {
            return false
        }
        lastgc := int64(atomic.Load64(&memstats.last_gc_nanotime))
        return lastgc != 0 && t.now-lastgc > forcegcperiod
    case gcTriggerCycle:
        return int32(t.n-work.cycles) > 0
    }
	
    return true
}
```

1. `gcTriggerHeap`, 堆内存的分配到达控制器计算的触发堆大小,

2. `gcTriggerTime`, 如果一定时间内没有触发, 就会触发新的循环, 该条件由 `runtime.forcegcperiod` 变量控制, 默认为
2 分钟

3. `gcTriggerCycle`, 如果当前没有开启垃圾收集, 则触发新的循环.

由于开启垃圾收集的方法 `runtime.gcStart` 会接收一个 `runtime.gcTrigger` 类型的参数, 根据这个触发 `_GCoff` 退出
的结构体找到所有触发的垃圾收集的代码:

- `runtime.sysmon()` 和 `runtime.forcegchelper()` -- 后台运行时定时检查和垃圾收集;

- `runtime.GC()` -- 用户程序手动触发垃圾收集;

- `runtime.mallocgc()` -- 申请内存时根据堆大小触发垃圾收集;

![image](/images/gc_tiggers.png)


### 后台触发


```cgo
func init() {
    go forcegchelper()
}

func forcegchelper() {
    forcegc.g = getg()
    for {
        lock(&forcegc.lock)
        if forcegc.idle != 0 {
            throw("forcegc: phase error")
        }
        atomic.Store(&forcegc.idle, 1)
        goparkunlock(&forcegc.lock, waitReasonForceGGIdle, traceEvGoBlock, 1)
        if debug.gctrace > 0 {
            println("GC forced")
        }
        gcStart(gcTrigger{kind: gcTriggerTime, now: nanotime()})
    }
}
```

为了减少对计算资源的占用, 该 Goroutine 会在循环中调用 `runtime.goparkunlock()` 主动陷入休眠等待其他 Goroutine 
的唤醒, `runtime,forcegchelper` 在大多数时间都是陷入休眠的, 但是它会被系统监控器 `runtime.sysmon()` 在满足垃圾
收集条件时唤醒;

```cgo
func sysmon() {
    ...
    for {
        ...
        // check if we need to force a GC
        if t := (gcTrigger{kind: gcTriggerTime, now: now}); t.test() && atomic.Load(&forcegc.idle) != 0 {
            lock(&forcegc.lock)
            forcegc.idle = 0
            var list gList
            list.push(forcegc.g)
            injectglist(&list)
            unlock(&forcegc.lock)
        }
        if debug.schedtrace > 0 && lasttrace+int64(debug.schedtrace)*1000000 <= now {
            lasttrace = now
            schedtrace(debug.scheddetail > 0)
        }
    }
}
```

forcegc 变量是连接 `runtime.sysmon()` 和 `runtime.forcegchelper()` 的桥梁