package mod

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/render"
	normalizeutil "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils/normalize"
)

// LoopHandler 循环处理器，从DefExecutor中独立出来
type LoopHandler struct {
	loopParams LoopParameters
}

// NewLoopHandler 创建新的循环处理器
func NewLoopHandler(loopParams LoopParameters) *LoopHandler {
	return &LoopHandler{
		loopParams: loopParams,
	}
}

// HandleLoopAction 处理循环动作的主逻辑
func (h *LoopHandler) HandleLoopAction(ctx context.Context, taskIns *entity.TaskInstance, params interface{}, act entity.Action) error {
	loopParams := h.loopParams
	// Set default mode if not specified
	if loopParams.Mode == "" {
		loopParams.Mode = "limit"
	}

	// For array mode, convert array to actual slice and determine limit
	var arrayData []interface{}
	var actualLimit int

	if loopParams.Mode == "array" {
		if loopParams.Array != nil {
			arrayData = loopParams.Array
			loopParams.Limit = len(arrayData)
		}
		actualLimit = len(arrayData)
	} else {
		// limit mode
		actualLimit = loopParams.Limit
	}

	// 确保循环任务在任务树中
	parser := GetParser().(*DefParser)
	taskTree, ok := parser.getTaskTree(taskIns.DagInsID)
	if !ok {
		log.Warnf("任务树不存在，尝试重新构建")
		// 获取所有任务实例
		tasks, err := GetStore().ListTaskInstance(ctx, &ListTaskInstanceInput{
			DagInsID: taskIns.DagInsID,
		})
		if err != nil {
			log.Errorf("获取任务实例失败: %v", err)
			return fmt.Errorf("get task instances failed: %w", err)
		}

		// 构建新的任务树
		root, err := BuildRootNode(MapTaskInsToGetter(tasks))
		if err != nil {
			log.Errorf("构建任务树失败: %v", err)
			return fmt.Errorf("build task tree failed: %w", err)
		}

		taskTree = &TaskTree{
			DagIns: taskIns.RelatedDagInstance,
			Root:   root,
		}
		parser.taskTrees.Store(taskIns.DagInsID, taskTree)
	}

	// 更新循环任务在任务树中的状态
	walkNode(taskTree.Root, func(node *TaskNode) bool {
		if node.TaskInsID == taskIns.ID {
			node.Status = taskIns.Status
			return false
		}
		return true
	}, true)

	// 如果actualLimit为0，直接标记循环任务为完成状态
	if actualLimit <= 0 {
		taskIns.Status = common.SuccessStatus
		taskIns.Reason = "循环完成，没有迭代项"

		if err := GetStore().UpdateTaskIns(ctx, taskIns); err != nil {
			log.Errorf("更新循环任务 [%s] 状态失败: %v", taskIns.ID, err)
			return fmt.Errorf("update task instance failed: %w", err)
		}

		// 更新任务树中的节点状态
		walkNode(taskTree.Root, func(node *TaskNode) bool {
			if node.TaskInsID == taskIns.ID {
				node.Status = entity.TaskInstanceStatusSuccess
				return false
			}
			return true
		}, true)

		// Notify parser to process next steps after loop
		GetParser().EntryTaskIns(taskIns)
		return nil
	}

	// 检查前一次迭代的任务状态
	if loopParams.CurrentIteration > 0 {
		// 获取当前循环任务的依赖任务
		if len(taskIns.DependOn) > 0 {
			tasks, err := GetStore().ListTaskInstance(ctx, &ListTaskInstanceInput{
				DagInsID: taskIns.DagInsID,
			})
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			// 检查所有依赖任务是否都已完成
			for _, dependTaskID := range taskIns.DependOn {
				dependTaskFound := false
				for _, task := range tasks {
					if task.TaskID == dependTaskID {
						dependTaskFound = true
						if task.Status == entity.TaskInstanceStatusFailed {
							return fmt.Errorf("依赖任务 [%s] 失败", dependTaskID)
						}
						if task.Status != entity.TaskInstanceStatusSuccess && task.Status != entity.TaskInstanceStatusSkipped {
							return nil
						}
						break
					}
				}
				if !dependTaskFound {
					log.Warnf("依赖任务 [%s] 未找到", dependTaskID)
					return nil
				}
			}
		}
	}

	loopParams.Steps = taskIns.Steps

	// Make sure the loop control ID is set to the loop task's ID if not already set
	if loopParams.LoopControlID == "" {
		loopParams.LoopControlID = taskIns.ID
		// If we're setting this for the first time, also store it in params
		if taskIns.GetParams() == nil {
			taskIns.SetParams(make(map[string]interface{}))
		}
		taskIns.SetParam("loop_control_id", taskIns.ID)
	}

	if loopParams.LoopTaskID == "" {
		loopParams.LoopTaskID = taskIns.TaskID
		taskIns.SetParam("loop_task_id", taskIns.TaskID)
	}

	// 如果循环任务的状态是失败或取消，但我们想继续循环，重置状态
	if taskIns.Status == entity.TaskInstanceStatusFailed ||
		taskIns.Status == entity.TaskInstanceStatusCanceled {
		log.Infof("循环任务 [%s] 状态为 %s，重置为初始化状态", taskIns.ID, taskIns.Status)
		taskIns.Status = entity.TaskInstanceStatusInit
		taskIns.Reason = nil

		// 更新任务树中的节点状态
		walkNode(taskTree.Root, func(node *TaskNode) bool {
			if node.TaskInsID == taskIns.ID {
				node.Status = entity.TaskInstanceStatusInit
				log.Infof("更新任务树中循环节点 [%s] 状态为初始化", node.TaskInsID)
				return false
			}
			return true
		}, true)

		if err := GetStore().UpdateTaskIns(ctx, taskIns); err != nil {
			log.Errorf("更新循环任务 [%s] 状态失败: %v", taskIns.ID, err)
			return fmt.Errorf("update task instance failed: %w", err)
		}
	}

	// Get current iteration from ShareData if exists, otherwise use the one from params
	loopKeyPrefix := fmt.Sprintf("__loop_%s_", loopParams.LoopTaskID)
	currentIterationKey := loopKeyPrefix + "current_iteration"

	// Use the iteration from params, do not overwrite from ShareData as it causes race conditions in parallel execution
	log.Infof("当前循环 [%s] 迭代计数: %d", taskIns.ID, loopParams.CurrentIteration)

	// 在增加迭代计数之前检查是否达到限制
	if actualLimit > 0 && loopParams.CurrentIteration >= actualLimit {
		// If we've reached the limit, mark the loop task as completed
		prevStatus := taskIns.Status
		taskIns.Status = common.SuccessStatus
		taskIns.Reason = fmt.Sprintf("循环完成，共执行 %d 次迭代", actualLimit)

		log.Infof("标记循环任务 [%s] 为成功。状态变化: %s -> %s",
			taskIns.ID, prevStatus, taskIns.Status)

		if err := GetStore().UpdateTaskIns(ctx, taskIns); err != nil {
			log.Errorf("更新循环任务 [%s] 状态失败: %v", taskIns.ID, err)
			return fmt.Errorf("update task instance failed: %w", err)
		}

		// 更新任务树中的节点状态
		walkNode(taskTree.Root, func(node *TaskNode) bool {
			if node.TaskInsID == taskIns.ID {
				node.Status = entity.TaskInstanceStatusSuccess
				log.Infof("更新任务树中循环节点 [%s] 状态为成功", node.TaskInsID)
				return false
			}
			return true
		}, true)

		// // Notify parser to process next steps after loop
		// log.Infof("通知解析器处理循环 [%s] 后的步骤", taskIns.ID)
		// GetParser().EntryTaskIns(taskIns)
		// 只在最后一个循环节点收集输出
		if err := h.collectLoopOutputs(ctx, taskIns, &loopParams, taskIns.RelatedDagInstance.ShareData); err != nil {
			log.Warnf("收集循环输出失败: %v", err)
		}
		return nil
	}

	// Store loop variables in ShareData for access by child tasks
	// Create a wrapper for all loop internal values
	loopInternalValues := map[string]interface{}{
		"index": loopParams.CurrentIteration,
	}

	// For array mode, add current value
	if loopParams.Mode == "array" && loopParams.CurrentIteration < len(arrayData) {
		currentValue := arrayData[loopParams.CurrentIteration]
		loopInternalValues["value"] = currentValue
	}

	// 幂等判断：如果已写过本轮，直接 return，避免重复写入和日志
	keys := strings.Split(taskIns.TaskID, "_")
	originKey := keys[0]
	loopKey := fmt.Sprintf("__%s", taskIns.TaskID)
	alreadyExists := false
	if taskIns.RelatedDagInstance != nil && taskIns.RelatedDagInstance.ShareData != nil {
		_, alreadyExists = taskIns.RelatedDagInstance.ShareData.Get(loopKey)
	}

	// Only log on first write
	if !alreadyExists {
		log.Infof("设置循环 [%s] 迭代 %d 的值到ShareData: %v",
			loopParams.LoopControlID, loopParams.CurrentIteration, loopInternalValues)
	}

	// Store all loop internal values under loop control ID namespace
	taskIns.RelatedDagInstance.ShareData.Set(fmt.Sprintf("__%s", originKey), loopInternalValues)
	taskIns.RelatedDagInstance.ShareData.Set(fmt.Sprintf("__%s", taskIns.TaskID), loopInternalValues)

	// Also store individual values for backward compatibility
	taskIns.RelatedDagInstance.ShareData.Set(fmt.Sprintf("__loop_%s_index", originKey), loopParams.CurrentIteration)
	if loopParams.Mode == "array" && loopParams.CurrentIteration < len(arrayData) {
		taskIns.RelatedDagInstance.ShareData.Set(fmt.Sprintf("__loop_%s_value", originKey), arrayData[loopParams.CurrentIteration])
	}

	// Increment the current iteration
	// loopParams.CurrentIteration++
	// log.Infof("增加循环 [%s] 的迭代计数: %d -> %d",
	// 	taskIns.ID, oldIteration, loopParams.CurrentIteration)

	// Store updated iteration count in ShareData
	taskIns.RelatedDagInstance.ShareData.Set(currentIterationKey, loopParams.CurrentIteration)
	taskIns.RelatedDagInstance.ShareData.Set(loopKeyPrefix+"last_iteration_task_id", loopParams.LastIterationTaskID)

	// Store array data for later use by loop executor
	loopParams.Array = arrayData

	_ = taskIns.Run(ctx, params, act, nil)

	// Create a loop executor
	loopExecutor := NewLoopExecutor(ctx, taskIns.RelatedDagInstance, &loopParams)

	// Generate tasks for the current iteration
	_, err := loopExecutor.GenerateIterationTasks()
	if err != nil {
		log.Errorf("为循环 [%s] 生成迭代任务失败: %v", taskIns.ID, err)
		return fmt.Errorf("generate iteration tasks failed: %w", err)
	}

	// Update the loop parameters for the next iteration in task params too
	// This is for backward compatibility
	if taskIns.GetParams() == nil {
		taskIns.SetParams(make(map[string]interface{}))
	}
	taskIns.SetParam("current_iteration", loopParams.CurrentIteration)
	taskIns.SetParam("last_iteration_task_id", loopParams.LastIterationTaskID)

	// Use UpdateTaskIns to fully update the task, including params
	taskIns.Status = entity.TaskInstanceStatusSuccess
	taskIns.Reason = fmt.Sprintf("循环迭代 %d 完成", loopParams.CurrentIteration)
	taskIns.Results = loopInternalValues

	if err := GetStore().UpdateTaskIns(ctx, taskIns); err != nil {
		log.Errorf("更新循环任务 [%s] 失败: %v", taskIns.ID, err)
		return fmt.Errorf("update task instance failed: %w", err)
	}

	// Update the task tree
	if err := loopExecutor.UpdateTaskTree(); err != nil {
		log.Errorf("更新循环 [%s] 的任务树失败: %v", taskIns.ID, err)
		return fmt.Errorf("update task tree failed: %w", err)
	}

	if err := loopExecutor.PushExecutableTasks(); err != nil {
		log.Errorf("推送循环 [%s] 的可执行任务失败: %v", taskIns.ID, err)
		return fmt.Errorf("push executable tasks failed: %w", err)
	}

	return nil
}

