package goproxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

const (
	defaultTargetConnectTimeout   = 10 * time.Second
	defaultTargetReadWriteTimeout = 30 * time.Second
	defaultClientReadWriteTimeout = 30 * time.Second
)

var tunnelResponseLine = []byte("HTTP/1.1 200 Connection established\r\n\r\n")
var tunnelRequestLine = func(addr string) []byte {
	return []byte(fmt.Sprintf("CONNECT %s HTTP/1.1\r\n\r\n", addr))
}

var BadGateway = []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", http.StatusBadGateway,
	http.StatusText(http.StatusBadGateway)))

type options struct {
	disableKeepAlive bool
	delegate         Delegate
	decryptHTTPS     bool
	certCache        Cache
	transport        *http.Transport
}

type Option func(*options)

func WithDisableKeepAlive(disableKeepAlive bool) Option {
	return func(opt *options) {
		opt.disableKeepAlive = disableKeepAlive
	}
}

func WithDelegate(delegate Delegate) Option {
	return func(opt *options) {
		opt.delegate = delegate
	}
}

func WithTransport(t *http.Transport) Option {
	return func(opt *options) {
		opt.transport = t
	}
}

func WithDecryptHTTPS(c Cache) Option {
	return func(opt *options) {
		opt.decryptHTTPS = true
		opt.certCache = c
	}
}

func New(opt ...Option) *Proxy {
	opts := &options{}
	for _, o := range opt {
		o(opts)
	}
	if opts.delegate == nil {
		opts.delegate = &DefaultDelegate{}
	}
	if opts.transport == nil {
		opts.transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   15 * time.Second,
			ExpectContinueTimeout: 15 * time.Second,
		}
	}

	p := &Proxy{}
	p.delegate = opts.delegate
	p.decryptHTTPS = opts.decryptHTTPS
	if p.decryptHTTPS {
		p.cert = NewCertificate(opts.certCache)
	}
	p.transport = opts.transport
	p.transport.DisableKeepAlives = opts.disableKeepAlive
	p.transport.Proxy = p.delegate.ParentProxy

	return p
}

// Proxy, server handler
type Proxy struct {
	delegate      Delegate
	clientConnNum int32
	decryptHTTPS  bool
	cert          *Certificate
	transport     *http.Transport
}

var _ http.Handler = &Proxy{}

func (p *Proxy) ClientConnNum() int32 {
	return atomic.LoadInt32(&p.clientConnNum)
}

func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}
	atomic.AddInt32(&p.clientConnNum, 1)
	defer func() {
		atomic.AddInt32(&p.clientConnNum, -1)
	}()
	ctx := &Context{
		Req:  req,
		Data: make(map[interface{}]interface{}),
	}
	defer p.delegate.Finish(ctx)
	p.delegate.Connect(ctx, rw)
	if ctx.abort {
		return
	}
	p.delegate.Auth(ctx, rw)
	if ctx.abort {
		return
	}
	fmt.Println(ctx.Req.Method, ctx.Req.Header)
	switch {
	case ctx.Req.Method == http.MethodConnect && p.decryptHTTPS:
		p.forwardHTTPS(ctx, rw)
	case ctx.Req.Method == http.MethodConnect:
		p.forwardTunnel(ctx, rw)
	default:
		p.forwardHTTP(ctx, rw)
	}
}

// DoRequest, send http(s) request, and call callback
func (p *Proxy) DoRequest(ctx *Context, responseFunc func(*http.Response, error)) {
	if ctx.Data == nil {
		ctx.Data = make(map[interface{}]interface{})
	}
	p.delegate.BeforeRequest(ctx) // callback
	if ctx.abort {
		return
	}

	newReq := new(http.Request)
	*newReq = *ctx.Req
	newReq.Header = CloneHeader(newReq.Header)
	removeConnectionHeaders(newReq.Header)
	for _, item := range hopHeaders {
		if newReq.Header.Get(item) != "" {
			newReq.Header.Del(item)
		}
	}
	resp, err := p.transport.RoundTrip(newReq)

	p.delegate.BeforeResponse(ctx, resp, err) // callback
	if ctx.abort {
		return
	}

	if err == nil {
		removeConnectionHeaders(resp.Header)
		for _, h := range hopHeaders {
			resp.Header.Del(h)
		}
	}

	// do other work
	responseFunc(resp, err)
}

// HTTP
func (p *Proxy) forwardHTTP(ctx *Context, rw http.ResponseWriter) {
	ctx.Req.URL.Scheme = "http"
	p.DoRequest(ctx, func(resp *http.Response, err error) {
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTP request fail: %s", ctx.Req.URL, err))
			rw.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		CopyHeader(rw.Header(), resp.Header)
		rw.WriteHeader(resp.StatusCode)
		io.Copy(rw, resp.Body)
	})
}

