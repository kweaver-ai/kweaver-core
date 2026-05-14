// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric

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

func Test_metricService_CheckMetricExistByID(t *testing.T) {
	Convey("Test CheckMetricExistByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ma := bmock.NewMockMetricAccess(mockCtrl)
		service := &metricService{
			appSetting: &common.AppSetting{},
			ma:         ma,
		}

		Convey("Success when metric exists\n", func() {
			ma.EXPECT().CheckMetricExistByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, "mid1").Return("mname1", true, nil)

			name, exist, err := service.CheckMetricExistByID(ctx, "kn1", interfaces.MAIN_BRANCH, "mid1")
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(name, ShouldEqual, "mname1")
		})

		Convey("Success when metric does not exist\n", func() {
			ma.EXPECT().CheckMetricExistByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, "mid1").Return("", false, nil)

			name, exist, err := service.CheckMetricExistByID(ctx, "kn1", interfaces.MAIN_BRANCH, "mid1")
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			ma.EXPECT().CheckMetricExistByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return("", false, sql.ErrConnDone)

			name, exist, err := service.CheckMetricExistByID(ctx, "kn1", interfaces.MAIN_BRANCH, "mid1")
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Metric_InternalError_CheckMetricIfExistFailed)
		})
	})
}

func Test_metricService_CheckMetricExistByName(t *testing.T) {
	Convey("Test CheckMetricExistByName\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ma := bmock.NewMockMetricAccess(mockCtrl)
		service := &metricService{
			appSetting: &common.AppSetting{},
			ma:         ma,
		}

		Convey("Success when metric exists\n", func() {
			ma.EXPECT().CheckMetricExistByName(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, "mname1").Return("mid1", true, nil)

			id, exist, err := service.CheckMetricExistByName(ctx, "kn1", interfaces.MAIN_BRANCH, "mname1")
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(id, ShouldEqual, "mid1")
		})

		Convey("Success when metric does not exist\n", func() {
			ma.EXPECT().CheckMetricExistByName(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, "mname1").Return("", false, nil)

			id, exist, err := service.CheckMetricExistByName(ctx, "kn1", interfaces.MAIN_BRANCH, "mname1")
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
		})

		Convey("Failed when access layer returns error\n", func() {
			ma.EXPECT().CheckMetricExistByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return("", false, sql.ErrConnDone)

			id, exist, err := service.CheckMetricExistByName(ctx, "kn1", interfaces.MAIN_BRANCH, "mname1")
			So(err, ShouldNotBeNil)
			So(exist, ShouldBeFalse)
			So(id, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Metric_InternalError_CheckMetricIfExistFailed)
		})
	})
}

func Test_metricService_GetMetricByID(t *testing.T) {
	Convey("Test GetMetricByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ma := bmock.NewMockMetricAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		service := &metricService{
			appSetting: &common.AppSetting{},
			ma:         ma,
			ps:         ps,
		}

		Convey("Success when metric found\n", func() {
			def := &interfaces.MetricDefinition{ID: "mid1", KnID: "kn1", Branch: interfaces.MAIN_BRANCH, Name: "n1"}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ma.EXPECT().GetMetricByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, "mid1").Return(def, nil)

			got, err := service.GetMetricByID(ctx, "kn1", interfaces.MAIN_BRANCH, "mid1")
			So(err, ShouldBeNil)
			So(got, ShouldEqual, def)
		})

		Convey("Failed when permission denied\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_InternalError_CheckPermissionFailed))

			got, err := service.GetMetricByID(ctx, "kn1", interfaces.MAIN_BRANCH, "mid1")
			So(err, ShouldNotBeNil)
			So(got, ShouldBeNil)
		})

		Convey("Failed when not found\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ma.EXPECT().GetMetricByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, sql.ErrNoRows)

			got, err := service.GetMetricByID(ctx, "kn1", interfaces.MAIN_BRANCH, "mid1")
			So(err, ShouldNotBeNil)
			So(got, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Metric_NotFound)
		})

		Convey("Failed when access returns other error\n", func() {
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ma.EXPECT().GetMetricByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, sql.ErrConnDone)

			got, err := service.GetMetricByID(ctx, "kn1", interfaces.MAIN_BRANCH, "mid1")
			So(err, ShouldNotBeNil)
			So(got, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Metric_InternalError)
		})
	})
}

