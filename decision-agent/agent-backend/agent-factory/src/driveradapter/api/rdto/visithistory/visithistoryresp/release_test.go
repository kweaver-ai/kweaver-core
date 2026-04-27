package releaseresp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := PublishResp{
		ReleaseId: "release-123",
		Version:   "1.0.0",
	}

	assert.Equal(t, "release-123", resp.ReleaseId)
	assert.Equal(t, "1.0.0", resp.Version)
}

func TestPublishResp_JSONTags(t *testing.T) {
	t.Parallel()

	resp := PublishResp{
		ReleaseId: "release-456",
		Version:   "2.0.0",
	}

	// Marshal to JSON
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	// Unmarshal to map to check JSON tags
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "release-456", result["release_id"])
	assert.Equal(t, "2.0.0", result["version"])
}

func TestPublishResp_EmptyValues(t *testing.T) {
	t.Parallel()

	resp := PublishResp{}

	assert.Empty(t, resp.ReleaseId)
	assert.Empty(t, resp.Version)
}

func TestPublishResp_WithRealVersionFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{
			name:    "semantic versioning",
			version: "1.2.3",
			valid:   true,
		},
		{
			name:    "version with prerelease",
			version: "2.0.0-alpha",
			valid:   true,
		},
		{
			name:    "version with build metadata",
			version: "1.0.0+build.123",
			valid:   true,
		},
		{
			name:    "complex version",
			version: "3.1.4-beta.2+build.456",
			valid:   true,
		},
		{
			name:    "empty version",
			version: "",
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := PublishResp{
				ReleaseId: "release-test",
				Version:   tt.version,
			}

			assert.Equal(t, tt.version, resp.Version)

			if tt.valid && tt.version != "" {
				assert.NotEmpty(t, resp.Version)
			}
		})
	}
}

func TestHistoryListItemResp_StructFields(t *testing.T) {
	t.Parallel()

	item := HistoryListItemResp{
		HistoryId:    "history-123",
		AgentId:      "agent-456",
		AgentVersion: "1.2.3",
		AgentDesc:    "Test agent description",
		CreateTime:   1640995200000,
	}

	assert.Equal(t, "history-123", item.HistoryId)
	assert.Equal(t, "agent-456", item.AgentId)
	assert.Equal(t, "1.2.3", item.AgentVersion)
	assert.Equal(t, "Test agent description", item.AgentDesc)
	assert.Equal(t, int64(1640995200000), item.CreateTime)
}

func TestHistoryListItemResp_JSONTags(t *testing.T) {
	t.Parallel()

	item := HistoryListItemResp{
		HistoryId:    "hist-1",
		AgentId:      "agent-1",
		AgentVersion: "1.0.0",
		AgentDesc:    "Description",
		CreateTime:   1640995200000,
	}

	// Marshal to JSON
	data, err := json.Marshal(item)
	require.NoError(t, err)

	// Unmarshal to map to check JSON tags
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "hist-1", result["history_id"])
	assert.Equal(t, "agent-1", result["agent_id"])
	assert.Equal(t, "1.0.0", result["agent_version"])
	assert.Equal(t, "Description", result["agent_desc"])
	assert.Equal(t, float64(1640995200000), result["create_time"]) // JSON numbers are float64
}

func TestHistoryListItemResp_EmptyValues(t *testing.T) {
	t.Parallel()

	item := HistoryListItemResp{}

	assert.Empty(t, item.HistoryId)
	assert.Empty(t, item.AgentId)
	assert.Empty(t, item.AgentVersion)
	assert.Empty(t, item.AgentDesc)
	assert.Equal(t, int64(0), item.CreateTime)
}

func TestHistoryListResp_Type(t *testing.T) {
	t.Parallel()

	// HistoryListResp is a slice type
	var list HistoryListResp

	assert.Nil(t, list)
	assert.IsType(t, HistoryListResp{}, list)
}

