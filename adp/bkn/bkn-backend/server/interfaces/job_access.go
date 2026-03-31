// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

// JobAccess 定义 job 访问接口
//
//go:generate mockgen -source ../interfaces/job_access.go -destination ../interfaces/mock/mock_job_access.go
type JobAccess interface {
	CreateJob(ctx context.Context, tx *sql.Tx, jobInfo *JobInfo) error
	DeleteJobsByIDs(ctx context.Context, tx *sql.Tx, jobIDs []string) (int64, error)
	DeleteTasksByJobIDs(ctx context.Context, tx *sql.Tx, jobIDs []string) (int64, error)
	ListJobs(ctx context.Context, queryParams JobsQueryParams) ([]*JobInfo, error)
	GetJobsTotal(ctx context.Context, queryParams JobsQueryParams) (int64, error)
	ListTasks(ctx context.Context, queryParams TasksQueryParams) ([]*TaskInfo, error)
	GetTasksTotal(ctx context.Context, queryParams TasksQueryParams) (int64, error)

	CreateTasks(ctx context.Context, tx *sql.Tx, tasks map[string]*TaskInfo) error
	GetJobByID(ctx context.Context, jobID string) (*JobInfo, error)
	GetJobsByIDs(ctx context.Context, jobIDs []string) (map[string]*JobInfo, error)
	GetJobIDsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) ([]string, error)

	UpdateJobState(ctx context.Context, tx *sql.Tx, jobID string, stateInfo JobStateInfo) error
	UpdateTaskState(ctx context.Context, taskID string, stateInfo TaskStateInfo) error
}
