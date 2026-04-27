package auditlogdto

type AgentCreateAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentUpdateAuditLogInfo struct {
	ID      string `json:"id"`
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
}

type AgentDeleteAuditLogInfo struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type AgentCopyAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentCopy2TplAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentCopy2TplAndPublishAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentPublishAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentUnPublishAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentImportAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentExportAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentModifyPublishAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
