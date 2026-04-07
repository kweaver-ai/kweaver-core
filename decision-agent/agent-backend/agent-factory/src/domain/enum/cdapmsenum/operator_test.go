package cdapmsenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperator_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		operator Operator
	}{
		{"agent publish", AgentPublish},
		{"agent unpublish", AgentUnpublish},
		{"agent unpublish other user", AgentUnpublishOtherUserAgent},
		{"agent publish to be skill agent", AgentPublishToBeSkillAgent},
		{"agent publish to be web sdk agent", AgentPublishToBeWebSdkAgent},
		{"agent publish to be api agent", AgentPublishToBeApiAgent},
		{"agent publish to be data flow agent", AgentPublishToBeDataFlowAgent},
		{"agent create system agent", AgentCreateSystemAgent},
		{"agent built in agent mgmt", AgentBuiltInAgentMgmt},
		{"agent see trajectory analysis", AgentSeeTrajectoryAnalysis},
		{"agent use", AgentUse},
		{"agent tpl publish", AgentTplPublish},
		{"agent tpl unpublish", AgentTplUnpublish},
		{"agent tpl unpublish other user", AgentTplUnpublishOtherUserAgentTpl},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.operator.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestOperator_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidOp := Operator("invalid_operator")
	err := invalidOp.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid operator")
}

func TestOperator_String(t *testing.T) {
	t.Parallel()

	op := AgentPublish
	assert.Equal(t, "publish", op.String())
}

func TestGetAllOperator(t *testing.T) {
	t.Parallel()

	operators := GetAllOperator()

	assert.Len(t, operators, 14)

	// Check for some key operators
	assert.Contains(t, operators, AgentPublish)
	assert.Contains(t, operators, AgentUnpublish)
	assert.Contains(t, operators, AgentUse)
	assert.Contains(t, operators, AgentTplPublish)
}

func TestGetAllAgentMgmtOperator(t *testing.T) {
	t.Parallel()

	operators := GetAllAgentMgmtOperator()

	assert.Len(t, operators, 10)

	// Check that it contains management operators
	assert.Contains(t, operators, AgentPublish)
	assert.Contains(t, operators, AgentUnpublish)
	assert.Contains(t, operators, AgentBuiltInAgentMgmt)

	// Check that it doesn't contain AgentUse
	assert.NotContains(t, operators, AgentUse)
}

func TestGetAllAgentUseOperator(t *testing.T) {
	t.Parallel()

	operators := GetAllAgentUseOperator()

	assert.Len(t, operators, 1)
	assert.Contains(t, operators, AgentUse)
}

func TestGetAllAgentOperator(t *testing.T) {
	t.Parallel()

	operators := GetAllAgentOperator()

	// Should be the combination of mgmt and use operators
	assert.Len(t, operators, 11) // 10 mgmt + 1 use

	// Check both categories
	assert.Contains(t, operators, AgentPublish)
	assert.Contains(t, operators, AgentUse)
}

func TestGetAllAgentTplOperator(t *testing.T) {
	t.Parallel()

	operators := GetAllAgentTplOperator()

	assert.Len(t, operators, 3)

	// Check template operators
	assert.Contains(t, operators, AgentTplPublish)
	assert.Contains(t, operators, AgentTplUnpublish)
	assert.Contains(t, operators, AgentTplUnpublishOtherUserAgentTpl)
}
