package main

import (
	"encoding/json"
	"fmt"
)

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
