## Client 解析

### 相关 struct

```cgo
type Client struct {	
	Transport RoundTripper
	CheckRedirect func(req *Request, via []*Request) error
	Jar CookieJar
	Timeout time.Duration
}
```

- Client 字段的解析:

Client 是HTTP客户端. 它的零值(DefaultClient)是使用DefaultTransport的可用客户端.

> Transport

`Client` 的 `Transport` 通常具有内部状态(缓存的TCP连接), 因此 `Client` 可以被重用. `Client` 可以
安全地被多个 `goroutine` 并发使用.

`Client` 是比 `RoundTripper` (例如`Transport`) 更高一级, 并且还处理 `HTTP详细信息`, 例如 `Cookie`
和 `Redirect`.

> Jar 和 CheckRedirect

在进行重定向后, `Client` 将 `Forward(转发)` 在初始请求上设置的所有标头, 但以下情况除外:

1) 将 `sensitive(敏感的) Header` (例如 "Authorization", "WWW-Authenticate" 和 "Cookie", 
"Cookie2") 转发到不受信任的目标时. 当 `redirect(重定向)` 到一个 "与子域不匹配" 或 "与初始域不完全匹配" 的域时, 
将忽略这些 `Header`. 例如, 从 "foo.com" 重定向到 "foo.com" 或 "sub.foo.com" 将转发 `sensitive Header`, 
但重定向到 "bar.com" 则不会.

2) 如果 `Jar` 非空, 则使用 `Jar` 转发 "Cookie" 头时, 由于每个重定向可能会更改 `CookieJar` 的状态, 
因此重定向可能会更改初始请求中设置的 "Cookie" 内容. 当转发 "Cookie" 头时, 任何修改的 `Cookie` 都将被
省略, 并希望 `Jar` 将插入那些修改的 `Cookie` (假设 origin 匹配).

如果 `Jar` 为 `nil`, 则将转发原始 `Cookie`, 而不进行任何更改.

> Timeout

Timeout, 用于设定 Client 发出请求的时间限制. 包括 `connect time`, `any redirects time`, 
`reading response body time`.

在 `Get`, `Head`, `Post`, 或 `Do` 返回之后, 计时器保持运行状态, 并且在读取 `Response.Body` 
时(出现error时候), 计时器中断.

Timeout 设置为0, 表示没有超时限制.

`Client` 可以使用 `Request.Cancel()` 取消使用 `Transport` 的请求. 传递给 `Client.Do()` 
的 `Request` 依旧可以设置 `Request.Cancel`.  两者都会取消请求.

为了兼容, `Client` 会在 `Transport` 中使用 `CancelRequest (deprecated)` 方法取消请求.
新版本的 `RoundTripper` 的实现应该使用 `Request.Cancel` 代替 `CancelRequest` 方法去取消
请求.


- RoundTripper 

```cgo
type RoundTripper interface {
	RoundTrip(*Request) (*Response, error)
}
```

`RoundTrip` 执行一个 `HTTP事务`, 为提供的 `Request` 返回一个 `Response`.

> 实现需要注意的细节点:

1. `RoundTrip` 不应尝试解析 `Response`. 特别是, 如果 `RoundTrip` 获得响应, 则必须返回 `error` 是
否为 `nil`, 而不管响应的 `HTTP` 状态代码如何.

2. 应该保留 `non-nil error`, 以获取失败的响应. 同样, `RoundTrip` 不应尝试处理更高级别的协议详细信息,
例如 `redirect`, `authentication` 或 `cookie`.

3. 除了 `consuming(读取)` 和 `closing Request's Body` 之外, `RoundTrip` 不应修改 `Request`. 
`RoundTrip` 可以在单独的 `goroutine` 中读取请求的字段.

> **在响应的 Body 关闭之前, 调用者不应更改 Request.**

4. `RoundTrip` 必须始终 **close Request Body**, 包括发生错误时, 但根据实现的不同, 即使 `RoundTrip` 
返回后, 也可能在单独的 goroutine 中关闭它. 这意味着希望 `reuse body` 以用于后续请求的调用者必须安排在等
待 `Close` 调用之后再这样做.

> 调用RoundTrip的前提: Request 的 URL 和 Header 字段必须初始化.


- CookieJar

```cgo
type CookieJar interface {
	// SetCookies 在给定URL的回复中处理cookie的接收.
  // 它可能会 或 可能不会 选择保存 Cookie, 具体取决于jar的策略和实现.
	SetCookies(u *url.URL, cookies []*Cookie)

  // Cookies 返回 cookie, 以发送对给定URL的请求. 具体实现取决于标准的cookie使用限制,
  // 例如RFC 6265中的限制.
    	
  Cookies(u *url.URL) []*Cookie
}
```

