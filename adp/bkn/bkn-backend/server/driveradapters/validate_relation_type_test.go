// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

func Test_ValidateRelationType(t *testing.T) {
	Convey("Test ValidateRelationType\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid relation type\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "relation1",
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					Type:               interfaces.RELATION_TYPE_DIRECT,
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "prop1"},
							TargetProp: interfaces.SimpleProperty{Name: "prop2"},
						},
					},
				},
				CommonInfo: interfaces.CommonInfo{
					Tags:          []string{"tag1", "tag2", "tag3"},
					Comment:       "test comment",
					Icon:          "icon1",
					Color:         "color1",
					BKNRawContent: "bkn1",
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with strictMode false despite empty endpoints and mapping\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt-relaxed",
					RTName:             "relation_relaxed",
					SourceObjectTypeID: "",
					TargetObjectTypeID: "",
					Type:               "",
					MappingRules:       nil,
				},
			}
			err := ValidateRelationType(ctx, rt, false)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid ID\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "_invalid_id",
					RTName: "relation1",
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty name\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "",
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_NullParameter_Name)
		})

		Convey("Failed with invalid type\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   "invalid_type",
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with mapping rules but empty type\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   "",
					MappingRules: []map[string]any{
						{
							"source_property": map[string]string{"name": "prop1"},
							"target_property": map[string]string{"name": "prop2"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with direct mapping rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "relation1",
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					Type:               interfaces.RELATION_TYPE_DIRECT,
					MappingRules: []map[string]any{
						{
							"source_property": map[string]string{"name": "prop1"},
							"target_property": map[string]string{"name": "prop2"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldBeNil)
		})

		Convey("Failed with direct mapping rules empty source prop\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DIRECT,
					MappingRules: []map[string]any{
						{
							"source_property": map[string]string{"name": ""},
							"target_property": map[string]string{"name": "prop2"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with direct mapping rules empty target prop\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DIRECT,
					MappingRules: []map[string]any{
						{
							"source_property": map[string]string{"name": "prop1"},
							"target_property": map[string]string{"name": ""},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with direct mapping rules invalid format\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:         "rt1",
					RTName:       "relation1",
					Type:         interfaces.RELATION_TYPE_DIRECT,
					MappingRules: "invalid_format",
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with resource backing_data_source.type\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "relation1",
					Type:               interfaces.RELATION_TYPE_DATA_VIEW,
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.DATA_SOURCE_TYPE_RESOURCE,
							"id":   "res1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "prop1"},
								"target_property": map[string]string{"name": "bridge1"},
							},
						},
						"target_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "bridge1"},
								"target_property": map[string]string{"name": "prop2"},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with data_view mapping rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "relation1",
					Type:               interfaces.RELATION_TYPE_DATA_VIEW,
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "prop1"},
								"target_property": map[string]string{"name": "bridge1"},
							},
						},
						"target_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "bridge1"},
								"target_property": map[string]string{"name": "prop2"},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldBeNil)
		})

		Convey("Failed with data_view mapping rules empty backing_data_source\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty backing_data_source.type\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": "",
							"id":   "dv1",
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules invalid backing_data_source.type\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": "invalid_type",
							"id":   "dv1",
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty backing_data_source.id\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "",
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty source_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty source prop in source_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": ""},
								"target_property": map[string]string{"name": "bridge1"},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty target prop in source_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "prop1"},
								"target_property": map[string]string{"name": ""},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty target_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "prop1"},
								"target_property": map[string]string{"name": "bridge1"},
							},
						},
						"target_mapping_rules": []map[string]any{},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty source prop in target_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "prop1"},
								"target_property": map[string]string{"name": "bridge1"},
							},
						},
						"target_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": ""},
								"target_property": map[string]string{"name": "prop2"},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules empty target prop in target_mapping_rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: map[string]any{
						"backing_data_source": map[string]any{
							"type": interfaces.RELATION_TYPE_DATA_VIEW,
							"id":   "dv1",
						},
						"source_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "prop1"},
								"target_property": map[string]string{"name": "bridge1"},
							},
						},
						"target_mapping_rules": []map[string]any{
							{
								"source_property": map[string]string{"name": "bridge1"},
								"target_property": map[string]string{"name": ""},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with data_view mapping rules invalid format\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:         "rt1",
					RTName:       "relation1",
					Type:         interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: "invalid_format",
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with invalid type in validateMappingRules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation1",
					Type:   "invalid_type",
					MappingRules: []map[string]any{
						{
							"source_property": map[string]string{"name": "prop1"},
							"target_property": map[string]string{"name": "prop2"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with filtered_cross_join mapping rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt-fcj",
					RTName:             "fcj",
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					Type:               interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN,
					MappingRules: map[string]any{
						"source_condition": map[string]any{"field": "a", "operation": "==", "value": 1},
						"target_condition": map[string]any{"field": "b", "operation": "==", "value": 2},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldBeNil)
			_, ok := rt.MappingRules.(*interfaces.FilteredCrossJoinMapping)
			So(ok, ShouldBeTrue)
		})

		Convey("Success with filtered_cross_join omitting target_condition (optional)\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt-fcj2",
					RTName:             "fcj2",
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					Type:               interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN,
					MappingRules: map[string]any{
						"source_condition": map[string]any{"field": "a", "operation": "==", "value": 1},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with filtered_cross_join empty mapping (both sides unconstrained)\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt-fcj3",
					RTName:             "fcj3",
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					Type:               interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN,
					MappingRules:       map[string]any{},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with scope_binding target rules\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt-sb",
					RTName: "scope_binding",
					Type:   interfaces.RELATION_TYPE_SCOPE_BINDING,
					MappingRules: map[string]any{
						"source": map[string]any{
							"object_type_id": "skill",
						},
						"target_rules": []map[string]any{
							{"scope": "kn"},
							{"scope": "object_type", "id": "contract"},
							{"scope": "relation_type", "id": "contract_relation"},
							{"scope": "risk_type", "id": "contract_risk"},
							{"scope": "concept_group", "id": "contract_group"},
							{"scope": "action_type", "id": "contract_action"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldBeNil)
			_, ok := rt.MappingRules.(*interfaces.ScopeBindingMapping)
			So(ok, ShouldBeTrue)
		})

		Convey("Success with scope_binding clearing top-level source and target object type id\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt-sb-top-level",
					RTName:             "scope_binding_top_level",
					SourceObjectTypeID: "skill",
					TargetObjectTypeID: "contract",
					Type:               interfaces.RELATION_TYPE_SCOPE_BINDING,
					MappingRules: map[string]any{
						"source": map[string]any{
							"object_type_id": "skill",
						},
						"target_rules": []map[string]any{
							{"scope": "kn"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldBeNil)
			So(rt.SourceObjectTypeID, ShouldEqual, "")
			So(rt.TargetObjectTypeID, ShouldEqual, "")
		})

		Convey("Failed with scope_binding kn target rule carrying id\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt-sb-kn-id",
					RTName: "scope_binding_kn_id",
					Type:   interfaces.RELATION_TYPE_SCOPE_BINDING,
					MappingRules: map[string]any{
						"source": map[string]any{
							"object_type_id": "skill",
						},
						"target_rules": []map[string]any{
							{"scope": "kn", "id": "kn1"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with scope_binding object_type target rule missing id\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt-sb-object-no-id",
					RTName: "scope_binding_object_no_id",
					Type:   interfaces.RELATION_TYPE_SCOPE_BINDING,
					MappingRules: map[string]any{
						"source": map[string]any{
							"object_type_id": "skill",
						},
						"target_rules": []map[string]any{
							{"scope": "object_type"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with scope_binding slash-style target rule scope\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt-sb-slash",
					RTName: "scope_binding_slash",
					Type:   interfaces.RELATION_TYPE_SCOPE_BINDING,
					MappingRules: map[string]any{
						"source": map[string]any{
							"object_type_id": "skill",
						},
						"target_rules": []map[string]any{
							{"scope": "object_type/contract"},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with scope_binding non object_type target rule carrying condition\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt-sb-kn-condition",
					RTName: "scope_binding_kn_condition",
					Type:   interfaces.RELATION_TYPE_SCOPE_BINDING,
					MappingRules: map[string]any{
						"source": map[string]any{
							"object_type_id": "skill",
						},
						"target_rules": []map[string]any{
							{
								"scope":     "kn",
								"condition": map[string]any{"field": "contract_type", "operation": "==", "value": "audit"},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with scope_binding object_type target rule carrying condition\n", func() {
			rt := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt-sb-object-condition",
					RTName: "scope_binding_object_condition",
					Type:   interfaces.RELATION_TYPE_SCOPE_BINDING,
					MappingRules: map[string]any{
						"source": map[string]any{
							"object_type_id": "skill",
						},
						"target_rules": []map[string]any{
							{
								"scope":     "object_type",
								"id":        "contract",
								"condition": map[string]any{"field": "contract_type", "operation": "==", "value": "audit"},
							},
						},
					},
				},
			}
			err := ValidateRelationType(ctx, rt, true)
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateRelationTypes(t *testing.T) {
	Convey("Test ValidateRelationTypes\n", t, func() {
		ctx := context.Background()
		knID := "kn1"

		makeRT := func(id, name string) *interfaces.RelationType {
			return &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               id,
					RTName:             name,
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					Type:               interfaces.RELATION_TYPE_DIRECT,
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "prop1"},
							TargetProp: interfaces.SimpleProperty{Name: "prop2"},
						},
					},
				},
			}
		}

		Convey("Success with two relation types having the same name\n", func() {
			relationTypes := []*interfaces.RelationType{
				makeRT("rt1", "same_name"),
				makeRT("rt2", "same_name"),
			}
			err := ValidateRelationTypes(ctx, knID, relationTypes, true)
			So(err, ShouldBeNil)
			So(relationTypes[0].KNID, ShouldEqual, knID)
			So(relationTypes[1].KNID, ShouldEqual, knID)
		})

		Convey("Failed when two relation types have the same ID\n", func() {
			relationTypes := []*interfaces.RelationType{
				makeRT("rt1", "name1"),
				makeRT("rt1", "name2"),
			}
			err := ValidateRelationTypes(ctx, knID, relationTypes, true)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_Duplicated_IDInFile)
		})

		Convey("Success with multiple distinct relation types\n", func() {
			relationTypes := []*interfaces.RelationType{
				makeRT("rt1", "name1"),
				makeRT("rt2", "name2"),
				makeRT("rt3", "name1"),
			}
			err := ValidateRelationTypes(ctx, knID, relationTypes, true)
			So(err, ShouldBeNil)
		})
	})
}
