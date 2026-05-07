// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "github.com/kweaver-ai/kweaver-go-lib/audit"

type contextKey string // 自定义专属的key类型

const (
	CONTENT_TYPE_NAME = "Content-Type"
	CONTENT_TYPE_JSON = "application/json"

	HTTP_HEADER_METHOD_OVERRIDE = "x-http-method-override"
	HTTP_HEADER_ACCOUNT_ID      = "x-account-id"
	HTTP_HEADER_ACCOUNT_TYPE    = "x-account-type"
	HTTP_HEADER_BUSINESS_DOMAIN = "x-business-domain"

	X_REQUEST_TOOK = "x-request-took"

	ACCOUNT_INFO_KEY contextKey = "x-account-info" // 避免直接使用string

	NAME_MAX_LENGTH        = 255
	DESCRIPTION_MAX_LENGTH = 1000
	TAGS_MAX_NUMBER        = 5
	TAG_MAX_LENGTH         = 40
	TAG_INVALID_CHARACTER  = "/:?\\\"<>|：？‘’“”！《》,#[]{}%&*$^!=.'"

	DEFAULT_OFFSET = "0"
	MIN_OFFSET     = 0

	DEFAULT_LIMIT = "20"
	MIN_LIMIT     = 1
	MAX_LIMIT     = 1000
	NO_LIMIT      = "-1"

	DEFAULT_DIRECTION = "desc"
	DESC_DIRECTION    = "desc"
	ASC_DIRECTION     = "asc"

	MODULE_TYPE_CATALOG        = "catalog"
	MODULE_TYPE_RESOURCE       = "resource"
	MODULE_TYPE_CONNECTOR_TYPE = "connector_type"
)

// AccountInfo represents user/account information.
type AccountInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// PaginationQueryParams holds common pagination parameters.
type PaginationQueryParams struct {
	Offset    int    `form:"offset" json:"offset"`
	Limit     int    `form:"limit" json:"limit"`
	Sort      string `form:"sort" json:"sort"`
	Direction string `form:"direction" json:"direction"`
}

func GenerateCatalogAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: MODULE_TYPE_CATALOG,
		ID:   id,
		Name: name,
	}
}

func GenerateConnectorTypeAuditObject(typ string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: MODULE_TYPE_CONNECTOR_TYPE,
		ID:   typ,
		Name: name,
	}
}

func GenerateResourceAuditObject(id string, name string) audit.AuditObject {
	return audit.AuditObject{
		Type: MODULE_TYPE_RESOURCE,
		ID:   id,
		Name: name,
	}
}
