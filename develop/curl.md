## curl 参数讲解

### 常用的参数

- 请求 Method

`-X, --request COMMAND` COMMAND 可以是 `GET`, `POST`, `PUT`, `DELETE`, `OPTION`

- 请求 Header

`-H, --header LINE` LINE 的格式是 `'key:value'`.

> 注意: 使用的是 `''` 包含的内容, 该参数可以出现多次.

`-D, --dump-header FILE`, 备份请求的Header 

- 请求 Cookie 

`-b, --cookie STRING/FILE` 从 STRING/FILE 当中读取cookie

`-c, --cookie-jar FILE` 将Cookie写入到 FILE 当中

> 设置 "--cookie" 和 "--cookie-jar" 就可以完成 Cookie 的读取和存储. 

- 请求 Body 

`-d, --data DATA`, POST 请求的 body 数据. 

`--data-raw DATA`, POST 请求的 body 数据.

`--data-binary DATA` POST 请求二进制数据.

> 注: 在默认状况下, `--data`, `--data-raw`, `--data-binary` 的 `Content-Type` 是 
`application/x-www-form-urlencoded`. 当然可以手动指定 `Content-Type` 的值.
> `--data`, `--data-raw`, `--data-binary` 可以使用字符串, 也可以使用特殊分隔符(参考下面的案例)

案例:

```
# 使用特殊的分隔符
curl -X POST https://reqbin.com/echo/post/json --data-raw @- <<DATA
{             
  "Id": 78912,
  "Customer": "Jason Sweet",
  "Quantity": 1,
  "Price": 18.00
}
DATA

# 使用字符串的形式
curl -X POST https://reqbin.com/echo/post/json --data-raw '{             
  "Id": 78912,
  "Customer": "Jason Sweet",
  "Quantity": 1,
  "Price": 18.00
}'
```

`--data-urlencode DATA` POST 请求使用 urlencode 编码, 格式为 `key1=value1&key2=value2`, 其中 key 和 value 
会进行 urlencode 编码. 在此种方式下, 其 `Content-Type` 值为 `application/x-www-form-urlencoded`.


`-F, --form CONTENT`, POST方式请求资源, CONTENT 的格式是 `key=value`, 该参数和 `-H` 一样, 可以出现多次. 对于
上传文件格式为 `-F file=@/path/to/file`, 即 value 使用 `@PATH` 的方式设置文件上传.
在此种方式下, 其 `Content-Type` 值为 `multipart/form-data; boundary=xxx`.

> 常见的 Content-Type:
> `application/x-www-form-urlencoded`, form表单数据被编码为key/value格式发送到服务器
> `multipart/form-data`, form 表单数据 (表单中进行文件上传)
> `text/plain`, 纯文本格式
> `text/html`, html 格式
> `text/xml`, xml 格式
> `application/json`, JSON 格式
> `application/octet-stream`, 二进制格式


- 响应内容

`-o, --output FILE` 将请求响应的内容存储到文件 FILE 当中. 默认状况, 请求响应内容直接打印在console上.


- 请求代理

`-x, --proxy [PROTOCOL://]HOST[:PORT]` 使用代理请求. 例如 `--proxy http://proxy.com:8888`, 表示使用 HTTP
代理. 

> HTTPS 转发代理:
> 1. Client 与 Proxy 建立 TCP 连接.
> 1. Client 发送 CONNECT 请求到 Proxy.
> 2. Proxy 与 Server 建立 TCP 连接后, 响应 Client `200 Connection established`, 此时就打开了通道.
> 3. Client 发送将加密数据到 Proxy, Proxy 将数据转发给 Server. (这个过程当中, 数据是加密的, 代理是无法知道数据的)
>
> 上述的过程就是一个转发代理. Proxy 完全就是一个中间人. Nginx 的代理就是上述的过程.

> HTTPS 中间人代理:
> 1. Client 与 Proxy 建立 TCP 连接.
> 1. Client 发送 CONNECT 请求到 Proxy.
> 2. Proxy 响应 Client `200 Connection established`. 
> 3. Client 与 Proxy 开始进行 TLS 握手, client hello.
> 4. Proxy 用生成一个假证书(这个证书顶替真正的服务器证书), 完成 server hello. (此时 client 与 proxy 建立的是 
> https连接, Proxy 完全看得见 HTTPS 内容)
> 5. Proxy 临时和 Server 建立 TCP 握手, 并完成 TLS 握手.
> 6. 数据传输. 注: 这个过程都是加密的, 由于 Client 与 Proxy 是 HTTPS 连接, 因此对于 HTTPS 而言, 数据完全是透明的,
> Proxy 只要将数据加密传输给 Server 即可, 返回的数据也是一样的.
>
> Proxy 要与 Client 完成 TLS 握手, 那么对于 Client 端需要内置 Proxy 的 RootCA 证书, 用于验证 Proxy 签发的假证
书. 对于手机而言, 就需要将 Proxy 的 RootCA 证书进行信任.

`--cacert FILE` 设置信任私有的内置 CA 证书. 

`--capath DIR` 设置信任私有的内置 CA 证书路径.

`--insecure` 表示客户端不需要验证服务器的证书(也就是说任何证书都可以使用)

> 当进行中间人代理 HTTPS 的时候, curl 是需要 `--cacert` 或 `--insecure` 其中之一的参数. 


### HTTPS 相关的参数

- SSL 连接参数

1) TLS/SSL 版本 
```
-2, --sslv2 使用SSLv2
-3, --sslv3 使用SSLv3

-l, --tlsv1 TLSv1 以上的版本
--tlsv1.0 使用 TLSv1.0
--tlsv1.1 使用 TLSv1.1
--tlsv1.2 使用 TLSv1.2
--tlsv1.3 使用 TLSv1.3
```
   
2) HTTP 版本
``` 
-0, --http1.0 使用 HTTP 1.0
--http1.1 使用 HTTP 1.1
--http2   使用 HTTP 2 
```

> 目前经常使用到是 `HTTP 1.1` 和 `HTTP 1.0`

3) `--ciphers LIST` 设置加密算法列表(这些算法是优先考虑的)

4) `-K, --insecure` 信任SSL连接的任何站点的证书.


- 客户端证书(双向认证, 用于服务器认证客户端的过程中, 该证书是私有 CA 证书签发的)

`--cert-status` 只是进行服务器证书验证, 不进行任何请求.

`-E, --cert CERT[:PASSWD]` 客户端证书和密码

`--cert-type TYPE` 设置证书文件的类型 (DER/PEM/ENG). 


- 客户端秘钥(双向认证, 用于服务器认证客户端的过程中.) 

`--key KEY` 私有证书

`--key-type TYPE` 私有证书类型(DER/PEM/ENG)


- CA 证书(用于HTTPS中间人代理, 信任自签发的私有证书)

`--cacert FILE` CA证书

`--capath DIR` CA证书路径


### HTTP 请求调试参数

`--trace FILE` 将请求调用过程打印到一个文件当中, 可以详细查看一个请求的过程.

`--trace-time` trace 文件当中增加时间.

`-v, -vv, --verbose` 打印请求的详细过程(TLS协商, 请求详细内容, 响应详细内容)

