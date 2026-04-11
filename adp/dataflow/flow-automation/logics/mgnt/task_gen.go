package mgnt

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"maps"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	aerr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	commonLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/log"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
)

// LoopParam 循环节点步骤参数
type LoopParam struct {
	LoopTaskID string
	Index      int
	LoopItems  []any
	StepMap    map[string]string
}

// GenerateTaskResults 动态生成任务执行结果
func (m *mgnt) GenerateTaskResults(ctx context.Context, dagID, dagInsID string, page, limit int64) (int64, []*entity.TaskInstance, error) {
	var (
		err      error
		allTasks []*entity.TaskInstance
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dagIns, err := m.mongo.GetDagInstance(ctx, dagInsID)
	if err != nil {
		log.Warnf("[logic.GenerateTaskResults] GetDagInstance err, detail: %s", err.Error())
		return 0, nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	dag, err := m.mongo.GetDagWithOptionalVersion(ctx, dagID, dagIns.VersionID)
	if err != nil {
		log.Warnf("[logic.GenerateTaskResults] GetDagWithOptionalVersion err, detail: %s", err.Error())
		return 0, nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	contents, exist := m.memoryCache.Get(dagInsID)
	if exist {
		err = json.Unmarshal([]byte(contents), &allTasks)
		if err != nil {
			log.Warnf("[logic.GenerateTaskResults] Unmarshal err, detail: %s", err.Error())
			return 0, nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
		}
	} else {
		switch dagIns.EventPersistence {
		case entity.DagInstanceEventPersistenceSql:
			events, err := dagIns.ListEvents(ctx, &rds.DagInstanceEventListOptions{
				DagInstanceID: dagIns.ID,
				Types: []rds.DagInstanceEventType{
					rds.DagInstanceEventTypeTaskStatus,
					rds.DagInstanceEventTypeVariable,
					rds.DagInstanceEventTypeTrace,
				},
			})
			if err != nil {
				log.Warnf("[logic.GenerateTaskResults] list events err, detail: %s", err.Error())
				return 0, nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.InternalError, err.Error())
			}
			allTasks = buildTaskInstanceFromEvents(events, dagIns, dag)
			m.memoryCache.Set(dagInsID, allTasks, time.Minute*5)
		case entity.DagInstanceEventPersistenceOss:
			events, err := dagIns.ListOssEvents(ctx)

			if err != nil {
				log.Warnf("[logic.GenerateTaskResults] list events err, detail: %s", err.Error())
				return 0, nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.InternalError, err.Error())
			}

			allTasks = buildTaskInstanceFromEvents(events, dagIns, dag)
			m.memoryCache.Set(dagInsID, allTasks, time.Minute*5)
		default:
			err = dagIns.LoadExtData(ctx)
			if err != nil {
				log.Warnf("[logic.GenerateTaskResults] LoadExtData err, detail: %s", err.Error())
				return 0, nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.InternalError, err.Error())
			}

			for _, task := range dag.Tasks {
				if task.ActionName == common.InternalReturnOpt {
					allTasks = append(allTasks, m.createBaseTaskInstance(task, dagIns))
					break
				}
				// 循环节点内的step信息不会在任务执行前初始化好，需要动态生成
				if task.ActionName == common.Loop {
					isRetrun, tasks := m.generateLoopTaskResults(task, dagIns)
					allTasks = append(allTasks, tasks...)
					if isRetrun {
						break
					}
				} else {
					baseTask := m.createBaseTaskInstance(task, dagIns)
					allTasks = append(allTasks, baseTask)
				}
			}

			sort.SliceStable(allTasks, func(i, j int) bool {
				return allTasks[i].LastModifiedAt < allTasks[j].LastModifiedAt
			})
			m.memoryCache.Set(dagInsID, allTasks, time.Minute*5)
		}
	}

	total := int64(len(allTasks))

	if limit == -1 {
		return total, allTasks, nil
	}

	start, end := page*limit, (page+1)*limit
	if start > total {
		start = total
	}

	if end > total {
		end = total
	}

	allTasks = allTasks[start:end]

	return total, allTasks, nil
}

