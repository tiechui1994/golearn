# openssl 使用

## RSA

- RSA 密钥生成

```bash
openssl genrsa -out private.pem 1024
```

常用的密钥加密选项:

① `-des`, 使用 DES CBC 模式加密密钥

② `-des3`,  使用 DES CBC 模式加密密钥 (168位的key)

③ `-aes128`, `-aes192`, `-aes256`, 使用 AES CBC


- RSA 公钥分离

```bash
openssl rsa -in private.pem -pubout -out public.pem
```

> 注: 不管是否加密密钥, 都使用上述方法分离公钥

- 签名与验证

```bash
# 签名
openssl rsautl -sign -inkey private.pem -in msg.txt -out enc.txt

# 恢复
openssl rsautl -verify -inkey private.pem -in enc.txt -out plian.txt
```

> 签名与验证都使用的是私钥. 


- 加密与解密

```bash
# 加密
openssl rsautl -encrypt -pubin  -inkey public.pem -in msg.txt -out enc.txt
openssl rsautl -encrypt -certin -inkey cert.pem -in msg.txt -out enc.txt

# 解密
openssl rsautl -decrypt -inkey private.pem -in enc.txt -out plain.txt
```

> 注: 加密使用的是公钥或证书, 解密使用的是私钥
> 
> `-pubin` 是使用公钥加密, `-certin` 是使用证书加密. 两者本质上是一样, 都是使用公钥, 因为证书当中是包含公钥的.

加密的其他选项:

① `-ssl`, 使用 SSL v2 padding

② `-pkcs`, 使用 PKCS#1 v1.5 padding (默认)

③ `-oeap`, 使用 PKCS#1 OAEP padding

④ `-raw`, no padding

