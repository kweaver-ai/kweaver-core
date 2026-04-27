package agentvo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPublishedAgentInfo(t *testing.T) {
	t.Parallel()

	info := NewPublishedAgentInfo()
	assert.NotNil(t, info)
	assert.Empty(t, info.PublishedBy)
	assert.Empty(t, info.PublishedByName)
	assert.Equal(t, int64(0), info.PublishedAt)
	assert.Empty(t, info.Profile)
	assert.Empty(t, info.Version)
	assert.Empty(t, info.Avatar)
}

func TestPublishedAgentInfo_LoadFromReleaseAgentPO(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		po        *dapo.PublishedJoinPo
		wantErr   bool
		checkInfo func(t *testing.T, info *PublishedAgentInfo)
	}{
		{
			name: "valid published agent po",
			po: &dapo.PublishedJoinPo{
				ReleasePartPo: dapo.ReleasePartPo{
					ReleaseID:   "release-1",
					PublishDesc: "Test publish",
					Version:     "v1.0.0",
					PublishedAt: 1000000,
					PublishedBy: "user-1",
					IsPmsCtrl:   1,
					PublishedToBeStruct: dapo.PublishedToBeStruct{
						IsAPIAgent:      1,
						IsWebSDKAgent:   0,
						IsSkillAgent:    1,
						IsDataFlowAgent: 0,
					},
				},
				DataAgentPo: dapo.DataAgentPo{
					ID:         "agent-1",
					Name:       "Test Agent",
					Profile:    strPtr("Test profile"),
					AvatarType: cdaenum.AvatarTypeBuiltIn,
					Avatar:     "🤖",
				},
			},
			wantErr: false,
			checkInfo: func(t *testing.T, info *PublishedAgentInfo) {
				assert.NotNil(t, info)
				assert.Equal(t, "user-1", info.PublishedBy)
				assert.Equal(t, int64(1000000), info.PublishedAt)
				assert.Equal(t, "v1.0.0", info.Version)
				assert.Equal(t, "Test profile", info.Profile)
				assert.Equal(t, "🤖", info.Avatar)
				assert.Equal(t, cdaenum.AvatarTypeBuiltIn, info.AvatarType)
				assert.Equal(t, 1, info.IsAPIAgent)
				assert.Equal(t, 0, info.IsWebSDKAgent)
				assert.Equal(t, 1, info.IsSkillAgent)
			},
		},
		{
			name: "minimal published agent po",
			po: &dapo.PublishedJoinPo{
				ReleasePartPo: dapo.ReleasePartPo{
					ReleaseID:   "release-2",
					PublishedAt: 2000000,
					PublishedBy: "user-2",
				},
			},
			wantErr: false,
			checkInfo: func(t *testing.T, info *PublishedAgentInfo) {
				assert.NotNil(t, info)
				assert.Equal(t, "user-2", info.PublishedBy)
				assert.Equal(t, int64(2000000), info.PublishedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info := NewPublishedAgentInfo()

			err := info.LoadFromReleaseAgentPO(tt.po)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.checkInfo != nil {
					tt.checkInfo(t, info)
				}
			}
		})
	}
}

func TestPublishedAgentInfo_Fields(t *testing.T) {
	t.Parallel()

	info := &PublishedAgentInfo{
		PublishedBy:     "user-1",
		PublishedByName: "User One",
		PublishedAt:     1000000,
		Profile:         "Agent profile",
		Version:         "v1.0",
		AvatarType:      cdaenum.AvatarTypeUserUploaded,
		Avatar:          "custom.png",
		PublishedToBeStruct: dapo.PublishedToBeStruct{
			IsAPIAgent:      1,
			IsWebSDKAgent:   1,
			IsSkillAgent:    0,
			IsDataFlowAgent: 1,
		},
	}

	assert.Equal(t, "user-1", info.PublishedBy)
	assert.Equal(t, "User One", info.PublishedByName)
	assert.Equal(t, int64(1000000), info.PublishedAt)
	assert.Equal(t, "Agent profile", info.Profile)
	assert.Equal(t, "v1.0", info.Version)
	assert.Equal(t, "custom.png", info.Avatar)
	assert.Equal(t, cdaenum.AvatarTypeUserUploaded, info.AvatarType)
	assert.Equal(t, 1, info.IsAPIAgent)
	assert.Equal(t, 1, info.IsWebSDKAgent)
	assert.Equal(t, 0, info.IsSkillAgent)
	assert.Equal(t, 1, info.IsDataFlowAgent)
}

func TestPublishUserInfo(t *testing.T) {
	t.Parallel()

	info := &PublishUserInfo{
		UserID:   "user-123",
		Username: "testuser",
	}

	assert.Equal(t, "user-123", info.UserID)
	assert.Equal(t, "testuser", info.Username)
}

