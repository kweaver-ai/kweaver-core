package dbaccess

// t_api_metadata 表单元测试

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	errMock = errors.New("mock db error")
	dbName  = "dip_data_operator_hub"
)

func TestNewAPIMetadataDB(t *testing.T) {
	Convey("TestNewAPIMetadataDB:初始化API元数据DB", t, func() {
		mockDB, _, err := sqlx.New()
		So(err, ShouldBeNil)
		defer func() {
			_ = mockDB.Close()
		}()
		p := gomonkey.ApplyFunc(config.NewConfigLoader, func() *config.Config {
			return &config.Config{
				DB: config.DBConfig{
					DBName: dbName,
				},
			}
		})
		defer p.Reset()
		p1 := gomonkey.ApplyFunc(db.NewDBPool, func() *sqlx.DB {
			return mockDB
		})
		defer p1.Reset()
		_ = NewAPIMetadataDB()
	})
}

func TestInsertAPIMetadata(t *testing.T) {
	Convey("TestInsertAPIMetadata:插入API元数据", t, func() {
		mockDB, mock, err := sqlx.New()
		So(err, ShouldBeNil)
		defer func() {
			_ = mockDB.Close()
		}()
		am := &apiMetadataDB{
			dbPool: mockDB,
			dbName: dbName,
		}
		query := "INSERT INTO `dip_data_operator_hub`.`t_metadata_api`(`f_summary`, `f_version`, `f_description`, " +
			"`f_path`, `f_svc_url`, `f_method`, `f_api_spec`, `f_create_user`, `f_create_time`,`f_update_user`,`f_update_time`)VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		metadata := &model.APIMetadataDB{}
		Convey("成功插入,新增1行", func() {
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(1, 1))
			_, err = am.InsertAPIMetadata(context.Background(), nil, metadata)
			So(err, ShouldBeNil)
		})
		Convey("插入失败，无新增", func() {
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 0))
			_, err = am.InsertAPIMetadata(context.Background(), nil, &model.APIMetadataDB{})
			So(err, ShouldNotBeNil)
		})
		Convey("插入失败", func() {
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(errMock)
			_, err := am.InsertAPIMetadata(context.Background(), nil, &model.APIMetadataDB{})
			So(err, ShouldNotBeNil)
		})
	})
}

