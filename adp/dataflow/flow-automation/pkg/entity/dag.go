package entity

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	libstore "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/store"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/utils"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/utils/value"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm/opcode"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/store"
	normalizeutil "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils/normalize"
	"go.mongodb.org/mongo-driver/bson"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewDag new a dag
func NewDag() *Dag {
	return &Dag{
		Status: DagStatusNormal,
	}
}

// Dag dag struct
type Dag struct {
	BaseInfo    `yaml:",inline" json:",inline" bson:"inline"`
	UserID      string     `yaml:"userid,omitempty" json:"userid,omitempty" bson:"userid,omitempty"`
	Name        string     `yaml:"name,omitempty" json:"name,omitempty" bson:"name,omitempty"`
	Desc        string     `yaml:"desc,omitempty" json:"desc,omitempty" bson:"desc,omitempty"`
	Trigger     Trigger    `yaml:"trigger,omitempty" json:"trigger,omitempty" bson:"trigger,omitempty"`
	Cron        string     `yaml:"cron,omitempty" json:"cron,omitempty" bson:"cron,omitempty"`
	Vars        DagVars    `yaml:"vars,omitempty" json:"vars,omitempty" bson:"vars,omitempty"`
	Status      DagStatus  `yaml:"status,omitempty" json:"status,omitempty" bson:"status,omitempty"`
	Tasks       []Task     `yaml:"tasks,omitempty" json:"tasks,omitempty" bson:"tasks,omitempty"`
	Steps       []Step     `yaml:"steps,omitempty" json:"steps,omitempty" bson:"steps,omitempty"`
	Description string     `yaml:"description,omitempty" json:"description,omitempty" bson:"description,omitempty"`
	Shortcuts   []string   `yaml:"shortcuts,omitempty" json:"shortcuts,omitempty" bson:"shortcuts,omitempty"`
	Accessors   []Accessor `yaml:"accessors,omitempty" json:"accessors,omitempty" bson:"accessors,omitempty"`
	Type        string     `yaml:"type,omitempty" json:"type,omitempty" bson:"type,omitempty"`
	PolicyType  string     `yaml:"policy_type,omitempty" json:"policy_type,omitempty" bson:"policy_type,omitempty"`
	AppInfo     AppInfo    `yaml:"appinfo,omitempty" json:"appinfo,omitempty" bson:"appinfo,omitempty"`
	Priority    string     `yaml:"priority,omitempty" json:"priority,omitempty" bson:"priority,omitempty"`
	Removed     bool       `yaml:"removed,omitempty" json:"removed,omitempty" bson:"removed,omitempty"`
	Emails      []string   `yaml:"emails,omitempty" json:"emails,omitempty" bson:"emails,omitempty"`
	Template    string     `yaml:"template,omitempty" json:"template,omitempty" bson:"template,omitempty"`
	Published   bool       `yaml:"publish,omitempty" json:"publish,omitempty" bson:"publish,omitempty"`

	pushMessage   func(topic string, message []byte) error `yaml:"-" json:"-" bson:"-"`
	checkRootNode func([]Task) error                       `yaml:"-" json:"-" bson:"-"`

	TriggerConfig *TriggerConfig `yaml:"trigger_config,omitempty" json:"trigger_config,omitempty" bson:"trigger_config,omitempty"`

	// SubIDs 当前流程绑定的子流程id，当前仅在组合算子注册时使用
	SubIDs []string `yaml:"sub_ids,omitempty" json:"sub_ids,omitempty" bson:"sub_ids,omitempty"`
	// ExecMode 组合算子执行模式
	ExecMode string `yaml:"exec_mode,omitempty" json:"exec_mode,omitempty" bson:"exec_mode,omitempty"`
	// Category 组合算子分类
	Category string `yaml:"category,omitempty" json:"category,omitempty" bson:"category,omitempty"`
	// 输出节点参数定义结构
	OutPuts []*OutPut `yaml:"outputs,omitempty" json:"outputs,omitempty" bson:"outputs,omitempty"`
	// 预编译指令
	Instructions []*opcode.Instruction `yaml:"instructions,omitempty" json:"instructions,omitempty" bson:"instructions,omitempty"`
	// 组合算子ID,用于权限校验时传递resourceID使用
	OperatorID string `yaml:"operator_id,omitempty" json:"operator_id,omitempty" bson:"operator_id,omitempty"`

	IncValues map[string]any `yaml:"-" json:"inc_values,omitempty" bson:"inc_values,omitempty"`
	// 流程语义化版本信息
	Version Version `yaml:"version" json:"version" bson:"version"`
	// 流程版本ID
	VersionID string `yaml:"versionId" json:"versionId" bson:"versionId"`
	// 流程更新人
	ModifyBy string `yaml:"modify_by" json:"modify_by" bson:"modify_by"`
	// 是否为调试流程
	IsDebug bool `yaml:"is_debug,omitempty" json:"is_debug,omitempty" bson:"is_debug,omitempty"`
	// 流程调试ID
	DeBugID string `yaml:"debug_id,omitempty" json:"debug_id,omitempty" bson:"debug_id,omitempty"`
	// 业务域ID
	BizDomainID string `yaml:"biz_domain_id,omitempty" json:"biz_domain_id,omitempty" bson:"biz_domain_id,omitempty"`
}

// OutPut 输出节点信息
type OutPut struct {
	Key         string         `json:"key,omitempty"`
	Name        string         `json:"name,omitempty"`
	Type        string         `json:"type,omitempty"`
	Description *OutPutVarDesc `json:"description,omitempty"`
}

type OutPutVarDesc struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

