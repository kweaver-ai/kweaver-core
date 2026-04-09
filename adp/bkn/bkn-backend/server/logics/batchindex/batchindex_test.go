// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package batchindex

import (
	"testing"

	"bkn-backend/interfaces"
	"bkn-backend/interfaces/data_type"
)

func TestCollectKNFromPayload_nestedConceptGroup(t *testing.T) {
	kn := &interfaces.KN{
		KNID:   "kn1",
		Branch: interfaces.MAIN_BRANCH,
		ConceptGroups: []*interfaces.ConceptGroup{
			{
				CGID: "cg1",
				ObjectTypes: []*interfaces.ObjectType{
					{ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "ot_a", OTName: "a"}},
				},
				RelationTypes: []*interfaces.RelationType{
					{RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{RTID: "rt1"}},
				},
				ActionTypes: []*interfaces.ActionType{
					{ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{ATID: "at1"}},
				},
			},
		},
		ObjectTypes: []*interfaces.ObjectType{
			{ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "ot_top"}},
		},
	}
	idx, err := CollectKNFromPayload(kn)
	if err != nil {
		t.Fatal(err)
	}
	if !HasObjectTypeID("ot_a", idx) || !HasObjectTypeID("ot_top", idx) {
		t.Fatalf("object types: %+v", idx.ObjectTypes)
	}
	if _, ok := idx.RelationTypeIDs["rt1"]; !ok {
		t.Fatal("expected rt1 in index")
	}
	if _, ok := idx.ActionTypeIDs["at1"]; !ok {
		t.Fatal("expected at1 in index")
	}
	if !HasConceptGroupID("cg1", idx) {
		t.Fatal("expected cg1 in index")
	}
}

func TestCollectKNFromPayload_duplicateObjectTypeConflict(t *testing.T) {
	kn := &interfaces.KN{
		KNID:   "kn1",
		Branch: interfaces.MAIN_BRANCH,
		ObjectTypes: []*interfaces.ObjectType{
			{ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "dup", OTName: "a"}},
			{ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "dup", OTName: "b"}},
		},
	}
	_, err := CollectKNFromPayload(kn)
	if err == nil {
		t.Fatal("expected conflicting definitions error")
	}
}

func TestEnsureObjectTypePropertyMap(t *testing.T) {
	ot := &interfaces.ObjectType{
		ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
			OTID: "o1",
			DataProperties: []*interfaces.DataProperty{
				{Name: "p1", DisplayName: "P1", Type: data_type.DATATYPE_STRING},
			},
		},
	}
	EnsureObjectTypePropertyMap(ot)
	if ot.PropertyMap["p1"] != "P1" {
		t.Fatalf("PropertyMap = %#v", ot.PropertyMap)
	}
}
