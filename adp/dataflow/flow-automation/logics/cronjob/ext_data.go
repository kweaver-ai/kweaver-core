package cronjob

import (
	"context"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
)

type ExtDataCronJob interface {
	DeleteDagInstanceExtData(ctx context.Context)
}

type extDataCronJob struct {
	og         drivenadapters.OssGateWay
	extDataDao rds.DagInstanceExtDataDao
}

var (
	extDataCronJobIns  ExtDataCronJob
	extDataCronJobOnce sync.Once
)

func NewExtDataCronJob() ExtDataCronJob {
	extDataCronJobOnce.Do(func() {
		extDataCronJobIns = &extDataCronJob{
			og:         drivenadapters.NewOssGateWay(),
			extDataDao: rds.GetDagInstanceExtDataDao(),
		}
	})

	return extDataCronJobIns
}

func (c *extDataCronJob) DeleteDagInstanceExtData(ctx context.Context) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	batchSize := 1000
	minID := ""
	maxRetries := 3
	timeout := time.Minute * 30

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Warnf("[DeleteDagInstanceExtData] context canceled: %s", ctx.Err().Error())
			return
		default:
		}

		// 分批查询数据
		items, err := c.extDataDao.List(ctx, &rds.ExtDataQueryOptions{
			Removed:     true,
			Limit:       batchSize,
			MinID:       minID,
			SelectField: []string{"f_id", "f_oss_id", "f_oss_key"},
		})

		if err != nil {
			log.Warnf("[DeleteDagInstanceExtData] list ext data err: %s", err.Error())
			break
		}

		if len(items) == 0 {
			break
		}

		deletedIds := make([]string, 0)
		failedItems := make([]string, 0)

		for _, item := range items {
			var lastErr error
			// 重试删除OSS文件
			for i := 0; i < maxRetries; i++ {
				err := c.og.DeleteFile(ctx, item.OssID, item.OssKey, true)
				if err == nil {
					deletedIds = append(deletedIds, item.ID)
					break
				}
				lastErr = err
				time.Sleep(time.Second * time.Duration(i+1))
			}

			if lastErr != nil {
				log.Warnf("[DeleteDagInstanceExtData] delete file failed after retries, err: %s", lastErr.Error())
				failedItems = append(failedItems, item.ID)
				continue
			}
		}

		// 只删除OSS和数据库都成功的记录
		if len(deletedIds) > 0 {
			for i := 0; i < maxRetries; i++ {
				err := c.extDataDao.Delete(ctx, &rds.ExtDataQueryOptions{IDs: deletedIds})
				if err == nil {
					break
				}
				log.Warnf("[DeleteDagInstanceExtData] delete db record failed (retry %d), err: %s", i+1, err.Error())
				time.Sleep(time.Second * time.Duration(i+1))
			}
		}

		// 记录失败项
		if len(failedItems) > 0 {
			log.Warnf("[DeleteDagInstanceExtData] failed to delete %d items: %v", len(failedItems), failedItems)
		}

		minID = items[len(items)-1].ID

		if len(items) < batchSize {
			break
		}
	}
}
