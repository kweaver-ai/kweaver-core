// 一个pipeline对应一个worker，一个worker对应一个task
// 一个worker一个消费者组，消费一个input_topic
// 可以通过扩展 pod 数与分区数相对应，提升消费速率
// 后续如果kafka支持单条提交，多个task可同时消费一个分区，一个worker会对应多个task
package logics

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/panjf2000/ants/v2"

	"flow-stream-data-pipeline/common"
	serrors "flow-stream-data-pipeline/errors"
	"flow-stream-data-pipeline/pipeline-worker/interfaces"
)

var (
	FlushBytes       int = 5 * 1024 * 1024
	FlushItems       int = 10000
	FlushInterval        = 5 * time.Second
	RetryInterval        = 3000 * time.Millisecond
	FailureThreshold     = 10
	PackagePoolSize  int = 20
	PackagePool, _       = ants.NewPool(PackagePoolSize, ants.WithPreAlloc(true), ants.WithNonblocking(false))
)

// 一个管道一个消费组
func ComsumerGroupID(tenant string, pipelineID string) string {
	return fmt.Sprintf("%s.sdp.%s", tenant, pipelineID)
}

type workerError struct {
	wID    string
	wError error
}

type WorkerService struct {
	pipelineID         string
	pipelineInfo       *interfaces.Pipeline          // manage the scrape configs ,key is pipelineID
	pipelineMgmtAccess interfaces.PipelineMgmtAccess // for communicating with server
	mqAccess           interfaces.MQAccess
	ims                interfaces.IndexBaseService
	worker             *Worker
	tasks              map[string]*Task
	triggerReload      chan struct{} // trigger for verifying if configs of workers have changed
	ctx                context.Context
	cancel             context.CancelFunc
	appSetting         *common.AppSetting
	errChan            chan workerError
}

func NewWorkerService(ctx context.Context, appSetting *common.AppSetting, pipelineID string) *WorkerService {
	FlushBytes = appSetting.ServerSetting.FlushMiB * 1024 * 1024
	FlushItems = appSetting.ServerSetting.FlushItems
	FlushInterval = time.Duration(appSetting.ServerSetting.FlushIntervalSec) * time.Second
	RetryInterval = time.Duration(appSetting.ServerSetting.RetryIntervalMs) * time.Millisecond
	FailureThreshold = appSetting.ServerSetting.FailureThreshold
	PackagePoolSize = appSetting.ServerSetting.PackagePoolSize
	PackagePool, _ = ants.NewPool(PackagePoolSize, ants.WithPreAlloc(true), ants.WithNonblocking(false))

	// 因为有pipelineID, 不使用 once.Do
	wService := &WorkerService{
		pipelineID:         pipelineID,
		pipelineMgmtAccess: PMAccess,
		mqAccess:           MQAccess,
		triggerReload:      make(chan struct{}, 5),
		ims:                NewIndexBaseService(appSetting),
		tasks:              map[string]*Task{},
		appSetting:         appSetting,
		errChan:            make(chan workerError, 100),
	}

	wService.ctx, wService.cancel = context.WithCancel(ctx)

	// 自动恢复管道任务 worker
	go wService.AutoRecoverWorker(ctx, pipelineID, appSetting.ServerSetting.WatchWorkersIntervalMin*time.Minute)

	// 开启一个协程监听 errChan，如果监听到错误，则停止运行此 worker，并更新状态为失败状态
	go wService.ListenToErrChan(ctx)

	return wService
}

func (wService *WorkerService) Start(ctx context.Context) {
	go func() {
		for {
			wService.longPolling(ctx)
		}
	}()

	wService.reloader(ctx)
}

func (wService *WorkerService) Stop(ctx context.Context) {
	logger.Info("worker stoping...")

	err := wService.pauseWorker(ctx, wService.pipelineID)
	if err != nil {
		logger.Infof("failed to stop worker, error: %s", err.Error())
	}

	logger.Info("worker cancel...")
	wService.cancel()
}

