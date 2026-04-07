// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

// KnOperationType Business knowledge network operator
type KnOperationType string

const (
	KnOperationTypeAnd            KnOperationType = "and"       // AND
	KnOperationTypeOr             KnOperationType = "or"        // OR
	KnOperationTypeEqual          KnOperationType = "=="        // Equal
	KnOperationTypeNotEqual       KnOperationType = "!="        // Not Equal
	KnOperationTypeGreater        KnOperationType = ">"         // Greater than
	KnOperationTypeLess           KnOperationType = "<"         // Less than
	KnOperationTypeGreaterOrEqual KnOperationType = ">="        // Greater than or equal
	KnOperationTypeLessOrEqual    KnOperationType = "<="        // Less than or equal
	KnOperationTypeIn             KnOperationType = "in"        // in
	KnOperationTypeNotIn          KnOperationType = "not_in"    // not_in
	KnOperationTypeLike           KnOperationType = "like"      // like
	KnOperationTypeNotLike        KnOperationType = "not_like"  // not_like
	KnOperationTypeRange          KnOperationType = "range"     // range
	KnOperationTypeOutRange       KnOperationType = "out_range" // out_range
	KnOperationTypeExist          KnOperationType = "exist"     // exist
	KnOperationTypeNotExist       KnOperationType = "not_exist" // not_exist
	KnOperationTypeRegex          KnOperationType = "regex"     // regex
	KnOperationTypeMatch          KnOperationType = "match"     // match
	KnOperationTypeKnn            KnOperationType = "knn"       // knn
)

// LogicPropertyType Logic property type
type LogicPropertyType string

const (
	LogicPropertyTypeMetric   LogicPropertyType = "metric"   // Metric type
	LogicPropertyTypeOperator LogicPropertyType = "operator" // Operator type
)

type KnBaseError struct {
	ErrorCode               string         `json:"error_code"`    // Error code
	Description             string         `json:"description"`   // Error description
	Solution                string         `json:"solution"`      // Solution
	ErrorLink               string         `json:"error_link"`    // Error link
	ErrorDetails            interface{}    `json:"error_details"` // Detail content
	DescriptionTemplateData map[string]any `json:"-"`             // Description parameters
	SolutionTemplateData    map[string]any `json:"-"`             // Solution parameters
}

type ResourceInfo struct {
	Type string `json:"type"` // Data source type
	ID   string `json:"id"`   // Data view ID
	Name string `json:"name"` // View name
}

type SimpleObjectType struct {
	OTID   string `json:"id"`
	OTName string `json:"name"`
	Icon   string `json:"icon"`
	Color  string `json:"color"`
}

// DataProperty Data property structure definition
type DataProperty struct {
	// Name is the property name. Can only contain lowercase letters, numbers, underscores (_),
	// hyphens (-), and cannot start with underscore or hyphen
	Name                string            `json:"name"`
	DisplayName         string            `json:"display_name"`         // Property display name
	Type                string            `json:"type"`                 // Property data type. In addition to view field types, there are metric, objective, event, trace, log, operator
	Comment             string            `json:"comment"`              // Comment
	MappedField         any               `json:"mapped_field"`         // View field info
	ConditionOperations []KnOperationType `json:"condition_operations"` // List of query condition operators supported by this data property
}

// LogicPropertyDef Logic property definition (extracted from object type definition)
type LogicPropertyDef struct {
	Name        string              `json:"name"`
	DisplayName string              `json:"display_name,omitempty"`
	Type        LogicPropertyType   `json:"type"` // Logic property type: metric or operator
	Comment     string              `json:"comment,omitempty"`
	DataSource  map[string]any      `json:"data_source,omitempty"`
	Parameters  []PropertyParameter `json:"parameters,omitempty"`
}

// PropertyParameter Property parameter definition
type PropertyParameter struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	ValueFrom        string `json:"value_from"` // "input", "property", "const"
	Value            any    `json:"value,omitempty"`
	IfSystemGenerate bool   `json:"if_system_generate,omitempty"`
	Comment          string `json:"comment,omitempty"`
}

// ObjectType Object type structure definition
type ObjectType struct {
	ModuleType      string              `json:"module_type"` // Module type
	ID              string              `json:"id"`          // Object ID
	Name            string              `json:"name"`        // Object name
	Tags            []string            `json:"tags"`        // Tags
	Comment         string              `json:"comment"`     // Comment
	Score           float64             `json:"_score"`      // Score
	DataSource      *ResourceInfo       `json:"data_source"`
	DataProperties  []*DataProperty     `json:"data_properties,omitempty"`  // Data properties
	LogicProperties []*LogicPropertyDef `json:"logic_properties,omitempty"` // Logic properties
	PrimaryKeys     []string            `json:"primary_keys"`               // Primary key fields
}

