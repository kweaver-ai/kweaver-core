package tplsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDataAgentTplSvc_Publish_PanicsWithoutAgentTplRepo(t *testing.T) {
	svc := &dataAgentTplSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	id := int64(123)
	req := &agenttplreq.PublishReq{}

	assert.Panics(t, func() {
		_, _, _ = svc.Publish(ctx, nil, req, id, false)
	})
}

func TestDataAgentTplSvc_Publish_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
		pmsSvc:       mockPmsSvc,
	}

	ctx := context.Background()
	id := int64(123)
	req := &agenttplreq.PublishReq{}

	// Permission denied
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	resp, auditLogInfo, err := svc.Publish(ctx, nil, req, id, false)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "do not have publish permission")
}

func TestDataAgentTplSvc_Publish_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
		pmsSvc:       mockPmsSvc,
	}

	ctx := context.Background()
	id := int64(123)
	req := &agenttplreq.PublishReq{}

	// Permission granted but BeginTx fails
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

	txErr := errors.New("transaction begin failed")
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, txErr)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	resp, auditLogInfo, err := svc.Publish(ctx, nil, req, id, false)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "begin transaction")
}
