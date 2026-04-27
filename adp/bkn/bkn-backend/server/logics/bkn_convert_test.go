// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"testing"

	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	. "github.com/smartystreets/goconvey/convey"

	"bkn-backend/interfaces"
)

// ── Network ──────────────────────────────────────────────────────────────────

func Test_ToADPNetWork(t *testing.T) {
	Convey("Test ToADPNetWork\n", t, func() {
		bknNet := &bknsdk.BknNetwork{
			BknNetworkFrontmatter: bknsdk.BknNetworkFrontmatter{
				ID:             "kn1",
				Name:           "Test KN",
				Tags:           []string{"tag1"},
				Branch:         "main",
				BusinessDomain: "domain1",
			},
			Description: "desc",
			RawContent:  "raw",
		}

		kn := ToADPNetWork(bknNet)

		So(kn.KNID, ShouldEqual, "kn1")
		So(kn.KNName, ShouldEqual, "Test KN")
		So(kn.Tags, ShouldResemble, []string{"tag1"})
		So(kn.Comment, ShouldEqual, "desc")
		So(kn.BKNRawContent, ShouldEqual, "raw")
		So(kn.Branch, ShouldEqual, "main")
		So(kn.BusinessDomain, ShouldEqual, "domain1")
	})
}

func Test_ToBKNNetWork(t *testing.T) {
	Convey("Test ToBKNNetWork\n", t, func() {
		kn := &interfaces.KN{
			KNID:   "kn1",
			KNName: "Test KN",
			CommonInfo: interfaces.CommonInfo{
				Tags:          []string{"tag1"},
				Comment:       "desc",
				BKNRawContent: "raw",
			},
			Branch:         "main",
			BusinessDomain: "domain1",
		}

		bknNet := ToBKNNetWork(kn)

		So(bknNet.ID, ShouldEqual, "kn1")
		So(bknNet.Name, ShouldEqual, "Test KN")
		So(bknNet.Tags, ShouldResemble, []string{"tag1"})
		So(bknNet.Description, ShouldEqual, "desc")
		So(bknNet.RawContent, ShouldEqual, "raw")
		So(bknNet.Branch, ShouldEqual, "main")
		So(bknNet.BusinessDomain, ShouldEqual, "domain1")
		So(bknNet.Type, ShouldEqual, interfaces.MODULE_TYPE_KN)
	})
}

// ── ObjectType ───────────────────────────────────────────────────────────────

