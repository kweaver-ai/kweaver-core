// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

type JobType string
type JobState string
type TaskState string

const (
	JobTypeFull        JobType = "full"
	JobTypeIncremental JobType = "incremental"

	MAX_STATE_DETAIL_SIZE int = 50000
)

const (
	JobStatePending   JobState = "pending"
	JobStateRunning   JobState = "running"
	JobStateCompleted JobState = "completed"
	JobStateCanceled  JobState = "canceled"
	JobStateFailed    JobState = "failed"
)

const (
	TaskStatePending   TaskState = "pending"
	TaskStateRunning   TaskState = "running"
	TaskStateCompleted TaskState = "completed"
	TaskStateCanceled  TaskState = "canceled"
	TaskStateFailed    TaskState = "failed"
)

type JobStateInfo struct {
	State       JobState `json:"state"`
	StateDetail string   `json:"state_detail"`
	FinishTime  int64    `json:"finish_time"`
	TimeCost    int64    `json:"time_cost"`
}

type TaskStateInfo struct {
	Index       string    `json:"index"`
	DocCount    int64     `json:"doc_count"`
	State       TaskState `json:"state"`
	StateDetail string    `json:"state_detail"`
	StartTime   int64     `json:"start_time"`
	FinishTime  int64     `json:"finish_time"`
	TimeCost    int64     `json:"time_cost"`
}

type ConceptConfig struct {
	ConceptType string `json:"concept_type"`
	ConceptID   string `json:"concept_id"`
}

// JobInfo 定义 job 结构
type JobInfo struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	KNID             string          `json:"kn_id"`
	Branch           string          `json:"branch"`
	JobType          JobType         `json:"job_type"`
	JobConceptConfig []ConceptConfig `json:"job_concept_config"`
	Creator          AccountInfo     `json:"creator"`
	CreateTime       int64           `json:"create_time"`

	JobStateInfo

	TaskInfos map[string]*TaskInfo `json:"tasks,omitempty"`
}

// TaskInfo 定义子任务结构
type TaskInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	JobID       string `json:"job_id"`
	ConceptType string `json:"concept_type"`
	ConceptID   string `json:"concept_id"`

	TaskStateInfo
}

var (
	JOB_SORT = map[string]string{
		"create_time": "f_create_time",
		"finish_time": "f_finish_time",
		"time_cost":   "f_time_cost",
	}
	TASK_SORT = map[string]string{
		"start_time":  "f_start_time",
		"finish_time": "f_finish_time",
		"time_cost":   "f_time_cost",
	}
)

// 任务的分页查询
type JobsQueryParams struct {
	PaginationQueryParameters
	NamePattern string
	KNID        string
	Branch      string
	JobType     JobType
	State       []JobState
}

type TasksQueryParams struct {
	PaginationQueryParameters
	KNID        string
	Branch      string
	JobID       string
	ConceptType string
	NamePattern string
	State       []TaskState
}
