package sandboxplatformdto

// CreateSessionReq 创建 Session 请求
// 按照 sandbox-design-v2.1.md 章节 5.3.1 设计，扩展支持依赖安装
type CreateSessionReq struct {
	ID                    *string                `json:"id,omitempty"`                       // 会话 ID（可选，手动指定时需确保唯一性），最大长度64
	TemplateID            string                 `json:"template_id"`                        // 模板 ID（必填）
	Timeout               int                    `json:"timeout,omitempty"`                  // 超时时间（秒），默认 300，最大 3600
	CPU                   string                 `json:"cpu,omitempty"`                      // CPU 核心数，如 "1", "2"，默认 "1"
	Memory                string                 `json:"memory,omitempty"`                   // 内存限制，如 "512Mi", "1Gi"，默认 "512Mi"
	Disk                  string                 `json:"disk,omitempty"`                     // 磁盘限制，如 "1Gi", "10Gi"，默认 "1Gi"
	EnvVars               map[string]string      `json:"env_vars,omitempty"`                 // 环境变量字典
	Event                 map[string]interface{} `json:"event,omitempty"`                    // 事件数据
	Dependencies          []DependencySpec       `json:"dependencies,omitempty"`             // 会话级依赖包列表
	InstallTimeout        int                    `json:"install_timeout,omitempty"`          // 依赖安装超时时间（秒），默认 300，范围 30-1800
	FailOnDependencyError *bool                  `json:"fail_on_dependency_error,omitempty"` // 依赖安装失败时是否终止会话创建，默认 true
	AllowVersionConflicts *bool                  `json:"allow_version_conflicts,omitempty"`  // 是否允许版本冲突，默认 false
}

// DependencySpec 依赖包规范
type DependencySpec struct {
	Name    string  `json:"name"`
	Version *string `json:"version,omitempty"`
}
