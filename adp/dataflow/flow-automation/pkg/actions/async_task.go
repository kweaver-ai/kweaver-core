package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	lock "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/lock"
	libstore "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/store"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/store"
)

const (
	asyncTaskLockPrefix  = "automation:async_task_cache"
	asyncTaskLockTTL     = 30 * time.Second
	asyncTaskLockWaitTTL = 60 * time.Second
)

// AsyncTaskResultNotification 异步任务结果通知消息统一结构
type AsyncTaskResultNotification struct {
	Hash        string         `json:"hash"`         // 任务哈希（用于查找关联的TaskInstance）
	TaskType    string         `json:"task_type"`    // 任务类型（如 "file_parse", "fulltext" 等）
	Status      string         `json:"status"`       // 任务状态：success, failed
	Result      map[string]any `json:"result"`       // 任务结果（Status=success时有值）
	Error       string         `json:"error"`        // 错误信息（Status=failed时有值）
	CompletedAt int64          `json:"completed_at"` // 完成时间戳（Unix秒）
}

// AsyncTaskExecutor 异步任务执行器接口
type AsyncTaskExecutor interface {
	// Execute 执行异步任务，返回结果和错误
	Execute(ctx context.Context) (map[string]any, error)
	// GetTaskType 获取任务类型，用于TaskCache.Type字段
	GetTaskType() string
	// GetHashContent 获取用于生成hash的输入字符串
	GetHashContent() string
	// GetExpireSeconds 获取任务过期时间（秒）
	GetExpireSeconds() int64
	// GetResultFileExt 获取结果文件扩展名
	GetResultFileExt() string
}

// TaskResultLoader 任务结果加载器接口
type TaskResultLoader interface {
	// LoadResult 从OSS加载任务结果
	LoadResult(ctx context.Context, task *rds.TaskCacheItem) (map[string]any, error)
}

// DefaultTaskResultLoader 默认的任务结果加载器，从OSS读取JSON格式的结果
type DefaultTaskResultLoader struct{}

func (l *DefaultTaskResultLoader) LoadResult(ctx context.Context, task *rds.TaskCacheItem) (map[string]any, error) {
	og := drivenadapters.NewOssGateWay()
	data, err := og.DownloadFile(ctx, task.OssID, task.OssKey, true)
	if err != nil {
		return nil, fmt.Errorf("load result from OSS failed: %w", err)
	}

	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal result failed: %w", err)
	}

	return result, nil
}

// AsyncTaskManager 异步任务缓存管理器
type AsyncTaskManager struct {
	taskCache      rds.TaskCache
	resultLoader   TaskResultLoader
	lockPrefix     string
	lockTTL        time.Duration
	lockWaitTTL    time.Duration
	executeMethods entity.ExecuteMethods
}

// NewAsyncTaskManager 创建异步任务缓存管理器（使用统一的 topic）
func NewAsyncTaskManager(executeMethods entity.ExecuteMethods) *AsyncTaskManager {
	return &AsyncTaskManager{
		taskCache:      rds.GetTaskCache(),
		resultLoader:   &DefaultTaskResultLoader{},
		lockPrefix:     asyncTaskLockPrefix,
		lockTTL:        asyncTaskLockTTL,
		lockWaitTTL:    asyncTaskLockWaitTTL,
		executeMethods: executeMethods,
	}
}

// WithResultLoader 设置自定义结果加载器
func (m *AsyncTaskManager) WithResultLoader(loader TaskResultLoader) *AsyncTaskManager {
	m.resultLoader = loader
	return m
}

// WithLockPrefix 设置锁前缀
func (m *AsyncTaskManager) WithLockPrefix(prefix string) *AsyncTaskManager {
	m.lockPrefix = prefix
	return m
}

