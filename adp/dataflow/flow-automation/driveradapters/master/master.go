package master

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	lock "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/lock"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	store "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/store"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/alarm"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/cronjob"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/outbox"
)

const defaultInternal = 30 * time.Second

// Master interface
type Master interface {
	Run()
}

type master struct {
	log        commonLog.Logger
	cronJob    cronjob.CronJob
	outbox     outbox.OutBox
	alarm      alarm.Alarm
	lockClient lock.DistributeLock
	running    bool
}

var (
	oM sync.Once
	m  Master
)

// NewOnMaster 实例化master
func NewOnMaster() Master {
	oM.Do(func() {
		hostName, _ := os.Hostname()
		m = &master{
			log:        commonLog.NewLogger(),
			cronJob:    cronjob.NewCronJob(),
			outbox:     outbox.NewOutBox(),
			alarm:      alarm.NewAlarm(),
			lockClient: lock.NewDistributeLock(store.NewRedis(), "ONMASTER", hostName),
			running:    false,
		}
	})
	return m
}

func (m *master) run() {
	ctx := context.Background()
	// 不关闭channel，防止发生写close panic，采用定时3秒清空channel
	m.lockClient.ClearErrChannel()
	subCtx, cancle := context.WithCancel(ctx)
	err := m.lockClient.TryLock(subCtx, common.DumpLogLockTime, true)
	if err != nil {
		cancle()
		m.running = false
		return
	}
	m.log.Infoln("master thread running...")

	defer func() {
		if rErr := recover(); rErr != nil {
			cancle()
			m.running = false
		}
	}()

	m.running = true
	m.cronJob.CronOnMaster(subCtx)
	m.outbox.StartPushMessage(subCtx)
	m.alarm.ErrorAlarm(subCtx)
	m.cronJob.CronDeleteRemovedDagInstanceExtData(subCtx)
	m.cronJob.CronDeleteExpiredTaskCache(subCtx)

	<-m.lockClient.GetErrChannel()
	cancle()
	m.running = false
}

func (m *master) Run() {
	// 启动一个30s的定时器来监控线程
	timer := time.NewTimer(defaultInternal)
	m.run()

	for {
		<-timer.C
		if !m.running {
			m.log.Infoln("master thread not running, try to restart.")
			m.run()
		}
		timer.Reset(defaultInternal)
	}
}
