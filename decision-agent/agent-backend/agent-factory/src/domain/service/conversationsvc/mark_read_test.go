package conversationsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMarkRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		conversationID string
		lastestReadIdx int
		setup          func(*gomock.Controller) (*conversationSvc, context.Context)
		wantErr        bool
		errContains    string
	}{
		{
			name:           "marks conversation as read successfully",
			conversationID: "conv-123",
			lastestReadIdx: 5,
			setup: func(ctrl *gomock.Controller) (*conversationSvc, context.Context) {
				ctx := context.Background()
				repo := idbaccessmock.NewMockIConversationRepo(ctrl)

				repo.EXPECT().GetByID(gomock.Any(), "conv-123").Return(&dapo.ConversationPO{ID: "conv-123"}, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

				svc := &conversationSvc{
					SvcBase:          service.NewSvcBase(),
					conversationRepo: repo,
				}

				return svc, ctx
			},
			wantErr: false,
		},
		{
			name:           "returns error when conversation not found",
			conversationID: "conv-999",
			lastestReadIdx: 5,
			setup: func(ctrl *gomock.Controller) (*conversationSvc, context.Context) {
				ctx := context.Background()
				repo := idbaccessmock.NewMockIConversationRepo(ctrl)

				repo.EXPECT().GetByID(gomock.Any(), "conv-999").Return(nil, sql.ErrNoRows)

				svc := &conversationSvc{
					SvcBase:          service.NewSvcBase(),
					conversationRepo: repo,
				}

				return svc, ctx
			},
			wantErr:     true,
			errContains: "NotFound",
		},
		{
			name:           "returns error when get conversation fails",
			conversationID: "conv-123",
			lastestReadIdx: 5,
			setup: func(ctrl *gomock.Controller) (*conversationSvc, context.Context) {
				ctx := context.Background()
				repo := idbaccessmock.NewMockIConversationRepo(ctrl)

				repo.EXPECT().GetByID(gomock.Any(), "conv-123").Return(nil, errors.New("database error"))

				svc := &conversationSvc{
					SvcBase:          service.NewSvcBase(),
					conversationRepo: repo,
				}

				return svc, ctx
			},
			wantErr:     true,
			errContains: "database error",
		},
		{
			name:           "returns error when update fails",
			conversationID: "conv-123",
			lastestReadIdx: 5,
			setup: func(ctrl *gomock.Controller) (*conversationSvc, context.Context) {
				ctx := context.Background()
				repo := idbaccessmock.NewMockIConversationRepo(ctrl)

				repo.EXPECT().GetByID(gomock.Any(), "conv-123").Return(&dapo.ConversationPO{ID: "conv-123"}, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))

				svc := &conversationSvc{
					SvcBase:          service.NewSvcBase(),
					conversationRepo: repo,
				}

				return svc, ctx
			},
			wantErr:     true,
			errContains: "update failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx := tt.setup(ctrl)
			err := svc.MarkRead(ctx, tt.conversationID, tt.lastestReadIdx)

			if tt.wantErr {
				assert.Error(t, err)

				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
