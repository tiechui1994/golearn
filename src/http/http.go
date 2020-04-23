package main

import (
	"net/http"
	"time"
)

type Client struct {
	Transport     RoundTripper
	CheckRedirect func(req *http.Request, via []*http.Request) error
	Jar           http.CookieJar
	Timeout       time.Duration
}

/**
Client 是HTTP客户端. 它的零值(DefaultClient)是使用DefaultTransport的可用客户端.

Client的 Transport 通常具有内部状态(缓存的TCP连接), 因此 Client 可以被重用.
Client 可以安全地被多个goroutine并发使用.

Client 是比 RoundTripper (例如Transport) 更高一级, 并且还处理HTTP详细信息, 例如cookie和重定向.

在进行重定向后, Client 将 forward(转发) 在初始请求上设置的所有标头, 但以下情况除外:

1) 将 sensitive(敏感的) Header (例如 "Authorization", "WWW-Authenticate" 和 "Cookie") 转发
到不受信任的目标时. 当 redirect(重定向) 到一个 "与子域不匹配" 或 "与初始域不完全匹配" 的域时, 将忽略这
些Header. 例如, 从 "foo.com" 重定向到 "foo.com" 或 "sub.foo.com" 将转发敏感Header, 但重定向到
"bar.com" 则不会.

2) 如果 Jar 非空, 则使用 Jar 转发 "Cookie" 头时, 由于每个重定向可能会更改 CookieJar 的状态, 因此重
定向可能会更改初始请求中设置的 "Cookie" 内容. 当转发 "Cookie" 头时, 任何修改的 Cookie 都将被省略, 并
希望 Jar 将插入那些修改的 Cookie (假设 origin 匹配).

如果 Jar 为 nil, 则将转发原始 cookie, 而不进行任何更改.
*/

/**
Jar 指定 CookieJar.

对于每一个出站的Request, 使用 Jar 插入相关的 Cookies.
对于每一个入站的 Response, 使用 Jar 更新相关的 Cookies.

如果 Jar 为 nil, 则仅当在 Request 上显示设置 Cookie, 才发送 Cookie.
**/

/**
Timeout, 用于设定 Client 发出请求的时间限制. 包括 connect time, any redirects, reading response body.

在Get, Head, Post, 或 Do 返回之后, 计时器保持运行状态, 并且在读取 Response.Body 时, 计时器中断.

Timeout 设置为0, 表示没有超时限制.

Client 可以使用 Request.Cancel() 取消使用 Transport 的请求. 传递给 Client.Do() 的 Request 依旧可以设置
Request.Cancel.  两者都会取消请求.

为了兼容, Client 会在 Transport 中使用 CancelRequest (deprecated) 方法取消请求.
新版本的 RoundTripper 的实现应该使用 Request.Cancel 代替 CancelRequest 方法去取消请求.
**/

/**
RoundTrip执行一个HTTP事务, 为提供的 Request 返回一个 Response.

// 实现注意的细节点:

1. RoundTrip 不应尝试解析 Response. 特别是, 如果 RoundTrip 获得响应, 则必须返回 err 是否为
nil, 而不管响应的 HTTP 状态代码如何.

2. 应该保留 non-nil err, 以获取失败的响应. 同样, RoundTrip 不应尝试处理更高级别的协议详细信息,
例如重定向, 身份验证 或 cookie.

3. 除了 consuming 和 closing Request's Body 之外,  RoundTrip 不应修改 Request. RoundTrip
可以在单独的 goroutine 中读取请求的字段.

> 在响应的 Body 关闭之前, 调用者不应更改请求.

4. RoundTrip 必须始终 close Body, 包括发生错误时, 但根据实现的不同, 即使 RoundTrip 返回后, 也
可能在单独的 goroutine 中关闭它. 这意味着希望 reuse body 以用于后续请求的调用者必须安排在等待 Close
调用之后再这样做.

请求的 URL 和 Header 字段必须初始化.
*/
type RoundTripper interface {
	RoundTrip(*http.Request) (*http.Response, error)
}

/**

Client Do() 说明

1. 如果返回的错误为nil, 则 Response 将包含一个非 null 的Body, 希望用户可以将其关闭.
如果未关闭主体, 则 Client 的 RoundTripper(通常为Transport) 可能无法将与服务器的持久
TCP连接重新用于后续的 "keep-alive" 请求.

2. 请求 Body (如果非空) 将被在 Transport 当中关闭, 即使发生错误也是如此.

3. 发生错误时, 任何 Response 都可以忽略. 一个 non-nil Response, 带有一个 non-nil
err, 仅仅仅发生在 CheckRedirect 失败时. 并且此时返回的 Response.Body 已关闭.


4. 如果服务器回复一个 Redirect, 则 Client 首先使用 CheckRedirect 函数确定是否进行重定向.
如果允许, 则301, 302或303重定向会导致后续请求使用 HTTP 方法GET (如果原始请求为HEAD, 则为
HEAD), 并且不携带任何Body. 如果定义了 Request.GetBody 函数, 则307或308重定向将保留原始的
HTTP 方法和 Body.

NewRequest 函数自动为通用标准库为Request设置 GetBody 方法.

**/
func Do(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func main() {
	request, _ := http.NewRequest("GET", "", nil)
	request.Close = false
	request.Proto = "HTTP/0.9"
	request.Proto = "HTTP/1.0"
	request.Proto = "HTTP/1.1"
	request.Proto = "HTTP/2.0"

	// 0.9
	request.ProtoMinor = 0
	request.ProtoMajor = 0

	// 1.0, 1.1
	request.ProtoMinor = 1
	request.ProtoMajor = 1

	// 2.0
	request.ProtoMinor = 2
	request.ProtoMajor = 2

	http.DefaultClient.Do(request)
}
