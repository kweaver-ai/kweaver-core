package entity

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/utils"
)

// Task 解析后的task数据
type Task struct {
	ID          string                 `yaml:"id,omitempty" json:"id,omitempty"  bson:"id,omitempty"`
	Name        string                 `yaml:"name,omitempty" json:"name,omitempty"  bson:"name,omitempty"`
	DependOn    []string               `yaml:"dependOn,omitempty" json:"dependOn,omitempty"  bson:"dependOn,omitempty"`
	ActionName  string                 `yaml:"actionName,omitempty" json:"actionName,omitempty"  bson:"actionName,omitempty"`
	TimeoutSecs int                    `yaml:"timeoutSecs,omitempty" json:"timeoutSecs,omitempty"  bson:"timeoutSecs,omitempty"`
	Params      map[string]interface{} `yaml:"params,omitempty" json:"params,omitempty"  bson:"params,omitempty"`
	PreChecks   PreChecks              `yaml:"preCheck,omitempty" json:"preCheck,omitempty"  bson:"preCheck,omitempty"`
	Steps       []Step                 `yaml:"steps,omitempty" json:"steps,omitempty"  bson:"steps,omitempty"`
	BranchsNum  int                    `yaml:"branchsNum,omitempty" json:"branchsNum,omitempty"  bson:"branchsNum,omitempty"`
	BranchID    string                 `yaml:"branchsID,omitempty" json:"branchsID,omitempty"  bson:"branchsID,omitempty"`
	Priority    int                    `json:"priority,omitempty" bson:"priority,omitempty"`
	Settings    *Settings              `json:"settings,omitempty" bson:"settings,omitempty"`
}

