package funcs

import (
	"context"
	"fmt"
)

type compare struct{}

func (c *compare) Call(ctx context.Context, name string, nrets int, args ...interface{}) (wait bool, rets []interface{}, err error) {

	var a, b interface{} = nil, nil

	if len(args) > 0 {
		a = args[0]
	}

	if len(args) > 1 {
		b = args[1]
	}

	switch name {
	case "eq":
		return false, []any{a == b}, nil
	case "ne":
		return false, []any{a != b}, nil
	case "lt":
		lv, _ := toFloat64(a)
		rv, _ := toFloat64(b)
		return false, []any{lv < rv}, nil
	case "lte":
		lv, _ := toFloat64(a)
		rv, _ := toFloat64(b)
		return false, []any{lv <= rv}, nil
	case "gt":
		lv, _ := toFloat64(a)
		rv, _ := toFloat64(b)
		return false, []any{lv > rv}, nil
	case "gte":
		lv, _ := toFloat64(a)
		rv, _ := toFloat64(b)
		return false, []any{lv >= rv}, nil
	default:
		return false, []any{}, fmt.Errorf("unknown method: %s", name)
	}
}
