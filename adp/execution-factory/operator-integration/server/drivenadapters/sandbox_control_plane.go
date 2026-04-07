package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

// 沙箱控制服务Client
type sandBoxControlPlaneClient struct {
	baseURL    string
	logger     interfaces.Logger
	httpClient interfaces.HTTPClient
	templateID string // 模版ID
}

var (
	sbcpInstance *sandBoxControlPlaneClient // 沙箱控制服务Client实例
	sbcpOnce     sync.Once                  // 沙箱控制服务Client实例初始化Once
)

// NewSandBoxControlPlaneClient 创建沙箱控制服务Client实例
func NewSandBoxControlPlaneClient() interfaces.SandBoxControlPlane {
	sbcpOnce.Do(func() {
		conf := config.NewConfigLoader()
		sbcpInstance = &sandBoxControlPlaneClient{
			baseURL: fmt.Sprintf("%s://%s:%d/api/v1", conf.SandboxControlPlane.PrivateProtocol,
				conf.SandboxControlPlane.PrivateHost, conf.SandboxControlPlane.PrivatePort),
			logger:     conf.GetLogger(),
			httpClient: rest.NewHTTPClient(),
			templateID: conf.SandboxControlPlane.PrivateHost,
		}
	})
	return sbcpInstance
}

// GetTemplateDetail 获取模版详情
func (c *sandBoxControlPlaneClient) GetTemplateDetail(ctx context.Context, tempID string) (any, error) {
	src := fmt.Sprintf("%s/templates/%s", c.baseURL, tempID)
	headers := common.GetHeaderFromCtx(ctx)
	_, respData, err := c.httpClient.Get(ctx, src, nil, headers)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("GetTemplateDetail failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, map[string]any{
			"error":       err.Error(),
			"template_id": tempID,
		})
		return nil, err
	}
	return respData, nil
}

// CreateSession 创建会话
func (c *sandBoxControlPlaneClient) CreateSession(ctx context.Context, req *interfaces.CreateSessionReq) (any, error) {
	src := fmt.Sprintf("%s/sessions", c.baseURL)
	headers := common.GetHeaderFromCtx(ctx)
	_, respData, err := c.httpClient.Post(ctx, src, headers, req)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("CreateSession failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, map[string]any{
			"error":      err.Error(),
			"session_id": req.ID,
		})
		return nil, err
	}
	return respData, nil
}

// QuerySession 查询会话
func (c *sandBoxControlPlaneClient) QuerySession(ctx context.Context, sessionID string) (exists bool, detail *interfaces.SessionDetail, err error) {
	src := fmt.Sprintf("%s/sessions/%s", c.baseURL, sessionID)
	headers := common.GetHeaderFromCtx(ctx)
	respCode, respData, err := c.httpClient.GetNoUnmarshal(ctx, src, nil, headers)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("QuerySession failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtSandboxControlPlaneFailed, map[string]any{
			"error":      err.Error(),
			"session_id": sessionID,
		})
		return false, nil, err
	}
	if respCode == http.StatusNotFound {
		c.logger.WithContext(ctx).Infof("QuerySession failed, session not found, session_id: %s", sessionID)
		return false, nil, nil
	}

	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		c.logger.WithContext(ctx).Errorf("QuerySession failed, unexpected status code: %d, session_id: %s", respCode, sessionID)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, map[string]any{
			"error":      fmt.Sprintf("unexpected status code: %d", respCode),
			"session_id": sessionID,
			"response":   string(respData),
			"http_code":  respCode,
		})
		return false, nil, err
	}
	detail = &interfaces.SessionDetail{
		RequestedDependencies: []*interfaces.DependencyInfo{},
		InstalledDependencies: []*interfaces.DependencyInfo{},
	}
	err = utils.StringToObject(string(respData), detail)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("QuerySession failed, StringToObject failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, map[string]any{
			"error":      err.Error(),
			"session_id": sessionID,
		})
		return false, nil, err
	}
	return true, detail, nil
}

