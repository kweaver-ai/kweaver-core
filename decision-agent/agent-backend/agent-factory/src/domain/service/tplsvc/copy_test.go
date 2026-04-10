package tplsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDataAgentTplSvc_Copy_SourceNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockBdAgentTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:           service.NewSvcBase(),
		agentTplRepo:      mockAgentTplRepo,
		bizDomainHttp:     mockBizDomainHttp,
		bdAgentTplRelRepo: mockBdAgentTplRelRepo,
	}

	ctx := context.Background()
	templateID := int64(123)

	notFoundErr := errors.New("record not found")
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(nil, notFoundErr)

	resp, auditLogInfo, err := svc.Copy(ctx, templateID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, auditLogInfo.ID)
	assert.Empty(t, auditLogInfo.Name)
}

func TestDataAgentTplSvc_Copy_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockBdAgentTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:           service.NewSvcBase(),
		agentTplRepo:      mockAgentTplRepo,
		bizDomainHttp:     mockBizDomainHttp,
		bdAgentTplRelRepo: mockBdAgentTplRelRepo,
	}

	ctx := context.Background()
	templateID := int64(123)

	dbErr := errors.New("database connection failed")
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(nil, dbErr)

	resp, auditLogInfo, err := svc.Copy(ctx, templateID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, resp)
	assert.Empty(t, auditLogInfo.ID)
	assert.Empty(t, auditLogInfo.Name)
}

func TestDataAgentTplSvc_Copy_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockBdAgentTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:           service.NewSvcBase(),
		agentTplRepo:      mockAgentTplRepo,
		bizDomainHttp:     mockBizDomainHttp,
		bdAgentTplRelRepo: mockBdAgentTplRelRepo,
	}

	ctx := context.Background()
	templateID := int64(123)

	sourcePo := &dapo.DataAgentTplPo{
		ID:   123,
		Name: "Test Template",
		Key:  "test-tpl-key",
	}

	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(sourcePo, nil)

	txErr := errors.New("transaction begin failed")
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, txErr)

	resp, auditLogInfo, err := svc.Copy(ctx, templateID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction")
	assert.Nil(t, resp)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.NotEmpty(t, auditLogInfo.Name)
}

func TestDataAgentTplSvc_Copy_PanicsWithoutAgentTplRepo(t *testing.T) {
	svc := &dataAgentTplSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	templateID := int64(123)

	assert.Panics(t, func() {
		_, _, _ = svc.Copy(ctx, templateID)
	})
}
