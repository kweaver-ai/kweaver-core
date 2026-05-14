// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package concept_group

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

var (
	testUpdateTime = int64(1735786555379)
	testTags       = []string{"tag1", "tag2", "tag3"}

	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)

	testConceptGroup = &interfaces.ConceptGroup{
		CGID:   "cg1",
		CGName: "Concept Group 1",
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
		ModuleType: interfaces.MODULE_TYPE_CONCEPT_GROUP,
	}
)

func MockNewConceptGroupAccess(appSetting *common.AppSetting) (*conceptGroupAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	cga := &conceptGroupAccess{
		appSetting: appSetting,
		db:         db,
	}
	return cga, smock
}

func Test_conceptGroupAccess_CheckConceptGroupExistByID(t *testing.T) {
	Convey("test CheckConceptGroupExistByID\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := "SELECT f_name FROM t_concept_group WHERE f_id = ? AND f_kn_id = ? AND f_branch = ?"

		knID := "kn1"
		branch := "main"
		cgID := "cg1"

		Convey("CheckConceptGroupExistByID Success \n", func() {
			rows := sqlmock.NewRows([]string{"f_name"}).AddRow("Concept Group 1")
			smock.ExpectQuery(sqlStr).WithArgs(cgID, knID, branch).WillReturnRows(rows)

			name, exists, err := cga.CheckConceptGroupExistByID(testCtx, knID, branch, cgID)
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(name, ShouldEqual, "Concept Group 1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckConceptGroupExistByID Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(cgID, knID, branch).WillReturnError(sql.ErrNoRows)

			name, exists, err := cga.CheckConceptGroupExistByID(testCtx, knID, branch, cgID)
			So(name, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckConceptGroupExistByID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(cgID, knID, branch).WillReturnError(expectedErr)

			name, exists, err := cga.CheckConceptGroupExistByID(testCtx, knID, branch, cgID)
			So(name, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_CheckConceptGroupExistByName(t *testing.T) {
	Convey("test CheckConceptGroupExistByName\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := "SELECT f_id FROM t_concept_group WHERE f_name = ? AND f_kn_id = ? AND f_branch = ?"

		knID := "kn1"
		branch := "main"
		name := "Concept Group 1"

		Convey("CheckConceptGroupExistByName Success \n", func() {
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("cg1")
			smock.ExpectQuery(sqlStr).WithArgs(name, knID, branch).WillReturnRows(rows)

			cgID, exists, err := cga.CheckConceptGroupExistByName(testCtx, knID, branch, name)
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(cgID, ShouldEqual, "cg1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckConceptGroupExistByName Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(name, knID, branch).WillReturnError(sql.ErrNoRows)

			cgID, exists, err := cga.CheckConceptGroupExistByName(testCtx, knID, branch, name)
			So(cgID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckConceptGroupExistByName Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(name, knID, branch).WillReturnError(expectedErr)

			cgID, exists, err := cga.CheckConceptGroupExistByName(testCtx, knID, branch, name)
			So(cgID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_CreateConceptGroup(t *testing.T) {
	Convey("test CreateConceptGroup\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id,f_name,f_tags,f_comment,f_icon,f_color,f_bkn_raw_content,"+
			"f_kn_id,f_branch,f_creator,f_creator_type,f_create_time,f_updater,f_updater_type,f_update_time) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", CONCEPT_GROUP_TABLE_NAME)

		Convey("CreateConceptGroup Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := cga.db.Begin()
			err := cga.CreateConceptGroup(testCtx, tx, testConceptGroup)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateConceptGroup Exec sql error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := cga.db.Begin()
			err := cga.CreateConceptGroup(testCtx, tx, testConceptGroup)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_ListConceptGroups(t *testing.T) {
	Convey("test ListConceptGroups\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
			"f_kn_id, f_branch, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_TABLE_NAME)

		rows := sqlmock.NewRows([]string{
			"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
			"f_kn_id", "f_branch", "f_creator", "f_creator_type", "f_create_time",
			"f_updater", "f_updater_type", "f_update_time",
		}).AddRow(
			"cg1", "Concept Group 1", `"tag1"`, "comment", "icon", "color", "detail",
			"kn1", "main", "admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		)

		query := interfaces.ConceptGroupsQueryParams{
			KNID:   "kn1",
			Branch: "main",
		}

		Convey("ListConceptGroups Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			conceptGroups, err := cga.ListConceptGroups(testCtx, query)
			So(err, ShouldBeNil)
			So(len(conceptGroups), ShouldEqual, 1)
			So(conceptGroups[0].CGID, ShouldEqual, "cg1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListConceptGroups Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(sqlmock.NewRows(nil))

			conceptGroups, err := cga.ListConceptGroups(testCtx, query)
			So(conceptGroups, ShouldResemble, []*interfaces.ConceptGroup{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListConceptGroups Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			conceptGroups, err := cga.ListConceptGroups(testCtx, query)
			So(conceptGroups, ShouldResemble, []*interfaces.ConceptGroup{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListConceptGroups with Sort ASC\n", func() {
			sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
				"f_kn_id, f_branch, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
				"FROM %s WHERE f_kn_id = ? AND f_branch = ? ORDER BY f_name ASC", CONCEPT_GROUP_TABLE_NAME)

			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"cg1", "Concept Group 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			queryWithSort := interfaces.ConceptGroupsQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Sort:      "f_name",
					Direction: "ASC",
				},
				KNID:   "kn1",
				Branch: "main",
			}

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			conceptGroups, err := cga.ListConceptGroups(testCtx, queryWithSort)
			So(err, ShouldBeNil)
			So(len(conceptGroups), ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListConceptGroups scan error\n", func() {
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("cg1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := cga.ListConceptGroups(testCtx, query)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_GetConceptGroupsTotal(t *testing.T) {
	Convey("test GetConceptGroupsTotal\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s WHERE f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"COUNT(f_id)"}).AddRow(1)

		query := interfaces.ConceptGroupsQueryParams{
			KNID:   "kn1",
			Branch: "main",
		}

		Convey("GetConceptGroupsTotal Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := cga.GetConceptGroupsTotal(testCtx, query)
			So(total, ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetConceptGroupsTotal Failed  Query error\n", func() {
			expectedErr := errors.New("Query error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := cga.GetConceptGroupsTotal(testCtx, query)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_GetConceptGroupsByIDs(t *testing.T) {
	Convey("test GetConceptGroupsByIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
			"f_kn_id, f_branch, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_id IN (?,?) AND f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_TABLE_NAME)

		rows := sqlmock.NewRows([]string{
			"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
			"f_kn_id", "f_branch", "f_creator", "f_creator_type", "f_create_time",
			"f_updater", "f_updater_type", "f_update_time",
		}).AddRow(
			"cg1", "Concept Group 1", `"tag1"`, "comment", "icon", "color", "detail",
			"kn1", "main", "admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		).AddRow(
			"cg2", "Concept Group 2", `"tag2"`, "comment2", "icon2", "color2", "detail2",
			"kn1", "main", "admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		)

		knID := "kn1"
		branch := "main"
		cgIDs := []string{"cg1", "cg2"}

		Convey("GetConceptGroupsByIDs Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tx, _ := cga.db.Begin()
			conceptGroups, err := cga.GetConceptGroupsByIDs(testCtx, tx, knID, branch, cgIDs)
			So(len(conceptGroups), ShouldEqual, 2)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetConceptGroupsByIDs Failed \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := cga.db.Begin()
			conceptGroups, err := cga.GetConceptGroupsByIDs(testCtx, tx, knID, branch, cgIDs)
			So(conceptGroups, ShouldResemble, []*interfaces.ConceptGroup{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetConceptGroupsByIDs scan error\n", func() {
			smock.ExpectBegin()
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("cg1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tx, _ := cga.db.Begin()
			_, err := cga.GetConceptGroupsByIDs(testCtx, tx, knID, branch, cgIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_GetConceptGroupByID(t *testing.T) {
	Convey("test GetConceptGroupByID\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
			"f_kn_id, f_branch, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_id = ? AND f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_TABLE_NAME)

		knID := "kn1"
		branch := "main"
		cgID := "cg1"

		Convey("GetConceptGroupByID Success \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"cg1", "Concept Group 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			smock.ExpectQuery(sqlStr).WithArgs(cgID, knID, branch).WillReturnRows(rows)

			conceptGroup, err := cga.GetConceptGroupByID(testCtx, knID, branch, cgID)
			So(err, ShouldBeNil)
			So(conceptGroup, ShouldNotBeNil)
			So(conceptGroup.CGID, ShouldEqual, "cg1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetConceptGroupByID Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(cgID, knID, branch).WillReturnError(sql.ErrNoRows)

			conceptGroup, err := cga.GetConceptGroupByID(testCtx, knID, branch, cgID)
			So(conceptGroup, ShouldBeNil)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetConceptGroupByID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(cgID, knID, branch).WillReturnError(expectedErr)

			conceptGroup, err := cga.GetConceptGroupByID(testCtx, knID, branch, cgID)
			So(conceptGroup, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_UpdateConceptGroup(t *testing.T) {
	Convey("Test UpdateConceptGroup\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_bkn_raw_content = ?, f_color = ?, f_comment = ?, f_icon = ?, f_name = ?, "+
			"f_tags = ?, f_update_time = ?, f_updater = ?, f_updater_type = ? "+
			"WHERE f_id = ? AND f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_TABLE_NAME)

		conceptGroup := &interfaces.ConceptGroup{
			CGID:   "cg1",
			CGName: "Updated Concept Group",
			CommonInfo: interfaces.CommonInfo{
				Tags:    testTags,
				Comment: "updated comment",
				Icon:    "icon1",
				Color:   "color1",
			},
			KNID:   "kn1",
			Branch: "main",
			Updater: interfaces.AccountInfo{
				ID:   "admin",
				Type: "admin",
			},
			UpdateTime: testUpdateTime,
		}

		Convey("UpdateConceptGroup Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := cga.db.Begin()
			err := cga.UpdateConceptGroup(testCtx, tx, conceptGroup)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateConceptGroup failed prepare \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("prepare error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := cga.db.Begin()
			err := cga.UpdateConceptGroup(testCtx, tx, conceptGroup)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateConceptGroup Failed RowsAffected \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := cga.db.Begin()
			err := cga.UpdateConceptGroup(testCtx, tx, conceptGroup)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateConceptGroup RowsAffected 2\n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := cga.db.Begin()
			err := cga.UpdateConceptGroup(testCtx, tx, conceptGroup)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_UpdateConceptGroupDetail(t *testing.T) {
	Convey("Test UpdateConceptGroupDetail\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_bkn_raw_content = ? WHERE f_id = ? AND f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_TABLE_NAME)

		knID := "kn1"
		branch := "main"
		cgID := "cg1"
		detail := "updated detail"

		Convey("UpdateConceptGroupDetail Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs(detail, cgID, knID, branch).WillReturnResult(sqlmock.NewResult(1, 1))

			err := cga.UpdateConceptGroupDetail(testCtx, knID, branch, cgID, detail)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateConceptGroupDetail Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs(detail, cgID, knID, branch).WillReturnError(expectedErr)

			err := cga.UpdateConceptGroupDetail(testCtx, knID, branch, cgID, detail)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateConceptGroupDetail RowsAffected error\n", func() {
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs(detail, cgID, knID, branch).WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			err := cga.UpdateConceptGroupDetail(testCtx, knID, branch, cgID, detail)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateConceptGroupDetail RowsAffected not 1\n", func() {
			smock.ExpectExec(sqlStr).WithArgs(detail, cgID, knID, branch).WillReturnResult(sqlmock.NewResult(1, 2))

			err := cga.UpdateConceptGroupDetail(testCtx, knID, branch, cgID, detail)
			So(err, ShouldBeNil) // 注意：代码中 RowsAffected != 1 时只记录警告，不返回错误

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_DeleteConceptGroupByID(t *testing.T) {
	Convey("Test DeleteConceptGroupByID\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_id = ? AND f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_TABLE_NAME)

		knID := "kn1"
		branch := "main"
		cgID := "cg1"

		Convey("DeleteConceptGroupByID Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs(cgID, knID, branch).WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := cga.db.Begin()
			rowsAffected, err := cga.DeleteConceptGroupByID(testCtx, tx, knID, branch, cgID)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteConceptGroupByID null \n", func() {
			smock.ExpectBegin()

			tx, _ := cga.db.Begin()
			rowsAffected, err := cga.DeleteConceptGroupByID(testCtx, tx, knID, branch, "")
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 0)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteConceptGroupByID Failed dbExec \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("dbExec error")
			smock.ExpectExec(sqlStr).WithArgs(cgID, knID, branch).WillReturnError(expectedErr)

			tx, _ := cga.db.Begin()
			_, err := cga.DeleteConceptGroupByID(testCtx, tx, knID, branch, cgID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteConceptGroupByID RowsAffected error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs(cgID, knID, branch).WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := cga.db.Begin()
			_, err := cga.DeleteConceptGroupByID(testCtx, tx, knID, branch, cgID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_GetConceptGroupIDsByKnID(t *testing.T) {
	Convey("test GetConceptGroupIDsByKnID\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id FROM %s WHERE f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_id"}).
			AddRow("cg1").
			AddRow("cg2").
			AddRow("cg3")

		knID := "kn1"
		branch := "main"

		Convey("GetConceptGroupIDsByKnID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			cgIDs, err := cga.GetConceptGroupIDsByKnID(testCtx, knID, branch)
			So(err, ShouldBeNil)
			So(len(cgIDs), ShouldEqual, 3)
			So(cgIDs[0], ShouldEqual, "cg1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetConceptGroupIDsByKnID Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(sqlmock.NewRows(nil))

			cgIDs, err := cga.GetConceptGroupIDsByKnID(testCtx, knID, branch)
			So(cgIDs, ShouldResemble, []string{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetConceptGroupIDsByKnID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnError(expectedErr)

			cgIDs, err := cga.GetConceptGroupIDsByKnID(testCtx, knID, branch)
			So(cgIDs, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_GetAllConceptGroupsByKnID(t *testing.T) {
	Convey("test GetAllConceptGroupsByKnID\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
			"f_kn_id, f_branch, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_TABLE_NAME)

		rows := sqlmock.NewRows([]string{
			"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
			"f_kn_id", "f_branch", "f_creator", "f_creator_type", "f_create_time",
			"f_updater", "f_updater_type", "f_update_time",
		}).AddRow(
			"cg1", "Concept Group 1", `"tag1"`, "comment", "icon", "color", "detail",
			"kn1", "main", "admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		).AddRow(
			"cg2", "Concept Group 2", `"tag2"`, "comment2", "icon2", "color2", "detail2",
			"kn1", "main", "admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		)

		knID := "kn1"
		branch := "main"

		Convey("GetAllConceptGroupsByKnID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			conceptGroups, err := cga.GetAllConceptGroupsByKnID(testCtx, knID, branch)
			So(err, ShouldBeNil)
			So(len(conceptGroups), ShouldEqual, 2)
			So(conceptGroups["cg1"], ShouldNotBeNil)
			So(conceptGroups["cg2"], ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllConceptGroupsByKnID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnError(expectedErr)

			conceptGroups, err := cga.GetAllConceptGroupsByKnID(testCtx, knID, branch)
			So(conceptGroups, ShouldResemble, map[string]*interfaces.ConceptGroup{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllConceptGroupsByKnID scan error\n", func() {
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("cg1")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			_, err := cga.GetAllConceptGroupsByKnID(testCtx, knID, branch)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_ListConceptGroupRelations(t *testing.T) {
	Convey("test ListConceptGroupRelations\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_kn_id, f_branch, f_group_id, f_concept_type, f_concept_id, f_create_time "+
			"FROM %s WHERE f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_RELATION_TABLE_NAME)

		rows := sqlmock.NewRows([]string{
			"f_id", "f_kn_id", "f_branch", "f_group_id", "f_concept_type", "f_concept_id", "f_create_time",
		}).AddRow(
			"cgr1", "kn1", "main", "cg1", interfaces.MODULE_TYPE_OBJECT_TYPE, "ot1", testUpdateTime,
		).AddRow(
			"cgr2", "kn1", "main", "cg1", interfaces.MODULE_TYPE_OBJECT_TYPE, "ot2", testUpdateTime,
		)

		query := interfaces.ConceptGroupRelationsQueryParams{
			KNID:   "kn1",
			Branch: "main",
		}

		Convey("ListConceptGroupRelations Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tx, _ := cga.db.Begin()
			relations, err := cga.ListConceptGroupRelations(testCtx, tx, query)
			So(err, ShouldBeNil)
			So(len(relations), ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListConceptGroupRelations Failed \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := cga.db.Begin()
			relations, err := cga.ListConceptGroupRelations(testCtx, tx, query)
			So(relations, ShouldResemble, []interfaces.ConceptGroupRelation{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListConceptGroupRelations scan error\n", func() {
			smock.ExpectBegin()
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("cgr1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			tx, _ := cga.db.Begin()
			_, err := cga.ListConceptGroupRelations(testCtx, tx, query)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_CreateConceptGroupRelation(t *testing.T) {
	Convey("test CreateConceptGroupRelation\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id,f_kn_id,f_branch,f_group_id,f_concept_type,f_concept_id,f_create_time) "+
			"VALUES (?,?,?,?,?,?,?)", CONCEPT_GROUP_RELATION_TABLE_NAME)

		conceptGroupRelation := &interfaces.ConceptGroupRelation{
			ID:          "cgr1",
			KNID:        "kn1",
			Branch:      "main",
			CGID:        "cg1",
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			ConceptID:   "ot1",
			CreateTime:  testUpdateTime,
			ModuleType:  interfaces.MODULE_TYPE_CONCEPT_GROUP_RELATION,
		}

		Convey("CreateConceptGroupRelation Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := cga.db.Begin()
			err := cga.CreateConceptGroupRelation(testCtx, tx, conceptGroupRelation)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateConceptGroupRelation Exec sql error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := cga.db.Begin()
			err := cga.CreateConceptGroupRelation(testCtx, tx, conceptGroupRelation)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_DeleteObjectTypesFromGroup(t *testing.T) {
	Convey("Test DeleteObjectTypesFromGroup\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_kn_id = ? AND f_branch = ? AND f_concept_type = ? AND f_group_id IN (?,?) AND f_concept_id IN (?,?)",
			CONCEPT_GROUP_RELATION_TABLE_NAME)

		query := interfaces.ConceptGroupRelationsQueryParams{
			KNID:        "kn1",
			Branch:      "main",
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			CGIDs:       []string{"cg1", "cg2"},
			OTIDs:       []string{"ot1", "ot2"},
		}

		Convey("DeleteObjectTypesFromGroup Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := cga.db.Begin()
			rowsAffected, err := cga.DeleteObjectTypesFromGroup(testCtx, tx, query)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteObjectTypesFromGroup Failed dbExec \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("dbExec error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := cga.db.Begin()
			_, err := cga.DeleteObjectTypesFromGroup(testCtx, tx, query)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteObjectTypesFromGroup RowsAffected error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := cga.db.Begin()
			_, err := cga.DeleteObjectTypesFromGroup(testCtx, tx, query)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_GetConceptIDsByConceptGroupIDs(t *testing.T) {
	Convey("test GetConceptIDsByConceptGroupIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		cga, smock := MockNewConceptGroupAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_concept_id FROM %s WHERE f_kn_id = ? AND f_branch = ? AND f_concept_type = ? AND f_group_id IN (?,?)",
			CONCEPT_GROUP_RELATION_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_concept_id"}).
			AddRow("ot1").
			AddRow("ot2").
			AddRow("ot3")

		knID := "kn1"
		branch := "main"
		cgIDs := []string{"cg1", "cg2"}
		conceptType := interfaces.MODULE_TYPE_OBJECT_TYPE

		Convey("GetConceptIDsByConceptGroupIDs Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, conceptType, "cg1", "cg2").WillReturnRows(rows)

			conceptIDs, err := cga.GetConceptIDsByConceptGroupIDs(testCtx, knID, branch, cgIDs, conceptType)
			So(err, ShouldBeNil)
			So(len(conceptIDs), ShouldEqual, 3)
			So(conceptIDs[0], ShouldEqual, "ot1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetConceptIDsByConceptGroupIDs Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, conceptType, "cg1", "cg2").WillReturnError(expectedErr)

			conceptIDs, err := cga.GetConceptIDsByConceptGroupIDs(testCtx, knID, branch, cgIDs, conceptType)
			So(conceptIDs, ShouldResemble, []string{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetConceptIDsByConceptGroupIDs scan error\n", func() {
			rows := sqlmock.NewRows([]string{"f_concept_id", "f_concept_id"}).AddRow("ot1", "ot2")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, conceptType, "cg1", "cg2").WillReturnRows(rows)

			_, err := cga.GetConceptIDsByConceptGroupIDs(testCtx, knID, branch, cgIDs, conceptType)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_ProcessQueryCondition(t *testing.T) {
	Convey("test processQueryCondition ", t, func() {
		appSetting := &common.AppSetting{}
		_, _ = MockNewConceptGroupAccess(appSetting)

		sqlBuilder := sq.Select("COUNT(f_id)").From(CONCEPT_GROUP_TABLE_NAME)

		Convey("NamePattern query ", func() {
			query := interfaces.ConceptGroupsQueryParams{
				NamePattern: "name_a",
				KNID:        "kn1",
				Branch:      "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE (instr(f_name, ?) > 0 OR instr(f_id, ?) > 0) AND f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("Tag query ", func() {
			query := interfaces.ConceptGroupsQueryParams{
				Tag:    "tag1",
				KNID:   "kn1",
				Branch: "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE instr(f_tags, ?) > 0 AND f_kn_id = ? AND f_branch = ?", CONCEPT_GROUP_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("CGIDs query ", func() {
			query := interfaces.ConceptGroupsQueryParams{
				CGIDs:  []string{"cg1", "cg2"},
				KNID:   "kn1",
				Branch: "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE f_kn_id = ? AND f_branch = ? AND f_id IN (?,?)", CONCEPT_GROUP_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})
	})
}

func Test_conceptGroupAccess_GetRelationTypeIDsFromConceptGroupRelation(t *testing.T) {
	Convey("test GetRelationTypeIDsFromConceptGroupRelation\n", t, func() {
		appSetting := &common.AppSetting{}
		// 使用正则表达式匹配器来处理复杂的 SQL
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		cga := &conceptGroupAccess{
			appSetting: appSetting,
			db:         db,
		}

		query := interfaces.ConceptGroupRelationsQueryParams{
			KNID:        "kn1",
			Branch:      "main",
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			CGIDs:       []string{"cg1"},
		}

		Convey("GetRelationTypeIDsFromConceptGroupRelation Success \n", func() {
			// 使用正则表达式匹配 SQL 开头
			rows := sqlmock.NewRows([]string{"f_id"}).
				AddRow("rt1").
				AddRow("rt2")
			smock.ExpectQuery("^SELECT f_id FROM t_relation_type").WillReturnRows(rows)

			rtIDs, err := cga.GetRelationTypeIDsFromConceptGroupRelation(testCtx, query)
			So(err, ShouldBeNil)
			So(len(rtIDs), ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypeIDsFromConceptGroupRelation Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery("^SELECT f_id FROM t_relation_type").WillReturnError(expectedErr)

			rtIDs, err := cga.GetRelationTypeIDsFromConceptGroupRelation(testCtx, query)
			So(rtIDs, ShouldResemble, []string{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypeIDsFromConceptGroupRelation scan error \n", func() {
			rows := sqlmock.NewRows([]string{"f_id", "f_id"}).AddRow(123, "123") // 使用 int 而不是 string
			smock.ExpectQuery("SELECT f_id FROM t_relation_type").WillReturnRows(rows)

			rtIDs, err := cga.GetRelationTypeIDsFromConceptGroupRelation(testCtx, query)
			So(rtIDs, ShouldResemble, []string{})
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_GetActionTypeIDsFromConceptGroupRelation(t *testing.T) {
	Convey("test GetActionTypeIDsFromConceptGroupRelation\n", t, func() {
		appSetting := &common.AppSetting{}
		// 使用正则表达式匹配器来处理复杂的 SQL
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		cga := &conceptGroupAccess{
			appSetting: appSetting,
			db:         db,
		}

		query := interfaces.ConceptGroupRelationsQueryParams{
			KNID:        "kn1",
			Branch:      "main",
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			CGIDs:       []string{"cg1"},
		}

		Convey("GetActionTypeIDsFromConceptGroupRelation Success \n", func() {
			rows := sqlmock.NewRows([]string{"f_id"}).
				AddRow("at1").
				AddRow("at2")
			smock.ExpectQuery("^SELECT f_id FROM t_action_type").WillReturnRows(rows)

			atIDs, err := cga.GetActionTypeIDsFromConceptGroupRelation(testCtx, query)
			So(err, ShouldBeNil)
			So(len(atIDs), ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypeIDsFromConceptGroupRelation Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery("^SELECT f_id FROM t_action_type").WillReturnError(expectedErr)

			atIDs, err := cga.GetActionTypeIDsFromConceptGroupRelation(testCtx, query)
			So(atIDs, ShouldResemble, []string{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypeIDsFromConceptGroupRelation scan error \n", func() {
			rows := sqlmock.NewRows([]string{"f_id", "f_id"}).AddRow(123, "123") // 使用 int 而不是 string
			smock.ExpectQuery("^SELECT f_id FROM t_action_type").WillReturnRows(rows)

			atIDs, err := cga.GetActionTypeIDsFromConceptGroupRelation(testCtx, query)
			So(atIDs, ShouldResemble, []string{})
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_conceptGroupAccess_GetConceptGroupsByOTIDs(t *testing.T) {
	Convey("test GetConceptGroupsByOTIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		// 使用正则表达式匹配器来处理复杂的 SQL
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		cga := &conceptGroupAccess{
			appSetting: appSetting,
			db:         db,
		}

		query := interfaces.ConceptGroupRelationsQueryParams{
			KNID:        "kn1",
			Branch:      "main",
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			OTIDs:       []string{"ot1", "ot2"},
		}

		Convey("GetConceptGroupsByOTIDs Success \n", func() {
			rows := sqlmock.NewRows([]string{
				"cgr.f_concept_id", "cg.f_id", "cg.f_name", "cg.f_tags", "cg.f_comment",
				"cg.f_icon", "cg.f_color", "cg.f_kn_id", "cg.f_branch",
			}).AddRow(
				"ot1", "cg1", "Concept Group 1", `"tag1"`, "comment",
				"icon", "color", "kn1", "main",
			).AddRow(
				"ot1", "cg2", "Concept Group 2", `"tag2"`, "comment2",
				"icon2", "color2", "kn1", "main",
			).AddRow(
				"ot2", "cg1", "Concept Group 1", `"tag1"`, "comment",
				"icon", "color", "kn1", "main",
			)

			smock.ExpectBegin()
			smock.ExpectQuery("^SELECT cgr.f_concept_id").WillReturnRows(rows)

			tx, _ := cga.db.Begin()
			results, err := cga.GetConceptGroupsByOTIDs(testCtx, tx, query)
			So(err, ShouldBeNil)
			So(len(results), ShouldEqual, 2)
			So(len(results["ot1"]), ShouldEqual, 2)
			So(len(results["ot2"]), ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetConceptGroupsByOTIDs Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectBegin()
			smock.ExpectQuery("^SELECT cgr.f_concept_id").WillReturnError(expectedErr)

			tx, _ := cga.db.Begin()
			results, err := cga.GetConceptGroupsByOTIDs(testCtx, tx, query)
			So(results, ShouldResemble, map[string][]*interfaces.ConceptGroup{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetConceptGroupsByOTIDs scan error \n", func() {
			rows := sqlmock.NewRows([]string{"cgr.f_concept_id"}).AddRow("ot1")
			smock.ExpectBegin()
			smock.ExpectQuery("^SELECT cgr.f_concept_id").WillReturnRows(rows)

			tx, _ := cga.db.Begin()
			results, err := cga.GetConceptGroupsByOTIDs(testCtx, tx, query)
			So(results, ShouldResemble, map[string][]*interfaces.ConceptGroup{})
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
