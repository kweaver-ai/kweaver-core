// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package worker

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	llq "github.com/emirpasic/gods/queues/linkedlistqueue"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"bkn-backend/common"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
)

var (
	jExecutorOnce sync.Once
	jExecutor     *jobExecutor
)

type jobExecutor struct {
	appSetting *common.AppSetting
	db         *sql.DB
	ja         interfaces.JobAccess
	ota        interfaces.ObjectTypeAccess
	rta        interfaces.RelationTypeAccess

	ReloadJobEnabled   bool
	MaxConcurrentTasks int

	mJobs      map[string]*Job
	mJobLock   sync.Mutex
	mTaskQueue *llq.Queue

	mTaskCallbackChan chan Task
}

func NewJobExecutor(appSetting *common.AppSetting) interfaces.JobExecutor {
	jExecutorOnce.Do(func() {
		jExecutor = &jobExecutor{
			appSetting: appSetting,
			db:         logics.DB,
			ja:         logics.JA,
			ota:        logics.OTA,
			rta:        logics.RTA,

			ReloadJobEnabled:   appSetting.ServerSetting.ReloadJobEnabled,
			MaxConcurrentTasks: appSetting.ServerSetting.MaxConcurrentTasks,

			mJobs:      make(map[string]*Job),
			mJobLock:   sync.Mutex{},
			mTaskQueue: llq.New(),

			mTaskCallbackChan: make(chan Task, 100),
		}
	})
	return jExecutor
}

func (je *jobExecutor) Start() {
	logger.Info("jobExecutor Start")

	if je.ReloadJobEnabled {
		err := je.reloadJobs()
		if err != nil {
			logger.Fatalf("Failed to reload jobs: %v", err)
			return
		}
	}

	if je.MaxConcurrentTasks <= 0 {
		je.MaxConcurrentTasks = 1
	}
	for i := 0; i < je.MaxConcurrentTasks; i++ {
		go je.StartTaskWorker()
	}
	go je.StartTaskCallbackWorker()
}

func (je *jobExecutor) reloadJobs() error {
	logger.Info("jobExecutor reloadJobs")

	ctx := context.Background()

	taskList, err := je.ja.ListTasks(ctx, interfaces.TasksQueryParams{
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Sort:      "f_id",
			Direction: interfaces.DESC_DIRECTION,
		},
		State: []interfaces.TaskState{
			interfaces.TaskStateRunning,
		},
	})
	if err != nil {
		return err
	}

	for _, task := range taskList {
		// 所有子任务完成，父任务完成
		stateInfo := interfaces.TaskStateInfo{
			State:       interfaces.TaskStateCanceled,
			StateDetail: fmt.Sprintf("task '%s' has been canceled caused by restart service", task.ID),
		}

		err = je.ja.UpdateTaskState(ctx, task.ID, stateInfo)
		if err != nil {
			return err
		}
	}

	jobList, err := je.ja.ListJobs(ctx, interfaces.JobsQueryParams{
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Sort:      "f_id",
			Direction: interfaces.DESC_DIRECTION,
		},
		State: []interfaces.JobState{
			interfaces.JobStateRunning,
			interfaces.JobStatePending,
		},
	})
	if err != nil {
		return err
	}

	for _, job := range jobList {
		// 所有子任务完成，父任务完成
		finishTime := time.Now().UnixMilli()
		info := interfaces.JobStateInfo{
			State:       interfaces.JobStateCanceled,
			StateDetail: fmt.Sprintf("job '%s' has been canceled caused by restart service", job.Name),
			FinishTime:  finishTime,
			TimeCost:    finishTime - job.CreateTime,
		}

		err = je.ja.UpdateJobState(ctx, nil, job.ID, info)
		if err != nil {
			return err
		}
	}

	return nil
}