func TestSelectByVersion(t *testing.T) {
	Convey("TestSelectByVersion:根据版本查询", t, func() {
		mockDB, mock, err := sqlx.New()
		So(err, ShouldBeNil)
		defer func() {
			_ = mockDB.Close()
		}()
		am := &apiMetadataDB{
			dbPool: mockDB,
			dbName: dbName,
		}
		fieldsKey := []string{
			`f_summary`,
			`f_version`,
			`f_description`,
			`f_path`,
			`f_svc_url`,
			`f_method`,
			`f_api_spec`,
			`f_create_user`,
			`f_create_time`,
			`f_update_user`,
			`f_update_time`,
		}
		query := "SELECT `f_summary`, `f_version`, `f_description`, `f_path`, `f_svc_url`, `f_method`, `f_api_spec`, " +
			"`f_create_user`, `f_create_time`, `f_update_user`, `f_update_time` FROM `dip_data_operator_hub`.`t_metadata_api` WHERE `f_version` = ?"
		Convey("查询成功:不存在", func() {
			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(sql.ErrNoRows)
			has, _, err := am.SelectByVersion(context.Background(), "v1")
			So(err, ShouldBeNil)
			So(has, ShouldBeFalse)
		})
		Convey("查询成功:存在", func() {
			mock.ExpectQuery(regexp.QuoteMeta(query)).
				WillReturnRows(sqlmock.NewRows(fieldsKey).
					AddRow("summary", "v1", "description", "path", "svc_url", "method", "api_spec",
						"create_user", 11111111, "update_user", 11111111))
			has, _, err := am.SelectByVersion(context.Background(), "v1")
			So(err, ShouldBeNil)
			So(has, ShouldBeTrue)
		})
		Convey("查询失败", func() {
			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errMock)
			_, _, err := am.SelectByVersion(context.Background(), "v1")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestUpdateByVersion(t *testing.T) {
	Convey("TestUpdateByVersion:根据版本更新", t, func() {
		mockDB, mock, err := sqlx.New()
		So(err, ShouldBeNil)
		defer func() {
			_ = mockDB.Close()
		}()
		am := &apiMetadataDB{
			dbPool: mockDB,
			dbName: dbName,
		}
		query := "UPDATE `dip_data_operator_hub`.`t_metadata_api` SET `f_summary` = ?, `f_description` = ?, `f_path` = ?, `f_svc_url` = ?, `f_method` = ?, `f_api_spec` = ?, " +
			"`f_update_user` = ?, `f_update_time` = ? WHERE `f_version` = ?"
		Convey("更新成功", func() {
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 1))
			err := am.UpdateByVersion(context.Background(), nil, "v1", &model.APIMetadataDB{})
			So(err, ShouldBeNil)
		})
		Convey("更新失败", func() {
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(errMock)
			err := am.UpdateByVersion(context.Background(), nil, "v1", &model.APIMetadataDB{})
			So(err, ShouldNotBeNil)
		})
	})
}

// t_op_register 表单元测试

// var (
// 	columns = []string{`f_op_id`, `f_name`, `f_metadata_version`, `f_metadata_type`, `f_status`, `f_operator_type`,
// 		`f_execution_mode`, `f_category`, `f_execute_control`,
// 		`f_extend_info`, `f_create_user`, `f_create_time`, `f_update_user`, `f_update_time`, `f_is_latest`, `f_source`}
// 	getValues = func() []driver.Value {
// 		var values []driver.Value
// 		values = append(values, "op_id", "name", "v1", "metadata_type", "status", "operator_type",
// 			"execution_mode", "category",
// 			"execute_control", "extend_info", "create_user", time.Now().UnixNano(), "update_user", time.Now().UnixNano(), true, "source")
// 		return values
// 	}
// )

func TestNewOperatorRegisterDB(t *testing.T) {
	Convey("TestNewOperatorRegisterDB:初始化Operator注册DB", t, func() {
		mockDB, _, err := sqlx.New()
		So(err, ShouldBeNil)
		defer func() {
			_ = mockDB.Close()
		}()
		p := gomonkey.ApplyFunc(config.NewConfigLoader, func() *config.Config {
			return &config.Config{
				DB: config.DBConfig{
					DBName: dbName,
				},
			}
		})
		defer p.Reset()
		p1 := gomonkey.ApplyFunc(db.NewDBPool, func() *sqlx.DB {
			return mockDB
		})
		defer p1.Reset()
		_ = NewOperatorManagerDB()
	})
}

// func TestInsertOperator(t *testing.T) {
// 	Convey("TestInsertOperator:插入Operator", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		query := "INSERT INTO `%s`.`t_op_registry` (%s)" +
// 			"VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
// 		query = fmt.Sprintf(query, dbName, queryFields)
// 		operator := &model.OperatorRegisterDB{}
// 		Convey("成功插入,新增1行", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(1, 1))
// 			_, _, err = om.InsertOperator(context.Background(), nil, operator)
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("插入失败，无新增", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 0))
// 			_, _, err = om.InsertOperator(context.Background(), nil, operator)
// 			So(err, ShouldNotBeNil)
// 		})
// 		Convey("插入失败", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			_, _, err := om.InsertOperator(context.Background(), nil, operator)
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestSelectByNameAndStatus(t *testing.T) {
// 	Convey("TestSelectByNameAndStatus:根据名称和状态查询", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		mockName := "mock name"
// 		mockStatus := string(interfaces.OperatorStatusPublished)
// 		query := fmt.Sprintf("SELECT %s FROM `%s`.`t_op_registry` WHERE `f_name` = ? AND `f_status`=? GROUP BY `f_op_id` ORDER BY `f_update_time` DESC LIMIT 1", queryFields, dbName)
// 		Convey("查询成功:不存在", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(mockName, mockStatus).WillReturnError(sql.ErrNoRows)
// 			has, _, err := om.SelectByNameAndStatus(context.Background(), nil, mockName, mockStatus)
// 			So(err, ShouldBeNil)
// 			So(has, ShouldBeFalse)
// 		})
// 		Convey("查询成功:存在", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(mockName, mockStatus).
// 				WillReturnRows(sqlmock.NewRows(columns).
// 					AddRow(getValues()...))
// 			has, _, err := om.SelectByNameAndStatus(context.Background(), nil, mockName, mockStatus)
// 			So(err, ShouldBeNil)
// 			So(has, ShouldBeTrue)
// 		})
// 		Convey("查询失败", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			_, _, err := om.SelectByNameAndStatus(context.Background(), nil, mockName, mockStatus)
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestSelectByOperatorIDAndStatus(t *testing.T) {
// 	Convey("TestSelectByOperatorIDAndStatus:根据OperatorID和状态查询", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		mockOperatorID := "mock operator id"
// 		mockStatus := string(interfaces.OperatorStatusPublished)
// 		query := fmt.Sprintf("SELECT %s FROM `%s`.`t_op_registry` WHERE `f_op_id` =? AND `f_status` = ? ORDER BY `f_update_time` DESC LIMIT 1", queryFields, dbName)
// 		Convey("查询成功:不存在", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(mockOperatorID, mockStatus).WillReturnError(sql.ErrNoRows)
// 			has, _, err := om.SelectByOperatorIDAndStatus(context.Background(), nil, mockOperatorID, mockStatus)
// 			So(err, ShouldBeNil)
// 			So(has, ShouldBeFalse)
// 		})
// 		Convey("查询成功:存在", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(mockOperatorID, mockStatus).
// 				WillReturnRows(sqlmock.NewRows(columns).
// 					AddRow(getValues()...))
// 			has, _, err := om.SelectByOperatorIDAndStatus(context.Background(), nil, mockOperatorID, mockStatus)
// 			So(err, ShouldBeNil)
// 			So(has, ShouldBeTrue)
// 		})
// 		Convey("查询失败", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			_, _, err := om.SelectByOperatorIDAndStatus(context.Background(), nil, mockOperatorID, mockStatus)
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestSelectByOperatorIDAndVersion(t *testing.T) {
// 	Convey("TestSelectByOperatorIDAndVersion:根据OperatorID和版本查询", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		mockOperatorID := "mock operator id"
// 		mockVersion := "mock version"
// 		query := fmt.Sprintf("SELECT %s FROM `%s`.`t_op_registry` WHERE `f_op_id` =? AND `f_metadata_version` =?", queryFields, dbName)
// 		Convey("查询成功:不存在", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(mockOperatorID, mockVersion).WillReturnError(sql.ErrNoRows)
// 			has, _, err := om.SelectByOperatorIDAndVersion(context.Background(), mockOperatorID, mockVersion)
// 			So(err, ShouldBeNil)
// 			So(has, ShouldBeFalse)
// 		})
// 		Convey("查询成功:存在", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(mockOperatorID, mockVersion).
// 				WillReturnRows(sqlmock.NewRows(columns).
// 					AddRow(getValues()...))
// 			has, _, err := om.SelectByOperatorIDAndVersion(context.Background(), mockOperatorID, mockVersion)
// 			So(err, ShouldBeNil)
// 			So(has, ShouldBeTrue)
// 		})
// 		Convey("查询失败", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			_, _, err := om.SelectByOperatorIDAndVersion(context.Background(), mockOperatorID, mockVersion)
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestSelectListByOperatorID(t *testing.T) {
// 	Convey("TestSelectListByOperatorID:根据OperatorID查询", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		mockOperatorID := "mock operator id"
// 		query := fmt.Sprintf("SELECT %s FROM `%s`.`t_op_registry` WHERE `f_op_id` =? ORDER BY `f_update_time` DESC", queryFields, dbName)
// 		Convey("查询成功", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(mockOperatorID).
// 				WillReturnRows(sqlmock.NewRows(columns).
// 					AddRow(getValues()...))
// 			_, err := om.SelectListByOperatorID(context.Background(), mockOperatorID)
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询失败", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			_, err := om.SelectListByOperatorID(context.Background(), mockOperatorID)
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestCountByWhereClause(t *testing.T) {
// 	Convey("TestCountByWhereClause:根据条件查询", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`t_op_registry`", dbName)
// 		Convey("查询成功:默认查询已发布最新算子", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_status`=? AND `f_is_latest`=1")).WillReturnRows(sqlmock.NewRows([]string{`COUNT(*)`}).AddRow(1))
// 			_, err := om.CountByWhereClause(context.Background(), map[string]interface{}{})
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询成功:根据f_op_id查询", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_op_id`=?")).WillReturnRows(sqlmock.NewRows([]string{`COUNT(*)`}).AddRow(1))
// 			_, err := om.CountByWhereClause(context.Background(), map[string]interface{}{
// 				"f_op_id": "mock op id",
// 			})
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询成功:根据f_op_id和version查询", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_op_id`=? AND `f_metadata_version`=?")).
// 				WillReturnRows(sqlmock.NewRows([]string{`COUNT(*)`}).AddRow(1))
// 			_, err := om.CountByWhereClause(context.Background(), map[string]interface{}{
// 				"f_op_id":            "mock op id",
// 				"f_metadata_version": "mock version",
// 			})
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询成功:根据f_create_user查询", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_create_user`=?")).
// 				WillReturnRows(sqlmock.NewRows([]string{`COUNT(*)`}).AddRow(1))
// 			_, err := om.CountByWhereClause(context.Background(), map[string]interface{}{
// 				"f_create_user": "mock create user",
// 			})
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询成功:根据f_name查询", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_name` LIKE ?")).
// 				WillReturnRows(sqlmock.NewRows([]string{`COUNT(*)`}).AddRow(1))
// 			_, err := om.CountByWhereClause(context.Background(), map[string]interface{}{
// 				"f_name": "mock name",
// 			})
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询成功:根据f_status查询,非发布状态", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_status`=?")).
// 				WillReturnRows(sqlmock.NewRows([]string{`COUNT(*)`}).AddRow(1))
// 			_, err := om.CountByWhereClause(context.Background(), map[string]interface{}{
// 				"f_status": interfaces.OperatorStatusOffline,
// 			})
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询成功:根据f_status查询,发布状态", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_status`=? AND `f_is_latest`=1")).
// 				WillReturnRows(sqlmock.NewRows([]string{`COUNT(*)`}).AddRow(1))
// 			_, err := om.CountByWhereClause(context.Background(), map[string]interface{}{
// 				"f_status": interfaces.OperatorStatusPublished,
// 			})
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询成功:根据f_category查询", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_category`=?")).
// 				WillReturnRows(sqlmock.NewRows([]string{`COUNT(*)`}).AddRow(1))
// 			_, err := om.CountByWhereClause(context.Background(), map[string]interface{}{
// 				"f_category": "mock category",
// 			})
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询成功:根据f_category、f_operator_type查询", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_category`=? AND `f_operator_type`=?")).
// 				WillReturnRows(sqlmock.NewRows([]string{`COUNT(*)`}).AddRow(1))
// 			_, err := om.CountByWhereClause(context.Background(), map[string]interface{}{
// 				"f_category":      "mock category",
// 				"f_operator_type": "mock operator type",
// 			})
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询失败", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			_, err := om.CountByWhereClause(context.Background(), map[string]interface{}{})
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestSelectListPage(t *testing.T) {
// 	Convey("TestSelectListPage:分页查询", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		query := fmt.Sprintf("SELECT %s FROM `%s`.`t_op_registry`", queryFields, dbName)
// 		Convey("默认查询已发布最新算子，根据f_update_time desc 排序，不分页", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_status`=? AND `f_is_latest`=1 ORDER BY `f_update_time` DESC")).WillReturnRows(sqlmock.NewRows(columns).
// 				AddRow(getValues()...))
// 			_, err := om.SelectListPage(context.Background(), -1, 0, map[string]interface{}{}, "", "")
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("根据名字查询，根据f_name ASE 排序，分页", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_name` LIKE ? ORDER BY `f_name` ASC LIMIT ? OFFSET ?")).WillReturnRows(sqlmock.NewRows(columns).
// 				AddRow(getValues()...))
// 			_, err := om.SelectListPage(context.Background(), 10, 0, map[string]interface{}{
// 				"f_name": "mock name",
// 			}, "name", "ASC")
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("根据状态查询，根据f_crate_time DESC 排序，分页", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query + " WHERE `f_status`=? ORDER BY `f_create_time` DESC LIMIT ? OFFSET ?")).WillReturnRows(sqlmock.NewRows(columns).
// 				AddRow(getValues()...))
// 			_, err := om.SelectListPage(context.Background(), 10, 0, map[string]interface{}{
// 				"f_status": interfaces.OperatorStatusOffline,
// 			}, "Create_time", "DESC")
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询失败", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			_, err := om.SelectListPage(context.Background(), 10, 0, map[string]interface{}{}, "", "")
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestSelectLatestByOperatorID(t *testing.T) {
// 	Convey("TestSelectLatestByOperatorID:根据OperatorID查询最新版本", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		query := fmt.Sprintf("SELECT %s FROM `%s`.`t_op_registry` WHERE `f_op_id`=? AND `f_is_latest`=1 ORDER BY `f_update_time` DESC LIMIT 1", queryFields, dbName)
// 		Convey("查询成功：存在", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows(columns).
// 				AddRow(getValues()...))
// 			has, _, err := om.SelectLatestByOperatorID(context.Background(), nil, "")
// 			So(err, ShouldBeNil)
// 			So(has, ShouldBeTrue)
// 		})
// 		Convey("查询成功：不存在", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(sql.ErrNoRows)
// 			has, _, err := om.SelectLatestByOperatorID(context.Background(), nil, "")
// 			So(err, ShouldBeNil)
// 			So(has, ShouldBeFalse)
// 		})
// 		Convey("查询失败", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			has, _, err := om.SelectLatestByOperatorID(context.Background(), nil, "")
// 			So(err, ShouldNotBeNil)
// 			So(has, ShouldBeFalse)
// 		})
// 	})
// }

// func TestSelectListPublishLatestVersion(t *testing.T) {
// 	Convey("TestSelectListPublishLatestVersion: 查询算子已发布版本最新列表", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		query := "SELECT %s FROM `%s`.`t_op_registry` WHERE `f_status` = 'published' AND `f_is_latest`=1 ORDER BY `f_update_time` DESC"
// 		query = fmt.Sprintf(query, queryFields, dbName)
// 		Convey("查询成功：存在", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows(columns).
// 				AddRow(getValues()...))
// 			_, err := om.SelectListPublishLatestVersion(context.Background())
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询失败", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			_, err := om.SelectListPublishLatestVersion(context.Background())
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestDeleteByOperatorIDAndVersion(t *testing.T) {
// 	Convey("TestDeleteByOperatorIDAndVersion: 根据id version 删除算子", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		query := "DELETE FROM `%s`.`t_op_registry` WHERE `f_op_id` = ? AND `f_metadata_version` = ?"
// 		query = fmt.Sprintf(query, dbName)
// 		Convey("删除成功", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(1, 1))
// 			err := om.DeleteByOperatorIDAndVersion(context.Background(), nil, "", "")
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("删除失败", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			err := om.DeleteByOperatorIDAndVersion(context.Background(), nil, "", "")
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestUpdateOperatorStatus(t *testing.T) {
// 	Convey("TestUpdateOperatorStatus: 更新算子状态", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		operator := &model.OperatorRegisterDB{}
// 		query := "UPDATE `%s`.`t_op_registry` SET `f_status` = ?, `f_is_latest`=?, `f_update_user` = ?, `f_update_time` = ? WHERE `f_op_id` = ? AND `f_metadata_version` = ?"
// 		query = fmt.Sprintf(query, dbName)
// 		Convey("更新成功", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(1, 1))
// 			err := om.UpdateOperatorStatus(context.Background(), nil, operator, "")
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("更新失败", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			err := om.UpdateOperatorStatus(context.Background(), nil, operator, "")
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestUpdateNameByOperatorID(t *testing.T) {
// 	Convey("TestUpdateNameByOperatorID: 更新算子名称", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		query := "UPDATE `%s`.`t_op_registry` SET `f_name` = ?, `f_update_user` = ?, `f_update_time` = ? WHERE `f_op_id` = ?"
// 		query = fmt.Sprintf(query, om.dbName)
// 		Convey("更新成功", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(1, 1))
// 			err := om.UpdateNameByOperatorID(context.Background(), nil, "", "", "")
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("更新失败", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			err := om.UpdateNameByOperatorID(context.Background(), nil, "", "", "")
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestUpdateMetadataVersionByOperatorID(t *testing.T) {
// 	Convey("TestUpdateMetadataVersionByOperatorID: 更新元数据版本", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		query := "UPDATE `%s`.`t_op_registry` SET `f_metadata_version` = ? WHERE `f_op_id` = ?"
// 		query = fmt.Sprintf(query, om.dbName)
// 		Convey("更新成功", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(1, 1))
// 			err := om.UpdateMetadataVersionByOperatorID(context.Background(), "", "")
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("更新失败", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			err := om.UpdateMetadataVersionByOperatorID(context.Background(), "", "")
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestSelectByOperatorIDAndMetadataType(t *testing.T) {
// 	Convey("TestSelectByOperatorIDAndMetadataType: 根据OperatorID和元数据类型查询", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 			},
// 		}
// 		query := fmt.Sprintf("SELECT %s FROM `%s`.`t_op_registry` WHERE `f_op_id` = ? AND `f_metadata_type` = ? ORDER BY `f_update_time` DESC", queryFields, dbName)
// 		Convey("查询成功：存在", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows(columns).
// 				AddRow(getValues()...))
// 			_, err := om.SelectByOperatorIDAndMetadataType(context.Background(), "", "")
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("查询失败", func() {
// 			mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			_, err := om.SelectByOperatorIDAndMetadataType(context.Background(), "", "")
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestUpdateByOperatorIDAndVersion(t *testing.T) {
// 	Convey("TestUpdateByOperatorIDAndVersion: 根据算子ID和版本更新算子", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 				logger: logger.DefaultLogger(),
// 			},
// 		}
// 		operator := &model.OperatorRegisterDB{}
// 		query := "UPDATE `%s`.`t_op_registry` SET `f_status` = ?, `f_operator_type`=?,  `f_execution_mode`=?, `f_category`=?, `f_execute_control`=?, " +
// 			"`f_extend_info`=?, `f_update_user`=?, `f_update_time`=?, `f_is_latest`=?, `f_source`=? WHERE `f_op_id` = ? AND `f_metadata_version` = ?"
// 		query = fmt.Sprintf(query, om.dbName)
// 		Convey("更新成功", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(1, 1))
// 			err := om.UpdateByOperatorIDAndVersion(context.Background(), nil, operator)
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("更新失败", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			err := om.UpdateByOperatorIDAndVersion(context.Background(), nil, operator)
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestUpdateIsLatest(t *testing.T) {
// 	Convey("TestUpdateIsLatest: 据算子ID和版本更新是否为最新版本", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		om := &operatorManagerDB{
// 			CommonDB: CommonDB{
// 				dbPool: mockDB,
// 				dbName: dbName,
// 				logger: logger.DefaultLogger(),
// 			},
// 		}
// 		query := "UPDATE `%s`.`t_op_registry` SET `f_is_latest` =? WHERE `f_op_id` =? AND `f_metadata_version` =?"
// 		query = fmt.Sprintf(query, om.dbName)
// 		Convey("更新成功", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(1, 1))
// 			err := om.UpdateIsLatest(context.Background(), nil, "", "", true)
// 			So(err, ShouldBeNil)
// 		})
// 		Convey("更新失败", func() {
// 			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(errMock)
// 			err := om.UpdateIsLatest(context.Background(), nil, "", "", false)
// 			So(err, ShouldNotBeNil)
// 		})
// 	})
// }

// // tx.go

// func TestGetTx(t *testing.T) {
// 	Convey("TestGetTx: 获取事务", t, func() {
// 		mockDB, mock, err := sqlx.New()
// 		So(err, ShouldBeNil)
// 		defer func() {
// 			_ = mockDB.Close()
// 		}()
// 		p := gomonkey.ApplyFunc(db.NewDBPool, func() *sqlx.DB {
// 			return mockDB
// 		})
// 		defer p.Reset()
// 		baseTx := NewBaseTx()
// 		mock.ExpectBegin().WillReturnError(errMock)
// 		_, err = baseTx.GetTx(context.Background())
// 		So(err, ShouldNotBeNil)
// 	})
// }
