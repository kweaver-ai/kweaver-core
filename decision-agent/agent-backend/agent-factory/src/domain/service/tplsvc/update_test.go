package tplsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDataAgentTplSvc_Update_PanicsWithoutAgentTplRepo(t *testing.T) {
	t.Parallel()

	svc := &dataAgentTplSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	id := int64(123)
	req := &agenttplreq.UpdateReq{}

	assert.Panics(t, func() {
		_, _ = svc.Update(ctx, req, id)
	})
}

func TestDataAgentTplSvc_Update_TemplateNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
	}

	ctx := context.Background()
	id := int64(999)
	req := &agenttplreq.UpdateReq{}

	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, sql.ErrNoRows)

	auditLogInfo, err := svc.Update(ctx, req, id)

	assert.Error(t, err)
	assert.Empty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "模板不存在")
}

func TestDataAgentTplSvc_Update_GetByIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
	}

	ctx := context.Background()
	id := int64(123)
	req := &agenttplreq.UpdateReq{}

	dbErr := errors.New("database connection failed")
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, dbErr)

	auditLogInfo, err := svc.Update(ctx, req, id)

	assert.Error(t, err)
	assert.Empty(t, auditLogInfo.ID)
}

func TestDataAgentTplSvc_Update_NameConflict(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
	}

	ctx := context.Background()
	id := int64(123)
	userID := "user-123"

	// Set up context with user ID
	visitor := &rest.Visitor{
		ID: userID,
	}
	ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

	req := &agenttplreq.UpdateReq{}

	oldPo := &dapo.DataAgentTplPo{
		ID:        123,
		Name:      "Old Name",
		CreatedBy: userID,
	}

	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), id).Return(oldPo, nil)
	mockAgentTplRepo.EXPECT().ExistsByNameExcludeID(gomock.Any(), gomock.Any(), id).Return(true, nil)

	auditLogInfo, err := svc.Update(ctx, req, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "模板名称已存在")
	assert.NotEmpty(t, auditLogInfo.ID)
}