func Test_metricService_GetMetricsByIDs(t *testing.T) {
	Convey("Test GetMetricsByIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ma := bmock.NewMockMetricAccess(mockCtrl)
		service := &metricService{
			appSetting: &common.AppSetting{},
			ma:         ma,
		}

		Convey("Success\n", func() {
			list := []*interfaces.MetricDefinition{{ID: "a", KnID: "kn1", Branch: interfaces.MAIN_BRANCH}}
			ma.EXPECT().GetMetricsByIDs(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, []string{"a"}).Return(list, nil)

			got, err := service.GetMetricsByIDs(ctx, "kn1", interfaces.MAIN_BRANCH, []string{"a", "a"})
			So(err, ShouldBeNil)
			So(len(got), ShouldEqual, 1)
		})

		Convey("Failed when access returns error\n", func() {
			ma.EXPECT().GetMetricsByIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, sql.ErrConnDone)

			got, err := service.GetMetricsByIDs(ctx, "kn1", interfaces.MAIN_BRANCH, []string{"x"})
			So(err, ShouldNotBeNil)
			So(got, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Metric_InternalError_GetMetricsByIDsFailed)
		})
	})
}

func Test_metricService_ListMetrics(t *testing.T) {
	Convey("Test ListMetrics\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ma := bmock.NewMockMetricAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		service := &metricService{
			appSetting: &common.AppSetting{},
			ma:         ma,
			ps:         ps,
		}

		Convey("Success with total and entries\n", func() {
			q := interfaces.MetricsListQueryParams{
				KNID: "kn1",
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Offset: 0,
					Limit:  -1,
				},
			}
			entries := []*interfaces.MetricDefinition{{ID: "m1", KnID: "kn1", Name: "n1"}}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ma.EXPECT().ListMetrics(gomock.Any(), q).Return(entries, nil)
			ma.EXPECT().GetMetricsTotal(gomock.Any(), q).Return(1, nil)

			out, err := service.ListMetrics(ctx, q)
			So(err, ShouldBeNil)
			So(out.TotalCount, ShouldEqual, 1)
			So(len(out.Entries), ShouldEqual, 1)
		})

		Convey("Pagination returns empty slice when offset beyond list length\n", func() {
			q := interfaces.MetricsListQueryParams{
				KNID: "kn1",
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Offset: 2,
					Limit:  10,
				},
			}
			entries := []*interfaces.MetricDefinition{
				{ID: "m1", KnID: "kn1"},
				{ID: "m2", KnID: "kn1"},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ma.EXPECT().ListMetrics(gomock.Any(), q).Return(entries, nil)
			ma.EXPECT().GetMetricsTotal(gomock.Any(), q).Return(99, nil)

			out, err := service.ListMetrics(ctx, q)
			So(err, ShouldBeNil)
			So(len(out.Entries), ShouldEqual, 0)
			So(out.TotalCount, ShouldEqual, 99)
		})

		Convey("Failed when permission denied\n", func() {
			q := interfaces.MetricsListQueryParams{KNID: "kn1"}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_InternalError_CheckPermissionFailed))

			out, err := service.ListMetrics(ctx, q)
			So(err, ShouldNotBeNil)
			So(out, ShouldBeNil)
		})

		Convey("Failed when ListMetrics returns error\n", func() {
			q := interfaces.MetricsListQueryParams{KNID: "kn1"}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ma.EXPECT().ListMetrics(gomock.Any(), gomock.Any()).Return(nil, sql.ErrConnDone)

			out, err := service.ListMetrics(ctx, q)
			So(err, ShouldNotBeNil)
			So(out, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Metric_InternalError)
		})
	})
}

