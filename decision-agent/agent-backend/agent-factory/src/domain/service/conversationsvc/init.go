package conversationsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

func (sv *conversationSvc) Init(ctx context.Context, req conversationreq.InitReq) (rt conversationresp.InitConversationResp, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)

	if req.Title == "" {
		req.Title = "新会话"
	}

	po := &dapo.ConversationPO{
		AgentAPPKey: req.AgentAPPKey,
		CreateBy:    req.UserID,
		UpdateBy:    req.UserID,
		Title:       req.Title,
		Ext:         new(string),
	}

	po, err = sv.conversationRepo.Create(ctx, po)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[Init] create conversation error, err: %v", err))
		return rt, errors.Wrapf(err, "update conversation title failed")
	}

	// 确保 Sandbox Session 存在并就绪（仅在启用沙箱时执行）
	var sandboxSessionID string

	if sv.sandboxPlatformConf.Enable {
		sessionID := cutil.GetSandboxSessionID()

		var sandboxErr error

		sandboxSessionID, sandboxErr = sv.ensureSandboxSession(ctx, sessionID, &req)
		if sandboxErr != nil {
			o11y.Warn(ctx, fmt.Sprintf("[Init] ensure sandbox session failed: %v", sandboxErr))
		}
	}

	rt = conversationresp.InitConversationResp{
		ID:               po.ID,
		SandboxSessionID: sandboxSessionID,
	}

	return
}

// ensureSandboxSession 确保 Sandbox Session 存在并就绪
func (sv *conversationSvc) ensureSandboxSession(ctx context.Context, sessionID string, req *conversationreq.InitReq) (string, error) {
	// 1. 检测 Session 状态
	sessionInfo, err := sv.sandboxPlatform.GetSession(ctx, sessionID)
	if err != nil {
		// 404 错误表示 Session 不存在，继续创建
		if sv.isSessionNotFoundError(err) {
			return sv.createNewSession(ctx, sessionID, req)
		}

		// 其他错误：尝试创建新 Session
		sv.logger.Warnf("[ensureSandboxSession] get session failed: %v, will create new session", err)

		return sv.createNewSession(ctx, sessionID, req)
	}

	// 2. 检查 Session 状态
	if sessionInfo.Status == "running" {
		sv.logger.Infof("[ensureSandboxSession] reuse existing session: %s", sessionID)
		return sessionID, nil
	}

	// 3. Session 状态非 running，自动重新创建
	sv.logger.Warnf("[ensureSandboxSession] session status is %s, will recreate: %s", sessionInfo.Status, sessionID)

	return sv.createNewSession(ctx, sessionID, req)
}

// createNewSession 创建新的 Sandbox Session
func (sv *conversationSvc) createNewSession(ctx context.Context, sessionID string, req *conversationreq.InitReq) (string, error) {
	cpu := sv.sandboxPlatformConf.DefaultCPU
	if cpu == "" {
		cpu = "1"
	}

	memory := sv.sandboxPlatformConf.DefaultMemory
	if memory == "" {
		memory = "512Mi"
	}

	disk := sv.sandboxPlatformConf.DefaultDisk
	if disk == "" {
		disk = "1Gi"
	}

	timeout := sv.sandboxPlatformConf.DefaultTimeout
	if timeout == 0 {
		timeout = 300
	}

	createReq := sandboxplatformdto.CreateSessionReq{
		ID:         &sessionID,
		TemplateID: sv.sandboxPlatformConf.DefaultTemplateID,
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
				"max_file_size":      sv.sandboxPlatformConf.DefaultFileUploadConfig.MaxFileSize,
				"max_file_size_unit": sv.sandboxPlatformConf.DefaultFileUploadConfig.MaxFileSizeUnit,
				"max_file_count":     sv.sandboxPlatformConf.DefaultFileUploadConfig.MaxFileCount,
				"allowed_file_types": sv.sandboxPlatformConf.DefaultFileUploadConfig.AllowedFileTypes,
			},
		},
	}

	createResp, err := sv.sandboxPlatform.CreateSession(ctx, createReq)
	if err != nil {
		if sv.isSessionAlreadyExistsError(err) {
			sv.logger.Infof("[createNewSession] session already exists: %s, will wait for ready", sessionID)
			return sv.waitForSessionReady(ctx, sessionID)
		}

		sv.logger.Errorf("[createNewSession] create failed: %v", err)

		return "", errors.Wrap(err, "create sandbox session failed")
	}

	actualSessionID := createResp.ID
	if createResp.ID == "" {
		actualSessionID = sessionID
	}

	return sv.waitForSessionReady(ctx, actualSessionID)
}

// waitForSessionReady 等待 Session 就绪
func (sv *conversationSvc) waitForSessionReady(ctx context.Context, sessionID string) (string, error) {
	maxRetries := sv.sandboxPlatformConf.MaxRetries
	retryInterval := sv.sandboxPlatformConf.RetryInterval

	retryIntervalDuration, err := time.ParseDuration(retryInterval)
	if err != nil {
		sv.logger.Warnf("[waitForSessionReady] failed to parse retry interval, using default 500ms")

		retryIntervalDuration = 500 * time.Millisecond
	}

	for i := 0; i < maxRetries; i++ {
		sessionInfo, err := sv.sandboxPlatform.GetSession(ctx, sessionID)
		if err != nil {
			sv.logger.Errorf("[waitForSessionReady] get session status failed (attempt %d): %v", i+1, err)
			time.Sleep(retryIntervalDuration)

			continue
		}

		if sessionInfo.Status == "running" {
			sv.logger.Infof("[waitForSessionReady] session ready: %s (attempts: %d)", sessionID, i+1)
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
func (sv *conversationSvc) isSessionNotFoundError(err error) bool {
	// 检查是否为 rest.HTTPError 类型
	var httpErr *rest.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.HTTPCode == http.StatusNotFound
	}

	return false
}

// isSessionAlreadyExistsError 判断是否为 Session 已存在的错误
func (sv *conversationSvc) isSessionAlreadyExistsError(err error) bool {
	var httpErr *rest.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.HTTPCode == http.StatusConflict
	}

	return strings.Contains(err.Error(), "already exists")
}
