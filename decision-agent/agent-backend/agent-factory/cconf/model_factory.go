package cconf

type ModelFactoryConf struct {
	ModelApiSvc     SvcConf `yaml:"model_api_svc"`
	ModelManagerSvc SvcConf `yaml:"model_manager_svc"`
	LLM             LLMConf `yaml:"llm"` // LLM配置
}

// LLMConf LLM 相关配置
type LLMConf struct {
	DefaultModelName string `yaml:"default_model_name"` // 默认模型名称
	APIKey           string `yaml:"api_key"`            // API Key
}
