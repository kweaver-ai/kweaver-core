// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knowledge_network

import (
	"context"
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
)

func Test_knowledgeNetworkService_CheckKNExistByID(t *testing.T) {
	Convey("Test CheckKNExistByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := bmock.NewMockKNAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
		}

		Convey("Success when KN exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			knName := "knowledge_network1"

			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(knName, true, nil)

			name, exist, err := service.CheckKNExistByID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(name, ShouldEqual, knName)
		})

		Convey("Success when KN does not exist\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH

			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			name, exist, err := service.CheckKNExistByID(ctx, knID, branch)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH

			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			name, exist, err := service.CheckKNExistByID(ctx, knID, branch)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_KnowledgeNetwork_InternalError_CheckKNIfExistFailed)
		})
	})
}

func Test_knowledgeNetworkService_CheckKNExistByName(t *testing.T) {
	Convey("Test CheckKNExistByName\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := bmock.NewMockKNAccess(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
		}

		Convey("Success when KN exists\n", func() {
			knName := "knowledge_network1"
			branch := interfaces.MAIN_BRANCH
			knID := "kn1"

			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(knID, true, nil)

			id, exist, err := service.CheckKNExistByName(ctx, knName, branch)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(id, ShouldEqual, knID)
		})

		Convey("Success when KN does not exist\n", func() {
			knName := "knowledge_network1"
			branch := interfaces.MAIN_BRANCH

			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			id, exist, err := service.CheckKNExistByName(ctx, knName, branch)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knName := "knowledge_network1"
			branch := interfaces.MAIN_BRANCH

			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			id, exist, err := service.CheckKNExistByName(ctx, knName, branch)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_KnowledgeNetwork_InternalError_CheckKNIfExistFailed)
		})
	})
}

func Test_knowledgeNetworkService_UpdateKNDetail(t *testing.T) {
	Convey("Test UpdateKNDetail\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := bmock.NewMockKNAccess(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
		}

		Convey("Success updating KN detail\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			detail := "updated detail"

			kna.EXPECT().UpdateKNDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.UpdateKNDetail(ctx, knID, branch, detail)
			So(err, ShouldBeNil)
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			detail := "updated detail"

			kna.EXPECT().UpdateKNDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			err := service.UpdateKNDetail(ctx, knID, branch, detail)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_KnowledgeNetwork_InternalError)
		})
	})
}

