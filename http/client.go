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

func init() {
	InitClient()
}

func InitClient() *http.Client {
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
		Certificates:       []tls.Certificate{cliCert}, // 客户端证书, 双向认证必须携带
		RootCAs:            caCertPool,                 // 校验服务端证书 [CA证书]
		InsecureSkipVerify: false,                      // 不用校验服务器证书
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

/*
curl -X GET \
	 --cert certs/client/client.crt --key certs/client/client.key.text \
	 --cacert certs/ca/ca.crt \
	 'https://localhost/'


*   Trying 127.0.0.1...
* Connected to localhost (127.0.0.1) port 443 (#0)
* found 1 certificates in certs/ca/ca.crt
* found 597 certificates in /etc/ssl/certs
* ALPN, offering http/1.1
* SSL connection using TLS1.2 / ECDHE_RSA_AES_128_GCM_SHA256
* 	 server certificate verification OK
* 	 server certificate status verification SKIPPED
* 	 common name: localhost (matched)
* 	 server certificate expiration date OK
* 	 server certificate activation date OK
* 	 certificate public key: RSA
* 	 certificate version: #1
* 	 subject: C=CN,ST=ZheJiang,L=Hangzhou,O=Broadlink,OU=Broadlink,CN=localhost
* 	 start date: Wed, 01 Jul 2020 07:03:40 GMT
* 	 expire date: Thu, 01 Jul 2021 07:03:40 GMT
* 	 issuer: C=CN,ST=ZheJiang,L=Hangzhou,O=Broadlink,OU=Broadlink,CN=master
* 	 compression: NULL
* ALPN, server accepted to use http/1.1
> GET / HTTP/1.1
> Host: localhost
> User-Agent: curl/7.47.0
>
< HTTP/1.1 200 OK
< Date: Wed, 01 Jul 2020 08:41:20 GMT
< Content-Length: 12
< Content-Type: text/plain; charset=utf-8
<
* Connection #0 to host localhost left intact
{"status":0}

**/

func ClientRequest() {
	request, _ := http.NewRequest("GET", "https://localhost/", nil)
	// 这里必须是 localhost, 因为证书当中的 `Common Name(服务器的名称)` 是 localhost
	response, err := client.Do(request)
	if err != nil {
		log.Println("DO err", err)
		return
	}

	data, _ := ioutil.ReadAll(response.Body)
	log.Println("data", string(data))
}

func main() {
	ClientRequest()
}