## 相关的 Method

- Do() 

send Request and return Response.

1. 如果是由 `Client` 策略(例如CheckRedirect) 或 无法 HTTP (例如网络连接问题) 引起的, 
则返回err. 非2xx状态代码不会导致错误.

2. 如果返回的错误为 `nil`, 则 `Response` 将包含一个非 `nil` 的 `Body`, **希望用户可以
将其关闭**. 如果未关闭 `Body`, 则 `Client` 的 `RoundTripper`(通常为 `Transport` ) 可
能无法将与服务器的持久TCP连接重新用于后续的 "keep-alive" 请求.

3. `Request Body` (如果非空) 将被在 `Transport` 当中关闭, 即使发生错误也是如此.

4. 发生错误时, 任何 `Response` 都可以忽略. 一个 `non-nil Response`, 带有一个 `non-nil
err`, 仅仅仅发生在 `CheckRedirect` 失败时. 并且此时返回的 `Response.Body` 已关闭.

5. 如果服务器回复一个 `Redirect`, 则 `Client` 首先使用 `CheckRedirect` 函数确定是否进行
重定向. 如果允许, 则 `301`, `302` 或 `303` 重定向会导致后续请求使用 `HTTP` 方法GET (如果
原始请求为`HEAD`, 则为`HEAD`), 并且不携带任何 `Body`. 如果定义了 `Request.GetBody` 函数, 
则 `307` 或 `308` 重定向将保留原始的 HTTP 方法和 `Body`.

> NewRequest 方法为每个 Request 自动设置了 GetBody 参数.
 
```cgo
func (c *Client) Do(req *Request) (*Response, error) {
  // request的URL不存在
	if req.URL == nil {
		req.closeBody()
		return nil, errors.New("http: nil Request.URL")
	}

	var (
		deadline      = c.deadline() // 获取请求的deadline,也就是Timeout
		reqs          []*Request
		resp          *Response
		copyHeaders   = c.makeHeadersCopier(req) // copy 当前的 Header, 并且在返回Response的时候修改Header
		reqBodyClosed = false // 当前的Request是否关闭

		// Redirect的参数: Method, Body
		redirectMethod string
		includeBody    bool
	)
	
	// 包装所有的返回的 Error
	uerr := func(err error) error {
		// 在 c.send() 当中可能已经关闭了 Body
		if !reqBodyClosed {
			req.closeBody()
		}
		method := valueOrDefault(reqs[0].Method, "GET") // 原始请求方法
		var urlStr string // 请求的URL
		if resp != nil && resp.Request != nil {
			urlStr = resp.Request.URL.String()
		} else {
			urlStr = req.URL.String()
		}
		return &url.Error{
			Op:  method[:1] + strings.ToLower(method[1:]),
			URL: urlStr,
			Err: err,
		}
	}
	
	for {
	  // Redirect 才会走到此处
		// 对于除第一个请求以外的所有请求, 创建下一个Request 并替换 req.
		if len(reqs) > 0 {
		  // 获取重定向的 Location, 即新的请求的 URL
			loc := resp.Header.Get("Location") // 
			if loc == "" {
				resp.closeBody()
				return nil, uerr(fmt.Errorf("%d response missing Location header", resp.StatusCode))
			}
			u, err := req.URL.Parse(loc)
			if err != nil {
				resp.closeBody()
				return nil, uerr(fmt.Errorf("failed to parse Location header %q: %v", loc, err))
			}
			
			// 如果调用方指定了自定义Host标头, 并且重定向 Location 是相对的, 则通过重定向保留Host标头.
			host := ""
			if req.Host != "" && req.Host != req.URL.Host {
				if u, _ := url.Parse(loc); u != nil && !u.IsAbs() {
					host = req.Host
				}
			}
			ireq := reqs[0]
			req = &Request{
				Method:   redirectMethod,
				Response: resp,
				URL:      u,
				Header:   make(Header),
				Host:     host,
				Cancel:   ireq.Cancel,
				ctx:      ireq.ctx,
			}
			if includeBody && ireq.GetBody != nil {
				req.Body, err = ireq.GetBody()
				if err != nil {
					resp.closeBody()
					return nil, uerr(err)
				}
				req.ContentLength = ireq.ContentLength
			}

			// 如果用户在第一个请求上设置了Referer, 则在设置Referer之前复制原始Header. 
			// 如果他们真的想覆盖, 则可以在其 CheckRedirect 函数中进行覆盖.
			copyHeaders(req)

		  // 如果不是 https-> http, 则将来自最新请求URL的Referer标头添加到新请求URL中:
			if ref := refererForURL(reqs[len(reqs)-1].URL, req.URL); ref != "" {
				req.Header.Set("Referer", ref)
			}
			
			// 检查重定向. 默认函数: 只有不超过10次重定向即可
			err = c.checkRedirect(req, reqs)

			// 特殊的错误, 前哨错误使用户可以选择前一个响应, 而无需关闭其主体.
			if err == ErrUseLastResponse {
				return resp, nil
			}

			// 关闭之前的Response的Body. 但是会读取一部分 Response 的 Body.
			// 如果 Body 很小, 则将重新使用之前的TCP连接. 无需检查读取的错误,
			// 如果失败, Transport 将不会再使用它
			const maxBodySlurpSize = 2 << 10
			if resp.ContentLength == -1 || resp.ContentLength <= maxBodySlurpSize {
				io.CopyN(ioutil.Discard, resp.Body, maxBodySlurpSize)
			}
			resp.Body.Close()

      // 当 CheckRedirect 失败之后, 返回已经关闭的 Response 和 err
			if err != nil {
				ue := uerr(err)
				ue.(*url.Error).URL = loc
				return resp, ue
			}
		}

		reqs = append(reqs, req)
		var err error
		var didTimeout func() bool // Timeout 函数回调函数
    // 调用 c.send() 发送请求, 获取响应
		if resp, didTimeout, err = c.send(req, deadline); err != nil {
			// c.send() 当中总是会关闭 Request 的 Body
			reqBodyClosed = true
			// 当没有 Timeout 的时候, 需要执行 Timeout 的回调, 做最后的清理工作
			if !deadline.IsZero() && didTimeout() {
				err = &httpError{
					err:     err.Error() + " (Client.Timeout exceeded while awaiting headers)",
					timeout: true,
				}
			}
			return nil, uerr(err)
		}
    
    // 请求响应成功, 需要处理是否是 Redirect 的状况
		var shouldRedirect bool
		redirectMethod, shouldRedirect, includeBody = redirectBehavior(req.Method, resp, reqs[0])
		if !shouldRedirect {
			return resp, nil // 一般正常的请求, 成功响应到此就返回了 
		}
    
    // 当进行Redirect时, 关闭 Request Body
		req.closeBody()
	}
}
```

