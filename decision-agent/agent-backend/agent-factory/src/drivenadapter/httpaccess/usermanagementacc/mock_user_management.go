package usermanagementacc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iusermanagementacc"
)

type mockClient struct{}

func NewMockClient() iusermanagementacc.UserMgnt {
	return &mockClient{}
}

func (m *mockClient) GetUserInfoByUserID(_ context.Context, userIDs []string, _ []string) (map[string]*iusermanagementacc.UserInfo, error) {
	usersInfo := make(map[string]*iusermanagementacc.UserInfo, len(userIDs))

	for _, userID := range userIDs {
		if userID == "" {
			continue
		}

		usersInfo[userID] = &iusermanagementacc.UserInfo{
			Name: userID,
			ID:   userID,
		}
	}

	return usersInfo, nil
}