// Step 每一步骤的执行数据
type Step struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Operator   string                 `json:"operator"`
	DataSource *DataSource            `json:"dataSource,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Cron       string                 `json:"cron,omitempty"`
	Branches   []Branch               `json:"branches,omitempty"`
	Steps      []Step                 `json:"steps,omitempty"`
	Settings   *Settings              `json:"settings,omitempty" bson:"settings,omitempty"`
}

type Settings struct {
	Retry   *RetryConfig   `json:"retry"`
	TimeOut *TimeoutConfig `json:"timeout"`
}

type RetryConfig struct {
	Max   int `json:"max"`
	Delay int `json:"delay"`
}

type TimeoutConfig struct {
	Delay int `json:"delay"`
}

// DataSource 数据源
type DataSource struct {
	ID         string           `json:"id,omitempty"`
	Operator   string           `json:"operator"`
	Parameters *DataSourceParam `json:"parameters,omitempty"`
}

// DataSourceParam 数据源参数
type DataSourceParam struct {
	DocID      string                   `json:"docid,omitempty"`
	DocIDs     []string                 `json:"docids,omitempty"`
	AccessorID string                   `json:"accessorid,omitempty"`
	Depth      int                      `json:"depth,omitempty"`
	Docs       []map[string]interface{} `json:"docs,omitempty"`
	Sources    []S3Source               `json:"sources,omitempty"` // S3数据源列表
	Mode       string                   `json:"mode,omitempty"`    // 数据源模式
}

// S3Source S3数据源配置
type S3Source struct {
	Bucket string `json:"bucket"` // S3 Bucket名称
	Path   string `json:"path"`   // S3对象路径前缀
	Key    string `json:"key"`    // S3对象Key
	Name   string `json:"name"`   // 文件名
	Size   int64  `json:"size"`   // 文件大小
}

// Branch 分支结构
type Branch struct {
	ID         string            `json:"id"`
	Conditions [][]TaskCondition `json:"conditions"`
	Steps      []Step            `json:"steps"`
}

// Condition 条件结构
type Condition struct {
	ID         string `json:"id"`
	Operator   string `json:"operator"`
	Parameters map[string]interface{}
}

// GetGraphID 获取task ID
func (t *Task) GetGraphID() string {
	return t.ID
}

// GetID 获取task ID
func (t *Task) GetID() string {
	return t.ID
}

// GetDepend 获取task 依赖
func (t *Task) GetDepend() []string {
	return t.DependOn
}

// GetStatus 获取task 状态
func (t *Task) GetStatus() TaskInstanceStatus {
	return ""
}

// PreChecks 预检查项
type PreChecks map[string]*Check

// Check 检查项
type Check struct {
	Conditions []TaskCondition `yaml:"conditions,omitempty" json:"conditions,omitempty"  bson:"conditions,omitempty"`
	Act        ActiveAction    `yaml:"act,omitempty" json:"act,omitempty"  bson:"act,omitempty"`
}

// IsMeet return if check is meet
func (c *Check) IsMeet(dagIns *DagInstance) bool {
	for _, cd := range c.Conditions {
		if !cd.IsMeet(dagIns) {
			return false
		}
	}
	return true
}

// ActiveAction active action
type ActiveAction string

const (
	// ActiveActionSkip skip action when all condition is meet, otherwise execute it
	ActiveActionSkip ActiveAction = "skip"
	// ActiveActionBlock block action when all condition is meet, otherwise execute it
	ActiveActionBlock ActiveAction = "block"
)

// Operator 比较操作类型
type Operator string

const (
	OperatorStringIn       Operator = "@internal/cmp/string-contains"     // OperatorStringIn 字符串包含
	OperatorStringNotIn    Operator = "@internal/cmp/string-not-contains" // OperatorStringNotIn 字符串不包含
	OperateStringEq        Operator = "@internal/cmp/string-eq"           // OperateStringEq 字符串等于
	OperateStringNeq       Operator = "@internal/cmp/string-neq"          // OperateStringNeq 字符串不等于
	OperateStringStartWith Operator = "@internal/cmp/string-start-with"   // OperateStringStartWith 以某字符串开头
	OperateStringEndWith   Operator = "@internal/cmp/string-end-with"     // OperateStringEndWith 以某字符串结尾
	OperateStringEmpty     Operator = common.BranchCmpStringEmpty         // OperateStringEmpty 字符串为空
	OperateStringNotEmpty  Operator = common.BranchCmpStringNotEmpty      // OperateStringNotEmpty 字符串不为空
	OperateStringMatch     Operator = common.BranchCmpStringMatch         // OperateStringMatch 字符串匹配正则表达式

	OperateNumberEq  Operator = "@internal/cmp/number-eq"  // OperateNumberEq 数字等于
	OperateNumberNeq Operator = "@internal/cmp/number-neq" // OperateNumberNeq 数字不等于
	OperateNumberGte Operator = "@internal/cmp/number-gte" // OperateNumberGte 数字不小于
	OperateNumberGt  Operator = "@internal/cmp/number-gt"  // OperateNumberGt 数字大于
	OperateNumberLt  Operator = "@internal/cmp/number-lt"  // OperateNumberLt 数字小于
	OperateNumberLte Operator = "@internal/cmp/number-lte" // OperateNumberLte 数字不大于

	OperateDateEq     Operator = "@internal/cmp/date-eq"           // OperateDateEq 时间等于
	OperateDateNeq    Operator = "@internal/cmp/date-neq"          // OperateDateNeq 时间不等于
	OperateDateBefore Operator = "@internal/cmp/date-earlier-than" // OperateDateBefore 时间早于
	OperateDateAfter  Operator = "@internal/cmp/date-later-than"   // OperateDateAfter 时间晚于

	OperateWorkflowEq  Operator = common.BranchWorkflowEq  // OperateWorkflowEq 审核结果等于
	OperateWorkflowNeq Operator = common.BranchWorkflowNeq // OperateWorkflowNeq 审核结果不等于
)

// TaskConditionSource 条件源数据
type TaskConditionSource string

const (
	// TaskConditionSourceVars 条件本地变量
	TaskConditionSourceVars TaskConditionSource = "vars"
	// TaskConditionSourceShareData 条件共享变量
	TaskConditionSourceShareData TaskConditionSource = "share-data"
)

// BuildKvGetter 变量获取
func (t TaskConditionSource) BuildKvGetter(dagIns *DagInstance) utils.KeyValueGetter {
	switch t {
	case TaskConditionSourceVars:
		return dagIns.VarsGetter()
	case TaskConditionSourceShareData:
		return dagIns.ShareData.Get
	default:
		return dagIns.ShareData.Get
	}
}

// TaskCondition 任务条件
type TaskCondition struct {
	ID           string                 `yaml:"id,omitempty" json:"id,omitempty"  bson:"id,omitempty"`
	Source       TaskConditionSource    `yaml:"source,omitempty" json:"source,omitempty"  bson:"source,omitempty"`
	Parameter    TaskConditionParameter `yaml:"parameters,omitempty" json:"parameters,omitempty"  bson:"parameters,omitempty"`
	Op           Operator               `yaml:"operator,omitempty" json:"operator,omitempty"  bson:"operator,omitempty"`
	ParamsRender ParamsRender           `yaml:"-,omitempty" json:"-"  bson:"-,omitempty"`
}

type TaskConditionParameter struct {
	A interface{} `yaml:"a,omitempty" json:"a,omitempty"  bson:"a,omitempty"`
	B interface{} `yaml:"b,omitempty" json:"b,omitempty"  bson:"b,omitempty"`
}

// IsMeet return if check is meet
func (c *TaskCondition) IsMeet(dagIns *DagInstance) bool { //nolint
	e := NewParamsRender()
	vl, _ := e.RenderParam(dagIns, c.Parameter.A)
	vr, _ := e.RenderParam(dagIns, c.Parameter.B)

	switch c.Op {
	case OperatorStringIn:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return strings.Contains(vls, vrs)
	case OperatorStringNotIn:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return !strings.Contains(vls, vrs)
	case OperateStringEq, OperateWorkflowEq:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return vrs == vls
	case OperateStringNeq, OperateWorkflowNeq:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return vrs != vls
	case OperateStringStartWith:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return strings.HasPrefix(vls, vrs)
	case OperateStringEndWith:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		return strings.HasSuffix(vls, vrs)
	case OperateStringEmpty:
		vls, _, ok := c.ParseString(vl, "")
		if !ok {
			return false
		}
		return vls == ""
	case OperateStringNotEmpty:
		vls, _, ok := c.ParseString(vl, "")
		if !ok {
			return false
		}
		return vls != ""
	case OperateStringMatch:
		vls, vrs, ok := c.ParseString(vl, vr)
		if !ok {
			return false
		}
		matched, err := regexp.MatchString(vrs, vls)
		if err != nil {
			return false
		}
		return matched
	case OperateNumberEq:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vrs == vls
	case OperateNumberNeq:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vrs != vls
	case OperateNumberGte:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vls >= vrs
	case OperateNumberGt:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vls > vrs
	case OperateNumberLte:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vls <= vrs
	case OperateNumberLt:
		vls, vrs, ok := c.ParseNumber(vl, vr)
		if !ok {
			return false
		}
		return vls < vrs
	case OperateDateEq:
		vls, vrs, ok := c.ParseDate(vl, vr)
		if !ok {
			return false
		}
		return vls == vrs
	case OperateDateNeq:
		vls, vrs, ok := c.ParseDate(vl, vr)
		if !ok {
			return true
		}
		return vls != vrs
	case OperateDateBefore:
		vls, vrs, ok := c.ParseDate(vl, vr)
		if !ok {
			return false
		}
		return vls < vrs
	case OperateDateAfter:
		vls, vrs, ok := c.ParseDate(vl, vr)
		if !ok {
			return false
		}
		return vls > vrs
	}

	return false
}

// ParseString 变量转换为string类型
func (c *TaskCondition) ParseString(vl, vr interface{}) (vls, vrs string, ok bool) {
	vls, ok = vl.(string)
	if !ok {
		return
	}
	vrs, ok = vr.(string)
	if !ok {
		return
	}

	return
}

// ParseNumber 转换数字
func (c *TaskCondition) ParseNumber(vl, vr interface{}) (vls, vrs float64, ok bool) {
	switch tl := vl.(type) {
	case string:
		floatvar, err := strconv.ParseFloat(tl, 64)
		if err != nil {
			vls = 0
			break
		}
		vls = floatvar
	case float64:
		vls = tl
	case int64:
		vls = float64(tl)
	}

	switch tr := vr.(type) {
	case string:
		floatvar, err := strconv.ParseFloat(tr, 64)
		if err != nil {
			vls = 0
			break
		}
		vrs = floatvar
	case float64:
		vrs = tr
	case int64:
		vrs = float64(tr)
	}
	return vls, vrs, true
}

// ParseDate 转换日期
func (c *TaskCondition) ParseDate(vl, vr interface{}) (vls, vrs int64, ok bool) {
	switch tl := vl.(type) {
	case string:
		v, err := time.Parse(time.RFC3339, tl)
		if err != nil {
			var newNum float64
			_, err := fmt.Sscanf(tl, "%e", &newNum)
			if err != nil {
				vls = 0
				break
			}
			vls = int64(newNum)
			break
		}
		vls = v.UnixNano() / 1e9 // 秒级
	case int64:
		vls = tl
	}

	switch tr := vr.(type) {
	case string:
		v, err := time.Parse(time.RFC3339, tr)
		if err != nil {
			var newNum float64
			_, err := fmt.Sscanf(tr, "%e", &newNum)
			if err != nil {
				vrs = 0
				break
			}
			vrs = int64(newNum)
			break
		}
		vrs = v.UnixNano() / 1e9 // 秒级
	case int64:
		vrs = tr
	}
	return vls, vrs, true
}

// TaskInstance task instance
type TaskInstance struct {
	BaseInfo `bson:"inline"`
	// Task's Id it should be unique in a dag instance
	TaskID      string                 `json:"taskId,omitempty" bson:"taskId,omitempty"`
	DagInsID    string                 `json:"dagInsId,omitempty" bson:"dagInsId,omitempty"`
	Name        string                 `json:"name,omitempty" bson:"name,omitempty"`
	DependOn    []string               `json:"dependOn,omitempty" bson:"dependOn,omitempty"`
	ActionName  string                 `json:"actionName,omitempty" bson:"actionName,omitempty"`
	TimeoutSecs int                    `json:"timeoutSecs" bson:"timeoutSecs"`
	Params      map[string]interface{} `json:"params,omitempty" bson:"params,omitempty"`
	Traces      []TraceInfo            `json:"traces,omitempty" bson:"traces,omitempty"`
	Status      TaskInstanceStatus     `json:"status,omitempty" bson:"status,omitempty"`
	Reason      interface{}            `json:"reason,omitempty" bson:"reason,omitempty"`
	PreChecks   PreChecks              `json:"preChecks,omitempty"  bson:"preChecks,omitempty"`
	Results     interface{}            `json:"results,omitempty"  bson:"results,omitempty"`
	Steps       []Step                 `json:"steps,omitempty" bson:"steps,omitempty"`
	// LastModifiedAt tracks the latest modification time with nanosecond precision
	LastModifiedAt int64                  `json:"lastModifiedAt,omitempty" bson:"lastModifiedAt,omitempty"`
	RenderedParams map[string]interface{} `json:"renderedParams,omitempty" bson:"renderedParams,omitempty"`
	Hash           string                 `json:"-" bson:"hash,omitempty"`
	Settings       *Settings              `json:"settings,omitempty" bson:"settings,omitempty"`
	MetaData       *TaskMetaData          `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// used to save changes
	Patch              func(context.Context, *TaskInstance) error `json:"-" bson:"-"`
	Context            ExecuteContext                             `json:"-" bson:"-"`
	RelatedDagInstance *DagInstance                               `json:"-" bson:"-"`

	// it used to buffer traces, and persist when status changed
	bufTraces []TraceInfo

	// mutex to protect Params field from concurrent access
	mu sync.RWMutex
}

