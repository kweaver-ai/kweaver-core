// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

// BKNService BKN 导入导出服务接口
//
//go:generate mockgen -source ../interfaces/bkn_service.go -destination ../interfaces/mock/mock_bkn_service.go
type BKNService interface {
	// ExportToTar 将知识网络导出为 tar 包
	ExportToTar(ctx context.Context, knID string, branch string) ([]byte, error)
}