// generateLoopTaskResults 生成循环节点结果
func (m *mgnt) generateLoopTaskResults(task entity.Task, dagIns *entity.DagInstance) (bool, []*entity.TaskInstance) {
	allTasks := make([]*entity.TaskInstance, 0)
	mode := task.Params["mode"].(string)
	iterations, loopParam := 0, &LoopParam{LoopTaskID: task.ID, Index: 0, StepMap: make(map[string]string)}
	switch mode {
	case "limit":
		loopTaskIns := m.createBaseTaskInstance(task, dagIns)
		switch v := loopTaskIns.Params["limit"].(type) {
		case int32:
			iterations = int(v)
		case int64:
			iterations = int(v)
		case int8:
			iterations = int(v)
		case int:
			iterations = v
		case float64:
			iterations = int(v)
		case string:
			iterations, _ = strconv.Atoi(v)
		}
	case "array":
		loopTaskIns := m.createBaseTaskInstance(task, dagIns)
		if v, ok := loopTaskIns.Params["array"].([]any); ok {
			iterations = len(v)
			loopParam.LoopItems = v
		} else {
			commonLog.NewLogger().Warnln("[logic.generateLoopTaskResults] parse loop params failed, not array")
		}
	}

	var loopOutputs map[string]any
	for i := range iterations {
		loopParam.Index = i
		// loop 节点循环结果, 生成副本数据防止原数据被破坏
		b, _ := json.Marshal(task)
		var taskCopy entity.Task
		_ = json.Unmarshal(b, &taskCopy)

		if i > 0 {
			taskCopy.ID = fmt.Sprintf("%s_i%d", task.ID, i)
		}
		baseTask := m.createBaseTaskInstance(taskCopy, dagIns)

		// 循环节点输出结果不全，手动补全
		result, ok := baseTask.Results.(map[string]any)
		if ok {
			if loopOutputs == nil && result["outputs"] != nil {
				loopOutputs = result["outputs"].(map[string]any)
			}

			switch mode {
			case "limit":
				result["index"] = i
			case "array":
				content := baseTask.Params["array"].([]any)
				result["index"] = i
				result["value"] = content[i]
			}

			if i == iterations-1 && loopOutputs != nil {
				result["outputs"] = loopOutputs
			} else {
				delete(result, "outputs")
			}
			baseTask.Results = result
		}
		allTasks = append(allTasks, baseTask)

		// loop内部step创建
		for _, step := range task.Steps {
			if step.Operator == common.BranchOpt || step.Operator == common.ControlFlowParallel {
				breanchTaskID := fmt.Sprintf("%s_i%v_s%v", task.ID, i, step.ID)
				isReturn, tasks := m.generateBranchTaskResults(breanchTaskID, step, dagIns, loopParam, false)
				allTasks = append(allTasks, tasks...)
				if isReturn {
					return true, allTasks
				}
			} else {
				loopTask := entity.Task{
					ID:         fmt.Sprintf("%s_i%v_s%v", task.ID, i, step.ID),
					Name:       step.Title,
					ActionName: step.Operator,
					Params:     m.renderLoopBranchParams(step.Parameters, loopParam, dagIns),
				}
				loopParam.StepMap[step.ID] = loopTask.ID
				allTasks = append(allTasks, m.createBaseTaskInstance(loopTask, dagIns))
				if step.Operator == common.InternalReturnOpt {
					return true, allTasks
				}
			}
		}
	}

	return false, allTasks
}

