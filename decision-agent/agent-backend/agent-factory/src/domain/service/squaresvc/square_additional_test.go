package squaresvc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/daconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/common"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squareresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iusermanagementacc"
	"github.com/stretchr/testify/assert"
)

type noopSquareLogger struct{}

func (noopSquareLogger) Infof(string, ...interface{})  {}
func (noopSquareLogger) Infoln(...interface{})         {}
func (noopSquareLogger) Debugf(string, ...interface{}) {}
func (noopSquareLogger) Debugln(...interface{})        {}
func (noopSquareLogger) Errorf(string, ...interface{}) {}
func (noopSquareLogger) Errorln(...interface{})        {}
func (noopSquareLogger) Warnf(string, ...interface{})  {}
func (noopSquareLogger) Warnln(...interface{})         {}
func (noopSquareLogger) Panicf(string, ...interface{}) {}
func (noopSquareLogger) Panicln(...interface{})        {}
func (noopSquareLogger) Fatalf(string, ...interface{}) {}
func (noopSquareLogger) Fatalln(...interface{})        {}

type fakeUserMgmt struct {
	infos map[string]*iusermanagementacc.UserInfo
	err   error
}

func (f *fakeUserMgmt) GetUserInfoByUserID(context.Context, []string, []string) (map[string]*iusermanagementacc.UserInfo, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.infos, nil
}

func (f *fakeUserMgmt) GetUserIDByAccount(context.Context, string) (bool, string, error) {
	return false, "", nil
}

func (f *fakeUserMgmt) IsSuperAdmin(context.Context, string) (bool, error) {
	return false, nil
}

func TestSquareSvc_NewSquareService_Construct(t *testing.T) {
	t.Parallel()

	oldGlobal := global.GConfig
	oldCGlobal := cglobal.GConfig
	oldOnce := squareSvcOnce //nolint:govet
	oldImpl := squareSvcImpl

	global.GConfig = &conf.Config{Config: cconf.BaseDefConfig(), SwitchFields: conf.NewSwitchFields()}
	cglobal.GConfig = cconf.BaseDefConfig()
	squareSvcOnce = sync.Once{}
	squareSvcImpl = nil

	t.Cleanup(func() {
		global.GConfig = oldGlobal
		cglobal.GConfig = oldCGlobal
		squareSvcOnce = oldOnce //nolint:govet
		squareSvcImpl = oldImpl
	})

	svc1 := NewSquareService()
	svc2 := NewSquareService()

	assert.NotNil(t, svc1)
	assert.Same(t, svc1, svc2)
}

func TestSquareSvc_GetAgentInfo_UnpublishedSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

	svc := &squareSvc{
		SvcBase:            &service.SvcBase{Logger: noopSquareLogger{}},
		agentConfRepo:      mockAgentConfRepo,
		releaseHistoryRepo: mockReleaseHistoryRepo,
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
		ID:     "a1",
		Key:    "k1",
		Name:   "n1",
		Config: "{}",
	}, nil)
	mockReleaseHistoryRepo.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").
		Return(&dapo.ReleaseHistoryPO{AgentVersion: "v1"}, nil)

	res, err := svc.GetAgentInfo(context.Background(), &squarereq.AgentInfoReq{
		AgentID:      "a1",
		AgentVersion: daconstant.AgentVersionUnpublished,
		IsVisit:      false,
	})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, daconstant.AgentVersionUnpublished, res.Version)
	assert.Equal(t, "v1", res.LatestVersion)
	assert.Equal(t, "a1", res.ID)
}

