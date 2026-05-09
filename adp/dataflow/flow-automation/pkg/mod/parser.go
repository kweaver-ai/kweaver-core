package mod

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	aErrs "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	libErrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/actions"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/event"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/utils"
	cutils "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
	"github.com/shiningrush/goevent"
	"github.com/spaolacci/murmur3"
)

const (
	TaskInstanceChan = 50 // 允许最大通道
)

var prioritySlice = []string{
	common.PriorityHighest,
	common.PriorityHigh,
	common.PriorityMedium,
	common.PriorityLow,
	common.PriorityLowest,
}

// DefParser def parser
type DefParser struct {
	workerNumber int
	workerQueue  []chan *entity.TaskInstance
	workerWg     sync.WaitGroup
	dagInsWg     sync.WaitGroup
	taskTrees    sync.Map
	taskTimeout  time.Duration

	closeCh chan struct{}
	lock    sync.RWMutex
	// log     common.Logger
	mq           MQHandler
	listinsCount int
}

// NewDefParser 实例化DefParser
func NewDefParser(workerNumber int, listinsCount int, taskTimeout time.Duration) *DefParser {
	return &DefParser{
		workerNumber: workerNumber,
		workerWg:     sync.WaitGroup{},
		dagInsWg:     sync.WaitGroup{},
		closeCh:      make(chan struct{}),
		taskTimeout:  taskTimeout,
		// log:          common.NewLogger(),
		mq:           NewMQHandler(),
		listinsCount: listinsCount,
	}
}

// Init 初始化
func (p *DefParser) Init() {
	for _, priority := range prioritySlice {
		v := priority
		p.workerWg.Add(1)
		go p.startWatcher(func() error {
			return p.watchScheduledDagIns(v)
		})
	}
	// p.workerWg.Add(1)
	// go p.startWatcher(p.watchDagInsCmd)

	for i := 0; i < p.workerNumber; i++ {
		p.workerWg.Add(1)
		ch := make(chan *entity.TaskInstance, TaskInstanceChan)
		p.workerQueue = append(p.workerQueue, ch)
		go p.goWorker(ch)
	}
	if err := p.initialRunningDagIns(); err != nil {
		log.Fatalf("parser init dags failed: %s", err)
	}
}

func (p *DefParser) startWatcher(do func() error) {
	timerCh := time.NewTicker(time.Second)
	closed := false
	for !closed {
		select {
		case <-p.closeCh:
			closed = true
			timerCh.Stop()
		case <-timerCh.C:
			if err := do(); err != nil {
				p.handleErr(err)
			}
		}
	}
	p.workerWg.Done()
}

func (p *DefParser) watchScheduledDagIns(priority string) (err error) {
	start := time.Now()
	e := &event.ParseScheduleDagInsCompleted{}
	defer func() {

		if err != nil {
			err = fmt.Errorf("watch scheduled dag ins failed: %w", err)
			e.Error = err
		}
		e.ElapsedMs = time.Since(start).Milliseconds()
		goevent.Publish(e)
	}()

	cons := []interface{}{priority}
	if priority == common.PriorityLowest {
		cons = append(cons, nil)
	}

	// 根据用户分别进行调度
	users, err := GetStore().DisdinctDagInstance(&ListDagInstanceInput{
		Status: []entity.DagInstanceStatus{
			entity.DagInstanceStatusScheduled,
		},
		Priority:      cons,
		DistinctField: "userid",
	})
	if err != nil {
		return err
	}

	for _, userid := range users {
		_ins, err := GetStore().ListDagInstance(context.TODO(), &ListDagInstanceInput{
			Status: []entity.DagInstanceStatus{
				entity.DagInstanceStatusScheduled,
			},
			Limit:    500,
			UserIDs:  []string{userid.(string)},
			Priority: cons,
			Worker:   GetKeeper().WorkerKey(),
		})
		if err != nil {
			log.Errorf("[watchScheduledDagIns] err: %v", err.Error())
		}

		for _, instance := range _ins {
			err = instance.LoadExtData(context.Background())
			if err != nil {
				log.Errorf("[watchScheduledDagIns] dagins load ext data failed priority: %s, count: %d, woker: %s, err: %s", priority, len(_ins), GetKeeper().WorkerKey(), err.Error())
				continue
			}
			if instance.Mode == entity.DagInstanceModeVM {
				go func() {
					_ = NewVMExt(context.Background(), instance, instance.UserID).Boot()
				}()
				continue
			}

			dIns, err := GetStore().GetDagInstance(context.TODO(), instance.ID)
			if err != nil {
				log.Errorf("[watchScheduledDagIns] check dag ins status failed priority: %s, count: %d, woker: %s, err: %s", priority, len(_ins), GetKeeper().WorkerKey(), err.Error())
				continue
			}

			if dIns.Status == entity.DagInstanceStatusCancled {
				continue
			}

			err = p.parseScheduleDagIns(context.TODO(), instance)
			if err != nil {
				log.Errorf("[watchScheduledDagIns] parser dag failed priority: %s, count: %d, woker: %s, err: %s", priority, len(_ins), GetKeeper().WorkerKey(), err.Error())
				continue
			}
			p.InitialDagIns(context.TODO(), instance)
		}
	}
	return
}

