package utils

import (
	"fmt"
	"oss-gateway/pkg/snowflake"
	"strconv"
	"strings"
	"time"
)

// GenerateStorageID 生成存储ID（使用19位雪花ID）
func GenerateStorageID() string {
	id, err := snowflake.GenerateID()
	if err != nil {
		// 如果雪花ID生成失败，降级为时间戳+随机数
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return strconv.FormatInt(id, 10)
}

func ParseEndpoint(endpoint string) (string, bool) {
	endpoint = strings.TrimSpace(endpoint)

	if strings.HasPrefix(endpoint, "https://") {
		return strings.TrimPrefix(endpoint, "https://"), true
	}

	if strings.HasPrefix(endpoint, "http://") {
		return strings.TrimPrefix(endpoint, "http://"), false
	}

	return endpoint, true
}

func CalculateExpiresAt(validSeconds int64) time.Time {
	return time.Now().Add(time.Duration(validSeconds) * time.Second)
}

func ValidateStorageID(storageID string) error {
	if storageID == "" {
		return fmt.Errorf("storage_id cannot be empty")
	}
	if len(storageID) > 64 {
		return fmt.Errorf("storage_id too long (max 64 characters)")
	}
	return nil
}

func ValidateObjectKey(objectKey string) error {
	if objectKey == "" {
		return fmt.Errorf("object_key cannot be empty")
	}
	if len(objectKey) > 512 {
		return fmt.Errorf("object_key too long (max 512 characters)")
	}
	return nil
}