func Test_knowledgeNetworkService_GetStatByKN(t *testing.T) {
	Convey("Test GetStatByKN\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ota := bmock.NewMockObjectTypeAccess(mockCtrl)
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		ata := bmock.NewMockActionTypeAccess(mockCtrl)
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		rtA := bmock.NewMockRiskTypeAccess(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			ota:        ota,
			rta:        rta,
			ata:        ata,
			cga:        cga,
			riskTypeA:  rtA,
		}

		Convey("Success getting statistics\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ota.EXPECT().GetObjectTypesTotal(gomock.Any(), gomock.Any()).Return(10, nil)
			rta.EXPECT().GetRelationTypesTotal(gomock.Any(), gomock.Any()).Return(5, nil)
			ata.EXPECT().GetActionTypesTotal(gomock.Any(), gomock.Any()).Return(3, nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(2, nil)
			rtA.EXPECT().GetRiskTypesTotal(gomock.Any(), gomock.Any()).Return(4, nil)

			stats, err := service.GetStatByKN(ctx, kn)
			So(err, ShouldBeNil)
			So(stats, ShouldNotBeNil)
			So(stats.OtTotal, ShouldEqual, 10)
			So(stats.RtTotal, ShouldEqual, 5)
			So(stats.AtTotal, ShouldEqual, 3)
			So(stats.CgTotal, ShouldEqual, 2)
			So(stats.RiskTypeTotal, ShouldEqual, 4)
		})

		Convey("Failed when getting object types total returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ota.EXPECT().GetObjectTypesTotal(gomock.Any(), gomock.Any()).Return(0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			stats, err := service.GetStatByKN(ctx, kn)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_KnowledgeNetwork_InternalError_GetObjectTypesTotalFailed)
		})

		Convey("Failed when getting relation types total returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ota.EXPECT().GetObjectTypesTotal(gomock.Any(), gomock.Any()).Return(10, nil)
			rta.EXPECT().GetRelationTypesTotal(gomock.Any(), gomock.Any()).Return(0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			stats, err := service.GetStatByKN(ctx, kn)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_KnowledgeNetwork_InternalError_GetRelationTypesTotalFailed)
		})

		Convey("Failed when getting action types total returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ota.EXPECT().GetObjectTypesTotal(gomock.Any(), gomock.Any()).Return(10, nil)
			rta.EXPECT().GetRelationTypesTotal(gomock.Any(), gomock.Any()).Return(5, nil)
			ata.EXPECT().GetActionTypesTotal(gomock.Any(), gomock.Any()).Return(0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			stats, err := service.GetStatByKN(ctx, kn)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_KnowledgeNetwork_InternalError_GetRelationTypesTotalFailed)
		})

		Convey("Failed when getting concept groups total returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ota.EXPECT().GetObjectTypesTotal(gomock.Any(), gomock.Any()).Return(10, nil)
			rta.EXPECT().GetRelationTypesTotal(gomock.Any(), gomock.Any()).Return(5, nil)
			ata.EXPECT().GetActionTypesTotal(gomock.Any(), gomock.Any()).Return(3, nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			stats, err := service.GetStatByKN(ctx, kn)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_KnowledgeNetwork_InternalError_GetRelationTypesTotalFailed)
		})

		Convey("Failed when getting risk types total returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ota.EXPECT().GetObjectTypesTotal(gomock.Any(), gomock.Any()).Return(10, nil)
			rta.EXPECT().GetRelationTypesTotal(gomock.Any(), gomock.Any()).Return(5, nil)
			ata.EXPECT().GetActionTypesTotal(gomock.Any(), gomock.Any()).Return(3, nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(2, nil)
			rtA.EXPECT().GetRiskTypesTotal(gomock.Any(), gomock.Any()).Return(0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			stats, err := service.GetStatByKN(ctx, kn)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_KnowledgeNetwork_InternalError_GetRiskTypesTotalFailed)
		})
	})
}

func Test_knowledgeNetworkService_ListKNs(t *testing.T) {
	Convey("Test ListKNs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := bmock.NewMockKNAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		ums := bmock.NewMockUserMgmtService(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
			ums:        ums,
		}

		Convey("Success listing KNs\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			knArr := []*interfaces.KN{
				{
					KNID:   "kn1",
					KNName: "kn1",
				},
			}

			kna.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return(knArr, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			kns, total, err := service.ListKNs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(kns), ShouldEqual, 1)
		})

		Convey("Success with empty result\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			kna.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return([]*interfaces.KN{}, nil)

			kns, total, err := service.ListKNs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(kns), ShouldEqual, 0)
		})

		Convey("Failed when access layer returns error\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			kna.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			kns, total, err := service.ListKNs(ctx, parameter)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(kns), ShouldEqual, 0)
		})

		Convey("Failed when FilterResources returns error\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			knArr := []*interfaces.KN{
				{
					KNID:   "kn1",
					KNName: "kn1",
				},
			}

			kna.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return(knArr, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			kns, total, err := service.ListKNs(ctx, parameter)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(kns), ShouldEqual, 0)
		})

		Convey("Failed when GetAccountNames returns error\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			knArr := []*interfaces.KN{
				{
					KNID:   "kn1",
					KNName: "kn1",
				},
			}

			kna.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return(knArr, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			kns, total, err := service.ListKNs(ctx, parameter)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(kns), ShouldEqual, 0)
		})

		Convey("Success with Limit = -1\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  -1,
					Offset: 0,
				},
			}
			knArr := []*interfaces.KN{
				{
					KNID:   "kn1",
					KNName: "kn1",
				},
			}

			kna.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return(knArr, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			kns, total, err := service.ListKNs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(kns), ShouldEqual, 1)
		})

		Convey("Success with Offset out of bounds\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 100,
				},
			}
			knArr := []*interfaces.KN{
				{
					KNID:   "kn1",
					KNName: "kn1",
				},
			}

			kna.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return(knArr, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)

			kns, total, err := service.ListKNs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(kns), ShouldEqual, 0)
		})

		Convey("Success with pagination\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  2,
					Offset: 1,
				},
			}
			knArr := []*interfaces.KN{
				{
					KNID:   "kn1",
					KNName: "kn1",
				},
				{
					KNID:   "kn2",
					KNName: "kn2",
				},
				{
					KNID:   "kn3",
					KNName: "kn3",
				},
			}

			kna.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return(knArr, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}},
					"kn2": {Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}},
					"kn3": {Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}},
				}, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			kns, total, err := service.ListKNs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 3)
			So(len(kns), ShouldEqual, 2)
			So(kns[0].KNID, ShouldEqual, "kn2")
		})
	})
}

