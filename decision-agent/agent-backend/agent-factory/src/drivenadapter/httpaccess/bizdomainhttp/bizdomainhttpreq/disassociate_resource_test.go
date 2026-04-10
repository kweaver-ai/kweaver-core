package bizdomainhttpreq

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestDisassociateResourceReq_StructFields(t *testing.T) {
	t.Parallel()

	req := DisassociateResourceReq{
		BdID: "bd-123",
		ID:   "resource-456",
		Type: cdaenum.ResourceTypeDataAgent,
	}

	assert.Equal(t, "bd-123", req.BdID)
	assert.Equal(t, "resource-456", req.ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgent, req.Type)
}

func TestDisassociateResourceReq_Empty(t *testing.T) {
	t.Parallel()

	req := DisassociateResourceReq{}

	assert.Empty(t, req.BdID)
	assert.Empty(t, req.ID)
	assert.Empty(t, req.Type)
}

func TestDisassociateResourceReq_WithDifferentTypes(t *testing.T) {
	t.Parallel()

	types := []cdaenum.ResourceType{
		cdaenum.ResourceTypeDataAgent,
		cdaenum.ResourceTypeDataAgentTpl,
	}

	for _, resourceType := range types {
		req := DisassociateResourceReq{
			BdID: "bd-123",
			ID:   "resource-456",
			Type: resourceType,
		}
		assert.Equal(t, resourceType, req.Type)
	}
}

func TestDisassociateResourceReq_WithBdID(t *testing.T) {
	t.Parallel()

	bdIDs := []string{
		"bd-001",
		"business-domain-123",
		"",
	}

	for _, bdID := range bdIDs {
		req := DisassociateResourceReq{
			BdID: bdID,
			ID:   "resource-456",
			Type: cdaenum.ResourceTypeDataAgent,
		}
		assert.Equal(t, bdID, req.BdID)
	}
}

func TestDisassociateResourceReq_WithResourceID(t *testing.T) {
	t.Parallel()

	resourceIDs := []string{
		"resource-001",
		"agent-123",
		"tpl-456",
		"",
	}

	for _, resourceID := range resourceIDs {
		req := DisassociateResourceReq{
			BdID: "bd-123",
			ID:   resourceID,
			Type: cdaenum.ResourceTypeDataAgent,
		}
		assert.Equal(t, resourceID, req.ID)
	}
}