func (p *DefParser) watchDagInsCmd() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("watch dag command failed: %w", err)
		}
	}()

	dagIns, err := GetStore().ListDagInstance(context.TODO(), &ListDagInstanceInput{
		Worker:        GetKeeper().WorkerKey(),
		HasCmd:        true,
		ExcludeModeVM: true,
	})
	if err != nil {
		return err
	}
	for i := range dagIns {
		if err = p.parseCmd(dagIns[i]); err != nil { //nolint
			return err
		}
	}
	return nil
}

func (p *DefParser) goWorker(queue <-chan *entity.TaskInstance) {
	for taskIns := range queue {
		if err := p.workerDo(taskIns); err != nil {
			p.handleErr(fmt.Errorf("worker do failed: %w", err))
		}
	}
	p.workerWg.Done()
}

func (p *DefParser) initialRunningDagIns() error {
	var err error
	ctx, span := trace.StartInternalSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	dagIns, err := GetStore().ListDagInstance(ctx, &ListDagInstanceInput{
		Worker: GetKeeper().WorkerKey(),
		Status: []entity.DagInstanceStatus{
			entity.DagInstanceStatusRunning,
		},
	})
	if err != nil {
		return err
	}

	for _, d := range dagIns {
		err = d.LoadExtData(context.Background())

		if err != nil {
			log.Errorf("dag instance[%s] LoadExtData failed: %s", d.ID, err.Error())
			continue
		}

		if d.Mode == entity.DagInstanceModeVM {
			_ = NewVMExt(ctx, d, d.UserID).Boot()
		} else {
			p.InitialDagIns(ctx, d)
		}
	}
	return nil
}

