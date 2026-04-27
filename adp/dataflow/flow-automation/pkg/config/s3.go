package config

import (
	"fmt"
	"os"
	"strconv"
)

// S3Config S3全局配置
type S3Config struct {
	Endpoint        string // S3服务端点
	Region          string // S3区域
	AccessKeyID     string // 访问密钥ID
	SecretAccessKey string // 访问密钥
	BucketName      string // Bucket名称
	UseSSL          bool   // 是否使用SSL
	SkipVerify      bool   // 是否跳过SSL验证
}

// LoadS3Config 从环境变量加载S3配置
func LoadS3Config() (*S3Config, error) {
	config := &S3Config{
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Region:          os.Getenv("S3_REGION"),
		AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		BucketName:      os.Getenv("S3_BUCKET_NAME"),
		UseSSL:          true,  // 默认使用SSL
		SkipVerify:      false, // 默认不跳过验证
	}

	// 解析SSL配置
	if useSSLStr := os.Getenv("S3_USE_SSL"); useSSLStr != "" {
		useSSL, err := strconv.ParseBool(useSSLStr)
		if err != nil {
			return nil, fmt.Errorf("invalid S3_USE_SSL value: %w", err)
		}
		config.UseSSL = useSSL
	}

	// 解析SkipVerify配置
	if skipVerifyStr := os.Getenv("S3_SKIP_VERIFY"); skipVerifyStr != "" {
		skipVerify, err := strconv.ParseBool(skipVerifyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid S3_SKIP_VERIFY value: %w", err)
		}
		config.SkipVerify = skipVerify
	}

	// 验证必需配置
	if config.Endpoint == "" {
		return nil, fmt.Errorf("S3_ENDPOINT is required")
	}
	if config.Region == "" {
		return nil, fmt.Errorf("S3_REGION is required")
	}
	if config.AccessKeyID == "" {
		return nil, fmt.Errorf("S3_ACCESS_KEY_ID is required")
	}
	if config.SecretAccessKey == "" {
		return nil, fmt.Errorf("S3_SECRET_ACCESS_KEY is required")
	}
	if config.BucketName == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME is required")
	}

	return config, nil
}

// IsConfigured 检查S3配置是否已配置
func IsS3Configured() bool {
	return os.Getenv("S3_ENDPOINT") != "" &&
		os.Getenv("S3_ACCESS_KEY_ID") != "" &&
		os.Getenv("S3_SECRET_ACCESS_KEY") != ""
}