// generateBranchTaskResults 生成分支任务结果
func (m *mgnt) generateBranchTaskResults(taskID string, step entity.Step, dagIns *entity.DagInstance, loopParam *LoopParam, isSkip bool) (bool, []*entity.TaskInstance) {
	var tasks []*entity.TaskInstance

	// 分支节点
	breanchTask := createIterationBranchTaskInstance(taskID, step, dagIns, isSkip)
	tasks = append(tasks, breanchTask)

	for branchIndex, branch := range step.Branches {
		branch := utils.DeepCopy(branch)
		for _, branchStep := range branch.Steps {
			prechecks := entity.PreChecks{}
			for idx, con := range branch.Conditions {
				conMap := map[string]any{}
				for i, v := range con {
					conMap[fmt.Sprintf("%v", i)] = map[string]any{
						"a": v.Parameter.A,
						"b": v.Parameter.B,
					}
				}
				conMap = m.renderLoopBranchParams(conMap, loopParam, dagIns)
				for i := range len(conMap) {
					key := fmt.Sprintf("%v", i)
					con[i].Parameter = entity.TaskConditionParameter{
						A: conMap[key].(map[string]any)["a"],
						B: conMap[key].(map[string]any)["b"],
					}
				}

				check := &entity.Check{
					Conditions: con,
					Act:        entity.ActiveActionSkip,
				}
				prechecks[fmt.Sprintf("%s_%v", taskID, idx)] = check
			}
			task := entity.Task{
				Name:       branchStep.Title,
				ActionName: branchStep.Operator,
				Params:     m.renderLoopBranchParams(branchStep.Parameters, loopParam, dagIns),
				PreChecks:  prechecks,
			}

			if branchStep.Operator == common.InternalReturnOpt {
				tasks = append(tasks, m.createBaseTaskInstance(task, dagIns))
				return true, tasks
			}
			if branchStep.Operator == common.BranchOpt {
				isSkip = tasks[len(tasks)-1].Status == entity.TaskInstanceStatusSkipped
				branchTaskID := buildBranchTaskID(dagIns.Mode, branchIndex, loopParam.Index, branchStep.ID, taskID)
				isReturn, subTasks := m.generateBranchTaskResults(branchTaskID, branchStep, dagIns, loopParam, isSkip)
				tasks = append(tasks, subTasks...)
				if isReturn {
					return true, tasks
				}
			} else {
				task.ID = buildBranchTaskID(dagIns.Mode, branchIndex, loopParam.Index, branchStep.ID, taskID)
				loopParam.StepMap[branchStep.ID] = task.ID
				branchStepTask := m.createBaseTaskInstance(task, dagIns)
				tasks = append(tasks, branchStepTask)
			}
		}
	}

	return false, tasks
}

// createBaseTaskInstance 创建节点实例
func (m *mgnt) createBaseTaskInstance(task entity.Task, dagIns *entity.DagInstance) *entity.TaskInstance {
	b, _ := json.Marshal(task)
	var taskCopy entity.Task
	_ = json.Unmarshal(b, &taskCopy)

	taskIns := &entity.TaskInstance{
		TaskID:             taskCopy.ID,
		DagInsID:           dagIns.ID,
		Name:               taskCopy.Name,
		ActionName:         taskCopy.ActionName,
		Params:             taskCopy.Params,
		Status:             entity.TaskInstanceStatusSuccess,
		PreChecks:          taskCopy.PreChecks,
		Results:            extractTaskResults(taskCopy.ID, dagIns),
		LastModifiedAt:     dagIns.EndedAt * 1e9,
		RelatedDagInstance: dagIns,
		MetaData: &entity.TaskMetaData{
			StartedAt: dagIns.CreatedAt * 1e3,
		},
	}
	taskIns.Initial()
	taskIns.CreatedAt = dagIns.CreatedAt
	taskIns.UpdatedAt = dagIns.EndedAt
	taskIns.DoPreCheck(dagIns)
	if taskIns.Status == entity.TaskInstanceStatusSkipped {
		taskIns.LastModifiedAt = time.Now().AddDate(200, 0, 0).UnixNano()
	}

	m.renderParamsV2(taskIns)
	if taskCopy.ActionName == common.InternalReturnOpt {
		taskIns.Results = taskIns.Params
	}
	// 组合算子结束节点无输出
	if dagIns.DagType == common.DagTypeComboOperator && taskCopy.ActionName == common.InternalReturnOpt {
		taskIns.Params = nil
	}
	return taskIns
}

// createIterationBranchTaskInstance 创建迭代分支节点
func createIterationBranchTaskInstance(taskID string, step entity.Step, dagIns *entity.DagInstance, isSkip bool) *entity.TaskInstance {
	taskIns := &entity.TaskInstance{
		TaskID:         taskID,
		Name:           step.Title,
		ActionName:     step.Operator,
		Status:         entity.TaskInstanceStatusSuccess,
		LastModifiedAt: dagIns.EndedAt * 1e9,
		MetaData: &entity.TaskMetaData{
			StartedAt: dagIns.CreatedAt * 1e3,
		},
	}
	taskIns.Initial()
	taskIns.CreatedAt = dagIns.CreatedAt
	taskIns.UpdatedAt = dagIns.EndedAt
	if isSkip {
		taskIns.Status = entity.TaskInstanceStatusSkipped
	}

	return taskIns
}

// extractTaskResults 根据运行模式提取任务结果
func extractTaskResults(taskID string, dagIns *entity.DagInstance) any {
	shareData := parseShareData(dagIns)
	taskKey := fmt.Sprintf("__%v", taskID)
	vars, _ := shareData.Get(taskKey)
	return vars
}

