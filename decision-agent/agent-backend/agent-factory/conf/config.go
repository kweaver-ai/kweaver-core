package conf

import (
	"sync"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel"
)

type AuthConf struct {
	Mechanism string `yaml:"mechanism"`
	Password  string `yaml:"password"`
	Username  string `yaml:"username"`
}

type MQConf struct {
	Auth          AuthConf `yaml:"auth"`
	ConnectorType string   `yaml:"connectorType"`
	MQHost        string   `yaml:"mqHost"`
	MQLookupdHost string   `yaml:"mqLookupdHost"`
	MQLookupdPort int      `yaml:"mqLookupdPort"`
	MQPort        int      `yaml:"mqPort"`
	MQType        string   `yaml:"mqType"`
	Protocol      string   `yaml:"protocol"`
	Tenant        string   `yaml:"tenant"`
}

// TODO?
func (c MQConf) IsDebug() bool {
	return true
}

type Config struct {
	*cconf.Config
	MQ MQConf

	// APP 配置字段
	AgentFactoryConf    *AgentFactoryConf    `yaml:"agent_factory"`
	AgentExecutorConf   *AgentExecutorConf   `yaml:"agent_executor"`
	FastConf            *EfastConf           `yaml:"efast"`
	DocsetConf          *DocsetConf          `yaml:"docset"`
	EcoConfigConf       *EcoConfigConf       `yaml:"ecoconfig"`
	UniqueryConf        *UniqueryConf        `yaml:"uniquery"`
	SandboxPlatformConf *SandboxPlatformConf `yaml:"sandbox_platform"`

	// 流式响应配置
	StreamDiffFrequency int `yaml:"stream_diff_frequency"`

	// OpenTelemetry 配置
	OtelConfig *OtelConfig `yaml:"opentelemetry"`

	// 新版 OTel Collector 配置
	OtelV2Config *otel.OtelV2Config `yaml:"otel"`

	// 特性开关配置
	SwitchFields *SwitchFields `yaml:"switch_fields"`

	// OpenAPI 文档配置
	EnableSwagger bool   `yaml:"enable_swagger"` // 是否启用 API 文档 UI（兼容旧配置名）
	SwaggerToken  string `yaml:"swagger_token"`  // 文档 UI 认证 token（开发模式，当前保留）
}

func (c Config) IsDebug() bool {
	return cenvhelper.IsDebugMode()
}

func (c *Config) IsBizDomainDisabled() bool {
	return c != nil && c.SwitchFields.IsBizDomainDisabled()
}

var (
	configOnce sync.Once
	configImpl *Config
)

func NewConfig() *Config {
	configOnce.Do(func() {
		configImpl = &Config{}

		configImpl.SwitchFields = NewSwitchFields()

		configImpl.Config = cconf.BaseDefConfig()

		configImpl.OtelConfig = &OtelConfig{}
		configImpl.OtelV2Config = &otel.OtelV2Config{}

		bys := cconf.GetConfigBys("agent-factory.yaml")
		cconf.LoadConfig(bys, configImpl.Config)
		// 同时加载扩展字段（AgentFactoryConf等）
		cconf.LoadConfig(bys, configImpl)

		setOtelDefaults(configImpl.OtelConfig)
		configImpl.OtelV2Config.SetDefaults()

		secretBys := cconf.GetConfigBys("secret/agent-factory-secret.yaml")
		cconf.LoadConfig(secretBys, configImpl.Config)

		mqBys := cconf.GetConfigBys("mq_config.yaml")
		cconf.LoadConfig(mqBys, &configImpl.MQ)
	})

	return configImpl
}
