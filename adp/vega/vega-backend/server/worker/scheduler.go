// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/robfig/cron/v3"

	"vega-backend/common"
	"vega-backend/interfaces"
)

var (
	schedulerOnce sync.Once  // 确保 Scheduler 只被初始化一次
	scheduler     *Scheduler // 全局唯一的调度器实例
)

// Scheduler 管理定时发现任务的调度器
// 使用 cron 表达式来定义任务的执行时间
type Scheduler struct {
	appSetting *common.AppSetting                 // 应用配置
	cron       *cron.Cron                         // cron 调度器实例
	dss        interfaces.DiscoverScheduleService // 定时发现任务服务

	taskEntries      map[string]cron.EntryID // 任务ID到cron条目ID的映射
	taskEntriesMutex sync.RWMutex            // 保护taskEntries的读写锁

	ctx    context.Context    // 上下文
	cancel context.CancelFunc // 取消函数
}

// NewScheduler 创建或返回单例 Scheduler
// 使用 sync.Once 确保只创建一个实例
// 参数:
//   - appSetting: 应用配置
//   - dss: 定时发现任务服务
//
// 返回:
//   - *Scheduler: 调度器实例
func NewScheduler(appSetting *common.AppSetting, dss interfaces.DiscoverScheduleService) *Scheduler {
	schedulerOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		scheduler = &Scheduler{
			appSetting: appSetting,
			cron:       cron.New(), // Support seconds in cron expression:cron.WithSeconds()
			dss:        dss,

			taskEntries: make(map[string]cron.EntryID),

			ctx:    ctx,
			cancel: cancel,
		}
	})
	return scheduler
}

// Start 启动调度器并安排所有已启用的任务
// 执行步骤:
//  1. 从数据库加载所有启用的定时任务
//  2. 为每个任务创建 cron 调度条目
//  3. 启动 cron 调度器
//
// 返回:
//   - error: 如果启动失败则返回错误
func (s *Scheduler) Start() error {
	logger.Info("Starting scheduler") // 记录调度器启动信息

	// 从数据库加载所有启用的任务
	tasks, err := s.dss.GetEnabledTasks(s.ctx)
	if err != nil {
		logger.Errorf("Failed to load enabled tasks: %v", err)
		return fmt.Errorf("failed to load enabled tasks: %w", err)
	}

	// 为每个启用的任务创建调度
	for _, task := range tasks {
		if err := s.scheduleTask(task); err != nil {
			logger.Errorf("Failed to schedule task %s: %v", task.ID, err)
		}
	}

	// 启动 cron 调度器
	s.cron.Start()
	logger.Info("Scheduler started")
	return nil
}

// Stop 停止调度器
// 执行步骤:
//  1. 取消上下文，停止所有正在执行的任务
//  2. 停止 cron 调度器
func (s *Scheduler) Stop() {
	logger.Info("Stopping scheduler")
	s.cancel()    // 取消上下文
	s.cron.Stop() // 停止 cron 调度器
	logger.Info("Scheduler stopped")
}

// Reload 重新加载所有启用的任务并重新调度
// 用于在任务配置变更后刷新调度器
// 执行步骤:
//  1. 移除所有现有的调度任务
//  2. 从数据库重新加载所有启用的任务
//  3. 为每个任务创建新的调度条目
//
// 返回:
//   - error: 如果重载失败则返回错误
func (s *Scheduler) Reload() error {
	logger.Info("Reloading scheduler")

	// 移除所有现有的调度任务
	s.taskEntriesMutex.Lock()
	for taskID, entryID := range s.taskEntries {
		s.cron.Remove(entryID)
		delete(s.taskEntries, taskID)
	}
	s.taskEntriesMutex.Unlock()

	// 从数据库重新加载所有启用的任务
	tasks, err := s.dss.GetEnabledTasks(s.ctx)
	if err != nil {
		logger.Errorf("Failed to load enabled tasks: %v", err)
		return fmt.Errorf("failed to load enabled tasks: %w", err)
	}

	// 为每个任务创建新的调度条目
	for _, task := range tasks {
		if err := s.scheduleTask(task); err != nil {
			logger.Errorf("Failed to schedule task %s: %v", task.ID, err)
		}
	}

	logger.Info("Scheduler reloaded")
	return nil
}

