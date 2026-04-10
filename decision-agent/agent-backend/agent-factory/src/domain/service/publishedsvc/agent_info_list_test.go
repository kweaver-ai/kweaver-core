package publishedsvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbret"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
)

func initCGlobalConfig(t *testing.T) {
	t.Helper()

	oldCfg := cglobal.GConfig
	cglobal.GConfig = cconf.BaseDefConfig()

	t.Cleanup(func() {
		cglobal.GConfig = oldCfg
	})
}

func TestPublishedSvc_GetPubedAgentInfoList_RepoError(t *testing.T) {
	// 不使用 t.Parallel(): initCGlobalConfig 修改全局 cglobal.GConfig
	initCGlobalConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
	svc := &publishedSvc{
		pubedAgentRepo: mockRepo,
		umHttp:         mockUm,
	}

	mockRepo.EXPECT().GetPubedListByXx(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("db failed"))

	res, err := svc.GetPubedAgentInfoList(context.Background(), &pubedreq.PAInfoListReq{
		AgentKeys: []string{"a1"},
	})

	assert.Error(t, err)
	assert.NotNil(t, res)
	assert.Contains(t, err.Error(), "get published agent list failed")
}

func TestPublishedSvc_GetPubedAgentInfoList_EmptyPos(t *testing.T) {
	// 不使用 t.Parallel(): initCGlobalConfig 修改全局 cglobal.GConfig
	initCGlobalConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
	svc := &publishedSvc{
		pubedAgentRepo: mockRepo,
		umHttp:         mockUm,
	}

	mockRepo.EXPECT().GetPubedListByXx(gomock.Any(), gomock.Any()).
		Return(&padbret.GetPaPoListByXxRet{JoinPos: []*dapo.PublishedJoinPo{}}, nil)

	res, err := svc.GetPubedAgentInfoList(context.Background(), &pubedreq.PAInfoListReq{
		AgentKeys: []string{"a1"},
	})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Empty(t, res.Entries)
}

func TestPublishedSvc_GetPubedAgentInfoList_ConvertError(t *testing.T) {
	// 不使用 t.Parallel(): initCGlobalConfig 修改全局 cglobal.GConfig
	initCGlobalConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
	svc := &publishedSvc{
		pubedAgentRepo: mockRepo,
		umHttp:         mockUm,
	}

	mockRepo.EXPECT().GetPubedListByXx(gomock.Any(), gomock.Any()).
		Return(&padbret.GetPaPoListByXxRet{
			JoinPos: []*dapo.PublishedJoinPo{
				{
					DataAgentPo: dapo.DataAgentPo{
						ID:     "a1",
						Config: "{invalid-json",
					},
					ReleasePartPo: dapo.ReleasePartPo{
						ReleaseID:   "r1",
						PublishedBy: "u1",
					},
				},
			},
		}, nil)
	mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).
		Return(umtypes.NewOsnInfoMapS(), nil).AnyTimes()

	res, err := svc.GetPubedAgentInfoList(context.Background(), &pubedreq.PAInfoListReq{
		AgentKeys:        []string{"a1"},
		NeedConfigFields: []string{"input"},
	})

	assert.Error(t, err)
	assert.NotNil(t, res)
	assert.Contains(t, err.Error(), "convert published agent list failed")
}

func TestPublishedSvc_GetPubedAgentInfoList_Success(t *testing.T) {
	// 不使用 t.Parallel(): initCGlobalConfig 修改全局 cglobal.GConfig
	initCGlobalConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
	svc := &publishedSvc{
		pubedAgentRepo: mockRepo,
		umHttp:         mockUm,
	}

	mockRepo.EXPECT().GetPubedListByXx(gomock.Any(), gomock.Any()).
		Return(&padbret.GetPaPoListByXxRet{
			JoinPos: []*dapo.PublishedJoinPo{
				{
					DataAgentPo: dapo.DataAgentPo{
						ID:     "a1",
						Key:    "agent-key-1",
						Name:   "agent-name-1",
						Config: "{}",
					},
					ReleasePartPo: dapo.ReleasePartPo{
						ReleaseID:   "r1",
						PublishedBy: "u1",
					},
				},
			},
		}, nil)
	mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).
		Return(umtypes.NewOsnInfoMapS(), nil).AnyTimes()

	res, err := svc.GetPubedAgentInfoList(context.Background(), &pubedreq.PAInfoListReq{
		AgentKeys:        []string{"a1"},
		NeedConfigFields: []string{"input"},
	})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Entries, 1)
	assert.Equal(t, "a1", res.Entries[0].ID)
}
