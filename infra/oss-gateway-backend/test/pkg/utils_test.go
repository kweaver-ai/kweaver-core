package pkg_test

import (
	"oss-gateway/pkg/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateStorageID(t *testing.T) {
	id := utils.GenerateStorageID()

	assert.NotEmpty(t, id)
	assert.True(t, len(id) > 0)
}

func TestGenerateStorageID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)

	// 生成100个ID，应该都是唯一的
	for i := 0; i < 100; i++ {
		id := utils.GenerateStorageID()
		assert.False(t, ids[id], "Duplicate ID generated: %s", id)
		ids[id] = true
	}

	assert.Equal(t, 100, len(ids))
}

func TestParseEndpoint_HTTPS(t *testing.T) {
	endpoint := "https://oss-cn-hangzhou.aliyuncs.com"
	result, useSSL := utils.ParseEndpoint(endpoint)

	assert.Equal(t, "oss-cn-hangzhou.aliyuncs.com", result)
	assert.True(t, useSSL)
}

func TestParseEndpoint_HTTP(t *testing.T) {
	endpoint := "http://oss-cn-hangzhou.aliyuncs.com"
	result, useSSL := utils.ParseEndpoint(endpoint)

	assert.Equal(t, "oss-cn-hangzhou.aliyuncs.com", result)
	assert.False(t, useSSL)
}

func TestParseEndpoint_NoScheme(t *testing.T) {
	endpoint := "oss-cn-hangzhou.aliyuncs.com"
	result, useSSL := utils.ParseEndpoint(endpoint)

	assert.Equal(t, "oss-cn-hangzhou.aliyuncs.com", result)
	assert.True(t, useSSL) // Default to HTTPS
}

func TestParseEndpoint_WithWhitespace(t *testing.T) {
	endpoint := "  https://oss-cn-hangzhou.aliyuncs.com  "
	result, useSSL := utils.ParseEndpoint(endpoint)

	assert.Equal(t, "oss-cn-hangzhou.aliyuncs.com", result)
	assert.True(t, useSSL)
}

func TestParseEndpoint_WithPort(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected string
		useSSL   bool
	}{
		{
			name:     "HTTPS with port",
			endpoint: "https://localhost:9000",
			expected: "localhost:9000",
			useSSL:   true,
		},
		{
			name:     "HTTP with port",
			endpoint: "http://localhost:9000",
			expected: "localhost:9000",
			useSSL:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, useSSL := utils.ParseEndpoint(tt.endpoint)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.useSSL, useSSL)
		})
	}
}

func TestCalculateExpiresAt(t *testing.T) {
	now := time.Now()
	validSeconds := int64(3600) // 1 hour

	expiresAt := utils.CalculateExpiresAt(validSeconds)

	// 验证过期时间大约是1小时后
	duration := expiresAt.Sub(now)
	assert.InDelta(t, float64(validSeconds), duration.Seconds(), 1.0)
}

func TestCalculateExpiresAt_Zero(t *testing.T) {
	now := time.Now()
	validSeconds := int64(0)

	expiresAt := utils.CalculateExpiresAt(validSeconds)

	// 过期时间应该约等于当前时间
	duration := expiresAt.Sub(now)
	assert.InDelta(t, 0, duration.Seconds(), 1.0)
}

func TestCalculateExpiresAt_Negative(t *testing.T) {
	now := time.Now()
	validSeconds := int64(-3600) // -1 hour

	expiresAt := utils.CalculateExpiresAt(validSeconds)

	// 过期时间应该是过去的时间
	assert.True(t, expiresAt.Before(now))
}

func TestCalculateExpiresAt_LargeValue(t *testing.T) {
	now := time.Now()
	validSeconds := int64(7 * 24 * 3600) // 7 days

	expiresAt := utils.CalculateExpiresAt(validSeconds)

	duration := expiresAt.Sub(now)
	assert.InDelta(t, float64(validSeconds), duration.Seconds(), 1.0)
}

func TestValidateStorageID_Valid(t *testing.T) {
	tests := []string{
		"storage123",
		"A1B2C3D4E5F6G7H8",
		"1234567890123456789",
		"a",
		"test-storage-id",
	}

	for _, id := range tests {
		t.Run(id, func(t *testing.T) {
			err := utils.ValidateStorageID(id)
			assert.NoError(t, err)
		})
	}
}

func TestValidateStorageID_Empty(t *testing.T) {
	err := utils.ValidateStorageID("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestValidateStorageID_TooLong(t *testing.T) {
	// 生成一个超过64字符的ID
	longID := ""
	for i := 0; i < 65; i++ {
		longID += "a"
	}

	err := utils.ValidateStorageID(longID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too long")
}

func TestValidateStorageID_ExactlyMaxLength(t *testing.T) {
	// 生成一个刚好64字符的ID
	maxID := ""
	for i := 0; i < 64; i++ {
		maxID += "a"
	}

	err := utils.ValidateStorageID(maxID)
	assert.NoError(t, err)
}

func TestValidateObjectKey_Valid(t *testing.T) {
	tests := []string{
		"test/file.txt",
		"folder/subfolder/file.pdf",
		"image.jpg",
		"测试文件.doc",
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
	}

	for _, key := range tests {
		t.Run(key, func(t *testing.T) {
			err := utils.ValidateObjectKey(key)
			assert.NoError(t, err)
		})
	}
}

func TestValidateObjectKey_Empty(t *testing.T) {
	err := utils.ValidateObjectKey("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestValidateObjectKey_TooLong(t *testing.T) {
	// 生成一个超过512字符的key
	longKey := ""
	for i := 0; i < 513; i++ {
		longKey += "a"
	}

	err := utils.ValidateObjectKey(longKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too long")
}

func TestValidateObjectKey_ExactlyMaxLength(t *testing.T) {
	// 生成一个刚好512字符的key
	maxKey := ""
	for i := 0; i < 512; i++ {
		maxKey += "a"
	}

	err := utils.ValidateObjectKey(maxKey)
	assert.NoError(t, err)
}
