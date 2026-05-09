// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package relation_type

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	cond "bkn-backend/common/condition"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
	"bkn-backend/logics/batchindex"
)

func Test_relationTypeService_CheckRelationTypeExistByID(t *testing.T) {
	Convey("Test CheckRelationTypeExistByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)

		service := &relationTypeService{
			appSetting: appSetting,
			rta:        rta,
		}

		Convey("Success when relation type exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtID := "rt1"
			rtName := "relation_type1"

			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rtName, true, nil)

			name, exist, err := service.CheckRelationTypeExistByID(ctx, knID, branch, rtID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(name, ShouldEqual, rtName)
		})

		Convey("Success when relation type does not exist\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtID := "rt1"

			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			name, exist, err := service.CheckRelationTypeExistByID(ctx, knID, branch, rtID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtID := "rt1"

			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			name, exist, err := service.CheckRelationTypeExistByID(ctx, knID, branch, rtID)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_InternalError_CheckRelationTypeIfExistFailed)
		})
	})
}

func Test_relationTypeService_GetRelationTypeIDsByKnID(t *testing.T) {
	Convey("Test GetRelationTypeIDsByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)

		service := &relationTypeService{
			appSetting: appSetting,
			rta:        rta,
		}

		Convey("Success getting relation type IDs\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1", "rt2"}

			rta.EXPECT().GetRelationTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(rtIDs, nil)

			result, err := service.GetRelationTypeIDsByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, rtIDs)
		})

		Convey("Success with empty result\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH

			rta.EXPECT().GetRelationTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)

			result, err := service.GetRelationTypeIDsByKnID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH

			rta.EXPECT().GetRelationTypeIDsByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			result, err := service.GetRelationTypeIDsByKnID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_InternalError_GetRelationTypesByIDsFailed)
		})
	})
}

