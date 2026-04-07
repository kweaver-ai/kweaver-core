package entity

import (
	"context"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/dependency"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/utils"
)

type DebugContext struct {
	ctx            context.Context
	taskID         string
	taskInstance   *TaskInstance
	memShareData   ShareDataOperator
	varsGetter     func(string) (interface{}, bool)
	varsIterator   utils.KeyValueIterator
	paramsGetter   func(string) (interface{}, bool)
	executeMethods ExecuteMethods
	repo           dependency.Repo
}

func NewDebugExecuteContext(
	ctx context.Context,
	op ShareDataOperator,
	dagVars utils.KeyValueGetter,
	varsIterator utils.KeyValueIterator,
	taskParams utils.KeyValueInterfaceGetter,
	taskID string,
	taskIns *TaskInstance,
	executeMethods ExecuteMethods,
	repo dependency.Repo,
) *DebugContext {
	return &DebugContext{
		ctx:            ctx,
		taskID:         taskID,
		taskInstance:   taskIns,
		memShareData:   op,
		varsGetter:     dagVars,
		varsIterator:   varsIterator,
		paramsGetter:   taskParams,
		executeMethods: executeMethods,
		repo:           repo,
	}
}

var _ ExecuteContext = &DebugContext{}

// SetContext sets the context
func (d *DebugContext) SetContext(ctx context.Context) {
	d.ctx = ctx
}

// Context returns the context
func (d *DebugContext) Context() context.Context {
	return d.ctx
}

// WithValue attaches a value to the context
func (d *DebugContext) WithValue(key, value interface{}) {
	d.ctx = context.WithValue(d.ctx, key, value)
}

// ShareData returns the share data operator
func (d *DebugContext) ShareData() ShareDataOperator {
	return d.memShareData
}

// Trace prints a message to traces
func (d *DebugContext) Trace(ctx context.Context, msg string, opt ...TraceOp) {
	// Implementation for trace logging
}

// Tracef prints a formatted message to traces
func (d *DebugContext) Tracef(ctx context.Context, msg string, a ...interface{}) {
}

// GetVar retrieves a variable by name
func (d *DebugContext) GetVar(varName string) (interface{}, bool) {
	return d.varsGetter(varName)
}

// IterateVars iterates through all variables
func (d *DebugContext) IterateVars(iterateFunc utils.KeyValueIterateFunc) {
	d.varsIterator(iterateFunc)
}

// GetParam retrieves a parameter by name
func (d *DebugContext) GetParam(paramName string) (interface{}, bool) {
	return d.paramsGetter(paramName)
}

// GetTaskID returns the task ID
func (d *DebugContext) GetTaskID() string {
	return d.taskID
}

// GetTaskInstance returns the task instance
func (d *DebugContext) GetTaskInstance() *TaskInstance {
	return d.taskInstance
}

// NewASDoc creates a new Efast document
func (d *DebugContext) NewASDoc() drivenadapters.Efast {
	return drivenadapters.NewEfast()
}

// NewExecuteMethods creates new execute methods
func (d *DebugContext) NewExecuteMethods() ExecuteMethods {
	return d.executeMethods
}

// NewRepo creates a new repository
func (d *DebugContext) NewRepo() dependency.Repo {
	return d.repo
}

// IsDebug returns true if the task is in debug mode.
// Debug mode is enabled by setting "single_debug" to "true" in the dagIns's runVars.
// In debug mode, the task will not write any trace to the database.
// This can be useful for testing tasks without polluting the database with test data.
func (d *DebugContext) IsDebug() bool {
	if v, ok := d.varsGetter("single_debug"); ok {
		return v == "true"
	}

	return false
}