// Header 的处理函数
```cgo
// 返回一个 Function, 用于从 initial Request (ireq) 当中 copy Header 到 Request.
// 对于每一个 redirect, 此函数都必须被调用从而复制Header到即将发送的Request
func (c *Client) makeHeadersCopier(ireq *Request) func(*Request) {
	// 使用 closure callbak 保存原始的 Headers
	var (
		ireqhdr  = ireq.Header.clone() // 原始的Header
		icookies map[string][]*Cookie // 原始的Cookie信息
	)
	
	// 当存在 Jar 的时候, Copy原始的Cookie信息
	if c.Jar != nil && ireq.Header.Get("Cookie") != "" {
		icookies = make(map[string][]*Cookie)
		for _, c := range ireq.Cookies() {
			icookies[c.Name] = append(icookies[c.Name], c)
		}
	}
  
  // 使用 preq 指向原始的 Request
	preq := ireq 
	return func(req *Request) {
	  // 如果 Jar 不为 nil, 并且存在一些初始Cookie(原始cookie), 那么在进行重定向时, 我们可能需要更改初始
	  // Cookie, 因为每次重定向最终都可能会修改现有的Cookie.
    //
    // 由于已在请求Header中设置的Cookie不包含有关原始的 "domain" 和 "path" 的信息, 因此以下逻辑假定任
    // 何新设置的Cookie都将覆盖原始cookie，而与 domain 或 path 无关.
		if c.Jar != nil && icookies != nil {
			var changed bool
			resp := req.Response // 重定向返回的Response, 删除掉在原始Request当中存在的Cookie信息
			for _, c := range resp.Cookies() {
				if _, ok := icookies[c.Name]; ok {
					delete(icookies, c.Name)
					changed = true
				}
			}
			
			// 重新设置原始 Header 当中 Cookie 信息
			if changed {
				ireqhdr.Del("Cookie") 
				var ss []string
				// 过滤之后的Cookie
				for _, cs := range icookies {
					for _, c := range cs {
						ss = append(ss, c.Name+"="+c.Value)
					}
				}
				sort.Strings(ss) // Ensure deterministic headers
				ireqhdr.Set("Cookie", strings.Join(ss, "; ")) // 将原始的Cookie设置为修改后的Cookie
			}
		}

		// 复制原始 Header 当中其他的信息. 
		// 特殊处理了 "Authorization", "Www-Authenticate", "Cookie", "Cookie2",
		// 如果是相同的 Domain, SubDomain, 则会进行设置, 否则, 这些Header会被处理掉
		for k, vv := range ireqhdr {
			if shouldCopyHeaderOnRedirect(k, preq.URL, req.URL) {
				req.Header[k] = vv
			}
		}
    
    // 更新 preq 为当前的 Request, 下次还可能调用.
		preq = req 
	}
}
```

