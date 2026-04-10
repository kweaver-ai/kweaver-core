package mgnt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	aerr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var triggerHandlers = map[string]TriggerHandler{}
var AnyshareSelectedTriggerList = []string{common.OpAnyShareSelectedFileTrigger, common.OpAnyShareSelectedFolderTrigger}

type (
	Efast          = drivenadapters.Efast
	UserManagement = drivenadapters.UserManagement
	Uniquery       = drivenadapters.UniqueryDriven
)

type SingleDeBugReq struct {
	ID          string         `json:"id"`
	Operator    string         `json:"operator"`
	Parameters  map[string]any `json:"parameters"`
	BizDomainID string         `json:"-"`
}

// DeBugDagTemplateReq 调试数据流程请求体
type FullDeBugReq struct {
	ID            string                 `json:"id"`
	Steps         []entity.Step          `json:"steps"`
	TriggerConfig *entity.TriggerConfig  `json:"trigger_config"`
	TriggerData   map[string]interface{} `json:"trigger_data"`
	BizDomainID   string                 `json:"-"`
}

// TriggerHandler 不同类型的数据源运行参数校验器
type TriggerHandler interface {
	Key() string
	Handle(ctx context.Context, user *drivenadapters.UserInfo, triggerStep entity.Step, triggerData map[string]interface{}, runVar map[string]string) (entity.Trigger, error)
}

// RegisterTriggerHandler 注册校验器
func RegisterTriggerHandler(h TriggerHandler) {
	triggerHandlers[h.Key()] = h
}

func RegisterTriggerHandlers(efast Efast, usermgnt UserManagement, uniquery Uniquery) {
	RegisterTriggerHandler(FormTriggerHandler{})
	RegisterTriggerHandler(TriggerDataFlowDocHandler{efast: efast})
	RegisterTriggerHandler(TriggerDataFlowDeptHandler{usermgnt: usermgnt})
	RegisterTriggerHandler(TriggerDataFlowUserHandler{usermgnt: usermgnt})
	RegisterTriggerHandler(TriggerDataFlowDataViewHandler{uniquery: uniquery})
	RegisterTriggerHandler(TriggerDataFlowOperatorHandler{})
	RegisterTriggerHandler(TriggerDataFlowEventHandler{efast: efast})
	RegisterTriggerHandler(TriggerDataFlowCronHandler{efast: efast})
	RegisterTriggerHandler(TriggerDataFlowManuallyHandler{efast: efast})
}

// HandleTriggerRunVar 校验器入口方法
func HandleTriggerRunVar(ctx context.Context, operator string, triggerStep entity.Step, user *drivenadapters.UserInfo, triggerData map[string]interface{}, runVar map[string]string) (entity.Trigger, error) {
	if strings.HasPrefix(operator, common.OperatorTrigger) {
		operator = common.OperatorTrigger
	}

	if utils.Contains(AnyshareEventTriggerList, operator) {
		operator = common.EventTrigger
	}

	if utils.Contains(AnyshareSelectedTriggerList, operator) {
		operator = common.MannualTrigger
	}

	if strings.HasPrefix(operator, common.CronTrigger) {
		operator = common.CronTrigger
	}

	if h, ok := triggerHandlers[operator]; ok {
		return h.Handle(ctx, user, triggerStep, triggerData, runVar)
	}
	return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, fmt.Sprintf("unsupported trigger: %s", operator))
}

// CheckTrigger 检查触发类型
func CheckTrigger(ctx context.Context, operator string) error {
	if _, ok := triggerHandlers[operator]; !ok {
		return ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, fmt.Sprintf("unsupported trigger: %s", operator))
	}
	return nil
}

// FormTriggerHandler 表单触发类型
type FormTriggerHandler struct{}

func (FormTriggerHandler) Key() string { return common.FormTrigger }

func (FormTriggerHandler) Handle(ctx context.Context, user *drivenadapters.UserInfo, triggerStep entity.Step, triggerData map[string]interface{}, runVar map[string]string) (entity.Trigger, error) {
	if len(triggerData) == 0 {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{
			"trigger_type": entity.TriggerForm,
			"operator":     common.FormTrigger,
			"trigger_data": "trigger_data is required",
		})
	}

	for key, val := range triggerData {
		switch x := val.(type) {
		case string, int, float64, bool:
			runVar[key] = fmt.Sprintf("%v", x)
		case map[string]interface{}, []interface{}, []string, []map[string]interface{}:
			b, _ := json.Marshal(x)
			runVar[key] = string(b)
		default:
			runVar[key] = fmt.Sprintf("%v", x)
		}
	}

	return entity.TriggerForm, nil
}

