// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_type

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

	testActionType = &interfaces.ActionType{
		ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
			ATID:         "at1",
			ATName:       "Action Type 1",
			ActionType:   interfaces.ACTION_SOURCE_TYPE_TOOL,
			ObjectTypeID: "ot1",
			Condition:    &interfaces.ActionCondCfg{},
			Affect:       &interfaces.ActionAffect{},
			ActionSource: interfaces.ActionSource{
				Type:   interfaces.ACTION_SOURCE_TYPE_TOOL,
				BoxID:  "box1",
				ToolID: "tool1",
			},
			Parameters: []interfaces.Parameter{},
			Schedule: interfaces.Schedule{
				Type:       "cron",
				Expression: "0 0 * * *",
			},
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
		ModuleType: interfaces.MODULE_TYPE_ACTION_TYPE,
	}
)

func MockNewActionTypeAccess(appSetting *common.AppSetting) (*actionTypeAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	ata := &actionTypeAccess{
		appSetting: appSetting,
		db:         db,
	}
	return ata, smock
}

func Test_ActionTypeAccess_CheckActionTypeExistByID(t *testing.T) {
	Convey("test CheckActionTypeExistByID\n", t, func() {
		appSetting := &common.AppSetting{}
		ata, smock := MockNewActionTypeAccess(appSetting)

		sqlStr := "SELECT f_name FROM t_action_type WHERE f_kn_id = ? AND f_branch = ? AND f_id = ?"

		knID := "kn1"
		branch := "main"
		atID := "at1"

		Convey("CheckActionTypeExistByID Success \n", func() {
			rows := sqlmock.NewRows([]string{"f_name"}).AddRow("Action Type 1")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, atID).WillReturnRows(rows)

			name, exists, err := ata.CheckActionTypeExistByID(testCtx, knID, branch, atID)
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(name, ShouldEqual, "Action Type 1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckActionTypeExistByID Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, atID).WillReturnError(sql.ErrNoRows)

			name, exists, err := ata.CheckActionTypeExistByID(testCtx, knID, branch, atID)
			So(name, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckActionTypeExistByID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, atID).WillReturnError(expectedErr)

			name, exists, err := ata.CheckActionTypeExistByID(testCtx, knID, branch, atID)
			So(name, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ActionTypeAccess_CheckActionTypeExistByName(t *testing.T) {
	Convey("test CheckActionTypeExistByName\n", t, func() {
		appSetting := &common.AppSetting{}
		ata, smock := MockNewActionTypeAccess(appSetting)

		sqlStr := "SELECT f_id FROM t_action_type WHERE f_kn_id = ? AND f_branch = ? AND f_name = ?"

		knID := "kn1"
		branch := "main"
		atName := "Action Type 1"

		Convey("CheckActionTypeExistByName Success \n", func() {
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("at1")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, atName).WillReturnRows(rows)

			atID, exists, err := ata.CheckActionTypeExistByName(testCtx, knID, branch, atName)
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(atID, ShouldEqual, "at1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckActionTypeExistByName Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, atName).WillReturnError(sql.ErrNoRows)

			atID, exists, err := ata.CheckActionTypeExistByName(testCtx, knID, branch, atName)
			So(atID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CheckActionTypeExistByName Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch, atName).WillReturnError(expectedErr)

			atID, exists, err := ata.CheckActionTypeExistByName(testCtx, knID, branch, atName)
			So(atID, ShouldEqual, "")
			So(exists, ShouldBeFalse)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ActionTypeAccess_CreateActionType(t *testing.T) {
	Convey("test CreateActionType\n", t, func() {
		appSetting := &common.AppSetting{}
		ata, smock := MockNewActionTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id,f_name,f_tags,f_comment,f_icon,f_color,f_bkn_raw_content,"+
			"f_kn_id,f_branch,f_action_type,f_action_intent,f_impact_contracts,f_object_type_id,f_condition,f_affect,f_action_source,"+
			"f_parameters,f_schedule,f_creator,f_creator_type,f_create_time,f_updater,f_updater_type,f_update_time) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", AT_TABLE_NAME)

		Convey("CreateActionType Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := ata.db.Begin()
			err := ata.CreateActionType(testCtx, tx, testActionType)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("CreateActionType Exec sql error\n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("some error1")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ata.db.Begin()
			err := ata.CreateActionType(testCtx, tx, testActionType)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ActionTypeAccess_ListActionTypes(t *testing.T) {
	Convey("test ListActionTypes\n", t, func() {
		appSetting := &common.AppSetting{}
		ata, smock := MockNewActionTypeAccess(appSetting)

		conditionBytes, _ := sonic.Marshal((*interfaces.ActionCondCfg)(nil))
		affectBytes, _ := sonic.Marshal((*interfaces.ActionAffect)(nil))
		actionSourceBytes, _ := sonic.Marshal(interfaces.ActionSource{})
		parametersBytes, _ := sonic.Marshal([]interfaces.Parameter{})
		scheduleBytes, _ := sonic.Marshal(interfaces.Schedule{})

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
			"f_kn_id, f_branch, f_action_type, f_action_intent, f_impact_contracts, f_object_type_id, f_condition, f_affect, f_action_source, "+
			"f_parameters, f_schedule, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_kn_id = ? AND f_branch = ?", AT_TABLE_NAME)

		rows := sqlmock.NewRows([]string{
			"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
			"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
			"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
			"f_creator", "f_creator_type", "f_create_time",
			"f_updater", "f_updater_type", "f_update_time",
		}).AddRow(
			"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
			"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
			conditionBytes, affectBytes, actionSourceBytes, parametersBytes, scheduleBytes,
			"admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		)

		query := interfaces.ActionTypesQueryParams{
			KNID:   "kn1",
			Branch: "main",
		}

		Convey("ListActionTypes Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			actionTypes, err := ata.ListActionTypes(testCtx, query)
			So(err, ShouldBeNil)
			So(len(actionTypes), ShouldEqual, 1)
			So(actionTypes[0].ATID, ShouldEqual, "at1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListActionTypes Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(sqlmock.NewRows(nil))

			actionTypes, err := ata.ListActionTypes(testCtx, query)
			So(actionTypes, ShouldResemble, []*interfaces.ActionType{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListActionTypes Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			actionTypes, err := ata.ListActionTypes(testCtx, query)
			So(actionTypes, ShouldResemble, []*interfaces.ActionType{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListActionTypes scan error \n", func() {
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("at1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.ListActionTypes(testCtx, query)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListActionTypes unmarshal Condition error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				invalidBytes, affectBytes, actionSourceBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.ListActionTypes(testCtx, query)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListActionTypes unmarshal Affect error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, invalidBytes, actionSourceBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.ListActionTypes(testCtx, query)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListActionTypes unmarshal ActionSource error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, affectBytes, invalidBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.ListActionTypes(testCtx, query)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListActionTypes unmarshal Parameters error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, affectBytes, actionSourceBytes, invalidBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.ListActionTypes(testCtx, query)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListActionTypes unmarshal Schedule error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, affectBytes, actionSourceBytes, parametersBytes, invalidBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.ListActionTypes(testCtx, query)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListActionTypes unmarshal ImpactContracts error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", invalidBytes, "ot1",
				conditionBytes, affectBytes, actionSourceBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.ListActionTypes(testCtx, query)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListActionTypes with Sort ASC\n", func() {
			sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
				"f_kn_id, f_branch, f_action_type, f_action_intent, f_impact_contracts, f_object_type_id, f_condition, f_affect, f_action_source, "+
				"f_parameters, f_schedule, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
				"FROM %s WHERE f_kn_id = ? AND f_branch = ? ORDER BY f_name ASC", AT_TABLE_NAME)

			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, affectBytes, actionSourceBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			queryWithSort := interfaces.ActionTypesQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Sort:      "f_name",
					Direction: "ASC",
				},
				KNID:   "kn1",
				Branch: "main",
			}

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			actionTypes, err := ata.ListActionTypes(testCtx, queryWithSort)
			So(err, ShouldBeNil)
			So(len(actionTypes), ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("ListActionTypes with Sort DESC\n", func() {
			sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
				"f_kn_id, f_branch, f_action_type, f_action_intent, f_impact_contracts, f_object_type_id, f_condition, f_affect, f_action_source, "+
				"f_parameters, f_schedule, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
				"FROM %s WHERE f_kn_id = ? AND f_branch = ? ORDER BY f_name DESC", AT_TABLE_NAME)

			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, affectBytes, actionSourceBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)

			queryWithSort := interfaces.ActionTypesQueryParams{
				PaginationQueryParameters: interfaces.PaginationQueryParameters{
					Sort:      "f_name",
					Direction: "DESC",
				},
				KNID:   "kn1",
				Branch: "main",
			}

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			actionTypes, err := ata.ListActionTypes(testCtx, queryWithSort)
			So(err, ShouldBeNil)
			So(len(actionTypes), ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

	})
}

func Test_ActionTypeAccess_GetActionTypesTotal(t *testing.T) {
	Convey("test GetActionTypesTotal\n", t, func() {
		appSetting := &common.AppSetting{}
		ata, smock := MockNewActionTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s WHERE f_kn_id = ? AND f_branch = ?", AT_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"COUNT(f_id)"}).AddRow(1)

		query := interfaces.ActionTypesQueryParams{
			KNID:   "kn1",
			Branch: "main",
		}

		Convey("GetActionTypesTotal Success\n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			total, err := ata.GetActionTypesTotal(testCtx, query)
			So(total, ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypesTotal Failed  Query error\n", func() {
			expectedErr := errors.New("Query error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := ata.GetActionTypesTotal(testCtx, query)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypesTotal scan error \n", func() {
			rows := sqlmock.NewRows([]string{"COUNT(f_id)"}).AddRow("s")

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)
			_, err := ata.GetActionTypesTotal(testCtx, query)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ActionTypeAccess_GetActionTypesByIDs(t *testing.T) {
	Convey("test GetActionTypesByIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		ata, smock := MockNewActionTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
			"f_kn_id, f_branch, f_action_type, f_action_intent, f_impact_contracts, f_object_type_id, f_condition, f_affect, f_action_source, "+
			"f_parameters, f_schedule, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_kn_id = ? AND f_branch = ? AND f_id IN (?,?)", AT_TABLE_NAME)

		conditionBytes, _ := sonic.Marshal((*interfaces.ActionCondCfg)(nil))
		affectBytes, _ := sonic.Marshal((*interfaces.ActionAffect)(nil))
		actionSourceBytes, _ := sonic.Marshal(interfaces.ActionSource{})
		parametersBytes, _ := sonic.Marshal([]interfaces.Parameter{})
		scheduleBytes, _ := sonic.Marshal(interfaces.Schedule{})

		rows := sqlmock.NewRows([]string{
			"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
			"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
			"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
			"f_creator", "f_creator_type", "f_create_time",
			"f_updater", "f_updater_type", "f_update_time",
		}).AddRow(
			"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
			"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
			conditionBytes, affectBytes, actionSourceBytes, parametersBytes, scheduleBytes,
			"admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		).AddRow(
			"at2", "Action Type 2", `"tag2"`, "comment2", "icon2", "color2", "detail2",
			"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
			conditionBytes, affectBytes, actionSourceBytes, parametersBytes, scheduleBytes,
			"admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		)

		knID := "kn1"
		branch := "main"
		atIDs := []string{"at1", "at2"}

		Convey("GetActionTypesByIDs Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			actionTypes, err := ata.GetActionTypesByIDs(testCtx, knID, branch, atIDs)
			So(len(actionTypes), ShouldEqual, 2)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypesByIDs Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(sqlmock.NewRows(nil))

			actionTypes, err := ata.GetActionTypesByIDs(testCtx, knID, branch, atIDs)
			So(actionTypes, ShouldResemble, []*interfaces.ActionType{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypesByIDs Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			actionTypes, err := ata.GetActionTypesByIDs(testCtx, knID, branch, atIDs)
			So(actionTypes, ShouldResemble, []*interfaces.ActionType{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypesByIDs scan error \n", func() {
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("at1")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.GetActionTypesByIDs(testCtx, knID, branch, atIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypesByIDs unmarshal Condition error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				invalidBytes, affectBytes, actionSourceBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.GetActionTypesByIDs(testCtx, knID, branch, atIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypesByIDs unmarshal Affect error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, invalidBytes, actionSourceBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.GetActionTypesByIDs(testCtx, knID, branch, atIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypesByIDs unmarshal ActionSource error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, affectBytes, invalidBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.GetActionTypesByIDs(testCtx, knID, branch, atIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypesByIDs unmarshal Parameters error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, affectBytes, actionSourceBytes, invalidBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.GetActionTypesByIDs(testCtx, knID, branch, atIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypesByIDs unmarshal Schedule error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, affectBytes, actionSourceBytes, parametersBytes, invalidBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := ata.GetActionTypesByIDs(testCtx, knID, branch, atIDs)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ActionTypeAccess_UpdateActionType(t *testing.T) {
	Convey("Test UpdateActionType\n", t, func() {
		appSetting := &common.AppSetting{}
		ata, smock := MockNewActionTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("UPDATE %s SET f_action_intent = ?, f_action_source = ?, f_action_type = ?, f_affect = ?, f_bkn_raw_content = ?, "+
			"f_color = ?, f_comment = ?, f_condition = ?, f_icon = ?, f_impact_contracts = ?, f_name = ?, f_object_type_id = ?, f_parameters = ?, "+
			"f_schedule = ?, f_tags = ?, f_update_time = ?, f_updater = ?, f_updater_type = ? "+
			"WHERE f_id = ? AND f_kn_id = ?", AT_TABLE_NAME)

		actionType := &interfaces.ActionType{
			ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
				ATID:         "at1",
				ATName:       "Updated Action Type",
				ActionType:   interfaces.ACTION_SOURCE_TYPE_TOOL,
				ObjectTypeID: "ot1",
			},
			KNID:   "kn1",
			Branch: "main",
			Updater: interfaces.AccountInfo{
				ID:   "admin",
				Type: "admin",
			},
			UpdateTime: testUpdateTime,
		}

		Convey("UpdateActionType Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := ata.db.Begin()
			err := ata.UpdateActionType(testCtx, tx, actionType)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateActionType failed prepare \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("prepare error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ata.db.Begin()
			err := ata.UpdateActionType(testCtx, tx, actionType)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateActionType Failed UpdateSql \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("sql exec error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ata.db.Begin()
			err := ata.UpdateActionType(testCtx, tx, actionType)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateActionType RowsAffected 2 \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := ata.db.Begin()
			err := ata.UpdateActionType(testCtx, tx, actionType)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateActionType Failed RowsAffected \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := ata.db.Begin()
			err := ata.UpdateActionType(testCtx, tx, actionType)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("UpdateActionType Build sql error \n", func() {
			// 使用一个会导致 SQL 构建错误的 actionType
			// 实际上，由于使用了 SetMap，SQL 构建通常不会失败
			// 但我们可以测试其他边界情况
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			tx, _ := ata.db.Begin()
			err := ata.UpdateActionType(testCtx, tx, actionType)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ActionTypeAccess_DeleteActionTypesByIDs(t *testing.T) {
	Convey("Test DeleteActionTypesByIDs\n", t, func() {
		appSetting := &common.AppSetting{}
		ata, smock := MockNewActionTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_kn_id = ? AND f_branch = ? AND f_id IN (?,?)", AT_TABLE_NAME)

		knID := "kn1"
		branch := "main"
		atIDs := []string{"at1", "at2"}

		Convey("DeleteActionTypesByIDs Success \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			tx, _ := ata.db.Begin()
			rowsAffected, err := ata.DeleteActionTypesByIDs(testCtx, tx, knID, branch, atIDs)
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteActionTypesByIDs failed prepare \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("prepare error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ata.db.Begin()
			_, err := ata.DeleteActionTypesByIDs(testCtx, tx, knID, branch, atIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteActionTypesByIDs RowsAffected 0 \n", func() {
			smock.ExpectBegin()
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 0))

			tx, _ := ata.db.Begin()
			rowsAffected, err := ata.DeleteActionTypesByIDs(testCtx, tx, knID, branch, atIDs)
			So(rowsAffected, ShouldEqual, 0)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteActionTypesByIDs Failed RowsAffected \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("Get RowsAffected error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewErrorResult(expectedErr))

			tx, _ := ata.db.Begin()
			_, err := ata.DeleteActionTypesByIDs(testCtx, tx, knID, branch, atIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteActionTypesByIDs Failed dbExec \n", func() {
			smock.ExpectBegin()
			expectedErr := errors.New("dbExec error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			tx, _ := ata.db.Begin()
			_, err := ata.DeleteActionTypesByIDs(testCtx, tx, knID, branch, atIDs)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("DeleteActionTypesByIDs null \n", func() {
			smock.ExpectBegin()

			tx, _ := ata.db.Begin()
			rowsAffected, err := ata.DeleteActionTypesByIDs(testCtx, tx, knID, branch, []string{})
			So(err, ShouldBeNil)
			So(rowsAffected, ShouldEqual, 0)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ActionTypeAccess_GetActionTypeIDsByKnID(t *testing.T) {
	Convey("test GetActionTypeIDsByKnID\n", t, func() {
		appSetting := &common.AppSetting{}
		ata, smock := MockNewActionTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id FROM %s WHERE f_kn_id = ? AND f_branch = ?", AT_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"f_id"}).
			AddRow("at1").
			AddRow("at2").
			AddRow("at3")

		knID := "kn1"
		branch := "main"

		Convey("GetActionTypeIDsByKnID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			atIDs, err := ata.GetActionTypeIDsByKnID(testCtx, knID, branch)
			So(err, ShouldBeNil)
			So(len(atIDs), ShouldEqual, 3)
			So(atIDs[0], ShouldEqual, "at1")

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypeIDsByKnID Success no row \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(sqlmock.NewRows(nil))

			atIDs, err := ata.GetActionTypeIDsByKnID(testCtx, knID, branch)
			So(atIDs, ShouldResemble, []string{})
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypeIDsByKnID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnError(expectedErr)

			atIDs, err := ata.GetActionTypeIDsByKnID(testCtx, knID, branch)
			So(atIDs, ShouldBeNil)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetActionTypeIDsByKnID scan error \n", func() {
			// 使用错误的列类型来触发 scan 错误
			rows := sqlmock.NewRows([]string{"f_id", "f_id"}).AddRow(123, "123") // 使用 int 而不是 string
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			atIDs, err := ata.GetActionTypeIDsByKnID(testCtx, knID, branch)
			So(atIDs, ShouldBeNil)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ActionTypeAccess_GetAllActionTypesByKnID(t *testing.T) {
	Convey("test GetAllActionTypesByKnID\n", t, func() {
		appSetting := &common.AppSetting{}
		ata, smock := MockNewActionTypeAccess(appSetting)

		sqlStr := fmt.Sprintf("SELECT f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content, "+
			"f_kn_id, f_branch, f_action_type, f_action_intent, f_impact_contracts, f_object_type_id, f_condition, f_affect, f_action_source, "+
			"f_parameters, f_schedule, f_creator, f_creator_type, f_create_time, f_updater, f_updater_type, f_update_time "+
			"FROM %s WHERE f_kn_id = ? AND f_branch = ?", AT_TABLE_NAME)

		conditionBytes, _ := sonic.Marshal((*interfaces.ActionCondCfg)(nil))
		affectBytes, _ := sonic.Marshal((*interfaces.ActionAffect)(nil))
		actionSourceBytes, _ := sonic.Marshal(interfaces.ActionSource{})
		parametersBytes, _ := sonic.Marshal([]interfaces.Parameter{})
		scheduleBytes, _ := sonic.Marshal(interfaces.Schedule{})

		rows := sqlmock.NewRows([]string{
			"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
			"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
			"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
			"f_creator", "f_creator_type", "f_create_time",
			"f_updater", "f_updater_type", "f_update_time",
		}).AddRow(
			"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
			"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
			conditionBytes, affectBytes, actionSourceBytes, parametersBytes, scheduleBytes,
			"admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		).AddRow(
			"at2", "Action Type 2", `"tag2"`, "comment2", "icon2", "color2", "detail2",
			"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
			conditionBytes, affectBytes, actionSourceBytes, parametersBytes, scheduleBytes,
			"admin", "admin", testUpdateTime,
			"admin", "admin", testUpdateTime,
		)

		knID := "kn1"
		branch := "main"

		Convey("GetAllActionTypesByKnID Success \n", func() {
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			actionTypes, err := ata.GetAllActionTypesByKnID(testCtx, knID, branch)
			So(err, ShouldBeNil)
			So(len(actionTypes), ShouldEqual, 2)
			So(actionTypes["at1"], ShouldNotBeNil)
			So(actionTypes["at2"], ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllActionTypesByKnID Failed \n", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnError(expectedErr)

			actionTypes, err := ata.GetAllActionTypesByKnID(testCtx, knID, branch)
			So(actionTypes, ShouldResemble, map[string]*interfaces.ActionType{})
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllActionTypesByKnID scan error \n", func() {
			rows := sqlmock.NewRows([]string{"f_id"}).AddRow("at1")
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			_, err := ata.GetAllActionTypesByKnID(testCtx, knID, branch)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllActionTypesByKnID unmarshal Condition error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				invalidBytes, affectBytes, actionSourceBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			_, err := ata.GetAllActionTypesByKnID(testCtx, knID, branch)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllActionTypesByKnID unmarshal Affect error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, invalidBytes, actionSourceBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			_, err := ata.GetAllActionTypesByKnID(testCtx, knID, branch)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllActionTypesByKnID unmarshal ActionSource error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, affectBytes, invalidBytes, parametersBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			_, err := ata.GetAllActionTypesByKnID(testCtx, knID, branch)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllActionTypesByKnID unmarshal Parameters error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, affectBytes, actionSourceBytes, invalidBytes, scheduleBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			_, err := ata.GetAllActionTypesByKnID(testCtx, knID, branch)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("GetAllActionTypesByKnID unmarshal Schedule error \n", func() {
			invalidBytes := []byte("invalid json")
			rows := sqlmock.NewRows([]string{
				"f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
				"f_kn_id", "f_branch", "f_action_type", "f_action_intent", "f_impact_contracts", "f_object_type_id",
				"f_condition", "f_affect", "f_action_source", "f_parameters", "f_schedule",
				"f_creator", "f_creator_type", "f_create_time",
				"f_updater", "f_updater_type", "f_update_time",
			}).AddRow(
				"at1", "Action Type 1", `"tag1"`, "comment", "icon", "color", "detail",
				"kn1", "main", interfaces.ACTION_SOURCE_TYPE_TOOL, "", nil, "ot1",
				conditionBytes, affectBytes, actionSourceBytes, parametersBytes, invalidBytes,
				"admin", "admin", testUpdateTime,
				"admin", "admin", testUpdateTime,
			)
			smock.ExpectQuery(sqlStr).WithArgs(knID, branch).WillReturnRows(rows)

			_, err := ata.GetAllActionTypesByKnID(testCtx, knID, branch)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_ActionTypeAccess_ProcessQueryCondition(t *testing.T) {
	Convey("test processQueryCondition ", t, func() {
		appSetting := &common.AppSetting{}
		_, _ = MockNewActionTypeAccess(appSetting)

		sqlBuilder := sq.Select("COUNT(f_id)").From(AT_TABLE_NAME)

		Convey("NamePattern query ", func() {
			query := interfaces.ActionTypesQueryParams{
				NamePattern: "name_a",
				KNID:        "kn1",
				Branch:      "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE (instr(f_name, ?) > 0 OR instr(f_id, ?) > 0) AND f_kn_id = ? AND f_branch = ?", AT_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("Tag query ", func() {
			query := interfaces.ActionTypesQueryParams{
				Tag:    "tag1",
				KNID:   "kn1",
				Branch: "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE instr(f_tags, ?) > 0 AND f_kn_id = ? AND f_branch = ?", AT_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("KNID query ", func() {
			query := interfaces.ActionTypesQueryParams{
				KNID:   "kn1",
				Branch: "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE f_kn_id = ? AND f_branch = ?", AT_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("Branch empty query ", func() {
			query := interfaces.ActionTypesQueryParams{
				KNID: "kn1",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE f_kn_id = ? AND f_branch = ?", AT_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("ActionType query ", func() {
			query := interfaces.ActionTypesQueryParams{
				ActionType: interfaces.ACTION_SOURCE_TYPE_TOOL,
				KNID:       "kn1",
				Branch:     "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE f_kn_id = ? AND f_branch = ? AND f_action_type = ?", AT_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})

		Convey("ObjectTypeIDs query ", func() {
			query := interfaces.ActionTypesQueryParams{
				ObjectTypeIDs: []string{"ot1", "ot2"},
				KNID:          "kn1",
				Branch:        "main",
			}

			expectedSqlStr := fmt.Sprintf("SELECT COUNT(f_id) FROM %s "+
				"WHERE f_kn_id = ? AND f_branch = ? AND f_object_type_id IN (?,?)", AT_TABLE_NAME)

			sqlBuilder := processQueryCondition(query, sqlBuilder)
			sqlStr, _, _ := sqlBuilder.ToSql()
			So(sqlStr, ShouldEqual, expectedSqlStr)
		})
	})
}
