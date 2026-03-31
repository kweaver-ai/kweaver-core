package agentresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileCheckResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := FileCheckResp{
		Progress: 50,
		ProcessInfo: []Info{
			{
				ID:     "info-1",
				Status: "completed",
			},
			{
				ID:     "info-2",
				Status: "processing",
			},
		},
	}

	assert.Equal(t, 50, resp.Progress)
	assert.Len(t, resp.ProcessInfo, 2)
	assert.Equal(t, "info-1", resp.ProcessInfo[0].ID)
	assert.Equal(t, "completed", resp.ProcessInfo[0].Status)
	assert.Equal(t, "info-2", resp.ProcessInfo[1].ID)
	assert.Equal(t, "processing", resp.ProcessInfo[1].Status)
}

func TestFileCheckResp_Empty(t *testing.T) {
	t.Parallel()

	resp := FileCheckResp{}

	assert.Equal(t, 0, resp.Progress)
	assert.Nil(t, resp.ProcessInfo)
}

func TestInfo_StructFields(t *testing.T) {
	t.Parallel()

	info := Info{
		ID:     "info-123",
		Status: "completed",
	}

	assert.Equal(t, "info-123", info.ID)
	assert.Equal(t, "completed", info.Status)
}

func TestInfo_Empty(t *testing.T) {
	t.Parallel()

	info := Info{}

	assert.Empty(t, info.ID)
	assert.Empty(t, info.Status)
}

func TestInfo_WithDifferentStatuses(t *testing.T) {
	t.Parallel()

	statuses := []string{
		"completed",
		"processing",
		"failed",
		"pending",
		"",
	}

	for _, status := range statuses {
		info := Info{
			ID:     "info-" + status,
			Status: status,
		}
		assert.Equal(t, status, info.Status)
	}
}

func TestFileCheckResp_WithDifferentProgress(t *testing.T) {
	t.Parallel()

	progressValues := []int{
		0,
		25,
		50,
		75,
		100,
	}

	for _, progress := range progressValues {
		resp := FileCheckResp{
			Progress: progress,
		}
		assert.Equal(t, progress, resp.Progress)
	}
}

func TestFileCheckResp_WithSingleInfo(t *testing.T) {
	t.Parallel()

	resp := FileCheckResp{
		Progress: 100,
		ProcessInfo: []Info{
			{
				ID:     "info-single",
				Status: "completed",
			},
		},
	}

	assert.Len(t, resp.ProcessInfo, 1)
	assert.Equal(t, "info-single", resp.ProcessInfo[0].ID)
}

func TestFileCheckResp_WithNoInfo(t *testing.T) {
	t.Parallel()

	resp := FileCheckResp{
		Progress:    100,
		ProcessInfo: []Info{},
	}

	assert.Equal(t, 100, resp.Progress)
	assert.NotNil(t, resp.ProcessInfo)
	assert.Len(t, resp.ProcessInfo, 0)
}

func TestFileCheckResp_WithMultipleInfo(t *testing.T) {
	t.Parallel()

	infos := []Info{
		{ID: "info-1", Status: "completed"},
		{ID: "info-2", Status: "completed"},
		{ID: "info-3", Status: "processing"},
		{ID: "info-4", Status: "failed"},
		{ID: "info-5", Status: "completed"},
	}

	resp := FileCheckResp{
		Progress:    60,
		ProcessInfo: infos,
	}

	assert.Len(t, resp.ProcessInfo, 5)
}

func TestFileCheckResp_WithCompletedStatus(t *testing.T) {
	t.Parallel()

	resp := FileCheckResp{
		Progress: 100,
		ProcessInfo: []Info{
			{
				ID:     "info-1",
				Status: "completed",
			},
		},
	}

	assert.Equal(t, 100, resp.Progress)
	assert.Equal(t, "completed", resp.ProcessInfo[0].Status)
}

func TestFileCheckResp_WithProcessingStatus(t *testing.T) {
	t.Parallel()

	resp := FileCheckResp{
		Progress: 50,
		ProcessInfo: []Info{
			{
				ID:     "info-1",
				Status: "processing",
			},
		},
	}

	assert.Equal(t, 50, resp.Progress)
	assert.Equal(t, "processing", resp.ProcessInfo[0].Status)
}

func TestFileCheckResp_WithFailedStatus(t *testing.T) {
	t.Parallel()

	resp := FileCheckResp{
		Progress: 75,
		ProcessInfo: []Info{
			{
				ID:     "info-1",
				Status: "failed",
			},
		},
	}

	assert.Equal(t, 75, resp.Progress)
	assert.Equal(t, "failed", resp.ProcessInfo[0].Status)
}

func TestFileCheckResp_MixedStatuses(t *testing.T) {
	t.Parallel()

	resp := FileCheckResp{
		Progress: 60,
		ProcessInfo: []Info{
			{ID: "info-1", Status: "completed"},
			{ID: "info-2", Status: "processing"},
			{ID: "info-3", Status: "failed"},
		},
	}

	assert.Len(t, resp.ProcessInfo, 3)
}

func TestInfo_WithChineseStatus(t *testing.T) {
	t.Parallel()

	info := Info{
		ID:     "info-中文",
		Status: "已完成",
	}

	assert.Equal(t, "info-中文", info.ID)
	assert.Equal(t, "已完成", info.Status)
}

func TestFileCheckResp_WithChineseStatus(t *testing.T) {
	t.Parallel()

	resp := FileCheckResp{
		Progress: 100,
		ProcessInfo: []Info{
			{
				ID:     "info-1",
				Status: "已完成",
			},
		},
	}

	assert.Equal(t, "已完成", resp.ProcessInfo[0].Status)
}