func (wService *WorkerService) longPolling(ctx context.Context) {
	pipelineInfo, exist, err := wService.pipelineMgmtAccess.GetConfigs(ctx, wService.pipelineID, true)
	if err != nil {
		return
	}

	if !exist {
		logger.Infof("pipeline '%s' is not exist in database", wService.pipelineID)
		err := wService.pauseWorker(ctx, wService.pipelineID)
		if err != nil {
			logger.Errorf("pipeline '%s' status is paused", wService.pipelineID)
		}

		return
	}

	// 抹平不需要的参数判断，unmarshal 时候的标签抹平多余的参数
	if reflect.DeepEqual(pipelineInfo, wService.pipelineInfo) {
		return
	}

	logger.Debugf("memory pipeline info: %s ", wService.pipelineInfo)
	// 将管道配置信息更新到内存中
	wService.pipelineInfo = pipelineInfo
	// 这一步是做啥的
	wService.triggerReload <- struct{}{}
}

func (wService *WorkerService) reloader(ctx context.Context) {
	for {
		select {
		case <-wService.triggerReload:
			wService.reload(ctx)
		case <-wService.ctx.Done():
			logger.Info("Worker server ctx done")
			return
		}
	}
}

func (wService *WorkerService) reload(ctx context.Context) {
	logger.Infof("reload worker: %s", wService.pipelineInfo)
	err := wService.reloadWorker(ctx, wService.pipelineInfo)
	if err != nil {
		logger.Errorf("failed to start worker '%s', error: %s", wService.pipelineInfo.PipelineName, err.Error())
	}
}

// ------------------ worker 定义

type Worker struct {
	*interfaces.Pipeline
	tasks []*Task
}

func (w *Worker) String() string {
	return fmt.Sprintf("{pipeline_id = %s, pipeline_name = %s,pipeline_builtin = %v, pipeline_output_type = %s, "+
		"pipeline_index_base = %s, pipeline_input_topic = %s, pipeline_output_topic = %s, pipeline_status = %s, "+
		"pipeline_error_topic = %s, pipeline_status_details = %s }",
		w.PipelineID, w.PipelineName, w.Builtin, w.OutputType, w.IndexBase, w.InputTopic, w.OutputTopic, w.PipelineStatus,
		w.ErrorTopic, w.PipelineStatusDetails)
}

// 新建一个worker
func newWorker(pipeline *interfaces.Pipeline) *Worker {
	worker := &Worker{
		Pipeline: pipeline,
		tasks:    make([]*Task, 0),
	}

	worker.PipelineStatus = interfaces.PipelineStatus_Running
	return worker
}

// 新建管道执行器
func (wService *WorkerService) startWorker(ctx context.Context, pipeline *interfaces.Pipeline) error {
	// 创建 worker 对象
	wService.worker = newWorker(pipeline)

	// 启动 tasks前的准备工作
	indexBaseInfo, err := wService.prepareWorker(ctx, wService.worker)
	if err != nil {
		logger.Errorf("failed to prepare worker %s, error: %v", pipeline.PipelineName, err)
		updateStatusErr := wService.updatePipelineStatus(ctx, wService.pipelineInfo, interfaces.PipelineStatus_Error, err.Error())
		if updateStatusErr != nil {
			return updateStatusErr
		}

		return err
	}

	// 启动 tasks
	wService.worker.StartTasks(ctx, wService.appSetting, indexBaseInfo, wService.errChan)

	return nil
}

