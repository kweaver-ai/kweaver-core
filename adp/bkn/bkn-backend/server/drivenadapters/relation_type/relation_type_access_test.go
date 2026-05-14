// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package relation_type

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

var (
	testUpdateTime = int64(1735786555379)
	testTags       = []string{"tag1", "tag2", "tag3"}

	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)

	testRelationType = &interfaces.RelationType{
		RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
			RTID:               "rt1",
			RTName:             "Relation Type 1",
			SourceObjectTypeID: "ot1",
			TargetObjectTypeID: "ot2",
			Type:               interfaces.RELATION_TYPE_DIRECT,
			MappingRules:       []interfaces.Mapping{},
		},
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
		ModuleType: interfaces.MODULE_TYPE_RELATION_TYPE,
	}
)

func MockNewRelationTypeAccess(appSetting *common.AppSetting) (*relationTypeAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	rta := &relationTypeAccess{
		appSetting: appSetting,
		db:         db,
	}
	return rta, smock
}

func Test_relationTypeAccess_CheckRelationTypeExistByID(t *testing.T) {
	Convey("test CheckRelationTypeExistByID\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRelationTypeAccess(appSetting)

		sqlStr := "SELECT f_name FROM t_relation_type WHERE f_kn_id = ? AND f_branch = ? AND f_id = ?"

		knID := "kn1"
		branch := "main"
		rtID := "rt1"

		Convey("CheckRelationTypeExistByID Success \n", func() {
			rows := sqlmock.NewRows([]string{"f_name"}).AddRow("Relation Type 1")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtID).WillReturnRows(rows)

			name, exists, err := rta.CheckRelationTypeExistByID(testCtx, knID, branch, rtID)
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(name, ShouldEqual, "Relation Type 1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckRelationTypeExistByID Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtID).WillReturnError(sql.ErrNoRows)

			name, exists, err := rta.CheckRelationTypeExistByID(testCtx, knID, branch, rtID)
			So(name, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckRelationTypeExistByID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtID).WillReturnError(expectedErr)

			name, exists, err := rta.CheckRelationTypeExistByID(testCtx, knID, branch, rtID)
			So(name, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_relationTypeAccess_CreateRelationType(t *testing.T) {
	Convey("test CreateRelationType\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRelationTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id,f_name,f_tags,f_comment,f_icon,f_color,f_bkn_raw_content,"+
			"f_kn_id,f_branch,f_source_object_type_id,f_target_object_type_id,f_type,f_mapping_rules,"+
			"f_creator,f_creator_type,f_create_time,f_updater,f_updater_type,f_update_time) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", RT_TABLE_NAME)

		Convey("CreateRelationType Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := rta.db.Begin()
			err := rta.CreateRelationType(testCtx, tx, testRelationType)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateRelationType Exec sql error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := rta.db.Begin()
			err := rta.CreateRelationType(testCtx, tx, testRelationType)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateRelationType Marshal error\n", func() {
			// 创建一个会导致marshal失败的relationType
			invalidRelationType := &interfaces.RelationType{
				RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
					RTID:               "rt1",
					RTName:             "Test Relation Type",
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					Type:               interfaces.RELATION_TYPE_DIRECT,
					MappingRules:       make(chan int), // 使用channel会导致marshal失败
				},
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
				ModuleType: interfaces.MODULE_TYPE_RELATION_TYPE,
			}

			tx, _ := rta.db.Begin()
			err := rta.CreateRelationType(testCtx, tx, invalidRelationType)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_relationTypeAccess_ListRelationTypes(t *testing.T) {
	Convey("test ListRelationTypes\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRelationTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
			"f_kn_id, f_branch, f_source_object_type_id, f_target_object_type_id, f_type, f_mapping_rules, "+
			"f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_kn_id = ? AND f_branch = ?", RT_TABLE_NAME)

		mappingRulesBytes, _ := sonic.Marshal(testRelationType.MappingRules)

		rows := sqlmock.NewRows([]string{
			"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
			"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
			"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
		}).AddRow(
			"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
			"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DIRECT, mappingRulesBytes,
			"admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		)

		query := interfaces.RelationTypesQueryParams{
			KNID:   "kn1",
			Branch: "main",
		}

		Convey("ListRelationTypes Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			relationTypes, err := rta.ListRelationTypes(testCtx, query)
			So(err, ShouldBeNil)
			So(len(relationTypes), ShouldEqual, 1)
			So(relationTypes[0].RTID, ShouldEqual, "rt1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListRelationTypes Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(sqlmock.NewRows(nil))

			relationTypes, err := rta.ListRelationTypes(testCtx, query)
			So(relationTypes, ShouldResemble, []*interfaces.RelationType{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListRelationTypes Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			relationTypes, err := rta.ListRelationTypes(testCtx, query)
			So(relationTypes, ShouldResemble, []*interfaces.RelationType{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListRelationTypes Scan error \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DIRECT, mappingRulesBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime, "f_update_time",
			)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			relationTypes, err := rta.ListRelationTypes(testCtx, query)
			So(relationTypes, ShouldResemble, []*interfaces.RelationType{})
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListRelationTypes Unmarshal mappingRules error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DIRECT, invalidBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			relationTypes, err := rta.ListRelationTypes(testCtx, query)
			So(relationTypes, ShouldResemble, []*interfaces.RelationType{})
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListRelationTypes with all query params \n", func() {
			queryWithAll := interfaces.RelationTypesQueryParams{
				NamePattern:         "test",
				Tag:                 "tag1",
				SourceObjectTypeIDs: []string{"ot1"},
				TargetObjectTypeIDs: []string{"ot2"},
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Offset:    10,
					Limit:     20,
					Sort:      "f_name",
					Direction: "ASC",
				},
			}
			sqlStrWithAll := `SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content,
			 f_kn_id, f_branch, f_source_object_type_id, f_target_object_type_id, f_type, f_mapping_rules, 
			 f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time 
			 FROM t_relation_type WHERE (instr(f_name, ?) > 0 OR instr(f_id, ?) > 0) AND instr(f_tags, ?) > 0 AND f_branch = ? 
			 AND f_source_object_type_id IN (?) AND f_target_object_type_id IN (?) ORDER BY f_name ASC`

			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
			})

			smock.ExpectQuery(sqlStrWithAll).WithArgs().WillReturnRows(rows)

			relationTypes, err := rta.ListRelationTypes(testCtx, queryWithAll)
			So(err, ShouldBeNil)
			So(relationTypes, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_relationTypeAccess_GetRelationTypesTotal(t *testing.T) {
	Convey("test GetRelationTypesTotal\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRelationTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s WHERE f_kn_id = ? AND f_branch = ?", RT_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"COUNT(f_id)"}).AddRow(1)

		query := interfaces.RelationTypesQueryParams{
			KNID:   "kn1",
			Branch: "main",
		}

		Convey("GetRelationTypesTotal Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := rta.GetRelationTypesTotal(testCtx, query)
			So(total, ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypesTotal Failed  Query error\n", func() {
			expectedErr := errors.New("Query error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := rta.GetRelationTypesTotal(testCtx, query)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypesTotal with all query params \n", func() {
			queryWithAll := interfaces.RelationTypesQueryParams{
				NamePattern:         "test",
				Tag:                 "tag1",
				SourceObjectTypeIDs: []string{"ot1"},
				TargetObjectTypeIDs: []string{"ot2"},
			}
			sqlStrWithAll := `SELECT COUNT(f_id) FROM t_relation_type WHERE (instr(f_name, ?) > 0 OR instr(f_id, ?) > 0) 
			AND instr(f_tags, ?) > 0 AND f_branch = ? AND f_source_object_type_id IN (?) 
			AND f_target_object_type_id IN (?)`

			rows := sqlmock.NewRows([]string{"COUNT(f_id)"}).AddRow(5)
			smock.ExpectQuery(sqlStrWithAll).WithArgs().WillReturnRows(rows)

			total, err := rta.GetRelationTypesTotal(testCtx, queryWithAll)
			So(total, ShouldEqual, 5)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_relationTypeAccess_GetRelationTypeByID(t *testing.T) {
	Convey("test GetRelationTypeByID\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRelationTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
			"f_kn_id, f_branch, f_source_object_type_id, f_target_object_type_id, f_type, f_mapping_rules, "+
			"f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_kn_id = ? AND f_branch = ? AND f_id = ?", RT_TABLE_NAME)

		knID := "kn1"
		branch := "main"
		rtID := "rt1"

		mappingRulesBytes, _ := sonic.Marshal(testRelationType.MappingRules)

		Convey("GetRelationTypeByID Success \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DIRECT, mappingRulesBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtID).WillReturnRows(rows)

			relationType, err := rta.GetRelationTypeByID(testCtx, knID, branch, rtID)
			So(err, ShouldBeNil)
			So(relationType, ShouldNotBeNil)
			So(relationType.RTID, ShouldEqual, "rt1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypeByID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtID).WillReturnError(expectedErr)

			relationType, err := rta.GetRelationTypeByID(testCtx, knID, branch, rtID)
			So(relationType, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypeByID Unmarshal mappingRules DIRECT type error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DIRECT, invalidBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtID).WillReturnRows(rows)

			relationType, err := rta.GetRelationTypeByID(testCtx, knID, branch, rtID)
			So(relationType, ShouldBeNil)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypeByID Unmarshal mappingRules DATA_VIEW type error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DATA_VIEW, invalidBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtID).WillReturnRows(rows)

			relationType, err := rta.GetRelationTypeByID(testCtx, knID, branch, rtID)
			So(relationType, ShouldBeNil)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypeByID Success DATA_VIEW type \n", func() {
			dataViewMapping := &interfaces.InDirectMapping{}
			dataViewMappingBytes, _ := sonic.Marshal(dataViewMapping)
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DATA_VIEW, dataViewMappingBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, rtID).WillReturnRows(rows)

			relationType, err := rta.GetRelationTypeByID(testCtx, knID, branch, rtID)
			So(err, ShouldBeNil)
			So(relationType, ShouldNotBeNil)
			So(relationType.RTID, ShouldEqual, "rt1")
			So(relationType.Type, ShouldEqual, interfaces.RELATION_TYPE_DATA_VIEW)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_relationTypeAccess_GetRelationTypesByIDs(t *testing.T) {
	Convey("test GetRelationTypesByIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRelationTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
			"f_kn_id, f_branch, f_source_object_type_id, f_target_object_type_id, f_type, f_mapping_rules, "+
			"f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_kn_id = ? AND f_branch = ? AND f_id IN (?,?)", RT_TABLE_NAME)

		knID := "kn1"
		branch := "main"
		rtIDs := []string{"rt1", "rt2"}

		mappingRulesBytes, _ := sonic.Marshal(testRelationType.MappingRules)

		Convey("GetRelationTypesByIDs Success \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DIRECT, mappingRulesBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			relationTypes, err := rta.GetRelationTypesByIDs(testCtx, knID, branch, rtIDs)
			So(len(relationTypes), ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypesByIDs Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			relationTypes, err := rta.GetRelationTypesByIDs(testCtx, knID, branch, rtIDs)
			So(relationTypes, ShouldResemble, []*interfaces.RelationType{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypesByIDs scan error \n", func() {
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("rt1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := rta.GetRelationTypesByIDs(testCtx, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypesByIDs unmarshal MappingRules DIRECT error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DIRECT, invalidBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := rta.GetRelationTypesByIDs(testCtx, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypesByIDs unmarshal MappingRules DATA_VIEW error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DATA_VIEW, invalidBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := rta.GetRelationTypesByIDs(testCtx, knID, branch, rtIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypesByIDs Success DATA_VIEW type \n", func() {
			dataViewMapping := &interfaces.InDirectMapping{}
			dataViewMappingBytes, _ := sonic.Marshal(dataViewMapping)
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DATA_VIEW, dataViewMappingBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			relationTypes, err := rta.GetRelationTypesByIDs(testCtx, knID, branch, rtIDs)
			So(err, ShouldBeNil)
			So(len(relationTypes), ShouldEqual, 1)
			So(relationTypes[0].Type, ShouldEqual, interfaces.RELATION_TYPE_DATA_VIEW)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_relationTypeAccess_UpdateRelationType(t *testing.T) {
	Convey("Test UpdateRelationType\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRelationTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_bkn_raw_content = ?, f_color = ?, f_comment = ?, f_icon = ?, "+
			"f_mapping_rules = ?, f_name = ?, f_source_object_type_id = ?, f_tags = ?, f_target_object_type_id = ?, "+
			"f_type = ?, f_update_time = ?, f_updater = ?, f_updater_type = ? WHERE f_id = ? AND f_kn_id = ?", RT_TABLE_NAME)

		relationType := &interfaces.RelationType{
			RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
				RTID:               "rt1",
				RTName:             "Updated Relation Type",
				SourceObjectTypeID: "ot1",
				TargetObjectTypeID: "ot2",
				Type:               interfaces.RELATION_TYPE_DIRECT,
				MappingRules:       []interfaces.Mapping{},
			},
			CommonInfo: interfaces.CommonInfo{
				Tags:    testTags,
				Comment: "updated comment",
				Icon:    "icon1",
				Color:   "color1",
			},
			KNID: "kn1",
			Updater: interfaces.AccountInfo{
				ID:   "admin",
				Type: "admin",
			},
			UpdateTime: testUpdateTime,
		}

		Convey("UpdateRelationType Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := rta.db.Begin()
			err := rta.UpdateRelationType(testCtx, tx, relationType)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateRelationType failed prepare \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("prepare error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := rta.db.Begin()
			err := rta.UpdateRelationType(testCtx, tx, relationType)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateRelationType RowsAffected != 1 \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(0, 0))

			tx, _ := rta.db.Begin()
			err := rta.UpdateRelationType(testCtx, tx, relationType)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateRelationType RowsAffected error \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := rta.db.Begin()
			err := rta.UpdateRelationType(testCtx, tx, relationType)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_relationTypeAccess_DeleteRelationTypesByIDs(t *testing.T) {
	Convey("Test DeleteRelationTypesByIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRelationTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_kn_id = ? AND f_branch = ? AND f_id IN (?,?)", RT_TABLE_NAME)

		knID := "kn1"
		branch := "main"
		rtIDs := []string{"rt1", "rt2"}

		Convey("DeleteRelationTypesByIDs Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs(knID, branch, "rt1", "rt2").WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := rta.db.Begin()
			rowsAffected, err := rta.DeleteRelationTypesByIDs(testCtx, tx, knID, branch, rtIDs)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteRelationTypesByIDs null \n", func() {
			smock.ExpectBegin()

			tx, _ := rta.db.Begin()
			rowsAffected, err := rta.DeleteRelationTypesByIDs(testCtx, tx, knID, branch, []string{})
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 0)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteRelationTypesByIDs Failed dbExec \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("dbExec error")
			smock.ExpectExec(sqlStr).WithArgs(knID, branch, "rt1", "rt2").WillReturnError(expectedErr)

			tx, _ := rta.db.Begin()
			_, err := rta.DeleteRelationTypesByIDs(testCtx, tx, knID, branch, rtIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteRelationTypesByIDs RowsAffected != len(rtIDs) \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs(knID, branch, "rt1", "rt2").WillReturnResult(sqlmock.NewResult(0, 1))

			tx, _ := rta.db.Begin()
			rowsAffected, err := rta.DeleteRelationTypesByIDs(testCtx, tx, knID, branch, rtIDs)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteRelationTypesByIDs RowsAffected error \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs(knID, branch, "rt1", "rt2").WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := rta.db.Begin()
			_, err := rta.DeleteRelationTypesByIDs(testCtx, tx, knID, branch, rtIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_relationTypeAccess_GetRelationTypeIDsByKnID(t *testing.T) {
	Convey("test GetRelationTypeIDsByKnID\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRelationTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id FROM %s WHERE f_kn_id = ? AND f_branch = ?", RT_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_id"}).
			AddRow("rt1").
			AddRow("rt2").
			AddRow("rt3")

		knID := "kn1"
		branch := "main"

		Convey("GetRelationTypeIDsByKnID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			rtIDs, err := rta.GetRelationTypeIDsByKnID(testCtx, knID, branch)
			So(err, ShouldBeNil)
			So(len(rtIDs), ShouldEqual, 3)
			So(rtIDs[0], ShouldEqual, "rt1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypeIDsByKnID Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(sqlmock.NewRows(nil))

			rtIDs, err := rta.GetRelationTypeIDsByKnID(testCtx, knID, branch)
			So(rtIDs, ShouldResemble, []string{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypeIDsByKnID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnError(expectedErr)

			rtIDs, err := rta.GetRelationTypeIDsByKnID(testCtx, knID, branch)
			So(rtIDs, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetRelationTypeIDsByKnID Scan error \n", func() {
			rows := sqlmock.NewRows([]string{"f_id", "f_id"}).
				AddRow("rt1", "rt1")

			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			rtIDs, err := rta.GetRelationTypeIDsByKnID(testCtx, knID, branch)
			So(rtIDs, ShouldBeNil)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_relationTypeAccess_GetAllRelationTypesByKnID(t *testing.T) {
	Convey("test GetAllRelationTypesByKnID\n", t, func() {
		appSetting := &common.AppSetting{}
		rta, smock := MockNewRelationTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
			"f_kn_id, f_branch, f_source_object_type_id, f_target_object_type_id, f_type, f_mapping_rules, "+
			"f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_kn_id = ? AND f_branch = ?", RT_TABLE_NAME)

		mappingRulesBytes, _ := sonic.Marshal(testRelationType.MappingRules)

		rows := sqlmock.NewRows([]string{
			"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
			"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
			"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
		}).AddRow(
			"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
			"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DIRECT, mappingRulesBytes,
			"admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		).AddRow(
			"rt2", "Relation Type 2", `"tag2"`, "comment2", "icon2", "color2", "detail2",
			"kn1", "main", "ot2", "ot3", interfaces.RELATION_TYPE_DIRECT, mappingRulesBytes,
			"admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		)

		knID := "kn1"
		branch := "main"

		Convey("GetAllRelationTypesByKnID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			relationTypes, err := rta.GetAllRelationTypesByKnID(testCtx, knID, branch)
			So(err, ShouldBeNil)
			So(len(relationTypes), ShouldEqual, 2)
			So(relationTypes["rt1"], ShouldNotBeNil)
			So(relationTypes["rt2"], ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllRelationTypesByKnID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnError(expectedErr)

			relationTypes, err := rta.GetAllRelationTypesByKnID(testCtx, knID, branch)
			So(relationTypes, ShouldResemble, map[string]*interfaces.RelationType{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllRelationTypesByKnID Scan error \n", func() {
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DIRECT, mappingRulesBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime, "f_update_time",
			)

			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			relationTypes, err := rta.GetAllRelationTypesByKnID(testCtx, knID, branch)
			So(relationTypes, ShouldResemble, map[string]*interfaces.RelationType{})
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllRelationTypesByKnID Unmarshal mappingRules error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
				"f_creator", "f_creator_type", "f_create_time", "f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"rt1", "Relation Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", "ot1", "ot2", interfaces.RELATION_TYPE_DIRECT, invalidBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			relationTypes, err := rta.GetAllRelationTypesByKnID(testCtx, knID, branch)
			So(relationTypes, ShouldResemble, map[string]*interfaces.RelationType{})
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_relationTypeAccess_ProcessQueryCondition(t *testing.T) {
	Convey("test processQueryCondition ", t, func() {
		appSetting := &common.AppSetting{}
		_, _ = MockNewRelationTypeAccess(appSetting)

		sqlBuilder := sq.Select("COUNT(f_id)").From(RT_TABLE_NAME)

		Convey("NamePattern query ", func() {
			query := interfaces.RelationTypesQueryParams{
				NamePattern: "name_a",
				KNID:        "kn1",
				Branch:      "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE (instr(f_name, ?) > 0 OR instr(f_id, ?) > 0) AND f_kn_id = ? AND f_branch = ?", RT_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("Tag query ", func() {
			query := interfaces.RelationTypesQueryParams{
				Tag:    "tag1",
				KNID:   "kn1",
				Branch: "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE instr(f_tags, ?) > 0 AND f_kn_id = ? AND f_branch = ?", RT_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("SourceObjectTypeIDs query ", func() {
			query := interfaces.RelationTypesQueryParams{
				SourceObjectTypeIDs: []string{"ot1", "ot2"},
				KNID:                "kn1",
				Branch:              "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE f_kn_id = ? AND f_branch = ? AND f_source_object_type_id IN (?,?)", RT_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("TargetObjectTypeIDs query ", func() {
			query := interfaces.RelationTypesQueryParams{
				TargetObjectTypeIDs: []string{"ot1", "ot2"},
				KNID:                "kn1",
				Branch:              "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE f_kn_id = ? AND f_branch = ? AND f_target_object_type_id IN (?,?)", RT_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("Empty Branch query ", func() {
			query := interfaces.RelationTypesQueryParams{
				KNID:   "kn1",
				Branch: "",
			}

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, vals, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldContainSubstring, "f_branch = ?")
			So(len(vals), ShouldBeGreaterThan, 0)
		})
	})
}