// collectLoopOutputs 收集循环输出 (这个方法保留，因为它是Loop特定的逻辑)
func (h *LoopHandler) collectLoopOutputs(ctx context.Context, taskIns *entity.TaskInstance, loopParams *LoopParameters, shareData *entity.ShareData) error {
	if len(loopParams.Outputs) == 0 {
		return nil
	}

	// Initialize outputs map in ShareData
	loopOutputsKey := fmt.Sprintf("__%s", strings.Split(taskIns.TaskID, "_")[0])
	outputsMap := make(map[string]interface{})

	// Initialize all output arrays
	for _, output := range loopParams.Outputs {
		outputsMap[output.Key] = make([]interface{}, 0)
	}

	// Get existing outputs if any
	if existingOutputs, ok := shareData.Get(loopOutputsKey); ok {
		if existingMap, ok := existingOutputs.(map[string]interface{}); ok {
			if outputs, ok := existingMap["outputs"].(map[string]interface{}); ok {
				outputsMap = outputs
			}
		}
	}

	// Collect outputs from all completed iterations
	tasks, err := GetStore().ListTaskInstance(ctx, &ListTaskInstanceInput{
		DagInsID: taskIns.DagInsID,
	})
	if err != nil {
		return fmt.Errorf("failed to list task instances: %w", err)
	}

	// Group tasks by iteration
	iterationTasks := make(map[int][]*entity.TaskInstance)
	loopControlPrefix := strings.Split(loopParams.LoopTaskID, "_i")[0] + "_i"

	for _, task := range tasks {
		if strings.HasPrefix(task.TaskID, loopControlPrefix) {
			// Extract iteration number from task ID format: loopid_iter_N_stepindex_stepid
			parts := strings.Split(task.TaskID, "_")
			if len(parts) == 2 {
				if iterStr := parts[1]; iterStr != "" {
					if iter, err := strconv.Atoi(iterStr[1:]); err == nil {
						iterationTasks[iter] = append(iterationTasks[iter], task)
					}
				}
			}
		}
	}

	// Process each completed iteration
	for iter := 1; iter <= loopParams.CurrentIteration; iter++ {

		// For each output configuration
		for _, outputConfig := range loopParams.Outputs {
			var outputValue interface{}

			// 从当前迭代的任务中获取输出
			if taskList, exists := iterationTasks[iter]; exists {
				for _, task := range taskList {
					// 从任务的参数中获取输出
					taskParams := task.GetParams()
					if outputs, ok := normalizeutil.AsSlice(taskParams["outputs"]); ok {
						for _, output := range outputs {
							if outputMap, ok := output.(map[string]interface{}); ok {
								if key, ok := outputMap["key"].(string); ok && key == outputConfig.Key {
									if value, ok := outputMap["value"]; ok {
										outputValue = value
										break
									}
								}
							}
						}
					}
					if outputValue != nil {
						break
					}
				}
			}

			temp := []interface{}{}
			if outputsMap[outputConfig.Key] != nil {
				temp = outputsMap[outputConfig.Key].([]interface{})
			}

			// Add to output array if not already present for this iteration
			if len(temp) <= iter {
				temp = append(temp, outputValue)
				outputsMap[outputConfig.Key] = temp
			}
		}
	}

	// Store collected outputs in ShareData
	shareData.Set(loopOutputsKey, map[string]interface{}{
		"outputs": outputsMap,
	})
	return nil
}

