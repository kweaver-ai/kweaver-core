package bizdomainhttpreq

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestAssociateResourceReq_StructFields(t *testing.T) {
	t.Parallel()

	req := AssociateResourceReq{
		BdID: "bd-123",
		ID:   "resource-456",
		Type: cdaenum.ResourceTypeDataAgent,
	}

	assert.Equal(t, "bd-123", req.BdID)
	assert.Equal(t, "resource-456", req.ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgent, req.Type)
}

func TestAssociateResourceReq_Empty(t *testing.T) {
	t.Parallel()

	req := AssociateResourceReq{}

	assert.Empty(t, req.BdID)
	assert.Empty(t, req.ID)
	assert.Empty(t, req.Type)
}

func TestAssociateResourceReq_WithDifferentTypes(t *testing.T) {
	t.Parallel()

	types := []cdaenum.ResourceType{
		cdaenum.ResourceTypeDataAgent,
		cdaenum.ResourceTypeDataAgentTpl,
	}

	for _, resourceType := range types {
		req := AssociateResourceReq{
			BdID: "bd-123",
			ID:   "resource-456",
			Type: resourceType,
		}
		assert.Equal(t, resourceType, req.Type)
	}
}

func TestAssociateResourceReq_WithBdID(t *testing.T) {
	t.Parallel()

	bdIDs := []string{
		"bd-001",
		"business-domain-123",
		"",
	}

	for _, bdID := range bdIDs {
		req := AssociateResourceReq{
			BdID: bdID,
			ID:   "resource-456",
			Type: cdaenum.ResourceTypeDataAgent,
		}
		assert.Equal(t, bdID, req.BdID)
	}
}

func TestAssociateResourceReq_WithResourceID(t *testing.T) {
	t.Parallel()

	resourceIDs := []string{
		"resource-001",
		"agent-123",
		"tpl-456",
		"",
	}

	for _, resourceID := range resourceIDs {
		req := AssociateResourceReq{
			BdID: "bd-123",
			ID:   resourceID,
			Type: cdaenum.ResourceTypeDataAgent,
		}
		assert.Equal(t, resourceID, req.ID)
	}
}

func TestAssociateResourceItem_StructFields(t *testing.T) {
	t.Parallel()

	item := &AssociateResourceItem{
		BdID: cenum.BizDomainPublic,
		ID:   "agent-123",
		Type: cdaenum.ResourceTypeDataAgent,
	}

	assert.Equal(t, cenum.BizDomainPublic, item.BdID)
	assert.Equal(t, "agent-123", item.ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgent, item.Type)
}

func TestNewInitAllAgentToPublicBusinessDomainReq(t *testing.T) {
	t.Parallel()

	agentIDs := []string{"agent-001", "agent-002", "agent-003"}
	req := NewInitAllAgentToPublicBusinessDomainReq(agentIDs)

	assert.Len(t, req, 3)
	assert.Equal(t, cenum.BizDomainPublic, req[0].BdID)
	assert.Equal(t, "agent-001", req[0].ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgent, req[0].Type)

	assert.Equal(t, cenum.BizDomainPublic, req[1].BdID)
	assert.Equal(t, "agent-002", req[1].ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgent, req[1].Type)

	assert.Equal(t, cenum.BizDomainPublic, req[2].BdID)
	assert.Equal(t, "agent-003", req[2].ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgent, req[2].Type)
}

func TestNewInitAllAgentToPublicBusinessDomainReq_Empty(t *testing.T) {
	t.Parallel()

	agentIDs := []string{}
	req := NewInitAllAgentToPublicBusinessDomainReq(agentIDs)

	assert.Len(t, req, 0)
}

func TestNewInitAllAgentToPublicBusinessDomainReq_Nil(t *testing.T) {
	t.Parallel()

	var agentIDs []string
	req := NewInitAllAgentToPublicBusinessDomainReq(agentIDs)

	assert.Len(t, req, 0)
}

func TestNewInitAllAgentTplToPublicBusinessDomainReq(t *testing.T) {
	t.Parallel()

	agentTplIDs := []string{"tpl-001", "tpl-002", "tpl-003"}
	req := NewInitAllAgentTplToPublicBusinessDomainReq(agentTplIDs)

	assert.Len(t, req, 3)
	assert.Equal(t, cenum.BizDomainPublic, req[0].BdID)
	assert.Equal(t, "tpl-001", req[0].ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgentTpl, req[0].Type)

	assert.Equal(t, cenum.BizDomainPublic, req[1].BdID)
	assert.Equal(t, "tpl-002", req[1].ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgentTpl, req[1].Type)

	assert.Equal(t, cenum.BizDomainPublic, req[2].BdID)
	assert.Equal(t, "tpl-003", req[2].ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgentTpl, req[2].Type)
}

func TestNewInitAllAgentTplToPublicBusinessDomainReq_Empty(t *testing.T) {
	t.Parallel()

	agentTplIDs := []string{}
	req := NewInitAllAgentTplToPublicBusinessDomainReq(agentTplIDs)

	assert.Len(t, req, 0)
}

func TestNewInitAllAgentTplToPublicBusinessDomainReq_Nil(t *testing.T) {
	t.Parallel()

	var agentTplIDs []string
	req := NewInitAllAgentTplToPublicBusinessDomainReq(agentTplIDs)

	assert.Len(t, req, 0)
}

func TestAssociateResourceBatchReq(t *testing.T) {
	t.Parallel()

	items := AssociateResourceBatchReq{
		&AssociateResourceItem{
			BdID: cenum.BizDomainPublic,
			ID:   "agent-001",
			Type: cdaenum.ResourceTypeDataAgent,
		},
		&AssociateResourceItem{
			BdID: cenum.BizDomainPublic,
			ID:   "tpl-001",
			Type: cdaenum.ResourceTypeDataAgentTpl,
		},
	}

	assert.Len(t, items, 2)
	assert.Equal(t, "agent-001", items[0].ID)
	assert.Equal(t, "tpl-001", items[1].ID)
}
