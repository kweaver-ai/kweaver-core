package mod

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/dependency"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/policy"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/render"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/utils/value"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"

	"encoding/json"

	ierrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/actions"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/mitchellh/mapstructure"
)

const (
	ReasonSuccessAfterCanceled = "success after canceled"
	ReasonParentCancel         = "parent success but already be canceled"
)

var queueTypes = []string{
	common.PriorityLowest,
	common.PriorityLow,
	common.PriorityMedium,
	common.PriorityHigh,
	common.PriorityHighest,
}

// DefExecutor
type DefExecutor struct {
	ctx           context.Context
	cancelMap     sync.Map
	workerNumbers []int
	// workerQueue  chan *entity.TaskInstance
	workerQueues map[string]chan *entity.TaskInstance
	workerWg     sync.WaitGroup
	initWg       sync.WaitGroup
	timeout      time.Duration
	initQueues   map[string]chan *initPayload
	// initQueue    chan *initPayload

	paramRender *render.TplRender

	closeCh chan struct{}
	lock    sync.RWMutex
	// 新增：用于同一个taskIns的更新加锁
	taskInsLocks sync.Map // key: taskIns.ID, value: *sync.Mutex
}

// initPayload
type initPayload struct {
	dagIns  *entity.DagInstance
	taskIns *entity.TaskInstance
}

// NewDefExecutor
func NewDefExecutor(timeout time.Duration, workers ...int) *DefExecutor {
	return &DefExecutor{
		ctx:           context.Background(),
		workerNumbers: workers,
		// workerQueue:  make(chan *entity.TaskInstance),
		workerQueues: make(map[string]chan *entity.TaskInstance),
		timeout:      timeout,
		initQueues:   make(map[string]chan *initPayload),
		// initQueue:    make(chan *initPayload),
		closeCh:     make(chan struct{}, 1),
		paramRender: render.NewTplRender(),
	}
}

// Init
func (e *DefExecutor) Init() {
	for index, queueType := range queueTypes {
		e.initWg.Add(1)
		e.initQueues[queueType] = make(chan *initPayload)
		e.workerQueues[queueType] = make(chan *entity.TaskInstance)
		go e.watchInitQueue(e.initQueues[queueType])
		for i := 0; i < e.workerNumbers[index]; i++ {
			e.workerWg.Add(1)
			go e.subWorkerQueue(e.workerQueues[queueType])
		}
	}
}

func (e *DefExecutor) subWorkerQueue(queue chan *entity.TaskInstance) {
	for taskIns := range queue {
		e.workerDo(taskIns)
	}
	e.workerWg.Done()
}

// CancelTaskIns
func (e *DefExecutor) CancelTaskIns(taskInsIds []string) error {
	for _, id := range taskInsIds {
		if cancel, ok := e.cancelMap.Load(id); ok {
			e.cancelMap.Delete(id)
			cancel.(context.CancelFunc)()
		}
	}

	return nil
}

func (e *DefExecutor) watchInitQueue(queue chan *initPayload) {
	for p := range queue {
		e.initWorkerTask(p.dagIns, p.taskIns)
	}
	e.initWg.Done()
}

func (e *DefExecutor) initWorkerTask(dagIns *entity.DagInstance, taskIns *entity.TaskInstance) {
	if _, ok := e.cancelMap.Load(taskIns.ID); ok {
		return
	}

	defTimeout := e.timeout
	if taskIns.TimeoutSecs != 0 {
		defTimeout = time.Duration(taskIns.TimeoutSecs) * time.Second
	}

	dag := func(ctx context.Context, id, versionID string) (*entity.Dag, error) {
		return GetStore().GetDagWithOptionalVersion(ctx, id, versionID)
	}

	patchDagIns := func(ctx context.Context, dagIns *entity.DagInstance, mustsPatchFields ...string) error {
		return GetStore().PatchDagIns(ctx, dagIns, mustsPatchFields...)
	}

	c, cancel := context.WithTimeout(context.TODO(), defTimeout)
	dagIns.ShareData.Save = func(data *entity.ShareData) error {
		patch := &entity.DagInstance{BaseInfo: entity.BaseInfo{ID: taskIns.DagInsID},
			DagID:            dagIns.DagID,
			ShareData:        data,
			EventPersistence: dagIns.EventPersistence,
		}
		if err := patch.SaveExtData(context.Background()); err != nil {
			return err
		}
		return GetStore().PatchDagIns(context.Background(), patch)
	}
	executeMethods := entity.ExecuteMethods{
		Publish:     NewMQHandler().Publish,
		GetDag:      dag,
		PatchDagIns: patchDagIns,
	}
	taskIns.InitialDep(
		entity.NewDefExecuteContext(c, dagIns.ShareData, taskIns.Trace, dagIns.VarsGetter(), dagIns.VarsIterator(), taskIns.ParamsGetter(), taskIns.GetGraphID(), taskIns, drivenadapters.NewEfast(), executeMethods, dependency.NewDriven()),
		func(ctx context.Context, instance *entity.TaskInstance) error {
			return GetStore().PatchTaskIns(ctx, instance)
		}, dagIns)
	e.cancelMap.Store(taskIns.ID, cancel)
	priority := dagIns.Priority
	if priority != "" {
		e.workerQueues[dagIns.Priority] <- taskIns
	} else {
		e.workerQueues[common.PriorityLowest] <- taskIns
	}
}

