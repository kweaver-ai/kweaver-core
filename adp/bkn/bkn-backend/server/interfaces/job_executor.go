// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// JobExecutor 定义 job 执行接口
//
//go:generate mockgen -source ../interfaces/job_executor.go -destination ../interfaces/mock/mock_job_executor.go
type JobExecutor interface {
	AddJob(ctx context.Context, jobInfo *JobInfo) error
	Start()
}
