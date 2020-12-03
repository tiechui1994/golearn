## Go context 包

什么是 context ?

context, 中文翻译为 "上下文", 准确的说它是 goroutine 的上下文, 包含 goroutine 的运行状态, 环境, 现场等信息.

context 主要用来在 goroutine 之间传递上下文信息, 包括: "取消信号", "超时时间", "截止时间", "k-v" 等.

context 包的引入, 标准库很多接口因此都加上了 context 参数, 例如 database/sql 包. context 几乎成为了并发控制和超
时控制的标准做法.

> context.Context 类型的值可以协调多个 goroutine 中的代码执行 "取消" 操作, 并且可以存储键值对. 最为重要的是它是并
发安全的.

为什么需要 context ?
