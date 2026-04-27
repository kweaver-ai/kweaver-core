package permissionsvc

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewPermissionService(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dto := &NewPermissionSvcDto{
		SvcBase:               service.NewSvcBase(),
		AgentConfigRepo:       idbaccessmock.NewMockIDataAgentConfigRepo(ctrl),
		ReleaseRepo:           idbaccessmock.NewMockIReleaseRepo(ctrl),
		ReleasePermissionRepo: idbaccessmock.NewMockIReleasePermissionRepo(ctrl),
		UmHttp:                httpaccmock.NewMockUmHttpAcc(ctrl),
		AuthZHttp:             authzaccmock.NewMockAuthZHttpAcc(ctrl),
	}

	svc := NewPermissionService(dto)

	assert.NotNil(t, svc)
}
