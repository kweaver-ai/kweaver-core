package common

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	wlCommon "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/common"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var config Config

// Config config info
type Config struct {
	ActionConfig             ActionConfig
	ActionConfigYaml         string                   `mapstructure:"action_config_yaml"`
	Lang                     string                   `mapstructure:"lang"`
	Debug                    string                   `mapstructure:"debug"`
	Server                   Server                   `mapstructure:"server"`
	DB                       DB                       `mapstructure:"db"`
	OAuth                    OAuth                    `mapstructure:"oauth"`
	Authentication           Authentication           `mapstructure:"authentication"`
	DeployService            DeployService            `mapstructure:"deployservice"`
	Efast                    Efast                    `mapstructure:"efast"`
	DocShare                 DocShare                 `mapstructure:"docshare"`
	Emetadata                Emetadata                `mapstructure:"emetadata"`
	UserManagement           UserManagement           `mapstructure:"usermanagement"`
	MQ                       MQ                       `mapstructure:"mq"`
	Kafka                    Kafka                    `mapstructure:"kafka"`
	MongoDB                  MongoDBConfig            `mapstructure:"mongodb"`
	Tika                     Tika                     `mapstructure:"tika"`
	DocConvert               DocConvert               `mapstructure:"doc_convert"`
	FastTextAnalysis         FastTextAnalysis         `mapstructure:"fasttextanalysis"`
	Document                 Document                 `mapstructure:"document"`
	T4th                     T4th                     `mapstructure:"t4th"`
	CodeRunner               CodeRunner               `mapstructure:"coderunner"`
	Telemetry                wlCommon.TelemetryConf   `mapstructure:"telemetry"`
	Appstore                 Appstore                 `mapstructure:"appstore"`
	Intelliinfo              Intelliinfo              `mapstructure:"intelliinfo"`
	ECron                    ECron                    `mapstructure:"ecron"`
	EcoTag                   EcoTag                   `mapstructure:"ecotag"`
	ContentAutomation        ContentAutomation        `mapstructure:"contentautomation"`
	OssGateWay               OSSGateWay               `mapstructure:"ossgateway"`
	OssGatewayBackend        OssGatewayBackend        `mapstructure:"ossgatewaybackend"`
	DumpLog                  DumpLog                  `mapstructure:"dumplog"`
	Redis                    RedisConfiguration       `mapstructure:"redis"`
	Kcmc                     Kcmc                     `mapstructure:"kcmc"`
	ShareMgnt                ShareMgnt                `mapstructure:"sharemgnt"`
	SpeechModel              SpeechModel              `mapstructure:"speechmodel"`
	CognitiveAssistant       CognitiveAssistant       `mapstructure:"cognitiveassistant"`
	DocSet                   DocSet                   `mapstructure:"docset"`
	PersonalConfig           PersonalConfig           `mapstructure:"personalconfig"`
	Metadata                 Metadata                 `mapstructure:"metadata"`
	Uie                      Uie                      `mapstructure:"uie"`
	AnyData                  AnyData                  `mapstructure:"anydata"`
	DataFlowTools            DataFlowTools            `mapstructure:"dataflowtools"`
	Ecoconfig                Ecoconfig                `mapstructure:"ecoconfig"`
	AgentOperatorIntegration AgentOperatorIntegration `mapstructure:"agentoperatorintegration"`
	MfModelManager           MfModelManager           `mapstructure:"mfmodelmanager"`
	AgentFactory             AgentFactory             `mapstructure:"agentfactory"`
	KnKnowledgeData          KnKnowledgeData          `mapstructure:"knknowledgedata"`
	AgentApp                 AgentApp                 `mapstructure:"agentapp"`
	MdlUniquery              MdlUniquery              `mapstructure:"mdl_uniquery"`
	MdlDataModel             MdlDataModel             `mapstructure:"mdl_datamodel"`
	MdlDataPipeline          MdlDataPipeline          `mapstructure:"mdl_datapipeline"`
	AuthorizationConf        AuthorizationConf        `mapstructure:"authorization"`
	OpenSearch               OpenSearch               `mapstructure:"opensearch"`
	DPDataSource             DPDataSource             `mapstructure:"dp_datasource"`
	ContentPipeline          ContentPipeline          `mapstructure:"contentpipeline"`
	OCR                      OCR                      `mapstructure:"ocr"`
	StructureExtractor       StructureExtractor       `mapstructure:"structureextractor"`
	MfModelApi               MfModelApi               `mapstructure:"mfmodelapi"`
	BusinessDomain           BusinessDomain           `mapstructure:"business_domain"`
	AccessAddress            AccessAddress            `mapstructure:"access_address"`
	Sandbox                  Sandbox                  `mapstructure:"sandbox"`
	FlowFileDownload         FlowFileDownload         `mapstructure:"flow_file_download"`
	VegaBackendConfig        VegaBackendConfig        `mapstructure:"vegabackend"`
}

