// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package concept_group

import (
	"context"
	"database/sql"
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

func Test_conceptGroupService_CheckConceptGroupExistByID(t *testing.T) {
	Convey("Test CheckConceptGroupExistByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
		}

		Convey("Success when concept group exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			cgName := "concept_group1"

			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(cgName, true, nil)

			name, exist, err := service.CheckConceptGroupExistByID(ctx, knID, branch, cgID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(name, ShouldEqual, cgName)
		})

		Convey("Success when concept group does not exist\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			name, exist, err := service.CheckConceptGroupExistByID(ctx, knID, branch, cgID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			name, exist, err := service.CheckConceptGroupExistByID(ctx, knID, branch, cgID)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ConceptGroup_InternalError_CheckConceptGroupIfExistFailed)
		})
	})
}

func Test_conceptGroupService_CheckConceptGroupExistByName(t *testing.T) {
	Convey("Test CheckConceptGroupExistByName\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
		}

		Convey("Success when concept group exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgName := "concept_group1"
			cgID := "cg1"

			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(cgID, true, nil)

			id, exist, err := service.CheckConceptGroupExistByName(ctx, knID, branch, cgName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(id, ShouldEqual, cgID)
		})

		Convey("Success when concept group does not exist\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgName := "concept_group1"

			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)

			id, exist, err := service.CheckConceptGroupExistByName(ctx, knID, branch, cgName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgName := "concept_group1"

			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			id, exist, err := service.CheckConceptGroupExistByName(ctx, knID, branch, cgName)
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ConceptGroup_InternalError_CheckConceptGroupIfExistFailed)
		})
	})
}

func Test_conceptGroupService_UpdateConceptGroupDetail(t *testing.T) {
	Convey("Test UpdateConceptGroupDetail\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
		}

		Convey("Success updating concept group detail\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			detail := "updated detail"

			cga.EXPECT().UpdateConceptGroupDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := service.UpdateConceptGroupDetail(ctx, knID, branch, cgID, detail)
			So(err, ShouldBeNil)
		})

		Convey("Failed when access layer returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			detail := "updated detail"

			cga.EXPECT().UpdateConceptGroupDetail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			err := service.UpdateConceptGroupDetail(ctx, knID, branch, cgID, detail)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ConceptGroup_InternalError)
		})
	})
}

func Test_conceptGroupService_GetStatByConceptGroup(t *testing.T) {
	Convey("Test GetStatByConceptGroup\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		rta := bmock.NewMockRelationTypeAccess(mockCtrl)
		ata := bmock.NewMockActionTypeAccess(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			rta:        rta,
			ata:        ata,
		}

		Convey("Success getting statistics\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			otIDs := []string{"ot1", "ot2"}

			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otIDs, nil)
			rta.EXPECT().GetRelationTypesTotal(gomock.Any(), gomock.Any()).Return(5, nil)
			ata.EXPECT().GetActionTypesTotal(gomock.Any(), gomock.Any()).Return(3, nil)

			stats, err := service.GetStatByConceptGroup(ctx, conceptGroup)
			So(err, ShouldBeNil)
			So(stats, ShouldNotBeNil)
			So(stats.OtTotal, ShouldEqual, 2)
			So(stats.RtTotal, ShouldEqual, 5)
			So(stats.AtTotal, ShouldEqual, 3)
		})

		Convey("Success with empty object types\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)

			stats, err := service.GetStatByConceptGroup(ctx, conceptGroup)
			So(err, ShouldBeNil)
			So(stats, ShouldNotBeNil)
			So(stats.OtTotal, ShouldEqual, 0)
			So(stats.RtTotal, ShouldEqual, 0)
			So(stats.AtTotal, ShouldEqual, 0)
		})

		Convey("Failed when getting concept IDs returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			stats, err := service.GetStatByConceptGroup(ctx, conceptGroup)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeNil)
		})

		Convey("Failed when GetRelationTypesTotal returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			otIDs := []string{"ot1"}

			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otIDs, nil)
			rta.EXPECT().GetRelationTypesTotal(gomock.Any(), gomock.Any()).Return(0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			stats, err := service.GetStatByConceptGroup(ctx, conceptGroup)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ConceptGroup_InternalError_GetRelationTypesTotalFailed)
		})

		Convey("Failed when GetActionTypesTotal returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			otIDs := []string{"ot1"}

			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otIDs, nil)
			rta.EXPECT().GetRelationTypesTotal(gomock.Any(), gomock.Any()).Return(5, nil)
			ata.EXPECT().GetActionTypesTotal(gomock.Any(), gomock.Any()).Return(0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			stats, err := service.GetStatByConceptGroup(ctx, conceptGroup)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ConceptGroup_InternalError_GetRelationTypesTotalFailed)
		})
	})
}

