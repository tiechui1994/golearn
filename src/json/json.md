# json 

- 忽略 struct 空字段

> 忽略 struct 空字段, 使用 `omitempty`

```cgo
type User struct{
    Email string `json:"email,omitempty"`
    Pwd  string `json:"pwd,omitempty"`
}
```

- 临时粘合两个struct

```
type BlogPost struct{
    URL string `json:"url"`
}

type Analytics struct{
    Visitors int `json:"visitors"`
}

json.Marshal(struct{
    *BlogPost
    *Analytics
}{post, analytics})
```

- 临时切开两个struct

```
json.Unmarshal([]byte(`{
    "url":"www@163.com",
    "visitors":10
}`), struct{
   *BlogPost
   *Analytics
}{&post, &analytics})
```


- 临时改名 struct 的字段

```
type Cache struct {
    Key    string `json:"key"`
    MaxAge int    `json:"cacheAge"`
}

json.Marshal(struct {
    *Cache
    OmitMaxAge int `json:"cacheAge,omitempty"` // remove bad key
    MaxAge     int `json:"max_age"` // add new key
}{cache, 0, 10})
```