package agentsvc

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Mock sandbox platform for testing EnsureSandboxSession
type mockGetSessionSandbox struct {
	getSessionFunc    func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error)
	createSessionFunc func(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error)
	deleteSessionFunc func(ctx context.Context, sessionID string) error
}

func (m *mockGetSessionSandbox) CreateSession(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(ctx, req)
	}

	return &sandboxplatformdto.CreateSessionResp{
		ID:     *req.ID,
		Status: "running",
	}, nil
}

func (m *mockGetSessionSandbox) GetSession(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
	if m.getSessionFunc != nil {
		return m.getSessionFunc(ctx, sessionID)
	}

	return &sandboxplatformdto.GetSessionResp{
		ID:     sessionID,
		Status: "running",
	}, nil
}

func (m *mockGetSessionSandbox) DeleteSession(ctx context.Context, sessionID string) error {
	if m.deleteSessionFunc != nil {
		return m.deleteSessionFunc(ctx, sessionID)
	}

	return nil
}

func (m *mockGetSessionSandbox) ListFiles(ctx context.Context, sessionID string, limit int) ([]string, error) {
	return []string{}, nil
}

func allowAnyLoggerCalls(mockLogger *cmpmock.MockLogger) {
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infoln(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnln(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorln(gomock.Any()).AnyTimes()
}

func TestIsSessionNotFoundError(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{}
	ctx := context.Background()

	t.Run("returns true for 404 HTTPError", func(t *testing.T) {
		t.Parallel()

		err := rest.NewHTTPError(ctx, http.StatusNotFound, rest.PublicError_NotFound)
		result := svc.isSessionNotFoundError(err)
		assert.True(t, result)
	})

	t.Run("returns false for 500 HTTPError", func(t *testing.T) {
		t.Parallel()

		err := rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError)
		result := svc.isSessionNotFoundError(err)
		assert.False(t, result)
	})

	t.Run("returns false for non-HTTPError", func(t *testing.T) {
		t.Parallel()

		err := errors.New("some error")
		result := svc.isSessionNotFoundError(err)
		assert.False(t, result)
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		t.Parallel()

		result := svc.isSessionNotFoundError(nil)
		assert.False(t, result)
	})
}

func TestIsSessionAlreadyExistsError_AgentSvc(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{}
	ctx := context.Background()

	t.Run("returns true for 409 HTTPError", func(t *testing.T) {
		t.Parallel()

		err := rest.NewHTTPError(ctx, http.StatusConflict, rest.PublicError_Conflict)
		result := svc.isSessionAlreadyExistsError(err)
		assert.True(t, result)
	})

	t.Run("returns true for error with 'already exists' message", func(t *testing.T) {
		t.Parallel()

		err := errors.New("session already exists")
		result := svc.isSessionAlreadyExistsError(err)
		assert.True(t, result)
	})

	t.Run("returns false for other HTTPError", func(t *testing.T) {
		t.Parallel()

		err := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest)
		result := svc.isSessionAlreadyExistsError(err)
		assert.False(t, result)
	})

	t.Run("returns false for non-HTTPError without 'already exists'", func(t *testing.T) {
		t.Parallel()

		err := errors.New("some other error")
		result := svc.isSessionAlreadyExistsError(err)
		assert.False(t, result)
	})

	t.Run("case sensitive check for 'already exists'", func(t *testing.T) {
		t.Parallel()

		err := errors.New("Session ALREADY EXISTS error")
		result := svc.isSessionAlreadyExistsError(err)
		assert.False(t, result) // Implementation is case-sensitive, only checks lowercase "already exists"
	})
}

func TestEnsureSandboxSession_SessionRunning(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			return &sandboxplatformdto.GetSessionResp{
				ID:     sessionID,
				Status: "running",
			}, nil
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			MaxRetries:    3,
			RetryInterval: "500ms",
		},
	}

	ctx := context.Background()
	sessionID := "test-session-123"
	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{
			UserID:            "user-123",
			XBusinessDomainID: "bd-789",
		},
		AgentID: "agent-456",
	}

	result, err := svc.EnsureSandboxSession(ctx, sessionID, req)

	assert.NoError(t, err)
	assert.Equal(t, sessionID, result)
}