// RelationType Relation type structure definition
type RelationType struct {
	ModuleType string   `json:"module_type"` // Module type
	ID         string   `json:"id"`          // Relation type ID
	Name       string   `json:"name"`        // Relation type name
	Tags       []string `json:"tags"`        // Tags
	Comment    string   `json:"comment"`     // Comment
	Score      float64  `json:"_score"`      // Score

	SourceObjectTypeID string `json:"source_object_type_id"`        // Source object type ID
	TargetObjectTypeID string `json:"target_object_type_id"`        // Target object type ID
	SourceObjectType   any    `json:"source_object_type,omitempty"` // Provide name when viewing details
	TargetObjectType   any    `json:"target_object_type,omitempty"` // Provide name when viewing details
	MappingRules       any    `json:"mapping_rules"`                // Mapping rules based on type, direct corresponds to []Mapping structure
	Type               string `json:"type"`                         // Relation type
}

// ActionType Action type structure definition
type ActionType struct {
	ModuleType string   `json:"module_type"` // Module type
	ID         string   `json:"id"`          // Action type ID
	Name       string   `json:"name"`        // Action type name
	Tags       []string `json:"tags"`        // Tags
	Comment    string   `json:"comment"`     // Comment
	Score      float64  `json:"_score"`      // Score

	ObjectTypeID string `json:"object_type_id"` // Object type ID bound to action type
}

type KnCondValueFrom string

const (
	CondValueFromConst KnCondValueFrom = "const"
)

type KnCondLimitKey string

const (
	CondLimitKeyK           KnCondLimitKey = "k"            // Pagination key
	CondLimitKeyMinScore    KnCondLimitKey = "min_score"    // Min score
	CondLimitKeyMinDistance KnCondLimitKey = "min_distance" // Min distance
)

// KnCondition Retrieval condition
type KnCondition struct {
	Field         string          `json:"field"`          // Field name
	Operation     KnOperationType `json:"operation"`      // Operator
	SubConditions []*KnCondition  `json:"sub_conditions"` // Sub filtering conditions
	Value         any             `json:"value"`          // Field value
	ValueFrom     KnCondValueFrom `json:"value_from"`     // Field value source
	LimitKey      KnCondLimitKey  `json:"limit_key"`
	LimitValue    any             `json:"limit_value"`
}

type KnSortParams struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// QueryConceptsReq Query concepts request
type QueryConceptsReq struct {
	KnID      string          `json:"-"`         // Knowledge network ID
	Cond      *KnCondition    `json:"condition"` // Retrieval condition
	Sort      []*KnSortParams `json:"sort"`
	Limit     int             `json:"limit"`      // Return count, default 10. Range 1-10000
	NeedTotal bool            `json:"need_total"` // Whether total count is needed, default false
}

// Concepts Retrieved concepts list
type Concepts struct {
	Entries     []any `json:"entries"`
	TotalCount  int64 `json:"total_count,omitempty"`
	SearchAfter []any `json:"search_after,omitempty"`
	OverallMs   int64 `json:"overall_ms"`
}

// ObjectTypeConcepts Object type concepts list
type ObjectTypeConcepts struct {
	Entries    []*ObjectType `json:"entries"`               // Object type data
	TotalCount int64         `json:"total_count,omitempty"` // Total count
}

// RelationTypeConcepts Relation type concepts list
type RelationTypeConcepts struct {
	Entries    []*RelationType `json:"entries"`               // Relation type data
	TotalCount int64           `json:"total_count,omitempty"` // Total count
}

// ActionTypeConcepts Action type concepts list
type ActionTypeConcepts struct {
	Entries    []*ActionType `json:"entries"`               // Action type data
	TotalCount int64         `json:"total_count,omitempty"` // Total count
}

// OntologyJobState Ontology job state
type OntologyJobState string

const (
	OntologyJobStatePending   OntologyJobState = "pending"   // Pending
	OntologyJobStateRunning   OntologyJobState = "running"   // Running
	OntologyJobStateCompleted OntologyJobState = "completed" // Completed
	OntologyJobStateCanceled  OntologyJobState = "canceled"  // Canceled
	OntologyJobStateFailed    OntologyJobState = "failed"    // Failed
)