// Push task to execute
func (e *DefExecutor) Push(dagIns *entity.DagInstance, taskIns *entity.TaskInstance) {
	isActive, err := taskIns.DoPreCheck(dagIns)
	if err != nil {
		log.Errorf("do task pre-check failed:%s", err)
		return
	}

	if isActive {
		if err := GetStore().PatchTaskIns(context.Background(), &entity.TaskInstance{
			BaseInfo: taskIns.BaseInfo,
			Status:   taskIns.Status,
		}); err != nil {
			log.Errorf("patch task[%s] failed: %s", taskIns.ID, err)
			return
		}

		// if pre-check is active, we should not execute task
		GetParser().EntryTaskIns(taskIns)
		return
	}

	e.lock.RLock()
	defer e.lock.RUnlock()

	// try to exit the sender goroutine as early as possible.
	// try-receive and try-send select blocks are specially optimized by the standard Go compiler,
	// so they are very efficient.
	select {
	case <-e.closeCh:
		log.Info("parser has already closed, so will not execute next task instances")
		return
	default:
	}

	priority := dagIns.Priority

	if priority != "" {
		// init task in single queue to prevent double check map
		e.initQueues[priority] <- &initPayload{
			dagIns:  dagIns,
			taskIns: taskIns,
		}
	} else {
		// init task in single queue to prevent double check map
		e.initQueues[common.PriorityLowest] <- &initPayload{
			dagIns:  dagIns,
			taskIns: taskIns,
		}
	}
}

