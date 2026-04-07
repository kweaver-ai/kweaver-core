package bizdomainhttpreq

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestQueryResourceAssociationsReq_StructFields(t *testing.T) {
	t.Parallel()

	req := QueryResourceAssociationsReq{
		BdID:   "bd-123",
		ID:     "resource-456",
		Type:   cdaenum.ResourceTypeDataAgent,
		Limit:  10,
		Offset: 0,
	}

	assert.Equal(t, "bd-123", req.BdID)
	assert.Equal(t, "resource-456", req.ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgent, req.Type)
	assert.Equal(t, 10, req.Limit)
	assert.Equal(t, 0, req.Offset)
}

func TestQueryResourceAssociationsReq_Empty(t *testing.T) {
	t.Parallel()

	req := QueryResourceAssociationsReq{}

	assert.Empty(t, req.BdID)
	assert.Empty(t, req.ID)
	assert.Empty(t, req.Type)
	assert.Equal(t, 0, req.Limit)
	assert.Equal(t, 0, req.Offset)
}

func TestQueryResourceAssociationsReq_WithDifferentTypes(t *testing.T) {
	t.Parallel()

	types := []cdaenum.ResourceType{
		cdaenum.ResourceTypeDataAgent,
		cdaenum.ResourceTypeDataAgentTpl,
	}

	for _, resourceType := range types {
		req := QueryResourceAssociationsReq{
			BdID:   "bd-123",
			ID:     "resource-456",
			Type:   resourceType,
			Limit:  10,
			Offset: 0,
		}
		assert.Equal(t, resourceType, req.Type)
	}
}

func TestQueryResourceAssociationsReq_WithBdID(t *testing.T) {
	t.Parallel()

	bdIDs := []string{
		"bd-001",
		"business-domain-123",
		"",
	}

	for _, bdID := range bdIDs {
		req := QueryResourceAssociationsReq{
			BdID:   bdID,
			ID:     "resource-456",
			Type:   cdaenum.ResourceTypeDataAgent,
			Limit:  10,
			Offset: 0,
		}
		assert.Equal(t, bdID, req.BdID)
	}
}

func TestQueryResourceAssociationsReq_WithPagination(t *testing.T) {
	t.Parallel()

	req := QueryResourceAssociationsReq{
		BdID:   "bd-123",
		Limit:  20,
		Offset: 40,
	}

	assert.Equal(t, "bd-123", req.BdID)
	assert.Equal(t, 20, req.Limit)
	assert.Equal(t, 40, req.Offset)
}

func TestQueryResourceAssociationsReq_WithZeroPagination(t *testing.T) {
	t.Parallel()

	req := QueryResourceAssociationsReq{
		BdID:   "bd-123",
		Limit:  0,
		Offset: 0,
	}

	assert.Equal(t, "bd-123", req.BdID)
	assert.Equal(t, 0, req.Limit)
	assert.Equal(t, 0, req.Offset)
}
