// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/fsnotify/fsnotify"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	"github.com/kweaver-ai/kweaver-go-lib/otel"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/spf13/viper"

	"bkn-backend/version"
)

// ServerSetting server配置项
type ServerSetting struct {
	RunMode                  string        `mapstructure:"runMode"`
	HttpPort                 int           `mapstructure:"httpPort"`
	Language                 string        `mapstructure:"language"`
	ReadTimeOut              time.Duration `mapstructure:"readTimeOut"`
	WriteTimeout             time.Duration `mapstructure:"writeTimeOut"`
	MaxConcurrentTasks       int           `mapstructure:"maxConcurrentTasks"`
	ViewDataLimit            int           `mapstructure:"viewDataLimit"`
	DefaultSmallModelName    string        `mapstructure:"defaultSmallModelName"`
	DefaultSmallModelEnabled bool          `mapstructure:"defaultSmallModelEnabled"`
	JobMaxRetryTimes         int           `mapstructure:"jobMaxRetryTimes"`
	ReloadJobEnabled         bool          `mapstructure:"reloadJobEnabled"`
	// Schedule worker settings
	SchedulePollInterval int `mapstructure:"schedulePollInterval"` // in seconds, default 10
	ScheduleLockTimeout  int `mapstructure:"scheduleLockTimeout"`  // in seconds, default 300 (5 min)
}

// AppSetting app配置项
type AppSetting struct {
	ServerSetting ServerSetting             `mapstructure:"server"`
	LogSetting    logger.LogSetting         `mapstructure:"log"`
	OtelSetting   otel.OtelConfig           `mapstructure:"otel"`
	DepServices   map[string]map[string]any `mapstructure:"depServices"`

	DBSetting         libdb.DBSetting
	MQSetting         libmq.MQSetting
	OpenSearchSetting rest.OpenSearchClientConfig
	HydraAdminSetting hydra.HydraAdminSetting

	// data model url
	DataModelUrl string
	// data view url
	DataViewUrl string
	// UniQuery url
	UniQueryUrl string

	// permission url
	PermissionUrl string
	// user management url
	UserMgmtUrl string
	// model factory url
	ModelFactoryManagerUrl string
	// model factory api url
	ModelFactoryAPIUrl string
	// business system url
	BusinessSystemUrl string
	// ontology query url
	OntologyQueryUrl string
	// vega backend url
	VegaBackendUrl string
	// AgentOperatorUrl is the single agent-operator-integration internal-v1 base, e.g.
	// {scheme}://{host}:{port}/api/agent-operator-integration/internal-v1
	// (A trailing /operator suffix is accepted for backward compatibility and normalized away.)
	AgentOperatorUrl string
}

const (
	// ConfigFile 配置文件信息
	configPath string = "./config/"
	configName string = "bkn-backend-config"
	configType string = "yaml"

	rdsServiceName                 string = "rds"
	mqServiceName                  string = "mq"
	opensearchServiceName          string = "opensearch"
	permissionServiceName          string = "authorization-private"
	userMgmtServiceName            string = "user-management"
	hydraAdminServiceName          string = "hydra-admin"
	modelFactoryManagerServiceName string = "mf-model-manager"
	modelFactoryAPIServiceName     string = "mf-model-api"
	dataModelServiceName           string = "data-model"
	dataViewServiceName            string = "data-model"
	uniQueryServiceName            string = "uniquery"
	businessSystemServiceName      string = "business-system"
	ontologyQueryServiceName       string = "ontology-query"
	vegaBackendServiceName         string = "vega-backend"
	agentOperatorServiceName       string = "agent-operator-integration"

	DATA_BASE_NAME string = "kweaver"
)

var (
	appSetting *AppSetting
	vp         *viper.Viper

	settingOnce sync.Once
)

// NewSetting 读取服务配置
func NewSetting() *AppSetting {
	settingOnce.Do(func() {
		appSetting = &AppSetting{}
		vp = viper.New()
		initSetting(vp)
	})

	return appSetting
}