// LoopParameters defines the parameters for loop execution
type LoopParameters struct {
	Mode                string        `json:"mode"`
	Limit               int           `json:"limit"`
	Array               []interface{} `json:"array"`
	Outputs             []LoopOutput  `json:"outputs"`
	CurrentIteration    int           `json:"current_iteration"`
	LastIterationTaskID string        `json:"last_iteration_task_id"`
	LoopControlID       string        `json:"loop_control_id"`
	Steps               []entity.Step `json:"steps"`
	LoopTaskID          string        `json:"loop_task_id"`
}

// LoopOutput defines output configuration for loop
type LoopOutput struct {
	Key   string `json:"key"`   // Output key name
	Value string `json:"value"` // Template for output value (e.g., "{{__2.data.result}}")
}

// LoopExecutor handles the dynamic generation and execution of loop tasks
type LoopExecutor struct {
	ctx        context.Context
	dagIns     *entity.DagInstance
	loopParams *LoopParameters
	taskTree   *TaskTree
	tplRender  *render.TplRender
	// 添加互斥锁防止并发创建任务
	taskCreationMutex sync.Mutex
	// 超时配置
	timeoutConfig *common.TimeoutConfig
}

// NewLoopExecutor creates a new loop executor with default timeout config
func NewLoopExecutor(ctx context.Context, dagIns *entity.DagInstance, loopParams *LoopParameters) *LoopExecutor {
	return &LoopExecutor{
		ctx:           ctx,
		dagIns:        dagIns,
		loopParams:    loopParams,
		tplRender:     render.NewTplRender(),
		timeoutConfig: common.NewTimeoutConfig(),
	}
}

// NewLoopExecutorWithConfig creates a new loop executor with custom timeout config
func NewLoopExecutorWithConfig(ctx context.Context, dagIns *entity.DagInstance, loopParams *LoopParameters, timeoutConfig *common.TimeoutConfig) *LoopExecutor {
	if timeoutConfig == nil {
		timeoutConfig = common.NewTimeoutConfig()
	}
	return &LoopExecutor{
		ctx:           ctx,
		dagIns:        dagIns,
		loopParams:    loopParams,
		tplRender:     render.NewTplRender(),
		timeoutConfig: timeoutConfig,
	}
}

