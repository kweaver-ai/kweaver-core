package common

import (
	"fmt"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// CacheConfig 缓存配置接口
type CacheConfig interface {
	GetHydraConfig() HydraConfig
	SetHydraConfig(config HydraConfig)
}

// ConfigLoader 配置文件
type ConfigLoader struct {
	Lang                  string `yaml:"lang"`
	JobFailures           int    `yaml:"job_failures"`
	AnalysisServiceID     string `yaml:"analysis_service_id"`
	ManagementServiceID   string `yaml:"management_service_id"`
	Webhook               string `yaml:"webhook"`
	MQConnectorType       string `yaml:"mq_connector_type"`
	CronAddr              string `yaml:"cron_addr"`
	CronPort              int    `yaml:"cron_port"`
	AnlysPort             int    `yaml:"anlys_port"`
	CronProtocol          string `yaml:"cron_protocol"`
	MultiNode             bool   `yaml:"multi_node"`
	DBAddr                string `yaml:"db_addr"`
	DBPort                int    `yaml:"db_port"`
	SystemId              string `yaml:"system_id"`
	DBName                string `yaml:"db_name"`
	UserName              string `yaml:"user_name"`
	UserPwd               string `yaml:"user_pwd"`
	DBFormat              string `yaml:"db_format"`
	TimeOut               int    `yaml:"timeout"`
	ReadTimeOut           int    `yaml:"read_timeout"`
	WriteTimeOut          int    `yaml:"write_timeout"`
	SSLOn                 bool   `yaml:"ssl_on"`
	SSLCertFile           string `yaml:"cert_file"`
	SSLKeyFile            string `yaml:"key_file"`
	AnalysisLoadSleep     int    `yaml:"analysis_load_sleep"`
	AnalysisRefreshSleep  int    `yaml:"analysis_refresh_sleep"`
	JobStatusRefreshSleep int    `yaml:"job_status_refresh_sleep"`
	MsmqClientOpSleep     int    `yaml:"msmq_client_op_sleep"`
	DBClientPingSleep     int    `yaml:"db_client_ping_sleep"`
	LostImmediateJobSleep int    `yaml:"lost_immediate_job_sleep"`
	TokenRefreshSleep     int    `yaml:"token_refresh_sleep"`
	MaxOpenConns          int    `yaml:"max_open_conns"`
	OAuthPublicAddr       string `yaml:"oauth_public_addr"`
	OAuthPublicPort       int    `yaml:"oauth_public_port"`
	OAuthPublicProtocol   string `yaml:"oauth_public_protocol"`
	OAuthAdminAddr        string `yaml:"oauth_admin_addr"`
	OAuthAdminPort        int    `yaml:"oauth_admin_port"`
	OAuthAdminProtocol    string `yaml:"oauth_admin_protocol"`
	AuthEnabled           string `yaml:"auth_enabled"`
	MQConfigPath          string
}

// Visitor 访问者信息
type Visitor struct {
	ClientID string `yaml:"client_id"`
	Name     string `yaml:"name"`
	Admin    bool   `yaml:"admin"`
}

// ServerInfo 服务器信息
type ServerInfo struct {
	ServiceID string //系统内服务唯一标识
	Addr      string //服务监听地址
	Port      int    //服务监听端口
	SSLOn     bool   //SSL启用标志，true=启用，false=禁用
	CertFile  string //SSL启用后，cert证书路径
	KeyFile   string //SSL启用后，key证书路径
	MultiNode bool   //端口不固定开关，true=选一个可用端口，false=选择固定端口
}

// LangLoader 语言加载器
type LangLoader struct {
	Lang map[string]string
}

// GetString 获取字符串型语言项
func (l LangLoader) GetString(key string) (value string) {
	if v, ok := l.Lang[key]; ok {
		value = v
	}
	return
}

// JobInfo 任务信息
type JobInfo struct {
	JobID       string     `json:"job_id"`
	JobName     string     `json:"job_name"`
	JobCronTime string     `json:"job_cron_time"`
	JobType     string     `json:"job_type"`
	Context     JobContext `json:"job_context"`
	TenantID    string     `json:"tenant_id"`
	Enabled     bool       `json:"enabled"`
	Remarks     string     `json:"remarks"`
	CreateTime  string     `json:"created_at"`
	UpdateTime  string     `json:"updated_at"`
}

// JobContext 任务上下文，任务执行的具体信息结构
type JobContext struct {
	Mode      string         `json:"mode"`
	Exec      string         `json:"exec"`
	Info      JobContextInfo `json:"info"`
	Notify    JobNotify      `json:"notify"`
	BeginTime string         `json:"begin_at"`
	EndTime   string         `json:"end_at"`
	ExecuteID string         `json:"execute_id"`
}

// JobContextInfo 任务执行信息结构
type JobContextInfo struct {
	Method     string                 `json:"method"`
	Params     map[string]string      `json:"params"`
	Headers    map[string]string      `json:"headers"`
	Body       map[string]interface{} `json:"body"`
	Kubernetes map[string]interface{} `json:"kubernetes"`
}

// JobNotify 任务通知
type JobNotify struct {
	Webhook string `json:"webhook"`
}

// JobStatus 任务状态
type JobStatus struct {
	ExecuteID    string                   `json:"execute_id"`
	JobID        string                   `json:"job_id"`
	JobType      string                   `json:"job_type"`
	JobName      string                   `json:"job_name"`
	JobStatus    string                   `json:"job_status"`
	BeginTime    string                   `json:"begin_at"`
	EndTime      string                   `json:"end_at"`
	Executor     []map[string]interface{} `json:"executor"` //任务执行者：{"executor_id":"123","execute_time":"..."}
	ExecuteTimes int                      `json:"execute_times"`
	ExtInfo      map[string]interface{}   `json:"ext_info"`
}

// JobTotal 定义用户分页获取时，先获取总数的数据结构
type JobTotal struct {
	Total     int    `json:"total"`
	TimeStamp string `json:"timestamp"`
}

// JobMsg 发布的定时任务通知
type JobMsg struct {
	Method string  `json:"method"`
	Data   JobInfo `json:"data"`
}

// JobInfoQueryParams 定时任务查询参数
type JobInfoQueryParams struct {
	JobID     []string `json:"job_id"`
	JobType   string   `json:"job_type"`
	Limit     int      `json:"limit"`
	Page      int      `json:"page"`
	TimeStamp string   `json:"timestamp"`
}

// JobStatusQueryParams 任务状态查询参数
type JobStatusQueryParams struct {
	JobID     string `json:"job_id"`
	JobType   string `json:"job_type"`
	JobStatus string `json:"job_status"`
	BeginTime string `json:"begin_at"`
	EndTime   string `json:"end_at"`
}

// JobTotalQueryParams 任务总数查询参数
type JobTotalQueryParams struct {
	BeginTime string `json:"begin_at"`
	EndTime   string `json:"end_at"`
}

// DJobType 任务类型字典
type DJobType struct {
	jobTypeInt    map[int]string
	jobTypeString map[string]int
}

// StringToInt 字符串转整型
func (d *DJobType) StringToInt(key string) (value int, ok bool) {
	if v, f := d.jobTypeString[key]; f {
		value = v
		ok = f
	}
	return
}

// IntToString 整形转字符串
func (d *DJobType) IntToString(key int) (value string, ok bool) {
	if v, f := d.jobTypeInt[key]; f {
		value = v
		ok = f
	}
	return
}

// DJobExecution 任务执行类型字典
type DJobExecution struct {
	jobExecutionInt    map[int]string
	jobExecutionString map[string]int
}

// StringToInt 字符串转整型
func (d *DJobExecution) StringToInt(key string) (value int, ok bool) {
	if v, f := d.jobExecutionString[key]; f {
		value = v
		ok = f
	}
	return
}

// IntToString 整形转字符串
func (d *DJobExecution) IntToString(key int) (value string, ok bool) {
	if v, f := d.jobExecutionInt[key]; f {
		value = v
		ok = f
	}
	return
}

// DJobOperation 任务操作类型字典
type DJobOperation struct {
	jobOperationInt    map[int]string
	jobOperationString map[string]int
}

// StringToInt 字符串转整型
func (d *DJobOperation) StringToInt(key string) (value int, ok bool) {
	if v, f := d.jobOperationString[key]; f {
		value = v
		ok = f
	}
	return
}

// IntToString 整形转字符串
func (d *DJobOperation) IntToString(key int) (value string, ok bool) {
	if v, f := d.jobOperationInt[key]; f {
		value = v
		ok = f
	}
	return
}

// DJobStatus 任务状态类型字典
type DJobStatus struct {
	jobStatusInt    map[int]string
	jobStatusString map[string]int
}

// StringToInt 字符串转整型
func (d *DJobStatus) StringToInt(key string) (value int, ok bool) {
	if v, f := d.jobStatusString[key]; f {
		value = v
		ok = f
	}
	return
}

// IntToString 整形转字符串
func (d *DJobStatus) IntToString(key int) (value string, ok bool) {
	if v, f := d.jobStatusInt[key]; f {
		value = v
		ok = f
	}
	return
}

// DataDict 数据字典
type DataDict struct {
	DJobType
	DJobExecution
	DJobOperation
	DJobStatus
}

var (
	dd      *DataDict
	ddMutex sync.Mutex
)

// NewDataDict 初始化字典
func NewDataDict() *DataDict {
	ddMutex.Lock()
	defer ddMutex.Unlock()

	if nil != dd {
		return dd
	}

	dd = &DataDict{
		DJobType: DJobType{
			jobTypeInt: map[int]string{
				1: TIMING,
				2: PERIODICITY,
				3: IMMEDIATE,
			},
			jobTypeString: map[string]int{
				TIMING:      1,
				PERIODICITY: 2,
				IMMEDIATE:   3,
			},
		},
		DJobExecution: DJobExecution{
			jobExecutionInt: map[int]string{
				1: HTTP,
				2: EXE,
				3: HTTPS,
			},
			jobExecutionString: map[string]int{
				HTTP:  1,
				EXE:   2,
				HTTPS: 3,
			},
		},
		DJobOperation: DJobOperation{
			jobOperationInt: map[int]string{
				1: CREATE,
				2: UPDATE,
				3: DELETE,
				4: ENABLE,
				5: NOTIFY,
			},
			jobOperationString: map[string]int{
				CREATE: 1,
				UPDATE: 2,
				DELETE: 3,
				ENABLE: 4,
				NOTIFY: 5,
			},
		},
		DJobStatus: DJobStatus{
			jobStatusInt: map[int]string{
				1: SUCCESS,
				2: EXECUTING,
				3: FAILURE,
				4: INTERRUPT,
				5: ABANDON,
			},
			jobStatusString: map[string]int{
				SUCCESS:   1,
				EXECUTING: 2,
				FAILURE:   3,
				INTERRUPT: 4,
				ABANDON:   5,
			},
		},
	}
	return dd
}

// 字典字符串
var (
	TIMING      = "timed"
	PERIODICITY = "scheduled"
	IMMEDIATE   = "real-time"

	HTTP  = "http"
	EXE   = "exe"
	HTTPS = "https"

	CREATE = "create"
	UPDATE = "update"
	DELETE = "delete"
	ENABLE = "enable"
	NOTIFY = "notify"

	SUCCESS   = "success"
	EXECUTING = "executing"
	FAILURE   = "failure"
	INTERRUPT = "interrupt"
	ABANDON   = "abandon"
)

// 消息队列topic名称
var (
	TopicImmediateJob = "cron.topic.immediate.job"
	TopicCronJob      = "cron.topic.cron.job"
	TopicJobStatus    = "cron.topic.job.status"
)

// 消息队列channel名称
var (
	ChannelImmediateJob = "cron.channel.immediate.job"
	ChannelCronJob      = "cron.channel.cron.job"
	ChannelJobStatus    = "cron.channel.job.status"
	ChannelECron        = "ECronChan"
)

// 定时服务错误码
const (
	BadRequest      = 400009001
	Unauthorized    = 401009001
	NotFound        = 404009001
	Conflict        = 409009001
	TooManyRequests = 429009001
	InternalError   = 500009001
)

// 公共key值
var (
	DetailConflicts           = "conflicts"
	DetailParameters          = "parameters"
	Authorization             = "Authorization"
	AdminSecret               = "Secret"
	AdminCode                 = "Code"
	Bearer                    = "Bearer"
	Basic                     = "Basic"
	ContentType               = "Content-Type"
	ApplicationFormUrlencoded = "application/x-www-form-urlencoded"
	ApplicationJSON           = "application/json"
	IsDeleted                 = "is_deleted"
	TenantID                  = "tenant_id"
)

// 通用错误字符串
var (
	ErrDataBaseUnavailable         = "Database is unavailable"              //数据库不可用
	ErrHTTPClientUnavailable       = "HTTP client is unavailable"           //HTTP客户端不可用
	ErrMSMQClientUnavailable       = "Msmq client is unavailable"           //消息队列客户端不可用
	ErrCronClientUnavailable       = "Cron client is unavailable"           //定时客户端不可用
	ErrExecutorUnavailable         = "Executor is unavailable"              //执行者不可用
	ErrAuthClientUnavailable       = "Auth client is unavailable"           //授权客户端不可用
	ErrGinContextUnavailable       = "Gin context is unavailable"           //gin服务上下文不可用
	ErrACLClientUnavailable        = "Acl client is unavailable"            //acl客户端不可用
	ErrUnsupportedExecutionMode    = "Unsupported execution mode"           //执行类型不支持
	ErrExecuteIDAndJobIDConfused   = "Executed ID and job ID are in chaos"  //执行ID和任务ID混乱
	ErrJobExecutedTooManyTimes     = "Too many job executions"              //任务执行次数太多
	ErrDataBaseDisconnected        = "Database is disconnected"             //数据库连接断开
	ErrJobNameExists               = "Job name already exists"              //任务名称已存在
	ErrJobNameEmpty                = "Job name is empty"                    //任务名称为空
	ErrDataBaseExecError           = "Database execution error"             //数据库执行错误
	ErrInsertJob                   = "Add job error"                        //添加任务错误
	ErrUpdateJob                   = "Job update error"                     //更新任务错误
	ErrTransactionBegin            = "Transaction begin error"              //事务开始错误
	ErrDeleteJob                   = "Delete job error"                     //删除任务错误
	ErrUpdateJobDeletedFlag        = "Error in updating job to delete flag" //更新任务删除标志错误
	ErrCommit                      = "Submission error"                     //提交错误
	ErrUpdateJobStatus             = "Error in job status updating"         //更新任务状态错误
	ErrBatchJobEnable              = "Error in batch job enabling"          //批量处理任务使能错误
	ErrBatchJobNotify              = "Error in batch job notification"      //批量处理任务通知错误
	ErrPrepare                     = "Error in database preparation"        //数据库准备错误
	ErrOpenDataBase                = "Database open error"                  //打开数据库错误
	ErrPingDataBase                = "Ping database error"                  //ping数据库错误
	ErrQueryJob                    = "Job query error"                      //查询任务错误
	ErrScanFieldValue              = "Error in field value scanning"        //扫描字段值错误
	ErrQueryJobTotal               = "Error in job count query"             //查询任务总数错误
	ErrQueryJobStatus              = "Error in job status query"            //查询任务状态错误
	ErrQueryParameterIsNull        = "Query parameters is null"             //查询参数为空
	ErrInvalidParameter            = "Invalid parameter"                    //无效参数
	ErrJobNotExist                 = "This job does not exist"              //任务不存在
	ErrBeginTimeGreaterThanEndTime = "End time is earlier than start time"  //开始时间大于结束时间
	ErrTimeIllegal                 = "Illegal time "                        //时间不合法
	ErrCronTime                    = "Cron time error"                      //定时时间错误
	ErrQueryJobName                = "Error in job name query"              //查询任务名称错误
	ErrQueryJobID                  = "Error in job ID query"                //查询任务ID错误
	ErrMarshalJSON                 = "Failed to serialize json"             //序列化json失败
	ErrUnMarshalJSON               = "Failed to deserialize json"           //反序列化json失败
	ErrJobTypeIllegal              = "Illegal job type"                     //任务类型非法
	ErrJobStatusIllegal            = "Illegal job status"                   //任务状态非法
	ErrLimitOrPageIllegal          = "Limit or page illegal"                //页码或限制页数非法
	ErrJobDeleted                  = "This job is deleted"                  //任务已删除
	ErrStatusEmpty                 = "Empty status"                         //状态为空
	ErrInvalidToken                = "Invalid token"                        //无效的token
	ErrTokenExpired                = "Token expired"                        //token已过期
	ErrInvalidExpiresTime          = "Invalid expiration time"              //截止时间无效
	ErrTokenEmpty                  = "Empty token"                          //token为空
)

// 通用提示字符串
var (
	InfoDataBaseConnected = "Database is connected"
)

// 内部版本号
var (
	Version = "1.0.202003241853"
)

// ECronError 定时服务错误结构
type ECronError struct {
	Cause   string                 `json:"cause"`
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Detail  map[string]interface{} `json:"detail"`
}

func (err ECronError) Error() string {
	errstr, _ := jsoniter.Marshal(err)
	return string(errstr)
}

// FormatAddress 格式化IP和端口
func FormatAddress(addr string, port int) (address string) {
	addr = ParseHost(addr)
	address = fmt.Sprintf("%s:%d", addr, port)
	if 0 == port {
		address = addr
	}
	return
}

// GetHTTPAccess 获取RESTful API访问方式
func GetHTTPAccess(addr string, port int, sslOn bool) (access string) {
	if sslOn {
		return fmt.Sprintf("https://%s", FormatAddress(addr, port))
	}
	return fmt.Sprintf("http://%s", FormatAddress(addr, port))
}

// TimeStampToString 时间戳转RFC3339格式的字符串
func TimeStampToString(t int64) string {
	if t <= 0 {
		return ""
	}

	return time.Unix(t, 0).Format(time.RFC3339)
}

// StringToTimeStamp RFC3339格式的字符串转时间戳
func StringToTimeStamp(t string) (int64, error) {
	if 0 == len(t) {
		return 0, nil
	}

	ts, err := time.Parse(time.RFC3339, t)
	if nil != err {
		return 0, err
	}
	return ts.Unix(), nil
}

// GetSleepDuration 校验线程等待时间
func GetSleepDuration(data int) time.Duration {
	return time.Duration(1e9 * GetIntMoreThanLowerLimit(data, 10))
}

// GetIntMoreThanLowerLimit 返回一个不小于下限的整数
func GetIntMoreThanLowerLimit(data int, lower int) int {
	if data < lower {
		return lower
	}
	return data
}

// HydraConfig hydra服务配置信息
type HydraConfig struct {
	VerifyTokenPath string // token内省path
}

type cacheConfig struct {
	sync.RWMutex
	hydraConfig HydraConfig
}

var (
	cacheConfigOnce    sync.Once
	cacheConfigHandler CacheConfig
)

// NewCacheConfig 创建缓存配置对象
func NewCacheConfig() CacheConfig {
	cacheConfigOnce.Do(func() {
		cacheConfigHandler = &cacheConfig{
			hydraConfig: HydraConfig{},
		}
	})
	return cacheConfigHandler
}

func (c *cacheConfig) GetHydraConfig() HydraConfig {
	c.RLock()
	defer c.RUnlock()
	return c.hydraConfig
}

func (c *cacheConfig) SetHydraConfig(config HydraConfig) {
	c.Lock()
	defer c.Unlock()
	c.hydraConfig.VerifyTokenPath = config.VerifyTokenPath
}
