// Package download_pool 文件下载线程池
package download_pool

import (
	"context"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
)

// Pool 下载线程池
type Pool interface {
	Start(ctx context.Context)
	Stop()
}

type pool struct {
	config     *Config
	ossGateway drivenadapters.OssGateWay
	mqHandler  mod.MQHandler
	workers    []*Worker
	wg         sync.WaitGroup
	cancel     context.CancelFunc
	log        traceLog.Logger
}

var (
	poolOnce sync.Once
	poolInst Pool
)

// NewPool 创建下载线程池
func NewPool() Pool {
	poolOnce.Do(func() {
		poolInst = &pool{
			config:     LoadConfig(),
			ossGateway: drivenadapters.NewOssGateWay(),
			mqHandler:  mod.NewMQHandler(),
			log:        traceLog.WithContext(context.Background()),
		}
	})
	return poolInst
}

// Start 启动线程池
func (p *pool) Start(ctx context.Context) {
	p.log.Infof("[DownloadPool] Starting with %d workers", p.config.WorkerCount)

	// 创建子 context 用于控制所有 worker
	workerCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	// 初始化 workers
	p.workers = make([]*Worker, p.config.WorkerCount)
	for i := 0; i < p.config.WorkerCount; i++ {
		p.workers[i] = NewWorker(i, p.config, p.ossGateway, p.mqHandler)
	}

	// 启动所有 workers
	for _, worker := range p.workers {
		p.wg.Add(1)
		go func(w *Worker) {
			defer p.wg.Done()
			w.Run(workerCtx)
		}(worker)
	}

	p.log.Infof("[DownloadPool] Started successfully")
}

// Stop 停止线程池
func (p *pool) Stop() {
	if p.cancel != nil {
		p.log.Infof("[DownloadPool] Stopping...")
		p.cancel()
		p.wg.Wait()
		p.log.Infof("[DownloadPool] Stopped")
	}
}

// GetPool 获取全局线程池实例
func GetPool() Pool {
	return poolInst
}