// TriggerDataFlowDocHandler 非结构化数据触发
type TriggerDataFlowDocHandler struct {
	efast drivenadapters.Efast
}

func (TriggerDataFlowDocHandler) Key() string { return common.DataflowDocTrigger }

func (d TriggerDataFlowDocHandler) Handle(ctx context.Context, user *drivenadapters.UserInfo, triggerStep entity.Step, triggerData map[string]interface{}, runVar map[string]string) (entity.Trigger, error) {
	id, ok := triggerData["id"].(string)
	if !ok || strings.TrimSpace(id) == "" {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{
			"trigger_type": entity.TriggerManually,
			"operator":     common.DataflowDocTrigger,
			"trigger_data": "docid is required",
		})
	}

	_, err := d.efast.GetFileAttr(ctx, id, strings.TrimPrefix(user.TokenID, "Bearer "), user.LoginIP)
	if err != nil {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, err.Error())
	}

	runVar["docid"] = id

	return entity.TriggerManually, nil
}

// TriggerDataFlowDeptHandler 结构化数据触发(部门)
type TriggerDataFlowDeptHandler struct {
	usermgnt drivenadapters.UserManagement
}

func (TriggerDataFlowDeptHandler) Key() string { return common.DataflowDeptTrigger }

func (d TriggerDataFlowDeptHandler) Handle(ctx context.Context, user *drivenadapters.UserInfo, triggerStep entity.Step, triggerData map[string]interface{}, runVar map[string]string) (entity.Trigger, error) {
	id, ok := triggerData["id"].(string)
	if !ok || strings.TrimSpace(id) == "" {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{
			"trigger_type": entity.TriggerManually,
			"operator":     common.DataflowDeptTrigger,
			"trigger_data": "deptmentid is required",
		})
	}

	_, err := d.usermgnt.GetDepartmentInfo(id)
	if err != nil {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, err.Error())
	}

	runVar["id"] = id

	return entity.TriggerManually, nil
}

// TriggerDataFlowUserHandler 结构化数据触发(用户)
type TriggerDataFlowUserHandler struct {
	usermgnt drivenadapters.UserManagement
}

func (TriggerDataFlowUserHandler) Key() string { return common.DataflowUserTrigger }

func (u TriggerDataFlowUserHandler) Handle(ctx context.Context, user *drivenadapters.UserInfo, triggerStep entity.Step, triggerData map[string]interface{}, runVar map[string]string) (entity.Trigger, error) {
	id, ok := triggerData["id"].(string)
	if !ok || strings.TrimSpace(id) == "" {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{
			"trigger_type": entity.TriggerManually,
			"operator":     common.DataflowUserTrigger,
			"trigger_data": "userid is required",
		})
	}

	_, err := u.usermgnt.GetUserInfo(id)
	if err != nil {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, err.Error())
	}

	runVar["id"] = id

	return entity.TriggerManually, nil
}

// TriggerDataFlowDataViewHandler 数据视图触发
type TriggerDataFlowDataViewHandler struct {
	uniquery drivenadapters.UniqueryDriven
}

func (dv TriggerDataFlowDataViewHandler) Key() string { return common.MDLDataViewTrigger }

func (dv TriggerDataFlowDataViewHandler) Handle(ctx context.Context, user *drivenadapters.UserInfo, triggerStep entity.Step, triggerData map[string]interface{}, runVar map[string]string) (entity.Trigger, error) {
	var param DataViewParam
	err := param.Init(triggerStep.Parameters)
	if err != nil {
		return entity.TriggerManually, err
	}

	data, err := dv.uniquery.UniqueryDataView(ctx, param.ID, &drivenadapters.UniqueryDataViewOptions{
		Start:          0,
		End:            0,
		Limit:          1,
		NeedTotal:      false,
		UseSearchAfter: true,
		SearchAfter:    []any{},
		Format:         "original",
	}, user.UserID, user.AccountType)

	if err != nil {
		log.Warnf("[logic.TriggerDataFlowDataViewHandler] Handle err, detail: %s", err.Error())
		return entity.TriggerManually, err
	}

	if len(data.Entries) == 0 {
		return entity.TriggerManually, nil
	}

	entryBytes, _ := json.Marshal(data.Entries[0:1])
	runVar["mdl_data"] = string(entryBytes)

	return entity.TriggerManually, nil
}

