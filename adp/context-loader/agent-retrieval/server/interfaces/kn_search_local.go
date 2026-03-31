// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines interfaces for knowledge network search (local implementation)
package interfaces

import "context"

// ==================== Request Structures ====================

// KnSearchLocalRequest 知识网络检索本地请求
type KnSearchLocalRequest struct {
	// Header Fields
	AccountID   string `json:"-" header:"x-account-id"`
	AccountType string `json:"-" header:"x-account-type"`

	// Request Body
	Query           string                   `json:"query" validate:"required"`
	KnID            string                   `json:"kn_id" validate:"required"`
	RetrievalConfig *KnSearchRetrievalConfig `json:"retrieval_config,omitempty"`
	OnlySchema      bool                     `json:"only_schema" default:"false"`
	EnableRerank    bool                     `json:"enable_rerank" default:"true"`
}

// KnSearchRetrievalConfig 召回配置参数
type KnSearchRetrievalConfig struct {
	ConceptRetrieval          *KnSearchConceptRetrievalConfig          `json:"concept_retrieval,omitempty"`
	SemanticInstanceRetrieval *KnSearchSemanticInstanceRetrievalConfig `json:"semantic_instance_retrieval,omitempty"`
	PropertyFilter            *KnSearchPropertyFilterConfig            `json:"property_filter,omitempty"`
}

// KnSearchConceptRetrievalConfig 概念召回配置参数
type KnSearchConceptRetrievalConfig struct {
	TopK                   int   `json:"top_k" default:"10"`
	IncludeSampleData      *bool `json:"include_sample_data" default:"false"`
	SchemaBrief            *bool `json:"schema_brief" default:"true"`
	EnableCoarseRecall     *bool `json:"enable_coarse_recall" default:"true"`
	CoarseObjectLimit      int   `json:"coarse_object_limit" default:"2000"`
	CoarseRelationLimit    int   `json:"coarse_relation_limit" default:"300"`
	CoarseMinRelationCount int   `json:"coarse_min_relation_count" default:"5000"`
	EnablePropertyBrief    *bool `json:"enable_property_brief" default:"true"`
	PerObjectPropertyTopK  int   `json:"per_object_property_top_k" default:"8"`
	GlobalPropertyTopK     int   `json:"global_property_top_k" default:"30"`
}

// KnSearchSemanticInstanceRetrievalConfig 语义实例召回配置参数
type KnSearchSemanticInstanceRetrievalConfig struct {
	InitialCandidateCount             int     `json:"initial_candidate_count" default:"50"`
	PerTypeInstanceLimit              int     `json:"per_type_instance_limit" default:"5"`
	MaxSemanticSubConditions          int     `json:"max_semantic_sub_conditions" default:"10"`
	SemanticFieldKeepRatio            float64 `json:"semantic_field_keep_ratio" default:"0.2"`
	SemanticFieldKeepMin              int     `json:"semantic_field_keep_min" default:"5"`
	SemanticFieldKeepMax              int     `json:"semantic_field_keep_max" default:"15"`
	SemanticFieldRerankBatchSize      int     `json:"semantic_field_rerank_batch_size" default:"128"`
	MinDirectRelevance                float64 `json:"min_direct_relevance" default:"0.3"`
	EnableGlobalFinalScoreRatioFilter *bool   `json:"enable_global_final_score_ratio_filter" default:"true"`
	GlobalFinalScoreRatio             float64 `json:"global_final_score_ratio" default:"0.25"`
	ExactNameMatchScore               float64 `json:"exact_name_match_score" default:"0.85"`
}

// KnSearchPropertyFilterConfig 实例属性过滤配置
type KnSearchPropertyFilterConfig struct {
	MaxPropertiesPerInstance int   `json:"max_properties_per_instance" default:"20"`
	MaxPropertyValueLength   int   `json:"max_property_value_length" default:"500"`
	EnablePropertyFilter     *bool `json:"enable_property_filter" default:"true"`
}

