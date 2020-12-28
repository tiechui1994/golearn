## 证书流程

先弄清楚几个概念: CER, CRT, CA, CSR

CER, CRT, 这两个表示的概念是一样的, 都表示证书. https请求的时候要获取证书, 就是获取的是这个东西.

CSR, 证书签名请求. 服务器要进行 https 通信, 就需要证书, 证书是怎么来的呢? 首先你得生成一个私钥, 然后通过私钥生成一个证
书签名请求(里面主要包含了你的域名, 公司, 公钥等一些信息), 然后你把这个证书签名请求发给第三方颁发证书的机构, 第三方颁发证
书的机构对你的证书签名请求进行签名(简单来说, 通过它的私钥给你加密一下, 然后生成一串hash), 生成一个证书, 然后发给你. 然
后你在服务器上进行配置, 然后你就可以愉快的进行https通信了.  

CA, 第三方颁发证书的机构. 服务器要进行 https 通信, 就需要它颁发的证书.

1. 生成 CA

```bash
# ca private key
openssl genrsa -des3 -out ca.key 4096 

# X.509 Certificate Signing Request (CSR) Management. CA 自己的证书(自签)
openssl req -new -x509 -days 365 -key ca.key -out ca.crt
```

> 注: 自己生成的 CA 是没有权威性, 也就是自个玩而已.

2. 生成服务器私钥

```bash
# private key (带有加密密码)
openssl genrsa -des3 -out server.key 1024

# private key (不带加密密码)
openssl rsa -in server.key -out server.key.text
```

3. 生成服务器证书的 CSR

```bash
# csr, 服务器证书签名请求
# Common Name, 服务器的名称
openssl req -new -key server.key -out server.csr
```

> 注: 这里的服务器的名称要认真填写.

4. CA 签发服务器证书

```bash
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt
```

5. 生成客户端秘钥

```bash
openssl genrsa -des3 -out client.key 1024
openssl rsa -in client.key -out client.key.text
```

6. 生成客户端 CSR

```bash
openssl req -new -key client.key -out client.csr
```

7. CA 签发客户端证书

```bash
openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out client.crt
```

### 双向认证实践, 这里主要介绍 TLSConfig 的配置

// 客户端

```
caCert, _ := ioutil.ReadFile("./ca.crt")
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)
	
cliCert, _ := tls.LoadX509KeyPair("./client.crt","./client.key")

config := &tls.Config{
    Certificates:       []tls.Certificate{cliCert}, // 客户端证书, 双向认证必须携带
    RootCAs:            caCertPool,                 // 校验服务端证书 [CA证书]
    InsecureSkipVerify: false,                      // 不用校验服务器证书
}

config.BuildNameToCertificate()
```

> ca.crt 是CA的证书, 自定义的CA签发
> client.crt 是客户端的证书, 自定义的CA签发
> client.key 是客户端的私钥
> 
> 注: 实际的开发当中, ca.crt 是第三方CA的证书, 已经内置在系统, 已经不需要. client.crt 是自定义CA签发的证书, client.key
> 是客户端的私钥

// 服务端

```
// 加载CA, 添加进 caCertPool
caCert, _ := ioutil.ReadFile("./ca.crt")
	
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

// 加载服务端证书
srvCert, _ := tls.LoadX509KeyPair("./server.crt", "./server.key")

config := &tls.Config{
    Certificates:       []tls.Certificate{srvCert},     // 服务器证书
    ClientCAs:          caCertPool,                     // 校验客户端的证书 [CA证书]
    ClientAuth:         tls.RequireAndVerifyClientCert, // 校验客户端证书
    InsecureSkipVerify: false,                          // 必须校验 tls

    CipherSuites: []uint16{
        tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
    },
}

config.BuildNameToCertificate()
```

> ca.crt 是CA的证书, 自定义的CA签发
> server.crt 是服务端的证书, 自定义的CA签发
> server.key 是服务端的私钥, 服务器自己生成的.
> 
> 注: 实际生产环境当中, server.crt 是由第三方 CA 进行签名的. ca.crt 还是自定义CA签发的CA证书. server.key 是服务
> 器的私钥.

如果使用 nginx 进行做代理, 配置应该是这样的:

```
server {
  server_name localhost;
  listen 443 ssl http2;
  
  # CERT, 源自第三方签名
  ssl_certificate ./server.crt;
  ssl_certificate_key ./server.key;
  
  # CA, 自定义
  ssl_client_certificate ./ca.crt;
  ssl_verify_client on;
  
  # ssl ciphers
  ssl_protocols TLSv1.1 TLSv1.2;
  ssl_prefer_server_ciphers on;
}
```

对于上 `server.crt`, `server.key` 和 `ca.crt` 的含义与上面的Server的是一致的, 并且相关的配置也是类似的.

更为详细的案例, 参考代码 `client.go` 和 `server.go`.

### tls 流程(单向) (ECDHE为例)

整体流程:

![image](/images/https_process.png)

- Client Hello (client -> server)

![image](/images/https_clienthello.png)


- Server Hello (server -> client)

![image](/images/https_serverhello.png)


- Certificate, Server Key Exchange, Server Hello Done (server -> client)

![image](/images/https_certs.png)

- Client Key Exchange, Change Cipher Spec, Encrypted Handshake Messsage (client -> server)

![images](/images/https_clientkey.png)

- New Session Ticket, Change Cipher Spec, Encrypted Handshake Message (server -> client)

![image](/images/https_sessionticket.png)
