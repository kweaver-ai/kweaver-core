package mod

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	liberrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm"
	vmerrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm/errors"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm/hook"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm/state"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
)

func (vm *VMExt) HookBeforeAssign(id string, target string, value any) {
	taskIns := &entity.TaskInstance{
		TaskID:     id,
		DagInsID:   vm.dagIns.ID,
		ActionName: common.InternalAssignOpt,
		Params: map[string]any{
			"target": target,
			"value":  value,
		},
		Status: entity.TaskInstanceStatusSuccess,
	}

	vm.AppendEvents(
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeTaskStatus,
			InstanceID: vm.dagIns.ID,
			Operator:   common.InternalAssignOpt,
			TaskID:     id,
			Status:     string(entity.TaskInstanceStatusRunning),
			Timestamp:  time.Now().UnixMicro(),
			Visibility: rds.DagInstanceEventVisibilityPublic,
		},
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeVariable,
			InstanceID: vm.dagIns.ID,
			Name:       target,
			Data:       value,
			Visibility: rds.DagInstanceEventVisibilityPublic,
			Timestamp:  time.Now().UnixMicro(),
		},
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeTaskStatus,
			InstanceID: vm.dagIns.ID,
			Operator:   common.InternalAssignOpt,
			TaskID:     id,
			Status:     string(entity.TaskInstanceStatusSuccess),
			Timestamp:  time.Now().UnixMicro(),
			Visibility: rds.DagInstanceEventVisibilityPublic,
		},
	)

	_, err := GetStore().BatchCreateTaskIns(vm.Context(), []*entity.TaskInstance{taskIns})

	if err != nil {
		vm.logger.Warnf("[VMExt.HookBeforeAssign] BatchCreateTaskIns err: %s", err.Error())
	}
}

func (vm *VMExt) HookBranchSkip(id string) {
	taskIns := &entity.TaskInstance{
		TaskID:     id,
		DagInsID:   vm.dagIns.ID,
		ActionName: common.BranchOpt,
		Status:     entity.TaskInstanceStatusSkipped,
	}

	vm.AppendEvents(
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeTaskStatus,
			InstanceID: vm.dagIns.ID,
			Operator:   common.BranchOpt,
			TaskID:     id,
			Status:     string(entity.TaskInstanceStatusSkipped),
			Timestamp:  time.Now().UnixMicro(),
			Visibility: rds.DagInstanceEventVisibilityPublic,
		},
	)

	_, err := GetStore().BatchCreateTaskIns(vm.Context(), []*entity.TaskInstance{taskIns})

	if err != nil {
		vm.logger.Warnf("[VMExt.HookBranchSkip] BatchCreateTaskIns err: %s", err.Error())
	}
}

func (vm *VMExt) HookBranchStart(id string) {

	taskIns := &entity.TaskInstance{
		TaskID:     id,
		DagInsID:   vm.dagIns.ID,
		ActionName: common.BranchOpt,
		Status:     entity.TaskInstanceStatusSuccess,
	}

	vm.AppendEvents(
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeTaskStatus,
			InstanceID: vm.dagIns.ID,
			Operator:   common.BranchOpt,
			TaskID:     id,
			Status:     string(entity.TaskInstanceStatusSuccess),
			Timestamp:  time.Now().UnixMicro(),
			Visibility: rds.DagInstanceEventVisibilityPublic,
		},
	)

	_, err := GetStore().BatchCreateTaskIns(vm.Context(), []*entity.TaskInstance{taskIns})

	if err != nil {
		vm.logger.Warnf("[VMExt.HookBranchStart] BatchCreateTaskIns err: %s", err.Error())
	}
}

func (vm *VMExt) HookLoopStart(id string, value any) {
	taskIns := &entity.TaskInstance{
		TaskID:     id,
		DagInsID:   vm.dagIns.ID,
		ActionName: common.Loop,
		Results:    value,
		Status:     entity.TaskInstanceStatusSuccess,
	}

	vm.AppendEvents(
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeTaskStatus,
			InstanceID: vm.dagIns.ID,
			Operator:   common.Loop,
			TaskID:     id,
			Status:     string(entity.TaskInstanceStatusRunning),
			Timestamp:  time.Now().UnixMicro(),
			Visibility: rds.DagInstanceEventVisibilityPublic,
		},
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeVariable,
			InstanceID: vm.dagIns.ID,
			Name:       fmt.Sprintf("__%s", id),
			Data:       value,
			Visibility: rds.DagInstanceEventVisibilityPublic,
			Timestamp:  time.Now().UnixMicro(),
		},
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeTaskStatus,
			InstanceID: vm.dagIns.ID,
			Operator:   common.Loop,
			TaskID:     id,
			Status:     string(entity.TaskInstanceStatusSuccess),
			Timestamp:  time.Now().UnixMicro(),
			Visibility: rds.DagInstanceEventVisibilityPublic,
		},
	)

	_, err := GetStore().BatchCreateTaskIns(vm.Context(), []*entity.TaskInstance{taskIns})

	if err != nil {
		vm.logger.Warnf("[VMExt.HookLoopStart] BatchCreateTaskIns err: %s", err.Error())
	}
}