// Helper function
func strPtr(s string) *string {
	return &s
}

func TestPublishedAgentInfo_LoadFromReleaseAgentPO_AllFields(t *testing.T) {
	t.Parallel()

	info := NewPublishedAgentInfo()
	po := &dapo.PublishedJoinPo{
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID:   "release-full",
			PublishDesc: "Full publish description",
			Version:     "v2.5.0",
			PublishedAt: 1700000000,
			PublishedBy: "admin-user",
			IsPmsCtrl:   1,
			PublishedToBeStruct: dapo.PublishedToBeStruct{
				IsAPIAgent:      1,
				IsWebSDKAgent:   1,
				IsSkillAgent:    1,
				IsDataFlowAgent: 1,
			},
		},
		DataAgentPo: dapo.DataAgentPo{
			ID:         "agent-full",
			Name:       "Full Agent",
			Profile:    strPtr("Full profile description"),
			AvatarType: cdaenum.AvatarTypeAIGenerated,
			Avatar:     "ai-generated.png",
		},
	}

	err := info.LoadFromReleaseAgentPO(po)
	require.NoError(t, err)
	assert.Equal(t, "release-full", po.ReleaseID)
	assert.Equal(t, "v2.5.0", info.Version)
	assert.Equal(t, "admin-user", info.PublishedBy)
	assert.Equal(t, int64(1700000000), info.PublishedAt)
	assert.Equal(t, "Full profile description", info.Profile)
	assert.Equal(t, "ai-generated.png", info.Avatar)
	assert.Equal(t, cdaenum.AvatarTypeAIGenerated, info.AvatarType)
	assert.Equal(t, 1, info.IsAPIAgent)
	assert.Equal(t, 1, info.IsWebSDKAgent)
	assert.Equal(t, 1, info.IsSkillAgent)
	assert.Equal(t, 1, info.IsDataFlowAgent)
}

func TestPublishedAgentInfo_LoadFromReleaseAgentPO_WithNilPO(t *testing.T) {
	t.Parallel()

	info := NewPublishedAgentInfo()
	// LoadFromReleaseAgentPO with nil po will panic on CopyStructUseJSON
	assert.Panics(t, func() {
		info.LoadFromReleaseAgentPO(nil) //nolint:errcheck
	})
}

func TestPublishedAgentInfo_LoadFromReleaseAgentPO_WithEmptyValues(t *testing.T) {
	t.Parallel()

	info := NewPublishedAgentInfo()
	po := &dapo.PublishedJoinPo{
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID:   "",
			Version:     "",
			PublishedBy: "",
		},
	}

	err := info.LoadFromReleaseAgentPO(po)
	require.NoError(t, err)
	assert.Empty(t, info.Version)
	assert.Empty(t, info.PublishedBy)
}

func TestPublishedAgentInfo_EmptyInitStruct(t *testing.T) {
	t.Parallel()

	info := &PublishedAgentInfo{}
	assert.Empty(t, info.PublishedBy)
	assert.Empty(t, info.Version)
	assert.Empty(t, info.Profile)
	assert.Empty(t, info.Avatar)
	assert.Equal(t, int64(0), info.PublishedAt)
}

func TestPublishUserInfo_Empty(t *testing.T) {
	t.Parallel()

	info := &PublishUserInfo{}
	assert.Empty(t, info.UserID)
	assert.Empty(t, info.Username)
}

func TestPublishUserInfo_WithChineseCharacters(t *testing.T) {
	t.Parallel()

	info := &PublishUserInfo{
		UserID:   "用户-123",
		Username: "中文用户名",
	}

	assert.Equal(t, "用户-123", info.UserID)
	assert.Equal(t, "中文用户名", info.Username)
}

func TestPublishedAgentInfo_LoadFromReleaseAgentPO_AllAvatarTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		avatarType cdaenum.AvatarType
	}{
		{"built-in avatar", cdaenum.AvatarTypeBuiltIn},
		{"user uploaded avatar", cdaenum.AvatarTypeUserUploaded},
		{"AI generated avatar", cdaenum.AvatarTypeAIGenerated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info := NewPublishedAgentInfo()
			po := &dapo.PublishedJoinPo{
				ReleasePartPo: dapo.ReleasePartPo{
					ReleaseID:   "release-test",
					PublishedBy: "user-1",
				},
				DataAgentPo: dapo.DataAgentPo{
					ID:         "agent-test",
					Name:       "Test Agent",
					AvatarType: tt.avatarType,
					Avatar:     "test-avatar.png",
				},
			}

			err := info.LoadFromReleaseAgentPO(po)
			require.NoError(t, err)
			assert.Equal(t, tt.avatarType, info.AvatarType)
			assert.Equal(t, "test-avatar.png", info.Avatar)
		})
	}
}
