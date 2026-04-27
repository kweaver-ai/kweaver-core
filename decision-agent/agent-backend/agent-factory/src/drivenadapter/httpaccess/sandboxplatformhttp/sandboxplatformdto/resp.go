package sandboxplatformdto

// CreateSessionResp 创建 Session 响应
type CreateSessionResp struct {
	ID             string            `json:"id"`
	TemplateID     string            `json:"template_id"`
	Status         string            `json:"status"`
	ResourceLimit  *ResourceLimit    `json:"resource_limit,omitempty"`
	WorkspacePath  *string           `json:"workspace_path,omitempty"`
	RuntimeType    string            `json:"runtime_type"`
	RuntimeNode    *string           `json:"runtime_node,omitempty"`
	ContainerID    *string           `json:"container_id,omitempty"`
	PodName        *string           `json:"pod_name,omitempty"`
	EnvVars        map[string]string `json:"env_vars,omitempty"`
	Timeout        int               `json:"timeout"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
	CompletedAt    *string           `json:"completed_at,omitempty"`
	LastActivityAt *string           `json:"last_activity_at,omitempty"`
}

// GetSessionResp 获取 Session 响应
type GetSessionResp struct {
	ID             string            `json:"id"`
	TemplateID     string            `json:"template_id"`
	Status         string            `json:"status"`
	ResourceLimit  *ResourceLimit    `json:"resource_limit,omitempty"`
	WorkspacePath  *string           `json:"workspace_path,omitempty"`
	RuntimeType    string            `json:"runtime_type"`
	RuntimeNode    *string           `json:"runtime_node,omitempty"`
	ContainerID    *string           `json:"container_id,omitempty"`
	PodName        *string           `json:"pod_name,omitempty"`
	EnvVars        map[string]string `json:"env_vars,omitempty"`
	Timeout        int               `json:"timeout"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
	CompletedAt    *string           `json:"completed_at,omitempty"`
	LastActivityAt *string           `json:"last_activity_at,omitempty"`
}

// ResourceLimit 资源限制响应
type ResourceLimit struct {
	CPU          string `json:"cpu"`
	Memory       string `json:"memory"`
	Disk         string `json:"disk"`
	MaxProcesses *int   `json:"max_processes,omitempty"`
}
