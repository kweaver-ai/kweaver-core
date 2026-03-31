package umcmp

import (
	"testing"
)

func TestObjectBaseInfo(t *testing.T) {
	t.Parallel()

	t.Run("valid object base info", func(t *testing.T) {
		t.Parallel()

		info := ObjectBaseInfo{
			ID:   "test-id",
			Name: "test-name",
			Type: "test-type",
		}

		if info.ID != "test-id" {
			t.Errorf("Expected ID to be 'test-id', got '%s'", info.ID)
		}

		if info.Name != "test-name" {
			t.Errorf("Expected Name to be 'test-name', got '%s'", info.Name)
		}

		if info.Type != "test-type" {
			t.Errorf("Expected Type to be 'test-type', got '%s'", info.Type)
		}
	})

	t.Run("empty object base info", func(t *testing.T) {
		t.Parallel()

		info := ObjectBaseInfo{}

		if info.ID != "" {
			t.Errorf("Expected ID to be empty, got '%s'", info.ID)
		}

		if info.Name != "" {
			t.Errorf("Expected Name to be empty, got '%s'", info.Name)
		}

		if info.Type != "" {
			t.Errorf("Expected Type to be empty, got '%s'", info.Type)
		}
	})
}

func TestGroupInfo(t *testing.T) {
	t.Parallel()

	t.Run("valid group info", func(t *testing.T) {
		t.Parallel()

		group := GroupInfo{
			ID:    "group-123",
			Name:  "Test Group",
			Notes: "Test notes",
		}

		if group.ID != "group-123" {
			t.Errorf("Expected ID to be 'group-123', got '%s'", group.ID)
		}

		if group.Name != "Test Group" {
			t.Errorf("Expected Name to be 'Test Group', got '%s'", group.Name)
		}

		if group.Notes != "Test notes" {
			t.Errorf("Expected Notes to be 'Test notes', got '%s'", group.Notes)
		}
	})

	t.Run("empty group info", func(t *testing.T) {
		t.Parallel()

		group := GroupInfo{}

		if group.ID != "" {
			t.Errorf("Expected ID to be empty, got '%s'", group.ID)
		}

		if group.Name != "" {
			t.Errorf("Expected Name to be empty, got '%s'", group.Name)
		}

		if group.Notes != "" {
			t.Errorf("Expected Notes to be empty, got '%s'", group.Notes)
		}
	})
}

func TestUserInfo_GroupIDs(t *testing.T) {
	t.Parallel()

	t.Run("with groups", func(t *testing.T) {
		t.Parallel()

		userInfo := &UserInfo{
			Id:   "user-123",
			Name: "Test User",
			Groups: []*GroupInfo{
				{ID: "group-1", Name: "Group 1"},
				{ID: "group-2", Name: "Group 2"},
				{ID: "group-3", Name: "Group 3"},
			},
		}

		groupIDs := userInfo.GroupIDs()

		if len(groupIDs) != 3 {
			t.Errorf("Expected 3 group IDs, got %d", len(groupIDs))
		}

		expectedIDs := []string{"group-1", "group-2", "group-3"}
		for i, id := range groupIDs {
			if id != expectedIDs[i] {
				t.Errorf("Expected group ID %d to be '%s', got '%s'", i, expectedIDs[i], id)
			}
		}
	})

	t.Run("with no groups", func(t *testing.T) {
		t.Parallel()

		userInfo := &UserInfo{
			Id:     "user-123",
			Name:   "Test User",
			Groups: []*GroupInfo{},
		}

		groupIDs := userInfo.GroupIDs()

		if len(groupIDs) != 0 {
			t.Errorf("Expected 0 group IDs, got %d", len(groupIDs))
		}
	})

	t.Run("with nil groups", func(t *testing.T) {
		t.Parallel()

		userInfo := &UserInfo{
			Id:     "user-123",
			Name:   "Test User",
			Groups: nil,
		}

		groupIDs := userInfo.GroupIDs()

		if len(groupIDs) != 0 {
			t.Errorf("Expected 0 group IDs, got %d", len(groupIDs))
		}
	})
}

