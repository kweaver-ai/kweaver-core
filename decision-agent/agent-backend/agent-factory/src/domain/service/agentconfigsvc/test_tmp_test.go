package v3agentconfigsvc

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUpdateStatusTest_CopyJSONError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: repo,
	}

	// params 为一个无法正确复制的 chan（会导致 JSON 序列化失败）
	req := &agentconfigreq.TestTmpReq{
		TestFlag: "update_status",
		Params:   make(chan int), // chan type → json.Marshal fails → CopyUseJSON error
	}

	err := svc.TmpTest(context.Background(), req)
	assert.Error(t, err)
}

func TestTmpTest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFlag string
		params   interface{}
		setup    func(*gomock.Controller) (*dataAgentConfigSvc, context.Context)
		wantErr  bool
	}{
		{
			name:     "unknown test flag returns nil",
			testFlag: "unknown",
			params:   nil,
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				repo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

				svc := &dataAgentConfigSvc{
					SvcBase:       service.NewSvcBase(),
					agentConfRepo: repo,
				}

				return svc, ctx
			},
			wantErr: false,
		},
		{
			name:     "update_status test flag calls UpdateStatus",
			testFlag: "update_status",
			params: map[string]interface{}{
				"id":     "agent-123",
				"status": "published",
			},
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				repo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

				// Expect UpdateStatus to be called with the correct parameters
				repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), cdaenum.StatusPublished, "agent-123", "").Return(nil)

				svc := &dataAgentConfigSvc{
					SvcBase:       service.NewSvcBase(),
					agentConfRepo: repo,
				}

				return svc, ctx
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx := tt.setup(ctrl)

			req := &agentconfigreq.TestTmpReq{
				TestFlag: tt.testFlag,
				Params:   tt.params,
			}

			err := svc.TmpTest(ctx, req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
