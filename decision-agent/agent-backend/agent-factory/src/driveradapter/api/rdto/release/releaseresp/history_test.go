package releaseresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistoryListResp_Type(t *testing.T) {
	t.Parallel()

	// HistoryListResp is a slice type
	var list HistoryListResp

	assert.Nil(t, list)
	assert.IsType(t, HistoryListResp{}, list)
}

func TestHistoryListResp_Empty(t *testing.T) {
	t.Parallel()

	list := HistoryListResp{}

	assert.Empty(t, list)
	assert.Len(t, list, 0)
}

func TestHistoryListResp_WithItems(t *testing.T) {
	t.Parallel()

	list := HistoryListResp{
		{
			HistoryId:    "hist-1",
			AgentId:      "agent-1",
			AgentVersion: "1.0.0",
			AgentDesc:    "Description 1",
			CreateTime:   1640995200000,
		},
		{
			HistoryId:    "hist-2",
			AgentId:      "agent-2",
			AgentVersion: "2.0.0",
			AgentDesc:    "Description 2",
			CreateTime:   1641081600000,
		},
	}

	assert.Len(t, list, 2)
	assert.Equal(t, "hist-1", list[0].HistoryId)
	assert.Equal(t, "agent-1", list[0].AgentId)
	assert.Equal(t, "1.0.0", list[0].AgentVersion)
	assert.Equal(t, "Description 1", list[0].AgentDesc)
	assert.Equal(t, int64(1640995200000), list[0].CreateTime)
}

func TestHistoryListItemResp_StructFields(t *testing.T) {
	t.Parallel()

	item := HistoryListItemResp{
		HistoryId:    "hist-123",
		AgentId:      "agent-456",
		AgentVersion: "1.5.0",
		AgentDesc:    "Test description",
		CreateTime:   1640995200000,
	}

	assert.Equal(t, "hist-123", item.HistoryId)
	assert.Equal(t, "agent-456", item.AgentId)
	assert.Equal(t, "1.5.0", item.AgentVersion)
	assert.Equal(t, "Test description", item.AgentDesc)
	assert.Equal(t, int64(1640995200000), item.CreateTime)
}

func TestHistoryListItemResp_Empty(t *testing.T) {
	t.Parallel()

	item := HistoryListItemResp{}

	assert.Empty(t, item.HistoryId)
	assert.Empty(t, item.AgentId)
	assert.Empty(t, item.AgentVersion)
	assert.Empty(t, item.AgentDesc)
	assert.Equal(t, int64(0), item.CreateTime)
}

func TestHistoryListResp_Append(t *testing.T) {
	t.Parallel()

	list := HistoryListResp{}

	item1 := HistoryListItemResp{
		HistoryId: "hist-1",
		AgentId:   "agent-1",
	}
	item2 := HistoryListItemResp{
		HistoryId: "hist-2",
		AgentId:   "agent-2",
	}

	list = append(list, item1)
	list = append(list, item2)

	assert.Len(t, list, 2)
	assert.Equal(t, "hist-1", list[0].HistoryId)
	assert.Equal(t, "hist-2", list[1].HistoryId)
}

func TestHistoryListResp_SliceOperations(t *testing.T) {
	t.Parallel()

	list := HistoryListResp{
		{HistoryId: "hist-1", AgentId: "agent-1"},
		{HistoryId: "hist-2", AgentId: "agent-2"},
		{HistoryId: "hist-3", AgentId: "agent-3"},
	}

	// Test length
	assert.Len(t, list, 3)

	// Test slicing
	subList := list[1:3]
	assert.Len(t, subList, 2)
	assert.Equal(t, "hist-2", subList[0].HistoryId)
	assert.Equal(t, "hist-3", subList[1].HistoryId)

	// Test iteration
	count := 0

	for _, item := range list {
		assert.NotEmpty(t, item.HistoryId)

		count++
	}

	assert.Equal(t, 3, count)
}

func TestHistoryListItemResp_WithTimestamps(t *testing.T) {
	t.Parallel()

	timestamps := []int64{
		1640995200000, // 2022-01-01
		1643673600000, // 2022-02-01
		1646092800000, // 2022-03-01
		1672531200000, // 2023-01-01
		1704067200000, // 2024-01-01
	}

	for i, ts := range timestamps {
		item := HistoryListItemResp{
			HistoryId:  "hist-" + string(rune('a'+i)),
			CreateTime: ts,
		}
		assert.Equal(t, ts, item.CreateTime)
	}
}

func TestHistoryListResp_Capacity(t *testing.T) {
	t.Parallel()

	list := make(HistoryListResp, 0, 100)

	assert.Len(t, list, 0)
	assert.NotNil(t, list)

	// Test that we can append up to capacity
	for i := 0; i < 100; i++ {
		list = append(list, HistoryListItemResp{
			HistoryId: "hist-" + string(rune(i)),
		})
	}

	assert.Len(t, list, 100)
}

func TestHistoryListItemResp_WithLongDescription(t *testing.T) {
	t.Parallel()

	longDesc := "This is a very long description that contains multiple words and should be stored properly in the AgentDesc field. It might contain special characters like 中文 and numbers like 12345."
	item := HistoryListItemResp{
		HistoryId: "hist-long",
		AgentDesc: longDesc,
	}

	assert.Equal(t, longDesc, item.AgentDesc)
	assert.Contains(t, item.AgentDesc, "中文")
	assert.Contains(t, item.AgentDesc, "12345")
}

func TestHistoryListItemResp_WithDifferentVersions(t *testing.T) {
	t.Parallel()

	versions := []string{
		"1.0.0",
		"2.1.3",
		"3.0.0-alpha",
		"4.5.2-beta.1",
		"latest",
	}

	for _, version := range versions {
		item := HistoryListItemResp{
			HistoryId:    "hist-" + version,
			AgentVersion: version,
		}
		assert.Equal(t, version, item.AgentVersion)
	}
}

func TestHistoryListItemResp_WithChineseCharacters(t *testing.T) {
	t.Parallel()

	item := HistoryListItemResp{
		HistoryId:    "hist-中文",
		AgentId:      "agent-中文",
		AgentVersion: "版本-1.0",
		AgentDesc:    "这是中文描述",
	}

	assert.Equal(t, "hist-中文", item.HistoryId)
	assert.Equal(t, "agent-中文", item.AgentId)
	assert.Equal(t, "版本-1.0", item.AgentVersion)
	assert.Equal(t, "这是中文描述", item.AgentDesc)
}

func TestHistoryListResp_WithDuplicateItems(t *testing.T) {
	t.Parallel()

	list := HistoryListResp{
		{
			HistoryId: "hist-1",
			AgentId:   "agent-1",
		},
		{
			HistoryId: "hist-1",  // Same ID
			AgentId:   "agent-1", // Same agent
		},
	}

	assert.Len(t, list, 2)
}

func TestHistoryListItemResp_WithAllFields(t *testing.T) {
	t.Parallel()

	item := HistoryListItemResp{
		HistoryId:    "hist-complete",
		AgentId:      "agent-complete",
		AgentVersion: "9.9.9",
		AgentDesc:    "Complete description with all details",
		CreateTime:   1704067200000,
	}

	assert.Equal(t, "hist-complete", item.HistoryId)
	assert.Equal(t, "agent-complete", item.AgentId)
	assert.Equal(t, "9.9.9", item.AgentVersion)
	assert.Equal(t, "Complete description with all details", item.AgentDesc)
	assert.Equal(t, int64(1704067200000), item.CreateTime)
}
