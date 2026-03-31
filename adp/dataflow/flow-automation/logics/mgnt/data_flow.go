package mgnt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/perm"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/store"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateDataFlowReq 创建数据流程请求
type CreateDataFlowReq struct {
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	Status        string                `json:"status"`
	TriggerConfig *entity.TriggerConfig `json:"trigger_config"`
	Steps         []entity.Step         `json:"steps"`
	AppInfo       entity.AppInfo        `json:"appinfo"`
	Version       entity.Version        `json:"version"`
	ChangeLog     string                `json:"change_log"`
	DeBugID       string                `json:"debug_id"`
	BizDomainID   string                `json:"-"`
}

// CreateDataFlow 创建数据流类型的流程
func (m *mgnt) CreateDataFlow(ctx context.Context, param *CreateDataFlowReq, userInfo *drivenadapters.UserInfo) (string, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	_, err = m.permPolicy.CheckPerm(ctx, userInfo.UserID, userInfo.AccountType, []string{perm.DataAdminResourceID}, perm.CreateOperation)
	if err != nil {
		return "", err
	}

	// 基本参数验证
	if err = m.validateDataFlowParams(param); err != nil {
		return "", err
	}

	// 验证用户权限
	userDetail, err := m.usermgnt.GetUserInfoByType(userInfo.UserID, userInfo.AccountType)
	if err != nil {
		log.Warnf("[logic.CreateDataFlow] GetUserInfoByType err, Type: %s, detail: %s", userInfo.AccountType, err.Error())
		return "", ierrors.NewIError(ierrors.ErrorDepencyService, "", err.Error())
	}
	userInfo.UserName = userDetail.UserName

	// check duplicated name
	dagInfo, err := m.mongo.GetDagByFields(ctx, map[string]interface{}{"name": param.Title})
	if err != nil && !ierrors.IsNotFoundErr(err) {
		log.Warnf("[logic.CreateDag] GetDagByFields err, deail: %s", err.Error())
		return "", ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	if dagInfo != nil {
		return "", ierrors.NewIError(ierrors.DuplicatedName, "", map[string]string{"title": param.Title})
	}

	if param.DeBugID != "" {
		_, err = m.mongo.GetDag(ctx, param.DeBugID)
		if err != nil {
			if ierrors.IsNotFoundErr(err) {
				return "", ierrors.NewIError(ierrors.InvalidParameter, "", "debug id not found")
			}
			log.Warnf("[logic.CreateDag] GetDag err, deail: %s", err.Error())
			return "", ierrors.NewIError(ierrors.InternalError, "", nil)
		}
	}

	var tasks = make([]entity.Task, 0)
	var stepList = make([]map[string]interface{}, 0)
	steps := make([]entity.Step, len(param.Steps))
	copy(steps, param.Steps)
	m.buildTasks(&steps[0], steps, &tasks, nil, &stepList, nil, nil)

	subDagIDs, err := m.validSubFlowInSteps(ctx, "", stepList)
	if err != nil {
		return "", err
	}

	triggerType := m.getTriggerType(param.TriggerConfig.Operator)

	// 创建 dag
	dag := &entity.Dag{
		UserID:      userInfo.UserID,
		Name:        param.Title,
		Description: param.Description,
		Status:      entity.DagStatusNormal,
		Steps:       param.Steps,
		Tasks:       tasks,
		Priority:    common.PriorityHigh,
		Type:        common.DagTypeDataFlow,
		Trigger:     triggerType,
		Vars: entity.DagVars{
			"userid": {DefaultValue: userInfo.UserID},
			"docid":  {DefaultValue: ""},
		},
		TriggerConfig: param.TriggerConfig,
		ModifyBy:      userInfo.UserID,
		DeBugID:       param.DeBugID,
		BizDomainID:   param.BizDomainID,
		SubIDs:        subDagIDs,
	}

	// 检查S3数据源，如果有临时文件，则移动到正式目录
	if dag.TriggerConfig != nil && dag.TriggerConfig.DataSource != nil &&
		dag.TriggerConfig.DataSource.Operator == common.S3DataListObjects {
		// 转换参数
		dataSource := dag.TriggerConfig.DataSource
		// 直接使用类型化的参数
		params := dataSource.Parameters

		if params != nil && params.Mode == "upload" && len(params.Sources) > 0 {
			var keysToMove []string
			var sourceMap = make(map[string]int) // key -> index

			for i, source := range params.Sources {
				// 优先使用Path，如果Path为空则使用Key
				path := source.Path
				if path == "" {
					path = source.Key
				}

				if strings.Contains(path, "/temp/") {
					keysToMove = append(keysToMove, path)
					sourceMap[path] = i
				}
			}

			if len(keysToMove) > 0 {
				// 我们需要在使用dag.ID之前确保它存在。
				if dag.ID == "" {
					dag.ID = store.NextStringID()
				}

				newKeys, err := m.MoveS3Files(ctx, keysToMove, dag.ID)
				if err != nil {
					log.Warnf("[logic.CreateDataFlow] MoveS3Files err: %v", err)
					return "", ierrors.NewIError(ierrors.InternalError, "failed to move s3 files", err.Error())
				}

				// 更新Sources中的Path/Key
				for i, oldKey := range keysToMove {
					if i < len(newKeys) {
						idx := sourceMap[oldKey]
						// 能够根据oldKey找到对应的source索引
						// 更新为新的路径
						// 此时我们统一更新Path字段，清空Key字段（或者两者都更?）
						// 为了规范，建议使用Path
						params.Sources[idx].Path = newKeys[i]
						params.Sources[idx].Key = "" // 可选：清空Key，避免混淆
					}
				}
			}
		}
	}

	// Double check dag ID generation in case we didn't enter the block above
	if dag.ID == "" {
		dag.ID = store.NextStringID()
	}

	dag.AppInfo = entity.AppInfo{
		Enable: false,
	}

	dag.SetCheckRootNode(func(t []entity.Task) error {
		_, bErr := mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
		if bErr != nil {
			return bErr
		}
		return nil
	})

	err = m.validSteps(&Validate{
		Ctx:         ctx,
		Steps:       stepList,
		IsAdminRole: true,
		UserInfo:    userInfo,
		ErrType:     ErrTypeV1,
		ParseFunc:   common.JSONSchemaValid,
	}).BuildError()
	if err != nil {
		return "", err
	}

	if param.Status == common.StoppedStatus {
		dag.Status = entity.DagStatusStopped
	}

	if dag.Steps[0].Operator == common.MDLDataViewTrigger {
		var params DataViewParam
		params.Init(dag.Steps[0].Parameters)

		if params.SyncMode == SyncModeIncremental {
			dag.IncValues = map[string]any{
				params.IncrementField: params.IncrementValue,
			}
		}
	}

	var dagID string
	dagVersions, err := m.buildDagVersions(&BuildDagVersionParams{
		Dag:        dag,
		CurVersion: &param.Version,
		ChangeLog:  param.ChangeLog,
		UserID:     userInfo.UserID,
		IsCreate:   true,
	})
	if err != nil {
		return "", err
	}

	err = m.mongo.WithTransaction(ctx, func(nctx context.Context, ms mod.Store) error {
		dagID, err = ms.CreateDag(nctx, dag)
		if err != nil {
			return err
		}

		config, _ := json.Marshal(dag)
		dagVersion := dagVersions[0]
		dagVersion.DagID = dagID
		dagVersion.Config = entity.Config(config)
		_, err = ms.CreateDagVersion(nctx, dagVersion)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Warnf("[logic.CreateDataFlow] Transaction err, detail: %s", err.Error())
		return "", ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 分配权限策略
	resourceID := fmt.Sprintf("%s:%s", dagID, common.DagTypeDataFlow)
	allow := []string{perm.ListOperation, perm.CreateOperation, perm.ModeifyOperation, perm.DeleteOperation,
		perm.ViewOperation, perm.ManualExecOperation, perm.RunStatisticsOperation}
	err = m.permPolicy.CreatePolicy(context.Background(), userInfo.UserID, userInfo.AccountType, userInfo.UserName, resourceID, dag.Name, allow, []string{})
	if err != nil {
		if berr := m.mongo.DeleteDag(ctx, dagID); berr != nil {
			log.Warnf("[logic.CreateDataFlow] CreatePolicy Failed, BatchDeleteDagWithTransaction err, detail: %s", berr.Error())
		}
		return "", err
	}

	bizDomainParams := drivenadapters.BizDomainResourceParams{
		BizDomainID:  param.BizDomainID,
		ResourceID:   fmt.Sprintf("%s:%s", dagID, dag.Type),
		ResourceType: perm.DataFlowResourceType,
	}
	err = m.bizDomain.BindResourceInternal(ctx, bizDomainParams)
	if err != nil {
		log.Warnf("[logic.CreateDataFlow] BindResourceInternal err, deail: %s", err.Error())
		m.mongo.DeleteDag(ctx, dagID)
		return "", ierrors.NewIError(ierrors.InternalError, ierrors.ErrorDepencyService, err.Error())
	}

	if param.TriggerConfig.Cron != "" && !strings.HasPrefix(param.TriggerConfig.Cron, common.CronTrigger) {
		jobID, exist, err := m.ecron.RegisterCronJob(
			ctx,
			fmt.Sprintf("auto_%s", dagID),
			fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/cron/%s",
				m.config.ContentAutomation.PrivateHost,
				m.config.ContentAutomation.PrivatePort,
				dagID),
			param.TriggerConfig.Cron,
		)
		if err != nil {
			if !exist {
				log.Warnf("[logic.CreateDataFlow] RegisterCronJob err, detail: %s", err.Error())
				// 删除已创建的 dag
				// if berr := m.mongo.BatchDeleteDagWithTransaction(ctx, []string{dagID}); berr != nil {
				// 	log.Warnf("[logic.CreateDataFlow] BatchDeleteDagWithTransaction err, detail: %s", berr.Error())
				// }
				if berr := m.mongo.DeleteDag(ctx, dagID); berr != nil {
					log.Warnf("[logic.CreateDataFlow] DeleteDag err, detail: %s", berr.Error())
				}

				// 创建失败清除已绑定的权限策略配置
				go m.permPolicy.DeletePolicy(ctx, resourceID)
				_ = m.bizDomain.UnBindResourceInternal(ctx, bizDomainParams)

				if err := m.extData.Remove(ctx, &rds.ExtDataQueryOptions{
					DagID: dagID,
				}); err != nil {
					log.Warnf("[logic.CreateDag] Remove extData err, deail: %s", err.Error())
				}

				return "", ierrors.NewIError(ierrors.InternalError, "", nil)
			}
		}

		// 更新 dag 的 cron 配置
		dag.Cron = jobID
		err = m.mongo.UpdateDag(ctx, dag)
		if err != nil {
			log.Warnf("[logic.CreateDataFlow] UpdateDag with cron job id err, detail: %s", err.Error())
			return "", ierrors.NewIError(ierrors.InternalError, "", nil)
		}
	}

	// 记录日志
	go func() {
		detail, extMsg := common.GetLogBody(common.CreateTask, []interface{}{dag.Name},
			[]interface{}{})

		object := map[string]interface{}{
			"type":     common.DagTypeDataFlow,
			"id":       dagID,
			"name":     dag.Name,
			"creator":  userInfo.UserID,
			"priority": dag.Priority,
		}

		userInfo.Type = common.User.ToString()
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		m.logger.Log(drivenadapters.LogTypeASOperationLog, &drivenadapters.BuildARLogParams{
			Operation:   common.ArLogCreateDag,
			Description: detail,
			UserInfo:    userInfo,
			Object:      object,
		}, write)

		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: dagID,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dagID,
			Operation: drivenadapters.CreateOperation,
			ObjID:     dagID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}()

	return dagID, nil
}

// validateDataFlowParams 验证数据流参数
func (m *mgnt) validateDataFlowParams(param *CreateDataFlowReq) error {
	// 验证标题
	if param.Title == "" {
		return ierrors.NewIError(ierrors.InvalidParameter, "", "title is required")
	}

	// 验证触发器配置
	if param.TriggerConfig == nil {
		return ierrors.NewIError(ierrors.InvalidParameter, "", "trigger configuration required for data-flow")
	}

	// 验证步骤配置
	if len(param.Steps) < 2 {
		return ierrors.NewIError(ierrors.InvalidParameter, "", "at least 2 steps are required")
	}

	return nil
}

// UpdateDataFlowReq 更新数据流程请求
type UpdateDataFlowReq struct {
	Title         *string               `json:"title"`
	Description   *string               `json:"description"`
	Status        *string               `json:"status"`
	TriggerConfig *entity.TriggerConfig `json:"trigger_config"`
	Steps         *[]entity.Step        `json:"steps"`
	AppInfo       *entity.AppInfo       `json:"appinfo"`
	Version       *entity.Version       `json:"version"`
	ChangeLog     string                `json:"change_log"`
	DeBugID       string                `json:"debug_id"`
}

// UpdateDataFlow 更新数据流类型的流程
func (m *mgnt) UpdateDataFlow(ctx context.Context, dagID string, param *UpdateDataFlowReq, userInfo *drivenadapters.UserInfo) error {
	var err error
	var stopRunningTask bool
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 检查dag是否存在
	query := map[string]interface{}{"_id": dagID, "type": common.DagTypeDataFlow}
	dag, err := m.mongo.GetDagByFields(ctx, query)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": dagID})
		}
		log.Warnf("[logic.UpdateDataFlow] GetDagByFields err, query: %v, detail: %s", query, err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	dagBytes, err := json.Marshal(dag)
	if err != nil {
		log.Warnf("[logic.UpdateDataFlow] Marshal err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err.Error())
	}

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow: {perm.ModeifyOperation},
		},
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dagID, userInfo, opMap)
	if err != nil {
		return err
	}

	dag.AppInfo.Enable = false

	dag.ModifyBy = userInfo.UserID
	// 更新标题
	if param.Title != nil {
		title := strings.TrimSpace(*param.Title)
		if title == "" {
			return ierrors.NewIError(ierrors.InvalidParameter, "", "title is required")
		}
		_query := map[string]interface{}{"name": title}
		_dag, merr := m.mongo.GetDagByFields(ctx, _query)
		if merr == nil && _dag.ID != dagID {
			return ierrors.NewIError(ierrors.DuplicatedName, "", map[string]string{"title": title})
		}
		dag.Name = title
	}

	if param.DeBugID != "" {
		dag.DeBugID = param.DeBugID
		_, err = m.mongo.GetDag(ctx, param.DeBugID)
		if err != nil {
			if ierrors.IsNotFoundErr(err) {
				return ierrors.NewIError(ierrors.InvalidParameter, "", "debug id not found")
			}
			log.Warnf("[logic.UpdateDataFlow] GetDag err, deail: %s", err.Error())
			return ierrors.NewIError(ierrors.InternalError, "", nil)
		}
	}

	// 更新描述
	if param.Description != nil {
		dag.Description = strings.TrimSpace(*param.Description)
	}

	// 更新步骤
	if param.Steps != nil {
		if len(*param.Steps) < 2 {
			return ierrors.NewIError(ierrors.InvalidParameter, "", "at least 2 steps are required")
		}
		var tasks = make([]entity.Task, 0)
		var stepList = make([]map[string]interface{}, 0)
		steps := make([]entity.Step, len(*param.Steps))
		copy(steps, *param.Steps)
		m.buildTasks(&steps[0], steps, &tasks, nil, &stepList, nil, nil)

		dag.SubIDs, err = m.validSubFlowInSteps(ctx, dagID, stepList)
		if err != nil {
			return err
		}

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

		triggerStep := (*param.Steps)[0]

		if triggerStep.Operator != common.MDLDataViewTrigger {
			dag.IncValues = map[string]any{}
		} else {
			oldTriggerStep := dag.Steps[0]
			var oldParam, newParam DataViewParam

			oldParam.Init(oldTriggerStep.Parameters)
			newParam.Init(triggerStep.Parameters)

			if newParam.SyncMode != SyncModeIncremental {
				dag.IncValues = map[string]any{}
			} else if oldParam.SyncMode != SyncModeIncremental || newParam.ID != oldParam.ID {
				dag.IncValues = map[string]any{
					newParam.IncrementField: newParam.IncrementValue,
				}
			} else if newParam.IncrementField != oldParam.IncrementField {
				if dag.IncValues == nil {
					dag.IncValues = map[string]any{}
				}
				dag.IncValues[newParam.IncrementField] = newParam.IncrementValue
			}
		}

		dag.Steps = *param.Steps
		dag.Tasks = tasks
	}

	// 更新触发器配置
	if param.TriggerConfig != nil {
		triggerType := m.getTriggerType(param.TriggerConfig.Operator)
		dag.Trigger = triggerType
		dag.TriggerConfig = param.TriggerConfig

		// 处理定时任务配置
		if param.TriggerConfig.Cron != "" && !strings.HasPrefix(param.TriggerConfig.Cron, common.CronTrigger) {
			if dag.Cron != "" {
				// 更新现有的定时任务
				err = m.ecron.UpdateCronJob(
					ctx,
					fmt.Sprintf("auto_%s", dagID),
					fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/cron/%s",
						m.config.ContentAutomation.PrivateHost,
						m.config.ContentAutomation.PrivatePort,
						dagID),
					param.TriggerConfig.Cron,
					dag.Cron,
				)
			} else {
				// 创建新的定时任务
				jobID, exist, err := m.ecron.RegisterCronJob(
					ctx,
					fmt.Sprintf("auto_%s", dagID),
					fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/cron/%s",
						m.config.ContentAutomation.PrivateHost,
						m.config.ContentAutomation.PrivatePort,
						dagID),
					param.TriggerConfig.Cron,
				)
				if err != nil {
					if !exist {
						log.Warnf("[logic.UpdateDataFlow] RegisterCronJob err, detail: %s", err.Error())
						return ierrors.NewIError(ierrors.InternalError, "", nil)
					}
				} else {
					dag.Cron = jobID
				}
			}
		} else if dag.Cron != "" {
			// 删除现有的定时任务
			err = m.ecron.DeleteEcronJob(ctx, dag.Cron)
			if err != nil {
				log.Warnf("[logic.UpdateDataFlow] DeleteEcronJob err, detail: %s", err.Error())
				return ierrors.NewIError(ierrors.InternalError, "", nil)
			}
			dag.Cron = ""
		}
	}

	// 更新状态
	if param.Status != nil {
		dag.Status = entity.DagStatusNormal
		if *param.Status == common.StoppedStatus {
			dag.Status = entity.DagStatusStopped
			stopRunningTask = true
		}
	}

	var dagVersions []*entity.DagVersion
	dagVersions, err = m.buildDagVersions(&BuildDagVersionParams{
		OldDagBytes: dagBytes,
		Dag:         dag,
		CurVersion:  param.Version,
		ChangeLog:   param.ChangeLog,
		UserID:      userInfo.UserID,
		IsCreate:    false,
	})
	if err != nil {
		return err
	}

	err = m.mongo.WithTransaction(ctx, func(nctx context.Context, ms mod.Store) error {
		if err = ms.UpdateDag(nctx, dag); err != nil {
			return err
		}

		for _, dagVersion := range dagVersions {
			_, err = ms.CreateDagVersion(nctx, dagVersion)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Warnf("[logic.UpdateDataFlow] Transaction err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 处理停止状态下的运行中任务
	go func(stopRunningTask bool) {
		if stopRunningTask {
			gctx := context.Background()
			var input = &mod.ListDagInstanceInput{
				DagIDs: []string{dagID},
				Status: []entity.DagInstanceStatus{
					entity.DagInstanceStatusRunning,
					entity.DagInstanceStatusScheduled,
					entity.DagInstanceStatusInit,
					entity.DagInstanceStatusBlocked,
				},
			}
			dagInsList, err := m.mongo.ListDagInstance(gctx, input)
			if err != nil {
				log.Warnf("[logic.UpdateDataFlow] ListDagInstance err, detail: %s", err.Error())
				return
			}
			var dagInsArr = make([]*entity.DagInstance, 0)
			for _, dagIns := range dagInsList {
				_dagIns := *dagIns
				_dagIns.Status = entity.DagInstanceStatusCancled
				dagInsArr = append(dagInsArr, &_dagIns)
			}
			if err = m.mongo.BatchUpdateDagIns(gctx, dagInsArr); err != nil {
				log.Warnf("[logic.UpdateDataFlow] BatchUpdateDagIns err, detail: %s", err.Error())
			}
		}
	}(stopRunningTask)

	// 记录日志
	go func() {
		detail, extMsg := common.GetLogBody(common.UpdateTask, []interface{}{dag.Name}, []interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: dagID,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dagID,
			Operation: drivenadapters.UpdateOperation,
			ObjID:     dagID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}()

	go func() {
		resourceID := fmt.Sprintf("%s:%s", dagID, common.DagTypeDataFlow)
		m.permPolicy.HandlePolicyNameChange(resourceID, dag.Name, common.DagTypeDataFlow)
	}()

	return nil
}

// DeleteDataFlow 管理员删除数据流
func (m *mgnt) DeleteDataFlow(ctx context.Context, dagID, bizDomainID string, userInfo *drivenadapters.UserInfo) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow: {perm.DeleteOperation},
		},
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dagID, userInfo, opMap)
	if err != nil {
		return err
	}

	// 管理员可以删除自己的DAG或任何dataflow类型的DAG
	query := bson.M{
		"_id": dagID,
		"$or": []bson.M{
			{"userid": userInfo.UserID},
			{"type": common.DagTypeDataFlow},
		},
	}

	dag, err := m.mongo.GetDagByFields(ctx, query)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return nil
		}
		log.Warnf("[logic.DeleteDataFlow] GetDagByFields err, detail: %s", err.Error())
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
	}

	// 执行删除逻辑...
	err = m.deleteDag(ctx, dag, userInfo, bizDomainID)
	if err != nil {
		log.Warnf("[logic.DeleteDataFlow] deleteDag err, detail: %s", err.Error())
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
	}

	go func() {
		resourceID := fmt.Sprintf("%s:%s", dagID, common.DagTypeDataFlow)
		cctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		err = m.permPolicy.DeletePolicy(cctx, resourceID)
		if err != nil {
			log.Warnf("[logic.DeleteDataFlow] DeletePolicy err, detail: %s", err.Error())
		}
	}()

	return nil
}
