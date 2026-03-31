package publishedsvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
)

func TestPublishedSvc_GetPubedTplList_LenGreaterThanNeedSize(t *testing.T) {
	t.Parallel()

	initCGlobalConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
	svc := &publishedSvc{
		bizDomainHttp:    mockBizDomain,
		publishedTplRepo: mockPublishedTplRepo,
		umHttp:           mockUm,
	}

	ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1") //nolint:staticcheck // SA1029
	req := &pubedreq.PubedTplListReq{Size: 1}

	mockBizDomain.EXPECT().GetAllAgentTplIDList(gomock.Any(), []string{"bd-1"}).
		Return([]string{"tpl-1"}, nil)
	mockPublishedTplRepo.EXPECT().GetPubTplList(gomock.Any(), gomock.Any()).
		Return([]*dapo.PublishedTplPo{
			{ID: 1, TplID: 101, Name: "tpl-1", PublishedBy: "u1"},
			{ID: 2, TplID: 102, Name: "tpl-2", PublishedBy: "u2"},
		}, nil)

	umRet := umtypes.NewOsnInfoMapS()
	umRet.UserNameMap["u1"] = "user-1"
	mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).
		Return(umRet, nil).AnyTimes()

	res, err := svc.GetPubedTplList(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Entries, 1)
	assert.False(t, res.IsLastPage)
	assert.NotEmpty(t, res.PaginationMarkerStr)
}

func TestPublishedSvc_GetPubedTplList_ConvertError(t *testing.T) {
	t.Parallel()

	initCGlobalConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
	svc := &publishedSvc{
		bizDomainHttp:    mockBizDomain,
		publishedTplRepo: mockPublishedTplRepo,
		umHttp:           mockUm,
	}

	ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1") //nolint:staticcheck // SA1029
	req := &pubedreq.PubedTplListReq{Size: 1}

	mockBizDomain.EXPECT().GetAllAgentTplIDList(gomock.Any(), []string{"bd-1"}).
		Return([]string{"tpl-1"}, nil)
	mockPublishedTplRepo.EXPECT().GetPubTplList(gomock.Any(), gomock.Any()).
		Return([]*dapo.PublishedTplPo{
			{ID: 1, TplID: 101, Name: "tpl-1", PublishedBy: "u1"},
		}, nil)
	mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("um failed")).AnyTimes()

	res, err := svc.GetPubedTplList(ctx, req)
	assert.Error(t, err)
	assert.NotNil(t, res)
	assert.Contains(t, err.Error(), "convert published agent template list failed")
}

func TestPublishedSvc_GetPubedTplList_SuccessLastPage(t *testing.T) {
	t.Parallel()

	initCGlobalConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
	svc := &publishedSvc{
		bizDomainHttp:    mockBizDomain,
		publishedTplRepo: mockPublishedTplRepo,
		umHttp:           mockUm,
	}

	ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1") //nolint:staticcheck // SA1029
	req := &pubedreq.PubedTplListReq{Size: 2}

	mockBizDomain.EXPECT().GetAllAgentTplIDList(gomock.Any(), []string{"bd-1"}).
		Return([]string{"tpl-1"}, nil)
	mockPublishedTplRepo.EXPECT().GetPubTplList(gomock.Any(), gomock.Any()).
		Return([]*dapo.PublishedTplPo{
			{ID: 1, TplID: 101, Name: "tpl-1", PublishedBy: "u1"},
		}, nil)

	umRet := umtypes.NewOsnInfoMapS()
	umRet.UserNameMap["u1"] = "user-1"
	mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).
		Return(umRet, nil).AnyTimes()

	res, err := svc.GetPubedTplList(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Entries, 1)
	assert.True(t, res.IsLastPage)
	assert.Empty(t, res.PaginationMarkerStr)
}
