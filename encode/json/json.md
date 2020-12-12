## 一次 json 解码引起的思考


这是一个简答 json 反序列化的过程, 大家想想最终的输出结果是什么?

```cgo
type Student struct {
	Child []int `json:"child"`
}

func main() {
	data := `{"child":[1,2,3]}`
	var a Student
	json.Unmarshal([]byte(data), &a)
	aa := a.Child
	fmt.Println(aa)

	data = `{"child":[3,4,5,6,7,8]}`
	json.Unmarshal([]byte(data), &a)
	fmt.Println(aa)
}
```

## Unmarshal 

`Unmarshal` 方法是 JSON 解析, 将一个JSON字符串转换成为一个对象. 这个方法的过程很简单, 就两步操作, 第一步是检查 JSON 
字符串的合法性, 第二步是将 JSON 字符串解析成为一个`指针对象`.

```cgo
func Unmarshal(data []byte, v interface{}) error {
	var d decodeState
	err := checkValid(data, &d.scan) // 检查 JSON 合法性
	if err != nil {
		return err
	}

	d.init(data)
	return d.unmarshal(v) // 解析
}
```

检查 JSON 字符串的合法性, 其实在 `json` 包当中专门提供了一个方法 `json.Valid()`. 方法的原理是 `栈` + 状态迁移函数.
状态迁移函数转换图如下:

![image](/images/encode_json_valid.png)



接下来详细的介绍一下 `unmarshal` 方法.

```cgo
func (d *decodeState) unmarshal(v interface{}) error {
    // 检查 v, 必须是一个非空指针
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	d.scan.reset() // 初始化 scanner 的状态迁移函数 `stateBeginValue`.
	d.scanWhile(scanSkipSpace) // scanSkipSpace 是扫描的一个标记, 表示可以忽略当前的空格字符
	// 解码rv而不是rv.Elem, 因为必须在值的顶层应用Unmarshaler接口测试.
	err := d.value(rv)
	if err != nil {
		return d.addErrorContext(err)
	}
	return d.savedError
}
```

