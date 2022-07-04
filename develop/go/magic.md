# go 当中的特殊链接

`//go:noescape` 禁止逃逸

`//go:nosplit` 禁止栈分裂

`//go:linkname local remote` 将 remote 包下的函数链接到 local 函数

`//go:systemstack` 仅允许在 g0 栈上执行的代码. 只允许出现在 runtime 包当中.

`//go:build` 进行条件编译. 只有满足条件, 才会编译当前的源文件.

`//go:cgo_import_dynamic local_func dynamic_func "dynamic_file.o"` 向动态库的方法导入到本地.