// DeleteSession 删除会话
func (c *sandBoxControlPlaneClient) DeleteSession(ctx context.Context, sessionID string) (err error) {
	src := fmt.Sprintf("%s/sessions/%s", c.baseURL, sessionID)
	headers := common.GetHeaderFromCtx(ctx)
	respCode, respData, err := c.httpClient.Delete(ctx, src, headers)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("DeleteSession failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, map[string]any{
			"error":      err.Error(),
			"session_id": sessionID,
		})
		return err
	}
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		c.logger.WithContext(ctx).Errorf("DeleteSession failed, unexpected status code: %d, session_id: %s", respCode, sessionID)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, map[string]any{
			"error":      fmt.Sprintf("unexpected status code: %d", respCode),
			"session_id": sessionID,
			"http_code":  respCode,
			"response":   respData,
		})
		return err
	}
	return nil
}

// ListSessions 列举会话
func (c *sandBoxControlPlaneClient) ListSessions(ctx context.Context, req *interfaces.ListSessionsReq) (resp *interfaces.ListSessionsResp, err error) {
	src := fmt.Sprintf("%s/sessions", c.baseURL)
	headers := common.GetHeaderFromCtx(ctx)
	query := url.Values{}
	if req.Limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", req.Limit))
	}
	if req.Offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", req.Offset))
	}
	if req.Status != "" {
		query.Add("status", string(req.Status))
	}
	respCode, respData, err := c.httpClient.GetNoUnmarshal(ctx, src, query, headers)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("ListSessions failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, err.Error())
		return nil, err
	}
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		c.logger.WithContext(ctx).Errorf("ListSessions failed, unexpected status code: %d", respCode)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, map[string]any{
			"error":     fmt.Sprintf("unexpected status code: %d", respCode),
			"http_code": respCode,
			"response":  respData,
		})
		return nil, err
	}
	resp = &interfaces.ListSessionsResp{}
	err = utils.StringToObject(string(respData), resp)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("ListSessions failed, StringToObject failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, err.Error())
		return nil, err
	}
	return resp, nil
}

// ExecuteCodeSync 执行函数(同步)
func (c *sandBoxControlPlaneClient) ExecuteCodeSync(ctx context.Context, sessionID string, req *interfaces.ExecuteCodeReq) (*interfaces.ExecuteCodeResp, error) {
	src := fmt.Sprintf("%s/executions/sessions/%s/execute-sync", c.baseURL, sessionID)
	headers := common.GetHeaderFromCtx(ctx)
	respCode, respData, err := c.httpClient.PostNoUnmarshal(ctx, src, headers, req)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("ExecuteCodeSync failed, err: %v", err)
		return nil, err
	}
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		c.logger.WithContext(ctx).Errorf("ExecuteCodeSync failed, unexpected status code: %d, session_id: %s", respCode, sessionID)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, map[string]any{
			"error":     fmt.Sprintf("unexpected status code: %d", respCode),
			"http_code": respCode,
			"response":  respData,
		})
		return nil, err
	}
	resp := &interfaces.ExecuteCodeResp{}
	err = utils.StringToObject(string(respData), resp)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("ExecuteCodeSync failed, StringToObject failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, err.Error())
		return nil, err
	}
	return resp, nil
}

// InstallPythonDependencies 安装python依赖库
func (c *sandBoxControlPlaneClient) InstallPythonDependencies(ctx context.Context, sessionID string, req *interfaces.InstallDependenciesReq) (detail *interfaces.SessionDetail, err error) {
	src := fmt.Sprintf("%s/sessions/%s/dependencies/install", c.baseURL, sessionID)
	headers := common.GetHeaderFromCtx(ctx)
	respCode, respData, err := c.httpClient.PostNoUnmarshal(ctx, src, headers, req)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("InstallPythonDependencies failed, err: %v", err)
		return nil, err
	}
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		c.logger.WithContext(ctx).Errorf("InstallPythonDependencies failed, unexpected status code: %d, session_id: %s", respCode, sessionID)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, map[string]any{
			"error":          fmt.Sprintf("unexpected status code: %d", respCode),
			"http_code":      respCode,
			"response":       string(respData),
			"session_id":     sessionID,
			"request_params": utils.ObjectToJSON(req),
		})
		return nil, err
	}
	detail = &interfaces.SessionDetail{}
	err = utils.StringToObject(string(respData), detail)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("InstallPythonDependencies failed, StringToObject failed, err: %v", err)
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtSandboxControlPlaneFailed, err.Error())
		return nil, err
	}
	return detail, nil
}
