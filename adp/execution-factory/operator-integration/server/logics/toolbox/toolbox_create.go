package toolbox

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// CreateToolBox 工具箱管理
func (s *ToolServiceImpl) CreateToolBox(ctx context.Context, req *interfaces.CreateToolBoxReq) (resp *interfaces.CreateToolBoxResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 检查新建权限
	var accessor *interfaces.AuthAccessor
	accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckCreatePermission(ctx, accessor, interfaces.AuthResourceTypeToolBox)
	if err != nil {
		return
	}
	// 1. 参数解析及校验
	metadatas, err := s.parseAndInitDefaultValues(ctx, req)
	if err != nil {
		return
	}
	// 2. 校验工具箱名称是否存在
	err = s.checkBoxDuplicateName(ctx, req.BoxName, "")
	if err != nil {
		return
	}
	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// 添加工具箱
	toolBox := &model.ToolboxDB{
		Name:         req.BoxName,
		Description:  req.BoxDesc,
		Status:       interfaces.BizStatusUnpublish.String(),
		Source:       req.Source,
		Category:     string(req.Category),
		ServerURL:    req.BoxSvcURL,
		CreateUser:   req.UserID,
		CreateTime:   time.Now().UnixNano(),
		UpdateUser:   req.UserID,
		UpdateTime:   time.Now().UnixNano(),
		MetadataType: string(req.MetadataType),
	}
	var boxID string
	boxID, err = s.ToolBoxDB.InsertToolBox(ctx, tx, toolBox)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("insert toolbox failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 检查是否存在元数据变更
	var detils []metric.AuditLogToolDetil
	if len(metadatas) > 0 {
		var tools []*model.ToolDB
		tools, _, _, err = s.parseOpenAPIToMetadata(ctx, boxID, req.UserID, metadatas, false)
		if err != nil {
			return
		}
		for i, tool := range tools {
			// 添加元数据
			tool.SourceID, err = s.MetadataService.RegisterMetadata(ctx, tx, metadatas[i])
			if err != nil {
				s.Logger.WithContext(ctx).Errorf("register metadata failed, err: %v", err)
				err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
				return
			}
			// 添加工具
			var toolID string
			toolID, err = s.ToolDB.InsertTool(ctx, tx, tool)
			if err != nil {
				s.Logger.WithContext(ctx).Errorf("insert tool failed, err: %v", err)
				err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
				return
			}
			detils = append(detils, metric.AuditLogToolDetil{
				ToolID:   toolID,
				ToolName: tool.Name,
			})
		}
	}
	// 关联业务域
	err = s.BusinessDomainService.AssociateResource(ctx, req.BusinessDomainID, boxID, interfaces.AuthResourceTypeToolBox)
	if err != nil {
		return
	}
	// 触发新建策略，创建人默认拥有对当前资源的所有操作权限
	err = s.AuthService.CreateOwnerPolicy(ctx, accessor, &interfaces.AuthResource{
		ID:   boxID,
		Type: string(interfaces.AuthResourceTypeToolBox),
		Name: req.BoxName,
	})
	if err != nil {
		return
	}
	// 记录审计日志
	go func() {
		accountAuthContext, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.Logger.WithContext(ctx).Warnf("[CreateToolBox] GetAccountAuthContextFromCtx err :%v", err)
			return
		}
		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationCreate,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectTool,
				Name: toolBox.Name,
				ID:   toolBox.BoxID,
			},
			Detils: &metric.AuditLogToolDetils{
				Infos:         detils,
				OperationCode: metric.AddTool,
			},
		})
	}()
	resp = &interfaces.CreateToolBoxResp{
		BoxID: boxID,
	}
	return
}