func Test_conceptGroupService_ListConceptGroups(t *testing.T) {
	Convey("Test ListConceptGroups\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		ums := bmock.NewMockUserMgmtService(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			ums:        ums,
		}

		Convey("Success listing concept groups\n", func() {
			query := interfaces.ConceptGroupsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			cgArr := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "cg1",
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return(cgArr, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil).AnyTimes()

			cgs, total, err := service.ListConceptGroups(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(cgs), ShouldEqual, 1)
		})

		Convey("Success with empty result\n", func() {
			query := interfaces.ConceptGroupsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return([]*interfaces.ConceptGroup{}, nil)

			cgs, total, err := service.ListConceptGroups(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(len(cgs), ShouldEqual, 0)
		})

		Convey("Failed when permission check fails\n", func() {
			query := interfaces.ConceptGroupsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_ConceptGroup_InternalError))

			cgs, total, err := service.ListConceptGroups(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(cgs), ShouldEqual, 0)
		})

		Convey("Failed when ListConceptGroups returns error\n", func() {
			query := interfaces.ConceptGroupsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			cgs, total, err := service.ListConceptGroups(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(cgs), ShouldEqual, 0)
		})

		Convey("Failed when GetAccountNames returns error\n", func() {
			query := interfaces.ConceptGroupsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			cgArr := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "cg1",
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return(cgArr, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			cgs, total, err := service.ListConceptGroups(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(cgs), ShouldEqual, 0)
		})

		Convey("Failed when GetStatByConceptGroup returns error\n", func() {
			query := interfaces.ConceptGroupsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			cgArr := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "cg1",
				},
			}
			rta := bmock.NewMockRelationTypeAccess(mockCtrl)
			ata := bmock.NewMockActionTypeAccess(mockCtrl)

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				ums:        ums,
				rta:        rta,
				ata:        ata,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return(cgArr, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			cgs, total, err := service.ListConceptGroups(ctx, query)
			So(err, ShouldNotBeNil)
			So(total, ShouldEqual, 0)
			So(len(cgs), ShouldEqual, 0)
		})

		Convey("Success with Limit = -1\n", func() {
			query := interfaces.ConceptGroupsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  -1,
					Offset: 0,
				},
			}
			cgArr := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "cg1",
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return(cgArr, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)

			cgs, total, err := service.ListConceptGroups(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(cgs), ShouldEqual, 1)
		})

		Convey("Success with Offset out of bounds\n", func() {
			query := interfaces.ConceptGroupsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 100,
				},
			}
			cgArr := []*interfaces.ConceptGroup{
				{
					CGID:   "cg1",
					CGName: "cg1",
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			cga.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return(cgArr, nil)

			cgs, total, err := service.ListConceptGroups(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(len(cgs), ShouldEqual, 0)
		})

		Convey("Success with pagination\n", func() {
			query := interfaces.ConceptGroupsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  2,
					Offset: 1,
				},
			}
			cgArr := []*interfaces.ConceptGroup{
				{CGID: "cg1", CGName: "cg1"},
				{CGID: "cg2", CGName: "cg2"},
				{CGID: "cg3", CGName: "cg3"},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return(cgArr, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil).AnyTimes()

			cgs, total, err := service.ListConceptGroups(ctx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 3)
			So(len(cgs), ShouldEqual, 2)
			So(cgs[0].CGID, ShouldEqual, "cg2")
		})
	})
}

