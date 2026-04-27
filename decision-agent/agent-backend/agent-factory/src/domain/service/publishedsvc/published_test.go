package publishedsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	pubedreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Helper function to create context with business domain ID
func createPublishedCtx(bdID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, cenum.BizDomainIDCtxKey.String(), bdID) //nolint:staticcheck // SA1029

	return ctx
}

func setPublishedDisableBizDomain(t *testing.T, disable bool) {
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

func TestGetPubedTplList_BizDomainHttpError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := NewPublishedService(&NewPublishedSvcDto{
		SvcBase:          service.NewSvcBase(),
		PublishedTplRepo: mockPublishedTplRepo,
		BizDomainHttp:    mockBizDomainHttp,
		UmHttp:           mockUmHttp,
	})

	ctx := createPublishedCtx("test-bd-id")
	req := &pubedreq.PubedTplListReq{
		Size: 10,
	}

	httpErr := errors.New("http request failed")

	mockBizDomainHttp.EXPECT().GetAllAgentTplIDList(gomock.Any(), gomock.Any()).Return(nil, httpErr)

	res, err := svc.GetPubedTplList(ctx, req)

	// The function returns both response and error
	assert.Error(t, err)
	assert.NotNil(t, res) // Response is initialized even on error
	assert.Contains(t, err.Error(), "bizDomainHttp.GetAllAgentTplIDList failed")
}

func TestGetPubedTplList_NoTemplatesInBusinessDomain(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := NewPublishedService(&NewPublishedSvcDto{
		SvcBase:          service.NewSvcBase(),
		PublishedTplRepo: mockPublishedTplRepo,
		BizDomainHttp:    mockBizDomainHttp,
		UmHttp:           mockUmHttp,
	})

	ctx := createPublishedCtx("test-bd-id")
	req := &pubedreq.PubedTplListReq{
		Size: 10,
	}

	// When GetAllAgentTplIDList returns empty, the function returns early without calling the repo
	mockBizDomainHttp.EXPECT().GetAllAgentTplIDList(gomock.Any(), gomock.Any()).Return([]string{}, nil)

	res, err := svc.GetPubedTplList(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.IsLastPage)
}

func TestGetPubedTplList_RepoError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := NewPublishedService(&NewPublishedSvcDto{
		SvcBase:          service.NewSvcBase(),
		PublishedTplRepo: mockPublishedTplRepo,
		BizDomainHttp:    mockBizDomainHttp,
		UmHttp:           mockUmHttp,
	})

	ctx := createPublishedCtx("test-bd-id")
	req := &pubedreq.PubedTplListReq{
		Size: 10,
	}

	// Mock HTTP to return template IDs
	mockBizDomainHttp.EXPECT().GetAllAgentTplIDList(gomock.Any(), gomock.Any()).Return([]string{"tpl1"}, nil)

	// Mock repository to fail
	repoErr := errors.New("repository query failed")
	mockPublishedTplRepo.EXPECT().GetPubTplList(gomock.Any(), gomock.Any()).Return(nil, repoErr)

	res, err := svc.GetPubedTplList(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, res)
	assert.Contains(t, err.Error(), "publishedTplRepo.GetPubTplList failed")
}

func TestGetPubedTplList_EmptyResults(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := NewPublishedService(&NewPublishedSvcDto{
		SvcBase:          service.NewSvcBase(),
		PublishedTplRepo: mockPublishedTplRepo,
		BizDomainHttp:    mockBizDomainHttp,
		UmHttp:           mockUmHttp,
	})

	ctx := createPublishedCtx("test-bd-id")
	req := &pubedreq.PubedTplListReq{
		Size: 10,
	}

	mockBizDomainHttp.EXPECT().GetAllAgentTplIDList(gomock.Any(), gomock.Any()).Return([]string{"tpl1"}, nil)
	mockPublishedTplRepo.EXPECT().GetPubTplList(gomock.Any(), gomock.Any()).Return([]*dapo.PublishedTplPo{}, nil)

	res, err := svc.GetPubedTplList(ctx, req)

	// When repo returns empty list, function returns without error (no p2e conversion needed)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestGetPubedTplList_DisableBizDomainSkipsFilter(t *testing.T) {
	setPublishedDisableBizDomain(t, true)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := NewPublishedService(&NewPublishedSvcDto{
		SvcBase:          service.NewSvcBase(),
		PublishedTplRepo: mockPublishedTplRepo,
		UmHttp:           mockUmHttp,
	})

	ctx := createPublishedCtx("test-bd-id")
	req := &pubedreq.PubedTplListReq{
		Size: 10,
	}

	mockPublishedTplRepo.EXPECT().GetPubTplList(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, arg *pubedreq.PubedTplListReq) ([]*dapo.PublishedTplPo, error) {
			assert.Nil(t, arg.TplIDsByBd)
			return []*dapo.PublishedTplPo{}, nil
		},
	)

	res, err := svc.GetPubedTplList(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}
