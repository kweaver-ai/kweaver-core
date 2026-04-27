package personalspacesvc

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbret"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/personal_space/personalspacereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
)

func setPersonalSpaceTestConfig(t *testing.T) {
	t.Helper()

	oldCfg := cglobal.GConfig
	cglobal.GConfig = cconf.BaseDefConfig()

	t.Cleanup(func() {
		cglobal.GConfig = oldCfg
	})
}

func TestPersonalSpaceService_AgentList_MoreBranches(t *testing.T) {
	t.Parallel()

	setPersonalSpaceTestConfig(t)

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		mockSpaceRepo := idbaccessmock.NewMockIPersonalSpaceRepo(ctrl)
		svc := &PersonalSpaceService{
			SvcBase:           service.NewSvcBase(),
			bizDomainHttp:     mockBiz,
			pmsSvc:            mockPms,
			personalSpaceRepo: mockSpaceRepo,
		}

		req := &personalspacereq.AgentListReq{Size: 1}
		ctx := createPersonalSpaceCtx("u1", "bd1")

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, gomock.Any()).Return(true, nil)
		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd1"}).Return([]string{"a1"}, map[string]string{"a1": "bd1"}, nil)
		mockSpaceRepo.EXPECT().ListPersonalSpaceAgent(gomock.Any(), gomock.Any()).Return(nil, errors.New("db failed"))

		resp, err := svc.AgentList(ctx, req)
		assert.Error(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("empty agent list from repo", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		mockSpaceRepo := idbaccessmock.NewMockIPersonalSpaceRepo(ctrl)
		svc := &PersonalSpaceService{
			SvcBase:           service.NewSvcBase(),
			bizDomainHttp:     mockBiz,
			pmsSvc:            mockPms,
			personalSpaceRepo: mockSpaceRepo,
		}

		req := &personalspacereq.AgentListReq{Size: 1}
		ctx := createPersonalSpaceCtx("u1", "bd1")

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, gomock.Any()).Return(true, nil)
		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd1"}).Return([]string{"a1"}, map[string]string{"a1": "bd1"}, nil)
		mockSpaceRepo.EXPECT().ListPersonalSpaceAgent(gomock.Any(), gomock.Any()).Return([]*dapo.DataAgentPo{}, nil)

		resp, err := svc.AgentList(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.IsLastPage)
	})

	t.Run("published map repo error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		mockSpaceRepo := idbaccessmock.NewMockIPersonalSpaceRepo(ctrl)
		mockPubedRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &PersonalSpaceService{
			SvcBase:           service.NewSvcBase(),
			bizDomainHttp:     mockBiz,
			pmsSvc:            mockPms,
			personalSpaceRepo: mockSpaceRepo,
			pubedAgentRepo:    mockPubedRepo,
			umHttp:            mockUm,
		}

		req := &personalspacereq.AgentListReq{Size: 1}
		ctx := createPersonalSpaceCtx("u1", "bd1")
		profile := "p1"
		pos := []*dapo.DataAgentPo{{ID: "a1", Key: "k1", Name: "n1", Profile: &profile, CreatedBy: "u1", UpdatedBy: "u1"}}

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, gomock.Any()).Return(true, nil)
		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd1"}).Return([]string{"a1"}, map[string]string{"a1": "bd1"}, nil)
		mockSpaceRepo.EXPECT().ListPersonalSpaceAgent(gomock.Any(), gomock.Any()).Return(pos, nil)
		mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).Return(umtypes.NewOsnInfoMapS(), nil)
		mockPubedRepo.EXPECT().GetPubedPoMapByXx(gomock.Any(), gomock.Any()).Return(nil, errors.New("query failed"))

		resp, err := svc.AgentList(ctx, req)
		assert.Error(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("success with next page marker", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		mockSpaceRepo := idbaccessmock.NewMockIPersonalSpaceRepo(ctrl)
		mockPubedRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &PersonalSpaceService{
			SvcBase:           service.NewSvcBase(),
			bizDomainHttp:     mockBiz,
			pmsSvc:            mockPms,
			personalSpaceRepo: mockSpaceRepo,
			pubedAgentRepo:    mockPubedRepo,
			umHttp:            mockUm,
		}

		req := &personalspacereq.AgentListReq{Size: 1}
		ctx := createPersonalSpaceCtx("u1", "bd1")
		profile := "p1"
		pos := []*dapo.DataAgentPo{
			{ID: "a1", Key: "k1", Name: "n1", Profile: &profile, CreatedBy: "u1", UpdatedBy: "u1", UpdatedAt: 1},
			{ID: "a2", Key: "k2", Name: "n2", Profile: &profile, CreatedBy: "u1", UpdatedBy: "u1", UpdatedAt: 2},
		}

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, gomock.Any()).Return(false, nil)
		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd1"}).Return([]string{"a1", "a2"}, map[string]string{"a1": "bd1", "a2": "bd1"}, nil)
		mockSpaceRepo.EXPECT().ListPersonalSpaceAgent(gomock.Any(), gomock.Any()).Return(pos, nil)
		mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).Return(umtypes.NewOsnInfoMapS(), nil)
		mockPubedRepo.EXPECT().GetPubedPoMapByXx(gomock.Any(), gomock.Any()).Return(padbret.NewGetPaPoMapByXxRet(), nil)

		resp, err := svc.AgentList(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.IsLastPage)
		assert.Len(t, resp.Entries, 1)
		assert.NotEmpty(t, resp.PaginationMarkerStr)
	})
}