// InitialDagIns
func (p *DefParser) InitialDagIns(ctx context.Context, dagIns *entity.DagInstance) {
	var err error
	ctx, span := trace.StartInternalSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	tasks, err := GetStore().ListTaskInstance(ctx, &ListTaskInstanceInput{
		DagInsID: dagIns.ID,
	})
	if err != nil {
		log.Errorf("dag instance[%s] list task instance failed: %s", dagIns.ID, err)
		return
	}

	if len(tasks) == 0 {
		return
	}
	root, err := BuildRootNode(MapTaskInsToGetter(tasks))
	if err != nil {
		log.Warnf("dag instance[%s] build task tree failed: %s", dagIns.ID, err)
		taskIDMap := make(map[string]*entity.TaskInstance, 0)
		deleteTask := make([]string, 0)
		fixTasks := make([]*entity.TaskInstance, 0)

		for _, task := range tasks {
			if _, ok := taskIDMap[task.TaskID]; ok {
				deleteTask = append(deleteTask, task.ID)
			} else {
				taskIDMap[task.TaskID] = task
				fixTasks = append(fixTasks, task)
			}
		}
		if len(deleteTask) > 0 {
			if derr := GetStore().BatchDeleteTaskIns(ctx, deleteTask); derr != nil {
				log.Errorf("delete repeat task instances failed, dagInsId: %s, err: %s", dagIns.ID, derr)
			}
		}
		root, err = BuildRootNode(MapTaskInsToGetter(fixTasks))
		if err != nil {
			log.Errorf("dag instance[%s] build task tree failed: %s", dagIns.ID, err)
			return
		}
		tasks = fixTasks
	}

	tree := &TaskTree{
		DagIns: dagIns,
		Root:   root,
	}
	executableTaskIds := tree.Root.GetExecutableTaskIds()

	if len(executableTaskIds) == 0 {
		sts, taskInsId := tree.Root.ComputeStatus()
		switch sts {
		case TreeStatusSuccess:
			tree.DagIns.Success()
			p.handleDagInsResult(dagIns)
		case TreeStatusBlocked:
			tree.DagIns.Block(fmt.Sprintf("initial blocked because task ins[%s]", taskInsId))
		case TreeStatusFailed:
			{
				var taskIns *entity.TaskInstance
				for _, t := range tasks {
					if t.ID == taskInsId {
						taskIns = t
						break
					}
				}

				if taskIns != nil {
					tree.DagIns.FailDetail(map[string]any{
						"taskId":     taskIns.TaskID,
						"name":       taskIns.Name,
						"actionName": taskIns.ActionName,
						"detail":     cutils.IfNot(taskIns.Reason != nil && taskIns.Reason != "", taskIns.Reason, "initial failed"),
					})
				} else {
					tree.DagIns.FailDetail(map[string]any{
						"detail": "initial failed",
					})
				}
				p.handleDagInsResult(dagIns)
			}
		default:
			log.Warn("initial a dag which has no executable tasks",
				utils.LogKeyDagInsID, dagIns.ID)
			return
		}

		if err := GetStore().PatchDagIns(ctx, &entity.DagInstance{
			BaseInfo:         entity.BaseInfo{ID: dagIns.ID},
			EventPersistence: dagIns.EventPersistence,
			Status:           dagIns.Status,
			Reason:           tree.DagIns.Reason,
			EndedAt:          tree.DagIns.EndedAt}); err != nil {
			log.Errorf("patch dag instance[%s] failed: %s", dagIns.ID, err)
			return
		}

		if tree.DagIns.Status == entity.DagInstanceStatusFailed || tree.DagIns.Status == entity.DagInstanceStatusSuccess {
			go p.handleDagInstanceCompletion(ctx, tree.DagIns.Status, tree, nil)
		}

		return
	}

	p.taskTrees.Store(dagIns.ID, tree)
	taskMap := getTasksMap(tasks)
	for _, tid := range executableTaskIds {
		GetExecutor().Push(dagIns, taskMap[tid])
	}
}

func (p *DefParser) RunDagIns(dagIns *entity.DagInstance) (err error) {
	if dagIns.DagType == common.DagTypeSecurityPolicy {
		if err = p.parseScheduleDagIns(context.TODO(), dagIns); err != nil {
			return err
		}
		p.InitialDagIns(context.TODO(), dagIns)
	}
	return
}

func getTasksMap(tasks []*entity.TaskInstance) map[string]*entity.TaskInstance {
	tmpMap := map[string]*entity.TaskInstance{}
	for i := range tasks {
		tmpMap[tasks[i].ID] = tasks[i]
	}
	return tmpMap
}