// TriggerDataFlowOperatorHandler 基础算子触发
type TriggerDataFlowOperatorHandler struct{}

func (TriggerDataFlowOperatorHandler) Key() string { return common.OperatorTrigger }

func (TriggerDataFlowOperatorHandler) Handle(ctx context.Context, user *drivenadapters.UserInfo, triggerStep entity.Step, triggerData map[string]interface{}, runVar map[string]string) (entity.Trigger, error) {
	return entity.TriggerManually, nil
}

// TriggerDataFlowEventHandler 事件触发
type TriggerDataFlowEventHandler struct {
	efast drivenadapters.Efast
}

func (e TriggerDataFlowEventHandler) Key() string { return common.EventTrigger }

func (e TriggerDataFlowEventHandler) Handle(ctx context.Context, user *drivenadapters.UserInfo, triggerStep entity.Step, triggerData map[string]interface{}, runVar map[string]string) (entity.Trigger, error) {
	id, ok := triggerData["id"].(string)
	if !ok || strings.TrimSpace(id) == "" {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{
			"trigger_type": entity.TriggerManually,
			"operator":     common.DataflowDocTrigger,
			"trigger_data": "docid is required",
		})
	}

	runVar["id"] = id
	runVar["docid"] = id
	runVar["new_id"] = id
	runVar["source_type"] = "doc"

	res, err := e.efast.CheckPerm(ctx, id, "display", strings.TrimPrefix(user.TokenID, "Bearer "), user.LoginIP)
	if err != nil {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}

	if res != 0 {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorForbidden, ierr.PErrorForbidden, map[string]interface{}{
			"info": "has no perm to get doc metadata",
			"doc": map[string]string{
				"docid": id,
			},
		})
	}

	attr, err := e.efast.GetDocMsg(ctx, id)
	if err != nil {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}
	runVar["name"] = attr.Name
	runVar["path"] = attr.Path
	runVar["new_path"] = attr.Path
	runVar["size"] = fmt.Sprintf("%v", attr.Size)

	return entity.TriggerManually, nil
}

// TriggerDataFlowCronHandler 定时触发
type TriggerDataFlowCronHandler struct {
	efast drivenadapters.Efast
}

func (c TriggerDataFlowCronHandler) Key() string { return common.CronTrigger }

func (c TriggerDataFlowCronHandler) Handle(ctx context.Context, user *drivenadapters.UserInfo, triggerStep entity.Step, triggerData map[string]interface{}, runVar map[string]string) (entity.Trigger, error) {
	dataSource := triggerStep.DataSource
	if dataSource == nil {
		return entity.TriggerManually, nil
	}

	docIDs := dataSource.Parameters.DocIDs

	if len(docIDs) == 0 {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{
			"trigger_type": entity.TriggerManually,
			"operator":     common.OpAnyShareSelectedFileTrigger,
			"info":         "docids empty",
		})
	}

	for _, id := range docIDs {
		// 权限校验
		res, err := c.efast.CheckPerm(ctx, id, "display", strings.TrimPrefix(user.TokenID, "Bearer "), user.LoginIP)
		if err != nil {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
		}
		if res != 0 {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorForbidden, ierr.PErrorForbidden, map[string]interface{}{
				"info": "has no perm to get doc metadata",
				"doc": map[string]string{
					"docid": id,
				},
			})
		}
	}

	switch dataSource.Operator {
	case common.AnyshareDataSpecifyFiles, common.AnyshareDataSpecifyFolders, common.AnyshareDataListFolders:
		runVar["docid"] = docIDs[0]
	case common.AnyshareDataListFiles:
		files, _, err := c.efast.ListDir(ctx, docIDs[0], strings.TrimPrefix(user.TokenID, "Bearer "), user.LoginIP)
		if err != nil {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
		}

		if len(files) == 0 {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{
				"trigger_type": entity.TriggerManually,
				"operator":     common.OpAnyShareSelectedFileTrigger,
				"info":         "doc not found",
			})
		}
		info := files[0].(map[string]interface{})
		runVar["docid"] = info["docid"].(string)
	default:
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, fmt.Sprintf("unsupported operator type: %v", dataSource.Operator))
	}

	return entity.TriggerManually, nil
}

