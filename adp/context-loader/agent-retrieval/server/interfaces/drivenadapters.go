// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines interfaces
// @file drivenadapters.go
// @description: Inbound interface definition
package interfaces

//go:generate mockgen -source=drivenadapters.go -destination=../mocks/drivenadapters.go -package=mocks
import (
	"context"
)

// AccountAuthContext Account authentication context
type AccountAuthContext struct {
	// AccountID Account unique identifier
	AccountID string `json:"account_id"`
	// AccountType Account Type
	AccountType AccessorType `json:"account_type"`
	// TokenInfo Token information
	TokenInfo *TokenInfo `json:"token_info"`
}

// VisitorType Visitor Type
type VisitorType string

// Visitor type definitions
const (
	RealName  VisitorType = "realname"  // Real-name user
	Anonymous VisitorType = "anonymous" // Anonymous user
	Business  VisitorType = "business"  // Application account
)

// ToAccessorType Converts to AccessorType
func (v VisitorType) ToAccessorType() AccessorType {
	switch v {
	case RealName:
		return AccessorTypeUser
	case Business:
		return AccessorTypeApp
	case Anonymous:
		return AccessorTypeAnonymous
	default:
		// Unknown visitor type, default to anonymous user
		return AccessorTypeAnonymous
	}
}

// AccessorType Accessor Type
type AccessorType string

const (
	AccessorTypeUser       AccessorType = "user"       // Real-name user
	AccessorTypeDepartment AccessorType = "department" // Department
	AccessorTypeGroup      AccessorType = "group"      // Organization
	AccessorTypeRole       AccessorType = "role"       // Role
	AccessorTypeApp        AccessorType = "app"        // Application account
	AccessorTypeAnonymous  AccessorType = "anonymous"  // Anonymous access
)

// ToVisitorType Converts AccessorType to VisitorType
func (a AccessorType) ToVisitorType() VisitorType {
	switch a {
	case AccessorTypeUser:
		return RealName
	case AccessorTypeApp:
		return Business
	case AccessorTypeAnonymous:
		return Anonymous
	case AccessorTypeDepartment, AccessorTypeGroup, AccessorTypeRole:
		return ""
	default:
		return ""
	}
}

// AccountType Login account type
type AccountType int32

// Login account type definition
const (
	Other  AccountType = 0
	IDCard AccountType = 1
)

const (
	// AccessedByUser Real-name user
	AccessedByUser string = "accessed_by_users"
	// AccessedByAnyOne Anonymous user
	AccessedByAnyOne string = "accessed_by_anyone"
)

// ClientType Device type
type ClientType int32

// ClientTypeMap Client type map
var ClientTypeMap = map[ClientType]string{
	Unknown:      "unknown",
	IOS:          "ios",
	Android:      "android",
	WindowsPhone: "windows_phone",
	Windows:      "windows",
	MacOS:        "mac_os",
	Web:          "web",
	MobileWeb:    "mobile_web",
	Nas:          "nas",
	ConsoleWeb:   "console_web",
	DeployWeb:    "deploy_web",
	Linux:        "linux",
	APP:          "app",
}

// ReverseClientTypeMap Reverse client type map
var ReverseClientTypeMap = map[string]ClientType{
	"unknown":       Unknown,
	"ios":           IOS,
	"android":       Android,
	"windows_phone": WindowsPhone,
	"windows":       Windows,
	"mac_os":        MacOS,
	"web":           Web,
	"mobile_web":    MobileWeb,
	"nas":           Nas,
	"console_web":   ConsoleWeb,
	"deploy_web":    DeployWeb,
	"linux":         Linux,
	"app":           APP,
}

// AccountTypeMap Account type map
var AccountTypeMap = map[AccountType]string{
	Other:  "other_category",
	IDCard: "id_card",
}

// ReverseAccountTypeMap Reverse account type map
var ReverseAccountTypeMap = map[string]AccountType{
	"other_category": Other,
	"id_card":        IDCard,
}

func (typ ClientType) String() string {
	str, ok := ClientTypeMap[typ]
	if !ok {
		str = ClientTypeMap[Unknown]
	}
	return str
}

// Device type definition
const (
	Unknown ClientType = iota
	IOS
	Android
	WindowsPhone
	Windows
	MacOS
	Web
	MobileWeb
	Nas
	ConsoleWeb
	DeployWeb
	Linux
	APP
)

