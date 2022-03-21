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

### Transport

- roundTrip

```cgo
func (t *Transport) roundTrip(req *Request) (*Response, error) {
	t.nextProtoOnce.Do(t.onceSetNextProtoDefaults)
	ctx := req.Context()
	trace := httptrace.ContextClientTrace(ctx)
	
	
    // ... URL, Header 校验 
    scheme := req.URL.Scheme
    isHTTP := scheme == "http" || scheme == "https"
    
	origReq := req
	cancelKey := cancelKey{origReq}
	req = setupRewindBody(req) // 对当前 req.Body 进行再包装. 包装之后的 req.Body 实现了 GetBody() 方法
	
	// 对于特殊的 req.URL.Scheme, alternateRoundTripper 返回用于此 req.URL.Scheme 的 RoundTripper
	// 对于正常使用 Transport 的情况下, alternateRoundTripper 返回值为 nil
	// 
	// 对于 req.URL.Scheme="https", 并且是 websocket 请求(Connection:upgrade, Upgrade:websocket), 返
	// 回值是 nil
    // 其他的情况, 则是从 altProto 当中获取对应的 req.URL.Scheme 的 RoundTripper. 
    // 注: 可以通过 RegisterProtocol 方法注册自定义的 RoundTripper
	if altRT := t.alternateRoundTripper(req); altRT != nil {
		if resp, err := altRT.RoundTrip(req); err != ErrSkipAltProtocol {
			return resp, err
		}
		// err == ErrSkipAltProtocol
		var err error
		req, err = rewindBody(req)
		if err != nil {
			return nil, err
		}
	}
	
	// scheme 必须是 http 或 https, 如果不是, 需要自己去实现 RoundTripper.
	if !isHTTP {
		req.closeBody()
		return nil, badStringError("unsupported protocol scheme", scheme)
	}
	
	// method, host 校验
	if req.Method != "" && !validMethod(req.Method) {
		req.closeBody()
		return nil, fmt.Errorf("net/http: invalid method %q", req.Method)
	}
	if req.URL.Host == "" {
		req.closeBody()
		return nil, errors.New("http: no Host in request URL")
	}
    
    // 建立连接 -> 请求
	for {
		select {
		case <-ctx.Done():
			req.closeBody()
			return nil, ctx.Err()
		default:
		}

		// treq gets modified by roundTrip, so we need to recreate for each retry.
		treq := &transportRequest{Request: req, trace: trace, cancelKey: cancelKey}
		cm, err := t.connectMethodForRequest(treq)
		if err != nil {
			req.closeBody()
			return nil, err
		}

		// Get the cached or newly-created connection to either the
		// host (for http or https), the http proxy, or the http proxy
		// pre-CONNECTed to https server. In any case, we'll be ready
		// to send it requests.
		
		// 获取一个 conn, 对端可能是:
		// to host(对于 http 或 https), 
		// to http proxy
		// to https server(CONNECT)
		// 无论如何, 我们都会准备好向它发送请求.
		pconn, err := t.getConn(treq, cm)
		if err != nil {
			t.setReqCanceler(cancelKey, nil)
			req.closeBody()
			return nil, err
		}

		var resp *Response
		if pconn.alt != nil {
			// HTTP/2 path.
			t.setReqCanceler(cancelKey, nil) // not cancelable with CancelRequest
			resp, err = pconn.alt.RoundTrip(req)
		} else {
			resp, err = pconn.roundTrip(treq)
		}
		if err == nil {
			resp.Request = origReq
			return resp, nil
		}
        
		// err != nil, 清理并确定是否重试.
		if http2isNoCachedConnError(err) {
			if t.removeIdleConn(pconn) {
				t.decConnsPerHost(pconn.cacheKey)
			}
		} else if !pconn.shouldRetryRequest(req, err) {
			// Issue 16465: return underlying net.Conn.Read error from peek,
			// as we've historically done.
			if e, ok := err.(transportReadFromServerError); ok {
				err = e.err
			}
			return nil, err
		}
		testHookRoundTripRetried()

		// Rewind the body if we're able to.
		req, err = rewindBody(req)
		if err != nil {
			return nil, err
		}
	}
}
```


