package skill

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	infracommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/category"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/sandbox"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"gopkg.in/yaml.v3"
)

type skillRegistry struct {
	parser                *skillParser
	skillRepo             model.ISkillRepository
	fileRepo              model.ISkillFileIndex
	releaseRepo           model.ISkillReleaseDB
	releaseHistoryRepo    model.ISkillReleaseHistoryDB
	assetStore            skillAssetStore
	indexSync             interfaces.SkillIndexSyncService
	sandboxClient         interfaces.SandBoxControlPlane
	sessionPool           sandbox.SessionPool
	dbTx                  model.DBTx
	AuthService           interfaces.IAuthorizationService
	BusinessDomainService interfaces.IBusinessDomainService
	UserMgnt              interfaces.UserManagement
	Logger                interfaces.Logger
	CategoryManager       interfaces.CategoryManager
}

var (
	registryOnce sync.Once
	registryInst interfaces.SkillRegistry
)

const maxSkillReleaseHistoryVersions = 10

// NewSkillRegistry 创建技能注册器
func NewSkillRegistry() interfaces.SkillRegistry {
	registryOnce.Do(func() {
		registryInst = &skillRegistry{
			Logger:                config.NewConfigLoader().GetLogger(),
			parser:                newSkillParser(),
			skillRepo:             dbaccess.NewSkillRepositoryDB(),
			fileRepo:              dbaccess.NewSkillFileIndexDB(),
			releaseRepo:           dbaccess.NewSkillReleaseDB(),
			releaseHistoryRepo:    dbaccess.NewSkillReleaseHistoryDB(),
			assetStore:            newOSSGatewaySkillAssetStore(),
			indexSync:             NewSkillIndexSyncService(),
			sandboxClient:         drivenadapters.NewSandBoxControlPlaneClient(),
			sessionPool:           sandbox.GetSessionPool(),
			dbTx:                  dbaccess.NewBaseTx(),
			AuthService:           auth.NewAuthServiceImpl(),
			BusinessDomainService: business_domain.NewBusinessDomainService(),
			UserMgnt:              drivenadapters.NewUserManagementClient(),
			CategoryManager:       category.NewCategoryManager(),
		}
	})
	return registryInst
}

// RegisterSkill 注册技能
func (r *skillRegistry) RegisterSkill(ctx context.Context, req *interfaces.RegisterSkillReq) (resp *interfaces.RegisterSkillResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"user_id": req.UserID,
		"bd_id":   req.BusinessDomainID,
	})
	// 检查新建权限
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckCreatePermission(ctx, accessor, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}
	// 检查分类是否合法
	if req.Category != "" {
		if !r.CategoryManager.CheckCategory(req.Category) {
			err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtSkillCategoryNotFound,
				fmt.Sprintf(" %s category not found", req.Category))
			return
		}
	}

	skill, files, assets, err := r.parser.parseRegisterReq(req)
	if err != nil {
		return nil, err
	}
	skill.FileManifest = utils.ObjectToJSON(files)

	tx, err := r.dbTx.GetTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tx failed: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	// 插入技能
	skillID, err := r.skillRepo.InsertSkill(ctx, tx, skill)
	if err != nil {
		return nil, err
	}
	if len(assets) > 0 {
		var fileIndices []*model.SkillFileIndexDB
		fileIndices, err = r.persistSkillAssets(ctx, skillID, skill.Version, assets)
		if err != nil {
			return nil, err
		}
		if err = r.fileRepo.BatchInsertSkillFiles(ctx, tx, fileIndices); err != nil {
			return nil, err
		}
	}
	// 关联技能到业务域
	err = r.BusinessDomainService.AssociateResource(ctx, req.BusinessDomainID, skillID, interfaces.AuthResourceTypeSkill)
	if err != nil {
		return nil, err
	}
	// 触发新建策略，创建人默认拥有对当前资源的所有操作权限
	err = r.AuthService.CreateOwnerPolicy(ctx, accessor, &interfaces.AuthResource{
		ID:   skill.SkillID,
		Type: string(interfaces.AuthResourceTypeSkill),
		Name: skill.Name,
	})
	if err != nil {
		return nil, err
	}

	filePaths := make([]string, 0, len(files))
	for _, file := range files {
		filePaths = append(filePaths, file.RelPath)
	}
	resp = &interfaces.RegisterSkillResp{
		SkillID:     skillID,
		Name:        skill.Name,
		Description: skill.Description,
		Version:     skill.Version,
		Status:      interfaces.BizStatus(skill.Status),
		Files:       filePaths,
	}
	// TODO: 待接入审计日志
	return resp, nil
}

// UpdateSkillMetadata 更新技能元数据
func (r *skillRegistry) UpdateSkillMetadata(ctx context.Context, req *interfaces.UpdateSkillMetadataReq) (resp *interfaces.UpdateSkillMetadataResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id": req.SkillID,
		"user_id":  req.UserID,
		"bd_id":    req.BusinessDomainID,
	})

	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.IsDeleted {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", req.SkillID))
	}
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckModifyPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}
	if req.Name != skill.Name {
		if err = r.checkSkillDuplicateName(ctx, req.Name, req.SkillID); err != nil {
			return nil, err
		}
	}
	if req.Category != "" && !r.CategoryManager.CheckCategory(req.Category) {
		return nil, errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtSkillCategoryNotFound,
			fmt.Sprintf(" %s category not found", req.Category))
	}

	// FR-6: 判断元数据是否有变更，决定是否需要重写 OSS SKILL.md
	nameChanged := req.Name != skill.Name
	descChanged := req.Description != skill.Description
	needsRewrite := nameChanged || descChanged

	tx, err := r.dbTx.GetTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tx failed: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.Logger.WithContext(ctx).Errorf("rollback skill metadata update failed, skill_id=%s, err=%v", req.SkillID, rollbackErr)
			}
		} else {
			commitErr := tx.Commit()
			if commitErr != nil {
				r.Logger.WithContext(ctx).Errorf("commit skill metadata update failed, skill_id=%s, err=%v", req.SkillID, commitErr)
			}
			// FR-6: 事务提交成功后，重写 OSS SKILL.md 的 frontmatter
			if needsRewrite {
				if rewriteErr := r.rewriteSkillMDFrontmatter(ctx, skill.SkillID, skill.Version, req.Name, req.Description); rewriteErr != nil {
					r.Logger.WithContext(ctx).Errorf("rewrite SKILL.md frontmatter failed, skill_id=%s, err=%v", skill.SkillID, rewriteErr)
				}
			}
		}
	}()

	skill.Name = req.Name
	skill.Description = req.Description
	skill.Category = req.Category.String()
	if req.Source != "" {
		skill.Source = req.Source
	}
	if req.ExtendInfo != nil {
		skill.ExtendInfo = string(req.ExtendInfo)
	}
	skill.UpdateUser = req.UserID
	if skill.Status == interfaces.BizStatusPublished.String() {
		skill.Status = interfaces.BizStatusEditing.String()
	}
	if err = r.skillRepo.UpdateSkill(ctx, tx, skill); err != nil {
		return nil, err
	}
	return &interfaces.UpdateSkillMetadataResp{
		SkillID: skill.SkillID,
		Version: skill.Version,
		Status:  interfaces.BizStatus(skill.Status),
	}, nil
}

