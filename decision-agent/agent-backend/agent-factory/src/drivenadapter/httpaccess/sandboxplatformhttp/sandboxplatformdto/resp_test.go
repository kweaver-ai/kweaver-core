package sandboxplatformdto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSessionResp_StructFields(t *testing.T) {
	t.Parallel()

	maxProcesses := 10
	workspacePath := "/workspace"
	runtimeNode := "node-1"
	containerID := "container-123"
	podName := "pod-1"
	completedAt := "2024-01-01T00:00:00Z"
	lastActivityAt := "2024-01-01T01:00:00Z"

	resp := &CreateSessionResp{
		ID:         "session-123",
		TemplateID: "template-456",
		Status:     "running",
		ResourceLimit: &ResourceLimit{
			CPU:          "2",
			Memory:       "4GB",
			Disk:         "20GB",
			MaxProcesses: &maxProcesses,
		},
		WorkspacePath:  &workspacePath,
		RuntimeType:    "docker",
		RuntimeNode:    &runtimeNode,
		ContainerID:    &containerID,
		PodName:        &podName,
		EnvVars:        map[string]string{"KEY": "value"},
		Timeout:        3600,
		CreatedAt:      "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
		CompletedAt:    &completedAt,
		LastActivityAt: &lastActivityAt,
	}

	assert.Equal(t, "session-123", resp.ID)
	assert.Equal(t, "template-456", resp.TemplateID)
	assert.Equal(t, "running", resp.Status)
	assert.NotNil(t, resp.ResourceLimit)
	assert.Equal(t, "2", resp.ResourceLimit.CPU)
	assert.Equal(t, "4GB", resp.ResourceLimit.Memory)
	assert.NotNil(t, resp.ResourceLimit.MaxProcesses)
	assert.Equal(t, 10, *resp.ResourceLimit.MaxProcesses)
}

func TestGetSessionResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := &GetSessionResp{
		ID:         "session-456",
		TemplateID: "template-789",
		Status:     "completed",
		Timeout:    1800,
	}

	assert.Equal(t, "session-456", resp.ID)
	assert.Equal(t, "template-789", resp.TemplateID)
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, 1800, resp.Timeout)
}

func TestResourceLimit_StructFields(t *testing.T) {
	t.Parallel()

	maxProcesses := 5

	limit := &ResourceLimit{
		CPU:          "1",
		Memory:       "2GB",
		Disk:         "10GB",
		MaxProcesses: &maxProcesses,
	}

	assert.Equal(t, "1", limit.CPU)
	assert.Equal(t, "2GB", limit.Memory)
	assert.Equal(t, "10GB", limit.Disk)
	assert.NotNil(t, limit.MaxProcesses)
	assert.Equal(t, 5, *limit.MaxProcesses)
}

func TestResourceLimit_WithoutOptionalFields(t *testing.T) {
	t.Parallel()

	limit := &ResourceLimit{
		CPU:    "1",
		Memory: "2GB",
		Disk:   "10GB",
	}

	assert.Equal(t, "1", limit.CPU)
	assert.Equal(t, "2GB", limit.Memory)
	assert.Equal(t, "10GB", limit.Disk)
	assert.Nil(t, limit.MaxProcesses)
}

func TestCreateSessionResp_PointerFields(t *testing.T) {
	t.Parallel()

	resp := &CreateSessionResp{
		ID:         "session-789",
		TemplateID: "template-101",
		Status:     "pending",
	}

	// Verify pointer fields are nil when not set
	assert.Nil(t, resp.ResourceLimit)
	assert.Nil(t, resp.WorkspacePath)
	assert.Nil(t, resp.RuntimeNode)
	assert.Nil(t, resp.ContainerID)
	assert.Nil(t, resp.PodName)
	assert.Nil(t, resp.CompletedAt)
	assert.Nil(t, resp.LastActivityAt)
}
