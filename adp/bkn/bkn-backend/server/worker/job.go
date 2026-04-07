// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package worker

import (
	"bkn-backend/interfaces"
)

type Job struct {
	mJobInfo     *interfaces.JobInfo
	mTasks       map[string]Task
	mFinishCount int
}

func (j *Job) GetJobInfo() *interfaces.JobInfo {
	return j.mJobInfo
}

type Task interface {
	GetTaskInfo() *interfaces.TaskInfo
}