// GenerateIterationTasks creates task instances for the current iteration
func (e *LoopExecutor) GenerateIterationTasks() ([]*entity.TaskInstance, error) {
	// 使用互斥锁防止并发创建任务
	e.taskCreationMutex.Lock()
	defer e.taskCreationMutex.Unlock()

	var taskInstances []*entity.TaskInstance
	var lastTaskIDs []string

	// Get the shared data
	shareData := e.dagIns.ShareData
	if shareData == nil {
		shareData = &entity.ShareData{
			Dict: make(map[string]interface{}),
		}
		e.dagIns.ShareData = shareData
	}

	// Use both specific loop control ID prefixed key and general __loop_index
	loopKeyPrefix := fmt.Sprintf("__loop_%s_", e.loopParams.LoopTaskID)

	// Update or set the current iteration in shared data
	shareData.Set(loopKeyPrefix+"current_iteration", e.loopParams.CurrentIteration)

	// Also set the general loop index for template rendering compatibility

	// 检查数据库中是否已存在当前迭代的任务，避免重复创建
	existingTasks, err := GetStore().ListTaskInstance(e.ctx, &ListTaskInstanceInput{
		DagInsID: e.dagIns.ID,
	})
	if err != nil {
		log.Errorf("Failed to list existing tasks: %v", err)
		return nil, fmt.Errorf("获取现有任务失败: %w", err)
	}

	// 创建现有任务ID的映射，用于快速查找
	existingTaskIDs := make(map[string]bool)
	existingTaskIDMap := make(map[string]*entity.TaskInstance)
	for _, task := range existingTasks {
		existingTaskIDs[task.ID] = true
		existingTaskIDMap[task.TaskID] = task
	}

	// 检查当前迭代的任务是否已经存在
	currentIterPrefix := fmt.Sprintf("%s_i%d_", strings.Split(e.loopParams.LoopTaskID, "_i")[0], e.loopParams.CurrentIteration)
	hasCurrentIterTasks := false
	for _, task := range existingTasks {
		if strings.HasPrefix(task.TaskID, currentIterPrefix) {
			hasCurrentIterTasks = true
			break
		}
	}

	// 如果当前迭代的任务已经存在，直接返回
	if hasCurrentIterTasks {
		log.Infof("当前迭代 %d 的任务已存在，跳过创建", e.loopParams.CurrentIteration)
		return []*entity.TaskInstance{}, nil
	}

	// 检查前一次迭代的任务是否都已完成
	if e.loopParams.CurrentIteration > 0 {
		prevIter := e.loopParams.CurrentIteration - 1
		prevIterPrefix := fmt.Sprintf("%s_iter_%d_", e.loopParams.LoopControlID, prevIter)

		var prevIterTasks []*entity.TaskInstance
		for _, task := range existingTasks {
			if strings.HasPrefix(task.ID, prevIterPrefix) {
				prevIterTasks = append(prevIterTasks, task)
			}
		}

		// 检查前一次迭代的所有任务状态
		allTasksCompleted := true
		for _, task := range prevIterTasks {
			if task.Status != entity.TaskInstanceStatusSuccess {
				allTasksCompleted = false
				break
			}
		}

		if !allTasksCompleted {
			return nil, nil
		}
	}

	// 获取前一次迭代的最后一个任务ID
	prevIterLastTaskID := ""
	if e.loopParams.CurrentIteration > 0 {
		prevIter := e.loopParams.CurrentIteration - 1
		prevIterPrefix := fmt.Sprintf("%s_iter_%d_", e.loopParams.LoopControlID, prevIter)

		// 查找前一次迭代的所有任务
		var prevIterTasks []*entity.TaskInstance
		for _, task := range existingTasks {
			if strings.HasPrefix(task.ID, prevIterPrefix) {
				prevIterTasks = append(prevIterTasks, task)
			}
		}

		// 按步骤索引排序，确保找到最后一个任务
		if len(prevIterTasks) > 0 {
			// 从任务ID中提取步骤索引
			lastStepIndex := -1
			for _, task := range prevIterTasks {
				parts := strings.Split(task.ID, "_")
				if len(parts) >= 4 {
					if stepIndex, err := strconv.Atoi(parts[3]); err == nil {
						if stepIndex > lastStepIndex {
							lastStepIndex = stepIndex
							prevIterLastTaskID = task.ID
						}
					}
				}
			}

			if prevIterLastTaskID == "" {
				log.Warnf("Could not determine last task from previous iteration %d for loop [%s]",
					prevIter, e.loopParams.LoopControlID)
			} else {
				log.Infof("Found last task from previous iteration %d: %s (step index: %d)",
					prevIter, prevIterLastTaskID, lastStepIndex)
			}
		} else {
			log.Warnf("No tasks found for previous iteration %d of loop [%s]",
				prevIter, e.loopParams.LoopControlID)
		}
	}

	// Create tasks for each step in the iteration
	for j, step := range e.loopParams.Steps {
		// Generate unique task ID with iteration number and step index
		baseTaskID := strings.Split(e.loopParams.LoopTaskID, "_i")[0]
		taskID := fmt.Sprintf("%s_i%d_s%s",
			baseTaskID,
			e.loopParams.CurrentIteration,
			step.ID)

		// 检查任务是否已存在
		if existingTaskIDs[taskID] {
			log.Warnf("Task with ID [%s] already exists, skipping creation", taskID)
			lastTaskIDs = []string{taskID}
			continue
		}

		// Create task dependencies
		var dependOn []string

		if j > 0 {
			prevStep := e.loopParams.Steps[j-1]
			baseTaskID := strings.Split(e.loopParams.LoopTaskID, "_i")[0]

			// Check if previous step is a parallel branch
			if prevStep.Operator == common.ControlFlowParallel {
				// Get all last task IDs from parallel branches
				prevTaskID := fmt.Sprintf("%s_i%d_s%s",
					baseTaskID,
					e.loopParams.CurrentIteration,
					prevStep.ID)

				// Use helper to get all branch last tasks
				for i, branch := range prevStep.Branches {
					branchLastTasks := getLoopLastTaskIDs(branch.Steps, prevTaskID, i)
					dependOn = append(dependOn, branchLastTasks...)
				}
			} else if prevStep.Operator == common.BranchOpt {
				// Check if previous step is a branch option
				// Get all last task IDs from branches
				prevTaskID := fmt.Sprintf("%s_i%d_s%s",
					baseTaskID,
					e.loopParams.CurrentIteration,
					prevStep.ID)

				// Use helper to get all branch last tasks
				for i, branch := range prevStep.Branches {
					branchLastTasks := getLoopLastTaskIDs(branch.Steps, prevTaskID, i)
					dependOn = append(dependOn, branchLastTasks...)
				}
			} else {
				// Regular step - single dependency
				prevStepTaskID := fmt.Sprintf("%s_i%d_s%s",
					baseTaskID,
					e.loopParams.CurrentIteration,
					prevStep.ID)
				dependOn = append(dependOn, prevStepTaskID)
			}
		} else {
			// 如果是第一个步骤，依赖循环节点
			dependOn = append(dependOn, e.loopParams.LoopTaskID)
		}

		// 设置超时时间
		getTimeout := func(operator string) int {
			return e.timeoutConfig.GetTimeout(operator)
		}

		// 递归处理步骤（包括嵌套分支）
		taskInstances, lastTaskIDs = e.processStepRecursively(
			taskInstances, step, taskID, dependOn, existingTaskIDs, lastTaskIDs, getTimeout)
	}

	// 为下一次迭代创建新的循环节点副本
	nextIteration := e.loopParams.CurrentIteration + 1
	baseTaskID := strings.Split(e.loopParams.LoopTaskID, "_i")[0]

	// 在第一次迭代时，查找并存储依赖循环节点的任务
	if e.loopParams.CurrentIteration == 0 {
		tasks, err := GetStore().ListTaskInstance(e.ctx, &ListTaskInstanceInput{
			DagInsID: e.dagIns.ID,
		})
		if err != nil {
			log.Errorf("获取任务实例失败: %v", err)
		} else {
			var dependentTasks []string
			// 查找依赖原循环节点的任务
			for _, task := range tasks {
				// 跳过循环内的任务
				if strings.HasPrefix(task.TaskID, baseTaskID+"_i") {
					continue
				}

				// 检查任务是否依赖原循环节点
				for _, dependID := range task.DependOn {
					if dependID == e.loopParams.LoopTaskID {
						dependentTasks = append(dependentTasks, task.ID)
						break
					}
				}
			}
			// 将依赖任务列表存储到 sharedata
			if len(dependentTasks) > 0 {
				shareData.Set(loopKeyPrefix+"dependent_tasks", dependentTasks)
			}
		}
	}

	if nextIteration <= e.loopParams.Limit || nextIteration <= len(e.loopParams.Array) {
		// 检查下一次迭代的循环任务是否已存在
		nextLoopTaskID := fmt.Sprintf("%s_i%d", baseTaskID, nextIteration)
		if _, exists := existingTaskIDMap[nextLoopTaskID]; exists {
			log.Infof("下一次迭代的循环任务 [%s] 已存在，跳过创建", nextLoopTaskID)
		} else {
			// 创建新的循环节点副本
			newLoopTask := &entity.TaskInstance{
				DagInsID:           e.dagIns.ID,
				ActionName:         common.Loop,
				Status:             entity.TaskInstanceStatusInit,
				Params:             make(map[string]interface{}),
				RelatedDagInstance: e.dagIns,
				Steps:              e.loopParams.Steps,
			}

			// 复制参数
			if taskIns, err := GetStore().GetTaskIns(e.ctx, e.loopParams.LoopControlID); err == nil && taskIns != nil {
				// 只使用基础任务ID和当前迭代号，不累积历史迭代信息
				newLoopTask.TaskID = nextLoopTaskID
				paramsCopy := taskIns.GetParams()
				for k, v := range paramsCopy {
					newLoopTask.SetParam(k, v)
				}
			}
			// 更新迭代相关参数
			newLoopTask.SetParam("current_iteration", nextIteration)
			// lastTaskIDs 可能有多个，取最后一个作为标记，或者在此处需要调整逻辑?
			// 循环节点的 last_iteration_task_id 通常用于记录上一次迭代的结束点
			// 如果有多个结束点，这里可能只需要一个作为引用，或者需要改为数组
			// 目前主要是 user case 中用到，暂时取第一个或 join
			var lastTaskIDStr string
			if len(lastTaskIDs) > 0 {
				lastTaskIDStr = lastTaskIDs[len(lastTaskIDs)-1]
			}
			newLoopTask.SetParam("last_iteration_task_id", lastTaskIDStr)

			// 设置依赖关系 - 依赖当前迭代的最后一个任务(s)
			newLoopTask.DependOn = lastTaskIDs

			taskInstances = append(taskInstances, newLoopTask)

			// 从 sharedata 获取依赖任务列表并更新依赖关系
			dependentTasksValue, exists := shareData.Get("__loop_" + baseTaskID + "_dependent_tasks")
			if exists {
				var dependentTasks []interface{}
				var ok bool

				// 尝试断言为[]any类型
				if dependentTasks, ok = dependentTasksValue.([]interface{}); ok {
					// 成功断言为[]interface{}，直接使用
				} else if stringSlice, ok := dependentTasksValue.([]string); ok {
					// 将[]string转换为[]interface{}
					dependentTasks = make([]interface{}, len(stringSlice))
					for i, v := range stringSlice {
						dependentTasks[i] = v
					}
				} else {
					// 类型不匹配，跳过处理
					log.Warnf("dependentTasksValue类型不匹配，跳过处理")
				}

				if len(dependentTasks) > 0 {
					for _, taskID := range dependentTasks {
						task, err := GetStore().GetTaskIns(e.ctx, taskID.(string))
						if err != nil {
							log.Errorf("获取任务实例失败: %v", err)
							continue
						}

						// 更新依赖关系
						for _, dependID := range task.DependOn {
							if dependID == e.loopParams.LoopTaskID {
								task.DependOn = append(task.DependOn, newLoopTask.TaskID)
								// 更新任务实例
								if err := GetStore().PatchTaskIns(e.ctx, task); err != nil {
									log.Errorf("更新任务依赖失败: %v", err)
									continue
								}
								break
							}
						}
					}
				}
			}
		}
	}

	// Update last iteration task ID and store in shared data
	var finalLastTaskID string
	if len(lastTaskIDs) > 0 {
		finalLastTaskID = lastTaskIDs[len(lastTaskIDs)-1]
	}
	e.loopParams.LastIterationTaskID = finalLastTaskID
	shareData.Set(loopKeyPrefix+"last_iteration_task_id", finalLastTaskID)

	// 如果没有需要创建的任务（可能全部已存在），直接返回
	if len(taskInstances) == 0 {
		// 确保我们有一个有效的lastTaskID
		if finalLastTaskID == "" {
			// 查找迭代中最后一个步骤的ID
			if len(e.loopParams.Steps) > 0 {
				lastStep := e.loopParams.Steps[len(e.loopParams.Steps)-1]
				computedLastID := fmt.Sprintf("%s_iter_%d_%d_%s",
					e.loopParams.LoopControlID,
					e.loopParams.CurrentIteration,
					len(e.loopParams.Steps)-1,
					lastStep.ID)
				log.Infof("Using computed last task ID: %s", computedLastID)
				finalLastTaskID = computedLastID
			}
		}

		e.loopParams.LastIterationTaskID = finalLastTaskID
		shareData.Set(loopKeyPrefix+"last_iteration_task_id", finalLastTaskID)

		return []*entity.TaskInstance{}, nil
	}

	// Batch create task instances
	createdTasks, err := GetStore().BatchCreateTaskIns(e.ctx, taskInstances)
	if err != nil {
		log.Errorf("Failed to create loop iteration tasks: %v", err)
		return nil, fmt.Errorf("创建任务实例失败: %w", err)
	}

	return createdTasks, nil
}