// TraceInfo
type TraceInfo struct {
	Time    int64  `json:"time,omitempty" bson:"time,omitempty"`
	Message string `json:"message,omitempty" bson:"message,omitempty"`
}

// TaskMetaData 任务执行元信息
type TaskMetaData struct {
	Attempts  int   `json:"attempts"`
	MaxRetry  int   `json:"max_retry,omitempty"`
	StartedAt int64 `json:"started_at"`
	EndedAt   int64 `json:"ended_at,omitempty"`
	// 节点调度到执行器上运行时间
	Duration int64 `json:"duration"`
	// 节点从开始到完成总耗时
	ElapsedTime int64 `json:"elapsed_time"`
}

// NewTaskInstance
func NewTaskInstance(dagInsID string, t *Task) *TaskInstance {
	// 为未执行的任务设置lastModifiedAt为当前时间+200年（纳秒级），避免溢出
	now := time.Now()
	futureTime := now.AddDate(200, 0, 0).UnixNano()

	return &TaskInstance{
		TaskID:         t.ID,
		DagInsID:       dagInsID,
		Name:           t.Name,
		DependOn:       t.DependOn,
		ActionName:     t.ActionName,
		TimeoutSecs:    t.TimeoutSecs,
		Params:         t.Params,
		Status:         TaskInstanceStatusInit,
		PreChecks:      t.PreChecks,
		Steps:          t.Steps,
		LastModifiedAt: futureTime, // 设置为当前时间+200年，确保未执行任务排在后面
		Settings:       t.Settings,
	}
}

