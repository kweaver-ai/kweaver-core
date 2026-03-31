package auditlogdto

type CustomSpaceCreateAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CustomSpaceUpdateAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CustomSpaceDeleteAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
