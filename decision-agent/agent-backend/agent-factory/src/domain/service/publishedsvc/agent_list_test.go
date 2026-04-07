package publishedsvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/cconf"
	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
)

func setDisablePmsCheck(t *testing.T, disable bool) {
	t.Helper()

	oldCfg := global.GConfig
	oldCConf := cglobal.GConfig

	global.GConfig = &conf.Config{
		Config:       cconf.BaseDefConfig(),
		SwitchFields: conf.NewSwitchFields(),
	}
	cglobal.GConfig = cconf.BaseDefConfig()
	global.GConfig.SwitchFields.DisablePmsCheck = disable

	t.Cleanup(func() {
		global.GConfig = oldCfg
		cglobal.GConfig = oldCConf
	})
}

func setDisableBizDomainForPublished(t *testing.T, disable bool) {
	t.Helper()

	oldCfg := global.GConfig
	oldCConf := cglobal.GConfig

	global.GConfig = &conf.Config{
		Config:       cconf.BaseDefConfig(),
		SwitchFields: conf.NewSwitchFields(),
	}
	cglobal.GConfig = cconf.BaseDefConfig()
	global.GConfig.SwitchFields.DisableBizDomain = disable

	t.Cleanup(func() {
		global.GConfig = oldCfg
		cglobal.GConfig = oldCConf
	})
}

func newPublishedJoinPo(id string, isPmsCtrl int, publishedAt int64) *dapo.PublishedJoinPo {
	profile := "profile-" + id

	return &dapo.PublishedJoinPo{
		DataAgentPo: dapo.DataAgentPo{
			ID:      id,
			Key:     "key-" + id,
			Name:    "name-" + id,
			Profile: &profile,
			Config:  "{}",
		},
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID:   "release-" + id,
			PublishedAt: publishedAt,
			PublishedBy: "u1",
			IsPmsCtrl:   isPmsCtrl,
		},
	}
}

func TestPublishedSvc_getPos(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	svc := &publishedSvc{pubedAgentRepo: mockRepo}

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()
		mockRepo.EXPECT().GetPubedList(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("db failed"))

		pos, err := svc.getPos(context.Background(), &pubedreq.PubedAgentListReq{}, []string{"a1"})
		assert.Error(t, err)
		assert.Nil(t, pos)
		assert.Contains(t, err.Error(), "get published agent list failed")
	})

	t.Run("filter by business domain ids", func(t *testing.T) {
		t.Parallel()
		mockRepo.EXPECT().GetPubedList(gomock.Any(), gomock.Any()).
			Return([]*dapo.PublishedJoinPo{
				newPublishedJoinPo("a1", 0, 11),
				newPublishedJoinPo("a2", 0, 12),
			}, nil)

		pos, err := svc.getPos(context.Background(), &pubedreq.PubedAgentListReq{}, []string{"a2"})
		assert.NoError(t, err)
		assert.Len(t, pos, 1)
		assert.Equal(t, "a2", pos[0].ID)
	})
}