// GetGraphID
func (t *TaskInstance) GetGraphID() string {
	return t.TaskID
}

// GetDataSourceID
func (t *TaskInstance) GetDataSourceID() string {
	return t.TaskID
}

// GetID
func (t *TaskInstance) GetID() string {
	return t.ID
}

// GetDepend
func (t *TaskInstance) GetDepend() []string {
	return t.DependOn
}

// GetStatus
func (t *TaskInstance) GetStatus() TaskInstanceStatus {
	return t.Status
}

// GetParam
func (t *TaskInstance) GetParam() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Params
}

// GetParams returns a thread-safe copy of the Params map
func (t *TaskInstance) GetParams() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.Params == nil {
		return make(map[string]interface{})
	}

	// Create a copy to avoid concurrent access issues
	result := make(map[string]interface{})
	for k, v := range t.Params {
		result[k] = v
	}
	return result
}

// SetParams sets the Params field in a thread-safe manner
func (t *TaskInstance) SetParams(params map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Params = params
}

// SetParam sets a single parameter in a thread-safe manner
func (t *TaskInstance) SetParam(key string, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Params == nil {
		t.Params = make(map[string]interface{})
	}
	t.Params[key] = value
}

// ParamsGetter
func (t *TaskInstance) ParamsGetter() utils.KeyValueInterfaceGetter {
	return func(key string) (interface{}, bool) {
		t.mu.RLock()
		defer t.mu.RUnlock()
		val, ok := t.Params[key]
		return val, ok
	}
}