// UpdateSkillPackage 更新技能包
func (r *skillRegistry) UpdateSkillPackage(ctx context.Context, req *interfaces.UpdateSkillPackageReq) (resp *interfaces.UpdateSkillPackageResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id":  req.SkillID,
		"user_id":   req.UserID,
		"bd_id":     req.BusinessDomainID,
		"file_type": req.FileType,
	})

	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.IsDeleted {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", req.SkillID))
	}
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckModifyPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}

	parsedSkill, files, assets, err := r.parser.parseRegisterReq(&interfaces.RegisterSkillReq{
		UserID:     req.UserID,
		FileType:   req.FileType,
		File:       req.File,
		Category:   interfaces.BizCategory(skill.Category),
		Source:     skill.Source,
		ExtendInfo: json.RawMessage(skill.ExtendInfo),
	})
	if err != nil {
		return nil, err
	}
	if parsedSkill.Name != skill.Name {
		if err = r.checkSkillDuplicateName(ctx, parsedSkill.Name, req.SkillID); err != nil {
			return nil, err
		}
	}
	replaceCurrentVersion := skill.Status == interfaces.BizStatusEditing.String() ||
		skill.Status == interfaces.BizStatusUnpublish.String()
	targetVersion := parsedSkill.Version
	if replaceCurrentVersion {
		targetVersion = skill.Version
	}

	tx, err := r.dbTx.GetTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tx failed: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.Logger.WithContext(ctx).Errorf("rollback skill package update failed, skill_id=%s, err=%v", req.SkillID, rollbackErr)
			}
		} else {
			commitErr := tx.Commit()
			if commitErr != nil {
				r.Logger.WithContext(ctx).Errorf("commit skill package update failed, skill_id=%s, err=%v", req.SkillID, commitErr)
			}
		}
	}()

	skill.Name = parsedSkill.Name
	skill.Description = parsedSkill.Description
	skill.SkillContent = parsedSkill.SkillContent
	skill.Version = targetVersion
	skill.Dependencies = parsedSkill.Dependencies
	skill.ExtendInfo = parsedSkill.ExtendInfo
	skill.FileManifest = utils.ObjectToJSON(files)
	skill.UpdateUser = req.UserID
	if skill.Status == interfaces.BizStatusPublished.String() || skill.Status == interfaces.BizStatusOffline.String() {
		skill.Status = interfaces.BizStatusEditing.String()
	}
	if replaceCurrentVersion {
		existingFiles, selectErr := r.fileRepo.SelectSkillFileBySkillID(ctx, tx, skill.SkillID, targetVersion)
		if selectErr != nil {
			return nil, selectErr
		}
		for _, file := range existingFiles {
			if deleteErr := r.assetStore.Delete(ctx, &interfaces.OssObject{
				StorageID:  file.StorageID,
				StorageKey: file.StorageKey,
			}); deleteErr != nil {
				return nil, deleteErr
			}
		}
		if err = r.fileRepo.DeleteSkillFileBySkillID(ctx, tx, skill.SkillID, targetVersion); err != nil {
			return nil, err
		}
	}
	if err = r.skillRepo.UpdateSkill(ctx, tx, skill); err != nil {
		return nil, err
	}
	if len(assets) > 0 {
		fileIndices, persistErr := r.persistSkillAssets(ctx, skill.SkillID, targetVersion, assets)
		if persistErr != nil {
			return nil, persistErr
		}
		if err = r.fileRepo.BatchInsertSkillFiles(ctx, tx, fileIndices); err != nil {
			return nil, err
		}
	}
	resp = &interfaces.UpdateSkillPackageResp{
		SkillID: skill.SkillID,
		Version: skill.Version,
		Status:  interfaces.BizStatus(skill.Status),
	}
	if r.indexSync != nil {
		if syncErr := r.indexSync.UpdateSkill(ctx, skill); syncErr != nil {
			r.Logger.WithContext(ctx).Errorf("update skill index failed after package update, skill_id=%s, err=%v", req.SkillID, syncErr)
		}
	}
	return resp, nil
}

// RepublishSkillHistory 将历史版本回灌到草稿态
func (r *skillRegistry) RepublishSkillHistory(ctx context.Context, req *interfaces.RepublishSkillHistoryReq) (resp *interfaces.RepublishSkillHistoryResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id": req.SkillID,
		"user_id":  req.UserID,
		"bd_id":    req.BusinessDomainID,
		"version":  req.Version,
	})

	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.IsDeleted {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", req.SkillID))
	}
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckModifyPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}

	history, err := r.releaseHistoryRepo.SelectBySkillIDAndVersion(ctx, nil, req.SkillID, req.Version)
	if err != nil {
		return nil, err
	}
	if history == nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill history not found: %s@%s", req.SkillID, req.Version))
	}
	release := utils.JSONToObject[*model.SkillReleaseDB](history.SkillRelease)
	if release == nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("invalid skill history snapshot: %s@%s", req.SkillID, req.Version))
	}

	tx, err := r.dbTx.GetTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tx failed: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.Logger.WithContext(ctx).Errorf("rollback skill history republish failed, skill_id=%s, version=%s, err=%v", req.SkillID, req.Version, rollbackErr)
			}
		} else {
			commitErr := tx.Commit()
			if commitErr != nil {
				r.Logger.WithContext(ctx).Errorf("commit skill history republish failed, skill_id=%s, version=%s, err=%v", req.SkillID, req.Version, commitErr)
			}
		}
	}()

	skill.Name = release.Name
	skill.Description = release.Description
	skill.SkillContent = release.SkillContent
	skill.Version = release.Version
	skill.Category = release.Category
	skill.Source = release.Source
	skill.ExtendInfo = release.ExtendInfo
	skill.Dependencies = release.Dependencies
	skill.FileManifest = release.FileManifest
	skill.Status = interfaces.BizStatusEditing.String()
	skill.UpdateUser = req.UserID
	if err = r.skillRepo.UpdateSkill(ctx, tx, skill); err != nil {
		return nil, err
	}
	return &interfaces.RepublishSkillHistoryResp{
		SkillID: skill.SkillID,
		Version: skill.Version,
		Status:  interfaces.BizStatus(skill.Status),
	}, nil
}