func Test_relationTypeService_GetRelationTypesByIDs(t *testing.T) {
	Convey("Test GetRelationTypesByIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		ums := bmock.NewMockUserMgmtService(mockCtrl)

		service := &relationTypeService{
			appSetting: appSetting,
			rta:        rta,
			ps:         ps,
			ots:        ots,
			ums:        ums,
		}

		Convey("Success getting relation types by IDs\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1", "rt2"}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "rt1",
					},
				},
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt2",
						RTName: "rt2",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().GetRelationTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rtArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil).AnyTimes()
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			result, err := service.GetRelationTypesByIDs(ctx, knID, branch, rtIDs)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("Failed when relation types count mismatch\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1", "rt2"}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "rt1",
						RTName: "rt1",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().GetRelationTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rtArr, nil)

			result, err := service.GetRelationTypesByIDs(ctx, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)
			So(result, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_RelationTypeNotFound)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_InternalError_CheckPermissionFailed))

			result, err := service.GetRelationTypesByIDs(ctx, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when GetRelationTypesByIDs returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().GetRelationTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			result, err := service.GetRelationTypesByIDs(ctx, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when GetObjectTypesMapByIDs returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1"}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "rt1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().GetRelationTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rtArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			result, err := service.GetRelationTypesByIDs(ctx, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Success with DIRECT type\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1"}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "rt1",
						Type:               interfaces.RELATION_TYPE_DIRECT,
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
						MappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{
									Name: "prop1",
								},
								TargetProp: interfaces.SimpleProperty{
									Name: "prop2",
								},
							},
						},
					},
				},
			}
			objectTypeMap := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
					PropertyMap: map[string]string{
						"prop1": "Property1",
					},
				},
				"ot2": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot2",
						OTName: "ot2",
					},
					PropertyMap: map[string]string{
						"prop2": "Property2",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().GetRelationTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rtArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypeMap, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			result, err := service.GetRelationTypesByIDs(ctx, knID, branch, rtIDs)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(result[0].SourceObjectType.OTID, ShouldEqual, "ot1")
			So(result[0].TargetObjectType.OTID, ShouldEqual, "ot2")
		})

		Convey("Success with DATA_VIEW type\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1"}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "rt1",
						Type:               interfaces.RELATION_TYPE_DATA_VIEW,
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
						MappingRules: &interfaces.InDirectMapping{
							BackingDataSource: &interfaces.ResourceInfo{
								ID: "dv1",
							},
							SourceMappingRules: []interfaces.Mapping{
								{
									SourceProp: interfaces.SimpleProperty{
										Name: "prop1",
									},
									TargetProp: interfaces.SimpleProperty{
										Name: "field1",
									},
								},
							},
							TargetMappingRules: []interfaces.Mapping{
								{
									SourceProp: interfaces.SimpleProperty{
										Name: "field2",
									},
									TargetProp: interfaces.SimpleProperty{
										Name: "prop2",
									},
								},
							},
						},
					},
				},
			}
			objectTypeMap := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
					PropertyMap: map[string]string{
						"prop1": "Property1",
					},
				},
				"ot2": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot2",
						OTName: "ot2",
					},
					PropertyMap: map[string]string{
						"prop2": "Property2",
					},
				},
			}
			dataView := &interfaces.DataView{
				ViewName: "data_view1",
				FieldsMap: map[string]*interfaces.ViewField{
					"field1": {
						DisplayName: "Field1",
					},
					"field2": {
						DisplayName: "Field2",
					},
				},
			}
			dva := bmock.NewMockDataViewAccess(mockCtrl)

			service := &relationTypeService{
				appSetting: appSetting,
				rta:        rta,
				ps:         ps,
				ots:        ots,
				dva:        dva,
				ums:        ums,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().GetRelationTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rtArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypeMap, nil)
			dva.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any()).Return(dataView, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			result, err := service.GetRelationTypesByIDs(ctx, knID, branch, rtIDs)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(result[0].MappingRules.(*interfaces.InDirectMapping).BackingDataSource.Name, ShouldEqual, "data_view1")
		})

		Convey("Failed when GetDataViewByID returns error for DATA_VIEW type\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1"}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "rt1",
						Type:               interfaces.RELATION_TYPE_DATA_VIEW,
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
						MappingRules: &interfaces.InDirectMapping{
							BackingDataSource: &interfaces.ResourceInfo{
								ID: "dv1",
							},
						},
					},
				},
			}
			objectTypeMap := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
				"ot2": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot2",
						OTName: "ot2",
					},
				},
			}
			dva := bmock.NewMockDataViewAccess(mockCtrl)

			service := &relationTypeService{
				appSetting: appSetting,
				rta:        rta,
				ps:         ps,
				ots:        ots,
				dva:        dva,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().GetRelationTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rtArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypeMap, nil)
			dva.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			result, err := service.GetRelationTypesByIDs(ctx, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})
	})
}

func Test_relationTypeService_ListRelationTypes(t *testing.T) {
	Convey("Test ListRelationTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		ums := bmock.NewMockUserMgmtService(mockCtrl)

		service := &relationTypeService{
			appSetting: appSetting,
			rta:        rta,
			ps:         ps,
			ots:        ots,
			ums:        ums,
		}

		Convey("Success listing relation types\n", func() {
			query := interfaces.RelationTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "rt1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return(rtArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			rts, total, err := service.ListRelationTypes(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(rts), ShouldEqual, 1)
		})

		Convey("Success with empty result\n", func() {
			query := interfaces.RelationTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, nil)

			rts, total, err := service.ListRelationTypes(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(rts), ShouldEqual, 0)
		})

		Convey("Failed when permission check fails\n", func() {
			query := interfaces.RelationTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_InternalError_CheckPermissionFailed))

			rts, total, err := service.ListRelationTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(rts), ShouldEqual, 0)
		})

		Convey("Failed when ListRelationTypes returns error\n", func() {
			query := interfaces.RelationTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			rts, total, err := service.ListRelationTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(rts), ShouldEqual, 0)
		})

		Convey("Failed when GetObjectTypesMapByIDs returns error\n", func() {
			query := interfaces.RelationTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "rt1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return(rtArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			rts, total, err := service.ListRelationTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(rts), ShouldEqual, 0)
		})

		Convey("Failed when GetAccountNames returns error\n", func() {
			query := interfaces.RelationTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "rt1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return(rtArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			rts, total, err := service.ListRelationTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(rts), ShouldEqual, 0)
		})

		Convey("Success with Limit = -1\n", func() {
			query := interfaces.RelationTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  -1,
					Offset: 0,
				},
			}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "rt1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return(rtArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			rts, total, err := service.ListRelationTypes(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(rts), ShouldEqual, 1)
		})

		Convey("Success with Offset out of bounds\n", func() {
			query := interfaces.RelationTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 100,
				},
			}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "rt1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return(rtArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			rts, total, err := service.ListRelationTypes(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(rts), ShouldEqual, 0)
		})

		Convey("Success with pagination\n", func() {
			query := interfaces.RelationTypesQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  2,
					Offset: 1,
				},
			}
			rtArr := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "rt1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
				},
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt2",
						RTName:             "rt2",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
				},
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt3",
						RTName:             "rt3",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return(rtArr, nil)
			ots.EXPECT().GetObjectTypesMapByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ObjectType{}, nil).AnyTimes()
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			rts, total, err := service.ListRelationTypes(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 3)
			So(len(rts), ShouldEqual, 2)
			So(rts[0].RTID, ShouldEqual, "rt2")
		})
	})
}