// Run 执行异步任务，处理缓存逻辑
func (m *AsyncTaskManager) Run(ctx entity.ExecuteContext, executor AsyncTaskExecutor) (any, error) {
	log := traceLog.WithContext(ctx.Context())
	taskIns := ctx.GetTaskInstance()

	// 生成hash并更新到TaskInstance
	hashContent := executor.GetHashContent()
	taskIns.Hash = hash(hashContent)

	err := taskIns.Patch(ctx.Context(), &entity.TaskInstance{
		BaseInfo: taskIns.BaseInfo,
		Hash:     taskIns.Hash,
		Results:  map[string]any{},
	})
	if err != nil {
		log.Warnf("[AsyncTaskManager] Patch hash err: %s, taskInsID: %s, hash: %s", err.Error(), taskIns.ID, taskIns.Hash)
		return nil, err
	}

	// 查询TaskCache
	task, err := m.taskCache.GetByHash(ctx.Context(), taskIns.Hash)
	if err != nil {
		log.Warnf("[AsyncTaskManager] GetByHash err: %s, hash: %s", err.Error(), taskIns.Hash)
		return nil, err
	}

	// 如果任务不存在，创建新任务并异步执行
	if task == nil {
		err = m.withTaskLock(ctx.Context(), taskIns.Hash, taskIns.ID, func(lockCtx context.Context) error {
			task, err = m.taskCache.GetByHash(lockCtx, taskIns.Hash)
			if err != nil {
				log.Warnf("[AsyncTaskManager] GetByHash with lock err: %s, hash: %s", err.Error(), taskIns.Hash)
				return err
			}

			if task != nil {
				return nil
			}

			// 创建新任务
			og := drivenadapters.NewOssGateWay()
			ossID, err := og.GetAvaildOSS(lockCtx)
			if err != nil {
				log.Warnf("[AsyncTaskManager] GetAvaildOSS err: %s", err.Error())
				return err
			}

			config := common.NewConfig()
			ossKey := fmt.Sprintf(`%s/task_results/%s%s`, config.Server.StoragePrefix, taskIns.Hash, executor.GetResultFileExt())
			now := time.Now().Unix()

			task = &rds.TaskCacheItem{
				ID:         store.NextID(),
				Hash:       taskIns.Hash,
				Type:       executor.GetTaskType(),
				Status:     rds.TaskStatusPending,
				OssID:      ossID,
				OssKey:     ossKey,
				Ext:        executor.GetResultFileExt(),
				Size:       0,
				ErrMsg:     "",
				CreateTime: now,
				ModifyTime: now,
				ExpireTime: now + executor.GetExpireSeconds(),
			}

			err = m.taskCache.Insert(lockCtx, task)
			if err != nil {
				log.Warnf("[AsyncTaskManager] Insert err: %s, hash: %s", err.Error(), taskIns.Hash)
				return err
			}

			go m.executeAsync(context.Background(), executor, task)

			return nil
		})

		if err != nil {
			return nil, err
		}

		// 重新查询任务
		if task == nil {
			task, err = m.taskCache.GetByHash(ctx.Context(), taskIns.Hash)
			if err != nil {
				log.Warnf("[AsyncTaskManager] GetByHash after lock err: %s, hash: %s", err.Error(), taskIns.Hash)
				return nil, err
			}
		}
	}

	// 根据任务状态返回结果
	result := map[string]any{}
	switch task.Status {
	case rds.TaskStatusFailed:
		ctx.ShareData().Set("__status_"+taskIns.ID, entity.TaskInstanceStatusFailed)
		return nil, fmt.Errorf("%v", task.ErrMsg)
	case rds.TaskStatusPending:
		ctx.ShareData().Set("__status_"+taskIns.ID, entity.TaskInstanceStatusBlocked)
		return result, nil
	default:
		// 从OSS加载结果
		result, err = m.resultLoader.LoadResult(ctx.Context(), task)
		if err != nil {
			log.Warnf("[AsyncTaskManager] LoadResult err: %s, hash: %s", err.Error(), taskIns.Hash)
			return nil, err
		}

		ctx.ShareData().Set(ctx.GetTaskID(), result)
	}

	return result, nil
}

func (m *AsyncTaskManager) RunAfter(ctx entity.ExecuteContext) (entity.TaskInstanceStatus, error) {
	taskIns := ctx.GetTaskInstance()
	status, ok := ctx.ShareData().Get("__status_" + taskIns.ID)

	if ok && status == entity.TaskInstanceStatusBlocked {
		return entity.TaskInstanceStatusBlocked, nil
	}

	return entity.TaskInstanceStatusSuccess, nil
}