// getLoopLastTaskIDs 递归获取循环迭代步骤中的最后一个(或多个)任务ID
// 对于普通步骤,返回该步骤的ID
// 对于并行分支,递归收集所有子分支的最后任务ID
func getLoopLastTaskIDs(steps []entity.Step, parentTaskID string, branchIndex int) []string {
	if len(steps) == 0 {
		return []string{}
	}

	lastStep := steps[len(steps)-1]

	// 如果最后一个步骤是并行节点,递归获取其所有分支的最后任务
	if lastStep.Operator == common.ControlFlowParallel {
		result := []string{}
		for i, branch := range lastStep.Branches {
			nestedParentID := fmt.Sprintf("%s_%d_%s", parentTaskID, branchIndex, lastStep.ID)
			branchLastTasks := getLoopLastTaskIDs(branch.Steps, nestedParentID, i)
			result = append(result, branchLastTasks...)
		}
		return result
	}

	// 如果最后一个步骤是条件分支节点,递归获取其所有分支的最后任务
	if lastStep.Operator == common.BranchOpt {
		result := []string{}
		for i, branch := range lastStep.Branches {
			currentID := fmt.Sprintf("%s_%d_%s", parentTaskID, branchIndex, lastStep.ID)
			branchLastTasks := getLoopLastTaskIDs(branch.Steps, currentID, i)
			result = append(result, branchLastTasks...)
		}
		// 如果条件分支没有有效的分支任务(例如空分支), 这里的result可能为空
		// 这种情况下, 后续任务应该依赖于条件分支节点本身
		if len(result) > 0 {
			return result
		}
	}

	// 否则,返回该步骤的ID
	return []string{fmt.Sprintf("%s_%d_%s", parentTaskID, branchIndex, lastStep.ID)}
}