// PublishSkillHistory 直接发布历史版本
func (r *skillRegistry) PublishSkillHistory(ctx context.Context, req *interfaces.PublishSkillHistoryReq) (resp *interfaces.PublishSkillHistoryResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id": req.SkillID,
		"user_id":  req.UserID,
		"bd_id":    req.BusinessDomainID,
		"version":  req.Version,
	})

	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.IsDeleted {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", req.SkillID))
	}
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckModifyPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckPublishPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}

	history, err := r.releaseHistoryRepo.SelectBySkillIDAndVersion(ctx, nil, req.SkillID, req.Version)
	if err != nil {
		return nil, err
	}
	if history == nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill history not found: %s@%s", req.SkillID, req.Version))
	}
	release := utils.JSONToObject[*model.SkillReleaseDB](history.SkillRelease)
	if release == nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("invalid skill history snapshot: %s@%s", req.SkillID, req.Version))
	}
	if err = r.checkSkillDuplicateName(ctx, release.Name, req.SkillID); err != nil {
		return nil, err
	}

	tx, err := r.dbTx.GetTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tx failed: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.Logger.WithContext(ctx).Errorf("rollback skill history publish failed, skill_id=%s, version=%s, err=%v", req.SkillID, req.Version, rollbackErr)
			}
		} else {
			commitErr := tx.Commit()
			if commitErr != nil {
				r.Logger.WithContext(ctx).Errorf("commit skill history publish failed, skill_id=%s, version=%s, err=%v", req.SkillID, req.Version, commitErr)
			}
		}
	}()

	skill.Name = release.Name
	skill.Description = release.Description
	skill.SkillContent = release.SkillContent
	skill.Version = release.Version
	skill.Category = release.Category
	skill.Source = release.Source
	skill.ExtendInfo = release.ExtendInfo
	skill.Dependencies = release.Dependencies
	skill.FileManifest = release.FileManifest
	skill.Status = interfaces.BizStatusPublished.String()
	skill.UpdateUser = req.UserID
	if err = r.skillRepo.UpdateSkill(ctx, tx, skill); err != nil {
		return nil, err
	}
	if err = r.publishSkillSnapshot(ctx, tx, skill, req.UserID); err != nil {
		return nil, err
	}
	if r.indexSync != nil {
		if syncErr := r.indexSync.UpsertSkill(ctx, skill); syncErr != nil {
			r.Logger.WithContext(ctx).Errorf("sync published historical skill index failed, skill_id=%s, version=%s, err=%v", req.SkillID, req.Version, syncErr)
		}
	}
	return &interfaces.PublishSkillHistoryResp{
		SkillID: skill.SkillID,
		Version: skill.Version,
		Status:  interfaces.BizStatus(skill.Status),
	}, nil
}

// DeleteSkill 删除技能
func (r *skillRegistry) DeleteSkill(ctx context.Context, req *interfaces.DeleteSkillReq) (err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id": req.SkillID,
		"user_id":  req.UserID,
		"bd_id":    req.BusinessDomainID,
	})
	// 检查删除权限
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return err
	}
	if err = r.AuthService.CheckDeletePermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return err
	}
	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("select skill by id failed: %s", err.Error()))
		return
	}
	// 技能不存在，或者已经删除
	if skill == nil || skill.IsDeleted {
		err = errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", req.SkillID))
		return
	}
	// 删除状态校验沿用公共状态管理
	if !common.CanDelete(interfaces.BizStatus(skill.Status)) {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtSkillUnSupportDelete,
			fmt.Sprintf("skill can not be deleted in status: %s", skill.Status))
		return
	}
	tx, err := r.dbTx.GetTx(ctx)
	if err != nil {
		return fmt.Errorf("get tx failed: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.Logger.WithContext(ctx).Errorf("rollback skill delete failed, skill_id=%s, err=%v", req.SkillID, rollbackErr)
				return
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				r.Logger.WithContext(ctx).Errorf("commit skill delete failed, skill_id=%s, err=%v", req.SkillID, commitErr)
				return
			}
		}
	}()

	// 将技能标记为删除中，TODO：需要设计一个单独的协程用于处理删除中断的兜底策略
	if err = r.skillRepo.UpdateSkillDeleted(ctx, tx, req.SkillID, true, req.UserID); err != nil {
		return err
	}
	// 查找索引文件
	files, err := r.fileRepo.SelectSkillFileBySkillID(ctx, tx, req.SkillID, skill.Version)
	if err != nil {
		return err
	}
	// 先删除对象存储中的记录，再删除数据库中的记录
	for _, file := range files {
		if err = r.assetStore.Delete(ctx, &interfaces.OssObject{
			StorageID:  file.StorageID,
			StorageKey: file.StorageKey,
		}); err != nil {
			r.Logger.WithContext(ctx).Warnf("delete file failed, err:%s", err.Error())
			return err
		}
	}
	if err = r.fileRepo.DeleteSkillFileBySkillID(ctx, tx, req.SkillID, skill.Version); err != nil {
		return err
	}
	if r.releaseRepo != nil {
		if err = r.deletePublishedSkillSnapshot(ctx, tx, req.SkillID); err != nil {
			return err
		}
	}
	if r.releaseHistoryRepo != nil {
		if err = r.releaseHistoryRepo.DeleteBySkillID(ctx, tx, req.SkillID); err != nil {
			return err
		}
	}
	if err = r.skillRepo.DeleteSkillByID(ctx, tx, req.SkillID); err != nil {
		return err
	}
	// 取消技能与业务域的关联
	if err = r.BusinessDomainService.DisassociateResource(ctx, req.BusinessDomainID, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return err
	}
	// 删除技能的权限策略
	if err = r.AuthService.DeletePolicy(ctx, []string{req.SkillID}, interfaces.AuthResourceTypeSkill); err != nil {
		return err
	}
	if r.indexSync != nil {
		if syncErr := r.indexSync.DeleteSkill(ctx, req.SkillID); syncErr != nil {
			r.Logger.WithContext(ctx).Errorf("delete skill index failed after skill delete, skill_id=%s, err=%v", req.SkillID, syncErr)
		}
	}
	return nil
}

