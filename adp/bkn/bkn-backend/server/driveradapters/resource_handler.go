// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"bkn-backend/interfaces"
)

const (
	RESOURCES_KEYWOED    = "keyword"
	RESOURCES_PAGE_LIMIT = "50"
)

// 分页获取指标模型资源列表
func (r *restHandler) ListResources(c *gin.Context) {
	logger.Debug("ListResources Start")

	// 获取分页参数
	resourceType := c.Query("resource_type")
	switch resourceType {
	case interfaces.RESOURCE_TYPE_KN:
		r.ListKnSrcs(c)
	default:
		// httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusNotFound,
		// 	derrors.DataModel_MetricModel_MetricTaskNotFound)

		// // 设置 trace 的错误信息的 attributes
		// rest.ReplyError(c, httpErr)
	}

}
