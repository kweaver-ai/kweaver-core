package skill

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
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

type skillManagementReader struct {
	skillRepo             model.ISkillRepository
	fileRepo              model.ISkillFileIndex
	assetStore            skillAssetStore
	AuthService           interfaces.IAuthorizationService
	BusinessDomainService interfaces.IBusinessDomainService
	Logger                interfaces.Logger
}

var (
	mgmtReaderOnce sync.Once
	mgmtReaderInst interfaces.SkillManagementReader
)

// NewSkillManagementReader 创建管理态技能读取服务
func NewSkillManagementReader() interfaces.SkillManagementReader {
	mgmtReaderOnce.Do(func() {
		conf := config.NewConfigLoader()
		mgmtReaderInst = &skillManagementReader{
			skillRepo:             dbaccess.NewSkillRepositoryDB(),
			fileRepo:              dbaccess.NewSkillFileIndexDB(),
			assetStore:            newOSSGatewaySkillAssetStore(),
			AuthService:           auth.NewAuthServiceImpl(),
			BusinessDomainService: business_domain.NewBusinessDomainService(),
			Logger:                conf.GetLogger(),
		}
	})
	return mgmtReaderInst
}

// GetManagementContent 获取管理态 SKILL.md 内容
func (r *skillManagementReader) GetManagementContent(ctx context.Context, req *interfaces.GetManagementContentReq) (resp *interfaces.GetManagementContentResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id": req.SkillID,
	})

	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.IsDeleted {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", req.SkillID))
	}

	if common.IsPublicAPIFromCtx(ctx) {
		accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		authorized, err := r.AuthService.OperationCheckAny(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill,
			interfaces.AuthOperationTypeView, interfaces.AuthOperationTypeModify)
		if err != nil {
			return nil, err
		}
		if !authorized {
			r.Logger.WithContext(ctx).Errorf("user has no permission to view/modify skill %s", req.SkillID)
			return nil, errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden,
				fmt.Sprintf("user has no permission to view/modify skill %s", req.SkillID))
		}
	}

	resp = &interfaces.GetManagementContentResp{
		SkillID:     skill.SkillID,
		Name:        skill.Name,
		Description: skill.Description,
		Version:     skill.Version,
		Status:      interfaces.BizStatus(skill.Status),
		Source:      skill.Source,
		FileType:    detectSkillFileType(skill),
	}

	resp.Files = utils.JSONToObject[[]*interfaces.SkillFileSummary](skill.FileManifest)
	if resp.Files == nil {
		resp.Files = []*interfaces.SkillFileSummary{}
	}

	// 统一查询 SKILL.md 的 OSS 记录，用于 URL 和 content 模式
	skillFile, err := r.fileRepo.SelectSkillFileByPath(ctx, nil, skill.SkillID, skill.Version, SkillMD)
	if err != nil {
		return nil, err
	}
	if skillFile != nil {
		downloadURL, err := r.assetStore.GetDownloadURL(ctx, &interfaces.OssObject{
			StorageID:  skillFile.StorageID,
			StorageKey: skillFile.StorageKey,
		})
		if err != nil {
			return nil, err
		}
		resp.URL = downloadURL
	}

	// 根据 response_mode 填充 Content:
	//   url(默认) — Content 为空，客户端通过 URL 下载
	//   content  — 优先用 DB 中的 skill_content，zip 类型从 OSS 下载 SKILL.md
	//   auto     — 仅 content 注册类型填充（向后兼容）
	switch req.ResponseMode {
	case "content":
		if skill.SkillContent != "" {
			resp.Content = skill.SkillContent
		} else if skillFile != nil {
			ossContent, err := r.assetStore.Download(ctx, &interfaces.OssObject{
				StorageID:  skillFile.StorageID,
				StorageKey: skillFile.StorageKey,
			})
			if err != nil {
				r.Logger.WithContext(ctx).Errorf("download SKILL.md from OSS failed: %v", err)
			} else {
				resp.Content = string(ossContent)
			}
		}
	case "auto":
		if resp.FileType == "content" {
			resp.Content = skill.SkillContent
		}
		// default(""/"url"): Content 保持零值，omitempty 跳过
	}

	return resp, nil
}