// processStepRecursively 递归处理步骤，支持嵌套分支
func (e *LoopExecutor) processStepRecursively(
	taskInstances []*entity.TaskInstance,
	step entity.Step,
	taskID string,
	dependOn []string,
	existingTaskIDs map[string]bool,
	lastTaskIDs []string, // Unused in this logic actually, kept for signature parallel? No, signature changed.
	getTimeout func(string) int,
) ([]*entity.TaskInstance, []string) {

	// 处理并行分支节点
	if step.Operator == common.ControlFlowParallel {
		// 并行分支节点不创建任务，只处理所有分支
		var parallelBranchLastTasks []string
		branchEntryDependOn := dependOn

		for i, branch := range step.Branches {
			// 处理分支内的每个步骤
			// 分支内的步骤是串行的，但第一个步骤依赖 `dependOn`
			currentBranchPathLastTasks := branchEntryDependOn

			for j, branchStep := range branch.Steps {
				// 创建步骤任务ID
				stepTaskID := fmt.Sprintf("%s_%d_%s", taskID, i, branchStep.ID)

				// 检查任务是否已存在
				if existingTaskIDs[stepTaskID] {
					log.Warnf("Parallel branch step task with ID [%s] already exists, skipping creation", stepTaskID)
					// 如果已存在，更新依赖链为该已存在任务
					currentBranchPathLastTasks = []string{stepTaskID}
					continue
				}

				// 确定依赖关系
				var branchStepDependOn []string
				if j == 0 {
					// 分支的第一个步骤依赖进入并行分支前的节点
					branchStepDependOn = branchEntryDependOn
				} else {
					// 后续步骤依赖分支内的前一个步骤的最后任务(s)
					branchStepDependOn = currentBranchPathLastTasks
				}

				// 递归处理分支内的步骤
				var stepResultTasks []string
				taskInstances, stepResultTasks = e.processStepRecursively(
					taskInstances, branchStep, stepTaskID, branchStepDependOn,
					existingTaskIDs, nil, getTimeout)

				// 更新当前分支路径的最后任务ID(s)
				if len(stepResultTasks) > 0 {
					currentBranchPathLastTasks = stepResultTasks
				}
			}

			// 收集该分支的最后任务ID
			// 如果分支为空，应该依赖并行节点前的依赖?
			// 逻辑上空分支意味着直通，所以 currentBranchPathLastTasks 保持为 branchEntryDependOn
			// 但这里 `getLoopLastTaskIDs` helper may be used?
			// No, verify with logic. If `branch.Steps` is empty, loop above doesn't run. `currentBranchPathLastTasks` == `branchEntryDependOn`.
			// So empty branch just propagates dependency. Correct.

			parallelBranchLastTasks = append(parallelBranchLastTasks, currentBranchPathLastTasks...)
		}

		// 返回所有并行分支的最后任务ID
		return taskInstances, parallelBranchLastTasks
	}

	// 处理条件分支节点
	if step.Operator == common.BranchOpt {
		// 创建分支任务
		branchTask := &entity.Task{
			ID:          taskID,
			ActionName:  common.BranchOpt,
			DependOn:    dependOn,
			Params:      step.Parameters,
			TimeoutSecs: getTimeout(common.BranchOpt),
			Name:        step.Title,
		}

		// 创建分支任务实例
		branchTaskInstance := entity.NewTaskInstance(e.dagIns.ID, branchTask)
		taskInstances = append(taskInstances, branchTaskInstance)

		// 记录所有分支的最后任务ID
		var allBranchesLastTasks []string

		// 分支节点本身的ID
		branchTaskID := taskID

		for i, branch := range step.Branches {
			// 处理分支
			prechecks := entity.PreChecks{}
			// 为每个条件创建预检查
			for j, conditions := range branch.Conditions {
				check := &entity.Check{
					Conditions: conditions,
					Act:        entity.ActiveActionSkip,
				}
				prechecks[fmt.Sprintf("%s_%d_%d", taskID, i, j)] = check
			}

			// 处理分支内的步骤（递归处理）
			// 分支内的初始依赖是分支任务节点本身
			branchPathLastTasks := []string{branchTaskID}

			for _, branchStep := range branch.Steps {
				// 创建步骤任务ID
				stepTaskID := fmt.Sprintf("%s_%d_%s", taskID, i, branchStep.ID)

				// 检查分支步骤任务是否已存在
				if existingTaskIDs[stepTaskID] {
					log.Warnf("Branch step task with ID [%s] already exists, skipping creation", stepTaskID)
					branchPathLastTasks = []string{stepTaskID}
					continue
				}

				// 分支步骤的依赖：依赖当前分支路径的前一(批)任务
				branchStepDependOn := branchPathLastTasks

				// 递归处理分支内的步骤
				var stepResultTasks []string
				taskInstances, stepResultTasks = e.processStepRecursively(
					taskInstances, branchStep, stepTaskID, branchStepDependOn,
					existingTaskIDs, nil, getTimeout)

				// 更新分支内步骤的最后一个任务ID
				if len(stepResultTasks) > 0 {
					branchPathLastTasks = stepResultTasks
				}

				// 为递归处理后的任务添加预检查
				// 注意：这里简单的预检查添加逻辑假设递归返回的任务是该步骤创建的顶层任务
				// 如果是嵌套结构，可能需要更复杂的逻辑，但目前先保持原有逻辑
				// 只是现在我们有多个 stepResultTasks，通常步骤的第一个任务添加precheck?
				// 原有逻辑: Check lastInstance.TaskID == stepTaskID.
				// 如果是并行节点，它不创建任务。步骤ID也不对应任务。
				// 如果是嵌套情况，需要确保只有直接属于该条件的第一个任务被添加precheck。
				// 原逻辑在`processStepRecursively`返回后做这个检查。
				// 如果是普通步骤，created 1 task.
				if len(taskInstances) > 0 {
					// 查找刚才创建的任务。
					// 由于 `processStepRecursively` append to `taskInstances`.
					// 我们可能需要知道哪些是新添加的?
					// 简化起见，我们只能检查最近的任务。
					// 但如果是并行，可能添加多个。
					// 暂且保留原逻辑的意图: 只有当任务ID匹配stepTaskID时(即普通节点或分支节点)才添加。
					for k := len(taskInstances) - 1; k >= 0; k-- {
						if taskInstances[k].TaskID == stepTaskID {
							taskInstances[k].PreChecks = prechecks
							break
						}
						// 如果不匹配，可能是嵌套内部的任务，直到找到顶层?
						// 或者，我们不应该添加给嵌套内部的任务?
						// 只给直接子步骤添加。
					}
				}
			}

			// 如果分支为空(branch.Steps为空), branchPathLastTasks 仍为 [branchTaskID]
			// 这种情况下，分支结束就即是分支节点本身。
			// 收集该分支路径的最后任务
			allBranchesLastTasks = append(allBranchesLastTasks, branchPathLastTasks...)
		}

		// 如果没有有效的分支结果(即没有分支)，依赖分支节点本身
		if len(allBranchesLastTasks) == 0 {
			allBranchesLastTasks = []string{branchTaskID}
		}

		return taskInstances, allBranchesLastTasks
	}

	// 处理普通节点
	task := &entity.Task{
		ID:          taskID,
		ActionName:  step.Operator,
		DependOn:    dependOn,
		Params:      step.Parameters,
		TimeoutSecs: getTimeout(step.Operator),
		Name:        step.Title,
	}
	taskInstance := entity.NewTaskInstance(e.dagIns.ID, task)
	taskInstances = append(taskInstances, taskInstance)

	return taskInstances, []string{taskID}
}

