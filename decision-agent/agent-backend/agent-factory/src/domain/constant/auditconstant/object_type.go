package auditconstant

import "github.com/kweaver-ai/kweaver-go-lib/audit"

const (
	OBJECT_TYPE_AGENT          = "agent"          // 智能体/智能體/Agent
	OBJECT_TYPE_AGENT_TEMPLATE = "agent_template" // 智能体模板/智能體模板/Agent Template
	OBJECT_TYPE_CUSTOM_SPACE   = "custom_space"   // 自定义空间/自定义空間/Custom Space
	OBJECT_TYPE_PRODUCT        = "product"        // 产品/產品/Product
)

func GenerateAgentAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_AGENT,
		ID:   id,
		Name: name,
	}
}

func GenerateAgentTemplateAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_AGENT_TEMPLATE,
		ID:   id,
		Name: name,
	}
}

func GenerateCustomSpaceAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_CUSTOM_SPACE,
		ID:   id,
		Name: name,
	}
}

func GenerateProductAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: OBJECT_TYPE_PRODUCT,
		ID:   id,
		Name: name,
	}
}
