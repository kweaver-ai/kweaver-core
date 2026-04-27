package utils

import (
	"fmt"
)

// ConvertToBytes  将当前限制转化为 B、KB、MB、GB、TB为单位的字符串
func ConvertToBytes(size int64) string {
	const (
		KB int64 = 1024
		MB       = KB * 1024
		GB       = MB * 1024
		TB       = GB * 1024
	)
	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}