func (p *DefParser) executeNext(taskIns *entity.TaskInstance) error {
	var err error
	ctx, span := trace.StartInternalSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	// 不再特殊处理循环任务，让它走正常的处理流程，以便后续节点能被执行
	// 循环任务由executor处理成SUCCESS状态后，会让parser找它的子节点

	tree, ok := p.getTaskTree(taskIns.DagInsID)
	if !ok {
		log.Warnf("dag instance[%s] does not found task tree", taskIns.DagInsID)
		return nil
	}

	ids, find := tree.Root.GetNextTaskIds(taskIns)
	if !find {
		err = fmt.Errorf("task instance[%s] does not found normal node", taskIns.ID)
		return err
	}

	shouldReturn := taskIns.ActionName == common.InternalReturnOpt && taskIns.Status != entity.TaskInstanceStatusSkipped

	// only the tasks which is not success has no next task ids
	if shouldReturn || len(ids) == 0 {

		if shouldReturn && taskIns.Status == entity.TaskInstanceStatusSuccess {
			if result, ok := taskIns.Results.(string); ok && result == actions.ReturnResultSuccess {
				tree.DagIns.Success()
				p.handleDagInsResult(tree.DagIns)
			} else {
				tree.DagIns.FailDetail(map[string]any{
					"taskId":     taskIns.TaskID,
					"name":       taskIns.Name,
					"actionName": taskIns.ActionName,
					"detail":     "return failed",
				})
				p.handleDagInsResult(tree.DagIns)
			}
		} else {
			treeStatus, taskId := tree.Root.ComputeStatus()
			switch treeStatus {
			case TreeStatusRunning:
				return nil
			case TreeStatusFailed:
				tree.DagIns.FailDetail(map[string]any{
					"taskId":     taskIns.TaskID,
					"name":       taskIns.Name,
					"actionName": taskIns.ActionName,
					"detail":     taskIns.Reason,
				})
				p.handleDagInsResult(tree.DagIns)
			case TreeStatusBlocked:
				tree.DagIns.Block(fmt.Sprintf("task[%s] blocked", taskId))
			case TreeStatusSuccess:
				tree.DagIns.Success()
				p.handleDagInsResult(tree.DagIns)
			}
		}

		status := tree.DagIns.Status
		// 每次执行后续节点的时候，判断当前dagIns状态是否已是取消状态
		dIns, err := GetStore().GetDagInstance(ctx, taskIns.DagInsID)
		if err != nil {
			return err
		}

		if dIns.Status == entity.DagInstanceStatusCancled {
			status = entity.DagInstanceStatusCancled
		}
		// tree has already completed, delete from map
		p.taskTrees.Delete(taskIns.DagInsID)
		if err := GetStore().PatchDagIns(ctx, &entity.DagInstance{
			BaseInfo:         entity.BaseInfo{ID: tree.DagIns.ID},
			Status:           status,
			EventPersistence: tree.DagIns.EventPersistence,
			Reason:           tree.DagIns.Reason,
			EndedAt:          tree.DagIns.EndedAt,
		}); err != nil {
			return err
		}

		if status == entity.DagInstanceStatusFailed || status == entity.DagInstanceStatusSuccess {
			go p.handleDagInstanceCompletion(ctx, status, tree, taskIns)
		}

		return nil
	}
	if taskIns.Reason == ReasonSuccessAfterCanceled {
		return p.cancelChildTasks(ctx, tree, ids)
	}

	return p.pushTasks(ctx, tree.DagIns, ids)
}

func (p *DefParser) pushTasks(ctx context.Context, dagIns *entity.DagInstance, ids []string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	tasks, err := GetStore().ListTaskInstance(ctx, &ListTaskInstanceInput{
		IDs: ids,
	})
	if err != nil {
		return err
	}
	for _, t := range tasks {
		GetExecutor().Push(dagIns, t)
	}

	return nil
}

func (p *DefParser) cancelChildTasks(ctx context.Context, tree *TaskTree, ids []string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	walkNode(tree.Root, func(node *TaskNode) bool {
		if utils.StringsContain(ids, node.TaskInsID) {
			node.Status = entity.TaskInstanceStatusCanceled
		}
		return true
	}, false)

	for _, id := range ids {
		if err := GetStore().PatchTaskIns(ctx, &entity.TaskInstance{
			BaseInfo: entity.BaseInfo{ID: id},
			Status:   entity.TaskInstanceStatusCanceled,
			Reason:   ReasonParentCancel,
		}); err != nil {
			return err
		}
	}

	// not equal running mean that all tasks already completed
	if sts, _ := tree.Root.ComputeStatus(); sts != TreeStatusRunning {
		p.taskTrees.Delete(tree.DagIns.ID)
	}

	if !tree.DagIns.CanModifyStatus() {
		return nil
	}
	tree.DagIns.FailDetail(map[string]any{
		"detail": fmt.Sprintf("task instance[%s] canceled", strings.Join(ids, ",")),
	})

	p.handleDagInsResult(tree.DagIns)
	if err := tree.DagIns.SaveExtData(context.Background()); err != nil {
		return err
	}
	return GetStore().PatchDagIns(ctx, tree.DagIns)
}

