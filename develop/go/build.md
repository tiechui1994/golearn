## go 命令

### go build

常用的参数:

- `-a` 用于强制重新编译所有涉及的 Go 语言代码包(包括Go语言标准库的代码包), 即使它们已经是最新的了. 该标记可以让我们有机会
通过改动底层的代码包做一些实验.

- `-n` 仅打印执行过程中用到的所有命令, 而不去真正执行命令.

- `-x` 打印命令执行过程中用到的所有命令, 并同时执行命令.

- `-race` 用于检测并报告指定 Go 语言程序存在的数据竞争问题. 当用 Go 语言编写并发程序的时候, 这是很重要的检测手段.

- `-v` 打印命令执行过程中涉及的代码包. 

- `-work` 打印命令执行时使用的临时工作目录的名字, 且命令执行完成后不删除它. 这个目录下的文件可以从侧面了解命令的执行过程. 
如果没有此标记, 那么临时目录会在命令执行完毕前删除.

- `-compiler name` 设置使用的编译器, 作为 runtime.Compiler (gc 或 gccgo), 默认值是 gc

- `-buildmode mode` 设置编译模式(archive, c-archive, c-shared, default, shared, exe, plugin), 详细内容参
考 `go help buildmode` 

- `-mod mode` 模块下载的模式(readonly, vendor, mod), 详细参考 `go help modules` 说明.

- `-gccgoflags '[pattern=]arg list'` 传递给 gccgo 编译器/连接器的参数.

- `-asmflags '[pattern=]arg list'` 传递给 `go tool asm` 调用的参数.

- `-ldflags '[pattern=]arg list'` 用于传递每个 `go tool link` 调用的参数. (gc编译器)

```
-I linker  添加搜索header文件的目录
-L directory 将指定目录添加到库路径
-X definition 使用 importpath.name=value 增加定义值

-c        dump call graph
-n        dump symbol table
-dumpdep  dump symbol dependency graph

-buildid id 设置编译唯一标识, 使用 file 命令可以查看次标识

-buildmode mode 设置编译模式(archive, c-archive, c-shared, default, shared, exe, plugin), 详细内容参考 
                `go help buildmode`, 默认值是 exe

-linkmode mode  设置 link mode (internal, external, auto), 具体描述在 cmd/cgo/doc.go

-linkshared 链接到已安装的 Go 共享库(实验性)

-extar string  存档程序, buildmode=c-archive
-extld linker  设置外部连接器(默认值是"clang" 或 "gcc"), 默认值是 gcc
-extldflasgs flags 设置外部连接器参数

-o file 将输出写入到指定文件

-race 启用竞态检测

-s 禁止生成符号表(symbol table)
-w 禁止生成 DWARF 
-v 打印link的追踪

-r path 设置 ELF 动态链接搜索路径, dir1:dir2:..., 相当于 gcc 当中 `-Wl,-rpath -Wl,dir1:dir2` 的作用
-pluginpath string 插件路径
```

- `-gcflags '[pattern=]arg list'` 用于传递每个 `go tool compile` 调用的参数. (gc编译器)

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


-std 编译标准库
```

## 编译压缩

- 编译使用 `-ldflags='-w -s'` 压缩编译内容

- 编译完成后, 可以使用 upx 进行压缩(`upx EXEC`).

upx 安装:

```
sudo apt-get install upx
```

