package auditlogdto

type ProductCreateAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProductUpdateAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProductDeleteAuditLogInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
