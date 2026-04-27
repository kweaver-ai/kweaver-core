package othersvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	otherreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/other/otherreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewOtherService(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dto := &NewOtherSvcDto{
		SvcBase:       service.NewSvcBase(),
		AgentConfRepo: idbaccessmock.NewMockIDataAgentConfigRepo(ctrl),
	}

	svc := NewOtherService(dto)

	assert.NotNil(t, svc)
}

func TestOtherSvc_DolphinTplList(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &otherSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentRepo,
	}

	config := &daconfvalobj.Config{}
	req := &otherreq.DolphinTplListReq{
		Config:          config,
		BuiltInAgentKey: "",
	}

	// Test with empty built-in agent key - should not call repo
	resp, err := svc.DolphinTplList(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestOtherSvc_DolphinTplList_WithAgent(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &otherSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentRepo,
	}

	builtInYes := cdaenum.BuiltInYes
	config := &daconfvalobj.Config{}
	req := &otherreq.DolphinTplListReq{
		Config:          config,
		BuiltInAgentKey: "test-agent-key",
	}

	// Mock successful agent retrieval
	agentPO := &dapo.DataAgentPo{
		ID:        "agent-123",
		Key:       "test-agent-key",
		IsBuiltIn: &builtInYes,
	}
	mockAgentRepo.EXPECT().GetByKey(ctx, "test-agent-key").Return(agentPO, nil)

	resp, err := svc.DolphinTplList(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestOtherSvc_DolphinTplList_AgentNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &otherSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentRepo,
	}

	config := &daconfvalobj.Config{}
	req := &otherreq.DolphinTplListReq{
		Config:          config,
		BuiltInAgentKey: "non-existent-agent",
	}

	// Mock agent not found with SQL ErrNoRows
	mockAgentRepo.EXPECT().GetByKey(ctx, "non-existent-agent").Return(nil, sql.ErrNoRows)

	resp, err := svc.DolphinTplList(ctx, req)

	assert.Error(t, err)
	// LoadFromConfig still runs, so resp won't be nil
	assert.NotNil(t, resp)
}

func TestOtherSvc_DolphinTplList_NotBuiltIn(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &otherSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentRepo,
	}

	builtInNo := cdaenum.BuiltInNo
	config := &daconfvalobj.Config{}
	req := &otherreq.DolphinTplListReq{
		Config:          config,
		BuiltInAgentKey: "custom-agent",
	}

	// Mock agent found but not built-in
	agentPO := &dapo.DataAgentPo{
		ID:        "agent-123",
		Key:       "custom-agent",
		IsBuiltIn: &builtInNo, // Not built-in
	}
	mockAgentRepo.EXPECT().GetByKey(ctx, "custom-agent").Return(agentPO, nil)

	resp, err := svc.DolphinTplList(ctx, req)

	assert.Error(t, err)
	// LoadFromConfig still runs, so resp won't be nil
	assert.NotNil(t, resp)
}

func TestOtherSvc_DolphinTplList_DBError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &otherSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentRepo,
	}

	config := &daconfvalobj.Config{}
	req := &otherreq.DolphinTplListReq{
		Config:          config,
		BuiltInAgentKey: "error-agent",
	}

	// Mock database error (not sql.ErrNotFound)
	dbErr := errors.New("database connection failed")
	mockAgentRepo.EXPECT().GetByKey(ctx, "error-agent").Return(nil, dbErr)

	resp, err := svc.DolphinTplList(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "[DolphinTplList]")
	assert.NotNil(t, resp)
}