func Test_conceptGroupService_GetConceptGroupByID(t *testing.T) {
	Convey("Test GetConceptGroupByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
		}

		Convey("Success getting concept group by ID\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			mode := ""
			cg := &interfaces.ConceptGroup{
				CGID:   cgID,
				CGName: "cg1",
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(cg, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)

			result, err := service.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.CGID, ShouldEqual, cgID)
		})

		Convey("Failed when concept group not found\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			mode := ""

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

			result, err := service.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ConceptGroup_ConceptGroupNotFound)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			mode := ""

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_ConceptGroup_InternalError))

			result, err := service.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("Failed when GetConceptGroupByID returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			mode := ""

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			result, err := service.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ConceptGroup_InternalError_GetConceptGroupByIDFailed)
		})

		Convey("Failed when GetConceptIDsByConceptGroupIDs returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			mode := ""
			cg := &interfaces.ConceptGroup{
				CGID:   cgID,
				CGName: "cg1",
				KNID:   knID,
				Branch: branch,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(cg, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			result, err := service.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Success with object types\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			mode := ""
			cg := &interfaces.ConceptGroup{
				CGID:   cgID,
				CGName: "cg1",
				KNID:   knID,
				Branch: branch,
			}
			otIDs := []string{"ot1"}
			ots := bmock.NewMockObjectTypeService(mockCtrl)
			rts := bmock.NewMockRelationTypeService(mockCtrl)
			ats := bmock.NewMockActionTypeService(mockCtrl)

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				ots:        ots,
				rts:        rts,
				ats:        ats,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(cg, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otIDs, nil)
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.ObjectType{}, 0, nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, 0, nil)
			ats.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{}, 0, nil)

			result, err := service.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.CGID, ShouldEqual, cgID)
		})

		Convey("Failed when ListObjectTypes returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			mode := ""
			cg := &interfaces.ConceptGroup{
				CGID:   cgID,
				CGName: "cg1",
				KNID:   knID,
				Branch: branch,
			}
			otIDs := []string{"ot1"}
			ots := bmock.NewMockObjectTypeService(mockCtrl)

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				ots:        ots,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(cg, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otIDs, nil)
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			result, err := service.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Failed when ListRelationTypes returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			mode := ""
			cg := &interfaces.ConceptGroup{
				CGID:   cgID,
				CGName: "cg1",
				KNID:   knID,
				Branch: branch,
			}
			otIDs := []string{"ot1"}
			ots := bmock.NewMockObjectTypeService(mockCtrl)
			rts := bmock.NewMockRelationTypeService(mockCtrl)

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				ots:        ots,
				rts:        rts,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(cg, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otIDs, nil)
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.ObjectType{}, 0, nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return(nil, 0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			result, err := service.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Failed when ListActionTypes returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			mode := ""
			cg := &interfaces.ConceptGroup{
				CGID:   cgID,
				CGName: "cg1",
				KNID:   knID,
				Branch: branch,
			}
			otIDs := []string{"ot1"}
			ots := bmock.NewMockObjectTypeService(mockCtrl)
			rts := bmock.NewMockRelationTypeService(mockCtrl)
			ats := bmock.NewMockActionTypeService(mockCtrl)

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				ots:        ots,
				rts:        rts,
				ats:        ats,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(cg, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(otIDs, nil)
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.ObjectType{}, 0, nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, 0, nil)
			ats.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return(nil, 0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			result, err := service.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func Test_conceptGroupService_InsertDatasetData(t *testing.T) {
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

		service := &conceptGroupService{
			appSetting: appSetting,
			vba:        vba,
		}

		Convey("Success inserting dataset data\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := service.InsertDatasetData(ctx, conceptGroup)
			So(err, ShouldBeNil)
		})

		Convey("Failed when WriteDatasetDocuments returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			err := service.InsertDatasetData(ctx, conceptGroup)
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

			serviceWithVector := &conceptGroupService{
				appSetting: appSettingWithVector,
				vba:        vbaWithVector,
				mfa:        mfa,
			}

			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				CommonInfo: interfaces.CommonInfo{
					Tags:          []string{"tag1"},
					Comment:       "comment",
					BKNRawContent: "bkn",
				},
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			vectors := []*cond.VectorResp{
				{
					Vector: []float32{0.1, 0.2, 0.3},
				},
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(vectors, nil)
			vbaWithVector.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := serviceWithVector.InsertDatasetData(ctx, conceptGroup)
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetDefaultModel returns error with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &conceptGroupService{
				appSetting: appSettingWithVector,
				mfa:        mfa,
			}

			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			err := serviceWithVector.InsertDatasetData(ctx, conceptGroup)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetVector returns error with vector enabled\n", func() {
			appSettingWithVector := &common.AppSetting{
				ServerSetting: common.ServerSetting{
					DefaultSmallModelEnabled: true,
				},
			}
			mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

			serviceWithVector := &conceptGroupService{
				appSetting: appSettingWithVector,
				mfa:        mfa,
			}

			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			mfa.EXPECT().GetDefaultModel(gomock.Any()).Return(&interfaces.SmallModel{ModelID: "model1"}, nil)
			mfa.EXPECT().GetVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))

			err := serviceWithVector.InsertDatasetData(ctx, conceptGroup)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_conceptGroupService_DeleteConceptGroupByID(t *testing.T) {
	Convey("Test DeleteConceptGroupByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			vba:        vba,
			db:         db,
		}

		Convey("Success deleting concept group\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			vba.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := service.DeleteConceptGroupByID(ctx, nil, knID, branch, cgID)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_ConceptGroup_InternalError))

			err := service.DeleteConceptGroupByID(ctx, nil, knID, branch, cgID)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteConceptGroupByID returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			err := service.DeleteConceptGroupByID(ctx, nil, knID, branch, cgID)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteObjectTypesFromGroup returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			err := service.DeleteConceptGroupByID(ctx, nil, knID, branch, cgID)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteDatasetDocumentByID returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			vba.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			err := service.DeleteConceptGroupByID(ctx, nil, knID, branch, cgID)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with rowsAffect != 1\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteConceptGroupByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			vba.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := service.DeleteConceptGroupByID(ctx, nil, knID, branch, cgID)
			So(err, ShouldBeNil)
		})
	})
}