func Test_knowledgeNetworkService_GetKNByID(t *testing.T) {
	Convey("Test GetKNByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := bmock.NewMockKNAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		ums := bmock.NewMockUserMgmtService(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
			ums:        ums,
		}

		Convey("Success getting KN by ID\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			mode := ""
			kn := &interfaces.KN{
				KNID:   knID,
				KNName: "kn1",
			}

			kna.EXPECT().GetKNByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(kn, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

			result, err := service.GetKNByID(ctx, knID, branch, mode)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.KNID, ShouldEqual, knID)
		})

		Convey("Failed when KN not found\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			mode := ""

			kna.EXPECT().GetKNByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

			result, err := service.GetKNByID(ctx, knID, branch, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_KnowledgeNetwork_NotFound)
		})

		Convey("Failed when GetKNByID returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			mode := ""

			kna.EXPECT().GetKNByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			result, err := service.GetKNByID(ctx, knID, branch, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_KnowledgeNetwork_InternalError_GetKNByIDFailed)
		})

		Convey("Failed when FilterResources returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			mode := ""
			kn := &interfaces.KN{
				KNID:   knID,
				KNName: "kn1",
			}

			kna.EXPECT().GetKNByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(kn, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			result, err := service.GetKNByID(ctx, knID, branch, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Failed when no permission\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			mode := ""
			kn := &interfaces.KN{
				KNID:   knID,
				KNName: "kn1",
			}

			kna.EXPECT().GetKNByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(kn, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{}, nil)

			result, err := service.GetKNByID(ctx, knID, branch, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, rest.PublicError_Forbidden)
		})

		Convey("Failed when GetAccountNames returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			mode := ""
			kn := &interfaces.KN{
				KNID:   knID,
				KNName: "kn1",
			}

			kna.EXPECT().GetKNByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(kn, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			result, err := service.GetKNByID(ctx, knID, branch, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Success with export mode\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			mode := interfaces.Mode_Export
			kn := &interfaces.KN{
				KNID:   knID,
				KNName: "kn1",
				Branch: branch,
			}
			cgs := bmock.NewMockConceptGroupService(mockCtrl)
			ots := bmock.NewMockObjectTypeService(mockCtrl)
			rts := bmock.NewMockRelationTypeService(mockCtrl)
			ats := bmock.NewMockActionTypeService(mockCtrl)

			service := &knowledgeNetworkService{
				appSetting: appSetting,
				kna:        kna,
				ps:         ps,
				ums:        ums,
				cgs:        cgs,
				ots:        ots,
				rts:        rts,
				ats:        ats,
			}

			kna.EXPECT().GetKNByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(kn, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)
			cgs.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return([]*interfaces.ConceptGroup{}, 0, nil)
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.ObjectType{}, 0, nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, 0, nil)
			ats.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{}, 0, nil)

			result, err := service.GetKNByID(ctx, knID, branch, mode)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.KNID, ShouldEqual, knID)
		})
	})
}

