// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knrerank

import (
	"context"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// mockLogger 测试用的mock logger
type mockLogger struct{}

func (m *mockLogger) Debug(args ...interface{})                                  {}
func (m *mockLogger) Debugf(format string, args ...interface{})                  {}
func (m *mockLogger) Info(args ...interface{})                                   {}
func (m *mockLogger) Infof(format string, args ...interface{})                   {}
func (m *mockLogger) Warn(args ...interface{})                                   {}
func (m *mockLogger) Warnf(format string, args ...interface{})                   {}
func (m *mockLogger) Error(args ...interface{})                                  {}
func (m *mockLogger) Errorf(format string, args ...interface{})                  {}
func (m *mockLogger) WithContext(ctx context.Context) interfaces.Logger          { return m }
func (m *mockLogger) WithField(key string, value interface{}) interfaces.Logger  { return m }
func (m *mockLogger) WithFields(fields map[string]interface{}) interfaces.Logger { return m }

// mockMFModelClient 测试用的统一MF-Model-API客户端
type mockMFModelClient struct {
	chatResponse string
	chatError    error
	rerankResp   *interfaces.RerankResp
	rerankError  error
}

func (m *mockMFModelClient) Chat(ctx context.Context, req *interfaces.LLMChatReq) (string, error) {
	return m.chatResponse, m.chatError
}

func (m *mockMFModelClient) Rerank(ctx context.Context, query string, documents []string) (*interfaces.RerankResp, error) {
	return m.rerankResp, m.rerankError
}

func TestKnowledgeReranker_ParseIndices(t *testing.T) {
	reranker := &KnowledgeReranker{logger: &mockLogger{}}

	tests := []struct {
		name     string
		content  string
		expected []int
	}{
		{
			name:     "JSON数组格式",
			content:  "[1, 3, 5]",
			expected: []int{0, 2, 4}, // 转为0-based
		},
		{
			name:     "带文本的JSON数组",
			content:  "相关的概念编号是：[2, 4]",
			expected: []int{1, 3},
		},
		{
			name:     "空数组",
			content:  "[]",
			expected: []int{},
		},
		{
			name:     "无效格式降级",
			content:  "1, 3, 5",
			expected: []int{0, 2, 4},
		},
		{
			name:     "无数字",
			content:  "没有相关概念",
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indices, err := reranker.parseIndices(tt.content)
			if err != nil {
				t.Fatalf("parseIndices failed: %v", err)
			}

			if len(indices) != len(tt.expected) {
				t.Errorf("Expected %d indices, got %d", len(tt.expected), len(indices))
				return
			}

			for i, idx := range indices {
				if idx != tt.expected[i] {
					t.Errorf("Index %d: expected %d, got %d", i, tt.expected[i], idx)
				}
			}
		})
	}
}

func TestKnowledgeReranker_FormatConceptText(t *testing.T) {
	reranker := &KnowledgeReranker{logger: &mockLogger{}}

	tests := []struct {
		name     string
		concept  *interfaces.ConceptResult
		expected string
		hasError bool
	}{
		{
			name: "基本概念",
			concept: &interfaces.ConceptResult{
				ConceptType: "object_type",
				ConceptName: "销售订单",
			},
			expected: "我们有一个'销售订单'的概念。",
			hasError: false,
		},
		{
			name: "带描述的概念",
			concept: &interfaces.ConceptResult{
				ConceptType: "object_type",
				ConceptName: "商品",
				ConceptDetail: map[string]interface{}{
					"comment": "表示可销售的商品信息",
				},
			},
			expected: "我们有一个'商品'的概念，描述为表示可销售的商品信息。",
			hasError: false,
		},
		{
			name: "缺少类型",
			concept: &interfaces.ConceptResult{
				ConceptName: "测试",
			},
			hasError: true,
		},
		{
			name: "缺少名称",
			concept: &interfaces.ConceptResult{
				ConceptType: "object_type",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, err := reranker.formatConceptText(tt.concept)
			if tt.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("formatConceptText failed: %v", err)
			}

			if text != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, text)
			}
		})
	}
}

func TestKnowledgeReranker_FormatIntentsText(t *testing.T) {
	reranker := &KnowledgeReranker{logger: &mockLogger{}}

	tests := []struct {
		name     string
		intents  []interface{}
		expected string
	}{
		{
			name:     "空意图",
			intents:  []interface{}{},
			expected: "",
		},
		{
			name: "单个意图",
			intents: []interface{}{
				map[string]interface{}{
					"query_segment": "销售额",
					"confidence":    0.9,
				},
			},
			expected: "意图1是关于'销售额'，置信度为0.9。",
		},
		{
			name: "多个意图",
			intents: []interface{}{
				map[string]interface{}{
					"query_segment": "销售额",
				},
				map[string]interface{}{
					"query_segment": "利润",
				},
			},
			expected: "意图1是关于'销售额'；意图2是关于'利润'。",
		},
		{
			name: "带相关概念的意图",
			intents: []interface{}{
				map[string]interface{}{
					"query_segment": "订单",
					"related_concepts": []interface{}{
						map[string]interface{}{
							"concept_name": "销售订单",
						},
						map[string]interface{}{
							"concept_id": "order_001",
						},
					},
				},
			},
			expected: "意图1是关于'订单'，相关概念包括：销售订单, order_001。",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reranker.formatIntentsText(tt.intents)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestKnowledgeReranker_RerankByLLM_EmptyConcepts(t *testing.T) {
	mockClient := &mockMFModelClient{
		chatResponse: "[1]",
		rerankResp:   &interfaces.RerankResp{},
	}
	reranker := &KnowledgeReranker{
		logger:        &mockLogger{},
		mfModelClient: mockClient,
	}

	req := &interfaces.KnowledgeRerankReq{
		QueryUnderstanding: &interfaces.QueryUnderstanding{
			OriginQuery: "测试查询",
		},
		KnowledgeConcepts: []*interfaces.ConceptResult{},
		Action:            interfaces.KnowledgeRerankActionLLM,
	}

	results, err := reranker.Rerank(context.Background(), req)
	if err != nil {
		t.Fatalf("Rerank failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty concepts, got %d", len(results))
	}
}

func TestKnowledgeReranker_RerankByVector(t *testing.T) {
	mockClient := &mockMFModelClient{
		rerankResp: &interfaces.RerankResp{
			Results: []interfaces.RerankResult{
				{Index: 1, RelevanceScore: 0.9},
				{Index: 0, RelevanceScore: 0.7},
				{Index: 2, RelevanceScore: 0.3},
			},
		},
	}

	reranker := &KnowledgeReranker{
		logger:        &mockLogger{},
		mfModelClient: mockClient,
	}

	concepts := []*interfaces.ConceptResult{
		{ConceptType: "object_type", ConceptName: "概念A"},
		{ConceptType: "object_type", ConceptName: "概念B"},
		{ConceptType: "object_type", ConceptName: "概念C"},
	}

	req := &interfaces.KnowledgeRerankReq{
		QueryUnderstanding: &interfaces.QueryUnderstanding{
			OriginQuery: "测试查询",
		},
		KnowledgeConcepts: concepts,
		Action:            interfaces.KnowledgeRerankActionVector,
	}

	results, err := reranker.Rerank(context.Background(), req)
	if err != nil {
		t.Fatalf("Rerank failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// 验证按分数降序排序
	if results[0].RerankScore < results[1].RerankScore {
		t.Error("Results should be sorted by score descending")
	}
}
