package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type validationErr []error

// CheckJSONValue 错误
func (e validationErr) Error() string {
	buff := bytes.NewBufferString("")
	for i := 0; i < len(e); i++ {
		buff.WriteString(e[i].Error())
		buff.WriteString("\n")
	}
	return strings.TrimSpace(buff.String())
}

// trans 错误消息与CheckJSONValue保持一致
func trans(fe validator.FieldError) (err error) {
	switch fe.Tag() {
	case "required":
		err = fmt.Errorf("%s is required", fe.Namespace())
	case "oneof":
		err = fmt.Errorf("%s should be one of %s", fe.Namespace(), fe.Param())
	case "gt":
		switch fe.Kind() { //nolint
		case reflect.Slice:
			err = fmt.Errorf("len of %s should greater than %s", fe.Namespace(), fe.Param())
		default:
			err = fmt.Errorf("%s validation failed on tag %s", fe.Namespace(), fe.Tag())
		}
	default:
		err = fmt.Errorf("%s validation failed on tag %s", fe.Namespace(), fe.Tag())
	}
	return
}

// ValidationTrans 使用validator校验参数并优化错误消息
func ValidationTrans(err error) error {
	var ve validator.ValidationErrors
	errs := validationErr{}
	if errors.As(err, &ve) {
		for _, fe := range ve {
			errs = append(errs, trans(fe))
		}
		return errs
	}

	ue, ok := err.(*json.UnmarshalTypeError)
	if ok {
		if ue.Struct != "" || ue.Field != "" {
			errs = append(errs, fmt.Errorf("type of %s should be %s", ue.Struct+"."+ue.Field, ue.Type.String()))
			return errs
		}
	}

	return err
}