func (p *DefParser) getTaskTree(dagInsId string) (*TaskTree, bool) {
	tasks, ok := p.taskTrees.Load(dagInsId)
	if !ok {
		return nil, false
	}
	return tasks.(*TaskTree), true
}

// EntryTaskIns
func (p *DefParser) EntryTaskIns(taskIns *entity.TaskInstance) {
	murmurHash := murmur3.New32()
	// murmur3 hash does not return error, so we don't need to handle it
	_, _ = murmurHash.Write([]byte(taskIns.DagInsID))
	mod := int(murmurHash.Sum32()) % p.workerNumber
	p.sendToChannel(mod, taskIns, true)
}

func (p *DefParser) sendToChannel(mod int, taskIns *entity.TaskInstance, newRoutineWhenFull bool) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	// try to exit the sender goroutine as early as possible.
	// try-receive and try-send select blocks are specially optimized by the standard Go compiler,
	// so they are very efficient.
	select {
	case <-p.closeCh:
		log.Info("parser has already closed, so will not execute next task instances")
		return
	default:
	}

	if !newRoutineWhenFull {
		p.workerQueue[mod] <- taskIns
		return
	}

	select {
	// ensure that same dag instance handled by same worker, so avoid parallel writing
	case p.workerQueue[mod] <- taskIns:
	// if queue is full, we can do it in a new goroutine to prevent dead lock
	default:
		go p.sendToChannel(mod, taskIns, false)
	}
}

func (p *DefParser) workerDo(taskIns *entity.TaskInstance) error {
	return p.executeNext(taskIns)
}

func (p *DefParser) parseScheduleDagIns(ctx context.Context, dagIns *entity.DagInstance) error {
	if dagIns.Status == entity.DagInstanceStatusScheduled {
		var err error
		ctx, span := trace.StartInternalSpan(ctx)
		defer func() { trace.TelemetrySpanEnd(span, err) }()

		dag, err := GetStore().GetDagWithOptionalVersion(ctx, dagIns.DagID, dagIns.VersionID)
		if err != nil {
			if aErrs.IsNotFoundErr(err) {
				if _err := GetStore().BatchDeleteDagIns(ctx, []string{dagIns.ID}); _err != nil {
					return _err
				}

				if err := rds.GetDagInstanceExtDataDao().Remove(ctx, &rds.ExtDataQueryOptions{
					DagInsID: dagIns.ID,
				}); err != nil {
					return err
				}
			}
			return err
		}
		tasks, err := GetStore().ListTaskInstance(ctx, &ListTaskInstanceInput{
			DagInsID: dagIns.ID,
		})
		if err != nil {

			return err
		}

		// the init of tasks is not complete, should continue/start it.
		if len(dag.Tasks) != len(tasks) {
			var needInitTaskIns []*entity.TaskInstance
			for i := range dag.Tasks {
				notFound := true
				for j := range tasks {
					if dag.Tasks[i].ID == tasks[j].TaskID {
						notFound = false
					}
				}

				if notFound {
					renderParams, err := dagIns.Vars.Render(dag.Tasks[i].Params)
					if err != nil {
						return err
					}
					dag.Tasks[i].Params = renderParams
					if dag.Tasks[i].TimeoutSecs == 0 {
						dag.Tasks[i].TimeoutSecs = int(p.taskTimeout.Seconds())
					}
					if dag.Tasks[i].ActionName == common.InternalToolPy3Opt {
						// 兼容已创建的流程
						dag.Tasks[i].TimeoutSecs = 24 * 60 * 60
					}
					needInitTaskIns = append(needInitTaskIns, entity.NewTaskInstance(dagIns.ID, &(dag.Tasks[i])))
				}
			}
			if len(needInitTaskIns) > 0 {
				if _, err := GetStore().BatchCreateTaskIns(ctx, needInitTaskIns); err != nil {
					fmt.Println("BatchCreateTaskIns failed: ", err.Error())
					return err
				}
			}
		}

		dagIns.Run()

		if err := GetStore().PatchDagIns(ctx, &entity.DagInstance{
			BaseInfo:         dagIns.BaseInfo,
			Status:           dagIns.Status,
			EventPersistence: dagIns.EventPersistence,
			Reason:           dagIns.Reason,
		}, "Reason"); err != nil {
			return err
		}
	}
	return nil
}

