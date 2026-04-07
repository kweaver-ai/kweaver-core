// Package actions 动作节点定义
package actions

import (
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

const (
	controlFlow string = "@control/flow"
	// ControlFlowBranches 逻辑分支
	ControlFlowBranches string = controlFlow + "/branches"
	// ControlFlowParallel 并行分支
	ControlFlowParallel string = controlFlow + "/parallel"
	// ControlFlowCondition 逻辑分支条件
	ControlFlowCondition string = controlFlow + "/condition"
)

// LogicBranch 逻辑分支
type LogicBranch struct {
}

// Name 操作名称
func (a *LogicBranch) Name() string {
	return ControlFlowBranches
}

// Run 操作方法
func (a *LogicBranch) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	ctx.Trace(ctx.Context(), "run end")
	return nil, nil
}

// ParameterNew 初始化参数
func (a *LogicBranch) ParameterNew() interface{} {
	return &LogicBranch{}
}

// ParallelBranch 并行分支
type ParallelBranch struct {
}

// Name 操作名称
func (a *ParallelBranch) Name() string {
	return ControlFlowParallel
}

// Run 操作方法
func (a *ParallelBranch) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	ctx.Trace(ctx.Context(), "parallel branch start", entity.TraceOpPersistAfterAction)
	ctx.Trace(ctx.Context(), "parallel branch end")
	return nil, nil
}

// ParameterNew 初始化参数
func (a *ParallelBranch) ParameterNew() interface{} {
	return &ParallelBranch{}
}

// LoopParameters defines the parameters for loop execution
type LoopParameters struct {
	Mode                string        `json:"mode"`              // "limit" or "array"
	Limit               int           `json:"limit"`             // Maximum iteration count for limit mode
	Array               interface{}   `json:"array"`             // Array to iterate over for array mode
	Outputs             []LoopOutput  `json:"outputs"`           // Output configuration
	CurrentIteration    int           `json:"current_iteration"` // Current iteration (0-based)
	LastIterationTaskID string        `json:"last_iteration_task_id"`
	LoopControlID       string        `json:"loop_control_id"`
	Steps               []entity.Step `json:"steps"`
}

// LoopOutput defines output configuration for loop
type LoopOutput struct {
	Key   string `json:"key"`   // Output key name
	Value string `json:"value"` // Template for output value (e.g., "{{__2.data.result}}")
}

// Loop represents a loop action that executes a set of steps multiple times
type Loop struct{}

// Name returns the name of the action
func (l *Loop) Name() string {
	return common.Loop
}

// Run executes the loop action
func (l *Loop) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	return nil, nil
}

// ParameterNew returns a new instance of LoopParameters
func (l *Loop) ParameterNew() interface{} {
	return &LoopParameters{}
}