func Test_relationTypeService_CreateRelationTypes(t *testing.T) {
	Convey("Test CreateRelationTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &relationTypeService{
			appSetting: appSetting,
			db:         db,
			rta:        rta,
			ps:         ps,
			ots:        ots,
			vba:        vba,
		}

		Convey("Success creating relation types with normal mode\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "relation_type1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ObjectType{ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "ot1"}}, nil).AnyTimes()
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", true, nil).AnyTimes()
			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil).AnyTimes()
			rta.EXPECT().CreateRelationType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			smock.ExpectCommit()

			result, err := service.CreateRelationTypes(ctx, nil, relationTypes, interfaces.ImportMode_Normal, true)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, "rt1")
		})

		Convey("Failed when permission check fails\n", func() {
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

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_InternalError_CheckPermissionFailed))

			result, err := service.CreateRelationTypes(ctx, nil, relationTypes, interfaces.ImportMode_Normal, true)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when relation type ID already exists in normal mode\n", func() {
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

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("rt1", true, nil)
			smock.ExpectRollback()

			result, err := service.CreateRelationTypes(ctx, nil, relationTypes, interfaces.ImportMode_Normal, true)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_RelationTypeIDExisted)
		})

		Convey("Success with ignore mode when relation type ID exists\n", func() {
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

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("rt1", true, nil)
			smock.ExpectCommit()

			result, err := service.CreateRelationTypes(ctx, nil, relationTypes, interfaces.ImportMode_Ignore, true)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Success with ignore mode when same name but different ID exists\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt2",
						RTName:             "relation_type1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ObjectType{}, nil).AnyTimes()
			rta.EXPECT().CreateRelationType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			result, err := service.CreateRelationTypes(ctx, nil, relationTypes, interfaces.ImportMode_Ignore, true)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("Success with Overwrite mode when ID exists\n", func() {
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

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("rt1", true, nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ObjectType{}, nil).AnyTimes()
			rta.EXPECT().UpdateRelationType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			smock.ExpectCommit()

			result, err := service.CreateRelationTypes(ctx, nil, relationTypes, interfaces.ImportMode_Overwrite, true)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Success with Overwrite mode when same name but different ID - creates new record\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt2",
						RTName:             "relation_type1",
						SourceObjectTypeID: "ot1",
						TargetObjectTypeID: "ot2",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ObjectType{}, nil).AnyTimes()
			rta.EXPECT().CreateRelationType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			result, err := service.CreateRelationTypes(ctx, nil, relationTypes, interfaces.ImportMode_Overwrite, true)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("Success with empty RTID generates new ID\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:   "",
						RTName: "relation_type1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx, knID, branch, rtID interface{}) {
				So(rtID, ShouldNotBeEmpty)
			}).Return("", false, nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.ObjectType{}, nil).AnyTimes()
			rta.EXPECT().CreateRelationType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			result, err := service.CreateRelationTypes(ctx, nil, relationTypes, interfaces.ImportMode_Normal, true)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldNotBeEmpty)
		})

		Convey("Failed when validateDependency returns error\n", func() {
			relationTypes := []*interfaces.RelationType{
				{
					RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
						RTID:               "rt1",
						RTName:             "relation_type1",
						SourceObjectTypeID: "ot1",
					},
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError)

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, httpErr)
			smock.ExpectRollback()

			result, err := service.CreateRelationTypes(ctx, nil, relationTypes, interfaces.ImportMode_Normal, true)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when CreateRelationType returns error\n", func() {
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

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.ObjectType{}, nil).AnyTimes()
			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			rta.EXPECT().CreateRelationType(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))
			smock.ExpectRollback()

			result, err := service.CreateRelationTypes(ctx, nil, relationTypes, interfaces.ImportMode_Normal, true)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when InsertDatasetData returns error\n", func() {
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

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.ObjectType{}, nil).AnyTimes()
			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			rta.EXPECT().CreateRelationType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))
			smock.ExpectRollback()

			result, err := service.CreateRelationTypes(ctx, nil, relationTypes, interfaces.ImportMode_Normal, true)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})
	})
}