func (p *DefParser) parseCmd(dagIns *entity.DagInstance) (err error) {
	if dagIns.Cmd != nil {
		switch dagIns.Cmd.Name {
		case entity.CommandNameRetry:
			hasAnyTaskRetried := false
			defer func() {
				if err == nil && hasAnyTaskRetried {
					p.InitialDagIns(context.TODO(), dagIns)
				}
			}()

			taskIns, err := GetStore().ListTaskInstance(context.TODO(), &ListTaskInstanceInput{
				IDs:    dagIns.Cmd.TargetTaskInsIDs,
				Status: []entity.TaskInstanceStatus{entity.TaskInstanceStatusFailed, entity.TaskInstanceStatusCanceled},
			})
			if err != nil {
				return err
			}

			for _, t := range taskIns {
				if t.Status != entity.TaskInstanceStatusFailed &&
					t.Status != entity.TaskInstanceStatusCanceled {
					continue
				}

				t.Status = entity.TaskInstanceStatusRetrying
				t.Reason = ""
				if err := GetStore().UpdateTaskIns(context.TODO(), t); err != nil {
					return err
				}
				hasAnyTaskRetried = true
			}
			dagIns.Run()
		case entity.CommandNameCancel:
			if err := GetExecutor().CancelTaskIns(dagIns.Cmd.TargetTaskInsIDs); err != nil {
				return err
			}
		}

		dagIns.Cmd = nil
		if err := GetStore().PatchDagIns(context.TODO(), &entity.DagInstance{
			BaseInfo:         dagIns.BaseInfo,
			Status:           dagIns.Status,
			EventPersistence: dagIns.EventPersistence,
			Cmd:              dagIns.Cmd,
			Reason:           dagIns.Reason,
		}, "Cmd", "Reason"); err != nil {
			return err
		}
	}
	return nil
}

// Close
func (p *DefParser) Close() {
	p.lock.Lock()
	defer p.lock.Unlock()

	close(p.closeCh)
	for i := range p.workerQueue {
		close(p.workerQueue[i])
	}
	p.workerWg.Wait()
}

func (p *DefParser) handleErr(err error) {
	log.Errorf("parser get some error",
		"module", "parser",
		"err", err)
}

func (p *DefParser) handleDagInsResult(dagIns *entity.DagInstance) {
	// 发布安全策略消息
	if dagIns.DagType == common.DagTypeSecurityPolicy {
		msg := &SecurityPolicyProcResultMsg{
			PID:        dagIns.ID,
			PolicyType: dagIns.PolicyType,
		}

		if dagIns.Status == entity.DagInstanceStatusSuccess {
			msg.Result = "success"
		} else {
			msg.Result = "failed"
		}

		log.Infof("[Parser.handleDagInsResult] publish topic: pid=%s, result=%s\n", msg.PID, msg.Result)

		result, _ := jsoniter.Marshal(msg)

		err := NewMQHandler().Publish(common.TopicSecurityPolicyProcResult, result)

		if err != nil {
			log.Infof("[Parser.handleDagInsResult] publish topic failed: %s\n", err.Error())
		}
	}
}

// handleDagInstanceCompletion 处理 DAG 实例完成后的所有操作
// 包括：数据清理、回调执行、日志记录
func (p *DefParser) handleDagInstanceCompletion(ctx context.Context, status entity.DagInstanceStatus, tree *TaskTree, taskIns *entity.TaskInstance) {
	// 执行回调
	HandleSubprocessCallback(ctx, status, tree.DagIns, taskIns)

	// 清理成功的 DAG 实例数据
	p.cleanupSuccessDagInstance(ctx, tree)

	// 记录日志
	p.logDagInstanceCompletion(ctx, tree.DagIns, taskIns)
}

