package mgnt

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
	normalizeutil "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils/normalize"
	"go.mongodb.org/mongo-driver/bson"
)

type CreateFlowParams struct {
	Steps      []entity.Step `json:"steps"`
	PolicyType string        `json:"policy_type"`
}

type Flow struct {
	ID         string        `json:"id"`
	Steps      []entity.Step `json:"steps"`
	PolicyType string        `json:"policy_type"`
}

type FormField struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required,omitempty"`
}

type ProcParams struct {
	UserID string                 `json:"user_id"`
	FlowID string                 `json:"flow_id"`
	Source interface{}            `json:"source"`
	Values map[string]interface{} `json:"values"`
}

func (m *mgnt) CreateSecurityPolicyFlow(ctx context.Context, param *CreateFlowParams, userInfo *drivenadapters.UserInfo) (string, error) {
	var dagID string
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var tasks = make([]entity.Task, 0)
	var stepList = make([]map[string]interface{}, 0)
	steps := make([]entity.Step, len(param.Steps))
	copy(steps, param.Steps)
	m.buildTasks(&steps[0], steps, &tasks, nil, &stepList, nil, nil)

	err = m.validSteps(&Validate{
		Ctx:         ctx,
		Steps:       stepList,
		IsAdminRole: true,
		UserInfo:    userInfo,
		ErrType:     ErrTypeV1,
		ParseFunc:   common.JSONSchemaValid,
	}).BuildError()
	if err != nil {
		return dagID, err
	}

	trigger := m.getTriggerType(param.Steps[0].Operator)

	if trigger != entity.TriggerSecurityPolicy {
		log.Warnf("[logic.CreateSecurityPolicy] Invalid Security Policy Trigger: %s", param.Steps[0].Operator)
		return dagID, ierrors.NewIError(ierrors.InvalidParameter, ierrors.ErrorIncorretOperator, map[string]string{"operator": param.Steps[0].Operator})
	}

	dag := &entity.Dag{
		UserID: userInfo.UserID,
		Name:   "",
		Vars: entity.DagVars{
			"userid": {DefaultValue: userInfo.UserID},
			"docid":  {DefaultValue: ""},
		},
		Trigger:     trigger,
		Tasks:       tasks,
		Steps:       param.Steps,
		Description: "",
		Status:      entity.DagStatusNormal,
		Shortcuts:   nil,
		Accessors:   nil,
		Type:        common.DagTypeSecurityPolicy,
		PolicyType:  param.PolicyType,
		Priority:    common.PriorityHighest,
		Removed:     false,
	}

	dag.SetCheckRootNode(func(t []entity.Task) error {
		_, bErr := mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
		if bErr != nil {
			return bErr
		}
		return nil
	})

	dagID, err = m.mongo.CreateDag(ctx, dag)

	if err != nil {
		log.Warnf("[logic.CreateSecurityPolicy] CreateDag err, deail: %s", err.Error())
		return dagID, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorln(err)
			}
		}()
		detail, extMsg := common.GetLogBody("CreateSecurityPolicyFlow", []interface{}{dag.ID},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: dag.ID,
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		}, &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish})
	}()

	return dagID, err
}

func (m *mgnt) UpdateSecurityPolicyFlow(ctx context.Context, dagID string, steps []entity.Step, userInfo *drivenadapters.UserInfo) error {
	var err error
	var appCountInfoMap = make(map[string]string, 0)

	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	query := map[string]interface{}{
		"_id":  dagID,
		"type": common.DagTypeSecurityPolicy,
		"removed": bson.M{
			"$ne": true,
		},
	}

	dag, err := m.mongo.GetDagByFields(ctx, query)

	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.TaskNotFound, "", nil)
		}

		log.Warnf("[logic.UpdateSecurityPolicyFlow] GetDagByFields err, id: %s, deail: %s", dagID, err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	resetPermApplyStepAppPwd(appCountInfoMap, dag.Steps, steps)
	var tasks = make([]entity.Task, 0)
	var stepList = make([]map[string]interface{}, 0)
	dag.Steps = steps
	m.buildTasks(&steps[0], steps, &tasks, nil, &stepList, nil, nil)

	err = m.validSteps(&Validate{
		Ctx:         ctx,
		Steps:       stepList,
		IsAdminRole: true,
		UserInfo:    userInfo,
		ErrType:     ErrTypeV1,
		ParseFunc:   common.JSONSchemaValid,
	}).BuildError()

	if err != nil {
		return err
	}

	dag.Tasks = tasks

	err = m.mongo.UpdateDag(ctx, dag)

	if err != nil {
		log.Warnf("[logic.UpdateSecurityPolicyFlow] UpdateDag err, id: %s, deail: %s", dagID, err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorln(err)
			}
		}()
		detail, extMsg := common.GetLogBody("UpdateSecurityPolicyFlow", []interface{}{dag.ID},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: dag.ID,
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		}, &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish})
	}()

	return err
}

