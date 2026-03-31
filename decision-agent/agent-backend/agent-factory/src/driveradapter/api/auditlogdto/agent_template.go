package auditlogdto

type AgentTemplateCreateAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentTemplateUpdateAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentTemplateDeleteAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentTemplateCopyAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentTemplatePublishAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentTemplateUnpublishAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentTemplateImportAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentTemplateExportAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentTemplateModifyPublishAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