func Test_ToADPObjectType(t *testing.T) {
	Convey("Test ToADPObjectType\n", t, func() {
		Convey("Minimal: no DataSource, no properties\n", func() {
			bknObj := &bknsdk.BknObjectType{
				BknObjectTypeFrontmatter: bknsdk.BknObjectTypeFrontmatter{
					ID: "ot1", Name: "OT1",
				},
			}
			adp := ToADPObjectType("kn1", "main", bknObj)

			So(adp.OTID, ShouldEqual, "ot1")
			So(adp.OTName, ShouldEqual, "OT1")
			So(adp.KNID, ShouldEqual, "kn1")
			So(adp.Branch, ShouldEqual, "main")
			So(adp.DataSource, ShouldBeNil)
			So(adp.DataProperties, ShouldBeEmpty)
			So(adp.LogicProperties, ShouldBeEmpty)
		})

		Convey("Full: DataSource, DataProperty with MappedField, LogicProperty with all fields\n", func() {
			bknObj := &bknsdk.BknObjectType{
				BknObjectTypeFrontmatter: bknsdk.BknObjectTypeFrontmatter{
					ID:   "ot1",
					Name: "OT1",
					Tags: []string{"t"},
				},
				Description:    "d",
				DataSource:     &bknsdk.ResourceInfo{ID: "ds1", Type: "mysql", Name: "DS1"},
				PrimaryKeys:    []string{"id"},
				DisplayKey:     "name",
				IncrementalKey: "ts",
				DataProperties: []*bknsdk.DataProperty{
					{Name: "dp1", DisplayName: "DP1", Type: "string", Description: "ddp", MappedField: "col1"},
					{Name: "dp2", Type: "int"},
				},
				LogicProperties: []*bknsdk.LogicProperty{
					{
						Name: "lp1", DisplayName: "LP1", Type: "metric",
						DataSource: &bknsdk.ResourceInfo{ID: "ds2", Type: "es", Name: "DS2"},
						Parameters: []bknsdk.Parameter{
							{Name: "p1", Type: "string", Source: "const", IfSystemGen: true, Description: "pdesc"},
						},
						AnalysisDims: []bknsdk.Field{
							{Name: "dim1", Type: "date", DisplayName: "Dim1"},
						},
					},
					{Name: "lp2", Type: "agg"},
				},
			}

			adp := ToADPObjectType("kn1", "main", bknObj)

			So(adp.DataSource.ID, ShouldEqual, "ds1")
			So(adp.PrimaryKeys, ShouldResemble, []string{"id"})
			So(adp.DisplayKey, ShouldEqual, "name")
			So(len(adp.DataProperties), ShouldEqual, 2)
			So(adp.DataProperties[0].MappedField.Name, ShouldEqual, "col1")
			So(adp.DataProperties[1].MappedField, ShouldBeNil)
			So(len(adp.LogicProperties), ShouldEqual, 2)
			So(adp.LogicProperties[0].DataSource.ID, ShouldEqual, "ds2")
			So(len(adp.LogicProperties[0].Parameters), ShouldEqual, 1)
			So(*adp.LogicProperties[0].Parameters[0].IfSystemGen, ShouldBeTrue)
			So(*adp.LogicProperties[0].Parameters[0].Comment, ShouldEqual, "pdesc")
			So(len(adp.LogicProperties[0].AnalysisDims), ShouldEqual, 1)
			So(adp.LogicProperties[0].AnalysisDims[0].Name, ShouldEqual, "dim1")
			So(adp.LogicProperties[1].DataSource, ShouldBeNil)
		})
	})
}

func Test_ToBKNObjectType(t *testing.T) {
	Convey("Test ToBKNObjectType\n", t, func() {
		Convey("Minimal: no DataSource, no properties\n", func() {
			adpObj := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "ot1", OTName: "OT1"},
			}
			bknObj := ToBKNObjectType(adpObj)

			So(bknObj.ID, ShouldEqual, "ot1")
			So(bknObj.Name, ShouldEqual, "OT1")
			So(bknObj.Type, ShouldEqual, interfaces.MODULE_TYPE_OBJECT_TYPE)
			So(bknObj.DataSource, ShouldBeNil)
		})

		Convey("Full: DataSource, DataProperty with/without MappedField, LogicProperty with all fields\n", func() {
			trueVal := true
			comment := "pdesc"
			adpObj := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1", OTName: "OT1",
					PrimaryKeys: []string{"id"}, DisplayKey: "name", IncrementalKey: "ts",
					DataSource: &interfaces.ResourceInfo{ID: "ds1", Type: "mysql", Name: "DS1"},
					DataProperties: []*interfaces.DataProperty{
						{Name: "dp1", MappedField: &interfaces.Field{Name: "col1"}},
						{Name: "dp2"},
					},
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:       "lp1",
							DataSource: &interfaces.ResourceInfo{ID: "ds2", Type: "es", Name: "DS2"},
							Parameters: []interfaces.Parameter{
								{Name: "p1", IfSystemGen: &trueVal, Comment: &comment},
							},
							AnalysisDims: []interfaces.Field{{Name: "dim1", Type: "date"}},
						},
						{Name: "lp2"},
					},
				},
				CommonInfo: interfaces.CommonInfo{Tags: []string{"t"}, Comment: "d"},
			}

			bknObj := ToBKNObjectType(adpObj)

			So(bknObj.DataSource.ID, ShouldEqual, "ds1")
			So(bknObj.PrimaryKeys, ShouldResemble, []string{"id"})
			So(len(bknObj.DataProperties), ShouldEqual, 2)
			So(bknObj.DataProperties[0].MappedField, ShouldEqual, "col1")
			So(bknObj.DataProperties[1].MappedField, ShouldBeEmpty)
			So(len(bknObj.LogicProperties), ShouldEqual, 2)
			So(bknObj.LogicProperties[0].DataSource.ID, ShouldEqual, "ds2")
			So(bknObj.LogicProperties[0].Parameters[0].IfSystemGen, ShouldBeTrue)
			So(bknObj.LogicProperties[0].Parameters[0].Description, ShouldEqual, "pdesc")
			So(bknObj.LogicProperties[0].AnalysisDims[0].Name, ShouldEqual, "dim1")
			So(bknObj.LogicProperties[1].DataSource, ShouldBeNil)
		})

		Convey("LogicProperty Parameters with nil Comment and IfSystemGen pointers\n", func() {
			adpObj := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					LogicProperties: []*interfaces.LogicProperty{
						{Parameters: []interfaces.Parameter{{Name: "p1", Comment: nil, IfSystemGen: nil}}},
					},
				},
			}
			bknObj := ToBKNObjectType(adpObj)
			So(bknObj.LogicProperties[0].Parameters[0].Description, ShouldBeEmpty)
			So(bknObj.LogicProperties[0].Parameters[0].IfSystemGen, ShouldBeFalse)
		})
	})
}

