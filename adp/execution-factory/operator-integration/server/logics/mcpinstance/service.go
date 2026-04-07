package mcpinstance

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

var (
	serviceOnce sync.Once
	service     interfaces.InstanceService
)

// instanceService 实现 MCP 实例的生命周期管理（对外暴露 interfaces.InstanceService）：
// - 期望态：runtime config 持久化到 t_resource_deploy
// - 运行态：实例由 InstancePool 按需创建并在内存中管理
// - 构建/卸载：由 InstanceManager 负责把 runtime config 组装为可服务的 instance
type instanceService struct {
	logger           interfaces.Logger
	dbTx             model.DBTx
	dbResourceDeploy model.DBResourceDeploy
	instancePool     *InstancePool
	// configStore      MCPConfigStore
}

// NewMCPInstanceService 新建 MCP 实例服务
func NewMCPInstanceService(executor interfaces.IMCPToolExecutor) interfaces.InstanceService {
	serviceOnce.Do(func() {
		service = &instanceService{
			logger:           config.NewConfigLoader().GetLogger(),
			dbTx:             dbaccess.NewBaseTx(),
			dbResourceDeploy: dbaccess.NewResourceDeployDBSingleton(),
			instancePool:     initInstancePool(executor),
		}
	})
	return service
}

// CreateMCPInstance 创建 MCPInstance
func (s *instanceService) CreateMCPInstance(ctx context.Context, req *interfaces.MCPInstanceCreateRequest) (*interfaces.MCPInstanceCreateResponse, error) {
	exists, err := s.dbResourceDeploy.Exists(ctx, req.MCPID, req.Version)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, infraerrors.NewHTTPError(ctx, http.StatusBadRequest, infraerrors.ErrExtMCPInstanceAlreadyExists, nil)
	}
	tx, err := s.dbTx.GetTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if tx != nil {
				_ = tx.Rollback()
			}
		} else {
			if tx != nil {
				_ = tx.Commit()
			}
		}
	}()

	cfg := buildRuntimeConfig(req)
	_, err = s.dbResourceDeploy.Insert(ctx, tx, &model.ResourceDeployDB{
		ResourceID:  req.MCPID,
		Type:        interfaces.ResourceDeployTypeMCP.String(),
		Version:     req.Version,
		Name:        req.Name,
		Description: req.Instructions,
		Config:      utils.ObjectToJSON(cfg),
	})
	if err != nil {
		return nil, err
	}

	// 创建实例
	instance, err := s.instancePool.GetOrCreateWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &interfaces.MCPInstanceCreateResponse{
		MCPID:     req.MCPID,
		Version:   req.Version,
		StreamURL: instance.StreamRoutePath,
		SSEURL:    instance.SSERoutePath,
	}, nil
}

func (s *instanceService) UpdateMCPInstance(ctx context.Context, mcpID string, version int, req *interfaces.MCPInstanceUpdateRequest) (*interfaces.MCPInstanceUpdateResponse, error) {
	tx, err := s.dbTx.GetTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if tx != nil {
				_ = tx.Rollback()
			}
		} else {
			if tx != nil {
				_ = tx.Commit()
			}
		}
	}()

	cfg := &interfaces.MCPRuntimeConfig{
		MCPID:        mcpID,
		Version:      version,
		Name:         req.MCPServerName,
		Instructions: req.Instructions,
		Tools:        convertToolConfigs(req.ToolConfigs),
	}

	err = s.dbResourceDeploy.Update(ctx, tx, &model.ResourceDeployDB{
		ResourceID:  mcpID,
		Type:        interfaces.ResourceDeployTypeMCP.String(),
		Version:     version,
		Name:        req.MCPServerName,
		Description: req.Instructions,
		Config:      utils.ObjectToJSON(cfg),
	})
	if err != nil {
		return nil, err
	}

	_ = s.instancePool.DeleteInstance(ctx, mcpID, version)
	instance, err := s.instancePool.GetOrCreateWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &interfaces.MCPInstanceUpdateResponse{
		MCPID:      mcpID,
		MCPVersion: version,
		StreamURL:  instance.StreamRoutePath,
		SSEURL:     instance.SSERoutePath,
	}, nil
}

func (s *instanceService) DeleteMCPInstance(ctx context.Context, mcpID string, version int) (err error) {
	tx, err := s.dbTx.GetTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if tx != nil {
				_ = tx.Rollback()
			}
		} else {
			if tx != nil {
				_ = tx.Commit()
			}
		}
	}()

	if err = s.dbResourceDeploy.Delete(ctx, tx, version, interfaces.ResourceDeployTypeMCP.String(), mcpID); err != nil {
		return err
	}
	return s.instancePool.DeleteInstance(ctx, mcpID, version)
}

func (s *instanceService) DeleteAllMCPInstances(ctx context.Context, mcpID string) error {
	resourceDeploys, err := s.dbResourceDeploy.SelectListByResourceID(ctx, mcpID)
	if err != nil {
		return err
	}
	for _, rd := range resourceDeploys {
		if err := s.DeleteMCPInstance(ctx, mcpID, rd.Version); err != nil {
			return err
		}
	}
	return nil
}

func (s *instanceService) UpgradeMCPInstance(ctx context.Context, req *interfaces.MCPInstanceCreateRequest) (*interfaces.MCPInstanceCreateResponse, error) {
	exists, err := s.dbResourceDeploy.Exists(ctx, req.MCPID, req.Version)
	if err != nil {
		return nil, err
	}
	if exists {
		resp, err := s.UpdateMCPInstance(ctx, req.MCPID, req.Version, &interfaces.MCPInstanceUpdateRequest{
			MCPServerName: req.Name,
			Instructions:  req.Instructions,
			ToolConfigs:   req.ToolConfigs,
		})
		if err != nil {
			return nil, err
		}
		return &interfaces.MCPInstanceCreateResponse{
			MCPID:     req.MCPID,
			Version:   req.Version,
			StreamURL: resp.StreamURL,
			SSEURL:    resp.SSEURL,
		}, nil
	}
	return s.CreateMCPInstance(ctx, req)
}

// GetMCPInstance 获取MCP实例并初始化
func (s *instanceService) GetMCPInstance(ctx context.Context, mcpID string, version int) (*interfaces.MCPServerInstance, error) {
	ins, err := s.instancePool.GetOrCreate(ctx, mcpID, version)
	if err == nil {
		return ins, nil
	}
	if errors.Is(err, ErrMCPInstanceConfigNotFound) {
		return nil, infraerrors.NewHTTPError(ctx, http.StatusNotFound, infraerrors.ErrExtMCPInstanceNotFound, nil)
	}
	return nil, err
}

func buildRuntimeConfig(req *interfaces.MCPInstanceCreateRequest) *interfaces.MCPRuntimeConfig {
	return &interfaces.MCPRuntimeConfig{
		MCPID:        req.MCPID,
		Version:      req.Version,
		Name:         req.Name,
		Instructions: req.Instructions,
		Tools:        convertToolConfigs(req.ToolConfigs),
	}
}

func convertToolConfigs(toolConfigs []*interfaces.MCPToolConfig) []*interfaces.MCPToolDeployConfig {
	out := make([]*interfaces.MCPToolDeployConfig, 0, len(toolConfigs))
	for _, t := range toolConfigs {
		out = append(out, &interfaces.MCPToolDeployConfig{
			ToolID:      t.ToolID,
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}
	return out
}