func Test_knowledgeNetworkService_InsertDatasetData(t *testing.T) {
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

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			vba:        vba,
		}

		Convey("Success inserting dataset data\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := service.InsertDatasetData(ctx, kn)
			So(err, ShouldBeNil)
		})

		Convey("Failed when WriteDatasetDocuments returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			err := service.InsertDatasetData(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Success inserting dataset data with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			vbaWithVector := bmock.NewMockVegaBackendAccess(mockCtrl)
			mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &knowledgeNetworkService{
				appSetting: appSettingWithVector,
				vba:        vbaWithVector,
				mfa:        mfa,
			}

			kn := &interfaces.KN{
				KNID:          "kn1",
				KNName:        "kn1",
				Tags:          []string{"tag1"},
				Comment:       "comment",
				BKNRawContent: "detail",
				Branch:        interfaces.MAIN_BRANCH,
			}
			vectors := []*cond.VectorResp{
				{
					Vector: []float32{0.1, 0.2, 0.3},
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)
			vbaWithVector.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := serviceWithVector.InsertDatasetData(ctx, kn)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &knowledgeNetworkService{
				appSetting: appSettingWithVector,
				mfa:        mfa,
			}

			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			err := serviceWithVector.InsertDatasetData(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &knowledgeNetworkService{
				appSetting: appSettingWithVector,
				mfa:        mfa,
			}

			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			err := serviceWithVector.InsertDatasetData(ctx, kn)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_knowledgeNetworkService_UpdateKN(t *testing.T) {
	Convey("Test UpdateKN\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		kna := bmock.NewMockKNAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
			vba:        vba,
			db:         db,
		}

		Convey("Success updating KN\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().UpdateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := service.UpdateKN(ctx, nil, kn, false)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_KnowledgeNetwork_InternalError))

			err := service.UpdateKN(ctx, nil, kn, false)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when UpdateKN returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().UpdateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.UpdateKN(ctx, nil, kn, false)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when InsertDatasetData returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().UpdateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.UpdateKN(ctx, nil, kn, false)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with IfNameModify true\n", func() {
			kn := &interfaces.KN{
				KNID:         "kn1",
				KNName:       "kn1",
				Branch:       interfaces.MAIN_BRANCH,
				IfNameModify: true,
			}
			ps2 := bmock.NewMockPermissionService(mockCtrl)

			service := &knowledgeNetworkService{
				appSetting: appSetting,
				kna:        kna,
				ps:         ps2,
				vba:        vba,
				db:         db,
			}

			smock.ExpectBegin()
			ps2.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().UpdateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			ps2.EXPECT().UpdateResource(gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := service.UpdateKN(ctx, nil, kn, false)
			So(err, ShouldBeNil)
		})

		Convey("Failed when UpdateResource returns error\n", func() {
			kn := &interfaces.KN{
				KNID:         "kn1",
				KNName:       "kn1",
				Branch:       interfaces.MAIN_BRANCH,
				IfNameModify: true,
			}
			ps2 := bmock.NewMockPermissionService(mockCtrl)

			service := &knowledgeNetworkService{
				appSetting: appSetting,
				kna:        kna,
				ps:         ps2,
				vba:        vba,
				db:         db,
			}

			smock.ExpectBegin()
			ps2.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().UpdateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			ps2.EXPECT().UpdateResource(gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectCommit()

			err := service.UpdateKN(ctx, nil, kn, false)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_knowledgeNetworkService_DeleteKN(t *testing.T) {
	Convey("Test DeleteKN\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := bmock.NewMockKNAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		bss := bmock.NewMockBusinessSystemService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		riskTypeS := bmock.NewMockRiskTypeService(mockCtrl)
		js := bmock.NewMockJobService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)

		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
			vba:        vba,
			bss:        bss,
			ots:        ots,
			rts:        rts,
			ats:        ats,
			riskTypeS:  riskTypeS,
			js:         js,
			cgs:        cgs,
			db:         db,
		}

		Convey("Success deleting KN\n", func() {
			kn := &interfaces.KN{
				KNID:           "kn1",
				KNName:         "kn1",
				Branch:         interfaces.MAIN_BRANCH,
				BusinessDomain: "bd1",
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().DeleteObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().DeleteActionTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			riskTypeS.EXPECT().DeleteRiskTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cgs.EXPECT().DeleteConceptGroupsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().DeleteJobsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			vba.EXPECT().DeleteDatasetDocumentsByQuery(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			bss.EXPECT().UnbindResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_KnowledgeNetwork_InternalError))

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteKN returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteObjectTypesByKnID returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().DeleteObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteRelationTypesByKnID returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().DeleteObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteActionTypesByKnID returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().DeleteObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().DeleteActionTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteConceptGroupsByKnID returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().DeleteObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().DeleteActionTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			riskTypeS.EXPECT().DeleteRiskTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cgs.EXPECT().DeleteConceptGroupsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteJobsByKnID returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().DeleteObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().DeleteActionTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			riskTypeS.EXPECT().DeleteRiskTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cgs.EXPECT().DeleteConceptGroupsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().DeleteJobsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteDatasetDocumentByID returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().DeleteObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().DeleteActionTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			riskTypeS.EXPECT().DeleteRiskTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cgs.EXPECT().DeleteConceptGroupsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().DeleteJobsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteDatasetDocumentsByQuery returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().DeleteObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().DeleteActionTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			riskTypeS.EXPECT().DeleteRiskTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cgs.EXPECT().DeleteConceptGroupsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().DeleteJobsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			vba.EXPECT().DeleteDatasetDocumentsByQuery(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteResources returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().DeleteObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().DeleteActionTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			riskTypeS.EXPECT().DeleteRiskTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cgs.EXPECT().DeleteConceptGroupsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().DeleteJobsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			vba.EXPECT().DeleteDatasetDocumentsByQuery(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when UnbindResource returns error\n", func() {
			kn := &interfaces.KN{
				KNID:           "kn1",
				KNName:         "kn1",
				Branch:         interfaces.MAIN_BRANCH,
				BusinessDomain: "bd1",
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().DeleteObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().DeleteActionTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			riskTypeS.EXPECT().DeleteRiskTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cgs.EXPECT().DeleteConceptGroupsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().DeleteJobsByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			vba.EXPECT().DeleteDatasetDocumentsByQuery(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			ps.EXPECT().DeleteResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			bss.EXPECT().UnbindResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteRiskTypesByKnID returns error\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)
			kna.EXPECT().DeleteKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			ots.EXPECT().DeleteObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().DeleteRelationTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().DeleteActionTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			riskTypeS.EXPECT().DeleteRiskTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			err := service.DeleteKN(ctx, kn)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_knowledgeNetworkService_ListKnSrcs(t *testing.T) {
	Convey("Test ListKnSrcs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := bmock.NewMockKNAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
		}

		Convey("Success listing KN sources\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			knList := []interfaces.PermissionResource{
				{
					ID:   "kn1",
					Name: "kn1",
				},
			}

			kna.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return(knList, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)

			resources, total, err := service.ListKnSrcs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(resources), ShouldEqual, 1)
		})

		Convey("Success with empty result\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			kna.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return([]interfaces.PermissionResource{}, nil)

			resources, total, err := service.ListKnSrcs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(resources), ShouldEqual, 0)
		})

		Convey("Failed when access layer returns error\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			kna.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			resources, total, err := service.ListKnSrcs(ctx, parameter)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(resources), ShouldEqual, 0)
		})

		Convey("Failed when FilterResources returns error\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			knList := []interfaces.PermissionResource{
				{
					ID:   "kn1",
					Name: "kn1",
				},
			}

			kna.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return(knList, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			resources, total, err := service.ListKnSrcs(ctx, parameter)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(resources), ShouldEqual, 0)
		})

		Convey("Success with Limit = -1\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  -1,
					Offset: 0,
				},
			}
			knList := []interfaces.PermissionResource{
				{
					ID:   "kn1",
					Name: "kn1",
				},
			}

			kna.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return(knList, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)

			resources, total, err := service.ListKnSrcs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(resources), ShouldEqual, 1)
		})

		Convey("Success with Offset out of bounds\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 100,
				},
			}
			knList := []interfaces.PermissionResource{
				{
					ID:   "kn1",
					Name: "kn1",
				},
			}

			kna.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return(knList, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {
						Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
					},
				}, nil)

			resources, total, err := service.ListKnSrcs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(resources), ShouldEqual, 0)
		})

		Convey("Success with pagination\n", func() {
			parameter := interfaces.KNsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  2,
					Offset: 1,
				},
			}
			knList := []interfaces.PermissionResource{
				{ID: "kn1", Name: "kn1"},
				{ID: "kn2", Name: "kn2"},
				{ID: "kn3", Name: "kn3"},
			}

			kna.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return(knList, nil)
			ps.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(map[string]interfaces.PermissionResourceOps{
					"kn1": {Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}},
					"kn2": {Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}},
					"kn3": {Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}},
				}, nil)

			resources, total, err := service.ListKnSrcs(ctx, parameter)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 3)
			So(len(resources), ShouldEqual, 2)
			So(resources[0].ID, ShouldEqual, "kn2")
		})
	})
}

