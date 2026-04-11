package drivenadapters

import "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"

const (
	mockUserManagementUserID         = "mock-user-id"
	mockUserManagementUserName       = "mock-user"
	mockUserManagementUserEmail      = "mock@example.com"
	mockUserManagementUserPhone      = "12345678901"
	mockUserManagementGroupID        = "mock-group-id"
	mockUserManagementGroupName      = "mock-group"
	mockUserManagementDepartmentID   = "mock-department-id"
	mockUserManagementDepartmentName = "mock-department"
	mockUserManagementContactorID    = "mock-contactor-id"
	mockUserManagementContactorName  = "mock-contactor"
	mockUserManagementAppID          = "mock-app-id-12345"
	mockUserManagementAppName        = "content-automation"
)

type mockUserManagement struct{}

func (m *mockUserManagement) GetUserMailList(userIDs []string) ([]string, error) {
	return []string{}, nil
}

func (m *mockUserManagement) GetDeptMailList(departmentIDs []string) ([]string, error) {
	return []string{}, nil
}

func (m *mockUserManagement) GetGroupUserList(groupID []string) ([]string, error) {
	return []string{mockUserManagementUserID}, nil
}

func (m *mockUserManagement) BatchGetUserInfo(userIDs []string) ([]UserInfo, error) {
	users := make([]UserInfo, 0, len(userIDs))
	users = append(users, m.newMockUserInfo())
	return users, nil
}

func (m *mockUserManagement) GetUserInfo(userID string) (UserInfo, error) {
	return m.newMockUserInfo(), nil
}

func (m *mockUserManagement) GetNameByAccessorIDs(accessorIDs map[string]string) (map[string]string, error) {
	names := make(map[string]string, len(accessorIDs))
	names[mockUserManagementUserID] = common.User.ToString()
	return names, nil
}

func (m *mockUserManagement) GetUserAccessorIDs(userID string) ([]string, error) {
	return []string{mockUserManagementUserID}, nil
}

func (m *mockUserManagement) RegisterInternalAccount(name, password string) (string, error) {
	return mockUserManagementAppID, nil
}

func (m *mockUserManagement) QueryInternalAccount(id string) (string, error) {
	return mockUserManagementAppName, nil
}

func (m *mockUserManagement) CreateInternalGroup() (string, error) {
	return mockUserManagementGroupID, nil
}

func (m *mockUserManagement) DeleteInternalGroup(ids []string) error {
	return nil
}

func (m *mockUserManagement) UpdateInternalGroupMember(groupID string, userIDs []string) error {
	return nil
}

func (m *mockUserManagement) GetInternalGroupMembers(groupID string) ([]string, error) {
	return []string{mockUserManagementUserID}, nil
}

func (m *mockUserManagement) GetDepartmentInfo(departmentID string) (*DepartInfo, error) {
	return &DepartInfo{
		DepartmentID: mockUserManagementDepartmentID,
		Name:         mockUserManagementDepartmentName,
		ParentDeps:   []DepInfo{},
	}, nil
}

func (m *mockUserManagement) GetDepartments(level int) (*[]DepInfo, error) {
	departments := []DepInfo{{
		ID:   mockUserManagementDepartmentID,
		Name: mockUserManagementDepartmentName,
		Type: common.Department.ToString(),
	}}
	return &departments, nil
}

func (m *mockUserManagement) GetDepartmentMemberIDs(deptID string) (*DepartmentMembers, error) {
	return &DepartmentMembers{
		UserIDs:       []string{mockUserManagementUserID},
		DepartmentIDs: []string{},
	}, nil
}

func (m *mockUserManagement) BatchGetNames(data map[string][]string) (*NamesInfo, error) {
	return &NamesInfo{
		UserNames:       mockUserAttributes([]string{mockUserManagementUserID}, mockUserManagementUserName),
		GroupNames:      mockUserAttributes([]string{mockUserManagementGroupID}, mockUserManagementGroupName),
		DepartmentNames: mockUserAttributes([]string{mockUserManagementDepartmentID}, mockUserManagementDepartmentName),
		ContactorNames:  mockUserAttributes([]string{mockUserManagementContactorID}, mockUserManagementContactorName),
		AppNames:        mockUserAttributes([]string{mockUserManagementAppID}, mockUserManagementAppName),
	}, nil
}

func (m *mockUserManagement) GetAppAccountInfo(mockUserManagementAppIDappID string) (AppAccountInfo, error) {
	return AppAccountInfo{
		AppID: mockUserManagementAppID,
		Name:  mockUserManagementAppName,
	}, nil
}

func (m *mockUserManagement) GetUserInfoByType(accessorID, accessorType string) (UserInfo, error) {
	switch accessorType {
	case common.Group.ToString():
		return UserInfo{
			UserID:      mockUserManagementGroupID,
			UserName:    mockUserManagementGroupName,
			AccountType: common.APP.ToString(),
		}, nil
	case common.Department.ToString():
		return UserInfo{
			UserID:      mockUserManagementDepartmentID,
			UserName:    mockUserManagementDepartmentName,
			AccountType: common.APP.ToString(),
		}, nil
	case common.Contactor.ToString():
		return UserInfo{
			UserID:      mockUserManagementContactorID,
			UserName:    mockUserManagementContactorName,
			AccountType: common.APP.ToString(),
		}, nil
	case common.APP.ToString():
		return UserInfo{
			UserID:      mockUserManagementAppID,
			UserName:    mockUserManagementAppName,
			AccountType: common.APP.ToString(),
		}, nil
	default:
		return m.GetUserInfo(accessorID)
	}
}

func (m *mockUserManagement) IsApp(appID string) (bool, error) {
	return appID == mockUserManagementAppID, nil
}

func (m *mockUserManagement) newMockUserInfo() UserInfo {
	parentDeps := []interface{}{
		[]interface{}{
			map[string]interface{}{
				"id":   mockUserManagementDepartmentID,
				"name": mockUserManagementDepartmentName,
			},
		},
	}

	userInfo := UserInfo{
		UserID:       mockUserManagementUserID,
		UserName:     mockUserManagementUserName,
		ParentDeps:   parentDeps,
		ParentDepIDs: []string{mockUserManagementDepartmentID},
		CsfLevel:     1,
		Roles:        []string{"super_admin"},
		Telephone:    mockUserManagementUserPhone,
		Email:        mockUserManagementUserEmail,
		Enabled:      true,
		CustomAttr: map[string]interface{}{
			"is_knowledge": "1",
		},
		IsKnowledge: true,
	}
	userInfo.SetFullDepPath()
	return userInfo
}

func mockUserAttributes(ids []string, fallbackName string) []*UserAttribute {
	attrs := make([]*UserAttribute, 0, len(ids))
	for _, id := range ids {
		attrs = append(attrs, &UserAttribute{
			ID:   id,
			Name: fallbackName,
		})
	}
	return attrs
}

func mockAccessorName(accessorID, accessorType string) string {
	switch accessorType {
	case common.Group.ToString():
		return mockUserManagementGroupName
	case common.Department.ToString():
		return mockUserManagementDepartmentName
	case common.Contactor.ToString():
		return mockUserManagementContactorName
	case common.APP.ToString():
		return mockUserManagementAppName
	default:
		if accessorID == "" {
			return mockUserManagementUserName
		}
		return mockUserManagementUserName
	}
}
