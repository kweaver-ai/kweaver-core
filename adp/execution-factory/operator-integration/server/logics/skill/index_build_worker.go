package skill

import (
	"context"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

type skillIndexBuildWorker struct {
	service *skillIndexBuildService
	stopCh  chan struct{}
	doneCh  chan struct{}
	once    sync.Once
}

var (
	skillIndexBuildWorkerOnce sync.Once
	skillIndexBuildWorkerInst *skillIndexBuildWorker
)

func NewSkillIndexBuildWorker() *skillIndexBuildWorker {
	skillIndexBuildWorkerOnce.Do(func() {
		svc, _ := NewSkillIndexBuildService().(*skillIndexBuildService)
		skillIndexBuildWorkerInst = &skillIndexBuildWorker{
			service: svc,
			stopCh:  make(chan struct{}),
			doneCh:  make(chan struct{}),
		}
	})
	return skillIndexBuildWorkerInst
}

func (w *skillIndexBuildWorker) Start() error {
	w.once.Do(func() {
		go w.loop()
	})
	return nil
}

func (w *skillIndexBuildWorker) Stop(ctx context.Context) {
	select {
	case <-w.stopCh:
	default:
		close(w.stopCh)
	}
	select {
	case <-w.doneCh:
	case <-ctx.Done():
	}
}

func (w *skillIndexBuildWorker) loop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer close(w.doneCh)

	_ = w.service.recoverStaleRunningTask(context.Background())
	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			_ = w.service.recoverStaleRunningTask(context.Background())
			_ = w.service.schedulePeriodicFullTask(context.Background())
			_ = w.service.cleanupExpiredFinishedTasks(context.Background())
			w.processPendingTask(context.Background())
		}
	}
}

func (w *skillIndexBuildWorker) processPendingTask(ctx context.Context) {
	if w.service == nil {
		return
	}
	task, err := w.service.taskRepo.SelectRunningTask(ctx, nil)
	if err != nil || task == nil {
		return
	}
	if interfaces.SkillIndexBuildStatus(task.Status) != interfaces.SkillIndexBuildStatusPending {
		return
	}
	ok, err := w.service.tryStartPendingTask(ctx, task.TaskID)
	if err != nil || !ok {
		return
	}
	_ = w.service.runTask(ctx, task.TaskID)
}
