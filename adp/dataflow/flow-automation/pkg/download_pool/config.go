// Package download_pool 文件下载线程池
package download_pool

import (
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
)

// Config 下载线程池配置
type Config struct {
	WorkerCount     int
	PollInterval    int
	BatchSize       int
	DownloadTimeout int
	MaxFileSize     int64
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		WorkerCount:     3,
		PollInterval:    5,
		BatchSize:       10,
		DownloadTimeout: 300,
		MaxFileSize:     100 * 1024 * 1024, // 100MB
	}
}

// LoadConfig 从全局配置加载
func LoadConfig() *Config {
	cfg := common.NewConfig()
	downloadCfg := cfg.FlowFileDownload

	config := DefaultConfig()

	if downloadCfg.WorkerCount > 0 {
		config.WorkerCount = downloadCfg.WorkerCount
	}
	if downloadCfg.PollInterval > 0 {
		config.PollInterval = downloadCfg.PollInterval
	}
	if downloadCfg.BatchSize > 0 {
		config.BatchSize = downloadCfg.BatchSize
	}
	if downloadCfg.DownloadTimeout > 0 {
		config.DownloadTimeout = downloadCfg.DownloadTimeout
	}
	if downloadCfg.MaxFileSize > 0 {
		config.MaxFileSize = downloadCfg.MaxFileSize
	}

	return config
}