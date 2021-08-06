## 条件编译

### 文件命名约定

如果一个 `go` 源文件的文件名, 去掉扩展名和可能存在的 `_test` 后缀之后, 符合下面的格式, 则只在特定操作系统或架构编译.

- `_GOOS`

- `_GOARCH`

- `_GOOS_GOARCH` 

案例:

```
xxx_windows_amd64.go
```


### 条件构建标记

条件构建标识通过代码注释的形式实现.

条件构建标识以 `// +build` 开头, **必须写在文件的最顶端, 前面只能有空白行或其他注释行, 构建标识与 `package` 之间必须
有一个空白行**

规则:

- 不同的 tag 之间以空格分割, tag与tag之间关系是 OR 关系
- 不同的 tag 之间以逗号分割, tag与tag之间关系是 AND 关系
- tag 由字母和数据区分, 以 `!` 开头表示条件非
- 当一个文件存在多个 tag 时, tag与tag之间的关系是AND

案例:

```
// +build linux,386 drawin,!cgo

package main
```

> 表示 (linux AND 386) OR (darwin AND !cgo)


```
// +build linux drawin
// +build 386

package main
```

> 表示 (linux OR darwin) AND 386


```
// +build ignore

package main
```

> 特例: 表示该文件在任何状况下都不被构建

在一次特定构建中, 条件构建标识需要满足以下条件:

- `runtime.GOOS` 中对操作系统的定义
- `runtime.GOARCH` 中对CPU架构的定义
- 使用的编译器, `gc` 或 `gccgo`
- 如果 `ctx.CgoEnabled` 为 true, 支持 cgo
- `go1.N` 代表 Go 编译器版本大于 `1.N` 时编译
- `ctx.BuildTags` 当中的其它标识.
