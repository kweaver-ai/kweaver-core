// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	cond "bkn-backend/common/condition"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

func TestNewConceptSyncer(t *testing.T) {
	Convey("Test NewConceptSyncer", t, func() {
		appSetting := &common.AppSetting{}

		syncer1 := NewConceptSyncer(appSetting)
		syncer2 := NewConceptSyncer(appSetting)

		Convey("Should return singleton instance", func() {
			So(syncer1, ShouldNotBeNil)
			So(syncer2, ShouldEqual, syncer1)
		})
	})
}

func TestConceptSyncer_handleKNs(t *testing.T) {
	Convey("Test handleKNs", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		kna := bmock.NewMockKNAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			kna:        kna,
			vba:        vba,
		}

		Convey("Success with no knowledge networks", func() {
			kna.EXPECT().GetAllKNs(ctx).Return(map[string]*interfaces.KN{}, nil)
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{},
			}
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(datasetResp, nil)

			err := cs.handleKNs()
			So(err, ShouldBeNil)
		})

		Convey("Success with knowledge networks needing update", func() {
			knID := "kn1"
			branch := "main"
			kn := &interfaces.KN{
				KNID:       knID,
				KNName:     "test_kn",
				Branch:     branch,
				UpdateTime: time.Now().UnixMilli(),
			}

			ota := bmock.NewMockObjectTypeAccess(mockCtrl)
			rta := bmock.NewMockRelationTypeAccess(mockCtrl)
			ata := bmock.NewMockActionTypeAccess(mockCtrl)
			cga := bmock.NewMockConceptGroupAccess(mockCtrl)

			cs.ota = ota
			cs.rta = rta
			cs.ata = ata
			cs.cga = cga

			// handleKNs 调用顺序：
			// 1. GetAllKNs
			kna.EXPECT().GetAllKNs(ctx).Return(map[string]*interfaces.KN{knID: kn}, nil)
			// 2. getAllKNsFromDataset (内部调用 QueryResourceData)
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{},
			}
			vba.EXPECT().QueryResourceData(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(datasetResp, nil)
			// 3. handleKnowledgeNetwork 会调用多个 getAllXXXFromDatasetByKnID
			// 每个都会调用 QueryResourceData
			vba.EXPECT().QueryResourceData(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(datasetResp, nil).Times(4)

			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)

			kna.EXPECT().UpdateKNDetail(ctx, knID, branch, gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := cs.handleKNs()
			So(err, ShouldBeNil)
		})

		Convey("Failed to get knowledge networks", func() {
			kna.EXPECT().GetAllKNs(ctx).Return(nil, errors.New("db error"))

			err := cs.handleKNs()
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to get knowledge networks from dataset", func() {
			kna.EXPECT().GetAllKNs(ctx).Return(map[string]*interfaces.KN{}, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil, errors.New("dataset error"))

			err := cs.handleKNs()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleKnowledgeNetwork(t *testing.T) {
	Convey("Test handleKnowledgeNetwork", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		kna := bmock.NewMockKNAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		ota := bmock.NewMockObjectTypeAccess(mockCtrl)
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		ata := bmock.NewMockActionTypeAccess(mockCtrl)
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			kna:        kna,
			vba:        vba,
			ota:        ota,
			rta:        rta,
			ata:        ata,
			cga:        cga,
		}

		knID := "kn1"
		branch := "main"
		kn := &interfaces.KN{
			KNID:       knID,
			KNName:     "test_kn",
			Branch:     branch,
			UpdateTime: time.Now().UnixMilli(),
		}

		Convey("Success handling knowledge network", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{},
			}
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(datasetResp, nil).Times(4)

			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)

			kna.EXPECT().UpdateKNDetail(ctx, knID, branch, gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldBeNil)
		})

		Convey("No update needed", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{},
			}
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(datasetResp, nil).Times(4)

			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)

			err := cs.handleKnowledgeNetwork(ctx, kn, false)
			So(err, ShouldBeNil)
		})

		Convey("Failed to handle object types", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleObjectTypes(t *testing.T) {
	Convey("Test handleObjectTypes", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		ota := bmock.NewMockObjectTypeAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			ota:        ota,
			vba:        vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success handling object types", func() {
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(objectTypes, nil)
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{},
			}
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(datasetResp, nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			simpleItems, needUpdate, err := cs.handleObjectTypes(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(needUpdate, ShouldBeTrue)
			So(len(simpleItems), ShouldEqual, 1)
			So(simpleItems[0].OTID, ShouldEqual, "ot1")
			So(simpleItems[0].OTName, ShouldEqual, "object_type1")
		})

		Convey("Failed to get object types", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			_, _, err := cs.handleObjectTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleRelationTypes(t *testing.T) {
	Convey("Test handleRelationTypes", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			rta:        rta,
			vba:        vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success handling relation types", func() {
			relationTypes := map[string]*interfaces.RelationType{
				"rt1": {
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "relation_type1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
				},
			}

			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(relationTypes, nil)
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{},
			}
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(datasetResp, nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			simpleItems, needUpdate, err := cs.handleRelationTypes(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(needUpdate, ShouldBeTrue)
			So(len(simpleItems), ShouldEqual, 1)
			So(simpleItems[0].RTID, ShouldEqual, "rt1")
			So(simpleItems[0].SourceObjectTypeID, ShouldEqual, "ot1")
			So(simpleItems[0].TargetObjectTypeID, ShouldEqual, "ot2")
		})

		Convey("Failed to get relation types", func() {
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			_, _, err := cs.handleRelationTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleActionTypes(t *testing.T) {
	Convey("Test handleActionTypes", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		ata := bmock.NewMockActionTypeAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			ata:        ata,
			vba:        vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success handling action types", func() {
			actionTypes := map[string]*interfaces.ActionType{
				"at1": {
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:         "at1",
						ATName:       "action_type1",
						ObjectTypeID: "ot1",
					},
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(actionTypes, nil)
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{},
			}
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(datasetResp, nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			simpleItems, needUpdate, err := cs.handleActionTypes(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(needUpdate, ShouldBeTrue)
			So(len(simpleItems), ShouldEqual, 1)
			So(simpleItems[0].ATID, ShouldEqual, "at1")
			So(simpleItems[0].ObjectTypeID, ShouldEqual, "ot1")
		})

		Convey("Failed to get action types", func() {
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			_, _, err := cs.handleActionTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleConceptGroups(t *testing.T) {
	Convey("Test handleConceptGroups", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			cga:        cga,
			vba:        vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success handling concept groups", func() {
			conceptGroups := map[string]*interfaces.ConceptGroup{
				"cg1": {
					CGID:       "cg1",
					CGName:     "concept_group1",
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(conceptGroups, nil)
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{},
			}
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(datasetResp, nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			simpleItems, needUpdate, err := cs.handleConceptGroups(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(needUpdate, ShouldBeTrue)
			So(len(simpleItems), ShouldEqual, 1)
			So(simpleItems[0].CGID, ShouldEqual, "cg1")
			So(simpleItems[0].CGName, ShouldEqual, "concept_group1")
		})

		Convey("Failed to get concept groups", func() {
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			_, _, err := cs.handleConceptGroups(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertDatasetDataForKN(t *testing.T) {
	Convey("Test insertDatasetDataForKN", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			vba:        vba,
			mfa:        mfa,
		}

		kn := &interfaces.KN{
			KNID:   "kn1",
			KNName: "test_kn",
			Branch: "main",
		}

		Convey("Success inserting KN data", func() {
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := cs.insertDatasetDataForKN(ctx, kn)
			So(err, ShouldBeNil)
		})

		Convey("Failed to insert KN data", func() {
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(errors.New("opensearch error"))

			err := cs.insertDatasetDataForKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_getAllKNsFromDataset(t *testing.T) {
	Convey("Test getAllKNsFromDataset", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			vba: vba,
		}

		Convey("Success getting KNs from dataset", func() {
			entry := map[string]any{
				"kn_id":   "kn1",
				"kn_name": "test_kn",
				"branch":  "main",
			}
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{entry},
			}

			vba.EXPECT().QueryResourceData(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(datasetResp, nil)

			kns, err := cs.getAllKNsFromDataset(ctx)
			So(err, ShouldBeNil)
			So(len(kns), ShouldEqual, 1)
		})

		Convey("Failed to query KNs", func() {
			vba.EXPECT().QueryResourceData(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil, errors.New("dataset error"))

			_, err := cs.getAllKNsFromDataset(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to decode KN from entry", func() {
			entry := map[string]any{
				"name": make(chan int), // 无法解码的类型
			}
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{entry},
			}

			vba.EXPECT().QueryResourceData(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(datasetResp, nil)

			_, err := cs.getAllKNsFromDataset(ctx)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertDatasetDataForKN_WithVector(t *testing.T) {
	Convey("Test insertDatasetDataForKN with vector enabled\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: true,
			},
		}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			vba:        vba,
			mfa:        mfa,
		}

		kn := &interfaces.KN{
			KNID:   "kn1",
			KNName: "test_kn",
			Branch: "main",
			CommonInfo: interfaces.CommonInfo{
				Tags:          []string{"tag1"},
				Comment:       "comment",
				BKNRawContent: "detail",
			},
		}
		vectors := []*cond.VectorResp{
			{
				Vector: []float32{0.1, 0.2, 0.3},
			},
		}

		Convey("Success inserting KN data with vector\n", func() {
			mfa.EXPECT().GetDefaultModel(ctx).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil).AnyTimes()
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil).AnyTimes()

			err := cs.insertDatasetDataForKN(ctx, kn)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error\n", func() {
			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, errors.New("model error"))

			err := cs.insertDatasetDataForKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error\n", func() {
			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("vector error"))

			err := cs.insertDatasetDataForKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil).AnyTimes()
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(errors.New("opensearch error")).AnyTimes()

			err := cs.insertDatasetDataForKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertDatasetDataForObjectTypes(t *testing.T) {
	Convey("Test insertDatasetDataForObjectTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			vba:        vba,
		}

		Convey("Success with empty list\n", func() {
			objectTypes := []*interfaces.ObjectType{}

			err := cs.insertDatasetDataForObjectTypes(ctx, objectTypes)
			So(err, ShouldBeNil)
		})

		Convey("Success inserting object types\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := cs.insertDatasetDataForObjectTypes(ctx, objectTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(errors.New("opensearch error"))

			err := cs.insertDatasetDataForObjectTypes(ctx, objectTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertDatasetDataForObjectTypes_WithVector(t *testing.T) {
	Convey("Test insertDatasetDataForObjectTypes with vector enabled\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: true,
			},
		}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			vba:        vba,
			mfa:        mfa,
		}

		Convey("Success inserting object types with vector\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					CommonInfo: interfaces.CommonInfo{
						Tags:          []string{"tag1"},
						Comment:       "comment",
						BKNRawContent: "bkn",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{
				{
					Vector: []float32{0.1, 0.2, 0.3},
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil).AnyTimes()
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil).AnyTimes()

			err := cs.insertDatasetDataForObjectTypes(ctx, objectTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, errors.New("model error"))

			err := cs.insertDatasetDataForObjectTypes(ctx, objectTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("vector error"))

			err := cs.insertDatasetDataForObjectTypes(ctx, objectTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when vector count mismatch\n", func() {
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)

			err := cs.insertDatasetDataForObjectTypes(ctx, objectTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertDatasetDataForRelationTypes(t *testing.T) {
	Convey("Test insertDatasetDataForRelationTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			vba:        vba,
		}

		Convey("Success with empty list\n", func() {
			relationTypes := []*interfaces.RelationType{}

			err := cs.insertDatasetDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldBeNil)
		})

		Convey("Success inserting relation types\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := cs.insertDatasetDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(errors.New("opensearch error"))

			err := cs.insertDatasetDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertDatasetDataForRelationTypes_WithVector(t *testing.T) {
	Convey("Test insertDatasetDataForRelationTypes with vector enabled\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: true,
			},
		}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			vba:        vba,
			mfa:        mfa,
		}

		Convey("Success inserting relation types with vector\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					CommonInfo: interfaces.CommonInfo{
						Tags:          []string{"tag1"},
						Comment:       "comment",
						BKNRawContent: "bkn",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{
				{
					Vector: []float32{0.1, 0.2, 0.3},
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := cs.insertDatasetDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, errors.New("model error"))

			err := cs.insertDatasetDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("vector error"))

			err := cs.insertDatasetDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when vector count mismatch\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil).AnyTimes()

			err := cs.insertDatasetDataForRelationTypes(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertDatasetDataForActionTypes(t *testing.T) {
	Convey("Test insertDatasetDataForActionTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			vba:        vba,
		}

		Convey("Success with empty list\n", func() {
			actionTypes := []*interfaces.ActionType{}

			err := cs.insertDatasetDataForActionTypes(ctx, actionTypes)
			So(err, ShouldBeNil)
		})

		Convey("Success inserting action types\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := cs.insertDatasetDataForActionTypes(ctx, actionTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(errors.New("opensearch error"))

			err := cs.insertDatasetDataForActionTypes(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertDatasetDataForActionTypes_WithVector(t *testing.T) {
	Convey("Test insertDatasetDataForActionTypes with vector enabled\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: true,
			},
		}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			vba:        vba,
			mfa:        mfa,
		}

		Convey("Success inserting action types with vector\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					CommonInfo: interfaces.CommonInfo{
						Tags:          []string{"tag1"},
						Comment:       "comment",
						BKNRawContent: "bkn",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{
				{
					Vector: []float32{0.1, 0.2, 0.3},
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := cs.insertDatasetDataForActionTypes(ctx, actionTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, errors.New("model error"))

			err := cs.insertDatasetDataForActionTypes(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("vector error"))

			err := cs.insertDatasetDataForActionTypes(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when vector count mismatch\n", func() {
			actionTypes := []*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)

			err := cs.insertDatasetDataForActionTypes(ctx, actionTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertDatasetDataForConceptGroups(t *testing.T) {
	Convey("Test insertDatasetDataForConceptGroups\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			vba:        vba,
		}

		Convey("Success with empty list\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{}

			err := cs.insertDatasetDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldBeNil)
		})

		Convey("Success inserting concept groups\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := cs.insertDatasetDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldBeNil)
		})

		Convey("Failed when InsertData returns error\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(errors.New("opensearch error"))

			err := cs.insertDatasetDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_insertDatasetDataForConceptGroups_WithVector(t *testing.T) {
	Convey("Test insertDatasetDataForConceptGroups with vector enabled\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: true,
			},
		}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			vba:        vba,
			mfa:        mfa,
		}

		Convey("Success inserting concept groups with vector\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
					CommonInfo: interfaces.CommonInfo{
						Tags:          []string{"tag1"},
						Comment:       "comment",
						BKNRawContent: "bkn",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{
				{
					Vector: []float32{0.1, 0.2, 0.3},
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := cs.insertDatasetDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, errors.New("model error"))

			err := cs.insertDatasetDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("vector error"))

			err := cs.insertDatasetDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when vector count mismatch\n", func() {
			conceptGroups := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "concept_group1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			vectors := []*cond.VectorResp{}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)

			err := cs.insertDatasetDataForConceptGroups(ctx, conceptGroups)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_getAllObjectTypesFromDatasetByKnID(t *testing.T) {
	Convey("Test getAllObjectTypesFromDatasetByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			vba: vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success getting object types from dataset\n", func() {
			response := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{
					{
						"ot_id":   "ot1",
						"ot_name": "object_type1",
					},
				},
				TotalCount: 1,
			}

			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(response, nil)

			objectTypes, err := cs.getAllObjectTypesFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(len(objectTypes), ShouldEqual, 1)
		})

		Convey("Failed to search object types\n", func() {
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil, errors.New("dataset error"))

			_, err := cs.getAllObjectTypesFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to decode object type from entry\n", func() {
			response := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{
					{
						"name": make(chan int), // 无法解码的类型
					},
				},
				TotalCount: 1,
			}

			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(response, nil)

			_, err := cs.getAllObjectTypesFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_getAllRelationTypesFromDatasetByKnID(t *testing.T) {
	Convey("Test getAllRelationTypesFromDatasetByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			vba: vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success getting relation types from dataset\n", func() {
			response := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{
					{
						"rt_id":   "rt1",
						"rt_name": "relation_type1",
					},
				},
				TotalCount: 1,
			}

			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(response, nil)

			relationTypes, err := cs.getAllRelationTypesFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(len(relationTypes), ShouldEqual, 1)
		})

		Convey("Failed to search relation types\n", func() {
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil, errors.New("dataset error"))

			_, err := cs.getAllRelationTypesFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to decode relation type from entry\n", func() {
			response := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{
					{
						"name": make(chan int), // 无法解码的类型
					},
				},
				TotalCount: 1,
			}

			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(response, nil)

			_, err := cs.getAllRelationTypesFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_getAllActionTypesFromDatasetByKnID(t *testing.T) {
	Convey("Test getAllActionTypesFromDatasetByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			vba: vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success getting action types from dataset\n", func() {
			response := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{
					{
						"at_id":   "at1",
						"at_name": "action_type1",
					},
				},
				TotalCount: 1,
			}

			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(response, nil)

			actionTypes, err := cs.getAllActionTypesFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(len(actionTypes), ShouldEqual, 1)
		})

		Convey("Failed to search action types\n", func() {
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil, errors.New("dataset error"))

			_, err := cs.getAllActionTypesFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to decode action type from entry\n", func() {
			response := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{
					{
						"name": make(chan int), // 无法解码的类型
					},
				},
				TotalCount: 1,
			}

			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(response, nil)

			_, err := cs.getAllActionTypesFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_getAllConceptGroupsFromDatasetByKnID(t *testing.T) {
	Convey("Test getAllConceptGroupsFromDatasetByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			vba: vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Success getting concept groups from dataset\n", func() {
			response := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{
					{
						"cg_id":   "cg1",
						"cg_name": "concept_group1",
					},
				},
				TotalCount: 1,
			}

			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(response, nil)

			conceptGroups, err := cs.getAllConceptGroupsFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(len(conceptGroups), ShouldEqual, 1)
		})

		Convey("Failed to search concept groups\n", func() {
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil, errors.New("dataset error"))

			_, err := cs.getAllConceptGroupsFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to decode concept group from entry\n", func() {
			response := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{
					{
						"name": make(chan int), // 无法解码的类型
					},
				},
				TotalCount: 1,
			}

			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(response, nil)

			_, err := cs.getAllConceptGroupsFromDatasetByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleKnowledgeNetwork_Errors(t *testing.T) {
	Convey("Test handleKnowledgeNetwork error cases\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		kna := bmock.NewMockKNAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		ota := bmock.NewMockObjectTypeAccess(mockCtrl)
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		ata := bmock.NewMockActionTypeAccess(mockCtrl)
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			kna:        kna,
			vba:        vba,
			ota:        ota,
			rta:        rta,
			ata:        ata,
			cga:        cga,
		}

		knID := "kn1"
		branch := "main"
		kn := &interfaces.KN{
			KNID:       knID,
			KNName:     "test_kn",
			Branch:     branch,
			UpdateTime: time.Now().UnixMilli(),
		}

		Convey("Failed to handle relation types\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(&interfaces.DatasetQueryResponse{Entries: []map[string]any{}, TotalCount: 0}, nil)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to handle action types\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(&interfaces.DatasetQueryResponse{Entries: []map[string]any{}, TotalCount: 0}, nil).Times(2)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to handle concept groups\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(&interfaces.DatasetQueryResponse{Entries: []map[string]any{}, TotalCount: 0}, nil).Times(3)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(nil, errors.New("db error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to update KN detail\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(&interfaces.DatasetQueryResponse{Entries: []map[string]any{}, TotalCount: 0}, nil).Times(4)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)
			kna.EXPECT().UpdateKNDetail(ctx, knID, branch, gomock.Any()).Return(errors.New("db error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to insert dataset data for KN\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(&interfaces.DatasetQueryResponse{Entries: []map[string]any{}, TotalCount: 0}, nil).Times(4)
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)
			kna.EXPECT().UpdateKNDetail(ctx, knID, branch, gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(errors.New("dataset error"))

			err := cs.handleKnowledgeNetwork(ctx, kn, true)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleObjectTypes_Errors(t *testing.T) {
	Convey("Test handleObjectTypes error cases\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		ota := bmock.NewMockObjectTypeAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			ota:        ota,
			vba:        vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Failed to get object types from dataset\n", func() {
			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ObjectType{}, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil, errors.New("dataset error"))

			_, _, err := cs.handleObjectTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to insert dataset data\n", func() {
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object_type1",
					},
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			ota.EXPECT().GetAllObjectTypesByKnID(ctx, knID, branch).Return(objectTypes, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(&interfaces.DatasetQueryResponse{Entries: []map[string]any{}, TotalCount: 0}, nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(errors.New("dataset error"))

			_, _, err := cs.handleObjectTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleRelationTypes_Errors(t *testing.T) {
	Convey("Test handleRelationTypes error cases\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			rta:        rta,
			vba:        vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Failed to get relation types from dataset\n", func() {
			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.RelationType{}, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil, errors.New("dataset error"))

			_, _, err := cs.handleRelationTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to insert dataset data\n", func() {
			relationTypes := map[string]*interfaces.RelationType{
				"rt1": {
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "relation_type1",
					},
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			rta.EXPECT().GetAllRelationTypesByKnID(ctx, knID, branch).Return(relationTypes, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(&interfaces.DatasetQueryResponse{Entries: []map[string]any{}, TotalCount: 0}, nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(errors.New("dataset error"))

			_, _, err := cs.handleRelationTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleActionTypes_Errors(t *testing.T) {
	Convey("Test handleActionTypes error cases\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		ata := bmock.NewMockActionTypeAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			ata:        ata,
			vba:        vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Failed to get action types from dataset\n", func() {
			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(map[string]*interfaces.ActionType{}, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil, errors.New("dataset error"))

			_, _, err := cs.handleActionTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to insert dataset data\n", func() {
			actionTypes := map[string]*interfaces.ActionType{
				"at1": {
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:   "at1",
						ATName: "action_type1",
					},
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			ata.EXPECT().GetAllActionTypesByKnID(ctx, knID, branch).Return(actionTypes, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(&interfaces.DatasetQueryResponse{Entries: []map[string]any{}, TotalCount: 0}, nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(errors.New("dataset error"))

			_, _, err := cs.handleActionTypes(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestConceptSyncer_handleConceptGroups_Errors(t *testing.T) {
	Convey("Test handleConceptGroups error cases\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}

		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		cs := &ConceptSyncer{
			appSetting: appSetting,
			cga:        cga,
			vba:        vba,
		}

		knID := "kn1"
		branch := "main"

		Convey("Failed to get concept groups from dataset\n", func() {
			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(map[string]*interfaces.ConceptGroup{}, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil, errors.New("dataset error"))

			_, _, err := cs.handleConceptGroups(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to insert dataset data\n", func() {
			conceptGroups := map[string]*interfaces.ConceptGroup{
				"cg1": {
					CGID:       "cg1",
					CGName:     "concept_group1",
					UpdateTime: time.Now().UnixMilli(),
				},
			}

			cga.EXPECT().GetAllConceptGroupsByKnID(ctx, knID, branch).Return(conceptGroups, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(&interfaces.DatasetQueryResponse{Entries: []map[string]any{}, TotalCount: 0}, nil)
			vba.EXPECT().WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, gomock.Any()).Return(errors.New("dataset error"))

			_, _, err := cs.handleConceptGroups(ctx, knID, branch)
			So(err, ShouldNotBeNil)
		})
	})
}
