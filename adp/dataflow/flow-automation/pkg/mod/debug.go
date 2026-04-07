package mod

import (
	"context"
	"fmt"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/actions"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/dependency"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

// DebugExecute debug执行上下文
type DebugExecute struct {
	ctx     context.Context
	dagIns  *entity.DagInstance
	taskIns *entity.TaskInstance
	token   *entity.Token
}

// NewDebugExecute 实例化
func NewDebugExecute(ctx context.Context, dagIns *entity.DagInstance, taskIns *entity.TaskInstance, token *entity.Token) *DebugExecute {
	return &DebugExecute{
		ctx:     ctx,
		dagIns:  dagIns,
		taskIns: taskIns,
		token:   token,
	}
}

// SingleDeBug 单步调试
func (d *DebugExecute) SingleDeBug() (err error) {
	ctx, span := trace.StartInternalSpan(d.ctx)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		trace.TelemetrySpanEnd(span, err)
	}()

	d.ctx = ctx
	executeMethods := entity.ExecuteMethods{
		Publish: NewMQHandler().Publish,
		GetDag:  func(ctx context.Context, id, versionID string) (*entity.Dag, error) { return &entity.Dag{}, nil },
		PatchDagIns: func(ctx context.Context, dagIns *entity.DagInstance, mustsPatchFields ...string) error {
			return nil
		},
	}

	debugCtx := entity.NewDebugExecuteContext(d.ctx, d.dagIns.MemoryShareData, d.dagIns.VarsGetter(), d.dagIns.VarsIterator(), d.taskIns.ParamsGetter(), d.taskIns.GetGraphID(), d.taskIns, executeMethods, dependency.NewDriven())

	d.taskIns.InitialDep(debugCtx,
		func(ctx context.Context, instance *entity.TaskInstance) error {
			return nil
		},
		d.dagIns,
	)

	var act entity.Action
	var p interface{}
	if strings.HasPrefix(d.taskIns.ActionName, "@operator/") {
		act = &actions.ComboOperator{
			Operator: d.taskIns.ActionName,
		}
		p = act
		err = weakDecode(d.taskIns.Params, p)
		if err != nil {
			return
		}
	} else {
		act = ActionMap[d.taskIns.ActionName]
		paramAct, ok := act.(entity.ParameterAction)
		if !ok {
			return fmt.Errorf("action %s not implement ParameterAction", d.taskIns.ActionName)
		}

		p = paramAct.ParameterNew()
		if p == nil {
			return fmt.Errorf("action %s not implement ParameterNew", d.taskIns.ActionName)
		}

		err = weakDecode(d.taskIns.Params, p)
		if err != nil {
			return
		}
	}

	// 节点执行前置操作
	if d.taskIns.Status == entity.TaskInstanceStatusInit {
		beforeAct, ok := act.(entity.BeforeAction)
		d.taskIns.Status = entity.TaskInstanceStatusRunning
		if ok {
			d.taskIns.Status, err = beforeAct.RunBefore(debugCtx, p)
			if err != nil {
				return
			}
		}

		d.taskIns.Results, err = act.Run(debugCtx, p, d.token)
		if err != nil {
			return
		}

		if d.taskIns.Status == entity.TaskInstanceStatusRunning {
			d.taskIns.Status = entity.TaskInstanceStatusEnding
		}
	}

	if d.taskIns.Status == entity.TaskInstanceStatusRunning {
		d.taskIns.Results, err = act.Run(debugCtx, p, d.token)
		if err != nil {
			return
		}

		d.taskIns.Status = entity.TaskInstanceStatusEnding
	}

	// 节点执行后置操作
	if d.taskIns.Status == entity.TaskInstanceStatusEnding {
		afterAct, ok := act.(entity.AfterAction)
		d.taskIns.Status = entity.TaskInstanceStatusSuccess
		if ok {
			d.taskIns.Status, err = afterAct.RunAfter(debugCtx, p)
			if err != nil {
				return
			}
		}
	}

	d.dagIns.Status = entity.DagInstanceStatusSuccess
	if d.taskIns.Status == entity.TaskInstanceStatusBlocked {
		d.dagIns.Status = entity.DagInstanceStatusBlocked
	}

	return
}