// Redirect 的行为判断
// 301, 302, 303: 修改原始的请求方法 
// 307, 308: 不修改原始的请方法
```cgo
func redirectBehavior(reqMethod string, resp *Response, ireq *Request) (redirectMethod string, shouldRedirect, includeBody bool) {
	switch resp.StatusCode {
	case 301, 302, 303:
		redirectMethod = reqMethod
		shouldRedirect = true
		includeBody = false

		if reqMethod != "GET" && reqMethod != "HEAD" {
			redirectMethod = "GET"
		}
	case 307, 308:
		redirectMethod = reqMethod
		shouldRedirect = true
		includeBody = true
    
    // 代码当中这个是多余的. 因为后面有统一处理
		if resp.Header.Get("Location") == "" {
			shouldRedirect = false
			break
		}
		if ireq.GetBody == nil && ireq.outgoingLength() != 0 {
			shouldRedirect = false
		}
	}
	return redirectMethod, shouldRedirect, includeBody
}
```

// 发送请求.
// CookieJar 的好处是单独独立维护Cookie信息, 每次在请求发送之前添加, 请求响应之后进行更新. 达到与业务逻辑分离的
// 目的. 这样一个Client只需要维护一份Cookie, 所有使用该Client都可以使用它.
```cgo
func (c *Client) send(req *Request, deadline time.Time) (resp *Response, didTimeout func() bool, err error) {
  // 发送请求前增加新的Cookie信息. CookieJar
	if c.Jar != nil {
		for _, cookie := range c.Jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
	}
	
	// 发送请求
	resp, didTimeout, err = send(req, c.transport(), deadline)
	if err != nil {
		return nil, didTimeout, err
	}
	
	// 请求响应之后, 将响应的Cookie信息保存到Jar当中
	if c.Jar != nil {
		if rc := resp.Cookies(); len(rc) > 0 {
			c.Jar.SetCookies(req.URL, rc)
		}
	}
	return resp, nil, nil
}
```

// 发送 HTTP Request, 响应 Response, Timeout 回调, error

```cgo
func send(ireq *Request, rt RoundTripper, deadline time.Time) (resp *Response, didTimeout func() bool, err error) {
	// req 是原始的 Request
	req := ireq 
	
	// rt, ireq 参数的判断.
	if rt == nil {
		req.closeBody()
		return nil, alwaysFalse, errors.New("http: no Client.Transport or DefaultTransport")
	}
	if req.URL == nil {
		req.closeBody()
		return nil, alwaysFalse, errors.New("http: nil Request.URL")
	}
	if req.RequestURI != "" {
		req.closeBody()
		return nil, alwaysFalse, errors.New("http: Request.RequestURI can't be set in client requests.")
	}

	// 浅层 Copy Request. 
	forkReq := func() {
		if ireq == req {
			req = new(Request)
			*req = *ireq // shallow clone
		}
	}
  
  // 初始化 Header
	if req.Header == nil {
		forkReq()
		req.Header = make(Header)
	}
    
  // 初始化 Authorization, 设置了 username 和 passwd, 但是没有设置 Authorization
	if u := req.URL.User; u != nil && req.Header.Get("Authorization") == "" {
		username := u.Username()
		password, _ := u.Password()
		forkReq()
		req.Header = cloneHeader(ireq.Header)
		req.Header.Set("Authorization", "Basic "+basicAuth(username, password))
	}
  
  // 保证在没有到达 deadline 的时候 Copy 一次
	if !deadline.IsZero() {
		forkReq()
	}
	
	// 设置请求的停止计时函数 和 判断是否超时
	stopTimer, didTimeout := setRequestCancel(req, rt, deadline)

  // 真正执行 HTTP 请求的方法 Transport 的 RoundTrip 方法
	resp, err = rt.RoundTrip(req)
	// 请求出现问题
	if err != nil {
		stopTimer() // 停止deadline的计时器, 没有超时, 但是失败了
		if resp != nil {
			log.Printf("RoundTripper returned a response & error; ignoring response")
		}
		
		// 检查是否是 HTTPS 的握手出现问题. HTTPS的 Client 请求 HTTP 的Server
		if tlsErr, ok := err.(tls.RecordHeaderError); ok {
			if string(tlsErr.RecordHeader[:]) == "HTTP/" {
				err = errors.New("http: server gave HTTP response to HTTPS client")
			}
		}
		
		return nil, didTimeout, err
	}
	
	// 请求成功, 且没有到达 deadline, 会设置 Response 的 Body 信息, 否则不会设置Body信息
	if !deadline.IsZero() {
		resp.Body = &cancelTimerBody{
			stop:          stopTimer,
			rc:            resp.Body,
			reqDidTimeout: didTimeout,
		}//如果截止日期不为零，则setRequestCancel设置要求的“取消”字段。 RoundTripper的类型用于确定是否应使用传统的CancelRequest行为。
//
//作为背景，有三种取消请求的方法：
//首先是Transport.CancelRequest。 （已弃用）
//第二个是Request.Cancel（此机制）。
//第三是Request.Context。
	}
	
	return resp, nil, nil
}
```