func (vm *VMExt) HookLoopEnd(id string, value any) {
	vm.AppendEvents(
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeVariable,
			InstanceID: vm.dagIns.ID,
			Name:       fmt.Sprintf("__%s", id),
			Data:       value,
			Visibility: rds.DagInstanceEventVisibilityPublic,
			Timestamp:  time.Now().UnixMicro(),
		})
}

func (vm *VMExt) HookBeforeReturn(id string, value any) {
	taskIns := &entity.TaskInstance{
		TaskID:     id,
		DagInsID:   vm.dagIns.ID,
		ActionName: common.InternalReturnOpt,
		Results:    value,
		Status:     entity.TaskInstanceStatusSuccess,
	}

	vm.AppendEvents(
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeTaskStatus,
			InstanceID: vm.dagIns.ID,
			Operator:   common.InternalReturnOpt,
			TaskID:     id,
			Status:     string(entity.TaskInstanceStatusSuccess),
			Timestamp:  time.Now().UnixMicro(),
			Visibility: rds.DagInstanceEventVisibilityPublic,
		},
	)

	_, err := GetStore().BatchCreateTaskIns(vm.Context(), []*entity.TaskInstance{taskIns})

	if err != nil {
		vm.logger.Warnf("[VMExt.HookBeforeReturn] BatchCreateTaskIns err: %s", err.Error())
	}
}

func (vmIns *VMExt) HookResume(name, id string, value any) {
	var data any

	if rets, ok := value.([]any); ok && len(rets) > 0 {
		data = rets[0]
	}

	vmIns.AppendEvents(
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeVariable,
			InstanceID: vmIns.dagIns.ID,
			Name:       fmt.Sprintf("__%s", id),
			Data:       data,
			Visibility: rds.DagInstanceEventVisibilityPublic,
			Timestamp:  time.Now().UnixMicro(),
		},
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeTaskStatus,
			InstanceID: vmIns.dagIns.ID,
			Operator:   name,
			TaskID:     id,
			Status:     string(entity.TaskInstanceStatusSuccess),
			Timestamp:  time.Now().UnixMicro(),
			Visibility: rds.DagInstanceEventVisibilityPublic,
		},
	)
}

func (vmIns *VMExt) HookResumeError(name, id string, err any) {
	vmIns.AppendEvents(
		&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeTaskStatus,
			InstanceID: vmIns.dagIns.ID,
			Operator:   name,
			TaskID:     id,
			Status:     string(entity.TaskInstanceStatusFailed),
			Data:       err,
			Timestamp:  time.Now().UnixMicro(),
			Visibility: rds.DagInstanceEventVisibilityPublic,
		},
	)
}

func (vmIns *VMExt) HookVMStop() {
	ctx := vmIns.Context()
	dagIns := vmIns.dagIns

	patch := &entity.DagInstance{
		BaseInfo:         dagIns.BaseInfo,
		DagID:            dagIns.DagID,
		EventPersistence: dagIns.EventPersistence,
		EndedAt:          time.Now().Unix(),
	}

	vmState, data, vmErr := vmIns.Result()

	switch vmState {
	case state.Done:
		patch.Status = entity.DagInstanceStatusSuccess
		vmIns.logger.Infof("[VMExt.HookVMStop] 运行完成: %v\n", data)
		go dagIns.SendSuccessCallback(data)

	case state.Error:
		patch.Status = entity.DagInstanceStatusFailed

		reason := make(map[string]any)

		if vmErrObj, ok := vmErr.(*vmerrors.Error); ok {
			reason["taskId"] = vmErrObj.Step

			if detail, ok := vmErrObj.Detail.(map[string]any); ok {
				reason["actionName"] = detail["name"]
				reason["detail"] = detail["cause"]
			} else {
				reason["detail"] = detail
			}
		} else {
			reason["detail"] = vmErr
		}

		b, _ := json.Marshal(reason)
		patch.Reason = string(b)

		log.Infof("[VMExt.HookVMStop] 运行失败: %v\n", vmErr)
		go dagIns.SendErrorCallback(
			liberrors.NewPublicRestError(context.Background(), liberrors.PErrorInternalServerError,
				liberrors.PErrorInternalServerError,
				vmErr))

	case state.Wait:
		log.Infof("[VMExt.HookVMStop] 运行挂起\n")
		patch.Status = entity.DagInstanceStatusBlocked
	}

	switch dagIns.EventPersistence {
	case entity.DagInstanceEventPersistenceOss:
		// nothing to do
	case entity.DagInstanceEventPersistenceSql:
		vmIns.AppendEvents(&entity.DagInstanceEvent{
			Type:       rds.DagInstanceEventTypeVM,
			InstanceID: dagIns.ID,
			Status:     string(patch.Status),
			Data: &vm.VM{
				State:     vmIns.State,
				PC:        vmIns.PC,
				RegID:     vmIns.RegID,
				Stack:     vmIns.Stack,
				Env:       vmIns.Env,
				Callstack: vmIns.Callstack,
				Traces:    vmIns.Traces,
				Err:       vmIns.Err,
			},
			Timestamp:  time.Now().UnixMicro(),
			Visibility: rds.DagInstanceEventVisibilityPrivate,
		})
		if err := vmIns.WriteEvents(); err != nil {
			log.Warnf("[VMExt.HookVMStop] SaveExtData err, detail: %s", err.Error())
		}

	default:
		bytes, err := json.Marshal(vmIns)

		if err != nil {
			vmIns.logger.Warnf("[VMExt.HookVMStop] PatchDagIns err, detail: %s", err.Error())
			return
		}
		patch.ShareData = dagIns.ShareData
		patch.Dump = string(bytes)
	}

	if err := patch.SaveExtData(context.Background()); err != nil {
		log.Warnf("[VMExt.HookVMStop] SaveExtData err, detail: %s", err.Error())
		return
	}

	if err := GetStore().PatchDagIns(ctx, patch); err != nil {
		log.Warnf("[VMExt.HookVMStop] PatchDagIns err, detail: %s", err.Error())
	}

	go vmIns.AfterStopLog(context.Background(), dagIns, patch.Status)
}

