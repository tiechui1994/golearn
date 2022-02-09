## runtime 总结

1. systemstack vs mcall

相同点: 切换到 g0 栈上, 然后执行指定的函数.

不同点:

mcall 执行的是函数类型 `func(*g)`, 其中 `g` 是调用执行的 goroutine. 被调用的函数 fn 绝对不能返回, 通常它以调用
schedule 结束, 然 m 运行其他 goroutine. 被调用是函数 fn 不能是 `go:noescape` (如果 fn 是堆栈分配的闭包, fn 将
g 放入运行队列, 并且 g 在 fn 返回之前执行, 则闭包在执行时).

mcall 只能从 g 堆栈上调用(不能是 g0 或 gsignal). 