func (m *mgnt) DeleteSecurityPolicyFlow(ctx context.Context, dagID string, userInfo *drivenadapters.UserInfo) error {
	var err error

	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	query := map[string]interface{}{
		"_id":  dagID,
		"type": common.DagTypeSecurityPolicy,
		"removed": bson.M{
			"$ne": true,
		},
	}

	dag, err := m.mongo.GetDagByFields(ctx, query)

	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.TaskNotFound, "", nil)
		}

		log.Warnf("[logic.DeleteSecurityPolicyFlow] GetDagByFields err, id: %s, deail: %s", dagID, err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	dag.Removed = true
	err = m.mongo.UpdateDag(ctx, dag)

	if err != nil {
		log.Warnf("[Logic.DeleteSecurityPolicyFlow] UpdateDag err, id: %s, deail: %s", dagID, err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	if userInfo != nil {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Errorln(err)
				}
			}()
			detail, extMsg := common.GetLogBody("DeleteSecurityPolicyFlow", []interface{}{dag.ID},
				[]interface{}{})
			log.Infof("detail: %s, extMsg: %s", detail, extMsg)
			m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
				UserInfo: userInfo,
				Msg:      detail,
				ExtMsg:   extMsg,
				OutBizID: dag.ID,
				LogLevel: drivenadapters.NcTLogLevel_NCT_LL_WARN,
			}, &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish})
		}()
	}

	return err
}

func (m *mgnt) GetSecurityPolicyFlowByID(ctx context.Context, dagID string) (flow Flow, err error) {
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	query := map[string]interface{}{
		"_id":  dagID,
		"type": common.DagTypeSecurityPolicy,
		"removed": bson.M{
			"$ne": true,
		},
	}

	dag, err := m.mongo.GetDagByFields(ctx, query)

	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			err = ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"id": dagID})
		}

		log.Warnf("[logic.GetSecurityPolicyFlowByID] GetDagByFields err, id: %s, deail: %s", dagID, err.Error())
		return
	}

	for _, task := range dag.Tasks {
		if task.ActionName == common.AnyshareFileSetPermOpt {
			delete(task.Params, "apppwd")
		}
	}

	for _, step := range dag.Steps {
		if step.Operator == common.AnyshareFileSetPermOpt {
			delete(step.Parameters, "apppwd")
		}
	}

	flow.ID = dag.ID
	flow.Steps = dag.Steps
	flow.PolicyType = dag.PolicyType

	return
}

