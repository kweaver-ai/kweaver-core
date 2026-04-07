package funcs

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type Hash struct{}

func convertToInt(step interface{}) (int, error) {
	rv := reflect.ValueOf(step)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int(rv.Float()), nil
	case reflect.String:
		s := rv.String()
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		return i, nil
	default:
		return 0, fmt.Errorf("cannot convert to int")
	}
}

func get(current interface{}, key interface{}) interface{} {
	rv := reflect.ValueOf(current)
	switch rv.Kind() {
	case reflect.String:
		index, err := convertToInt(key)
		if err != nil || index < 0 || index >= rv.Len() {
			return ""
		}
		return string(rv.String()[index])

	case reflect.Slice, reflect.Array:
		index, err := convertToInt(key)
		if err != nil || index < 0 || index >= rv.Len() {
			return nil
		}
		return rv.Index(index).Interface()

	case reflect.Map:
		keyVal := reflect.ValueOf(key)
		if !keyVal.Type().ConvertibleTo(rv.Type().Key()) {
			return nil
		}
		convertedKey := keyVal.Convert(rv.Type().Key())
		if value := rv.MapIndex(convertedKey); value.IsValid() {
			return value.Interface()
		}
		return nil

	default:
		return nil
	}
}

func convertToString(target interface{}) string {
	if target == nil {
		return ""
	}
	val := reflect.ValueOf(target)
	switch val.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return fmt.Sprintf("%v", target)
	default:
		b, err := json.Marshal(target)
		if err != nil {
			return fmt.Sprintf("%v", target)
		}
		return string(b)
	}
}

func deepCopy(target interface{}) interface{} {
	b, err := json.Marshal(target)
	if err != nil {
		return target
	}
	var copy interface{}
	err = json.Unmarshal(b, &copy)
	if err != nil {
		return target
	}
	return copy
}

func set(target interface{}, paths []interface{}, value interface{}) interface{} {
	if len(paths) == 0 {
		return value
	}
	target = deepCopy(target)
	rv := reflect.ValueOf(target)

	switch rv.Kind() {
	case reflect.Map:
		m, ok := target.(map[string]interface{})
		if !ok {
			return target
		}
		path := paths[0]
		var key string
		switch v := path.(type) {
		case string:
			key = v
		default:
			key = fmt.Sprintf("%v", v)
		}
		if len(paths) == 1 {
			m[key] = value
		} else {
			m[key] = set(m[key], paths[1:], value)
		}
		return m

	case reflect.Slice:
		s, ok := target.([]interface{})
		if !ok {
			return target
		}
		path := paths[0]
		var index int
		switch v := path.(type) {
		case int:
			index = v
		default:
			s := fmt.Sprintf("%v", v)
			i, err := strconv.Atoi(s)
			if err != nil {
				return target
			}
			index = i
		}
		if index >= len(s) {
			newSlice := make([]interface{}, index+1)
			copy(newSlice, s)
			s = newSlice
		}
		if len(paths) == 1 {
			s[index] = value
		} else {
			s[index] = set(s[index], paths[1:], value)
		}
		return s

	case reflect.String:
		str := target.(string)
		path := paths[0]
		var index int
		switch v := path.(type) {
		case int:
			index = v
		default:
			s := fmt.Sprintf("%v", v)
			i, err := strconv.Atoi(s)
			if err != nil {
				return str
			}
			index = i
		}
		if index < 0 {
			return str
		}
		runes := []rune(str)
		if index >= len(runes) {
			padLength := index - len(runes) + 1
			for i := 0; i < padLength; i++ {
				runes = append(runes, ' ')
			}
		}
		valueStr := convertToString(value)
		valueRunes := []rune(valueStr)
		var newRunes []rune
		if index <= len(runes) {
			newRunes = append(runes[:index], valueRunes...)
			if index+len(valueRunes) <= len(runes) {
				newRunes = append(newRunes, runes[index+len(valueRunes):]...)
			}
		} else {
			return string(runes)
		}
		return string(newRunes)

	default:
		if len(paths) == 0 {
			return value
		}
		var newStructure interface{}
		path := paths[0]
		switch path.(type) {
		case int:
			newStructure = make([]interface{}, 0)
		default:
			newStructure = make(map[string]interface{})
		}
		return set(newStructure, paths, value)
	}
}

func (c *Hash) Call(ctx context.Context, name string, nrets int, args ...interface{}) (wait bool, rets []interface{}, err error) {
	switch name {
	case "get":
		if len(args) < 1 {
			return false, nil, nil
		}
		target := args[0]
		var path []interface{}
		if len(args) >= 2 {
			if pathSlice, ok := args[1].([]interface{}); ok {
				path = pathSlice
			} else {
				path = []interface{}{args[1]}
			}
		} else {
			path = []interface{}{}
		}

		current := target
		for _, key := range path {
			current = get(current, key)
		}
		return false, []interface{}{current}, nil

	case "set":
		var target, pathArg, value any
		var path []interface{}

		switch len(args) {
		case 0:
			return false, nil, nil
		case 1, 2:
			return false, []any{args[0]}, nil
		default:
			target, pathArg, value = args[0], args[1], args[2]
		}

		if pathSlice, ok := pathArg.([]interface{}); ok {
			path = pathSlice
		} else if pathArg != nil {
			path = []interface{}{pathArg}
		}

		result := set(target, path, value)
		return false, []interface{}{result}, nil
	default:
		return false, nil, fmt.Errorf("unknown method: %s", name)
	}
}
