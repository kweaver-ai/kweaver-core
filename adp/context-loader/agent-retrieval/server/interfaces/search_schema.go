// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

// SearchSchemaReq search_schema standard request
type SearchSchemaReq struct {
	XAccountID   string `header:"x-account-id"`
	XAccountType string `header:"x-account-type"`
	XKnID        string `header:"x-kn-id"`

	Query        string             `json:"query" validate:"required"`
	KnID         string             `json:"kn_id,omitempty"`
	SearchScope  *SearchSchemaScope `json:"search_scope,omitempty"`
	MaxConcepts  *int               `json:"max_concepts,omitempty" default:"10"`
	SchemaBrief  *bool              `json:"schema_brief,omitempty" default:"true"`
	EnableRerank *bool              `json:"enable_rerank,omitempty" default:"true"`
}

// SearchSchemaScope search_schema scope controls
type SearchSchemaScope struct {
	IncludeObjectTypes   *bool `json:"include_object_types,omitempty" default:"true"`
	IncludeRelationTypes *bool `json:"include_relation_types,omitempty" default:"true"`
	IncludeActionTypes   *bool `json:"include_action_types,omitempty" default:"true"`
}

// SearchSchemaResp search_schema standard response
type SearchSchemaResp struct {
	ObjectTypes   []any `json:"object_types"`
	RelationTypes []any `json:"relation_types"`
	ActionTypes   []any `json:"action_types"`
}
