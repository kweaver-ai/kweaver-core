package personalspacesvc

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewPersonalSpaceService(t *testing.T) {
	t.Parallel()

	t.Run("creates service with all dependencies", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dto := &NewPersonalSpaceSvcDto{
			SvcBase:           service.NewSvcBase(),
			AgentTplRepo:      idbaccessmock.NewMockIDataAgentTplRepo(ctrl),
			AgentConfigRepo:   idbaccessmock.NewMockIDataAgentConfigRepo(ctrl),
			PersonalSpaceRepo: idbaccessmock.NewMockIPersonalSpaceRepo(ctrl),
			ReleaseRepo:       idbaccessmock.NewMockIReleaseRepo(ctrl),
			PubedAgentRepo:    idbaccessmock.NewMockIPubedAgentRepo(ctrl),
		}

		svc := NewPersonalSpaceService(dto)

		assert.NotNil(t, svc)
		assert.IsType(t, &PersonalSpaceService{}, svc)
	})

	t.Run("creates service with minimal dependencies", func(t *testing.T) {
		t.Parallel()

		dto := &NewPersonalSpaceSvcDto{
			SvcBase:           service.NewSvcBase(),
			AgentTplRepo:      nil,
			AgentConfigRepo:   nil,
			PersonalSpaceRepo: nil,
			ReleaseRepo:       nil,
			PubedAgentRepo:    nil,
		}

		svc := NewPersonalSpaceService(dto)

		assert.NotNil(t, svc)
	})
}