func (m *mgnt) StartSecurityPolicyFlowProc(ctx context.Context, params ProcParams) (string, error) {
	var pid string
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	fmt.Printf("[logic.StartSecurityPolicyFlowProc] params: %+v", params)
	query := map[string]interface{}{
		"_id":  params.FlowID,
		"type": common.DagTypeSecurityPolicy,
		"removed": bson.M{
			"$ne": true,
		},
	}

	dag, err := m.mongo.GetDagByFields(ctx, query)

	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			log.Warnf("[logic.StartSecurityPolicyFlowProc] GetDagByFields err, id: %s, deail: %s", params.FlowID, err.Error())
			return pid, ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"id": params.FlowID})
		}

		log.Warnf("[logic.RunSecurityPolicyFlow] GetDagByFields err, id: %s, deail: %s", params.FlowID, err.Error())
		return pid, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	userDetail, err := m.usermgnt.GetUserInfo(params.UserID)

	if err != nil {
		log.Warnf("[logic.RunSecurityPolicyFlow] GetUserInfo err, id: %s, deail: %s", params.UserID, err.Error())
		return pid, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	runVar := map[string]string{
		"userid":        userDetail.UserID,
		"operator_id":   userDetail.UserID,
		"operator_name": userDetail.UserName,
		"operator_type": common.User.ToString(),
		"dag_type":      common.DagTypeSecurityPolicy,
		"policy_type":   dag.PolicyType,
	}

	triggerType := m.getTriggerType(dag.Steps[0].Operator)

	if triggerType != entity.TriggerSecurityPolicy {
		err = ierrors.NewIError(ierrors.Forbidden, ierrors.ErrorIncorretTrigger, map[string]interface{}{
			"trigger": fmt.Sprintf("invalid security policy, trigger operator: %s", dag.Steps[0].Operator),
		})
		return pid, err
	}

	if params.Source != nil {
		bytes, _ := json.Marshal(params.Source)
		runVar["source"] = string(bytes)
	}

	if fields, ok := normalizeutil.AsSlice(dag.Steps[0].Parameters["fields"]); ok {
		for _, f := range fields {
			if field, ok := f.(map[string]interface{}); ok {
				key := field["key"].(string)
				typ := field["type"].(string)
				required, ok := field["required"].(bool)
				required = ok && required

				if params.Values[key] == nil {

					if required {
						log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is required", key)
						return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("field: %s is required", key)})
					}

					continue
				}

				switch typ {
				case "string", "long_string":
					if val, ok := params.Values[key].(string); ok {
						runVar[key] = val
					} else {
						log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is invalid", key)
						return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)})
					}

				case "number":
					if val, ok := params.Values[key].(int64); ok {
						runVar[key] = fmt.Sprintf("%v", val)
					} else if val, ok := params.Values[key].(float64); ok {
						runVar[key] = fmt.Sprintf("%v", val)
					} else {
						log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is invalid", key)
						return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)})
					}

				case "asFile", "asFolder", "asDoc":
					if val, ok := params.Values[key].(string); ok && utils.IsGNS(val) {
						runVar[key] = val
					} else {
						log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is invalid", key)
						return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)})
					}

				case "asTags":
					if val, ok := params.Values[key].([]interface{}); ok {
						var tags = make([]string, 0)
						for _, v := range val {
							tags = append(tags, fmt.Sprintf("%v", v))
						}
						bytes, _ := json.Marshal(tags)
						runVar[key] = string(bytes)
					} else {
						log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is invalid", key)
						return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)})
					}

				case "asLevel":
					if val, ok := params.Values[key].(int64); ok {
						runVar[key] = fmt.Sprintf("%v", val)
					} else if val, ok := params.Values[key].(map[string]interface{}); ok {
						bytes, _ := json.Marshal(val)
						err := common.JSONSchemaValid(bytes, "values/as-level-info.json")

						if err != nil {

							log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is invalid", key)
							return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)})
						}

						runVar[key] = string(bytes)
					} else {
						log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is invalid", key)
						return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)})
					}

				case "asMetadata":
					{
						bytes, _ := json.Marshal(params.Values[key])
						err := common.JSONSchemaValid(bytes, "values/as-metadata.json")

						if err != nil {
							log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is invalid", key)
							return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)})
						}

						runVar[key] = string(bytes)
					}

				case "asPerm":
					{
						bytes, _ := json.Marshal(params.Values[key])
						err := common.JSONSchemaValid(bytes, "values/as-perm.json")
						if err != nil {
							log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is invalid", key)
							return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)})
						}
						permOrigin, _ := params.Values[key].(map[string]interface{})
						_allows := permOrigin["allow"].([]interface{})
						foundDisplay := false
						for _, al := range _allows {
							if al == "display" {
								foundDisplay = true
								break
							}
						}

						if !foundDisplay {
							_allows = append(_allows, "display")
							permOrigin["allow"] = _allows
							bytes, _ = json.Marshal(permOrigin)
						}

						runVar[key] = string(bytes)
					}
				case "asAccessorPerms":
					{
						bytes, _ := json.Marshal(params.Values[key])
						err := common.JSONSchemaValid(bytes, "values/as-accessor-perms.json")

						runVar[key] = string(bytes)
						if err != nil {
							log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is invalid", key)
							return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)})
						}
					}

				case "asSpaceQuota":
					{
						switch params.Values[key].(type) {
						case float64:
							runVar[key] = strconv.FormatFloat(params.Values[key].(float64), 'f', -1, 64)
						default:
							log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is invalid", key)
							return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)})
						}
					}

				case "asAllowSuffixDoc":
					{
						bytes, _ := json.Marshal(params.Values[key])
						err := common.JSONSchemaValid(bytes, "values/as-allow-suffix-doc.json")
						runVar[key] = string(bytes)
						if err != nil {
							log.Warnf("[logic.RunSecurityPolicyFlow] field: %s is invalid", key)
							return pid, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)})
						}
					}

				default:
					{
						switch v := params.Values[key].(type) {
						case string, int, float64:
							runVar[key] = fmt.Sprintf("%v", v)
						case map[string]interface{}:
							bytes, _ := json.Marshal(v)
							runVar[key] = string(bytes)
						default:
							runVar[key] = fmt.Sprintf("%v", v)
						}
					}
				}
			}
		}
	}

	dagIns, dagErr := dag.Run(ctx, triggerType, runVar, entity.WithKeyWords(userDetail.UserName))
	if dagErr != nil {
		log.Warnf("[logic.RunSecurityPolicyFlow] dag.Run err, deail: %s", dagErr.Error())
		return pid, ierrors.NewIError(ierrors.Forbidden, ierrors.DagStatusNotNormal, map[string]interface{}{"id:": dag.ID, "status": dag.Status})
	}

	// worker := mod.GetKeeper().WorkerKey()

	// dagIns.Worker = worker
	// dagIns.Status = entity.DagInstanceStatusScheduled
	daginstances := []*entity.DagInstance{dagIns}
	pids, err := m.mongo.BatchCreateDagIns(ctx, daginstances)

	if err != nil {
		log.Warnf("[logic.RunSecurityPolicyFlow] BatchCreateDagIns err, deail: %s", err.Error())

		// parser := mod.GetParser()
		// err := parser.RunDagIns(dagIns)

		// if err != nil {
		// 	return pid, ierrors.NewIError(ierrors.InternalError, "", nil)
		// }
	}
	pid = pids[0].ID

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorln(err)
			}
		}()
		userInfo := &drivenadapters.UserInfo{
			UserID: params.UserID,
			UdID:   userDetail.UdID,
		}

		detail, extMsg := common.GetLogBody("RunSecurityPolicyFlow", []interface{}{dag.ID},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: dag.ID,
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_WARN,
		}, &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish})
	}()

	return pid, err
}

