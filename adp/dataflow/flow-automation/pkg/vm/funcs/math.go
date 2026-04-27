package funcs

import (
	"context"
	"fmt"
	"reflect"
)

type Math struct{}

// toFloat64 将任意类型转换为 float64
func toFloat64(v interface{}) (float64, error) {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return val.Float(), nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}

func (*Math) Call(ctx context.Context, name string, nrets int, args ...interface{}) (wait bool, rets []interface{}, err error) {
	nums := make([]float64, len(args))
	for i, arg := range args {
		nums[i], err = toFloat64(arg)
		if err != nil {
			return false, nil, fmt.Errorf("invalid argument %d: %v", i, err)
		}
	}

	var result float64
	switch name {
	case "add":
		result = 0
		for _, num := range nums {
			result += num
		}
	case "sub":
		if len(nums) == 0 {
			result = 0.0
		} else {
			result = nums[0]
			for i := 1; i < len(nums); i++ {
				result -= nums[i]
			}
		}
	case "mul":
		result = 1.0
		for _, num := range nums {
			result *= num
		}
	case "div":
		if len(nums) == 0 {
			result = 0.0
		} else {
			result = nums[0]
			for i := 1; i < len(nums); i++ {
				if nums[i] == 0 {
					return false, nil, fmt.Errorf("division by zero")
				}
				result /= nums[i]
			}
		}
	default:
		return false, nil, fmt.Errorf("unsupported operation: %s", name)
	}

	rets = append(rets, result)
	return false, rets, nil
}