func Test_knowledgeNetworkService_GetRelationTypePaths(t *testing.T) {
	Convey("Test GetRelationTypePaths\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := bmock.NewMockKNAccess(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ots:        ots,
		}

		Convey("Success getting relation type paths\n", func() {
			query := interfaces.RelationTypePathsBaseOnSource{
				KNID:              "kn1",
				Branch:            interfaces.MAIN_BRANCH,
				SourceObjecTypeId: "ot1",
				Direction:         "out",
				PathLength:        1,
			}
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "ot1",
				},
			}
			neighborPathsMap := map[string][]interfaces.RelationTypePath{
				"ot1": {
					{
						ObjectTypes: []interfaces.ObjectTypeWithKeyField{
							{OTID: "ot1"},
							{OTID: "ot2"},
						},
						TypeEdges: []interfaces.TypeEdge{
							{RelationTypeId: "rt1"},
						},
						Length: 1,
					},
				},
			}

			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&objectType, nil).AnyTimes()
			kna.EXPECT().GetNeighborPathsBatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(neighborPathsMap, nil)

			paths, err := service.GetRelationTypePaths(ctx, query)
			So(err, ShouldBeNil)
			So(len(paths), ShouldBeGreaterThanOrEqualTo, 0)
		})

		Convey("Success with path length 0\n", func() {
			query := interfaces.RelationTypePathsBaseOnSource{
				KNID:              "kn1",
				Branch:            interfaces.MAIN_BRANCH,
				SourceObjecTypeId: "ot1",
				Direction:         "out",
				PathLength:        0,
			}
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "ot1",
				},
			}

			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&objectType, nil).AnyTimes()

			paths, err := service.GetRelationTypePaths(ctx, query)
			So(err, ShouldBeNil)
			So(len(paths), ShouldEqual, 1)
		})

		Convey("Failed when GetObjectTypesByIDs returns error\n", func() {
			query := interfaces.RelationTypePathsBaseOnSource{
				KNID:              "kn1",
				Branch:            interfaces.MAIN_BRANCH,
				SourceObjecTypeId: "ot1",
				Direction:         "out",
				PathLength:        1,
			}

			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			paths, err := service.GetRelationTypePaths(ctx, query)
			So(err, ShouldNotBeNil)
			So(paths, ShouldBeNil)
		})

		Convey("Failed when GetNeighborPathsBatch returns error\n", func() {
			query := interfaces.RelationTypePathsBaseOnSource{
				KNID:              "kn1",
				Branch:            interfaces.MAIN_BRANCH,
				SourceObjecTypeId: "ot1",
				Direction:         "out",
				PathLength:        1,
			}
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "ot1",
				},
			}

			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&objectType, nil).AnyTimes()
			kna.EXPECT().GetNeighborPathsBatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))

			paths, err := service.GetRelationTypePaths(ctx, query)
			So(err, ShouldNotBeNil)
			So(paths, ShouldBeNil)
		})

		Convey("Success with no neighbors\n", func() {
			query := interfaces.RelationTypePathsBaseOnSource{
				KNID:              "kn1",
				Branch:            interfaces.MAIN_BRANCH,
				SourceObjecTypeId: "ot1",
				Direction:         "out",
				PathLength:        1,
			}
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "ot1",
				},
			}
			neighborPathsMap := map[string][]interfaces.RelationTypePath{
				"ot1": {},
			}

			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&objectType, nil).AnyTimes()
			kna.EXPECT().GetNeighborPathsBatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(neighborPathsMap, nil)

			paths, err := service.GetRelationTypePaths(ctx, query)
			So(err, ShouldBeNil)
			So(len(paths), ShouldEqual, 1)
		})
	})
}

