package cutil

import (
	"reflect"
)

func IsStringOrNumber(value interface{}) bool {
	valueType := reflect.TypeOf(value)

	switch valueType.Kind() {
	case reflect.String:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func IsZeroValue(value interface{}) bool {
	// 使用 reflect.ValueOf 获取值的反射对象
	v := reflect.ValueOf(value)
	// 使用 reflect.Zero 获取类型的零值
	zeroValue := reflect.Zero(v.Type())
	// 使用 reflect.DeepEqual 比较值和零值
	return reflect.DeepEqual(v.Interface(), zeroValue.Interface())
}