// 更新管道
func (wService *WorkerService) updateWorker(ctx context.Context, pipeline *interfaces.Pipeline) error {
	// 更新内存里的pipeline信息
	wService.pipelineInfo = pipeline

	// 如果任务是暂停状态，只更新任务信息
	if pipeline.PipelineStatus == interfaces.PipelineStatus_Close {
		logger.Infof("the status of worker '%s' is paused", pipeline.PipelineName)
		return nil
	}

	// 如果任务是运行中或失败状态时，停止运行 tasks
	if wService.worker != nil {
		wService.worker.StopTasks()
	}

	// 创建 worker 对象
	wService.worker = newWorker(pipeline)

	indexBaseInfo, err := wService.prepareWorker(ctx, wService.worker)
	if err != nil {
		logger.Errorf("failed to prepare worker %s", pipeline.PipelineName)
		updateStatusErr := wService.updatePipelineStatus(ctx, pipeline, interfaces.PipelineStatus_Error, err.Error())
		if updateStatusErr != nil {
			logger.Errorf("failed to update pipeline status of worker '%s', error: %v", pipeline.PipelineName, err)
			return updateStatusErr
		}

		return err
	}

	nowPipelineStatus := interfaces.PipelineStatus_Running

	if wService.pipelineInfo.PipelineStatus != nowPipelineStatus {
		// 更改数据库的任务状态为running
		updateStatusErr := wService.updatePipelineStatus(ctx, wService.pipelineInfo, nowPipelineStatus, "")
		if updateStatusErr != nil {
			logger.Errorf("failed to update pipeline status of worker '%s' from pipeline-mgmt, error: %s",
				wService.pipelineInfo.PipelineName, updateStatusErr.Error())
			return updateStatusErr
		}
	}

	// 启动 tasks
	wService.worker.StartTasks(ctx, wService.appSetting, indexBaseInfo, wService.errChan)

	return nil
}

// 暂停任务
func (wService *WorkerService) pauseWorker(ctx context.Context, workerID string) error {
	logger.Infof("Check %s worker's status for pause...", workerID)
	isStatusChanged, err := wService.checkAndSetPipelineStatus(ctx, wService.pipelineInfo, interfaces.PipelineStatus_Closing)
	if err != nil {
		logger.Errorf("failed to check and set %s worker's status", workerID)
		return err
	}
	if !isStatusChanged {
		logger.Infof("worker %s is already paused", wService.pipelineInfo.PipelineName)
		return nil
	}

	// 停掉 tasks
	if wService.pipelineInfo.PipelineStatus != interfaces.PipelineStatus_Closing {
		if wService.worker != nil {
			wService.worker.StopTasks()
		}
	}

	// 更新任务信息
	wService.pipelineInfo.PipelineStatus = interfaces.PipelineStatus_Close

	return nil
}

// 1. 修改任务信息的时候，任务状态不变 2. 暂停重启任务时，任务状态变化
func (wService *WorkerService) reloadWorker(ctx context.Context, pipeline *interfaces.Pipeline) error {
	if wService.pipelineInfo == nil {
		err := wService.startWorker(ctx, pipeline)
		if err != nil {
			return err
		}

		return nil
	}

	// 更新任务的时候，任务状态不变
	switch pipeline.PipelineStatus {
	case wService.pipelineInfo.PipelineStatus:
		return wService.updateWorker(ctx, pipeline)
	case interfaces.PipelineStatus_Running:
		return wService.resumeWorker(ctx, wService.worker)
	case interfaces.PipelineStatus_Close:
		return wService.pauseWorker(ctx, wService.pipelineID)
	default:
		return nil
	}
}

// 开启任务
func (wService *WorkerService) resumeWorker(ctx context.Context, worker *Worker) error {
	logger.Infof("Check %s worker's status for resume...", wService.pipelineInfo.PipelineName)
	isStatusChanged, err := wService.checkAndSetPipelineStatus(ctx, wService.pipelineInfo, wService.pipelineInfo.PipelineStatus)
	if err != nil {
		logger.Errorf("Worker %s failed to check and set status", wService.pipelineInfo.PipelineName)
		return err
	}
	if !isStatusChanged {
		logger.Infof("Worker %s is already running or sleeping", wService.pipelineInfo.PipelineName)
		return nil
	}

	wService.worker = worker

	indexBaseInfo, err := wService.prepareWorker(ctx, worker)
	if err != nil {
		logger.Errorf("Worker %s prepare failed", wService.pipelineInfo.PipelineName)
		return err
	}

	worker.StartTasks(ctx, wService.appSetting, indexBaseInfo, wService.errChan)

	return nil
}

