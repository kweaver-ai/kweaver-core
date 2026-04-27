package executor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
)

var (
	executorOnce sync.Once
	executor     ExecutorHandler
)

type ExecutorAccessorDto struct {
	ID   *string `json:"id"`
	Type *string `json:"type"`
	Name *string `json:"name"`
}

type ExecutorActionInputDto struct {
	Key      *string `json:"key"`
	Name     *string `json:"name"`
	Type     *string `json:"type"`
	Required *bool   `json:"required"`
}

type ExecutorActionOutputDto struct {
	Key  *string `json:"key"`
	Name *string `json:"name"`
	Type *string `json:"type"`
}

type ExecutorActionDto struct {
	ID          *string                    `json:"id"`
	Name        *string                    `json:"name"`
	Operator    *string                    `json:"operator"`
	Description *string                    `json:"description"`
	Group       *string                    `json:"group,omitempty"`
	Type        *string                    `json:"type"`
	Inputs      *[]ExecutorActionInputDto  `json:"inputs,omitempty"`
	Outputs     *[]ExecutorActionOutputDto `json:"outputs,omitempty"`
	Config      *map[string]interface{}    `json:"config,omitempty"`
	CreatedAt   *int64                     `json:"created_at"`
	UpdatedAt   *int64                     `json:"updated_at"`
}

type ExecutorCreatorDto struct {
	ID   *string `json:"id"`
	Name *string `json:"name"`
}

type ExecutorDto struct {
	ID          *string                `json:"id"`
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	Status      *int                   `json:"status"`
	Creator     *ExecutorCreatorDto    `json:"creator"`
	CreatedAt   *int64                 `json:"created_at"`
	UpdatedAt   *int64                 `json:"updated_at"`
	Accessors   []*ExecutorAccessorDto `json:"accessors,omitempty"`
	Actions     []*ExecutorActionDto   `json:"actions,omitempty"`
}

type ImportAgentsDto struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Accessors   []*ExecutorAccessorDto `json:"accessors,omitempty"`
	Status      int                    `json:"status"`
	AgentIDs    []string               `json:"agent_ids"`
	OnDup       int                    `json:"ondup"`
	ActionOnDup int                    `json:"action_ondup"`
}

type ImportAgentsResult struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
}

type CheckNameDto struct {
	Name *string `json:"name"`
}

type ExecutorHandler interface {
	CheckCreateExecutor(ctx context.Context, dto ExecutorDto, userInfo *drivenadapters.UserInfo) (bool, error)
	CheckUpdateExecutor(ctx context.Context, id uint64, dto ExecutorDto, userInfo *drivenadapters.UserInfo) (bool, error)
	CheckCreateExecutorAction(ctx context.Context, executorID uint64, dto ExecutorActionDto, userInfo *drivenadapters.UserInfo) (bool, error)
	CheckUpdateExecutorAction(ctx context.Context, executorID uint64, actionID uint64, dto ExecutorActionDto, userInfo *drivenadapters.UserInfo) (bool, error)

	CreateExecutor(ctx context.Context, dto ExecutorDto, userInfo *drivenadapters.UserInfo) (uint64, error)
	UpdateExecutor(ctx context.Context, id uint64, dto ExecutorDto, userInfo *drivenadapters.UserInfo) error
	GetExecutors(ctx context.Context, userInfo *drivenadapters.UserInfo) ([]*ExecutorDto, error)
	GetExecutor(ctx context.Context, id uint64, userInfo *drivenadapters.UserInfo) (*ExecutorDto, error)
	DeleteExecutor(ctx context.Context, id uint64, userInfo *drivenadapters.UserInfo) error
	CreateExecutorAction(ctx context.Context, executorID uint64, dto ExecutorActionDto, userInfo *drivenadapters.UserInfo) (uint64, error)
	UpdateExecutorAction(ctx context.Context, executorID uint64, actionID uint64, dto ExecutorActionDto, userInfo *drivenadapters.UserInfo) error
	DeleteExecutorAction(ctx context.Context, executorID uint64, actionID uint64, userInfo *drivenadapters.UserInfo) error
	GetAccessableExecutors(ctx context.Context, userInfo *drivenadapters.UserInfo) ([]*ExecutorDto, error)

	ImportAgents(ctx context.Context, userInfo *drivenadapters.UserInfo, dto *ImportAgentsDto) (result *ImportAgentsResult, err error)
}

type ExecutorHandlerImpl struct {
	executorDao    rds.ExecutorDao
	userManagement drivenadapters.UserManagement
	executeMethods entity.ExecuteMethods
	logger         drivenadapters.Logger
	ad             drivenadapters.AnyData
}

func NewExecutorHandler() ExecutorHandler {

	executorOnce.Do(func() {
		em := entity.ExecuteMethods{
			Publish: mod.NewMQHandler().Publish,
		}
		executor = &ExecutorHandlerImpl{
			executorDao:    rds.GetExecutorDao(),
			userManagement: drivenadapters.NewUserManagement(),
			executeMethods: em,
			logger:         drivenadapters.NewLogger(),
			ad:             drivenadapters.NewAnyData(),
		}
	})

	return executor
}

