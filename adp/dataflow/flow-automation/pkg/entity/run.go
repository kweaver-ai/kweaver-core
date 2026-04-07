package entity

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/dependency"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/utils"
)

// PushMessage NSQ Publish method
type PushMessage func(topic string, message []byte) error

// GetDag query dag info
type GetDag func(ctx context.Context, id, versionID string) (*Dag, error)

type PatchDagIns func(ctx context.Context, dagIns *DagInstance, mustsPatchFields ...string) error

// ExecuteMethods 自定义方法结构体，用于外部方法注入
type ExecuteMethods struct {
	Publish     PushMessage
	GetDag      GetDag
	PatchDagIns PatchDagIns
}

// NewDefExecuteContext
func NewDefExecuteContext(
	ctx context.Context,
	op ShareDataOperator,
	trace func(ctx context.Context, msg string, opt ...TraceOp),
	dagVars utils.KeyValueGetter,
	varsIterator utils.KeyValueIterator,
	taskParams utils.KeyValueInterfaceGetter,
	taskID string,
	taskIns *TaskInstance,
	efast drivenadapters.Efast,
	executeMethods ExecuteMethods,
	repo dependency.Repo,
) *DefExecuteContext {
	return &DefExecuteContext{
		ctx:            ctx,
		op:             op,
		trace:          trace,
		varsGetter:     dagVars,
		varsIterator:   varsIterator,
		paramsGetter:   taskParams,
		taskID:         taskID,
		taskIns:        taskIns,
		efast:          efast,
		executeMethods: executeMethods,
		repo:           repo,
	}
}

// ExecuteContext is a context using by action
//go:generate mockgen -package entity -source ./run.go -destination ./run_mock.go

type ExecuteContext interface {
	SetContext(ctx context.Context)
	Context() context.Context
	// WithValue can attach value to context,so can share data between action
	// however it is base on memory, it is possible to lose changes such as application crash
	WithValue(key, value interface{})
	ShareData() ShareDataOperator
	// Trace print msg to the TaskInstance.Traces.
	Trace(ctx context.Context, msg string, opt ...TraceOp)
	// Tracef print msg to the TaskInstance.Traces.
	// Arguments are handled in the manner of fmt.Printf.
	// Opt can only be placed at the end of args.
	// Tracef("{format_str}",{format_val},{opts})
	// e.g. Tracef("%d", 1, TraceOpPersistAfterAction)
	// wrong case: Tracef("%d", TraceOpPersistAfterAction, 1)
	Tracef(ctx context.Context, msg string, a ...interface{})
	GetVar(varName string) (interface{}, bool)
	IterateVars(iterateFunc utils.KeyValueIterateFunc)
	GetParam(paramName string) (interface{}, bool)
	GetTaskID() string
	GetTaskInstance() *TaskInstance
	NewASDoc() drivenadapters.Efast
	NewExecuteMethods() ExecuteMethods
	NewRepo() dependency.Repo
	IsDebug() bool
}

// ShareDataOperator used to operate share data
type ShareDataOperator interface {
	Get(key string) (interface{}, bool)
	Set(key string, val interface{})
}

var _ ExecuteContext = &DefExecuteContext{}

// Default Executor context
type DefExecuteContext struct {
	ctx            context.Context
	op             ShareDataOperator
	taskID         string
	trace          func(ctx context.Context, msg string, opt ...TraceOp)
	varsGetter     func(string) (interface{}, bool)
	varsIterator   utils.KeyValueIterator
	paramsGetter   func(string) (interface{}, bool)
	efast          drivenadapters.Efast
	taskIns        *TaskInstance
	result         func(results interface{}) error
	executeMethods ExecuteMethods
	repo           dependency.Repo
}

func (e *DefExecuteContext) SetContext(ctx context.Context) {
	e.ctx = ctx
}

// Context
func (e *DefExecuteContext) Context() context.Context {
	return e.ctx
}

// WithValue is wrapper of "context.WithValue"
func (e *DefExecuteContext) WithValue(key, value interface{}) {
	e.ctx = context.WithValue(e.ctx, key, value)
}

