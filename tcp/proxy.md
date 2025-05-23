## TCP 使用代理

目前本地代理主流是 HTTP 代理与 SOCKS5 代理, 如何在 TCP 请求当中使用这些代理呢? 

代理的本质就是建立 tunnel(需要告知代理端你访问的目标地址), 然后将数据包丢向 tunnel, 并从 tunnel 当中读取响应的数据包. 

HTTP 代理:

发起 `CONNECT请求`, 请求当中携带目标地址(host + port), 解析请求的响应(成功与否), 后续在这个 tunnel 上进行数据包交互.

SOCKS 代理:

发起 `SOCKS5 请求`(本质上就是一个 TCP 请求, 协议当中携带目标地址), 解析响应的数据包, 后续在这个 tunnel 上进行数据包交互.


### 具体实践

- HTTP

```
func Dial(network string, addr string, proxyUrl string) (net.Conn, error) {
	p, err := url.Parse(proxyUrl)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", p.Host)
	if err != nil {
		return nil, err
	}
	
	connectHeader := make(http.Header)
	if user := p.User; user != nil {
		proxyUser := user.Username()
		if proxyPassword, passwordSet := user.Password(); passwordSet {
			credential := base64.StdEncoding.EncodeToString([]byte(proxyUser + ":" + proxyPassword))
			connectHeader.Set("Proxy-Authorization", "Basic "+credential)
		}
	}

	connectReq := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: addr},
		Host:   addr,  // note: 请求目标地址
		Header: connectHeader,
	}

	if err := connectReq.Write(conn); err != nil {
		if err := conn.Close(); err != nil {
			log.Printf("httpProxyDialer: failed to close connection: %v", err)
		}
		return nil, err
	}

	// Read response. It's OK to use and discard buffered reader here becaue
	// the remote server does not speak until spoken to.
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, connectReq)
	if err != nil {
		if err := conn.Close(); err != nil {
			log.Printf("httpProxyDialer: failed to close connection: %v", err)
		}
		return nil, err
	}

	if resp.StatusCode != 200 {
		if err := conn.Close(); err != nil {
			log.Printf("httpProxyDialer: failed to close connection: %v", err)
		}
		f := strings.SplitN(resp.Status, " ", 2)
		return nil, errors.New(f[1])
	}
	return conn, nil
}
```

- SOCKS

```
// 基于 golang.org/x/net/proxy 
func Dial(network string, addr string, proxyUrl string) (net.Conn, error) {
    p, err := url.Parse(proxyUrl)
    if err != nil {
        return nil, err
    }
    
    var auth *proxy.Auth
    if user := p.User; user != nil {
        pwd, _ := user.Password()
        auth = &proxy.Auth{
            User: user.Username(),
            Password: pwd,
        }
    }
    
    socks, err := proxy.SOCKS5("tcp", proxyUrl, auth, nil)
    if err != nil {
        return nil, err
    }
    
    return socks.Dial(network, addr)
}
```

- 使用

```
# 不使用代理
conn, err := net.Dial("tcp4", "www.baidu.com")

# 使用代理, HTTP or SOCKS
conn, err := Dial("tcp4", "www.baidu.com", "http://127.0.0.1:1080")

conn, err := Dial("tcp4", "www.baidu.com", "socks5://127.0.0.1:1080")
```
