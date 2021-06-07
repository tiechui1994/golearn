## go 命令

### go build

常用的参数:

- `-a` 用于强制重新编译所有涉及的 Go 语言代码包(包括Go语言标准库的代码包), 即使它们已经是最新的了. 该标记可以让我们有
机会通过改动底层的代码包做一些实验.

- `-n` 仅打印执行过程中用到的所有命令, 而不去真正执行命令.

- `-x` 打印命令执行过程中用到的所有命令, 并同时执行命令.

- `-race` 用于检测并报告指定 Go 语言程序存在的数据竞争问题. 当用 Go 语言编写并发程序的时候, 这是很重要的检测手段.

- `-v` 打印命令执行过程中涉及的代码包. 

- `-work` 打印命令执行时使用的临时工作目录的名字, 且命令执行完成后不删除它. 这个目录下的文件可以从侧面了解命令的执行过
程. 如果没有此标记, 那么临时目录会在命令执行完毕前删除.

- `-ldflags '[pattern=]arg list'` 用于传递每个 `go tool link` 调用的参数.

```
-I linker  添加搜索header文件的目录
-L directory 将指定目录添加到库路径
-X definition 使用 importpath.name=value 增加定义值

-c        dump call graph
-n        dump symbol table
-dumpdep  dump symbol dependency graph

-buildid id 设置编译唯一标识, 使用 file 命令可以查看次标识
-buildmode mode 设置编译mode
-linkmode mode 设置链接mode

-extar string  存档程序, buildmode=c-archive
-extld linker  在external mode下链接时使用的链接器.
-extldflasgs flags 将链接参数传递给外部链接器

-o file 将输出写入到指定文件

-race 启用竞态检测

-s 禁止生成符号表(symbol table)
-w 禁止生成 DWARF 
-v 打印link的追踪

-r path 设置 ELF 动态链接搜索路径, dir1:dir2:...
-pluginpath string 插件路径
```

- `-gcflags '[pattern=]arg list'` 用于传递每个 `go tool compile` 调用的参数.

```
-B 禁止边界检测
-C 禁止打印错误消息的行

-D path 设置本地导入的相对路径.
-I directory 将directory添加到导入搜索的目录
-E 调试符号表导出

-N 禁止编译器优化
-l 禁止内联

-S 打印汇编(assembly)结果
-V 打印版本并退出

-W debug parse tree after type checking
-w debug type checking

-asmhdr file 将汇编的header写入到文件file

-m 打印优化的决策
```

## 编译压缩

- 编译使用 `-ldflags='-w -s'` 压缩编译内容

- 编译完成后, 可以使用 upx 进行压缩(`upx EXEC`).

upx 安装:

```
sudo apt-get install upx
```

