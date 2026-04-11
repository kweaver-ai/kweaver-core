package v3agentconfigsvc

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewDataAgentConfigService(t *testing.T) {
	t.Parallel()

	t.Run("creates service with all dependencies", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dto := &NewDaConfSvcDto{
			SvcBase:           service.NewSvcBase(),
			AgentConfRepo:     idbaccessmock.NewMockIDataAgentConfigRepo(ctrl),
			AgentTplRepo:      idbaccessmock.NewMockIDataAgentTplRepo(ctrl),
			ReleaseRepo:       idbaccessmock.NewMockIReleaseRepo(ctrl),
			PubedAgentRepo:    idbaccessmock.NewMockIPubedAgentRepo(ctrl),
			ProductRepo:       idbaccessmock.NewMockIProductRepo(ctrl),
			BdAgentRelRepo:    idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl),
			BdAgentTplRelRepo: idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl),
		}

		svc := NewDataAgentConfigService(dto)

		assert.NotNil(t, svc)
		assert.IsType(t, &dataAgentConfigSvc{}, svc)
	})

	t.Run("creates service with minimal dependencies", func(t *testing.T) {
		t.Parallel()

		dto := &NewDaConfSvcDto{
			SvcBase:           service.NewSvcBase(),
			AgentConfRepo:     nil,
			AgentTplRepo:      nil,
			ReleaseRepo:       nil,
			PubedAgentRepo:    nil,
			ProductRepo:       nil,
			BdAgentRelRepo:    nil,
			BdAgentTplRelRepo: nil,
		}

		svc := NewDataAgentConfigService(dto)

		assert.NotNil(t, svc)
	})
}
