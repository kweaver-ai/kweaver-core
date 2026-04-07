package auditconstant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "agent", OBJECT_TYPE_AGENT)
	assert.Equal(t, "agent_template", OBJECT_TYPE_AGENT_TEMPLATE)
	assert.Equal(t, "custom_space", OBJECT_TYPE_CUSTOM_SPACE)
	assert.Equal(t, "product", OBJECT_TYPE_PRODUCT)
}

func TestOperation_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "create", CREATE)
	assert.Equal(t, "delete", DELETE)
	assert.Equal(t, "update", UPDATE)
	assert.Equal(t, "copy", COPY)
	assert.Equal(t, "publish", PUBLISH)
	assert.Equal(t, "unpublish", UNPUBLISH)
	assert.Equal(t, "import", IMPORT)
	assert.Equal(t, "export", EXPORT)
	assert.Equal(t, "modify_publish", MODIFY_PUBLISH)
	assert.Equal(t, "copy_publish", COPY_PUBLISH)
}

func TestGenerateAgentAuditObject(t *testing.T) {
	t.Parallel()

	obj := GenerateAgentAuditObject("agent-1", "Test Agent")
	assert.Equal(t, "agent", obj.Type)
	assert.Equal(t, "agent-1", obj.ID)
	assert.Equal(t, "Test Agent", obj.Name)
}

func TestGenerateAgentTemplateAuditObject(t *testing.T) {
	t.Parallel()

	obj := GenerateAgentTemplateAuditObject("tpl-1", "Test Template")
	assert.Equal(t, "agent_template", obj.Type)
	assert.Equal(t, "tpl-1", obj.ID)
	assert.Equal(t, "Test Template", obj.Name)
}

func TestGenerateCustomSpaceAuditObject(t *testing.T) {
	t.Parallel()

	obj := GenerateCustomSpaceAuditObject("space-1", "Test Space")
	assert.Equal(t, "custom_space", obj.Type)
	assert.Equal(t, "space-1", obj.ID)
	assert.Equal(t, "Test Space", obj.Name)
}

func TestGenerateProductAuditObject(t *testing.T) {
	t.Parallel()

	obj := GenerateProductAuditObject("product-1", "Test Product")
	assert.Equal(t, "product", obj.Type)
	assert.Equal(t, "product-1", obj.ID)
	assert.Equal(t, "Test Product", obj.Name)
}