// 初始化配置
func initSetting(vp *viper.Viper) {
	logger.Infof("Init Setting From File %s%s.%s", configPath, configName, configType)

	vp.AddConfigPath(configPath)
	vp.SetConfigName(configName)
	vp.SetConfigType(configType)

	loadSetting(vp)

	vp.WatchConfig()
	vp.OnConfigChange(func(e fsnotify.Event) {
		logger.Infof("Config file changed:%s", e)
		loadSetting(vp)
	})
}

// 读取配置文件
func loadSetting(vp *viper.Viper) {
	logger.Infof("Load Setting File %s%s.%s", configPath, configName, configType)

	if err := vp.ReadInConfig(); err != nil {
		logger.Fatalf("err:%s\n", err)
	}

	if err := vp.Unmarshal(appSetting); err != nil {
		logger.Fatalf("err:%s\n", err)
	}

	SetLogSetting(appSetting.LogSetting)

	SetDBSetting()

	SetMQSetting()

	SetOpenSearchSetting()

	SetHydraAdminSetting()

	SetPermissionSetting()

	SetUserMgmtSetting()

	SetDataModelSetting()

	SetDataViewSetting()

	SetUniQuerySetting()

	SetModelFactoryManagerSetting()

	SetModelFactoryAPISetting()

	SetBusinessSystemSetting()

	SetOntologyQuerySetting()

	SetVegaBackendSetting()

	SetAgentOperatorSetting()

	logger.Infof("ServerName: %s, ServerVersion: %s, Language: %s, GoVersion: %s, GoArch: %s",
		version.ServerName, version.ServerVersion, version.LanguageGo,
		version.GoVersion, version.GoArch)

	s, _ := sonic.MarshalString(appSetting)
	logger.Debug(s)
}

func SetDBSetting() {
	setting, ok := appSetting.DepServices[rdsServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", rdsServiceName)
	}

	appSetting.DBSetting = libdb.DBSetting{
		Host:     setting["host"].(string),
		Port:     setting["port"].(int),
		Username: setting["user"].(string),
		Password: setting["password"].(string),
		DBName:   DATA_BASE_NAME,
	}
}

func SetMQSetting() {
	setting, ok := appSetting.DepServices[mqServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", mqServiceName)
	}
	authSetting, ok := setting["auth"].(map[string]any)
	if !ok {
		logger.Fatalf("service %s auth not found in depServices", mqServiceName)
	}

	appSetting.MQSetting = libmq.MQSetting{
		MQType: setting["mqtype"].(string),
		MQHost: setting["mqhost"].(string),
		MQPort: setting["mqport"].(int),
		Tenant: setting["tenant"].(string),
		Auth: libmq.MQAuthSetting{
			Username:  authSetting["username"].(string),
			Password:  authSetting["password"].(string),
			Mechanism: authSetting["mechanism"].(string),
		},
	}
}

func SetOpenSearchSetting() {
	setting, ok := appSetting.DepServices[opensearchServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", opensearchServiceName)
	}

	appSetting.OpenSearchSetting = rest.OpenSearchClientConfig{
		Host:     setting["host"].(string),
		Port:     setting["port"].(int),
		Protocol: setting["protocol"].(string),
		Username: setting["user"].(string),
		Password: setting["password"].(string),
	}
}

// GetAuthEnabled 获取认证开关状态
// 通过环境变量 AUTH_ENABLED 控制，默认 true（安全优先）
func GetAuthEnabled() bool {
	envVal := os.Getenv("AUTH_ENABLED")
	// 仅当显式设置为 false 或 0 时禁用认证
	return envVal != "false" && envVal != "0"
}

func SetHydraAdminSetting() {
	if !GetAuthEnabled() {
		logger.Info("ISF authentication disabled via AUTH_ENABLED env, skipping hydra-admin configuration")
		return
	}
	setting, ok := appSetting.DepServices[hydraAdminServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", hydraAdminServiceName)
	}
	appSetting.HydraAdminSetting = hydra.HydraAdminSetting{
		HydraAdminProcotol: setting["protocol"].(string),
		HydraAdminHost:     setting["host"].(string),
		HydraAdminPort:     setting["port"].(int),
	}
}