func Test_conceptGroupService_UpdateConceptGroup(t *testing.T) {
	Convey("Test UpdateConceptGroup\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			vba:        vba,
			db:         db,
		}

		Convey("Success updating concept group\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().UpdateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			smock.ExpectCommit()

			err := service.UpdateConceptGroup(ctx, nil, conceptGroup)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_ConceptGroup_InternalError))

			err := service.UpdateConceptGroup(ctx, nil, conceptGroup)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when UpdateConceptGroup returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().UpdateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			err := service.UpdateConceptGroup(ctx, nil, conceptGroup)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when InsertDatasetData returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().UpdateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			err := service.UpdateConceptGroup(ctx, nil, conceptGroup)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_conceptGroupService_ListConceptGroupRelations(t *testing.T) {
	Convey("Test ListConceptGroupRelations\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			db:         db,
		}

		Convey("Success listing concept group relations\n", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				KNID:        "kn1",
				Branch:      interfaces.MAIN_BRANCH,
				CGIDs:       []string{"cg1"},
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}
			cgrArr := []interfaces.ConceptGroupRelation{
				{
					ID:          "cgr1",
					CGID:        "cg1",
					ConceptID:   "ot1",
					ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return(cgrArr, nil)
			smock.ExpectCommit()

			result, err := service.ListConceptGroupRelations(ctx, query)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("Success with empty result\n", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				KNID:        "kn1",
				Branch:      interfaces.MAIN_BRANCH,
				CGIDs:       []string{"cg1"},
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 0,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.ConceptGroupRelation{}, nil)
			smock.ExpectCommit()

			result, err := service.ListConceptGroupRelations(ctx, query)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when permission check fails\n", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				KNID:        "kn1",
				Branch:      interfaces.MAIN_BRANCH,
				CGIDs:       []string{"cg1"},
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_ConceptGroup_InternalError))

			result, err := service.ListConceptGroupRelations(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when ListConceptGroupRelations returns error\n", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				KNID:        "kn1",
				Branch:      interfaces.MAIN_BRANCH,
				CGIDs:       []string{"cg1"},
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			result, err := service.ListConceptGroupRelations(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Success with Limit = -1\n", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				KNID:        "kn1",
				Branch:      interfaces.MAIN_BRANCH,
				CGIDs:       []string{"cg1"},
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  -1,
					Offset: 0,
				},
			}
			cgrArr := []interfaces.ConceptGroupRelation{
				{
					ID:          "cgr1",
					CGID:        "cg1",
					ConceptID:   "ot1",
					ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return(cgrArr, nil)
			smock.ExpectCommit()

			result, err := service.ListConceptGroupRelations(ctx, query)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("Success with Offset out of bounds\n", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				KNID:        "kn1",
				Branch:      interfaces.MAIN_BRANCH,
				CGIDs:       []string{"cg1"},
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  10,
					Offset: 100,
				},
			}
			cgrArr := []interfaces.ConceptGroupRelation{
				{
					ID:          "cgr1",
					CGID:        "cg1",
					ConceptID:   "ot1",
					ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return(cgrArr, nil)
			smock.ExpectCommit()

			result, err := service.ListConceptGroupRelations(ctx, query)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Success with pagination\n", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				KNID:        "kn1",
				Branch:      interfaces.MAIN_BRANCH,
				CGIDs:       []string{"cg1"},
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Limit:  2,
					Offset: 1,
				},
			}
			cgrArr := []interfaces.ConceptGroupRelation{
				{ID: "cgr1", CGID: "cg1", ConceptID: "ot1", ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE},
				{ID: "cgr2", CGID: "cg1", ConceptID: "ot2", ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE},
				{ID: "cgr3", CGID: "cg1", ConceptID: "ot3", ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE},
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return(cgrArr, nil)
			smock.ExpectCommit()

			result, err := service.ListConceptGroupRelations(ctx, query)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
			So(result[0].ID, ShouldEqual, "cgr2")
		})
	})
}