type DagInstanceEventArchivePolicy string

const (
	DagInstanceEventArchivePolicyNever       DagInstanceEventArchivePolicy = "never"
	DagInstanceEventArchivePolicySuccessOnly DagInstanceEventArchivePolicy = "successOnly"
)

type Edition string

const (
	EditionCommunity  Edition = "community"
	EditionCommercial Edition = "commercial"
)

type StorageBackend string

const (
	StorageBackendOssGateway        StorageBackend = "ossgateway"
	StorageBackendS3                StorageBackend = "s3"
	StorageBackendOssGatewayBackend StorageBackend = "ossgatewaybackend"
)

// Server 服务基础配置
type Server struct {
	LowestExecutorCount           int                           `mapstructure:"lowest_executor_count"`
	LowExecutorCount              int                           `mapstructure:"low_executor_count"`
	MediumExecutorCount           int                           `mapstructure:"medium_executor_count"`
	HighExecutorCount             int                           `mapstructure:"high_executor_count"`
	HighestExecutorCount          int                           `mapstructure:"highest_executor_count"`
	DebugExecutorCount            int                           `mapstructure:"debug_executor_count"`
	ParserrCount                  int                           `mapstructure:"parser_count"`
	ExecutorTimeout               int                           `mapstructure:"executor_timeout"`
	ScheduleTimeout               int                           `mapstructure:"schedule_timeout"`
	ListInsCount                  int                           `mapstructure:"listins_count"`
	MongoMaxInlineSize            int                           `mapstructure:"mongo_max_inline_size"`
	StoragePrefix                 string                        `mapstructure:"storage_prefix"`
	DeleteExtDataCron             string                        `mapstructure:"delete_ext_data_cron"`
	DeleteExpiredTaskCache        string                        `mapstructure:"delete_expired_task_cache_cron"`
	DagInstanceEventMaxInlineSize int                           `mapstructure:"dag_instance_event_max_inline_size"`
	DagInstanceEventArchivePolicy DagInstanceEventArchivePolicy `mapstructure:"dag_instance_event_archive_policy"`
	Edition                       Edition                       `mapstructure:"edition"`
	DBType                        string                        `mapstructure:"db_type"`
	AuthEnabled                   string                        `mapstructure:"auth_enabled"`
	BusinessDomainEnabled         string                        `mapstructure:"businessdomain_enabled"`
	StorageBackend                StorageBackend                `mapstructure:"storage_backend"`
}

// DB database config
type DB struct {
	HOST     string `mapstructure:"host"`
	PORT     string `mapstructure:"port"`
	TYPE     string `mapstructure:"type"`
	NAME     string `mapstructure:"name"`
	USER     string `mapstructure:"user"`
	PASSWORD string `mapstructure:"password"`
}