func (wService *WorkerService) prepareWorker(ctx context.Context, worker *Worker) (*interfaces.IndexBaseInfo, error) {
	// 获取索引库信息,如果索引库为空（mdl-model-persistence），则不获取
	var baseInfo = new(interfaces.IndexBaseInfo)
	var err error
	if worker.IndexBase != "" {
		baseInfo, err = wService.ims.GetIndexBaseByBaseType(ctx, worker.IndexBase)
		if err != nil {
			logger.Errorf("failed to get index base info for worker %s", worker.PipelineName)
			return nil, err
		}
	}

	return baseInfo, nil
}

// 先更新数据库，再更新内存
func (wService *WorkerService) updatePipelineStatus(ctx context.Context, pipeline *interfaces.Pipeline, status, details string) error {
	statusInfo := &interfaces.PipelineStatusInfo{
		Status:  status,
		Details: details,
	}
	err := wService.pipelineMgmtAccess.UpdatePipelineStatus(ctx, pipeline.PipelineID, statusInfo)
	if err != nil {
		logger.Errorf("Update worker %s's status in DB failed", pipeline.PipelineID)

		pipeline.PipelineStatus = interfaces.PipelineStatus_Error
		pipeline.PipelineStatusDetails = details
		return fmt.Errorf("failed to update the worker's status and details in DB, %w", err)
	}

	pipeline.PipelineStatus = status
	pipeline.PipelineStatusDetails = details
	logger.Infof("Worker '%s' has updated status to be %s in memory and DB", pipeline.PipelineName, status)

	return nil
}

// 检查状态，并在条件满足时设置状态。参数 expectedStatus 表示想要的状态。currentStatus 表示当前状态。
//
//		  检查规则：
//	       1. 想要的状态不能和当前状态一样。
//	       2. 当想要的状态为 paused, 当前状态只能为 running或者sleeping。
func (wService *WorkerService) checkAndSetPipelineStatus(ctx context.Context, rPipeline *interfaces.Pipeline,
	expectedStatus string) (isStatusChanged bool, err error) {
	currentStatus := rPipeline.PipelineStatus
	if expectedStatus == currentStatus {
		logger.Warnf("'%s' status is already expectedStatus: %s (error, running, closed, closing).", rPipeline.PipelineName, expectedStatus)
		return false, nil
	}

	if expectedStatus == interfaces.PipelineStatus_Closing && currentStatus != interfaces.PipelineStatus_Running {
		logger.Warnf("%s currentPipelineStatus is not running, cannot pause", rPipeline.PipelineName)
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			serrors.StreamDataPipeline_InternalError_PauseNotRunningPipelineFailed).WithErrorDetails("currentStatus is not running, cannot pause")
	}

	// 满足条件下设置状态
	rPipeline.PipelineStatus = expectedStatus
	return true, nil
}

// 每隔10分钟会自动恢复失败的任务和重启后内存丢失的任务，实现断电断网自动恢复
func (wService *WorkerService) AutoRecoverWorker(ctx context.Context, workerID string, interval time.Duration) {
	logger.Infof("Auto recover worker, worker id is '%s', interval: %s", workerID, interval)
	for {
		wService.recoverWorker(ctx, workerID)
		time.Sleep(interval)
	}
}

// 恢复当前任务
func (wService *WorkerService) recoverWorker(ctx context.Context, workerID string) {
	rPipeline, err := wService.syncMemoryPipelineBasedOnDB(ctx, workerID)
	if err != nil {
		return
	}

	if rPipeline == nil {
		return
	}

	logger.Infof("Recover worker %s", rPipeline)

	nowPipelineStatus := interfaces.PipelineStatus_Running

	// 设置 worker 状态为运行中
	updatePipelineStatusErr := wService.updatePipelineStatus(ctx, rPipeline, nowPipelineStatus, "")
	if updatePipelineStatusErr != nil {
		return
	}

	// 启动 worker
	err = wService.startWorker(ctx, rPipeline)
	if err != nil {
		return
	}
}

