// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

func GenerateUniqueKey(id string, label map[string]string) string {
	var keys []string
	for k := range label {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	key := id
	for _, k := range keys {
		key = key + "-" + k + ":" + label[k]
	}
	return key
}

func IsSlice(i any) bool {
	kind := reflect.ValueOf(i).Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

func IsSameType(arr []any) bool {
	if len(arr) == 0 {
		return true
	}

	firstType := reflect.TypeOf(arr[0])
	for _, v := range arr {
		if reflect.TypeOf(v) != firstType {
			return false
		}
	}

	return true
}

func SplitString2InterfaceArray(s, sep string) []any {
	count := strings.Count(s, sep) + 1
	anys := make([]any, count)
	idx := 0
	for {
		idx = strings.Index(s, sep)
		if idx == -1 {
			anys[len(anys)-1] = s
			break
		}
		anys[len(anys)-count] = s[:idx]
		s = s[idx+len(sep):]
		count--
	}
	return anys

	// strs := strings.Split(s, sep)
	// anys := make([]any, len(strs))

	// for i, subStr := range strs {
	// 	anys[i] = subStr
	// }
	// return anys
}

func Any2String(val any) string {
	switch t := val.(type) {
	case string:
		return val.(string)
	case uint:
		it := val.(uint)
		return strconv.Itoa(int(it))
	case uint8:
		it := val.(uint8)
		return strconv.Itoa(int(it))
	case uint16:
		it := val.(uint16)
		return strconv.Itoa(int(it))
	case uint32:
		it := val.(uint32)
		return strconv.Itoa(int(it))
	case uint64:
		it := val.(uint64)
		return strconv.FormatUint(it, 10)
	case int:
		it := val.(int)
		return strconv.Itoa(it)
	case int8:
		it := val.(int8)
		return strconv.Itoa(int(it))
	case int16:
		it := val.(int16)
		return strconv.Itoa(int(it))
	case int32:
		it := val.(int32)
		return strconv.Itoa(int(it))
	case int64:
		it := val.(int64)
		return strconv.FormatInt(it, 10)
	case float32:
		ft := val.(float32)
		return strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case float64:
		ft := val.(float64)
		return strconv.FormatFloat(ft, 'f', -1, 64)
	case []byte:
		return string(val.([]byte))
	default:
		logger.Warnf("unspported interface dynamic type: %v", t)
		return ""
	}
}

func CloneStringMap(originalMap map[string]string) map[string]string {
	newMap := make(map[string]string)
	for k, v := range originalMap {
		newMap[k] = v
	}
	return newMap
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
