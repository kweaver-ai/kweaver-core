package releasereq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionRange_Type(t *testing.T) {
	t.Parallel()

	// PermissionRange is a slice type
	var pr PermissionRange

	assert.Nil(t, pr)
	assert.IsType(t, PermissionRange{}, pr)
}

func TestPermissionRange_Empty(t *testing.T) {
	t.Parallel()

	pr := PermissionRange{}

	assert.Empty(t, pr)
	assert.Len(t, pr, 0)
}

func TestPermissionRange_WithItems(t *testing.T) {
	t.Parallel()

	pr := PermissionRange{
		{
			ObjectID:   "obj-1",
			ObjectType: "type-a",
		},
		{
			ObjectID:   "obj-2",
			ObjectType: "type-b",
		},
	}

	assert.Len(t, pr, 2)
	assert.Equal(t, "obj-1", pr[0].ObjectID)
	assert.Equal(t, "type-a", pr[0].ObjectType)
	assert.Equal(t, "obj-2", pr[1].ObjectID)
	assert.Equal(t, "type-b", pr[1].ObjectType)
}

func TestPermissionRangeObject_StructFields(t *testing.T) {
	t.Parallel()

	obj := PermissionRangeObject{
		ObjectID:   "test-obj-123",
		ObjectType: "test-type",
	}

	assert.Equal(t, "test-obj-123", obj.ObjectID)
	assert.Equal(t, "test-type", obj.ObjectType)
}

func TestPermissionRangeObject_Empty(t *testing.T) {
	t.Parallel()

	obj := PermissionRangeObject{}

	assert.Empty(t, obj.ObjectID)
	assert.Empty(t, obj.ObjectType)
}

func TestPermissionRangeObject_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	obj := PermissionRangeObject{
		ObjectID:   "obj-中文-123",
		ObjectType: "type-特殊",
	}

	assert.Equal(t, "obj-中文-123", obj.ObjectID)
	assert.Equal(t, "type-特殊", obj.ObjectType)
}

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
			HistoryID:    "hist-1",
			AgentID:      "agent-1",
			AgentVersion: "1.0.0",
			AgentDesc:    "Description 1",
			CreateTime:   1640995200000,
		},
		{
			HistoryID:    "hist-2",
			AgentID:      "agent-2",
			AgentVersion: "2.0.0",
			AgentDesc:    "Description 2",
			CreateTime:   1641081600000,
		},
	}

	assert.Len(t, list, 2)
	assert.Equal(t, "hist-1", list[0].HistoryID)
	assert.Equal(t, "agent-1", list[0].AgentID)
	assert.Equal(t, "1.0.0", list[0].AgentVersion)
	assert.Equal(t, "Description 1", list[0].AgentDesc)
	assert.Equal(t, int64(1640995200000), list[0].CreateTime)
}

func TestHistoryListItemResp_StructFields(t *testing.T) {
	t.Parallel()

	item := HistoryListItemResp{
		HistoryID:    "hist-123",
		AgentID:      "agent-456",
		AgentVersion: "1.5.0",
		AgentDesc:    "Test description",
		CreateTime:   1640995200000,
	}

	assert.Equal(t, "hist-123", item.HistoryID)
	assert.Equal(t, "agent-456", item.AgentID)
	assert.Equal(t, "1.5.0", item.AgentVersion)
	assert.Equal(t, "Test description", item.AgentDesc)
	assert.Equal(t, int64(1640995200000), item.CreateTime)
}

func TestHistoryListItemResp_Empty(t *testing.T) {
	t.Parallel()

	item := HistoryListItemResp{}

	assert.Empty(t, item.HistoryID)
	assert.Empty(t, item.AgentID)
	assert.Empty(t, item.AgentVersion)
	assert.Empty(t, item.AgentDesc)
	assert.Equal(t, int64(0), item.CreateTime)
}

func TestHistoryListItemResp_WithVersionFormats(t *testing.T) {
	t.Parallel()

	versions := []string{
		"1.0.0",
		"2.1.3",
		"3.0.0-alpha",
		"4.5.2-beta.1",
	}

	for _, version := range versions {
		item := HistoryListItemResp{
			HistoryID:    "hist-" + version,
			AgentVersion: version,
		}
		assert.Equal(t, version, item.AgentVersion)
	}
}