- getConn, 获取连接到"对端"的持久化连接

```cgo
func (t *Transport) getConn(treq *transportRequest, cm connectMethod) (pc *persistConn, err error) {
	req := treq.Request
	trace := treq.trace
	ctx := req.Context()
	if trace != nil && trace.GetConn != nil {
		trace.GetConn(cm.addr())
	}

	w := &wantConn{
		cm:         cm,
		key:        cm.key(),
		ctx:        ctx,
		ready:      make(chan struct{}, 1),
		beforeDial: testHookPrePendingDial,
		afterDial:  testHookPostPendingDial,
	}
	defer func() {
		if err != nil {
			w.cancel(t, err)
		}
	}()

	// Queue for "idle connection". 
	// 尝试从空闲连接队列当中查找可以使用的连接. 一旦成功, 设置好请求 cancel 的函数.
	if delivered := t.queueForIdleConn(w); delivered {
		pc := w.pc
		// Trace only for HTTP/1.
		// HTTP/2 calls trace.GotConn itself.
		if pc.alt == nil && trace != nil && trace.GotConn != nil {
			trace.GotConn(pc.gotIdleConnTrace(pc.idleAt))
		}
		// set request canceler to some non-nil function so we
		// can detect whether it was cleared between now and when
		// we enter roundTrip
		t.setReqCanceler(treq.cancelKey, func(error) {})
		return pc, nil
	}

    // 为 dail connection 做准备工作. 
	cancelc := make(chan error, 1)
	t.setReqCanceler(treq.cancelKey, func(err error) { cancelc <- err })

	// Queue for "dial connection".
	t.queueForDial(w)

	select {
	case <-w.ready:
		// Trace success but only for HTTP/1.
		// HTTP/2 calls trace.GotConn itself.
		if w.pc != nil && w.pc.alt == nil && trace != nil && trace.GotConn != nil {
			trace.GotConn(httptrace.GotConnInfo{Conn: w.pc.conn, Reused: w.pc.isReused()})
		}
		if w.err != nil {
			// If the request has been cancelled, that's probably
			// what caused w.err; if so, prefer to return the
			// cancellation error (see golang.org/issue/16049).
			select {
			case <-req.Cancel:
				return nil, errRequestCanceledConn
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case err := <-cancelc:
				if err == errRequestCanceled {
					err = errRequestCanceledConn
				}
				return nil, err
			default:
				// return below
			}
		}
		return w.pc, w.err
	case <-req.Cancel:
		return nil, errRequestCanceledConn
	case <-req.Context().Done():
		return nil, req.Context().Err()
	case err := <-cancelc:
		if err == errRequestCanceled {
			err = errRequestCanceledConn
		}
		return nil, err
	}
}
```

// queueForIdleConn, 查询持久化的 idle 连接

