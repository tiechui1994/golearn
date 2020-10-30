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
			Certificates:       []tls.Certificate{srvCert},     // 服务器证书
			ClientCAs:          caCertPool,                     // 专门校验客户端的证书 [CA证书]
			InsecureSkipVerify: false,                          // 必须校验
			ClientAuth:         tls.RequireAndVerifyClientCert, // 校验客户端证书

			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
		},
		ErrorLog: log.New(os.Stdout, "", log.Lshortfile|log.Ldate|log.Ltime),
	}

	server.TLSConfig.BuildNameToCertificate() // 生成NameToCertificate

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Println("ListenAndServeTLS err", err)
		return
	}
}

func main() {
	Server()
}