func TestPublishedSvc_getPmsAgentPos(t *testing.T) {
	// 不使用 t.Parallel(): 子测试通过 setDisablePmsCheck 修改全局 global.GConfig/cglobal.GConfig
	t.Run("biz domain http error", func(t *testing.T) {
		// 不使用 t.Parallel(): setDisablePmsCheck 修改全局配置
		setDisablePmsCheck(t, false)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &publishedSvc{
			pubedAgentRepo: mockRepo,
			bizDomainHttp:  mockBizDomain,
			authZHttp:      mockAuthz,
		}

		mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).
			Return(nil, nil, errors.New("bd error"))

		pos, bdMap, isLastPage, err := svc.getPmsAgentPos(context.Background(), &pubedreq.PubedAgentListReq{Size: 10})
		assert.Error(t, err)
		assert.Nil(t, pos)
		assert.Nil(t, bdMap)
		assert.False(t, isLastPage)
	})

	t.Run("no agent id from biz domain", func(t *testing.T) {
		// 不使用 t.Parallel(): setDisablePmsCheck 修改全局配置
		setDisablePmsCheck(t, false)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &publishedSvc{
			pubedAgentRepo: mockRepo,
			bizDomainHttp:  mockBizDomain,
			authZHttp:      mockAuthz,
		}

		mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).
			Return([]string{}, map[string]string{}, nil)

		pos, bdMap, isLastPage, err := svc.getPmsAgentPos(context.Background(), &pubedreq.PubedAgentListReq{Size: 10})
		assert.NoError(t, err)
		assert.Empty(t, pos)
		assert.Empty(t, bdMap)
		assert.True(t, isLastPage)
	})

	t.Run("repo get list error", func(t *testing.T) {
		// 不使用 t.Parallel(): setDisablePmsCheck 修改全局配置
		setDisablePmsCheck(t, false)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &publishedSvc{
			pubedAgentRepo: mockRepo,
			bizDomainHttp:  mockBizDomain,
			authZHttp:      mockAuthz,
		}

		mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).
			Return([]string{"a1"}, map[string]string{"a1": "bd-1"}, nil)
		mockRepo.EXPECT().GetPubedList(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("repo error"))

		pos, bdMap, isLastPage, err := svc.getPmsAgentPos(context.Background(), &pubedreq.PubedAgentListReq{Size: 10})
		assert.Error(t, err)
		assert.Nil(t, pos)
		assert.Equal(t, "bd-1", bdMap["a1"])
		assert.False(t, isLastPage)
	})

	t.Run("disable pms check returns original pos", func(t *testing.T) {
		// 不使用 t.Parallel(): setDisablePmsCheck 修改全局配置
		setDisablePmsCheck(t, true)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &publishedSvc{
			pubedAgentRepo: mockRepo,
			bizDomainHttp:  mockBizDomain,
			authZHttp:      mockAuthz,
		}

		mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).
			Return([]string{"a1"}, map[string]string{"a1": "bd-1"}, nil)
		mockRepo.EXPECT().GetPubedList(gomock.Any(), gomock.Any()).
			Return([]*dapo.PublishedJoinPo{newPublishedJoinPo("a1", 1, 11)}, nil)

		pos, bdMap, isLastPage, err := svc.getPmsAgentPos(context.Background(), &pubedreq.PubedAgentListReq{Size: 10})
		assert.NoError(t, err)
		assert.Len(t, pos, 1)
		assert.Equal(t, "a1", pos[0].ID)
		assert.Equal(t, "bd-1", bdMap["a1"])
		assert.True(t, isLastPage)
	})

	t.Run("disable biz domain skips biz domain filter", func(t *testing.T) {
		setDisablePmsCheck(t, false)
		setDisableBizDomainForPublished(t, true)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &publishedSvc{
			pubedAgentRepo: mockRepo,
			authZHttp:      mockAuthz,
		}

		mockRepo.EXPECT().GetPubedList(gomock.Any(), gomock.Any()).
			Return([]*dapo.PublishedJoinPo{newPublishedJoinPo("a1", 0, 11)}, nil)
		mockAuthz.EXPECT().FilterCanUseAgentIDMap(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(map[string]struct{}{}, nil).AnyTimes()

		pos, bdMap, isLastPage, err := svc.getPmsAgentPos(context.Background(), &pubedreq.PubedAgentListReq{Size: 10})
		assert.NoError(t, err)
		assert.Len(t, pos, 1)
		assert.Empty(t, bdMap)
		assert.True(t, isLastPage)
	})

	t.Run("authz filter error", func(t *testing.T) {
		// 不使用 t.Parallel(): setDisablePmsCheck 修改全局配置
		setDisablePmsCheck(t, false)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &publishedSvc{
			pubedAgentRepo: mockRepo,
			bizDomainHttp:  mockBizDomain,
			authZHttp:      mockAuthz,
		}

		mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).
			Return([]string{"a1"}, map[string]string{"a1": "bd-1"}, nil)
		mockRepo.EXPECT().GetPubedList(gomock.Any(), gomock.Any()).
			Return([]*dapo.PublishedJoinPo{newPublishedJoinPo("a1", 1, 11)}, nil)
		mockAuthz.EXPECT().FilterCanUseAgentIDMap(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("authz failed"))

		pos, bdMap, isLastPage, err := svc.getPmsAgentPos(context.Background(), &pubedreq.PubedAgentListReq{Size: 1001})
		assert.Error(t, err)
		assert.Nil(t, pos)
		assert.Equal(t, "bd-1", bdMap["a1"])
		assert.True(t, isLastPage)
	})

	t.Run("pms filter loops and doubles page size", func(t *testing.T) {
		// 不使用 t.Parallel(): setDisablePmsCheck 修改全局配置
		setDisablePmsCheck(t, false)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &publishedSvc{
			pubedAgentRepo: mockRepo,
			bizDomainHttp:  mockBizDomain,
			authZHttp:      mockAuthz,
		}

		req := &pubedreq.PubedAgentListReq{Size: 1001}

		mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).
			Return([]string{"a1"}, map[string]string{"a1": "bd-1"}, nil)

		firstPagePos := make([]*dapo.PublishedJoinPo, 0, 1001)
		for i := 0; i < 1001; i++ {
			firstPagePos = append(firstPagePos, newPublishedJoinPo("a1", 1, int64(i+1)))
		}

		callCount := 0

		mockRepo.EXPECT().GetPubedList(gomock.Any(), gomock.Any()).Times(2).
			DoAndReturn(func(_ context.Context, r *pubedreq.PubedAgentListReq) ([]*dapo.PublishedJoinPo, error) {
				callCount++
				if callCount == 1 {
					assert.Equal(t, 1001, r.Size)
					return firstPagePos, nil
				}

				assert.Equal(t, 2002, r.Size)
				assert.NotNil(t, r.Marker)

				return []*dapo.PublishedJoinPo{newPublishedJoinPo("a1", 1, 2000)}, nil
			})

		authzCount := 0

		mockAuthz.EXPECT().FilterCanUseAgentIDMap(gomock.Any(), gomock.Any(), gomock.Any()).Times(2).
			DoAndReturn(func(_ context.Context, _ string, agentIDs []string) (map[string]struct{}, error) {
				authzCount++
				if authzCount == 1 {
					assert.Len(t, agentIDs, 1001)
					return map[string]struct{}{}, nil
				}

				assert.Equal(t, []string{"a1"}, agentIDs)

				return map[string]struct{}{"a1": {}}, nil
			})

		pos, bdMap, isLastPage, err := svc.getPmsAgentPos(context.Background(), req)
		assert.NoError(t, err)
		assert.Len(t, pos, 1)
		assert.Equal(t, "a1", pos[0].ID)
		assert.Equal(t, "bd-1", bdMap["a1"])
		assert.True(t, isLastPage)
	})
}