func TestSquareSvc_notUnpublished_Additional(t *testing.T) {
	t.Parallel()

	t.Run("release repo error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

		svc := &squareSvc{
			SvcBase:            &service.SvcBase{Logger: noopSquareLogger{}},
			releaseRepo:        mockReleaseRepo,
			releaseHistoryRepo: mockReleaseHistoryRepo,
		}

		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(nil, errors.New("db failed"))

		err := svc.notUnpublished(context.Background(), &squarereq.AgentInfoReq{
			AgentID:      "a1",
			AgentVersion: "v1",
		}, &squareresp.AgentMarketAgentInfoResp{})
		assert.Error(t, err)
	})

	t.Run("empty version returns nil due wrapf(nil,...)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

		svc := &squareSvc{
			SvcBase:            &service.SvcBase{Logger: noopSquareLogger{}},
			releaseRepo:        mockReleaseRepo,
			releaseHistoryRepo: mockReleaseHistoryRepo,
		}

		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(nil, nil)
		mockReleaseHistoryRepo.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").Return(nil, nil)

		err := svc.notUnpublished(context.Background(), &squarereq.AgentInfoReq{
			AgentID:      "a1",
			AgentVersion: "",
		}, &squareresp.AgentMarketAgentInfoResp{})
		assert.NoError(t, err)
	})

	t.Run("history not found", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

		svc := &squareSvc{
			SvcBase:            &service.SvcBase{Logger: noopSquareLogger{}},
			releaseRepo:        mockReleaseRepo,
			releaseHistoryRepo: mockReleaseHistoryRepo,
		}

		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(&dapo.ReleasePO{AgentVersion: "v1"}, nil)
		mockReleaseHistoryRepo.EXPECT().GetByAgentIdVersion(gomock.Any(), "a1", "v2").Return(nil, nil)

		err := svc.notUnpublished(context.Background(), &squarereq.AgentInfoReq{
			AgentID:      "a1",
			AgentVersion: "v2",
		}, &squareresp.AgentMarketAgentInfoResp{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not exist")
	})

	t.Run("unmarshal error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

		svc := &squareSvc{
			SvcBase:                  &service.SvcBase{Logger: noopSquareLogger{}},
			releaseRepo:              mockReleaseRepo,
			releaseHistoryRepo:       mockReleaseHistoryRepo,
			usermanagementHttpClient: &fakeUserMgmt{},
		}

		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(&dapo.ReleasePO{AgentVersion: "v1"}, nil)
		mockReleaseHistoryRepo.EXPECT().GetByAgentIdVersion(gomock.Any(), "a1", "v1").
			Return(&dapo.ReleaseHistoryPO{AgentConfig: "{bad-json"}, nil)

		err := svc.notUnpublished(context.Background(), &squarereq.AgentInfoReq{
			AgentID:      "a1",
			AgentVersion: "v1",
		}, &squareresp.AgentMarketAgentInfoResp{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "json.Unmarshal")
	})

	t.Run("success with user name", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

		agentCfg, _ := json.Marshal(&dapo.DataAgentPo{
			ID:     "a1",
			Key:    "k1",
			Name:   "n1",
			Config: "{}",
		})
		releasePo := &dapo.ReleasePO{AgentVersion: "v1", AgentDesc: "desc", UpdateBy: "u1", UpdateTime: 10}
		historyPo := &dapo.ReleaseHistoryPO{
			AgentConfig: string(agentCfg),
			AgentDesc:   "history-desc",
			UpdateBy:    "u1",
			UpdateTime:  11,
		}

		svc := &squareSvc{
			SvcBase:            &service.SvcBase{Logger: noopSquareLogger{}},
			releaseRepo:        mockReleaseRepo,
			releaseHistoryRepo: mockReleaseHistoryRepo,
			usermanagementHttpClient: &fakeUserMgmt{
				infos: map[string]*iusermanagementacc.UserInfo{"u1": {Name: "user-1"}},
			},
		}

		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(releasePo, nil)
		mockReleaseHistoryRepo.EXPECT().GetByAgentIdVersion(gomock.Any(), "a1", "v1").Return(historyPo, nil)

		res := squareresp.NewAgentMarketAgentInfoResp()
		err := svc.notUnpublished(context.Background(), &squarereq.AgentInfoReq{
			AgentID:      "a1",
			AgentVersion: daconstant.AgentVersionLatest,
		}, res)
		assert.NoError(t, err)
		assert.Equal(t, "v1", res.Version)
		assert.Equal(t, "u1", res.PublishedBy)
		assert.Equal(t, "user-1", res.PublishedByName)
	})
}