// ShareData
func (e *DefExecuteContext) ShareData() ShareDataOperator {
	return e.op
}

// GetTaskID
func (e *DefExecuteContext) GetTaskID() string {
	return e.taskID
}

// GetTaskIns
func (e *DefExecuteContext) GetTaskInstance() *TaskInstance {
	return e.taskIns
}

// NewEfast
func (e *DefExecuteContext) NewASDoc() drivenadapters.Efast {
	return e.efast
}

// NewExecuteMethods 获取自定义封装方法体
func (e *DefExecuteContext) NewExecuteMethods() ExecuteMethods {
	return e.executeMethods
}

func (e *DefExecuteContext) NewRepo() dependency.Repo {
	return e.repo
}

// Trace print msg to the TaskInstance.Traces.
func (e *DefExecuteContext) Trace(ctx context.Context, msg string, opt ...TraceOp) {
	e.trace(ctx, msg, opt...)
}

// Tracef print msg to the TaskInstance.Traces.
// Arguments are handled in the manner of fmt.Printf.
// Opt can only be placed at the end of args.
// Tracef("{format_str}",{format_val},{opts})
// e.g. Tracef("%d", 1, TraceOpPersistAfterAction)
// wrong case: Tracef("%d", TraceOpPersistAfterAction, 1)
func (e *DefExecuteContext) Tracef(ctx context.Context, msg string, a ...interface{}) {
	args, ops := splitArgsAndOpt(a...)
	e.trace(ctx, fmt.Sprintf(msg, args...), ops...)
}

// splitArgsAndOpt split args and opt, opt must be placed at the end of args
func splitArgsAndOpt(a ...interface{}) ([]interface{}, []TraceOp) {
	optStartIndex := len(a)
	for i := len(a) - 1; i >= 0; i -= 1 {
		if _, ok := a[i].(TraceOp); !ok {
			optStartIndex = i + 1
			break
		}
		if i == 0 {
			optStartIndex = 0
		}
	}

	var traceOps []TraceOp
	for i := optStartIndex; i < len(a); i++ {
		traceOps = append(traceOps, a[i].(TraceOp))
	}

	return a[:optStartIndex], traceOps
}

// GetVar used to get key from ShareData
func (e *DefExecuteContext) GetVar(varName string) (interface{}, bool) {
	return e.varsGetter(varName)
}

// IterateVars used to iterate ShareData
func (e *DefExecuteContext) IterateVars(iterateFunc utils.KeyValueIterateFunc) {
	e.varsIterator(iterateFunc)
}

// GetParam used to get key from task
func (e *DefExecuteContext) GetParam(paramName string) (interface{}, bool) {
	return e.paramsGetter(paramName)
}

// IsDebug returns true if the task is in debug mode.
// Debug mode is enabled by setting "single_debug" to "true" in the dagIns's runVars.
// In debug mode, the task will not write any trace to the database.
// This can be useful for testing tasks without polluting the database with test data.
func (e *DefExecuteContext) IsDebug() bool {
	if v, ok := e.varsGetter("single_debug"); ok {
		return v == "true"
	}

	return false
}

// TraceOption
type TraceOption struct {
	Priority PersistPriority
}
type TraceOp func(opt *TraceOption)

// NewTraceOption
func NewTraceOption(ops ...TraceOp) *TraceOption {
	opt := &TraceOption{}
	for i := range ops {
		if ops[i] != nil {
			ops[i](opt)
		}
	}
	return opt
}

var (
	// TraceOpPersistAfterAction
	// Patch change when after execute each action("RunBefore", "Run" or "RunAfter")
	// this will be high performance, but here is a risk to lost trace when application crashed
	TraceOpPersistAfterAction TraceOp = func(opt *TraceOption) {
		opt.Priority = PersistPriorityAfterAction
	}
)

// PersistPriority
type PersistPriority string

const (
	// Patch change immediately, this will increase the burden of storage
	// the default behavior
	PersistPriorityImmediately = "Immediately"
	// Patch change when after execute each action("RunBefore", "Run" or "RunAfter")
	// this will be high performance, but here is a risk to lost trace when application crashed
	PersistPriorityAfterAction = "AfterAction"
)
