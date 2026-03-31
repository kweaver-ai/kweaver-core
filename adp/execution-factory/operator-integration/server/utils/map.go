package utils

// GetValueOrDefault 获取map中key对应的值，如果不存在则返回默认值
func GetValueOrDefault(m map[string]string, key, defaultValue string) string {
	if key == "" {
		return defaultValue
	}
	if value, exists := m[key]; exists {
		return value
	}
	if defaultValue == "" {
		return key
	}
	return defaultValue
}