func (m *mgnt) StopSecurityPolicyFlowProc(ctx context.Context, pid string, userInfo *drivenadapters.UserInfo) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dagIns, err := m.mongo.GetDagInstance(ctx, pid)

	if err != nil {
		log.Warnf("[logic.StopSecurityPolicyFlowProc] GetDagInstance err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err.Error())
	}

	if userInfo != nil && dagIns.UserID != userInfo.UserID {
		return ierrors.NewIError(ierrors.DagInsNotFound, "", map[string]string{"dagInsID": pid})
	}

	query := map[string]interface{}{
		"_id":  dagIns.DagID,
		"type": common.DagTypeSecurityPolicy,
	}

	_, err = m.mongo.GetDagByFields(ctx, query)

	if err != nil {

		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.DagInsNotFound, "", map[string]string{"dagInsID": pid})
		}
		log.Warnf("[logic.StopSecurityPolicyFlowProc] GetDagByFields err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	if dagIns.Status == entity.DagInstanceStatusSuccess ||
		dagIns.Status == entity.DagInstanceStatusFailed ||
		dagIns.Status == entity.DagInstanceStatusCancled ||
		dagIns.Status == "" {
		return ierrors.NewIError(ierrors.Forbidden, ierrors.DagInsNotRunning, map[string]string{"status": fmt.Sprintf("status: %s, dag Ins is success failed cancled or ' '", dagIns.Status)})
	}

	dagIns.Status = entity.DagInstanceStatusCancled
	dagIns.EndedAt = time.Now().Unix()

	taskIns, err := m.mongo.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{
		DagInsID: dagIns.ID,
		Status:   []entity.TaskInstanceStatus{entity.TaskInstanceStatusInit, entity.TaskInstanceStatusBlocked},
	})

	if err != nil {
		log.Warnf("[logic.StopSecurityPolicyFlowProc] ListTaskInstance err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	for _, taskIn := range taskIns {

		if taskIn.Status == entity.TaskInstanceStatusBlocked {
			act := mod.ActionMap[taskIn.ActionName]
			if act == nil {
				log.Warnf("[logic.StopSecurityPolicyFlowProc] action not found, deail: %s", taskIn.ActionName)
				continue
			}

			cancelAct, ok := act.(entity.CancelAction)
			if ok {
				err := cancelAct.Cancel(*taskIn, m.executeMethods)

				if err != nil {
					log.Warnf("[logic.StopSecurityPolicyFlowProc] Cancel err, deail: %s", err.Error())
					continue
				}
			}
		}

		taskIn.Status = entity.TaskInstanceStatusCanceled
		err = m.mongo.UpdateTaskIns(ctx, taskIn)

		if err != nil {
			log.Warnf("[logic.StopSecurityPolicyFlowProc] UpdateTaskIns err, deail: %s", err.Error())
			continue
		}
	}

	err = m.mongo.UpdateDagIns(ctx, dagIns)

	if err != nil {
		log.Warnf("[logic.StopSecurityPolicyFlowProc] UpdateDagIns err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 发布安全策略消息
	if dagIns.DagType == common.DagTypeSecurityPolicy {
		msg := &mod.SecurityPolicyProcResultMsg{
			PID:        dagIns.ID,
			Result:     "canceled",
			PolicyType: dagIns.PolicyType,
		}
		result, _ := jsoniter.Marshal(msg)
		m.executeMethods.Publish(common.TopicSecurityPolicyProcResult, result)
	}

	return nil
}
