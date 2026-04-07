// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// KnConceptType Knowledge Network Concept Type
type KnConceptType string

const (
	KnConceptTypeObject   KnConceptType = "object_type"   // Object Type
	KnConceptTypeRelation KnConceptType = "relation_type" // Relation Type
	KnConceptTypeAction   KnConceptType = "action_type"   // Action Type
)

// QueryObjectInstancesReq Request object for querying detailed object instances
type QueryObjectInstancesReq struct {
	KnID               string       `form:"kn_id"`                          // Knowledge Network ID
	OtID               string       `form:"ot_id"`                          // Object Type ID
	IncludeTypeInfo    bool         `form:"include_type_info"`              // Whether to include object type info
	IncludeLogicParams bool         `form:"include_logic_params"`           // Include calculation parameters for logic properties, default false
	Cond               *KnCondition `json:"condition"`                      // Retrieval conditions
	Limit              int          `json:"limit" validate:"min=1,max=10000" default:"10"` // Quantity limit, default 10, range 1-10000
	Properties         []string     `json:"properties"`                     // 指定返回的对象属性字段列表，默认返回所有属性
}

type QueryObjectInstancesResp struct {
	Data          []any          `json:"datas"`                 // List of object instances
	ObjectConcept map[string]any `json:"object_type,omitempty"` // Object type definition，由 req.include_type_info 控制是否返回
}

// QueryLogicPropertiesReq Request for querying logic properties values
type QueryLogicPropertiesReq struct {
	KnID               string                   `json:"kn_id"`
	OtID               string                   `json:"ot_id"`
	InstanceIdentities []map[string]interface{} `json:"_instance_identities"`
	Properties         []string                 `json:"properties"`
	DynamicParams      map[string]interface{}   `json:"dynamic_params"`
}

// QueryLogicPropertiesResp Response for querying logic properties values
type QueryLogicPropertiesResp struct {
	Datas []map[string]interface{} `json:"datas"`
}

// QueryInstanceSubgraphReq Subgraph query request
type QueryInstanceSubgraphReq struct {
	// Path parameters
	KnID string `form:"kn_id"`

	// Query parameters
	IncludeLogicParams bool `form:"include_logic_params"`

	// Body parameters - use interface{} to avoid explicit struct definition
	// Corresponds to SubGraphQueryBaseOnTypePath struct in ontology-query interface
	RelationTypePaths interface{} `json:"relation_type_paths"`
}

// QueryInstanceSubgraphResp Subgraph query response
type QueryInstanceSubgraphResp struct {
	// Use interface{} to directly return the original structure from the underlying interface
	// Corresponds to PathEntries struct in ontology-query interface
	Entries interface{} `json:"entries"`
}

// DrivenOntologyQuery Ontology query interface
type DrivenOntologyQuery interface {
	// QueryObjectInstances retrieves detailed data of objects for a specified object class
	QueryObjectInstances(ctx context.Context, req *QueryObjectInstancesReq) (resp *QueryObjectInstancesResp, err error)
	// QueryLogicProperties queries logic property values
	QueryLogicProperties(ctx context.Context, req *QueryLogicPropertiesReq) (resp *QueryLogicPropertiesResp, err error)
	// QueryActions queries actions
	QueryActions(ctx context.Context, req *QueryActionsRequest) (resp *QueryActionsResponse, err error)
	// QueryInstanceSubgraph queries object subgraph
	QueryInstanceSubgraph(ctx context.Context, req *QueryInstanceSubgraphReq) (resp *QueryInstanceSubgraphResp, err error)
}
