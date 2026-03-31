// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package setup

import (
	"fmt"

	"github.com/spf13/viper"
)

// TestConfig AT测试配置
type TestConfig struct {
	BKNBackend BKNBackendConfig `mapstructure:"bkn_backend"`
	MariaDB    MariaDBConfig    `mapstructure:"mariadb"`
	OpenSearch OpenSearchConfig `mapstructure:"opensearch"`
}

// BKNBackendConfig BKN Backend服务配置
type BKNBackendConfig struct {
	BaseURL string `mapstructure:"base_url"` // BKN Backend HTTP服务地址
}

// MariaDBConfig 测试目标MariaDB配置
type MariaDBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// OpenSearchConfig 测试目标OpenSearch配置
type OpenSearchConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	UseSSL   bool   `mapstructure:"use_ssl"`
}

// LoadTestConfig 加载测试配置
// 优先从 testdata/test-config.yaml 读取
// 支持环境变量覆盖 (BKN_TEST_前缀)
func LoadTestConfig() (*TestConfig, error) {
	viper.SetConfigName("test-config")
	viper.SetConfigType("yaml")

	// 添加多个可能的配置文件路径
	viper.AddConfigPath("./testdata")                         // 从测试目录运行
	viper.AddConfigPath("./integration_tests/testdata")       // 从tests目录运行
	viper.AddConfigPath("./tests/integration_tests/testdata") // 从server目录运行
	viper.AddConfigPath("../testdata")                        // 从子目录运行
	viper.AddConfigPath("../../testdata")                     // 从深层子目录运行
	viper.AddConfigPath("../../../testdata")                  // 从深层子目录运行

	// 支持环境变量覆盖
	viper.SetEnvPrefix("BKN_TEST")
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取test-config.yaml失败: %w\n提示: 请确保配置文件存在于tests/integration_tests/testdata/目录", err)
	}

	var config TestConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证必填字段
	if config.BKNBackend.BaseURL == "" {
		return nil, fmt.Errorf("配置错误: bkn_backend.base_url 不能为空")
	}

	return &config, nil
}
