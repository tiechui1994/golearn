## runtime 总结

### systemstack VS mcall

- mcall

mcall 切换到 g0 栈上, 执行指定的函数. 执行的是函数类型 `func(*g)`, 其中 `g` 是调用执行的 goroutine. 被调用的函数 
fn 绝对不能返回, 通常它以调用 schedule 结束, 然 m 运行其他 goroutine. 被调用是函数 fn 不能是 `go:noescape` (如
果 fn 是堆栈分配的闭包, fn 将 g 放入运行队列, 并且 g 在 fn 返回之前执行, 则闭包在执行时).

mcall 只能从 g 堆栈上调用(不能是 g0 或 gsignal). 

- systemstack

systemstack 在系统堆栈上运行 fn. 

如果从 per-OS-thread(g0) 栈调用 systemstack, 或者如果从 gsignal 栈调用 systemstack, 则 systemstack 直接调用
fn 并返回.

否则, systemstack 将从 goroutine 的栈中调用. 在这种情况下, systemstack 切换到 per-OS-thread 栈, 调用 fn, 然后
切换回来.

通常使用 func 字面量作为参数, 以便与调用系统栈共享输入和输出:

```cgo
... set up y ...

systemstack(func() {
    x = bigcall(y)
})

... use x ...
```


