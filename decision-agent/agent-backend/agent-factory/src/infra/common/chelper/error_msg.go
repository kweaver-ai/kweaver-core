package chelper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type IErrMsgBindStruct interface {
	// GetErrMsgMap 获取错误信息map
	GetErrMsgMap() map[string]string
}

// ErrMsg 获取err对应的，在dto结构图中事先配置的错误信息
// 【注意】：这个只能处理部分自定义的错误信息，其他的会直接返回err.Error()
// s: dto结构体
// err: 错误
// 返回值：错误信息
func ErrMsg(err error, s IErrMsgBindStruct) (retMsg string) {
	defer func() {
		if retMsg == "" {
			retMsg = err.Error()
		}
	}()

	var validationErrs validator.ValidationErrors

	if errors.As(err, &validationErrs) {
		v := reflect.ValueOf(s)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		sName := v.Type().Name()

		msgs := getErrsMsgs(validationErrs, sName, v, s)
		sno := 1

		for i := range msgs {
			msgs[i] = fmt.Sprintf("%d、%s", sno, msgs[i])
			sno++
		}

		return strings.Join(msgs, "；")
	}

	return err.Error()
}

// getErrsMsgs 获取err对应的，在dto结构图中事先配置的错误信息
func getErrsMsgs(validationErrs validator.ValidationErrors, sName string, v reflect.Value, s IErrMsgBindStruct) (msgs []string) {
	for _, e := range validationErrs {
		var m map[string]string

		fieldName := e.StructField()
		// paramVal := e.Value()
		key := fmt.Sprintf("%s.%s", fieldName, e.Tag())

		fullName := e.StructNamespace()

		leftName := strings.TrimPrefix(fullName, sName+".")
		midFieldStr := ""

		if len(leftName) > len(fieldName) {
			midFieldStr = leftName[:len(leftName)-len(fieldName)-1]
		}

		if midFieldStr != "" { // composite，内嵌结构体字段
			m = getMsgMapFromMidFields(midFieldStr, v)
		} else { // 非composite字段
			m = s.GetErrMsgMap()
		}

		if m != nil {
			if msg, ok := m[key]; ok {
				msgs = append(msgs, msg)
			}
		}
	}

	return
}

// getMsgMapFromMidFields 从中间字段获取错误信息map
func getMsgMapFromMidFields(midFieldStr string, v reflect.Value) (m map[string]string) {
	midFields := strings.Split(midFieldStr, ".")

	var bottomField reflect.Value

	for _, midField := range midFields {
		if f := v.FieldByName(midField); f.IsValid() {
			// 如果不是结构体或结构体指针，跳过
			if f.Kind() == reflect.Ptr {
				if f.Elem().Kind() != reflect.Struct {
					continue
				}
			} else if f.Kind() != reflect.Struct {
				continue
			}

			bottomField = f

			if ss, ok := bottomField.Interface().(IErrMsgBindStruct); ok {
				m = ss.GetErrMsgMap()
			} else {
				panic(fmt.Sprintf("struct %s not implement IErrMsgBindStruct", bottomField.Type().Name()))
			}
		}
	}

	return
}