// ── RelationType ─────────────────────────────────────────────────────────────

func Test_ToADPRelationType(t *testing.T) {
	Convey("Test ToADPRelationType\n", t, func() {
		Convey("No MappingRules\n", func() {
			bknRel := &bknsdk.BknRelationType{
				BknRelationTypeFrontmatter: bknsdk.BknRelationTypeFrontmatter{ID: "rt1", Name: "RT1"},
				Endpoint:                   bknsdk.Endpoint{Source: "ot1", Target: "ot2", Type: "direct"},
			}
			adp := ToADPRelationType("kn1", "main", bknRel)
			So(adp.RTID, ShouldEqual, "rt1")
			So(adp.SourceObjectTypeID, ShouldEqual, "ot1")
			So(adp.MappingRules, ShouldBeNil)
		})

		Convey("DirectMappingRule\n", func() {
			bknRel := &bknsdk.BknRelationType{
				BknRelationTypeFrontmatter: bknsdk.BknRelationTypeFrontmatter{ID: "rt1"},
				Endpoint:                   bknsdk.Endpoint{Source: "ot1", Target: "ot2"},
				MappingRules: bknsdk.DirectMappingRule{
					{SourceProperty: "src_p", TargetProperty: "tgt_p"},
				},
			}
			adp := ToADPRelationType("kn1", "main", bknRel)
			mappings, ok := adp.MappingRules.([]interfaces.Mapping)
			So(ok, ShouldBeTrue)
			So(len(mappings), ShouldEqual, 1)
			So(mappings[0].SourceProp.Name, ShouldEqual, "src_p")
			So(mappings[0].TargetProp.Name, ShouldEqual, "tgt_p")
		})

		Convey("InDirectMappingRule\n", func() {
			bknRel := &bknsdk.BknRelationType{
				BknRelationTypeFrontmatter: bknsdk.BknRelationTypeFrontmatter{ID: "rt1"},
				Endpoint:                   bknsdk.Endpoint{Source: "ot1", Target: "ot2"},
				MappingRules: &bknsdk.InDirectMappingRule{
					BackingDataSource:  &bknsdk.ResourceInfo{ID: "ds1", Type: "mysql"},
					SourceMappingRules: []bknsdk.MappingRule{{SourceProperty: "sp", TargetProperty: "tp"}},
					TargetMappingRules: []bknsdk.MappingRule{{SourceProperty: "sp2", TargetProperty: "tp2"}},
				},
			}
			adp := ToADPRelationType("kn1", "main", bknRel)
			indirect, ok := adp.MappingRules.(*interfaces.InDirectMapping)
			So(ok, ShouldBeTrue)
			So(indirect.BackingDataSource.ID, ShouldEqual, "ds1")
			So(len(indirect.SourceMappingRules), ShouldEqual, 1)
			So(len(indirect.TargetMappingRules), ShouldEqual, 1)
		})
	})
}

