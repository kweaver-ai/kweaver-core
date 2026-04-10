// Package download_pool 文件下载线程池
package download_pool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/lock"
	libstore "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/store"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/store"
)

// FlowFileDownloadResult 下载结果消息体
type FlowFileDownloadResult struct {
	FileID     uint64                 `json:"file_id"`
	Status     string                 `json:"status"` // "success" or "failed"
	ErrorCode  string                 `json:"error_code,omitempty"`
	ErrorMsg   string                 `json:"error_msg,omitempty"`
	NodeOutput map[string]interface{} `json:"node_output,omitempty"` // 节点输出结果
	RetryCount int                    `json:"retry_count,omitempty"`
}

// Worker 下载任务执行器
type Worker struct {
	id             int
	config         *Config
	ossGateway     drivenadapters.OssGateWay
	mqHandler      mod.MQHandler
	flowFileDao    rds.FlowFileDao
	flowStorageDao rds.FlowStorageDao
	downloadJobDao rds.FlowFileDownloadJobDao
	taskResumeDao  rds.FlowTaskResumeDao
	log            traceLog.Logger
}

// NewWorker 创建 Worker
func NewWorker(id int, config *Config, ossGateway drivenadapters.OssGateWay, mqHandler mod.MQHandler) *Worker {
	return &Worker{
		id:             id,
		config:         config,
		ossGateway:     ossGateway,
		mqHandler:      mqHandler,
		flowFileDao:    rds.GetFlowFileDao(),
		flowStorageDao: rds.GetFlowStorageDao(),
		downloadJobDao: rds.GetFlowFileDownloadJobDao(),
		taskResumeDao:  rds.GetFlowTaskResumeDao(),
		log:            traceLog.WithContext(context.Background()),
	}
}

// Run 启动 Worker
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(w.config.PollInterval) * time.Second)
	defer ticker.Stop()

	w.log.Infof("[DownloadPool] Worker %d started", w.id)

	for {
		select {
		case <-ctx.Done():
			w.log.Infof("[DownloadPool] Worker %d stopped", w.id)
			return
		case <-ticker.C:
			w.processJobs(ctx)
		}
	}
}

// processJobs 处理一批下载任务
func (w *Worker) processJobs(ctx context.Context) {
	now := time.Now().Unix()

	// 1. 查询待处理任务
	pendingStatus := rds.FlowFileDownloadJobStatusPending
	jobs, err := w.downloadJobDao.List(ctx, &rds.FlowFileDownloadJobQueryOptions{
		Status:      &pendingStatus,
		RetryBefore: now,
		Limit:       w.config.BatchSize,
	})
	if err != nil {
		w.log.Warnf("[DownloadPool] Worker %d: query jobs failed: %s", w.id, err.Error())
		return
	}

	if len(jobs) == 0 {
		return
	}

	// 2. 处理每个任务
	for _, job := range jobs {
		if ctx.Err() != nil {
			return
		}
		w.processJob(ctx, job)
	}
}

// processJob 处理单个下载任务
func (w *Worker) processJob(ctx context.Context, job *rds.FlowFileDownloadJob) {
	// 1. Redis 分布式锁前置过滤（保护 ClaimJob 操作）
	lockKey := fmt.Sprintf("download_job:%d", job.ID)
	lockField := fmt.Sprintf("worker_%d", w.id)
	dlock := lock.NewDistributeLock(libstore.NewRedis(), lockKey, lockField)

	// 尝试获取锁，5秒足够完成 ClaimJob
	if err := dlock.TryLock(ctx, 5*time.Second, false); err != nil {
		// 锁被其他 worker 占用，跳过此任务
		return
	}

	now := time.Now().Unix()

	// 2. 乐观锁抢占任务
	claimed, err := w.downloadJobDao.ClaimJob(ctx, job.ID, now)

	// ClaimJob 完成后立即释放 Redis 锁
	// 因为状态已变为 running，其他 worker 不会再查询到此任务
	dlock.Release()

	if err != nil {
		w.log.Warnf("[DownloadPool] Worker %d: failed to claim job %d: %s", w.id, job.ID, err.Error())
		return
	}
	if !claimed {
		// 被其他 worker 抢走了
		return
	}

	// 3. 查询关联的 flow_file
	flowFile, err := w.flowFileDao.GetByID(ctx, job.FileID)
	if err != nil || flowFile == nil {
		w.markJobFailed(ctx, job, "FILE_NOT_FOUND", "flow_file not found")
		return
	}

	// 4. 执行下载
	w.log.Infof("[DownloadPool] Worker %d: downloading file %d from %s", w.id, job.FileID, job.DownloadURL)

	result := w.downloadAndUpload(ctx, job, flowFile)

	// 5. 发送结果消息到 MQ
	w.publishResult(ctx, result)
}

