package conversationsvc

import (
	"context"
	"database/sql"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     *conversationreq.UpdateReq
		setup   func(*gomock.Controller) (*conversationSvc, context.Context)
		wantErr bool
	}{
		{
			name: "updates conversation title successfully",
			req: &conversationreq.UpdateReq{
				ID:    "conv-123",
				Title: "New Title",
			},
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
			name: "truncates title to 50 characters",
			req: &conversationreq.UpdateReq{
				ID:    "conv-123",
				Title: "This is a very long title that exceeds fifty characters limit",
			},
			setup: func(ctrl *gomock.Controller) (*conversationSvc, context.Context) {
				ctx := context.Background()
				repo := idbaccessmock.NewMockIConversationRepo(ctrl)

				repo.EXPECT().GetByID(gomock.Any(), "conv-123").Return(&dapo.ConversationPO{ID: "conv-123"}, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Do(func(_ context.Context, po *dapo.ConversationPO) {
					assert.Equal(t, 50, len([]rune(po.Title)))
				}).Return(nil)

				svc := &conversationSvc{
					SvcBase:          service.NewSvcBase(),
					conversationRepo: repo,
				}

				return svc, ctx
			},
			wantErr: false,
		},
		{
			name: "handles empty title",
			req: &conversationreq.UpdateReq{
				ID:    "conv-123",
				Title: "",
			},
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
			name: "returns error when conversation not found",
			req: &conversationreq.UpdateReq{
				ID:    "conv-999",
				Title: "New Title",
			},
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
			wantErr: true,
		},
		{
			name: "handles title with exactly 50 characters",
			req: &conversationreq.UpdateReq{
				ID:    "conv-123",
				Title: "12345678901234567890123456789012345678901234567890",
			},
			setup: func(ctrl *gomock.Controller) (*conversationSvc, context.Context) {
				ctx := context.Background()
				repo := idbaccessmock.NewMockIConversationRepo(ctrl)

				repo.EXPECT().GetByID(gomock.Any(), "conv-123").Return(&dapo.ConversationPO{ID: "conv-123"}, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Do(func(_ context.Context, po *dapo.ConversationPO) {
					assert.Equal(t, 50, len([]rune(po.Title)))
					assert.Equal(t, "12345678901234567890123456789012345678901234567890", po.Title)
				}).Return(nil)

				svc := &conversationSvc{
					SvcBase:          service.NewSvcBase(),
					conversationRepo: repo,
				}

				return svc, ctx
			},
			wantErr: false,
		},
		{
			name: "handles title with unicode characters",
			req: &conversationreq.UpdateReq{
				ID:    "conv-123",
				Title: "标题测试这是一个中文标题",
			},
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
			name: "returns error when update fails",
			req: &conversationreq.UpdateReq{
				ID:    "conv-123",
				Title: "New Title",
			},
			setup: func(ctrl *gomock.Controller) (*conversationSvc, context.Context) {
				ctx := context.Background()
				repo := idbaccessmock.NewMockIConversationRepo(ctrl)

				repo.EXPECT().GetByID(gomock.Any(), "conv-123").Return(&dapo.ConversationPO{ID: "conv-123"}, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(assert.AnError)

				svc := &conversationSvc{
					SvcBase:          service.NewSvcBase(),
					conversationRepo: repo,
				}

				return svc, ctx
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx := tt.setup(ctrl)
			err := svc.Update(ctx, *tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
