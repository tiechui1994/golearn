# go get 以及自定义私有包

## go get 从网络上获取源码包的协议是什么?

环境变量: 
默认情况下, 获取 vcs 信息使用的是 https 协议, 在获取失败的情况下, 对于 `GOINSECURE` 当中配置的域名, 可以尝试使用 `http` 方式获取.`

1. 转换包路径为 HTTP URL. 例如: `golang.org/x/net` => `https://golang.org/x/net?go-get=1`, 下载到对应的 html 文件. 

2. 解析下载的 html 文件, 查询 `<meta name="go-import" content="xxx git|svn|fossil|mod yyy">`, 其中 xxx 表示包名, yyy 表示包的目标路径(只支持特定的几类协议).
对于中间部分, git|hg|svn|fossil 表示使用对应协议去下载目标地址. mod 是比较特殊, 不做任何处理.

> note: 获取包目标路径, 默认支持 mod, git, hg 协议方式, 可以通过 `GOPRIVATE` 来支持其他的协议方式. 

3. 根据不同的协议去下载. 其中 git 是执行 `clone`, svn 是 `checkout`, hg 是 `clone`, `mod` 是做任何处理.

| vcs 协议 | 目标链接 schema |
| --- | --- |
| git | git, https, http, ssh, git+ssh |
| svn | svn, https, http, svn+ssh |

## 搭建私有包

针对不同的包路径请求地址, 返回 html 内容当中携带 `<meta name="go-import" content="xxx git|svn yyy">`.
例如： `<meta name="go-import" content="hello.com/xyz/ppp git ssh://hello.local/xyz/ppp">`

