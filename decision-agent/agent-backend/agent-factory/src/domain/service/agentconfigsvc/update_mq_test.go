package v3agentconfigsvc

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccessmock"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentConfigSvc_HandleUpdateNameMq_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMq := imqaccessmock.NewMockIMqAccess(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:  service.NewSvcBase(),
		mqAccess: mockMq,
	}

	mockMq.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	err := svc.handleUpdateNameMq("agent-001", "新Agent名称")
	assert.NoError(t, err)
}

func TestDataAgentConfigSvc_HandleUpdateNameMq_PublishError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMq := imqaccessmock.NewMockIMqAccess(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:  service.NewSvcBase(),
		mqAccess: mockMq,
	}

	mockMq.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("mq unavailable"))

	err := svc.handleUpdateNameMq("agent-001", "新Agent名称")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish msg failed")
}