// HTTPS
func (p *Proxy) forwardHTTPS(ctx *Context, rw http.ResponseWriter) {
	clientConn, err := hijacker(rw)
	if err != nil {
		p.delegate.ErrorLog(err)
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer clientConn.Close()
	// response CONNECT to client
	_, err = clientConn.Write(tunnelResponseLine)
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS, connect fail: %s", ctx.Req.URL.Host, err))
		return
	}

	// cert and tls conn
	tlsConfig, err := p.cert.GenerateTlsConfig(ctx.Req.URL.Host)
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS, generate cert fail: %s", ctx.Req.URL.Host, err))
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	// tls handshark
	tlsClientConn := tls.Server(clientConn, tlsConfig)
	tlsClientConn.SetDeadline(time.Now().Add(defaultClientReadWriteTimeout))
	defer tlsClientConn.Close()

	err = tlsClientConn.Handshake()
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS, handshark fail: %s", ctx.Req.URL.Host, err))
		return
	}

	// request
	buf := bufio.NewReader(tlsClientConn)
	tlsReq, err := http.ReadRequest(buf)
	if err != nil {
		if err != io.EOF {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS, read request fail: %s", ctx.Req.URL.Host, err))
		}
		return
	}

	tlsReq.RemoteAddr = ctx.Req.RemoteAddr
	tlsReq.URL.Scheme = "https"
	tlsReq.URL.Host = tlsReq.Host

	// send
	ctx.Req = tlsReq
	p.DoRequest(ctx, func(resp *http.Response, err error) {
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS, request fail: %s", ctx.Req.URL, err))
			tlsClientConn.Write(BadGateway)
		}

		err = resp.Write(tlsClientConn)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS, response write fail: %s", ctx.Req.URL, err))
		}

		resp.Body.Close()
	})
}

// Tunel
func (p *Proxy) forwardTunnel(ctx *Context, rw http.ResponseWriter) {
	clientConn, err := hijacker(rw)
	if err != nil {
		p.delegate.ErrorLog(err)
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer clientConn.Close()

	// parent proxy
	parentProxyURL, err := p.delegate.ParentProxy(ctx.Req)
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - Tunnel, parse parent proxy: %s", ctx.Req.URL.Host, err))
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	// tcp conn, if exist parentProxyURL forward to proxy host, otherwise forward to target host
	targetAddr := ctx.Req.URL.Host
	if parentProxyURL != nil {
		targetAddr = parentProxyURL.Host
	}
	targetConn, err := net.DialTimeout("tcp", targetAddr, defaultTargetConnectTimeout)
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - Tunnel, dail target: %s", ctx.Req.URL.Host, err))
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer targetConn.Close()
	clientConn.SetDeadline(time.Now().Add(defaultClientReadWriteTimeout))
	targetConn.SetDeadline(time.Now().Add(defaultTargetReadWriteTimeout))

	// if parent proxy exist, CONNECT to parent proxy,
	// otherwise response client connect success to establish.
	if parentProxyURL == nil {
		_, err = clientConn.Write(tunnelResponseLine)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - Tunel connect success, write fail: %s", ctx.Req.URL.Host, err))
			return
		}
	} else {
		targetConn.Write(tunnelRequestLine(ctx.Req.URL.Host)) // send CONNECT to proxy
	}

	// traffic forwarding
	p.transfer(clientConn, targetConn)
}

// src and dst transfer
func (p *Proxy) transfer(src net.Conn, dst net.Conn) {
	go func() {
		io.Copy(src, dst)
		src.Close()
		dst.Close()
	}()

	io.Copy(dst, src)
	dst.Close()
	src.Close()
}

// 获取底层连接
func hijacker(rw http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := rw.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("web server not support Hijacker")
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		return nil, fmt.Errorf("hijacker faield: %s", err)
	}

	return conn, nil
}

func CopyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// deepin copy header
func CloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

// deepin copy body
func CloneBody(b io.ReadCloser) (r io.ReadCloser, body []byte, err error) {
	if b == nil {
		return http.NoBody, nil, nil
	}
	body, err = ioutil.ReadAll(b)
	if err != nil {
		return http.NoBody, nil, err
	}
	r = ioutil.NopCloser(bytes.NewReader(body))

	return r, body, nil
}

var hopHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

func removeConnectionHeaders(h http.Header) {
	if c := h.Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			if f = strings.TrimSpace(f); f != "" {
				h.Del(f)
			}
		}
	}
}
