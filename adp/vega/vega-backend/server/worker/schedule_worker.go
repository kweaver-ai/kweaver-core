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
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/robfig/cron/v3"

	"vega-backend/common"
	"vega-backend/interfaces"
)

var (
	scheduleWorkerOnce sync.Once       // 确保 ScheduleWorker 只被初始化一次
	scheduleWorker     *ScheduleWorker // 全局唯一的调度器实例
)

// ScheduleWorker 管理定时发现任务的调度器
// 使用 cron 表达式来定义任务的执行时间
type ScheduleWorker struct {
	appSetting *common.AppSetting                 // 应用配置
	cron       *cron.Cron                         // cron 调度器实例
	dss        interfaces.DiscoverScheduleService // 定时发现任务服务

	scheduleEntries      map[string]cron.EntryID // 任务ID到cron条目ID的映射
	scheduleEntriesMutex sync.RWMutex            // 保护scheduleEntries的读写锁

	ctx    context.Context    // 上下文
	cancel context.CancelFunc // 取消函数
}

// NewScheduleWorker 创建或返回单例 ScheduleWorker
// 使用 sync.Once 确保只创建一个实例
// 参数:
//   - appSetting: 应用配置
//   - dss: 定时发现任务服务
//
// 返回:
//   - *ScheduleWorker: 调度器实例
func NewScheduleWorker(appSetting *common.AppSetting, dss interfaces.DiscoverScheduleService) *ScheduleWorker {
	scheduleWorkerOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		scheduleWorker = &ScheduleWorker{
			appSetting: appSetting,
			cron:       cron.New(), // Support seconds in cron expression:cron.WithSeconds()
			dss:        dss,

			scheduleEntries: make(map[string]cron.EntryID),

			ctx:    ctx,
			cancel: cancel,
		}
	})
	return scheduleWorker
}

// Start 启动调度器并安排所有已启用的任务
// 执行步骤:
//  1. 从数据库加载所有启用的定时任务
//  2. 为每个任务创建 cron 调度条目
//  3. 启动 cron 调度器
//
// 返回:
//   - error: 如果启动失败则返回错误
func (sw *ScheduleWorker) Start() error {
	logger.Info("Starting schedule worker") // 记录调度器启动信息

	// 从数据库加载所有启用的任务
	schedules, err := sw.dss.GetEnabledSchedules(sw.ctx)
	if err != nil {
		logger.Errorf("Failed to load enabled tasks: %v", err)
		return fmt.Errorf("failed to load enabled tasks: %w", err)
	}

	// 为每个启用的任务创建调度
	for _, schedule := range schedules {
		if err := sw.schedule(schedule); err != nil {
			logger.Errorf("Failed to schedule schedule %s: %v", schedule.ID, err)
		}
	}

	// 启动 cron 调度器
	sw.cron.Start()
	logger.Info("Schedule worker started")
	return nil
}

