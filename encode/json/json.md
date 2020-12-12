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

