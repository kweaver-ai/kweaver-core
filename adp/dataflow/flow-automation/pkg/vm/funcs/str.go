package funcs

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type str struct{}

func (c *str) Call(ctx context.Context, name string, nrets int, args ...interface{}) (wait bool, rets []interface{}, err error) {

	switch name {
	case "str":
		parts := make([]string, len(args))

		for i, arg := range args {
			if arg == nil {
				parts[i] = ""
				continue
			}

			rv := reflect.ValueOf(arg)
			switch rv.Kind() {
			case reflect.Array, reflect.Slice, reflect.Map:
				bytes, _ := json.Marshal(arg)
				parts[i] = string(bytes)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				parts[i] = strconv.FormatInt(int64(rv.Int()), 10)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				parts[i] = strconv.FormatUint(rv.Uint(), 10)
			case reflect.Float32:
				parts[i] = strconv.FormatFloat(rv.Float(), 'f', -1, 32)
			case reflect.Float64:
				parts[i] = strconv.FormatFloat(rv.Float(), 'f', -1, 64)
			default:
				parts[i] = fmt.Sprintf("%v", arg)
			}
		}

		s := strings.Join(parts, "")
		return false, []interface{}{s}, nil
	default:
		return false, nil, fmt.Errorf("unknown method: %s", name)
	}
}