type TriggerConfig struct {
	Operator   string                 `json:"operator,omitempty"`
	Cron       string                 `json:"cron,omitempty"`
	DataSource *DataSource            `json:"dataSource,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// AppInfo 应用账户信息
type AppInfo struct {
	Enable bool `yaml:"enable,omitempty" json:"enable,omitempty" bson:"enable,omitempty"`
}

// SpecifiedVar special variable
type SpecifiedVar struct {
	Name  string
	Value string
}

// DagRunOption dag run option
type DagRunOption func(*DagInstance)

// WithKeyWords
func WithKeyWords(keywords ...string) DagRunOption {
	return func(dagIns *DagInstance) {
		dagIns.Keywords = keywords
	}
}

// WithCallChain
func WithCallChain(chains []string) DagRunOption {
	return func(dagIns *DagInstance) {
		dagIns.CallChain = chains
	}
}

// WithCallBack
func WithCallBack(successCallBack, failCallBack string) DagRunOption {
	return func(dagIns *DagInstance) {
		dagIns.SuccessCallback = successCallBack
		dagIns.ErrorCallback = failCallBack
	}
}

func (d *Dag) SetPushMessage(publish func(topic string, message []byte) error) {
	d.pushMessage = publish
}

func (d *Dag) SetCheckRootNode(fn func([]Task) error) {
	d.checkRootNode = fn
}

func (d *Dag) CheckRootNode(tasks []Task) error {
	if d.checkRootNode == nil {
		return nil
	}
	return d.checkRootNode(tasks)
}

// Run used to build a new DagInstance, then you also need save it to Store
func (d *Dag) Run(ctx context.Context, trigger Trigger, specVars map[string]string, opts ...DagRunOption) (*DagInstance, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if d.Status != DagStatusNormal {
		return nil, fmt.Errorf("you cannot run a stopeed dag")
	}

	specVarsCopy := maps.Clone(specVars)

	if v, ok := specVarsCopy["call_chain"]; ok {
		delete(specVarsCopy, "call_chain")
		opts = append(opts, WithCallChain(strings.Split(v, ",")))
	}

	if v, ok := specVarsCopy["call_back"]; ok {
		delete(specVarsCopy, "call_back")
		opts = append(opts, WithCallBack(v, ""))
	}

	dagInsVars := DagInstanceVars{}

	for key, value := range specVarsCopy {
		dagInsVars[key] = DagInstanceVar{
			Value: value,
		}
	}

	var priority = d.Priority
	if priority == "" {
		priority = common.PriorityLowest
	}

	userID := d.UserID

	if d.Type == common.DagTypeComboOperator {
		userID = dagInsVars["operator_id"].Value
	}

	go func() {
		if d.IsDebug {
			return
		}

		did, _ := strconv.Atoi(d.ID)
		object := map[string]interface{}{
			"type":     trigger,
			"id":       d.ID,
			"did":      did,
			"name":     d.Name,
			"priority": d.Priority,
		}

		var userInfo drivenadapters.UserInfo
		userInfo, err0 := drivenadapters.NewUserManagement().GetUserInfoByType(dagInsVars["operator_id"].Value, dagInsVars["operator_type"].Value)
		if err0 != nil {
			log.Warnf("[dag.Run] GetUserInfoByType failed: %s", err0.Error())
			userInfo = drivenadapters.UserInfo{
				UserID:   dagInsVars["operator_id"].Value,
				UserName: dagInsVars["operator_name"].Value,
				Type:     dagInsVars["operator_type"].Value,
			}
		}
		userInfo.VisitorType = common.AuthenticatedUserType

		detail, _ := common.GetLogBody(common.StartRunDag, []interface{}{d.Name}, []interface{}{})
		drivenadapters.NewLogger().Log(drivenadapters.LogTypeASOperationLog, &drivenadapters.BuildARLogParams{
			Operation:   common.ArLogStartDagIns,
			Description: detail,
			UserInfo:    &userInfo,
			Object:      object,
		}, &drivenadapters.JSONLogWriter{SendFunc: d.pushMessage})
	}()

	dagIns := &DagInstance{
		ctx:              ctx,
		DagID:            d.ID,
		Trigger:          trigger,
		Vars:             dagInsVars,
		EventPersistence: DagInstanceEventPersistenceSql,
		Status:           DagInstanceStatusInit,
		UserID:           userID,
		DagType:          d.Type,
		PolicyType:       d.PolicyType,
		AppInfo:          d.AppInfo,
		Priority:         priority,
		Version:          d.Version,
		VersionID:        d.VersionID,
		BizDomainID:      d.BizDomainID,
	}

	dagIns.ShareData = &ShareData{
		Dict:        map[string]any{},
		DagInstance: dagIns,
	}

	for _, opt := range opts {
		opt(dagIns)
	}

	return dagIns, nil
}

// DagVars dag variables type
type DagVars map[string]DagVar

// DagVar dag variable
type DagVar struct {
	Desc         string `yaml:"desc,omitempty" json:"desc,omitempty" bson:"desc,omitempty"`
	DefaultValue string `yaml:"defaultValue,omitempty" json:"defaultValue,omitempty" bson:"defaultValue,omitempty"`
}

// DagInstanceVar dag instance variable
type DagInstanceVar struct {
	Value string `json:"value,omitempty" bson:"value,omitempty"`
}

// DagStatus dag status
type DagStatus string

const (
	// DagStatusNormal normal status
	DagStatusNormal DagStatus = "normal"
	// DagStatusStopped stopped status
	DagStatusStopped DagStatus = "stopped"
)

type DagInstanceMode int

const (
	DagInstanceModeSchedule DagInstanceMode = 0
	DagInstanceModeVM       DagInstanceMode = 1
)

type DagInstanceEventPersistence int

const (
	DagInstanceEventPersistenceNone DagInstanceEventPersistence = 0
	DagInstanceEventPersistenceSql  DagInstanceEventPersistence = 1
	DagInstanceEventPersistenceOss  DagInstanceEventPersistence = 2
)

// DagInstance dag instance
type DagInstance struct {
	BaseInfo         `bson:"inline"`
	ctx              context.Context
	DagID            string                      `json:"dagId,omitempty" bson:"dagId,omitempty"`
	Trigger          Trigger                     `json:"trigger,omitempty" bson:"trigger,omitempty"`
	Worker           string                      `json:"worker,omitempty" bson:"worker,omitempty"`
	Source           string                      `json:"source,omitempty" bson:"source,omitempty"`
	Vars             DagInstanceVars             `json:"vars,omitempty" bson:"vars,omitempty"`
	Keywords         []string                    `json:"keywords,omitempty" bson:"keywords,omitempty"`
	EventPersistence DagInstanceEventPersistence `json:"eventPersistence,omitempty" bson:"eventPersistence,omitempty"`
	EventOssPath     string                      `json:"eventOssPath,omitempty" bson:"eventOssPath,omitempty"`
	ShareData        *ShareData                  `json:"shareData,omitempty" bson:"shareData,omitempty"`
	ShareDataExt     *DagInstanceExtData         `json:"-" bson:"shareDataExt,omitempty"`
	Status           DagInstanceStatus           `json:"status,omitempty" bson:"status,omitempty"`
	Reason           string                      `json:"reason,omitempty" bson:"reason,omitempty"`
	Cmd              *Command                    `json:"cmd,omitempty" bson:"cmd,omitempty"`
	UserID           string                      `json:"userid,omitempty" bson:"userid,omitempty"`
	EndedAt          int64                       `json:"endedAt,omitempty" bson:"endedAt,omitempty"`
	DagType          string                      `json:"dag_type,omitempty" bson:"dag_type,omitempty"`
	PolicyType       string                      `json:"policy_type,omitempty" bson:"policy_type,omitempty"`
	AppInfo          AppInfo                     `json:"appinfo,omitempty" bson:"appinfo,omitempty"`
	Priority         string                      `json:"priority,omitempty" bson:"priority,omitempty"`
	Mode             DagInstanceMode             `json:"mode,omitempty" bson:"mode,omitempty"`
	Dump             string                      `json:"dump,omitempty" bson:"dump,omitempty"`
	DumpExt          *DagInstanceExtData         `json:"-" bson:"dumpExt,omitempty"`
	SuccessCallback  string                      `json:"success_callback,omitempty" bson:"success_callback,omitempty"`
	ErrorCallback    string                      `json:"error_callback,omitempty" bson:"error_callback,omitempty"`
	CallChain        []string                    `json:"-" bson:"call_chain,omitempty"`
	ResumeData       string                      `json:"-" bson:"resume_data,omitempty"`
	ResumeStatus     TaskInstanceStatus          `json:"-" bson:"resume_status,omitempty"`
	Version          Version                     `json:"version,omitempty" bson:"version,omitempty"`
	VersionID        string                      `json:"versionId,omitempty" bson:"versionId,omitempty"`
	BizDomainID      string                      `json:"biz_domain_id,omitempty" bson:"biz_domain_id,omitempty"`
	MemoryShareData  *MemoryShareData            `json:"-" bson:"-"`

	// 添加互斥锁用于保护 SaveExtData 方法
	extDataMutex sync.Mutex `json:"-" bson:"-"`
}

var (
	// StoreMarshal store marshal
	StoreMarshal func(interface{}) ([]byte, error)
	// StoreUnmarshal store unmarshal
	StoreUnmarshal func([]byte, interface{}) error
)

// ShareData can read/write within all tasks and will persist it
// if you want a high performance just within same task, you can use
// ExecuteContext's Context
type ShareData struct {
	Dict        map[string]interface{} `json:"dict,omitempty" bson:"dict,omitempty"`
	Save        func(data *ShareData) error
	mutex       sync.RWMutex // 改用读写锁
	DagInstance *DagInstance
}

// MarshalBSON used by mongo
func (d *ShareData) MarshalBSON() ([]byte, error) {

	if d.DagInstance != nil && d.DagInstance.EventPersistence != DagInstanceEventPersistenceNone {
		return StoreMarshal(map[string]any{})
	}

	d.mutex.RLock() // 使用读锁
	defer d.mutex.RUnlock()

	// 创建 map 的副本
	dictCopy := make(map[string]interface{})

	if d.Dict == nil {
		return StoreMarshal(bson.M{})
	}

	for k, v := range d.Dict {
		dictCopy[k] = v
	}

	bytes, err := json.Marshal(dictCopy)

	if err != nil {
		return nil, err
	}

	return StoreMarshal(map[string]any{
		"__internal_type": "object",
		"data":            string(bytes),
	})

}

// UnmarshalBSON used by mongo
func (d *ShareData) UnmarshalBSON(data []byte) error {
	dict := make(map[string]any)

	if len(data) == 0 || d.DagInstance != nil && d.DagInstance.EventPersistence != DagInstanceEventPersistenceNone {
		d.Dict = dict
		return nil
	}

	err := StoreUnmarshal(data, &dict)

	if err != nil {
		return err
	}

	internalType, ok := dict["__internal_type"]

	if !ok {
		d.Dict = dict
		return nil
	}

	if internalType != "object" {
		return fmt.Errorf("invalid sharedata")
	}

	if d.Dict == nil {
		d.Dict = make(map[string]interface{})
	}

	innerData, ok := dict["data"]

	if !ok {
		return nil
	}

	innerJson, ok := innerData.(string)

	if !ok {
		return fmt.Errorf("invalid sharedata")
	}

	return json.Unmarshal([]byte(innerJson), &d.Dict)
}

// MarshalJSON used by json
func (d *ShareData) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Dict)
}

// UnmarshalJSON used by json
func (d *ShareData) UnmarshalJSON(data []byte) error {
	if d.Dict == nil {
		d.Dict = make(map[string]interface{})
	}
	return json.Unmarshal(data, &d.Dict)
}

// Get value from share data, it is thread-safe.
func (d *ShareData) Get(key string) (interface{}, bool) {
	_ = d.Load(context.Background(), []string{key})
	d.mutex.RLock() // 使用读锁
	defer d.mutex.RUnlock()

	v, ok := d.Dict[key]
	return v, ok
}

func (d *ShareData) Load(ctx context.Context, keys []string) error {
	if d.DagInstance == nil || d.DagInstance.EventPersistence != DagInstanceEventPersistenceSql {
		return nil
	}

	names := []string{}

	if len(keys) == 0 {
		names = keys
	} else {
		if d.Dict == nil {
			names = keys
		} else {
			for _, k := range keys {
				if _, ok := d.Dict[k]; ok {
					continue
				}
				names = append(names, k)
			}
		}

		if len(names) == 0 {
			return nil
		}
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.Dict == nil {
		d.Dict = make(map[string]interface{})
	}

	events, err := d.DagInstance.ListEvents(context.Background(), &rds.DagInstanceEventListOptions{
		DagInstanceID: d.DagInstance.ID,
		Types:         []rds.DagInstanceEventType{rds.DagInstanceEventTypeVariable},
		Names:         names,
		LatestOnly:    true,
	})

	if err != nil {
		return err
	}

	for _, event := range events {
		d.Dict[event.Name] = event.Data
	}

	return nil
}

// Set value to share data, it is thread-safe.
func (d *ShareData) Set(key string, val interface{}) {
	d.mutex.Lock() // 写操作使用写锁
	defer d.mutex.Unlock()

	dict := make(map[string]any)

	if strings.HasPrefix(key, "__") {
		dict[key] = val
	} else if strings.Contains(key, "_i") {
		// 使用正则表达式匹配 _i{number}_s{id} 格式
		re := regexp.MustCompile(`_i\d+_s([^_]+)$`)
		matches := re.FindStringSubmatch(key)
		if len(matches) > 1 {
			originKey := matches[1]
			dict[fmt.Sprintf("__%s", originKey)] = val
			dict[fmt.Sprintf("__%s", key)] = val
		} else {
			// 如果不是 _i{number}_s{id} 格式，则取最后一个部分
			keys := strings.Split(key, "_")
			originKey := keys[len(keys)-1]
			dict[fmt.Sprintf("__%s", originKey)] = val
			dict[fmt.Sprintf("__%s", key)] = val
		}
	} else if key != "" {
		dict[fmt.Sprintf("__%s", key)] = val
	}

	if d.Dict == nil {
		d.Dict = dict
	} else {
		for k, v := range dict {
			d.Dict[k] = v
		}
	}

	if len(dict) > 0 &&
		d.DagInstance != nil &&
		d.DagInstance.EventPersistence == DagInstanceEventPersistenceSql &&
		d.DagInstance.Mode != DagInstanceModeVM {
		_ = d.DagInstance.WriteEventByVariableMap(context.Background(), dict, time.Now().UnixMicro())
	}

	if d.Save != nil {
		// 在保存之前释放锁，避免在持有锁的情况下进行 I/O 操作
		d.mutex.Unlock()
		err := d.Save(d)
		d.mutex.Lock() // 重新获取锁

		if err != nil {
			delete(d.Dict, key)
			log.Error("save share data failed",
				"err", err,
				"key", key,
				"value", val)
		}
	}
}

// 添加新的批量操作方法
func (d *ShareData) BatchSet(updates map[string]interface{}) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.Dict == nil {
		d.Dict = make(map[string]interface{})
	}

	for key, val := range updates {
		if strings.HasPrefix(key, "__") {
			d.Dict[key] = val
		} else if strings.Contains(key, "_i") {
			// 使用正则表达式匹配 _i{number}_s{id} 格式
			re := regexp.MustCompile(`_i\d+_s([^_]+)$`)
			matches := re.FindStringSubmatch(key)
			if len(matches) > 1 {
				originKey := matches[1]
				d.Dict[fmt.Sprintf("__%s", originKey)] = val
			} else {
				// 如果不是 _i{number}_s{id} 格式，则取最后一个部分
				keys := strings.Split(key, "_")
				originKey := keys[len(keys)-1]
				d.Dict[fmt.Sprintf("__%s", originKey)] = val
			}
		} else if key != "" {
			d.Dict[fmt.Sprintf("__%s", key)] = val
		}
	}

	if len(updates) > 0 &&
		d.DagInstance != nil &&
		d.DagInstance.EventPersistence == DagInstanceEventPersistenceSql &&
		d.DagInstance.Mode != DagInstanceModeVM {
		_ = d.DagInstance.WriteEventByVariableMap(context.Background(), updates, time.Now().UnixMicro())
	}

	if d.Save != nil {
		// 在保存之前释放锁
		d.mutex.Unlock()
		err := d.Save(d)
		d.mutex.Lock() // 重新获取锁

		if err != nil {
			log.Error("batch save share data failed",
				"err", err,
				"updates", updates)
		}
	}
}

// 添加新的批量获取方法
func (d *ShareData) BatchGet(keys []string) map[string]interface{} {
	_ = d.Load(context.Background(), keys)
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if d.Dict == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})
	for _, key := range keys {
		if val, ok := d.Dict[key]; ok {
			result[key] = val
		}
	}
	return result
}

// GetAll 获取所有数据的副本，线程安全
func (d *ShareData) GetAll() map[string]interface{} {
	_ = d.Load(context.Background(), nil)

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if d.Dict == nil {
		return make(map[string]interface{})
	}

	// 创建副本
	result := make(map[string]interface{})
	for k, v := range d.Dict {
		result[k] = v
	}
	return result
}

type MemoryShareData struct {
	mutex sync.RWMutex
	dict  map[string]interface{}
}

// Get value from share data, it is thread-safe.
func (md *MemoryShareData) Get(key string) (interface{}, bool) {
	if md.dict == nil {
		return "", false
	}
	md.mutex.RLock() // 使用读锁
	defer md.mutex.RUnlock()

	v, ok := md.dict[key]
	return v, ok
}

// Set value to share data, it is thread-safe.
func (md *MemoryShareData) Set(key string, val interface{}) {
	md.mutex.Lock() // 写操作使用写锁
	defer md.mutex.Unlock()

	if md.dict == nil {
		md.dict = make(map[string]interface{})
	}

	if strings.HasPrefix(key, "__") {
		md.dict[key] = val
	} else if strings.Contains(key, "_i") {
		// 使用正则表达式匹配 _i{number}_s{id} 格式
		re := regexp.MustCompile(`_i\d+_s([^_]+)$`)
		matches := re.FindStringSubmatch(key)
		if len(matches) > 1 {
			originKey := matches[1]
			md.dict[fmt.Sprintf("__%s", originKey)] = val
			md.dict[fmt.Sprintf("__%s", key)] = val
		} else {
			// 如果不是 _i{number}_s{id} 格式，则取最后一个部分
			keys := strings.Split(key, "_")
			originKey := keys[len(keys)-1]
			md.dict[fmt.Sprintf("__%s", originKey)] = val
			md.dict[fmt.Sprintf("__%s", key)] = val

		}
	} else if key != "" {
		md.dict[fmt.Sprintf("__%s", key)] = val
	}
}

// DagInstanceVars dag instance variables
type DagInstanceVars map[string]DagInstanceVar

// Cancel a task, it is just set a command, command will execute by Parser
func (dagIns *DagInstance) Cancel(taskInsIds []string) error {
	if dagIns.Status != DagInstanceStatusRunning {
		return fmt.Errorf("you can only cancel a running dag instance")
	}
	if dagIns.Cmd != nil {
		return fmt.Errorf("dag instance have a incomplete command")
	}
	dagIns.Cmd = &Command{
		Name:             CommandNameCancel,
		TargetTaskInsIDs: taskInsIds,
	}
	return nil
}

var (
	// HookDagInstance hook dag instance
	HookDagInstance DagInstanceLifecycleHook
)

// DagInstanceHookFunc type
type DagInstanceHookFunc func(dagIns *DagInstance)

// DagInstanceLifecycleHook dag instance lifecycle hook
type DagInstanceLifecycleHook struct {
	BeforeRun     DagInstanceHookFunc
	BeforeSuccess DagInstanceHookFunc
	BeforeFail    DagInstanceHookFunc
	BeforeBlock   DagInstanceHookFunc
	BeforeRetry   DagInstanceHookFunc
}

// VarsGetter the method of get variables
func (dagIns *DagInstance) VarsGetter() utils.KeyValueGetter {
	return func(key string) (interface{}, bool) {
		val, ok := dagIns.Vars[key]
		return val.Value, ok
	}
}

// VarsIterator variables iterator
func (dagIns *DagInstance) VarsIterator() utils.KeyValueIterator {
	return func(iterateFunc utils.KeyValueIterateFunc) {
		for k, v := range dagIns.Vars {
			if iterateFunc(k, v.Value) {
				break
			}
		}
	}
}

type DagInstanceEvent struct {
	ID         uint64                         `json:"-"`
	Type       rds.DagInstanceEventType       `json:"type,omitempty"`
	InstanceID string                         `json:"-"`
	Operator   string                         `json:"operator,omitempty"`
	TaskID     string                         `json:"task_id,omitempty"`
	Status     string                         `json:"status,omitempty"`
	Name       string                         `json:"name,omitempty"`
	Data       any                            `json:"data,omitempty"`
	Size       int                            `json:"-"`
	Inline     bool                           `json:"-"`
	Visibility rds.DagInstanceEventVisibility `json:"-"`
	Timestamp  int64                          `json:"timestamp,omitempty"`
}

func ToRdsEvent(ctx context.Context, ev *DagInstanceEvent) (*rds.DagInstanceEvent, error) {
	config := common.NewConfig()
	og := drivenadapters.NewOssGateWay()

	event := &rds.DagInstanceEvent{
		ID:         store.NextID(),
		Type:       ev.Type,
		InstanceID: ev.InstanceID,
		Operator:   ev.Operator,
		TaskID:     ev.TaskID,
		Status:     ev.Status,
		Name:       ev.Name,
		Data:       "",
		Inline:     true,
		Visibility: ev.Visibility,
		Timestamp:  ev.Timestamp,
	}

	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMicro()
	}

	if ev.Data != nil {
		b, _ := json.Marshal(ev.Data)
		event.Data = string(b)
		event.Size = len(event.Data)
		if config.Server.DagInstanceEventMaxInlineSize != -1 && event.Size > config.Server.DagInstanceEventMaxInlineSize {
			ossID, err := og.GetAvaildOSS(ctx)
			if err != nil {
				return nil, err
			}

			size := int64(event.Size)
			ossKey := fmt.Sprintf("%s/dag_instance/%s/event_%d", config.Server.StoragePrefix, event.InstanceID, event.ID)
			err = og.UploadFile(ctx, ossID, ossKey, true, bytes.NewReader([]byte(event.Data)), size)
			if err != nil {
				return nil, err
			}
			event.Inline = false
			event.Data = fmt.Sprintf("%s/%s", ossID, ossKey)
		}
	}

	return event, nil
}

func FromRdsEvent(ctx context.Context, ev *rds.DagInstanceEvent) (*DagInstanceEvent, error) {

	var (
		dataStr string
		data    any
	)

	if ev.Inline {
		dataStr = ev.Data
	} else {
		parts := strings.SplitN(ev.Data, "/", 2)
		ossID, ossKey := parts[0], parts[1]
		if ossID != "" && ossKey != "" {
			og := drivenadapters.NewOssGateWay()
			b, err := og.DownloadFile(ctx, ossID, ossKey, true)
			if err != nil {
				return nil, err
			}
			dataStr = string(b)
		}
	}

	switch ev.Type {
	case rds.DagInstanceEventTypeInstructions, rds.DagInstanceEventTypeVM:
		data = dataStr
	default:
		if dataStr != "" {
			if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
				return nil, err
			}
		}
	}

	event := &DagInstanceEvent{
		ID:         ev.ID,
		Type:       ev.Type,
		InstanceID: ev.InstanceID,
		Operator:   ev.Operator,
		TaskID:     ev.TaskID,
		Status:     ev.Status,
		Name:       ev.Name,
		Data:       data,
		Size:       ev.Size,
		Inline:     ev.Inline,
		Visibility: ev.Visibility,
		Timestamp:  ev.Timestamp,
	}

	return event, nil
}

func (dagIns *DagInstance) WriteEventByVariableMap(ctx context.Context, m map[string]any, timestamp int64) error {

	var events []*DagInstanceEvent

	if timestamp == 0 {
		timestamp = time.Now().UnixMicro()
	}

	for k, v := range m {
		event := &DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeVariable,
			InstanceID: dagIns.ID,
			Name:       k,
			Data:       v,
			Visibility: rds.DagInstanceEventVisibilityPrivate,
			Timestamp:  timestamp,
		}

		if regexp.MustCompile(`^__[0-9]+$`).MatchString(k) {
			event.Visibility = rds.DagInstanceEventVisibilityPublic
		}

		events = append(events, event)
	}

	return dagIns.WriteEvents(ctx, events)
}

// WriteTraceEvent 写入Trace变更信息
func (dagIns *DagInstance) WriteTraceEvent(ctx context.Context, m map[string]any) error {
	var events []*DagInstanceEvent

	now := time.Now().UnixMicro()
	keyReg := regexp.MustCompile(`__[a-zA-Z0-9_]+_trace`)

	for k, v := range m {
		if !keyReg.MatchString(k) {
			continue
		}

		event := &DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeTrace,
			InstanceID: dagIns.ID,
			Name:       k,
			Data:       v,
			Visibility: rds.DagInstanceEventVisibilityPublic,
			Timestamp:  now,
		}

		events = append(events, event)
	}

	return dagIns.WriteEvents(ctx, events)
}

func (dagIns *DagInstance) WriteEvents(ctx context.Context, events []*DagInstanceEvent) error {
	eventRepo := rds.GetDagInstanceEventRepository()
	rdsEvents := make([]*rds.DagInstanceEvent, 0, len(events))
	for _, event := range events {
		ev, err := ToRdsEvent(ctx, event)

		if err != nil {
			return err
		}
		rdsEvents = append(rdsEvents, ev)
	}

	err := eventRepo.InsertMany(ctx, rdsEvents)

	if err != nil {
		log.Warnf("[dagIns.WriteEvents] err: %s", err.Error())
	}

	return err
}

func (dagIns *DagInstance) UploadEvents(ctx context.Context) error {

	config := common.NewConfig()
	og := drivenadapters.NewOssGateWay()
	eventRepo := rds.GetDagInstanceEventRepository()

	ossId, err := og.GetAvaildOSS(ctx)

	if err != nil {
		return err
	}

	ossKey := fmt.Sprintf("%s/dag_instance/%s/events_%s.jsonl", config.Server.StoragePrefix, dagIns.ID, time.Now().Format("20060102150405"))

	pr, pw := io.Pipe()
	go func() {
		var err error

		defer func() {
			if err != nil {
				pw.CloseWithError(err)
			} else {
				pw.Close()
			}
		}()

		batchSize := 50
		opts := &rds.DagInstanceEventListOptions{
			DagInstanceID: dagIns.ID,
			Offset:        0,
			Limit:         batchSize,
			Visibilities:  []rds.DagInstanceEventVisibility{rds.DagInstanceEventVisibilityPublic},
		}

		for {
			var rdsEvents []*rds.DagInstanceEvent
			if rdsEvents, err = eventRepo.List(context.Background(), opts); err != nil {
				return
			}

			if len(rdsEvents) == 0 {
				return
			}

			for _, ev := range rdsEvents {
				var event *DagInstanceEvent
				if event, err = FromRdsEvent(ctx, ev); err != nil {
					return
				}

				b, _ := json.Marshal(event)

				if _, err = pw.Write(b); err != nil {
					return
				}

				if _, err = pw.Write([]byte("\n\n")); err != nil {
					return
				}
			}

			opts.Offset += batchSize
		}
	}()

	err = og.SimpleUpload(ctx, ossId, ossKey, true, pr)

	if err != nil {
		return err
	}

	dagIns.EventPersistence = DagInstanceEventPersistenceOss
	dagIns.EventOssPath = fmt.Sprintf("%s/%s", ossId, ossKey)

	return nil
}

func (dagIns *DagInstance) ListEvents(ctx context.Context, opts *rds.DagInstanceEventListOptions) ([]*DagInstanceEvent, error) {

	eventRepo := rds.GetDagInstanceEventRepository()

	if opts == nil {
		opts = &rds.DagInstanceEventListOptions{
			DagInstanceID: dagIns.ID,
		}
	} else if opts.DagInstanceID == "" {
		opts.DagInstanceID = dagIns.ID
	}

	rdsEvents, err := eventRepo.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	events := make([]*DagInstanceEvent, 0, len(rdsEvents))

	for _, ev := range rdsEvents {
		event, err := FromRdsEvent(ctx, ev)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (dagIns *DagInstance) ListOssEvents(ctx context.Context) (events []*DagInstanceEvent, err error) {

	if dagIns.EventPersistence != DagInstanceEventPersistenceOss || dagIns.EventOssPath == "" {
		return
	}

	og := drivenadapters.NewOssGateWay()
	parts := strings.SplitN(dagIns.EventOssPath, "/", 2)
	ossID, ossKey := parts[0], parts[1]
	url, err := og.GetDownloadURL(ctx, ossID, ossKey, 0, true)
	if err != nil {
		return
	}
	client := &http.Client{
		Transport: otelhttp.NewTransport(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}),
	}

	resp, err := client.Get(url)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch file: %s", string(body))
	}

	decoder := json.NewDecoder(resp.Body)

	for {
		var elem DagInstanceEvent

		if err := decoder.Decode(&elem); err == io.EOF {
			break
		} else if err != nil {
			log.Warnf("[dagIns.ListOssEvents] decode err %s", err.Error())
			continue
		}

		events = append(events, &elem)
	}

	return
}

// Run the dag instance
func (dagIns *DagInstance) Run() {
	dagIns.executeHook(HookDagInstance.BeforeRun)
	dagIns.Status = DagInstanceStatusRunning
	dagIns.Reason = ""
}

// Success the dag instance
func (dagIns *DagInstance) Success() {
	dagIns.executeHook(HookDagInstance.BeforeSuccess)
	dagIns.Status = DagInstanceStatusSuccess
	dagIns.Reason = ""
	dagIns.EndedAt = time.Now().Unix()
}

// Fail the dag instance
func (dagIns *DagInstance) Fail(reason string) {
	dagIns.Reason = reason
	dagIns.executeHook(HookDagInstance.BeforeFail)
	dagIns.Status = DagInstanceStatusFailed
	dagIns.EndedAt = time.Now().Unix()
}

func (dagIns *DagInstance) FailDetail(reason map[string]any) {
	if detail, ok := reason["detail"]; ok {
		reason["detail"] = normalizeutil.NormalizeContainer(detail)
	}
	b, _ := json.Marshal(reason)
	dagIns.Fail(string(b))
}

// Block the dag instance
func (dagIns *DagInstance) Block(reason string) {
	dagIns.executeHook(HookDagInstance.BeforeBlock)
	dagIns.Status = DagInstanceStatusBlocked
	dagIns.EndedAt = time.Now().Unix()
}

// Retry a task, it is just set a command, command will execute by Parser
func (dagIns *DagInstance) Retry(taskInsIds []string) error {
	if dagIns.Cmd != nil {
		return fmt.Errorf("dag instance have a incomplete command")
	}

	dagIns.executeHook(HookDagInstance.BeforeRetry)
	dagIns.Cmd = &Command{
		Name:             CommandNameRetry,
		TargetTaskInsIDs: taskInsIds,
	}
	return nil
}

func (dagIns *DagInstance) executeHook(hookFunc DagInstanceHookFunc) {
	if hookFunc != nil {
		hookFunc(dagIns)
	}
}

// CanModifyStatus indicate if the dag instance can modify status
func (dagIns *DagInstance) CanModifyStatus() bool {
	return dagIns.Status != DagInstanceStatusFailed
}

// Render variables
func (vars DagInstanceVars) Render(p map[string]interface{}) (map[string]interface{}, error) {
	err := value.MapValue(p).WalkString(func(walkContext *value.WalkContext, s string) error {
		for varKey, varValue := range vars {
			s = strings.ReplaceAll(s, fmt.Sprintf("{{%s}}", varKey), varValue.Value)
		}
		walkContext.Setter(s)
		return nil
	})
	return p, err
}

func (dagIns *DagInstance) Lock(ttl time.Duration) bool {
	rdb := libstore.NewRedis()
	return rdb.TryLock(DagInstanceLock+dagIns.ID, "", ttl) == nil
}

func (dagIns *DagInstance) Unlock() {
	rdb := libstore.NewRedis()
	_, _ = rdb.Unlock(DagInstanceLock+dagIns.ID, "")
}

type AsyncResponseStatus string

const (
	AsyncResponseStatusCompleted AsyncResponseStatus = "COMPLETED"
	AsyncResponseStatusFailed    AsyncResponseStatus = "FAILED"
)

type AsyncResponseData struct {
	TaskID    string              `json:"task_id"`
	Status    AsyncResponseStatus `json:"status"`
	Timestamp string              `json:"timestamp"`
	Data      any                 `json:"data,omitempty"`
	Error     any                 `json:"error,omitempty"`
}

func (dagIns *DagInstance) SendSuccessCallback(data any) (err error) {
	if dagIns.SuccessCallback == "" {
		return
	}

	client := drivenadapters.NewOtelHTTPClient()
	_, _, err = client.Post(context.Background(), dagIns.SuccessCallback, map[string]string{},
		&AsyncResponseData{
			TaskID:    dagIns.ID,
			Status:    AsyncResponseStatusCompleted,
			Timestamp: time.Now().Format(time.RFC3339),
			Data:      data,
		})
	if err != nil {
		log.Warnf("[dagIns.SendSuccessCallback] 调用 SuccessCallback %s 失败: %v \n", dagIns.SuccessCallback, err)
	}

	return
}

func (dagIns *DagInstance) SendErrorCallback(e error) (err error) {
	if dagIns.ErrorCallback == "" {
		return
	}

	client := drivenadapters.NewOtelHTTPClient()
	_, _, err = client.Post(context.Background(), dagIns.ErrorCallback, map[string]string{},
		&AsyncResponseData{
			TaskID:    dagIns.ID,
			Status:    AsyncResponseStatusFailed,
			Timestamp: time.Now().Format(time.RFC3339),
			Error:     e,
		})
	if err != nil {
		log.Warnf("[dagIns.SendErrorCallback] 调用 ErrorCallback %s 失败: %v \n", dagIns.ErrorCallback, err)
	}
	return
}

// Command struct
type Command struct {
	Name             CommandName
	TargetTaskInsIDs []string
}

// CommandName name
type CommandName string

const (
	// CommandNameRetry retry
	CommandNameRetry = "retry"
	// CommandNameCancel cancel
	CommandNameCancel = "cancel"
)

// DagInstanceStatus status
type DagInstanceStatus string

const (
	// DagInstanceStatusInit init status
	DagInstanceStatusInit DagInstanceStatus = "init"
	// DagInstanceStatusScheduled scheduled status
	DagInstanceStatusScheduled DagInstanceStatus = "scheduled"
	// DagInstanceStatusRunning running status
	DagInstanceStatusRunning DagInstanceStatus = "running"
	// DagInstanceStatusBlocked blocked status
	DagInstanceStatusBlocked DagInstanceStatus = "blocked"
	// DagInstanceStatusFailed failed status
	DagInstanceStatusFailed DagInstanceStatus = "failed"
	// DagInstanceStatusSuccess success status
	DagInstanceStatusSuccess DagInstanceStatus = "success"
	// DagInstanceStatusCancled canceled status
	DagInstanceStatusCancled DagInstanceStatus = "canceled"
)

// Trigger trigger
type Trigger string

const (
	// TriggerManually 手动触发器
	TriggerManually Trigger = "manually"
	// TriggerCron 定时触发器
	TriggerCron Trigger = "cron"
	// TriggerEvent 事件触发器
	TriggerEvent Trigger = "event"
	// TriggerWebhook webhook触发
	TriggerWebhook Trigger = "webhook"
	// FormWebhook webhook触发
	TriggerForm Trigger = "form"
	// 右键文件触发
	TriggerDocument Trigger = "document"
	// 安全策略触发
	TriggerSecurityPolicy Trigger = "security-policy"
)

type DagInstanceGroup struct {
	Total  int64        `bson:"total"`
	DagIns *DagInstance `bson:"latestRecord"`
}

type DagInstanceExtData struct {
	*rds.DagInstanceExtData `bson:"inline"`
}

func NewDagInstanceExtData(dagId string, dagInsID string, field string) *DagInstanceExtData {
	extData := &DagInstanceExtData{
		DagInstanceExtData: &rds.DagInstanceExtData{
			ID:        store.NextStringID(),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			DagID:     dagId,
			DagInsID:  dagInsID,
			Field:     field,
			Removed:   false,
		},
	}
	return extData
}

func (a *DagInstanceExtData) Save(ctx context.Context, data []byte) error {
	og := drivenadapters.NewOssGateWay()
	config := common.NewConfig()

	ossID, err := og.GetAvaildOSS(ctx)
	if err != nil {
		return err
	}
	size := int64(len(data))
	ossKey := fmt.Sprintf("%s/dag_instance/%s/%s_%s", config.Server.StoragePrefix, a.DagInsID, a.Field, a.ID)
	err = og.UploadFile(ctx, ossID, ossKey, true, bytes.NewReader(data), size)
	if err != nil {
		return err
	}

	a.OssID = ossID
	a.OssKey = ossKey
	a.Size = size
	return nil
}

func (a *DagInstanceExtData) Read(ctx context.Context) ([]byte, error) {
	if a.OssID == "" || a.OssKey == "" {
		return nil, nil
	}

	og := drivenadapters.NewOssGateWay()

	maxRetries := 3
	count := 0

	for {
		data, err := og.DownloadFile(ctx, a.OssID, a.OssKey, true)
		if err == nil {
			return data, nil
		}

		count += 1
		if count > maxRetries {
			return nil, err
		}

		time.Sleep(time.Second)
	}
}

func (a *DagInstanceExtData) Delete(ctx context.Context) error {
	if a.OssID == "" || a.OssKey == "" {
		return nil
	}

	og := drivenadapters.NewOssGateWay()
	return og.DeleteFile(ctx, a.OssID, a.OssKey, true)
}

func (dagIns *DagInstance) uploadExtData(ctx context.Context, data []byte, field string, maxRetries int) (*DagInstanceExtData, error) {
	count := 0

	for {
		extData := NewDagInstanceExtData(dagIns.DagID, dagIns.ID, field)
		err := extData.Save(ctx, data)
		if err == nil {
			return extData, nil
		}
		count += 1

		if count > maxRetries {
			return nil, err
		}

		time.Sleep(time.Second)
	}
}

func (dagIns *DagInstance) SaveExtData(ctx context.Context) (err error) {

	switch dagIns.EventPersistence {
	case DagInstanceEventPersistenceOss, DagInstanceEventPersistenceSql:
		// 数据已存到 event 表
	default:
		// 使用互斥锁保护整个方法
		dagIns.extDataMutex.Lock()
		defer dagIns.extDataMutex.Unlock()

		config := common.NewConfig()
		extDataItems := make([]*DagInstanceExtData, 0)

		if dagIns.ShareData != nil {
			var copyDict map[string]any = dagIns.ShareData.GetAll()

			data, err := json.Marshal(copyDict)
			if err != nil {
				return err
			}
			size := len(data)
			if size > config.Server.MongoMaxInlineSize {
				extData, err := dagIns.uploadExtData(ctx, data, "shareData", 3)
				if err != nil {
					log.Warnf("[dagIns.SaveExtData] upload shareData failed, dagInsId: %s, err: %s", dagIns.ID, err.Error())
					return err
				}
				dagIns.ShareDataExt = extData
				extDataItems = append(extDataItems, extData)
			} else {
				dagIns.ShareDataExt = nil
			}
		}

		if dagIns.Dump != "" {
			data := []byte(dagIns.Dump)
			size := len(data)
			if size > config.Server.MongoMaxInlineSize {
				extData, err := dagIns.uploadExtData(ctx, data, "dump", 3)
				if err != nil {
					log.Warnf("[dagIns.SaveExtData] upload dump failed, dagInsId: %s, err: %s", dagIns.ID, err.Error())
					return err
				}
				dagIns.DumpExt = extData
				extDataItems = append(extDataItems, extData)
			} else {
				dagIns.DumpExt = nil
			}
		}

		if len(extDataItems) > 0 {
			var items []*rds.DagInstanceExtData
			for _, item := range extDataItems {
				items = append(items, item.DagInstanceExtData)
			}
			go func() {
				err := rds.GetDagInstanceExtDataDao().InsertMany(context.Background(), items)
				if err != nil {
					for _, item := range extDataItems {
						_ = item.Delete(context.Background())
					}
				}
			}()
		}
	}

	return
}

func (dagIns *DagInstance) LoadExtData(ctx context.Context) (err error) {

	if dagIns.ShareData == nil {
		dagIns.ShareData = &ShareData{}
	}

	dagIns.ShareData.DagInstance = dagIns

	switch dagIns.EventPersistence {
	case DagInstanceEventPersistenceOss, DagInstanceEventPersistenceSql:
		// nothing to do:
	default:
		{
			if dagIns.ShareDataExt != nil {
				data, err := dagIns.ShareDataExt.Read(ctx)
				if err != nil {
					return err
				}

				dict := make(map[string]any)

				err = json.Unmarshal(data, &dict)

				if err != nil {
					log.Warnf("[dagIns.LoadExtData] download shareData failed, dagInsId: %s, err: %s", dagIns.ID, err.Error())
					return err
				}

				dagIns.ShareData.mutex.Lock()
				dagIns.ShareData.Dict = dict
				dagIns.ShareData.mutex.Unlock()
			}

			if dagIns.DumpExt != nil {
				data, err := dagIns.DumpExt.Read(ctx)
				if err != nil {
					log.Warnf("[dagIns.LoadExtData] download dump failed, dagInsId: %s, err: %s", dagIns.ID, err.Error())
					return err
				}

				dagIns.Dump = string(data)
			}
		}
	}

	return nil
}