// ReadManagementFile 读取管理态指定文件
func (r *skillManagementReader) ReadManagementFile(ctx context.Context, req *interfaces.ReadManagementFileReq) (resp *interfaces.ReadManagementFileResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id": req.SkillID,
		"rel_path": req.RelPath,
	})

	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.IsDeleted {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", req.SkillID))
	}

	if common.IsPublicAPIFromCtx(ctx) {
		accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		authorized, err := r.AuthService.OperationCheckAny(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill,
			interfaces.AuthOperationTypeView, interfaces.AuthOperationTypeModify)
		if err != nil {
			return nil, err
		}
		if !authorized {
			r.Logger.WithContext(ctx).Errorf("user has no permission to view/modify skill %s", req.SkillID)
			return nil, errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden,
				fmt.Sprintf("user has no permission to view/modify skill %s", req.SkillID))
		}
	}

	relPath, err := normalizeZipPath(req.RelPath)
	if err != nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
	}

	file, err := r.fileRepo.SelectSkillFileByPath(ctx, nil, req.SkillID, skill.Version, relPath)
	if err != nil {
		r.Logger.WithContext(ctx).Errorf("read management file failed: %v", err)
		return nil, err
	}
	if file == nil {
		r.Logger.WithContext(ctx).Warnf("management file not found: %s", relPath)
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("file not found: %s", relPath))
	}

	downloadURL, err := r.assetStore.GetDownloadURL(ctx, &interfaces.OssObject{
		StorageID:  file.StorageID,
		StorageKey: file.StorageKey,
	})
	if err != nil {
		r.Logger.WithContext(ctx).Errorf("read management file failed: %v", err)
		return nil, err
	}

	return &interfaces.ReadManagementFileResp{
		SkillID:  req.SkillID,
		RelPath:  relPath,
		URL:      downloadURL,
		MimeType: file.MimeType,
		FileType: file.FileType,
		Size:     file.Size,
	}, nil
}

// DownloadManagementSkill 下载管理态完整技能包
func (r *skillManagementReader) DownloadManagementSkill(ctx context.Context, req *interfaces.DownloadManagementSkillReq) (resp *interfaces.DownloadSkillResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"skill_id": req.SkillID,
	})

	skill, err := r.skillRepo.SelectSkillByID(ctx, nil, req.SkillID)
	if err != nil {
		return nil, err
	}
	if skill == nil || skill.IsDeleted {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("skill not found: %s", req.SkillID))
	}

	if common.IsPublicAPIFromCtx(ctx) {
		accessor, err := r.AuthService.GetAccessor(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		authorized, err := r.AuthService.OperationCheckAny(ctx, accessor, req.SkillID, interfaces.AuthResourceTypeSkill,
			interfaces.AuthOperationTypeView, interfaces.AuthOperationTypeModify)
		if err != nil {
			return nil, err
		}
		if !authorized {
			r.Logger.WithContext(ctx).Errorf("user has no permission to view/modify skill %s", req.SkillID)
			return nil, errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden,
				fmt.Sprintf("user has no permission to view/modify skill %s", req.SkillID))
		}
	}

	files, err := r.fileRepo.SelectSkillFileBySkillID(ctx, nil, req.SkillID, skill.Version)
	if err != nil {
		r.Logger.WithContext(ctx).Errorf("select management files failed: %v", err)
		return nil, err
	}

	skill, zipName, content, err := buildArchiveFromFiles(ctx, r.assetStore, skill, files)
	if err != nil {
		r.Logger.WithContext(ctx).Errorf("build management archive failed: %v", err)
		return nil, err
	}

	return &interfaces.DownloadSkillResp{
		SkillID:  req.SkillID,
		FileName: zipName,
		Content:  content,
	}, nil
}

// buildArchiveFromFiles 从文件列表构建 ZIP 归档
func buildArchiveFromFiles(ctx context.Context, store skillAssetStore, skill *model.SkillRepositoryDB,
	files []*model.SkillFileIndexDB) (*model.SkillRepositoryDB, string, []byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, file := range files {
		content, readErr := store.Download(ctx, &interfaces.OssObject{
			StorageID:  file.StorageID,
			StorageKey: file.StorageKey,
		})
		if readErr != nil {
			_ = zw.Close()
			return nil, "", nil, readErr
		}
		w, createErr := zw.Create(file.RelPath)
		if createErr != nil {
			_ = zw.Close()
			return nil, "", nil, createErr
		}
		if _, writeErr := io.Copy(w, bytes.NewReader(content)); writeErr != nil {
			_ = zw.Close()
			return nil, "", nil, writeErr
		}
	}
	if err := zw.Close(); err != nil {
		return nil, "", nil, err
	}
	return skill, fmt.Sprintf("%s.zip", skill.Name), buf.Bytes(), nil
}

// detectSkillFileType 从 repository 记录推断注册类型
// FR-5: content 注册的 manifest 仅有 SKILL.md 一条记录，zip 注册有更多文件
func detectSkillFileType(skill *model.SkillRepositoryDB) string {
	manifest := utils.JSONToObject[[]*interfaces.SkillFileSummary](skill.FileManifest)
	if len(manifest) == 0 {
		return "content"
	}
	if len(manifest) == 1 && manifest[0].RelPath == SkillMD && skill.SkillContent != "" {
		return "content"
	}
	return "zip"
}
