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


### tls 流程 (ECDHE为例)

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
