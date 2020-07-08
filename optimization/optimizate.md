## Go 性能优化

### 定位瓶颈

golang 可以通过 benchmark + pprof 来定位具体的性能瓶颈.

#### benchmark 简介

```
go test -v gate_test.go -run=none -bench=. -benchtime=3s -cpuprofile cpu.prof -memprofile mem.prof  
```

- `-run` 单次执行, 一般用于代码逻辑验证

- `-bench` 执行所有的 Benchmark, 也可以通过用例函数名来指定部分测试用例.

- `-benchmem` 输出内存分配状况

- `-count N` 执行 tests 和 benchmarks N 次(默认是1次)

- `-benchtime D` 指定执行时长 

- `-parallel N` 最多并行运行 N 个 tests (默认为8个)

- `-timeout D` 持续时间 D 之后的 panic (默认0,禁用超时)

- `-trace FILE` 将执行跟踪写入文件

- `-cpuprofile FILE` 输出 cpu 的 pprof 信息文件

- `-memprofile FILE` 输出 heap 的 pprof 信息文件

- `-blockprofile FILE` 阻塞分析, 记录 goroutine 阻塞等待同步 (包括定时器通道) 的位置

- `-mutextprofile` 互斥锁分析, 报告互斥锁的竞争状况


benchmark 测试用例常用函数

- b.ReportAllocs() 输出单次循环使用的内存数量和对象 allocs 信息

- b.RunParallel() 使用协程并发测试

- b.SetBytes(n int64) 设置单次循环使用的内存数量


#### pprof 简介

> 生成方式

- **runtime/pprof**: 手动调用如 `runtime.StartCPUProfile` 或者 `runtime.StopCPUProfile` 等 API 来生成和
写入采样文件, 灵活度高. 主要用于本地测试.

- **net/http.pprof**: 通过 http 服务获取 Profile 采样文件, 简单易用, 适合于对应用程序的整体监控. 通过 `runtime/pprof`
实现. 主要用于服务器测试.

- **go test**: 通过 `go test -bench=. -cpuorofile cpu.prof`从生成采样文件, 主要用于本地基准测试. 可用于重点
测试某些函数.


> 查看方式

- `go tool pprof <format> [options] [binary] <source> ...`

**format**, 一个或多个

1) -text 纯文本

2) -web 生成 svg 并用浏览器打开(如果 svg 的默认打开方式是浏览器)

3) -svg 只生成 svg

4) -list function 筛选正则匹配 function 的函数的信息

5) **-http=":port" 直接本地浏览器打开 profile 查看 (包括 top, graph, 火焰图等)**


**source**

1) -seconds 基于时间的 profile 收集的持续时间

2) -timeout profile 收集超时(以秒为单位)

2) http://host/profile

3) profile.pb.gz


- go tool pprof -base prof1 prof2

1) 对比查看2个profile, 一般用于代码修改前后对比, 定位差异点.


- 通过命令行方式查看 profile 时, 查看信息.

```
flat  flat%   sum%        cum   cum%
5.66s 18.98% 18.98%     16.20s 54.33%  command-line-arguments.RandStringRunes
5.37s 18.01% 36.99%      5.37s 18.01%  sync.(*Mutex).Lock
4.98s 16.70% 53.69%      4.98s 16.70%  sync.(*Mutex).Unlock
4.62s 15.49% 69.18%     16.84s 56.47%  math/rand.(*lockedSource).Int63
2.11s  7.08% 76.26%     10.66s 35.75%  math/rand.(*Rand).Int31n
1.12s  3.76% 80.01%      1.12s  3.76%  math/rand.(*rngSource).Uint64
```

1)flat: 采样时, 该函数正在运行的次数*采样频率(10ms), 即得到 `估算` 的函数运行 "采样时间". *这里不包括函数等待子函
数返回*

2)flat%: flat/总采样时间

3)sum%: 前面所有行的 `flat%` 的累加值. 如第三行 53.69 = 18.98 + 18.01 + 16.70

4)cum: 采样时, 该函数出现在调用堆栈的采样时间, 包括函数等待子函数返回, 因此 flat <= cum

5)cum%: cum/总采样时间

>> 可以使用的命令, topN, list ncname


