// scheduleTask 调度一个定时发现任务
// 该方法为指定的任务创建 cron 调度条目
// 参数:
//   - task: 指向 DiscoverSchedule 结构体的指针，包含要调度的任务信息
//
// 返回值:
//   - error: 如果任务调度成功则返回 nil，否则返回相应的错误信息
func (s *Scheduler) scheduleTask(task *interfaces.DiscoverSchedule) error {
	// 检查任务是否已经被调度
	s.taskEntriesMutex.RLock()
	if _, exists := s.taskEntries[task.ID]; exists {
		s.taskEntriesMutex.RUnlock()
		logger.Warnf("Task %s is already scheduled", task.ID)
		return nil
	}
	s.taskEntriesMutex.RUnlock()

	// 使用 cron 表达式添加任务到调度器
	// 当到达执行时间时，会调用 executeTask 方法执行任务
	entryID, err := s.cron.AddFunc(task.CronExpr, func() {
		s.executeTask(task)
	})
	if err != nil {
		logger.Errorf("Failed to add cron job for task %s: %v", task.ID, err)
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// 保存 cron 条目 ID，用于后续的取消调度操作
	s.taskEntriesMutex.Lock()
	s.taskEntries[task.ID] = entryID
	s.taskEntriesMutex.Unlock()

	logger.Infof("Scheduled task %s with cron expression: %s", task.ID, task.CronExpr)
	return nil
}

// unscheduleTask 取消单个任务的调度
// 该方法从 cron 调度器中移除指定的任务
// 参数:
//   - taskID: 要取消调度的任务 ID
//
// 返回值:
//   - error: 总是返回 nil
func (s *Scheduler) unscheduleTask(taskID string) error {
	s.taskEntriesMutex.Lock()
	defer s.taskEntriesMutex.Unlock()

	// 查找任务的 cron 条目 ID
	entryID, exists := s.taskEntries[taskID]
	if !exists {
		logger.Warnf("Task %s is not scheduled", taskID)
		return nil
	}

	// 从 cron 调度器中移除任务
	s.cron.Remove(entryID)
	delete(s.taskEntries, taskID)

	logger.Infof("Unscheduled task %s", taskID)
	return nil
}

// executeTask 执行定时发现任务
// 该方法在任务到达执行时间时被 cron 调度器调用
// 执行步骤:
//  1. 从数据库获取最新的任务状态
//  2. 检查任务是否仍然启用
//  3. 检查任务是否已过期
//  4. 执行任务
//
// 参数:
//   - task: 要执行的任务
func (s *Scheduler) executeTask(task *interfaces.DiscoverSchedule) {
	logger.Infof("Executing scheduled discover task: id=%s, catalog_id=%s", task.ID, task.CatalogID)

	// 从数据库获取最新的任务状态
	currentTask, err := s.dss.GetByID(s.ctx, task.ID)
	if err != nil {
		logger.Errorf("Failed to get task %s: %v", task.ID, err)
		o11y.Error(s.ctx, fmt.Sprintf("Failed to get task %s", task.ID))
		return
	}

	// 检查任务是否仍然启用
	if !currentTask.Enabled {
		logger.Warnf("Task %s is disabled, skipping execution", task.ID)

		// 从调度器中移除已禁用的任务
		if err := s.unscheduleTask(task.ID); err != nil {
			logger.Errorf("Failed to unschedule disabled task %s: %v", task.ID, err)
			o11y.Error(s.ctx, fmt.Sprintf("Failed to unschedule disabled task %s", task.ID))
			return
		}

		logger.Infof("Successfully unscheduled disabled task %s", task.ID)
		return
	}

	// 检查任务是否已过期
	if currentTask.EndTime > 0 && time.Now().UnixMilli() > currentTask.EndTime {
		logger.Warnf("Task %s has expired, disabling and unscheduling", task.ID)

		// 禁用过期任务
		if err := s.dss.Disable(s.ctx, task.ID); err != nil {
			logger.Errorf("Failed to disable expired task %s: %v", task.ID, err)
			o11y.Error(s.ctx, fmt.Sprintf("Failed to disable expired task %s", task.ID))
			return
		}

		// 从调度器中移除任务
		if err := s.unscheduleTask(task.ID); err != nil {
			logger.Errorf("Failed to unschedule expired task %s: %v", task.ID, err)
			o11y.Error(s.ctx, fmt.Sprintf("Failed to unschedule expired task %s", task.ID))
			return
		}

		logger.Infof("Successfully disabled and unscheduled expired task %s", task.ID)
		return
	}

	// 执行任务
	if err := s.dss.ExecuteTask(s.ctx, currentTask); err != nil {
		logger.Errorf("Failed to execute task %s: %v", task.ID, err)
		o11y.Error(s.ctx, fmt.Sprintf("Failed to execute task %s", task.ID))
		return
	}

	logger.Infof("Successfully executed scheduled discover task: id=%s", task.ID)
}

// ScheduleTask 调度一个任务
// 该方法在创建新任务或启用任务时被调用
// 参数:
//   - taskID: 要调度的任务 ID
//
// 返回值:
//   - error: 如果调度失败则返回错误
func (s *Scheduler) ScheduleTask(taskID string) error {
	// 从数据库获取任务信息
	task, err := s.dss.GetByID(s.ctx, taskID)
	if err != nil {
		logger.Errorf("Failed to get task %s: %v", taskID, err)
		return fmt.Errorf("failed to get task: %w", err)
	}

	// 如果任务未启用，则不进行调度
	if !task.Enabled {
		logger.Warnf("Task %s is disabled, not scheduling", taskID)
		return nil
	}

	// 调用内部方法进行调度
	return s.scheduleTask(task)
}

// UnscheduleTask 取消任务的调度
// 该方法在任务被禁用或删除时被调用
// 参数:
//   - taskID: 要取消调度的任务 ID
//
// 返回值:
//   - error: 如果取消失败则返回错误
func (s *Scheduler) UnscheduleTask(taskID string) error {
	return s.unscheduleTask(taskID)
}

// UpdateTask 更新定时任务的调度
// 该方法在任务配置变更时被调用
// 执行步骤:
//  1. 取消旧的调度
//  2. 从数据库获取更新后的任务
//  3. 如果任务启用，则重新调度
//
// 参数:
//   - taskID: 要更新的任务 ID
//
// 返回值:
//   - error: 如果更新失败则返回错误
func (s *Scheduler) UpdateTask(taskID string) error {
	// 取消旧的调度
	if err := s.unscheduleTask(taskID); err != nil {
		logger.Errorf("Failed to unschedule task %s: %v", taskID, err)
	}

	// 从数据库获取更新后的任务
	task, err := s.dss.GetByID(s.ctx, taskID)
	if err != nil {
		logger.Errorf("Failed to get task %s: %v", taskID, err)
		return fmt.Errorf("failed to get task: %w", err)
	}

	// 如果任务启用，则重新调度
	if task.Enabled {
		return s.scheduleTask(task)
	}

	return nil
}
