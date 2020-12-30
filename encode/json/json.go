package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Student struct {
	Child []int `json:"child"`
}

func example() {
	data := `{"child":[1,2,3]}`
	var a Student
	json.Unmarshal([]byte(data), &a)
	aa := a.Child
	fmt.Println(aa)

	data = `{"child":[3,4,5,6,7,8]}`
	json.Unmarshal([]byte(data), &a)
	fmt.Println(aa)
}

func valid() {
	// json 合法的开始头部
	case1 := "123"
	fmt.Println(case1, json.Valid([]byte(case1)))

	case2 := "0.123"
	fmt.Println(case2, json.Valid([]byte(case2)))

	case3 := "-0"
	fmt.Println(case3, json.Valid([]byte(case3)))

	case4 := "true"
	fmt.Println(case4, json.Valid([]byte(case4)))

	case5 := "false"
	fmt.Println(case5, json.Valid([]byte(case5)))

	case6 := "{}"
	fmt.Println(case6, json.Valid([]byte(case6)))

	case7 := "[]"
	fmt.Println(case7, json.Valid([]byte(case7)))

	case8 := `""`
	fmt.Println(case8, json.Valid([]byte(case8)))
}

func write() {
	data := make([]byte, 8192)
	copy(data[0:4096], bytes.Repeat([]byte{'·'}, 4096))
	ioutil.WriteFile("./www.data", data, 066)
}
func main() {
	write()
}