func parseShareData(dagIns *entity.DagInstance) *entity.ShareData {
	var shareData = &entity.ShareData{}
	switch dagIns.Mode {
	case entity.DagInstanceModeVM:
		data := map[string]any{}
		_ = json.Unmarshal([]byte(dagIns.Dump), &data)
		shareData.Dict = data["env"].(map[string]any)
	default:
		shareData = dagIns.ShareData
	}

	return shareData
}

// buildBranchTaskID 构建分支任务ID
func buildBranchTaskID(mode entity.DagInstanceMode, branchIndex, iter int, branchStepID, taskID string) string {
	if mode == entity.DagInstanceModeVM {
		loopTaskID := taskID
		if idx := strings.Index(taskID, "_i"); idx != -1 {
			loopTaskID = taskID[:idx]
		}
		return fmt.Sprintf("%s_i%v_s%v", loopTaskID, iter, branchStepID)
	}

	// 统一处理非VM模式
	return fmt.Sprintf("%s_%v_%v", taskID, branchIndex, branchStepID)
}

// renderLoopBranchParams 渲染引用循环节点Value和Index节点入参
func (m *mgnt) renderLoopBranchParams(params map[string]any, loopParam *LoopParam, dagIns *entity.DagInstance) map[string]any {
	b, _ := json.Marshal(params)
	jsonStr := string(b)

	loopOutputPattern := `{{__(\d+)\.([^}]+)}}`
	loopOutputRe := regexp.MustCompile(loopOutputPattern)
	// 处理循环内节点输出结果变量引用
	loopOutputMatches := loopOutputRe.FindAllStringSubmatch(jsonStr, -1)
	if len(loopOutputMatches) > 0 {
		for _, match := range loopOutputMatches {
			if len(match) >= 3 {
				loopOutputKey := match[2]
				if v, ok := loopParam.StepMap[match[1]]; ok {
					loopOutputKey = fmt.Sprintf("{{__%v.%v}}", v, loopOutputKey)
					jsonStr = strings.ReplaceAll(jsonStr, match[0], loopOutputKey)
				}
			}
		}
	}

	copyParams := make(map[string]any)
	_ = json.Unmarshal([]byte(jsonStr), &copyParams)

	vmIns := vm.NewVM()
	vmIns.AddGlobals(mod.NewGlobals(dagIns))

	g := vm.NewGenerator(vmIns)

	err := g.GenerateValue(copyParams)
	if err != nil {
		commonLog.NewLogger().Warnf("[logic.renderLoopBranchParams] GenerateValue failed, detail: %v", err.Error())
		return params
	}

	vmIns.LoadInstructions(g.Instructions)

	env := parseShareData(dagIns).GetAll()

	vmIns.Env = env

	pattern := `{{__(\d+)\.(value|index)}}`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(jsonStr, -1)
	if len(matches) > 0 {
		var arrVal any
		if loopParam.Index < len(loopParam.LoopItems) {
			arrVal = loopParam.LoopItems[loopParam.Index]
		} else {
			arrVal = ""
		}
		vmIns.Env["__"+loopParam.LoopTaskID] = map[string]any{
			"value": arrVal,
			"index": loopParam.Index,
		}
	}

	vmIns.Run()

	_, ret, err := vmIns.Result()

	if err != nil {
		commonLog.NewLogger().Warnf("[logic.renderLoopBranchParams] vmIns.Result failed, detail: %v", err.Error())
		return params
	}

	if resultMap, ok := ret.(map[string]interface{}); ok {
		maps.Copy(copyParams, resultMap)
	}

	return copyParams
}

func buildStepMap(steps []entity.Step, stepMap map[string]*entity.Step) {
	for _, step := range steps {
		switch step.Operator {
		case common.Loop:
			stepMap[step.ID] = &step
			buildStepMap(step.Steps, stepMap)
		case common.BranchOpt, common.ControlFlowParallel:
			for _, branch := range step.Branches {
				buildStepMap(branch.Steps, stepMap)
			}
		default:
			stepMap[step.ID] = &step
		}
	}
}