// executeAsync 异步执行任务
func (m *AsyncTaskManager) executeAsync(ctx context.Context, executor AsyncTaskExecutor, task *rds.TaskCacheItem) {
	var err error
	taskType := executor.GetTaskType()
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() {
		trace.TelemetrySpanEnd(span, err)
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			traceLog.WithContext(ctx).Errorf("[AsyncTaskManager.executeAsync] panic: %v, taskType: %s", r, taskType)
			m.deleteTaskCache(ctx, task.Hash)
			m.notifyResult(ctx, task.Hash, taskType, nil, err)
		}
	}()
	log := traceLog.WithContext(ctx)

	// 执行任务
	result, err := executor.Execute(ctx)
	if err != nil {
		log.Warnf("[AsyncTaskManager.executeAsync] Execute err: %s, hash: %s, taskType: %s", err.Error(), task.Hash, taskType)
		m.deleteTaskCache(ctx, task.Hash)
		m.notifyResult(ctx, task.Hash, taskType, nil, err)
		return
	}

	// 将结果存储到OSS
	resultBytes, err := json.Marshal(result)
	if err != nil {
		log.Warnf("[AsyncTaskManager.executeAsync] Marshal result err: %v, taskType: %s", err, taskType)
		m.deleteTaskCache(ctx, task.Hash)
		m.notifyResult(ctx, task.Hash, taskType, nil, err)
		return
	}

	og := drivenadapters.NewOssGateWay()
	err = og.UploadFile(ctx, task.OssID, task.OssKey, true, bytes.NewReader(resultBytes), int64(len(resultBytes)))
	if err != nil {
		log.Warnf("[AsyncTaskManager.executeAsync] UploadFile err: %s, taskType: %s", err.Error(), taskType)
		m.deleteTaskCache(ctx, task.Hash)
		m.notifyResult(ctx, task.Hash, taskType, nil, err)
		return
	}

	// 更新TaskCache为成功状态
	m.updateTaskStatus(ctx, task.Hash, rds.TaskStatusSuccess, "", int64(len(resultBytes)))

	// 发送MQ消息通知完成
	log.Infof("[AsyncTaskManager.executeAsync] Task completed successfully, hash: %s, taskType: %s, size: %d bytes", task.Hash, taskType, len(resultBytes))
	m.notifyResult(ctx, task.Hash, taskType, result, nil)
}

func (m *AsyncTaskManager) deleteTaskCache(ctx context.Context, hash string) {
	err := m.taskCache.DeleteByHash(ctx, hash)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[AsyncTaskManager.deleteTaskCache] Delete err: %s", err.Error())
	}
}

// updateTaskStatus 更新任务状态
func (m *AsyncTaskManager) updateTaskStatus(ctx context.Context, hash string, status rds.TaskStatus, errMsg string, size int64) {
	update := &rds.TaskCacheItem{
		Hash:       hash,
		Status:     status,
		ErrMsg:     errMsg,
		Size:       size,
		ModifyTime: time.Now().Unix(),
	}
	err := m.taskCache.Update(ctx, update)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[AsyncTaskManager.updateTaskStatus] Update err: %s", err.Error())
		// 更新失败删除缓存
		m.deleteTaskCache(ctx, hash)
	}
}

// notifyResult 发送MQ消息通知任务完成（使用统一的消息结构和 topic）
func (m *AsyncTaskManager) notifyResult(ctx context.Context, hash, taskType string, result map[string]any, err error) {
	PublishAsyncTaskResult(ctx, m.executeMethods, hash, taskType, result, err)
}

// PublishAsyncTaskResult 统一发布异步任务结果通知
func PublishAsyncTaskResult(ctx context.Context, executeMethods entity.ExecuteMethods, hash, taskType string, result map[string]any, err error) {
	log := traceLog.WithContext(ctx)

	// 构建统一的通知消息
	notification := AsyncTaskResultNotification{
		Hash:        hash,
		TaskType:    taskType,
		CompletedAt: time.Now().Unix(),
	}

	if err != nil {
		notification.Status = "failed"
		notification.Error = err.Error()
	} else {
		notification.Status = "success"
	}

	msgBytes, marshalErr := json.Marshal(notification)
	if marshalErr != nil {
		log.Warnf("[PublishAsyncTaskResult] Marshal msg err: %v", marshalErr)
		return
	}

	publishErr := executeMethods.Publish(common.TopicAsyncTaskResult, msgBytes)
	if publishErr != nil {
		log.Warnf("[PublishAsyncTaskResult] Publish err: %v, topic: %s", publishErr, common.TopicAsyncTaskResult)
	} else {
		log.Infof("[PublishAsyncTaskResult] Published result notification, hash: %s, taskType: %s, status: %s", hash, taskType, notification.Status)
	}
}

// withTaskLock 使用分布式锁保护任务创建
func (m *AsyncTaskManager) withTaskLock(ctx context.Context, hash, owner string, fn func(context.Context) error) error {
	lockClient := lock.NewDistributeLock(libstore.NewRedis(), fmt.Sprintf("%s:%s", m.lockPrefix, hash), owner)
	lockCtx, cancel := context.WithTimeout(ctx, m.lockWaitTTL)
	defer cancel()

	err := lockClient.TryLock(lockCtx, m.lockTTL, false)
	if err != nil {
		cachedTask, err := m.taskCache.GetByHash(ctx, hash)
		if err == nil && cachedTask != nil {
			return nil
		}
		return fmt.Errorf("acquire task cache lock timeout, hash: %s", hash)
	}
	defer lockClient.Release()

	return fn(lockCtx)
}
