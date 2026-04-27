package sandboxplatformhttp

import (
	"context"

	sandboxdto "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/isandboxhtpp"
)

type mockSandboxPlatform struct {
	logger icmp.Logger
}

func NewMockSandboxPlatform(logger icmp.Logger) isandboxhtpp.ISandboxPlatform {
	return &mockSandboxPlatform{
		logger: logger,
	}
}

func (m *mockSandboxPlatform) CreateSession(ctx context.Context, req sandboxdto.CreateSessionReq) (*sandboxdto.CreateSessionResp, error) {
	m.logger.Infof("[MockSandboxPlatform] create session: templateID=%s, id=%v", req.TemplateID, req.ID)

	cpu := "1"
	if req.CPU != "" {
		cpu = req.CPU
	}

	memory := "512Mi"
	if req.Memory != "" {
		memory = req.Memory
	}

	disk := "1Gi"
	if req.Disk != "" {
		disk = req.Disk
	}

	timeout := 300
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	sessionID := "mock-session-" + req.TemplateID
	if req.ID != nil && *req.ID != "" {
		sessionID = *req.ID
	}

	resp := &sandboxdto.CreateSessionResp{
		ID:          sessionID,
		TemplateID:  req.TemplateID,
		Status:      "running",
		RuntimeType: "python3.11",
		ResourceLimit: &sandboxdto.ResourceLimit{
			CPU:          cpu,
			Memory:       memory,
			Disk:         disk,
			MaxProcesses: new(int),
		},
		EnvVars:   req.EnvVars,
		Timeout:   timeout,
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}
	*(resp.ResourceLimit.MaxProcesses) = 128

	m.logger.Infof("[MockSandboxPlatform] create session success: %s", resp.ID)

	return resp, nil
}

func (m *mockSandboxPlatform) GetSession(ctx context.Context, sessionID string) (*sandboxdto.GetSessionResp, error) {
	m.logger.Infof("[MockSandboxPlatform] get session: %s", sessionID)

	maxProcesses := 128
	resp := &sandboxdto.GetSessionResp{
		ID:          sessionID,
		TemplateID:  "python3.11",
		Status:      "running",
		RuntimeType: "python3.11",
		ResourceLimit: &sandboxdto.ResourceLimit{
			CPU:          "1",
			Memory:       "512Mi",
			Disk:         "1Gi",
			MaxProcesses: &maxProcesses,
		},
		EnvVars:   map[string]string{},
		Timeout:   300,
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	m.logger.Infof("[MockSandboxPlatform] get session success: %s, status: %s", sessionID, resp.Status)

	return resp, nil
}

func (m *mockSandboxPlatform) DeleteSession(ctx context.Context, sessionID string) error {
	m.logger.Infof("[MockSandboxPlatform] delete session: %s", sessionID)
	m.logger.Infof("[MockSandboxPlatform] delete session success: %s", sessionID)

	return nil
}

func (m *mockSandboxPlatform) ListFiles(ctx context.Context, sessionID string, limit int) ([]string, error) {
	m.logger.Infof("[MockSandboxPlatform] list files: sessionID=%s, limit=%d", sessionID, limit)

	files := []string{
		sessionID + "/file1.txt",
		sessionID + "/file2.py",
		sessionID + "/file3.json",
	}

	// Apply limit if specified
	if limit > 0 && limit < len(files) {
		files = files[:limit]
	}

	m.logger.Infof("[MockSandboxPlatform] list files success: found %d files", len(files))

	return files, nil
}