func buildTaskInstanceFromEvents(events []*entity.DagInstanceEvent, dagIns *entity.DagInstance, dag *entity.Dag) (tasks []*entity.TaskInstance) {
	var (
		allTasksMap = make(map[string]*entity.TaskInstance)
		env         = make(map[string]any)
		stepMap     = make(map[string]*entity.Step)
	)

	buildStepMap(dag.Steps, stepMap)

	for _, event := range events {
		switch event.Type {
		case rds.DagInstanceEventTypeVariable:
			env[event.Name] = event.Data
			stepId := event.Name[2:]
			// 尝试更新Loop节点的输出结果
			// 注意：对于循环节点，这里总是更新该TaskID对应的"最新"一个实例
			if step, ok := stepMap[stepId]; ok && step.Operator == common.Loop {
				if data, ok := event.Data.(map[string]any); ok {
					if _, ok := data["outputs"]; ok {
						if taskIns, ok := allTasksMap[stepId]; ok {
							resultsMap, isMap := taskIns.Results.(map[string]any)
							if resultsMap == nil || !isMap {
								resultsMap = make(map[string]any)
							}

							for k, v := range data {
								resultsMap[k] = v
							}

							taskIns.Results = resultsMap
						}
					}
				}
			}
		case rds.DagInstanceEventTypeTaskStatus:
			current, exists := allTasksMap[event.TaskID]
			isNewInstance := false

			// 检测是否需要创建新实例：
			// 1. Map中不存在该TaskID
			// 2. Map中存在，但已处于最终状态，且当前事件为非最终状态（循环节点的下一次迭代复用ID）
			if !exists {
				isNewInstance = true
			} else {
				isFinished := current.Status == entity.TaskInstanceStatusSuccess ||
					current.Status == entity.TaskInstanceStatusFailed ||
					current.Status == entity.TaskInstanceStatusSkipped

				// 如果是开始类状态（Running/Blocked/Pending），且旧实例已完成，视为新迭代
				isStarting := event.Status == "running" || event.Status == "blocked" || event.Status == "pending"

				if isFinished && isStarting {
					isNewInstance = true
				}
			}

			if isNewInstance {
				timestampSec := event.Timestamp / 1e6
				current = &entity.TaskInstance{
					BaseInfo: entity.BaseInfo{
						ID:        fmt.Sprintf("%d", len(tasks)),
						CreatedAt: timestampSec,
						UpdatedAt: timestampSec,
					},
					TaskID:         event.TaskID,
					DagInsID:       dagIns.ID,
					ActionName:     event.Operator,
					Status:         entity.TaskInstanceStatus(event.Status),
					LastModifiedAt: event.Timestamp,
					MetaData: &entity.TaskMetaData{
						StartedAt: event.Timestamp / 1e3,
					},
				}

				if step, ok := stepMap[event.TaskID]; ok {
					current.Name = step.Title

					switch step.Operator {
					case common.InternalAssignOpt:
						target := step.Parameters["target"]
						value, valueOk := step.Parameters["value"]
						if valueOk {
							value = renderParams(value, env)
						}
						current.Params = map[string]any{
							"target": target,
							"value":  value,
						}
					case common.BranchOpt:
					default:
						realParam := renderParams(step.Parameters, env)
						if p, ok := realParam.(map[string]any); ok {
							current.Params = p
						} else {
							current.Params = step.Parameters
						}
					}
				}

				// 加入列表和Map
				tasks = append(tasks, current)
				allTasksMap[event.TaskID] = current
			} else {
				// 更新现有实例
				current.Status = entity.TaskInstanceStatus(event.Status)
				current.UpdatedAt = event.Timestamp / 1e6
				current.LastModifiedAt = current.UpdatedAt
				// 注意：这里更新MetaData.ElapsedTime依赖于正确的StartedAt，而StartedAt是在创建时设置的
				current.MetaData.ElapsedTime = event.Timestamp/1e3 - current.MetaData.StartedAt
			}

			// 处理完成状态的Result/Reason更新
			switch current.Status {
			case entity.TaskInstanceStatusSuccess:
				if result, ok := env[fmt.Sprintf("__%s", current.TaskID)]; ok {
					current.Results = result
				}
			case entity.TaskInstanceStatusFailed:
				current.Reason = event.Data
			}

		case rds.DagInstanceEventTypeTrace:
			if current, ok := allTasksMap[event.TaskID]; ok {
				dataBytes, _ := json.Marshal(event.Data)
				_ = json.Unmarshal(dataBytes, current.MetaData)
			}
		}
	}

	return
}

func renderParams(params any, env map[string]any) any {
	vmIns := vm.NewVM()
	g := vm.NewGenerator(vmIns)
	err := g.GenerateValue(params)
	if err != nil {
		return params
	}
	vmIns.LoadInstructions(g.Instructions)
	vmIns.Env = env
	vmIns.Run()
	_, ret, err := vmIns.Result()

	if err != nil {
		return params
	}

	return ret
}