func TestPersonalSpaceService_AgentTplList_MoreBranches(t *testing.T) {
	t.Parallel()

	setPersonalSpaceTestConfig(t)

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockSpaceRepo := idbaccessmock.NewMockIPersonalSpaceRepo(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &PersonalSpaceService{
			SvcBase:           service.NewSvcBase(),
			bizDomainHttp:     mockBiz,
			personalSpaceRepo: mockSpaceRepo,
			umHttp:            mockUm,
		}

		req := &personalspacereq.AgentTplListReq{Size: 1}
		ctx := createPersonalSpaceCtx("u1", "bd1")

		mockBiz.EXPECT().GetAllAgentTplIDList(gomock.Any(), []string{"bd1"}).Return([]string{"1"}, nil)
		mockSpaceRepo.EXPECT().ListPersonalSpaceTpl(gomock.Any(), gomock.Any()).Return(nil, errors.New("db failed"))

		resp, err := svc.AgentTplList(ctx, req)
		assert.Error(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("empty tpl list from repo", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockSpaceRepo := idbaccessmock.NewMockIPersonalSpaceRepo(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &PersonalSpaceService{
			SvcBase:           service.NewSvcBase(),
			bizDomainHttp:     mockBiz,
			personalSpaceRepo: mockSpaceRepo,
			umHttp:            mockUm,
		}

		req := &personalspacereq.AgentTplListReq{Size: 1}
		ctx := createPersonalSpaceCtx("u1", "bd1")

		mockBiz.EXPECT().GetAllAgentTplIDList(gomock.Any(), []string{"bd1"}).Return([]string{"1"}, nil)
		mockSpaceRepo.EXPECT().ListPersonalSpaceTpl(gomock.Any(), gomock.Any()).Return([]*dapo.DataAgentTplPo{}, nil)

		resp, err := svc.AgentTplList(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.IsLastPage)
	})

	t.Run("success with next page marker", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockSpaceRepo := idbaccessmock.NewMockIPersonalSpaceRepo(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &PersonalSpaceService{
			SvcBase:           service.NewSvcBase(),
			bizDomainHttp:     mockBiz,
			personalSpaceRepo: mockSpaceRepo,
			umHttp:            mockUm,
		}

		req := &personalspacereq.AgentTplListReq{Size: 1}
		ctx := createPersonalSpaceCtx("u1", "bd1")
		profile := "p1"
		pos := []*dapo.DataAgentTplPo{
			{ID: 1, Key: "k1", Name: "n1", Profile: &profile, CreatedBy: "u1", UpdatedBy: "u1", UpdatedAt: 1},
			{ID: 2, Key: "k2", Name: "n2", Profile: &profile, CreatedBy: "u1", UpdatedBy: "u1", UpdatedAt: 2},
		}

		mockBiz.EXPECT().GetAllAgentTplIDList(gomock.Any(), []string{"bd1"}).Return([]string{"1", "2"}, nil)
		mockSpaceRepo.EXPECT().ListPersonalSpaceTpl(gomock.Any(), gomock.Any()).Return(pos, nil)
		mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).Return(umtypes.NewOsnInfoMapS(), nil)

		resp, err := svc.AgentTplList(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.IsLastPage)
		assert.Len(t, resp.Entries, 1)
		assert.NotEmpty(t, resp.PaginationMarkerStr)
	})
}