```cgo
// queueForIdleConn 将 w 排队以接收 w.cm 的下一个空闲连接.
// 作为对调用者的优化提示, queueForIdleConn 报告它是否成功传递了一个已经空闲的连接.
func (t *Transport) queueForIdleConn(w *wantConn) (delivered bool) {
    // 禁止 KeepAlive 了, 每次都会建立新的连接.
	if t.DisableKeepAlives {
		return false
	}

	t.idleMu.Lock()
	defer t.idleMu.Unlock()

	// Stop closing connections that become idle - we might want one.
	// (That is, undo the effect of t.CloseIdleConnections.)
	t.closeIdle = false

	if w == nil {
		return false // Happens in test hook.
	}

	// 当设置了 IdleConnTimeout, 需要计算最小的 persistConn.idleAt 的时间. 
	// 如果 persistConn 超过缓存时间, 则需要进行清理操作.
	var oldTime time.Time
	if t.IdleConnTimeout > 0 {
		oldTime = time.Now().Add(-t.IdleConnTimeout)
	}

	// 查询最近的空闲连接
	if list, ok := t.idleConn[w.key]; ok {
		stop := false
		delivered := false
		for len(list) > 0 && !stop {
			pconn := list[len(list)-1]

			// 查看此连接是否空闲时间过长
			tooOld := !oldTime.IsZero() && pconn.idleAt.Round(0).Before(oldTime) 
			if tooOld {
				// 进行异步清理. 它获取持有的 idleMu, 并执行同步 net.Conn.Close
				go pconn.closeConnIfStillIdle()
			}
			if pconn.isBroken() || tooOld {
				// 如果 persistConn.readLoop 被标记为连接断开, 但 Transport.removeIdleConn 尚未将其从空闲列
				// 表中删除, 或者如果此 persistConn 太老(空闲时间过长), 则忽略它, 并查询下一个. 
				// 在这两种情况下, 它都已经处于关闭过程中.
				list = list[:len(list)-1]
				continue
			}
			
			// 尝试将 pconn, err 传递给 w. 在 w.pc 和 w.err 都为 nil 状况下, 尝试传递.
			delivered = w.tryDeliver(pconn, nil)
			if delivered {
				if pconn.alt != nil {
					// HTTP/2: multiple clients can share pconn.
					// Leave it in the list.
				} else {
					// HTTP/1: only one client can use pconn.
					// Remove it from the list.
					t.idleLRU.remove(pconn)
					list = list[:len(list)-1]
				}
			}
			stop = true
		}
		
		// 对 t.idleConn 的值重新处理.
		if len(list) > 0 {
			t.idleConn[w.key] = list
		} else {
			delete(t.idleConn, w.key)
		}
		
		if stop {
			return delivered // 是否过度成功的的标志
		}
	}
    
    // idleConn 当中不存在 w.key
    // 需要往 t.idleConnWait 当中添加一个新的等待的请求, 当下一次空闲出来了, 可以进行请求
	if t.idleConnWait == nil {
		t.idleConnWait = make(map[connectMethodKey]wantConnQueue)
	}
	
	// wantConnQueue 是一个队列. 实现上比较巧妙: 两个数据交换使用.
	// 连个队列数组: head, tail, 使用 headPos 指定当前的头位置
	// pushBack 添加尾元素, 直接添加到 tail
	// popFront 弹出头元素的时候可能会将 head 和 tail 进行交换.
	q := t.idleConnWait[w.key]
	q.cleanFront() // 清理头部元素
	q.pushBack(w)
	t.idleConnWait[w.key] = q
	return false
}
```


// queueForDial, 连接新连接, 并将当前请求加入到新连接的队列当中.

```cgo
// 将 w 加入到排队建立连接的队列当中. 一旦允许建立连接, 则开启新 goroutine 异步操作.
func (t *Transport) queueForDial(w *wantConn) {
	w.beforeDial() // 空函数
	// MaxConnsPerHost, 一个 host 最多的连接数
	if t.MaxConnsPerHost <= 0 {
		go t.dialConnFor(w)
		return
	}

	t.connsPerHostMu.Lock()
	defer t.connsPerHostMu.Unlock()
    
    // 可以为 host 建立新连接
	if n := t.connsPerHost[w.key]; n < t.MaxConnsPerHost {
		if t.connsPerHost == nil {
			t.connsPerHost = make(map[connectMethodKey]int)
		}
		t.connsPerHost[w.key] = n + 1 // 修改 connsPerHost 的值.
		go t.dialConnFor(w)
		return
	}
    
    // 只能等待.
	if t.connsPerHostWait == nil {
		t.connsPerHostWait = make(map[connectMethodKey]wantConnQueue)
	}
	q := t.connsPerHostWait[w.key]
	q.cleanFront()
	q.pushBack(w)
	t.connsPerHostWait[w.key] = q
}
```

// dialConnFor, 建立连接

