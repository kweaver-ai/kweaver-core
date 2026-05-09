package mgnt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	aerr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/perm"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils/openapi"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils/ptr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v2"
)

const (
	// OperatorSource 算子来源，取值system/unknown, system不支持修改, 默认使用unknown
	OperatorSource = "unknown"
	// OperatorType 算子类型, 工作流注册算子都为组合算子
	OperatorType = "composite"
	// OperatorMetadataType 工作流生成的算子API文档都为OpenAPI格式
	OperatorMetadataType = "openapi"
	// PublishedStatus 已发布状态
	PublishedStatus = "published"
	// PublishedStatus 已发布状态
	UnPublishedStatus = "unpublish"
	// RegisterSuccessStatus 注册算子状态
	RegisterSuccessStatus = "success"
)

// 依赖服务错误码
const (
	OperatorErrorConflict = "AgentOperatorIntegration.Conflict.OperatorExistsSameName"
)

// ComboOperatorReq 算子编排请求体
type ComboOperatorReq struct {
	DagID       string           `json:"-"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Steps       []entity.Step    `json:"steps"`
	Category    string           `json:"category"`
	OutPuts     []*entity.OutPut `json:"outputs"`
	Status      string           `json:"status"`
	BizDomainID string           `json:"-"`
}

// ComboOperatorReq 算子编排请求体
type OptionalComboOperatorReq struct {
	DagID       string            `json:"-"`
	Title       *string           `json:"title"`
	Description *string           `json:"description"`
	Steps       *[]entity.Step    `json:"steps"`
	Category    *string           `json:"category"`
	OutPuts     *[]*entity.OutPut `json:"outputs"`
	OperatorID  string            `json:"operator_id"` // 组合算子ID
	Version     string            `json:"version"`     // 算子版本 uuid
	Status      string            `json:"status"`
	BizDomainID string            `json:"-"`
}

// CycleError 循环引用错误
type CycleError struct {
	Cycle      bool   `json:"-"`
	CurrID     string `json:"id,omitempty"`
	ReferDagID string `json:"refer_dag_id,omitempty"`
	ReferName  string `json:"refer_name,omitempty"`
}

type ComboOperatorItem struct {
	OperatorID   string `json:"operator_id"`
	OperatorName string `json:"operator_name"`
	Version      string `json:"version"`
	Description  string `json:"description"`
	OperatorType string `json:"operator_type"`
	Category     string `json:"category"`
	Status       string `json:"status"`
	DagID        string `json:"dag_id"`
	CreatorName  string `json:"creator_name"`
	CreatedAt    int64  `json:"created_at"`
	UpdaterName  string `json:"updater_name"`
	UpdatedAt    int64  `json:"updated_at"`
}

type ComboOperatorList struct {
	Ops   []*ComboOperatorItem `json:"ops"`
	Page  int64                `json:"page"`
	Limit int64                `json:"limit"`
	Total int64                `json:"total"`
}

func (oc *ComboOperatorReq) ToOptinalComboOperator() *OptionalComboOperatorReq {
	return &OptionalComboOperatorReq{
		Title:       &oc.Title,
		Description: &oc.Description,
		Steps:       &oc.Steps,
		Category:    &oc.Category,
		OutPuts:     &oc.OutPuts,
	}
}

type OperatorInfo = drivenadapters.OperatorResponse

type ValidOperatorsResult struct {
	RefDagIDs []string
	OpInfoMap map[string]*OperatorInfo
}

func (oc *OptionalComboOperatorReq) BuildAPI() string {
	if oc.Steps == nil {
		return ""
	}
	config := common.NewConfig()

	var outputs = make([]interface{}, 0)
	outputsBytes, _ := json.Marshal(*oc.OutPuts)
	_ = json.Unmarshal(outputsBytes, &outputs)

	docBuilder := openapi.NewOpenAPIDocumentBuilder()
	operation := openapi.NewOperationBuilder().
		WithOperatorMeta(*oc.Title, *oc.Description).
		WithRequestParam("Authorization", "User Token", "string", "header", true).
		WithRequestBody("application/json", (*oc.Steps)[0].Parameters, openapi.ConvertFlowParamsToSchema).
		WithResponse("200", "OK").
		WithResponseContent(200, "application/json", map[string]interface{}{"fields": outputs}, openapi.ConvertFlowParamsToSchema).Build()
	// WithResponseContent(200, "application/json", map[string]interface{}{}, openapi.ConvertMapToSchema).Build()

	docBuilder.AddOperation(fmt.Sprintf("/operators/%v/executions", oc.DagID), "POST", operation)

	doc := openapi.NewOpenAPI().
		AddOpenAPIInfo(*oc.Title, *oc.Description).
		AddServers(fmt.Sprintf("http://%v:%v/api/automation/v1", config.ContentAutomation.PublicHost, config.ContentAutomation.PublicPort)).
		AddPaths(docBuilder.Build()).
		Build()
	yamlData, _ := yaml.Marshal(doc)
	return string(yamlData)
}

// OperatorImportExportItem 算子导入导出项
type OperatorImportExportItem struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Steps       []entity.Step    `json:"steps"`
	Category    string           `json:"category"`
	OutPuts     []*entity.OutPut `json:"outputs"`
	OperatorID  string           `json:"operator_id"`
	IsRoot      bool             `json:"is_root,omitempty"`
}

// ExportOperator 算子导出结果
type ExportOperator struct {
	Configs     []*OperatorImportExportItem `json:"configs"`
	OperatorIDs []string                    `json:"operator_ids"`
}

type ImportOperatorReq struct {
	Mode    string                     `json:"mode"`
	Configs []OperatorImportExportItem `json:"configs"`
}

func (m *mgnt) CreateComboOperator(ctx context.Context, param *ComboOperatorReq, userInfo *drivenadapters.UserInfo) (string, string, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	userDetail, err := m.usermgnt.GetUserInfoByType(userInfo.UserID, userInfo.AccountType)
	if err != nil {
		log.Warnf("[logic.CreateComboOperator] GetUserInfoByType err, detail: %s", err.Error())
		return "", "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, err.Error())
	}
	userInfo.UserName = userDetail.UserName

	dag := &entity.Dag{
		UserID: userInfo.UserID,
		Vars: entity.DagVars{
			"userid": {DefaultValue: userInfo.UserID},
			"docid":  {DefaultValue: ""},
		},
		Priority:    common.PriorityLowest,
		Type:        common.DagTypeComboOperator,
		Status:      entity.DagStatusNormal,
		BizDomainID: param.BizDomainID,
	}

	oparam := param.ToOptinalComboOperator()
	err = m.CheckAndBuildDag(ctx, oparam, dag, userInfo)
	if err != nil {
		return "", "", err
	}

	dag.SetCheckRootNode(func(t []entity.Task) error {
		_, bErr := mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
		if bErr != nil {
			return bErr
		}
		return nil
	})

	dagID, err := m.mongo.CreateDag(ctx, dag)
	if err != nil {
		log.Warnf("[logic.CreateComboOperator] CreateDag err, detail: %s", err.Error())
		return "", "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	oparam.DagID = dagID
	// 生成OpenAPI文件
	doc := oparam.BuildAPI()
	// 调用算子注册接口
	isPublish := param.Status == PublishedStatus
	data := &drivenadapters.RegisterOperatorReq{
		Data:                 doc,
		OperatorMetadataType: OperatorMetadataType,
		OperatorInfo: &drivenadapters.OperatorInfo{
			OperatorType:  OperatorType,
			ExecutionMode: dag.ExecMode,
			Category:      param.Category,
			Source:        OperatorSource,
		},
		ExtendInfo: map[string]interface{}{
			"dag_id": oparam.DagID,
		},
		DirectPublish: isPublish,
		UserToken:     strings.TrimPrefix(userInfo.TokenID, "Bearer "),
		BizDomainID:   param.BizDomainID,
	}

	results, err := m.operator.RegisterOperator(ctx, data, userInfo)
	err = m.addOperatorAfter(ctx, results, err)
	if err != nil {
		_ = m.mongo.DeleteDag(ctx, dagID)
		log.Warnf("[logic.CreateComboOperator] RegisterOperator err, detail: %s", err.Error())
		return "", "", err
	}
	operatorID := results[0].OperatorID

	dag.OperatorID = operatorID
	err = m.mongo.UpdateDag(ctx, dag)
	if err != nil {
		log.Warnf("[logic.CreateComboOperator] UpdateDag err, detail: %s", err.Error())
	}

	return dagID, operatorID, nil
}

func (m *mgnt) UpdateComboOperator(ctx context.Context, param *OptionalComboOperatorReq, userInfo *drivenadapters.UserInfo) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeComboOperator: {perm.ModeifyOperation},
			common.DagTypeDefault:       {perm.OldShareOperation},
		},
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, param.DagID, userInfo, opMap)
	if err != nil {
		return err
	}

	op, err := m.validOperators(ctx, []string{fmt.Sprintf("%v:%v", param.OperatorID, param.Version)}, userInfo, param.BizDomainID)
	if err != nil {
		return err
	}

	query := map[string]interface{}{"_id": param.DagID}

	// check dag whether exisis
	dag, err := m.mongo.GetDagByFields(ctx, query)
	if err != nil {
		if aerr.IsNotFoundErr(err) {
			return ierr.NewPublicRestError(ctx, ierr.PErrorNotFound, aerr.DescKeyTaskNotFound, map[string]string{"dagId": param.DagID})
		}
		log.Warnf("[logic.UpdateComboOperator] GetDagByFields err, query: %v, deail: %s", query, err.Error())
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	// 生成steps md5 指纹信息，对比是否有节点更新
	var preStepHash, curStepHash string
	if param.Steps != nil {
		stepByte, _ := json.Marshal(dag.Steps)
		preStepHash = utils.ComputeHash(string(stepByte))
		stepByte, _ = json.Marshal(*param.Steps)
		curStepHash = utils.ComputeHash(string(stepByte))
	}
	var preOutPutsHash, curOutPutsHash string
	if param.OutPuts != nil {
		opByte, _ := json.Marshal(dag.OutPuts)
		preOutPutsHash = utils.ComputeHash(string(opByte))
		opByte, _ = json.Marshal(*param.OutPuts)
		curOutPutsHash = utils.ComputeHash(string(opByte))
	} else {
		// 回填原始数据，目的为了openapi返回值结构保持不变
		param.OutPuts = &dag.OutPuts
	}

	preDag := utils.DeepCopy(dag)

	err = m.CheckAndBuildDag(ctx, param, dag, userInfo)
	if err != nil {
		return err
	}

	// 获取当前算子发布状态
	var opStatus string
	for _, v := range op.OpInfoMap {
		if v.OperatorID == param.OperatorID && v.Version == param.Version {
			opStatus = v.Status
		}
	}

	dag.SetCheckRootNode(func(t []entity.Task) error {
		_, bErr := mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
		if bErr != nil {
			return bErr
		}
		return nil
	})

	updateFlag := true
	// case1：如果存在更新，则需要生成新版本
	// case2：如果不存在，则仅更新当前dag配置信息即可
	// 已发布的算子流程有更新或从已发布到未发布都应该生成一条新的记录
	if opStatus != UnPublishedStatus && (preStepHash != curStepHash || preOutPutsHash != curOutPutsHash || param.Status == UnPublishedStatus) {
		dag.ID = ""
		dagID, err := m.mongo.CreateDag(ctx, dag)
		if err != nil {
			log.Warnf("[logic.UpdateComboOperator] CreateDag err, detail: %s", err.Error())
			return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
		}
		param.DagID = dagID
		updateFlag = false
	} else {
		err = m.mongo.UpdateDag(ctx, dag)
		if err != nil {
			log.Warnf("[logic.UpdateComboOperator] UpdateDag err, detail: %s", err.Error())
			return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
		}
	}

	// 生成OpenAPI文件
	doc := param.BuildAPI()
	// 调用算子注册接口
	isPublish := param.Status == PublishedStatus
	data := &drivenadapters.UpdateOperatorReq{
		OperatorID:           param.OperatorID,
		Data:                 doc,
		OperatorMetadataType: OperatorMetadataType,
		OperatorInfo: &drivenadapters.OperatorInfo{
			OperatorType:  OperatorType,
			ExecutionMode: dag.ExecMode,
			Category:      dag.Category,
			Source:        OperatorSource,
		},
		ExtendInfo: map[string]interface{}{
			"dag_id": param.DagID,
		},
		DirectPublish: isPublish,
		UserToken:     strings.TrimPrefix(userInfo.TokenID, "Bearer "),
		BizDomainID:   param.BizDomainID,
	}

	results, err := m.operator.UpdateOperator(ctx, data, userInfo)
	err = m.addOperatorAfter(ctx, results, err)
	if err != nil {
		if updateFlag {
			_ = m.mongo.UpdateDag(ctx, preDag)
		} else {
			_ = m.mongo.DeleteDag(ctx, param.DagID)
		}
		log.Warnf("[logic.UpdateComboOperator] UpdateOperator err, detail: %s", err.Error())
		return err
	}

	return nil
}

// ListComboOperator 列举组合算子
func (m *mgnt) ListComboOperator(ctx context.Context, params map[string]interface{}, userInfo *drivenadapters.UserInfo) (*ComboOperatorList, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	page := params["page"].(int64)
	limit := params["limit"].(int64)

	query := &drivenadapters.QueryParams{}
	paramsBytes, _ := json.Marshal(params)
	_ = json.Unmarshal(paramsBytes, query)
	query.Page = ptr.Int64(page + 1)
	query.PageSize = &limit

	if v, ok := params["sortby"].(string); ok {
		query.SortBy = &v
	}
	if v, ok := params["order"].(string); ok {
		query.SortOrder = &v
	}

	opratores, err := m.operator.OperatorList(ctx, query, userInfo)
	if err != nil {
		log.Warnf("[logic.ListComboOperator] OperatorList err, detail: %s", err.Error())
		return nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, err.Error())
	}

	ops := &ComboOperatorList{
		Ops:   []*ComboOperatorItem{},
		Page:  page,
		Limit: limit,
		Total: opratores.Total,
	}

	for _, v := range opratores.Data {
		dagID, ok := v.ExtendInfo["dag_id"]
		if !ok {
			dagID = ""
		}

		ops.Ops = append(ops.Ops, &ComboOperatorItem{
			OperatorID:   v.OperatorID,
			OperatorName: v.Name,
			Version:      v.Version,
			Description:  v.Metadata.Description,
			OperatorType: v.OperatorInfo.OperatorType,
			Category:     v.OperatorInfo.Category,
			Status:       v.Status,
			DagID:        dagID.(string),
			CreatorName:  v.CreateUser,
			CreatedAt:    v.CreateTime,
			UpdaterName:  v.UpdateUser,
			UpdatedAt:    v.UpdateTime,
		})
	}

	return ops, nil
}

func (m *mgnt) CheckAndBuildDag(ctx context.Context, param *OptionalComboOperatorReq, dag *entity.Dag, userInfo *drivenadapters.UserInfo) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if param.Title != nil {
		dag.Name = *param.Title
	}

	if param.Description != nil {
		dag.Description = *param.Description
	}

	if param.Steps != nil {
		var tasks = make([]entity.Task, 0)
		var stepList = make([]map[string]interface{}, 0)
		steps := make([]entity.Step, len(*param.Steps))
		copy(steps, *param.Steps)
		m.buildTasks(&steps[0], steps, &tasks, nil, &stepList, nil, nil)

		// step参数校验
		err = m.validSteps(&Validate{
			Ctx:         ctx,
			Steps:       stepList,
			IsAdminRole: false,
			UserInfo:    userInfo,
			ErrType:     ErrTypeV2,
			ParseFunc:   common.JSONSchemaValidV2,
		}).BuildError()
		if err != nil {
			return err
		}

		var vRes *ValidOperatorsResult
		vRes, err = m.validOperatorInSteps(ctx, stepList, userInfo, dag.BizDomainID)
		if err != nil {
			return err
		}

		dag.SubIDs = vRes.RefDagIDs
		// 开始节点必须为表单触发
		if steps[0].Operator != common.FormTrigger {
			return ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, aerr.DescKeyUnSupportedTrigger, map[string]interface{}{"trigger": steps[0].Operator})
		}

		// 流程引用子流程是否存在环
		cycle, err := m.hasCycle(ctx, param.DagID, dag.SubIDs, []string{common.DagTypeComboOperator})
		if err != nil {
			return err
		}

		if cycle.Cycle {
			return ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorLoopDetected, cycle)
		}

		// 获取算子运行方式
		dag.ExecMode, err = m.getOperatorExecutionMode(vRes.OpInfoMap, stepList)
		if err != nil {
			return err
		}

		dag.Trigger = m.getTriggerType((*param.Steps)[0].Operator)
		dag.Tasks = tasks
		dag.Steps = *param.Steps
	}

	if param.Category != nil {
		dag.Category = *param.Category
	}

	if param.OutPuts != nil {
		dag.OutPuts = *param.OutPuts
	}

	return nil
}

// hasCycle 当前创建和更新流程是否存在环
func (m *mgnt) hasCycle(ctx context.Context, dagID string, dagIDs []string, dagTypes []string) (*CycleError, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var visited = map[string]struct{}{}
	var graphMap = map[string][]string{}
	var queue []string
	queue = append(queue, dagIDs...)

	if dagID != "" {
		graphMap[dagID] = append(graphMap[dagID], queue...)
		visited[dagID] = struct{}{}
	}

	cycleErr := &CycleError{
		CurrID: dagID,
	}
	for len(queue) > 0 {
		size := len(queue)
		nodes := queue[0:size]
		queue = queue[size:]
		fillter := bson.M{
			"_id": bson.M{"$in": nodes},
		}
		dags, err := m.mongo.ListDagByFields(ctx, fillter, options.FindOptions{})
		if err != nil {
			log.Warnf("[logic.hasCycle] ListDagByFields err, detail: %s", err.Error())
			return cycleErr, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
		}
		for _, dag := range dags {
			dagType := utils.IfNot(dag.Type == "", common.DagTypeDefault, dag.Type)
			if !utils.Contains(dagTypes, dagType) || len(dag.SubIDs) == 0 && dagID != dag.ID {
				continue
			}

			// 标记当前节点已被访问
			if _, ok := visited[dag.ID]; ok {
				cycleErr.Cycle = true
				cycleErr.ReferDagID = dag.ID
				cycleErr.ReferName = dag.Name
				return cycleErr, nil
			}
			visited[dag.ID] = struct{}{}
			graphMap[dag.ID] = dag.SubIDs
			queue = append(queue, dag.SubIDs...)
		}
	}

	return cycleErr, nil
}

// validOperatorInSteps 校验组合算子中引用的算子信息
func (m *mgnt) validOperatorInSteps(ctx context.Context, stepList []map[string]interface{}, userInfo *drivenadapters.UserInfo, bizDomainID string) (*ValidOperatorsResult, error) {
	var opIDs []string
	var hasReturn bool
	for _, v := range stepList {
		opType := v["operator"].(string)
		if !hasReturn && opType == common.InternalReturnOpt {
			hasReturn = true
		}

		var opID string

		if strings.HasPrefix(opType, common.ComboOperatorPrefix) {
			opID = strings.TrimPrefix(opType, common.ComboOperatorPrefix)

		} else if strings.HasPrefix(opType, common.TriggerOperatorPrefix) {
			opID = strings.TrimPrefix(opType, common.TriggerOperatorPrefix)
		}

		if opID == "" {
			continue
		}
		paramsIface := v["parameters"].(map[string]interface{})
		version := fmt.Sprintf("%v", paramsIface["version"])

		opIDs = append(opIDs, fmt.Sprintf("%v:%v", opID, version))
	}

	// 校验流程中是否包含retrun节点，至少包含一个
	if !hasReturn {
		return nil, ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, aerr.RequiredRetrunNode, nil)
	}

	return m.validOperators(ctx, opIDs, userInfo, bizDomainID)
}

// validOperators 校验算子, 包括算子状态或算子存在性, 暂未提供批量获取接口只能挨个请求，是否存在性能问题待考虑
func (m *mgnt) validOperators(ctx context.Context, ops []string, userInfo *drivenadapters.UserInfo, bizDomainID string) (*ValidOperatorsResult, error) {
	result := &ValidOperatorsResult{
		RefDagIDs: []string{},
		OpInfoMap: map[string]*OperatorInfo{},
	}
	var invalidOpMap = make([]map[string]string, 0)
	var visited = map[string]struct{}{}

	for _, v := range ops {
		if _, ok := visited[v]; ok {
			continue
		}

		arr := strings.Split(v, ":")
		opID, version := arr[0], arr[1]

		opInfo, err := m.operator.LatestOperatorInfo(ctx, opID, bizDomainID, userInfo)
		if err != nil {
			parseErr, eErr := ierr.ExHTTPErrorParser(err)
			if eErr == nil && parseErr.Status == http.StatusNotFound {
				return nil, ierr.NewPublicRestError(ctx, ierr.PErrorNotFound, ierr.PErrorNotFound, map[string]interface{}{"operator_id": opID})
			}
			return nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, err.Error())
		}

		// if opInfo.Status != PublishedStatus {
		// 	invalidOpMap = append(invalidOpMap, map[string]string{"operator_id": opInfo.OperatorID, "version": opInfo.Version})
		// }

		key := fmt.Sprintf("%v:%v", opID, version)
		result.OpInfoMap[key] = opInfo

		visited[v] = struct{}{}
		if v, ok := opInfo.ExtendInfo["dag_id"]; ok && v != "" {
			result.RefDagIDs = append(result.RefDagIDs, fmt.Sprintf("%v", v))
		}
	}

	if len(invalidOpMap) > 0 {
		return nil, ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, aerr.UnPublishedOperator, map[string]interface{}{"ops": invalidOpMap})
	}

	return result, nil
}

// addOperatorAfter 注册或更新算子后操作，校验服务返回错误信息
func (m *mgnt) addOperatorAfter(ctx context.Context, results []*drivenadapters.OperatorModifyResp, err error) error {
	// 如果是接口错误，则需要判断是否是非法参数，非400错误则返回服务依赖错误
	if err != nil {
		parseErr, eErr := ierr.ExHTTPErrorParser(err)
		if eErr == nil && parseErr.Status == http.StatusBadRequest {
			return ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, parseErr.Body["details"])
		}
		if err != nil && parseErr.Status == http.StatusForbidden {
			return ierr.NewPublicRestError(ctx, ierr.PErrorForbidden, aerr.NoPermission, parseErr.Body["details"])
		}
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, err.Error())
	}

	if len(results) == 0 {
		return nil
	}

	result := results[0]
	if result.Status == RegisterSuccessStatus {
		return nil
	}

	if result.Error["code"] == OperatorErrorConflict {
		return ierr.NewPublicRestError(ctx, ierr.PErrorConflict, ierr.PErrorConflict, result.Error["details"])
	} else {
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, result.Error)
	}
}

// getOperatorExecutionMode  获取算子运行模式
func (m *mgnt) getOperatorExecutionMode(opInfoMap map[string]*OperatorInfo, stepList []map[string]interface{}) (string, error) {
	for _, v := range stepList {
		opType := v["operator"].(string)

		if opType == common.InternalToolPy3Opt {
			p := v["parameters"].(map[string]any)
			if mode, ok := p["mode"]; ok && mode == "sync" {
				continue
			}
		}

		// 当前内置节点异步节点： python节点、审核节点、自定义节点、沙箱节点
		if opType == common.InternalToolPy3Opt ||
			opType == common.WorkflowApproval ||
			opType == common.OpSandboxExecute ||
			strings.HasPrefix(opType, common.CustomOperatorPrefix) {
			return common.ExecutionModeAsync, nil
		}
	}

	for _, v := range opInfoMap {
		if v.OperatorInfo.ExecutionMode == common.ExecutionModeAsync {
			return common.ExecutionModeAsync, nil
		}
	}

	return common.ExecutionModeSync, nil
}

// ExportOperator 导出算子
func (m *mgnt) ExportOperator(ctx context.Context, dagIDs []string) (ExportOperator, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	exportRes := ExportOperator{}
	dagIDs = utils.RemoveRepByMap(dagIDs)
	if len(dagIDs) == 0 || len(dagIDs) == 1 && dagIDs[0] == "" {
		return exportRes, nil
	}

	levelNodes := [][]string{}
	nodeMap := map[string]*entity.Dag{}

	dags, err := m.mongo.ListDag(ctx, &mod.ListDagInput{
		DagIDs: dagIDs,
		Type:   common.DagTypeComboOperator,
	})
	if err != nil {
		log.Warnf("[logic.ExportOperator] ListDag err, detail: %s", err.Error())
		return exportRes, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	if len(dagIDs) != len(dags) {
		existIDs := []string{}
		for _, dag := range dags {
			existIDs = append(existIDs, dag.ID)
		}
		_, del := utils.Arrcmp(dagIDs, existIDs)
		return exportRes, ierr.NewPublicRestError(ctx, ierr.PErrorNotFound, ierr.PErrorNotFound, map[string]interface{}{"ids": del})
	}

	subIDs := []string{}
	for _, dag := range dags {
		nodeMap[dag.ID] = dag
		exportRes.OperatorIDs = append(exportRes.OperatorIDs, dag.OperatorID)
		subIDs = append(subIDs, dag.SubIDs...)
	}

	levelNodes = append(levelNodes, dagIDs)
	exportRes.OperatorIDs = append(exportRes.OperatorIDs, findOperatorID(dags...)...)

	for len(subIDs) > 0 {
		subDags, err := m.mongo.ListDag(ctx, &mod.ListDagInput{
			DagIDs: subIDs,
			Type:   common.DagTypeComboOperator,
		})
		if err != nil {
			log.Warnf("[logic.ExportOperator] GetDags err, detail: %s", err.Error())
			return exportRes, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
		}

		if len(subDags) == 0 {
			break
		}

		levelNodes = append(levelNodes, subIDs)
		subIDs = []string{}
		for _, subDag := range subDags {
			nodeMap[subDag.ID] = subDag
			subIDs = append(subIDs, subDag.SubIDs...)

		}
		exportRes.OperatorIDs = append(exportRes.OperatorIDs, findOperatorID(subDags...)...)
	}

	seen := map[string]struct{}{}
	for i := len(levelNodes) - 1; i >= 0; i-- {
		for _, id := range levelNodes[i] {
			if _, ok := seen[id]; ok {
				continue
			}
			dag := nodeMap[id]
			exportRes.Configs = append(exportRes.Configs, &OperatorImportExportItem{
				ID:          dag.ID,
				Title:       dag.Name,
				Description: dag.Description,
				Steps:       dag.Steps,
				Category:    dag.Category,
				OutPuts:     dag.OutPuts,
				OperatorID:  dag.OperatorID,
				IsRoot:      utils.IsContain(dag.ID, dagIDs),
			})
			seen[id] = struct{}{}
		}
	}

	exportRes.OperatorIDs = utils.RemoveRepByMap(exportRes.OperatorIDs)
	return exportRes, nil
}

func findOperatorID(dags ...*entity.Dag) []string {
	b, _ := json.Marshal(dags)
	re := regexp.MustCompile(`@operator/([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})`)
	matches := re.FindAllStringSubmatch(string(b), -1)
	seen := map[string]struct{}{}
	var ids []string
	for _, g := range matches {
		if len(g) >= 2 {
			id := g[1]
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// DeleteComboOperator 删除组合算子
func (m *mgnt) DeleteComboOperator(ctx context.Context, operatorID string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dags, err := m.mongo.ListDagByFields(ctx, bson.M{"operator_id": operatorID}, options.FindOptions{})
	if err != nil {
		log.Warnf("[logic.DeleteOperator] ListDagByFields err, detail: %s", err.Error())
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	if len(dags) == 0 {
		return nil
	}

	dagIDs := []string{}
	for _, dag := range dags {
		dagIDs = append(dagIDs, dag.ID)
	}

	err = m.mongo.DeleteDag(ctx, dagIDs...)
	if err != nil {
		log.Warnf("[logic.DeleteOperator] DeleteDag err, detail: %s", err.Error())
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	return nil
}

// ImportOperator 导入算子
func (m *mgnt) ImportOperator(ctx context.Context, params *ImportOperatorReq, userInfo *drivenadapters.UserInfo) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dagIDs := []string{}
	for _, item := range params.Configs {
		dagIDs = append(dagIDs, item.ID)
	}

	dags, err := m.mongo.ListDag(ctx, &mod.ListDagInput{
		DagIDs: dagIDs,
		Type:   "all",
	})
	if err != nil {
		log.Warnf("[logic.ExportOperator] GetDags err, detail: %s", err.Error())
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	conflectIDs, coverIDs := []string{}, []string{}
	for _, dag := range dags {
		if params.Mode == "upsert" {
			coverIDs = append(coverIDs, dag.ID)
			continue
		}
		conflectIDs = append(conflectIDs, dag.ID)
	}

	if len(conflectIDs) > 0 {
		return ierr.NewPublicRestError(ctx, ierr.PErrorConflict, ierr.PErrorConflict, map[string]interface{}{"dag_ids": conflectIDs})
	}

	// 构建dag结构
	dags = []*entity.Dag{}
	for _, item := range params.Configs {
		dag := &entity.Dag{
			BaseInfo: entity.BaseInfo{
				ID: item.ID,
			},
			UserID: userInfo.UserID,
			Vars: entity.DagVars{
				"userid": {DefaultValue: userInfo.UserID},
				"docid":  {DefaultValue: ""},
			},
			Priority:   common.PriorityLowest,
			Type:       common.DagTypeComboOperator,
			Status:     entity.DagStatusNormal,
			OperatorID: item.OperatorID,
		}

		oparam := &OptionalComboOperatorReq{
			DagID:       item.ID,
			Title:       &item.Title,
			Description: &item.Description,
			Steps:       &item.Steps,
			Category:    &item.Category,
			OutPuts:     &item.OutPuts,
			OperatorID:  item.OperatorID,
		}
		err = m.CheckAndBuildDag(ctx, oparam, dag, userInfo)
		if err != nil {
			return err
		}

		dags = append(dags, dag)
	}

	err = m.mongo.WithTransaction(ctx, func(nctx context.Context, ms mod.Store) error {
		err = ms.DeleteDag(nctx, coverIDs...)
		if err != nil {
			log.Warnf("[logic.ExportOperator] DeleteDag err, detail: %s", err.Error())
			return err
		}

		_, err := ms.BatchCreateDag(nctx, dags)
		if err != nil {
			log.Warnf("[logic.ExportOperator] BatchCreateDag err, detail: %s", err.Error())
			return err
		}

		return nil
	})
	if err != nil {
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	return nil
}
