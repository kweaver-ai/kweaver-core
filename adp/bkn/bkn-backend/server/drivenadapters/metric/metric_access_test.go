// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

func mockMetricAccess(t *testing.T) (*metricAccess, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	ma := &metricAccess{appSetting: &common.AppSetting{}, db: db}
	cleanup := func() { _ = db.Close() }
	return ma, mock, cleanup
}

func TestMetricAccess_CreateMetric(t *testing.T) {
	Convey("CreateMetric inserts one row", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		def := &interfaces.MetricDefinition{
			ID:     "mid1",
			KnID:   "kn1",
			Branch: interfaces.MAIN_BRANCH,
			Name:   "m1",
			CommonInfo: interfaces.CommonInfo{
				Comment: "c",
			},
			UnitType:   "",
			Unit:       "",
			MetricType: interfaces.MetricTypeAtomic,
			ScopeType:  interfaces.ScopeTypeObjectType,
			ScopeRef:   "ot1",
			CalculationFormula: &interfaces.MetricCalculationFormula{
				Aggregation: interfaces.MetricAggregation{Property: "x", Aggr: interfaces.MetricAggrCount},
			},
			Creator:    interfaces.AccountInfo{ID: "u1", Type: "user"},
			Updater:    interfaces.AccountInfo{ID: "u1", Type: "user"},
			CreateTime: 1,
			UpdateTime: 1,
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO " + METRIC_TABLE_NAME).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := ma.db.Begin()
		So(err, ShouldBeNil)
		err = ma.CreateMetric(context.Background(), tx, def)
		So(err, ShouldBeNil)
		err = tx.Commit()
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestMetricAccess_GetMetricByID(t *testing.T) {
	Convey("GetMetricByID returns row", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		cols := metricSelectColumns()
		rows := sqlmock.NewRows(cols).
			AddRow("mid1", "kn1", interfaces.MAIN_BRANCH, "m1", "", "", "", "", "", "", "", interfaces.MetricTypeAtomic,
				interfaces.ScopeTypeObjectType, "ot1", nil, `{"aggregation":{"property":"x","aggr":"count"}}`, nil,
				"u1", "user", int64(1), "u1", "user", int64(1))

		mock.ExpectQuery("SELECT (.+) FROM "+METRIC_TABLE_NAME).
			WithArgs("kn1", interfaces.MAIN_BRANCH, "mid1").
			WillReturnRows(rows)

		def, err := ma.GetMetricByID(context.Background(), "kn1", interfaces.MAIN_BRANCH, "mid1")
		So(err, ShouldBeNil)
		So(def.ID, ShouldEqual, "mid1")
		So(def.Name, ShouldEqual, "m1")
		So(def.CalculationFormula.Aggregation.Property, ShouldEqual, "x")
		So(def.CalculationFormula.Aggregation.Aggr, ShouldEqual, interfaces.MetricAggrCount)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestMetricAccess_GetMetricByID_timeDimensionLegacyFieldJSON(t *testing.T) {
	Convey("GetMetricByID maps legacy time_dimension JSON key field to Property", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		cols := metricSelectColumns()
		rows := sqlmock.NewRows(cols).
			AddRow("mid1", "kn1", interfaces.MAIN_BRANCH, "m1", "", "", "", "", "", "", "", interfaces.MetricTypeAtomic,
				interfaces.ScopeTypeObjectType, "ot1", `{"field":"@timestamp","default_range_policy":"last_1h"}`,
				`{"aggregation":{"property":"x","aggr":"count"}}`, nil,
				"u1", "user", int64(1), "u1", "user", int64(1))

		mock.ExpectQuery("SELECT (.+) FROM "+METRIC_TABLE_NAME).
			WithArgs("kn1", interfaces.MAIN_BRANCH, "mid1").
			WillReturnRows(rows)

		def, err := ma.GetMetricByID(context.Background(), "kn1", interfaces.MAIN_BRANCH, "mid1")
		So(err, ShouldBeNil)
		So(def.TimeDimension, ShouldNotBeNil)
		So(def.TimeDimension.Property, ShouldEqual, "@timestamp")
		So(def.TimeDimension.DefaultRangePolicy, ShouldEqual, interfaces.MetricTimeDefaultRangePolicyLast1h)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestMetricAccess_CheckMetricExistByID(t *testing.T) {
	Convey("CheckMetricExistByID when row exists", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		rows := sqlmock.NewRows([]string{"f_name"}).AddRow("m1")
		mock.ExpectQuery("SELECT (.+) FROM "+METRIC_TABLE_NAME).
			WithArgs("kn1", interfaces.MAIN_BRANCH, "mid1").
			WillReturnRows(rows)

		name, exist, err := ma.CheckMetricExistByID(context.Background(), "kn1", interfaces.MAIN_BRANCH, "mid1")
		So(err, ShouldBeNil)
		So(exist, ShouldBeTrue)
		So(name, ShouldEqual, "m1")
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})

	Convey("CheckMetricExistByID when row missing", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		rows := sqlmock.NewRows([]string{"f_name"})
		mock.ExpectQuery("SELECT (.+) FROM "+METRIC_TABLE_NAME).
			WithArgs("kn1", interfaces.MAIN_BRANCH, "mid1").
			WillReturnRows(rows)

		name, exist, err := ma.CheckMetricExistByID(context.Background(), "kn1", interfaces.MAIN_BRANCH, "mid1")
		So(err, ShouldBeNil)
		So(exist, ShouldBeFalse)
		So(name, ShouldEqual, "")
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestMetricAccess_CheckMetricExistByName(t *testing.T) {
	Convey("CheckMetricExistByName when row exists", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		rows := sqlmock.NewRows([]string{"f_id"}).AddRow("mid1")
		mock.ExpectQuery("SELECT (.+) FROM "+METRIC_TABLE_NAME).
			WithArgs("kn1", interfaces.MAIN_BRANCH, "m1").
			WillReturnRows(rows)

		id, exist, err := ma.CheckMetricExistByName(context.Background(), "kn1", interfaces.MAIN_BRANCH, "m1")
		So(err, ShouldBeNil)
		So(exist, ShouldBeTrue)
		So(id, ShouldEqual, "mid1")
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})

	Convey("CheckMetricExistByName when row missing", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		rows := sqlmock.NewRows([]string{"f_id"})
		mock.ExpectQuery("SELECT (.+) FROM "+METRIC_TABLE_NAME).
			WithArgs("kn1", interfaces.MAIN_BRANCH, "m1").
			WillReturnRows(rows)

		id, exist, err := ma.CheckMetricExistByName(context.Background(), "kn1", interfaces.MAIN_BRANCH, "m1")
		So(err, ShouldBeNil)
		So(exist, ShouldBeFalse)
		So(id, ShouldEqual, "")
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestMetricAccess_GetMetricsByIDs(t *testing.T) {
	Convey("GetMetricsByIDs returns multiple rows", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		cols := metricSelectColumns()
		rows := sqlmock.NewRows(cols).
			AddRow("mid1", "kn1", interfaces.MAIN_BRANCH, "m1", "", "", "", "", "", "", "", interfaces.MetricTypeAtomic,
				interfaces.ScopeTypeObjectType, "ot1", nil, `{"aggregation":{"property":"x","aggr":"count"}}`, nil,
				"u1", "user", int64(1), "u1", "user", int64(1)).
			AddRow("mid2", "kn1", interfaces.MAIN_BRANCH, "m2", "", "", "", "", "", "", "", interfaces.MetricTypeAtomic,
				interfaces.ScopeTypeObjectType, "ot1", nil, `{"aggregation":{"property":"y","aggr":"sum"}}`, nil,
				"u1", "user", int64(1), "u1", "user", int64(1))

		mock.ExpectQuery("SELECT (.+) FROM "+METRIC_TABLE_NAME).
			WithArgs("kn1", interfaces.MAIN_BRANCH, "mid1", "mid2").
			WillReturnRows(rows)

		list, err := ma.GetMetricsByIDs(context.Background(), "kn1", interfaces.MAIN_BRANCH, []string{"mid1", "mid2"})
		So(err, ShouldBeNil)
		So(len(list), ShouldEqual, 2)
		So(list[0].ID, ShouldEqual, "mid1")
		So(list[1].ID, ShouldEqual, "mid2")
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})

	Convey("GetMetricsByIDs empty ids returns empty slice", t, func() {
		ma, _, cleanup := mockMetricAccess(t)
		defer cleanup()

		list, err := ma.GetMetricsByIDs(context.Background(), "kn1", interfaces.MAIN_BRANCH, []string{})
		So(err, ShouldBeNil)
		So(len(list), ShouldEqual, 0)
	})
}

func TestMetricAccess_UpdateMetric(t *testing.T) {
	Convey("UpdateMetric executes UPDATE", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		m := &interfaces.MetricDefinition{
			ID:     "mid1",
			KnID:   "kn1",
			Branch: interfaces.MAIN_BRANCH,
			CommonInfo: interfaces.CommonInfo{
				Comment: "c2",
			},
			UnitType:   "numUnit",
			Unit:       "none",
			MetricType: interfaces.MetricTypeAtomic,
			CalculationFormula: &interfaces.MetricCalculationFormula{
				Aggregation: interfaces.MetricAggregation{Property: "x", Aggr: interfaces.MetricAggrCount},
			},
			Updater:    interfaces.AccountInfo{ID: "u2", Type: "user"},
			UpdateTime: 99,
		}

		mock.ExpectBegin()
		tx, err := ma.db.Begin()
		So(err, ShouldBeNil)
		mock.ExpectExec("UPDATE " + METRIC_TABLE_NAME).
			WillReturnResult(sqlmock.NewResult(0, 1)) // RowsAffected 1
		err = ma.UpdateMetric(context.Background(), tx, m)
		So(err, ShouldBeNil)
		mock.ExpectCommit()
		err = tx.Commit()
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestMetricAccess_DeleteMetricsByIDs(t *testing.T) {
	Convey("DeleteMetricsByIDs executes DELETE", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		mock.ExpectBegin()
		tx, err := ma.db.Begin()
		So(err, ShouldBeNil)
		mock.ExpectExec("DELETE FROM "+METRIC_TABLE_NAME).
			WithArgs("kn1", interfaces.MAIN_BRANCH, "mid1", "mid2").
			WillReturnResult(sqlmock.NewResult(0, 2))

		err = ma.DeleteMetricsByIDs(context.Background(), tx, "kn1", interfaces.MAIN_BRANCH, []string{"mid1", "mid2"})
		So(err, ShouldBeNil)
		mock.ExpectCommit()
		err = tx.Commit()
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestMetricAccess_ListMetrics_and_GetMetricsTotal(t *testing.T) {
	Convey("ListMetrics returns rows with filters", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		cols := metricSelectColumns()
		rows := sqlmock.NewRows(cols).
			AddRow("mid1", "kn1", interfaces.MAIN_BRANCH, "m1", "", "", "", "", "", "", "", interfaces.MetricTypeAtomic,
				interfaces.ScopeTypeObjectType, "ot1", nil, `{"aggregation":{"property":"x","aggr":"count"}}`, nil,
				"u1", "user", int64(1), "u1", "user", int64(1))

		mock.ExpectQuery("SELECT (.+) FROM "+METRIC_TABLE_NAME).
			WithArgs("kn1", interfaces.MAIN_BRANCH, interfaces.ScopeTypeObjectType, "ot1").
			WillReturnRows(rows)

		q := interfaces.MetricsListQueryParams{
			KNID:      "kn1",
			Branch:    interfaces.MAIN_BRANCH,
			ScopeType: interfaces.ScopeTypeObjectType,
			ScopeRef:  "ot1",
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Offset:    0,
				Limit:     -1,
				Sort:      "f_update_time",
				Direction: interfaces.DESC_DIRECTION,
			},
		}
		list, err := ma.ListMetrics(context.Background(), q)
		So(err, ShouldBeNil)
		So(len(list), ShouldEqual, 1)
		So(list[0].ID, ShouldEqual, "mid1")
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})

	Convey("GetMetricsTotal returns count", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		countRows := sqlmock.NewRows([]string{"COUNT(f_id)"}).AddRow(3)
		mock.ExpectQuery("SELECT (.+) FROM "+METRIC_TABLE_NAME).
			WithArgs("kn1", interfaces.MAIN_BRANCH).
			WillReturnRows(countRows)

		total, err := ma.GetMetricsTotal(context.Background(), interfaces.MetricsListQueryParams{
			KNID:   "kn1",
			Branch: interfaces.MAIN_BRANCH,
		})
		So(err, ShouldBeNil)
		So(total, ShouldEqual, 3)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestMetricAccess_GetMetricByID_notFound(t *testing.T) {
	Convey("GetMetricByID returns sql.ErrNoRows when missing", t, func() {
		ma, mock, cleanup := mockMetricAccess(t)
		defer cleanup()

		rows := sqlmock.NewRows(metricSelectColumns())
		mock.ExpectQuery("SELECT (.+) FROM "+METRIC_TABLE_NAME).
			WithArgs("kn1", interfaces.MAIN_BRANCH, "missing").
			WillReturnRows(rows)

		def, err := ma.GetMetricByID(context.Background(), "kn1", interfaces.MAIN_BRANCH, "missing")
		So(err, ShouldEqual, sql.ErrNoRows)
		So(def, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}
