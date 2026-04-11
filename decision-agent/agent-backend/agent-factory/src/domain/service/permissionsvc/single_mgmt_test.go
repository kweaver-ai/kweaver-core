package permissionsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func init() {
	// Initialize global config for tests
	if global.GConfig == nil {
		global.GConfig = &conf.Config{
			SwitchFields: conf.NewSwitchFields(),
		}
	}
}

func TestPermissionSvc_GetSingleMgmtPermission_EmptyUserID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZHttp,
		umHttp:    mockUmHttp,
	}

	ctx := context.Background()

	allAllowed, err := svc.GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentPublish)

	assert.Error(t, err)
	assert.False(t, allAllowed)
	assert.Contains(t, err.Error(), "user id is empty")
}

func TestPermissionSvc_GetSingleMgmtPermission_InvalidResourceType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZHttp,
		umHttp:    mockUmHttp,
	}

	ctx := createContextWithUserID("user-123")

	allAllowed, err := svc.GetSingleMgmtPermission(ctx, "invalid_resource_type", cdapmsenum.AgentPublish)

	assert.Error(t, err)
	assert.False(t, allAllowed)
	assert.Contains(t, err.Error(), "invalid resource type")
}

func TestPermissionSvc_GetSingleMgmtPermission_AgentResourceOpsError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZHttp,
		umHttp:    mockUmHttp,
	}

	ctx := createContextWithUserID("user-123")
	httpErr := errors.New("http request failed")

	mockAuthZHttp.EXPECT().GetAgentResourceOpsByUid(gomock.Any(), "user-123").Return(nil, httpErr)

	allAllowed, err := svc.GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentPublish)

	assert.Error(t, err)
	assert.False(t, allAllowed)
}

func TestPermissionSvc_GetSingleMgmtPermission_AgentTplResourceOpsError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZHttp,
		umHttp:    mockUmHttp,
	}

	ctx := createContextWithUserID("user-123")
	httpErr := errors.New("http request failed")

	mockAuthZHttp.EXPECT().GetAgentTplResourceOpsByUid(gomock.Any(), "user-123").Return(nil, httpErr)

	allAllowed, err := svc.GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplPublish)

	assert.Error(t, err)
	assert.False(t, allAllowed)
}

func TestPermissionSvc_GetSingleMgmtPermission_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZHttp,
		umHttp:    mockUmHttp,
	}

	ctx := createContextWithUserID("user-123")
	agentOps := map[cdapmsenum.Operator]bool{
		cdapmsenum.AgentPublish: true,
		cdapmsenum.AgentUse:     false,
	}

	mockAuthZHttp.EXPECT().GetAgentResourceOpsByUid(gomock.Any(), "user-123").Return(agentOps, nil)

	allAllowed, err := svc.GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentPublish)

	assert.NoError(t, err)
	assert.True(t, allAllowed)
}

func TestPermissionSvc_GetSingleMgmtPermission_NotAllowed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZHttp,
		umHttp:    mockUmHttp,
	}

	ctx := createContextWithUserID("user-123")
	agentOps := map[cdapmsenum.Operator]bool{
		cdapmsenum.AgentPublish: false,
	}

	mockAuthZHttp.EXPECT().GetAgentResourceOpsByUid(gomock.Any(), "user-123").Return(agentOps, nil)

	allAllowed, err := svc.GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentPublish)

	assert.NoError(t, err)
	assert.False(t, allAllowed)
}