func Test_conceptGroupService_DeleteObjectTypesFromGroup(t *testing.T) {
	Convey("Test DeleteObjectTypesFromGroup\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			db:         db,
		}

		Convey("Success deleting object types from group\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []string{"ot1", "ot2"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(2), nil)
			smock.ExpectCommit()

			err := service.DeleteObjectTypesFromGroup(ctx, nil, knID, branch, cgID, otIDs)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []string{"ot1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_ConceptGroup_InternalError))

			err := service.DeleteObjectTypesFromGroup(ctx, nil, knID, branch, cgID, otIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteObjectTypesFromGroup returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []string{"ot1"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			err := service.DeleteObjectTypesFromGroup(ctx, nil, knID, branch, cgID, otIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with rowsAffect != len(otIDs)\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []string{"ot1", "ot2"}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			smock.ExpectCommit()

			err := service.DeleteObjectTypesFromGroup(ctx, nil, knID, branch, cgID, otIDs)
			So(err, ShouldBeNil)
		})
	})
}

func Test_conceptGroupService_CreateConceptGroup(t *testing.T) {
	Convey("Test CreateConceptGroup\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ps:         ps,
			vba:        vba,
			db:         db,
		}

		Convey("Success creating concept group with normal mode\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			smock.ExpectCommit()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldBeNil)
			So(cgID, ShouldNotBeEmpty)
		})

		Convey("Failed when permission check fails\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_ConceptGroup_InternalError))

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldNotBeNil)
			So(cgID, ShouldEqual, "")
		})

		Convey("Failed when concept group ID already exists in normal mode\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("cg1", true, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			smock.ExpectRollback()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldNotBeNil)
			So(cgID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ConceptGroup_ConceptGroupIDExisted)
		})

		Convey("Success with empty CGID generates new ID\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx, knID, branch, cgID interface{}) {
				So(cgID, ShouldNotBeEmpty)
			}).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			smock.ExpectCommit()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldBeNil)
			So(cgID, ShouldNotBeEmpty)
		})

		Convey("Success with Ignore mode when concept group exists\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Ignore

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("cg1", true, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			smock.ExpectCommit()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldBeNil)
			So(cgID, ShouldEqual, "cg1")
		})

		Convey("Success with Overwrite mode when ID exists\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Overwrite

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("cg1", true, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("cg1", true, nil)
			cga.EXPECT().UpdateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil).AnyTimes()
			smock.ExpectCommit()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldBeNil)
			So(cgID, ShouldEqual, "cg1")
		})

		Convey("Failed when CreateConceptGroup returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldNotBeNil)
			So(cgID, ShouldEqual, "")
		})

		Convey("Failed when InsertDatasetData returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			mode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldNotBeNil)
			So(cgID, ShouldEqual, "")
		})

		Convey("Success creating concept group with ObjectTypes\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
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

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				vba:        vba,
				db:         db,
				ots:        ots,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]*interfaces.ObjectType{{ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "ot1"}}}, 1, nil).AnyTimes()
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().CreateObjectTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"ot1"}, nil).AnyTimes()
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.ConceptGroupRelation{}, nil).AnyTimes()
			cga.EXPECT().CreateConceptGroupRelation(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil).AnyTimes()
			smock.ExpectCommit()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldBeNil)
			So(cgID, ShouldNotBeEmpty)
		})

		Convey("Failed when CreateObjectTypes returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
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

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				vba:        vba,
				db:         db,
				ots:        ots,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().CreateObjectTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldNotBeNil)
			So(cgID, ShouldEqual, "")
		})

		Convey("Failed when AddObjectTypesToConceptGroup returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
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

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				vba:        vba,
				db:         db,
				ots:        ots,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ots.EXPECT().CreateObjectTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"ot1"}, nil)
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldNotBeNil)
			So(cgID, ShouldEqual, "")
		})

		Convey("Success creating concept group with RelationTypes\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
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

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				vba:        vba,
				db:         db,
				rts:        rts,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().CreateRelationTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"rt1"}, nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			smock.ExpectCommit()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldBeNil)
			So(cgID, ShouldNotBeEmpty)
		})

		Convey("Failed when CreateRelationTypes returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
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

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				vba:        vba,
				db:         db,
				rts:        rts,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().CreateRelationTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldNotBeNil)
			So(cgID, ShouldEqual, "")
		})

		Convey("Success creating concept group with ActionTypes\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
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

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				vba:        vba,
				db:         db,
				ats:        ats,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().CreateActionTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"at1"}, nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)
			smock.ExpectCommit()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldBeNil)
			So(cgID, ShouldNotBeEmpty)
		})

		Convey("Failed when CreateActionTypes returns error\n", func() {
			conceptGroup := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "cg1",
				KNID:   "kn1",
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

			service := &conceptGroupService{
				appSetting: appSetting,
				cga:        cga,
				ps:         ps,
				vba:        vba,
				db:         db,
				ats:        ats,
			}

			smock.ExpectBegin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().CheckConceptGroupExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CheckConceptGroupExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", false, nil)
			cga.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ats.EXPECT().CreateActionTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			cgID, err := service.CreateConceptGroup(ctx, nil, conceptGroup, mode, true)
			So(err, ShouldNotBeNil)
			So(cgID, ShouldEqual, "")
		})
	})
}

