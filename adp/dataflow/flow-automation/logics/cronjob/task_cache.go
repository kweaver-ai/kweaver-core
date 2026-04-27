package cronjob

import (
	"context"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
)

type TaskCacheCronJob interface {
	DeleteExpiredTaskCache(ctx context.Context)
}

type taskCacheCronJob struct {
	og        drivenadapters.OssGateWay
	taskCache rds.TaskCache
}

var (
	taskCacheCronJobIns  TaskCacheCronJob
	taskCacheCronJobOnce sync.Once
)

func NewTaskCacheCronJob() TaskCacheCronJob {
	taskCacheCronJobOnce.Do(func() {
		taskCacheCronJobIns = &taskCacheCronJob{
			og:        drivenadapters.NewOssGateWay(),
			taskCache: rds.GetTaskCache(),
		}
	})
	return taskCacheCronJobIns
}

func (c *taskCacheCronJob) DeleteExpiredTaskCache(ctx context.Context) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	batchSize := int64(1000)
	maxRetries := 3
	timeout := time.Minute * 30

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	tableSuffixes := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
		"a", "b", "c", "d", "e", "f"}
	for _, suffix := range tableSuffixes {
		var minID uint64
		for {
			select {
			case <-ctx.Done():
				log.Warnf("[DeleteExpiredTaskCache] context canceled: %s", ctx.Err().Error())
				return
			default:
			}

			// 查询过期记录
			tasks, err := c.taskCache.ListTaskCache(ctx, rds.ListTaskCacheOptions{
				TableSuffix: suffix,
				Expired:     func() *bool { b := true; return &b }(),
				Limit:       batchSize,
				MinID:       minID,
			})
			if err != nil {
				log.Warnf("[DeleteExpiredTaskCache] ListTaskCache (table: %s) err: %s", suffix, err.Error())
				break
			}
			if len(tasks) == 0 {
				break
			}

			deletedHashes := make([]any, 0, len(tasks))
			failedHashes := make([]string, 0)
			for _, task := range tasks {
				// 只删除有 OSS 文件的
				if task.OssID != "" && task.OssKey != "" {
					var lastErr error
					// 删除 OSS 文件，重试机制
					for i := 0; i < maxRetries; i++ {
						err := c.og.DeleteFile(ctx, task.OssID, task.OssKey, true)
						if err == nil {
							deletedHashes = append(deletedHashes, task.Hash)
							break
						}
						lastErr = err
						time.Sleep(time.Second * time.Duration(i+1))
					}
					if lastErr != nil {
						log.Warnf("[DeleteExpiredTaskCache] delete OSS file failed after retries: hash=%s, err=%s", task.Hash, lastErr.Error())
						failedHashes = append(failedHashes, task.Hash)
					}
				} else {
					// 没有 oss 信息也计为直接可删
					deletedHashes = append(deletedHashes, task.Hash)
				}
			}

			// 只删除 OSS 已经成功清理的记录
			if len(deletedHashes) > 0 {
				for i := 0; i < maxRetries; i++ {
					err := c.taskCache.BatchDeleteByHash(ctx, deletedHashes)
					if err == nil {
						break
					}
					log.Warnf("[DeleteExpiredTaskCache] delete db record failed (retry %d), err: %s", i+1, err.Error())
					time.Sleep(time.Second * time.Duration(i+1))
				}
			}

			if len(failedHashes) > 0 {
				log.Warnf("[DeleteExpiredTaskCache] failed to delete some OSS files for hashes: %v", failedHashes)
			}

			minID = tasks[len(tasks)-1].ID
			if int64(len(tasks)) < batchSize {
				break
			}
		}
	}
}
