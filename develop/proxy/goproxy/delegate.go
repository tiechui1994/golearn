package goproxy

import (
	"log"
	"net/http"
	"net/url"
)

type Context struct {
	Req   *http.Request
	Data  map[interface{}]interface{}
	abort bool
}

// Abort, break request
func (c *Context) Abort() {
	c.abort = true
}

func (c *Context) IsAborted() bool {
	return c.abort
}

type Delegate interface {
	// Connect 收到客户端连接
	Connect(ctx *Context, rw http.ResponseWriter)
	// Auth proxy authentication
	Auth(ctx *Context, rw http.ResponseWriter)
	// BeforeRequest before send request to target, can set "Header", "Body" etc
	BeforeRequest(ctx *Context)
	// BeforeResponse before send response to client, 修改 "Header", "Body", "Status Code" etc
	BeforeResponse(ctx *Context, resp *http.Response, err error)
	// ParentProxy parent proxy, example: local proxy
	ParentProxy(*http.Request) (*url.URL, error)
	// Finish finish once request
	Finish(ctx *Context)
	ErrorLog(err error)
}

var _ Delegate = &DefaultDelegate{}

// DefaultDelegate, do nothing
type DefaultDelegate struct {
	Delegate
}

func (h *DefaultDelegate) Connect(ctx *Context, rw http.ResponseWriter) {}

func (h *DefaultDelegate) Auth(ctx *Context, rw http.ResponseWriter) {}

func (h *DefaultDelegate) BeforeRequest(ctx *Context) {}

func (h *DefaultDelegate) BeforeResponse(ctx *Context, resp *http.Response, err error) {}

func (h *DefaultDelegate) ParentProxy(req *http.Request) (*url.URL, error) {
	return http.ProxyFromEnvironment(req)
}

func (h *DefaultDelegate) Finish(ctx *Context) {}

func (h *DefaultDelegate) ErrorLog(err error) {
	log.Println(err)
}