// cleanupSuccessDagInstance 清理成功完成的 DAG 实例数据
func (p *DefParser) cleanupSuccessDagInstance(ctx context.Context, tree *TaskTree) {
	if tree.DagIns.Status != entity.DagInstanceStatusSuccess {
		return
	}

	// 删除任务实例数据
	if dErr := GetStore().DeleteTaskInsByDagInsID(ctx, tree.DagIns.ID); dErr != nil {
		log.Warnf("run success, delete task instance failed: %s", dErr.Error())
	}

	// 上传事件到 OSS
	if tree.DagIns.EventPersistence == entity.DagInstanceEventPersistenceSql {
		_ = UploadDagInstanceEvents(context.Background(), tree.DagIns)
	}
}

// logDagInstanceCompletion 记录 DAG 实例完成日志
func (p *DefParser) logDagInstanceCompletion(ctx context.Context, dagIns *entity.DagInstance, taskIns *entity.TaskInstance) {
	dag, err := GetStore().GetDagWithOptionalVersion(context.Background(), dagIns.DagID, dagIns.VersionID)
	if err != nil {
		log.Warnf("[parser.logDagInstanceCompletion] GetDagWithOptionalVersion failed: %s", err.Error())
		return
	}

	// 调试模式下不执行后续操作
	if dag == nil || dag.IsDebug {
		return
	}
	// 构建日志详情
	var detail, extMsg string
	if dagIns.DagType == common.DagTypeSecurityPolicy {
		bodyType := "RunSecurityPolicyFlowFailed"
		if dagIns.Status == entity.DagInstanceStatusSuccess {
			bodyType = "RunSecurityPolicyFlowSuccess"
		}
		detail, extMsg = common.GetLogBody(bodyType, []any{dag.ID}, []any{})
	} else {
		bodyType := common.CompleteTaskWithFailed
		if dagIns.Status == entity.DagInstanceStatusSuccess {
			bodyType = common.CompleteTaskWithSuccess
		}
		detail, extMsg = common.GetLogBody(bodyType, []any{dag.Name}, []any{})
	}

	// 构建日志对象
	object := map[string]interface{}{
		"type":          dag.Trigger,
		"id":            dagIns.ID,
		"dagId":         dagIns.DagID,
		"name":          dag.Name,
		"priority":      dagIns.Priority,
		"status":        dagIns.Status,
		"biz_domain_id": cutils.IfNot(dag.BizDomainID == "", common.BizDomainDefaultID, dag.BizDomainID),
	}
	if len(dag.Type) != 0 {
		object["dagType"] = dag.Type
	} else {
		object["dagType"] = common.DagTypeDefault
	}
	if dagIns.EndedAt < dagIns.CreatedAt {
		object["duration"] = time.Now().Unix() - dagIns.CreatedAt
	} else {
		object["duration"] = dagIns.EndedAt - dagIns.CreatedAt
	}

	// 获取用户信息
	varsGetter := dagIns.VarsGetter()
	userID, _ := varsGetter("operator_id")
	userType, _ := varsGetter("operator_type")
	userInfo, err := drivenadapters.NewUserManagement().GetUserInfoByType(fmt.Sprintf("%v", userID), fmt.Sprintf("%v", userType))
	if err != nil {
		log.Warnf("[parser.logDagInstanceCompletion] GetUserInfoByType failed: %s", err.Error())
		userName, _ := varsGetter("operator_name")
		userInfo = drivenadapters.UserInfo{
			UserID:   fmt.Sprintf("%v", userID),
			UserName: fmt.Sprintf("%v", userName),
			Type:     fmt.Sprintf("%v", userType),
		}
	}
	userInfo.VisitorType = common.InternalServiceUserType

	// 记录可观测性日志
	logger := drivenadapters.NewLogger()
	logger.LogO11y(&drivenadapters.BuildARLogParams{
		Operation:   common.ArLogEndDagIns,
		Description: detail,
		UserInfo:    &userInfo,
		Object:      object,
	}, &drivenadapters.O11yLogWriter{Logger: traceLog.NewFlowO11yLogger()})

	traceLog.WithContext(ctx).Infof("detail: %s, extMsg: %s", detail, extMsg)

	// 确定 OutBizID
	outBizID := dagIns.ID
	if taskIns != nil {
		outBizID = taskIns.ID
	}

	// 记录 DIP 审计日志
	write := &drivenadapters.JSONLogWriter{SendFunc: p.mq.Publish}
	logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
		UserInfo:  &userInfo,
		Msg:       detail,
		ExtMsg:    extMsg,
		OutBizID:  outBizID,
		Operation: drivenadapters.ExecuteOperation,
		ObjID:     dag.ID,
		ObjName:   dag.Name,
		LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
	}, write)
}

