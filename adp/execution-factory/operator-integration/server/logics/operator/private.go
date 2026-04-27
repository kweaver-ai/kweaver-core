package operator

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// CheckAddAsTool 检查算子是否允许添加为工具
func (m *operatorManager) CheckAddAsTool(ctx context.Context, operatorID, userID string) (resp *interfaces.CheckAddAsToolResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	var accessor *interfaces.AuthAccessor
	accessor, err = m.AuthService.GetAccessor(ctx, userID)
	if err != nil {
		return
	}
	// 校验是否拥有算子的公开访问和使用权限
	var authorized bool
	authorized, err = m.AuthService.OperationCheckAll(ctx, accessor, operatorID, interfaces.AuthResourceTypeOperator,
		interfaces.AuthOperationTypePublicAccess, interfaces.AuthOperationTypeExecute)
	if err != nil {
		return
	}
	if !authorized {
		err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonUseForbidden, nil)
		return
	}
	// 查询算子发布信息
	exist, releaseDB, err := m.OpReleaseDB.SelectByOpID(ctx, operatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("query operator release failed, err: %v", err)
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "query operator release failed")
	}
	if !exist {
		return nil, errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtOperatorNotFound, "operator release not found")
	}
	// 检查算子是否可用
	if releaseDB.Status != interfaces.BizStatusPublished.String() {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtOperatorNotAvailable,
			fmt.Sprintf("operator %s is not available", releaseDB.OpID), releaseDB.Name)
		return
	}
	// 检查是否是同步执行
	if releaseDB.ExecutionMode != interfaces.ExecutionModeSync.String() {
		// 仅支持同步算子转换为工具
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolConvertOnlySupportSync,
			"only sync operators can be published as tools ")
		return
	}
	// 获取元数据信息
	metadataDB, err := m.MetadataService.GetMetadataByVersion(ctx, interfaces.MetadataType(releaseDB.MetadataType), releaseDB.MetadataVersion)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("query metadata failed, err: %v", err)
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "query metadata failed")
	}
	resp = &interfaces.CheckAddAsToolResp{
		OperatorID: releaseDB.OpID,
		Name:       releaseDB.Name,
		Metadata:   metadataDB,
	}
	return
}
