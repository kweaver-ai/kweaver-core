package skill

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	infracommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

type skillIndexBuildWorker struct {
	server  *asynq.Server
	mux     *asynq.ServeMux
	logger  interfaces.Logger
	service *skillIndexBuildService
}

var (
	skillIndexBuildWorkerOnce sync.Once
	skillIndexBuildWorkerInst *skillIndexBuildWorker
)

func NewSkillIndexBuildWorker() *skillIndexBuildWorker {
	skillIndexBuildWorkerOnce.Do(func() {
		conf := config.NewConfigLoader()
		redisOpt, err := buildSkillIndexBuildRedisConnOpt()
		if err != nil {
			panic(err)
		}
		svc, _ := NewSkillIndexBuildService().(*skillIndexBuildService)
		mux := asynq.NewServeMux()
		worker := &skillIndexBuildWorker{
			server: asynq.NewServer(redisOpt, asynq.Config{
				Concurrency:     1,
				Queues:          map[string]int{skillIndexBuildQueueName: 1},
				BaseContext:     func() context.Context { return context.Background() },
				ShutdownTimeout: 30 * time.Second,
			}),
			mux:     mux,
			logger:  conf.GetLogger(),
			service: svc,
		}
		mux.HandleFunc(skillIndexBuildTaskType, worker.processTask)
		skillIndexBuildWorkerInst = worker
	})
	return skillIndexBuildWorkerInst
}

func (w *skillIndexBuildWorker) Start() error {
	return w.server.Start(w.mux)
}

func (w *skillIndexBuildWorker) Stop(ctx context.Context) {
	_ = ctx
	w.server.Shutdown()
}

func (w *skillIndexBuildWorker) processTask(ctx context.Context, task *asynq.Task) error {
	payload := &skillIndexBuildTaskPayload{}
	if err := json.Unmarshal(task.Payload(), payload); err != nil {
		return err
	}
	ctx = decorateSkillIndexBuildTaskContext(ctx, payload)
	if w.service != nil && payload.TaskID != "" {
		if taskDB, err := w.service.taskRepo.SelectByTaskID(ctx, nil, payload.TaskID); err == nil && taskDB != nil {
			if retried, ok := asynq.GetRetryCount(ctx); ok {
				taskDB.RetryCount = int64(retried)
			}
			if maxRetry, ok := asynq.GetMaxRetry(ctx); ok {
				taskDB.MaxRetry = int64(maxRetry)
			}
			_ = w.service.taskRepo.UpdateByTaskID(ctx, nil, taskDB)
		}
	}
	return w.service.runTask(ctx, payload.TaskID)
}

func decorateSkillIndexBuildTaskContext(ctx context.Context, payload *skillIndexBuildTaskPayload) context.Context {
	if payload == nil {
		return ctx
	}
	if payload.BusinessDomainID != "" {
		ctx = infracommon.SetBusinessDomainToCtx(ctx, payload.BusinessDomainID)
	}
	if payload.AccountID != "" {
		ctx = infracommon.SetAccountAuthContextToCtx(ctx, &interfaces.AccountAuthContext{
			AccountID:   payload.AccountID,
			AccountType: payload.AccountType,
		})
	}
	return ctx
}
