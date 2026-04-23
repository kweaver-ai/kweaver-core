package bootstrap

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type ToolDependencySync struct {
	logger              interfaces.Logger
	operatorIntegration interfaces.DrivenOperatorIntegration
	config              config.ToolDependencySyncConfig
	readFile            func(path string) ([]byte, error)
	wait                func(ctx context.Context, d time.Duration) bool
}

var (
	toolDependencySyncOnce sync.Once
	toolDependencySync     *ToolDependencySync
)

const (
	toolDependencyPackagePath = "docs/release/tool-deps/execution_factory_tools.adp"
	toolDependencyVersionPath = "VERSION"
)

// NewToolDependencySync 创建 ToolDependencySync 实例
// 该实例会自动启动工具依赖同步任务
func NewToolDependencySync() *ToolDependencySync {
	toolDependencySyncOnce.Do(func() {
		cfg := config.NewConfigLoader()
		toolDependencySync = &ToolDependencySync{
			logger:              cfg.GetLogger(),
			operatorIntegration: drivenadapters.NewOperatorIntegrationClient(),
			config:              cfg.ToolDependencySync,
			readFile:            os.ReadFile,
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
		resp, err := s.syncOnce(ctx)
		if err == nil {
			if resp != nil {
				s.logger.WithContext(ctx).Infof("[ToolDependencySync] sync completed, status: %s, message: %s", resp.Status, resp.Message)
			}
			return
		}
		s.logger.WithContext(ctx).Warnf("[ToolDependencySync] sync failed, retry after %s, err: %v", delay.String(), err)
		if !s.wait(ctx, delay) {
			return
		}
		delay = s.nextRetryDelay(delay)
	}
}

func (s *ToolDependencySync) syncOnce(ctx context.Context) (*interfaces.SyncToolDependencyPackageResponse, error) {
	packageData, err := s.readFile(toolDependencyPackagePath)
	if err != nil {
		return nil, err
	}
	versionData, err := s.readFile(toolDependencyVersionPath)
	if err != nil {
		return nil, err
	}

	req := &interfaces.SyncToolDependencyPackageRequest{
		Mode:           "upsert",
		PackageVersion: strings.TrimSpace(string(versionData)),
		PackageData:    packageData,
	}
	return s.operatorIntegration.SyncToolDependencyPackage(ctx, req)
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