func TestSquareSvc_GetRecentAgentList_SuccessAndUserInfoErr(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	svc := &squareSvc{
		SvcBase:                  &service.SvcBase{Logger: noopSquareLogger{}},
		releaseRepo:              mockReleaseRepo,
		usermanagementHttpClient: &fakeUserMgmt{err: errors.New("um failed")},
	}

	publishedCfg, _ := json.Marshal(&dapo.DataAgentPo{
		ID:     "a2",
		Key:    "k2",
		Name:   "n2",
		Config: "{}",
	})

	pos := []*dapo.RecentVisitAgentPO{
		{
			ReleaseAgentPO: dapo.ReleaseAgentPO{
				DataAgentPo:  dapo.DataAgentPo{ID: "a1", Key: "k1", Name: "n1", Config: "{}"},
				AgentVersion: sql.NullString{String: daconstant.AgentVersionUnpublished, Valid: true},
			},
			LastVisitTime: sql.NullInt64{Int64: 200, Valid: true},
		},
		{
			ReleaseAgentPO: dapo.ReleaseAgentPO{
				AgentConfig:   sql.NullString{String: string(publishedCfg), Valid: true},
				AgentVersion:  sql.NullString{String: "v1", Valid: true},
				AgentDesc:     sql.NullString{String: "desc", Valid: true},
				PublishTime:   sql.NullInt64{Int64: 10, Valid: true},
				PublishUserId: sql.NullString{String: "u1", Valid: true},
				DataAgentPo:   dapo.DataAgentPo{UpdatedBy: "u2"},
			},
			LastVisitTime: sql.NullInt64{Int64: 100, Valid: true},
		},
	}

	req := squarereq.AgentSquareRecentAgentReq{
		PageSize: common.PageSize{Page: 1, Size: 10},
	}
	mockReleaseRepo.EXPECT().ListRecentAgentForMarket(gomock.Any(), req).Return(pos, nil)

	res, err := svc.GetRecentAgentList(context.Background(), req)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "a1", res[0].ID)
	assert.Equal(t, "v1", res[1].Version)
}

func TestSquareSvc_GetRecentAgentList_UserInfoSuccessAndBadJSON(t *testing.T) {
	t.Parallel()

	t.Run("user info success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		svc := &squareSvc{
			SvcBase:     &service.SvcBase{Logger: noopSquareLogger{}},
			releaseRepo: mockReleaseRepo,
			usermanagementHttpClient: &fakeUserMgmt{
				infos: map[string]*iusermanagementacc.UserInfo{
					"u1": {Name: "user-1"},
					"u2": {Name: "user-2"},
				},
			},
		}

		publishedCfg, _ := json.Marshal(&dapo.DataAgentPo{
			ID:     "a2",
			Key:    "k2",
			Name:   "n2",
			Config: "{}",
		})
		req := squarereq.AgentSquareRecentAgentReq{PageSize: common.PageSize{Page: 1, Size: 10}}
		mockReleaseRepo.EXPECT().ListRecentAgentForMarket(gomock.Any(), req).Return([]*dapo.RecentVisitAgentPO{
			{
				ReleaseAgentPO: dapo.ReleaseAgentPO{
					AgentConfig:   sql.NullString{String: string(publishedCfg), Valid: true},
					AgentVersion:  sql.NullString{String: "v1", Valid: true},
					AgentDesc:     sql.NullString{String: "desc", Valid: true},
					PublishTime:   sql.NullInt64{Int64: 10, Valid: true},
					PublishUserId: sql.NullString{String: "u1", Valid: true},
					DataAgentPo:   dapo.DataAgentPo{UpdatedBy: "u2"},
				},
				LastVisitTime: sql.NullInt64{Int64: 1, Valid: true},
			},
		}, nil)

		res, err := svc.GetRecentAgentList(context.Background(), req)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "user-1", res[0].PublishedByName)
	})

	t.Run("bad json in published config", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		svc := &squareSvc{
			SvcBase:                  &service.SvcBase{Logger: noopSquareLogger{}},
			releaseRepo:              mockReleaseRepo,
			usermanagementHttpClient: &fakeUserMgmt{},
		}

		req := squarereq.AgentSquareRecentAgentReq{PageSize: common.PageSize{Page: 1, Size: 10}}
		mockReleaseRepo.EXPECT().ListRecentAgentForMarket(gomock.Any(), req).Return([]*dapo.RecentVisitAgentPO{
			{
				ReleaseAgentPO: dapo.ReleaseAgentPO{
					AgentConfig:  sql.NullString{String: "{bad-json", Valid: true},
					AgentVersion: sql.NullString{String: "v1", Valid: true},
				},
				LastVisitTime: sql.NullInt64{Int64: 1, Valid: true},
			},
		}, nil)

		res, err := svc.GetRecentAgentList(context.Background(), req)
		assert.Error(t, err)
		assert.Empty(t, res)
	})
}