// TokenInfo Authorization verification information
type TokenInfo struct {
	Active     bool        // Token status
	VisitorID  string      // Visitor ID
	Scope      string      // Permission scope
	ClientID   string      // Client ID
	VisitorTyp VisitorType // Visitor type
	// Following fields exist only when visitorType=realname (real-name user)
	LoginIP     string      // Login IP
	Udid        string      // Device ID
	AccountTyp  AccountType // Account type
	ClientTyp   ClientType  // Device type
	PhoneNumber string      // Phone number for anonymous users
	VisitorName string      // Nickname for anonymous visitors
	MAC         string      // MAC address
	UserAgent   string      // User agent info
}

// Hydra Authorization service interface
type Hydra interface {
	Introspect(ctx context.Context, token string) (tokenInfo *TokenInfo, err error)
}

// KnowledgeRerankActionType Result set rerank type based on business knowledge network
type KnowledgeRerankActionType string

const (
	KnowledgeRerankActionLLM     KnowledgeRerankActionType = "llm"     // Rerank based on LLM
	KnowledgeRerankActionVector  KnowledgeRerankActionType = "vector"  // Rerank based on vector
	KnowledgeRerankActionDefault KnowledgeRerankActionType = "default" // Default rerank
)

// KnowledgeRerankReq Knowledge rerank request
type KnowledgeRerankReq struct {
	QueryUnderstanding *QueryUnderstanding       `json:"query_understanding" validate:"required"`                      // Query understanding
	KnowledgeConcepts  []*ConceptResult          `json:"concepts" validate:"required"`                                 // Business knowledge network concepts
	Action             KnowledgeRerankActionType `json:"action" validate:"required,oneof=llm vector" default:"vector"` // Action: llm based rerank, vector based rerank
}

// KnDataSourceConfig Knowledge network data source configuration
type KnDataSourceConfig struct {
	KnowledgeNetworkID string `json:"knowledge_network_id"` // Knowledge Network ID
}

// ConceptRetrievalConfig 概念召回配置
type ConceptRetrievalConfig struct {
	TopK                   int  `json:"top_k,omitempty"`                     // 默认10
	IncludeSampleData      bool `json:"include_sample_data,omitempty"`       // 默认false
	SchemaBrief            bool `json:"schema_brief,omitempty"`              // 默认false
	PerObjectPropertyTopK  int  `json:"per_object_property_top_k,omitempty"` // 默认8
	GlobalPropertyTopK     int  `json:"global_property_top_k,omitempty"`     // 默认30
	EnablePropertyBrief    bool `json:"enable_property_brief,omitempty"`     // 默认true
	EnableCoarseRecall     bool `json:"enable_coarse_recall,omitempty"`      // 默认true，启用粗召回
	CoarseObjectLimit      int  `json:"coarse_object_limit,omitempty"`       // 默认2000
	CoarseRelationLimit    int  `json:"coarse_relation_limit,omitempty"`     // 默认300
	CoarseMinRelationCount int  `json:"coarse_min_relation_count,omitempty"` // 默认5000，触发粗召回的最小关系数量
}

// PropertyFilterConfig 属性过滤配置
type PropertyFilterConfig struct {
	MaxPropertiesPerInstance int  `json:"max_properties_per_instance,omitempty"` // 默认20
	MaxPropertyValueLength   int  `json:"max_property_value_length,omitempty"`   // 默认500
	EnablePropertyFilter     bool `json:"enable_property_filter,omitempty"`      // 默认true
}

// SemanticInstanceRetrievalConfig 语义实例检索配置
type SemanticInstanceRetrievalConfig struct {
	PerTypeInstanceLimit              int     `json:"per_type_instance_limit,omitempty"`                // 默认5
	InitialCandidateCount             int     `json:"initial_candidate_count,omitempty"`                // 默认50
	EnableGlobalFinalScoreRatioFilter bool    `json:"enable_global_final_score_ratio_filter,omitempty"` // 默认true
	GlobalFinalScoreRatio             float64 `json:"global_final_score_ratio,omitempty"`               // 默认0.25
	PreFilterPerTypeLimit             int     `json:"pre_filter_per_type_limit,omitempty"`              // 可选
	MaxKeywords                       int     `json:"max_keywords,omitempty"`                           // 多关键词最大数量，默认5
	MaxSemanticSubConditions          int     `json:"max_semantic_sub_conditions,omitempty"`            // 默认10
	SemanticFieldKeepRatio            float64 `json:"semantic_field_keep_ratio,omitempty"`              // 默认0.2
	SemanticFieldKeepMin              int     `json:"semantic_field_keep_min,omitempty"`                // 默认5
	SemanticFieldKeepMax              int     `json:"semantic_field_keep_max,omitempty"`                // 默认15
	SemanticFieldRerankBatchSize      int     `json:"semantic_field_rerank_batch_size,omitempty"`       // 默认128
	MinDirectRelevance                float64 `json:"min_direct_relevance,omitempty"`                   // 默认0.3
	ExactNameMatchScore               float64 `json:"exact_name_match_score,omitempty"`                 // 默认0.85
}

