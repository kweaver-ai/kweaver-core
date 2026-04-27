package mocks

import "fmt"

// MockFuncErr mock func error
func MockFuncErr(errStr string) error {
	return fmt.Errorf("mock %s error", errStr)
}

const (
	asciiLen = 26
)

// MockDescription mock description str
func MockDescription(l int64) string {
	result := make([]byte, l)
	for i := int64(0); i < l; i++ {
		result[i] = 'a' + byte(i%asciiLen)
	}
	return string(result)
}