// UpdateSkillStatus 更新技能状态
func (r *skillRegistry) UpdateSkillStatus(ctx context.Context, req *interfaces.UpdateSkillStatusReq) (resp *interfaces.UpdateSkillStatusResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id": req.SkillID,
		"user_id":  req.UserID,
		"bd_id":    req.BusinessDomainID,
		"status":   req.Status,
	})
	// 获取技能
	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	// 技能不存在，或者已经删除
	if skill == nil || skill.IsDeleted {
		err = errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", req.SkillID))
		return
	}
	// 检查状态变换是否合法
	if !common.CheckStatusTransition(interfaces.BizStatus(skill.Status), req.Status) {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtSkillStatusInvalid,
			fmt.Sprintf("skill status can not be updated from %s to %s", skill.Status, req.Status))
		return
	}
	// 检查更新权限
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	switch req.Status {
	case interfaces.BizStatusPublished:
		// 检查是否有发布权限
		err = r.AuthService.CheckPublishPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill)
		if err != nil {
			return nil, err
		}
		// 检查是否重名
		err = r.checkSkillDuplicateName(ctx, skill.Name, skill.SkillID)
		if err != nil {
			return nil, err
		}
	case interfaces.BizStatusUnpublish, interfaces.BizStatusEditing:
	case interfaces.BizStatusOffline:
		err = r.AuthService.CheckUnpublishPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill)
		if err != nil {
			return nil, err
		}
	}
	tx, err := r.dbTx.GetTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tx failed: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.Logger.WithContext(ctx).Errorf("rollback skill status update failed, skill_id=%s, err=%v", req.SkillID, rollbackErr)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				r.Logger.WithContext(ctx).Errorf("commit skill status update failed, skill_id=%s, err=%v", req.SkillID, commitErr)
				return
			}
		}
	}()
	// 更新技能状态
	if err = r.skillRepo.UpdateSkillStatus(ctx, tx, req.SkillID, string(req.Status), req.UserID); err != nil {
		return nil, err
	}
	switch req.Status {
	case interfaces.BizStatusPublished:
		if err = r.publishSkillSnapshot(ctx, tx, skill, req.UserID); err != nil {
			return nil, err
		}
	case interfaces.BizStatusOffline:
		if err = r.deletePublishedSkillSnapshot(ctx, tx, req.SkillID); err != nil {
			return nil, err
		}
	}
	resp = &interfaces.UpdateSkillStatusResp{
		SkillID: req.SkillID,
		Status:  req.Status,
	}
	// 将skill数据写到dataset，但是不阻塞主流程
	switch req.Status {
	case interfaces.BizStatusPublished:
		if r.indexSync != nil {
			if syncErr := r.indexSync.UpsertSkill(ctx, skill); syncErr != nil {
				r.Logger.WithContext(ctx).Errorf("sync published skill index failed, skill_id=%s, err=%v", req.SkillID, syncErr)
			}
		}
	case interfaces.BizStatusOffline:
		if r.indexSync != nil {
			if syncErr := r.indexSync.DeleteSkill(ctx, req.SkillID); syncErr != nil {
				r.Logger.WithContext(ctx).Errorf("delete skill index failed after status update, skill_id=%s, err=%v", req.SkillID, syncErr)
			}
		}
	}
	return resp, nil
}

// 重名检查
func (r *skillRegistry) checkSkillDuplicateName(ctx context.Context, name string, skillID string) (err error) {
	has, skillDB, err := r.skillRepo.SelectSkillByName(ctx, nil, name, []string{string(interfaces.BizStatusPublished)})
	if err != nil {
		r.Logger.WithContext(ctx).Errorf("select skill by name failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "select skill by name failed")
		return
	}
	if !has || (skillID != "" && skillDB.SkillID == skillID) {
		return
	}
	// 存在
	err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtSkillNameDuplicate,
		fmt.Sprintf("skill name %s already exists", name), name)
	return
}

// DownloadSkill 下载技能
func (r *skillRegistry) DownloadSkill(ctx context.Context, req *interfaces.DownloadSkillReq) (resp *interfaces.DownloadSkillResp, err error) {
	// 记录可观测性
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id": req.SkillID,
		"user_id":  req.UserID,
		"bd_id":    req.BusinessDomainID,
	})

	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	// 检查是否有查看或者公开访问权限
	authorized, err := r.AuthService.OperationCheckAny(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill,
		interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess)
	if err != nil {
		return
	}
	if !authorized {
		err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden, nil)
		return
	}
	skill, fileName, archive, err := r.buildSkillArchive(ctx, req.SkillID)
	if err != nil {
		return nil, err
	}

	return &interfaces.DownloadSkillResp{
		SkillID:  skill.SkillID,
		FileName: fileName,
		Content:  archive,
	}, nil
}

// ExecuteSkill 执行技能
func (r *skillRegistry) ExecuteSkill(ctx context.Context, req *interfaces.ExecuteSkillReq) (resp *interfaces.ExecuteSkillResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"user_id":  req.UserID,
		"bd_id":    req.BusinessDomainID,
		"skill_id": req.SkillID,
	})

	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	authorized, err := r.AuthService.OperationCheckAny(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill,
		interfaces.AuthOperationTypeExecute, interfaces.AuthOperationTypePublicAccess)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden, nil)
	}

	skill, fileName, archive, err := r.buildSkillArchive(ctx, req.SkillID)
	if err != nil {
		return nil, err
	}

	sessionID, err := r.sessionPool.AcquireSession(ctx)
	if err != nil {
		return nil, err
	}
	defer r.sessionPool.ReleaseSession(sessionID)

	uploadWorkDir := path.Join("skills", req.SkillID)

	uploadResp, err := r.sandboxClient.UploadSkillArchive(ctx, sessionID, &interfaces.UploadSkillArchiveReq{
		WorkDir:  uploadWorkDir,
		FileName: fileName,
		Content:  archive,
	})
	if err != nil {
		return nil, err
	}
	execResp, err := r.sandboxClient.ExecuteShell(ctx, sessionID, &interfaces.ExecuteShellReq{
		WorkDir: uploadResp.WorkDir,
		Command: req.EntryShell,
		Timeout: req.Timeout,
	})
	if err != nil {
		return nil, err
	}

	return &interfaces.ExecuteSkillResp{
		SkillID:       skill.SkillID,
		SessionID:     sessionID,
		WorkDir:       uploadResp.WorkDir,
		FileName:      uploadResp.FileName,
		UploadedPath:  uploadResp.UploadedPath,
		Command:       execResp.Command,
		ExitCode:      execResp.ExitCode,
		Stdout:        execResp.Stdout,
		Stderr:        execResp.Stderr,
		ExecutionTime: execResp.ExecutionTime,
		Mocked:        uploadResp.Mocked || execResp.Mocked,
	}, nil
}