// RetrievalConfig 检索配置
type RetrievalConfig struct {
	ConceptRetrieval          *ConceptRetrievalConfig          `json:"concept_retrieval,omitempty"`
	SemanticInstanceRetrieval *SemanticInstanceRetrievalConfig `json:"semantic_instance_retrieval,omitempty"`
	PropertyFilter            *PropertyFilterConfig            `json:"property_filter,omitempty"`
}

// KnSearchReq kn_search request
type KnSearchReq struct {
	// Header Parameters
	XAccountID   string `header:"x-account-id"`
	XAccountType string `header:"x-account-type"`

	// Body Parameters - use any to avoid defining complex structures explicitly
	// Corresponds to the complete request structure of data-retrieval interface
	Query           string                `json:"query" validate:"required"`
	KnID            string                `json:"kn_id" validate:"required"`
	knIDs           []*KnDataSourceConfig // Internal use, converted from KnID, not exposed
	RetrievalConfig any                   `json:"retrieval_config,omitempty"`
	OnlySchema      *bool                 `json:"only_schema,omitempty"`
	EnableRerank    *bool                 `json:"enable_rerank,omitempty"`
}

// SetKnIDs Sets knIDs (internal use, converted from KnID)
func (r *KnSearchReq) SetKnIDs(knIDs []*KnDataSourceConfig) {
	r.knIDs = knIDs
}

// GetKnIDs Gets knIDs (internal use)
func (r *KnSearchReq) GetKnIDs() []*KnDataSourceConfig {
	return r.knIDs
}

// KnSearchResp kn_search response
type KnSearchResp struct {
	// Use any to directly return the original structure from the underlying interface
	// Corresponds to the complete response structure of data-retrieval interface
	ObjectTypes   any     `json:"object_types,omitempty"`
	RelationTypes any     `json:"relation_types,omitempty"`
	ActionTypes   any     `json:"action_types,omitempty"`
	Nodes         any     `json:"nodes,omitempty"`
	Message       *string `json:"message,omitempty"`
}

// DataRetrieval Data retrieval interface
type DataRetrieval interface {
	KnowledgeRerank(ctx context.Context, req *KnowledgeRerankReq) (results []*ConceptResult, err error)
	// KnSearch Knowledge network retrieval
	KnSearch(ctx context.Context, req *KnSearchReq) (resp *KnSearchResp, err error)
}

// LLMMessage LLM对话消息
type LLMMessage struct {
	Role    string `json:"role"`    // "system" | "user" | "assistant"
	Content string `json:"content"` // 消息内容
}

// LLMChatReq LLM对话请求
type LLMChatReq struct {
	Model            string       `json:"model"`                       // 模型名称
	Messages         []LLMMessage `json:"messages"`                    // 对话消息列表
	Temperature      float64      `json:"temperature,omitempty"`       // 温度参数
	TopK             int          `json:"top_k,omitempty"`             // TopK采样
	TopP             float64      `json:"top_p,omitempty"`             // TopP采样
	FrequencyPenalty float64      `json:"frequency_penalty,omitempty"` // 频率惩罚
	PresencePenalty  float64      `json:"presence_penalty,omitempty"`  // 存在惩罚
	MaxTokens        int          `json:"max_tokens,omitempty"`        // 最大token数
	Stream           bool         `json:"stream,omitempty"`            // 是否流式
	AccountID        string       `json:"-"`                           // 账号ID（用于Header）
	AccountType      string       `json:"-"`                           // 账号类型（用于Header）
}

// DrivenMFModelAPIClient MF-Model API客户端接口
// 统一提供LLM对话和向量重排序能力
type DrivenMFModelAPIClient interface {
	// Chat 对话，返回完整响应内容
	Chat(ctx context.Context, req *LLMChatReq) (content string, err error)
	// Rerank 对文档进行重排序
	Rerank(ctx context.Context, query string, documents []string) (*RerankResp, error)
}

// RerankResult 单个重排结果
type RerankResult struct {
	Index          int     `json:"index"`           // 文档索引
	RelevanceScore float64 `json:"relevance_score"` // 相关性分数
	Document       *string `json:"document"`        // 原始文档（通常为null）
}

// RerankResp 重排响应
type RerankResp struct {
	Results []RerankResult `json:"results"` // 重排结果列表
}