// OAuth oauth config
type OAuth struct {
	AdminHost    string `mapstructure:"admin_host"`
	AdminPort    int    `mapstructure:"admin_port"`
	AdminPrefix  string `mapstructure:"admin_prefix"`
	PublicHost   string `mapstructure:"public_host"`
	PublicPort   int    `mapstructure:"public_port"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	ClientName   string `mapstructure:"client_name"`
	RedirectURI  string `mapstructure:"redirect_uri"`
}

// Authentication authentication config
type Authentication struct {
	PublicHost  string `mapstructure:"public_host"`
	PublicPort  int    `mapstructure:"public_port"`
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort int    `mapstructure:"private_port"`
}

// DeployService  deploy service config
type DeployService struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

// UserManagement  user-management service config
type UserManagement struct {
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

// Efast efast service
type Efast struct {
	PublicHost  string `mapstructure:"public_host"`
	PublicPort  int    `mapstructure:"public_port"`
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort int    `mapstructure:"private_port"`
}

// DocShare doc-share service
type DocShare struct {
	PublicHost  string `mapstructure:"public_host"`
	PublicPort  int    `mapstructure:"public_port"`
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort int    `mapstructure:"private_port"`
}

type DocSet struct {
	PublicHost  string `mapstructure:"public_host"`
	PublicPort  int    `mapstructure:"public_port"`
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort int    `mapstructure:"private_port"`
}

type PersonalConfig struct {
	PublicHost  string `mapstructure:"public_host"`
	PublicPort  int    `mapstructure:"public_port"`
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort int    `mapstructure:"private_port"`
}

// Emetadata 元数据配置
type Emetadata struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// Metadata 元数据配置
type Metadata struct {
	PublicHost  string `mapstructure:"public_host"`
	PublicPort  int    `mapstructure:"public_port"`
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort int    `mapstructure:"private_port"`
}

// EacpLog 配置
// type EacpLog struct {
// 	Host string `mapstructure:"host"`
// 	Port int    `mapstructure:"port"`
// }

// Tika 配置
type Tika struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DocConvert struct {
	Host          string `mapstructure:"host"`
	GotenbergPort int    `mapstructure:"gotenberg_port"`
	TikaPort      int    `mapstructure:"tika_port"`
}

// FastTextAnalysis 配置
type FastTextAnalysis struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// Document document service配置
type Document struct {
	PublicHost  string `mapstructure:"public_host"`
	PublicPort  int    `mapstructure:"public_port"`
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort int    `mapstructure:"private_port"`
}

// T4th 第四范式配置
type T4th struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Protocol string `mapstructure:"protocol"`
	Enable   bool   `mapstructure:"enable"`
	Type     string `mapstructure:"connectortype"`
}

// CodeRunner 配置
type CodeRunner struct {
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

// Appstore 配置
type Appstore struct {
	PublicHost string `mapstructure:"public_host"`
	PublicPort string `mapstructure:"public_port"`
}

// Intelliinfo 配置
type Intelliinfo struct {
	PublicHost  string `mapstructure:"public_host"`
	PublicPort  string `mapstructure:"public_port"`
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

// ECron 配置
type ECron struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Protocol string `mapstructure:"protocol"`
}

// EcoTag ecotag服务配置
type EcoTag struct {
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

// ContentAutomation 服务配置
type ContentAutomation struct {
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
	PublicHost  string `mapstructure:"public_host"`
	PublicPort  string `mapstructure:"public_port"`
}

// MQ 消息队列配置
type MQ struct {
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	ConnectorType string `mapstructure:"connectortype"`
	LookupdHost   string `mapstructure:"lookupd_host"`
	LookupdPort   int    `mapstructure:"lookupd_port"`
	Auth          Auth   `mapstructure:"auth"`
	Protocol      string `mapstructure:"protocol"`
}

// Kafka 配置
type Kafka struct {
	SessionTimeoutMs              int    `mapstructure:"session_timeout_ms"`
	SocketTimeoutMs               int    `mapstructure:"socket_timeout_ms"`
	MaxPollIntervalMs             int    `mapstructure:"max_poll_interval_ms"`
	HeartbeatIntervalMs           int    `mapstructure:"heartbeat_interval_ms"`
	TransactionTimeoutMs          int    `mapstructure:"transaction_timeout_ms"`
	RetentionTime                 string `mapstructure:"retention_time"`
	RetentionSize                 int    `mapstructure:"retention_size"`
	AutoOffsetReset               string `mapstructure:"auto_offset_reset"`
	AdminClientRequestTimeoutMs   int    `mapstructure:"admin_client_request_timeout_ms"`
	AdminClientOperationTimeoutMs int    `mapstructure:"admin_client_operation_timeout_ms"`
	Retries                       int    `mapstructure:"retries"`
	RetryBackoffMs                int    `mapstructure:"retry_backoff_ms"`
}

// Auth kafka auth参数
type Auth struct {
	UserName  string `mapstructure:"username"`
	PassWord  string `mapstructure:"password"`
	Mechanism string `mapstructure:"mechanism"`
}

// MongoDBConfig mongodb 配置信息
type MongoDBConfig struct {
	User       string `mapstructure:"user"`
	Password   string `mapstructure:"password"`
	Hosts      string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Name       string `mapstructure:"name"`
	DSNDB      string `mapstructure:"dsndb"`
	ReplicaSet string `mapstructure:"replicaset"`
	Direct     bool   `mapstructure:"direct"`
	PoolMax    uint64 `mapstructure:"pool_max"`
	PoolMin    uint64 `mapstructure:"pool_min"`
	SSL        bool   `mapstructure:"ssl"`
}

// OSSGateWay oss 配置信息
type OSSGateWay struct {
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

// OssGatewayBackend 新版 OssGateway Backend 服务配置
type OssGatewayBackend struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DumpLog 日志转储阈值设置
type DumpLog struct {
	DagThreshold      int64  `mapstructure:"dag_threshold"`
	TaskThreshold     int64  `mapstructure:"task_threshold"`
	KeepHistory       string `mapstructure:"keep_history"`
	CronJobExpression string `mapstructure:"cronjob_expression"`
}

// RedisConfiguration redis config
type RedisConfiguration struct {
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	SlaveHost        string `mapstructure:"slave_host"`
	SlavePort        int    `mapstructure:"slave_port"`
	UserName         string `mapstructure:"username"`
	Password         string `mapstructure:"password"`
	SentinelUsername string `mapstructure:"sentinel_username"`
	SentinelPassword string `mapstructure:"sentinel_password"`
	MasterGroupName  string `mapstructure:"master_groupname"`
	// 是否是云模式
	ClusterMode string `mapstructure:"cluster_mode"`
}

// Kcmc kc服务配置
type Kcmc struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

// ShareMgnt 邮件服务配置
type ShareMgnt struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// SpeechModel 音频转文字服务配置
type SpeechModel struct {
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

// Uie 自定义文本提取服务配置
type Uie struct {
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

type CognitiveAssistant struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// AnyData 配置
type AnyData struct {
	Protocol string `mapstructure:"protocol"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	AppID    string `mapstructure:"appid"`
	Model    string `mapstructure:"model"`
}