func TestSquareSvc_GetAgentInfo_PublishedVersionSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

	agentCfg, _ := json.Marshal(&dapo.DataAgentPo{
		ID:     "a1",
		Key:    "k1",
		Name:   "n1",
		Config: "{}",
	})
	releasePo := &dapo.ReleasePO{AgentVersion: "v1", AgentDesc: "desc", UpdateBy: "u1", UpdateTime: 10}
	historyPo := &dapo.ReleaseHistoryPO{
		AgentConfig: string(agentCfg),
		AgentDesc:   "history-desc",
		UpdateBy:    "u1",
		UpdateTime:  11,
	}

	svc := &squareSvc{
		SvcBase:            &service.SvcBase{Logger: noopSquareLogger{}},
		agentConfRepo:      mockAgentConfRepo,
		releaseRepo:        mockReleaseRepo,
		releaseHistoryRepo: mockReleaseHistoryRepo,
		usermanagementHttpClient: &fakeUserMgmt{
			infos: map[string]*iusermanagementacc.UserInfo{"u1": {Name: "user-1"}},
		},
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
		ID:     "a1",
		Key:    "k1",
		Name:   "n1",
		Config: "{}",
	}, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(releasePo, nil)
	mockReleaseHistoryRepo.EXPECT().GetByAgentIdVersion(gomock.Any(), "a1", "v1").Return(historyPo, nil)

	res, err := svc.GetAgentInfo(context.Background(), &squarereq.AgentInfoReq{
		AgentID:      "a1",
		AgentVersion: daconstant.AgentVersionLatest,
		IsVisit:      false,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "v1", res.Version)
	assert.Equal(t, "user-1", res.PublishedByName)
}

func TestSquareSvc_notUnpublished_MoreBranches(t *testing.T) {
	t.Parallel()

	t.Run("GetByAgentIdVersion error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

		svc := &squareSvc{
			SvcBase:            &service.SvcBase{Logger: noopSquareLogger{}},
			releaseRepo:        mockReleaseRepo,
			releaseHistoryRepo: mockReleaseHistoryRepo,
		}

		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(&dapo.ReleasePO{AgentVersion: "v1"}, nil)
		mockReleaseHistoryRepo.EXPECT().GetByAgentIdVersion(gomock.Any(), "a1", "v1").
			Return(nil, errors.New("history failed"))

		err := svc.notUnpublished(context.Background(), &squarereq.AgentInfoReq{
			AgentID:      "a1",
			AgentVersion: "v1",
		}, squareresp.NewAgentMarketAgentInfoResp())
		assert.Error(t, err)
	})

	t.Run("user info error ignored", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

		agentCfg, _ := json.Marshal(&dapo.DataAgentPo{
			ID:     "a1",
			Key:    "k1",
			Name:   "n1",
			Config: "{}",
		})
		svc := &squareSvc{
			SvcBase:            &service.SvcBase{Logger: noopSquareLogger{}},
			releaseRepo:        mockReleaseRepo,
			releaseHistoryRepo: mockReleaseHistoryRepo,
			usermanagementHttpClient: &fakeUserMgmt{
				err: errors.New("um failed"),
			},
		}

		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(&dapo.ReleasePO{AgentVersion: "v1"}, nil)
		mockReleaseHistoryRepo.EXPECT().GetByAgentIdVersion(gomock.Any(), "a1", "v1").
			Return(&dapo.ReleaseHistoryPO{
				AgentConfig: string(agentCfg),
				AgentDesc:   "desc",
				UpdateBy:    "u1",
				UpdateTime:  1,
			}, nil)

		err := svc.notUnpublished(context.Background(), &squarereq.AgentInfoReq{
			AgentID:      "a1",
			AgentVersion: "v1",
		}, squareresp.NewAgentMarketAgentInfoResp())
		assert.NoError(t, err)
	})
}

func TestSquareSvc_GetAgentInfo_RecordVisitLogErrorIgnored(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockVisitRepo := idbaccessmock.NewMockIVisitHistoryRepo(ctrl)
	svc := &squareSvc{
		SvcBase:            &service.SvcBase{Logger: noopSquareLogger{}},
		agentConfRepo:      mockAgentConfRepo,
		releaseHistoryRepo: mockReleaseHistoryRepo,
		visitHistoryRepo:   mockVisitRepo,
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
		ID:     "a1",
		Key:    "k1",
		Name:   "n1",
		Config: "{}",
	}, nil)
	mockVisitRepo.EXPECT().IncVisitCount(gomock.Any(), gomock.Any()).Return(errors.New("visit failed"))
	mockReleaseHistoryRepo.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").
		Return(&dapo.ReleaseHistoryPO{AgentVersion: "v1"}, nil)

	res, err := svc.GetAgentInfo(context.Background(), &squarereq.AgentInfoReq{
		AgentID:      "a1",
		AgentVersion: daconstant.AgentVersionUnpublished,
		IsVisit:      true,
		UserID:       "u1",
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
