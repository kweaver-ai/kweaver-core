package ptr

// Int64 returns a pointer to the given int64 value.
func Int64(v int64) *int64 {
	return &v
}

// String returns a pointer to the given string value.
func String(v string) *string {
	return &v
}

// Bool returns a pointer to the given bool value.
func Bool(v bool) *bool {
	return &v
}

// Float64 returns a pointer to the given float64 value.
func Float64(v float64) *float64 {
	return &v
}

// ToPtr returns a pointer to the given value, or nil if the given value is
// already a pointer to T. This is useful for avoiding allocations when
// converting a value to a pointer, as the compiler can optimize away the
// allocation if the value is already a pointer.
func ToPtr[T any](v T) *T {
	// any 转 interface{}，检查是否是 *T
	if p, ok := any(v).(*T); ok {
		return p
	}
	return &v
}
