// Package common 公共模块操作接口
package common

import (
	"context"
	"net/http"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

type UpgradeHandler interface {
	MigrateHistoryData(c *gin.Context)
	UpgradeSkillV070(c *gin.Context)
}

var (
	upgradeHandlerOnce sync.Once
	upgradeH           UpgradeHandler
)

type upgradeHandler struct {
	Logger                  interfaces.Logger
	DBMCPServerConfig       model.DBMCPServerConfig
	ToolBoxDB               model.IToolboxDB
	DBOperatorManager       model.IOperatorRegisterDB
	SkillRepo               model.ISkillRepository
	SkillReleaseRepo        model.ISkillReleaseDB
	SkillReleaseHistoryRepo model.ISkillReleaseHistoryDB
	DBTx                    model.DBTx
}

type MigrateHistoryDataRequest struct {
	ResourceType   interfaces.AuthResourceType `form:"resource_type"`    // 资源类型
	Page           int                         `form:"page" default:"0"` // 页码
	PageSize       int                         `form:"page_size"`
	ALL            bool                        `form:"all" default:"false"` // 是否迁移所有历史数据
	CurrentVersion string                      `form:"current_version"`     // 当前版本
	TargetVersion  string                      `form:"target_version"`      // 目标版本
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
			Logger:                  confLoader.GetLogger(),
			DBMCPServerConfig:       dbaccess.NewMCPServerConfigDBSingleton(),
			ToolBoxDB:               dbaccess.NewToolboxDB(),
			DBOperatorManager:       dbaccess.NewOperatorManagerDB(),
			SkillRepo:               dbaccess.NewSkillRepositoryDB(),
			SkillReleaseRepo:        dbaccess.NewSkillReleaseDB(),
			SkillReleaseHistoryRepo: dbaccess.NewSkillReleaseHistoryDB(),
			DBTx:                    dbaccess.NewBaseTx(),
		}
	})
	return upgradeH
}

// UpgradeSkillV060 升级技能V0.6.0 -> V0.7.0
func (uh *upgradeHandler) UpgradeSkillV070(c *gin.Context) {
	var err error
	req := &MigrateHistoryDataRequest{}
	ctx := c.Request.Context()

	if err = c.ShouldBindWith(req, binding.Form); err != nil {
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
	if req.ResourceType != interfaces.AuthResourceTypeSkill {
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "resource_type not supported")
		rest.ReplyError(c, err)
		return
	}
	resp, err := uh.migrateHistoryDataForSkill(ctx, req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
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

func (uh *upgradeHandler) migrateHistoryDataForSkill(ctx context.Context, req *MigrateHistoryDataRequest) (resp *MigrateHistoryDataResponse, err error) {
	if err = validateSkillUpgradeVersion(ctx, req.CurrentVersion, req.TargetVersion); err != nil {
		return nil, err
	}
	resp = &MigrateHistoryDataResponse{
		Items: []*HistoryData{},
	}
	filter := map[string]interface{}{
		"status": interfaces.BizStatusPublished.String(),
	}

	total, err := uh.SkillRepo.CountByWhereClause(ctx, nil, filter)
	if err != nil {
		return nil, err
	}
	resp.Total = total

	if req.ALL {
		if total == 0 {
			return resp, nil
		}
		filter["all"] = req.ALL
		skills, err := uh.SkillRepo.SelectSkillListPage(ctx, nil, filter, nil, nil)
		if err != nil {
			return nil, err
		}
		for _, skill := range skills {
			if err = uh.migrateSkillReleaseData(ctx, skill); err != nil {
				return nil, err
			}
			resp.Items = append(resp.Items, &HistoryData{Id: skill.SkillID})
		}
		return resp, nil
	}

	actualOffset := int64(req.Page * req.PageSize)
	if actualOffset >= total {
		return resp, nil
	}

	filter["limit"] = req.PageSize
	filter["offset"] = req.Page
	skills, err := uh.SkillRepo.SelectSkillListPage(ctx, nil, filter, nil, nil)
	if err != nil {
		return nil, err
	}

	for _, skill := range skills {
		if err = uh.migrateSkillReleaseData(ctx, skill); err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, &HistoryData{Id: skill.SkillID})
	}
	return resp, nil
}

func validateSkillUpgradeVersion(ctx context.Context, currentVersion, targetVersion string) error {
	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		return errors.DefaultHTTPError(ctx, http.StatusBadRequest, "current_version is invalid")
	}
	target, err := semver.NewVersion(targetVersion)
	if err != nil {
		return errors.DefaultHTTPError(ctx, http.StatusBadRequest, "target_version is invalid")
	}

	maxCurrent := semver.MustParse("0.6.0")
	minTarget := semver.MustParse("0.7.0")
	if current.GreaterThan(maxCurrent) || target.LessThan(minTarget) {
		return errors.DefaultHTTPError(ctx, http.StatusBadRequest,
			"skill upgrade only supports current_version <= 0.6.0 and target_version >= 0.7.0")
	}
	return nil
}

func (uh *upgradeHandler) migrateSkillReleaseData(ctx context.Context, skill *model.SkillRepositoryDB) (err error) {
	tx, err := uh.DBTx.GetTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	release := &model.SkillReleaseDB{
		SkillID:      skill.SkillID,
		Name:         skill.Name,
		Description:  skill.Description,
		SkillContent: skill.SkillContent,
		Version:      skill.Version,
		Category:     skill.Category,
		Source:       skill.Source,
		ExtendInfo:   skill.ExtendInfo,
		Dependencies: skill.Dependencies,
		FileManifest: skill.FileManifest,
		Status:       interfaces.BizStatusPublished.String(),
		CreateTime:   skill.CreateTime,
		CreateUser:   skill.CreateUser,
		UpdateTime:   skill.UpdateTime,
		UpdateUser:   skill.UpdateUser,
		ReleaseTime:  skill.UpdateTime,
		ReleaseUser:  skill.UpdateUser,
		ReleaseDesc:  "",
	}

	existingRelease, err := uh.SkillReleaseRepo.SelectBySkillID(ctx, tx, skill.SkillID)
	if err != nil {
		return err
	}
	// 如果技能已存在，直接返回
	if existingRelease != nil {
		return nil
	}
	if err = uh.SkillReleaseRepo.Insert(ctx, tx, release); err != nil {
		return err
	}
	existingHistory, err := uh.SkillReleaseHistoryRepo.SelectBySkillIDAndVersion(ctx, tx, skill.SkillID, skill.Version)
	if err != nil {
		return err
	}
	if existingHistory == nil {
		history := &model.SkillReleaseHistoryDB{
			SkillID:      skill.SkillID,
			Version:      skill.Version,
			SkillRelease: utils.ObjectToJSON(release),
			ReleaseDesc:  "",
			CreateTime:   skill.UpdateTime,
			CreateUser:   skill.UpdateUser,
			UpdateTime:   skill.UpdateTime,
			UpdateUser:   skill.UpdateUser,
		}
		if err = uh.SkillReleaseHistoryRepo.Insert(ctx, tx, history); err != nil {
			return err
		}
	}
	return nil
}