func HandleSubprocessCallback(ctx context.Context, status entity.DagInstanceStatus, dagIns *entity.DagInstance, taskIns *entity.TaskInstance) {
	if dagIns.SuccessCallback == "" {
		return
	}

	store := GetStore()
	if dagIns.DagType == common.DagTypeDataFlow {
		varsGetter := dagIns.VarsGetter()
		batchRunID, ok := varsGetter("batch_run_id")
		if !ok {
			return
		}

		dagIns, err := store.ListDagInstance(ctx, &ListDagInstanceInput{
			DagIDs: []string{dagIns.DagID},
			Status: []entity.DagInstanceStatus{entity.DagInstanceStatusInit, entity.DagInstanceStatusScheduled, entity.DagInstanceStatusRunning, entity.DagInstanceStatusBlocked},
			MatchQuery: &MatchQuery{
				Field: "vars.batch_run_id.value",
				Value: batchRunID,
			},
			SelectField: []string{"_id", "dagId"},
		})
		if err != nil {
			log.Warnf("[executeCallback] ListDagInstance failed: %s", err.Error())
			return
		}

		if len(dagIns) > 0 {
			return
		}
	}

	var body any
	if status == entity.DagInstanceStatusSuccess {
		var data = map[string]any{}
		if dagIns.DagType != common.DagTypeDataFlow {
			dag, err := store.GetDagWithOptionalVersion(ctx, dagIns.DagID, dagIns.VersionID)
			if err != nil {
				log.Warnf("[executeCallback] GetDagWithOptionalVersion failed: %s", err.Error())
				return
			}

			lastStep := dag.Steps[len(dag.Steps)-1]
			if lastStep.Operator != common.BranchOpt {
				events, _ := GetDagInstanceEvents(ctx, &rds.DagInstanceEventListOptions{
					DagInstanceID: dagIns.ID,
					Names:         []string{fmt.Sprintf("__%s", lastStep.ID)},
				})
				if len(events) > 0 {
					_ = json.Unmarshal([]byte(events[0].Data), &data)
				}
			}
		}
		body = data
	} else {
		// 处理失败情况
		var reason any
		if taskIns != nil {
			reason = taskIns.Reason
		} else {
			// taskIns 为 nil 时，从 dagIns 中获取 reason
			reason = dagIns.Reason
			data := map[string]any{}
			mErr := json.Unmarshal([]byte(dagIns.Reason), &data)
			if mErr == nil {
				detail, ok := data["detail"]
				reason = cutils.IfNot(ok, detail, reason)
			}
		}

		switch reason.(type) {
		case *aErrs.IError, aErrs.IError, *libErrors.RestError, libErrors.RestError:
			body = reason
		case map[string]any:
			if _, ok := reason.(map[string]any)["code"]; ok {
				body = reason
			} else {
				body = libErrors.NewPublicRestError(ctx, libErrors.PErrorInternalServerError, libErrors.PErrorInternalServerError, reason)
			}
		default:
			body = libErrors.NewPublicRestError(ctx, libErrors.PErrorInternalServerError, libErrors.PErrorInternalServerError, reason)
		}
	}

	client := drivenadapters.NewOtelHTTPClient()
	_, _, err := client.Post(context.Background(), dagIns.SuccessCallback, map[string]string{}, body)
	if err != nil {
		log.Warnf("[parser.executeCallback] 调用 callBack %s 失败: %v \n", dagIns.SuccessCallback, err)
	}

	return
}
