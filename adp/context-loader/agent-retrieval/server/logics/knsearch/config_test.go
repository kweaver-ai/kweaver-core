package knsearch

import (
	"reflect"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

func TestDefaultConceptRetrievalConfig(t *testing.T) {
	config := DefaultConceptRetrievalConfig()
	if config == nil {
		t.Fatal("DefaultConceptRetrievalConfig returned nil")
	}

	if config.TopK != 10 {
		t.Errorf("Expected TopK 10, got %d", config.TopK)
	}
	if !boolValue(config.EnableCoarseRecall) {
		t.Error("Expected EnableCoarseRecall true")
	}
	if config.CoarseObjectLimit != 2000 {
		t.Errorf("Expected CoarseObjectLimit 2000, got %d", config.CoarseObjectLimit)
	}
}

func TestDefaultSemanticInstanceRetrievalConfig(t *testing.T) {
	config := DefaultSemanticInstanceRetrievalConfig()
	if config == nil {
		t.Fatal("DefaultSemanticInstanceRetrievalConfig returned nil")
	}

	if config.InitialCandidateCount != 50 {
		t.Errorf("Expected InitialCandidateCount 50, got %d", config.InitialCandidateCount)
	}
	if config.GlobalFinalScoreRatio != 0.25 {
		t.Errorf("Expected GlobalFinalScoreRatio 0.25, got %f", config.GlobalFinalScoreRatio)
	}
}

func TestDefaultPropertyFilterConfig(t *testing.T) {
	config := DefaultPropertyFilterConfig()
	if config == nil {
		t.Fatal("DefaultPropertyFilterConfig returned nil")
	}

	if config.MaxPropertiesPerInstance != 20 {
		t.Errorf("Expected MaxPropertiesPerInstance 20, got %d", config.MaxPropertiesPerInstance)
	}
	if !boolValue(config.EnablePropertyFilter) {
		t.Error("Expected EnablePropertyFilter true")
	}
}

func TestMergeRetrievalConfig(t *testing.T) {
	tests := []struct {
		name       string
		userConfig *interfaces.KnSearchRetrievalConfig
		check      func(*testing.T, *interfaces.KnSearchRetrievalConfig)
	}{
		{
			name:       "nil user config",
			userConfig: nil,
			check: func(t *testing.T, result *interfaces.KnSearchRetrievalConfig) {
				if result.ConceptRetrieval.TopK != 10 {
					t.Errorf("Expected default TopK 10, got %d", result.ConceptRetrieval.TopK)
				}
			},
		},
		{
			name: "merge concept config",
			userConfig: &interfaces.KnSearchRetrievalConfig{
				ConceptRetrieval: &interfaces.KnSearchConceptRetrievalConfig{
					TopK:               20,
					EnableCoarseRecall: boolPtr(false),
				},
			},
			check: func(t *testing.T, result *interfaces.KnSearchRetrievalConfig) {
				if result.ConceptRetrieval.TopK != 20 {
					t.Errorf("Expected TopK 20, got %d", result.ConceptRetrieval.TopK)
				}
				if boolValue(result.ConceptRetrieval.EnableCoarseRecall) {
					t.Error("Expected EnableCoarseRecall false")
				}
				// Check default preserved
				if result.ConceptRetrieval.CoarseObjectLimit != 2000 {
					t.Errorf("Expected CoarseObjectLimit 2000, got %d", result.ConceptRetrieval.CoarseObjectLimit)
				}
			},
		},
		{
			name: "merge semantic config",
			userConfig: &interfaces.KnSearchRetrievalConfig{
				SemanticInstanceRetrieval: &interfaces.KnSearchSemanticInstanceRetrievalConfig{
					InitialCandidateCount: 100,
					MinDirectRelevance:    0.5,
				},
			},
			check: func(t *testing.T, result *interfaces.KnSearchRetrievalConfig) {
				if result.SemanticInstanceRetrieval.InitialCandidateCount != 100 {
					t.Errorf("Expected InitialCandidateCount 100, got %d", result.SemanticInstanceRetrieval.InitialCandidateCount)
				}
				if result.SemanticInstanceRetrieval.MinDirectRelevance != 0.5 {
					t.Errorf("Expected MinDirectRelevance 0.5, got %f", result.SemanticInstanceRetrieval.MinDirectRelevance)
				}
			},
		},
		{
			name: "merge property config",
			userConfig: &interfaces.KnSearchRetrievalConfig{
				PropertyFilter: &interfaces.KnSearchPropertyFilterConfig{
					MaxPropertiesPerInstance: 50,
					EnablePropertyFilter:     boolPtr(false),
				},
			},
			check: func(t *testing.T, result *interfaces.KnSearchRetrievalConfig) {
				if result.PropertyFilter.MaxPropertiesPerInstance != 50 {
					t.Errorf("Expected MaxPropertiesPerInstance 50, got %d", result.PropertyFilter.MaxPropertiesPerInstance)
				}
				if boolValue(result.PropertyFilter.EnablePropertyFilter) {
					t.Error("Expected EnablePropertyFilter false")
				}
			},
		},
		{
			name: "zero values do not override",
			userConfig: &interfaces.KnSearchRetrievalConfig{
				ConceptRetrieval: &interfaces.KnSearchConceptRetrievalConfig{
					TopK: 0, // Should not override default
				},
			},
			check: func(t *testing.T, result *interfaces.KnSearchRetrievalConfig) {
				if result.ConceptRetrieval.TopK != 10 {
					t.Errorf("Expected default TopK 10, got %d", result.ConceptRetrieval.TopK)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeRetrievalConfig(tt.userConfig)
			tt.check(t, result)
		})
	}
}

func TestMergeHelpers(t *testing.T) {
	// Test mergeConceptRetrievalConfig directly
	baseC := DefaultConceptRetrievalConfig()
	userC := &interfaces.KnSearchConceptRetrievalConfig{
		TopK:                   5,
		IncludeSampleData:      boolPtr(true),
		SchemaBrief:            boolPtr(false),
		EnableCoarseRecall:     boolPtr(false),
		CoarseObjectLimit:      100,
		CoarseRelationLimit:    50,
		CoarseMinRelationCount: 1000,
		EnablePropertyBrief:    boolPtr(false),
		PerObjectPropertyTopK:  20,
		GlobalPropertyTopK:     100,
	}
	mergeConceptRetrievalConfig(baseC, userC)

	if baseC.TopK != 5 {
		t.Error("mergeConceptRetrievalConfig failed on basic fields")
	}
	if baseC.CoarseObjectLimit != 100 || baseC.PerObjectPropertyTopK != 20 {
		t.Error("mergeConceptRetrievalConfig failed on numeric fields")
	}

	// Test mergeSemanticInstanceRetrievalConfig
	baseS := DefaultSemanticInstanceRetrievalConfig()
	userS := &interfaces.KnSearchSemanticInstanceRetrievalConfig{
		InitialCandidateCount:             20,
		PerTypeInstanceLimit:              10,
		MaxSemanticSubConditions:          5,
		SemanticFieldKeepRatio:            0.5,
		SemanticFieldKeepMin:              2,
		SemanticFieldKeepMax:              10,
		SemanticFieldRerankBatchSize:      64,
		MinDirectRelevance:                0.6,
		EnableGlobalFinalScoreRatioFilter: boolPtr(false),
		GlobalFinalScoreRatio:             0.5,
		ExactNameMatchScore:               1.0,
	}
	mergeSemanticInstanceRetrievalConfig(baseS, userS)

	if baseS.InitialCandidateCount != 20 || baseS.MinDirectRelevance != 0.6 {
		t.Error("mergeSemanticInstanceRetrievalConfig failed")
	}
	if boolValue(baseS.EnableGlobalFinalScoreRatioFilter) != false {
		t.Error("mergeSemanticInstanceRetrievalConfig failed boolean merge")
	}

	// Test mergePropertyFilterConfig
	baseP := DefaultPropertyFilterConfig()
	userP := &interfaces.KnSearchPropertyFilterConfig{
		MaxPropertiesPerInstance: 100,
		MaxPropertyValueLength:   1000,
		EnablePropertyFilter:     boolPtr(false),
	}
	mergePropertyFilterConfig(baseP, userP)

	if baseP.MaxPropertiesPerInstance != 100 || boolValue(baseP.EnablePropertyFilter) != false {
		t.Error("mergePropertyFilterConfig failed")
	}
}

// Ensure pointers are different (deep copy not fully implemented but structure should be new)
func TestDefaultConfigsReturnNewInstances(t *testing.T) {
	c1 := DefaultConceptRetrievalConfig()
	c2 := DefaultConceptRetrievalConfig()
	if reflect.ValueOf(c1).Pointer() == reflect.ValueOf(c2).Pointer() {
		t.Error("DefaultConceptRetrievalConfig should return new instance")
	}
}