func Test_relationTypeService_UpdateRelationType(t *testing.T) {
	Convey("Test UpdateRelationType\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &relationTypeService{
			appSetting: appSetting,
			db:         db,
			rta:        rta,
			ps:         ps,
			ots:        ots,
			vba:        vba,
		}

		Convey("Success updating relation type\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "relation_type1",
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.ObjectType{}, nil).AnyTimes()
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", true, nil).AnyTimes()
			rta.EXPECT().UpdateRelationType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := service.UpdateRelationType(ctx, nil, relationType, false)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation_type1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_InternalError_CheckPermissionFailed))

			err := service.UpdateRelationType(ctx, nil, relationType, false)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when validateDependency returns error\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "relation_type1",
					SourceObjectTypeID: "ot1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError)

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, httpErr)
			smock.ExpectRollback()

			// strictMode must be true; otherwise validateDependency short-circuits and never calls GetObjectTypeByID
			err := service.UpdateRelationType(ctx, nil, relationType, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when UpdateRelationType returns error\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation_type1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.ObjectType{}, nil).AnyTimes()
			rta.EXPECT().UpdateRelationType(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))
			smock.ExpectRollback()

			err := service.UpdateRelationType(ctx, nil, relationType, false)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when InsertDatasetData returns error\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "relation_type1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.ObjectType{}, nil).AnyTimes()
			rta.EXPECT().UpdateRelationType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))
			smock.ExpectRollback()

			err := service.UpdateRelationType(ctx, nil, relationType, false)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_relationTypeService_DeleteRelationTypesByIDs(t *testing.T) {
	Convey("Test DeleteRelationTypesByIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &relationTypeService{
			appSetting: appSetting,
			db:         db,
			rta:        rta,
			ps:         ps,
			vba:        vba,
		}

		Convey("Success deleting relation types\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1", "rt2"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().DeleteRelationTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(2), nil)
			vba.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
			smock.ExpectCommit()

			err := service.DeleteRelationTypesByIDs(ctx, nil, knID, branch, rtIDs)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_InternalError_CheckPermissionFailed))

			err := service.DeleteRelationTypesByIDs(ctx, nil, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteRelationTypesByIDs returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().DeleteRelationTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))
			smock.ExpectRollback()

			err := service.DeleteRelationTypesByIDs(ctx, nil, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteData returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			rtIDs := []string{"rt1"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rta.EXPECT().DeleteRelationTypesByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			vba.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))
			smock.ExpectRollback()

			err := service.DeleteRelationTypesByIDs(ctx, nil, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_relationTypeService_InsertDatasetData(t *testing.T) {
	Convey("Test InsertDatasetData\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		service := &relationTypeService{
			appSetting: appSetting,
			vba:        vba,
		}

		Convey("Success inserting empty list\n", func() {
			relationTypes := []*interfaces.RelationType{}

			err := service.InsertDatasetData(ctx, relationTypes)
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

			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.InsertDatasetData(ctx, relationTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when WriteDatasetDocuments returns error\n", func() {
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

			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			err := service.InsertDatasetData(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Success inserting relation types with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			vbaWithVector := bmock.NewMockVegaBackendAccess(mockCtrl)
			mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &relationTypeService{
				appSetting: appSettingWithVector,
				vba:        vbaWithVector,
				mfa:        mfa,
			}

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
			vbaWithVector.EXPECT().WriteDatasetDocuments(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := serviceWithVector.InsertDatasetData(ctx, relationTypes)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &relationTypeService{
				appSetting: appSettingWithVector,
				mfa:        mfa,
			}

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

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			err := serviceWithVector.InsertDatasetData(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &relationTypeService{
				appSetting: appSettingWithVector,
				mfa:        mfa,
			}

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
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			err := serviceWithVector.InsertDatasetData(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when vector count mismatch with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &relationTypeService{
				appSetting: appSettingWithVector,
				mfa:        mfa,
			}

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
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)

			err := serviceWithVector.InsertDatasetData(ctx, relationTypes)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_relationTypeService_GetTotal(t *testing.T) {
	Convey("Test GetTotal\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		service := &relationTypeService{
			appSetting: appSetting,
			vba:        vba,
		}

		Convey("Success getting total\n", func() {
			filterCondition := map[string]any{
				"field":      "module_type",
				"operation":  "==",
				"value":      interfaces.MODULE_TYPE_RELATION_TYPE,
				"value_from": "const",
			}

			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DatasetQueryResponse{
				TotalCount: 10,
			}, nil)

			total, err := service.GetTotal(ctx, filterCondition)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 10)
		})

		Convey("Failed when QueryResourceData fails\n", func() {
			filterCondition := map[string]any{}

			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			total, err := service.GetTotal(ctx, filterCondition)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})
	})
}

func Test_relationTypeService_GetTotalWithLargeRTIDs(t *testing.T) {
	Convey("Test GetTotalWithLargeRTIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		service := &relationTypeService{
			appSetting: appSetting,
			vba:        vba,
		}

		Convey("Success getting total with large RTIDs\n", func() {
			filterCondition := map[string]any{
				"field":      "module_type",
				"operation":  "==",
				"value":      interfaces.MODULE_TYPE_RELATION_TYPE,
				"value_from": "const",
			}
			rtIDs := []string{"rt1", "rt2", "rt3"}

			// Mock GetTotalWithRTIDs calls
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DatasetQueryResponse{
				TotalCount: 5,
			}, nil).Times(1)

			total, err := service.GetTotalWithLargeRTIDs(ctx, filterCondition, rtIDs)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 5)
		})

		Convey("Success with empty RTIDs\n", func() {
			filterCondition := map[string]any{}
			rtIDs := []string{}

			total, err := service.GetTotalWithLargeRTIDs(ctx, filterCondition, rtIDs)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
		})

		Convey("Failed when GetTotalWithRTIDs returns error\n", func() {
			filterCondition := map[string]any{}
			rtIDs := []string{"rt1"}

			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			total, err := service.GetTotalWithLargeRTIDs(ctx, filterCondition, rtIDs)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})
	})
}

