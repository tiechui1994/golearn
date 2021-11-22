## go mod 

### go mod edit

修改 go.mod 文件

- `-require=path@version` 和 `-droprequire=path` 添加和删除给定的路径和版本的模块. 需要注意, `-require` 会覆
盖任何现有的模块.

- `-exclude=path@version` 和 `-dropexclude=path@version` 添加和删除给定模块路径和版本的排除项. 

- `-replace=old[@v]=new[@v]`, 添加给定的replace模块和版本对. 如果省略了old@v 当中的 @v, 则一个添加了左侧的没有
版本的replace, 适用于旧模块路径的所有版本; 如果省略了new@v 当中的 @v, 则新路径应该是本地模块根目录, 而不是模块路径. 
注意: -replace 会覆盖 `old[@v]` 的任何冗余目录, 因此省略了 `@v` 将删除特定版本的现有替换.

- `-dropreplace=old[@v]`, 删除给定模块路径和版本对的replace. 如果省略了 `@v`, 左侧没有版本的replace将被删除.

> 注: go1.13, 针对 require, exclude, 其 path 必须以域名开始, 例如: `domain.com/module`, 但是在 go1.16 当中, 
对 path 的格式已经没有特别的规定.


### 常见的问题

- `invalid path: malformed module path "xxx": missing dot in first path element` 

在 go1.13 当中, 对 `require`, `exclude` 的 path 的格式是类似 `domain.com/xxx`, 如果格式错误, 则会出现上述的问
题.