func TestHistoryListResp_WithMultipleItems(t *testing.T) {
	t.Parallel()

	list := HistoryListResp{
		{
			HistoryId:    "hist-1",
			AgentId:      "agent-1",
			AgentVersion: "1.0.0",
			AgentDesc:    "First agent",
			CreateTime:   1640995200000,
		},
		{
			HistoryId:    "hist-2",
			AgentId:      "agent-2",
			AgentVersion: "1.0.1",
			AgentDesc:    "Second agent",
			CreateTime:   1641081600000,
		},
		{
			HistoryId:    "hist-3",
			AgentId:      "agent-3",
			AgentVersion: "1.1.0",
			AgentDesc:    "Third agent",
			CreateTime:   1641168000000,
		},
	}

	assert.Len(t, list, 3)
	assert.Equal(t, "hist-1", list[0].HistoryId)
	assert.Equal(t, "hist-2", list[1].HistoryId)
	assert.Equal(t, "hist-3", list[2].HistoryId)
}

func TestHistoryListResp_Empty(t *testing.T) {
	t.Parallel()

	list := HistoryListResp{}

	assert.Empty(t, list)
	assert.Len(t, list, 0)
}

func TestHistoryListResp_JSONMarshaling(t *testing.T) {
	t.Parallel()

	list := HistoryListResp{
		{
			HistoryId:    "hist-1",
			AgentId:      "agent-1",
			AgentVersion: "1.0.0",
			AgentDesc:    "Agent 1",
			CreateTime:   1640995200000,
		},
		{
			HistoryId:    "hist-2",
			AgentId:      "agent-2",
			AgentVersion: "2.0.0",
			AgentDesc:    "Agent 2",
			CreateTime:   1641081600000,
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(list)
	require.NoError(t, err)

	// Unmarshal back
	var result HistoryListResp
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, "hist-1", result[0].HistoryId)
	assert.Equal(t, "hist-2", result[1].HistoryId)
	assert.Equal(t, "1.0.0", result[0].AgentVersion)
	assert.Equal(t, "2.0.0", result[1].AgentVersion)
}

func TestHistoryListItemResp_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	item := HistoryListItemResp{
		HistoryId:    "hist-中文-123",
		AgentId:      "agent-测试",
		AgentVersion: "1.0.0-β",
		AgentDesc:    "This is a description with \"quotes\" and 'apostrophes'",
		CreateTime:   1640995200000,
	}

	assert.Equal(t, "hist-中文-123", item.HistoryId)
	assert.Equal(t, "agent-测试", item.AgentId)
	assert.Equal(t, "1.0.0-β", item.AgentVersion)
	assert.Contains(t, item.AgentDesc, "quotes")
}

func TestHistoryListResp_Append(t *testing.T) {
	t.Parallel()

	list := HistoryListResp{}

	// Append items
	list = append(list, HistoryListItemResp{
		HistoryId:    "hist-1",
		AgentId:      "agent-1",
		AgentVersion: "1.0.0",
		AgentDesc:    "Agent 1",
		CreateTime:   1640995200000,
	})

	list = append(list, HistoryListItemResp{
		HistoryId:    "hist-2",
		AgentId:      "agent-2",
		AgentVersion: "2.0.0",
		AgentDesc:    "Agent 2",
		CreateTime:   1641081600000,
	})

	assert.Len(t, list, 2)
	assert.Equal(t, "Agent 1", list[0].AgentDesc)
	assert.Equal(t, "Agent 2", list[1].AgentDesc)
}

func TestHistoryListItemResp_TimestampComparison(t *testing.T) {
	t.Parallel()

	olderItem := HistoryListItemResp{
		HistoryId:  "hist-older",
		CreateTime: 1640995200000, // Earlier
	}

	newerItem := HistoryListItemResp{
		HistoryId:  "hist-newer",
		CreateTime: 1641081600000, // Later
	}

	assert.True(t, newerItem.CreateTime > olderItem.CreateTime)
	assert.True(t, olderItem.CreateTime < newerItem.CreateTime)
}