// 手动触发
type TriggerDataFlowManuallyHandler struct {
	efast drivenadapters.Efast
}

func (m TriggerDataFlowManuallyHandler) Key() string { return common.MannualTrigger }

func (m TriggerDataFlowManuallyHandler) Handle(ctx context.Context, user *drivenadapters.UserInfo, triggerStep entity.Step, triggerData map[string]interface{}, runVar map[string]string) (entity.Trigger, error) {
	var id string
	var docIDs []interface{}
	var ok bool
	params := triggerStep.Parameters

	if triggerStep.Operator == common.OpAnyShareSelectedFileTrigger || triggerStep.Operator == common.OpAnyShareSelectedFolderTrigger {
		docIDs, ok = params["docids"].([]interface{})
		if !ok {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{
				"trigger_type": entity.TriggerManually,
				"operator":     triggerStep.Operator,
				"info":         "docids is required",
			})
		}

		if len(docIDs) == 0 {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{
				"trigger_type": entity.TriggerManually,
				"operator":     triggerStep.Operator,
				"info":         "docids empty",
			})
		}

		for _, id := range docIDs {
			res, err := m.efast.CheckPerm(ctx, fmt.Sprintf("%v", id), "display", strings.TrimPrefix(user.TokenID, "Bearer "), user.LoginIP)
			if err != nil {
				return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
			}

			if res != 0 {
				return "", ierr.NewPublicRestError(ctx, ierr.PErrorForbidden, ierr.PErrorForbidden, map[string]interface{}{
					"info": "has no perm to get doc metadata",
					"doc": map[string]string{
						"docid": fmt.Sprintf("%v", id),
					},
				})
			}
		}
	}

	switch triggerStep.Operator {
	case common.OpAnyShareSelectedFileTrigger:
		files, _, err := m.efast.ListDir(ctx, fmt.Sprintf("%v", docIDs[0]), strings.TrimPrefix(user.TokenID, "Bearer "), user.LoginIP)
		if err != nil {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
		}

		if len(files) == 0 {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{
				"trigger_type": entity.TriggerManually,
				"operator":     common.OpAnyShareSelectedFileTrigger,
				"info":         "doc not found",
			})
		}
		info := files[0].(map[string]interface{})
		id = info["docid"].(string)
	case common.OpAnyShareSelectedFolderTrigger:
		id = fmt.Sprintf("%v", docIDs[0])
	case common.MannualTrigger:
	default:
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, "unsupported trigger type "+triggerStep.Operator)
	}

	if id != "" {
		metadata, err := m.efast.GetDocMsg(ctx, id)
		if err != nil {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
		}

		sourceType := "file"
		if metadata.Size == -1 {
			sourceType = "folder"
		}

		source := map[string]interface{}{
			"type": sourceType,
			"id":   metadata.DocID,
			"name": metadata.Name,
			"rev":  metadata.Rev,
			"size": metadata.Size,
			"path": metadata.Path,
		}

		bytes, _ := json.Marshal(source)

		runVar["source"] = string(bytes)
	}

	if fields, ok := triggerStep.Parameters["fields"].([]interface{}); ok {
		data, ok := triggerData["data"].(map[string]interface{})
		if !ok {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{
				"trigger_type": entity.TriggerManually,
				"operator":     triggerStep.Operator,
				"trigger_data": "data is required",
			})
		}

		err := ParseFields(ctx, primitive.A(fields), data, runVar, ErrTypeV2).BuildError()
		if err != nil {
			log.Warnf("[logic.TriggerDataFlowManuallyHandler] ParseFields err, deail: %s", err.Error())
			return entity.TriggerManually, err
		}
	}

	return entity.TriggerManually, nil
}

