// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	lib_common "github.com/kweaver-ai/kweaver-go-lib/common"
)

// string 转 []string
func StringToStringSlice(str string) []string {
	if str == "" {
		return []string{}
	}

	strSlice := []string{}
	strs := strings.Split(str, ",")
	for _, v := range strs {
		v = strings.Trim(v, " ")
		if v != "" {
			strSlice = append(strSlice, v)
		}
	}
	return strSlice
}

const (
	oneGiB = 1024 * 1024 * 1024 //1073741824.0 定义1GB的字节数

	CALENDAR_STEP_MINUTE  string = "minute"
	CALENDAR_STEP_HOUR    string = "hour"
	CALENDAR_STEP_DAY     string = "day"
	CALENDAR_STEP_WEEK    string = "week"
	CALENDAR_STEP_MONTH   string = "month"
	CALENDAR_STEP_QUARTER string = "quarter"
	CALENDAR_STEP_YEAR    string = "year"
)

func BytesToGiB(bytes int64) float64 {
	return math.Round(float64(bytes)/oneGiB*100) / 100 // 四舍五入到小数点后两位
}

func GiBToBytes(gib int64) int64 {
	return gib * oneGiB
}

// AnyToFloat64 converts a numeric or string value to float64 (JSON unmarshal / API responses).
func AnyToFloat64(value any) (float64, error) {
	if value == nil {
		return 0, errors.New("value is nil")
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.String:
		return strconv.ParseFloat(v.String(), 64)
	default:
		return 0, fmt.Errorf("无法将类型 %T 转换为 float64", value)
	}
}

// AnyToInt64 尝试将 interface{} 转换为 int64，如果转换失败则返回错误。
func AnyToInt64(value any) (int64, error) {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int64(v.Float()), nil
	case reflect.String:
		return strconv.ParseInt(v.String(), 10, 64)
	default:
		return 0, fmt.Errorf("无法将类型 %T 转换为 int64", value)
	}
}

// 使用类型断言和标准库进行转换
func AnyToBool(value any) (bool, error) {
	// 检查是否是字符串类型
	if s, ok := value.(string); ok {
		// 使用 strconv.ParseBool 转换字符串
		// 它接受 "1", "t", "T", "TRUE", "true", "True" 为真
		// 接受 "0", "f", "F", "FALSE", "false", "False" 为假 [citation:3][citation:5][citation:8]
		return strconv.ParseBool(s)
	}
	// 检查是否是布尔类型本身，如果是则直接返回
	if b, ok := value.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("unsupported type: %T", value)
}

func AnyToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ReplaceLikeWildcards，把 like 的通配符替换成正则表达式里的字符
func ReplaceLikeWildcards(input string) string {
	if input == "" {
		return input
	}

	var result strings.Builder
	escaped := false
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if escaped {
			// 转义字符后的字符
			switch r {
			case '%', '_', '\\':
				result.WriteRune(r)
			default:
				// 如果转义了非特殊字符，保留转义符和字符
				result.WriteRune('\\')
				result.WriteRune(r)
			}
			escaped = false
		} else if r == '\\' {
			// 遇到转义符，检查是否是最后一个字符
			if i == len(runes)-1 {
				// 转义符在末尾，直接输出
				result.WriteRune(r)
			} else {
				// 标记转义状态，但不立即输出转义符
				escaped = true
			}
		} else if r == '%' {
			result.WriteString(".*")
		} else if r == '_' {
			result.WriteString(".")
		} else {
			result.WriteRune(r)
		}
	}

	// 处理以转义符结尾的情况
	if escaped {
		result.WriteRune('\\')
	}

	return result.String()
}

// lastDayOfMonth 返回给定日期所在月份的最后一天
func LastDayOfMonth(t time.Time) time.Time {
	firstOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	nextMonth := firstOfMonth.AddDate(0, 1, 0)
	return nextMonth.AddDate(0, 0, -1)
}

// isLeap 判断是否是闰年
func IsLeap(year int) bool {
	return year%400 == 0 || (year%100 != 0 && year%4 == 0)
}