// Stop 停止调度器
// 执行步骤:
//  1. 取消上下文，停止所有正在执行的任务
//  2. 停止 cron 调度器
func (sw *ScheduleWorker) Stop() {
	logger.Info("Stopping schedule worker")
	sw.cancel()    // 取消上下文
	sw.cron.Stop() // 停止 cron 调度器
	logger.Info("Schedule worker stopped")
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
func (sw *ScheduleWorker) Reload() error {
	logger.Info("Reloading schedule worker")

	// 移除所有现有的调度任务
	sw.scheduleEntriesMutex.Lock()
	for scheduleID, entryID := range sw.scheduleEntries {
		sw.cron.Remove(entryID)
		delete(sw.scheduleEntries, scheduleID)
	}
	sw.scheduleEntriesMutex.Unlock()

	// 从数据库重新加载所有启用的任务
	schedules, err := sw.dss.GetEnabledSchedules(sw.ctx)
	if err != nil {
		logger.Errorf("Failed to load enabled tasks: %v", err)
		return fmt.Errorf("failed to load enabled tasks: %w", err)
	}

	// 为每个任务创建新的调度条目
	for _, schedule := range schedules {
		if err := sw.schedule(schedule); err != nil {
			logger.Errorf("Failed to schedule schedule %s: %v", schedule.ID, err)
		}
	}

	logger.Info("Schedule worker reloaded")
	return nil
}

// schedule 调度一个定时发现任务
// 该方法为指定的任务创建 cron 调度条目
// 参数:
//   - schedule: 指向 DiscoverSchedule 结构体的指针，包含要调度的任务信息
//
// 返回值:
//   - error: 如果任务调度成功则返回 nil，否则返回相应的错误信息
func (sw *ScheduleWorker) schedule(schedule *interfaces.DiscoverSchedule) error {
	// 检查任务是否已经被调度
	sw.scheduleEntriesMutex.RLock()
	if _, exists := sw.scheduleEntries[schedule.ID]; exists {
		sw.scheduleEntriesMutex.RUnlock()
		logger.Warnf("Schedule %sw is already scheduled", schedule.ID)
		return nil
	}
	sw.scheduleEntriesMutex.RUnlock()

	// 使用 cron 表达式添加任务到调度器
	// 当到达执行时间时，会调用 executeSchedule 方法执行任务
	entryID, err := sw.cron.AddFunc(schedule.CronExpr, func() {
		sw.executeSchedule(schedule.ID)
	})
	if err != nil {
		logger.Errorf("Failed to add cron job for schedule %sw: %v", schedule.ID, err)
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// 保存 cron 条目 ID，用于后续的取消调度操作
	sw.scheduleEntriesMutex.Lock()
	sw.scheduleEntries[schedule.ID] = entryID
	sw.scheduleEntriesMutex.Unlock()

	logger.Infof("Scheduled schedule %sw with cron expression: %s", schedule.ID, schedule.CronExpr)
	return nil
}

// unschedule 取消单个任务的调度
// 该方法从 cron 调度器中移除指定的任务
// 参数:
//   - scheduleID: 要取消调度的任务 ID
//
// 返回值:
//   - error: 总是返回 nil
func (sw *ScheduleWorker) unschedule(scheduleID string) error {
	sw.scheduleEntriesMutex.Lock()
	defer sw.scheduleEntriesMutex.Unlock()

	// 查找任务的 cron 条目 ID
	entryID, exists := sw.scheduleEntries[scheduleID]
	if !exists {
		logger.Warnf("Schedule %s is not scheduled", scheduleID)
		return nil
	}

	// 从 cron 调度器中移除任务
	sw.cron.Remove(entryID)
	delete(sw.scheduleEntries, scheduleID)

	logger.Infof("Unscheduled schedule %s", scheduleID)
	return nil
}

// executeSchedule 执行定时发现任务
// 该方法在任务到达执行时间时被 cron 调度器调用
// 执行步骤:
//  1. 从数据库获取最新的任务状态
//  2. 检查任务是否仍然启用
//  3. 检查任务是否已过期
//  4. 执行任务
//
// 参数:
//   - schedule: 要执行的任务
func (sw *ScheduleWorker) executeSchedule(scheduleID string) {
	logger.Infof("Executing discover schedule: id=%s", scheduleID)

	// 从数据库获取最新的任务状态
	schedule, err := sw.dss.GetByID(sw.ctx, scheduleID)
	if err != nil {
		otellog.LogError(sw.ctx, fmt.Sprintf("Failed to get schedule %s", scheduleID), err)
		return
	}
	if schedule == nil {
		otellog.LogError(sw.ctx, fmt.Sprintf("Schedule %s not found", scheduleID), nil)
		return
	}

	// 检查任务是否仍然启用
	if !schedule.Enabled {
		logger.Warnf("Schedule %s is disabled, skipping execution", schedule.ID)

		// 从调度器中移除已禁用的任务
		if err := sw.unschedule(schedule.ID); err != nil {
			otellog.LogError(sw.ctx, fmt.Sprintf("Failed to unschedule disabled schedule %s", schedule.ID), err)
			return
		}

		logger.Infof("Successfully unscheduled disabled schedule %s", schedule.ID)
		return
	}

	// 检查任务是否已过期
	if schedule.EndTime > 0 && time.Now().UnixMilli() > schedule.EndTime {
		logger.Warnf("Schedule %s has expired, disabling and unscheduling", schedule.ID)

		// 禁用过期任务
		if err := sw.dss.Disable(sw.ctx, schedule.ID); err != nil {
			otellog.LogError(sw.ctx, fmt.Sprintf("Failed to disable expired task %s", schedule.ID), err)
			return
		}

		// 从调度器中移除任务
		if err := sw.unschedule(schedule.ID); err != nil {
			otellog.LogError(sw.ctx, fmt.Sprintf("Failed to unschedule expired schedule %s", schedule.ID), err)
			return
		}

		logger.Infof("Successfully disabled and unscheduled expired schedule %s", schedule.ID)
		return
	}

	// 执行任务
	if err := sw.dss.ExecuteSchedule(sw.ctx, schedule); err != nil {
		otellog.LogError(sw.ctx, fmt.Sprintf("Failed to execute schedule %s", schedule.ID), err)
		return
	}

	logger.Infof("Successfully executed scheduled discover schedule: id=%s", schedule.ID)
}

// Schedule 调度一个任务
// 该方法在创建新任务或启用任务时被调用
// 参数:
//   - scheduleID: 要调度的任务 ID
//
// 返回值:
//   - error: 如果调度失败则返回错误
func (sw *ScheduleWorker) Schedule(scheduleID string) error {
	// 从数据库获取任务信息
	schedule, err := sw.dss.GetByID(sw.ctx, scheduleID)
	if err != nil {
		logger.Errorf("Failed to get schedule %s: %v", scheduleID, err)
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	// 如果任务未启用，则不进行调度
	if !schedule.Enabled {
		logger.Warnf("Schedule %s is disabled, not scheduling", scheduleID)
		return nil
	}

	// 调用内部方法进行调度
	return sw.schedule(schedule)
}

// Unschedule 取消任务的调度
// 该方法在任务被禁用或删除时被调用
// 参数:
//   - scheduleID: 要取消调度的任务 ID
//
// 返回值:
//   - error: 如果取消失败则返回错误
func (sw *ScheduleWorker) Unschedule(scheduleID string) error {
	return sw.unschedule(scheduleID)
}

// UpdateSchedule 更新定时任务的调度
// 该方法在任务配置变更时被调用
// 执行步骤:
//  1. 取消旧的调度
//  2. 从数据库获取更新后的任务
//  3. 如果任务启用，则重新调度
//
// 参数:
//   - scheduleID: 要更新的任务 ID
//
// 返回值:
//   - error: 如果更新失败则返回错误
func (sw *ScheduleWorker) UpdateSchedule(scheduleID string) error {
	// 取消旧的调度
	if err := sw.unschedule(scheduleID); err != nil {
		logger.Errorf("Failed to unschedule schedule %s: %v", scheduleID, err)
	}

	// 从数据库获取更新后的任务
	schedule, err := sw.dss.GetByID(sw.ctx, scheduleID)
	if err != nil {
		logger.Errorf("Failed to get schedule %s: %v", scheduleID, err)
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	// 如果任务启用，则重新调度
	if schedule.Enabled {
		return sw.schedule(schedule)
	}

	return nil
}
