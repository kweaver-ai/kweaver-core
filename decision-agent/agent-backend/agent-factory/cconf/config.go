package cconf

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"gopkg.in/yaml.v3"
)

var _configPath string

func GetConfigPath() string {
	if _configPath != "" {
		return _configPath
	}

	_configPath = "/sysvol/conf/"
	if _, err := os.Stat(_configPath); os.IsNotExist(err) {
		_configPath = "./conf"
	}

	if cenvhelper.ConfigPathFromEnv() != "" {
		_configPath = cenvhelper.ConfigPathFromEnv()
	}

	return _configPath
}

type Config struct {
	Project Project   `yaml:"project"`
	DB      DBConf    `yaml:"db"`
	Redis   RedisConf `yaml:"redis"`
	Hydra   HydraCfg  `yaml:"hydra"`

	ModelFactory  *ModelFactoryConf `yaml:"model_factory"`
	Authorization *AuthzCfg         `yaml:"authorization"`
	AgentFactory  *AgentFactoryConf `yaml:"agent_factory"`
	BizDomain     *BizDomainConf    `yaml:"biz_domain"`

	MqCfgPath string
}

func (c *Config) IsDebug() bool {
	return cenvhelper.IsDebugMode()
}

func (c *Config) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("Config{error: %v}", err)
	}

	return "======= Config =======\n" + string(b) + "\n======= End Config ======="
}

func (c *Config) Check() (err error) {
	err = c.Project.Check()
	if err != nil {
		return
	}

	return
}

func (c *Config) GetDefaultLanguage() rest.Language {
	return c.Project.Language
}

// GetLogLevelString 获取日志级别字符串
func (c *Config) GetLogLevelString() string {
	return c.Project.LoggerLevel.String()
}

func BaseDefConfig() (defConf *Config) {
	defConf = &Config{
		Project: Project{
			Host:        "0.0.0.0",
			Port:        30777,
			Language:    rest.SimplifiedChinese,
			LoggerLevel: 1,
		},
		DB: DBConf{
			UserName:         "anyshare",
			Password:         "eisoo.com123",
			DBHost:           "",
			DBPort:           3330,
			DBName:           "dip_data_agent",
			Charset:          "utf8mb4",
			Timeout:          10,
			TimeoutRead:      10,
			TimeoutWrite:     10,
			MaxOpenConns:     30,
			MaxOpenReadConns: 30,
		},
		Redis: RedisConf{
			ConnectType:        "",
			UserName:           "",
			Password:           "",
			Host:               "",
			Port:               "",
			MasterGroupName:    "",
			SentinelHost:       "",
			SentinelPort:       "",
			SentinelUsername:   "",
			SentinelPwd:        "",
			MasterHost:         "",
			MasterPort:         "",
			SlaveHost:          "",
			SlavePort:          "",
			ClusterHosts:       nil,
			ClusterPwd:         "",
			DB:                 3,
			MaxRetries:         0,
			PoolSize:           0,
			ReadTimeout:        0,
			WriteTimeout:       0,
			IdleTimeout:        0,
			IdleCheckFrequency: 0,
			MaxConnAge:         0,
			PoolTimeout:        0,
		},
	}

	mqConfigPath := filepath.Join(GetConfigPath(), "mq_config.yaml")
	defConf.MqCfgPath = mqConfigPath

	return
}

func GetConfigBys(fileName string) []byte {
	configFilePath := filepath.Join(GetConfigPath(), fileName)
	log.Printf("Loading config file: %s\n", configFilePath)

	file, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("load %v failed: %v", configFilePath, err)
	}

	return file
}

func LoadConfig(file []byte, configImpl IConf) IConf {
	err := yaml.Unmarshal(file, configImpl)
	if err != nil {
		log.Fatalf("unmarshal yaml file failed: %v", err)
	}

	if configImpl.IsDebug() {
		conf := configImpl
		log.Println(conf)
	}

	return configImpl
}
