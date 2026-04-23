// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

const MaxMatchScore float64 = 100 // Maximum match score

// SemanticQueryStrategyType Semantic query strategy type
type SemanticQueryStrategyType string

const (
	ConceptGetStrategy              SemanticQueryStrategyType = "concept_get"               // Concept get
	ConceptDiscoveryStrategy        SemanticQueryStrategyType = "concept_discovery"         // Concept discovery
	ObjectInstanceDiscoveryStrategy SemanticQueryStrategyType = "object_instance_discovery" // Object instance discovery
)

// KnBaseConceptField Knowledge network base concept field
type KnBaseConceptField string

const (
	ConceptFieldID    KnBaseConceptField = "id"     // Concept ID
	ConceptFieldName  KnBaseConceptField = "name"   // Concept Name
	ConceptFieldScore KnBaseConceptField = "_score" // Match score
	ConceptFieldAny   KnBaseConceptField = "*"      // Any field
)

// SemanticQueryMode Semantic query strategy mode
type SemanticQueryMode string

const (
	AgentIntentPlanning    SemanticQueryMode = "agent_intent_planning"    // Strategy based on agent intent analysis and planning
	AgentIntentRetrieval   SemanticQueryMode = "agent_intent_retrieval"   // Intent analysis agent + traditional retrieval strategy
	KeywordVectorRetrieval SemanticQueryMode = "keyword_vector_retrieval" // Keyword + Vector retrieval
)

// SearchScopeConfig Search scope configuration
type SearchScopeConfig struct {
	ConceptGroups        []string `json:"concept_groups"`
	IncludeObjectTypes   *bool    `json:"include_object_types" default:"true"`
	IncludeRelationTypes *bool    `json:"include_relation_types" default:"true"`
	IncludeActionTypes   *bool    `json:"include_action_types" default:"true"`
}

// KnowledgeConcept Knowledge network concept definition
type KnowledgeConcept struct {
	ConceptType KnConceptType `json:"concept_type"` // Concept type
	ConceptID   string        `json:"concept_id"`   // Concept ID
	ConceptName string        `json:"concept_name"` // Concept Name
}

// SemanticQueryIntent Semantic query intent
type SemanticQueryIntent struct {
	QuerySegment      string              `json:"query_segment"`      // Query segment
	Confidence        float32             `json:"confidence"`         // Confidence
	Reasoning         string              `json:"reasoning"`          // Reasoning
	RequiresReasoning bool                `json:"requires_reasoning"` // Whether further reasoning is required
	RelatedConcepts   []*KnowledgeConcept `json:"related_concepts"`   // Related concepts
}

// QueryStrategyCondition Strategy filtering condition
type QueryStrategyCondition struct {
	Field     string `json:"field"`     // Field name
	Operation string `json:"operation"` // Operator
	Value     any    `json:"value"`     // Field value
}

// QueryStrategyFilter Query strategy filter item
type QueryStrategyFilter struct {
	ConceptType KnConceptType             `json:"concept_type"` // Concept type
	ConceptID   string                    `json:"concept_id"`   // Concept ID
	ConceptIDs  []string                  `json:"concept_ids"`  // Concept IDs
	Conditions  []*QueryStrategyCondition `json:"conditions"`   // Filtering conditions
}

// SemanticQueryStrategy Semantic query strategy
type SemanticQueryStrategy struct {
	StrategyType SemanticQueryStrategyType `json:"strategy_type"` // Strategy type
	Filter       *QueryStrategyFilter      `json:"filter"`        // Filtering conditions
}

// QueryUnderstanding Query understanding
type QueryUnderstanding struct {
	OriginQuery    string                   `json:"origin_query"`    // Original Query
	ProcessedQuery string                   `json:"processed_query"` // Processed Query
	Intent         []*SemanticQueryIntent   `json:"intent"`          // Semantic query intent
	QueryStrategys []*SemanticQueryStrategy `json:"query_strategy"`  // Semantic query strategies
}

// SemanticSearchRequest Semantic search request
type SemanticSearchRequest struct {
	// Mode is the semantic retrieval strategy mode
	Mode SemanticQueryMode `form:"mode" validate:"required,oneof=keyword_vector_retrieval agent_intent_planning agent_intent_retrieval" default:"keyword_vector_retrieval"`
	// RerankAction is the rerank action: llm based rerank, vector based rerank
	RerankAction KnowledgeRerankActionType `json:"rerank_action" validate:"required,oneof=default llm vector" default:"vector"`
	// ReturnQueryUnderstanding indicates whether to return query understanding information
	ReturnQueryUnderstanding *bool    `json:"return_query_understanding" default:"false"`
	Query                    string   `json:"query" validate:"required"` // User Query
	KnID                     string   `json:"kn_id" validate:"required"` // Knowledge network ID
	PreviousQueries          []string `json:"previous_queries"`          // History Queries
	// SearchScope is the search scope configuration
	SearchScope *SearchScopeConfig `json:"search_scope"`
	MaxConcepts int                `json:"max_concepts" default:"10"` // Max concepts count
}

// ConceptResult Concept result
type ConceptResult struct {
	ConceptType   KnConceptType `json:"concept_type"`           // Concept type
	ConceptID     string        `json:"concept_id"`             // Concept ID
	ConceptName   string        `json:"concept_name"`           // Concept Name
	ConceptDetail any           `json:"concept_detail"`         // Concept Detail
	IntentScore   float64       `json:"intent_score,omitempty"` // Intent Score
	MatchScore    float64       `json:"match_score,omitempty"`  // Match Score
	RerankScore   float64       `json:"rerank_score,omitempty"` // Rerank Score
	Samples       []any         `json:"samples,omitempty"`      // Samples
}

type SemanticSearchResponse struct {
	QueryUnderstanding *QueryUnderstanding `json:"query_understanding,omitempty" validate:"required"` // Query understanding
	KnowledgeConcepts  []*ConceptResult    `json:"concepts" validate:"required"`                      // Knowledge network concepts
	HitsTotal          int                 `json:"hits_total,omitempty"`                              // Total hits
}

// IKnRetrievalService Knowledge network based retrieval service
type IKnRetrievalService interface {
	// AgentIntentPlanning Semantic retrieval: Intent analysis agent + Planning strategy
	AgentIntentPlanning(ctx context.Context, req *SemanticSearchRequest) (*SemanticSearchResponse, error)
	// AgentIntentRetrieval Semantic retrieval: Intent analysis agent + Retrieval strategy
	AgentIntentRetrieval(ctx context.Context, req *SemanticSearchRequest) (resp *SemanticSearchResponse, err error)
	// KeywordVectorRetrieval Semantic retrieval: Keyword + Vector retrieval
	KeywordVectorRetrieval(ctx context.Context, req *SemanticSearchRequest) (resp *SemanticSearchResponse, err error)
}