func Test_knowledgeNetworkService_CreateKN(t *testing.T) {
	Convey("Test CreateKN\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		kna := bmock.NewMockKNAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		bss := bmock.NewMockBusinessSystemService(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			ps:         ps,
			vba:        vba,
			bss:        bss,
			db:         db,
		}

		Convey("Success creating KN with normal mode\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			bss.EXPECT().BindResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			knID, err := service.CreateKN(ctx, kn, mode, true)
			So(err, ShouldBeNil)
			So(knID, ShouldNotBeEmpty)
		})

		Convey("Failed when permission check fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_KnowledgeNetwork_InternalError))

			knID, err := service.CreateKN(ctx, kn, mode, true)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
		})

		Convey("Failed when KN ID already exists in normal mode\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", true, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			smock.ExpectRollback()

			knID, err := service.CreateKN(ctx, kn, mode, true)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_KnowledgeNetwork_KNIDExisted)
		})

		Convey("Success with empty KNID generates new ID\n", func() {
			kn := &interfaces.KN{
				KNID:   "",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx, knID, branch interface{}) {
				So(knID, ShouldNotBeEmpty)
			}).Return("", false, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			bss.EXPECT().BindResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			knID, err := service.CreateKN(ctx, kn, mode, true)
			So(err, ShouldBeNil)
			So(knID, ShouldNotBeEmpty)
		})

		Convey("Success with Ignore mode when KN exists\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Ignore

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", true, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			smock.ExpectCommit()

			knID, err := service.CreateKN(ctx, kn, mode, true)
			So(err, ShouldBeNil)
			So(knID, ShouldEqual, "kn1")
		})

		Convey("Success with Overwrite mode when ID exists\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Overwrite
			kna2 := bmock.NewMockKNAccess(mockCtrl)
			ots := bmock.NewMockObjectTypeService(mockCtrl)
			rts := bmock.NewMockRelationTypeService(mockCtrl)
			ats := bmock.NewMockActionTypeService(mockCtrl)
			cgs := bmock.NewMockConceptGroupService(mockCtrl)

			service := &knowledgeNetworkService{
				appSetting: appSetting,
				kna:        kna2,
				ps:         ps,
				vba:        vba,
				bss:        bss,
				db:         db,
				ots:        ots,
				rts:        rts,
				ats:        ats,
				cgs:        cgs,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			// CreateKN handleKNImportMode + UpdateKN(strict) → ValidateKN(..., Overwrite) → handleKNImportMode again
			kna2.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", true, nil).AnyTimes()
			kna2.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", true, nil).AnyTimes()
			kna2.EXPECT().UpdateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil).AnyTimes()
			smock.ExpectCommit()

			knID, err := service.CreateKN(ctx, kn, mode, true)
			So(err, ShouldBeNil)
			So(knID, ShouldEqual, "kn1")
		})

		Convey("Failed when Begin transaction fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			// 模拟Begin失败
			db2, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			db2.Close() // 关闭数据库连接以模拟Begin失败
			service2 := &knowledgeNetworkService{
				appSetting: appSetting,
				kna:        kna,
				ps:         ps,
				vba:        vba,
				bss:        bss,
				db:         db2,
			}

			knID, err := service2.CreateKN(ctx, kn, mode, true)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
		})

		Convey("Failed when CreateKN fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			knID, err := service.CreateKN(ctx, kn, mode, true)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
		})

		Convey("Failed when CreateConceptGroup fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
				ConceptGroups: []*interfaces.ConceptGroup{
					{
						CGID:   "cg1",
						CGName: "cg1",
					},
				},
			}
			mode := interfaces.ImportMode_Normal
			cgs := bmock.NewMockConceptGroupService(mockCtrl)

			service3 := &knowledgeNetworkService{
				appSetting: appSetting,
				kna:        kna,
				ps:         ps,
				vba:        vba,
				bss:        bss,
				db:         db,
				cgs:        cgs,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cgs.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			knID, err := service3.CreateKN(ctx, kn, mode, true)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
		})

		Convey("Failed when CreateObjectTypes fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
				ObjectTypes: []*interfaces.ObjectType{
					{
						ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
							OTID:   "ot1",
							OTName: "ot1",
						},
					},
				},
			}
			mode := interfaces.ImportMode_Normal
			ots := bmock.NewMockObjectTypeService(mockCtrl)

			service4 := &knowledgeNetworkService{
				appSetting: appSetting,
				kna:        kna,
				ps:         ps,
				vba:        vba,
				bss:        bss,
				db:         db,
				ots:        ots,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().CreateObjectTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			knID, err := service4.CreateKN(ctx, kn, mode, true)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
		})

		Convey("Failed when CreateRelationTypes fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
				RelationTypes: []*interfaces.RelationType{
					{
						RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
							RTID:   "rt1",
							RTName: "rt1",
						},
					},
				},
			}
			mode := interfaces.ImportMode_Normal
			rts := bmock.NewMockRelationTypeService(mockCtrl)

			service5 := &knowledgeNetworkService{
				appSetting: appSetting,
				kna:        kna,
				ps:         ps,
				vba:        vba,
				bss:        bss,
				db:         db,
				rts:        rts,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().CreateRelationTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			knID, err := service5.CreateKN(ctx, kn, mode, true)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
		})

		Convey("Failed when CreateActionTypes fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
				ActionTypes: []*interfaces.ActionType{
					{
						ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
							ATID:   "at1",
							ATName: "at1",
						},
					},
				},
			}
			mode := interfaces.ImportMode_Normal
			ats := bmock.NewMockActionTypeService(mockCtrl)

			service6 := &knowledgeNetworkService{
				appSetting: appSetting,
				kna:        kna,
				ps:         ps,
				vba:        vba,
				bss:        bss,
				db:         db,
				ats:        ats,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().CreateActionTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			knID, err := service6.CreateKN(ctx, kn, mode, true)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
		})

		Convey("Failed when InsertDatasetData fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			knID, err := service.CreateKN(ctx, kn, mode, true)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
		})

		Convey("Failed when CreateResources fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			knID, err := service.CreateKN(ctx, kn, mode, true)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
		})

		Convey("Failed when BindResource fails\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			kna.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			ps.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			bss.EXPECT().BindResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_KnowledgeNetwork_InternalError))
			smock.ExpectRollback()

			knID, err := service.CreateKN(ctx, kn, mode, true)
			So(err, ShouldNotBeNil)
			So(knID, ShouldEqual, "")
		})
	})
}