func Test_ToBKNRelationType(t *testing.T) {
	Convey("Test ToBKNRelationType\n", t, func() {
		Convey("No MappingRules\n", func() {
			adpRel := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{RTID: "rt1", RTName: "RT1"},
			}
			bknRel := ToBKNRelationType(adpRel)
			So(bknRel.ID, ShouldEqual, "rt1")
			So(bknRel.Type, ShouldEqual, interfaces.MODULE_TYPE_RELATION_TYPE)
			So(bknRel.MappingRules, ShouldBeNil)
		})

		Convey("DirectMappingRule ([]interfaces.Mapping)\n", func() {
			adpRel := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID: "rt1",
					MappingRules: []interfaces.Mapping{
						{SourceProp: interfaces.SimpleProperty{Name: "sp"}, TargetProp: interfaces.SimpleProperty{Name: "tp"}},
					},
				},
			}
			bknRel := ToBKNRelationType(adpRel)
			direct, ok := bknRel.MappingRules.(bknsdk.DirectMappingRule)
			So(ok, ShouldBeTrue)
			So(len(direct), ShouldEqual, 1)
			So(direct[0].SourceProperty, ShouldEqual, "sp")
		})

		Convey("InDirectMappingRule (interfaces.InDirectMapping)\n", func() {
			adpRel := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID: "rt1",
					MappingRules: &interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{ID: "ds1", Type: "mysql"},
						SourceMappingRules: []interfaces.Mapping{
							{SourceProp: interfaces.SimpleProperty{Name: "sp"}, TargetProp: interfaces.SimpleProperty{Name: "tp"}},
						},
						TargetMappingRules: []interfaces.Mapping{
							{SourceProp: interfaces.SimpleProperty{Name: "sp2"}, TargetProp: interfaces.SimpleProperty{Name: "tp2"}},
						},
					},
				},
			}
			bknRel := ToBKNRelationType(adpRel)
			indirect, ok := bknRel.MappingRules.(*bknsdk.InDirectMappingRule)
			So(ok, ShouldBeTrue)
			So(indirect.BackingDataSource.ID, ShouldEqual, "ds1")
			So(len(indirect.SourceMappingRules), ShouldEqual, 1)
			So(len(indirect.TargetMappingRules), ShouldEqual, 1)
		})
	})
}

// ── ActionType ───────────────────────────────────────────────────────────────

func Test_ToADPActionType(t *testing.T) {
	Convey("Test ToADPActionType\n", t, func() {
		Convey("Minimal: no optional fields\n", func() {
			bknAct := &bknsdk.BknActionType{
				BknActionTypeFrontmatter: bknsdk.BknActionTypeFrontmatter{ID: "at1", Name: "AT1"},
			}
			adp := ToADPActionType("kn1", "main", bknAct)
			So(adp.ATID, ShouldEqual, "at1")
			So(adp.Affect, ShouldBeNil)
			So(adp.Condition, ShouldBeNil)
			So(adp.ActionSource.Type, ShouldBeEmpty)
			So(adp.Schedule.Type, ShouldBeEmpty)
		})

		Convey("Full: Affect, Condition, ActionSource, Parameters, Schedule\n", func() {
			bknAct := &bknsdk.BknActionType{
				BknActionTypeFrontmatter: bknsdk.BknActionTypeFrontmatter{
					ID: "at1", Name: "AT1", ActionType: "trigger",
				},
				BoundObject:  "ot1",
				AffectObject: &bknsdk.ActionAffect{ObjectType: "ot2", Description: "aff"},
				TriggerCondition: &bknsdk.ActionCondCfg{
					ObjectTypeID: "ot1", Field: "status", Operation: "eq", Value: "active",
				},
				ActionSource: &bknsdk.ActionSource{Type: "tool", BoxID: "box1", ToolID: "tool1"},
				Parameters:   []bknsdk.Parameter{{Name: "p1", Type: "string", IfSystemGen: true, Description: "d1"}},
				Schedule:     &bknsdk.Schedule{Type: "CRON", Expression: "0 * * * *"},
			}
			adp := ToADPActionType("kn1", "main", bknAct)

			So(adp.Affect.ObjectTypeID, ShouldEqual, "ot2")
			So(adp.Condition.ObjectTypeID, ShouldEqual, "ot1")
			So(adp.ActionSource.Type, ShouldEqual, "tool")
			So(adp.ActionSource.BoxID, ShouldEqual, "box1")
			So(len(adp.Parameters), ShouldEqual, 1)
			So(*adp.Parameters[0].IfSystemGen, ShouldBeTrue)
			So(*adp.Parameters[0].Comment, ShouldEqual, "d1")
			So(adp.Schedule.Type, ShouldEqual, "CRON")
			So(adp.Schedule.Expression, ShouldEqual, "0 * * * *")
		})
	})
}

