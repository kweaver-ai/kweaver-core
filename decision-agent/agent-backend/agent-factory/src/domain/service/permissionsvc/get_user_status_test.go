package permissionsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
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

// Helper function to create context with user ID
func createContextWithUserID(userID string) context.Context {
	visitor := &rest.Visitor{
		ID: userID,
	}

	return context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029
}

func TestPermissionSvc_GetUserStatus_EmptyUserID(t *testing.T) {
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

	// Create context without user ID
	ctx := context.Background()

	resp, err := svc.GetUserStatus(ctx)

	// Response is initialized before error check, so it's not nil
	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "user id is empty")
}

func TestPermissionSvc_GetUserStatus_GetAgentResourceOpsError(t *testing.T) {
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

	resp, err := svc.GetUserStatus(ctx)

	// Response is initialized before error check, so it's not nil
	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "get agent resource ops by uid failed")
}

func TestPermissionSvc_GetUserStatus_GetAgentTplResourceOpsError(t *testing.T) {
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
	}
	httpErr := errors.New("http request failed")

	mockAuthZHttp.EXPECT().GetAgentResourceOpsByUid(gomock.Any(), "user-123").Return(agentOps, nil)
	mockAuthZHttp.EXPECT().GetAgentTplResourceOpsByUid(gomock.Any(), "user-123").Return(nil, httpErr)

	resp, err := svc.GetUserStatus(ctx)

	// Response is initialized before error check, so it's not nil
	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "get agent tpl resource ops by uid failed")
}

func TestPermissionSvc_GetUserStatus_Success(t *testing.T) {
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
		cdapmsenum.AgentPublish:                  true,
		cdapmsenum.AgentUnpublish:                false,
		cdapmsenum.AgentUnpublishOtherUserAgent:  false,
		cdapmsenum.AgentPublishToBeSkillAgent:    true,
		cdapmsenum.AgentPublishToBeWebSdkAgent:   false,
		cdapmsenum.AgentPublishToBeApiAgent:      false,
		cdapmsenum.AgentPublishToBeDataFlowAgent: false,
		cdapmsenum.AgentCreateSystemAgent:        false,
		cdapmsenum.AgentBuiltInAgentMgmt:         false,
		cdapmsenum.AgentSeeTrajectoryAnalysis:    true,
	}
	agentTplOps := map[cdapmsenum.Operator]bool{
		cdapmsenum.AgentTplPublish:                    true,
		cdapmsenum.AgentTplUnpublish:                  false,
		cdapmsenum.AgentTplUnpublishOtherUserAgentTpl: false,
	}

	mockAuthZHttp.EXPECT().GetAgentResourceOpsByUid(gomock.Any(), "user-123").Return(agentOps, nil)
	mockAuthZHttp.EXPECT().GetAgentTplResourceOpsByUid(gomock.Any(), "user-123").Return(agentTplOps, nil)

	resp, err := svc.GetUserStatus(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, true, resp.Agent.Publish)
	assert.Equal(t, false, resp.Agent.Unpublish)
	assert.Equal(t, true, resp.AgentTpl.Publish)
	assert.Equal(t, false, resp.AgentTpl.Unpublish)
}
