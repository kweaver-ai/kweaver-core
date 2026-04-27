package conf

import "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"

// SandboxPlatformConf Sandbox Platform 配置
type SandboxPlatformConf struct {
	Enable                  bool             `yaml:"enable"` // 是否启用沙箱功能
	PublicSvc               cconf.SvcConf    `yaml:"public_svc"`
	PrivateSvc              cconf.SvcConf    `yaml:"private_svc"`
	DefaultTTL              int64            `yaml:"default_ttl"`    // 默认 Session TTL（秒）
	MaxRetries              int              `yaml:"max_retries"`    // 等待 Session 就绪的最大重试次数
	RetryInterval           string           `yaml:"retry_interval"` // 重试间隔（如 "500ms"）
	DefaultFileUploadConfig FileUploadConfig `yaml:"file_upload_config"`
	DefaultTemplateID       string           `yaml:"default_template_id"` // 默认模板 ID
	DefaultCPU              string           `yaml:"default_cpu"`         // 默认 CPU 核心数（如 "1"）
	DefaultMemory           string           `yaml:"default_memory"`      // 默认内存（如 "512Mi"）
	DefaultDisk             string           `yaml:"default_disk"`        // 默认磁盘（如 "1Gi"）
	DefaultTimeout          int64            `yaml:"default_timeout"`     // 默认超时时间（秒）
}

// FileUploadConfig 文件上传配置
type FileUploadConfig struct {
	MaxFileSize      int64    `yaml:"max_file_size"`      // 最大文件大小（数值）
	MaxFileSizeUnit  string   `yaml:"max_file_size_unit"` // 单位：KB/MB/GB
	MaxFileCount     int      `yaml:"max_file_count"`     // 最大文件数量
	AllowedFileTypes []string `yaml:"allowed_file_types"` // 允许的文件类型
}
