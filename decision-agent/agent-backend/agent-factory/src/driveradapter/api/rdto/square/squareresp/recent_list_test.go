package squareresp

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/publishvo"
	"github.com/stretchr/testify/assert"
)

func TestRecentListAgentResp_Type(t *testing.T) {
	t.Parallel()

	// RecentListAgentResp is a slice type
	var list RecentListAgentResp

	assert.Nil(t, list)
	assert.IsType(t, RecentListAgentResp{}, list)
}

func TestRecentListAgentResp_Empty(t *testing.T) {
	t.Parallel()

	list := RecentListAgentResp{}

	assert.Empty(t, list)
	assert.Len(t, list, 0)
}

func TestRecentListAgentResp_WithMultipleItems(t *testing.T) {
	t.Parallel()

	list := RecentListAgentResp{
		{
			CategoryId:      "cat-1",
			CategoryName:    "Category 1",
			Version:         "1.0.0",
			Description:     "Description 1",
			PublishedAt:     1640995200000,
			PublishedBy:     "user-1",
			PublishedByName: "User One",
			PublishInfo:     &publishvo.ListPublishInfo{},
		},
		{
			CategoryId:      "cat-2",
			CategoryName:    "Category 2",
			Version:         "2.0.0",
			Description:     "Description 2",
			PublishedAt:     1641081600000,
			PublishedBy:     "user-2",
			PublishedByName: "User Two",
			PublishInfo:     &publishvo.ListPublishInfo{},
		},
	}

	assert.Len(t, list, 2)
	assert.Equal(t, "cat-1", list[0].CategoryId)
	assert.Equal(t, "cat-2", list[1].CategoryId)
}

func TestRecentAgentListItem_StructFields(t *testing.T) {
	t.Parallel()

	item := RecentAgentListItem{
		CategoryId:      "cat-123",
		CategoryName:    "TestCategory",
		Version:         "1.5.0",
		Description:     "Test agent description",
		PublishedAt:     1640995200000,
		PublishedBy:     "user-456",
		PublishedByName: "Test User",
		PublishInfo:     &publishvo.ListPublishInfo{},
	}

	assert.Equal(t, "cat-123", item.CategoryId)
	assert.Equal(t, "TestCategory", item.CategoryName)
	assert.Equal(t, "1.5.0", item.Version)
	assert.Equal(t, "Test agent description", item.Description)
	assert.Equal(t, int64(1640995200000), item.PublishedAt)
	assert.Equal(t, "user-456", item.PublishedBy)
	assert.Equal(t, "Test User", item.PublishedByName)
	assert.NotNil(t, item.PublishInfo)
}

func TestRecentAgentListItem_EmptyValues(t *testing.T) {
	t.Parallel()

	item := RecentAgentListItem{}

	assert.Empty(t, item.CategoryId)
	assert.Empty(t, item.CategoryName)
	assert.Empty(t, item.Version)
	assert.Empty(t, item.Description)
	assert.Equal(t, int64(0), item.PublishedAt)
	assert.Empty(t, item.PublishedBy)
	assert.Empty(t, item.PublishedByName)
	assert.Nil(t, item.PublishInfo)
}

func TestRecentAgentListItem_NilPublishInfo(t *testing.T) {
	t.Parallel()

	item := RecentAgentListItem{
		CategoryId:   "cat-nil",
		CategoryName: "NilCategory",
		PublishInfo:  nil,
	}

	assert.Nil(t, item.PublishInfo)
}

func TestRecentListAgentResp_Append(t *testing.T) {
	t.Parallel()

	list := RecentListAgentResp{}

	// Append items
	list = append(list, RecentAgentListItem{
		CategoryId:   "cat-1",
		CategoryName: "Category 1",
		Version:      "1.0.0",
		PublishInfo:  &publishvo.ListPublishInfo{},
	})

	list = append(list, RecentAgentListItem{
		CategoryId:   "cat-2",
		CategoryName: "Category 2",
		Version:      "2.0.0",
		PublishInfo:  &publishvo.ListPublishInfo{},
	})

	assert.Len(t, list, 2)
	assert.Equal(t, "Category 1", list[0].CategoryName)
	assert.Equal(t, "Category 2", list[1].CategoryName)
}

func TestRecentAgentListItem_TimestampComparison(t *testing.T) {
	t.Parallel()

	olderItem := RecentAgentListItem{
		CategoryId:  "cat-older",
		PublishedAt: 1640995200000, // Earlier
	}

	newerItem := RecentAgentListItem{
		CategoryId:  "cat-newer",
		PublishedAt: 1641081600000, // Later
	}

	assert.True(t, newerItem.PublishedAt > olderItem.PublishedAt)
	assert.True(t, olderItem.PublishedAt < newerItem.PublishedAt)
}

func TestRecentAgentListItem_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	item := RecentAgentListItem{
		CategoryId:      "cat-中文-123",
		CategoryName:    "分类名称",
		Version:         "1.0.0-β",
		Description:     "This is a description with \"quotes\" and 'apostrophes'",
		PublishedByName: "用户名称",
		PublishInfo:     &publishvo.ListPublishInfo{},
	}

	assert.Equal(t, "cat-中文-123", item.CategoryId)
	assert.Equal(t, "分类名称", item.CategoryName)
	assert.Equal(t, "1.0.0-β", item.Version)
	assert.Contains(t, item.Description, "quotes")
	assert.Equal(t, "用户名称", item.PublishedByName)
}

func TestRecentListAgentResp_SliceOperations(t *testing.T) {
	t.Parallel()

	list := RecentListAgentResp{
		{CategoryId: "cat-1", PublishInfo: &publishvo.ListPublishInfo{}},
		{CategoryId: "cat-2", PublishInfo: &publishvo.ListPublishInfo{}},
		{CategoryId: "cat-3", PublishInfo: &publishvo.ListPublishInfo{}},
	}

	// Test length
	assert.Len(t, list, 3)

	// Test iteration
	count := 0

	for _, item := range list {
		assert.NotEmpty(t, item.CategoryId)

		count++
	}

	assert.Equal(t, 3, count)

	// Test slicing
	subList := list[1:3]
	assert.Len(t, subList, 2)
	assert.Equal(t, "cat-2", subList[0].CategoryId)
	assert.Equal(t, "cat-3", subList[1].CategoryId)
}

func TestRecentAgentListItem_VersionFormats(t *testing.T) {
	t.Parallel()

	versions := []string{
		"1.0.0",
		"2.0.0-alpha",
		"3.0.0-beta.2+build.123",
		"4.1.3",
	}

	for _, version := range versions {
		item := RecentAgentListItem{
			CategoryId:  "cat-version",
			Version:     version,
			PublishInfo: &publishvo.ListPublishInfo{},
		}

		assert.Equal(t, version, item.Version)
	}
}