func (vm *VMExt) AfterStopLog(ctx context.Context, dagIns *entity.DagInstance, status entity.DagInstanceStatus) {
	if status != entity.DagInstanceStatusFailed && status != entity.DagInstanceStatusSuccess {
		return
	}

	if dagIns.EventPersistence == entity.DagInstanceEventPersistenceSql {
		_ = UploadDagInstanceEvents(ctx, dagIns)
	}

	// 流程执行成功删除所有task信息
	if status == entity.DagInstanceStatusSuccess {
		dErr := GetStore().DeleteTaskInsByDagInsID(ctx, dagIns.ID)
		if dErr != nil {
			log.Warnf("run success,delete task instance failed: %s", dErr.Error())
		}
	}

	dag, err := GetStore().GetDagWithOptionalVersion(ctx, dagIns.DagID, dagIns.VersionID)
	if err != nil {
		log.Warnf("[logic.AfterStopLog] get dag[%s] failed: %s", dagIns.DagID, err)
		return
	}

	if dag.IsDebug {
		return
	}

	var detail string

	bodyType := common.CompleteTaskWithFailed
	if dagIns.Status == entity.DagInstanceStatusSuccess {
		bodyType = common.CompleteTaskWithSuccess
	}
	detail, _ = common.GetLogBody(bodyType, []interface{}{dag.Name}, []interface{}{})

	object := map[string]interface{}{
		"type":          dag.Trigger,
		"id":            dagIns.ID,
		"dagId":         dagIns.DagID,
		"name":          dag.Name,
		"priority":      dagIns.Priority,
		"status":        status,
		"biz_domain_id": utils.IfNot(dag.BizDomainID == "", common.BizDomainDefaultID, dag.BizDomainID),
	}

	if len(dag.Type) != 0 {
		object["dagType"] = dag.Type
	} else {
		object["dagType"] = common.DagTypeDefault
	}

	if dagIns.EndedAt < dagIns.CreatedAt {
		endedAt := time.Now().Unix()
		object["duration"] = endedAt - dagIns.CreatedAt
	} else {
		object["duration"] = dagIns.EndedAt - dagIns.CreatedAt
	}

	varsGetter := dagIns.VarsGetter()
	userID, _ := varsGetter("operator_id")
	userType, _ := varsGetter("operator_type")

	var userInfo drivenadapters.UserInfo
	userInfo, err0 := drivenadapters.NewUserManagement().GetUserInfoByType(fmt.Sprintf("%v", userID), fmt.Sprintf("%v", userType))
	if err0 != nil {
		log.Warnf("[logic.AfterStopLog] GetUserInfoByType failed: %s", err0.Error())
		userName, _ := varsGetter("operator_name")
		userInfo = drivenadapters.UserInfo{
			UserID:   fmt.Sprintf("%v", userID),
			UserName: fmt.Sprintf("%v", userName),
			Type:     fmt.Sprintf("%v", userType),
		}
	}
	userInfo.VisitorType = common.AuthenticatedUserType

	drivenadapters.NewLogger().LogO11y(&drivenadapters.BuildARLogParams{
		Operation:   common.ArLogEndDagIns,
		Description: detail,
		UserInfo:    &userInfo,
		Object:      object,
	}, &drivenadapters.O11yLogWriter{Logger: traceLog.NewFlowO11yLogger()})
}

var _ hook.BeforeAssign = (*VMExt)(nil)
var _ hook.BeforeReturn = (*VMExt)(nil)
var _ hook.LoopStart = (*VMExt)(nil)
var _ hook.BranchStart = (*VMExt)(nil)
var _ hook.BranchSkip = (*VMExt)(nil)
var _ hook.VMStop = (*VMExt)(nil)
var _ hook.Resume = (*VMExt)(nil)
var _ hook.ResumeError = (*VMExt)(nil)