func (e *ExecutorHandlerImpl) CheckCreateExecutor(ctx context.Context, dto ExecutorDto, userInfo *drivenadapters.UserInfo) (bool, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	result, err := e.executorDao.CheckExecutor(ctx, &rds.ExecutorModel{
		Name:      dto.Name,
		CreatorID: &userInfo.UserID,
	})

	if err != nil {
		log.Warnf("logic.CheckCreateExecutor err, detail: %s", err.Error())
		return false, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	return result, nil
}

func (e *ExecutorHandlerImpl) CheckUpdateExecutor(ctx context.Context, id uint64, dto ExecutorDto, userInfo *drivenadapters.UserInfo) (bool, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	result, err := e.executorDao.CheckExecutor(ctx, &rds.ExecutorModel{
		ID:        &id,
		Name:      dto.Name,
		CreatorID: &userInfo.UserID,
	})

	if err != nil {
		log.Warnf("logic.CheckUpdateExecutor err, detail: %s", err.Error())
		return false, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	return result, nil
}

func (e *ExecutorHandlerImpl) CheckCreateExecutorAction(ctx context.Context, executorID uint64, dto ExecutorActionDto, userInfo *drivenadapters.UserInfo) (bool, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	executor, err := e.executorDao.GetExecutor(ctx, executorID)

	if err != nil {
		log.Warnf("logic.CheckCreateExecutorAction err, detail: %s", err.Error())
		return false, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if executor == nil {
		log.Warnf("logic.CheckCreateExecutorAction err, detail: executor not found, id: %s", executorID)
		return false, ierrors.NewIError(ierrors.ExecutorNotFound, "", map[string]interface{}{"id": executorID})
	}

	if *executor.CreatorID != userInfo.UserID {
		log.Warnf("logic.CheckCreateExecutorAction err, detail: is not owner, id: %s, userid: %s", executorID, userInfo.UserID)
		return false, ierrors.NewIError(ierrors.ExecutorForbidden, "", map[string]interface{}{"id": executorID})
	}

	result, err := e.executorDao.CheckExecutorAction(ctx, &rds.ExecutorActionModel{
		ExecutorID: &executorID,
		Name:       dto.Name,
	})

	if err != nil {
		log.Warnf("logic.CheckCreateExecutorAction err, detail: %s", err.Error())
		return false, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	return result, nil
}

func (e *ExecutorHandlerImpl) CheckUpdateExecutorAction(ctx context.Context, executorID uint64, actionID uint64, dto ExecutorActionDto, userInfo *drivenadapters.UserInfo) (bool, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	executor, err := e.executorDao.GetExecutor(ctx, executorID)

	if err != nil {
		log.Warnf("logic.CheckUpdateExecutorAction err, detail: %s", err.Error())
		return false, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if executor == nil {
		log.Warnf("logic.CheckUpdateExecutorAction err, detail: executor not found, id: %s", executorID)
		return false, ierrors.NewIError(ierrors.ExecutorNotFound, "", map[string]interface{}{"id": executorID})
	}

	if *executor.CreatorID != userInfo.UserID {
		log.Warnf("logic.CheckUpdateExecutorAction err, detail: is not owner, id: %s, userid: %s", executorID, userInfo.UserID)
		return false, ierrors.NewIError(ierrors.ExecutorForbidden, "", map[string]interface{}{"id": executorID})
	}

	result, err := e.executorDao.CheckExecutorAction(ctx, &rds.ExecutorActionModel{
		ID:         &actionID,
		ExecutorID: &executorID,
		Name:       dto.Name,
	})

	if err != nil {
		log.Warnf("logic.CheckUpdateExecutorAction err, detail: %s", err.Error())
		return false, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	return result, nil
}

func (e *ExecutorHandlerImpl) CreateExecutor(ctx context.Context, dto ExecutorDto, userInfo *drivenadapters.UserInfo) (uint64, error) {

	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	result, err := e.executorDao.CheckExecutor(ctx, &rds.ExecutorModel{
		Name:      dto.Name,
		CreatorID: &userInfo.UserID,
	})

	if err != nil {
		log.Warnf("logic.CheckCreateExecutor err, detail: %s", err.Error())
		return 0, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if !result {
		return 0, ierrors.NewIError(ierrors.DuplicatedName, "", map[string]interface{}{"name": dto.Name})
	}

	executorId, _ := utils.GetUniqueID()
	now := time.Now().UnixMilli()

	accessors := make([]*rds.ExecutorAccessorModel, 0)
	actions := make([]*rds.ExecutorActionModel, 0)
	accessorNames := make([]string, 0)
	actionNames := make([]string, 0)

	if dto.Accessors != nil {
		accessorIDs := make(map[string]string, 0)
		for _, accessor := range dto.Accessors {
			accessorIDs[*accessor.ID] = *accessor.Type
		}

		accessorNameMap, _ := e.userManagement.GetNameByAccessorIDs(accessorIDs)

		if len(accessorNameMap) > 0 {
			for itemID, accessorName := range accessorNameMap {
				accessorID := itemID
				accessorNames = append(accessorNames, accessorName)
				accessorType := accessorIDs[accessorID]
				id, _ := utils.GetUniqueID()
				accessors = append(accessors, &rds.ExecutorAccessorModel{
					ID:           &id,
					ExecutorID:   &executorId,
					AccessorID:   &accessorID,
					AccessorType: &accessorType,
				})
			}
		}
	}

	if dto.Actions != nil {
		for _, action := range dto.Actions {

			var (
				inputs      = ""
				outputs     = ""
				config      = ""
				group       = ""
				description = ""
			)

			if action.Inputs != nil {
				inputs, _ = jsoniter.MarshalToString(action.Inputs)
			}

			if action.Outputs != nil {
				outputs, _ = jsoniter.MarshalToString(action.Outputs)
			}

			if action.Config != nil {
				config, _ = jsoniter.MarshalToString(action.Config)
			}

			if action.Group != nil {
				group = *action.Group
			}

			if action.Description != nil {
				description = *action.Description
			}

			actionNames = append(actionNames, *action.Name)

			id, _ := utils.GetUniqueID()
			operator := fmt.Sprintf("@custom/%d/%d", executorId, id)

			actions = append(actions, &rds.ExecutorActionModel{
				ID:          &id,
				ExecutorID:  &executorId,
				Operator:    &operator,
				Name:        action.Name,
				Description: &description,
				Group:       &group,
				Type:        action.Type,
				Inputs:      &inputs,
				Outputs:     &outputs,
				Config:      &config,
				CreatedAt:   &now,
				UpdatedAt:   &now,
			})
		}
	}

	var description = ""

	if dto.Description != nil {
		description = *dto.Description
	}

	executor := &rds.ExecutorModel{
		ID:          &executorId,
		Name:        dto.Name,
		Description: &description,
		Status:      dto.Status,
		CreatorID:   &userInfo.UserID,
		CreatedAt:   &now,
		UpdatedAt:   &now,
		Accessors:   accessors,
		Actions:     actions,
	}

	if executor.Description == nil {
		executor.Description = new(string)
	}

	err = e.executorDao.CreateExecutor(ctx, executor)

	if err != nil {
		log.Warnf("logic.CreateExecutor err, detail: %s", err.Error())
		return 0, err
	}

	go func() {

		var status string
		if *executor.Status == 0 {
			status = common.GetLocale("disabled")
		} else {
			status = common.GetLocale("enabled")
		}
		detail, extMsg := common.GetLogBody(common.CreateCustomExecutor, []interface{}{*executor.Name},
			[]interface{}{
				strings.Join(accessorNames, ","),
				status,
				strings.Join(actionNames, ","),
			})

		object := map[string]interface{}{
			"id":      *executor.ID,
			"name":    *executor.Name,
			"creator": userInfo.UserID,
		}
		userInfo.Type = common.User.ToString()

		writer := &drivenadapters.JSONLogWriter{SendFunc: e.executeMethods.Publish}
		e.logger.Log(drivenadapters.LogTypeASOperationLog, &drivenadapters.BuildARLogParams{
			Operation:   common.CreateCustomExecutor,
			Description: detail,
			UserInfo:    userInfo,
			Object:      object,
		}, writer)

		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		e.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: fmt.Sprintf("%v", executorId),
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		}, writer)
	}()

	return executorId, nil
}

func (e *ExecutorHandlerImpl) UpdateExecutor(ctx context.Context, id uint64, dto ExecutorDto, userInfo *drivenadapters.UserInfo) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	originExecutor, err := e.executorDao.GetExecutor(ctx, id)

	if err != nil {
		log.Warnf("logic.UpdateExecutor err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if originExecutor == nil {
		log.Warnf("logic.UpdateExecutor err, detail: executor not found, id: %s")
		return ierrors.NewIError(ierrors.ExecutorNotFound, "", map[string]interface{}{"id": *dto.ID})
	}

	if *originExecutor.CreatorID != userInfo.UserID {
		log.Warnf("logic.UpdateExecutor err, detail: is not owner, id: %s")
		return ierrors.NewIError(ierrors.ExecutorForbidden, "", map[string]interface{}{"id": *dto.ID})
	}

	if dto.Name != nil && *dto.Name != "" {
		result, err := e.executorDao.CheckExecutor(ctx, &rds.ExecutorModel{
			ID:        &id,
			Name:      dto.Name,
			CreatorID: &userInfo.UserID,
		})

		if err != nil {
			log.Warnf("logic.UpdateExecutor err, detail: %s", err.Error())
			return ierrors.NewIError(ierrors.InternalError, "", err)
		}

		if !result {
			return ierrors.NewIError(ierrors.DuplicatedName, "", map[string]interface{}{"name": dto.Name})
		}
	}

	now := time.Now().UnixMilli()
	executor := &rds.ExecutorModel{
		ID:        originExecutor.ID,
		UpdatedAt: &now,
	}

	if dto.Name != nil && *dto.Name != "" {
		executor.Name = dto.Name
	}

	if dto.Description != nil {
		executor.Description = dto.Description
	}

	if dto.Status != nil {
		executor.Status = dto.Status
	}

	accessorNames := make([]string, 0)
	actionNames := make([]string, 0)

	if dto.Accessors != nil {
		accessors := make([]*rds.ExecutorAccessorModel, 0)
		accessorIDs := make(map[string]string, 0)
		for _, accessor := range dto.Accessors {
			accessorIDs[*accessor.ID] = *accessor.Type
		}

		accessorNameMap, _ := e.userManagement.GetNameByAccessorIDs(accessorIDs)

		if len(accessorNameMap) > 0 {
			for itemID, accessorName := range accessorNameMap {
				accessorID := itemID
				accessorNames = append(accessorNames, accessorName)
				accessorType := accessorIDs[accessorID]
				id, _ := utils.GetUniqueID()
				accessors = append(accessors, &rds.ExecutorAccessorModel{
					ID:           &id,
					ExecutorID:   executor.ID,
					AccessorID:   &accessorID,
					AccessorType: &accessorType,
				})
			}

			executor.Accessors = accessors
		}
	}

	err = e.executorDao.UpdateExecutor(ctx, executor)

	if err != nil {
		log.Warnf("logic.UpdateExecutor err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err)
	}

	go func() {
		var status string
		if executor.Status == nil {
			if *originExecutor.Status == 0 {
				status = common.GetLocale("disabled")
			} else {
				status = common.GetLocale("enabled")
			}
		} else {
			if *executor.Status == 0 {
				status = common.GetLocale("disabled")
			} else {
				status = common.GetLocale("enabled")
			}
		}
		executorName := *originExecutor.Name
		if executor.Name != nil {
			executorName = *executor.Name
		}
		detail, extMsg := common.GetLogBody(common.UpdateCustomExecutor, []interface{}{executorName},
			[]interface{}{
				strings.Join(accessorNames, ","),
				status,
				strings.Join(actionNames, ","),
			})

		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		e.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: fmt.Sprintf("%v", id),
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		}, &drivenadapters.JSONLogWriter{SendFunc: e.executeMethods.Publish})
	}()

	return nil
}

func (e *ExecutorHandlerImpl) GetExecutors(ctx context.Context, userInfo *drivenadapters.UserInfo) ([]*ExecutorDto, error) {

	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	executors, err := e.executorDao.GetExecutors(ctx, userInfo.UserID)

	if err != nil {
		log.Warnf("logic.GetExecutors err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	results, err := e.executorModelToDto(executors)

	if err != nil {
		log.Warnf("logic.GetExecutors err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	return results, nil
}

func (e *ExecutorHandlerImpl) GetExecutor(ctx context.Context, id uint64, userInfo *drivenadapters.UserInfo) (*ExecutorDto, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	executor, err := e.executorDao.GetExecutor(ctx, id)

	if err != nil {
		log.Warnf("logic.GetExecutor err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if executor == nil {
		log.Warnf("logic.GetExecutor err, detail: executor not found, id: %s", id)
		return nil, ierrors.NewIError(ierrors.ExecutorNotFound, "", map[string]interface{}{"id": id})
	}

	if err != nil {
		log.Warnf("logic.GetExecutor err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if *executor.CreatorID != userInfo.UserID {
		log.Warnf("logic.GetExecutor err, detail: is not owner, id: %s, userid: %s", id, userInfo.UserID)
		return nil, ierrors.NewIError(ierrors.ExecutorForbidden, "", map[string]interface{}{"id": id})
	}

	accessors, err := e.executorDao.GetExecutorAccessors(ctx, id)
	if err != nil {
		log.Warnf("logic.GetExecutor err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}
	executor.Accessors = accessors

	actions, err := e.executorDao.GetExecutorActions(ctx, id)
	if err != nil {
		log.Warnf("logic.GetExecutor err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}
	executor.Actions = actions

	accessorMap := map[string]string{*executor.CreatorID: common.User.ToString()}

	if executor.Accessors != nil {
		for _, accessor := range executor.Accessors {
			accessorMap[*accessor.AccessorID] = *accessor.AccessorType
		}
	}

	accessorNames, err := e.userManagement.GetNameByAccessorIDs(accessorMap)

	if err != nil {
		log.Warnf("logic.GetExecutor err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	creator := &ExecutorCreatorDto{
		ID: executor.CreatorID,
	}

	if name, ok := accessorNames[*executor.CreatorID]; ok {
		creator.Name = &name
	}

	dtoID := fmt.Sprintf("%d", *executor.ID)

	dto := &ExecutorDto{
		ID:          &dtoID,
		Name:        executor.Name,
		Description: executor.Description,
		Status:      executor.Status,
		Creator:     creator,
		CreatedAt:   executor.CreatedAt,
		UpdatedAt:   executor.UpdatedAt,
	}

	if executor.Accessors != nil {
		accessors := make([]*ExecutorAccessorDto, 0)
		for _, accessor := range executor.Accessors {
			if name, ok := accessorNames[*accessor.AccessorID]; ok {
				dto := ExecutorAccessorDto{
					ID:   accessor.AccessorID,
					Type: accessor.AccessorType,
					Name: &name,
				}
				accessors = append(accessors, &dto)
			}
		}
		dto.Accessors = accessors
	}

	if executor.Actions != nil {
		actionDtos, err := e.actionModelToDto(executor.Actions)

		if err != nil {
			log.Warnf("logic.GetExecutor err, detail: %s", err.Error())
			return nil, ierrors.NewIError(ierrors.InternalError, "", err)
		}

		dto.Actions = actionDtos
	}

	return dto, nil
}

func (e *ExecutorHandlerImpl) DeleteExecutor(ctx context.Context, id uint64, userInfo *drivenadapters.UserInfo) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	executor, err := e.executorDao.GetExecutor(ctx, id)

	if err != nil {
		log.Warnf("logic.DeleteExecutor err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if executor == nil {
		log.Warnf("logic.DeleteExecutor err, detail: executor not found, id: %s", id)
		return ierrors.NewIError(ierrors.ExecutorNotFound, "", map[string]interface{}{"id": id})
	}

	if *executor.CreatorID != userInfo.UserID {
		log.Warnf("logic.DeleteExecutor err, detail: is not owner, id: %s, userid: %s", id, userInfo.UserID)
		return ierrors.NewIError(ierrors.ExecutorForbidden, "", map[string]interface{}{"id": id})
	}

	err = e.executorDao.DeleteExecutor(ctx, id)

	if err != nil {
		log.Warnf("logic.DeleteExecutor err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err)
	}

	go func() {
		var status string
		if *executor.Status == 0 {
			status = common.GetLocale("disabled")
		} else {
			status = common.GetLocale("enabled")
		}

		detail, extMsg := common.GetLogBody(common.DeleteCustomExecutor, []interface{}{*executor.Name},
			[]interface{}{
				status,
			})

		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		e.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: fmt.Sprintf("%v", id),
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		}, &drivenadapters.JSONLogWriter{SendFunc: e.executeMethods.Publish})
	}()

	return nil
}

func (e *ExecutorHandlerImpl) CreateExecutorAction(ctx context.Context, executorID uint64, dto ExecutorActionDto, userInfo *drivenadapters.UserInfo) (uint64, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	executor, err := e.executorDao.GetExecutor(ctx, executorID)

	if err != nil {
		log.Warnf("logic.CreateExecutorAction err, detail: %s", err.Error())
		return 0, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if executor == nil {
		log.Warnf("logic.CreateExecutorAction err, detail: executor not found, id: %s", executorID)
		return 0, ierrors.NewIError(ierrors.ExecutorNotFound, "", map[string]interface{}{"id": executorID})
	}

	if *executor.CreatorID != userInfo.UserID {
		log.Warnf("logic.CreateExecutorAction err, detail: is not owner, id: %s, userid: %s", executorID, userInfo.UserID)
		return 0, ierrors.NewIError(ierrors.ExecutorForbidden, "", map[string]interface{}{"id": executorID})
	}

	result, err := e.executorDao.CheckExecutorAction(ctx, &rds.ExecutorActionModel{
		ExecutorID: &executorID,
		Name:       dto.Name,
	})

	if err != nil {
		log.Warnf("logic.CheckCreateExecutorAction err, detail: %s", err.Error())
		return 0, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if !result {
		return 0, ierrors.NewIError(ierrors.DuplicatedName, "", map[string]interface{}{"name": dto.Name})
	}

	actionID, _ := utils.GetUniqueID()
	operator := fmt.Sprintf("@custom/%d/%d", executorID, actionID)
	now := time.Now().UnixMilli()

	var (
		inputs      = ""
		outputs     = ""
		config      = ""
		group       = ""
		description = ""
	)

	if dto.Inputs != nil {
		inputs, _ = jsoniter.MarshalToString(dto.Inputs)
	}

	if dto.Outputs != nil {
		outputs, _ = jsoniter.MarshalToString(dto.Outputs)
	}

	if dto.Config != nil {
		config, _ = jsoniter.MarshalToString(dto.Config)
	}

	if dto.Group != nil {
		group = *dto.Group
	}

	if dto.Description != nil {
		description = *dto.Description
	}

	action := &rds.ExecutorActionModel{
		ID:          &actionID,
		ExecutorID:  &executorID,
		Operator:    &operator,
		Name:        dto.Name,
		Description: &description,
		Group:       &group,
		Type:        dto.Type,
		Inputs:      &inputs,
		Outputs:     &outputs,
		Config:      &config,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	err = e.executorDao.CreateExecutorAction(ctx, action)

	if err != nil {
		log.Warnf("logic.CreateExecutorAction err, detail: %s", err.Error())
		return 0, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	go func() {
		detail, extMsg := common.GetLogBody(common.CreateCustomExecutorAction, []interface{}{*action.Name},
			[]interface{}{})

		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		e.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: fmt.Sprintf("%v", executorID),
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		}, &drivenadapters.JSONLogWriter{SendFunc: e.executeMethods.Publish})
	}()

	return actionID, nil
}

func (e *ExecutorHandlerImpl) UpdateExecutorAction(ctx context.Context, executorID uint64, actionID uint64, dto ExecutorActionDto, userInfo *drivenadapters.UserInfo) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	executor, err := e.executorDao.GetExecutor(ctx, executorID)

	if err != nil {
		log.Warnf("logic.UpdateExecutorAction err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if executor == nil {
		log.Warnf("logic.UpdateExecutorAction err, detail: executor not found, id: %s", executorID)
		return ierrors.NewIError(ierrors.ExecutorNotFound, "", map[string]interface{}{"id": executorID})
	}

	if *executor.CreatorID != userInfo.UserID {
		log.Warnf("logic.UpdateExecutorAction err, detail: is not owner, id: %s, userid: %s", executorID, userInfo.UserID)
		return ierrors.NewIError(ierrors.ExecutorForbidden, "", map[string]interface{}{"id": executorID})
	}

	now := time.Now().UnixMilli()

	action := &rds.ExecutorActionModel{
		ID:         &actionID,
		ExecutorID: &executorID,
		UpdatedAt:  &now,
	}

	if dto.Name != nil && *dto.Name != "" {
		result, err := e.executorDao.CheckExecutorAction(ctx, &rds.ExecutorActionModel{
			ExecutorID: &executorID,
			ID:         &actionID,
			Name:       dto.Name,
		})

		if err != nil {
			log.Warnf("logic.CheckCreateExecutorAction err, detail: %s", err.Error())
			return ierrors.NewIError(ierrors.InternalError, "", err)
		}

		if !result {
			return ierrors.NewIError(ierrors.DuplicatedName, "", map[string]interface{}{"name": dto.Name})
		}

		action.Name = dto.Name
	}

	var inputs string
	var outputs string
	var config string

	if dto.Inputs != nil {
		inputs, _ = jsoniter.MarshalToString(dto.Inputs)
		action.Inputs = &inputs
	}

	if dto.Outputs != nil {
		outputs, _ = jsoniter.MarshalToString(dto.Outputs)
		action.Outputs = &outputs
	}

	if dto.Config != nil {
		config, _ = jsoniter.MarshalToString(dto.Config)
		action.Config = &config
	}

	if dto.Description != nil {
		action.Description = dto.Description
	}

	err = e.executorDao.UpdateExecutorAction(ctx, action)

	if err != nil {
		log.Warnf("logic.UpdateExecutorAction err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err)
	}

	go func() {
		detail, extMsg := common.GetLogBody(common.UpdateCustomExecutorAction, []interface{}{*action.Name},
			[]interface{}{})

		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		e.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: fmt.Sprintf("%v", executorID),
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		}, &drivenadapters.JSONLogWriter{SendFunc: e.executeMethods.Publish})
	}()

	return nil
}

func (e *ExecutorHandlerImpl) DeleteExecutorAction(ctx context.Context, executorID uint64, actionID uint64, userInfo *drivenadapters.UserInfo) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	executor, err := e.executorDao.GetExecutor(ctx, executorID)

	if err != nil {
		log.Warnf("logic.DeleteExecutorAction err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if executor == nil {
		log.Warnf("logic.DeleteExecutorAction err, detail: executor not found, id: %s", executorID)
		return ierrors.NewIError(ierrors.ExecutorNotFound, "", map[string]interface{}{"id": executorID})
	}

	if *executor.CreatorID != userInfo.UserID {
		log.Warnf("logic.DeleteExecutorAction err, detail: is not owner, id: %s, userid: %s", executorID, userInfo.UserID)
		return ierrors.NewIError(ierrors.ExecutorForbidden, "", map[string]interface{}{"id": executorID})
	}

	now := time.Now().UnixMilli()

	action := &rds.ExecutorActionModel{
		ID:         &actionID,
		ExecutorID: &executorID,
		UpdatedAt:  &now,
	}

	err = e.executorDao.DeleteExecutorAction(ctx, action)

	if err != nil {
		log.Warnf("logic.DeleteExecutorAction err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err)
	}

	go func() {
		detail, extMsg := common.GetLogBody(common.DeleteCustomExecutorAction, []interface{}{*action.ID},
			[]interface{}{})

		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		e.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: fmt.Sprintf("%v", executorID),
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		}, &drivenadapters.JSONLogWriter{SendFunc: e.executeMethods.Publish})
	}()

	return nil
}

func (e *ExecutorHandlerImpl) GetAccessableExecutors(ctx context.Context, userInfo *drivenadapters.UserInfo) ([]*ExecutorDto, error) {

	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	accessorIDs, err := e.userManagement.GetUserAccessorIDs(userInfo.UserID)

	if err != nil {
		log.Warnf("logic.GetAccessableExecutors err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	executors, err := e.executorDao.GetAccessableExecutors(ctx, userInfo.UserID, accessorIDs)

	if err != nil {
		log.Warnf("logic.GetAccessableExecutors err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	results, err := e.executorModelToDto(executors)

	if err != nil {
		log.Warnf("logic.GetAccessableExecutors err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	return results, nil
}

func (e *ExecutorHandlerImpl) executorModelToDto(models []*rds.ExecutorModel) ([]*ExecutorDto, error) {

	executors := make([]*ExecutorDto, 0)

	if len(models) == 0 {
		return executors, nil
	}

	accessorMap := make(map[string]string, 0)

	for _, model := range models {
		accessorMap[*model.CreatorID] = common.User.ToString()

		if model.Accessors != nil {
			for _, accessor := range model.Accessors {
				accessorMap[*accessor.AccessorID] = *accessor.AccessorType
			}
		}
	}

	accessorNames, err := e.userManagement.GetNameByAccessorIDs(accessorMap)

	if err != nil {
		return nil, err
	}

	for _, model := range models {
		dtoID := fmt.Sprintf("%d", *model.ID)
		dto := &ExecutorDto{
			ID:          &dtoID,
			Name:        model.Name,
			Description: model.Description,
			Status:      model.Status,
			Creator:     &ExecutorCreatorDto{ID: model.CreatorID},
			CreatedAt:   model.CreatedAt,
			UpdatedAt:   model.UpdatedAt,
		}

		if creatorName, ok := accessorNames[*model.CreatorID]; ok {
			dto.Creator.Name = &creatorName
		}

		if model.Accessors != nil {
			accessors := make([]*ExecutorAccessorDto, 0)
			for _, accessor := range model.Accessors {
				accessorName, ok := accessorNames[*accessor.AccessorID]
				if !ok {
					continue
				}
				accessors = append(accessors, &ExecutorAccessorDto{
					ID:   accessor.AccessorID,
					Type: accessor.AccessorType,
					Name: &accessorName,
				})
			}

			dto.Accessors = accessors
		}

		if model.Actions != nil {

			actions, err := e.actionModelToDto(model.Actions)
			if err != nil {
				return nil, err
			}

			dto.Actions = actions
		}

		executors = append(executors, dto)
	}

	return executors, nil
}

func (e *ExecutorHandlerImpl) actionModelToDto(models []*rds.ExecutorActionModel) ([]*ExecutorActionDto, error) {
	actions := make([]*ExecutorActionDto, 0)

	if len(models) > 0 {
		for _, action := range models {
			dtoID := fmt.Sprintf("%d", *action.ID)
			actionDto := ExecutorActionDto{
				ID:          &dtoID,
				Operator:    action.Operator,
				Name:        action.Name,
				Description: action.Description,
				Type:        action.Type,
				CreatedAt:   action.CreatedAt,
				UpdatedAt:   action.UpdatedAt,
			}

			if action.Group != nil && len(*action.Group) > 0 {
				actionDto.Group = action.Group
			}

			if action.Inputs != nil && len(*action.Inputs) > 0 {
				inputDto := make([]ExecutorActionInputDto, 0)
				err := jsoniter.UnmarshalFromString(*action.Inputs, &inputDto)
				if err != nil {
					return nil, err
				}
				actionDto.Inputs = &inputDto
			}

			if action.Outputs != nil && len(*action.Outputs) > 0 {
				outputDto := make([]ExecutorActionOutputDto, 0)
				err := jsoniter.UnmarshalFromString(*action.Outputs, &outputDto)
				if err != nil {
					return nil, err
				}
				actionDto.Outputs = &outputDto
			}

			if action.Config != nil && len(*action.Config) > 0 {
				configDto := make(map[string]interface{})
				err := jsoniter.UnmarshalFromString(*action.Config, &configDto)
				if err != nil {
					return nil, err
				}
				actionDto.Config = &configDto
			}

			actions = append(actions, &actionDto)
		}
	}

	return actions, nil
}

func (e *ExecutorHandlerImpl) ImportAgents(ctx context.Context, userInfo *drivenadapters.UserInfo, dto *ImportAgentsDto) (result *ImportAgentsResult, err error) {

	result = &ImportAgentsResult{}
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(newCtx)

	config := common.NewConfig()

	if config.AnyData.Host == "" {
		log.Warnf("[logic.ImportAgents] AnyData host is not configured")
		return nil, ierrors.NewIError(ierrors.InternalError, "", "AnyData host is not configured")
	}

	agentIDs := dto.AgentIDs

	if len(agentIDs) == 0 {
		agents, err := e.ad.GetAgents(ctx, "", nil, nil)

		if err != nil {
			log.Warnf("[logic.ImportAgents] Get agents error: %v", err)
			return nil, ierrors.NewIError(ierrors.InternalError, "", err)
		}

		for _, agent := range agents {
			agentIDs = append(agentIDs, agent.AgentID)
		}
	}

	if len(agentIDs) == 0 {
		log.Infof("[logic.ImportAgents] No agents found")
		return result, nil
	}

	result.Total = len(agentIDs)
	existsExecutor, err := e.executorDao.GetExecutorByName(ctx, userInfo.UserID, dto.Name)

	if err != nil {
		log.Warnf("[logic.ImportAgents] Get executor by name error: %v", err)
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}

	if existsExecutor != nil {
		if dto.OnDup == 0 {
			result.Skipped = result.Total
			return result, nil
		}

		if dto.OnDup == 1 {
			var actions []*rds.ExecutorActionModel
			actions, err = e.executorDao.GetExecutorActions(ctx, *existsExecutor.ID)

			if err != nil {
				log.Warnf("[logic.ImportAgents] get executor actions error: %v", err)
				return nil, ierrors.NewIError(ierrors.InternalError, "", err)
			}

			actionMap := make(map[string]*rds.ExecutorActionModel)
			for _, action := range actions {
				actionMap[*action.Name] = action
			}

			for _, agentID := range agentIDs {
				var err1 error

				agent, err1 := e.ad.GetAgent(ctx, agentID)

				if err1 != nil {
					log.Warnf("[logic.ImportAgents] get agent failed, id: %s, err: %v", agentID, err1)
					result.Failed++
					continue
				}

				action := e.convertAgentToAction(agent)

				if action == nil {
					result.Failed++
					continue
				}
				existsAction, exists := actionMap[*action.Name]

				if exists {
					if dto.ActionOnDup == 0 {
						result.Skipped++
						continue
					}

					if dto.ActionOnDup == 1 {
						err1 = e.UpdateExecutorAction(ctx, *existsExecutor.ID, *existsAction.ID, *action, userInfo)
						if err1 != nil {
							result.Failed++
							log.Warnf("[logic.ImportAgents] logic.ImportAgents update executor action failed: action: %v, err: %v", action.ID, err1)
						}
						result.Success++
						continue
					}

					actionName := utils.StringSlice(*action.Name, 0, 240) + "_" + time.Now().Format("20060102150405")
					action.Name = &actionName
				}

				_, err1 = e.CreateExecutorAction(ctx, *existsAction.ExecutorID, *action, userInfo)
				if err1 != nil {
					result.Failed++
					log.Warnf("[logic.ImportAgents] logic.ImportAgents create executor action failed: action: %v, err: %v", action.ID, err1)
					continue
				}

				result.Success++
			}

			return
		}

		dto.Name = utils.StringSlice(dto.Name, 0, 240) + "_" + time.Now().Format("20060102150405")
		return e.ImportAgents(ctx, userInfo, dto)
	}

	executorDto := &ExecutorDto{
		Name:        &dto.Name,
		Description: &dto.Description,
		Status:      &dto.Status,
		Accessors:   dto.Accessors,
		Actions:     []*ExecutorActionDto{},
	}

	for _, agentID := range agentIDs {
		var err1 error

		agent, err1 := e.ad.GetAgent(ctx, agentID)

		if err1 != nil {
			log.Warnf("[logic.ImportAgents] get agent failed, id: %s, err: %v", agentID, err1)
			result.Failed++
			continue
		}

		action := e.convertAgentToAction(agent)

		if action != nil {
			executorDto.Actions = append(executorDto.Actions, action)
			result.Success++
		} else {
			result.Failed++
		}
	}

	_, err = e.CreateExecutor(ctx, *executorDto, userInfo)

	if err != nil {
		return nil, ierrors.NewIError(ierrors.InternalError, "", err)
	}
	return result, nil
}

const AgentActionCode = `
import requests
import base64
import json

agent_data = "{agent_data}"


def main(*args):

    agent_json = base64.b64decode(agent_data).decode('utf-8')
    agent = json.loads(agent_json)

    base_url = agent["base_url"]
    app_id = agent["app_id"]
    agent_id = agent["agent_id"]

    input_keys = agent["input_keys"]
    output_keys = agent["output_keys"]
    block_output_keys = agent["block_output_keys"]

    inputs = {
        "_options": {
            "stream": False
        }
    }
    for i in range(len(input_keys)):
        if i >= len(args):
            break

        input_key = input_keys[i]

        if input_key == "history":
            try:
                inputs[input_key] = json.loads(args[i])
            except:
                inputs[input_key] = []
        else:
            inputs[input_key] = args[i]

    url = base_url + "/api/agent-factory/v2/agent/" + agent_id
    response = requests.post(url, headers={
        "appid": app_id
    }, json=inputs, verify=False)

    if response.status_code != 200:
        err = response.json()
        raise Exception(err["Description"])

    result = response.json()["res"]
    outputs = []

    for output_key in output_keys:
        if output_key in result["answer"]:
            outputs.append(result["answer"][output_key])
        else:
            outputs.append("")

    for block_output_key in block_output_keys:
        if block_output_key in result["block_answer"]:
            outputs.append(result["block_answer"][block_output_key])
        else:
            outputs.append("")

    return tuple(outputs)`

func (e *ExecutorHandlerImpl) convertAgentToAction(agent *drivenadapters.AgentInfo) *ExecutorActionDto {
	if agent.ReleaseConfig == nil {
		return nil
	}

	inputs := make([]ExecutorActionInputDto, 0)
	outputs := make([]ExecutorActionOutputDto, 0)

	inputKeys := make([]string, 0)
	outputKeys := make([]string, 0)
	blockOutputKeys := make([]string, 0)
	releaseConfig := agent.ReleaseConfig.(map[string]interface{})

	if inputConfig, ok := releaseConfig["input"].(map[string]interface{}); ok {
		if inputFields, ok := inputConfig["fields"].([]interface{}); ok {
			for _, inputField := range inputFields {
				field := inputField.(map[string]interface{})
				fieldName := field["name"].(string)
				if fieldName == "history" {
					continue
				}
				fieldType := "string"
				required := true

				inputKeys = append(inputKeys, field["name"].(string))
				inputs = append(inputs, ExecutorActionInputDto{
					Key:      &fieldName,
					Name:     &fieldName,
					Type:     &fieldType,
					Required: &required,
				})
			}
		}
	}

	if outputConfig, ok := releaseConfig["output"].(map[string]interface{}); ok {
		if answerConfig, ok := outputConfig["answer"].([]interface{}); ok {
			for _, answerField := range answerConfig {
				field := answerField.(map[string]interface{})
				fieldName := field["name"].(string)
				fieldType := "string"
				outputKeys = append(outputKeys, fieldName)
				outputs = append(outputs, ExecutorActionOutputDto{
					Key:  &fieldName,
					Name: &fieldName,
					Type: &fieldType,
				})
			}
		}

		if blockAnswerConfig, ok := outputConfig["block_answer"].([]interface{}); ok {
			for _, answerField := range blockAnswerConfig {
				field := answerField.(map[string]interface{})
				fieldName := field["name"].(string)
				fieldType := "string"
				outputKeys = append(outputKeys, fieldName)
				outputs = append(outputs, ExecutorActionOutputDto{
					Key:  &fieldName,
					Name: &fieldName,
					Type: &fieldType,
				})
			}
		}
	}

	agent_data, _ := json.Marshal(map[string]interface{}{
		"base_url":          e.ad.GetBaseURL(),
		"app_id":            e.ad.GetAppID(),
		"agent_id":          agent.AgentID,
		"input_keys":        inputKeys,
		"output_keys":       outputKeys,
		"block_output_keys": blockOutputKeys,
	})

	agent_data_base64 := base64.StdEncoding.EncodeToString(agent_data)

	description := agent.Description

	if description == "" {
		description = agent.Name
	}

	actionType := "python"
	return &ExecutorActionDto{
		Name:        &agent.Name,
		Description: &description,
		Type:        &actionType,
		Inputs:      &inputs,
		Outputs:     &outputs,
		Config: &map[string]interface{}{
			"agent_id": agent.AgentID,
			"code":     strings.Replace(AgentActionCode, "{agent_data}", agent_data_base64, 1),
		},
	}
}