func TestPublishedSvc_GetPublishedAgentList(t *testing.T) {
	// 不使用 t.Parallel(): 子测试通过 setDisablePmsCheck 修改全局配置
	t.Run("get pms agent pos error", func(t *testing.T) {
		// 不使用 t.Parallel(): setDisablePmsCheck 修改全局配置
		setDisablePmsCheck(t, false)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &publishedSvc{
			pubedAgentRepo: mockRepo,
			bizDomainHttp:  mockBizDomain,
			authZHttp:      mockAuthz,
			umHttp:         mockUm,
		}

		mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).
			Return(nil, nil, errors.New("biz domain error"))

		res, err := svc.GetPublishedAgentList(context.Background(), &pubedreq.PubedAgentListReq{Size: 10})
		assert.Error(t, err)
		assert.NotNil(t, res)
	})

	t.Run("convert error from p2e", func(t *testing.T) {
		// 不使用 t.Parallel(): setDisablePmsCheck 修改全局配置
		setDisablePmsCheck(t, true)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &publishedSvc{
			pubedAgentRepo: mockRepo,
			bizDomainHttp:  mockBizDomain,
			authZHttp:      mockAuthz,
			umHttp:         mockUm,
		}

		mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).
			Return([]string{"a1"}, map[string]string{"a1": "bd-1"}, nil)
		mockRepo.EXPECT().GetPubedList(gomock.Any(), gomock.Any()).
			Return([]*dapo.PublishedJoinPo{newPublishedJoinPo("a1", 0, 11)}, nil)
		mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("um failed")).AnyTimes()

		res, err := svc.GetPublishedAgentList(context.Background(), &pubedreq.PubedAgentListReq{Size: 10})
		assert.Error(t, err)
		assert.NotNil(t, res)
		assert.Contains(t, err.Error(), "convert published agent list failed")
	})

	t.Run("success", func(t *testing.T) {
		// 不使用 t.Parallel(): setDisablePmsCheck 修改全局配置
		setDisablePmsCheck(t, true)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &publishedSvc{
			pubedAgentRepo: mockRepo,
			bizDomainHttp:  mockBizDomain,
			authZHttp:      mockAuthz,
			umHttp:         mockUm,
		}

		mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).
			Return([]string{"a1"}, map[string]string{"a1": "bd-1"}, nil)
		mockRepo.EXPECT().GetPubedList(gomock.Any(), gomock.Any()).
			Return([]*dapo.PublishedJoinPo{newPublishedJoinPo("a1", 0, 11)}, nil)

		umRet := umtypes.NewOsnInfoMapS()
		umRet.UserNameMap["u1"] = "user-1"
		mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).
			Return(umRet, nil).AnyTimes()

		res, err := svc.GetPublishedAgentList(context.Background(), &pubedreq.PubedAgentListReq{Size: 10})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Entries, 1)
		assert.Equal(t, "a1", res.Entries[0].ID)
		assert.Equal(t, "bd-1", res.Entries[0].BusinessDomainID)
		assert.True(t, res.IsLastPage)
	})
}