// 解析并初始化默认值
func (s *ToolServiceImpl) parseAndInitDefaultValues(ctx context.Context, req *interfaces.CreateToolBoxReq) (metadatas []interfaces.IMetadataDB, err error) {
	switch req.MetadataType {
	case interfaces.MetadataTypeAPI:
		if req.OpenAPIInput != nil && req.OpenAPIInput.Data != nil {
			// 解析API数据
			var rawContent any
			rawContent, err = s.MetadataService.ParseRawContent(ctx, req.MetadataType, req.OpenAPIInput)
			if err != nil {
				s.Logger.WithContext(ctx).Infof("parse openapi failed, err: %v", err)
				return
			}
			var content *interfaces.OpenAPIContent
			content, ok := rawContent.(*interfaces.OpenAPIContent)
			if !ok {
				err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "openapi content type error")
				s.Logger.WithContext(ctx).Infof("parse openapi failed, err: %v", err)
				return
			}
			if req.BoxName == "" {
				req.BoxName = content.Info.Title
			}
			if req.BoxDesc == "" {
				req.BoxDesc = content.Info.Description
			}
			if req.BoxSvcURL == "" {
				req.BoxSvcURL = content.SererURL
			}
			// 解析元数据
			metadatas, err = s.MetadataService.ParseMetadata(ctx, req.MetadataType, req.OpenAPIInput)
			if err != nil {
				s.Logger.WithContext(ctx).Infof("parse openapi failed, err: %v", err)
				return
			}
		}
		err = s.Validator.ValidatorURL(ctx, req.BoxSvcURL)
		if err != nil {
			return
		}
	case interfaces.MetadataTypeFunc:
		req.BoxSvcURL = interfaces.AOIServerURL
	default:
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("unsupported metadata type: %s", req.MetadataType))
		return
	}
	// 当描述为空时，默认使用名称
	if req.BoxDesc == "" {
		req.BoxDesc = req.BoxName
	}
	err = s.Validator.ValidatorToolBoxName(ctx, req.BoxName)
	if err != nil {
		return
	}
	err = s.Validator.ValidatorToolBoxDesc(ctx, req.BoxDesc)
	return
}

// 从元数据中提取工具信息
func (s *ToolServiceImpl) parseOpenAPIToMetadata(ctx context.Context, boxID, userID string,
	metadatas []interfaces.IMetadataDB, isInternalTool bool) (tools []*model.ToolDB, validatorNameMap, validatorMethodPathMap map[string]bool, err error) {
	// 检查工具是否重名
	validatorMethodPathMap = make(map[string]bool)
	validatorNameMap = make(map[string]bool)
	for _, metadata := range metadatas {
		// 检查工具名称
		err = s.Validator.ValidatorToolName(ctx, metadata.GetSummary())
		if err != nil {
			return
		}
		var useRule string
		description := metadata.GetDescription()
		// 如果是内置工具，desc超出限制将内容注入到useRule,desc使用summary填充
		err = s.Validator.ValidatorToolDesc(ctx, description)
		if err != nil {
			if !isInternalTool {
				return
			}
			err = nil
			useRule = description
			description = metadata.GetSummary()
		}
		// 工具名称是否重复
		if validatorNameMap[metadata.GetSummary()] {
			err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolExists,
				fmt.Sprintf("tool name %s duplicate", metadata.GetSummary()), metadata.GetSummary())
			return
		}
		validatorNameMap[metadata.GetSummary()] = true
		// 检查工具路径是否重复是否存在
		if metadata.GetType() == string(interfaces.MetadataTypeAPI) {
			val := validatorMethodPath(metadata.GetMethod(), metadata.GetPath())
			if validatorMethodPathMap[val] { // 重复
				err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolExists,
					fmt.Sprintf("tool info %s duplicate", val), val)
				return
			}
			validatorMethodPathMap[val] = true
		}
		// 添加基础信息
		metadata.SetCreateInfo(userID)
		metadata.SetUpdateInfo(userID)
		if metadata.GetVersion() == "" {
			metadata.SetVersion(uuid.New().String())
		}
		tools = append(tools, &model.ToolDB{
			BoxID:       boxID,
			Name:        metadata.GetSummary(),
			Description: description,
			UseRule:     useRule,
			SourceID:    metadata.GetVersion(),
			SourceType:  model.SourceType(metadata.GetType()),
			Status:      string(interfaces.ToolStatusTypeDisabled),
			CreateUser:  userID,
			UpdateUser:  userID,
		})
	}
	return
}
