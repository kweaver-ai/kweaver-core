package iuniqueryhttp

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/uniqueryaccess/uniquerydto"
)

type IUniquery interface {
	// GetDataViews 根据视图ID查询数据
	// 请求单个数据试图
	GetDataView(ctx context.Context, viewID string, reqData uniquerydto.ReqDataView) (uniquerydto.ViewResults, error)
}
