package main

import (
	"net/http"
	"path/filepath"
	"crypto/tls"
	"io/ioutil"
	"crypto/x509"
	"time"
	"log"
)

var (
	client *http.Client
)

func NewClient() *http.Client {
	if client != nil {
		return client
	}

	dir, _ := filepath.Abs("./http/certs")

	// 加载客户端证书
	cliCert, err := tls.LoadX509KeyPair(dir+"/client/client.crt", dir+"/client/client.key.text")
	if err != nil {
		panic("try to load key & key err, " + err.Error())
	}

	// 加载CA证书
	caCrt, err := ioutil.ReadFile(dir + "/ca/ca.crt")
	if err != nil {
		panic("try to load ca err, " + err.Error())
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCrt)

	config := tls.Config{
		Certificates:       []tls.Certificate{cliCert},
		RootCAs:            caCertPool, // 校验服务端证书的 CA Pool
		InsecureSkipVerify: false,
	}
	config.BuildNameToCertificate()

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &config,
		},
		Timeout: 10 * time.Second,
	}

	return client
}

func main() {
	client := NewClient()
	response, err := client.Get("http://127.0.0.1:443/")
	if err != nil {
		log.Println("DO err", err)
		return
	}

	data, _ := ioutil.ReadAll(response.Body)
	log.Println("data", string(data))
}
