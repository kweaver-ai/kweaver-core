package mgnt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	aerr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	liberrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/perm"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm/state"
)

const MaxCallDepth = 10

func (m *mgnt) RunOperator(ctx context.Context,
	id string,
	formData map[string]any,
	successCallback string,
	errorCallback string,
	parentDagInsID string,
	userInfo *drivenadapters.UserInfo) (dagIns *entity.DagInstance, vmIns *mod.VMExt, err error) {

	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dag, err := m.mongo.GetDag(ctx, id)
	if err != nil {
		if aerr.IsNotFoundErr(err) {
			return nil, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorNotFound,
				liberrors.PErrorNotFound,
				map[string]string{"dagId": id})
		}

		log.Warnf("[logic.RunFormInstanceV2] GetDagByFields err, deail: %s", err.Error())
		return nil, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorInternalServerError,
			liberrors.PErrorInternalServerError,
			nil)
	}

	if dag.Type != common.DagTypeComboOperator {
		return nil, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorForbidden,
			liberrors.PErrorForbidden,
			map[string]any{
				"type": "dag type must be combo-operator",
			})
	}

	if userInfo.AccountType != common.APP.ToString() {
		opMap := &perm.MapOperationProvider{
			OpMap: map[string][]string{
				common.DagTypeComboOperator: {perm.OpExecuteOperation},
			},
		}

		_, err = m.permCheck.CheckDagAndPerm(ctx, dag.ID, userInfo, opMap)
		if err != nil {
			log.Warnf("[logic.RunFormInstanceV2] CheckDagAndPerm failed, err: %s", err.Error())
			return nil, nil, err
		}
	}

	return m.runFormInstanceVM(ctx, dag, formData, successCallback, errorCallback, parentDagInsID, userInfo, userInfo.UserID)
}

func (m *mgnt) runFormInstanceVM(
	ctx context.Context,
	dag *entity.Dag,
	formData map[string]interface{},
	successCallback, errorCallback string,
	parentDagInsID string,
	userInfo *drivenadapters.UserInfo,
	userID string,
) (dagIns *entity.DagInstance, vmIns *mod.VMExt, err error) {
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	triggerType := m.getTriggerType(dag.Steps[0].Operator)
	if triggerType != entity.TriggerForm {
		return nil, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorForbidden,
			liberrors.PErrorForbidden, map[string]interface{}{
				"trigger": fmt.Sprintf("%s trigger type is not allowed to run form", triggerType),
			})
	}

	callChain := []string{dag.ID}

	if parentDagInsID != "" {
		parentDagIns, err1 := m.mongo.GetDagInstance(ctx, parentDagInsID)
		if err1 == nil {
			callChain = append(parentDagIns.CallChain, dag.ID)
			if len(callChain) > MaxCallDepth {
				return nil, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorBadRequest,
					liberrors.PErrorBadRequest, map[string]interface{}{
						"message":        fmt.Sprintf("Maximum call chain depth exceeded (limit: %d).", MaxCallDepth),
						"max_call_depth": MaxCallDepth,
					},
				)
			}
		}
	}

	dag.SetPushMessage(m.executeMethods.Publish)
	bytes, _ := json.Marshal(formData)

	runVar := map[string]string{
		"userid":        userInfo.UserID,
		"operator_id":   userInfo.UserID,
		"operator_name": userInfo.UserName,
		"operator_type": userInfo.AccountType,
		"run_mode":      "vm",
		"run_args":      string(bytes),
	}

	dagIns, dagErr := dag.Run(ctx, triggerType, runVar, entity.WithKeyWords(userInfo.UserName))
	if dagErr != nil {
		return nil, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorForbidden,
			liberrors.PErrorForbidden,
			map[string]interface{}{"id:": dag.ID, "status": dag.Status})
	}

	dagIns.Mode = entity.DagInstanceModeVM
	dagIns.Status = entity.DagInstanceStatusRunning
	dagIns.SuccessCallback = successCallback
	dagIns.ErrorCallback = errorCallback
	dagIns.CallChain = callChain

	// Initialize ShareData if it doesn't exist
	if dagIns.ShareData == nil {
		dagIns.ShareData = &entity.ShareData{
			Dict:        map[string]interface{}{},
			DagInstance: dagIns,
		}
	}

	dagIns.ID, err = m.mongo.CreateDagIns(ctx, dagIns)

	if err != nil {
		log.Warnf("[logic.runFormInstanceVM] CreateDagIns err, deail: %s", err.Error())
		return nil, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorInternalServerError,
			liberrors.PErrorInternalServerError, nil,
		)
	}

	if dag.ExecMode == "async" {
		go func() {
			vmIns = mod.NewVMExt(context.Background(), dagIns, userID)
			_ = vmIns.Boot()
		}()
		return dagIns, nil, nil
	} else {
		vmIns = mod.NewVMExt(ctx, dagIns, userID)
		err = vmIns.Boot()
		if err != nil {
			return nil, nil, err
		}
	}

	go func() {
		detail, extMsg := common.GetLogBody(common.TriggerTaskManually, []interface{}{dag.Name},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: dag.ID,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dagIns.ID,
			Operation: drivenadapters.ExecuteOperation,
			ObjID:     dagIns.DagID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}()

	return dagIns, vmIns, nil
}

func (m *mgnt) GetDagInstanceResultVM(ctx context.Context, id string, userInfo *drivenadapters.UserInfo) (state.State, any, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dagIns, err := m.mongo.GetDagInstance(ctx, id)
	if err != nil {
		log.Warnf("[logic.GetDagInstanceResultVM] GetDagInstance err: %s", err.Error())
		if aerr.IsNotFoundErr(err) {
			return 0, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorNotFound,
				liberrors.PErrorNotFound, nil)
		}
		return 0, nil, err
	}

	if dagIns.DagType != common.DagTypeComboOperator || dagIns.Mode != entity.DagInstanceModeVM {
		log.Warnf("[logic.GetDagInstanceResultVM] invalid dagIns dagType %s, mode %s", dagIns.DagType, dagIns.Mode)
		return 0, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorBadRequest,
			liberrors.PErrorBadRequest, nil)
	}

	if dagIns.UserID != userInfo.UserID {
		return 0, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorForbidden,
			liberrors.PErrorForbidden, nil)
	}

	err = dagIns.LoadExtData(ctx)

	if err != nil {
		log.Warnf("[logic.GetDagInstanceResultVM] LoadExtData err: %s", err.Error())
		return 0, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorInternalServerError,
			liberrors.PErrorInternalServerError, "invalid dagIns")
	}

	switch dagIns.Status {
	case entity.DagInstanceStatusInit, entity.DagInstanceStatusScheduled, entity.DagInstanceStatusRunning:
		return state.Run, nil, nil
	default:
		vmIns := mod.NewVMExt(ctx, dagIns, dagIns.UserID)
		err = json.Unmarshal([]byte(dagIns.Dump), vmIns)

		if err != nil {
			log.Warnf("[logic.GetDagInstanceResultVM] Unmarshal dump err: %s", err.Error())
			return 0, nil, liberrors.NewPublicRestError(ctx, liberrors.PErrorInternalServerError,
				liberrors.PErrorInternalServerError, "invalid dagIns")
		}
		return vmIns.Result()
	}
}