// Ecoconfig
type Ecoconfig struct {
	PublicHost  string `mapstructure:"public_host"`
	PublicPort  string `mapstructure:"public_port"`
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

// DataFlowTools 数据流工具配置
type DataFlowTools struct {
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

// AgentOperatorIntegration
type AgentOperatorIntegration struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type MfModelManager struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type MfModelApi struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type AgentFactory struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type KnKnowledgeData struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type AgentApp struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type MdlUniquery struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type MdlDataModel struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type AuthorizationConf struct {
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

type MdlDataPipeline struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type OpenSearch struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Protocol string `mapstructure:"protocol"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// DataSource 数据源服务配置
type DPDataSource struct {
	Protocol string `mapstructure:"protocol"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
}

type ContentPipeline struct {
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

type OCR struct {
	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
}

type StructureExtractor struct {
	UseMineru     bool   `mapstructure:"use_mineru"`
	MineruBaseURL string `mapstructure:"mineru_base_url"`
	MineruToken   string `mapstructure:"mineru_token"`

	PrivateHost string `mapstructure:"private_host"`
	PrivatePort string `mapstructure:"private_port"`
	OutputDir   string `mapstructure:"output_dir"`
	Backend     string `mapstructure:"backend"`
	ServerUrl   string `mapstructure:"server_url"`
	FileHost    string `mapstructure:"file_host"`
	FilePort    string `mapstructure:"file_port"`
}

type BusinessDomain struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type AccessAddress struct {
	Host   string `mapstructure:"host"`
	Port   string `mapstructure:"port"`
	Schema string `mapstructure:"schema"`
	Path   string `mapstructure:"path"`
}

// DSN 获取连接地址
func (conf *MongoDBConfig) DSN() string {
	query := url.Values{}
	if conf.Direct {
		query.Set("directConnection", "true")
	}
	if conf.ReplicaSet != "" {
		query.Set("replicaSet", conf.ReplicaSet)
	}

	if conf.SSL {
		for _, k := range []string{"tls", "tlsInsecure", "tlsAllowInvalidHostnames", "tlsAllowInvalidCertificates"} {
			query.Set(k, "true")
		}
		sslPath := "/opt/ssl/mongo.ca.pem"
		if _, err := os.Stat(sslPath); os.IsNotExist(err) {
			pwd, _ := os.Getwd()
			sslPath = filepath.Join(pwd, "ssl", "mongo.ca.pem")
		}
		rootPEM, err := os.ReadFile(sslPath)
		if err != nil {
			panic(err)
		}
		if len(rootPEM) > 0 {
			query.Set("tlsCertificateKeyFile", sslPath)
		}
	}

	addrs := []string{}
	arr := strings.Split(conf.Hosts, ",")
	for _, add := range arr {
		addrs = append(addrs, fmt.Sprintf("%s:%d", strings.TrimSpace(add), conf.Port))
	}

	if conf.DSNDB == "" {
		conf.DSNDB = conf.Name
	}
	dsn := url.URL{
		Scheme:   "mongodb",
		User:     url.UserPassword(conf.User, conf.Password),
		Host:     strings.Join(addrs, ","),
		Path:     url.PathEscape(conf.DSNDB),
		RawQuery: query.Encode(),
	}
	dsnStr := strings.Replace(dsn.String(), "+", "%2B", -1)
	return dsnStr
}

// DBName 获取DB名称
func (conf *MongoDBConfig) DBName() string {
	return conf.Name
}

// MaxPool 连接池最大值
func (conf *MongoDBConfig) MaxPool() *uint64 {
	return &conf.PoolMax
}

// MinPool 连接池最小值
func (conf *MongoDBConfig) MinPool() *uint64 {
	return &conf.PoolMin
}

// Sandbox 沙箱服务配置
type Sandbox struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	TemplateID string `mapstructure:"template_id"` // 默认模板ID
	Cpu        string `mapstructure:"cpu"`
	Memory     string `mapstructure:"memory"`
	Disk       string `mapstructure:"disk"`
	Timeout    int    `mapstructure:"timeout"` // 默认超时时间（秒）
}

// FlowFileDownload 文件下载线程池配置
type FlowFileDownload struct {
	WorkerCount     int   `mapstructure:"worker_count"`     // 每实例 worker 数
	PollInterval    int   `mapstructure:"poll_interval"`    // 轮询间隔(秒)
	BatchSize       int   `mapstructure:"batch_size"`       // 每次查询任务数
	DownloadTimeout int   `mapstructure:"download_timeout"` // 下载超时(秒)
	MaxFileSize     int64 `mapstructure:"max_file_size"`    // 最大文件大小(字节)
}

// VegaBackendConfig vega-backend 服务配置
type VegaBackendConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// BindEnvs bind envs
func BindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		switch v.Kind() { //nolint
		case reflect.Struct:
			BindEnvs(v.Interface(), append(parts, tv)...)
		default:
			viper.BindEnv(strings.Join(append(parts, tv), ".")) //nolint
		}
	}
}

// InitConfig 初始化配置
func InitConfig() (*Config, error) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.ReadInConfig() //nolint
	BindEnvs(config)
	err := viper.Unmarshal(&config)

	if err != nil {
		fmt.Printf("Unable to load config: %s\n", err)
	}

	if config.ActionConfigYaml != "" {
		f, err := os.Open(config.ActionConfigYaml)

		if err != nil {
			fmt.Printf("Unable to open action_config.yaml: %s", err.Error())
		}

		defer f.Close()

		if err := yaml.NewDecoder(f).Decode(&config.ActionConfig); err != nil {
			fmt.Printf("Unable to decode action_config.yaml: %s", err.Error())
		}
	}

	return &config, err
}

// NewConfig new config instance
func NewConfig() *Config {
	return &config
}

func (c *Config) IsAuthEnabled() bool {
	if c == nil {
		return true
	}

	switch strings.ToLower(strings.TrimSpace(c.Server.AuthEnabled)) {
	case "", "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

func (c *Config) IsBusinessDomainEnabled() bool {
	if c == nil {
		return true
	}

	switch strings.ToLower(strings.TrimSpace(c.Server.BusinessDomainEnabled)) {
	case "", "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}