func TestEnsureSandboxSession_SessionNotFound_CreatesNew(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	sessionCreated := false
	getCallCount := 0

	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			getCallCount++
			if getCallCount == 1 {
				return nil, rest.NewHTTPError(ctx, http.StatusNotFound, rest.PublicError_NotFound)
			}
			return &sandboxplatformdto.GetSessionResp{
				ID:     sessionID,
				Status: "running",
			}, nil
		},
		createSessionFunc: func(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
			sessionCreated = true
			return &sandboxplatformdto.CreateSessionResp{
				ID:     *req.ID,
				Status: "running",
			}, nil
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			DefaultTemplateID: "python3.11",
			MaxRetries:        3,
			RetryInterval:     "500ms",
			DefaultCPU:        "1",
			DefaultMemory:     "512Mi",
			DefaultDisk:       "1Gi",
			DefaultTimeout:    300,
		},
	}

	ctx := context.Background()
	sessionID := "test-session-new"
	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{
			UserID:            "user-123",
			XBusinessDomainID: "bd-789",
		},
		AgentID: "agent-456",
	}

	result, err := svc.EnsureSandboxSession(ctx, sessionID, req)

	assert.NoError(t, err)
	assert.Equal(t, sessionID, result)
	assert.True(t, sessionCreated, "create session should have been called")
}

func TestEnsureSandboxSession_SessionFailed_DeletesAndRecreates(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	sessionDeleted := false
	sessionCreated := false
	getCallCount := 0

	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			getCallCount++
			if getCallCount == 1 {
				return &sandboxplatformdto.GetSessionResp{
					ID:     sessionID,
					Status: "failed",
				}, nil
			}
			return &sandboxplatformdto.GetSessionResp{
				ID:     sessionID,
				Status: "running",
			}, nil
		},
		deleteSessionFunc: func(ctx context.Context, sessionID string) error {
			sessionDeleted = true
			return nil
		},
		createSessionFunc: func(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
			sessionCreated = true
			return &sandboxplatformdto.CreateSessionResp{
				ID:     *req.ID,
				Status: "running",
			}, nil
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			DefaultTemplateID: "python3.11",
			MaxRetries:        3,
			RetryInterval:     "500ms",
		},
	}

	ctx := context.Background()
	sessionID := "test-session-failed"
	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{
			UserID:            "user-123",
			XBusinessDomainID: "bd-789",
		},
		AgentID: "agent-456",
	}

	result, err := svc.EnsureSandboxSession(ctx, sessionID, req)

	assert.NoError(t, err)
	assert.Equal(t, sessionID, result)
	assert.True(t, sessionDeleted, "delete session should have been called")
	assert.True(t, sessionCreated, "create session should have been called")
}

func TestEnsureSandboxSession_SessionErrorStatus_DeletesAndRecreates(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	sessionDeleted := false
	sessionCreated := false
	getCallCount := 0

	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			getCallCount++
			if getCallCount == 1 {
				return &sandboxplatformdto.GetSessionResp{
					ID:     sessionID,
					Status: "error",
				}, nil
			}
			return &sandboxplatformdto.GetSessionResp{
				ID:     sessionID,
				Status: "running",
			}, nil
		},
		deleteSessionFunc: func(ctx context.Context, sessionID string) error {
			sessionDeleted = true
			return nil
		},
		createSessionFunc: func(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
			sessionCreated = true
			return &sandboxplatformdto.CreateSessionResp{
				ID:     *req.ID,
				Status: "running",
			}, nil
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			DefaultTemplateID: "python3.11",
			MaxRetries:        3,
			RetryInterval:     "500ms",
		},
	}

	ctx := context.Background()
	sessionID := "test-session-error"
	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{
			UserID:            "user-123",
			XBusinessDomainID: "bd-789",
		},
		AgentID: "agent-456",
	}

	result, err := svc.EnsureSandboxSession(ctx, sessionID, req)

	assert.NoError(t, err)
	assert.Equal(t, sessionID, result)
	assert.True(t, sessionDeleted, "delete session should have been called")
	assert.True(t, sessionCreated, "create session should have been called")
}

