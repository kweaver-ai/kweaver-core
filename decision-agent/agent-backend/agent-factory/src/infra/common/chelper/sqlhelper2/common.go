package sqlhelper2

import (
	"fmt"
	"reflect"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// struct2SQLValPairsMapByTag 将结构体转换为insert或update的键值对
// 1. 只支持string和数字 或 *string和*数字（指针）
// 2. 如果是指针，取指针指向的值
// 3. 如果是零值，不插入
func struct2SQLValPairsMapByTag(obj interface{}, tag string) (pairs map[string]interface{}, err error) {
	v := reflect.ValueOf(obj)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic("obj must be a struct or a pointer to struct")
	}

	pairs = make(map[string]interface{})

	for i := 0; i < v.NumField(); i++ {
		key, ok := t.Field(i).Tag.Lookup(tag)
		if !ok {
			continue
		}

		if key == "" || key == "-" {
			continue
		}

		// 如果是零值，不插入
		if cutil.IsZeroValue(v.Field(i).Interface()) {
			continue
		}

		field := v.Field(i)

		// 如果是指针，取指针指向的值
		if field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		// 只支持字符串和数字
		if !cutil.IsStringOrNumber(field.Interface()) {
			if kind := v.Field(i).Kind(); kind == reflect.Ptr {
				//nolint:goerr113
				err = fmt.Errorf("only support string number *string *number, but field %s is %s and underlying type is %s",
					t.Field(i).Name, kind.String(), field.Kind().String())
			} else {
				//nolint:goerr113
				err = fmt.Errorf("only support string number *string *number, but field %s is %s", t.Field(i).Name, v.Field(i).Kind().String())
			}

			return
		}

		pairs[key] = field.Interface()
	}

	return
}