// InitialDep
func (t *TaskInstance) InitialDep(ctx ExecuteContext, patch func(context.Context, *TaskInstance) error, dagIns *DagInstance) {
	t.Patch = patch
	t.Context = ctx
	t.RelatedDagInstance = dagIns
}

func IsCtxDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// SetStatus will persist task instance
func (t *TaskInstance) SetStatus(ctx context.Context, s TaskInstanceStatus) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	done := IsCtxDone(ctx)
	if done {
		s = TaskInstanceStatusFailed
	}

	t.Status = s
	t.LastModifiedAt = time.Now().UnixNano()
	patch := &TaskInstance{BaseInfo: BaseInfo{ID: t.ID}, Status: t.Status, Reason: t.Reason, LastModifiedAt: t.LastModifiedAt}
	if len(t.bufTraces) != 0 {
		copy(patch.Traces, t.Traces)
		patch.Traces = append(patch.Traces, t.bufTraces...)
	}

	dagIns := t.RelatedDagInstance
	if dagIns != nil &&
		dagIns.EventPersistence == DagInstanceEventPersistenceSql &&
		dagIns.Mode != DagInstanceModeVM &&
		t.Status != TaskInstanceStatusEnding {

		taskID := t.TaskID

		for _, re := range []string{`^\d+_i\d+_s\d+.+_(\d+)$`, `^\d+_i\d+_s(\d+)$`, `^(\d+)_i\d+$`} {
			if matches := regexp.MustCompile(re).FindStringSubmatch(t.TaskID); len(matches) == 2 {
				taskID = matches[1]
				break
			}
		}

		err = dagIns.WriteEvents(ctx, []*DagInstanceEvent{
			{
				Type:       rds.DagInstanceEventTypeTaskStatus,
				InstanceID: dagIns.ID,
				Operator:   t.ActionName,
				TaskID:     taskID,
				Status:     string(t.Status),
				Data:       t.Reason,
				Timestamp:  time.Now().UnixMicro(),
				Visibility: rds.DagInstanceEventVisibilityPublic,
			},
		})

		if err != nil {
			return err
		}
	}

	err = t.Patch(ctx, patch)

	if err == nil && done {
		err = ctx.Err()
	}
	return err
}

