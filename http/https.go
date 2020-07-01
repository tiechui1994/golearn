package main

import (
	"net/http"
	"crypto/tls"
	"log"
	"io/ioutil"
	"crypto/x509"
	"io"
	"path/filepath"
	"os"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `{"status":0}`)
}

func Server() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HomeHandler)

	// 加载服务端证书
	dir, _ := filepath.Abs("./http/certs")
	log.Println("root", dir)
	srvCert, err := tls.LoadX509KeyPair(dir+"/server/server.crt", dir+"/server/server.key.text")
	if err != nil {
		log.Println("try to load key & crt err", err)
		return
	}

	// 加载CA, 添加进 caCertPool
	caCert, err := ioutil.ReadFile(dir + "/ca/ca.crt")
	if err != nil {
		log.Println("try to load ca err", err)
		return
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates:       []tls.Certificate{srvCert},
			ClientCAs:          caCertPool,                     // 专门校验客户端证书的 CA
			InsecureSkipVerify: false,                          // 必须校验
			ClientAuth:         tls.RequireAndVerifyClientCert, // 校验客户端证书
		},
		ErrorLog: log.New(os.Stdout, "", log.Lshortfile|log.Ldate|log.Ltime),
	}

	server.TLSConfig.BuildNameToCertificate() // 生成NameToCertificate

	server.ListenAndServeTLS("", "")
}

func main() {
	Server()
}