func Test_relationTypeService_GetTotalWithRTIDs(t *testing.T) {
	Convey("Test GetTotalWithRTIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)

		service := &relationTypeService{
			appSetting: appSetting,
			vba:        vba,
		}

		Convey("Success getting total with RTIDs\n", func() {
			filterCondition := map[string]any{
				"field":      "module_type",
				"operation":  "==",
				"value":      interfaces.MODULE_TYPE_RELATION_TYPE,
				"value_from": "const",
			}
			rtIDs := []string{"rt1", "rt2"}

			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DatasetQueryResponse{
				TotalCount: 2,
			}, nil)

			total, err := service.GetTotalWithRTIDs(ctx, filterCondition, rtIDs)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)
		})

		Convey("Failed when GetTotal returns error\n", func() {
			filterCondition := map[string]any{}
			rtIDs := []string{"rt1"}

			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			total, err := service.GetTotalWithRTIDs(ctx, filterCondition, rtIDs)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
		})
	})
}

func Test_relationTypeService_SearchRelationTypes(t *testing.T) {
	Convey("Test SearchRelationTypes\n", t, func() {
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
		ps := bmock.NewMockPermissionService(mockCtrl)

		service := &relationTypeService{
			appSetting: appSetting,
			cga:        cga,
			vba:        vba,
			ps:         ps,
		}

		Convey("Success searching relation types without concept groups\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				Limit:  10,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DatasetQueryResponse{
				Entries:     []map[string]any{},
				TotalCount:  0,
				SearchAfter: nil,
			}, nil)

			result, err := service.SearchRelationTypes(ctx, query)
			So(err, ShouldBeNil)
			So(result.Entries, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Success searching relation types with concept groups\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				ConceptGroups: []string{"cg1"},
				ActualCondition: &cond.CondCfg{
					Operation: "and",
					SubConds: []*cond.CondCfg{
						{
							Field:     "name",
							Operation: cond.OperationEq,
							ValueOptCfg: cond.ValueOptCfg{
								ValueFrom: "const",
								Value:     "rt1",
							},
						},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			cga.EXPECT().GetRelationTypeIDsFromConceptGroupRelation(gomock.Any(), gomock.Any()).Return([]string{"rt1"}, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DatasetQueryResponse{
				Entries:     []map[string]any{},
				TotalCount:  0,
				SearchAfter: nil,
			}, nil)

			result, err := service.SearchRelationTypes(ctx, query)
			So(err, ShouldBeNil)
			So(result.Entries, ShouldNotBeNil)
		})

		Convey("Failed when concept groups not found\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				NeedTotal:     false,
				Limit:         10,
				ConceptGroups: []string{"cg1"},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(0, nil)

			result, err := service.SearchRelationTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Failed when GetConceptGroupsTotal returns error\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				ConceptGroups: []string{"cg1"},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			result, err := service.SearchRelationTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Failed when GetRelationTypeIDsFromConceptGroupRelation returns error\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				ConceptGroups: []string{"cg1"},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			cga.EXPECT().GetRelationTypeIDsFromConceptGroupRelation(gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			result, err := service.SearchRelationTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Success with empty concept groups\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				Limit:  10,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DatasetQueryResponse{
				Entries:     []map[string]any{},
				TotalCount:  0,
				SearchAfter: nil,
			}, nil)

			result, err := service.SearchRelationTypes(ctx, query)
			So(err, ShouldBeNil)
			So(result.Entries, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})
	})
}