// ParseDurationMilli parses a string into a millisecod,
// assuming that a year always has 365d, a week always has 7d, and a day always has 24h.
func ParseDuration(s string) (time.Duration, error) {
	switch s {
	case "0":
		// Allow 0 without a unit.
		return 0, nil
	case "":
		return 0, errors.New("empty duration string")
	}

	orig := s
	var dur uint64
	lastUnitPos := 0

	for s != "" {
		if !isdigit(s[0]) {
			return 0, fmt.Errorf("not a valid duration string: %q", orig)
		}
		// Consume [0-9]*
		i := 0
		for ; i < len(s) && isdigit(s[i]); i++ {
		}
		v, err := strconv.ParseUint(s[:i], 10, 0)
		if err != nil {
			return 0, fmt.Errorf("not a valid duration string: %q", orig)
		}
		s = s[i:]

		// Consume unit.
		for i = 0; i < len(s) && !isdigit(s[i]); i++ {
		}
		if i == 0 {
			return 0, fmt.Errorf("not a valid duration string: %q", orig)
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, fmt.Errorf("unknown unit %q in duration %q", u, orig)
		}
		if unit.pos <= lastUnitPos { // Units must go in order from biggest to smallest.
			return 0, fmt.Errorf("not a valid duration string: %q", orig)
		}
		lastUnitPos = unit.pos
		// Check if the provided duration overflows time.Duration (> ~ 290years).
		if v > 1<<63/unit.mult {
			return 0, errors.New("duration out of range")
		}
		dur += v * unit.mult
		if dur > 1<<63-1 {
			return 0, errors.New("duration out of range")
		}
	}
	return time.Duration(dur), nil
}

func isdigit(c byte) bool { return c >= '0' && c <= '9' }

// Units are required to go in order from biggest to smallest.
// This guards against confusion from "1m1d" being 1 minute + 1 day, not 1 month + 1 day.
var unitMap = map[string]struct {
	pos  int
	mult uint64
}{
	"ms": {7, uint64(time.Millisecond)},
	"s":  {6, uint64(time.Second)},
	"m":  {5, uint64(time.Minute)},
	"h":  {4, uint64(time.Hour)},
	"d":  {3, uint64(24 * time.Hour)},
	"w":  {2, uint64(7 * 24 * time.Hour)},
	"y":  {1, uint64(365 * 24 * time.Hour)},
}

// appLocationOrUTC returns APP_LOCATION or UTC when the app is not initialized (e.g. unit tests).
func appLocationOrUTC() *time.Location {
	if APP_LOCATION != nil {
		return APP_LOCATION
	}
	return time.UTC
}

// AppLocationOrUTC is the exported form of appLocationOrUTC for callers outside package common.
func AppLocationOrUTC() *time.Location { return appLocationOrUTC() }

func FormatTimeMiliis(ts int64, formatType string) string {
	t := time.UnixMilli(ts).In(appLocationOrUTC())
	switch formatType {
	case CALENDAR_STEP_MINUTE:
		return t.Format("2006-01-02 15:04")
	case CALENDAR_STEP_HOUR:
		return t.Format("2006-01-02 15")
	case CALENDAR_STEP_DAY:
		return t.Format("2006-01-02")
	case CALENDAR_STEP_WEEK:
		// 周：年-周 (如 2025-46)
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-%02d", year, week)
	case CALENDAR_STEP_MONTH:
		return t.Format("2006-01")
	case CALENDAR_STEP_QUARTER:
		quarter := (t.Month()-1)/3 + 1
		return fmt.Sprintf("%d-Q%d", t.Year(), quarter)
	case CALENDAR_STEP_YEAR:
		return t.Format("2006")
	default:
		return FormatRFC3339Milli(ts)
	}
}

// 按环境变量中的时区格式化
func FormatRFC3339Milli(timestamp int64) string {
	t := time.UnixMilli(timestamp)

	// 转换为指定时区并格式化
	return t.In(appLocationOrUTC()).Format(lib_common.RFC3339Milli)
}
