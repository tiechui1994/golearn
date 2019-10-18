package main

import (
	"fmt"
	"reflect"
)

//type People struct{}
//
//func (p *People) ShowA() {
//	fmt.Println("showA")
//	p.ShowB() // 这里它并不知道它的上级是什么。
//}
//
//func (p *People) ShowB() {
//	fmt.Println("showB")
//}
//
//type Teacher struct {
//	People
//}
//
//func (t *Teacher) ShowB() {
//	fmt.Println("teacher showB")
//}
//
//func main() {
//	t := Teacher{}
//	t.ShowA() // showA showB
//	t.ShowB() // teacher showB
//}


type People interface {
	Show()
}

type Student struct{}

func (stu *Student) Show() {

}


// x(TYPE,VALUE)
// 如果VALUE == nil, 则 x == nil
// x(TYPE, VALUE) x'(TYPE, VALUE)
// 如果
func live() People {
	var stu People
	return stu
}

func main() {
	stu := new(Student)
	var x People = stu
	var t interface{} = stu
	if x == t {
		fmt.Println("x==t")
	}

	var p People
	fmt.Println(reflect.TypeOf(p))
	fmt.Println(reflect.ValueOf(p))

	if p == nil {
		fmt.Println("++++++++")
	}

	var i interface{}
	fmt.Println(reflect.TypeOf(i))
	fmt.Println(reflect.ValueOf(i))
	if p == i {
		fmt.Println("============")
	}

	fmt.Println(reflect.TypeOf(live()))
	fmt.Println(reflect.ValueOf(live()))
	if live() == i {
		fmt.Println("...........")
	}
}