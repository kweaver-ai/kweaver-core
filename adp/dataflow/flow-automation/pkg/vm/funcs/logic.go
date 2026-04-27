package funcs

import (
	"context"
	"fmt"
	"math"
	"reflect"
)

type Logic struct{}

func isFalsy(v interface{}) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return math.Float64bits(val.Float()) == math.Float64bits(0.0)
	case reflect.String:
		return val.String() == ""
	case reflect.Slice, reflect.Array:
		return val.Len() == 0
	case reflect.Map:
		return val.Len() == 0
	default:
		return false
	}
}

func isTruthy(v interface{}) bool {
	return !isFalsy(v)
}

func (*Logic) Call(ctx context.Context, name string, nrets int, args ...interface{}) (wait bool, rets []interface{}, err error) {
	values := make([]bool, len(args))
	for i, v := range args {
		values[i] = isTruthy(v)
	}

	var result bool
	switch name {
	case "and":
		result = true
		for _, val := range values {
			result = result && val
		}
	case "or":
		result = false
		for _, val := range values {
			result = result || val
		}
	case "not":
		if len(values) == 0 {
			result = true
		} else {
			result = !values[0]
		}
	default:
		return false, nil, fmt.Errorf("unsupported operation: %s", name)
	}

	rets = append(rets, result)
	return false, rets, nil
}
