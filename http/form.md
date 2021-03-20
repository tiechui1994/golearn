## Go 当中的 form

### Form, PostForm 和 MultipartForm

http.Request 对象存在, `Form`, `PostForm`, `MultipartForm` 三个关于 form 相关的字段.

- Form 包含已解析的 `form data`, 包括 URL 字段的 "query参数" 和 `PATCH`, `POST` 或 `PUT` 的 Body 参数解析的
`form data`. 此字段仅在调用ParseForm之后可用. HTTP 客户端会忽略 Form 并改用 Body.

- PostForm 包含从 `PATCH`, `POST` 或 `PUT` 的 Body 参数解析的 `form data` 此字段仅在调用ParseForm之后可用. 
HTTP 客户端会忽略 PostForm 并改用 Body.

- MultipartForm 是已解析的 `multipart form`, 包括文件上传. 仅在调用ParseMultipartForm之后, 此字段才可用.
HTTP 客户端会忽略 MultipartForm 并改用 Body.


### Form, PostForm, MultipartForm 的产生

不论是是任何请求, 调用 `ParseForm()` 方法, 从 `http.Request` 的 `URL.RawQuery` 当中解析产生 `Form`.

当请求是 `POST`, `PUT`, `PATCH` 时, 且 `Content-Type` 是 `application/x-www-form-urlencoded` 时, 调用
`ParseForm()` 方法, 会从Body 当中去解析产生 `PostForm`, `Form`. 

当 `Content-Type` 是 `multipart/form-data` 时, 调用 `ParseMultipartForm()` 方法, 会从 Body 当中解析产生 
`MultipartForm`, `PostForm` 和 `Form`.

当请求是 `POST` 时, 且 `Content-Type` 是 `multipart/mixed`, 调用 `MultipartReader()` 方法, 可以将 Body 转
换为 `multipart.Reader`.
