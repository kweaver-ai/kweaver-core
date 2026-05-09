package conf

// SwitchFields 定义了系统中的各种开关配置
type SwitchFields struct {
	// 是否保留老的APP路径，默认false
	KeepLegacyAppPath bool `yaml:"keep_legacy_app_path"`

	// 是否禁用权限检查，默认false
	DisablePmsCheck bool `yaml:"disable_pms_check"`

	// 是否禁用业务域，默认false；开启后业务域相关逻辑全部失效
	DisableBizDomain bool `yaml:"disable_biz_domain"`

	// 是否禁用审计日志初始化，默认false
	DisableAuditInit bool `yaml:"disable_audit_init"`

	// Mock服务的开关配置
	Mock *MockSwitchFields `yaml:"mock"`
}

func NewSwitchFields() *SwitchFields {
	return &SwitchFields{
		Mock: &MockSwitchFields{},
	}
}

func (sf *SwitchFields) IsBizDomainDisabled() bool {
	return sf != nil && sf.DisableBizDomain
}

// MockSwitchFields 定义了各种Mock服务的开关配置
type MockSwitchFields struct {
	// 是否使用Mock MQ客户端（本地开发时建议设置为true）
	MockMQClient bool `yaml:"mock_mq_client"`

	// 是否使用Mock Sandbox Platform（本地开发时建议设置为true）
	MockSandboxPlatform bool `yaml:"mock_sandbox_platform"`

	// 是否使用Mock Hydra（本地开发时建议设置为true）
	MockHydra bool `yaml:"mock_hydra"`

	// 是否使用Mock AuthZ（本地开发时建议设置为true）
	MockAuthZ bool `yaml:"mock_authz"`

	// 是否使用Mock BizDomain（本地开发时建议设置为true）
	MockBizDomain bool `yaml:"mock_biz_domain"`

	// 是否使用 Mock UserManagerModule（本地开发时建议设置为true）
	MockUserManagerModule bool `yaml:"mock_user_manager_module"`

	// Mock Hydra 返回的用户 ID
	MockUserID string `yaml:"mock_user_id"`
}