func TestPublishedSvc_getPmsAgentPos_SizeCapAt10000(t *testing.T) {
	// 不使用 t.Parallel(): setDisablePmsCheck 修改全局 global.GConfig/cglobal.GConfig
	setDisablePmsCheck(t, false)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	svc := &publishedSvc{
		pubedAgentRepo: mockRepo,
		bizDomainHttp:  mockBizDomain,
		authZHttp:      mockAuthz,
	}

	req := &pubedreq.PubedAgentListReq{Size: 6000}

	mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).
		Return([]string{"a1"}, map[string]string{"a1": "bd-1"}, nil)

	callCount := 0

	firstPagePos := make([]*dapo.PublishedJoinPo, 0, 6000)
	for i := 0; i < 6000; i++ {
		firstPagePos = append(firstPagePos, newPublishedJoinPo("a1", 1, int64(i+1)))
	}

	mockRepo.EXPECT().GetPubedList(gomock.Any(), gomock.Any()).Times(2).
		DoAndReturn(func(_ context.Context, r *pubedreq.PubedAgentListReq) ([]*dapo.PublishedJoinPo, error) {
			callCount++
			if callCount == 1 {
				assert.Equal(t, 6000, r.Size)
				return firstPagePos, nil
			}

			assert.Equal(t, 10000, r.Size)
			assert.NotNil(t, r.Marker)

			return []*dapo.PublishedJoinPo{newPublishedJoinPo("a1", 1, 10001)}, nil
		})
	mockAuthz.EXPECT().FilterCanUseAgentIDMap(gomock.Any(), gomock.Any(), gomock.Any()).Times(2).
		DoAndReturn(func(_ context.Context, _ string, _ []string) (map[string]struct{}, error) {
			return map[string]struct{}{}, nil
		})

	_, _, _, err := svc.getPmsAgentPos(context.Background(), req)
	assert.NoError(t, err)
}

func TestPublishedSvc_GetPublishedAgentList_EmptyPos(t *testing.T) {
	// 不使用 t.Parallel(): setDisablePmsCheck 修改全局 global.GConfig/cglobal.GConfig
	setDisablePmsCheck(t, true)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
	svc := &publishedSvc{
		pubedAgentRepo: mockRepo,
		bizDomainHttp:  mockBizDomain,
		authZHttp:      mockAuthz,
		umHttp:         mockUm,
	}

	mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).
		Return([]string{"a1"}, map[string]string{"a1": "bd-1"}, nil)
	mockRepo.EXPECT().GetPubedList(gomock.Any(), gomock.Any()).
		Return([]*dapo.PublishedJoinPo{}, nil)

	res, err := svc.GetPublishedAgentList(context.Background(), &pubedreq.PubedAgentListReq{Size: 10})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Empty(t, res.Entries)
}