func Test_conceptGroupService_AddObjectTypesToConceptGroup(t *testing.T) {
	Convey("Test AddObjectTypesToConceptGroup\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &conceptGroupService{
			appSetting: appSetting,
			cga:        cga,
			ots:        ots,
			db:         db,
		}

		Convey("Success adding object types to concept group\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []interfaces.ID{
				{ID: "ot1"},
				{ID: "ot2"},
			}
			importMode := interfaces.ImportMode_Normal
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot2",
						OTName: "ot2",
					},
				},
			}

			smock.ExpectBegin()
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, 2, nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.ConceptGroupRelation{}, nil)
			cga.EXPECT().CreateConceptGroupRelation(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
			smock.ExpectCommit()

			cgrIDs, err := service.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, otIDs, importMode)
			So(err, ShouldBeNil)
			So(len(cgrIDs), ShouldEqual, 2)
		})

		Convey("Failed when object types not found\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []interfaces.ID{
				{ID: "ot1"},
			}
			importMode := interfaces.ImportMode_Normal
			objectTypes := []*interfaces.ObjectType{}

			smock.ExpectBegin()
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, 0, nil)
			smock.ExpectRollback()

			cgrIDs, err := service.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, otIDs, importMode)
			So(err, ShouldNotBeNil)
			So(len(cgrIDs), ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ConceptGroup_ObjectTypeNotFound)
		})

		Convey("Failed when relation already exists in normal mode\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []interfaces.ID{
				{ID: "ot1"},
			}
			importMode := interfaces.ImportMode_Normal
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
			}
			cgRelations := []interfaces.ConceptGroupRelation{
				{
					ID:          "cgr1",
					CGID:        "cg1",
					ConceptID:   "ot1",
					ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				},
			}

			smock.ExpectBegin()
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, 1, nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return(cgRelations, nil)
			smock.ExpectRollback()

			cgrIDs, err := service.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, otIDs, importMode)
			So(err, ShouldNotBeNil)
			So(len(cgrIDs), ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ConceptGroup_ConceptGroupRelationExisted)
		})

		Convey("Failed when ListObjectTypes returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []interfaces.ID{
				{ID: "ot1"},
			}
			importMode := interfaces.ImportMode_Normal

			smock.ExpectBegin()
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			cgrIDs, err := service.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, otIDs, importMode)
			So(err, ShouldNotBeNil)
			So(len(cgrIDs), ShouldEqual, 0)
		})

		Convey("Failed when ListConceptGroupRelations returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []interfaces.ID{
				{ID: "ot1"},
			}
			importMode := interfaces.ImportMode_Normal
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
			}

			smock.ExpectBegin()
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, 1, nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			cgrIDs, err := service.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, otIDs, importMode)
			So(err, ShouldNotBeNil)
			So(len(cgrIDs), ShouldEqual, 0)
		})

		Convey("Failed when CreateConceptGroupRelation returns error\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []interfaces.ID{
				{ID: "ot1"},
			}
			importMode := interfaces.ImportMode_Normal
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
			}

			smock.ExpectBegin()
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, 1, nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.ConceptGroupRelation{}, nil)
			cga.EXPECT().CreateConceptGroupRelation(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 500, berrors.BknBackend_ConceptGroup_InternalError))
			smock.ExpectRollback()

			cgrIDs, err := service.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, otIDs, importMode)
			So(err, ShouldNotBeNil)
			So(len(cgrIDs), ShouldEqual, 0)
		})

		Convey("Success with Ignore mode when relation exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []interfaces.ID{
				{ID: "ot1"},
				{ID: "ot2"},
			}
			importMode := interfaces.ImportMode_Ignore
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot2",
						OTName: "ot2",
					},
				},
			}
			cgRelations := []interfaces.ConceptGroupRelation{
				{
					ID:          "cgr1",
					CGID:        "cg1",
					ConceptID:   "ot1",
					ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				},
			}

			smock.ExpectBegin()
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, 2, nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return(cgRelations, nil)
			cga.EXPECT().CreateConceptGroupRelation(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			cgrIDs, err := service.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, otIDs, importMode)
			So(err, ShouldBeNil)
			So(len(cgrIDs), ShouldEqual, 1)
		})

		Convey("Success with Overwrite mode when relation exists\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			cgID := "cg1"
			otIDs := []interfaces.ID{
				{ID: "ot1"},
				{ID: "ot2"},
			}
			importMode := interfaces.ImportMode_Overwrite
			objectTypes := []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "ot1",
					},
				},
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot2",
						OTName: "ot2",
					},
				},
			}
			cgRelations := []interfaces.ConceptGroupRelation{
				{
					ID:          "cgr1",
					CGID:        "cg1",
					ConceptID:   "ot1",
					ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				},
			}

			smock.ExpectBegin()
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, 2, nil)
			cga.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any(), gomock.Any()).Return(cgRelations, nil)
			cga.EXPECT().CreateConceptGroupRelation(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()

			cgrIDs, err := service.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, otIDs, importMode)
			So(err, ShouldBeNil)
			So(len(cgrIDs), ShouldEqual, 1)
		})
	})
}

