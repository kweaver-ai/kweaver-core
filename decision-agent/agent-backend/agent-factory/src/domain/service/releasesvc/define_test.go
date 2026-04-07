package releasesvc

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewReleaseService(t *testing.T) {
	t.Parallel()

	t.Run("creates service with all dependencies", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dto := &NewReleaseSvcDto{
			SvcBase:               service.NewSvcBase(),
			ReleaseRepo:           idbaccessmock.NewMockIReleaseRepo(ctrl),
			ReleaseHistoryRepo:    idbaccessmock.NewMockIReleaseHistoryRepo(ctrl),
			AgentConfigRepo:       idbaccessmock.NewMockIDataAgentConfigRepo(ctrl),
			ReleaseCategoryRepo:   idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl),
			ReleasePermissionRepo: idbaccessmock.NewMockIReleasePermissionRepo(ctrl),
			CategoryRepo:          idbaccessmock.NewMockICategoryRepo(ctrl),
		}

		svc := NewReleaseService(dto)

		assert.NotNil(t, svc)
		assert.IsType(t, &releaseSvc{}, svc)
	})

	t.Run("creates service with minimal dependencies", func(t *testing.T) {
		t.Parallel()

		dto := &NewReleaseSvcDto{
			SvcBase:               service.NewSvcBase(),
			ReleaseRepo:           nil,
			ReleaseHistoryRepo:    nil,
			AgentConfigRepo:       nil,
			ReleaseCategoryRepo:   nil,
			ReleasePermissionRepo: nil,
			CategoryRepo:          nil,
		}

		svc := NewReleaseService(dto)

		assert.NotNil(t, svc)
	})
}
