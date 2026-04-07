package permissionsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBuildAgentOperationItem_AgentPublish(t *testing.T) {
	t.Parallel()

	t.Run("builds operation item for AgentPublish", func(t *testing.T) {
		t.Parallel()

		op := cdapmsenum.AgentPublish
		item := buildAgentOperationItem(op)

		assert.NotNil(t, item)
		assert.Equal(t, string(op), item.ID)
		assert.Len(t, item.Name, 3)
		assert.Len(t, item.Scope, 1)
		assert.Contains(t, item.Scope, "type")
	})
}

func TestBuildAgentOperationItem_AgentUnpublish(t *testing.T) {
	t.Parallel()

	t.Run("builds operation item for AgentUnpublish", func(t *testing.T) {
		t.Parallel()

		op := cdapmsenum.AgentUnpublish
		item := buildAgentOperationItem(op)

		assert.NotNil(t, item)
		assert.Equal(t, string(op), item.ID)
		assert.Len(t, item.Name, 3)
		assert.Len(t, item.Scope, 1)
	})
}

func TestBuildAgentOperationItem_AgentUse(t *testing.T) {
	t.Parallel()

	t.Run("builds operation item for AgentUse", func(t *testing.T) {
		t.Parallel()

		op := cdapmsenum.AgentUse
		item := buildAgentOperationItem(op)

		assert.NotNil(t, item)
		assert.Equal(t, string(op), item.ID)
		assert.Len(t, item.Name, 3)
		assert.Len(t, item.Scope, 2)
		assert.Contains(t, item.Scope, "type")
		assert.Contains(t, item.Scope, "instance")
	})
}

func TestBuildAgentOperationItem_AllAgentOperators(t *testing.T) {
	t.Parallel()

	t.Run("builds operation items for all agent operators", func(t *testing.T) {
		t.Parallel()

		allOps := cdapmsenum.GetAllAgentOperator()

		for _, op := range allOps {
			item := buildAgentOperationItem(op)
			// Some operators may not have items (return nil)
			if item != nil {
				assert.NotEmpty(t, item.ID)
				assert.NotEmpty(t, item.Name)
				assert.NotEmpty(t, item.Scope)
			}
		}
	})
}

func TestBuildAgentOperationItem_UnknownOperator(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for unknown operator", func(t *testing.T) {
		t.Parallel()

		op := cdapmsenum.Operator("unknown_operator")
		item := buildAgentOperationItem(op)

		assert.Nil(t, item)
	})
}

func TestPermissionSvc_UpdateAgentResourceType_SetResourceTypeError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZHttp,
	}

	ctx := context.Background()
	httpErr := errors.New("http request failed")

	mockAuthZHttp.EXPECT().SetResourceType(gomock.Any(), gomock.Any(), gomock.Any()).Return(httpErr)

	err := svc.updateAgentResourceType(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "set resource type failed")
}

func TestPermissionSvc_UpdateAgentResourceType_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZHttp,
	}

	ctx := context.Background()

	mockAuthZHttp.EXPECT().SetResourceType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	err := svc.updateAgentResourceType(ctx)

	assert.NoError(t, err)
}
