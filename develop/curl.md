## curl 参数讲解

- 请求 Method

`-X, --request COMMAND` COMMAND 可以是 `GET`, `POST`, `PUT`, `DELETE`, `OPTION`

- 请求 Header

`-H, --header LINE` LINE 的格式是 `'key:value'`.

> 注意: 使用的是 `''` 包含的内容, 该参数可以出现多次.

`-D, --dump-header FILE`, 备份请求的Header 

- 请求 Body 

`-d, --data DATA`, POST 请求的 body 数据. 

`--data-raw DATA`, POST 请求的 body 数据.

> 注: 默认状况下, `-d` 和 `--data-raw` 的 `Content-Type` 是 `application/x-www-form-urlencoded`. 可以手
动指定 Header


`--data-binary DATA` POST 请求二进制数据.


`--data-urlencode DATA` POST 请求使用 urlencode 编码, 格式为 `key1=value1&key2=value2`, 其中 key 和 value 
会进行 urlencode 编码. 在此种方式下, 其 `Content-Type` 值为 `application/x-www-form-urlencoded`.


`-F, --form CONTENT`, POST方式请求资源, CONTENT 的格式是 `key=value`, 该参数和 `-H` 一样, 可以出现多次. 对于
上传文件格式为 `-F file=@/path/to/file`, 即 value 使用 `@PATH` 的方式设置文件上传.
在此种方式下, 其 `Content-Type` 值为 `multipart/form-data; boundary=-----xxxx`.


- 响应内容

`-o, --output FILE` 将请求响应的内容存储到文件 FILE 当中. 默认状况, 请求响应内容直接打印在console上.


- 请求代理

`-x, --proxy [PROTOCOL://]HOST[:PORT]` 使用代理请求. 例如 `--proxy http://proxy.com:8888`, 表示使用 HTTP
代理. 

`-ca`