// Trace info
func (t *TaskInstance) Trace(ctx context.Context, msg string, ops ...TraceOp) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	opt := NewTraceOption(ops...)
	if opt.Priority == PersistPriorityAfterAction {
		t.bufTraces = append(t.bufTraces, TraceInfo{
			Time:    time.Now().Unix(),
			Message: msg,
		})
		return
	}

	t.Traces = append(t.Traces, TraceInfo{
		Time:    time.Now().Unix(),
		Message: msg,
	})

	t.LastModifiedAt = time.Now().UnixNano()
	if err := t.Patch(ctx, &TaskInstance{BaseInfo: BaseInfo{ID: t.ID}, Traces: t.Traces, LastModifiedAt: t.LastModifiedAt}); err != nil {
		log.Error("save trace failed",
			"err", err,
			"trace", t.Traces)
	}
}

// Run action
func (t *TaskInstance) Run(ctx context.Context, params interface{}, act Action, tokenInfo *Token) (err error) { //nolint
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	defer func() {
		if rErr := recover(); rErr != nil {
			var stacktrace string
			for i := 2; ; i++ {
				_, f, l, got := runtime.Caller(i)
				if !got {
					break
				}
				stacktrace += fmt.Sprintf("%s:%d\n", f, l)
			}
			err = fmt.Errorf("get panic when running action: %s, err: %s, stack: %s", act.Name(), rErr, stacktrace)
		}
	}()

	// 将ar生成的ctx传递给taskIns的ctx用于链路记录
	t.Context.SetContext(ctx)
	if t.Status == TaskInstanceStatusInit {
		beforeAct, ok := act.(BeforeAction)
		beforeRunStatus := TaskInstanceStatusRunning
		var err error
		if ok {
			if beforeRunStatus, err = beforeAct.RunBefore(t.Context, params); err != nil {
				return err
			}
		}
		if err = t.SetStatus(ctx, beforeRunStatus); err != nil { //nolint
			return err
		}

		result, err := act.Run(t.Context, params, tokenInfo)
		if err != nil {
			return err
		}

		err = t.SetResult(ctx, result)
		if err != nil {
			return err
		}

		if beforeRunStatus == TaskInstanceStatusRunning {
			if err := t.SetStatus(ctx, TaskInstanceStatusEnding); err != nil {
				return err
			}
		}
	}

	if t.Status == TaskInstanceStatusRunning {
		result, err := act.Run(t.Context, params, tokenInfo)
		if err != nil {
			return err
		}

		err = t.SetResult(ctx, result)
		if err != nil {
			return err
		}

		if err := t.SetStatus(ctx, TaskInstanceStatusEnding); err != nil {
			return err
		}
	}

	if t.Status == TaskInstanceStatusEnding {
		afterAct, ok := act.(AfterAction)
		afterRunStatus := TaskInstanceStatusSuccess
		var err error
		if ok {
			if afterRunStatus, err = afterAct.RunAfter(t.Context, params); err != nil {
				return err
			}
		}
		if err := t.SetStatus(ctx, afterRunStatus); err != nil {
			return err
		}
	}

	if t.Status == TaskInstanceStatusRetrying {
		retryAct, ok := act.(RetryBeforeAction)
		if ok {
			if err := retryAct.RetryBefore(t.Context, params); err != nil {
				return fmt.Errorf("run retryBefore failed: %w", err)
			}
		}
		if err := t.SetStatus(ctx, TaskInstanceStatusInit); err != nil {
			return err
		}
	}
	return nil
}

