package bootstrap

import (
	"context"
	_ "embed"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type embeddedToolDependencyPackage struct {
	name string
	data []byte
}

type ToolDependencySync struct {
	logger              interfaces.Logger
	operatorIntegration interfaces.DrivenOperatorIntegration
	config              config.ToolDependencySyncConfig
	loadPackages        func() []embeddedToolDependencyPackage
	wait                func(ctx context.Context, d time.Duration) bool
}

var (
	toolDependencySyncOnce sync.Once
	toolDependencySync     *ToolDependencySync
)

//go:embed tool_dependencies/execution_factory_tools.adp
var executionFactoryToolsData []byte

//go:embed tool_dependencies/context_loader_toolset.adp
var contextLoaderToolsetData []byte

// NewToolDependencySync 创建 ToolDependencySync 实例
// 该实例会自动启动工具依赖同步任务
func NewToolDependencySync() *ToolDependencySync {
	toolDependencySyncOnce.Do(func() {
		cfg := config.NewConfigLoader()
		toolDependencySync = &ToolDependencySync{
			logger:              cfg.GetLogger(),
			operatorIntegration: drivenadapters.NewOperatorIntegrationClient(),
			config:              cfg.ToolDependencySync,
			loadPackages:        loadEmbeddedToolDependencyPackages,
			wait:                waitWithContext,
		}
	})
	return toolDependencySync
}

func (s *ToolDependencySync) Start(ctx context.Context) {
	if !s.config.Enabled {
		s.logger.WithContext(ctx).Info("[ToolDependencySync] disabled, skip startup sync")
		return
	}

	delay := s.initialRetryDelay()
	for {
		err := s.syncOnce(ctx)
		if err == nil {
			return
		}
		s.logger.WithContext(ctx).Warnf("[ToolDependencySync] sync failed, retry after %s, err: %v", delay.String(), err)
		if !s.wait(ctx, delay) {
			return
		}
		delay = s.nextRetryDelay(delay)
	}
}

func (s *ToolDependencySync) syncOnce(ctx context.Context) error {
	packages := s.loadPackages()
	for _, pkg := range packages {
		s.logger.WithContext(ctx).Infof("[ToolDependencySync] sync tool dependency package: %s", pkg.name)

		req := &interfaces.SyncToolDependencyPackageRequest{
			Mode:        "upsert",
			PackageData: pkg.data,
		}
		if err := s.operatorIntegration.SyncToolDependencyPackage(ctx, req); err != nil {
			return err
		}
		s.logger.WithContext(ctx).Infof("[ToolDependencySync] sync tool dependency package: %s completed", pkg.name)
	}
	return nil
}

func loadEmbeddedToolDependencyPackages() []embeddedToolDependencyPackage {
	return []embeddedToolDependencyPackage{
		{
			name: "tool_dependencies/execution_factory_tools.adp",
			data: executionFactoryToolsData,
		},
		{
			name: "tool_dependencies/context_loader_toolset.adp",
			data: contextLoaderToolsetData,
		},
	}
}

func (s *ToolDependencySync) initialRetryDelay() time.Duration {
	seconds := s.config.InitialRetryIntervalSeconds
	if seconds <= 0 {
		seconds = 5
	}
	return time.Duration(seconds) * time.Second
}

func (s *ToolDependencySync) nextRetryDelay(current time.Duration) time.Duration {
	maxDelay := time.Duration(s.config.MaxRetryIntervalSeconds) * time.Second
	if maxDelay <= 0 {
		maxDelay = 60 * time.Second
	}
	next := current * 2
	if next > maxDelay {
		return maxDelay
	}
	return next
}

func waitWithContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
