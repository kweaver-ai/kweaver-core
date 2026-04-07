package publishvo

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/pmsvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestPublishInfo_Fields(t *testing.T) {
	t.Parallel()

	pmsControl := &pmsvo.PmsControlObjS{
		RoleIDs:       []string{"role-1"},
		UserIDs:       []string{"user-1"},
		UserGroupIDs:  []string{"group-1"},
		DepartmentIDs: []string{"dept-1"},
		AppAccountIDs: []string{"app-1"},
	}

	info := &PublishInfo{
		CategoryIDs:    []string{"cat-1", "cat-2"},
		Description:    "Test publish",
		PublishToWhere: []daenum.PublishToWhere{daenum.PublishToWhereCustomSpace, daenum.PublishToWhereSquare},
		PmsControl:     pmsControl,
		PublishToBes:   []cdaenum.PublishToBe{cdaenum.PublishToBeSkillAgent, cdaenum.PublishToBeAPIAgent},
	}

	assert.Equal(t, []string{"cat-1", "cat-2"}, info.CategoryIDs)
	assert.Equal(t, "Test publish", info.Description)
	assert.Len(t, info.PublishToWhere, 2)
	assert.Equal(t, []string{"role-1"}, info.PmsControl.RoleIDs)
	assert.Equal(t, []string{"user-1"}, info.PmsControl.UserIDs)
	assert.Equal(t, []string{"group-1"}, info.PmsControl.UserGroupIDs)
	assert.Equal(t, []string{"dept-1"}, info.PmsControl.DepartmentIDs)
	assert.Equal(t, []string{"app-1"}, info.PmsControl.AppAccountIDs)
	assert.Len(t, info.PublishToBes, 2)
}

func TestPublishInfo_Empty(t *testing.T) {
	t.Parallel()

	info := &PublishInfo{}

	assert.Nil(t, info.CategoryIDs)
	assert.Empty(t, info.Description)
	assert.Nil(t, info.PublishToWhere)
	assert.Nil(t, info.PmsControl)
	assert.Nil(t, info.PublishToBes)
}

func TestListPublishInfo_New(t *testing.T) {
	t.Parallel()

	info := NewListPublishInfo()
	assert.NotNil(t, info)
	assert.Equal(t, 0, info.IsAPIAgent)
	assert.Equal(t, 0, info.IsWebSDKAgent)
	assert.Equal(t, 0, info.IsSkillAgent)
	assert.Equal(t, 0, info.IsDataFlowAgent)
}

func TestListPublishInfo_Fields(t *testing.T) {
	t.Parallel()

	info := &ListPublishInfo{
		PublishedToBeStruct: dapo.PublishedToBeStruct{
			IsAPIAgent:      1,
			IsWebSDKAgent:   1,
			IsSkillAgent:    1,
			IsDataFlowAgent: 0,
		},
	}

	assert.Equal(t, 1, info.IsAPIAgent)
	assert.Equal(t, 1, info.IsWebSDKAgent)
	assert.Equal(t, 1, info.IsSkillAgent)
	assert.Equal(t, 0, info.IsDataFlowAgent)
}
