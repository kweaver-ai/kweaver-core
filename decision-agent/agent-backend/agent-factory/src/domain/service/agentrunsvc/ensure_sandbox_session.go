package agentsvc

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// EnsureSandboxSession 确保 Sandbox Session 存在并就绪
// 完全移除 sync.Map 缓存，每次直接调用 Sandbox Platform 检测
func (s *agentSvc) EnsureSandboxSession(ctx context.Context, sessionID string, req *agentreq.ChatReq) (string, error) {
	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, nil)

	// o11y.String("user_id", req.UserID),
	// o11y.String("agent_id", req.AgentID),

	// 1. 检测 Session 状态
	sessionInfo, err := s.sandboxPlatform.GetSession(ctx, sessionID)
	if err != nil {
		// 404 错误表示 Session 不存在，继续创建
		if s.isSessionNotFoundError(err) {
			// o11y.SetAttributes(ctx, o11y.String("action", "create_new"))
			return s.createNewSession(ctx, sessionID, req)
		}

		// 其他错误：尝试创建新 Session
		// o11y.SetAttributes(ctx, o11y.String("action", "recover_from_error"))
		s.logger.Warnf("[EnsureSandboxSession] get session failed: %v, will create new session", err)

		return s.createNewSession(ctx, sessionID, req)
	}

	// 2. 检查 Session 状态
	if sessionInfo.Status == "running" {
		// o11y.SetAttributes(ctx, o11y.String("action", "reuse_existing"))
		s.logger.Infof("[EnsureSandboxSession] reuse existing session: %s", sessionID)
		return sessionID, nil
	}

	// 3. Session 状态为 failed 或 error，先删除再创建
	if sessionInfo.Status == "failed" || sessionInfo.Status == "error" || sessionInfo.Status == "stopped" {
		// o11y.SetAttributes(ctx, o11y.String("action", "delete_and_recreate"))
		s.logger.Warnf("[EnsureSandboxSession] session status is %s, will delete and recreate: %s", sessionInfo.Status, sessionID)

		if delErr := s.sandboxPlatform.DeleteSession(ctx, sessionID); delErr != nil {
			s.logger.Errorf("[EnsureSandboxSession] delete session failed: %v", delErr)
		}

		return s.createNewSession(ctx, sessionID, req)
	}

	// 4. Session 状态非 running，自动重新创建
	// o11y.SetAttributes(ctx, o11y.String("action", "recreate"))
	s.logger.Warnf("[EnsureSandboxSession] session status is %s, will recreate: %s", sessionInfo.Status, sessionID)

	return s.createNewSession(ctx, sessionID, req)
}

// createNewSession 创建新的 Sandbox Session
func (s *agentSvc) createNewSession(ctx context.Context, sessionID string, req *agentreq.ChatReq) (string, error) {
	cpu := s.sandboxPlatformConf.DefaultCPU
	if cpu == "" {
		cpu = "1"
	}

	memory := s.sandboxPlatformConf.DefaultMemory
	if memory == "" {
		memory = "512Mi"
	}

	disk := s.sandboxPlatformConf.DefaultDisk
	if disk == "" {
		disk = "1Gi"
	}

	timeout := s.sandboxPlatformConf.DefaultTimeout
	if timeout == 0 {
		timeout = 300
	}

	createReq := sandboxplatformdto.CreateSessionReq{
		ID:         &sessionID,
		TemplateID: s.sandboxPlatformConf.DefaultTemplateID,
		Timeout:    int(timeout),
		CPU:        cpu,
		Memory:     memory,
		Disk:       disk,
		Event: map[string]interface{}{
			"session_id":         sessionID,
			"user_id":            req.UserID,
			"agent_id":           req.AgentID,
			"business_domain_id": req.XBusinessDomainID,
			"file_upload_config": map[string]interface{}{
				"max_file_size":      s.sandboxPlatformConf.DefaultFileUploadConfig.MaxFileSize,
				"max_file_size_unit": s.sandboxPlatformConf.DefaultFileUploadConfig.MaxFileSizeUnit,
				"max_file_count":     s.sandboxPlatformConf.DefaultFileUploadConfig.MaxFileCount,
				"allowed_file_types": s.sandboxPlatformConf.DefaultFileUploadConfig.AllowedFileTypes,
			},
		},
	}

	createResp, err := s.sandboxPlatform.CreateSession(ctx, createReq)
	if err != nil {
		if s.isSessionAlreadyExistsError(err) {
			s.logger.Infof("[createNewSession] session already exists: %s, will wait for ready", sessionID)
			return s.waitForSessionReady(ctx, sessionID)
		}

		s.logger.Errorf("[createNewSession] create failed: %v", err)

		return "", errors.Wrap(err, "create sandbox session failed")
	}

	actualSessionID := createResp.ID
	if createResp.ID == "" {
		actualSessionID = sessionID
	}

	return s.waitForSessionReady(ctx, actualSessionID)
}

// waitForSessionReady 等待 Session 就绪
func (s *agentSvc) waitForSessionReady(ctx context.Context, sessionID string) (string, error) {
	maxRetries := s.sandboxPlatformConf.MaxRetries
	retryInterval := s.sandboxPlatformConf.RetryInterval

	retryIntervalDuration, err := time.ParseDuration(retryInterval)
	if err != nil {
		s.logger.Warnf("[waitForSessionReady] failed to parse retry interval, using default 500ms")

		retryIntervalDuration = 500 * time.Millisecond
	}

	for i := 0; i < maxRetries; i++ {
		sessionInfo, err := s.sandboxPlatform.GetSession(ctx, sessionID)
		if err != nil {
			s.logger.Errorf("[waitForSessionReady] get session status failed (attempt %d): %v", i+1, err)
			time.Sleep(retryIntervalDuration)

			continue
		}

		if sessionInfo.Status == "running" {
			s.logger.Infof("[waitForSessionReady] session ready: %s (attempts: %d)", sessionID, i+1)
			return sessionID, nil
		}

		// 如果状态是 error/stopped，直接失败
		if sessionInfo.Status == "error" || sessionInfo.Status == "stopped" {
			return "", errors.Errorf("session in invalid state: %s", sessionInfo.Status)
		}

		time.Sleep(retryIntervalDuration)
	}

	return "", errors.New("timeout waiting for session ready")
}

// isSessionNotFoundError 判断是否为 Session 不存在的错误
func (s *agentSvc) isSessionNotFoundError(err error) bool {
	// 检查是否为 rest.HTTPError 类型
	var httpErr *rest.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.HTTPCode == http.StatusNotFound
	}

	return false
}

// isSessionAlreadyExistsError 判断是否为 Session 已存在的错误
func (s *agentSvc) isSessionAlreadyExistsError(err error) bool {
	var httpErr *rest.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.HTTPCode == http.StatusConflict
	}

	return strings.Contains(err.Error(), "already exists")
}
