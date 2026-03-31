package personalspacesvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/cconf"
	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/personalspacedbacc/psdbarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/personal_space/personalspacereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Helper function to create context with user ID and business domain ID
func createPersonalSpaceCtx(userID, bdID string) context.Context {
	visitor := &rest.Visitor{
		ID: userID,
	}
	ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck
	ctx = context.WithValue(ctx, cenum.BizDomainIDCtxKey.String(), bdID)                        //nolint:staticcheck

	return ctx
}

func setDisableBizDomain(t *testing.T, disable bool) {
	t.Helper()

	oldCfg := global.GConfig
	global.GConfig = &conf.Config{
		Config:       cconf.BaseDefConfig(),
		SwitchFields: conf.NewSwitchFields(),
	}
	global.GConfig.SwitchFields.DisableBizDomain = disable

	t.Cleanup(func() {
		global.GConfig = oldCfg
	})
}

func TestAgentTplList_NilRequest(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &PersonalSpaceService{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	resp, err := svc.AgentTplList(ctx, nil)

	assert.Error(t, err)
	assert.NotNil(t, resp) // Response is created before error check
	assert.Contains(t, err.Error(), "请求参数不能为空")
}

func TestAgentTplList_NoUserInContext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)

	req := &personalspacereq.AgentTplListReq{
		Size: 10,
	}

	svc := &PersonalSpaceService{
		SvcBase:       service.NewSvcBase(),
		bizDomainHttp: mockBizDomainHttp,
	}

	ctx := context.Background() // No user in context
	resp, err := svc.AgentTplList(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "用户未登录")
}

func TestAgentTplList_BizDomainHttpError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)

	req := &personalspacereq.AgentTplListReq{
		Size: 10,
	}

	expectedErr := errors.New("business domain error")
	mockBizDomainHttp.EXPECT().GetAllAgentTplIDList(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

	svc := &PersonalSpaceService{
		SvcBase:       service.NewSvcBase(),
		bizDomainHttp: mockBizDomainHttp,
	}

	ctx := createPersonalSpaceCtx("user-123", "bd-123")
	resp, err := svc.AgentTplList(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "get all agent tpl id list failed")
}

func TestAgentTplList_NoTemplatesInBusinessDomain(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)

	req := &personalspacereq.AgentTplListReq{
		Size: 10,
	}

	mockBizDomainHttp.EXPECT().GetAllAgentTplIDList(gomock.Any(), []string{"bd-123"}).Return([]string{}, nil)

	svc := &PersonalSpaceService{
		SvcBase:       service.NewSvcBase(),
		bizDomainHttp: mockBizDomainHttp,
	}

	ctx := createPersonalSpaceCtx("user-123", "bd-123")
	resp, err := svc.AgentTplList(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestAgentList_NilRequest(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &PersonalSpaceService{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	resp, err := svc.AgentList(ctx, nil)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "请求参数不能为空")
}

func TestAgentList_NoUserInContext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)

	req := &personalspacereq.AgentListReq{
		Size: 10,
	}

	svc := &PersonalSpaceService{
		SvcBase:       service.NewSvcBase(),
		bizDomainHttp: mockBizDomainHttp,
	}

	ctx := context.Background() // No user in context
	resp, err := svc.AgentList(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "用户未登录")
}

func TestAgentList_NoAgentsInBusinessDomain(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	req := &personalspacereq.AgentListReq{
		Size: 10,
	}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, gomock.Any()).Return(true, nil)
	mockBizDomainHttp.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd-123"}).Return([]string{}, map[string]string{}, nil)

	svc := &PersonalSpaceService{
		SvcBase:       service.NewSvcBase(),
		bizDomainHttp: mockBizDomainHttp,
		pmsSvc:        mockPmsSvc,
	}

	ctx := createPersonalSpaceCtx("user-123", "bd-123")
	resp, err := svc.AgentList(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestAgentList_PermissionError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	req := &personalspacereq.AgentListReq{
		Size: 10,
	}

	expectedErr := errors.New("permission error")
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, gomock.Any()).Return(false, expectedErr)

	svc := &PersonalSpaceService{
		SvcBase:       service.NewSvcBase(),
		bizDomainHttp: mockBizDomainHttp,
		pmsSvc:        mockPmsSvc,
	}

	ctx := createPersonalSpaceCtx("user-123", "bd-123")
	resp, err := svc.AgentList(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "get agent list from repo failed")
}

func TestAgentList_BizDomainHttpError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	req := &personalspacereq.AgentListReq{
		Size: 10,
	}

	expectedErr := errors.New("business domain error")

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, gomock.Any()).Return(true, nil)
	mockBizDomainHttp.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd-123"}).Return(nil, nil, expectedErr)

	svc := &PersonalSpaceService{
		SvcBase:       service.NewSvcBase(),
		bizDomainHttp: mockBizDomainHttp,
		pmsSvc:        mockPmsSvc,
	}

	ctx := createPersonalSpaceCtx("user-123", "bd-123")
	resp, err := svc.AgentList(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "get agent list from repo failed")
}

func TestAgentList_DisableBizDomainSkipsFilter(t *testing.T) {
	setDisableBizDomain(t, true)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockRepo := idbaccessmock.NewMockIPersonalSpaceRepo(ctrl)

	req := &personalspacereq.AgentListReq{
		Size: 10,
	}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, gomock.Any()).Return(true, nil)
	mockRepo.EXPECT().ListPersonalSpaceAgent(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, arg *psdbarg.AgentListArg) ([]*dapo.DataAgentPo, error) {
			assert.Nil(t, arg.AgentIDsByBizDomain)
			return []*dapo.DataAgentPo{}, nil
		},
	)

	svc := &PersonalSpaceService{
		SvcBase:           service.NewSvcBase(),
		pmsSvc:            mockPmsSvc,
		personalSpaceRepo: mockRepo,
	}

	ctx := createPersonalSpaceCtx("user-123", "bd-123")
	resp, err := svc.AgentList(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestAgentTplList_DisableBizDomainSkipsFilter(t *testing.T) {
	setDisableBizDomain(t, true)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIPersonalSpaceRepo(ctrl)

	req := &personalspacereq.AgentTplListReq{
		Size: 10,
	}

	mockRepo.EXPECT().ListPersonalSpaceTpl(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, arg *psdbarg.TplListArg) ([]*dapo.DataAgentTplPo, error) {
			assert.Nil(t, arg.TplIDsByBd)
			return []*dapo.DataAgentTplPo{}, nil
		},
	)

	svc := &PersonalSpaceService{
		SvcBase:           service.NewSvcBase(),
		personalSpaceRepo: mockRepo,
	}

	ctx := createPersonalSpaceCtx("user-123", "bd-123")
	resp, err := svc.AgentTplList(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
