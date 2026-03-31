package bizdomainsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBizDomainSvc_InitBizDomainAgentRel_BeginTxError_Full(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewBizDomainService(&NewBizDomainSvcDto{
		SvcBase: service.NewSvcBase(),
	})

	ctx := context.Background()

	txErr := errors.New("transaction begin failed")

	mockBdAgentRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
	mockBdAgentRelRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, txErr)

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	err := svc.InitBizDomainAgentRel(ctx, mockAgentRepo, mockBdAgentRelRepo)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin tx failed")
}