func (worker *Worker) StartTasks(ctx context.Context, appSetting *common.AppSetting,
	indexBaseInfo *interfaces.IndexBaseInfo, errChan chan workerError) {
	// 一个 task 来消费一个 input_topic
	go func() {
		task := NewTask(appSetting, worker.Pipeline, indexBaseInfo)
		worker.tasks = append(worker.tasks, task)

		defer func() {
			task.runningChannel <- false
		}()

		logger.Infof("Start run task %s", task)
		err := task.Run(ctx)
		if err != nil {
			errChan <- workerError{wID: worker.PipelineID, wError: err}
			logger.Info("An error has been sent to the error channel")
		}
	}()
}

func (worker *Worker) StopTasks() {
	var wg sync.WaitGroup
	for _, task := range worker.tasks {
		wg.Add(1)
		go func(task *Task) {
			defer wg.Done()
			err := task.Stop()
			if err != nil {
				return
			}
		}(task)
	}
	wg.Wait()

	// 删除所有 tasks,长度为0, 容量保持不变
	worker.tasks = worker.tasks[:0]
}

// 同步内存和flow-stream-data-pipeline的Pipeline信息
// 1. pipeline 在内存里, 但没在flow-stream-data-pipeline, 停掉它的 tasks，删除内存里的 rPipeline
// 2. pipeline 在内存和flow-stream-data-pipeline都有, 将失败的 pipeline 添加到 needRecoveredPipeline
// 3. pipeline 不在内存里, 但是在flow-stream-data-pipeline里, 把 rPipeline 添加到内存里, 将失败的和正在运行的 pipeline 添加到 needRecoveredPipeline
func (wService *WorkerService) syncMemoryPipelineBasedOnDB(ctx context.Context, pipelineID string) (*interfaces.Pipeline, error) {
	pipelineInfo, exist, err := wService.pipelineMgmtAccess.GetConfigs(ctx, pipelineID, false)
	if err != nil {
		logger.Errorf("Recover: failed to get pipeline info from pipeline-mgmt, %v", err)
		return nil, err
	}

	if exist && wService.pipelineInfo != nil {
		// pipeline-mgmt 存在，内存中也存在
		if wService.pipelineInfo.PipelineStatus == interfaces.PipelineStatus_Error {
			return wService.pipelineInfo, nil
		}
	} else if exist {
		// pipeline-mgmt 存在，内存不存在，则加到内存里
		wService.pipelineInfo = pipelineInfo

		// 如果 pipeline 状态为 close，则不恢复
		if pipelineInfo.PipelineStatus != interfaces.PipelineStatus_Close {
			return pipelineInfo, nil
		}
	} else if wService.pipelineInfo != nil {
		// pipeline-mgmt 不存在，内存中存在，停掉内存里的任务
		if wService.worker != nil {
			wService.worker.StopTasks()
		}

		wService.pipelineInfo = nil
		logger.Infof("Recover: pipeline-mgmt not exist, delete pipeline object %s in memory", wService.pipelineInfo.PipelineName)
	}

	return nil, nil
}

// 监听worker的errChan
func (wService *WorkerService) ListenToErrChan(ctx context.Context) {
	for workerErr := range wService.errChan {
		logger.Infof("Received error of worker '%s' from channel", workerErr.wID)
		go func(workerErr workerError) {
			// 停止 tasks，任务失败的时候不删除topic和消费者组
			if wService.worker != nil {
				wService.worker.StopTasks()
			}

			// 更新任务状态为失败
			if err := wService.updatePipelineStatus(ctx, wService.worker.Pipeline, interfaces.PipelineStatus_Error,
				workerErr.wError.Error()); err != nil {
				logger.Errorf("Update pipeline status to 'error' failed, %v", err)
				return
			}
		}(workerErr)
	}

	// 正常情况下errChan不应该被关闭
	logger.Errorf("Error channel closed, exiting loop")
}