func (r *skillRegistry) buildSkillArchive(ctx context.Context, skillID string) (*model.SkillRepositoryDB, string, []byte, error) {
	skillDB, err := r.skillRepo.SelectSkillByID(ctx, nil, skillID)
	if err != nil {
		return nil, "", nil, err
	}
	if skillDB == nil {
		return nil, "", nil, fmt.Errorf("skill not found: %s", skillID)
	}
	files, err := r.fileRepo.SelectSkillFileBySkillID(ctx, nil, skillID, skillDB.Version)
	if err != nil {
		return nil, "", nil, err
	}
	return r.buildSkillArchiveFromSnapshot(ctx, skillDB, files)
}

func (r *skillRegistry) buildSkillArchiveFromSnapshot(ctx context.Context, skill *model.SkillRepositoryDB,
	files []*model.SkillFileIndexDB) (*model.SkillRepositoryDB, string, []byte, error) {
	var buf bytes.Buffer
	var err error
	zw := zip.NewWriter(&buf)
	writeFile := func(name string, content []byte) error {
		w, createErr := zw.Create(name)
		if createErr != nil {
			return createErr
		}
		_, writeErr := io.Copy(w, bytes.NewReader(content))
		return writeErr
	}

	for _, file := range files {
		content, readErr := r.assetStore.Download(ctx, &interfaces.OssObject{
			StorageID:  file.StorageID,
			StorageKey: file.StorageKey,
		})
		if readErr != nil {
			_ = zw.Close()
			return nil, "", nil, readErr
		}
		if err = writeFile(file.RelPath, content); err != nil {
			_ = zw.Close()
			return nil, "", nil, err
		}
	}
	if err = zw.Close(); err != nil {
		return nil, "", nil, err
	}

	return skill, fmt.Sprintf("%s.zip", skill.Name), buf.Bytes(), nil
}

// QuerySkillList 查询技能列表（管理接口）
func (r *skillRegistry) QuerySkillList(ctx context.Context, req *interfaces.QuerySkillListReq) (resp *interfaces.QuerySkillListResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"user_id": req.UserID,
		"bd_id":   req.BusinessDomainID,
	})
	resp = &interfaces.QuerySkillListResp{
		CommonPageResult: interfaces.CommonPageResult{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		Data: []*interfaces.SkillInfo{},
	}
	// 条件构建
	filter := map[string]interface{}{
		"all":         req.All,
		"name":        req.Name,
		"create_user": req.CreateUser,
		"status":      req.Status.String(),
	}
	// 检查分类是否合法
	if req.Category != "" {
		if !r.CategoryManager.CheckCategory(req.Category) {
			err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtSkillCategoryNotFound,
				fmt.Sprintf(" %s category not found", req.Category))
			return
		}
		filter["category"] = req.Category.String()
	}

	authResp, resourceToBdMap, err := r.querySkillListPage(ctx, filter, req.CommonPageParams, req.UserID, interfaces.AuthOperationTypeView)
	if err != nil {
		return nil, err
	}
	resp.CommonPageResult = authResp.CommonPageResult
	if len(authResp.Data) == 0 {
		return resp, nil
	}
	skillInfos, err := r.assembleSkillInfoList(ctx, authResp.Data, resourceToBdMap)
	if err != nil {
		return nil, err
	}
	resp.Data = skillInfos
	return resp, nil
}

// 组装技能市场摘要列表
func (r *skillRegistry) assembleMarketSkillInfoList(ctx context.Context, releaseDB []*model.SkillReleaseDB, resourceToBdMap map[string]string) (skillInfos []*interfaces.SkillInfo, err error) {
	var userIDs []string
	skillInfos = []*interfaces.SkillInfo{}
	for _, relese := range releaseDB {
		skillInfos = append(skillInfos, convertSkillMarketDetail(relese, r.CategoryManager.GetCategoryName(ctx, interfaces.BizCategory(relese.Category))))
		userIDs = append(userIDs, relese.CreateUser, relese.UpdateUser, relese.ReleaseUser)
	}
	// 获取用户名称
	userMap, err := r.UserMgnt.GetUsersName(ctx, userIDs)
	if err != nil {
		return
	}
	businessDomainIDStr, _ := infracommon.GetBusinessDomainFromCtx(ctx)
	for _, skill := range skillInfos {
		skill.CreateUser = utils.GetValueOrDefault(userMap, skill.CreateUser, interfaces.UnknownUser)
		skill.UpdateUser = utils.GetValueOrDefault(userMap, skill.UpdateUser, interfaces.UnknownUser)
		skill.ReleaseUser = utils.GetValueOrDefault(userMap, skill.ReleaseUser, interfaces.UnknownUser)
		skill.BusinessDomainID = utils.GetValueOrDefault(resourceToBdMap, skill.SkillID, businessDomainIDStr)
	}
	return
}

// 组装技能返回信息列表
func (r *skillRegistry) assembleSkillInfoList(ctx context.Context, skillDBs []*model.SkillRepositoryDB, resourceToBdMap map[string]string) (skillInfos []*interfaces.SkillInfo, err error) {
	var userIDs []string
	skillInfos = []*interfaces.SkillInfo{}
	for _, skill := range skillDBs {
		skillInfos = append(skillInfos, convertSkillDetail(skill, r.CategoryManager.GetCategoryName(ctx, interfaces.BizCategory(skill.Category))))
		userIDs = append(userIDs, skill.CreateUser, skill.UpdateUser)
	}
	// 获取用户名称
	userMap, err := r.UserMgnt.GetUsersName(ctx, userIDs)
	if err != nil {
		return
	}
	businessDomainIDStr, _ := infracommon.GetBusinessDomainFromCtx(ctx)
	for _, skill := range skillInfos {
		skill.CreateUser = utils.GetValueOrDefault(userMap, skill.CreateUser, interfaces.UnknownUser)
		skill.UpdateUser = utils.GetValueOrDefault(userMap, skill.UpdateUser, interfaces.UnknownUser)
		skill.BusinessDomainID = utils.GetValueOrDefault(resourceToBdMap, skill.SkillID, businessDomainIDStr)
	}
	return
}