func TestEnsureSandboxSession_SessionStopped_DeletesAndRecreates(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	sessionDeleted := false
	sessionCreated := false
	getCallCount := 0

	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			getCallCount++
			if getCallCount == 1 {
				return &sandboxplatformdto.GetSessionResp{
					ID:     sessionID,
					Status: "stopped",
				}, nil
			}
			return &sandboxplatformdto.GetSessionResp{
				ID:     sessionID,
				Status: "running",
			}, nil
		},
		deleteSessionFunc: func(ctx context.Context, sessionID string) error {
			sessionDeleted = true
			return nil
		},
		createSessionFunc: func(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
			sessionCreated = true
			return &sandboxplatformdto.CreateSessionResp{
				ID:     *req.ID,
				Status: "running",
			}, nil
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			DefaultTemplateID: "python3.11",
			MaxRetries:        3,
			RetryInterval:     "500ms",
		},
	}

	ctx := context.Background()
	sessionID := "test-session-stopped"
	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{
			UserID:            "user-123",
			XBusinessDomainID: "bd-789",
		},
		AgentID: "agent-456",
	}

	result, err := svc.EnsureSandboxSession(ctx, sessionID, req)

	assert.NoError(t, err)
	assert.Equal(t, sessionID, result)
	assert.True(t, sessionDeleted, "delete session should have been called")
	assert.True(t, sessionCreated, "create session should have been called")
}

func TestEnsureSandboxSession_GetSessionError_CreatesNew(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	sessionCreated := false
	getCallCount := 0

	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			getCallCount++
			if getCallCount == 1 {
				return nil, errors.New("network error")
			}
			return &sandboxplatformdto.GetSessionResp{
				ID:     sessionID,
				Status: "running",
			}, nil
		},
		createSessionFunc: func(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
			sessionCreated = true
			return &sandboxplatformdto.CreateSessionResp{
				ID:     *req.ID,
				Status: "running",
			}, nil
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			DefaultTemplateID: "python3.11",
			MaxRetries:        3,
			RetryInterval:     "500ms",
		},
	}

	ctx := context.Background()
	sessionID := "test-session-network-error"
	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{
			UserID:            "user-123",
			XBusinessDomainID: "bd-789",
		},
		AgentID: "agent-456",
	}

	result, err := svc.EnsureSandboxSession(ctx, sessionID, req)

	assert.NoError(t, err)
	assert.Equal(t, sessionID, result)
	assert.True(t, sessionCreated, "create session should have been called")
}

func TestEnsureSandboxSession_SessionAlreadyExists_WaitsForReady(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	createCallCount := 0
	getCallCount := 0

	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			getCallCount++
			if getCallCount == 1 {
				return nil, rest.NewHTTPError(ctx, http.StatusNotFound, rest.PublicError_NotFound)
			}
			return &sandboxplatformdto.GetSessionResp{
				ID:     sessionID,
				Status: "running",
			}, nil
		},
		createSessionFunc: func(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
			createCallCount++
			return nil, rest.NewHTTPError(ctx, http.StatusConflict, rest.PublicError_Conflict)
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			DefaultTemplateID: "python3.11",
			MaxRetries:        3,
			RetryInterval:     "500ms",
		},
	}

	ctx := context.Background()
	sessionID := "test-session-exists"
	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{
			UserID:            "user-123",
			XBusinessDomainID: "bd-789",
		},
		AgentID: "agent-456",
	}

	result, err := svc.EnsureSandboxSession(ctx, sessionID, req)

	assert.NoError(t, err)
	assert.Equal(t, sessionID, result)
	assert.Equal(t, 1, createCallCount, "create should have been called once")
	assert.Equal(t, 2, getCallCount, "get should have been called twice (initial check + wait for ready)")
}
