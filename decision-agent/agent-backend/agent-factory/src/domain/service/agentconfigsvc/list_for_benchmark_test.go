package v3agentconfigsvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/cconf"
	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	idbaccessmock "github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/stretchr/testify/assert"
)

func setListForBenchmarkDisableBizDomain(t *testing.T, disable bool) {
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

func TestDataAgentConfigSvc_ListForBenchmark_BizDomainError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		bizDomainHttp: mockBizDomain,
	}

	ctx := context.Background()
	req := &agentconfigreq.ListForBenchmarkReq{Name: "test"}

	mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("bizdomain error"))

	_, err := svc.ListForBenchmark(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get agent list by biz domain failed")
}

func TestDataAgentConfigSvc_ListForBenchmark_EmptyAgentIDs(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		bizDomainHttp: mockBizDomain,
	}

	ctx := context.Background()
	req := &agentconfigreq.ListForBenchmarkReq{Name: "test"}

	mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).Return([]string{}, nil, nil)

	resp, err := svc.ListForBenchmark(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Entries)
}

func TestDataAgentConfigSvc_ListForBenchmark_RepoError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		bizDomainHttp: mockBizDomain,
		agentConfRepo: mockAgentConfRepo,
	}

	ctx := context.Background()
	req := &agentconfigreq.ListForBenchmarkReq{Name: "test"}

	mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).Return([]string{"agent-1", "agent-2"}, nil, nil)
	mockAgentConfRepo.EXPECT().ListForBenchmark(gomock.Any(), gomock.Any()).Return(nil, int64(0), errors.New("db error"))

	_, err := svc.ListForBenchmark(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "[ListForBenchmark] failed")
}

func TestDataAgentConfigSvc_ListForBenchmark_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		bizDomainHttp: mockBizDomain,
		agentConfRepo: mockAgentConfRepo,
	}

	ctx := context.Background()
	req := &agentconfigreq.ListForBenchmarkReq{Name: "test"}

	agentIDs := []string{"agent-1"}
	pos := []*dapo.ListForBenchmarkPo{
		{ID: "agent-1", Name: "Test Agent", Key: "key-1"},
	}

	mockBizDomain.EXPECT().GetAllAgentIDList(gomock.Any(), gomock.Any()).Return(agentIDs, nil, nil)
	mockAgentConfRepo.EXPECT().ListForBenchmark(gomock.Any(), gomock.Any()).Return(pos, int64(1), nil)

	resp, err := svc.ListForBenchmark(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.Total)
	assert.Len(t, resp.Entries, 1)
}

func TestDataAgentConfigSvc_ListForBenchmark_DisableBizDomainSkipsFilter(t *testing.T) {
	setListForBenchmarkDisableBizDomain(t, true)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
	}

	ctx := context.Background()
	req := &agentconfigreq.ListForBenchmarkReq{Name: "test"}

	pos := []*dapo.ListForBenchmarkPo{
		{ID: "agent-1", Name: "Test Agent", Key: "key-1"},
	}

	mockAgentConfRepo.EXPECT().ListForBenchmark(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, arg *agentconfigreq.ListForBenchmarkReq) ([]*dapo.ListForBenchmarkPo, int64, error) {
			assert.Nil(t, arg.AgentIDsByBizDomain)
			return pos, int64(1), nil
		},
	)

	resp, err := svc.ListForBenchmark(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.Total)
	assert.Len(t, resp.Entries, 1)
}