func Test_relationTypeService_validateDependency(t *testing.T) {
	Convey("Test validateDependency\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		dva := bmock.NewMockDataViewAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &relationTypeService{
			appSetting: appSetting,
			db:         db,
			ots:        ots,
			dva:        dva,
		}

		Convey("Failed when source object type not found\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "rt1",
					SourceObjectTypeID: "ot1",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError)

			smock.ExpectBegin()
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, httpErr)
			smock.ExpectRollback()

			err := service.validateDependency(ctx, nil, relationType, true, nil)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_InternalError)
		})

		Convey("Failed when target object type not found\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "rt1",
					TargetObjectTypeID: "ot2",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError)

			smock.ExpectBegin()
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, httpErr)
			smock.ExpectRollback()

			err := service.validateDependency(ctx, nil, relationType, true, nil)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_InternalError)
		})

		Convey("Failed when source property not found in DIRECT type\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "rt1",
					Type:               interfaces.RELATION_TYPE_DIRECT,
					SourceObjectTypeID: "ot1",
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{
								Name: "prop1",
							},
						},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			sourceObjectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "ot1",
				},
				PropertyMap: map[string]string{},
			}

			smock.ExpectBegin()
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sourceObjectType, nil)
			smock.ExpectRollback()

			err := service.validateDependency(ctx, nil, relationType, true, nil)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_InvalidParameter)
		})

		Convey("Success DIRECT when GetObjectTypeByID returns OTs with DataProperties only (no PropertyMap; regression)\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "rt1",
					Type:               interfaces.RELATION_TYPE_DIRECT,
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "prop1"},
							TargetProp: interfaces.SimpleProperty{Name: "prop2"},
						},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			sourceOT := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "员工",
					DataProperties: []*interfaces.DataProperty{
						{Name: "prop1", DisplayName: "展示名"},
					},
				},
			}
			targetOT := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot2",
					OTName: "部门",
					DataProperties: []*interfaces.DataProperty{
						{Name: "prop2", DisplayName: "P2"},
					},
				},
			}
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "ot1").Return(sourceOT, nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "ot2").Return(targetOT, nil)

			err := service.validateDependency(ctx, nil, relationType, true, nil)
			So(err, ShouldBeNil)
		})

		Convey("Failed when target property not found in DIRECT type\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "rt1",
					Type:               interfaces.RELATION_TYPE_DIRECT,
					TargetObjectTypeID: "ot2",
					MappingRules: []interfaces.Mapping{
						{
							TargetProp: interfaces.SimpleProperty{
								Name: "prop2",
							},
						},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			targetObjectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot2",
					OTName: "ot2",
				},
				PropertyMap: map[string]string{},
			}

			smock.ExpectBegin()
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(targetObjectType, nil)
			smock.ExpectRollback()

			err := service.validateDependency(ctx, nil, relationType, true, nil)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_InvalidParameter)
		})

		Convey("Failed when data view not found in DATA_VIEW type\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "rt1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: &interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							ID: "dv1",
						},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			dva.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any()).Return(nil, nil)

			err := service.validateDependency(ctx, nil, relationType, true, nil)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_InvalidParameter)
		})

		Convey("Failed when GetDataViewByID returns error in DATA_VIEW type\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:   "rt1",
					RTName: "rt1",
					Type:   interfaces.RELATION_TYPE_DATA_VIEW,
					MappingRules: &interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							ID: "dv1",
						},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			dva.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))

			err := service.validateDependency(ctx, nil, relationType, true, nil)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when source mapping field not found in data view\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "rt1",
					Type:               interfaces.RELATION_TYPE_DATA_VIEW,
					SourceObjectTypeID: "ot1",
					MappingRules: &interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							ID: "dv1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								TargetProp: interfaces.SimpleProperty{
									Name: "field1",
								},
							},
						},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			sourceObjectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "ot1",
				},
				PropertyMap: map[string]string{
					"prop1": "Property1",
				},
			}
			dataView := &interfaces.DataView{
				ViewName:  "data_view1",
				FieldsMap: map[string]*interfaces.ViewField{},
			}

			smock.ExpectBegin()
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sourceObjectType, nil)
			dva.EXPECT().GetDataViewByID(gomock.Any(), gomock.Any()).Return(dataView, nil)
			smock.ExpectRollback()

			err := service.validateDependency(ctx, nil, relationType, true, nil)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_InvalidParameter)
		})

		Convey("Success when source/target resolved from batch preflight index\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "rt1",
					Type:               interfaces.RELATION_TYPE_DIRECT,
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "p1"},
							TargetProp: interfaces.SimpleProperty{Name: "p2"},
						},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			batch, err := batchindex.CollectKNFromPayload(&interfaces.KN{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				ObjectTypes: []*interfaces.ObjectType{
					{
						ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
							OTID: "ot1",
							DataProperties: []*interfaces.DataProperty{
								{Name: "p1", DisplayName: "P1", Type: "string"},
							},
						},
					},
					{
						ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
							OTID: "ot2",
							DataProperties: []*interfaces.DataProperty{
								{Name: "p2", DisplayName: "P2", Type: "string"},
							},
						},
					},
				},
			})
			So(err, ShouldBeNil)
			err = service.validateDependency(ctx, nil, relationType, true, batch)
			So(err, ShouldBeNil)
		})

		Convey("Success when batch OT has no data properties (mapping validation degraded)\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					Type:               interfaces.RELATION_TYPE_DIRECT,
					SourceObjectTypeID: "ot1",
					MappingRules: []interfaces.Mapping{
						{SourceProp: interfaces.SimpleProperty{Name: "would_fail_if_mapped"}},
					},
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			batch := batchindex.NewBatchIDIndex("kn1", interfaces.MAIN_BRANCH)
			batch.ObjectTypes["ot1"] = &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "ot1"},
			}
			err := service.validateDependency(ctx, nil, relationType, true, batch)
			So(err, ShouldBeNil)
		})

		Convey("strictMode false returns nil without resolving dependencies\n", func() {
			relationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "rt1",
					SourceObjectTypeID: "ot_missing",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			err := service.validateDependency(ctx, nil, relationType, false, nil)
			So(err, ShouldBeNil)
		})
	})
}