// 如果deadline不为零, 则 setRequestCancel 设置要求的 "Cancel" 字段. 
// RoundTripper 的类型用于确定是否应使用传统的 `CancelRequest` 行为.
//
//As Background, 有三种取消请求的方法:
//首先是 Transport.CancelRequest.(已弃用)
//第二个是 Request.Cancel (此机制).
//第三是 Request.Context

```cgo
func setRequestCancel(req *Request, rt RoundTripper, deadline time.Time) (stopTimer func(), didTimeout func() bool) {
	if deadline.IsZero() {
		return nop, alwaysFalse
	}
    
  // 原始请求的 Cancel, Channel
	initialReqCancel := req.Cancel 
	
	// 重新定义的 Cancel. 由于在调用 setRequestCancel 之前, Request 已经进行至少一次的浅拷贝. 此时的 Request 和
	// 用户传入的 Request 的指向已经发生了改变. 但是Request当中的指针数据被保存了下来, 因此在这里可以进行重新设置 Cancel
	// 而不影响用户的 Cancel
	cancel := make(chan struct{})
	req.Cancel = cancel

  // Cancel 的 回调函数
	doCancel := func() {
		close(cancel) // 关闭这里定义的 cancel
    
    // 兼容 1.5, 1.6 版本的的 Transport
		type canceler interface {
			CancelRequest(*Request)
		}
		switch v := rt.(type) {
		case *Transport, *http2Transport: // 当前版本的RoundTripper
			// Do nothing. The net/http package's transports
			// support the new Request.Cancel channel
		case canceler: // 老版本的 RoundTripper
			v.CancelRequest(req)
		}
	}
	
	// 计时功能
	stopTimerCh := make(chan struct{})
	var once sync.Once
	stopTimer = func() { once.Do(func() { close(stopTimerCh) }) }

	timer := time.NewTimer(time.Until(deadline))
	var timedOut atomicBool

	go func() {
		select {
		case <-initialReqCancel: // 用户调用 Request.Cancel 取消请求
			doCancel()
			timer.Stop()
		case <-timer.C: // 函数超时, 系统
			timedOut.setTrue()
			doCancel()
		case <-stopTimerCh: // 停止计时(标准库调用), 此时请求已经有响应, 成功(读取body) 或者 失败
			timer.Stop()
		}
	}()

	return stopTimer, timedOut.isSet
}
```

- cancelTimerBody

```cgo
// cancelTimerBody是一个io.ReadCloser, 它使用以下两个功能包装rc:
// 1) 在读取错误或关闭时, 调用stop函数. 停止计时
// 2）在读取失败时, 如果 reqDidTimeout 为true, 则包装错误并标记为 net.Error 达到超时. 判断超时函数
type cancelTimerBody struct {
	stop          func() 
	rc            io.ReadCloser
	reqDidTimeout func() bool
}

func (b *cancelTimerBody) Read(p []byte) (n int, err error) {
	n, err = b.rc.Read(p)
	if err == nil {
		return n, nil
	}
	b.stop() // 停止计时
	if err == io.EOF {
		return n, err
	}
	
	// 超时错误
	if b.reqDidTimeout() {
		err = &httpError{
			err:     err.Error() + " (Client.Timeout exceeded while reading body)",
			timeout: true,
		}
	}
	return n, err
}

func (b *cancelTimerBody) Close() error {
	err := b.rc.Close()
	b.stop() // 停止计时
	return err
}
```

