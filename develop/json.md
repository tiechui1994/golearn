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


- 使用字符串传递数字

```
type Object struct {
    Num int `json:"num,string"`
}

这个对应的json是 {"num":"100"}, 如果json是 {"num":100} 则会报错
```


- 使用 MarshalJSON 支持 time.Time ^^

go 默认会把 `time.Time` 用字符串方式序列化. 如果想用其他方式表示 `time.Time`, 需要自定义类型并
定义 `MarshalJSON`

```
type timeImplMarshaler time.Time

func (o timeImplMarshaler) MarshalJSON() ([]byte, err) {
    seconds := time.Time(o).Unix()
    return []byte(strconv.FormatInt(seconds, 10)), nil
}
```


- 使用 RegisterTypeEncoder 支持 time.Time ^^

jsoniter 能够对不是自定义的 type 定义 JSON 编解码方式. 比如对于 `time.Time` 可以使用 epoch int64 
序列化.


```
import "github.com/json-iterator/go/extra"

extra.RegisterTimeAsInt64Codec(time.Microsecond)
ouput, err := jsoniter.Marshal(time.Unix(1,1002))
should.Equal("1000001", string(output))
```


- 使用 `MarshalText`, `UnmarshalText` 支持非字符串作为key的map

虽然 JSON 标准里只支持 `string` 作为 `key` 的 `map`. 但是 go 通过 `MarshalText()` 接口, 使得
其他类型也可以作为 `map` 的 `key`.

```
type Key struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// 必须是传值调用(编码)
// 函数当中不能调用 json.Marshal() 函数
func (k Key) MarshalText() ([]byte, error) {
	data := fmt.Sprintf("%v_%v", k.Name, k.Value)
	return []byte(data), nil
}

// 必须是传指针调用(解码), 否则解析的值是空. 
// 函数当中不能调用 json.Unmarshal() 函数
func (k *Key) UnmarshalText(text []byte) (error) {
	tokens := strings.Split(string(text), "_")
	if len(tokens) == 2 {
		k.Name = tokens[0]
		k.Value = tokens[1]
		return nil
	}
	return errors.New("invalid text")
}

func main() {
	val := map[Key]string{{"12", "22"}: "2"}
	data, err := json.Marshal(val)
	fmt.Println(string(data), err)

	var vv map[Key]string
	json.Unmarshal(data, &vv)
	fmt.Println(vv)
}

```


- 使用 json.RawMessage

如果部分 json 文档没有标准格式, 可以把原始的信息使用 `[]byte` 保存下来.

```
type RawObject struct {
	Key string
	Raw json.RawMessage
}

func main() {
	var data RawObject
	json.Unmarshal([]byte(`{"key":"111", "raw":[1,2,3]}`), &data)
	fmt.Println(string(data.Raw)) // [1,2,3]
}
```

- 使用 json.Number

默认情况下, 如果是 `interface` 对应数字的情况会是 `float64` 类型的. 如果输入的数字比较大, 这个会有损精度. 
可以使用 `UseNumber()` 启用 `json.Number` 来用字符串表示数字.


- 统一更改字段的命名风格

① 经常 JSON 里的字段名和 Go 里的字段名是不一样的. 可以使用 `field tag` 来修改.

② 但是一个个字段来设置, 太过于麻烦, 如果使用 jsoniter 可以统一设置命名风格.

> 按照 ①, ② 的方式去选择 JSON 的字段名. 如果方式 ① 的字段名存在, 则使用方式 ①, 否则再使用方式 ②

```
type T struct {
	UserName      string
	FirstLanguage string `json:"language"`
}

extra.SetNamingStrategy(func(s string) string {
    return strings.ToLower(s)
})
d, _ := jsoniter.Marshal(T{})
fmt.Println(string(d)) // {"username":"","language":""}
```

- 支持私有的字段

Go 标准库只支持 public 的 field. jsoniter 额外支持了 private 的 field. 需要使用 `SupportPrivateFields()`
来开启.

```
type T struct {
	UserName      string
	FirstLanguage string `json:"language"`
	private       string
}

extra.SetNamingStrategy(func(s string) string {
    return strings.ToLower(s)
})
extra.SupportPrivateFields()

var val T
jsoniter.Unmarshal([]byte(`{"username":"", "language":"en", "private":"private"}`), &val)
```