func Test_conceptGroupService_DeleteConceptGroupsByKnID(t *testing.T) {
	Convey("Test DeleteConceptGroupsByKnID\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		cga := bmock.NewMockConceptGroupAccess(mockCtrl)
		service := &conceptGroupService{appSetting: &common.AppSetting{}, cga: cga}

		knID := "kn1"
		branch := interfaces.MAIN_BRANCH

		Convey("Failed when tx is nil\n", func() {
			err := service.DeleteConceptGroupsByKnID(context.Background(), nil, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteConceptGroupsByKnID access returns error\n", func() {
			tx := new(sql.Tx)
			cga.EXPECT().DeleteConceptGroupsByKnID(gomock.Any(), tx, knID, branch).Return(int64(0), rest.NewHTTPError(context.Background(), 500, berrors.BknBackend_ConceptGroup_InternalError))
			err := service.DeleteConceptGroupsByKnID(context.Background(), tx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteConceptGroupRelationsByKnID access returns error\n", func() {
			tx := new(sql.Tx)
			cga.EXPECT().DeleteConceptGroupsByKnID(gomock.Any(), tx, knID, branch).Return(int64(1), nil)
			cga.EXPECT().DeleteConceptGroupRelationsByKnID(gomock.Any(), tx, knID, branch).Return(int64(0), rest.NewHTTPError(context.Background(), 500, berrors.BknBackend_ConceptGroup_InternalError))
			err := service.DeleteConceptGroupsByKnID(context.Background(), tx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Success\n", func() {
			tx := new(sql.Tx)
			cga.EXPECT().DeleteConceptGroupsByKnID(gomock.Any(), tx, knID, branch).Return(int64(2), nil)
			cga.EXPECT().DeleteConceptGroupRelationsByKnID(gomock.Any(), tx, knID, branch).Return(int64(5), nil)
			err := service.DeleteConceptGroupsByKnID(context.Background(), tx, knID, branch)
			So(err, ShouldBeNil)
		})
	})
}
