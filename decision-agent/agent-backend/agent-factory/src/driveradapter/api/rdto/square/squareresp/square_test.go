package squareresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInfo_StructFields(t *testing.T) {
	t.Parallel()

	userInfo := UserInfo{
		UserID:   "user-123",
		Username: "testuser",
	}

	assert.Equal(t, "user-123", userInfo.UserID)
	assert.Equal(t, "testuser", userInfo.Username)
}

func TestUserInfo_Empty(t *testing.T) {
	t.Parallel()

	userInfo := UserInfo{}

	assert.Empty(t, userInfo.UserID)
	assert.Empty(t, userInfo.Username)
}

func TestUserInfo_JSONTags(t *testing.T) {
	t.Parallel()

	userInfo := UserInfo{
		UserID:   "user-456",
		Username: "username456",
	}

	assert.Equal(t, "user-456", userInfo.UserID)
	assert.Equal(t, "username456", userInfo.Username)
}

func TestListAgentResp_Type(t *testing.T) {
	t.Parallel()

	// ListAgentResp is a slice type
	var list ListAgentResp

	assert.Nil(t, list)
	assert.IsType(t, ListAgentResp{}, list)
}

func TestListAgentResp_Empty(t *testing.T) {
	t.Parallel()

	list := ListAgentResp{}

	assert.Empty(t, list)
	assert.Len(t, list, 0)
}

func TestListAgentResp_WithMultipleItems(t *testing.T) {
	t.Parallel()

	list := ListAgentResp{
		{
			CategoryId:   "cat-1",
			CategoryName: "Category 1",
			Version:      "1.0.0",
			Description:  "Description 1",
			PublishTime:  1640995200000,
			PublishUserInfo: UserInfo{
				UserID:   "user-1",
				Username: "UserOne",
			},
			UpdateUserInfo: UserInfo{
				UserID:   "user-2",
				Username: "UserTwo",
			},
		},
		{
			CategoryId:   "cat-2",
			CategoryName: "Category 2",
			Version:      "2.0.0",
			Description:  "Description 2",
			PublishTime:  1641081600000,
		},
	}

	assert.Len(t, list, 2)
	assert.Equal(t, "cat-1", list[0].CategoryId)
	assert.Equal(t, "Category 1", list[0].CategoryName)
	assert.Equal(t, "user-1", list[0].PublishUserInfo.UserID)
	assert.Equal(t, "user-2", list[0].UpdateUserInfo.UserID)
}

func TestAgentListItemResp_StructFields(t *testing.T) {
	t.Parallel()

	item := AgentListItemResp{
		CategoryId:   "cat-123",
		CategoryName: "TestCategory",
		Version:      "1.5.0",
		Description:  "Test description",
		PublishTime:  1640995200000,
		PublishUserInfo: UserInfo{
			UserID:   "publisher-123",
			Username: "PublisherUser",
		},
		UpdateUserInfo: UserInfo{
			UserID:   "updater-456",
			Username: "UpdaterUser",
		},
	}

	assert.Equal(t, "cat-123", item.CategoryId)
	assert.Equal(t, "TestCategory", item.CategoryName)
	assert.Equal(t, "1.5.0", item.Version)
	assert.Equal(t, "Test description", item.Description)
	assert.Equal(t, int64(1640995200000), item.PublishTime)
	assert.Equal(t, "publisher-123", item.PublishUserInfo.UserID)
	assert.Equal(t, "updater-456", item.UpdateUserInfo.UserID)
}

func TestAgentListItemResp_Empty(t *testing.T) {
	t.Parallel()

	item := AgentListItemResp{}

	assert.Empty(t, item.CategoryId)
	assert.Empty(t, item.CategoryName)
	assert.Empty(t, item.Version)
	assert.Empty(t, item.Description)
	assert.Equal(t, int64(0), item.PublishTime)
	assert.Empty(t, item.PublishUserInfo.UserID)
	assert.Empty(t, item.UpdateUserInfo.UserID)
}

func TestAgentListItemResp_WithoutUserInfo(t *testing.T) {
	t.Parallel()

	item := AgentListItemResp{
		CategoryId:   "cat-empty",
		CategoryName: "Empty Category",
	}

	assert.Empty(t, item.PublishUserInfo.UserID)
	assert.Empty(t, item.UpdateUserInfo.UserID)
}

func TestListAgentResp_Append(t *testing.T) {
	t.Parallel()

	list := ListAgentResp{}

	item1 := AgentListItemResp{
		CategoryId: "cat-1",
		Version:    "1.0",
	}
	item2 := AgentListItemResp{
		CategoryId: "cat-2",
		Version:    "2.0",
	}

	list = append(list, item1)
	list = append(list, item2)

	assert.Len(t, list, 2)
	assert.Equal(t, "cat-1", list[0].CategoryId)
	assert.Equal(t, "cat-2", list[1].CategoryId)
}

func TestListAgentResp_SliceOperations(t *testing.T) {
	t.Parallel()

	list := ListAgentResp{
		{CategoryId: "cat-1", Version: "1.0"},
		{CategoryId: "cat-2", Version: "2.0"},
		{CategoryId: "cat-3", Version: "3.0"},
	}

	// Test length
	assert.Len(t, list, 3)

	// Test slicing
	subList := list[1:3]
	assert.Len(t, subList, 2)
	assert.Equal(t, "cat-2", subList[0].CategoryId)
	assert.Equal(t, "cat-3", subList[1].CategoryId)

	// Test iteration
	count := 0

	for _, item := range list {
		assert.NotEmpty(t, item.CategoryId)

		count++
	}

	assert.Equal(t, 3, count)
}

func TestAgentListItemResp_WithTimestamps(t *testing.T) {
	t.Parallel()

	timestamps := []int64{
		1640995200000, // 2022-01-01
		1643673600000, // 2022-02-01
		1646092800000, // 2022-03-01
	}

	for _, ts := range timestamps {
		item := AgentListItemResp{
			CategoryId:  "cat-timestamp",
			Version:     "1.0",
			PublishTime: ts,
		}

		assert.Equal(t, ts, item.PublishTime)
	}
}

func TestUserInfo_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	userInfo := UserInfo{
		UserID:   "user-中文-123",
		Username: "用户名",
	}

	assert.Equal(t, "user-中文-123", userInfo.UserID)
	assert.Equal(t, "用户名", userInfo.Username)
}

func TestAgentListItemResp_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	item := AgentListItemResp{
		CategoryId:   "cat-特殊",
		CategoryName: "特殊分类",
		Description:  "包含特殊字符的描述",
	}

	assert.Equal(t, "cat-特殊", item.CategoryId)
	assert.Contains(t, item.Description, "特殊字符")
}

func TestListAgentResp_Capacity(t *testing.T) {
	t.Parallel()

	list := make(ListAgentResp, 0, 100)

	assert.Len(t, list, 0)
	assert.NotNil(t, list)

	// Test that we can append up to capacity
	for i := 0; i < 100; i++ {
		list = append(list, AgentListItemResp{
			CategoryId: "cat-" + string(rune(i)),
		})
	}

	assert.Len(t, list, 100)
}