func (e *DefExecutor) workerDo(taskIns *entity.TaskInstance) {
	var err error
	ctx, span := trace.StartInternalSpan(e.ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	switch taskIns.Status {
	case entity.TaskInstanceStatusInit, entity.TaskInstanceStatusEnding, entity.TaskInstanceStatusRetrying:
	case entity.TaskInstanceStatusFailed, entity.TaskInstanceStatusCanceled:
		// 允许Loop任务在失败或取消状态下进入runAction，以便在那里的特殊逻辑重置状态
		if taskIns.ActionName == common.Loop {
			break
		}
		fallthrough
	default:
		log.Warnf("this task instance[%s] is not executable, status[%s]", taskIns.ID, taskIns.Status)
		e.cancelMap.Delete(taskIns.ID)
		GetParser().EntryTaskIns(taskIns)
		return
	}

	// 执行任务
	err = e.runAction(ctx, taskIns)

	e.handleTaskError(ctx, taskIns, err)
	e.cancelMap.Delete(taskIns.ID)

	GetParser().EntryTaskIns(taskIns)
}

// serializeRenderedParams converts RenderedParams to a MongoDB-serializable format
func (e *DefExecutor) serializeRenderedParams(params interface{}) map[string]interface{} {
	if params == nil {
		return nil
	}

	// Convert to map[string]interface{} for MongoDB compatibility
	jsonData, err := json.Marshal(params)
	if err != nil {
		log.Errorf("Failed to marshal rendered params: %v", err)
		// If marshaling fails, set to nil to avoid MongoDB errors
		return nil
	}

	var mapData map[string]interface{}
	if err := json.Unmarshal(jsonData, &mapData); err != nil {
		log.Errorf("Failed to unmarshal rendered params: %v", err)
		return nil
	}

	return mapData
}

func (e *DefExecutor) runAction(ctx context.Context, taskIns *entity.TaskInstance) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	// 特殊处理Loop任务，记录开始状态
	if taskIns.ActionName == common.Loop {
		// 如果循环任务之前是failed或canceled状态，重置它为init状态再次尝试
		if taskIns.Status == entity.TaskInstanceStatusFailed ||
			taskIns.Status == entity.TaskInstanceStatusCanceled {
			prevStatus := taskIns.Status
			taskIns.Status = entity.TaskInstanceStatusInit
			now := time.Now().UnixNano()
			taskIns.UpdatedAt = now
			log.Infof("循环任务 [%s] 状态从 %s 重置为 %s 以允许重新执行",
				taskIns.ID, prevStatus, taskIns.Status)

			// 保存状态变更
			if err := GetStore().PatchTaskIns(ctx, &entity.TaskInstance{
				BaseInfo: taskIns.BaseInfo,
				Status:   taskIns.Status,
				Reason:   "重置循环任务以继续执行",
			}); err != nil {
				log.Warnf("保存循环任务状态重置失败: %v", err)
			} else {
				log.Infof("成功重置循环任务 [%s] 为init状态", taskIns.ID)
			}
		}
	}

	if strings.HasPrefix(taskIns.ActionName, "@custom/") {

		segment := strings.Split(taskIns.ActionName, "/")

		if len(segment) != 3 {
			return fmt.Errorf("invalid action: %s", taskIns.ActionName)
		}

		executorID, err := strconv.ParseUint(segment[1], 10, 64)

		if err != nil {
			return fmt.Errorf("invalid action: %s", taskIns.ActionName)
		}

		actionID, err := strconv.ParseUint(segment[2], 10, 64)

		if err != nil {
			return fmt.Errorf("invalid action: %s", taskIns.ActionName)
		}

		parameters := make(map[string]interface{}, 0)

		act := &actions.CustomAction{
			ExecutorID: executorID,
			ActionID:   actionID,
			Parameters: parameters,
		}

		if taskIns.GetParams() != nil {
			if err := e.getFromTaskInstance(taskIns, &parameters); err != nil {
				return fmt.Errorf("get task params from task instance failed: %w", err)
			}
			serializableParams := e.serializeRenderedParams(parameters)
			err = taskIns.Patch(ctx, &entity.TaskInstance{
				BaseInfo:       taskIns.BaseInfo,
				RenderedParams: serializableParams,
			})
			if err != nil {
				return fmt.Errorf("patch task instance failed: %w", err)
			}
		}

		return e.runActionWithRetry(ctx, taskIns, act, act)
	}

	if strings.HasPrefix(taskIns.ActionName, "@operator/") {
		act := &actions.ComboOperator{}
		if taskIns.GetParams() != nil {
			if err := e.getFromTaskInstance(taskIns, &act); err != nil {
				return fmt.Errorf("get task params from task instance failed: %w", err)
			}
		}
		act.Operator = taskIns.ActionName
		serializableParams := e.serializeRenderedParams(act)
		err = taskIns.Patch(ctx, &entity.TaskInstance{
			BaseInfo:       taskIns.BaseInfo,
			RenderedParams: serializableParams,
		})
		if err != nil {
			return fmt.Errorf("patch task instance failed: %w", err)
		}
		return e.runActionWithRetry(ctx, taskIns, act, act)
	}

	if strings.HasPrefix(taskIns.ActionName, "@trigger/operator/") {
		act := &actions.TriggerOperator{}
		if taskIns.GetParams() != nil {
			if err := e.getFromTaskInstance(taskIns, &act); err != nil {
				return fmt.Errorf("get task params from task instance failed: %w", err)
			}
		}
		act.Operator = taskIns.ActionName
		serializableParams := e.serializeRenderedParams(act)
		err = taskIns.Patch(ctx, &entity.TaskInstance{
			BaseInfo:       taskIns.BaseInfo,
			RenderedParams: serializableParams,
		})
		if err != nil {
			return fmt.Errorf("patch task instance failed: %w", err)
		}
		return e.runActionWithRetry(ctx, taskIns, act, act)
	}

	act := ActionMap[taskIns.ActionName]
	if act == nil {
		return fmt.Errorf("action not found: %s", taskIns.ActionName)
	}

	if taskIns.GetParams() == nil {
		return e.runActionWithRetry(ctx, taskIns, nil, act)
	}
	paramAct, ok := act.(entity.ParameterAction)
	if !ok {
		return e.runActionWithRetry(ctx, taskIns, nil, act)
	}
	p := paramAct.ParameterNew()
	if p == nil {
		return e.runActionWithRetry(ctx, taskIns, nil, act)
	}
	if err := e.getFromTaskInstance(taskIns, p); err != nil {
		return fmt.Errorf("get task params from task instance failed: %w", err)
	}

	// Serialize RenderedParams to MongoDB-compatible format
	serializableParams := e.serializeRenderedParams(p)
	err = taskIns.Patch(ctx, &entity.TaskInstance{
		BaseInfo:       taskIns.BaseInfo,
		RenderedParams: serializableParams,
	})
	if err != nil {
		return fmt.Errorf("patch task instance failed: %w", err)
	}
	return e.runActionWithRetry(ctx, taskIns, p, act)
}

