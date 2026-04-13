// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines interfaces for find_skills skill recall
package interfaces

import "context"

// ==================== Request and Response Structures ====================

// FindSkillsReq find_skills request
type FindSkillsReq struct {
	// Header Fields
	AccountID   string `json:"-" header:"x-account-id"`
	AccountType string `json:"-" header:"x-account-type"`

	// Body Parameters
	KnID               string                   `json:"kn_id" validate:"required"`
	ObjectTypeID       string                   `json:"object_type_id" validate:"required"`
	InstanceIdentities []map[string]interface{} `json:"instance_identities"`
	SkillQuery         string                   `json:"skill_query"`
	TopK               int                      `json:"top_k" default:"10" validate:"min=1,max=20"`
}

// FindSkillsResp find_skills response
type FindSkillsResp struct {
	Entries []*SkillItem `json:"entries"`
	Message string       `json:"message,omitempty"`
}

// EmptyResultHint indicates why the coordinator returned empty matches.
// Only set when the empty result is due to a structural reason that the
// coordinator uniquely knows about (e.g. relations exist, no binding).
type EmptyResultHint string

const (
	HintNone                EmptyResultHint = ""
	HintNetworkScopeTooWide EmptyResultHint = "find_skills.network_scope_too_wide"
	HintObjectTypeNoBinding EmptyResultHint = "find_skills.object_type_no_binding"
)

// SkillItem candidate skill metadata
type SkillItem struct {
	SkillID     string `json:"skill_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ==================== Internal Structures ====================

// RecallMode recall mode determined by request parameters
type RecallMode int

const (
	RecallModeNetwork    RecallMode = 1 // kn_id only
	RecallModeObjectType RecallMode = 2 // kn_id + object_type_id
	RecallModeInstance   RecallMode = 3 // kn_id + object_type_id + instance_identities
)

// SkillMatch internal intermediate result for a single skill match
type SkillMatch struct {
	SkillID      string
	Name         string
	Description  string
	MatchedScope string  // "network" / "object_type" / "object_selector"
	Priority     int     // 100=instance, 50=object_type, 10=network
	Score        float64 // _score from ontology-query
}

// ==================== Service Interface ====================

// IFindSkillsService find_skills service interface
type IFindSkillsService interface {
	FindSkills(ctx context.Context, req *FindSkillsReq) (*FindSkillsResp, error)
}