func (r *skillRegistry) queryReleaseListPage(ctx context.Context, filter map[string]interface{}, pageParamsReq interfaces.CommonPageParams, userID string, operations ...interfaces.AuthOperationType) (
	authResp *interfaces.QueryResponse[model.SkillReleaseDB], resourceToBdMap map[string]string, err error) {
	// 构建查询执行器
	sortField := "f_update_time"
	switch pageParamsReq.SortBy {
	case "create_time":
		sortField = "f_create_time"
	case "name":
		sortField = "f_name"
	}
	sortOrder := ormhelper.SortOrderDesc
	if pageParamsReq.SortOrder == "asc" {
		sortOrder = ormhelper.SortOrderAsc
	}
	sort := &ormhelper.SortParams{Fields: []ormhelper.SortField{{Field: sortField, Order: sortOrder}}}
	// 统计总条数
	queryTotal := func(newCtx context.Context) (int64, error) {
		var count int64
		count, err = r.releaseRepo.CountByWhereClause(ctx, nil, filter)
		if err != nil {
			r.Logger.WithContext(newCtx).Errorf("count skill list failed, err: %v", err)
			err = errors.DefaultHTTPError(newCtx, http.StatusInternalServerError, "count skill list failed")
			return 0, err
		}
		return count, nil
	}
	// queryBatch 查询技能列表分页
	queryBatch := func(newCtx context.Context, pageSize int, offset int, cursorValue *model.SkillReleaseDB) ([]*model.SkillReleaseDB, error) {
		var skills []*model.SkillReleaseDB
		var cursor *ormhelper.CursorParams
		if cursorValue != nil {
			cursor = &ormhelper.CursorParams{
				Field:     sortField,
				Direction: ormhelper.SortOrder(pageParamsReq.SortOrder),
			}
			switch sortField {
			case "f_update_time":
				cursor.Value = cursorValue.UpdateTime
			case "f_create_time":
				cursor.Value = cursorValue.CreateTime
			case "f_name":
				cursor.Value = cursorValue.Name
			}
			// 如果使用游标不需要offset
			offset = 0
		}
		filter["limit"] = pageSize
		filter["offset"] = offset
		skills, err = r.releaseRepo.SelectListPage(ctx, nil, filter, sort, cursor)
		if err != nil {
			r.Logger.WithContext(newCtx).Errorf("select skill list page failed, err: %v", err)
			err = errors.DefaultHTTPError(newCtx, http.StatusInternalServerError, "select skill list page failed")
			return nil, err
		}
		return skills, nil
	}
	businessDomainStr, _ := infracommon.GetBusinessDomainFromCtx(ctx)
	businessDomainIDs := strings.Split(businessDomainStr, ",")
	resourceToBdMap, err = r.BusinessDomainService.BatchResourceList(ctx, businessDomainIDs, interfaces.AuthResourceTypeSkill)
	if err != nil {
		return
	}
	queryBuilder := auth.NewQueryBuilder[model.SkillReleaseDB]().
		SetPage(pageParamsReq.Page, pageParamsReq.PageSize).SetAll(pageParamsReq.All).
		SetQueryFunctions(queryTotal, queryBatch).
		SetFilteredQueryFunctions( // 带过滤条件的查询函数
			func(newCtx context.Context, ids []string) (int64, error) {
				filter["in"] = ids
				return queryTotal(newCtx)
			},
			func(newCtx context.Context, pageSize int, offset int, ids []string, cursorValue *model.SkillReleaseDB) ([]*model.SkillReleaseDB, error) {
				filter["in"] = ids
				return queryBatch(newCtx, pageSize, offset, cursorValue)
			},
		).
		SetBusinessDomainFilter(func(newCtx context.Context) ([]string, error) {
			// 从业务域中过滤相关技能
			resourceIDs := make([]string, 0, len(resourceToBdMap))
			for resourceID := range resourceToBdMap {
				resourceIDs = append(resourceIDs, resourceID)
			}
			return resourceIDs, nil
		})
	if infracommon.IsPublicAPIFromCtx(ctx) {
		// 如果是外部接口，权限检查
		queryBuilder.SetAuthFilter(func(newCtx context.Context) ([]string, error) {
			// 检查查看权限
			var accessor *interfaces.AuthAccessor
			accessor, err = r.AuthService.GetAccessor(newCtx, userID)
			if err != nil {
				return nil, err
			}
			return r.AuthService.ResourceListIDs(newCtx, accessor, interfaces.AuthResourceTypeSkill, operations...)
		})
	}
	authResp, err = queryBuilder.Execute(ctx)
	return
}

func (r *skillRegistry) querySkillListPage(ctx context.Context, filter map[string]interface{}, pageParamsReq interfaces.CommonPageParams, userID string, operations ...interfaces.AuthOperationType) (
	authResp *interfaces.QueryResponse[model.SkillRepositoryDB], resourceToBdMap map[string]string, err error) {
	// 构建查询执行器
	sortField := "f_update_time"
	switch pageParamsReq.SortBy {
	case "create_time":
		sortField = "f_create_time"
	case "name":
		sortField = "f_name"
	}
	sortOrder := ormhelper.SortOrderDesc
	if pageParamsReq.SortOrder == "asc" {
		sortOrder = ormhelper.SortOrderAsc
	}
	sort := &ormhelper.SortParams{Fields: []ormhelper.SortField{{Field: sortField, Order: sortOrder}}}
	// 统计总条数
	queryTotal := func(newCtx context.Context) (int64, error) {
		var count int64
		count, err = r.skillRepo.CountByWhereClause(ctx, nil, filter)
		if err != nil {
			r.Logger.WithContext(newCtx).Errorf("count skill list failed, err: %v", err)
			err = errors.DefaultHTTPError(newCtx, http.StatusInternalServerError, "count skill list failed")
			return 0, err
		}
		return count, nil
	}
	// queryBatch 查询技能列表分页
	queryBatch := func(newCtx context.Context, pageSize int, offset int, cursorValue *model.SkillRepositoryDB) ([]*model.SkillRepositoryDB, error) {
		var skills []*model.SkillRepositoryDB
		var cursor *ormhelper.CursorParams
		if cursorValue != nil {
			cursor = &ormhelper.CursorParams{
				Field:     sortField,
				Direction: ormhelper.SortOrder(pageParamsReq.SortOrder),
			}
			switch sortField {
			case "f_update_time":
				cursor.Value = cursorValue.UpdateTime
			case "f_create_time":
				cursor.Value = cursorValue.CreateTime
			case "f_name":
				cursor.Value = cursorValue.Name
			}
			// 如果使用游标不需要offset
			offset = 0
		}
		filter["limit"] = pageSize
		filter["offset"] = offset
		skills, err = r.skillRepo.SelectSkillListPage(ctx, nil, filter, sort, cursor)
		if err != nil {
			r.Logger.WithContext(newCtx).Errorf("select skill list page failed, err: %v", err)
			err = errors.DefaultHTTPError(newCtx, http.StatusInternalServerError, "select skill list page failed")
			return nil, err
		}
		return skills, nil
	}
	businessDomainStr, _ := infracommon.GetBusinessDomainFromCtx(ctx)
	businessDomainIDs := strings.Split(businessDomainStr, ",")
	resourceToBdMap, err = r.BusinessDomainService.BatchResourceList(ctx, businessDomainIDs, interfaces.AuthResourceTypeSkill)
	if err != nil {
		return
	}
	queryBuilder := auth.NewQueryBuilder[model.SkillRepositoryDB]().
		SetPage(pageParamsReq.Page, pageParamsReq.PageSize).SetAll(pageParamsReq.All).
		SetQueryFunctions(queryTotal, queryBatch).
		SetFilteredQueryFunctions( // 带过滤条件的查询函数
			func(newCtx context.Context, ids []string) (int64, error) {
				filter["in"] = ids
				return queryTotal(newCtx)
			},
			func(newCtx context.Context, pageSize int, offset int, ids []string, cursorValue *model.SkillRepositoryDB) ([]*model.SkillRepositoryDB, error) {
				filter["in"] = ids
				return queryBatch(newCtx, pageSize, offset, cursorValue)
			},
		).
		SetBusinessDomainFilter(func(newCtx context.Context) ([]string, error) {
			// 从业务域中过滤相关技能
			resourceIDs := make([]string, 0, len(resourceToBdMap))
			for resourceID := range resourceToBdMap {
				resourceIDs = append(resourceIDs, resourceID)
			}
			return resourceIDs, nil
		})
	if infracommon.IsPublicAPIFromCtx(ctx) {
		// 如果是外部接口，权限检查
		queryBuilder.SetAuthFilter(func(newCtx context.Context) ([]string, error) {
			// 检查查看权限
			var accessor *interfaces.AuthAccessor
			accessor, err = r.AuthService.GetAccessor(newCtx, userID)
			if err != nil {
				return nil, err
			}
			return r.AuthService.ResourceListIDs(newCtx, accessor, interfaces.AuthResourceTypeSkill, operations...)
		})
	}
	authResp, err = queryBuilder.Execute(ctx)
	return
}