func Test_relationTypeService_ValidateRelationTypes(t *testing.T) {
	Convey("Test ValidateRelationTypes\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ps := bmock.NewMockPermissionService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		dva := bmock.NewMockDataViewAccess(mockCtrl)
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &relationTypeService{
			appSetting: appSetting,
			db:         db,
			ps:         ps,
			ots:        ots,
			dva:        dva,
			rta:        rta,
		}

		expectRTImportOK := func() {
			rta.EXPECT().CheckRelationTypeExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
		}

		relationTypes := []*interfaces.RelationType{
			{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "rt1",
					SourceObjectTypeID: "ot_missing",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			},
		}

		Convey("strictMode false skips validateDependency\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			expectRTImportOK()
			err := service.ValidateRelationTypes(ctx, "kn1", interfaces.MAIN_BRANCH, relationTypes, false, nil, interfaces.ImportMode_Normal)
			So(err, ShouldBeNil)
		})

		Convey("strictMode true runs validateDependency and fails when source OT missing\n", func() {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError)
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			expectRTImportOK()
			smock.ExpectBegin()
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, httpErr)
			smock.ExpectRollback()
			err := service.ValidateRelationTypes(ctx, "kn1", interfaces.MAIN_BRANCH, relationTypes, true, nil, interfaces.ImportMode_Normal)
			So(err, ShouldNotBeNil)
			So(err.(*rest.HTTPError).BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_RelationType_InternalError)
		})
	})
}

