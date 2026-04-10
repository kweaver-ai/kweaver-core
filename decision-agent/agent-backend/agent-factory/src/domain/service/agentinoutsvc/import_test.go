package agentinoutsvc

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCheckSystemAgentCreatePermission_NoSystemAgents(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	// Non-system agent
	isSystemAgent := cenum.YesNoInt8No
	exportData := &agentinoutresp.ExportResp{
		Agents: []*agentinoutresp.ExportAgentItem{
			{
				DataAgentPo: &dapo.DataAgentPo{
					Key:           "agent-1",
					Name:          "Regular Agent",
					IsSystemAgent: &isSystemAgent,
				},
			},
		},
	}
	resp := agentinoutresp.NewImportResp()

	svc := &agentInOutSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	ctx := context.Background()
	err := svc.checkSystemAgentCreatePermission(ctx, exportData, resp)

	assert.NoError(t, err)
	assert.Empty(t, resp.NoCreateSystemAgentPms)
}

func TestCheckSystemAgentCreatePermission_HasPermission(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	// System agent
	isSystemAgent := cenum.YesNoInt8Yes
	exportData := &agentinoutresp.ExportResp{
		Agents: []*agentinoutresp.ExportAgentItem{
			{
				DataAgentPo: &dapo.DataAgentPo{
					Key:           "sys-agent-1",
					Name:          "System Agent",
					IsSystemAgent: &isSystemAgent,
				},
			},
		},
	}
	resp := agentinoutresp.NewImportResp()

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentCreateSystemAgent).Return(true, nil)

	svc := &agentInOutSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	ctx := context.Background()
	err := svc.checkSystemAgentCreatePermission(ctx, exportData, resp)

	assert.NoError(t, err)
	assert.Empty(t, resp.NoCreateSystemAgentPms)
}

func TestCheckSystemAgentCreatePermission_NoPermission(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	// System agent
	isSystemAgent := cenum.YesNoInt8Yes
	exportData := &agentinoutresp.ExportResp{
		Agents: []*agentinoutresp.ExportAgentItem{
			{
				DataAgentPo: &dapo.DataAgentPo{
					Key:           "sys-agent-1",
					Name:          "System Agent",
					IsSystemAgent: &isSystemAgent,
				},
			},
		},
	}
	resp := agentinoutresp.NewImportResp()

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentCreateSystemAgent).Return(false, nil)

	svc := &agentInOutSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	ctx := context.Background()
	err := svc.checkSystemAgentCreatePermission(ctx, exportData, resp)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.NoCreateSystemAgentPms)
	assert.Len(t, resp.NoCreateSystemAgentPms, 1)
}

func TestCheckSystemAgentCreatePermission_PermissionError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	// System agent
	isSystemAgent := cenum.YesNoInt8Yes
	exportData := &agentinoutresp.ExportResp{
		Agents: []*agentinoutresp.ExportAgentItem{
			{
				DataAgentPo: &dapo.DataAgentPo{
					Key:           "sys-agent-1",
					Name:          "System Agent",
					IsSystemAgent: &isSystemAgent,
				},
			},
		},
	}
	resp := agentinoutresp.NewImportResp()

	expectedErr := errors.New("permission error")
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentCreateSystemAgent).Return(false, expectedErr)

	svc := &agentInOutSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	ctx := context.Background()
	err := svc.checkSystemAgentCreatePermission(ctx, exportData, resp)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check system agent create permission failed")
}

func TestCheckSystemAgentCreatePermission_MultipleSystemAgents(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	// Multiple system agents
	isSystemAgent := cenum.YesNoInt8Yes
	exportData := &agentinoutresp.ExportResp{
		Agents: []*agentinoutresp.ExportAgentItem{
			{
				DataAgentPo: &dapo.DataAgentPo{
					Key:           "sys-agent-1",
					Name:          "System Agent 1",
					IsSystemAgent: &isSystemAgent,
				},
			},
			{
				DataAgentPo: &dapo.DataAgentPo{
					Key:           "sys-agent-2",
					Name:          "System Agent 2",
					IsSystemAgent: &isSystemAgent,
				},
			},
		},
	}
	resp := agentinoutresp.NewImportResp()

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentCreateSystemAgent).Return(false, nil)

	svc := &agentInOutSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	ctx := context.Background()
	err := svc.checkSystemAgentCreatePermission(ctx, exportData, resp)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.NoCreateSystemAgentPms)
	assert.Len(t, resp.NoCreateSystemAgentPms, 2)
}

func TestCheckAgentConfigValid_EmptyAgents(t *testing.T) {
	t.Parallel()

	svc := &agentInOutSvc{
		SvcBase: service.NewSvcBase(),
	}

	exportData := &agentinoutresp.ExportResp{
		Agents: []*agentinoutresp.ExportAgentItem{},
	}
	resp := agentinoutresp.NewImportResp()

	ctx := context.Background()
	svc.checkAgentConfigValid(ctx, exportData, resp)

	assert.False(t, resp.HasFail())
	assert.Empty(t, resp.ConfigInvalid)
}

func TestImport_MissingUserID(t *testing.T) {
	t.Parallel()

	svc := &agentInOutSvc{
		SvcBase: service.NewSvcBase(),
	}

	req := agentinoutreq.NewImportReq()
	req.ImportType = agentinoutreq.ImportTypeCreate

	ctx := context.Background()
	resp, err := svc.Import(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无法获取用户ID")
	assert.NotNil(t, resp)
	assert.False(t, resp.IsSuccess)
}

func TestImport_FileTooLarge(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	svc := &agentInOutSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	req := agentinoutreq.NewImportReq()
	req.ImportType = agentinoutreq.ImportTypeCreate
	req.File = &multipart.FileHeader{
		Size: 11 * 1024 * 1024, // 11MB, exceeds MaxFileSize
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
	}

	ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ //nolint:staticcheck // SA1029
		ID: "user-123",
	})
	resp, err := svc.Import(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件大小不能超过10MB")
	assert.NotNil(t, resp)
	assert.False(t, resp.IsSuccess)
}

func TestImport_InvalidContentType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	svc := &agentInOutSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	req := agentinoutreq.NewImportReq()
	req.ImportType = agentinoutreq.ImportTypeCreate
	req.File = &multipart.FileHeader{
		Size: 1024,
		Header: map[string][]string{
			"Content-Type": {"text/plain"},
		},
	}

	ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ //nolint:staticcheck // SA1029
		ID: "user-123",
	})
	resp, err := svc.Import(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "只支持JSON格式文件")
	assert.NotNil(t, resp)
	assert.False(t, resp.IsSuccess)
}
