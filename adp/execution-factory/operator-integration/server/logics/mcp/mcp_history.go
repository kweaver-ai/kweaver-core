package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

const (
	// MaxMCPHistoryRecords 定义MCP服务器历史记录的最大保留数量
	MaxMCPHistoryRecords = 10
)

func (s *mcpServiceImpl) addMCPHistory(ctx context.Context, tx *sql.Tx, mcpRelease *model.MCPServerReleaseDB, userID string) error {
	now := time.Now().UnixNano()
	history := &model.MCPServerReleaseHistoryDB{
		CreateUser:  userID,
		CreateTime:  now,
		UpdateUser:  userID,
		UpdateTime:  now,
		MCPID:       mcpRelease.MCPID,
		MCPRelease:  utils.ObjectToJSON(mcpRelease),
		Version:     mcpRelease.Version,
		ReleaseDesc: mcpRelease.ReleaseDesc,
	}

	// 查询现有历史记录
	histories, err := s.DBMCPServerReleaseHistory.SelectByMCPID(ctx, tx, mcpRelease.MCPID)
	if err != nil {
		s.logger.WithContext(ctx).Warnf("failed to query existing MCP history records: %v", err)
		return fmt.Errorf("failed to query existing MCP history records: %w", err)
	}

	// 如果历史记录数量达到上限，删除多余的记录（保持FIFO策略）
	// 注意：SelectByMCPID返回的是降序排序，最新记录在前，最旧记录在后
	if len(histories) >= MaxMCPHistoryRecords {
		// 计算需要删除的记录数量，确保插入新记录后不超过上限
		recordsToDelete := len(histories) - MaxMCPHistoryRecords + 1
		// 从末尾开始删除最旧的记录
		startIndex := len(histories) - recordsToDelete
		for i := startIndex; i < len(histories); i++ {
			if err = s.DBMCPServerReleaseHistory.DeleteByID(ctx, tx, histories[i].ID); err != nil {
				s.logger.WithContext(ctx).Warnf("failed to delete old MCP history record: %v", err)
				return fmt.Errorf("failed to delete old MCP history record: %w", err)
			}

			// 移除MCP 工具配置信息
			mcpReleaseHistory := utils.JSONToObject[model.MCPServerReleaseDB](histories[i].MCPRelease)
			if mcpReleaseHistory.CreationType == interfaces.MCPCreationTypeToolImported.String() {
				err = s.removeMCPTools(ctx, tx, mcpReleaseHistory.MCPID, mcpReleaseHistory.Version)
				if err != nil {
					s.logger.WithContext(ctx).Warnf("failed to remove MCP tools: %v", err)
					return fmt.Errorf("failed to remove MCP tools: %w", err)
				}
			}
		}
	}

	var lastMCPReleaseHistory *model.MCPServerReleaseHistoryDB
	if len(histories) > 0 {
		lastMCPReleaseHistory = histories[0]
	}

	// 删除最近一条是历史MCP Server实例
	if lastMCPReleaseHistory != nil {
		lastMCPRelease := utils.JSONToObject[model.MCPServerReleaseDB](lastMCPReleaseHistory.MCPRelease)
		if lastMCPRelease.CreationType == interfaces.MCPCreationTypeToolImported.String() {
			// 删除mcp Server实例
			err = s.MCPInstanceService.DeleteMCPInstance(ctx, lastMCPRelease.MCPID, lastMCPRelease.Version)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("failed to remove MCP server instance: %v", err)
				return fmt.Errorf("failed to remove MCP server instance: %w", err)
			}
		}
	}

	if lastMCPReleaseHistory != nil {
		if lastMCPReleaseHistory.Version > mcpRelease.Version {
			return nil
		}
		if lastMCPReleaseHistory.Version == mcpRelease.Version {
			// 删除当前版本发布历史记录
			err = s.DBMCPServerReleaseHistory.DeleteByMCPIDAndVersion(ctx, tx, history.MCPID, history.Version)
			if err != nil {
				return err
			}
		}
	}
	// 插入新的历史记录
	if _, err := s.DBMCPServerReleaseHistory.Insert(ctx, tx, history); err != nil {
		s.logger.WithContext(ctx).Warnf("failed to insert new MCP history record: %v", err)
		return fmt.Errorf("failed to insert new MCP history record: %w", err)
	}
	return nil
}
