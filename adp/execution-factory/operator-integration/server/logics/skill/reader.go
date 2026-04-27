package skill

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

type skillReader struct {
	skillRepo             model.ISkillRepository
	releaseRepo           model.ISkillReleaseDB
	releaseHistoryRepo    model.ISkillReleaseHistoryDB
	fileRepo              model.ISkillFileIndex
	assetStore            skillAssetStore
	AuthService           interfaces.IAuthorizationService
	BusinessDomainService interfaces.IBusinessDomainService
	Logger                interfaces.Logger
}

var (
	readerOnce sync.Once
	readerInst interfaces.SkillReader
)

// NewSkillReader 创建技能读取服务对象
func NewSkillReader() interfaces.SkillReader {
	readerOnce.Do(func() {
		conf := config.NewConfigLoader()
		readerInst = &skillReader{
			skillRepo:             dbaccess.NewSkillRepositoryDB(),
			releaseRepo:           dbaccess.NewSkillReleaseDB(),
			releaseHistoryRepo:    dbaccess.NewSkillReleaseHistoryDB(),
			fileRepo:              dbaccess.NewSkillFileIndexDB(),
			assetStore:            newOSSGatewaySkillAssetStore(),
			AuthService:           auth.NewAuthServiceImpl(),
			BusinessDomainService: business_domain.NewBusinessDomainService(),
			Logger:                conf.GetLogger(),
		}
	})
	return readerInst
}

// GetSkillContent 获取技能内容
func (r *skillReader) GetSkillContent(ctx context.Context, req *interfaces.GetSkillContentReq) (resp *interfaces.GetSkillContentResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id": req.SkillID,
	})

	skill, err := r.getPublishedSkill(ctx, req.SkillID)
	if err != nil {
		return
	}
	// 如果是外部接口
	if common.IsPublicAPIFromCtx(ctx) {
		// 有执行、查看、公开访问权限
		accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		authorized, err := r.AuthService.OperationCheckAny(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill,
			interfaces.AuthOperationTypeExecute, interfaces.AuthOperationTypePublicAccess, interfaces.AuthOperationTypeView)
		if err != nil {
			return nil, err
		}
		if !authorized {
			r.Logger.WithContext(ctx).Errorf("user has no permission to execute、view、public access skill %s", req.SkillID)
			err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden, fmt.Sprintf("user has no permission to execute、view、public access skill %s", req.SkillID))
			return nil, err
		}
	}
	// 查询对应的"SKILL.md文件
	skillFile, err := r.fileRepo.SelectSkillFileByPath(ctx, nil, skill.SkillID, skill.Version, SkillMD)
	if err != nil {
		r.Logger.WithContext(ctx).Errorf("select skill file failed: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, err
	}
	if skillFile == nil {
		err = errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill file not found: %s", SkillMD))
		return nil, err
	}
	contentObject := &interfaces.OssObject{
		StorageID:  skillFile.StorageID,
		StorageKey: skillFile.StorageKey,
	}
	downloadURL, err := r.assetStore.GetDownloadURL(ctx, contentObject)
	if err != nil {
		return nil, err
	}
	// TODO: 待接入审计日志
	return &interfaces.GetSkillContentResp{
		SkillID: skill.SkillID,
		URL:     downloadURL,
		Files:   utils.JSONToObject[[]*interfaces.SkillFileSummary](skill.FileManifest),
		Status:  interfaces.BizStatus(skill.Status),
	}, nil
}

