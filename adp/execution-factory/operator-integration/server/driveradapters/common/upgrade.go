// Package common 公共模块操作接口
package common

import (
	"context"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UpgradeHandler interface {
	MigrateHistoryData(c *gin.Context)
}

var (
	upgradeHandlerOnce sync.Once
	upgradeH           UpgradeHandler
)

type upgradeHandler struct {
	Logger            interfaces.Logger
	DBMCPServerConfig model.DBMCPServerConfig
	ToolBoxDB         model.IToolboxDB
	DBOperatorManager model.IOperatorRegisterDB
}

type MigrateHistoryDataRequest struct {
	ResourceType interfaces.AuthResourceType `form:"resource_type"`    // 资源类型
	Page         int                         `form:"page" default:"0"` // 页码
	PageSize     int                         `form:"page_size"`
}

type HistoryData struct {
	Id string `json:"id"` // 历史数据ID
}

type MigrateHistoryDataResponse struct {
	Total int64          `json:"total" default:"0"`
	Items []*HistoryData `json:"items"` // 历史数据列表
}

// NewUpgradeHandler 升级操作接口
func NewUpgradeHandler() UpgradeHandler {
	upgradeHandlerOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		upgradeH = &upgradeHandler{
			Logger:            confLoader.GetLogger(),
			DBMCPServerConfig: dbaccess.NewMCPServerConfigDBSingleton(),
			ToolBoxDB:         dbaccess.NewToolboxDB(),
			DBOperatorManager: dbaccess.NewOperatorManagerDB(),
		}
	})
	return upgradeH
}

// MigrateHistoryData 迁移历史数据
// 此接口仅在从旧版本升级到5.0.0.3版本时使用，用于迁移历史数据
func (uh *upgradeHandler) MigrateHistoryData(c *gin.Context) {
	var err error
	req := &MigrateHistoryDataRequest{}

	ctx := c.Request.Context()

	if err = c.ShouldBindQuery(req); err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	if err = defaults.Set(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	var resp *MigrateHistoryDataResponse
	switch req.ResourceType {
	case interfaces.AuthResourceTypeOperator:
		resp, err = uh.migrateHistoryDataForOperator(ctx, req)
	case interfaces.AuthResourceTypeMCP:
		resp, err = uh.migrateHistoryDataForForMcp(ctx, req)
	case interfaces.AuthResourceTypeToolBox:
		resp, err = uh.migrateHistoryDataForToolBox(ctx, req)
	default:
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "resource_type is invalid")
		rest.ReplyError(c, err)
		return
	}

	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	rest.ReplyOK(c, http.StatusOK, resp)
}

func (uh *upgradeHandler) migrateHistoryDataForOperator(ctx context.Context, req *MigrateHistoryDataRequest) (resp *MigrateHistoryDataResponse, err error) {
	resp = &MigrateHistoryDataResponse{
		Items: []*HistoryData{},
	}
	filter := make(map[string]interface{})

	var total int64
	total, err = uh.DBOperatorManager.CountByWhereClause(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp.Total = total

	// 计算实际的offset
	actualOffset := int64(req.Page * req.PageSize)

	// 如果offset超过total，直接返回空items
	if actualOffset >= total {
		return resp, nil
	}

	filter["limit"] = req.PageSize
	filter["offset"] = req.Page
	configList, err := uh.DBOperatorManager.SelectListPage(ctx, filter, nil, nil)
	if err != nil {
		return nil, err
	}

	for _, config := range configList {
		resp.Items = append(resp.Items, &HistoryData{Id: config.OperatorID})
	}

	return resp, nil
}

func (uh *upgradeHandler) migrateHistoryDataForToolBox(ctx context.Context, req *MigrateHistoryDataRequest) (resp *MigrateHistoryDataResponse, err error) {
	resp = &MigrateHistoryDataResponse{
		Items: []*HistoryData{},
	}
	filter := make(map[string]interface{})

	var total int64
	total, err = uh.ToolBoxDB.CountToolBox(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp.Total = total

	// 计算实际的offset
	actualOffset := int64(req.Page * req.PageSize)

	// 如果offset超过total，直接返回空items
	if actualOffset >= total {
		return resp, nil
	}

	filter["limit"] = req.PageSize
	filter["offset"] = req.Page
	configList, err := uh.ToolBoxDB.SelectToolBoxList(ctx, filter, nil, nil)
	if err != nil {
		return nil, err
	}

	for _, config := range configList {
		resp.Items = append(resp.Items, &HistoryData{Id: config.BoxID})
	}

	return resp, nil
}

func (uh *upgradeHandler) migrateHistoryDataForForMcp(ctx context.Context, req *MigrateHistoryDataRequest) (resp *MigrateHistoryDataResponse, err error) {
	resp = &MigrateHistoryDataResponse{
		Items: []*HistoryData{},
	}
	filter := make(map[string]interface{})
	var total int64
	total, err = uh.DBMCPServerConfig.CountByWhereClause(ctx, nil, filter)
	if err != nil {
		return nil, err
	}

	resp.Total = total

	// 计算实际的offset
	actualOffset := int64(req.Page * req.PageSize)

	// 如果offset超过total，直接返回空items
	if actualOffset >= total {
		return
	}

	filter["limit"] = req.PageSize
	filter["offset"] = req.Page
	configList, err := uh.DBMCPServerConfig.SelectListPage(ctx, nil, filter, nil, nil)
	if err != nil {
		return nil, err
	}

	for _, config := range configList {
		resp.Items = append(resp.Items, &HistoryData{Id: config.MCPID})
	}
	return resp, nil
}
