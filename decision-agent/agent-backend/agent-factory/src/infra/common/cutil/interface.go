package cutil

import (
	"reflect"
)

// IsImplement 判断i是否实现了inter接口
// 参数i：需要判断的对象
// 参数inter：需要判断的接口
// 返回值：如果i实现了inter接口，返回true；否则返回false
// 示例：
// type MyInterface interface {
//     MyMethod()
// }
//
// type MyStruct struct {}
//
// func (m MyStruct) MyMethod() {}
//
// var myVar MyInterface = MyStruct{}
// var isImplement = IsImplement(myVar, (*MyInterface)(nil)) // isImplement为true

func IsImplement(i, inter interface{}) bool {
	return reflect.TypeOf(i).
		Implements(reflect.TypeOf(inter).Elem())
}

// MustStrSlice 将[]interface{}转换为[]string
func MustStrSlice(i []interface{}) (s []string) {
	s = make([]string, len(i))
	for j := range i {
		s[j] = i[j].(string)
	}

	return
}

// MustStrSlice2 将interface{}转换为[]string
func MustStrSlice2(i interface{}) (s []string) {
	s = make([]string, len(i.([]interface{})))
	for j := range i.([]interface{}) {
		s[j] = i.([]interface{})[j].(string)
	}

	return
}