func (m *mgnt) SingleDeBug(ctx context.Context, params SingleDeBugReq, userInfo *drivenadapters.UserInfo) (string, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	logger := traceLog.WithContext(ctx)

	// 单个调试节点参数校验
	operator := params.Operator
	if strings.HasPrefix(operator, "@operator") {
		operator = "@operator/"
	}

	if strings.HasPrefix(operator, common.TriggerOperatorPrefix) {
		operator = common.TriggerOperatorPrefix
	}

	path, ok := common.ActionMap[operator]
	if !ok {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, map[string]interface{}{"operator": fmt.Sprintf("unsupported operation type: %s", operator)})
	}

	if path != "" {
		taskByte, _ := json.Marshal(params)
		err = common.JSONSchemaValidV2(ctx, taskByte, path)
		if err != nil {
			return "", err
		}
	}

	runVar := entity.DagInstanceVars{
		"userid":        entity.DagInstanceVar{Value: userInfo.UserID},
		"operator_id":   entity.DagInstanceVar{Value: userInfo.UserID},
		"operator_name": entity.DagInstanceVar{Value: userInfo.UserName},
		"operator_type": entity.DagInstanceVar{Value: userInfo.AccountType},
		"single_debug":  entity.DagInstanceVar{Value: "true"},
	}

	dagIns := &entity.DagInstance{
		Vars:            runVar,
		UserID:          userInfo.UserID,
		Priority:        common.PriorityLowest,
		Status:          entity.DagInstanceStatusInit,
		MemoryShareData: &entity.MemoryShareData{},
		BizDomainID:     params.BizDomainID,
	}

	dagIns.Initial()

	taskIns := &entity.TaskInstance{
		TaskID:         params.ID,
		DagInsID:       dagIns.ID,
		ActionName:     params.Operator,
		TimeoutSecs:    60,
		Params:         params.Parameters,
		Traces:         []entity.TraceInfo{},
		RenderedParams: params.Parameters,
		Patch: func(context.Context, *entity.TaskInstance) error {
			return nil
		},
		Status:             entity.TaskInstanceStatusInit,
		RelatedDagInstance: dagIns,
	}
	taskIns.Initial()

	token := &entity.Token{
		UserID:   userInfo.UserID,
		UserName: userInfo.UserName,
		Token:    userInfo.TokenID,
		LoginIP:  userInfo.LoginIP,
	}

	key := fmt.Sprintf("DEBUG:%v", taskIns.ID)
	ttl := m.taskTimeoutConfig.GetTimeout(taskIns.ActionName)
	ttl = utils.IfNot(ttl > 24*60*60, 24*60*60, ttl)
	m.memoryCache.Set(key, taskIns, time.Duration(ttl)*time.Second)

	sctx, cancle := context.WithTimeout(context.Background(), time.Duration(ttl)*time.Second)
	execute := mod.NewDebugExecute(sctx, dagIns, taskIns, token)
	err = m.pool.Submit(func() {
		dErr := execute.SingleDeBug()
		if dErr != nil {
			dagIns.Status = entity.DagInstanceStatusFailed
			taskIns.Status = entity.TaskInstanceStatusFailed
			taskIns.Results = dErr.Error()
		}

		m.memoryCache.Set(key, taskIns, 5*time.Minute)
		cancle()
	})

	if err != nil {
		m.memoryCache.Delete(key)
		logger.Warnf("[logic.SingleDeBug] SubmitWithTimeout err, detail: %s", err.Error())
		if err.Error() == "too many goroutines blocked on submit or Nonblocking is set" {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorForbidden, aerr.ConcurrencyLimit, map[string]any{"limit": m.config.Server.DebugExecutorCount})
		}
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}

	return taskIns.ID, nil
}

// SingleDeBugResult 获取单步调试结果
func (m *mgnt) SingleDeBugResult(ctx context.Context, id string) (entity.TaskInstanceStatus, any, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	key := fmt.Sprintf("DEBUG:%v", id)

	raw, exist := m.memoryCache.GetRaw(key)
	if !exist {
		return "", nil, ierr.NewPublicRestError(ctx, ierr.PErrorNotFound, ierr.PErrorNotFound, map[string]interface{}{"id": id})
	}

	taskIns, ok := raw.(*entity.TaskInstance)
	if !ok {
		return "", nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, "type assert failed")
	}

	return taskIns.Status, taskIns.Results, nil
}

