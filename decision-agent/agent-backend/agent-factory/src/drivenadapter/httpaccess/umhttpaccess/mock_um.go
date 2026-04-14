package umhttpaccess

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc"
)

type mockUmHttpAcc struct{}

var _ iumacc.UmHttpAcc = &mockUmHttpAcc{}

func NewMockUmHttpAcc() iumacc.UmHttpAcc {
	return &mockUmHttpAcc{}
}

func (m *mockUmHttpAcc) GetUserInfo(_ context.Context, dto *umarg.GetUserInfoArgDto) (umcmp.UserInfoMap, error) {
	infoMap := make(umcmp.UserInfoMap, len(dto.UserIds))

	for _, userID := range dto.UserIds {
		if userID == "" {
			continue
		}

		infoMap[userID] = &umcmp.UserInfo{
			Id:      userID,
			Name:    userID,
			Enabled: true,
		}
	}

	return infoMap, nil
}

func (m *mockUmHttpAcc) GetAppIDNameKv(_ context.Context, appIDs []string) (map[string]string, error) {
	appNameMap := make(map[string]string, len(appIDs))
	for _, appID := range appIDs {
		if appID == "" {
			continue
		}

		appNameMap[appID] = appID
	}

	return appNameMap, nil
}

func (m *mockUmHttpAcc) GetUserDeptIDs(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}

func (m *mockUmHttpAcc) GetUserUserGroupIDs(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}

func (m *mockUmHttpAcc) GetDeptInfoMapByIDs(_ context.Context, deptIDs []string) (map[string]*umtypes.DepartmentInfo, error) {
	deptInfoMap := make(map[string]*umtypes.DepartmentInfo, len(deptIDs))
	for _, deptID := range deptIDs {
		if deptID == "" {
			continue
		}

		deptInfoMap[deptID] = &umtypes.DepartmentInfo{
			DepartmentId: deptID,
			Name:         deptID,
		}
	}

	return deptInfoMap, nil
}

func (m *mockUmHttpAcc) GetUserDep(_ context.Context, _ string) ([][]umcmp.ObjectBaseInfo, error) {
	return [][]umcmp.ObjectBaseInfo{}, nil
}

func (m *mockUmHttpAcc) GetOsnNames(_ context.Context, dto *umarg.GetOsnArgDto) (*umtypes.OsnInfoMapS, error) {
	ret := umtypes.NewOsnInfoMapS()

	for _, userID := range dto.UserIDs {
		if userID == "" {
			continue
		}

		ret.UserNameMap[userID] = userID
	}

	for _, departmentID := range dto.DepartmentIDs {
		if departmentID == "" {
			continue
		}

		ret.DepartmentNameMap[departmentID] = departmentID
	}

	for _, groupID := range dto.GroupIDs {
		if groupID == "" {
			continue
		}

		ret.GroupNameMap[groupID] = groupID
	}

	for _, appID := range dto.AppIDs {
		if appID == "" {
			continue
		}

		ret.AppNameMap[appID] = appID
	}

	return ret, nil
}

func (m *mockUmHttpAcc) GetUserIDNameMap(ctx context.Context, userIDs []string) (map[string]string, error) {
	ret, err := m.GetOsnNames(ctx, &umarg.GetOsnArgDto{UserIDs: userIDs})
	if err != nil {
		return nil, err
	}

	return ret.UserNameMap, nil
}

func (m *mockUmHttpAcc) GetSingleUserName(_ context.Context, userID string) (string, error) {
	return userID, nil
}