func (e *DefExecutor) runActionWithRetry(ctx context.Context, taskIns *entity.TaskInstance, params interface{}, act entity.Action) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if act.Name() == common.Loop {
		loopLock := e.getTaskInsLock(taskIns.ID)
		loopLock.Lock()
		defer loopLock.Unlock()
		// Parse loop parameters from the task instance
		var loopParams LoopParameters
		if err := e.getFromTaskInstance(taskIns, &loopParams); err != nil {
			log.Errorf("获取循环参数失败 [%s]: %v", taskIns.ID, err)
			return fmt.Errorf("get loop params from task instance failed: %w", err)
		}

		return NewLoopHandler(loopParams).HandleLoopAction(ctx, taskIns, params, act)
	}

	var opts []policy.Option
	var skipPolicyMap = map[string]bool{
		common.BranchOpt: true,
	}

	actName := act.Name()
	start := time.Now()

	if _, ok := skipPolicyMap[actName]; !ok {
		config := utils.IfNot(taskIns.Settings != nil && !reflect.DeepEqual(taskIns.Settings, &entity.Settings{}), taskIns.Settings, &entity.Settings{
			Retry: &entity.RetryConfig{
				Max:   3,
				Delay: 3,
			},
			TimeOut: &entity.TimeoutConfig{
				// 默认配置 节点配置的超时时间超过24小时，最大不超过30分钟, 否则使用节点配置的超时时间（减60是因为在配置时手动增加了60秒）
				Delay: utils.IfNot(taskIns.TimeoutSecs > 24*60*60 || taskIns.TimeoutSecs <= 0, 30*60, taskIns.TimeoutSecs-60),
			},
		})
		opts = append(opts, policy.WithTimeout(config.TimeOut.Delay))
		opts = append(opts, policy.WithRetry(config.Retry.Max, config.Retry.Delay, func(err error) bool {
			if err != nil {
				if act.Name() == common.InternalToolPy3Opt ||
					act.Name() == "@custom" ||
					ierrors.Is(err, ierrors.Forbidden, ierrors.NonProcessManager) {
					return false
				}
				return true
			}
			return true
		}))
	}

	// 执行前置操作，记录开始时间
	opts = append(opts, policy.WithBeforeExecute(func(ctx context.Context) {
		metadata := &entity.TaskMetaData{
			StartedAt: start.UnixMilli(),
		}
		if pErr := taskIns.Patch(ctx, &entity.TaskInstance{
			BaseInfo: taskIns.BaseInfo,
			MetaData: metadata}); pErr != nil {
			log.Warnf("[mod.RunActionWithRetry] dag instance %v, before execute %s update task metadata failed, detail: %s", taskIns.RelatedDagInstance.ID, actName, pErr.Error())
		}
	}))

	// 失败时，日志输出
	opts = append(opts, policy.WithOnError(func(ctx context.Context, err error) {
		if err != nil {
			log.Warnf("[mod.RunActionWithRetry] dag instance %v, execute %s failed, detail: %s", taskIns.RelatedDagInstance.ID, actName, err.Error())
		}
	}))

	// 执行后置操作，记录节点执行过程性信息
	opts = append(opts, policy.WithAfterExecute(func(ctx context.Context, collector *policy.ResultCollector) {
		var metadata *entity.TaskMetaData
		startSec, endSec := start.UnixMilli(), time.Now().UnixMilli()
		retry, ok := policy.GetAs[*policy.RetryData](collector, policy.RetryPolicyName)
		if ok {
			key := fmt.Sprintf("__%s_trace", taskIns.TaskID)
			trace := map[string]any{
				key: map[string]any{
					// "started_at": startSec,
					// "ended_at":   endSec,
					"attempts":  retry.Data.Attempts,
					"max_retry": retry.Data.Max,
					"duration":  collector.Duration,
				},
			}

			err := taskIns.RelatedDagInstance.WriteTraceEvent(ctx, trace)
			if err != nil {
				log.Warnf("[mod.RunActionWithRetry] dag instance %v, after execute %s write trace event failed, detail: %s", taskIns.RelatedDagInstance.ID, actName, err.Error())
			}

			metadata = &entity.TaskMetaData{
				Attempts:    retry.Data.Attempts,
				MaxRetry:    retry.Data.Max,
				StartedAt:   startSec,
				Duration:    collector.Duration,
				ElapsedTime: endSec - startSec,
			}
		} else {
			metadata = &entity.TaskMetaData{
				StartedAt:   startSec,
				Duration:    collector.Duration,
				ElapsedTime: endSec - startSec,
			}
		}

		if pErr := taskIns.Patch(ctx, &entity.TaskInstance{
			BaseInfo: taskIns.BaseInfo,
			MetaData: metadata}); pErr != nil {
			log.Warnf("[mod.RunActionWithRetry] dag instance %v, after execute %s update task metadata failed, detail: %s", taskIns.RelatedDagInstance.ID, actName, pErr.Error())
		}

	}))

	p := policy.NewComposite(opts...)

	err = p.Do(ctx, func(ctx context.Context) error {
		var tokenInfo *entity.Token
		var gErr error
		tokenInfo, gErr = NewTokenStrategyManager().GetToken(&TokenContext{
			TaskIns: taskIns,
			ActName: act.Name(),
			Ctx:     ctx,
		})

		if gErr != nil {
			return gErr
		}

		return taskIns.Run(ctx, params, act, tokenInfo)
	})

	return ierrors.ParseError(err)
}

