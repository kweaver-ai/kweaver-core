package personalspacereq

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/personal_space/personalspaceresp"
	"github.com/stretchr/testify/assert"
)

func TestAgentListReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{
		Name:                "test agent",
		PublishStatus:       cdaenum.StatusThreeStatePublished,
		PublishToBe:         cdaenum.PublishToBeAPIAgent,
		AgentCreatedType:    daenum.AgentCreatedTypeCreate,
		Size:                25,
		PaginationMarkerStr: "marker-string",
	}

	assert.Equal(t, "test agent", req.Name)
	assert.Equal(t, cdaenum.StatusThreeStatePublished, req.PublishStatus)
	assert.Equal(t, cdaenum.PublishToBeAPIAgent, req.PublishToBe)
	assert.Equal(t, daenum.AgentCreatedTypeCreate, req.AgentCreatedType)
	assert.Equal(t, 25, req.Size)
}

func TestAgentListReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{}

	errMap := req.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Empty(t, errMap)
}

func TestAgentListReq_CustomCheck_EmptyEnums(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{}

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestAgentListReq_CustomCheck_ValidPublishStatus(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{
		PublishStatus: cdaenum.StatusThreeStateUnpublished,
	}

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestAgentListReq_CustomCheck_ValidPublishToBe(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{
		PublishToBe: cdaenum.PublishToBeWebSDKAgent,
	}

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestAgentListReq_CustomCheck_ValidAgentCreatedType(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{
		AgentCreatedType: daenum.AgentCreatedTypeCopy,
	}

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestAgentListReq_CustomCheck_InvalidPublishStatus(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{
		PublishStatus: cdaenum.StatusThreeState("invalid"),
	}

	err := req.CustomCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish_status is invalid")
}

func TestAgentListReq_CustomCheck_InvalidPublishToBe(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{
		PublishToBe: cdaenum.PublishToBe("invalid"),
	}

	err := req.CustomCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish_to_be is invalid")
}

func TestAgentListReq_CustomCheck_InvalidAgentCreatedType(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{
		AgentCreatedType: daenum.AgentCreatedType("invalid_type"),
	}

	err := req.CustomCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_created_type is invalid")
}

func TestAgentListReq_LoadMarkerStr_Empty(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{
		PaginationMarkerStr: "",
	}

	err := req.LoadMarkerStr()

	assert.NoError(t, err)
	assert.Nil(t, req.Marker)
}

func TestAgentListReq_LoadMarkerStr_Valid(t *testing.T) {
	t.Parallel()

	marker := &personalspaceresp.PAListPaginationMarker{}
	marker.UpdatedAt = 987654
	marker.LastAgentID = "456"

	markerStr, _ := marker.ToString()

	req := &AgentListReq{
		PaginationMarkerStr: markerStr,
	}

	err := req.LoadMarkerStr()

	assert.NoError(t, err)
	assert.NotNil(t, req.Marker)
}

func TestAgentListReq_LoadMarkerStr_Invalid(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{
		PaginationMarkerStr: "not-valid-base64!",
	}

	err := req.LoadMarkerStr()

	assert.Error(t, err)
}

func TestAgentListReq_DefaultValues(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{}

	assert.Empty(t, req.Name)
	assert.Equal(t, cdaenum.StatusThreeState(""), req.PublishStatus)
	assert.Equal(t, cdaenum.PublishToBe(""), req.PublishToBe)
	assert.Equal(t, daenum.AgentCreatedType(""), req.AgentCreatedType)
	assert.Zero(t, req.Size)
	assert.Empty(t, req.PaginationMarkerStr)
}

func TestAgentListReq_WithAllFields(t *testing.T) {
	t.Parallel()

	req := &AgentListReq{
		Name:             "My Agent",
		PublishStatus:    cdaenum.StatusThreeStatePublishedEdited,
		PublishToBe:      cdaenum.PublishToBeSkillAgent,
		AgentCreatedType: daenum.AgentCreatedTypeCopy,
		Size:             100,
	}

	assert.Equal(t, "My Agent", req.Name)
	assert.Equal(t, cdaenum.StatusThreeStatePublishedEdited, req.PublishStatus)
	assert.Equal(t, cdaenum.PublishToBeSkillAgent, req.PublishToBe)
	assert.Equal(t, daenum.AgentCreatedTypeCopy, req.AgentCreatedType)
	assert.Equal(t, 100, req.Size)
}
