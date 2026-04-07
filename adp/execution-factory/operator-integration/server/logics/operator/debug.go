package operator

import (
	"context"
	"fmt"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// DebugOperator 调试接口
func (m *operatorManager) DebugOperator(ctx context.Context, req *interfaces.DebugOperatorReq) (resp *interfaces.HTTPResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 检查算子是否存在
	exist, operator, err := m.DBOperatorManager.SelectByOperatorID(ctx, nil, req.OperatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("select operator by id failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exist {
		m.Logger.WithContext(ctx).Warnf("operator not exist, operator_id: %s", req.OperatorID)
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtOperatorNotFound, "operator not exist")
		return
	}
	// 校验使用权限
	accessor, err := m.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = m.AuthService.CheckExecutePermission(ctx, accessor, req.OperatorID, interfaces.AuthResourceTypeOperator)
	if err != nil {
		return
	}
	executionMode := operator.ExecutionMode
	if operator.MetadataVersion != req.Version {
		// 查询历史记录
		var historyDB *model.OperatorReleaseHistoryDB
		exist, historyDB, err = m.OpReleaseHistoryDB.SelectByOpIDAndMetdata(ctx, req.OperatorID, req.Version)
		if err != nil {
			m.Logger.WithContext(ctx).Warnf("select operator release history by id failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		if !exist {
			m.Logger.WithContext(ctx).Warnf("operator not exist, operator_id: %s, version: %s", req.OperatorID, req.Version)
			err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtOperatorNotFound, "operator not exist")
			return
		}
		releaseDB := &model.OperatorReleaseDB{}
		err = jsoniter.Unmarshal([]byte(historyDB.OpRelease), releaseDB)
		if err != nil {
			m.Logger.WithContext(ctx).Warnf("unmarshal operator release failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		executionMode = releaseDB.ExecutionMode
	}

	// 检查执行模式
	if executionMode != string(interfaces.ExecutionModeSync) {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtOnlySyncModeDebug, fmt.Sprintf("operator execution mode is %s, not supported", operator.ExecutionMode))
		return
	}
	if req.Timeout == 0 {
		executeControl := &interfaces.OperatorExecuteControl{}
		_ = utils.StringToObject(operator.ExecuteControl, executeControl)
		req.Timeout = int(executeControl.Timeout)
	}
	resp, err = m.executeOperator(ctx, req.OperatorID, req.HTTPRequestParams,
		interfaces.MetadataType(operator.MetadataType), req.Version, int64(req.Timeout))
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("execute operator failed, err: %v", err)
		return
	}
	// 记录日志
	go func() {
		accountAuth, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			m.Logger.WithContext(ctx).Errorf("get account auth context from ctx failed")
			return
		}
		m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuth.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationExecute,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectOperator,
				ID:   operator.OperatorID,
				Name: operator.Name,
			},
		})
	}()
	return
}

// ExecuteOperator 执行算子
func (m *operatorManager) ExecuteOperator(ctx context.Context, req *interfaces.ExecuteOperatorReq) (resp *interfaces.HTTPResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 检查算子是否存在
	exist, operator, err := m.OpReleaseDB.SelectByOpID(ctx, req.OperatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("select release operator by id failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 检查算子状态
	if !exist || operator.Status != string(interfaces.BizStatusPublished) {
		m.Logger.WithContext(ctx).Warnf("release operator not exist or not published, operator_id: %s", req.OperatorID)
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtOperatorNotFound, "release operator not exist or not published")
		return
	}
	// 校验使用权限
	accessor, err := m.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = m.AuthService.CheckExecutePermission(ctx, accessor, req.OperatorID, interfaces.AuthResourceTypeOperator)
	if err != nil {
		return
	}
	executionMode := operator.ExecutionMode
	// 检查执行模式
	if executionMode != string(interfaces.ExecutionModeSync) {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtOnlySyncModeDebug, fmt.Sprintf("operator execution mode is %s, not supported", operator.ExecutionMode))
		return
	}
	if req.Timeout == 0 {
		executeControl := &interfaces.OperatorExecuteControl{}
		_ = utils.StringToObject(operator.ExecuteControl, executeControl)
		req.Timeout = int(executeControl.Timeout)
	}
	resp, err = m.executeOperator(ctx, req.OperatorID, req.HTTPRequestParams,
		interfaces.MetadataType(operator.MetadataType), operator.MetadataVersion, int64(req.Timeout))
	if err != nil {
		return
	}
	// 记录日志
	go func() {
		accountAuth, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			m.Logger.WithContext(ctx).Errorf("get account auth context from ctx failed")
			return
		}
		m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuth.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationExecute,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectOperator,
				ID:   operator.OpID,
				Name: operator.Name,
			},
		})
	}()
	return
}

func (m *operatorManager) executeOperator(ctx context.Context, operatorID string, reqParam interfaces.HTTPRequestParams,
	metadataType interfaces.MetadataType, metadataVersion string, timeout int64) (resp *interfaces.HTTPResponse, err error) {
	// 获取元数据
	exists, metadataDB, err := m.MetadataService.CheckMetadataExists(ctx, metadataType, metadataVersion)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("select metadata by id failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtMetadataNotFound, fmt.Sprintf("metadata not found, type: %s, version: %s", metadataType, metadataVersion))
		return
	}
	// 执行算子
	url := fmt.Sprintf("%s%s", metadataDB.GetServerURL(), metadataDB.GetPath())
	resp, err = m.Proxy.HandlerRequest(ctx, &interfaces.HTTPRequest{
		ClientID: operatorID,
		HTTPRouter: interfaces.HTTPRouter{
			URL:    url,
			Method: metadataDB.GetMethod(),
		},
		HTTPRequestParams: reqParam,
		ExecutionMode:     interfaces.ExecutionModeSync,
		Timeout:           time.Duration(timeout) * time.Second,
	})
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("handler request failed, err: %v", err)
	}
	return
}
