// Package common common operator manage
// @file status_manage.go
// @description: 统一转台管理器
package common

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

var statusTransitions = map[interfaces.BizStatus][]interfaces.BizStatus{
	// 从未发布状态的转换
	interfaces.BizStatusUnpublish: {
		interfaces.BizStatusPublished,
		interfaces.BizStatusUnpublish,
	},

	// 从已发布状态的转换
	interfaces.BizStatusPublished: {
		interfaces.BizStatusOffline,
		interfaces.BizStatusEditing,
	},

	// 从已发布编辑中状态的转换
	interfaces.BizStatusEditing: {
		interfaces.BizStatusEditing,
		interfaces.BizStatusPublished,
	},

	// 从已下架状态的转换
	interfaces.BizStatusOffline: {
		interfaces.BizStatusPublished,
		interfaces.BizStatusUnpublish,
	},
}

// CheckStatusTransition 检查状态是否可以转换
func CheckStatusTransition(fromState, toState interfaces.BizStatus) bool {
	allowedTargetStates, exists := statusTransitions[fromState]
	if !exists {
		return false
	}
	for _, allowedTargetState := range allowedTargetStates {
		if allowedTargetState == toState {
			return true
		}
	}
	return false
}

// 编辑操作下允许的状态转换
var editStatusTrans = map[interfaces.BizStatus]interfaces.BizStatus{
	interfaces.BizStatusUnpublish: interfaces.BizStatusUnpublish,
	interfaces.BizStatusPublished: interfaces.BizStatusEditing,
	interfaces.BizStatusEditing:   interfaces.BizStatusEditing,
	interfaces.BizStatusOffline:   interfaces.BizStatusUnpublish,
}

var deletableStatus = map[interfaces.BizStatus]bool{
	interfaces.BizStatusUnpublish: true,
	interfaces.BizStatusOffline:   true,
}

// GetEditStatusTrans 获取编辑操作下允许的状态转换
func GetEditStatusTrans(ctx context.Context, fromState interfaces.BizStatus) (interfaces.BizStatus, error) {
	targetState, exists := editStatusTrans[fromState]
	if !exists {
		return "", errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtMCPUnSupportEdit, "current mcp does not support editing")
	}
	return targetState, nil
}

// CanDelete 判断当前业务状态是否允许进入删除流程
func CanDelete(fromState interfaces.BizStatus) bool {
	return deletableStatus[fromState]
}