// ReadSkillFile 读取技能文件内容
func (r *skillReader) ReadSkillFile(ctx context.Context, req *interfaces.ReadSkillFileReq) (resp *interfaces.ReadSkillFileResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"user_id":  req.UserID,
		"bd_id":    req.BusinessDomainID,
		"skill_id": req.SkillID,
		"rel_path": req.RelPath,
	})
	skill, err := r.getPublishedSkill(ctx, req.SkillID)
	if err != nil {
		r.Logger.WithContext(ctx).Errorf("read skill file failed: %v", err)
		return nil, err
	}
	if common.IsPublicAPIFromCtx(ctx) {
		// 有执行、查看、公开访问权限
		accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		authorized, err := r.AuthService.OperationCheckAny(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill,
			interfaces.AuthOperationTypeExecute, interfaces.AuthOperationTypePublicAccess, interfaces.AuthOperationTypeView)
		if err != nil {
			return nil, err
		}
		if !authorized {
			r.Logger.WithContext(ctx).Errorf("user %s has no permission to execute skill %s", req.UserID, req.SkillID)
			err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden, fmt.Sprintf("user %s has no permission to execute skill %s", req.UserID, req.SkillID))
			return nil, err
		}
	}
	relPath, err := normalizeZipPath(req.RelPath)
	if err != nil {
		return nil, err
	}
	file, err := r.fileRepo.SelectSkillFileByPath(ctx, nil, req.SkillID, skill.Version, relPath)
	if err != nil {
		r.Logger.WithContext(ctx).Errorf("read skill file failed: %v", err)
		return nil, err
	}
	if file == nil {
		r.Logger.WithContext(ctx).Warnf("skill file not found: %s", relPath)
		err = errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill file not found: %s", relPath))
		return nil, err
	}
	downloadURL, err := r.assetStore.GetDownloadURL(ctx, &interfaces.OssObject{
		StorageID:  file.StorageID,
		StorageKey: file.StorageKey,
	})
	if err != nil {
		r.Logger.WithContext(ctx).Errorf("read skill file failed: %v", err)
		return nil, err
	}

	return &interfaces.ReadSkillFileResp{
		SkillID:  req.SkillID,
		RelPath:  relPath,
		URL:      downloadURL,
		MimeType: file.MimeType,
		FileType: file.FileType,
	}, nil
}

// GetSkillReleaseHistory 查询 Skill 发布历史
func (r *skillReader) GetSkillReleaseHistory(ctx context.Context, req *interfaces.GetSkillReleaseHistoryReq) (resp []*interfaces.SkillReleaseHistoryInfo, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id": req.SkillID,
	})

	histories, err := r.releaseHistoryRepo.SelectBySkillID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	if len(histories) == 0 {
		return []*interfaces.SkillReleaseHistoryInfo{}, nil
	}
	resp = make([]*interfaces.SkillReleaseHistoryInfo, 0, len(histories))
	for _, history := range histories {
		release := &model.SkillReleaseDB{}
		if history.SkillRelease != "" {
			release = utils.JSONToObject[*model.SkillReleaseDB](history.SkillRelease)
		}
		if release == nil {
			release = &model.SkillReleaseDB{}
		}
		resp = append(resp, &interfaces.SkillReleaseHistoryInfo{
			SkillID:     history.SkillID,
			Name:        release.Name,
			Description: release.Description,
			Version:     history.Version,
			Status:      interfaces.BizStatus(release.Status),
			Category:    interfaces.BizCategory(release.Category),
			Source:      release.Source,
			ReleaseDesc: history.ReleaseDesc,
			ReleaseUser: release.ReleaseUser,
			ReleaseTime: release.ReleaseTime,
			CreateUser:  release.CreateUser,
			CreateTime:  release.CreateTime,
			UpdateUser:  release.UpdateUser,
			UpdateTime:  release.UpdateTime,
		})
	}
	return resp, nil
}

func (r *skillReader) getPublishedSkill(ctx context.Context, skillID string) (*model.SkillRepositoryDB, error) {
	release, err := r.releaseRepo.SelectBySkillID(ctx, nil, skillID)
	if err != nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	if release == nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", skillID))
	}
	return &model.SkillRepositoryDB{
		SkillID:      release.SkillID,
		Name:         release.Name,
		Description:  release.Description,
		SkillContent: release.SkillContent,
		Version:      release.Version,
		Status:       release.Status,
		Source:       release.Source,
		Dependencies: release.Dependencies,
		ExtendInfo:   release.ExtendInfo,
		FileManifest: release.FileManifest,
		CreateUser:   release.CreateUser,
		CreateTime:   release.CreateTime,
		UpdateUser:   release.UpdateUser,
		UpdateTime:   release.UpdateTime,
		Category:     release.Category,
	}, nil
}