func SetPermissionSetting() {
	if !GetAuthEnabled() {
		logger.Info("ISF authentication disabled via AUTH_ENABLED env, skipping authorization configuration")
		return
	}
	setting, ok := appSetting.DepServices[permissionServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", permissionServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.PermissionUrl = fmt.Sprintf("%s://%s:%d/api/authorization/v1", protocol, host, port)
}

func SetUserMgmtSetting() {
	if !GetAuthEnabled() {
		logger.Info("ISF authentication disabled via AUTH_ENABLED env, skipping user-management configuration")
		return
	}
	setting, ok := appSetting.DepServices[userMgmtServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", userMgmtServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.UserMgmtUrl = fmt.Sprintf("%s://%s:%d", protocol, host, port)
}

func SetDataModelSetting() {
	setting, ok := appSetting.DepServices[dataModelServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", dataModelServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.DataModelUrl = fmt.Sprintf("%s://%s:%d/api/mdl-data-model/in/v1", protocol, host, port)
}

func SetDataViewSetting() {
	setting, ok := appSetting.DepServices[dataViewServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", dataViewServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.DataViewUrl = fmt.Sprintf("%s://%s:%d/api/mdl-data-model/in/v1", protocol, host, port)
}

func SetUniQuerySetting() {
	setting, ok := appSetting.DepServices[uniQueryServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", uniQueryServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.UniQueryUrl = fmt.Sprintf("%s://%s:%d/api/mdl-uniquery/in/v1", protocol, host, port)
}

func SetModelFactoryManagerSetting() {
	setting, ok := appSetting.DepServices[modelFactoryManagerServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", modelFactoryManagerServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.ModelFactoryManagerUrl = fmt.Sprintf("%s://%s:%d/api/private/mf-model-manager/v1", protocol, host, port)
}

func SetModelFactoryAPISetting() {
	setting, ok := appSetting.DepServices[modelFactoryAPIServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", modelFactoryAPIServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.ModelFactoryAPIUrl = fmt.Sprintf("%s://%s:%d/api/private/mf-model-api/v1", protocol, host, port)
}

// GetBusinessDomainEnabled 获取业务域开关状态
// 通过环境变量 BUSINESS_DOMAIN_ENABLED 控制，默认 true（安全优先）
func GetBusinessDomainEnabled() bool {
	envVal := os.Getenv("BUSINESS_DOMAIN_ENABLED")
	return envVal != "false" && envVal != "0"
}

func SetBusinessSystemSetting() {
	if !GetBusinessDomainEnabled() {
		logger.Info("Business domain disabled via BUSINESS_DOMAIN_ENABLED env, skipping business-system configuration")
		return
	}
	setting, ok := appSetting.DepServices[businessSystemServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", businessSystemServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.BusinessSystemUrl = fmt.Sprintf("%s://%s:%d/internal/api/business-system/v1", protocol, host, port)
}

func SetOntologyQuerySetting() {
	setting, ok := appSetting.DepServices[ontologyQueryServiceName]
	if !ok {
		// Optional service, default to localhost for development
		logger.Warnf("service %s not found in depServices, using default", ontologyQueryServiceName)
		appSetting.OntologyQueryUrl = "http://localhost:8080"
		return
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.OntologyQueryUrl = fmt.Sprintf("%s://%s:%d", protocol, host, port)
}

func SetVegaBackendSetting() {
	setting, ok := appSetting.DepServices[vegaBackendServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaBackendServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.VegaBackendUrl = fmt.Sprintf("%s://%s:%d/api/vega-backend/in/v1", protocol, host, port)
}

func SetAgentOperatorSetting() {
	setting, ok := appSetting.DepServices[agentOperatorServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", agentOperatorServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.AgentOperatorUrl = fmt.Sprintf("%s://%s:%d/api/agent-operator-integration/internal-v1", protocol, host, port)
}