func (e *DefExecutor) getFromTaskInstance(taskIns *entity.TaskInstance, params interface{}) error {
	err := e.renderParamsV2(taskIns)
	if err != nil {
		return fmt.Errorf("renderParams failed: %w", err)
	}

	// 创建 taskIns.Params 的安全副本，避免并发 map 访问
	var safeParams map[string]interface{}
	paramsCopy := taskIns.GetParams()
	if paramsCopy != nil {
		safeParams = make(map[string]interface{})
		for k, v := range paramsCopy {
			safeParams[k] = v
		}
	}

	return weakDecode(safeParams, params)
}

func weakDecode(input, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Metadata:         nil,
		Result:           output,
		TagName:          "json", // Find target field using json tag
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func (e *DefExecutor) renderParamsV2(taskIns *entity.TaskInstance) error {

	vmIns := vm.NewVM()
	vmIns.AddGlobals(NewGlobals(taskIns.RelatedDagInstance))

	g := vm.NewGenerator(vmIns)

	rawParams := make(map[string]any)
	params := taskIns.GetParams()

	switch taskIns.ActionName {
	case common.InternalAssignOpt:
		rawParams["target"] = params["target"]
		delete(params, "target")
	case common.OpJsonTemplate:
		if template, ok := params["template"]; ok {
			rawParams["template"] = template
			delete(params, "template")
		}
	case common.InternalToolPy3Opt:
		if code, ok := params["code"]; ok {
			rawParams["code"] = code
			delete(params, "code")
		}
	}

	err := g.GenerateValue(params)

	if err != nil {
		return err
	}

	vmIns.LoadInstructions(g.Instructions)

	// 使用ShareData的GetAll方法安全地获取数据副本作为VM的Env
	env := make(map[string]interface{})
	if taskIns.RelatedDagInstance.ShareData != nil {
		// Optimize: Use GetAll() directly if we trust it returns a safe copy or if we are fine with it.
		// However, we want to inject local variables "index" which might conflict if we modify the map directly returned by GetAll if it was a pointer (it returns value).
		env = taskIns.RelatedDagInstance.ShareData.GetAll()

		// Inject loop index if available in task params
		// "current_iteration" is set by LoopHandler/LoopExecutor
		if val, ok := params["current_iteration"]; ok {
			env["index"] = val
			env["__loop_index"] = val
		}
	}

	vmIns.Env = env
	vmIns.Run()

	_, ret, err := vmIns.Result()

	if err != nil {
		return err
	}

	if resultMap, ok := ret.(map[string]interface{}); ok {
		if len(rawParams) > 0 {
			for k, v := range rawParams {
				resultMap[k] = v
			}
		}
		taskIns.SetParams(resultMap)
	}

	return nil
}

func (e *DefExecutor) renderParams(taskIns *entity.TaskInstance) error {
	data := map[string]interface{}{}

	dagInstance := taskIns.RelatedDagInstance
	if dagInstance != nil {
		data["vars"] = dagInstance.Vars
		if dagInstance.ShareData != nil {
			data["shareData"] = dagInstance.ShareData.GetAll()
		}
	}

	params := taskIns.GetParams()
	err := value.MapValue(params).Walk(func(walkContext *value.WalkContext, v interface{}) error {
		if m, ok := v.(string); ok {
			// 赋值操作不解析 target 的值
			if taskIns.ActionName == common.InternalAssignOpt && walkContext.Path() == "target" {
				return nil
			}

			if strings.Contains(m, "{{__") && strings.Contains(m, "}}") {
				// n := strings.Replace(m, "{{", "{{.shareData.", 1)
				n := strings.ReplaceAll(m, "{{", "{{.shareData.")
				result, err := e.paramRender.Render(n, data)
				if err != nil {
					return err
				}
				walkContext.Setter(result)
			}
		}
		// 遍历 slice
		// var res = make([]interface{}, 0)

		// if s, ok := v.(primitive.A); ok {
		// 	for index := range s {
		// 		if _s, ok := s[index].(string); ok {
		// 			if strings.Contains(_s, "{{") && strings.Contains(_s, "}}") {
		// 				n := strings.Replace(_s, "{{", "{{.shareData.", 1)
		// 				result, err := e.paramRender.Render(n, data)
		// 				if err != nil {
		// 					return err
		// 				}
		// 				res = append(res, result)
		// 			} else {
		// 				res = append(res, s[index])

		// 			}
		// 		}
		// 	}
		// 	walkContext.Setter(res)
		// }
		return nil
	})
	if err != nil {
		log.Errorf("WalkString failed: %v", err)
		return err
	}
	taskIns.SetParams(params)
	return nil
}

// Close
func (e *DefExecutor) Close() {
	e.lock.Lock()
	defer e.lock.Unlock()

	defer close(e.closeCh)
	e.closeCh <- struct{}{}
	for _, queueType := range queueTypes {
		close(e.initQueues[queueType])
		close(e.workerQueues[queueType])
	}
	e.initWg.Wait()
	e.workerWg.Wait()
}

func (e *DefExecutor) handleTaskError(ctx context.Context, taskIns *entity.TaskInstance, err error) {
	_, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	_, ok := e.cancelMap.Load(taskIns.ID)
	if err != nil {
		taskIns.Reason = err
		setStatus := entity.TaskInstanceStatusFailed
		if !ok {
			setStatus = entity.TaskInstanceStatusCanceled
		}

		taskIns.Reason = err
		if err := taskIns.SetStatus(ctx, setStatus); err != nil {
			log.Error("set status failed",
				"task_id", taskIns.ID,
				"err", err)
		}
		return
	}

	// 对于循环任务，无论状态如何，都不自动标记为取消状态
	// 循环任务的状态由循环执行器单独控制
	if taskIns.ActionName == common.Loop {
		return
	}

	// 如果任务在cancelMap中，说明它已被取消，直接返回
	if ok {
		return
	}

	taskIns.Reason = ReasonSuccessAfterCanceled
	if pErr := taskIns.Patch(ctx, &entity.TaskInstance{
		BaseInfo: taskIns.BaseInfo,
		Reason:   ReasonSuccessAfterCanceled}); pErr != nil {
		log.Errorf("tag canceled task instance[%s] failed: %s", taskIns.ID, pErr)
	}
}

func (e *DefExecutor) getTaskInsLock(id string) *sync.Mutex {
	lock, _ := e.taskInsLocks.LoadOrStore(id, &sync.Mutex{})
	return lock.(*sync.Mutex)
}