func Test_knowledgeNetworkService_ValidateKN_preflightBatch(t *testing.T) {
	Convey("ValidateKN passes BatchIDIndex from payload to ValidateConceptGroups\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := bmock.NewMockKNAccess(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)

		kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
		kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			cgs:        cgs,
		}

		kn := &interfaces.KN{
			KNID:   "kn1",
			Branch: interfaces.MAIN_BRANCH,
			ConceptGroups: []*interfaces.ConceptGroup{
				{
					CGID: "cg1",
					ObjectTypes: []*interfaces.ObjectType{
						{ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "ot_inner"}},
					},
				},
			},
		}

		cgs.EXPECT().ValidateConceptGroups(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, kn.ConceptGroups, true, gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, knID, br string, groups []*interfaces.ConceptGroup, sm bool, b *interfaces.BatchIDIndex, mode string) error {
				if b == nil || b.ObjectTypes["ot_inner"] == nil {
					return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_KnowledgeNetwork_InternalError)
				}
				return nil
			})

		err := service.ValidateKN(ctx, kn, true, interfaces.ImportMode_Normal)
		So(err, ShouldBeNil)
	})
}

func Test_knowledgeNetworkService_ValidateKN_strictMode_propagation(t *testing.T) {
	Convey("ValidateKN passes strictMode=false to all nested Validate calls\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		kna := bmock.NewMockKNAccess(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)

		kna.EXPECT().CheckKNExistByID(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
		kna.EXPECT().CheckKNExistByName(gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			kna:        kna,
			cgs:        cgs,
			ots:        ots,
			rts:        rts,
			ats:        ats,
		}

		kn := &interfaces.KN{
			KNID:   "kn1",
			Branch: interfaces.MAIN_BRANCH,
			ConceptGroups: []*interfaces.ConceptGroup{
				{CGID: "cg1"},
			},
			ObjectTypes: []*interfaces.ObjectType{
				{ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "ot1", OTName: "ot1"}},
			},
			RelationTypes: []*interfaces.RelationType{
				{RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{RTID: "rt1", RTName: "rt1"}},
			},
			ActionTypes: []*interfaces.ActionType{
				{ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{ATID: "at1", ATName: "at1"}},
			},
		}

		cgs.EXPECT().ValidateConceptGroups(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, kn.ConceptGroups, false, gomock.Any(), gomock.Any()).Return(nil)
		ots.EXPECT().ValidateObjectTypes(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, kn.ObjectTypes, false, gomock.Any(), gomock.Any()).Return(nil)
		rts.EXPECT().ValidateRelationTypes(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, kn.RelationTypes, false, gomock.Any(), gomock.Any()).Return(nil)
		ats.EXPECT().ValidateActionTypes(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, kn.ActionTypes, false, gomock.Any(), gomock.Any()).Return(nil)

		err := service.ValidateKN(ctx, kn, false, interfaces.ImportMode_Normal)
		So(err, ShouldBeNil)
	})
}