func Test_metricService_UpdateMetric(t *testing.T) {
	Convey("Test UpdateMetric\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ma := bmock.NewMockMetricAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &metricService{
			appSetting: appSetting,
			db:         db,
			ma:         ma,
			ps:         ps,
			vba:        vba,
		}

		Convey("Failed when kn_id branch or id missing\n", func() {
			req := &interfaces.MetricDefinition{ID: "m1", Branch: interfaces.MAIN_BRANCH}
			err := service.UpdateMetric(ctx, nil, req, false)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Metric_InvalidParameter)
		})

		Convey("Success with external transaction\n", func() {
			smock.ExpectBegin()
			tx, errBegin := db.Begin()
			So(errBegin, ShouldBeNil)

			req := &interfaces.MetricDefinition{
				ID:     "mid1",
				KnID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				Name:   "n1",
				CommonInfo: interfaces.CommonInfo{
					Comment: "c",
				},
				CalculationFormula: &interfaces.MetricCalculationFormula{
					Aggregation: interfaces.MetricAggregation{Property: "p", Aggr: interfaces.MetricAggrSum},
				},
			}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ma.EXPECT().UpdateMetric(gomock.Any(), tx, gomock.Any()).Return(nil)
			vba.EXPECT().WriteDatasetDocuments(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil)

			err := service.UpdateMetric(ctx, tx, req, false)
			So(err, ShouldBeNil)
		})
	})
}