func Test_ToBKNActionType(t *testing.T) {
	Convey("Test ToBKNActionType\n", t, func() {
		Convey("Minimal: no optional fields\n", func() {
			adpAct := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{ATID: "at1", ATName: "AT1"},
			}
			bknAct := ToBKNActionType(adpAct)
			So(bknAct.ID, ShouldEqual, "at1")
			So(bknAct.Type, ShouldEqual, interfaces.MODULE_TYPE_ACTION_TYPE)
			So(bknAct.AffectObject, ShouldBeNil)
			So(bknAct.TriggerCondition, ShouldBeNil)
			So(bknAct.ActionSource, ShouldBeNil)
			So(bknAct.Schedule, ShouldBeNil)
		})

		Convey("Full: Affect, Condition, ActionSource, Parameters, Schedule\n", func() {
			trueVal := true
			comment := "d1"
			adpAct := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID: "at1", ATName: "AT1",
					Affect: &interfaces.ActionAffect{ObjectTypeID: "ot2", Comment: "aff"},
					Condition: &interfaces.ActionCondCfg{
						ObjectTypeID: "ot1", Field: "status", Operation: "eq",
					},
					ActionSource: interfaces.ActionSource{Type: "tool", BoxID: "box1"},
					Parameters:   []interfaces.Parameter{{Name: "p1", IfSystemGen: &trueVal, Comment: &comment}},
					Schedule:     interfaces.Schedule{Type: "CRON", Expression: "0 * * * *"},
				},
			}
			bknAct := ToBKNActionType(adpAct)

			So(bknAct.AffectObject.ObjectType, ShouldEqual, "ot2")
			So(bknAct.TriggerCondition.ObjectTypeID, ShouldEqual, "ot1")
			So(bknAct.ActionSource.Type, ShouldEqual, "tool")
			So(bknAct.ActionSource.BoxID, ShouldEqual, "box1")
			So(len(bknAct.Parameters), ShouldEqual, 1)
			So(bknAct.Parameters[0].IfSystemGen, ShouldBeTrue)
			So(bknAct.Parameters[0].Description, ShouldEqual, "d1")
			So(bknAct.Schedule.Type, ShouldEqual, "CRON")
		})

		Convey("Parameters with nil Comment and IfSystemGen\n", func() {
			adpAct := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:       "at1",
					Parameters: []interfaces.Parameter{{Name: "p1", Comment: nil, IfSystemGen: nil}},
				},
			}
			bknAct := ToBKNActionType(adpAct)
			So(bknAct.Parameters[0].Description, ShouldBeEmpty)
			So(bknAct.Parameters[0].IfSystemGen, ShouldBeFalse)
		})
	})
}

// ── ConceptGroup ─────────────────────────────────────────────────────────────