// UpdateTaskTree updates or creates the task tree for the DAG
func (e *LoopExecutor) UpdateTaskTree() error {
	parser := GetParser().(*DefParser)
	taskTree, ok := parser.getTaskTree(e.dagIns.ID)
	if !ok {
		// Create new task tree
		tasks, err := GetStore().ListTaskInstance(e.ctx, &ListTaskInstanceInput{
			DagInsID: e.dagIns.ID,
		})
		if err != nil {
			return fmt.Errorf("获取任务实例失败: %w", err)
		}

		// 检测和处理重复的任务ID
		taskIDMap := make(map[string]*entity.TaskInstance)
		uniqueTasks := make([]*entity.TaskInstance, 0, len(tasks))
		duplicateTasks := make([]string, 0)

		for _, task := range tasks {
			if _, ok := taskIDMap[task.TaskID]; ok {
				duplicateTasks = append(duplicateTasks, task.ID)
			} else {
				taskIDMap[task.TaskID] = task
				uniqueTasks = append(uniqueTasks, task)
			}
		}

		// 如果有重复任务，删除它们
		if len(duplicateTasks) > 0 {
			if derr := GetStore().BatchDeleteTaskIns(e.ctx, duplicateTasks); derr != nil {
				log.Errorf("删除重复任务失败: %v", derr)
				// 继续尝试使用去重后的任务列表
			}
			tasks = uniqueTasks
		}

		root, err := BuildRootNode(MapTaskInsToGetter(tasks))
		if err != nil {
			return fmt.Errorf("构建任务树失败: %w", err)
		}

		taskTree = &TaskTree{
			DagIns: e.dagIns,
			Root:   root,
		}

		parser.taskTrees.Store(e.dagIns.ID, taskTree)
	} else {
		// Update existing task tree
		tasks, err := GetStore().ListTaskInstance(e.ctx, &ListTaskInstanceInput{
			DagInsID: e.dagIns.ID,
		})
		if err != nil {
			return fmt.Errorf("获取任务实例失败: %w", err)
		}

		// 同样在更新时检测和处理重复的任务ID
		taskIDMap := make(map[string]*entity.TaskInstance)
		uniqueTasks := make([]*entity.TaskInstance, 0, len(tasks))
		duplicateTasks := make([]string, 0)

		for _, task := range tasks {
			if _, ok := taskIDMap[task.TaskID]; ok {
				duplicateTasks = append(duplicateTasks, task.ID)
			} else {
				taskIDMap[task.TaskID] = task
				uniqueTasks = append(uniqueTasks, task)
			}
		}

		// 如果有重复任务，删除它们
		if len(duplicateTasks) > 0 {
			if derr := GetStore().BatchDeleteTaskIns(e.ctx, duplicateTasks); derr != nil {
				log.Errorf("删除重复任务失败: %v", derr)
				// 继续尝试使用去重后的任务列表
			}
			tasks = uniqueTasks
		}

		root, err := BuildRootNode(MapTaskInsToGetter(tasks))
		if err != nil {
			return fmt.Errorf("构建任务树失败: %w", err)
		}

		taskTree.Root = root
	}

	e.taskTree = taskTree
	return nil
}