func (je *jobExecutor) AddJob(ctx context.Context, jobInfo *interfaces.JobInfo) error {
	logger.Infof("jobExecutor AddJob: %v", jobInfo)

	job := &Job{
		mJobInfo:     jobInfo,
		mTasks:       make(map[string]Task),
		mFinishCount: 0,
	}

	for _, taskInfo := range jobInfo.TaskInfos {
		switch taskInfo.ConceptType {
		case interfaces.MODULE_TYPE_OBJECT_TYPE:
			ot, err := je.ota.GetObjectTypeByID(ctx, nil, jobInfo.KNID, jobInfo.Branch, taskInfo.ConceptID)
			if err != nil {
				return err
			}

			ott := NewObjectTypeTask(je.appSetting, taskInfo, ot)
			job.mTasks[taskInfo.ID] = ott
		}
	}

	je.mJobLock.Lock()
	defer je.mJobLock.Unlock()

	je.mJobs[jobInfo.ID] = job
	for _, task := range job.mTasks {
		je.mTaskQueue.Enqueue(task)
	}

	stateInfo := interfaces.JobStateInfo{
		State: interfaces.JobStateRunning,
	}
	err := je.ja.UpdateJobState(ctx, nil, jobInfo.ID, stateInfo)
	if err != nil {
		return err
	}

	return nil
}

func (je *jobExecutor) StartTaskCallbackWorker() {
	logger.Info("jobExecutor StartTaskCallbackWorker")
	for {
		task := <-je.mTaskCallbackChan
		je.HandleTaskCallback(task)
	}
}

func (je *jobExecutor) HandleTaskCallback(task Task) {
	defer func() {
		if rerr := recover(); rerr != nil {
			logger.Errorf("[handleTaskCallback] Failed: %v", rerr)
			debug.PrintStack()
			return
		}
	}()

	logger.Infof("jobExecutor HandleTaskCallback: %v", task)

	je.mJobLock.Lock()
	defer je.mJobLock.Unlock()

	taskInfo := task.GetTaskInfo()

	job, ok := je.mJobs[taskInfo.JobID]
	if !ok {
		logger.Errorf("Failed to get job %s", taskInfo.JobID)
		return
	}

	ctx := context.WithValue(context.Background(), interfaces.ACCOUNT_INFO_KEY, job.mJobInfo.Creator)

	// 检查任务是否完成
	job.mFinishCount++
	if job.mFinishCount != len(job.mTasks) {
		return
	}

	// 所有子任务完成，父任务完成
	finishTime := time.Now().UnixMilli()
	info := interfaces.JobStateInfo{
		State:      interfaces.JobStateCompleted,
		FinishTime: finishTime,
		TimeCost:   finishTime - job.mJobInfo.CreateTime,
	}
	for _, task := range job.mTasks {
		taskInfo := task.GetTaskInfo()
		if taskInfo.State == interfaces.TaskStateFailed {
			info.State = interfaces.JobStateFailed
			info.StateDetail += fmt.Sprintf("%s: %s\n", taskInfo.Name, taskInfo.StateDetail)
		}
	}

	if info.State == interfaces.JobStateFailed {
		err := je.ja.UpdateJobState(ctx, nil, job.mJobInfo.ID, info)
		if err != nil {
			return
		}

		delete(je.mJobs, job.mJobInfo.ID)
		return
	}

	tx, err := je.db.Begin()
	if err != nil {
		logger.Errorf("Failed to begin transaction: %v", err)
		return
	}

	for _, task := range job.mTasks {
		taskInfo := task.GetTaskInfo()
		if taskInfo.ConceptType == interfaces.MODULE_TYPE_OBJECT_TYPE {
			ott := task.(*ObjectTypeTask)
			err = je.ota.UpdateObjectTypeStatus(ctx, tx, job.mJobInfo.KNID,
				job.mJobInfo.Branch, taskInfo.ConceptID, *ott.objectTypeStatus)
			if err != nil {
				return
			}
		}
	}

	err = je.ja.UpdateJobState(ctx, tx, job.mJobInfo.ID, info)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		logger.Errorf("Failed to commit transaction: %v", err)
		return
	}

	delete(je.mJobs, job.mJobInfo.ID)
}

