package knsearch

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

func TestConceptRetrieval_MainFlow(t *testing.T) {
	// 准备基础测试数据
	mockDetail := createMockNetworkDetail(5, 5, 2)
	baseConfig := DefaultConceptRetrievalConfig()
	baseConfig.EnableCoarseRecall = boolPtr(false)

	tests := []struct {
		name        string
		req         *interfaces.KnSearchLocalRequest
		config      *interfaces.KnSearchConceptRetrievalConfig
		mockSetup   func(*mockBknBackend, *mockRerankClient)
		checkResult func(*testing.T, *interfaces.KnSearchConceptResult, error)
	}{
		{
			name: "Success - Basic Retrieval",
			req: &interfaces.KnSearchLocalRequest{
				KnID:  "129",
				Query: "对象类型_0",
			},
			config: baseConfig,
			mockSetup: func(m *mockBknBackend, r *mockRerankClient) {
				m.networkDetail = mockDetail
			},
			checkResult: func(t *testing.T, res *interfaces.KnSearchConceptResult, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if len(res.ObjectTypes) == 0 {
					t.Error("Expected object types, got 0")
				}
				// 检查关联过滤：对象类型_0 应该被召回 (因为有关系连接)
				found := false
				for _, obj := range res.ObjectTypes {
					if obj.ConceptID == "obj_0" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected obj_0 to be found")
				}
			},
		},
		{
			name: "Failure - Network Detail Error",
			req: &interfaces.KnSearchLocalRequest{
				KnID: "129",
			},
			config: baseConfig,
			mockSetup: func(m *mockBknBackend, r *mockRerankClient) {
				m.networkError = errors.New("network error")
			},
			checkResult: func(t *testing.T, res *interfaces.KnSearchConceptResult, err error) {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			},
		},
		{
			name: "Success - With Rerank",
			req: &interfaces.KnSearchLocalRequest{
				KnID:         "129",
				Query:        "test query",
				EnableRerank: true,
			},
			config: func() *interfaces.KnSearchConceptRetrievalConfig {
				cfg := DefaultConceptRetrievalConfig()
				cfg.EnableCoarseRecall = boolPtr(false)
				return cfg
			}(),
			mockSetup: func(m *mockBknBackend, r *mockRerankClient) {
				m.networkDetail = mockDetail
				r.rerankResp = &interfaces.RerankResp{
					Results: []interfaces.RerankResult{
						{Index: 0, RelevanceScore: 0.1},
						{Index: 1, RelevanceScore: 0.9},
					},
				}
			},
			checkResult: func(t *testing.T, res *interfaces.KnSearchConceptResult, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				// 验证排序结果：rel_1 分数更高，应排在第一
				if len(res.RelationTypes) > 0 {
					if res.RelationTypes[0].ConceptID != "rel_1" {
						t.Errorf("Expected rel_1 first, got %s", res.RelationTypes[0].ConceptID)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := &mockBknBackend{}
			mockRerank := &mockRerankClient{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockManager, mockRerank)
			}

			svc := &localSearchImpl{
				logger:       &mockLogger{},
				bknBackend:   mockManager,
				rerankClient: mockRerank,
			}

			res, err := svc.conceptRetrieval(context.Background(), tt.req, tt.config)
			tt.checkResult(t, res, err)
		})
	}
}

func TestConceptRetrieval_NoRelations_ObjectTopByScore(t *testing.T) {
	detail := createMockNetworkDetail(20, 0, 0)
	for i := range detail.ObjectTypes {
		detail.ObjectTypes[i].Score = 0
	}
	detail.ObjectTypes[7].Score = 0.9
	detail.ObjectTypes[3].Score = 0.8
	detail.ObjectTypes[1].Score = 0.7

	cfg := DefaultConceptRetrievalConfig()
	cfg.EnableCoarseRecall = boolPtr(false)
	cfg.TopK = 5

	mockManager := &mockBknBackend{networkDetail: detail}
	svc := &localSearchImpl{
		logger:     &mockLogger{},
		bknBackend: mockManager,
	}

	req := &interfaces.KnSearchLocalRequest{KnID: "129", Query: "q"}
	res, err := svc.conceptRetrieval(context.Background(), req, cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(res.RelationTypes) != 0 {
		t.Fatalf("Expected 0 relations, got %d", len(res.RelationTypes))
	}
	if len(res.ObjectTypes) != 10 {
		t.Fatalf("Expected 10 objects (top_k*2), got %d", len(res.ObjectTypes))
	}
	if res.ObjectTypes[0].ConceptID != "obj_7" {
		t.Fatalf("Expected obj_7 first by score, got %s", res.ObjectTypes[0].ConceptID)
	}
	if res.ObjectTypes[1].ConceptID != "obj_3" {
		t.Fatalf("Expected obj_3 second by score, got %s", res.ObjectTypes[1].ConceptID)
	}
	if res.ObjectTypes[2].ConceptID != "obj_1" {
		t.Fatalf("Expected obj_1 third by score, got %s", res.ObjectTypes[2].ConceptID)
	}
}

func TestConceptRetrieval_ObjectFallback_FillByScore(t *testing.T) {
	detail := createMockNetworkDetail(20, 2, 0)
	for i := range detail.ObjectTypes {
		detail.ObjectTypes[i].Score = 0
	}
	detail.ObjectTypes[10].Score = 0.95
	detail.ObjectTypes[11].Score = 0.94
	detail.ObjectTypes[12].Score = 0.93
	detail.ObjectTypes[13].Score = 0.92

	cfg := DefaultConceptRetrievalConfig()
	cfg.EnableCoarseRecall = boolPtr(false)
	cfg.TopK = 10

	mockManager := &mockBknBackend{networkDetail: detail}
	svc := &localSearchImpl{
		logger:     &mockLogger{},
		bknBackend: mockManager,
	}

	req := &interfaces.KnSearchLocalRequest{KnID: "129", Query: "q"}
	res, err := svc.conceptRetrieval(context.Background(), req, cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(res.RelationTypes) != 2 {
		t.Fatalf("Expected 2 relations, got %d", len(res.RelationTypes))
	}
	if len(res.ObjectTypes) != 10 {
		t.Fatalf("Expected 10 objects with fallback fill, got %d", len(res.ObjectTypes))
	}
	found := map[string]bool{}
	for _, obj := range res.ObjectTypes {
		found[obj.ConceptID] = true
	}
	for _, id := range []string{"obj_10", "obj_11", "obj_12", "obj_13"} {
		if !found[id] {
			t.Fatalf("Expected %s to be included by fallback scoring", id)
		}
	}
}

func TestConceptRetrieval_CoarseRecall(t *testing.T) {
	// 创建大量关系以触发粗召回
	mockDetail := createMockNetworkDetail(10, 6000, 10)

	config := DefaultConceptRetrievalConfig()
	config.CoarseMinRelationCount = 5000 // 设定阈值

	mockManager := &mockBknBackend{
		networkDetail: mockDetail,
		// 模拟粗召回返回部分对象和关系
		objectTypesResp: &interfaces.ObjectTypeConcepts{
			Entries: []*interfaces.ObjectType{
				{ID: "obj_0"}, {ID: "obj_1"},
			},
		},
		relationTypesResp: &interfaces.RelationTypeConcepts{
			Entries: []*interfaces.RelationType{
				{ID: "rel_0"}, {ID: "rel_1"},
			},
		},
	}

	svc := &localSearchImpl{
		logger:     &mockLogger{},
		bknBackend: mockManager,
	}

	req := &interfaces.KnSearchLocalRequest{
		KnID:  "129",
		Query: "query",
	}

	res, err := svc.conceptRetrieval(context.Background(), req, config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证结果是否被过滤
	// 原始有6000个关系，粗召回模拟返回2个
	if len(res.RelationTypes) > 2 { // 考虑到可能后续还有排序截断，这里简单验证应该很少
		// 注意：rankRelationTypes 默认 TopK 是 10，所以这里最多返回 10
		// 但如果粗召回生效，实际上只有 2 个候选，所以应该返回 2
		if len(res.RelationTypes) != 2 {
			t.Errorf("Expected 2 relations after coarse recall, got %d", len(res.RelationTypes))
		}
	}
}

func TestRankRelationTypes(t *testing.T) {
	svc := &localSearchImpl{
		logger: &mockLogger{},
	}

	relations := []*interfaces.RelationType{
		{ID: "r1", Name: "Alpha"},
		{ID: "r2", Name: "Beta"},
		{ID: "r3", Name: "Gamma"},
	}

	// Case 1: Simple Match
	t.Run("Simple Match", func(t *testing.T) {
		// "Alpha" 匹配 "Alpha" 得分最高
		res := svc.rankRelationTypesBySimpleMatch("Alpha", relations, 10)
		if res[0].ID != "r1" {
			t.Errorf("Expected r1 first, got %s", res[0].ID)
		}
	})

	// Case 2: TopK Limit
	t.Run("TopK Limit", func(t *testing.T) {
		res := svc.rankRelationTypesBySimpleMatch("Alpha", relations, 1)
		if len(res) != 1 {
			t.Errorf("Expected 1 result, got %d", len(res))
		}
	})
}

func TestRerankRelationPathsAcrossNetworks(t *testing.T) {
	mockRerank := &mockRerankClient{
		rerankResp: &interfaces.RerankResp{
			Results: []interfaces.RerankResult{
				{Index: 0, RelevanceScore: 0.1},
				{Index: 1, RelevanceScore: 0.9},
				{Index: 2, RelevanceScore: 0.8},
				{Index: 3, RelevanceScore: 0.2},
			},
		},
	}

	svc := &localSearchImpl{
		logger:       &mockLogger{},
		rerankClient: mockRerank,
	}

	objectTypes := []*interfaces.ObjectType{
		{ID: "obj_1", Name: "对象1"},
		{ID: "obj_2", Name: "对象2"},
	}
	relationTypes := []*interfaces.RelationType{
		{ID: "rel_a", Name: "关系A", Comment: "注释A", SourceObjectTypeID: "obj_1", TargetObjectTypeID: "obj_2"},
		{ID: "rel_b", Name: "关系B", Comment: "注释B", SourceObjectTypeID: "obj_2", TargetObjectTypeID: "obj_1"},
		{ID: "rel_c", Name: "关系C", Comment: "注释C", SourceObjectTypeID: "obj_1", TargetObjectTypeID: "obj_2"},
		{ID: "rel_d", Name: "关系D", Comment: "注释D", SourceObjectTypeID: "obj_2", TargetObjectTypeID: "obj_1"},
	}

	t.Run("PerNetworkAndTotal", func(t *testing.T) {
		res := svc.rankRelationTypes(context.Background(), "query", objectTypes, relationTypes, 3, true)
		if len(res) != 3 {
			t.Fatalf("Expected 3 relations, got %d", len(res))
		}
		if res[0].ID != "rel_b" {
			t.Fatalf("Expected rel_b first, got %s", res[0].ID)
		}
		if res[1].ID != "rel_c" {
			t.Fatalf("Expected rel_c second, got %s", res[1].ID)
		}
	})

	t.Run("GlobalTotalLimit", func(t *testing.T) {
		res := svc.rankRelationTypes(context.Background(), "query", objectTypes, relationTypes, 1, true)
		if len(res) != 1 {
			t.Fatalf("Expected 1 relation, got %d", len(res))
		}
		if res[0].ID != "rel_b" {
			t.Fatalf("Expected rel_b first, got %s", res[0].ID)
		}
	})
}

func TestCalculateRelevanceScore(t *testing.T) {
	svc := &localSearchImpl{}

	tests := []struct {
		query    string
		name     string
		comment  string
		minScore float64
	}{
		{"Test", "Test", "", 1.0},           // Exact match
		{"Test", "Test Case", "", 0.5},      // Name contains query
		{"Case", "Test Case", "", 0.5},      // Name contains query
		{"Test", "Other", "Test desc", 0.2}, // Comment contains query
		{"XYZ", "ABC", "DEF", 0.0},          // No match
	}

	for _, tt := range tests {
		score := svc.calculateRelevanceScore(tt.query, tt.name, tt.comment)
		if score < tt.minScore {
			t.Errorf("Score for %s/%s/%s = %f, expected >= %f",
				tt.query, tt.name, tt.comment, score, tt.minScore)
		}
	}
}

func TestContainsFold(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "WORLD", true},
		{"Hello", "lo", true},
		{"Hello", "xyz", false},
		{"A", "a", true},
		{"", "a", false},
		{"a", "", true},
	}

	for _, tt := range tests {
		if got := containsFold(tt.s, tt.substr); got != tt.want {
			t.Errorf("containsFold(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}

func TestPruneProperties(t *testing.T) {
	svc := &localSearchImpl{
		logger: &mockLogger{},
	}

	// Setup objects with many properties
	obj := &interfaces.KnSearchObjectType{
		ConceptID: "obj1",
		DataProperties: []*interfaces.KnSearchDataProperty{
			{Name: "p1", Comment: "Important Property"}, // Score high for "Important"
			{Name: "p2", Comment: "Unrelated"},
			{Name: "p3", Comment: "Important Data"},
		},
		LogicProperties: []*interfaces.KnSearchLogicProperty{
			{Name: "l1", Comment: "Important Logic"},
			{Name: "l2", Comment: "Other"},
		},
	}

	objects := []*interfaces.KnSearchObjectType{obj}
	config := &interfaces.KnSearchConceptRetrievalConfig{
		PerObjectPropertyTopK: 1, // Strict limit per object
		GlobalPropertyTopK:    10,
	}

	// Query matches "Important"
	res := svc.pruneProperties(context.Background(), "Important", objects, config)

	// Should keep p1 (highest score data) and l1 (highest score logic)
	// Note: The implementation logic separates logic and data properties limit?
	// Let's check the implementation:
	// It uses perObjectDataCount and perObjectLogicCount separately?
	// Reading code: "if perObjectLogicCount... < config.PerObjectPropertyTopK"
	// So it limits BOTH data and logic properties separately to PerObjectPropertyTopK?
	// Actually looking at code:
	// It iterates all properties sorted by score.
	// If isLogic, check perObjectLogicCount. If !isLogic, check perObjectDataCount.
	// So effectively we get TopK data props AND TopK logic props.

	if len(res[0].DataProperties) != 1 {
		t.Errorf("Expected 1 data property, got %d", len(res[0].DataProperties))
	}
	if res[0].DataProperties[0].Name != "p1" {
		t.Errorf("Expected p1 (Important Property), got %s", res[0].DataProperties[0].Name)
	}

	if len(res[0].LogicProperties) != 1 {
		t.Errorf("Expected 1 logic property, got %d", len(res[0].LogicProperties))
	}
}

func TestFetchSampleData(t *testing.T) {
	mockQuery := &mockOntologyQuery{
		instancesResp: &interfaces.QueryObjectInstancesResp{
			Data: []any{
				map[string]any{"key": "value", "_score": 0.12},
			},
		},
	}

	svc := &localSearchImpl{
		logger:        &mockLogger{},
		ontologyQuery: mockQuery,
	}

	objects := []*interfaces.KnSearchObjectType{
		{ConceptID: "obj1"},
	}

	svc.fetchSampleData(context.Background(), "kn1", objects, true)

	if objects[0].SampleData == nil {
		t.Error("Sample data not populated")
	}
	if _, ok := objects[0].SampleData["_score"]; ok {
		t.Error("Expected _score to be removed in schema_brief mode")
	}
	if mockQuery.callCount != 1 {
		t.Errorf("Expected 1 query call, got %d", mockQuery.callCount)
	}
}