func Test_ToADPConceptGroup(t *testing.T) {
	Convey("Test ToADPConceptGroup\n", t, func() {
		bknCG := &bknsdk.BknConceptGroup{
			BknConceptGroupFrontmatter: bknsdk.BknConceptGroupFrontmatter{
				ID:   "cg1",
				Name: "CG1",
				Tags: []string{"t"},
			},
			Description: "d",
		}
		adp := ToADPConceptGroup("kn1", "main", bknCG)

		So(adp.CGID, ShouldEqual, "cg1")
		So(adp.CGName, ShouldEqual, "CG1")
		So(adp.Tags, ShouldResemble, []string{"t"})
		So(adp.Comment, ShouldEqual, "d")
		So(adp.KNID, ShouldEqual, "kn1")
		So(adp.Branch, ShouldEqual, "main")
	})
}

func Test_ToBKNConceptGroup(t *testing.T) {
	Convey("Test ToBKNConceptGroup\n", t, func() {
		Convey("No ObjectTypes\n", func() {
			adpCG := &interfaces.ConceptGroup{CGID: "cg1", CGName: "CG1"}
			bknCG := ToBKNConceptGroup(adpCG)
			So(bknCG.ID, ShouldEqual, "cg1")
			So(bknCG.Type, ShouldEqual, interfaces.MODULE_TYPE_CONCEPT_GROUP)
			So(bknCG.ObjectTypes, ShouldBeEmpty)
		})

		Convey("With ObjectTypes (including nil element)\n", func() {
			adpCG := &interfaces.ConceptGroup{
				CGID:          "cg1",
				ObjectTypeIDs: []string{"ot1", "ot2"},
				ObjectTypes: []*interfaces.ObjectType{
					{ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "ot1"}},
					nil,
					{ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "ot2"}},
				},
			}
			bknCG := ToBKNConceptGroup(adpCG)
			// nil element is skipped
			So(bknCG.ObjectTypes, ShouldResemble, []string{"ot1", "ot2"})
		})
	})
}

// ── CondCfg (private, tested via ActionType) ─────────────────────────────────

func Test_condCfgConverters(t *testing.T) {
	Convey("Test toADPCondCfg / toBKNCondCfg\n", t, func() {
		Convey("nil input returns nil\n", func() {
			So(toADPActionCondCfg(nil), ShouldBeNil)
			So(toBKNActionCondCfg(nil), ShouldBeNil)
		})

		Convey("Single condition, no SubConds\n", func() {
			bknCond := &bknsdk.ActionCondCfg{
				ObjectTypeID: "ot1", Field: "f1", Operation: "eq", Value: "v1", ValueFrom: "const",
			}
			adp := toADPActionCondCfg(bknCond)
			So(adp.ObjectTypeID, ShouldEqual, "ot1")
			So(adp.Field, ShouldEqual, "f1")
			So(adp.ValueOptCfg.Value, ShouldEqual, "v1")
			So(adp.SubConds, ShouldBeEmpty)

			// round-trip
			bknBack := toBKNActionCondCfg(adp)
			So(bknBack.ObjectTypeID, ShouldEqual, "ot1")
			So(bknBack.Value, ShouldEqual, "v1")
			So(bknBack.SubConds, ShouldBeEmpty)
		})

		Convey("Nested SubConds (recursive)\n", func() {
			bknCond := &bknsdk.ActionCondCfg{
				ObjectTypeID: "root",
				SubConds: []*bknsdk.ActionCondCfg{
					{ObjectTypeID: "child1"},
					{ObjectTypeID: "child2", SubConds: []*bknsdk.ActionCondCfg{{ObjectTypeID: "grandchild"}}},
				},
			}
			adp := toADPActionCondCfg(bknCond)
			So(len(adp.SubConds), ShouldEqual, 2)
			So(adp.SubConds[0].ObjectTypeID, ShouldEqual, "child1")
			So(len(adp.SubConds[1].SubConds), ShouldEqual, 1)
			So(adp.SubConds[1].SubConds[0].ObjectTypeID, ShouldEqual, "grandchild")
		})
	})
}
