// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// ChatRequest chat request
type ChatRequest struct {
	AgentKey     string         `json:"agent_key"`                                // agent key
	AgentVersion string         `json:"agent_version,omitempty" default:"latest"` // agent version, default request latest version
	Stream       bool           `json:"stream"`                                   // Whether to return stream
	Query        string         `json:"query"`                                    // User question
	CustomQuerys map[string]any `json:"custom_querys"`                            // Custom input variables in input configuration
}

type Answer struct {
	Text string `json:"text"` // Answer content
}

// FinalAnswer Final Result
type FinalAnswer struct {
	Query  string  `json:"query"`  // Original Query
	Answer *Answer `json:"answer"` // Answer
}

type ChatContent struct {
	FinalAnswer *FinalAnswer `json:"final_answer"` // Final Result
}

// ChatMessage Message Content
type ChatMessage struct {
	ID             string       `json:"id"`              // Message ID
	ConversationID string       `json:"conversation_id"` // Conversation ID
	Role           string       `json:"role"`            // Message Role
	Content        *ChatContent `json:"content"`         // Message Content
}

// ChatResponse Chat Response
type ChatResponse struct {
	ConversationID     string       `json:"conversation_id"`      // Conversation ID
	UserMessageID      string       `json:"user_message_id"`      // User Message ID
	AssistantMessageID string       `json:"assistant_message_id"` // Assistant Message ID
	Message            *ChatMessage `json:"message"`              // Message Content
}

// ConceptIntentionAnalysisAgentReq Concept Intention Analysis Agent Request
type ConceptIntentionAnalysisAgentReq struct {
	PreviousQueries []string `json:"previous_queries"` // History Queries
	Query           string   `json:"query"`            // Current Query
	KnID            string   `json:"kn_id"`            // Knowledge Network ID
}

// ConceptRetrievalStrategistReq Concept Retrieval Strategist Agent Request
type ConceptRetrievalStrategistReq struct {
	QueryParam      *ConceptRetrievalStrategistQueryParam `json:"query_param"`      // Concept Retrieval Strategist Agent Query
	PreviousQueries []string                              `json:"previous_queries"` // History Queries
	KnID            string                                `json:"kn_id"`            // Knowledge Network ID
}

// ConceptRetrievalStrategistQueryParam Concept Retrieval Strategist Agent Query
type ConceptRetrievalStrategistQueryParam struct {
	OriginalQuery        string               `json:"original_query"`         // Original Query
	CurrentIntentSegment *SemanticQueryIntent `json:"current_intent_segment"` // Current Intent Segment
	ConceptCandidates    []*ConceptResult     `json:"concept_candidates"`     // Concept Candidates
}

// AgentApp Agent Interface
type AgentApp interface {
	APIChat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	// ConceptIntentionAnalysisAgent Concept Intention Analysis Agent
	ConceptIntentionAnalysisAgent(ctx context.Context, req *ConceptIntentionAnalysisAgentReq) (*QueryUnderstanding, error)
	// ConceptRetrievalStrategistAgent Concept Retrieval Strategist Agent
	ConceptRetrievalStrategistAgent(ctx context.Context, req *ConceptRetrievalStrategistReq) ([]*SemanticQueryStrategy, error)
	// MetricDynamicParamsGeneratorAgent Metric Dynamic Params Generator Agent
	MetricDynamicParamsGeneratorAgent(ctx context.Context, req *MetricDynamicParamsGeneratorReq) (dynamicParams map[string]any, missingParams *MissingPropertyParams, err error)
	// OperatorDynamicParamsGeneratorAgent Operator Dynamic Params Generator Agent
	OperatorDynamicParamsGeneratorAgent(ctx context.Context, req *OperatorDynamicParamsGeneratorReq) (dynamicParams map[string]any, missingParams *MissingPropertyParams, err error)
}

// MetricDynamicParamsGeneratorReq Metric Dynamic Params Generator Request
type MetricDynamicParamsGeneratorReq struct {
	LogicProperty     *LogicPropertyDef `json:"logic_property"`
	Query             string            `json:"query"`
	UniqueIdentities  []map[string]any  `json:"unique_identities"`
	AdditionalContext string            `json:"additional_context,omitempty"`
	NowMs             int64             `json:"now_ms,omitempty"`
	Timezone          string            `json:"timezone,omitempty"`
}

// OperatorDynamicParamsGeneratorReq Operator Dynamic Params Generator Request
type OperatorDynamicParamsGeneratorReq struct {
	OperatorID        string            `json:"operator_id"`
	LogicProperty     *LogicPropertyDef `json:"logic_property"`
	Query             string            `json:"query"`
	UniqueIdentities  []map[string]any  `json:"unique_identities"`
	AdditionalContext string            `json:"additional_context,omitempty"`
	// ObjectInstances removed, object instance information is passed via AdditionalContext
}
