package funcs

import (
	"context"
	"reflect"
)

type array struct{}

func toArray(value any) (arr []any) {
	if value != nil {
		rv := reflect.ValueOf(value)

		if rv.Kind() == reflect.Slice {
			arr = value.([]any)
		} else {
			arr = append(arr, value)
		}
	}
	return
}

func (c *array) Call(ctx context.Context, name string, nrets int, args ...interface{}) (wait bool, rets []interface{}, err error) {
	switch name {
	case "array":
		arr := make([]any, 0)
		if len(args) > 0 {
			arr = toArray(args[0])
		}
		return false, []any{arr}, nil

	case "len":
		if len(args) < 1 {
			return false, []any{0}, nil
		}
		rv := reflect.ValueOf(args[0])
		switch rv.Kind() {
		case reflect.Array, reflect.Slice, reflect.String, reflect.Map:
			return false, []any{rv.Len()}, nil
		default:
			return false, []any{0}, nil
		}

	case "append":
		arr := make([]any, 0)

		if len(args) > 0 {
			arr = append(toArray(args[0]), args[1:]...)
		}

		return false, []any{arr}, nil
	}

	return
}