// DoPreCheck
func (t *TaskInstance) DoPreCheck(dagIns *DagInstance) (isActive bool, err error) {
	if t.PreChecks == nil {
		return
	}

	status := t.Status

	checkResMap := make(map[string]bool, 0)
	for k, c := range t.PreChecks {
		key := strings.Split(k, "_")[0]
		active := false

		if !c.IsMeet(dagIns) {
			switch c.Act {
			case ActiveActionSkip:
				status = TaskInstanceStatusSkipped
			case ActiveActionBlock:
				status = TaskInstanceStatusBlocked
			default:
				status = TaskInstanceStatusSkipped
			}
			active = true
		} else {
			active = false
		}
		if _, ok := checkResMap[key]; ok {
			if !active {
				checkResMap[key] = active
			}
		} else {
			checkResMap[key] = active
		}
	}

	for _, c := range checkResMap {
		if c {
			isActive = c
			break
		}
	}

	if isActive {
		t.Status = status
	}

	return
}

func (t *TaskInstance) SetResult(ctx context.Context, results interface{}) error {
	if results == nil {
		return nil
	}

	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	done := IsCtxDone(ctx)
	if done {
		t.Status = TaskInstanceStatusFailed
		results = ctx.Err()
	}

	t.Results = results
	t.LastModifiedAt = time.Now().UnixNano()
	patch := &TaskInstance{BaseInfo: BaseInfo{ID: t.ID}, Results: t.Results, LastModifiedAt: t.LastModifiedAt}
	if err := t.Patch(ctx, patch); err != nil {
		return fmt.Errorf("set result err, dagins id: %s, taskins id: %s, result: %v", t.DagInsID, t.TaskID, results)
	}

	if err == nil && done {
		err = ctx.Err()
	}

	return nil
}

// TaskInstanceStatus
type TaskInstanceStatus string

func (ts TaskInstanceStatus) ToString() string {
	return string(ts)
}

const (
	TaskInstanceStatusInit     TaskInstanceStatus = "init"
	TaskInstanceStatusCanceled TaskInstanceStatus = "canceled"
	TaskInstanceStatusRunning  TaskInstanceStatus = "running"
	TaskInstanceStatusEnding   TaskInstanceStatus = "ending"
	TaskInstanceStatusFailed   TaskInstanceStatus = "failed"
	TaskInstanceStatusRetrying TaskInstanceStatus = "retrying"
	TaskInstanceStatusSuccess  TaskInstanceStatus = "success"
	TaskInstanceStatusBlocked  TaskInstanceStatus = "blocked"
	TaskInstanceStatusSkipped  TaskInstanceStatus = "skipped"
)

const ContentEntityKeyPrefix string = "automation:action:contententity:"
const EcoconfigReindexKeyPrefix string = "automation:action:ecoconfigindex:"

const DagInstanceLock string = "automation:daginslock:"
const DataViewTriggerLock string = "automation:dataviewtrigger:"
