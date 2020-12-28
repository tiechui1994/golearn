## 证书流程

1. 生成 CA

```bash
# ca private key
openssl genrsa -des3 -out ca.key 4096 

# X.509 Certificate Signing Request (CSR) Management. CA 自己的证书(自签)
openssl req -new -x509 -days 365 -key ca.key -out ca.crt
```

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

4. 通过 CSR 向 CA 签发服务端证书

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

7. 通过 CSR 向 CA 签发客户端证书

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
