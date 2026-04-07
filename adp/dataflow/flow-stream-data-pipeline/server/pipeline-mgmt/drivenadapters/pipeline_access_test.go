package drivenadapters

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	libcomm "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
)

var (
	testCtx = context.WithValue(context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage),
		interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
			ID:   "visitor.TokenID",
			Type: "user",
		})

	testNow = time.Now().UnixMilli()

	ppTags    = []string{"a", "b", "c", "d", "e"}
	ppTagsStr = libcomm.TagSlice2TagString(ppTags)

	testdeployConfig = &interfaces.DeploymentConfig{
		CpuLimit:    1,
		MemoryLimit: 2048,
	}

	testdeployConfigBytes, _ = sonic.Marshal(testdeployConfig)
)

func MockNewPipelineAccess() (*pipelineMgmtAccess, sqlmock.Sqlmock) {
	db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	pma := &pipelineMgmtAccess{
		db: db,
	}

	return pma, smock
}

func Test_PipelineAccess_CreatePipeline(t *testing.T) {
	Convey("Test CreatePipeline", t, func() {
		pma, smock := MockNewPipelineAccess()

		sqlStr := fmt.Sprintf("INSERT INTO %s "+
			"(f_builtin,f_comment,f_create_time,f_creator,f_creator_type,f_deployment_config,f_index_base,f_output_type,"+
			"f_pipeline_id,f_pipeline_name,f_pipeline_status,f_pipeline_status_details,f_tags,f_update_time,f_updater,"+
			"f_updater_type,f_use_index_base_in_data) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", PIPELINE_TABLE_NAME)

		pp := &interfaces.Pipeline{
			PipelineID:         "x",
			PipelineName:       "x",
			Tags:               []string{"a"},
			Comment:            "",
			Builtin:            false,
			OutputType:         "index_base",
			IndexBase:          "",
			UseIndexBaseInData: true,
			DeploymentConfig: &interfaces.DeploymentConfig{
				CpuLimit:    1,
				MemoryLimit: 2048,
			},
			PipelineStatus: interfaces.PipelineStatus_Running,
			CreateTime:     testNow,
			UpdateTime:     testNow,
			Creator: interfaces.AccountInfo{
				ID:   "visitor.TokenID",
				Type: "user",
			},
			Updater: interfaces.AccountInfo{
				ID:   "visitor.TokenID",
				Type: "user",
			},
		}

		Convey("Create failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.InsertBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := pma.CreatePipeline(testCtx, pp)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := pma.CreatePipeline(testCtx, pp)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Create succeed", func() {
			// WithArgs()中不传参数, 应该代表任意参数下都会被 smock
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := pma.CreatePipeline(testCtx, pp)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_PipelineAccess_DeletePipeline(t *testing.T) {
	Convey("Test DeletePipeline", t, func() {
		pma, smock := MockNewPipelineAccess()

		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_pipeline_id = ?", PIPELINE_TABLE_NAME)

		Convey("Delete failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.DeleteBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := pma.DeletePipeline(testCtx, "1")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := pma.DeletePipeline(testCtx, "1")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Delete succeed", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 2))

			err := pma.DeletePipeline(testCtx, "1")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_PipelineAccess_UpdatePipeline(t *testing.T) {
	Convey("Test UpdatePipeline", t, func() {
		pma, smock := MockNewPipelineAccess()

		sqlStr := fmt.Sprintf("UPDATE %s SET f_comment = ?, f_deployment_config = ?, f_index_base = ?, "+
			"f_pipeline_name = ?, f_tags = ?, f_update_time = ?, f_updater = ?, f_updater_type = ?, f_use_index_base_in_data = ? "+
			"WHERE f_pipeline_id = ?", PIPELINE_TABLE_NAME)

		pp := &interfaces.Pipeline{
			PipelineID:         "x",
			PipelineName:       "x",
			Tags:               []string{"a"},
			Comment:            "",
			Builtin:            false,
			OutputType:         "index_base",
			IndexBase:          "",
			UseIndexBaseInData: true,
			DeploymentConfig: &interfaces.DeploymentConfig{
				CpuLimit:    1,
				MemoryLimit: 2048,
			},
			PipelineStatus: interfaces.PipelineStatus_Running,
			CreateTime:     testNow,
			UpdateTime:     testNow,
			Creator: interfaces.AccountInfo{
				ID:   "visitor.TokenID",
				Type: "user",
			},
			Updater: interfaces.AccountInfo{
				ID:   "visitor.TokenID",
				Type: "user",
			},
		}

		Convey("Update failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := pma.UpdatePipeline(testCtx, pp)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update failed, caused by exec sql error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := pma.UpdatePipeline(testCtx, pp)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update succeed", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := pma.UpdatePipeline(testCtx, pp)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_PipelineAccess_GetPipeline(t *testing.T) {
	Convey("Test GetPipeline", t, func() {
		pma, smock := MockNewPipelineAccess()

		pp := &interfaces.Pipeline{
			PipelineID:         "x",
			PipelineName:       "x",
			Comment:            "",
			Builtin:            false,
			OutputType:         "index_base",
			IndexBase:          "",
			UseIndexBaseInData: true,

			PipelineStatus: interfaces.PipelineStatus_Running,
			CreateTime:     testNow,
			UpdateTime:     testNow,
			Creator: interfaces.AccountInfo{
				ID:   "visitor.TokenID",
				Type: "user",
			},
			Updater: interfaces.AccountInfo{
				ID:   "visitor.TokenID",
				Type: "user",
			},
		}

		rows1 := sqlmock.NewRows([]string{
			"f_pipeline_id",
			"f_pipeline_name",
			"f_tags",
			"f_comment",
			"f_builtin",
			"f_output_type",
			"f_index_base",
			"f_use_index_base_in_data",
			"f_pipeline_status",
			"f_pipeline_status_details",
			"f_deployment_config",
			"f_create_time",
			"f_update_time",
			"f_creator",
			"f_creator_type",
			"f_updater",
			"f_updater_type",
		}).AddRow(
			pp.PipelineID,
			pp.PipelineName,
			ppTagsStr,
			pp.Comment,
			pp.Builtin,
			pp.OutputType,
			pp.IndexBase,
			pp.UseIndexBaseInData,
			pp.PipelineStatus,
			pp.PipelineStatusDetails,
			testdeployConfigBytes,
			pp.CreateTime,
			pp.UpdateTime,
			pp.Creator.ID,
			pp.Creator.Type,
			pp.Updater.ID,
			pp.Updater.Type,
		)

		sqlStr := fmt.Sprintf("SELECT f_pipeline_id, f_pipeline_name, f_tags, f_comment, f_builtin, "+
			"f_output_type, f_index_base, f_use_index_base_in_data, f_pipeline_status, f_pipeline_status_details, "+
			"f_deployment_config, f_create_time, f_update_time, f_creator, f_creator_type, f_updater, f_updater_type FROM %s "+
			"WHERE f_pipeline_id = ?", PIPELINE_TABLE_NAME)

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, _, err := pma.GetPipeline(testCtx, "a")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, _, err := pma.GetPipeline(testCtx, "a")
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by row scan error", func() {
			rows := sqlmock.NewRows([]string{"f_pipeline_id", "f_pipeline_name"}).
				AddRow(pp.PipelineID, pp.PipelineName)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, _, err := pma.GetPipeline(testCtx, "a")
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed", func() {
			sqlStr := fmt.Sprintf("SELECT f_pipeline_id, f_pipeline_name, f_tags, f_comment, f_builtin, "+
				"f_output_type, f_index_base, f_use_index_base_in_data, f_pipeline_status, f_pipeline_status_details, "+
				"f_deployment_config, f_create_time, f_update_time, f_creator, f_creator_type, f_updater, f_updater_type FROM %s "+
				"WHERE f_pipeline_id = ?", PIPELINE_TABLE_NAME)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows1)

			_, _, err := pma.GetPipeline(testCtx, "a")
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_PipelineAccess_ListPipeline(t *testing.T) {
	Convey("Test ListPipeline", t, func() {
		pma, smock := MockNewPipelineAccess()

		pp := &interfaces.Pipeline{
			PipelineID:         "x",
			PipelineName:       "x",
			Comment:            "",
			Builtin:            false,
			OutputType:         "index_base",
			IndexBase:          "",
			UseIndexBaseInData: true,

			PipelineStatus: interfaces.PipelineStatus_Running,
			CreateTime:     testNow,
			UpdateTime:     testNow,
			Creator: interfaces.AccountInfo{
				ID:   "visitor.TokenID",
				Type: "user",
			},
			Updater: interfaces.AccountInfo{
				ID:   "visitor.TokenID",
				Type: "user",
			},
		}

		rows1 := sqlmock.NewRows([]string{
			"f_pipeline_id",
			"f_pipeline_name",
			"f_tags",
			"f_comment",
			"f_builtin",
			"f_output_type",
			"f_index_base",
			"f_use_index_base_in_data",
			"f_pipeline_status",
			"f_pipeline_status_details",
			"f_deployment_config",
			"f_create_time",
			"f_update_time",
			"f_creator",
			"f_creator_type",
			"f_updater",
			"f_updater_type",
		}).AddRow(
			pp.PipelineID,
			pp.PipelineName,
			ppTagsStr,
			pp.Comment,
			pp.Builtin,
			pp.OutputType,
			pp.IndexBase,
			pp.UseIndexBaseInData,
			pp.PipelineStatus,
			pp.PipelineStatusDetails,
			testdeployConfigBytes,
			pp.CreateTime,
			pp.UpdateTime,
			pp.Creator.ID,
			pp.Creator.Type,
			pp.Updater.ID,
			pp.Updater.Type,
		)
		rows2 := sqlmock.NewRows([]string{
			"f_pipeline_id",
			"f_pipeline_name",
			"f_tags",
			"f_comment",
			"f_builtin",
			"f_output_type",
			"f_index_base",
			"f_use_index_base_in_data",
			"f_pipeline_status",
			"f_pipeline_status_details",
			"f_deployment_config",
			"f_create_time",
			"f_update_time",
			"f_creator",
			"f_creator_type",
			"f_updater",
			"f_updater_type",
		}).AddRow(
			pp.PipelineID,
			pp.PipelineName,
			ppTagsStr,
			pp.Comment,
			pp.Builtin,
			pp.OutputType,
			pp.IndexBase,
			pp.UseIndexBaseInData,
			pp.PipelineStatus,
			pp.PipelineStatusDetails,
			testdeployConfigBytes,
			pp.CreateTime,
			pp.UpdateTime,
			pp.Creator.ID,
			pp.Creator.Type,
			pp.Updater.ID,
			pp.Updater.Type,
		).AddRow(
			pp.PipelineID,
			pp.PipelineName,
			ppTagsStr,
			pp.Comment,
			pp.Builtin,
			pp.OutputType,
			pp.IndexBase,
			pp.UseIndexBaseInData,
			pp.PipelineStatus,
			pp.PipelineStatusDetails,
			testdeployConfigBytes,
			pp.CreateTime,
			pp.UpdateTime,
			pp.Creator.ID,
			pp.Creator.Type,
			pp.Updater.ID,
			pp.Updater.Type,
		)

		sqlStr := fmt.Sprintf("SELECT f_pipeline_id, f_pipeline_name, f_tags, f_comment, f_builtin, "+
			"f_output_type, f_index_base, f_use_index_base_in_data, f_pipeline_status, f_pipeline_status_details, "+
			"f_deployment_config, f_create_time, f_update_time, f_creator, f_creator_type, f_updater, f_updater_type FROM %s "+
			"WHERE f_builtin IN (?) ORDER BY f_update_time desc", PIPELINE_TABLE_NAME)

		listQuery := &interfaces.ListPipelinesQuery{
			Builtin: []bool{false},
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.TABLE_SORT["update_time"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}

		Convey("List failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, err := pma.ListPipelines(testCtx, listQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, err := pma.ListPipelines(testCtx, listQuery)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List failed, caused by row scan error", func() {
			rows := sqlmock.NewRows([]string{"f_pipeline_id", "f_pipeline_name"}).
				AddRow(pp.PipelineID, pp.PipelineName)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, err := pma.ListPipelines(testCtx, listQuery)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List succeed with limit is equal to 1", func() {
			sqlStr := fmt.Sprintf("SELECT f_pipeline_id, f_pipeline_name, f_tags, f_comment, f_builtin, "+
				"f_output_type, f_index_base, f_use_index_base_in_data, f_pipeline_status, f_pipeline_status_details, "+
				"f_deployment_config, f_create_time, f_update_time, f_creator, f_creator_type, f_updater, f_updater_type FROM %s "+
				"WHERE f_builtin IN (?) ORDER BY f_update_time desc", PIPELINE_TABLE_NAME)

			listQuery.Limit = 1
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows1)

			entries, err := pma.ListPipelines(testCtx, listQuery)
			So(err, ShouldBeNil)
			So(len(entries), ShouldEqual, 1)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("List succeed with no limit", func() {
			sqlStr := fmt.Sprintf("SELECT f_pipeline_id, f_pipeline_name, f_tags, f_comment, f_builtin, "+
				"f_output_type, f_index_base, f_use_index_base_in_data, f_pipeline_status, f_pipeline_status_details, "+
				"f_deployment_config, f_create_time, f_update_time, f_creator, f_creator_type, f_updater, f_updater_type FROM %s "+
				"WHERE f_builtin IN (?) ORDER BY f_update_time desc", PIPELINE_TABLE_NAME)

			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows2)

			listQuery.Limit = -1
			entries, err := pma.ListPipelines(testCtx, listQuery)
			So(err, ShouldBeNil)
			So(len(entries), ShouldEqual, 2)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_PipelineAccess_GetPipelinesTotal(t *testing.T) {
	Convey("Test GetPipelinesTotal", t, func() {
		pmaMock, smock := MockNewPipelineAccess()

		sqlStr := fmt.Sprintf("SELECT COUNT(f_pipeline_id) FROM %s "+
			"WHERE f_builtin IN (?)", PIPELINE_TABLE_NAME)

		rows := sqlmock.NewRows([]string{"COUNT(f_pipeline_id)"}).AddRow(1)

		listQuery := &interfaces.ListPipelinesQuery{
			Builtin: []bool{true},
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit:     interfaces.MAX_LIMIT,
				Offset:    interfaces.MIN_OFFSET,
				Sort:      interfaces.TABLE_SORT["name"],
				Direction: interfaces.DESC_DIRECTION,
			},
		}

		Convey("Get failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			total, err := pmaMock.GetPipelinesTotal(testCtx, listQuery)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by the scan error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			total, err := pmaMock.GetPipelinesTotal(testCtx, listQuery)
			So(total, ShouldEqual, 0)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get succeed", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)
			total, err := pmaMock.GetPipelinesTotal(testCtx, listQuery)

			So(total, ShouldEqual, 1)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_PipelineAccess_UpdatePipelineStatus(t *testing.T) {
	Convey("Test UpdatePipelineStatus", t, func() {
		pma, smock := MockNewPipelineAccess()

		sqlStr := fmt.Sprintf("UPDATE %s SET f_pipeline_status = ?, f_pipeline_status_details = ? WHERE f_pipeline_id = ?", PIPELINE_TABLE_NAME)

		pp := &interfaces.Pipeline{
			PipelineID:            "x",
			PipelineStatus:        interfaces.PipelineStatus_Running,
			PipelineStatusDetails: "test",
		}

		Convey("Update failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.UpdateBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			err := pma.UpdatePipelineStatus(testCtx, pp, true)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update failed, caused by execute error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectExec(sqlStr).WithArgs().WillReturnError(expectedErr)

			err := pma.UpdatePipelineStatus(testCtx, pp, true)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Update succeed", func() {
			smock.ExpectExec(sqlStr).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

			err := pma.UpdatePipelineStatus(testCtx, pp, true)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_PipelineAccess_CheckPipelineExistByID(t *testing.T) {
	Convey("Test CheckPipelineExistByID", t, func() {
		pma, smock := MockNewPipelineAccess()

		sqlStr := fmt.Sprintf("SELECT f_pipeline_name FROM %s WHERE f_pipeline_id = ?", PIPELINE_TABLE_NAME)

		pp := &interfaces.Pipeline{
			PipelineID:         "x",
			PipelineName:       "x",
			Tags:               []string{"a"},
			Comment:            "",
			Builtin:            false,
			OutputType:         "index_base",
			IndexBase:          "",
			UseIndexBaseInData: true,
			DeploymentConfig: &interfaces.DeploymentConfig{
				CpuLimit:    1,
				MemoryLimit: 2048,
			},
			PipelineStatus: interfaces.PipelineStatus_Running,
			CreateTime:     testNow,
			UpdateTime:     testNow,
		}

		rows := sqlmock.NewRows([]string{"f_group_id"}).AddRow(pp.PipelineID)

		Convey("Check failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, _, err := pma.CheckPipelineExistByID(testCtx, pp.PipelineID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, _, err := pma.CheckPipelineExistByID(testCtx, pp.PipelineID)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check failed, caused by the scan error", func() {
			rows = sqlmock.NewRows([]string{"f_group_id", "f_group_name"}).
				AddRow(pp.PipelineID, pp.PipelineName)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, _, err := pma.CheckPipelineExistByID(testCtx, pp.PipelineID)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by ErrNoRows", func() {
			expectedErr := sql.ErrNoRows
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, exist, err := pma.CheckPipelineExistByID(testCtx, pp.PipelineID)
			So(exist, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check succeed", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, exist, err := pma.CheckPipelineExistByID(testCtx, pp.PipelineID)
			So(exist, ShouldBeTrue)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func Test_PipelineAccess_CheckPipelineExistByName(t *testing.T) {
	Convey("Test CheckPipelineExistByName", t, func() {
		pma, smock := MockNewPipelineAccess()

		sqlStr := fmt.Sprintf("SELECT f_pipeline_id FROM %s WHERE f_pipeline_name = ?", PIPELINE_TABLE_NAME)

		pp := &interfaces.Pipeline{
			PipelineID:         "x",
			PipelineName:       "x",
			Tags:               []string{"a"},
			Comment:            "",
			Builtin:            false,
			OutputType:         "index_base",
			IndexBase:          "",
			UseIndexBaseInData: true,
			DeploymentConfig: &interfaces.DeploymentConfig{
				CpuLimit:    1,
				MemoryLimit: 2048,
			},
			PipelineStatus: interfaces.PipelineStatus_Running,
			CreateTime:     testNow,
			UpdateTime:     testNow,
		}

		rows := sqlmock.NewRows([]string{"f_group_id"}).AddRow(pp.PipelineID)

		Convey("Check failed, caused by the error from squirrel func ToSql", func() {
			expectedErr := errors.New("some error")
			patch := ApplyMethodReturn(sq.SelectBuilder{}, "ToSql", "", []interface{}{}, expectedErr)
			defer patch.Reset()

			_, _, err := pma.CheckPipelineExistByName(testCtx, pp.PipelineName)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by query error", func() {
			expectedErr := errors.New("some error")
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, _, err := pma.CheckPipelineExistByName(testCtx, pp.PipelineName)
			So(err, ShouldResemble, expectedErr)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check failed, caused by the scan error", func() {
			rows = sqlmock.NewRows([]string{"f_group_id", "f_group_name"}).
				AddRow(pp.PipelineID, pp.PipelineName)
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, _, err := pma.CheckPipelineExistByName(testCtx, pp.PipelineName)
			So(err, ShouldNotBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Get failed, caused by ErrNoRows", func() {
			expectedErr := sql.ErrNoRows
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnError(expectedErr)

			_, exist, err := pma.CheckPipelineExistByName(testCtx, pp.PipelineName)
			So(exist, ShouldBeFalse)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("Check succeed", func() {
			smock.ExpectQuery(sqlStr).WithArgs().WillReturnRows(rows)

			_, exist, err := pma.CheckPipelineExistByName(testCtx, pp.PipelineName)
			So(exist, ShouldBeTrue)
			So(err, ShouldBeNil)

			if err := smock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