func Test_relationTypeService_DeleteRelationTypesByKnID(t *testing.T) {
	Convey("Test DeleteRelationTypesByKnID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		service := &relationTypeService{appSetting: &common.AppSetting{}, rta: rta}

		Convey("Failed when tx is nil\n", func() {
			err := service.DeleteRelationTypesByKnID(ctx, nil, "kn1", interfaces.MAIN_BRANCH)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when access layer returns error\n", func() {
			tx := new(sql.Tx)
			rta.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), tx, "kn1", interfaces.MAIN_BRANCH).
				Return(int64(0), rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))
			err := service.DeleteRelationTypesByKnID(ctx, tx, "kn1", interfaces.MAIN_BRANCH)
			So(err, ShouldNotBeNil)
		})

		Convey("Success\n", func() {
			tx := new(sql.Tx)
			rta.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), tx, "kn1", interfaces.MAIN_BRANCH).
				Return(int64(3), nil)
			err := service.DeleteRelationTypesByKnID(ctx, tx, "kn1", interfaces.MAIN_BRANCH)
			So(err, ShouldBeNil)
		})
	})
}

func Test_relationTypeService_SearchRelationTypes_extraCases(t *testing.T) {
	Convey("Test SearchRelationTypes extra cases\n", t, func() {
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
		ps := bmock.NewMockPermissionService(mockCtrl)

		service := &relationTypeService{
			appSetting: appSetting,
			cga:        cga,
			vba:        vba,
			ps:         ps,
		}

		Convey("Failed when CheckPermission returns error\n", func() {
			query := &interfaces.ConceptsQuery{KNID: "kn1", Branch: interfaces.MAIN_BRANCH, Limit: 10}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_InternalError_CheckPermissionFailed))
			result, err := service.SearchRelationTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Failed when QueryResourceData returns error\n", func() {
			query := &interfaces.ConceptsQuery{KNID: "kn1", Branch: interfaces.MAIN_BRANCH, Limit: 10}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_RelationType_InternalError))
			result, err := service.SearchRelationTypes(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Success when GetRelationTypeIDsFromConceptGroupRelation returns empty list\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				ConceptGroups: []string{"cg1"},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			cga.EXPECT().GetRelationTypeIDsFromConceptGroupRelation(gomock.Any(), gomock.Any()).Return([]string{}, nil)
			result, err := service.SearchRelationTypes(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Success with NeedTotal and no concept groups\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:      "kn1",
				Branch:    interfaces.MAIN_BRANCH,
				Limit:     10,
				NeedTotal: true,
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			// NeedTotal block: QueryResourceData with Limit=1, NeedTotal=true
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DatasetQueryResponse{
				Entries: []map[string]any{}, TotalCount: 5,
			}, nil)
			// Main loop: QueryResourceData returns empty → break
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DatasetQueryResponse{
				Entries: []map[string]any{},
			}, nil)
			result, err := service.SearchRelationTypes(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 5)
		})

		Convey("Success returning actual relation type entries\n", func() {
			query := &interfaces.ConceptsQuery{KNID: "kn1", Branch: interfaces.MAIN_BRANCH, Limit: 10}
			entry := map[string]any{
				"rt_id":   "rt1",
				"rt_name": "relation1",
				"_score":  float64(0.95),
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.DatasetQueryResponse{
				Entries: []map[string]any{entry},
			}, nil)
			result, err := service.SearchRelationTypes(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Entries), ShouldEqual, 1)
		})
	})
}
