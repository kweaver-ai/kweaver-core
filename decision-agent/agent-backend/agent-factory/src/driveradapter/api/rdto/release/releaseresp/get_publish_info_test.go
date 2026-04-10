package releaseresp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestNewPublishInfoResp(t *testing.T) {
	t.Parallel()

	resp := NewPublishInfoResp()
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Categories)
	assert.NotNil(t, resp.PublishToWhere)
	assert.NotNil(t, resp.PublishToBes)
	assert.Empty(t, resp.Categories)
	assert.Empty(t, resp.PublishToWhere)
	assert.Empty(t, resp.PublishToBes)
}

func TestPublishInfoResp_SetPublishedToBes(t *testing.T) {
	t.Parallel()

	t.Run("API agent only", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishInfoResp()
		isAPIAgent := 1
		po := &dapo.ReleasePO{
			IsAPIAgent: &isAPIAgent,
		}

		resp.SetPublishedToBes(po)
		assert.Len(t, resp.PublishToBes, 1)
		assert.Equal(t, cdaenum.PublishToBeAPIAgent, resp.PublishToBes[0])
	})

	t.Run("WebSDK agent only", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishInfoResp()
		isWebSDKAgent := 1
		po := &dapo.ReleasePO{
			IsWebSDKAgent: &isWebSDKAgent,
		}

		resp.SetPublishedToBes(po)
		assert.Len(t, resp.PublishToBes, 1)
		assert.Equal(t, cdaenum.PublishToBeWebSDKAgent, resp.PublishToBes[0])
	})

	t.Run("multiple publish targets", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishInfoResp()
		isAPIAgent := 1
		isWebSDKAgent := 1
		isSkillAgent := 1
		isDataFlowAgent := 1
		po := &dapo.ReleasePO{
			IsAPIAgent:      &isAPIAgent,
			IsWebSDKAgent:   &isWebSDKAgent,
			IsSkillAgent:    &isSkillAgent,
			IsDataFlowAgent: &isDataFlowAgent,
		}

		resp.SetPublishedToBes(po)
		assert.Len(t, resp.PublishToBes, 4)
	})

	t.Run("no publish targets", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishInfoResp()
		po := &dapo.ReleasePO{}

		resp.SetPublishedToBes(po)
		assert.Empty(t, resp.PublishToBes)
	})
}

func TestPublishInfoResp_SetPublishToWhere(t *testing.T) {
	t.Parallel()

	t.Run("custom space only", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishInfoResp()
		isToCustomSpace := 1
		po := &dapo.ReleasePO{
			IsToCustomSpace: &isToCustomSpace,
		}

		resp.SetPublishToWhere(po)
		assert.Len(t, resp.PublishToWhere, 1)
		assert.Equal(t, daenum.PublishToWhereCustomSpace, resp.PublishToWhere[0])
	})

	t.Run("square only", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishInfoResp()
		isToSquare := 1
		po := &dapo.ReleasePO{
			IsToSquare: &isToSquare,
		}

		resp.SetPublishToWhere(po)
		assert.Len(t, resp.PublishToWhere, 1)
		assert.Equal(t, daenum.PublishToWhereSquare, resp.PublishToWhere[0])
	})

	t.Run("both custom space and square", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishInfoResp()
		isToCustomSpace := 1
		isToSquare := 1
		po := &dapo.ReleasePO{
			IsToCustomSpace: &isToCustomSpace,
			IsToSquare:      &isToSquare,
		}

		resp.SetPublishToWhere(po)
		assert.Len(t, resp.PublishToWhere, 2)
	})

	t.Run("no publish targets", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishInfoResp()
		po := &dapo.ReleasePO{}

		resp.SetPublishToWhere(po)
		assert.Empty(t, resp.PublishToWhere)
	})
}

func TestCategoryInfo_Fields(t *testing.T) {
	t.Parallel()

	info := &CategoryInfo{
		ID:   "cat-123",
		Name: "Test Category",
	}

	assert.Equal(t, "cat-123", info.ID)
	assert.Equal(t, "Test Category", info.Name)
}

func TestNewPmsControlResp(t *testing.T) {
	t.Parallel()

	resp := NewPmsControlResp()
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Roles)
	assert.NotNil(t, resp.Users)
	assert.NotNil(t, resp.UserGroups)
	assert.NotNil(t, resp.Departments)
	assert.NotNil(t, resp.AppAccounts)
	assert.Empty(t, resp.Roles)
	assert.Empty(t, resp.Users)
	assert.Empty(t, resp.UserGroups)
	assert.Empty(t, resp.Departments)
	assert.Empty(t, resp.AppAccounts)
}

func TestPublishInfoResp_Fields(t *testing.T) {
	t.Parallel()

	resp := &PublishInfoResp{
		Description: "Test description",
	}

	assert.Equal(t, "Test description", resp.Description)
	assert.Nil(t, resp.Categories)
	assert.Nil(t, resp.PublishToWhere)
	assert.Nil(t, resp.PublishToBes)
}

func TestCustomSpaceInfo_Fields(t *testing.T) {
	t.Parallel()

	info := &CustomSpaceInfo{
		SpaceID:   "space-123",
		SpaceName: "Test Space",
	}

	assert.Equal(t, "space-123", info.SpaceID)
	assert.Equal(t, "Test Space", info.SpaceName)
}

func TestPublishInfoResp_WithAllPublishTargets(t *testing.T) {
	t.Parallel()

	resp := NewPublishInfoResp()
	isAPIAgent := 1
	isWebSDKAgent := 1
	isSkillAgent := 1
	isDataFlowAgent := 1
	isToCustomSpace := 1
	isToSquare := 1
	po := &dapo.ReleasePO{
		IsAPIAgent:      &isAPIAgent,
		IsWebSDKAgent:   &isWebSDKAgent,
		IsSkillAgent:    &isSkillAgent,
		IsDataFlowAgent: &isDataFlowAgent,
		IsToCustomSpace: &isToCustomSpace,
		IsToSquare:      &isToSquare,
	}

	resp.SetPublishedToBes(po)
	resp.SetPublishToWhere(po)

	assert.Len(t, resp.PublishToBes, 4)
	assert.Len(t, resp.PublishToWhere, 2)
}