func TestUserInfo(t *testing.T) {
	t.Parallel()

	t.Run("full user info", func(t *testing.T) {
		t.Parallel()

		userInfo := &UserInfo{
			Id:   "user-123",
			Name: "Test User",
			ParentDeps: [][]ObjectBaseInfo{
				{
					{ID: "dep-1", Name: "Department 1", Type: "department"},
				},
			},
			Enabled: true,
			Roles:   []string{"admin", "user"},
			Account: "testuser",
			Groups: []*GroupInfo{
				{ID: "group-1", Name: "Group 1"},
			},
		}

		if userInfo.Id != "user-123" {
			t.Errorf("Expected Id to be 'user-123', got '%s'", userInfo.Id)
		}

		if userInfo.Name != "Test User" {
			t.Errorf("Expected Name to be 'Test User', got '%s'", userInfo.Name)
		}

		if !userInfo.Enabled {
			t.Error("Expected Enabled to be true")
		}

		if len(userInfo.Roles) != 2 {
			t.Errorf("Expected 2 roles, got %d", len(userInfo.Roles))
		}

		if userInfo.Account != "testuser" {
			t.Errorf("Expected Account to be 'testuser', got '%s'", userInfo.Account)
		}
	})

	t.Run("minimal user info", func(t *testing.T) {
		t.Parallel()

		userInfo := &UserInfo{
			Id:   "user-456",
			Name: "Minimal User",
		}

		if userInfo.Id != "user-456" {
			t.Errorf("Expected Id to be 'user-456', got '%s'", userInfo.Id)
		}

		if userInfo.Enabled {
			t.Error("Expected Enabled to be false for zero value")
		}

		if userInfo.Roles != nil {
			t.Error("Expected Roles to be nil for zero value")
		}
	})
}

func TestUserInfos(t *testing.T) {
	t.Parallel()

	t.Run("create user infos slice", func(t *testing.T) {
		t.Parallel()

		userInfos := UserInfos{
			&UserInfo{Id: "user-1", Name: "User 1"},
			&UserInfo{Id: "user-2", Name: "User 2"},
		}

		if len(userInfos) != 2 {
			t.Errorf("Expected 2 user infos, got %d", len(userInfos))
		}
	})

	t.Run("nil user infos", func(t *testing.T) {
		t.Parallel()

		var userInfos UserInfos

		if len(userInfos) != 0 {
			t.Errorf("Expected 0 user infos, got %d", len(userInfos))
		}
	})
}

func TestUserInfoMap(t *testing.T) {
	t.Parallel()

	t.Run("create user info map", func(t *testing.T) {
		t.Parallel()

		userMap := UserInfoMap{
			"user-1": {Id: "user-1", Name: "User 1"},
			"user-2": {Id: "user-2", Name: "User 2"},
		}

		if len(userMap) != 2 {
			t.Errorf("Expected 2 entries in map, got %d", len(userMap))
		}

		user, ok := userMap["user-1"]
		if !ok {
			t.Error("Expected user-1 to exist in map")
		}

		if user.Name != "User 1" {
			t.Errorf("Expected name to be 'User 1', got '%s'", user.Name)
		}
	})

	t.Run("nil user info map", func(t *testing.T) {
		t.Parallel()

		var userMap UserInfoMap

		if len(userMap) != 0 {
			t.Errorf("Expected 0 entries in map, got %d", len(userMap))
		}
	})
}

func TestUmNotFound(t *testing.T) {
	t.Parallel()

	t.Run("constant value", func(t *testing.T) {
		t.Parallel()

		if UmNotFound != 404019001 {
			t.Errorf("Expected UmNotFound to be 404019001, got %d", UmNotFound)
		}
	})
}
