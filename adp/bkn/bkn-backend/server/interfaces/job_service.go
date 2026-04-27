// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

// JobService 定义 job 服务接口
//
//go:generate mockgen -source ../interfaces/job_service.go -destination ../interfaces/mock/mock_job_service.go
type JobService interface {
	CreateJob(ctx context.Context, jobInfo *JobInfo) (string, error)
	DeleteJobsByIDs(ctx context.Context, knID string, branch string, jobIDs []string) error
	ListJobs(ctx context.Context, queryParams JobsQueryParams) ([]*JobInfo, int64, error)
	ListTasks(ctx context.Context, queryParams TasksQueryParams) ([]*TaskInfo, int64, error)

	// 内部接口，不鉴权
	GetJobByID(ctx context.Context, jobID string) (*JobInfo, error)
	GetJobsByIDs(ctx context.Context, jobIDs []string) (map[string]*JobInfo, error)
	DeleteJobsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error
}
