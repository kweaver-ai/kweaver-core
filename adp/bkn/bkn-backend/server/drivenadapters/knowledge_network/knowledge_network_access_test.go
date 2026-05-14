// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knowledge_network

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

	testKN = &interfaces.KN{
		KNID:   "kn1",
		KNName: "Knowledge Network 1",
		CommonInfo: interfaces.CommonInfo{
			Tags:          testTags,
			Comment:       "test comment",
			Icon:          "icon1",
			Color:         "color1",
			BKNRawContent: "detail1",
		},
		Branch:         interfaces.MAIN_BRANCH,
		BusinessDomain: "domain1",
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
		ModuleType: interfaces.MODULE_TYPE_KN,
	}
)

func MockNewKNAccess(appSetting *common.AppSetting) (*knowledgeNetworkAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	kna := &knowledgeNetworkAccess{
		appSetting: appSetting,
		db:         db,
	}
	return kna, smock
}

func Test_knowledgeNetworkAccess_CheckKNExistByID(t *testing.T) {
	Convey("test CheckKNExistByID\n", t, func() {
		appSetting := &common.AppSetting{}
		kna, smock := MockNewKNAccess(appSetting)

		sqlStr := "SELECT f_name FROM t_knowledge_network WHERE f_id = ? AND f_branch = ?"

		knID := "kn1"
		branch := "main"

		Convey("CheckKNExistByID Success \n", func() {
			rows := sqlmock.NewRows([]string{"f_name"}).AddRow("Knowledge Network 1")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			name, exists, err := kna.CheckKNExistByID(testCtx, knID, branch)
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(name, ShouldEqual, "Knowledge Network 1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckKNExistByID Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnError(sql.ErrNoRows)

			name, exists, err := kna.CheckKNExistByID(testCtx, knID, branch)
			So(name, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckKNExistByID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnError(expectedErr)

			name, exists, err := kna.CheckKNExistByID(testCtx, knID, branch)
			So(name, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_knowledgeNetworkAccess_CheckKNExistByName(t *testing.T) {
	Convey("test CheckKNExistByName\n", t, func() {
		appSetting := &common.AppSetting{}
		kna, smock := MockNewKNAccess(appSetting)

		sqlStr := "SELECT f_id FROM t_knowledge_network WHERE f_name = ? AND f_branch = ?"

		knName := "Knowledge Network 1"
		branch := "main"

		Convey("CheckKNExistByName Success \n", func() {
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("kn1")
			smock.ExpectQuery(sqlStr).WithArgs(knName, branch).WillReturnRows(rows)

			knID, exists, err := kna.CheckKNExistByName(testCtx, knName, branch)
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(knID, ShouldEqual, "kn1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckKNExistByName Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knName, branch).WillReturnError(sql.ErrNoRows)

			knID, exists, err := kna.CheckKNExistByName(testCtx, knName, branch)
			So(knID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckKNExistByName Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knName, branch).WillReturnError(expectedErr)

			knID, exists, err := kna.CheckKNExistByName(testCtx, knName, branch)
			So(knID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_knowledgeNetworkAccess_CreateKN(t *testing.T) {
	Convey("test CreateKN\n", t, func() {
		appSetting := &common.AppSetting{}
		kna, smock := MockNewKNAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id,f_name,f_tags,f_comment,f_icon,f_color,f_bkn_raw_content,f_skill_content,"+
			"f_branch,f_business_domain,f_creator,f_creator_type,f_create_time,f_updater,f_updater_type,f_update_time) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", KN_TABLE_NAME)

		Convey("CreateKN Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := kna.db.Begin()
			err := kna.CreateKN(testCtx, tx, testKN)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateKN Exec sql error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := kna.db.Begin()
			err := kna.CreateKN(testCtx, tx, testKN)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

// Test_NewKNAccess 跳过测试，因为NewKNAccess需要实际的数据库连接
// 在单元测试中使用MockNewKNAccess代替
// func Test_NewKNAccess(t *testing.T) {
// 	Convey("test NewKNAccess\n", t, func() {
// 		appSetting := &common.AppSetting{
// 			DBSetting: libdb.DBSetting{
// 				Host:     "localhost",
// 				Port:     3306,
// 				Username: "test",
// 				Password: "test",
// 				DBName:   "test",
// 			},
// 		}
//
// 		Convey("NewKNAccess Success\n", func() {
// 			access := NewKNAccess(appSetting)
// 			So(access, ShouldNotBeNil)
//
// 			// 第二次调用应该返回同一个实例（单例模式）
// 			access2 := NewKNAccess(appSetting)
// 			So(access2, ShouldEqual, access)
// 		})
// 	})
// }

func Test_knowledgeNetworkAccess_ListKNs(t *testing.T) {
	Convey("test ListKNs\n", t, func() {
		appSetting := &common.AppSetting{}
		kna, smock := MockNewKNAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, f_skill_content, "+
			"f_branch, f_business_domain, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s", KN_TABLE_NAME)

		rows := sqlmock.NewRows([]string{
			"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content", "f_skill_content",
			"f_branch", "f_business_domain", "f_creator", "f_creator_type", "f_create_time",
			"f_updater", "f_updater_type", "f_update_time",
		}).AddRow(
			"kn1", "Knowledge Network 1", `"tag1"`, "comment", "icon", "color", "detail", "skill",
			"main", "domain1", "admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		)

		query := interfaces.KNsQueryParams{}

		Convey("ListKNs Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			kns, err := kna.ListKNs(testCtx, query)
			So(err, ShouldBeNil)
			So(len(kns), ShouldEqual, 1)
			So(kns[0].KNID, ShouldEqual, "kn1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListKNs Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(sqlmock.NewRows(nil))

			kns, err := kna.ListKNs(testCtx, query)
			So(kns, ShouldResemble, []*interfaces.KN{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListKNs Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			kns, err := kna.ListKNs(testCtx, query)
			So(kns, ShouldResemble, []*interfaces.KN{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListKNs Scan error \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content", "f_skill_content",
				"f_branch", "f_business_domain", "f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time", "f_update_time",
			}).AddRow(
				"kn1", "Knowledge Network 1", `"tag1"`, "comment", "icon", "color", "detail", "detail",
				"main", "domain1", "admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime, "testUpdateTime",
			)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			kns, err := kna.ListKNs(testCtx, query)
			So(kns, ShouldResemble, []*interfaces.KN{})
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListKNs with all query params \n", func() {
			queryWithAll := interfaces.KNsQueryParams{
				NamePattern:    "test",
				Tag:            "tag1",
				BusinessDomain: "domain1",
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Offset:    10,
					Limit:     20,
					Sort:      "f_name",
					Direction: "ASC",
				},
			}
			sqlStrWithAll := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, f_skill_content, "+
				"f_branch, f_business_domain, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
				"FROM %s WHERE (instr(f_name, ?) > 0 OR instr(f_id, ?) > 0) AND instr(f_tags, ?) > 0 AND f_business_domain = ? ORDER BY f_name ASC",
				KN_TABLE_NAME)

			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content", "f_skill_content",
				"f_branch", "f_business_domain", "f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			})

			smock.ExpectQuery(sqlStrWithAll).WithArgs().WillReturnRows(rows)

			kns, err := kna.ListKNs(testCtx, queryWithAll)
			So(err, ShouldBeNil)
			So(kns, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_knowledgeNetworkAccess_GetKNsTotal(t *testing.T) {
	Convey("test GetKNsTotal\n", t, func() {
		appSetting := &common.AppSetting{}
		kna, smock := MockNewKNAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s", KN_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"COUNT(f_id)"}).AddRow(1)

		query := interfaces.KNsQueryParams{}

		Convey("GetKNsTotal Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := kna.GetKNsTotal(testCtx, query)
			So(total, ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetKNsTotal Failed  Query error\n", func() {
			expectedErr := errors.New("Query error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := kna.GetKNsTotal(testCtx, query)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetKNsTotal with all query params \n", func() {
			queryWithAll := interfaces.KNsQueryParams{
				NamePattern:    "test",
				Tag:            "tag1",
				BusinessDomain: "domain1",
			}
			sqlStrWithAll := fmt.Sprintf("SELECT COUNT(f_id) FROM %s WHERE (instr(f_name, ?) > 0 OR instr(f_id, ?) > 0) AND instr(f_tags, ?) > 0 AND f_business_domain = ?", KN_TABLE_NAME)

			rows := sqlmock.NewRows([]string{"COUNT(f_id)"}).AddRow(5)
			smock.ExpectQuery(sqlStrWithAll).WithArgs().WillReturnRows(rows)

			total, err := kna.GetKNsTotal(testCtx, queryWithAll)
			So(total, ShouldEqual, 5)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_knowledgeNetworkAccess_GetKNByID(t *testing.T) {
	Convey("test GetKNByID\n", t, func() {
		appSetting := &common.AppSetting{}
		kna, smock := MockNewKNAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, f_skill_content, "+
			"f_branch, f_business_domain, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_id = ? AND f_branch = ?", KN_TABLE_NAME)

		knID := "kn1"
		branch := "main"

		Convey("GetKNByID Success \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content", "f_skill_content",
				"f_branch", "f_business_domain", "f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"kn1", "Knowledge Network 1", `"tag1"`, "comment", "icon", "color", "detail", "detail",
				"main", "domain1", "admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			kn, err := kna.GetKNByID(testCtx, knID, branch)
			So(err, ShouldBeNil)
			So(kn, ShouldNotBeNil)
			So(kn.KNID, ShouldEqual, "kn1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetKNByID Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnError(sql.ErrNoRows)

			kn, err := kna.GetKNByID(testCtx, knID, branch)
			So(kn, ShouldBeNil)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetKNByID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnError(expectedErr)

			kn, err := kna.GetKNByID(testCtx, knID, branch)
			So(kn, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_knowledgeNetworkAccess_UpdateKN(t *testing.T) {
	Convey("Test UpdateKN\n", t, func() {
		appSetting := &common.AppSetting{}
		kna, smock := MockNewKNAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_bkn_raw_content = ?, f_color = ?, f_comment = ?, f_icon = ?, f_name = ?, "+
			"f_skill_content = ?, f_tags = ?, f_update_time = ?, f_updater = ?, f_updater_type = ? WHERE f_id = ?", KN_TABLE_NAME)

		kn := &interfaces.KN{
			KNID:   "kn1",
			KNName: "Updated Knowledge Network",
			CommonInfo: interfaces.CommonInfo{
				Tags:    testTags,
				Comment: "updated comment",
				Icon:    "icon1",
				Color:   "color1",
			},
			Updater: interfaces.AccountInfo{
				ID:   "admin",
				Type: "admin",
			},
			UpdateTime: testUpdateTime,
		}

		Convey("UpdateKN Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := kna.db.Begin()
			err := kna.UpdateKN(testCtx, tx, kn)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateKN failed prepare \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("prepare error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := kna.db.Begin()
			err := kna.UpdateKN(testCtx, tx, kn)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateKN RowsAffected != 1 \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(0, 0))

			tx, _ := kna.db.Begin()
			err := kna.UpdateKN(testCtx, tx, kn)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateKN RowsAffected error \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := kna.db.Begin()
			err := kna.UpdateKN(testCtx, tx, kn)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_knowledgeNetworkAccess_UpdateKNDetail(t *testing.T) {
	Convey("Test UpdateKNDetail\n", t, func() {
		appSetting := &common.AppSetting{}
		kna, smock := MockNewKNAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_bkn_raw_content = ? WHERE f_id = ? AND f_branch = ?", KN_TABLE_NAME)

		knID := "kn1"
		branch := "main"
		detail := "updated detail"

		Convey("UpdateKNDetail Success \n", func() {
			smock.ExpectExec(sqlStr).WithArgs(detail, knID, branch).WillReturnResult(sqlmock.NewResult(1, 1))

			err := kna.UpdateKNDetail(testCtx, knID, branch, detail)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateKNDetail Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs(detail, knID, branch).WillReturnError(expectedErr)

			err := kna.UpdateKNDetail(testCtx, knID, branch, detail)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateKNDetail RowsAffected != 1 \n", func() {
			smock.ExpectExec(sqlStr).WithArgs(detail, knID, branch).WillReturnResult(sqlmock.NewResult(0, 0))

			err := kna.UpdateKNDetail(testCtx, knID, branch, detail)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateKNDetail RowsAffected error \n", func() {
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs(detail, knID, branch).WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			err := kna.UpdateKNDetail(testCtx, knID, branch, detail)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_knowledgeNetworkAccess_DeleteKN(t *testing.T) {
	Convey("Test DeleteKN\n", t, func() {
		appSetting := &common.AppSetting{}
		kna, smock := MockNewKNAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_id = ? AND f_branch = ?", KN_TABLE_NAME)

		knID := "kn1"
		branch := "main"

		Convey("DeleteKN Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs(knID, branch).WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := kna.db.Begin()
			rowsAffected, err := kna.DeleteKN(testCtx, tx, knID, branch)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteKN Failed dbExec \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("dbExec error")
			smock.ExpectExec(sqlStr).WithArgs(knID, branch).WillReturnError(expectedErr)

			tx, _ := kna.db.Begin()
			_, err := kna.DeleteKN(testCtx, tx, knID, branch)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteKN Failed RowsAffected \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs(knID, branch).WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := kna.db.Begin()
			_, err := kna.DeleteKN(testCtx, tx, knID, branch)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteKN RowsAffected != 1 \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs(knID, branch).WillReturnResult(sqlmock.NewResult(0, 0))

			tx, _ := kna.db.Begin()
			rowsAffected, err := kna.DeleteKN(testCtx, tx, knID, branch)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 0)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_knowledgeNetworkAccess_GetAllKNs(t *testing.T) {
	Convey("test GetAllKNs\n", t, func() {
		appSetting := &common.AppSetting{}
		kna, smock := MockNewKNAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, f_skill_content, "+
			"f_branch, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s", KN_TABLE_NAME)

		rows := sqlmock.NewRows([]string{
			"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content", "f_skill_content",
			"f_branch", "f_creator", "f_creator_type", "f_create_time",
			"f_updater", "f_updater_type", "f_update_time",
		}).AddRow(
			"kn1", "Knowledge Network 1", `"tag1"`, "comment", "icon", "color", "detail", "skill",
			"main", "admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		).AddRow(
			"kn2", "Knowledge Network 2", `"tag2"`, "comment2", "icon2", "color2", "detail2", "skill2",
			"main", "admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		)

		Convey("GetAllKNs Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			kns, err := kna.GetAllKNs(testCtx)
			So(err, ShouldBeNil)
			So(len(kns), ShouldEqual, 2)
			So(kns["kn1"], ShouldNotBeNil)
			So(kns["kn2"], ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllKNs Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			kns, err := kna.GetAllKNs(testCtx)
			So(kns, ShouldResemble, map[string]*interfaces.KN{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllKNs Scan error \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content", "f_skill_content",
				"f_branch", "f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time", "f_update_time",
			}).AddRow(
				"kn1", "Knowledge Network 1", `"tag1"`, "comment", "icon", "color", "detail", "skill",
				"main", "admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime, "testUpdateTime",
			)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			kns, err := kna.GetAllKNs(testCtx)
			So(kns, ShouldResemble, map[string]*interfaces.KN{})
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_knowledgeNetworkAccess_ListKnSrcs(t *testing.T) {
	Convey("test ListKnSrcs\n", t, func() {
		appSetting := &common.AppSetting{}
		kna, smock := MockNewKNAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name FROM %s", KN_TABLE_NAME)

		query := interfaces.KNsQueryParams{}

		Convey("ListKnSrcs Success \n", func() {
			rows := sqlmock.NewRows([]string{"f_id", "f_name"}).
				AddRow("kn1", "Knowledge Network 1").
				AddRow("kn2", "Knowledge Network 2")

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			srcs, err := kna.ListKnSrcs(testCtx, query)
			So(err, ShouldBeNil)
			So(len(srcs), ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListKnSrcs Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			srcs, err := kna.ListKnSrcs(testCtx, query)
			So(srcs, ShouldResemble, []interfaces.PermissionResource{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListKnSrcs Scan error \n", func() {
			rows := sqlmock.NewRows([]string{"f_id", "f_name", "f_update_time"}).
				AddRow("kn1", "Knowledge Network 1", "testUpdateTime")

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			srcs, err := kna.ListKnSrcs(testCtx, query)
			So(srcs, ShouldResemble, []interfaces.PermissionResource{})
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListKnSrcs with NamePattern and Sort \n", func() {
			queryWithParams := interfaces.KNsQueryParams{
				NamePattern: "test",
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Sort:      "f_name",
					Direction: "ASC",
				},
			}
			sqlStrWithParams := fmt.Sprintf("SELECT f_id, f_name FROM %s WHERE (instr(f_name, ?) > 0 OR instr(f_id, ?) > 0) ORDER BY f_name ASC", KN_TABLE_NAME)

			rows := sqlmock.NewRows([]string{"f_id", "f_name"}).
				AddRow("kn1", "Knowledge Network 1")

			smock.ExpectQuery(sqlStrWithParams).WithArgs().WillReturnRows(rows)

			srcs, err := kna.ListKnSrcs(testCtx, queryWithParams)
			So(err, ShouldBeNil)
			So(len(srcs), ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_knowledgeNetworkAccess_ProcessQueryCondition(t *testing.T) {
	Convey("test processQueryCondition ", t, func() {
		appSetting := &common.AppSetting{}
		_, _ = MockNewKNAccess(appSetting)

		sqlBuilder := sq.Select("COUNT(f_id)").From(KN_TABLE_NAME)

		Convey("NamePattern query ", func() {
			query := interfaces.KNsQueryParams{
				NamePattern: "name_a",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE (instr(f_name, ?) > 0 OR instr(f_id, ?) > 0)", KN_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("Tag query ", func() {
			query := interfaces.KNsQueryParams{
				Tag: "tag1",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE instr(f_tags, ?) > 0", KN_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("BusinessDomain query ", func() {
			query := interfaces.KNsQueryParams{
				BusinessDomain: "domain1",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE f_business_domain = ?", KN_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})
	})
}

func Test_knowledgeNetworkAccess_ProcessConceptGroupRelationsQueryCondition(t *testing.T) {
	Convey("test processConceptGroupRelationsQueryCondition", t, func() {
		appSetting := &common.AppSetting{}
		_, _ = MockNewKNAccess(appSetting)

		sqlBuilder := sq.Select("f_concept_id").From("t_concept_group_relation AS cgr")

		Convey("KNID query", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				KNID: "kn1",
			}

			sqlBuilder := processConceptGroupRelationsQueryCondition(query, sqlBuilder, "cgr.")
			sqlStr, vals, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldContainSubstring, "f_kn_id")
			So(len(vals), ShouldBeGreaterThan, 0)
		})

		Convey("Branch query", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				Branch: "main",
			}

			sqlBuilder := processConceptGroupRelationsQueryCondition(query, sqlBuilder, "cgr.")
			sqlStr, vals, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldContainSubstring, "f_branch")
			So(len(vals), ShouldBeGreaterThan, 0)
		})

		Convey("Empty Branch query", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				Branch: "",
			}

			sqlBuilder := processConceptGroupRelationsQueryCondition(query, sqlBuilder, "cgr.")
			sqlStr, vals, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldContainSubstring, "f_branch")
			So(len(vals), ShouldBeGreaterThan, 0)
		})

		Convey("CGIDs query", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				CGIDs: []string{"cg1", "cg2"},
			}

			sqlBuilder := processConceptGroupRelationsQueryCondition(query, sqlBuilder, "cgr.")
			sqlStr, vals, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldContainSubstring, "f_group_id")
			So(len(vals), ShouldBeGreaterThan, 0)
		})

		Convey("ConceptType query", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				ConceptType: "object_type",
			}

			sqlBuilder := processConceptGroupRelationsQueryCondition(query, sqlBuilder, "cgr.")
			sqlStr, vals, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldContainSubstring, "f_concept_type")
			So(len(vals), ShouldBeGreaterThan, 0)
		})

		Convey("OTIDs query", func() {
			query := interfaces.ConceptGroupRelationsQueryParams{
				OTIDs: []string{"ot1", "ot2"},
			}

			sqlBuilder := processConceptGroupRelationsQueryCondition(query, sqlBuilder, "cgr.")
			sqlStr, vals, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldContainSubstring, "f_concept_id")
			So(len(vals), ShouldBeGreaterThan, 0)
		})
	})
}

func Test_knowledgeNetworkAccess_GetNeighborPathsBatch(t *testing.T) {
	Convey("test GetNeighborPathsBatch", t, func() {
		appSetting := &common.AppSetting{}
		// 使用 QueryMatcherRegexp 来匹配复杂的 SQL
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		kna := &knowledgeNetworkAccess{
			appSetting: appSetting,
			db:         db,
		}

		otIDs := []string{"ot1"}
		query := interfaces.RelationTypePathsBaseOnSource{
			KNID:   "kn1",
			Branch: "main",
		}

		Convey("GetNeighborPathsBatch forward direction query error", func() {
			query.Direction = interfaces.DIRECTION_FORWARD
			expectedErr := errors.New("query error")
			// 匹配包含 forward 和 rt.f_source_object_type_id 的 SQL
			smock.ExpectQuery(`.*forward.*rt\.f_source_object_type_id.*`).WillReturnError(expectedErr)

			result, err := kna.GetNeighborPathsBatch(testCtx, otIDs, query)
			So(result, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetNeighborPathsBatch backward direction query error", func() {
			query.Direction = interfaces.DIRECTION_BACKWARD
			expectedErr := errors.New("query error")
			// 匹配包含 backward 和 rt.f_target_object_type_id 的 SQL
			smock.ExpectQuery(`.*backward.*rt\.f_target_object_type_id.*`).WillReturnError(expectedErr)

			result, err := kna.GetNeighborPathsBatch(testCtx, otIDs, query)
			So(result, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetNeighborPathsBatch bidirectional direction query error", func() {
			query.Direction = interfaces.DIRECTION_BIDIRECTIONAL
			expectedErr := errors.New("query error")
			// 匹配包含 UNION ALL 的 SQL
			smock.ExpectQuery(`.*UNION ALL.*`).WillReturnError(expectedErr)

			result, err := kna.GetNeighborPathsBatch(testCtx, otIDs, query)
			So(result, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetNeighborPathsBatch scan error", func() {
			query.Direction = interfaces.DIRECTION_FORWARD
			rows := sqlmock.NewRows([]string{
				"direction", "source_id", "neighbor_id", "rt_id", "rt_name",
				"source_ot_id", "target_ot_id", "rt_type", "mapping_rules",
				"ot_id", "ot_name", "data_source", "data_properties",
				"logic_properties", "primary_keys", "display_key", "display_key",
			}).AddRow(
				"forward", "ot1", "ot2", "rt1", "Relation 1",
				"ot1", "ot2", "direct", []byte("[]"),
				"ot2", "Object Type 2", []byte("{}"), []byte("[]"),
				[]byte("[]"), []byte("[]"), "id", "id",
			)

			smock.ExpectQuery(`.*forward.*rt\.f_source_object_type_id.*`).WillReturnRows(rows)

			result, err := kna.GetNeighborPathsBatch(testCtx, otIDs, query)
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetNeighborPathsBatch unmarshal mappingRules error", func() {
			query.Direction = interfaces.DIRECTION_FORWARD
			rows := sqlmock.NewRows([]string{
				"direction", "source_id", "neighbor_id", "rt_id", "rt_name",
				"source_ot_id", "target_ot_id", "rt_type", "mapping_rules",
				"ot_id", "ot_name", "data_source", "data_properties",
				"logic_properties", "primary_keys", "display_key",
			}).AddRow(
				"forward", "ot1", "ot2", "rt1", "Relation 1",
				"ot1", "ot2", "direct", []byte("invalid json"),
				"ot2", "Object Type 2", []byte("{}"), []byte("[]"),
				[]byte("[]"), []byte("[]"), "id",
			)

			smock.ExpectQuery(`.*forward.*rt\.f_source_object_type_id.*`).WillReturnRows(rows)

			result, err := kna.GetNeighborPathsBatch(testCtx, otIDs, query)
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetNeighborPathsBatch unmarshal dataSource error", func() {
			query.Direction = interfaces.DIRECTION_FORWARD
			rows := sqlmock.NewRows([]string{
				"direction", "source_id", "neighbor_id", "rt_id", "rt_name",
				"source_ot_id", "target_ot_id", "rt_type", "mapping_rules",
				"ot_id", "ot_name", "data_source", "data_properties",
				"logic_properties", "primary_keys", "display_key",
			}).AddRow(
				"forward", "ot1", "ot2", "rt1", "Relation 1",
				"ot1", "ot2", "direct", []byte("[]"),
				"ot2", "Object Type 2", []byte("invalid json"), []byte("[]"),
				[]byte("[]"), []byte("[]"), "id",
			)

			smock.ExpectQuery(`.*forward.*rt\.f_source_object_type_id.*`).WillReturnRows(rows)

			result, err := kna.GetNeighborPathsBatch(testCtx, otIDs, query)
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetNeighborPathsBatch unmarshal dataProperties error", func() {
			query.Direction = interfaces.DIRECTION_FORWARD
			rows := sqlmock.NewRows([]string{
				"direction", "source_id", "neighbor_id", "rt_id", "rt_name",
				"source_ot_id", "target_ot_id", "rt_type", "mapping_rules",
				"ot_id", "ot_name", "data_source", "data_properties",
				"logic_properties", "primary_keys", "display_key",
			}).AddRow(
				"forward", "ot1", "ot2", "rt1", "Relation 1",
				"ot1", "ot2", "direct", []byte("[]"),
				"ot2", "Object Type 2", []byte("{}"), []byte("invalid json"),
				[]byte("[]"), []byte("[]"), "id",
			)

			smock.ExpectQuery(`.*forward.*rt\.f_source_object_type_id.*`).WillReturnRows(rows)

			result, err := kna.GetNeighborPathsBatch(testCtx, otIDs, query)
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetNeighborPathsBatch unmarshal logicProperties error", func() {
			query.Direction = interfaces.DIRECTION_FORWARD
			rows := sqlmock.NewRows([]string{
				"direction", "source_id", "neighbor_id", "rt_id", "rt_name",
				"source_ot_id", "target_ot_id", "rt_type", "mapping_rules",
				"ot_id", "ot_name", "data_source", "data_properties",
				"logic_properties", "primary_keys", "display_key",
			}).AddRow(
				"forward", "ot1", "ot2", "rt1", "Relation 1",
				"ot1", "ot2", "direct", []byte("[]"),
				"ot2", "Object Type 2", []byte("{}"), []byte("[]"),
				[]byte("invalid json"), []byte("[]"), "id",
			)

			smock.ExpectQuery(`.*forward.*rt\.f_source_object_type_id.*`).WillReturnRows(rows)

			result, err := kna.GetNeighborPathsBatch(testCtx, otIDs, query)
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetNeighborPathsBatch unmarshal primaryKeys error", func() {
			query.Direction = interfaces.DIRECTION_FORWARD
			rows := sqlmock.NewRows([]string{
				"direction", "source_id", "neighbor_id", "rt_id", "rt_name",
				"source_ot_id", "target_ot_id", "rt_type", "mapping_rules",
				"ot_id", "ot_name", "data_source", "data_properties",
				"logic_properties", "primary_keys", "display_key",
			}).AddRow(
				"forward", "ot1", "ot2", "rt1", "Relation 1",
				"ot1", "ot2", "direct", []byte("[]"),
				"ot2", "Object Type 2", []byte("{}"), []byte("[]"),
				[]byte("[]"), []byte("invalid json"), "id",
			)

			smock.ExpectQuery(`.*forward.*rt\.f_source_object_type_id.*`).WillReturnRows(rows)

			result, err := kna.GetNeighborPathsBatch(testCtx, otIDs, query)
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