```cgo
// 建立 persistConn, 并将结果传递给 w
// dialConnFor 是已获得建立 persistConn 的权限, 计数存入了 t.connsPerHost[w.key] 当中
func (t *Transport) dialConnFor(w *wantConn) {
	defer w.afterDial() // 空函数.

	pc, err := t.dialConn(w.ctx, w.cm)
	delivered := w.tryDeliver(pc, err) // 将 pc 传递给 w 
	if err == nil && (!delivered || pc.alt != nil) {
		// pconn 无法传递给 w, 或当前是可以共享的 HTTP/2 请求,
		// 将 pc 加入到 idle 连接池当中.
		t.putOrCloseIdleConn(pc)
	}
	if err != nil {
		t.decConnsPerHost(w.key) // 减少 t.connsPerHost 的值
	}
}

// 建立 persistConn
// connectMethod:
//   proxyURL,  代理连接(包含有授权信息, 可以是 http, https, socket5)
//   targetScheme, 目标连接使用的协议, 只能是 http 或 https
//   targetAddr, 目标连接地址
// 
// 特例: 当 targetScheme 使用的是 http, 但 proxyURL 是https|http代理, 此时 targetAddr 的值将被重置为空
func (t *Transport) dialConn(ctx context.Context, cm connectMethod) (pconn *persistConn, err error) {
	pconn = &persistConn{
		t:             t,
		cacheKey:      cm.key(),
		reqch:         make(chan requestAndChan, 1),
		writech:       make(chan writeRequest, 1),
		closech:       make(chan struct{}),
		writeErrCh:    make(chan error, 1),
		writeLoopDone: make(chan struct{}),
	}
	trace := httptrace.ContextClientTrace(ctx)
	wrapErr := func(err error) error {
		if cm.proxyURL != nil {
			return &net.OpError{Op: "proxyconnect", Net: "tcp", Err: err}
		}
		return err
	}
	
	// https, 并且自定义了 DialTLS 或 DialTLSContext 方法
	if cm.scheme() == "https" && t.hasCustomTLSDialer() {
		var err error
		// 优先选择 DialTLSContext.
		pconn.conn, err = t.customDialTLS(ctx, "tcp", cm.addr())
		if err != nil {
			return nil, wrapErr(err)
		}
		if tc, ok := pconn.conn.(*tls.Conn); ok {
			// Handshake here, in case DialTLS didn't. TLSNextProto below
			// depends on it for knowing the connection state.
			if trace != nil && trace.TLSHandshakeStart != nil {
				trace.TLSHandshakeStart()
			}
			if err := tc.Handshake(); err != nil {
				go pconn.conn.Close()
				if trace != nil && trace.TLSHandshakeDone != nil {
					trace.TLSHandshakeDone(tls.ConnectionState{}, err)
				}
				return nil, err
			}
			cs := tc.ConnectionState()
			if trace != nil && trace.TLSHandshakeDone != nil {
				trace.TLSHandshakeDone(cs, nil)
			}
			pconn.tlsState = &cs
		}
	} else {
	    // http, https但未自定义TLS连接方法. 
	    // cm.addr(), 如果存在代理, 则返回代理地址, 否则返回目标地址.
	    // cm.scheme(), 如果存在代理, 则返回代理协议, 否则返回目标协议.
		conn, err := t.dial(ctx, "tcp", cm.addr())
		if err != nil {
			return nil, wrapErr(err)
		}
		pconn.conn = conn
		if cm.scheme() == "https" {
			var firstTLSHost string
			if firstTLSHost, _, err = net.SplitHostPort(cm.addr()); err != nil {
				return nil, wrapErr(err)
			}
			if err = pconn.addTLS(firstTLSHost, trace); err != nil {
				return nil, wrapErr(err)
			}
		}
	}

	// 初始化代理.
	switch {
	case cm.proxyURL == nil:
		// 不使用任何代理
	case cm.proxyURL.Scheme == "socks5":
	    // socks5代理
		conn := pconn.conn
		d := socksNewDialer("tcp", conn.RemoteAddr().String())
		// 代理账号密码处理
		if u := cm.proxyURL.User; u != nil {
			auth := &socksUsernamePassword{
				Username: u.Username(),
			}
			auth.Password, _ = u.Password()
			d.AuthMethods = []socksAuthMethod{
				socksAuthMethodNotRequired,
				socksAuthMethodUsernamePassword,
			}
			d.Authenticate = auth.Authenticate
		}
		if _, err := d.DialWithConn(ctx, conn, "tcp", cm.targetAddr); err != nil {
			conn.Close()
			return nil, err
		}
	case cm.targetScheme == "http":
	    // http 代理
		pconn.isProxy = true
		if pa := cm.proxyAuth(); pa != "" {
			pconn.mutateHeaderFunc = func(h Header) {
				h.Set("Proxy-Authorization", pa)
			}
		}
	case cm.targetScheme == "https":
	    // https 代理, 发送 CONNECT 请求
		conn := pconn.conn
		hdr := t.ProxyConnectHeader
		if hdr == nil {
			hdr = make(Header)
		}
		if pa := cm.proxyAuth(); pa != "" {
			hdr = hdr.Clone()
			hdr.Set("Proxy-Authorization", pa)
		}
		connectReq := &Request{
			Method: "CONNECT",
			URL:    &url.URL{Opaque: cm.targetAddr},
			Host:   cm.targetAddr,
			Header: hdr,
		}

		// If there's no done channel (no deadline or cancellation
		// from the caller possible), at least set some (long)
		// timeout here. This will make sure we don't block forever
		// and leak a goroutine if the connection stops replying
		// after the TCP connect.
		connectCtx := ctx
		if ctx.Done() == nil {
			newCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
			defer cancel()
			connectCtx = newCtx
		}

		didReadResponse := make(chan struct{}) // closed after CONNECT write+read is done or fails
		var (
			resp *Response
			err  error // write or read error
		)
		// 发送 https 代理 CONNECT 请求, 并读取响应
		go func() {
			defer close(didReadResponse)
			err = connectReq.Write(conn)
			if err != nil {
				return
			}
			// Okay to use and discard buffered reader here, because
			// TLS server will not speak until spoken to.
			br := bufio.NewReader(conn)
			resp, err = ReadResponse(br, connectReq)
		}()
		select {
		case <-connectCtx.Done():
			conn.Close()
			<-didReadResponse
			return nil, connectCtx.Err()
		case <-didReadResponse:
			// resp or err now set
		}
		if err != nil {
			conn.Close()
			return nil, err
		}
		if resp.StatusCode != 200 {
			f := strings.SplitN(resp.Status, " ", 2)
			conn.Close()
			if len(f) < 2 {
				return nil, errors.New("unknown status code")
			}
			return nil, errors.New(f[1])
		}
	}
    
    // 针对有代理的 https 请求
	if cm.proxyURL != nil && cm.targetScheme == "https" {
	    // cm.tlsHost(), 目标地址
		if err := pconn.addTLS(cm.tlsHost(), trace); err != nil {
			return nil, err
		}
	}
    
    // https 带有协商协议的
	if s := pconn.tlsState; s != nil && s.NegotiatedProtocolIsMutual && s.NegotiatedProtocol != "" {
		if next, ok := t.TLSNextProto[s.NegotiatedProtocol]; ok {
			alt := next(cm.targetAddr, pconn.conn.(*tls.Conn))
			if e, ok := alt.(http2erringRoundTripper); ok {
				// pconn.conn was closed by next (http2configureTransport.upgradeFn).
				return nil, e.err
			}
			return &persistConn{t: t, cacheKey: pconn.cacheKey, alt: alt}, nil
		}
	}

	pconn.br = bufio.NewReaderSize(pconn, t.readBufferSize())
	pconn.bw = bufio.NewWriterSize(persistConnWriter{pconn}, t.writeBufferSize())

	go pconn.readLoop()
	go pconn.writeLoop()
	return pconn, nil
}
```