func Test_metricService_DeleteMetricsByIDs(t *testing.T) {
	Convey("Test DeleteMetricsByIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ma := bmock.NewMockMetricAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &metricService{
			appSetting: &common.AppSetting{},
			db:         db,
			ma:         ma,
			ps:         ps,
			vba:        vba,
		}

		Convey("No-op when metricIDs empty\n", func() {
			err := service.DeleteMetricsByIDs(ctx, nil, "kn1", interfaces.MAIN_BRANCH, nil)
			So(err, ShouldBeNil)
		})

		Convey("Success with external transaction\n", func() {
			smock.ExpectBegin()
			tx, errBegin := db.Begin()
			So(errBegin, ShouldBeNil)

			ids := []string{"a", "b"}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ma.EXPECT().DeleteMetricsByIDs(gomock.Any(), tx, "kn1", interfaces.MAIN_BRANCH, ids).Return(nil)
			vba.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), interfaces.BKN_DATASET_ID, gomock.Any()).Return(nil).Times(2)

			err := service.DeleteMetricsByIDs(ctx, tx, "kn1", interfaces.MAIN_BRANCH, ids)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission denied\n", func() {
			smock.ExpectBegin()
			tx, _ := db.Begin()
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_InternalError_CheckPermissionFailed))

			err := service.DeleteMetricsByIDs(ctx, tx, "kn1", interfaces.MAIN_BRANCH, []string{"x"})
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_metricService_SearchMetrics(t *testing.T) {
	Convey("Test SearchMetrics\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				DefaultSmallModelEnabled: false,
			},
		}
		ps := bmock.NewMockPermissionService(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		cga := bmock.NewMockConceptGroupAccess(mockCtrl)

		service := &metricService{
			appSetting: appSetting,
			ps:         ps,
			vba:        vba,
			cga:        cga,
		}

		Convey("permission denied\n", func() {
			q := &interfaces.ConceptsQuery{KNID: "kn1"}
			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_InternalError_CheckPermissionFailed))

			res, err := service.SearchMetrics(ctx, q)
			So(err, ShouldNotBeNil)
			So(res.Type, ShouldEqual, interfaces.MODULE_TYPE_METRIC)
		})

		Convey("Success without concept groups\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				Limit:  10,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{},
			}
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(datasetResp, nil)

			result, err := service.SearchMetrics(ctx, query)
			So(err, ShouldBeNil)
			So(result.Entries, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Success with concept groups\n", func() {
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
								Value:     "m1",
							},
						},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"ot1"}, nil)
			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{},
			}
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(datasetResp, nil)

			result, err := service.SearchMetrics(ctx, query)
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

			result, err := service.SearchMetrics(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("Success with concept groups returning empty otIDs\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				ConceptGroups: []string{"cg1"},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil)

			result, err := service.SearchMetrics(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Entries), ShouldEqual, 0)
		})

		Convey("filters entries by scope_ref when concept groups set\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				Limit:         10,
				ConceptGroups: []string{"cg1"},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"ot_keep"}, nil)

			inDoc := map[string]any{
				"id":          "metric_drop",
				"kn_id":       "kn1",
				"branch":      interfaces.MAIN_BRANCH,
				"module_type": interfaces.MODULE_TYPE_METRIC,
				"name":        "x",
				"scope_ref":   "ot_other",
			}
			keepDoc := map[string]any{
				"id":          "metric_keep",
				"kn_id":       "kn1",
				"branch":      interfaces.MAIN_BRANCH,
				"module_type": interfaces.MODULE_TYPE_METRIC,
				"name":        "y",
				"scope_ref":   "ot_keep",
			}

			datasetResp := &interfaces.DatasetQueryResponse{
				Entries: []map[string]any{inDoc, keepDoc},
			}
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(datasetResp, nil)

			result, err := service.SearchMetrics(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Entries), ShouldEqual, 1)
			So(result.Entries[0].ID, ShouldEqual, "metric_keep")
			So(result.Entries[0].ScopeRef, ShouldEqual, "ot_keep")
		})

		Convey("NeedTotal with concept groups uses batched scope_ref total\n", func() {
			query := &interfaces.ConceptsQuery{
				KNID:          "kn1",
				Branch:        interfaces.MAIN_BRANCH,
				NeedTotal:     true,
				Limit:         5,
				ConceptGroups: []string{"cg1"},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			cga.EXPECT().GetConceptGroupsTotal(gomock.Any(), gomock.Any()).Return(1, nil)
			cga.EXPECT().GetConceptIDsByConceptGroupIDs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"ot1"}, nil)

			totalResp := &interfaces.DatasetQueryResponse{
				TotalCount: 7,
				Entries:    []map[string]any{},
			}
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(totalResp, nil)
			emptyPage := &interfaces.DatasetQueryResponse{Entries: []map[string]any{}}
			vba.EXPECT().QueryResourceData(gomock.Any(), gomock.Any(), gomock.Any()).Return(emptyPage, nil)

			result, err := service.SearchMetrics(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 7)
			So(len(result.Entries), ShouldEqual, 0)
		})
	})
}

func Test_metricService_ValidateMetrics_strict_batch_scopeNotInPayload(t *testing.T) {
	Convey("ValidateMetrics strictMode + batch rejects unknown scope_ref without DB lookup\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ots := bmock.NewMockObjectTypeService(mockCtrl)
		service := &metricService{
			appSetting: &common.AppSetting{},
			ots:        ots,
		}

		batch := batchindex.NewBatchIDIndex("kn1", interfaces.MAIN_BRANCH)
		entries := []*interfaces.MetricDefinition{
			{
				ID:       "m1",
				ScopeRef: "no_such_ot",
				CalculationFormula: &interfaces.MetricCalculationFormula{
					Aggregation: interfaces.MetricAggregation{Property: "x", Aggr: interfaces.MetricAggrCount},
				},
			},
		}

		err := service.ValidateMetrics(ctx, entries, true, interfaces.ImportMode_Overwrite, batch)
		So(err, ShouldNotBeNil)
		httpErr := err.(*rest.HTTPError)
		So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
	})
}
