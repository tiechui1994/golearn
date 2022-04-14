## go tool pprof 命令详解


命令格式: 
```
pprof [options] [binary] <source> ...
```

### 关于 options:

#### 输出格式(最多选择一个)

- `-svg`, SVG 格式

- `-text`, -top, 文本格式(经过了处理)

- `-png`, PNG 图片

- `-raw`, 原始文件的文本格式

- `-web`, 通过Web浏览器可视化图形(本质上还是一个svg图片).

- `-http [host]:[port]`, 通过HTTP的方式查看.

- `-tree`

- `-traces`, 所有函数调用栈, 以及调用栈的指标信息.

- `-list <regex>`, 正则表达式函数的注释源代码.

- `-weblist <regex>`, 正则表达式函数的注释源代码, web端展示.

- `-peek <regex>`, 正则表达式函数的caller/callee


#### 其他选项



### 关于输出格式

- flat, 本函数占用的内存量.

- flat%, 本函数内存占使用中的内存总量的百分比.

- sum%, 前面所有行 flat 百分比的和. 

- cum, 累积量, 假如 main 函数调用函数 f, 函数 f 占用的内存量, 也会计算进来.

- cum%, 累计量占总量的百分比.

> 重点关注的是 flat, sum%, 这关系到每一行代码使用的内存情况.