package sqlhelper2

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// Clause 搜索
type Clause struct {
	Key      string      `json:"key"`
	Operator Operator    `json:"operator"`
	Value    interface{} `json:"value"`
}

func (c *Clause) Build() (sqlStr string, args []interface{}, err error) {
	val := c.Value
	key := c.Key
	op := string(c.Operator)
	args = make([]interface{}, 0)

	switch c.Operator {
	case OperatorIn, OperatorNotIn:
		switch s := val.(type) {
		case []string:
			sqlStr, args, err = parseInClause(s)
		case []int:
			sqlStr, args, err = parseInClause(s)
		case []int8:
			sqlStr, args, err = parseInClause(s)
		case []int16:
			sqlStr, args, err = parseInClause(s)
		case []int32:
			sqlStr, args, err = parseInClause(s)
		case []int64:
			sqlStr, args, err = parseInClause(s)
		case []uint:
			sqlStr, args, err = parseInClause(s)
		case []uint8:
			sqlStr, args, err = parseInClause(s)
		case []uint16:
			sqlStr, args, err = parseInClause(s)
		case []uint32:
			sqlStr, args, err = parseInClause(s)
		case []uint64:
			sqlStr, args, err = parseInClause(s)
		case []float32:
			sqlStr, args, err = parseInClause(s)
		case []float64:
			sqlStr, args, err = parseInClause(s)
		case []interface{}:
			sqlStr, args, err = c.parseInClauseForInterfaceSlice(s)
		default:
			//nolint:goerr113
			err = errors.New("OperatorNotIn and OperatorIn only support " +
				"[]string, []int, []int8, []int16, []int32, []int64, []uint" +
				", []uint8, []uint16, []uint32, []uint64, []float32, []float64, []interface{}")
		}

		if err != nil {
			return
		}

		sqlStr = key + op + sqlStr
	case OperatorLike, OperatorNotLike:
		sqlStr = key + op + "?"

		args = append(args, fmt.Sprintf("%%%v%%", val))
	case OperatorIsNull, OperatorIsNotNull:
		sqlStr = key + op
	default:
		sqlStr = key + op + "?"

		args = append(args, val)
	}

	// sqlStr = "(" + sqlStr + ")"
	return
}

//nolint:funlen
func (c *Clause) parseInClauseForInterfaceSlice(s []interface{}) (sqlStr string, args []interface{}, err error) {
	if !isAllElementsSameType(s) {
		//nolint:goerr113
		err = errors.New("when operator is OperatorIn or OperatorNotIn, all elements should be of the same type")
		return
	}

	// 判断s的长度
	if len(s) == 0 {
		//nolint:goerr113
		err = errors.New("[parseInClauseForInterfaceSlice]:OperatorNotIn and OperatorIn must have at least one element")
		return
	}

	// 判断s的元素类型
	switch s[0].(type) {
	case string:
		ns := make([]string, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(string)
		}

		sqlStr, args, err = parseInClause(ns)
	case int:
		ns := make([]int, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(int)
		}

		sqlStr, args, err = parseInClause(ns)
	case int8:
		ns := make([]int8, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(int8)
		}

		sqlStr, args, err = parseInClause(ns)
	case int16:
		ns := make([]int16, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(int16)
		}

		sqlStr, args, err = parseInClause(ns)
	case int32:
		ns := make([]int32, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(int32)
		}

		sqlStr, args, err = parseInClause(ns)
	case int64:
		ns := make([]int64, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(int64)
		}

		sqlStr, args, err = parseInClause(ns)
	case uint:
		ns := make([]uint, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(uint)
		}

		sqlStr, args, err = parseInClause(ns)
	case uint8:
		ns := make([]uint8, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(uint8)
		}

		sqlStr, args, err = parseInClause(ns)
	case uint16:
		ns := make([]uint16, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(uint16)
		}

		sqlStr, args, err = parseInClause(ns)
	case uint32:
		ns := make([]uint32, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(uint32)
		}

		sqlStr, args, err = parseInClause(ns)
	case uint64:
		ns := make([]uint64, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(uint64)
		}

		sqlStr, args, err = parseInClause(ns)
	case float32:
		ns := make([]float32, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(float32)
		}

		sqlStr, args, err = parseInClause(ns)
	case float64:
		ns := make([]float64, len(s))
		for k, v := range s {
			//nolint:forcetypeassert //上面isAllElementsSameType等已保证这里的类型转换不会出错
			ns[k] = v.(float64)
		}

		sqlStr, args, err = parseInClause(ns)
	default:
		//nolint:goerr113
		err = errors.New("[parseInClauseForInterfaceSlice]: unsupported type")
	}

	return sqlStr, args, err
}

// 判断所有元素的类型是否相同
func isAllElementsSameType(items []interface{}) bool {
	if len(items) == 0 {
		return true
	}

	// 获取第一个元素的类型
	firstType := reflect.TypeOf(items[0])

	// 遍历切片并比较每个元素的类型
	for _, item := range items {
		if reflect.TypeOf(item) != firstType {
			return false
		}
	}

	return true
}

func parseInClause[T uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string |
	int | int8 | int16 | int32 | int64](ts []T) (sqlStr string, args []interface{}, err error) {
	args = make([]interface{}, 0)

	if len(ts) == 0 {
		//nolint:goerr113
		err = errors.New("OperatorNotIn and OperatorIn must have at least one element")
		return
	}

	// 去重
	ts = cutil.DeduplGeneric(ts)

	b := strings.Builder{}
	b.WriteByte('(')

	for i := range ts {
		b.WriteByte('?')

		if i != len(ts)-1 {
			b.WriteString(",")
		}

		args = append(args, ts[i])
	}

	b.WriteByte(')')

	sqlStr = b.String()

	return
}
