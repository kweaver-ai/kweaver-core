package util

// 根据size，截取字符串的前size部分，并添加...，形成新的字符串
// 示例： LeftTrimEllipsisSize("123456789", 5) -> "12..."
func LeftTrimEllipsisSize(str string, size int) (newStr string) {
	if size <= 3 {
		panic("size must be greater than 3")
	}

	runes := []rune(str)
	if len(runes) > size {
		newStr = string(runes[:size-3]) + "..."
	} else {
		newStr = str
	}

	return
}