// OntologyJobType Ontology job type
type OntologyJobType string

const (
	OntologyJobTypeFull OntologyJobType = "full" // Full build job
)

// AccountInfo Account information
type AccountInfo struct {
	ID   string `json:"id"`   // Account ID
	Type string `json:"type"` // Account type
	Name string `json:"name"` // Account name
}

// CreateFullBuildOntologyJobReq Request to create full ontology build job
type CreateFullBuildOntologyJobReq struct {
	Name string `json:"name" validate:"required"` // Job name
}

// CreateJobResp Response when creating a job
type CreateJobResp struct {
	ID string `json:"id"` // Job ID
}

// OntologyJob Ontology job details
type OntologyJob struct {
	ID           string           `json:"id"`            // Job ID
	Name         string           `json:"name"`          // Job name
	KnID         string           `json:"kn_id"`         // Knowledge network ID
	State        OntologyJobState `json:"state"`         // Job state
	StateDetail  string           `json:"state_detail"`  // State details
	Creator      *AccountInfo     `json:"creator"`       // Creator
	CreateTime   int64            `json:"create_time"`   // Create time (timestamp)
	FinishedTime int64            `json:"finished_time"` // Finished time (timestamp)
	TimeCost     int64            `json:"time_cost"`     // Time cost (seconds)
	JobType      OntologyJobType  `json:"job_type"`      // Job type
}

// ListOntologyJobsReq Request to list ontology jobs
type ListOntologyJobsReq struct {
	NamePattern string           `form:"name_pattern" json:"name_pattern"` // Job name pattern filter
	State       OntologyJobState `form:"state" json:"state"`               // Job state filter
	JobType     OntologyJobType  `form:"job_type" json:"job_type"`         // Job type filter
	Limit       int              `form:"limit" json:"limit"`               // Return count
	Direction   string           `form:"direction" json:"direction"`       // Sort direction (asc/desc)
	Offset      int              `form:"offset" json:"offset"`             // Pagination offset
}

// ListOntologyJobsResp Response for listing ontology jobs
type ListOntologyJobsResp struct {
	Entries    []*OntologyJob `json:"entries"`     // Job list
	TotalCount int64          `json:"total_count"` // Total count
}

// KnowledgeNetworkDetail Knowledge network detail with full schema
type KnowledgeNetworkDetail struct {
	ID            string          `json:"id"`             // Knowledge network ID
	Name          string          `json:"name"`           // Knowledge network name
	Comment       string          `json:"comment"`        // Comment/description
	ObjectTypes   []*ObjectType   `json:"object_types"`   // Object types
	RelationTypes []*RelationType `json:"relation_types"` // Relation types
	ActionTypes   []*ActionType   `json:"action_types"`   // Action types
}

// BknBackendAccess BKN backend ontology management interface
type BknBackendAccess interface {
	// GetKnowledgeNetworkDetail Get knowledge network detail with full schema (include_detail=true, mode=export)
	GetKnowledgeNetworkDetail(ctx context.Context, knID string) (*KnowledgeNetworkDetail, error)

	// SearchObjectTypes Search object types
	SearchObjectTypes(ctx context.Context, query *QueryConceptsReq) (objectTypes *ObjectTypeConcepts, err error)
	// GetObjectTypeDetail Get object type details
	GetObjectTypeDetail(ctx context.Context, knID string, otIds []string, includeDetail bool) ([]*ObjectType, error)

	// SearchRelationTypes Search relation types
	SearchRelationTypes(ctx context.Context, query *QueryConceptsReq) (releationTypes *RelationTypeConcepts, err error)
	// GetRelationTypeDetail Get relation type details
	GetRelationTypeDetail(ctx context.Context, knID string, rtIDs []string, includeDetail bool) ([]*RelationType, error)

	// SearchActionTypes Search action types
	SearchActionTypes(ctx context.Context, query *QueryConceptsReq) (actionTypes *ActionTypeConcepts, err error)
	// GetActionTypeDetail Get action type details
	GetActionTypeDetail(ctx context.Context, knID string, atIDs []string, includeDetail bool) ([]*ActionType, error)

	// CreateFullBuildOntologyJob Create a full ontology build job
	CreateFullBuildOntologyJob(ctx context.Context, knID string, req *CreateFullBuildOntologyJobReq) (resp *CreateJobResp, err error)
	// ListOntologyJobs List ontology jobs with filters
	ListOntologyJobs(ctx context.Context, knID string, req *ListOntologyJobsReq) (resp *ListOntologyJobsResp, err error)
}