## http 请求转发在Go当中的实践

请求转发: 请求转发是服务器行为. 当客户端请求一个资源, 服务器会从另外的位置(可能是当前的域名, 也可能是别的域名)请求获取这
个资源, 然后返回给客户端.

请求重定向: 请求重定向是客户端行为. 当客户端请求一个资源, 服务器会向客户端发送一个状态码(通常是301,302,303,307,308当
中的一个)和一个重定向的地址(在响应的header的Location当中). 然后客户端再次针对重定向的地址再次发起请求(一般这个过程是浏
览器自动的行为), 获取资源.

两者主要的区别在哪里?

- 数据共享

请求转发: 请求转发共享一个 request
请求重定向: 需要两次请求, 数据无法进行共享. 还需要注意的时, 服务器不同的状态码会导致客户端再次发起请求使用的方法是不一样
的, 从而可能造成两次请求数据不一致的问题. 这个在实际开放当中需要考虑到.

- 请求

请求转发: 用户一次请求, 即可以获取数据. 相对效率比较高.
请求重定向: 用户至少需要两次请求. 相对效率比较低.

> 这里只能说 `相对效率`, 原因在于请求转发在服务端会再次产生一次请求, 这次请求对用户是无感知的. 所以从整个链路来考虑, 都
会产生两次请求. 只不过, 请求转发请求第二次请求的时间可能比请求重定向第二次请求实际更短一些.(资源的位置导致的)

### 请求重定向

一般状况下, 请求重定向是直接在 `nginx` 当中进行配置的. 当然了, 代码当中也是可以处理的(原理都是一样的). 下面展示一个在
`nginx` 当中重定向的例子:

```
server {
    server_name www.jd.com jd.com;
    listen 80;

    location / {
        return 308 https://www.jd.com$request_uri;
    }
}
```

上述展示的是一个将 jd 的将 HTTP 请求转发到 HTTPS 请求的案例.


### 请求转发

请求转发也是可以使用 `nginx` 实现的. 使用 `rewrite`(改写请求的url, 里面包括请求协议, 域名, 端口, URI等内容) 就可以
办到了. 但是这里, 我将主要介绍在业务代码当中去实现请求转发.

首先, 请求转发, 服务器得知道请求转到哪里去, 也就是要知道目标资源的位置. 

其次, 就是发送请求了.

实现方式一(通用):

```cgo
func ForwardHandler(writer http.ResponseWriter, request *http.Request) {
    u, err := url.Parse("https://target.com/path/to/uri")
    if nil != err {
        log.Println(err)
        return
    }
    
    proxy := httputil.ReverseProxy{
        Director: func(request *http.Request) {
            request.URL = u
        },
    }

    proxy.ServeHTTP(writer, request)
}
```

在 go 当中 `*http.Request` 要被转发, 只需要修改其 URL 的指向即可, 不用再创建一个新的 `*http.Request` 对象了.

使用了 `net/http/httputil` 包当中专门处理请求的结构体. 它的内容如下:

```cgo
type ReverseProxy struct {
    Director func(*http.Request)

    Transport http.RoundTripper

    FlushInterval time.Duration

    ErrorLog *log.Logger

    BufferPool BufferPool

    ModifyResponse func(*http.Response) error

    ErrorHandler func(http.ResponseWriter, *http.Request, error)
}
```

在 `httputil.ReverseProxy` 当中 `Director` 主要是处理转发请求的, 在上面的例子当中就是修改了 `*http.Request` 对
象的 URL;  `ErrorHandler` 是用于处理当转发的目标出现错误的时候, 一个处理回调函数; `ModifyResponse` 可用于修改转发
请求的的响应内容; `Transport` 本质上就是一个 client, 用于做请求发送的一些限制, 比如超时时间, TLS连接 等等. 其他的字
段就不做说明了, 平时很少用到, 感兴趣的可以去源码中查看说明.

实现方式二:

```cgo
func ForwardHandler(writer http.ResponseWriter, request *http.Request) {
    u := &url.URL{
        Scheme: "https",
        Host:   "target.com",
    }
    
    proxy := httputil.NewSingleHostReverseProxy(u)
    request.URL.Path = "/path/to/uri"
    proxy.ServeHTTP(writer, request)
}
```

这种方式其实和上面的方式是一样的, 只是这种方式简洁一些. 但是需要注意的是, 上面的参数都是必须设置的, 如果缺少, 则会出现
一些错误的. 建议实际使用的是上面的第一种方式.


request.URL.Host 和 request.Host 的区别?

对于服务器请求, request.Host 指定在其搜索 URL 的主机. 它的值是 request.Header 当中 "Host" 的对于的值, 或 URL本
身提供的主机名. 它的形式可能是 "host:port". 对于国际域名, 主机可以采用 Unicode 形式. 为了防止 DNS 重新绑定攻击, 服
务器应验证 request.Header 当中 "Host" 的对于的值是否具有权威性. 

对于客户端请求, request.Host 可以选择覆盖要发送的 request.Header 当中 "Host" 的值. 如果为空, 则 Request.Write
方法会使用 URL.Host 的值替代.


request.URL 字段是通过解析 HTTP 请求URI创建的, 因此 request.URL.Host 和 request.URL.Schema 都是 ""; 而且 
request.URL.Path 和 request.RequestURI 的值是一样的.

request.Host 字段是 request.Header 中 "Host" 的值(request.Header.Get("Host")的值是空的).

r.URL.Host 和 r.Host的值几乎总是不同的. 在代理服务器上, r.URL.Host 是目标服务器的主机, r.Host是代理服务器本身的主机.
当不通过代理连接时, 客户端不会在请求URI中指定主机. 在这种情况下, r.URL.Host 是 "".
