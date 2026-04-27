// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package risk_type

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

var (
	testUpdateTime = int64(1735786555379)
	testTags       = []string{"tag1", "tag2", "tag3"}

	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)

	testRiskType = &interfaces.RiskType{
		RTID:   "rt1",
		RTName: "Risk Type 1",
		CommonInfo: interfaces.CommonInfo{
			Tags:          testTags,
			Comment:       "test comment",
			Icon:          "icon1",
			Color:         "color1",
			BKNRawContent: "bkn1",
		},
		KNID:   "kn1",
		Branch: interfaces.MAIN_BRANCH,
		Creator: interfaces.AccountInfo{
			ID:   "admin",
			Type: "admin",
		},
		CreateTime: testUpdateTime,
		Updater: interfaces.AccountInfo{
			ID:   "admin",
			Type: "admin",
		},
		UpdateTime: testUpdateTime,
		ModuleType: interfaces.MODULE_TYPE_RISK_TYPE,
	}
)

func MockNewRiskTypeAccess(appSetting *common.AppSetting) (*riskTypeAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	rta := &riskTypeAccess{
		appSetting: appSetting,
		db:         db,
	}
	return rta, smock
}

// rtSelectCols 是 ListRiskTypes / GetRiskTypesByIDs 的 SELECT 列列表（与实现保持一致）
var rtSelectCols = []string{
	"f_id", "f_name", "f_comment", "f_tags", "f_icon", "f_color", "f_bkn_raw_content",
	"f_kn_id", "f_branch", "f_creator", "f_creator_type", "f_create_time",
	"f_updater", "f_updater_type", "f_update_time",
}

// addRTRow 向 Rows 追加一行风险类测试数据
func addRTRow(rows *sqlmock.Rows, id, name string) *sqlmock.Rows {
	return rows.AddRow(
		id, name, "test comment", `"tag1"`, "icon1", "color1", "bkn1",
		"kn1", "main", "admin", "admin", testUpdateTime,
		"admin", "admin", testUpdateTime,
	)
}

// ---- CheckRiskTypeExistByID ----

