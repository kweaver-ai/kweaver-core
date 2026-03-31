package agentinoutsvc

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewAgentInOutService(t *testing.T) {
	t.Parallel()

	t.Run("creates service with all dependencies", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dto := &NewAgentInOutSvcDto{
			SvcBase:        service.NewSvcBase(),
			Logger:         nil,
			AgentConfRepo:  idbaccessmock.NewMockIDataAgentConfigRepo(ctrl),
			PmsSvc:         nil,
			BizDomainHttp:  bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl),
			BdAgentRelRepo: idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl),
		}

		svc := NewAgentInOutService(dto)

		assert.NotNil(t, svc)
		assert.IsType(t, &agentInOutSvc{}, svc)
	})

	t.Run("creates service with minimal dependencies", func(t *testing.T) {
		t.Parallel()

		dto := &NewAgentInOutSvcDto{
			SvcBase:        service.NewSvcBase(),
			Logger:         nil,
			AgentConfRepo:  nil,
			PmsSvc:         nil,
			BizDomainHttp:  nil,
			BdAgentRelRepo: nil,
		}

		svc := NewAgentInOutService(dto)

		assert.NotNil(t, svc)
	})
}
