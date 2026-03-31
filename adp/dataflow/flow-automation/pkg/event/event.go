package event

import "github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"

const (
	KeyDagInstanceUpdated = "DagInstanceUpdated"
	KeyDagInstancePatched = "DagInstancePatched"

	KeyTaskCompleted = "TaskCompleted"
	KeyTaskBegin     = "TaskBegin"

	KeyLeaderChanged                = "LeaderChanged"
	KeyDispatchInitDagInsCompleted  = "DispatchInitDagInsCompleted"
	KeyParseScheduleDagInsCompleted = "ParseScheduleDagInsCompleted"
)

// DagInstanceUpdated will raise when dag instance he updated
type DagInstanceUpdated struct {
	Payload *entity.DagInstance
}

// Topic
func (e *DagInstanceUpdated) Topic() []string {
	return []string{KeyDagInstanceUpdated}
}

// DagInstanceUpdated will raise when dag instance he updated
type DagInstancePatched struct {
	Payload         *entity.DagInstance
	MustPatchFields []string
}

// Topic
func (e *DagInstancePatched) Topic() []string {
	return []string{KeyDagInstancePatched}
}

// TaskCompleted will raise when executor completed a task instance,
type TaskCompleted struct {
	TaskIns *entity.TaskInstance
}

// Topic
func (e *TaskCompleted) Topic() []string {
	return []string{KeyTaskCompleted}
}

// TaskCompleted will raise when executor completed a task instance,
type TaskBegin struct {
	TaskIns *entity.TaskInstance
}

// Topic
func (e *TaskBegin) Topic() []string {
	return []string{KeyTaskBegin}
}

// LeaderChanged will raise when leader changed such as campaign success or continue leader failed
type LeaderChanged struct {
	IsLeader  bool
	WorkerKey string
}

// Topic
func (e *LeaderChanged) Topic() []string {
	return []string{KeyLeaderChanged}
}

// DispatchInitDagInsCompleted will raise when leader changed such as campaign success or continue leader failed
type DispatchInitDagInsCompleted struct {
	ElapsedMs int64
	Error     error
}

// Topic
func (e *DispatchInitDagInsCompleted) Topic() []string {
	return []string{KeyDispatchInitDagInsCompleted}
}

// ParseScheduleDagInsCompleted will raise when leader changed such as campaign success or continue leader failed
type ParseScheduleDagInsCompleted struct {
	ElapsedMs int64
	Error     error
}

// Topic
func (e *ParseScheduleDagInsCompleted) Topic() []string {
	return []string{KeyParseScheduleDagInsCompleted}
}