// downloadAndUpload 下载文件并上传到 OSS
func (w *Worker) downloadAndUpload(ctx context.Context, job *rds.FlowFileDownloadJob, flowFile *rds.FlowFile) *FlowFileDownloadResult {
	// 1. 创建 HTTP 客户端
	client := &http.Client{
		Timeout: time.Duration(w.config.DownloadTimeout) * time.Second,
	}

	// 2. 发起下载请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, job.DownloadURL, nil)
	if err != nil {
		return w.markJobFailed(ctx, job, "INVALID_URL", fmt.Sprintf("invalid download URL: %s", err.Error()))
	}

	resp, err := client.Do(req)
	if err != nil {
		return w.handleDownloadError(ctx, job, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return w.markJobFailed(ctx, job, "DOWNLOAD_FAILED", fmt.Sprintf("download failed with status: %d", resp.StatusCode))
	}

	// 3. 检查文件大小
	contentLength := resp.ContentLength
	if contentLength > w.config.MaxFileSize {
		return w.markJobFailed(ctx, job, "FILE_TOO_LARGE", fmt.Sprintf("file size %d exceeds limit %d", contentLength, w.config.MaxFileSize))
	}

	// 4. 获取可用 OSS
	ossID, err := w.ossGateway.GetAvaildOSS(ctx)
	if err != nil {
		return w.handleUploadError(ctx, job, err)
	}

	// 5. 生成对象存储路径
	objectKey, err := common.BuildFlowFileObjectKey(flowFile.ID, flowFile.Name)
	if err != nil {
		return w.markJobFailed(ctx, job, "INVALID_FILENAME", "failed to build object key")
	}

	// 6. 上传到 OSS
	// 先读取数据以获取实际大小（解决 Content-Length 未知的问题）
	data, err := io.ReadAll(io.LimitReader(resp.Body, w.config.MaxFileSize+1))
	if err != nil {
		return w.markJobFailed(ctx, job, "READ_ERROR", fmt.Sprintf("failed to read response body: %s", err.Error()))
	}

	actualSize := int64(len(data))
	if actualSize > w.config.MaxFileSize {
		return w.markJobFailed(ctx, job, "FILE_TOO_LARGE", fmt.Sprintf("file size %d exceeds limit %d", actualSize, w.config.MaxFileSize))
	}

	if err := w.ossGateway.UploadFile(ctx, ossID, objectKey, false, bytes.NewReader(data), actualSize); err != nil {
		return w.handleUploadError(ctx, job, err)
	}

	// 7. 创建 flow_storage
	now := time.Now().Unix()
	flowStorage := &rds.FlowStorage{
		ID:        store.NextID(),
		OssID:     ossID,
		ObjectKey: objectKey,
		Name:      flowFile.Name,
		Size:      uint64(actualSize),
		Status:    rds.FlowStorageStatusNormal,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := w.flowStorageDao.Insert(ctx, flowStorage); err != nil {
		w.log.Warnf("[DownloadPool] Worker %d: failed to create storage: %s", w.id, err.Error())
		// 尝试删除已上传的文件
		w.ossGateway.DeleteFile(ctx, ossID, objectKey, false)
		return w.handleUploadError(ctx, job, err)
	}

	// 8. 更新 flow_file
	readyStatus := rds.FlowFileStatusReady
	if err := w.flowFileDao.Update(ctx, flowFile.ID, &rds.FlowFileUpdateParams{
		StorageID: &flowStorage.ID,
		Status:    &readyStatus,
	}); err != nil {
		w.log.Warnf("[DownloadPool] Worker %d: failed to update flow_file: %s", w.id, err.Error())
		return w.markJobFailed(ctx, job, "DB_ERROR", "failed to update flow_file")
	}

	// 9. 更新下载任务状态为成功
	successStatus := rds.FlowFileDownloadJobStatusSuccess
	finishedAt := time.Now().Unix()
	w.downloadJobDao.Update(ctx, job.ID, &rds.FlowFileDownloadJobUpdateParams{
		Status:     &successStatus,
		FinishedAt: &finishedAt,
	})

	// 10. 构建节点输出结果
	dfsURI := common.BuildDFSURI(flowFile.ID)
	nodeOutput := map[string]interface{}{
		"id":         dfsURI,
		"docid":      dfsURI,
		"file_id":    fmt.Sprintf("%d", flowFile.ID),
		"storage_id": fmt.Sprintf("%d", flowStorage.ID),
		"name":       flowFile.Name,
		"size":       actualSize,
		"status":     "ready",
	}

	w.log.Infof("[DownloadPool] Worker %d: successfully downloaded file %d", w.id, job.FileID)

	return &FlowFileDownloadResult{
		FileID:     job.FileID,
		Status:     "success",
		NodeOutput: nodeOutput,
	}
}

// handleDownloadError 处理下载错误
func (w *Worker) handleDownloadError(ctx context.Context, job *rds.FlowFileDownloadJob, err error) *FlowFileDownloadResult {
	return w.markJobFailed(ctx, job, "DOWNLOAD_ERROR", err.Error())
}

// handleUploadError 处理上传错误
func (w *Worker) handleUploadError(ctx context.Context, job *rds.FlowFileDownloadJob, err error) *FlowFileDownloadResult {
	return w.markJobFailed(ctx, job, "UPLOAD_ERROR", err.Error())
}

// markJobFailed 标记任务失败
func (w *Worker) markJobFailed(ctx context.Context, job *rds.FlowFileDownloadJob, errorCode, errorMsg string) *FlowFileDownloadResult {
	now := time.Now().Unix()
	retryCount := job.RetryCount + 1

	// 检查是否达到最大重试次数
	if retryCount >= job.MaxRetry {
		// 标记为失败
		failedStatus := rds.FlowFileDownloadJobStatusFailed
		finishedAt := now
		w.downloadJobDao.Update(ctx, job.ID, &rds.FlowFileDownloadJobUpdateParams{
			Status:       &failedStatus,
			RetryCount:   &retryCount,
			ErrorCode:    &errorCode,
			ErrorMessage: &errorMsg,
			FinishedAt:   &finishedAt,
		})

		// 同时更新 flow_file 为 invalid
		invalidStatus := rds.FlowFileStatusInvalid
		w.flowFileDao.UpdateStatus(ctx, job.FileID, invalidStatus)

		w.log.Warnf("[DownloadPool] Worker %d: job %d failed after %d retries: %s - %s",
			w.id, job.ID, retryCount, errorCode, errorMsg)

		return &FlowFileDownloadResult{
			FileID:     job.FileID,
			Status:     "failed",
			ErrorCode:  errorCode,
			ErrorMsg:   errorMsg,
			RetryCount: retryCount,
		}
	}

	// 还可以重试
	pendingStatus := rds.FlowFileDownloadJobStatusPending
	nextRetryAt := now + 60 // 60秒后重试
	w.downloadJobDao.Update(ctx, job.ID, &rds.FlowFileDownloadJobUpdateParams{
		Status:       &pendingStatus,
		RetryCount:   &retryCount,
		NextRetryAt:  &nextRetryAt,
		ErrorCode:    &errorCode,
		ErrorMessage: &errorMsg,
	})

	w.log.Warnf("[DownloadPool] Worker %d: job %d failed (retry %d/%d): %s - %s",
		w.id, job.ID, retryCount, job.MaxRetry, errorCode, errorMsg)

	return nil // 不发送消息，等待重试
}

// publishResult 发送结果消息到 MQ
func (w *Worker) publishResult(ctx context.Context, result *FlowFileDownloadResult) {
	if result == nil {
		return
	}

	msg, err := json.Marshal(result)
	if err != nil {
		w.log.Warnf("[DownloadPool] Worker %d: failed to marshal result: %s", w.id, err.Error())
		return
	}

	if err := w.mqHandler.Publish(common.TopicFlowFileDownloadResult, msg); err != nil {
		w.log.Warnf("[DownloadPool] Worker %d: failed to publish result: %s", w.id, err.Error())
		return
	}

	w.log.Infof("[DownloadPool] Worker %d: published result for file %d, status: %s",
		w.id, result.FileID, result.Status)
}
