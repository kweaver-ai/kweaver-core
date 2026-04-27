package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	icommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// RegisterBuiltinMCPServer 注册内置MCP服务
func (s *mcpServiceImpl) RegisterBuiltinMCPServer(ctx context.Context, req *interfaces.MCPBuiltinRegisterRequest) (resp *interfaces.MCPBuiltinRegisterResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	if req.UserID == "" {
		req.UserID = interfaces.SystemUser
	}
	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get tx failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, "get tx failed")
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	// check 内置组件
	check := &interfaces.IntCompConfig{
		ComponentID:   req.MCPID,
		ComponentType: interfaces.ComponentTypeMCP,
		ConfigVersion: req.ConfigVersion,
		ConfigSource:  req.ConfigSource,
		ProtectedFlag: req.ProtectedFlag,
	}

	action, err := s.IntCompConfigService.CompareConfig(ctx, check)
	if err != nil {
		return
	}

	if action == interfaces.IntCompConfigActionTypeSkip {
		// 跳过
		resp = &interfaces.MCPBuiltinRegisterResponse{
			MCPID:  req.MCPID,
			Status: interfaces.BizStatusPublished,
		}
		return
	}
	// 检查mcp是否存在，存在就新增，不存在就更新
	var config *model.MCPServerConfigDB
	config, err = s.DBMCPServerConfig.SelectByID(ctx, tx, req.MCPID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("select mcp config failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("select mcp config failed, err: %v", err))
		return
	}
	if config == nil {
		// 新增
		mcpConfig := s.mcpBuiltinRegisterToModel(req)
		mcpConfig.CreateUser = req.UserID
		mcpConfig.UpdateUser = req.UserID
		_, err = s.addMCPConfig(ctx, tx, mcpConfig)
		if err != nil {
			return
		}
		// 创建内置组件权限策略
		err = s.AuthService.CreateIntCompPolicyForAllUsers(ctx, &interfaces.AuthResource{
			ID:   req.MCPID,
			Type: interfaces.AuthResourceTypeMCP.String(),
			Name: req.Name,
		})
		if err != nil {
			return
		}
	} else {
		// 更新
		// 下架内置MCP服务
		err = s.unpublishBuiltinMCPServer(ctx, tx, req.MCPID, req.UserID)
		if err != nil {
			return
		}
		// 更新MCP服务
		mcpConfig := s.mcpBuiltinRegisterToModel(req)
		mcpConfig.UpdateUser = req.UserID
		_, _, _, err = s.updateMCPConfig(ctx, tx, mcpConfig)
		if err != nil {
			return
		}
	}

	// 发布MCP服务
	_, err = s.publishBuiltinMCPServer(ctx, tx, req.MCPID, req.UserID)
	if err != nil {
		return
	}
	// 更新内置配置表
	err = s.IntCompConfigService.UpdateConfig(ctx, tx, check)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("update config failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("update config failed, err: %v", err))
		return
	}
	resp = &interfaces.MCPBuiltinRegisterResponse{
		MCPID:  req.MCPID,
		Status: interfaces.BizStatusPublished,
	}
	return
}

// UnregisterBuiltinMCPServer 注销内置MCP服务
func (s *mcpServiceImpl) UnregisterBuiltinMCPServer(ctx context.Context, req *interfaces.MCPBuiltinUnregisterRequest) (err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 校验用户ID
	if req.UserID == "" {
		req.UserID = interfaces.SystemUser
	}
	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get tx failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, "get tx failed")
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// 校验，如果是非内置MCP，不允许进行该操作
	config, err := s.DBMCPServerConfig.SelectByID(ctx, tx, req.MCPID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("select mcp config failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("select mcp config failed, err: %v", err))
		return
	}
	if config == nil || !config.IsInternal {
		err = infraerrors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("mcp_id: %s is not a builtin MCP, unregistration through this method is not allowed", req.MCPID))
		return
	}

	// 下架内置MCP服务
	err = s.unpublishBuiltinMCPServer(ctx, tx, req.MCPID, req.UserID)
	if err != nil {
		return
	}
	// 删除内置MCP服务
	err = s.deleteBuiltinMCPServer(ctx, tx, req.MCPID)
	if err != nil {
		return
	}
	// 删除内置配置表
	err = s.IntCompConfigService.DeleteConfig(ctx, tx, interfaces.ComponentTypeMCP.String(), req.MCPID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("delete config failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("delete config failed, err: %v", err))
		return
	}

	// 取消关联业务域
	businessDomainId, _ := icommon.GetBusinessDomainFromCtx(ctx)
	err = s.BusinessDomainService.DisassociateResource(ctx, businessDomainId, req.MCPID, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}

	// 触发权限策略删除
	err = s.AuthService.DeletePolicy(ctx, []string{req.MCPID}, interfaces.AuthResourceTypeMCP)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("delete policy failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("delete policy failed, err: %v", err))
	}
	return
}

// unpublishBuiltinMCPServer 下架内置MCP服务
func (s *mcpServiceImpl) unpublishBuiltinMCPServer(ctx context.Context, tx *sql.Tx, mcpID, userID string) (err error) {
	_, _, err = s.modifyMCPStatus(ctx, tx, &interfaces.UpdateMCPStatusRequest{
		MCPID:  mcpID,
		Status: interfaces.BizStatusOffline,
		UserID: userID,
	})
	return
}

// publishBuiltinMCPServer 发布内置MCP服务
func (s *mcpServiceImpl) publishBuiltinMCPServer(ctx context.Context, tx *sql.Tx, mcpID, userID string) (*interfaces.UpdateMCPStatusResponse, error) {
	_, result, err := s.modifyMCPStatus(ctx, tx, &interfaces.UpdateMCPStatusRequest{
		MCPID:  mcpID,
		Status: interfaces.BizStatusPublished,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// deleteBuiltinMCPServer 删除内置MCP服务
func (s *mcpServiceImpl) deleteBuiltinMCPServer(ctx context.Context, tx *sql.Tx, mcpID string) error {
	// 查询是否是内置MCP服务
	config, err := s.DBMCPServerConfig.SelectByID(ctx, tx, mcpID)
	if err != nil {
		return err
	}
	if config == nil {
		return fmt.Errorf("mcp not found")
	}
	// 删除MCP Server配置
	_, err = s.removeMCPConfig(ctx, tx, mcpID)
	if err != nil {
		return err
	}
	return nil
}

// mcpBuiltinRegisterToModel 内置MCP服务注册请求转换为MCP Server配置表
func (s *mcpServiceImpl) mcpBuiltinRegisterToModel(req *interfaces.MCPBuiltinRegisterRequest) *model.MCPServerConfigDB {
	return &model.MCPServerConfigDB{
		MCPID:       req.MCPID,
		Name:        req.Name,
		Description: req.Description,
		Mode:        string(req.Mode),
		URL:         req.URL,
		Headers:     utils.ObjectToJSON(req.Headers),
		Command:     req.Command,
		Env:         utils.ObjectToJSON(req.Env),
		Args:        utils.ObjectToJSON(req.Args),
		Status:      string(interfaces.BizStatusUnpublish),
		Category:    string(interfaces.CategoryTypeSystem),
		Source:      req.Source,
		IsInternal:  true,
	}
}
