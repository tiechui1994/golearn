package main

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `{"status":0}`)
}

func Server() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HomeHandler)

	dir, _ := filepath.Abs("./http/certs")
	log.Println("root", dir)

	// 加载CA, 添加进 caCertPool
	caCert, err := ioutil.ReadFile(dir + "/ca/ca.crt")
	if err != nil {
		log.Println("try to load ca err", err)
		return
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// 加载服务端证书(生产环境当中, 证书是第三方进行签名的, 而非自定义CA)
	srvCert, err := tls.LoadX509KeyPair(dir+"/server/server.crt", dir+"/server/server.key.text")
	if err != nil {
		log.Println("try to load key & crt err", err)
		return
	}

	config := &tls.Config{
		Certificates:       []tls.Certificate{srvCert},     // 服务器证书
		ClientCAs:          caCertPool,                     // 专门校验客户端的证书 [CA证书]
		InsecureSkipVerify: true,                           // 必须校验
		ClientAuth:         tls.RequireAndVerifyClientCert, // 校验客户端证书

		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	config.BuildNameToCertificate()

	server := &http.Server{
		Addr:    ":1443",
		Handler: mux,
		//TLSConfig: config,
		ErrorLog: log.New(os.Stdout, "", log.Lshortfile|log.Ldate|log.Ltime),
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Println("ListenAndServeTLS err", err)
		return
	}
}

func main() {
	Server()
}
