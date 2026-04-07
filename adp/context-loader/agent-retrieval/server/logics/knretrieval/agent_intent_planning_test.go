// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knretrieval

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// TestDeduplicateConcepts 测试 deduplicateConcepts 函数
func TestDeduplicateConcepts(t *testing.T) {
	convey.Convey("TestDeduplicateConcepts", t, func() {
		service := &knRetrievalServiceImpl{}

		convey.Convey("去重重复概念", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "obj-001"},
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "obj-001"}, // 重复
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "obj-002"},
				{ConceptType: interfaces.KnConceptTypeRelation, ConceptID: "rel-001"},
				{ConceptType: interfaces.KnConceptTypeRelation, ConceptID: "rel-001"}, // 重复
			}

			result := service.deduplicateConcepts(concepts)
			convey.So(len(result), convey.ShouldEqual, 3)
		})

		convey.Convey("相同 ID 不同类型不去重", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "id-001"},
				{ConceptType: interfaces.KnConceptTypeAction, ConceptID: "id-001"},
			}

			result := service.deduplicateConcepts(concepts)
			convey.So(len(result), convey.ShouldEqual, 2)
		})

		convey.Convey("空数组", func() {
			result := service.deduplicateConcepts([]*interfaces.ConceptResult{})
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("无重复", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "obj-001"},
				{ConceptType: interfaces.KnConceptTypeObject, ConceptID: "obj-002"},
			}

			result := service.deduplicateConcepts(concepts)
			convey.So(len(result), convey.ShouldEqual, 2)
		})
	})
}

// TestFilterQueryStrategysBySearchScope 测试 filterQueryStrategysBySearchScope 函数
func TestFilterQueryStrategysBySearchScope(t *testing.T) {
	convey.Convey("TestFilterQueryStrategysBySearchScope", t, func() {
		service := &knRetrievalServiceImpl{}

		convey.Convey("包含所有概念类型", func() {
			includeAll := true
			searchScope := &interfaces.SearchScopeConfig{
				IncludeObjectTypes:   &includeAll,
				IncludeRelationTypes: &includeAll,
				IncludeActionTypes:   &includeAll,
			}

			strategies := []*interfaces.SemanticQueryStrategy{
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeObject}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeRelation}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeAction}},
			}

			result := service.filterQueryStrategysBySearchScope(strategies, searchScope)
			convey.So(len(result), convey.ShouldEqual, 3)
		})

		convey.Convey("排除对象类型", func() {
			includeTrue := true
			includeFalse := false
			searchScope := &interfaces.SearchScopeConfig{
				IncludeObjectTypes:   &includeFalse,
				IncludeRelationTypes: &includeTrue,
				IncludeActionTypes:   &includeTrue,
			}

			strategies := []*interfaces.SemanticQueryStrategy{
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeObject}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeRelation}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeAction}},
			}

			result := service.filterQueryStrategysBySearchScope(strategies, searchScope)
			convey.So(len(result), convey.ShouldEqual, 2)
			convey.So(result[0].Filter.ConceptType, convey.ShouldEqual, interfaces.KnConceptTypeRelation)
		})

		convey.Convey("仅包含行动类型", func() {
			includeTrue := true
			includeFalse := false
			searchScope := &interfaces.SearchScopeConfig{
				IncludeObjectTypes:   &includeFalse,
				IncludeRelationTypes: &includeFalse,
				IncludeActionTypes:   &includeTrue,
			}

			strategies := []*interfaces.SemanticQueryStrategy{
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeObject}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeRelation}},
				{Filter: &interfaces.QueryStrategyFilter{ConceptType: interfaces.KnConceptTypeAction}},
			}

			result := service.filterQueryStrategysBySearchScope(strategies, searchScope)
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result[0].Filter.ConceptType, convey.ShouldEqual, interfaces.KnConceptTypeAction)
		})

		convey.Convey("策略无 Filter", func() {
			includeTrue := true
			searchScope := &interfaces.SearchScopeConfig{
				IncludeObjectTypes:   &includeTrue,
				IncludeRelationTypes: &includeTrue,
				IncludeActionTypes:   &includeTrue,
			}

			strategies := []*interfaces.SemanticQueryStrategy{
				{Filter: nil}, // 无 Filter 的策略应该被保留
			}

			result := service.filterQueryStrategysBySearchScope(strategies, searchScope)
			convey.So(len(result), convey.ShouldEqual, 1)
		})

		convey.Convey("空策略数组", func() {
			includeTrue := true
			searchScope := &interfaces.SearchScopeConfig{
				IncludeObjectTypes:   &includeTrue,
				IncludeRelationTypes: &includeTrue,
				IncludeActionTypes:   &includeTrue,
			}

			result := service.filterQueryStrategysBySearchScope([]*interfaces.SemanticQueryStrategy{}, searchScope)
			convey.So(len(result), convey.ShouldEqual, 0)
		})
	})
}
