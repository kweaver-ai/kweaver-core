package cutil

type Ordered interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64
}

// 检查值是否在指定范围内
func CheckInRange[T Ordered](value, min, max T) (ok bool) {
	if value < min || value > max {
		return false
	}

	return true
}

func CheckMin[T Ordered](value, min T) (ok bool) {
	return value >= min
}
