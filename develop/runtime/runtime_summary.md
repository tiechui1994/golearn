## go 调度 - runtime 函数小结

### systemstack VS mcall

- mcall

mcall 从 g 切换到 g0 栈上, 并调用 `fn(g)`, 其中 `g` 是进行调用的 goroutine. 

macll 将 g 的当前 PC/SP 保存在 g->sched 中, 以便以后可以恢复. fn 在稍后执行, 通常通过将 g 记录在数据结构当中, 导致
稍后调用 ready(g).

函数 fn 绝对不能返回, 通常它以调用 schedule 结束, 然 m 运行其他 goroutine. 被调用是函数 fn 不能是 `go:noescape`
(如果 fn 是堆栈分配的闭包, fn 将 g 放入运行队列, 并且 g 在 fn 返回之前执行, 则闭包在执行时).

mcall 只能从 g 堆栈上调用(不能是 g0 或 gsignal). 

- systemstack

systemstack 在系统堆栈上运行 fn. 

如果从 per-OS-thread(g0) 栈调用 systemstack, 或者如果从 gsignal 栈调用 systemstack, 则 systemstack 直接调用
fn 并返回.

否则, systemstack 需要切换到 per-OS-thread(g0) 栈, 调用 fn, 然后切换回来.

通常使用 func 字面量(实际传递给 systemstack 的是一个指针)作为参数, 以便与调用系统栈共享输入和输出:

```cgo
... set up y ...

systemstack(func() {
    x = bigcall(y)
})

... use x ...
```

### gogo VS gosave

gogo 函数原型: `func gogo(buf *gobuf)`

gosave 函数原型: `func gosave(buf *gobuf)`

gogo 作用是切换到 `gobuf.g` 栈上, 并恢复关联的 SP/PC 寄存器.

gosave 作用是将当前 g 栈和 SP/PC 寄存器保存到 `gobuf` 当中. 与 runtime.save() 函数功能是一致的.

### newproc(siz int32, fn *funcval)

创建协程, 全程必须要在 g0 栈上执行(systemstack), 创建一个协程 g(newproc1), 将 g 加入到当前的 p 对应的本地队列当中.

newproc1 当中做的事情:

- 栈分配, 必须在 g0 栈上(systemstack), 2048

- 参数拷贝到栈上

- 设置 g.sched, 重点是 sp, pc(先指向了 goexit, 后面会修改), 有个很重要的函数 `gostartcallfn`, 制造一次函数调用, 
使 g.sched.pc 指向了 fn 的第一个指令上, 当 g 被执行起来, 直接从该指令开始执行.

### newm(fn func(), p *p, id int64)

fn: m.mstartfn 

p: 临时借用的 p, 也是 m 启动后绑定的 p(如果 fn 是一个forever-loop函数, p 可以为空. 目前存在: sysmon, templateThread)

创建线程 m, 创建 m 对象的同时分配 g0 对象(栈大小无限制)(在 allocm 中完成); 创建 g0 运行的函数(newm1的 newosproc)

- allocm 当中, 创建 m, 创建 g0, 并将两者进行绑定关联, m.mstartfn 设置为 fn, m.nextp 设置为 p

- newm1 当中, 调用 clone(汇编实现, 系统调用) 创建新的系统线程.

```
int32 clone(int32 flags, void *stk, M *mp, G *gp, void (*fn)(void));

// flags 标识符
// stk 栈大小
// mp 可选参数, m, 拷贝到 child 当中
// gp 可选参数, g0, 拷贝到 child 当中
// fn, clone 完成后开始执行的函数, runtime.mstart
```

关于 mstart, 汇编实现, 直接调用 mstart1(运行 m.mstartfn 函数, 可能永不返回), 开启执行 schedule()
