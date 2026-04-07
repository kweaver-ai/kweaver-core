package cutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultTimeFormat = "2006-01-02 15:04:05"
)

// NowStr 获取当前时间字符串
func NowStr() string {
	return time.Now().Format(DefaultTimeFormat)
}

// 获取当前毫秒时间戳
func GetCurrentMSTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func FormatTime(t time.Time) string {
	return t.Format(DefaultTimeFormat)
}

func FormatTimeUnix(t int64) string {
	return time.Unix(t, 0).Format(DefaultTimeFormat)
}

// ParseTime 解析时间字符串
func ParseTime(timeStr string) (hour, min, sec int, err error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid time format: %s, expected HH:MM:SS", timeStr)
	}

	// 解析小时
	hour, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hour format: %s", parts[0])
	}

	// 解析分钟
	min, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minute format: %s", parts[1])
	}

	// 解析秒
	sec, err = strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid second format: %s", parts[2])
	}

	// 验证时间范围
	if hour >= 0 && hour <= 23 && min >= 0 && min <= 59 && sec >= 0 && sec <= 59 {
		return hour, min, sec, nil
	}

	return 0, 0, 0, fmt.Errorf("time values out of range: hour=%d, min=%d, sec=%d", hour, min, sec)
}
