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

