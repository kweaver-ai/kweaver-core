// Package cronjob 定时任务
package cronjob

import (
	"context"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/robfig/cron/v3"
)

const restartInternal = 3 * time.Second

// CronJob 定时任务接口
type CronJob interface {
	Start()
	CronOnMaster(ctx context.Context)
	CronDeleteRemovedDagInstanceExtData(ctx context.Context)
	CronDeleteExpiredTaskCache(ctx context.Context)
}

type cronJob struct {
	config           *common.Config
	store            mod.Store
	dumpLog          DumpLog
	extDataCronJob   ExtDataCronJob
	taskCacheCronJob TaskCacheCronJob
}

var (
	cOnce sync.Once
	c     CronJob
)

// NewCronJob new cron job instance
func NewCronJob() CronJob {
	cOnce.Do(func() {
		c = &cronJob{
			config:           common.NewConfig(),
			store:            mod.GetStore(),
			dumpLog:          NewDumpLog(),
			extDataCronJob:   NewExtDataCronJob(),
			taskCacheCronJob: NewTaskCacheCronJob(),
		}
	})

	return c
}

// Start start cron
func (c *cronJob) Start() {
	clog := commonLog.NewLogger()
	job := cron.New(cron.WithSeconds())
	// 每秒执行一次
	job.AddFunc("* * * * * *", func() { //nolint
		var err error
		msgs, err := c.store.ListInbox(context.TODO(), &mod.ListInboxInput{Now: time.Now().UTC().Unix(), Limit: 500})
		if err != nil {
			clog.Warnf("ListInbox err: %s", err.Error())
			return
		}
		ids := []string{}
		for _, msg := range msgs {
			ids = append(ids, msg.ID)
		}
		err = c.store.DeleteInbox(context.TODO(), ids)
		if err != nil {
			clog.Warnf("DeleteInbox err: %s", err.Error())
			return
		}
	})
	job.Start()

	defer job.Stop()

	select {}
}

// CronOnMaster 仅在主节点上执行的定时任务
func (c *cronJob) CronOnMaster(ctx context.Context) {
	go func(ctx context.Context) {
		clog := commonLog.NewLogger()
		clog.Infof("[CronOnMaster] history record save to oss thread start...")
		job := cron.New(cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))

		defer func() {
			clog.Errorf("[CronOnMaster] thread closed...")
			if rErr := recover(); rErr != nil {
				job.Stop()
				clog.Errorf("[CronOnMaster] panic occurred, detail: %v", rErr)
				time.Sleep(restartInternal)
				go c.CronOnMaster(ctx)
			}
		}()

		if _, err := job.AddFunc(c.config.DumpLog.CronJobExpression, func() {
			var err error
			newCtx, span := trace.StartInternalSpan(ctx)
			defer func() { trace.TelemetrySpanEnd(span, err) }()

			c.dumpLog.DumpLogToOSS(newCtx)
		}); err != nil {
			clog.Errorf("[CronOnMaster] add cron job failed, detail: %s", err.Error())
		}
		job.Start()

		defer job.Stop()

		<-ctx.Done()
	}(ctx)
}

func (c *cronJob) CronDeleteRemovedDagInstanceExtData(ctx context.Context) {
	go func(ctx context.Context) {
		clog := commonLog.NewLogger()
		clog.Infof("[CronDeleteRemovedDagInstanceExtData] start...")
		job := cron.New(cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))

		defer func() {
			clog.Errorf("[CronDeleteRemovedDagInstanceExtData] thread closed...")
			if rErr := recover(); rErr != nil {
				job.Stop()
				clog.Errorf("[CronDeleteRemovedDagInstanceExtData] panic occurred, detail: %v", rErr)
				time.Sleep(restartInternal)
				go c.CronDeleteRemovedDagInstanceExtData(ctx)
			}
		}()

		if _, err := job.AddFunc(c.config.Server.DeleteExtDataCron, func() {
			var err error
			newCtx, span := trace.StartInternalSpan(ctx)
			defer func() { trace.TelemetrySpanEnd(span, err) }()

			c.extDataCronJob.DeleteDagInstanceExtData(newCtx)
		}); err != nil {
			clog.Errorf("[CronDeleteRemovedDagInstanceExtData] add cron job failed, detail: %s", err.Error())
		}
		job.Start()

		defer job.Stop()

		<-ctx.Done()
	}(ctx)
}

func (c *cronJob) CronDeleteExpiredTaskCache(ctx context.Context) {
	go func(ctx context.Context) {
		clog := commonLog.NewLogger()
		clog.Infof("[CronDeleteExpiredTaskCache] start...")
		job := cron.New(cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))

		defer func() {
			clog.Errorf("[CronDeleteExpiredTaskCache] thread closed...")
			if rErr := recover(); rErr != nil {
				job.Stop()
				clog.Errorf("[CronDeleteExpiredTaskCache] panic occurred, detail: %v", rErr)
				time.Sleep(restartInternal)
				go c.CronDeleteRemovedDagInstanceExtData(ctx)
			}
		}()

		if _, err := job.AddFunc(c.config.Server.DeleteExpiredTaskCache, func() {
			var err error
			newCtx, span := trace.StartInternalSpan(ctx)
			defer func() { trace.TelemetrySpanEnd(span, err) }()

			c.taskCacheCronJob.DeleteExpiredTaskCache(newCtx)
		}); err != nil {
			clog.Errorf("[CronDeleteExpiredTaskCache] add cron job failed, detail: %s", err.Error())
		}
		job.Start()

		defer job.Stop()

		<-ctx.Done()
	}(ctx)
}