func Test_RiskTypeAccess_CheckRiskTypeExistByID(t *testing.T) {
	Convey("test CheckRiskTypeExistByID\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRiskTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_name FROM %s WHERE f_kn_id = ? AND f_branch = ? AND f_id = ?", RT_TABLE_NAME)
		knID, branch, rtID := "kn1", "main", "rt1"

		Convey("CheckRiskTypeExistByID Success\n", func() {
			rows := sqlmock.NewRows([]string{"f_name"}).AddRow("Risk Type 1")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtID).WillReturnRows(rows)

			name, exist, err := rta.CheckRiskTypeExistByID(testCtx, knID, branch, rtID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(name, ShouldEqual, "Risk Type 1")

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("CheckRiskTypeExistByID Not Found\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtID).WillReturnRows(sqlmock.NewRows(nil))

			name, exist, err := rta.CheckRiskTypeExistByID(testCtx, knID, branch, rtID)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(name, ShouldEqual, "")

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("CheckRiskTypeExistByID DB Error\n", func() {
			expectedErr := errors.New("db error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtID).WillReturnError(expectedErr)

			_, exist, err := rta.CheckRiskTypeExistByID(testCtx, knID, branch, rtID)
			So(err, ShouldResemble, expectedErr)
			So(exist, ShouldBeFalse)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// ---- CheckRiskTypeExistByName ----

func Test_RiskTypeAccess_CheckRiskTypeExistByName(t *testing.T) {
	Convey("test CheckRiskTypeExistByName\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRiskTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id FROM %s WHERE f_kn_id = ? AND f_branch = ? AND f_name = ?", RT_TABLE_NAME)
		knID, branch, rtName := "kn1", "main", "Risk Type 1"

		Convey("CheckRiskTypeExistByName Success\n", func() {
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("rt1")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtName).WillReturnRows(rows)

			rtID, exist, err := rta.CheckRiskTypeExistByName(testCtx, knID, branch, rtName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			So(rtID, ShouldEqual, "rt1")

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("CheckRiskTypeExistByName Not Found\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtName).WillReturnRows(sqlmock.NewRows(nil))

			rtID, exist, err := rta.CheckRiskTypeExistByName(testCtx, knID, branch, rtName)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			So(rtID, ShouldEqual, "")

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("CheckRiskTypeExistByName DB Error\n", func() {
			expectedErr := errors.New("db error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtName).WillReturnError(expectedErr)

			_, exist, err := rta.CheckRiskTypeExistByName(testCtx, knID, branch, rtName)
			So(err, ShouldResemble, expectedErr)
			So(exist, ShouldBeFalse)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// ---- CreateRiskType ----

func Test_RiskTypeAccess_CreateRiskType(t *testing.T) {
	Convey("test CreateRiskType\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRiskTypeAccess(appSetting)

		sqlStr := fmt.Sprintf(
			"INSERT INTO %s (f_id,f_name,f_comment,f_tags,f_icon,f_color,f_bkn_raw_content,f_kn_id,f_branch,"+
				"f_creator,f_creator_type,f_create_time,f_updater,f_updater_type,f_update_time) "+
				"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
			RT_TABLE_NAME,
		)

		Convey("CreateRiskType Success\n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := rta.db.Begin()
			err := rta.CreateRiskType(testCtx, tx, testRiskType)
			So(err, ShouldBeNil)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("CreateRiskType Exec Error\n", func() {
			expectedErr := errors.New("exec error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := rta.db.Begin()
			err := rta.CreateRiskType(testCtx, tx, testRiskType)
			So(err, ShouldResemble, expectedErr)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// ---- ListRiskTypes ----

func Test_RiskTypeAccess_ListRiskTypes(t *testing.T) {
	Convey("test ListRiskTypes\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRiskTypeAccess(appSetting)

		baseSelect := fmt.Sprintf(
			"SELECT f_id, f_name, f_comment, f_tags, f_icon, f_color, f_bkn_raw_content, f_kn_id, f_branch, "+
				"f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
				"FROM %s", RT_TABLE_NAME,
		)

		knID, branch := "kn1", "main"
		query := interfaces.RiskTypesQueryParams{KNID: knID, Branch: branch}

		rows := addRTRow(sqlmock.NewRows(rtSelectCols), "rt1", "Risk Type 1")

		Convey("ListRiskTypes Success\n", func() {
			sqlStr := baseSelect + " WHERE f_kn_id = ? AND f_branch = ?"
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			list, err := rta.ListRiskTypes(testCtx, query)
			So(err, ShouldBeNil)
			So(len(list), ShouldEqual, 1)
			So(list[0].RTID, ShouldEqual, "rt1")

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("ListRiskTypes No Rows\n", func() {
			sqlStr := baseSelect + " WHERE f_kn_id = ? AND f_branch = ?"
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(sqlmock.NewRows(nil))

			list, err := rta.ListRiskTypes(testCtx, query)
			So(err, ShouldBeNil)
			So(list, ShouldBeNil)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("ListRiskTypes with NamePattern\n", func() {
			sqlStr := baseSelect + " WHERE (instr(f_name, ?) > 0 OR instr(f_id, ?) > 0) AND f_kn_id = ? AND f_branch = ?"
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			q := interfaces.RiskTypesQueryParams{KNID: knID, Branch: branch, NamePattern: "Risk"}
			list, err := rta.ListRiskTypes(testCtx, q)
			So(err, ShouldBeNil)
			So(len(list), ShouldEqual, 1)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("ListRiskTypes with Tag\n", func() {
			sqlStr := baseSelect + " WHERE instr(f_tags, ?) > 0 AND f_kn_id = ? AND f_branch = ?"
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			q := interfaces.RiskTypesQueryParams{KNID: knID, Branch: branch, Tag: "tag1"}
			list, err := rta.ListRiskTypes(testCtx, q)
			So(err, ShouldBeNil)
			So(len(list), ShouldEqual, 1)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("ListRiskTypes with Sort ASC\n", func() {
			sqlStr := baseSelect + " WHERE f_kn_id = ? AND f_branch = ? ORDER BY f_name ASC"
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			q := interfaces.RiskTypesQueryParams{
				KNID:   knID,
				Branch: branch,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Sort:      "f_name",
					Direction: "ASC",
				},
			}
			list, err := rta.ListRiskTypes(testCtx, q)
			So(err, ShouldBeNil)
			So(len(list), ShouldEqual, 1)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("ListRiskTypes with Sort DESC\n", func() {
			sqlStr := baseSelect + " WHERE f_kn_id = ? AND f_branch = ? ORDER BY f_name DESC"
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			q := interfaces.RiskTypesQueryParams{
				KNID:   knID,
				Branch: branch,
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Sort:      "f_name",
					Direction: "DESC",
				},
			}
			list, err := rta.ListRiskTypes(testCtx, q)
			So(err, ShouldBeNil)
			So(len(list), ShouldEqual, 1)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("ListRiskTypes DB Error\n", func() {
			sqlStr := baseSelect + " WHERE f_kn_id = ? AND f_branch = ?"
			expectedErr := errors.New("db error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := rta.ListRiskTypes(testCtx, query)
			So(err, ShouldResemble, expectedErr)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("ListRiskTypes Scan Error\n", func() {
			sqlStr := baseSelect + " WHERE f_kn_id = ? AND f_branch = ?"
			badRows := sqlmock.NewRows([]string{"f_id"}).AddRow("rt1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(badRows)

			_, err := rta.ListRiskTypes(testCtx, query)
			So(err, ShouldNotBeNil)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// ---- GetRiskTypesTotal ----

func Test_RiskTypeAccess_GetRiskTypesTotal(t *testing.T) {
	Convey("test GetRiskTypesTotal\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRiskTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s WHERE f_kn_id = ? AND f_branch = ?", RT_TABLE_NAME)
		query := interfaces.RiskTypesQueryParams{KNID: "kn1", Branch: "main"}

		Convey("GetRiskTypesTotal Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(
				sqlmock.NewRows([]string{"count"}).AddRow(3),
			)

			total, err := rta.GetRiskTypesTotal(testCtx, query)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 3)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("GetRiskTypesTotal DB Error\n", func() {
			expectedErr := errors.New("db error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := rta.GetRiskTypesTotal(testCtx, query)
			So(err, ShouldResemble, expectedErr)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// ---- GetRiskTypesByIDs ----

func Test_RiskTypeAccess_GetRiskTypesByIDs(t *testing.T) {
	Convey("test GetRiskTypesByIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRiskTypeAccess(appSetting)

		sqlStr := fmt.Sprintf(
			"SELECT f_id, f_name, f_comment, f_tags, f_icon, f_color, f_bkn_raw_content, f_kn_id, f_branch, "+
				"f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
				"FROM %s WHERE f_kn_id = ? AND f_branch = ? AND f_id IN (?,?)",
			RT_TABLE_NAME,
		)

		knID, branch := "kn1", "main"
		rtIDs := []string{"rt1", "rt2"}

		rows := addRTRow(addRTRow(sqlmock.NewRows(rtSelectCols), "rt1", "Risk Type 1"), "rt2", "Risk Type 2")

		Convey("GetRiskTypesByIDs Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			list, err := rta.GetRiskTypesByIDs(testCtx, knID, branch, rtIDs)
			So(err, ShouldBeNil)
			So(len(list), ShouldEqual, 2)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("GetRiskTypesByIDs Empty IDs\n", func() {
			list, err := rta.GetRiskTypesByIDs(testCtx, knID, branch, []string{})
			So(err, ShouldBeNil)
			So(list, ShouldResemble, []*interfaces.RiskType{})
		})

		Convey("GetRiskTypesByIDs No Rows\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(sqlmock.NewRows(nil))

			list, err := rta.GetRiskTypesByIDs(testCtx, knID, branch, rtIDs)
			So(err, ShouldBeNil)
			So(list, ShouldBeNil)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("GetRiskTypesByIDs DB Error\n", func() {
			expectedErr := errors.New("db error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			list, err := rta.GetRiskTypesByIDs(testCtx, knID, branch, rtIDs)
			So(err, ShouldResemble, expectedErr)
			So(list, ShouldBeNil)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("GetRiskTypesByIDs Scan Error\n", func() {
			badRows := sqlmock.NewRows([]string{"f_id"}).AddRow("rt1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(badRows)

			_, err := rta.GetRiskTypesByIDs(testCtx, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// ---- UpdateRiskType ----

func Test_RiskTypeAccess_UpdateRiskType(t *testing.T) {
	Convey("test UpdateRiskType\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRiskTypeAccess(appSetting)

		// squirrel SetMap sorts keys alphabetically
		sqlStr := fmt.Sprintf(
			"UPDATE %s SET f_bkn_raw_content = ?, f_color = ?, f_comment = ?, f_icon = ?, f_name = ?, f_tags = ?, "+
				"f_update_time = ?, f_updater = ?, f_updater_type = ? "+
				"WHERE f_id = ? AND f_kn_id = ? AND f_branch = ?",
			RT_TABLE_NAME,
		)

		Convey("UpdateRiskType Success\n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := rta.db.Begin()
			err := rta.UpdateRiskType(testCtx, tx, testRiskType)
			So(err, ShouldBeNil)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("UpdateRiskType Exec Error\n", func() {
			expectedErr := errors.New("exec error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := rta.db.Begin()
			err := rta.UpdateRiskType(testCtx, tx, testRiskType)
			So(err, ShouldResemble, expectedErr)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// ---- DeleteRiskTypesByIDs ----

func Test_RiskTypeAccess_DeleteRiskTypesByIDs(t *testing.T) {
	Convey("test DeleteRiskTypesByIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRiskTypeAccess(appSetting)

		sqlStr := fmt.Sprintf(
			"DELETE FROM %s WHERE f_kn_id = ? AND f_branch = ? AND f_id IN (?,?)",
			RT_TABLE_NAME,
		)
		knID, branch := "kn1", "main"
		rtIDs := []string{"rt1", "rt2"}

		Convey("DeleteRiskTypesByIDs Success\n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(0, 2))

			tx, _ := rta.db.Begin()
			affected, err := rta.DeleteRiskTypesByIDs(testCtx, tx, knID, branch, rtIDs)
			So(err, ShouldBeNil)
			So(affected, ShouldEqual, 2)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("DeleteRiskTypesByIDs Empty IDs\n", func() {
			affected, err := rta.DeleteRiskTypesByIDs(testCtx, nil, knID, branch, []string{})
			So(err, ShouldBeNil)
			So(affected, ShouldEqual, 0)
		})

		Convey("DeleteRiskTypesByIDs Exec Error\n", func() {
			expectedErr := errors.New("exec error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := rta.db.Begin()
			_, err := rta.DeleteRiskTypesByIDs(testCtx, tx, knID, branch, rtIDs)
			So(err, ShouldResemble, expectedErr)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// ---- GetAllRiskTypesByKnID ----

func Test_RiskTypeAccess_GetAllRiskTypesByKnID(t *testing.T) {
	Convey("test GetAllRiskTypesByKnID\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRiskTypeAccess(appSetting)

		sqlStr := fmt.Sprintf(
			"SELECT f_id, f_name, f_comment, f_tags, f_icon, f_color, f_bkn_raw_content, f_kn_id, f_branch, "+
				"f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
				"FROM %s WHERE f_kn_id = ? AND f_branch = ?",
			RT_TABLE_NAME,
		)
		knID, branch := "kn1", "main"

		rows := addRTRow(addRTRow(sqlmock.NewRows(rtSelectCols), "rt1", "Risk Type 1"), "rt2", "Risk Type 2")

		Convey("GetAllRiskTypesByKnID Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			list, err := rta.GetAllRiskTypesByKnID(testCtx, knID, branch)
			So(err, ShouldBeNil)
			So(len(list), ShouldEqual, 2)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("GetAllRiskTypesByKnID DB Error\n", func() {
			expectedErr := errors.New("db error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnError(expectedErr)

			list, err := rta.GetAllRiskTypesByKnID(testCtx, knID, branch)
			So(err, ShouldResemble, expectedErr)
			So(list, ShouldBeNil)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

// ---- DeleteRiskTypesByKnID ----

func Test_RiskTypeAccess_DeleteRiskTypesByKnID(t *testing.T) {
	Convey("test DeleteRiskTypesByKnID\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRiskTypeAccess(appSetting)

		sqlStr := fmt.Sprintf(
			"DELETE FROM %s WHERE f_kn_id = ? AND f_branch = ?",
			RT_TABLE_NAME,
		)
		knID, branch := "kn1", "main"

		Convey("DeleteRiskTypesByKnID Success\n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(0, 3))

			tx, _ := rta.db.Begin()
			affected, err := rta.DeleteRiskTypesByKnID(testCtx, tx, knID, branch)
			So(err, ShouldBeNil)
			So(affected, ShouldEqual, 3)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("DeleteRiskTypesByKnID Exec Error\n", func() {
			expectedErr := errors.New("exec error")
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := rta.db.Begin()
			_, err := rta.DeleteRiskTypesByKnID(testCtx, tx, knID, branch)
			So(err, ShouldResemble, expectedErr)

			So(smock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}
