package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"golearn/develop/proxy/goproxy"
)

type Cache struct {
	m sync.Map
}

func (c *Cache) Set(host string, cert *tls.Certificate) {
	c.m.Store(host, cert)
}

func (c *Cache) Get(host string) *tls.Certificate {
	v, ok := c.m.Load(host)
	if !ok {
		return nil
	}

	return v.(*tls.Certificate)
}

type Intercept struct {
	goproxy.DefaultDelegate
}

func (h *Intercept) BeforeRequest(ctx *goproxy.Context) {
	data, err := httputil.DumpRequest(ctx.Req, true)
	if err != nil {
		log.Println("DumpRequest", err)
		return
	}

	ctx.Data["request"] = data

	log.Println("Method", ctx.Req.Method)
	log.Println("Scheme", ctx.Req.URL.Scheme)
}

func (h *Intercept) BeforeResponse(ctx *goproxy.Context, response *http.Response, err error) {
	data, err := httputil.DumpResponse(response, true)
	if err != nil {
		log.Println("DumpResponse", err)
		return
	}

	log.Printf("request:\n%s\n", string(ctx.Data["request"].([]byte)))
	log.Printf("response:\n%s\n\n", string(data))
}

/*
curl --proxy http://localhost:1433 --cacert ca.cert https://www.baidu.com
*/

func main() {
	p := flag.Int("p",80, "port")
	flag.Parse()
	/*
	// DecryptHTTPS and Delegate
	proxy := goproxy.New(goproxy.WithDecryptHTTPS(&Cache{}),
		goproxy.WithDelegate(&Intercept{}))
	*/

	proxy := goproxy.New(goproxy.WithDisableKeepAlive(true))
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *p),
		Handler:      proxy,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}
	log.Printf("Start Server [%s]", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