// ==================== Response Structures ====================

// KnSearchLocalResponse 知识网络检索本地响应
type KnSearchLocalResponse struct {
	ObjectTypes   []*KnSearchObjectType   `json:"object_types,omitempty"`
	RelationTypes []*KnSearchRelationType `json:"relation_types,omitempty"`
	ActionTypes   []*KnSearchActionType   `json:"action_types,omitempty"`
	Nodes         []*KnSearchNode         `json:"nodes,omitempty"`
	Message       string                  `json:"message,omitempty"`
}

// KnSearchObjectType object type (local response shape)
type KnSearchObjectType struct {
	ConceptType     string                   `json:"concept_type,omitempty"`
	ConceptID       string                   `json:"concept_id"`
	ConceptName     string                   `json:"concept_name"`
	Comment         string                   `json:"comment,omitempty"`
	Tags            []string                 `json:"tags,omitempty"`
	DataSource      *ResourceInfo            `json:"data_source,omitempty"`
	DataProperties  []*KnSearchDataProperty  `json:"data_properties,omitempty"`
	LogicProperties []*KnSearchLogicProperty `json:"logic_properties,omitempty"`
	PrimaryKeys     []string                 `json:"primary_keys,omitempty"`
	SampleData      map[string]any           `json:"sample_data,omitempty"`
}

// KnSearchDataProperty data property (local response shape)
type KnSearchDataProperty struct {
	Name                string            `json:"name,omitempty"`
	Comment             string            `json:"comment,omitempty"`
	Type                string            `json:"type,omitempty"`
	ConditionOperations []KnOperationType `json:"condition_operations,omitempty"`
}

// KnSearchLogicProperty logic property (local response shape)
type KnSearchLogicProperty struct {
	Name       string              `json:"name,omitempty"`
	Comment    string              `json:"comment,omitempty"`
	Type       string              `json:"type,omitempty"`
	DataSource map[string]any      `json:"data_source,omitempty"`
	Parameters []PropertyParameter `json:"parameters,omitempty"`
}

// KnSearchRelationType relation type (local response shape)
type KnSearchRelationType struct {
	ConceptType        string `json:"concept_type,omitempty"`
	ConceptID          string `json:"concept_id"`
	ConceptName        string `json:"concept_name"`
	Comment            string `json:"comment,omitempty"`
	SourceObjectTypeID string `json:"source_object_type_id"`
	TargetObjectTypeID string `json:"target_object_type_id"`
}

// KnSearchActionType action type (local response shape)
type KnSearchActionType struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ObjectTypeID   string   `json:"object_type_id"`
	ObjectTypeName string   `json:"object_type_name"`
	Comment        string   `json:"comment,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	KnID           string   `json:"kn_id"`
}

// KnSearchNode node (semantic instance) in local response
type KnSearchNode struct {
	ObjectTypeID     string         `json:"object_type_id"`
	ObjectTypeName   string         `json:"object_type_name,omitempty"`
	InstanceName     string         `json:"instance_name,omitempty"`
	UniqueIdentities map[string]any `json:"unique_identities,omitempty"`
	Properties       map[string]any `json:"properties,omitempty"`
	Score            float64        `json:"score,omitempty"`
}

// ==================== Internal Structures ====================

// KnSearchConceptResult concept retrieval result (internal)
type KnSearchConceptResult struct {
	ObjectTypes   []*KnSearchObjectType
	RelationTypes []*KnSearchRelationType
	ActionTypes   []*KnSearchActionType
}

// KnSearchSemanticInstanceResult semantic instance retrieval result (internal)
type KnSearchSemanticInstanceResult struct {
	Nodes   []*KnSearchNode
	Message string
}

// ==================== Service Interface ====================

// IKnSearchLocalService kn_search local service interface
type IKnSearchLocalService interface {
	// Search Knowledge network retrieval (local implementation)
	Search(ctx context.Context, req *KnSearchLocalRequest) (*KnSearchLocalResponse, error)
}