// QuerySkillMarketList 查询技能市场列表
func (r *skillRegistry) QuerySkillMarketList(ctx context.Context, req *interfaces.QuerySkillMarketListReq) (resp *interfaces.QuerySkillMarketListResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"user_id": req.UserID,
		"bd_id":   req.BusinessDomainID,
	})

	resp = &interfaces.QuerySkillMarketListResp{
		CommonPageResult: interfaces.CommonPageResult{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		Data: []*interfaces.SkillInfo{},
	}
	// 条件构建
	filter := map[string]interface{}{
		"all":         req.All,
		"name":        req.Name,
		"create_user": req.CreateUser,
		"status":      interfaces.BizStatusPublished.String(),
	}
	// 检查分类是否合法
	if req.Category != "" {
		if !r.CategoryManager.CheckCategory(req.Category) {
			err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtSkillCategoryNotFound,
				fmt.Sprintf(" %s category not found", req.Category))
			return
		}
		filter["category"] = req.Category.String()
	}

	authResp, resourceToBdMap, err := r.queryReleaseListPage(ctx, filter, req.CommonPageParams, req.UserID,
		interfaces.AuthOperationTypePublicAccess)
	if err != nil {
		return nil, err
	}
	resp.CommonPageResult = authResp.CommonPageResult
	if len(authResp.Data) == 0 {
		return resp, nil
	}
	skillInfos, err := r.assembleMarketSkillInfoList(ctx, authResp.Data, resourceToBdMap)
	if err != nil {
		return nil, err
	}
	resp.Data = skillInfos
	return resp, nil
}

// GetSkillMarketDetail 获取技能市场详情
func (r *skillRegistry) GetSkillMarketDetail(ctx context.Context, req *interfaces.GetSkillMarketDetailReq) (resp *interfaces.SkillInfo, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"user_id":  req.UserID,
		"bd_id":    req.BusinessDomainID,
		"skill_id": req.SkillID,
	})
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckPublicAccessPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}
	release, err := r.releaseRepo.SelectBySkillID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if release == nil {
		err = errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", req.SkillID))
		return nil, err
	}
	skillInfo := convertSkillMarketDetail(release, r.CategoryManager.GetCategoryName(ctx, interfaces.BizCategory(release.Category)))
	userNames, err := r.UserMgnt.GetUsersName(ctx, []string{release.CreateUser, release.UpdateUser, release.ReleaseUser})
	if err != nil {
		return nil, err
	}
	skillInfo.CreateUser = utils.GetValueOrDefault(userNames, release.CreateUser, interfaces.UnknownUser)
	skillInfo.UpdateUser = utils.GetValueOrDefault(userNames, release.UpdateUser, interfaces.UnknownUser)
	skillInfo.ReleaseUser = utils.GetValueOrDefault(userNames, release.ReleaseUser, interfaces.UnknownUser)
	return skillInfo, nil
}

// GetSkillDetail 获取技能详情
func (r *skillRegistry) GetSkillDetail(ctx context.Context, req *interfaces.GetSkillDetailReq) (resp *interfaces.SkillInfo, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"user_id":  req.UserID,
		"bd_id":    req.BusinessDomainID,
		"skill_id": req.SkillID,
	})
	accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if err = r.AuthService.CheckViewPermission(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill); err != nil {
		return nil, err
	}
	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.IsDeleted {
		return nil, fmt.Errorf("skill not found: %s", req.SkillID)
	}
	skillInfo := convertSkillDetail(skill, r.CategoryManager.GetCategoryName(ctx, interfaces.BizCategory(skill.Category)))
	// 获取用户信息
	userNames, err := r.UserMgnt.GetUsersName(ctx, []string{skill.CreateUser, skill.UpdateUser})
	if err != nil {
		return nil, err
	}
	skillInfo.CreateUser = utils.GetValueOrDefault(userNames, skill.CreateUser, interfaces.UnknownUser)
	skillInfo.UpdateUser = utils.GetValueOrDefault(userNames, skill.UpdateUser, interfaces.UnknownUser)
	return skillInfo, nil
}

func convertSkillDetail(skill *model.SkillRepositoryDB, categoryName string) *interfaces.SkillInfo {
	return &interfaces.SkillInfo{
		SkillID:      skill.SkillID,
		Name:         skill.Name,
		Description:  skill.Description,
		Version:      skill.Version,
		Category:     interfaces.BizCategory(skill.Category),
		CategoryName: categoryName,
		Status:       interfaces.BizStatus(skill.Status),
		Source:       skill.Source,
		Dependencies: utils.JSONToObject[map[string]interface{}](skill.Dependencies),
		ExtendInfo:   utils.JSONToObject[map[string]interface{}](skill.ExtendInfo),
		CreateUser:   skill.CreateUser,
		CreateTime:   skill.CreateTime,
		UpdateUser:   skill.UpdateUser,
		UpdateTime:   skill.UpdateTime,
	}
}