// FullDebug 整体调试数据流
func (m *mgnt) FullDebug(ctx context.Context, params FullDeBugReq, userInfo *drivenadapters.UserInfo) (string, string, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 判断是否具有执行权限

	var tasks = make([]entity.Task, 0)
	var stepList = make([]map[string]interface{}, 0)
	steps := make([]entity.Step, len(params.Steps))
	copy(steps, params.Steps)
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
		return "", "", err
	}

	var dag *entity.Dag

	// 创建调试任务
	if params.ID == "" {
		// 手动构造dag配置信息
		dag = &entity.Dag{
			UserID:  userInfo.UserID,
			Trigger: entity.TriggerManually,
			Cron:    "",
			Vars: entity.DagVars{
				"userid": {DefaultValue: userInfo.UserID},
				"docid":  {DefaultValue: ""},
			},
			Status:        entity.DagStatusNormal,
			Tasks:         tasks,
			Steps:         params.Steps,
			AppInfo:       entity.AppInfo{},
			TriggerConfig: params.TriggerConfig,
			IsDebug:       true,
		}

		dag.Initial()
		dag.SetCheckRootNode(func(t []entity.Task) error {
			_, bErr := mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
			if bErr != nil {
				return bErr
			}
			return nil
		})
		dag.Name = fmt.Sprintf("%v_DEBUG", dag.ID)
		_, err = m.mongo.CreateDag(ctx, dag)
		if err != nil {
			log.Warnf("[logic.FullDebug] CreateDag err, detail: %s", err.Error())
			return "", "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
		}
	} else {
		dag, err = m.mongo.GetDag(ctx, params.ID)
		if err != nil {
			if aerr.IsNotFoundErr(err) {
				return "", "", ierr.NewPublicRestError(ctx, ierr.PErrorNotFound, ierr.PErrorNotFound, map[string]string{"dagId": params.ID})
			}
			log.Warnf("[logic.FullDebug] GetDag err, detail: %s", err.Error())
			return "", "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
		}

		dag.Steps = params.Steps
		dag.Tasks = tasks
		err = m.mongo.UpdateDag(ctx, dag)
		if err != nil {
			log.Warnf("[logic.FullDebug] UpdateDag err, detail: %s", err.Error())
			return "", "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
		}
	}

	runVar := map[string]string{
		"userid":        userInfo.UserID,
		"operator_id":   userInfo.UserID,
		"operator_name": userInfo.UserName,
		"operator_type": common.User.ToString(),
	}

	if dag.TriggerConfig != nil {
		runVar["source_type"] = m.getDataSourceType(dag.TriggerConfig.DataSource)
		runVar["datasourceid"] = dag.Steps[0].ID
	} else {
		runVar["source_type"] = m.getDataSourceType(dag.Steps[0].DataSource)
		if dag.Steps[0].DataSource != nil {
			runVar["datasourceid"] = dag.Steps[0].DataSource.ID
		}
	}

	trigger, err := HandleTriggerRunVar(ctx, params.Steps[0].Operator, params.Steps[0], userInfo, params.TriggerData, runVar)
	if err != nil {
		return "", "", err
	}

	dagIns, err := dag.Run(ctx, trigger, runVar)
	if err != nil {
		log.Warnf("[logic.FullDebug] RunDag err, detail: %s", err.Error())
		return "", "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	dagIns.Mode = entity.DagInstanceModeVM
	dagIns.Status = entity.DagInstanceStatusRunning
	dagIns.Initial()

	// 数据视图手动构造触发数据源
	if dag.Steps[0].Operator == common.MDLDataViewTrigger {
		data, ok := runVar["mdl_data"]
		if ok {
			var entries []any
			jerr := json.Unmarshal([]byte(data), &entries)
			if err != nil {
				log.Warnf("[logic.FullDebug] json.Unmarshal err, detail: %s", jerr.Error())
				return "", "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
			}

			werr := dagIns.WriteEventByVariableMap(ctx, map[string]any{
				"__" + dag.Steps[0].ID: map[string]any{
					"data": entries,
				},
			}, time.Now().UnixMicro())

			if werr != nil {
				dagIns.Status = entity.DagInstanceStatusFailed
				dagIns.Reason = werr.Error()
			}

			delete(runVar, "mdl_data")
			delete(dagIns.Vars, "mdl_data")
		}
	}

	dagInsID, err := m.mongo.CreateDagIns(ctx, dagIns)
	if err != nil {
		log.Warnf("[logic.FullDebug] CreateDagIns err, detail: %s", err.Error())
		return "", "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	err = m.pool.Submit(func() {
		vmIns := mod.NewVMExt(context.Background(), dagIns, userInfo.UserID)
		_ = vmIns.Boot()
	})
	if err != nil {
		log.Warnf("[logic.FullDebug] Submit err, detail: %s", err.Error())
		if err.Error() == "too many goroutines blocked on submit or Nonblocking is set" {
			return "", "", ierr.NewPublicRestError(ctx, ierr.PErrorForbidden, aerr.ConcurrencyLimit, map[string]any{"limit": m.config.Server.DebugExecutorCount})
		}
		return "", "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}

	return dag.ID, dagInsID, nil
}