func (je *jobExecutor) StartTaskWorker() {
	logger.Info("jobExecutor StartTaskWorker")
	for {
		je.mJobLock.Lock()
		queObj, ok := je.mTaskQueue.Dequeue()
		je.mJobLock.Unlock()
		if !ok {
			time.Sleep(time.Second)
			continue
		}

		ctx := context.Background()
		select {
		case <-ctx.Done(): // 监听取消信号
			logger.Errorf("Operation canceled: %v", ctx.Err())
			continue
		default:
			task := queObj.(Task)
			err := je.HandleTask(ctx, task)
			if err != nil {
				logger.Errorf("Failed to handle task %s: %v", task.GetTaskInfo().ID, err)
				continue
			}
		}
	}
}

func (je *jobExecutor) UpdateTaskStateFailed(ctx context.Context, taskInfo *interfaces.TaskInfo, err error) {
	logger.Infof("jobExecutor UpdateTaskStateFailed: %v, err: %v", taskInfo, err)

	finishTime := time.Now().UnixMilli()
	stateInfo := interfaces.TaskStateInfo{
		State:       interfaces.TaskStateFailed,
		StateDetail: err.Error(),
		FinishTime:  finishTime,
		TimeCost:    finishTime - taskInfo.StartTime,
	}

	err = je.ja.UpdateTaskState(ctx, taskInfo.ID, stateInfo)
	if err != nil {
		return
	}
	taskInfo.State = stateInfo.State
	taskInfo.StateDetail = stateInfo.StateDetail
	taskInfo.FinishTime = stateInfo.FinishTime
	taskInfo.TimeCost = stateInfo.TimeCost
}

func (je *jobExecutor) UpdateTaskStateCompleted(ctx context.Context, taskInfo *interfaces.TaskInfo) {
	logger.Infof("jobExecutor UpdateTaskStateCompleted: %v", taskInfo)

	finishTime := time.Now().UnixMilli()
	stateInfo := interfaces.TaskStateInfo{
		State:      interfaces.TaskStateCompleted,
		FinishTime: finishTime,
		TimeCost:   finishTime - taskInfo.StartTime,
	}

	err := je.ja.UpdateTaskState(ctx, taskInfo.ID, stateInfo)
	if err != nil {
		return
	}
	taskInfo.State = stateInfo.State
	taskInfo.FinishTime = stateInfo.FinishTime
	taskInfo.TimeCost = stateInfo.TimeCost
}

func (je *jobExecutor) HandleTask(ctx context.Context, task Task) (err error) {
	taskInfo := task.GetTaskInfo()

	defer func() {
		if rerr := recover(); rerr != nil {
			logger.Errorf("[handleTask] Failed: %v, task为: %v", rerr, taskInfo)
			debug.PrintStack()
			return
		}

		if err != nil {
			je.UpdateTaskStateFailed(ctx, taskInfo, err)
			je.mTaskCallbackChan <- task
		}
	}()

	logger.Infof("Start to execute task %s, concept type: %s, concept id: %s",
		taskInfo.ID, taskInfo.ConceptType, taskInfo.ConceptID)

	if taskInfo.State != interfaces.TaskStatePending {
		err = fmt.Errorf("task %s is not pending, current state: %s", taskInfo.ID, taskInfo.State)
		return err
	}

	job, ok := je.mJobs[taskInfo.JobID]
	if !ok {
		err = fmt.Errorf("failed to get job %s", taskInfo.JobID)
		return err
	}

	accountInfo := job.mJobInfo.Creator
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	startTime := time.Now().UnixMilli()
	stateInfo := interfaces.TaskStateInfo{
		State:     interfaces.TaskStateRunning,
		StartTime: startTime,
	}
	err = je.ja.UpdateTaskState(ctx, taskInfo.ID, stateInfo)
	if err != nil {
		return err
	}
	taskInfo.State = stateInfo.State
	taskInfo.StartTime = stateInfo.StartTime

	switch t := task.(type) {
	case *ObjectTypeTask:
		ott := t
		err = ott.HandleObjectTypeTask(ctx, job.mJobInfo, taskInfo, ott.objectType)
		if err != nil {
			logger.Error(err.Error())
			return err
		}

		je.UpdateTaskStateCompleted(ctx, taskInfo)
		je.mTaskCallbackChan <- task
	}

	return nil
}
