package releasesvc

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCompareVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "v1 greater than v2 - numeric",
			v1:       "v100",
			v2:       "v50",
			expected: 50, // 100 - 50
		},
		{
			name:     "v2 greater than v1 - numeric",
			v1:       "v10",
			v2:       "v20",
			expected: -10, // 10 - 20
		},
		{
			name:     "equal versions - numeric",
			v1:       "v100",
			v2:       "v100",
			expected: 0,
		},
		{
			name:     "equal versions - no v prefix",
			v1:       "100",
			v2:       "100",
			expected: 0,
		},
		{
			name:     "v1 greater than v2 - string comparison",
			v1:       "v2.0",
			v2:       "v1.0",
			expected: 1, // "2.0" > "1.0" lexicographically
		},
		{
			name:     "v2 greater than v1 - string comparison",
			v1:       "v1.0",
			v2:       "v2.0",
			expected: -1,
		},
		{
			name:     "equal versions - string comparison",
			v1:       "v1.0",
			v2:       "v1.0",
			expected: 0,
		},
		{
			name:     "mixed - v1 numeric, v2 string",
			v1:       "100",
			v2:       "v2.0",
			expected: -1, // "100" < "2.0" lexicographically
		},
		{
			name:     "mixed - v1 string, v2 numeric",
			v1:       "v1.0",
			v2:       "100",
			expected: -1, // "1.0" < "100" lexicographically
		},
		{
			name:     "v1 with v prefix, v2 without",
			v1:       "v100",
			v2:       "100",
			expected: 0,
		},
		{
			name:     "v2 with v prefix, v1 without",
			v1:       "100",
			v2:       "v100",
			expected: 0,
		},
		{
			name:     "both without v prefix - numeric",
			v1:       "200",
			v2:       "100",
			expected: 100, // 200 - 100
		},
		{
			name:     "empty strings",
			v1:       "",
			v2:       "",
			expected: 0,
		},
		{
			name:     "v1 empty, v2 not",
			v1:       "",
			v2:       "v1.0",
			expected: -1, // "" < "1.0"
		},
		{
			name:     "v2 empty, v1 not",
			v1:       "v1.0",
			v2:       "",
			expected: 1, // "1.0" > ""
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := compareVersion(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPublishHistoryList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		agentID string
		setup   func(*gomock.Controller) (*releaseSvc, context.Context)
		wantLen int
		wantErr bool
	}{
		{
			name:    "successful get history list",
			agentID: "agent-123",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context) {
				ctx := context.Background()
				releaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

				poList := []*dapo.ReleaseHistoryPO{
					{
						ID:           "history-1",
						AgentID:      "agent-123",
						AgentVersion: "v1",
						AgentDesc:    "version 1",
						CreateTime:   1000,
					},
					{
						ID:           "history-2",
						AgentID:      "agent-123",
						AgentVersion: "v3",
						AgentDesc:    "version 3",
						CreateTime:   3000,
					},
					{
						ID:           "history-3",
						AgentID:      "agent-123",
						AgentVersion: "v2",
						AgentDesc:    "version 2",
						CreateTime:   2000,
					},
				}

				releaseHistoryRepo.EXPECT().
					ListByAgentID(ctx, "agent-123").
					Return(poList, int64(3), nil)

				svc := &releaseSvc{
					SvcBase:            service.NewSvcBase(),
					releaseHistoryRepo: releaseHistoryRepo,
				}

				return svc, ctx
			},
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "empty history list",
			agentID: "agent-no-history",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context) {
				ctx := context.Background()
				releaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

				releaseHistoryRepo.EXPECT().
					ListByAgentID(ctx, "agent-no-history").
					Return([]*dapo.ReleaseHistoryPO{}, int64(0), nil)

				svc := &releaseSvc{
					SvcBase:            service.NewSvcBase(),
					releaseHistoryRepo: releaseHistoryRepo,
				}

				return svc, ctx
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "repository error",
			agentID: "agent-error",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context) {
				ctx := context.Background()
				releaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

				releaseHistoryRepo.EXPECT().
					ListByAgentID(ctx, "agent-error").
					Return(nil, int64(0), assert.AnError)

				svc := &releaseSvc{
					SvcBase:            service.NewSvcBase(),
					releaseHistoryRepo: releaseHistoryRepo,
				}

				return svc, ctx
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx := tt.setup(ctrl)
			res, total, err := svc.GetPublishHistoryList(ctx, tt.agentID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantLen, len(res))
				assert.Equal(t, int64(tt.wantLen), total)

				// Verify sorting (should be in descending order by version)
				if len(res) > 1 {
					for i := 1; i < len(res); i++ {
						// Each subsequent version should be <= previous (descending)
						assert.GreaterOrEqual(t, compareVersion(res[i-1].AgentVersion, res[i].AgentVersion), 0)
					}
				}
			}
		})
	}
}

func TestGetPublishHistoryInfo(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &releaseSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	res, err := svc.GetPublishHistoryInfo(ctx, nil)

	require.NoError(t, err)
	assert.Equal(t, "", res)
}