func convertSkillMarketDetail(skill *model.SkillReleaseDB, categoryName string) *interfaces.SkillInfo {
	return &interfaces.SkillInfo{
		SkillID:      skill.SkillID,
		Name:         skill.Name,
		Description:  skill.Description,
		Version:      skill.Version,
		Category:     interfaces.BizCategory(skill.Category),
		CategoryName: categoryName,
		Status:       interfaces.BizStatus(skill.Status),
		Source:       skill.Source,
		Dependencies: utils.JSONToObject[map[string]interface{}](skill.Dependencies),
		ExtendInfo:   utils.JSONToObject[map[string]interface{}](skill.ExtendInfo),
		CreateUser:   skill.CreateUser,
		CreateTime:   skill.CreateTime,
		UpdateUser:   skill.UpdateUser,
		UpdateTime:   skill.UpdateTime,
		ReleaseTime:  skill.ReleaseTime,
		ReleaseUser:  skill.ReleaseUser,
	}
}

func (r *skillRegistry) persistSkillAssets(ctx context.Context, skillID, version string, assets []*skillAsset) ([]*model.SkillFileIndexDB, error) {
	indices := make([]*model.SkillFileIndexDB, 0, len(assets))
	for _, asset := range assets {
		object, checksum, err := r.assetStore.Upload(ctx, skillID, version, asset.RelPath, asset.Content)
		if err != nil {
			return nil, err
		}
		indices = append(indices, &model.SkillFileIndexDB{
			SkillID:       skillID,
			SkillVersion:  version,
			RelPath:       asset.RelPath,
			PathHash:      utils.MD5(asset.RelPath),
			StorageID:     object.StorageID,
			StorageKey:    object.StorageKey,
			FileType:      asset.FileType,
			ContentSHA256: checksum,
			MimeType:      asset.MimeType,
			Size:          int64(len(asset.Content)),
		})
	}
	return indices, nil
}

func (r *skillRegistry) publishSkillSnapshot(ctx context.Context, tx *sql.Tx, skill *model.SkillRepositoryDB, userID string) error {
	now := time.Now().UnixNano()
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
		ReleaseTime:  now,
		ReleaseUser:  userID,
	}
	history := &model.SkillReleaseHistoryDB{
		SkillID:     skill.SkillID,
		Version:     skill.Version,
		ReleaseDesc: "",
		CreateTime:  now,
		CreateUser:  userID,
		UpdateTime:  now,
		UpdateUser:  userID,
	}
	release.SkillContent = skill.SkillContent
	history.SkillRelease = utils.ObjectToJSON(release)

	existingRelease, err := r.releaseRepo.SelectBySkillID(ctx, tx, skill.SkillID)
	if err != nil {
		return err
	}
	if existingRelease == nil {
		if err = r.releaseRepo.Insert(ctx, tx, release); err != nil {
			return err
		}
	} else {
		if err = r.releaseRepo.UpdateBySkillID(ctx, tx, release); err != nil {
			return err
		}
	}
	// 添加历史记录
	existingHistory, err := r.releaseHistoryRepo.SelectBySkillIDAndVersion(ctx, tx, skill.SkillID, skill.Version)
	if err != nil {
		return err
	}
	histories, err := r.releaseHistoryRepo.SelectBySkillID(ctx, tx, skill.SkillID)
	if err != nil {
		return err
	}
	if existingHistory != nil {
		if err = r.releaseHistoryRepo.DeleteByID(ctx, tx, existingHistory.ID); err != nil {
			return err
		}
	} else if len(histories) >= maxSkillReleaseHistoryVersions {
		recordsToDelete := len(histories) - maxSkillReleaseHistoryVersions + 1
		startIndex := len(histories) - recordsToDelete
		for i := startIndex; i < len(histories); i++ {
			if err = r.releaseHistoryRepo.DeleteByID(ctx, tx, histories[i].ID); err != nil {
				return err
			}
		}
	}
	if err = r.releaseHistoryRepo.Insert(ctx, tx, history); err != nil {
		return err
	}
	return nil
}

func (r *skillRegistry) deletePublishedSkillSnapshot(ctx context.Context, tx *sql.Tx, skillID string) error {
	release, err := r.releaseRepo.SelectBySkillID(ctx, tx, skillID)
	if err != nil {
		return err
	}
	if release == nil {
		return nil
	}
	if err = r.releaseRepo.DeleteBySkillID(ctx, tx, skillID); err != nil {
		return err
	}
	return nil
}

// ========== FR-6: OSS SKILL.md Frontmatter Rewrite ==========

// rewriteSkillMDFrontmatter 重写 OSS 中 SKILL.md 的 name/description
// 在 UpdateSkillMetadata 事务提交成功后调用
// 失败只记录日志，不阻塞主流程
func (r *skillRegistry) rewriteSkillMDFrontmatter(ctx context.Context, skillID, version, newName, newDesc string) error {
	skillFile, err := r.fileRepo.SelectSkillFileByPath(ctx, nil, skillID, version, SkillMD)
	if err != nil {
		return fmt.Errorf("query SKILL.md file_index failed: %w", err)
	}
	if skillFile == nil {
		return fmt.Errorf("SKILL.md not found in file_index: skill_id=%s, version=%s", skillID, version)
	}

	content, err := r.assetStore.Download(ctx, &interfaces.OssObject{
		StorageID:  skillFile.StorageID,
		StorageKey: skillFile.StorageKey,
	})
	if err != nil {
		return fmt.Errorf("download SKILL.md from OSS failed: %w", err)
	}

	newContent, err := updateFrontmatterNameDesc(string(content), newName, newDesc)
	if err != nil {
		return fmt.Errorf("update SKILL.md frontmatter failed: %w", err)
	}

	_, _, err = r.assetStore.Upload(ctx, skillID, version, SkillMD, []byte(newContent))
	if err != nil {
		return fmt.Errorf("upload rewritten SKILL.md to OSS failed: %w", err)
	}
	return nil
}

// updateFrontmatterNameDesc 只替换 YAML frontmatter 中的 name 和 description
// 其余所有自定义字段保持不动
func updateFrontmatterNameDesc(rawMD, newName, newDesc string) (string, error) {
	parts := strings.SplitN(rawMD, "---", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid SKILL.md format: missing frontmatter")
	}

	frontmatter := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(parts[1]), &frontmatter); err != nil {
		return "", fmt.Errorf("failed to unmarshal frontmatter: %w", err)
	}

	if newName != "" {
		frontmatter["name"] = newName
	}
	if newDesc != "" {
		frontmatter["description"] = newDesc
	}

	newFrontmatter, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	return "---\n" + string(newFrontmatter) + "---\n" + strings.TrimPrefix(parts[2], "\n"), nil
}