func TestHistoryListResp_Append(t *testing.T) {
	t.Parallel()

	list := HistoryListResp{}

	item1 := HistoryListItemResp{
		HistoryID: "hist-1",
		AgentID:   "agent-1",
	}
	item2 := HistoryListItemResp{
		HistoryID: "hist-2",
		AgentID:   "agent-2",
	}

	list = append(list, item1)
	list = append(list, item2)

	assert.Len(t, list, 2)
	assert.Equal(t, "hist-1", list[0].HistoryID)
	assert.Equal(t, "hist-2", list[1].HistoryID)
}

func TestHistoryListResp_SliceOperations(t *testing.T) {
	t.Parallel()

	list := HistoryListResp{
		{HistoryID: "hist-1", AgentID: "agent-1"},
		{HistoryID: "hist-2", AgentID: "agent-2"},
		{HistoryID: "hist-3", AgentID: "agent-3"},
	}

	// Test length
	assert.Len(t, list, 3)

	// Test slicing
	subList := list[1:3]
	assert.Len(t, subList, 2)
	assert.Equal(t, "hist-2", subList[0].HistoryID)
	assert.Equal(t, "hist-3", subList[1].HistoryID)

	// Test iteration
	count := 0

	for _, item := range list {
		assert.NotEmpty(t, item.HistoryID)

		count++
	}

	assert.Equal(t, 3, count)
}

func TestPermissionRange_Append(t *testing.T) {
	t.Parallel()

	pr := PermissionRange{}

	obj1 := PermissionRangeObject{
		ObjectID:   "obj-1",
		ObjectType: "type-1",
	}
	obj2 := PermissionRangeObject{
		ObjectID:   "obj-2",
		ObjectType: "type-2",
	}

	pr = append(pr, obj1)
	pr = append(pr, obj2)

	assert.Len(t, pr, 2)
	assert.Equal(t, "obj-1", pr[0].ObjectID)
	assert.Equal(t, "obj-2", pr[1].ObjectID)
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
			HistoryID:  "hist-" + string(rune('a'+i)),
			CreateTime: ts,
		}
		assert.Equal(t, ts, item.CreateTime)
	}
}

func TestPermissionRange_WithDuplicateObjects(t *testing.T) {
	t.Parallel()

	pr := PermissionRange{
		{
			ObjectID:   "obj-1",
			ObjectType: "type-a",
		},
		{
			ObjectID:   "obj-1",  // Same ID
			ObjectType: "type-a", // Same type
		},
		{
			ObjectID:   "obj-1",  // Same ID
			ObjectType: "type-b", // Different type
		},
	}

	assert.Len(t, pr, 3)
	assert.Equal(t, "obj-1", pr[0].ObjectID)
	assert.Equal(t, "obj-1", pr[1].ObjectID)
	assert.Equal(t, "obj-1", pr[2].ObjectID)
}

func TestHistoryListResp_Capacity(t *testing.T) {
	t.Parallel()

	list := make(HistoryListResp, 0, 100)

	assert.Len(t, list, 0)
	assert.NotNil(t, list)

	// Test that we can append up to capacity
	for i := 0; i < 100; i++ {
		list = append(list, HistoryListItemResp{
			HistoryID: "hist-" + string(rune(i)),
		})
	}

	assert.Len(t, list, 100)
}

func TestHistoryListItemResp_WithLongDescription(t *testing.T) {
	t.Parallel()

	longDesc := "This is a very long description that contains multiple words and should be stored properly in the AgentDesc field. It might contain special characters like 中文 and numbers like 12345."
	item := HistoryListItemResp{
		HistoryID: "hist-long",
		AgentDesc: longDesc,
	}

	assert.Equal(t, longDesc, item.AgentDesc)
	assert.Contains(t, item.AgentDesc, "中文")
	assert.Contains(t, item.AgentDesc, "12345")
}
