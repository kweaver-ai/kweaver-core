// Package rest 对RESTful API请求参数进行检查
// 参数必须性、参数类型、非必需参数是否存在
package rest

import (
	"fmt"
	"reflect"
)

// JSONValueDesc json格式描述
// reflect.Bool , for JSON booleans
// reflect.Float64 , for JSON numbers
// reflect.String , for JSON strings
// reflect.Slice , for JSON arrays
// reflect.Map , for JSON objects
type JSONValueDesc struct {
	Kind      reflect.Kind
	Required  bool
	Exist     bool
	ValueDesc map[string]*JSONValueDesc
}

// CheckJSONValue 检查请求参数json格式
func CheckJSONValue(key string, jsonV interface{}, jsonValueDesc *JSONValueDesc) error {
	kind := reflect.ValueOf(jsonV).Kind()
	if kind != jsonValueDesc.Kind {
		return NewHTTPError(fmt.Sprintf("type of %s should be %v", key, jsonValueDesc.Kind), BadRequest, nil)
	} else if kind == reflect.Map {
		obj := jsonV.(map[string]interface{})
		for k, valueDesc := range jsonValueDesc.ValueDesc {
			if v, ok := obj[k]; ok {
				err := CheckJSONValue(fmt.Sprintf("%s.%s", key, k), v, valueDesc)
				if err != nil {
					return err
				}
				valueDesc.Exist = true
			} else if valueDesc.Required {
				return NewHTTPError(fmt.Sprintf("%v is required", fmt.Sprintf("%s.%s", key, k)), BadRequest, nil)
			}
		}
	} else if kind == reflect.Slice {
		arr := jsonV.([]interface{})
		for i, element := range arr {
			err := CheckJSONValue(fmt.Sprintf(`%s[%d]`, key, i), element, jsonValueDesc.ValueDesc["element"])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
