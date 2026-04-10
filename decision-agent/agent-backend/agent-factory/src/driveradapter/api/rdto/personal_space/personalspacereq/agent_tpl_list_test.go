package personalspacereq

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/personal_space/personalspaceresp"
	"github.com/stretchr/testify/assert"
)

func TestAgentTplListReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &AgentTplListReq{
		Name:                "test template",
		ProductKey:          "product-123",
		CategoryID:          "category-456",
		Size:                20,
		PaginationMarkerStr: "marker-string",
	}

	assert.Equal(t, "test template", req.Name)
	assert.Equal(t, "product-123", req.ProductKey)
	assert.Equal(t, "category-456", req.CategoryID)
	assert.Equal(t, 20, req.Size)
	assert.Equal(t, "marker-string", req.PaginationMarkerStr)
}

func TestAgentTplListReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &AgentTplListReq{}

	errMap := req.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Empty(t, errMap)
}

func TestAgentTplListReq_CustomCheck_EmptyEnums(t *testing.T) {
	t.Parallel()

	req := &AgentTplListReq{}

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestAgentTplListReq_CustomCheck_ValidPublishStatus(t *testing.T) {
	t.Parallel()

	req := &AgentTplListReq{
		PublishStatus: cdaenum.StatusPublished,
	}

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestAgentTplListReq_CustomCheck_ValidAgentTplCreatedType(t *testing.T) {
	t.Parallel()

	req := &AgentTplListReq{
		AgentTplCreatedType: daenum.AgentTplCreatedTypeCopyFromAgent,
	}

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestAgentTplListReq_CustomCheck_InvalidPublishStatus(t *testing.T) {
	t.Parallel()

	req := &AgentTplListReq{
		PublishStatus: cdaenum.Status("invalid"),
	}

	err := req.CustomCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish_status is invalid")
}

func TestAgentTplListReq_CustomCheck_InvalidAgentTplCreatedType(t *testing.T) {
	t.Parallel()

	req := &AgentTplListReq{
		AgentTplCreatedType: daenum.AgentTplCreatedType("invalid_type"),
	}

	err := req.CustomCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_tpl_created_type is invalid")
}

func TestAgentTplListReq_LoadMarkerStr_Empty(t *testing.T) {
	t.Parallel()

	req := &AgentTplListReq{
		PaginationMarkerStr: "",
	}

	err := req.LoadMarkerStr()

	assert.NoError(t, err)
	assert.Nil(t, req.Marker)
}

func TestAgentTplListReq_LoadMarkerStr_Valid(t *testing.T) {
	t.Parallel()

	// Create a marker string by serializing a marker
	marker := &personalspaceresp.PTplListPaginationMarker{}
	marker.UpdatedAt = 123456
	marker.LastTplID = 789

	markerStr, _ := marker.ToString()

	req := &AgentTplListReq{
		PaginationMarkerStr: markerStr,
	}

	err := req.LoadMarkerStr()

	assert.NoError(t, err)
	assert.NotNil(t, req.Marker)
}

func TestAgentTplListReq_LoadMarkerStr_Invalid(t *testing.T) {
	t.Parallel()

	req := &AgentTplListReq{
		PaginationMarkerStr: "invalid-base64-string",
	}

	err := req.LoadMarkerStr()

	assert.Error(t, err)
}

func TestAgentTplListReq_DefaultValues(t *testing.T) {
	t.Parallel()

	req := &AgentTplListReq{}

	assert.Empty(t, req.Name)
	assert.Empty(t, req.ProductKey)
	assert.Empty(t, req.CategoryID)
	assert.Equal(t, cdaenum.Status(""), req.PublishStatus)
	assert.Equal(t, daenum.AgentTplCreatedType(""), req.AgentTplCreatedType)
	assert.Zero(t, req.Size)
	assert.Empty(t, req.PaginationMarkerStr)
}

func TestAgentTplListReq_WithAllFields(t *testing.T) {
	t.Parallel()

	req := &AgentTplListReq{
		Name:                "My Template",
		ProductKey:          "product-key",
		CategoryID:          "cat-id",
		PublishStatus:       cdaenum.StatusPublished,
		AgentTplCreatedType: daenum.AgentTplCreatedTypeCopyFromTpl,
		Size:                50,
		PaginationMarkerStr: "eyJ1cGRhdGVkX2F0IjoxMjM0NTYsImxhc3RfdHBsX2lkIjo3ODl9",
	}

	assert.Equal(t, "My Template", req.Name)
	assert.Equal(t, cdaenum.StatusPublished, req.PublishStatus)
	assert.Equal(t, daenum.AgentTplCreatedTypeCopyFromTpl, req.AgentTplCreatedType)
	assert.Equal(t, 50, req.Size)
}