// PushExecutableTasks pushes executable tasks to the executor
func (e *LoopExecutor) PushExecutableTasks() error {
	if e.taskTree == nil {
		log.Errorf("Task tree for loop [%s] is nil, cannot push executable tasks", e.loopParams.LoopControlID)
		return fmt.Errorf("任务树未初始化")
	}

	executableTaskIds := e.taskTree.Root.GetExecutableTaskIds()

	tasks, err := GetStore().ListTaskInstance(e.ctx, &ListTaskInstanceInput{
		DagInsID: e.dagIns.ID,
	})
	if err != nil {
		log.Errorf("Failed to list task instances for DAG [%s]: %v", e.dagIns.ID, err)
		return fmt.Errorf("获取任务实例失败: %w", err)
	}

	taskMap := make(map[string]*entity.TaskInstance)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// 检查当前循环迭代中的任务是否有失败的
	currentIterFailure := false
	failedTasks := []string{}
	failureReasons := []string{}
	currentIterPrefix := fmt.Sprintf("%s_iter_%d_", e.loopParams.LoopControlID, e.loopParams.CurrentIteration)

	for _, task := range tasks {
		if strings.HasPrefix(task.ID, currentIterPrefix) {

			if task.Status == entity.TaskInstanceStatusFailed ||
				task.Status == entity.TaskInstanceStatusCanceled {
				currentIterFailure = true
				failedTasks = append(failedTasks, task.ID)
				reason := fmt.Sprintf("%v", task.Reason)
				failureReasons = append(failureReasons, reason)
			}
		}
	}

	// 如果当前迭代有失败的任务，但循环本身没有被取消，重置循环任务状态为初始状态
	if currentIterFailure {
		// 添加更详细的错误信息
		for i, taskID := range failedTasks {
			if task, ok := taskMap[taskID]; ok {
				log.Errorf("失败任务 %d: [%s] 状态: %s, 原因: %v",
					i+1, taskID, task.Status, task.Reason)
			}
		}

		// 查找循环控制任务
		loopTask, exists := taskMap[e.loopParams.LoopControlID]
		if exists {

			if loopTask.Status != entity.TaskInstanceStatusCanceled {
				// 重置循环任务为初始状态，允许在失败后继续执行下一迭代

				loopTask.Status = entity.TaskInstanceStatusInit
				failReason := fmt.Sprintf("迭代 %d 有失败任务: %v; 原因: %v",
					e.loopParams.CurrentIteration, failedTasks, failureReasons)
				loopTask.Reason = failReason

				if err := GetStore().UpdateTaskIns(e.ctx, loopTask); err != nil {
					log.Warnf("重置循环任务状态失败: %v", err)
				}

				// 将循环任务加入到可执行任务列表
				GetExecutor().Push(e.dagIns, loopTask)
				return nil
			} else {
				log.Warnf("循环任务 [%s] 处于canceled状态，无法重置", loopTask.ID)
			}
		} else {
			log.Errorf("循环控制任务 [%s] 在任务映射中未找到", e.loopParams.LoopControlID)
		}
	}

	for _, id := range executableTaskIds {
		if taskInstance, ok := taskMap[id]; ok {
			// 只有当cancelMap中不存在该任务ID时才执行Push，防止重复推送
			if _, exists := GetExecutor().(*DefExecutor).cancelMap.Load(taskInstance.ID); !exists {
				GetExecutor().Push(e.dagIns, taskInstance)
			} else {
				log.Infof("任务 [%s] 已在cancelMap中，跳过推送", taskInstance.ID)
			}
		} else {
			log.Warnf("可执行任务 [%s] 在任务映射中未找到", id)
		}
	}

	return nil
}
