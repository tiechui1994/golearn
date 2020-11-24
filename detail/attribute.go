package main

import (
	"fmt"
	"golearn/detail/t"
	"reflect"
	"unsafe"
)

func GetUnExportFiled(s interface{}, filed string) reflect.Value {
	v := reflect.ValueOf(s).Elem().FieldByName(filed)
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}

func SetUnExportFiled(s interface{}, filed string, val interface{}) error {
	v := GetUnExportFiled(s, filed)
	rv := reflect.ValueOf(val)
	if v.Kind() != rv.Kind() {
		return fmt.Errorf("invalid kind, expected kind:%v, got kind:%v", v.Kind(), rv.Kind())
	}

	v.Set(rv)
	return nil
}

func main() {
	s := t.T{}
	err := SetUnExportFiled(&s, "a", "AAA")
	if err == nil {
		fmt.Printf("%+v\n", s)
	}
}
